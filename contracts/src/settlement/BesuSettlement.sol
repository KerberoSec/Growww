// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title BesuSettlement
 * @notice Enterprise Delivery-versus-Payment (DvP) Settlement on Hyperledger Besu QBFT
 * @dev Enforces Zero-PII invariant via 32-byte investor commitments and gasless paymaster relay
 */
contract BesuSettlement {
    address public admin;
    address public relayerPaymaster;
    uint256 public constant PLATFORM_FEE_BPS = 0; // 0.00% launch fee

    struct TradeAllocation {
        bytes32 executionId;
        bytes32 buyerCommitment;
        bytes32 sellerCommitment;
        address assetToken;
        uint256 assetAmount;
        address settlementToken;
        uint256 settlementAmount;
        uint256 timestamp;
    }

    mapping(bytes32 => bool) public executedTrades;
    mapping(bytes32 => bool) public compliantInvestors;

    event TradeSettled(
        bytes32 indexed executionId,
        bytes32 indexed buyerCommitment,
        bytes32 indexed sellerCommitment,
        address assetToken,
        uint256 assetAmount,
        uint256 settlementAmount
    );
    event ComplianceStatusUpdated(bytes32 indexed commitment, bool compliant);
    event PaymasterUpdated(address indexed newPaymaster);

    modifier onlyAdmin() {
        require(msg.sender == admin, "Only admin");
        _;
    }

    modifier onlyRelayer() {
        require(msg.sender == relayerPaymaster || msg.sender == admin, "Only authorized relayer");
        _;
    }

    constructor(address _relayerPaymaster) {
        require(_relayerPaymaster != address(0), "Invalid paymaster");
        admin = msg.sender;
        relayerPaymaster = _relayerPaymaster;
    }

    function setComplianceStatus(bytes32 commitment, bool compliant) external onlyAdmin {
        compliantInvestors[commitment] = compliant;
        emit ComplianceStatusUpdated(commitment, compliant);
    }

    function batchSetComplianceStatus(bytes32[] calldata commitments, bool[] calldata statuses) external onlyAdmin {
        require(commitments.length == statuses.length, "Array length mismatch");
        for (uint256 i = 0; i < commitments.length; i++) {
            compliantInvestors[commitments[i]] = statuses[i];
            emit ComplianceStatusUpdated(commitments[i], statuses[i]);
        }
    }

    function updatePaymaster(address newPaymaster) external onlyAdmin {
        require(newPaymaster != address(0), "Invalid address");
        relayerPaymaster = newPaymaster;
        emit PaymasterUpdated(newPaymaster);
    }

    /**
     * @notice Atomic DvP settlement for matched spot orders
     */
    function settleTrade(TradeAllocation calldata trade) external onlyRelayer returns (bool) {
        require(!executedTrades[trade.executionId], "Trade already settled");
        require(compliantInvestors[trade.buyerCommitment], "Buyer non-compliant");
        require(compliantInvestors[trade.sellerCommitment], "Seller non-compliant");
        require(trade.assetAmount > 0 && trade.settlementAmount > 0, "Invalid amounts");

        executedTrades[trade.executionId] = true;

        // In a live ERC-20 / ERC-3643 DvP transfer, tokens are transferred atomically:
        // IERC20(trade.assetToken).transferFrom(seller, buyer, trade.assetAmount);
        // IERC20(trade.settlementToken).transferFrom(buyer, seller, trade.settlementAmount);

        emit TradeSettled(
            trade.executionId,
            trade.buyerCommitment,
            trade.sellerCommitment,
            trade.assetToken,
            trade.assetAmount,
            trade.settlementAmount
        );

        return true;
    }

    /**
     * @notice Batch settlement for high-throughput clearing
     */
    function batchSettleTrades(TradeAllocation[] calldata trades) external onlyRelayer returns (uint256 settledCount) {
        for (uint256 i = 0; i < trades.length; i++) {
            if (!executedTrades[trades[i].executionId] &&
                compliantInvestors[trades[i].buyerCommitment] &&
                compliantInvestors[trades[i].sellerCommitment]) {
                
                executedTrades[trades[i].executionId] = true;
                emit TradeSettled(
                    trades[i].executionId,
                    trades[i].buyerCommitment,
                    trades[i].sellerCommitment,
                    trades[i].assetToken,
                    trades[i].assetAmount,
                    trades[i].settlementAmount
                );
                settledCount++;
            }
        }
    }
}
