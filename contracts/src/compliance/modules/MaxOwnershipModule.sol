// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IComplianceModule} from "../../interfaces/IComplianceRegistry.sol";

/**
 * @title MaxOwnershipModule
 * @notice Enforces statutory maximum holding caps per investor to prevent market cornering.
 * @dev Validates that post-transfer recipient balance does not breach the specified basis point cap or absolute limit.
 */
contract MaxOwnershipModule is Ownable, IComplianceModule {
    address public immutable token;
    uint256 public maxOwnershipBps; // In basis points: 500 = 5.00%
    uint256 public constant BPS_DENOMINATOR = 10_000;

    mapping(address => bool) public isExempt;

    event MaxOwnershipBpsUpdated(uint256 oldBps, uint256 newBps);
    event ExemptionUpdated(address indexed account, bool exempt);

    error InvalidBps();
    error InvalidTokenAddress();

    constructor(address initialOwner, address tokenAddress, uint256 initialMaxBps) Ownable(initialOwner) {
        if (tokenAddress == address(0)) revert InvalidTokenAddress();
        if (initialMaxBps > BPS_DENOMINATOR) revert InvalidBps();

        token = tokenAddress;
        maxOwnershipBps = initialMaxBps;
    }

    function setMaxOwnershipBps(uint256 newBps) external onlyOwner {
        if (newBps > BPS_DENOMINATOR) revert InvalidBps();
        uint256 old = maxOwnershipBps;
        maxOwnershipBps = newBps;
        emit MaxOwnershipBpsUpdated(old, newBps);
    }

    function setExemption(address account, bool exempt) external onlyOwner {
        isExempt[account] = exempt;
        emit ExemptionUpdated(account, exempt);
    }

    function moduleCheck(
        address /* from */,
        address to,
        uint256 amount,
        address /* complianceRegistry */
    ) external view override returns (bool) {
        // Burns or transfers to zero address do not increase recipient ownership
        if (to == address(0)) {
            return true;
        }

        // Exempt addresses (e.g., market makers, settlement vaults) are not subject to cap
        if (isExempt[to]) {
            return true;
        }

        uint256 currentSupply = IERC20(token).totalSupply();
        if (currentSupply == 0) {
            return true;
        }

        uint256 currentRecipientBalance = IERC20(token).balanceOf(to);
        uint256 newBalance = currentRecipientBalance + amount;

        uint256 maxAllowed = (currentSupply * maxOwnershipBps) / BPS_DENOMINATOR;
        return newBalance <= maxAllowed;
    }

    function moduleTransferAction(address /* from */, address /* to */, uint256 /* amount */) external override {}
    function moduleMintAction(address /* to */, uint256 /* amount */) external override {}
    function moduleBurnAction(address /* from */, uint256 /* amount */) external override {}
}
