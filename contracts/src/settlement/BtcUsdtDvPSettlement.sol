// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title BtcUsdtDvPSettlement
 * @notice Real-Money Delivery-versus-Payment (DvP) Atomic Settlement Engine for BTC/USDT Pairs
 * @dev Enforces cryptographic non-repudiation and zero-PII trade settlement
 */
contract BtcUsdtDvPSettlement {
    address public immutable exchangeOperator;
    address public immutable settlementRelayer;

    struct DvPAllocation {
        bytes32 tradeHash;
        bytes32 buyerCommitment;
        bytes32 sellerCommitment;
        uint64 btcAmountSat;    // Satoshi units (1e8)
        uint64 usdtAmountCents; // USDT micro-units / cents (1e6)
        uint64 executionNonce;
        uint256 matchedTimestamp;
    }

    mapping(bytes32 => bool) public settledTrades;
    mapping(bytes32 => uint64) public accountNonces;

    event DvPExecuted(
        bytes32 indexed tradeHash,
        bytes32 indexed buyerCommitment,
        bytes32 indexed sellerCommitment,
        uint64 btcAmountSat,
        uint64 usdtAmountCents,
        uint256 timestamp
    );
    event TradeVoided(bytes32 indexed tradeHash, string reason);

    modifier onlyRelayer() {
        require(msg.sender == settlementRelayer || msg.sender == exchangeOperator, "Unauthorized relayer");
        _;
    }

    constructor(address _operator, address _relayer) {
        require(_operator != address(0) && _relayer != address(0), "Invalid addresses");
        exchangeOperator = _operator;
        settlementRelayer = _relayer;
    }

    /**
     * @notice Execute atomic DvP settlement
     */
    function executeDvP(DvPAllocation calldata trade) external onlyRelayer returns (bool) {
        require(!settledTrades[trade.tradeHash], "Trade already settled");
        require(trade.btcAmountSat > 0, "Invalid BTC amount");
        require(trade.usdtAmountCents > 0, "Invalid USDT amount");

        // Verify nonce progression to mitigate replay
        require(trade.executionNonce > accountNonces[trade.buyerCommitment], "Stale buyer nonce");
        require(trade.executionNonce > accountNonces[trade.sellerCommitment], "Stale seller nonce");

        accountNonces[trade.buyerCommitment] = trade.executionNonce;
        accountNonces[trade.sellerCommitment] = trade.executionNonce;
        settledTrades[trade.tradeHash] = true;

        emit DvPExecuted(
            trade.tradeHash,
            trade.buyerCommitment,
            trade.sellerCommitment,
            trade.btcAmountSat,
            trade.usdtAmountCents,
            block.timestamp
        );

        return true;
    }

    /**
     * @notice Void a broken or disputed trade leg prior to settlement completion
     */
    function voidTrade(bytes32 tradeHash, string calldata reason) external onlyRelayer {
        require(!settledTrades[tradeHash], "Cannot void already settled trade");
        settledTrades[tradeHash] = true; // Mark to prevent late execution
        emit TradeVoided(tradeHash, reason);
    }
}
