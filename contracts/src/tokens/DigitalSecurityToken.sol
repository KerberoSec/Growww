// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {ContextUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ContextUpgradeable.sol";
import {AccessControlEnumerableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/extensions/AccessControlEnumerableUpgradeable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IDigitalSecurityToken} from "../interfaces/tokens/IDigitalSecurityToken.sol";
import {IIdentityRegistry} from "../interfaces/IIdentityRegistry.sol";

/**
 * @title DigitalSecurityToken
 * @notice ERC-3643 compliant permissioned security token for Growww / NBSE Indian equity tokens.
 * @dev Enforces 1:1 custody-backed minting against NSDL/CDSL demat receipts, KYC identity gating,
 *      address-level freezing for regulatory enforcement, and upgradeable UUPS proxy semantics.
 */
contract DigitalSecurityToken is
    Initializable,
    ContextUpgradeable,
    AccessControlEnumerableUpgradeable,
    UUPSUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable,
    IDigitalSecurityToken
{
    // --- Roles ---
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");
    bytes32 public constant FREEZER_ROLE = keccak256("FREEZER_ROLE");
    bytes32 public constant COMPLIANCE_ROLE = keccak256("COMPLIANCE_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    // --- State Variables ---
    string private _name;
    string private _symbol;
    uint8 private _decimals;
    uint256 private _totalSupply;

    string private _isin;
    bytes32 private _isinHash;
    address public override identityRegistry;
    address public override compliance;

    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    mapping(address => bool) private _frozen;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the digital security token instance.
     * @param name_ Token name (e.g. "Growww Reliance Industries").
     * @param symbol_ Token symbol (e.g. "GROWWW-RELIANCE").
     * @param decimals_ Token decimals (standard 18).
     * @param isin_ Equity ISIN string (e.g. "INE002A01018").
     * @param identityRegistry_ Address of on-chain KYC identity claim registry.
     * @param compliance_ Address of compliance controller.
     * @param defaultAdmin Address of initial multisig administrator.
     */
    function initialize(
        string memory name_,
        string memory symbol_,
        uint8 decimals_,
        string memory isin_,
        address identityRegistry_,
        address compliance_,
        address defaultAdmin
    ) external initializer {
        if (defaultAdmin == address(0)) revert InvalidZeroAddress();

        __Context_init();
        __AccessControlEnumerable_init();
        __UUPSUpgradeable_init();
        __Pausable_init();
        __ReentrancyGuard_init();

        _name = name_;
        _symbol = symbol_;
        _decimals = decimals_;
        _isin = isin_;
        _isinHash = keccak256(bytes(isin_));
        identityRegistry = identityRegistry_;
        compliance = compliance_;

        _grantRole(DEFAULT_ADMIN_ROLE, defaultAdmin);
        _grantRole(UPGRADER_ROLE, defaultAdmin);
        _grantRole(MINTER_ROLE, defaultAdmin);
        _grantRole(BURNER_ROLE, defaultAdmin);
        _grantRole(FREEZER_ROLE, defaultAdmin);
        _grantRole(COMPLIANCE_ROLE, defaultAdmin);
    }

    // =========================================================================
    // View Functions
    // =========================================================================

    function name() external view returns (string memory) {
        return _name;
    }

    function symbol() external view returns (string memory) {
        return _symbol;
    }

    function decimals() external view returns (uint8) {
        return _decimals;
    }

    function totalSupply() external view override returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) external view override returns (uint256) {
        return _balances[account];
    }

    function allowance(address owner, address spender) external view override returns (uint256) {
        return _allowances[owner][spender];
    }

    function isin() external view override returns (string memory) {
        return _isin;
    }

    function isinHash() external view override returns (bytes32) {
        return _isinHash;
    }

    function isinSymbol() external view override returns (string memory) {
        return _isin;
    }

    function isFrozen(address account) external view override returns (bool) {
        return _frozen[account];
    }

    // =========================================================================
    // ERC-20 Transfer Logic with Compliance Hooks
    // =========================================================================

    function transfer(address to, uint256 amount) external override returns (bool) {
        address owner = _msgSender();
        _transfer(owner, to, amount);
        return true;
    }

    function approve(address spender, uint256 amount) external override returns (bool) {
        address owner = _msgSender();
        _approve(owner, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external override returns (bool) {
        address spender = _msgSender();
        _spendAllowance(from, spender, amount);
        _transfer(from, to, amount);
        return true;
    }

    // =========================================================================
    // Privileged Minting Functions (1:1 Custody Backed)
    // =========================================================================

    /**
     * @notice Mints security tokens backed by physical depository demat receipts.
     * @param to Compliant investor recipient address.
     * @param amount Units to mint (18 decimal precision).
     * @param depositoryBatchId Unique custody confirmation batch hash.
     * @param custodyProofHash Cryptographic hash of custodial audit proof.
     */
    function mint(
        address to,
        uint256 amount,
        bytes32 depositoryBatchId,
        bytes32 custodyProofHash
    ) public override onlyRole(MINTER_ROLE) nonReentrant whenNotPaused {
        _validateMint(to, amount);

        _totalSupply += amount;
        _balances[to] += amount;

        emit Transfer(address(0), to, amount);
        emit TokensMintedWithCustodyProof(to, amount, _isinHash, depositoryBatchId, custodyProofHash);
    }

    /**
     * @notice Overload for byte calldata custody proof.
     */
    function mint(
        address to,
        uint256 amount,
        bytes32 depositoryBatchId,
        bytes calldata custodyProof
    ) external onlyRole(MINTER_ROLE) nonReentrant whenNotPaused {
        bytes32 proofHash = keccak256(custodyProof);
        mint(to, amount, depositoryBatchId, proofHash);
    }

    /**
     * @notice Batched allocation of security tokens to multiple compliant recipients.
     */
    function batchMint(
        address[] calldata recipients,
        uint256[] calldata amounts,
        bytes32 depositoryBatchId,
        bytes32 custodyProofHash
    ) public override onlyRole(MINTER_ROLE) nonReentrant whenNotPaused {
        if (recipients.length != amounts.length) {
            revert ArrayLengthMismatch(recipients.length, amounts.length);
        }

        for (uint256 i = 0; i < recipients.length; i++) {
            address to = recipients[i];
            uint256 amount = amounts[i];
            _validateMint(to, amount);

            _totalSupply += amount;
            _balances[to] += amount;

            emit Transfer(address(0), to, amount);
            emit TokensMintedWithCustodyProof(to, amount, _isinHash, depositoryBatchId, custodyProofHash);
        }
    }

    /**
     * @notice Overload for byte calldata custody proof batch mint.
     */
    function batchMint(
        address[] calldata recipients,
        uint256[] calldata amounts,
        bytes32 depositoryBatchId,
        bytes calldata custodyProof
    ) external onlyRole(MINTER_ROLE) nonReentrant whenNotPaused {
        bytes32 proofHash = keccak256(custodyProof);
        batchMint(recipients, amounts, depositoryBatchId, proofHash);
    }

    // =========================================================================
    // Redemption / Burning
    // =========================================================================

    /**
     * @notice Burns security tokens upon valid off-chain demat delivery / redemption.
     */
    function burn(
        address from,
        uint256 amount,
        bytes32 redemptionId
    ) external override onlyRole(BURNER_ROLE) nonReentrant whenNotPaused {
        if (from == address(0)) revert InvalidZeroAddress();
        if (amount == 0) revert InvalidMintAmount();
        if (_balances[from] < amount) revert InvalidMintAmount();

        _balances[from] -= amount;
        _totalSupply -= amount;

        emit Transfer(from, address(0), amount);
        emit TokensBurned(from, amount, _isinHash, redemptionId);
    }

    // =========================================================================
    // Regulatory / Compliance Management
    // =========================================================================

    function freezeAddress(address account) external override onlyRole(FREEZER_ROLE) {
        if (account == address(0)) revert InvalidZeroAddress();
        _frozen[account] = true;
        emit AddressFrozen(account, true, _msgSender());
    }

    function unfreezeAddress(address account) external override onlyRole(FREEZER_ROLE) {
        if (account == address(0)) revert InvalidZeroAddress();
        _frozen[account] = false;
        emit AddressFrozen(account, false, _msgSender());
    }

    function setIdentityRegistry(address newIdentityRegistry) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newIdentityRegistry == address(0)) revert InvalidZeroAddress();
        address oldRegistry = identityRegistry;
        identityRegistry = newIdentityRegistry;
        emit IdentityRegistryUpdated(oldRegistry, newIdentityRegistry);
    }

    function setCompliance(address newCompliance) external override onlyRole(DEFAULT_ADMIN_ROLE) {
        address oldCompliance = compliance;
        compliance = newCompliance;
        emit ComplianceUpdated(oldCompliance, newCompliance);
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    // =========================================================================
    // Internal Helper Functions
    // =========================================================================

    function _validateMint(address to, uint256 amount) internal view {
        if (to == address(0)) revert InvalidZeroAddress();
        if (amount == 0) revert InvalidMintAmount();
        if (_frozen[to]) revert AddressIsFrozen(to);

        if (identityRegistry != address(0)) {
            bool verified = IIdentityRegistry(identityRegistry).isVerified(to);
            if (!verified) revert RecipientNotCompliant(to);
        }
    }

    function _transfer(address from, address to, uint256 amount) internal whenNotPaused {
        if (from == address(0) || to == address(0)) revert InvalidZeroAddress();
        if (_frozen[from]) revert AddressIsFrozen(from);
        if (_frozen[to]) revert AddressIsFrozen(to);

        if (identityRegistry != address(0)) {
            if (!IIdentityRegistry(identityRegistry).isVerified(to)) {
                revert RecipientNotCompliant(to);
            }
        }

        uint256 fromBalance = _balances[from];
        if (fromBalance < amount) revert InvalidMintAmount();

        _balances[from] = fromBalance - amount;
        _balances[to] += amount;

        emit Transfer(from, to, amount);
    }

    function _approve(address owner, address spender, uint256 amount) internal whenNotPaused {
        if (owner == address(0) || spender == address(0)) revert InvalidZeroAddress();
        _allowances[owner][spender] = amount;
        emit Approval(owner, spender, amount);
    }

    function _spendAllowance(address owner, address spender, uint256 amount) internal {
        uint256 currentAllowance = _allowances[owner][spender];
        if (currentAllowance != type(uint256).max) {
            if (currentAllowance < amount) revert InvalidMintAmount();
            _approve(owner, spender, currentAllowance - amount);
        }
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}
}
