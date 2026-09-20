// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IComplianceModule} from "../../interfaces/IComplianceRegistry.sol";

/**
 * @title LockupModule
 * @notice Enforces statutory lock-in periods on token balances for promoters and pre-IPO investors.
 * @dev Verifies that transfers do not spend balances subject to unexpired lockups.
 */
contract LockupModule is Ownable, IComplianceModule {
    struct LockEntry {
        uint256 amount;
        uint64 unlockTimestamp;
    }

    address public immutable token;
    mapping(address => LockEntry[]) private _userLocks;

    event TokensLocked(address indexed account, uint256 amount, uint64 unlockTimestamp);
    event LockReleased(address indexed account, uint256 index);

    error InvalidTokenAddress();
    error InvalidLockParams();

    constructor(address initialOwner, address tokenAddress) Ownable(initialOwner) {
        if (tokenAddress == address(0)) revert InvalidTokenAddress();
        token = tokenAddress;
    }

    function lockTokens(address account, uint256 amount, uint64 unlockTimestamp) external onlyOwner {
        if (account == address(0) || amount == 0 || unlockTimestamp <= block.timestamp) {
            revert InvalidLockParams();
        }

        _userLocks[account].push(LockEntry({
            amount: amount,
            unlockTimestamp: unlockTimestamp
        }));

        emit TokensLocked(account, amount, unlockTimestamp);
    }

    function getLockedBalance(address account) public view returns (uint256 totalLocked) {
        LockEntry[] storage locks = _userLocks[account];
        uint256 length = locks.length;
        for (uint256 i = 0; i < length; ++i) {
            if (locks[i].unlockTimestamp > block.timestamp) {
                totalLocked += locks[i].amount;
            }
        }
    }

    function getLockCount(address account) external view returns (uint256) {
        return _userLocks[account].length;
    }

    function getLockEntry(address account, uint256 index) external view returns (uint256 amount, uint64 unlockTimestamp) {
        LockEntry storage entry = _userLocks[account][index];
        return (entry.amount, entry.unlockTimestamp);
    }

    function moduleCheck(
        address from,
        address /* to */,
        uint256 amount,
        address /* complianceRegistry */
    ) external view override returns (bool) {
        // Mints from address(0) are not subject to sender lockup
        if (from == address(0)) {
            return true;
        }

        uint256 totalBalance = IERC20(token).balanceOf(from);
        uint256 locked = getLockedBalance(from);

        if (totalBalance < locked) {
            return false;
        }

        uint256 unlockedBalance = totalBalance - locked;
        return unlockedBalance >= amount;
    }

    function moduleTransferAction(address /* from */, address /* to */, uint256 /* amount */) external override {}
    function moduleMintAction(address /* to */, uint256 /* amount */) external override {}
    function moduleBurnAction(address /* from */, uint256 /* amount */) external override {}
}
