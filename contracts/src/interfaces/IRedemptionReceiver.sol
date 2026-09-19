// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IRedemptionReceiver
 * @notice Optional callback interface for downstream services notified upon redemption completion.
 */
interface IRedemptionReceiver {
    function onRedemptionCompleted(
        bytes32 redemptionId,
        address investor,
        address token,
        uint256 amount
    ) external;
}
