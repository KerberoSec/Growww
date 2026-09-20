// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "../interfaces/settlement/ICommitRevealSettlement.sol";

/**
 * @title CommitRevealSettlement
 * @notice Two-phase commit-reveal order settlement contract preventing MEV, front-running, and sandwich attacks on DEX orders.
 * @dev Enforces cryptographic commitments, delay windows, slashing for unrevealed orders, and atomic batch settlement.
 */
contract CommitRevealSettlement is
    ICommitRevealSettlement,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable,
    PausableUpgradeable
{
    using SafeERC20 for IERC20;

    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant EMERGENCY_GUARDIAN_ROLE = keccak256("EMERGENCY_GUARDIAN_ROLE");
    bytes32 public constant SETTLEMENT_OPERATOR_ROLE = keccak256("SETTLEMENT_OPERATOR_ROLE");

    address public bondToken;
    uint256 public minCommitBond;
    uint256 public minBlockDelay;
    uint256 public maxBlockDelay;
    address public treasury;

    mapping(bytes32 => OrderCommitment) private _commitments;
    mapping(bytes32 => RevealedOrder) private _revealedOrders;
    mapping(address => mapping(uint256 => bool)) public override isNonceUsed;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        address _bondToken,
        uint256 _minCommitBond,
        uint256 _minBlockDelay,
        uint256 _maxBlockDelay,
        address _treasury
    ) external initializer {
        if (_bondToken == address(0) || _treasury == address(0)) revert InvalidZeroAddress();

        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        __Pausable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(EMERGENCY_GUARDIAN_ROLE, admin);
        _grantRole(SETTLEMENT_OPERATOR_ROLE, admin);

        bondToken = _bondToken;
        minCommitBond = _minCommitBond;
        minBlockDelay = _minBlockDelay;
        maxBlockDelay = _maxBlockDelay;
        treasury = _treasury;
    }

    function setDelayParameters(uint256 _minBlockDelay, uint256 _maxBlockDelay) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(_maxBlockDelay > _minBlockDelay, "Invalid delay bounds");
        minBlockDelay = _minBlockDelay;
        maxBlockDelay = _maxBlockDelay;
    }

    function setMinCommitBond(uint256 _minCommitBond) external onlyRole(DEFAULT_ADMIN_ROLE) {
        minCommitBond = _minCommitBond;
    }

    function setTreasury(address _treasury) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (_treasury == address(0)) revert InvalidZeroAddress();
        treasury = _treasury;
    }

    function pause() external onlyRole(EMERGENCY_GUARDIAN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    function computeCommitmentHash(
        address trader,
        OrderParams calldata order,
        bytes32 salt
    ) public pure override returns (bytes32) {
        return keccak256(
            abi.encode(
                trader,
                order.baseToken,
                order.quoteToken,
                order.side,
                order.amount,
                order.limitPrice,
                order.nonce,
                order.deadline,
                salt
            )
        );
    }

    function commitOrder(bytes32 commitmentHash, uint256 bondAmount) external override whenNotPaused nonReentrant {
        if (commitmentHash == bytes32(0)) revert InvalidCommitmentHash();
        if (bondAmount < minCommitBond) revert InsufficientCommitBond(bondAmount, minCommitBond);
        if (_commitments[commitmentHash].status != CommitmentStatus.UNINITIALIZED) {
            revert CommitmentAlreadyExists(commitmentHash);
        }

        IERC20(bondToken).safeTransferFrom(msg.sender, address(this), bondAmount);

        _commitments[commitmentHash] = OrderCommitment({
            commitmentHash: commitmentHash,
            trader: msg.sender,
            commitBlock: block.number,
            commitTimestamp: block.timestamp,
            commitBond: bondAmount,
            status: CommitmentStatus.COMMITTED
        });

        emit OrderCommitted(commitmentHash, msg.sender, bondAmount, block.number, block.timestamp);
    }

    function revealOrder(
        OrderParams calldata order,
        bytes32 salt
    ) external override whenNotPaused nonReentrant returns (bytes32 commitmentHash) {
        if (block.timestamp > order.deadline) {
            revert OrderExpired(order.deadline, block.timestamp);
        }
        if (isNonceUsed[msg.sender][order.nonce]) {
            revert NonceAlreadyUsed(msg.sender, order.nonce);
        }

        commitmentHash = computeCommitmentHash(msg.sender, order, salt);
        OrderCommitment storage commit = _commitments[commitmentHash];

        if (commit.status == CommitmentStatus.UNINITIALIZED) {
            revert CommitmentDoesNotExist(commitmentHash);
        }
        if (commit.status != CommitmentStatus.COMMITTED) {
            revert InvalidCommitmentHash();
        }
        if (commit.trader != msg.sender) {
            revert UnauthorizedTrader(msg.sender, commit.trader);
        }

        if (block.number < commit.commitBlock + minBlockDelay) {
            revert RevealTooEarly(block.number, commit.commitBlock + minBlockDelay);
        }
        if (block.number > commit.commitBlock + maxBlockDelay) {
            revert RevealTooLate(block.number, commit.commitBlock + maxBlockDelay);
        }

        isNonceUsed[msg.sender][order.nonce] = true;
        commit.status = CommitmentStatus.REVEALED;

        // Refund commit bond
        uint256 bond = commit.commitBond;
        if (bond > 0) {
            IERC20(bondToken).safeTransfer(msg.sender, bond);
        }

        _revealedOrders[commitmentHash] = RevealedOrder({
            commitmentHash: commitmentHash,
            trader: msg.sender,
            order: order,
            revealBlock: block.number,
            revealTimestamp: block.timestamp,
            isSettled: false
        });

        emit OrderRevealed(
            commitmentHash,
            msg.sender,
            order.baseToken,
            order.quoteToken,
            order.side,
            order.amount,
            order.limitPrice
        );
    }

    function slashExpiredCommitment(bytes32 commitmentHash) external override whenNotPaused nonReentrant {
        OrderCommitment storage commit = _commitments[commitmentHash];
        if (commit.status != CommitmentStatus.COMMITTED) {
            revert CommitmentNotSlashable(commitmentHash);
        }
        if (block.number <= commit.commitBlock + maxBlockDelay) {
            revert CommitmentNotSlashable(commitmentHash);
        }

        commit.status = CommitmentStatus.SLASHED;
        uint256 slashedAmount = commit.commitBond;

        if (slashedAmount > 0) {
            uint256 slasherReward = slashedAmount / 2;
            uint256 treasuryShare = slashedAmount - slasherReward;

            IERC20(bondToken).safeTransfer(msg.sender, slasherReward);
            IERC20(bondToken).safeTransfer(treasury, treasuryShare);
        }

        emit CommitmentSlashed(commitmentHash, commit.trader, msg.sender, slashedAmount);
    }

    function batchSettleOrders(
        bytes32[] calldata buyCommitmentHashes,
        bytes32[] calldata sellCommitmentHashes,
        uint256 clearingPrice
    ) external override onlyRole(SETTLEMENT_OPERATOR_ROLE) whenNotPaused nonReentrant {
        require(buyCommitmentHashes.length == sellCommitmentHashes.length, "Mismatched orders count");
        uint256 totalVolume = 0;

        for (uint256 i = 0; i < buyCommitmentHashes.length; i++) {
            bytes32 buyHash = buyCommitmentHashes[i];
            bytes32 sellHash = sellCommitmentHashes[i];

            RevealedOrder storage buyOrder = _revealedOrders[buyHash];
            RevealedOrder storage sellOrder = _revealedOrders[sellHash];

            if (buyOrder.commitmentHash == bytes32(0) || buyOrder.isSettled) {
                revert OrderNotRevealed(buyHash);
            }
            if (sellOrder.commitmentHash == bytes32(0) || sellOrder.isSettled) {
                revert OrderNotRevealed(sellHash);
            }

            require(buyOrder.order.side == OrderSide.BUY, "Invalid buy order");
            require(sellOrder.order.side == OrderSide.SELL, "Invalid sell order");
            require(buyOrder.order.baseToken == sellOrder.order.baseToken, "Base token mismatch");
            require(buyOrder.order.quoteToken == sellOrder.order.quoteToken, "Quote token mismatch");
            require(clearingPrice <= buyOrder.order.limitPrice, "Clearing price exceeds buy limit");
            require(clearingPrice >= sellOrder.order.limitPrice, "Clearing price below sell limit");

            uint256 fillAmount = buyOrder.order.amount < sellOrder.order.amount ? buyOrder.order.amount : sellOrder.order.amount;
            uint256 quoteAmount = (fillAmount * clearingPrice) / 1e18;

            buyOrder.isSettled = true;
            sellOrder.isSettled = true;
            _commitments[buyHash].status = CommitmentStatus.SETTLED;
            _commitments[sellHash].status = CommitmentStatus.SETTLED;

            // Atomic DvP Swap
            // Buyer sends quote tokens -> Seller receives quote tokens
            IERC20(buyOrder.order.quoteToken).safeTransferFrom(buyOrder.trader, sellOrder.trader, quoteAmount);
            // Seller sends base tokens -> Buyer receives base tokens
            IERC20(sellOrder.order.baseToken).safeTransferFrom(sellOrder.trader, buyOrder.trader, fillAmount);

            totalVolume += fillAmount;

            emit OrderSettled(buyHash, buyOrder.trader, sellOrder.trader, clearingPrice, fillAmount);
            emit OrderSettled(sellHash, sellOrder.trader, buyOrder.trader, clearingPrice, fillAmount);
        }

        emit BatchSettlementExecuted(buyCommitmentHashes, sellCommitmentHashes, clearingPrice, totalVolume);
    }

    function getCommitment(bytes32 commitmentHash) external view override returns (OrderCommitment memory) {
        return _commitments[commitmentHash];
    }

    function getRevealedOrder(bytes32 commitmentHash) external view override returns (RevealedOrder memory) {
        return _revealedOrders[commitmentHash];
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[45] private __gap;
}
