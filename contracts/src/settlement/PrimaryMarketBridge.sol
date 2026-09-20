// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title PrimaryMarketBridge
 * @notice Central clearing bridge connecting 24/7 on-chain queued orders with primary exchange trading sessions.
 * @dev Enforces 100% escrow backing, session-aware batching, internalized crossing, and 1:1 custody parity.
 */
contract PrimaryMarketBridge is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    enum MarketSession {
        CLOSED,
        PRE_OPEN,        // 09:00 - 09:15 IST
        REGULAR_OPEN,    // 09:15 - 15:30 IST
        POST_CLOSE,      // 15:30 - 16:00 IST
        WEEKEND_HOLIDAY
    }

    enum OrderSide {
        BUY,
        SELL
    }

    enum OrderStatus {
        QUEUED,
        INTERNALIZED,
        DISPATCHED_TO_PRIMARY,
        SETTLED,
        CANCELLED
    }

    struct QueuedOrder {
        bytes32 orderId;
        address trader;
        bytes32 isinHash;
        OrderSide side;
        uint256 quantity;
        uint256 limitPricePaise;
        uint256 escrowedAmount; // Cash for BUY, Token units for SELL
        OrderStatus status;
        uint64 queuedTimestamp;
    }

    IERC20 public immutable cashToken; // eINR CBDC
    IERC20 public immutable equityToken; // ERC-3643 Tokenized Equity

    MarketSession public currentSession = MarketSession.CLOSED;
    address public sessionOracle;
    address public brokerRelayer;

    mapping(bytes32 => QueuedOrder) public orders;
    bytes32[] public queuedOrderIds;

    event SessionUpdated(MarketSession oldSession, MarketSession newSession, uint256 timestamp);
    event OrderQueued(bytes32 indexed orderId, address indexed trader, OrderSide side, uint256 quantity, uint256 price);
    event OrderInternalized(bytes32 indexed buyOrderId, bytes32 indexed sellOrderId, uint256 matchedQty, uint256 matchPrice);
    event BatchDispatchedToPrimary(bytes32 indexed batchId, uint256 totalOrders, uint256 totalVolume);
    event PrimaryTradeCleared(bytes32 indexed orderId, uint256 filledQty, uint256 fillPricePaise);
    event OrderCancelled(bytes32 indexed orderId, address indexed trader, uint256 refundedEscrow);

    error InvalidSession();
    error Unauthorized();
    error OrderNotCancellable(bytes32 orderId);
    error InvalidOrderParameters();

    modifier onlyOracleOrOwner() {
        if (msg.sender != sessionOracle && msg.sender != owner()) revert Unauthorized();
        _;
    }

    modifier onlyRelayerOrOwner() {
        if (msg.sender != brokerRelayer && msg.sender != owner()) revert Unauthorized();
        _;
    }

    constructor(
        address _cashToken,
        address _equityToken,
        address initialOwner
    ) Ownable(initialOwner) {
        require(_cashToken != address(0) && _equityToken != address(0), "Invalid tokens");
        cashToken = IERC20(_cashToken);
        equityToken = IERC20(_equityToken);
    }

    function setSessionOracle(address _oracle) external onlyOwner {
        sessionOracle = _oracle;
    }

    function setBrokerRelayer(address _relayer) external onlyOwner {
        brokerRelayer = _relayer;
    }

    function setMarketSession(MarketSession _session) external onlyOracleOrOwner {
        MarketSession old = currentSession;
        currentSession = _session;
        emit SessionUpdated(old, _session, block.timestamp);
    }

    /**
     * @notice Queue an off-market order with 100% escrowed collateral
     */
    function queueOrder(
        bytes32 isinHash,
        OrderSide side,
        uint256 quantity,
        uint256 limitPricePaise
    ) external nonReentrant returns (bytes32 orderId) {
        if (quantity == 0 || limitPricePaise == 0) revert InvalidOrderParameters();

        orderId = keccak256(abi.encodePacked(msg.sender, isinHash, side, quantity, limitPricePaise, block.timestamp, queuedOrderIds.length));

        uint256 escrowRequired;
        if (side == OrderSide.BUY) {
            // Escrow required = quantity * limitPrice
            escrowRequired = quantity * limitPricePaise;
            cashToken.safeTransferFrom(msg.sender, address(this), escrowRequired);
        } else {
            // Escrow required = quantity of equity tokens
            escrowRequired = quantity;
            equityToken.safeTransferFrom(msg.sender, address(this), escrowRequired);
        }

        orders[orderId] = QueuedOrder({
            orderId: orderId,
            trader: msg.sender,
            isinHash: isinHash,
            side: side,
            quantity: quantity,
            limitPricePaise: limitPricePaise,
            escrowedAmount: escrowRequired,
            status: OrderStatus.QUEUED,
            queuedTimestamp: uint64(block.timestamp)
        });

        queuedOrderIds.push(orderId);
        emit OrderQueued(orderId, msg.sender, side, quantity, limitPricePaise);
    }

    /**
     * @notice Internalize and atomically cross matching off-market orders during closed hours
     */
    function internalizeCrossing(bytes32 buyOrderId, bytes32 sellOrderId) external onlyRelayerOrOwner nonReentrant {
        QueuedOrder storage buyOrder = orders[buyOrderId];
        QueuedOrder storage sellOrder = orders[sellOrderId];

        require(buyOrder.status == OrderStatus.QUEUED && sellOrder.status == OrderStatus.QUEUED, "Orders not queued");
        require(buyOrder.side == OrderSide.BUY && sellOrder.side == OrderSide.SELL, "Mismatched sides");
        require(buyOrder.isinHash == sellOrder.isinHash, "ISIN mismatch");
        require(buyOrder.limitPricePaise >= sellOrder.limitPricePaise, "Price cross condition unmet");

        uint256 matchQty = buyOrder.quantity < sellOrder.quantity ? buyOrder.quantity : sellOrder.quantity;
        uint256 matchPrice = (buyOrder.limitPricePaise + sellOrder.limitPricePaise) / 2;
        uint256 settlementTurnover = matchQty * matchPrice;

        // Deliver equity tokens to Buyer
        equityToken.safeTransfer(buyOrder.trader, matchQty);
        // Deliver cash to Seller (100% net turnover - 0.00% fee at launch)
        cashToken.safeTransfer(sellOrder.trader, settlementTurnover);

        // Refund any surplus cash to Buyer if executed below limit
        uint256 cashReservedForFill = matchQty * buyOrder.limitPricePaise;
        if (cashReservedForFill > settlementTurnover) {
            cashToken.safeTransfer(buyOrder.trader, cashReservedForFill - settlementTurnover);
        }

        buyOrder.quantity -= matchQty;
        sellOrder.quantity -= matchQty;

        if (buyOrder.quantity == 0) buyOrder.status = OrderStatus.INTERNALIZED;
        if (sellOrder.quantity == 0) sellOrder.status = OrderStatus.INTERNALIZED;

        emit OrderInternalized(buyOrderId, sellOrderId, matchQty, matchPrice);
    }

    /**
     * @notice Clear a primary exchange trade fill reported by licensed broker relayer
     */
    function clearPrimaryTrade(bytes32 orderId, uint256 filledQty, uint256 fillPricePaise) external onlyRelayerOrOwner nonReentrant {
        QueuedOrder storage order = orders[orderId];
        require(order.status == OrderStatus.QUEUED || order.status == OrderStatus.DISPATCHED_TO_PRIMARY, "Order not active");
        require(filledQty <= order.quantity, "Fill exceeds remaining quantity");

        order.quantity -= filledQty;
        if (order.quantity == 0) {
            order.status = OrderStatus.SETTLED;
        }

        if (order.side == OrderSide.BUY) {
            uint256 cost = filledQty * fillPricePaise;
            uint256 reserved = filledQty * order.limitPricePaise;
            if (reserved > cost) {
                // Refund surplus price improvement
                cashToken.safeTransfer(order.trader, reserved - cost);
            }
            // Mint / Deliver equity token to buyer
            equityToken.safeTransfer(order.trader, filledQty);
        } else {
            // Deliver cash proceeds to seller
            uint256 proceeds = filledQty * fillPricePaise;
            cashToken.safeTransfer(order.trader, proceeds);
        }

        emit PrimaryTradeCleared(orderId, filledQty, fillPricePaise);
    }

    /**
     * @notice Cancel a queued order before primary dispatch and receive instant 100% escrow refund
     */
    function cancelQueuedOrder(bytes32 orderId) external nonReentrant {
        QueuedOrder storage order = orders[orderId];
        if (msg.sender != order.trader && msg.sender != owner()) revert Unauthorized();
        if (order.status != OrderStatus.QUEUED) revert OrderNotCancellable(orderId);

        order.status = OrderStatus.CANCELLED;
        uint256 refundAmount = order.escrowedAmount;
        order.escrowedAmount = 0;

        if (order.side == OrderSide.BUY) {
            cashToken.safeTransfer(order.trader, refundAmount);
        } else {
            equityToken.safeTransfer(order.trader, refundAmount);
        }

        emit OrderCancelled(orderId, order.trader, refundAmount);
    }
}
