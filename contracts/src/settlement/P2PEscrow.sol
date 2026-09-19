// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title P2PEscrow
 * @notice P2P Fiat-to-USDT Escrow Contract with Dispute Arbitration Desk
 * @dev Protects buyer and seller during peer-to-peer bank/UPI fiat transfers
 */
contract P2PEscrow {
    address public immutable arbitrator;
    
    enum EscrowStatus {
        CREATED,
        FUNDED,
        FIAT_PAID,
        COMPLETED,
        DISPUTED,
        RESOLVED_TO_BUYER,
        RESOLVED_TO_SELLER,
        CANCELLED
    }

    struct EscrowTrade {
        bytes32 tradeId;
        address payable seller;
        address payable buyer;
        uint256 amount;
        uint256 paymentWindowExpiry;
        EscrowStatus status;
        string fiatReference;
    }

    mapping(bytes32 => EscrowTrade) public escrows;

    event EscrowCreated(bytes32 indexed tradeId, address indexed seller, address indexed buyer, uint256 amount);
    event FiatPaymentMarked(bytes32 indexed tradeId, string fiatReference);
    event EscrowReleased(bytes32 indexed tradeId);
    event EscrowDisputed(bytes32 indexed tradeId, address raisedBy);
    event DisputeResolved(bytes32 indexed tradeId, EscrowStatus finalOutcome);
    event EscrowCancelled(bytes32 indexed tradeId);

    modifier onlyArbitrator() {
        require(msg.sender == arbitrator, "Only arbitrator allowed");
        _;
    }

    constructor(address _arbitrator) {
        require(_arbitrator != address(0), "Invalid arbitrator");
        arbitrator = _arbitrator;
    }

    /**
     * @notice Seller creates and funds escrow
     */
    function createEscrow(bytes32 tradeId, address payable buyer, uint256 paymentWindowSeconds) external payable {
        require(msg.value > 0, "No funds locked");
        require(buyer != address(0) && buyer != msg.sender, "Invalid buyer");
        require(escrows[tradeId].amount == 0, "Trade ID already exists");
        require(paymentWindowSeconds >= 15 minutes, "Window too short");

        escrows[tradeId] = EscrowTrade({
            tradeId: tradeId,
            seller: payable(msg.sender),
            buyer: buyer,
            amount: msg.value,
            paymentWindowExpiry: block.timestamp + paymentWindowSeconds,
            status: EscrowStatus.FUNDED,
            fiatReference: ""
        });

        emit EscrowCreated(tradeId, msg.sender, buyer, msg.value);
    }

    /**
     * @notice Buyer marks fiat payment completed (e.g. UPI / IMPS reference)
     */
    function markFiatPaid(bytes32 tradeId, string calldata fiatRef) external {
        EscrowTrade storage trade = escrows[tradeId];
        require(msg.sender == trade.buyer, "Only buyer can mark paid");
        require(trade.status == EscrowStatus.FUNDED, "Invalid status");
        require(block.timestamp <= trade.paymentWindowExpiry, "Payment window expired");

        trade.status = EscrowStatus.FIAT_PAID;
        trade.fiatReference = fiatRef;

        emit FiatPaymentMarked(tradeId, fiatRef);
    }

    /**
     * @notice Seller confirms fiat receipt and releases escrowed funds to buyer
     */
    function releaseEscrow(bytes32 tradeId) external {
        EscrowTrade storage trade = escrows[tradeId];
        require(msg.sender == trade.seller, "Only seller can release");
        require(trade.status == EscrowStatus.FIAT_PAID || trade.status == EscrowStatus.FUNDED, "Cannot release");

        trade.status = EscrowStatus.COMPLETED;
        (bool sent, ) = trade.buyer.call{value: trade.amount}("");
        require(sent, "Payout failed");

        emit EscrowReleased(tradeId);
    }

    /**
     * @notice Either party can open a dispute if fiat payment is contested
     */
    function raiseDispute(bytes32 tradeId) external {
        EscrowTrade storage trade = escrows[tradeId];
        require(msg.sender == trade.buyer || msg.sender == trade.seller, "Unauthorized party");
        require(trade.status == EscrowStatus.FIAT_PAID || trade.status == EscrowStatus.FUNDED, "Cannot dispute");

        trade.status = EscrowStatus.DISPUTED;
        emit EscrowDisputed(tradeId, msg.sender);
    }

    /**
     * @notice Arbitrator settles dispute after inspecting bank statement proof
     */
    function resolveDispute(bytes32 tradeId, bool releaseToBuyer) external onlyArbitrator {
        EscrowTrade storage trade = escrows[tradeId];
        require(trade.status == EscrowStatus.DISPUTED, "Not disputed");

        if (releaseToBuyer) {
            trade.status = EscrowStatus.RESOLVED_TO_BUYER;
            (bool sent, ) = trade.buyer.call{value: trade.amount}("");
            require(sent, "Transfer to buyer failed");
        } else {
            trade.status = EscrowStatus.RESOLVED_TO_SELLER;
            (bool sent, ) = trade.seller.call{value: trade.amount}("");
            require(sent, "Refund to seller failed");
        }

        emit DisputeResolved(tradeId, trade.status);
    }

    /**
     * @notice Seller can cancel unfunded / unpaid expired trade
     */
    function cancelExpired(bytes32 tradeId) external {
        EscrowTrade storage trade = escrows[tradeId];
        require(msg.sender == trade.seller, "Only seller can cancel");
        require(trade.status == EscrowStatus.FUNDED, "Cannot cancel in current state");
        require(block.timestamp > trade.paymentWindowExpiry, "Payment window still active");

        trade.status = EscrowStatus.CANCELLED;
        (bool sent, ) = trade.seller.call{value: trade.amount}("");
        require(sent, "Refund failed");

        emit EscrowCancelled(tradeId);
    }
}
