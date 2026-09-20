// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title IDigitalSecurityToken
 * @notice ERC-3643 compliant security token interface for Growww / NBSE asset-backed equity tokens.
 * @dev Each token instance represents a specific ISIN (e.g. INE002A01018 for Reliance Industries).
 *      Minting is restricted to authorized custodial agents who provide depository batch proofs.
 *      Zero PII on-chain: only wallet addresses, numeric balances, and cryptographic hashes are stored.
 */
interface IDigitalSecurityToken is IERC20 {
    // =========================================================================
    // Events
    // =========================================================================

    event TokensMintedWithCustodyProof(
        address indexed to,
        uint256 amount,
        bytes32 indexed isinHash,
        bytes32 indexed depositoryBatchId,
        bytes32 custodyProofHash
    );

    event TokensBurned(
        address indexed from,
        uint256 amount,
        bytes32 indexed isinHash,
        bytes32 indexed redemptionId
    );

    event AddressFrozen(address indexed target, bool isFrozen, address indexed operator);
    event IdentityRegistryUpdated(address indexed oldRegistry, address indexed newRegistry);
    event ComplianceUpdated(address indexed oldCompliance, address indexed newCompliance);

    // =========================================================================
    // Custom Errors
    // =========================================================================

    error CallerNotMinter(address caller);
    error RecipientNotCompliant(address recipient);
    error InvalidZeroAddress();
    error InvalidMintAmount();
    error AddressIsFrozen(address account);
    error ISINMismatch(bytes32 expected, bytes32 provided);
    error ArrayLengthMismatch(uint256 recipientsLength, uint256 amountsLength);

    // =========================================================================
    // View Functions
    // =========================================================================

    function isin() external view returns (string memory);
    function isinHash() external view returns (bytes32);
    function isinSymbol() external view returns (string memory);
    function identityRegistry() external view returns (address);
    function compliance() external view returns (address);
    function isFrozen(address account) external view returns (bool);

    // =========================================================================
    // Privileged Minting Functions
    // =========================================================================

    function mint(
        address to,
        uint256 amount,
        bytes32 depositoryBatchId,
        bytes32 custodyProofHash
    ) external;

    function batchMint(
        address[] calldata recipients,
        uint256[] calldata amounts,
        bytes32 depositoryBatchId,
        bytes32 custodyProofHash
    ) external;

    // =========================================================================
    // Burn (Redemption)
    // =========================================================================

    function burn(
        address from,
        uint256 amount,
        bytes32 redemptionId
    ) external;

    // =========================================================================
    // Regulatory / Compliance Management
    // =========================================================================

    function freezeAddress(address account) external;
    function unfreezeAddress(address account) external;
    function setIdentityRegistry(address newIdentityRegistry) external;
    function setCompliance(address newCompliance) external;
}
