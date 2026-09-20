// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICommitRevealSettlement {
    enum OrderSide {
        BUY,
        SELL
    }

    enum CommitmentStatus {
        UNINITIALIZED,
        COMMITTED,
        REVEALED,
        SETTLED,
        SLASHED,
        CANCELLED
    }

    struct OrderParams {
        address baseToken;
        address quoteToken;
        OrderSide side;
        uint256 amount;
        uint256 limitPrice;
        uint256 nonce;
        uint256 deadline;
    }

    struct OrderCommitment {
        bytes32 commitmentHash;
        address trader;
        uint256 commitBlock;
        uint256 commitTimestamp;
        uint256 commitBond;
        CommitmentStatus status;
    }

    struct RevealedOrder {
        bytes32 commitmentHash;
        address trader;
        OrderParams order;
        uint256 revealBlock;
        uint256 revealTimestamp;
        bool isSettled;
    }

    // Events
    event OrderCommitted(
        bytes32 indexed commitmentHash,
        address indexed trader,
        uint256 commitBond,
        uint256 commitBlock,
        uint256 commitTimestamp
    );
    event OrderRevealed(
        bytes32 indexed commitmentHash,
        address indexed trader,
        address baseToken,
        address quoteToken,
        OrderSide side,
        uint256 amount,
        uint256 limitPrice
    );
    event OrderSettled(
        bytes32 indexed commitmentHash,
        address indexed trader,
        address counterparty,
        uint256 executionPrice,
        uint256 amountFilled
    );
    event CommitmentSlashed(
        bytes32 indexed commitmentHash,
        address indexed trader,
        address indexed slasher,
        uint256 slashedAmount
    );
    event BatchSettlementExecuted(
        bytes32[] buyCommitmentHashes,
        bytes32[] sellCommitmentHashes,
        uint256 uniformClearingPrice,
        uint256 totalVolume
    );

    // Custom Errors
    error CommitmentAlreadyExists(bytes32 commitmentHash);
    error CommitmentDoesNotExist(bytes32 commitmentHash);
    error InvalidCommitmentHash();
    error InsufficientCommitBond(uint256 provided, uint256 required);
    error RevealTooEarly(uint256 currentBlock, uint256 minRevealBlock);
    error RevealTooLate(uint256 currentBlock, uint256 maxRevealBlock);
    error InvalidOrderHash();
    error OrderExpired(uint256 deadline, uint256 currentTimestamp);
    error CommitmentNotSlashable(bytes32 commitmentHash);
    error UnauthorizedTrader(address caller, address expected);
    error InvalidZeroAddress();
    error NonceAlreadyUsed(address trader, uint256 nonce);
    error OrderNotRevealed(bytes32 commitmentHash);

    function computeCommitmentHash(
        address trader,
        OrderParams calldata order,
        bytes32 salt
    ) external pure returns (bytes32);

    function commitOrder(bytes32 commitmentHash, uint256 bondAmount) external;
    function revealOrder(OrderParams calldata order, bytes32 salt) external returns (bytes32 commitmentHash);
    function slashExpiredCommitment(bytes32 commitmentHash) external;
    function batchSettleOrders(
        bytes32[] calldata buyCommitmentHashes,
        bytes32[] calldata sellCommitmentHashes,
        uint256 clearingPrice
    ) external;

    function getCommitment(bytes32 commitmentHash) external view returns (OrderCommitment memory);
    function getRevealedOrder(bytes32 commitmentHash) external view returns (RevealedOrder memory);
    function isNonceUsed(address trader, uint256 nonce) external view returns (bool);
}
