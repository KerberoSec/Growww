// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";

/**
 * @title DigitalRupee (eINR)
 * @notice Sovereign Central Bank Digital Currency (CBDC) token for Hyperledger Besu DvP settlement.
 * @dev 1:1 backed by RBI e-Rupee in nodal escrow custody.
 *      Enforces compliance with RBI Payment and Settlement Systems Act 2007.
 */
contract DigitalRupee is ERC20, AccessControl, Pausable {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant COMPLIANCE_ROLE = keccak256("COMPLIANCE_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");

    // Frozen accounts for FIU-IND / PMLA statutory compliance
    mapping(address => bool) private _frozenAccounts;

    // Processed RBI transaction references to prevent replay attacks
    mapping(bytes32 => bool) private _processedRBITransactions;

    event DigitalRupeeMinted(
        address indexed to,
        uint256 amount,
        bytes32 indexed rbiTxRef,
        string payerVPA,
        uint256 timestamp
    );

    event DigitalRupeeRedeemed(
        address indexed from,
        uint256 amount,
        string payeeVPA,
        uint256 timestamp
    );

    event AccountFrozen(address indexed account, string reason);
    event AccountUnfrozen(address indexed account);

    error AccountIsFrozen(address account);
    error InvalidAddress();
    error InvalidAmount();
    error RBITransactionAlreadyProcessed(bytes32 rbiTxRef);
    error EmptyVPA();

    constructor(address admin) ERC20("RBI Digital Rupee", "eINR") {
        if (admin == address(0)) revert InvalidAddress();
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(MINTER_ROLE, admin);
        _grantRole(COMPLIANCE_ROLE, admin);
        _grantRole(PAUSER_ROLE, admin);
    }

    /**
     * @notice Mints eINR upon verified deposit of RBI Digital Rupee in nodal custody account
     * @param to Recipient address on Hyperledger Besu
     * @param amount Amount in wei (18 decimals, 1 eINR = 1 INR = 100 paise)
     * @param rbiTxRef Deterministic hash of RBI CBDC core banking transaction receipt
     * @param payerVPA The source e-Rupee VPA
     */
    function mint(
        address to,
        uint256 amount,
        bytes32 rbiTxRef,
        string calldata payerVPA
    ) external onlyRole(MINTER_ROLE) whenNotPaused {
        if (to == address(0)) revert InvalidAddress();
        if (amount == 0) revert InvalidAmount();
        if (_frozenAccounts[to]) revert AccountIsFrozen(to);
        if (_processedRBITransactions[rbiTxRef]) revert RBITransactionAlreadyProcessed(rbiTxRef);
        if (bytes(payerVPA).length == 0) revert EmptyVPA();

        _processedRBITransactions[rbiTxRef] = true;
        _mint(to, amount);

        emit DigitalRupeeMinted(to, amount, rbiTxRef, payerVPA, block.timestamp);
    }

    /**
     * @notice Burns eINR to trigger fiat off-ramp redemption back to retail/wholesale bank VPA
     * @param amount Amount to burn
     * @param payeeVPA Recipient RBI e-Rupee VPA (e.g. user@rbi.edr)
     */
    function redeem(
        uint256 amount,
        string calldata payeeVPA
    ) external whenNotPaused {
        if (amount == 0) revert InvalidAmount();
        if (_frozenAccounts[msg.sender]) revert AccountIsFrozen(msg.sender);
        if (bytes(payeeVPA).length == 0) revert EmptyVPA();

        _burn(msg.sender, amount);

        emit DigitalRupeeRedeemed(msg.sender, amount, payeeVPA, block.timestamp);
    }

    /**
     * @notice Checks if an account is currently frozen under PMLA / FIU-IND orders
     */
    function isFrozen(address account) external view returns (bool) {
        return _frozenAccounts[account];
    }

    /**
     * @notice Checks if an RBI transaction reference has already been minted
     */
    function isRBITxProcessed(bytes32 rbiTxRef) external view returns (bool) {
        return _processedRBITransactions[rbiTxRef];
    }

    /**
     * @notice Freezes an account under regulatory / sanctions mandate
     */
    function freezeAccount(address account, string calldata reason) external onlyRole(COMPLIANCE_ROLE) {
        if (account == address(0)) revert InvalidAddress();
        _frozenAccounts[account] = true;
        emit AccountFrozen(account, reason);
    }

    /**
     * @notice Unfreezes an account after compliance clearance
     */
    function unfreezeAccount(address account) external onlyRole(COMPLIANCE_ROLE) {
        if (account == address(0)) revert InvalidAddress();
        _frozenAccounts[account] = false;
        emit AccountUnfrozen(account);
    }

    /**
     * @notice Emergency circuit breaker pause
     */
    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }

    /**
     * @notice Unpause circuit breaker
     */
    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    /**
     * @dev Overrides ERC20 _update hook to enforce freeze checks on transfer
     */
    function _update(
        address from,
        address to,
        uint256 value
    ) internal override whenNotPaused {
        if (from != address(0) && _frozenAccounts[from]) {
            revert AccountIsFrozen(from);
        }
        if (to != address(0) && _frozenAccounts[to]) {
            revert AccountIsFrozen(to);
        }
        super._update(from, to, value);
    }
}
