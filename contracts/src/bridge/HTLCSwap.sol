// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title HTLCSwap
 * @notice Cross-Chain Hash Time-Locked Contract (HTLC) for Atomic Swaps
 * @dev Enables trustless peer-to-peer and bridge swaps between BTC, EVM, and USDT rails
 */
contract HTLCSwap {
    struct SwapOrder {
        bytes32 hashLock;
        uint256 timelock;
        uint256 value;
        address payable sender;
        address payable receiver;
        bool withdrawn;
        bool refunded;
        bytes preimage;
    }

    mapping(bytes32 => SwapOrder) public swaps;

    event SwapInitiated(
        bytes32 indexed swapId,
        bytes32 indexed hashLock,
        uint256 timelock,
        uint256 value,
        address sender,
        address receiver
    );
    event SwapWithdrawn(bytes32 indexed swapId, bytes preimage);
    event SwapRefunded(bytes32 indexed swapId);

    /**
     * @notice Initiate an HTLC escrow
     */
    function initiateSwap(
        bytes32 swapId,
        bytes32 hashLock,
        uint256 timelockDuration,
        address payable receiver
    ) external payable {
        require(msg.value > 0, "No funds provided");
        require(swaps[swapId].value == 0, "Swap already exists");
        require(receiver != address(0), "Invalid receiver");
        require(timelockDuration >= 1 hours, "Timelock duration too short");

        uint256 timelock = block.timestamp + timelockDuration;

        swaps[swapId] = SwapOrder({
            hashLock: hashLock,
            timelock: timelock,
            value: msg.value,
            sender: payable(msg.sender),
            receiver: receiver,
            withdrawn: false,
            refunded: false,
            preimage: ""
        });

        emit SwapInitiated(swapId, hashLock, timelock, msg.value, msg.sender, receiver);
    }

    /**
     * @notice Withdraw funds by presenting secret preimage before timelock expires
     */
    function withdraw(bytes32 swapId, bytes calldata preimage) external {
        SwapOrder storage swap = swaps[swapId];
        require(swap.value > 0, "Swap not found");
        require(!swap.withdrawn, "Already withdrawn");
        require(!swap.refunded, "Already refunded");
        require(sha256(preimage) == swap.hashLock, "Hashlock mismatch");

        swap.withdrawn = true;
        swap.preimage = preimage;

        (bool sent, ) = swap.receiver.call{value: swap.value}("");
        require(sent, "Transfer failed");

        emit SwapWithdrawn(swapId, preimage);
    }

    /**
     * @notice Refund funds back to sender after timelock expiration
     */
    function refund(bytes32 swapId) external {
        SwapOrder storage swap = swaps[swapId];
        require(swap.value > 0, "Swap not found");
        require(!swap.withdrawn, "Already withdrawn");
        require(!swap.refunded, "Already refunded");
        require(block.timestamp >= swap.timelock, "Timelock not yet expired");

        swap.refunded = true;

        (bool sent, ) = swap.sender.call{value: swap.value}("");
        require(sent, "Refund transfer failed");

        emit SwapRefunded(swapId);
    }
}
