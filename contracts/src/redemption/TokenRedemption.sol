// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {EIP712Upgradeable} from "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

import {ITokenRedemption} from "../interfaces/ITokenRedemption.sol";
import {IDigitalSecurityToken} from "../interfaces/tokens/IDigitalSecurityToken.sol";
import {IRedemptionReceiver} from "../interfaces/IRedemptionReceiver.sol";

/**
 * @title TokenRedemption
 * @notice Two-phase token redemption and custodial asset release manager.
 * @dev Escrows DigitalSecurityToken units from verified investors, verifies custodial
 *      settlement proofs via EIP-712 structured signatures, and irreversibly burns tokens.
 */
contract TokenRedemption is
    Initializable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable,
    PausableUpgradeable,
    EIP712Upgradeable,
    ITokenRedemption
{
    using SafeERC20 for IERC20;

    // --- Roles ---
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    bytes32 public constant CUSTODIAN_ROLE = keccak256("CUSTODIAN_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    // --- EIP-712 TypeHash ---
    bytes32 public constant REDEMPTION_TYPEHASH =
        keccak256("RedemptionSettlement(bytes32 redemptionId,bytes32 dematDebitRef,uint256 timestamp)");

    // --- State Variables ---
    uint64 public defaultTimeoutSeconds;
    mapping(address => bool) public supportedTokens;
    mapping(bytes32 => RedemptionRequest) private _requests;
    mapping(address => uint256) public nonces;

    // Optional callback receiver
    IRedemptionReceiver public redemptionReceiver;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address adminMultisig,
        address initialRelayer,
        uint64 defaultTimeout
    ) external initializer {
        if (adminMultisig == address(0) || initialRelayer == address(0)) revert InvalidZeroAddress();
        if (defaultTimeout == 0) revert InvalidAmount();

        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        __Pausable_init();
        __EIP712_init("GrowwwRedemptionManager", "1");

        _grantRole(DEFAULT_ADMIN_ROLE, adminMultisig);
        _grantRole(UPGRADER_ROLE, adminMultisig);
        _grantRole(RELAYER_ROLE, initialRelayer);
        _grantRole(CUSTODIAN_ROLE, initialRelayer);

        defaultTimeoutSeconds = defaultTimeout;
    }

    // =========================================================================
    // Admin Configuration
    // =========================================================================

    function setSupportedToken(address token, bool supported) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (token == address(0)) revert InvalidZeroAddress();
        supportedTokens[token] = supported;
        emit TokenSupportUpdated(token, supported);
    }

    function setDefaultTimeout(uint64 newTimeout) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newTimeout == 0) revert InvalidAmount();
        defaultTimeoutSeconds = newTimeout;
    }

    function setRedemptionReceiver(address receiver) external onlyRole(DEFAULT_ADMIN_ROLE) {
        redemptionReceiver = IRedemptionReceiver(receiver);
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    // =========================================================================
    // User Operations
    // =========================================================================

    function requestRedemption(
        address token,
        uint256 amount
    ) external override nonReentrant whenNotPaused returns (bytes32 redemptionId) {
        if (amount == 0) revert InvalidAmount();
        if (!supportedTokens[token]) revert TokenNotSupported(token);

        redemptionId = keccak256(
            abi.encode(msg.sender, token, amount, block.timestamp, nonces[msg.sender]++)
        );

        uint64 expiry = uint64(block.timestamp + defaultTimeoutSeconds);

        _requests[redemptionId] = RedemptionRequest({
            id: redemptionId,
            investor: msg.sender,
            tokenAddress: token,
            amount: amount,
            requestTimestamp: uint64(block.timestamp),
            expiryTimestamp: expiry,
            status: RedemptionStatus.ESCROWED,
            dematDebitRef: bytes32(0)
        });

        // Escrow tokens into this contract
        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);

        emit RedemptionInitiated(redemptionId, msg.sender, token, amount, expiry);
    }

    // =========================================================================
    // Custodian / Relayer Settlement
    // =========================================================================

    function executeRedemptionWithProof(
        bytes32 redemptionId,
        bytes32 dematDebitRef,
        bytes calldata custodianSignature
    ) public override nonReentrant whenNotPaused {
        RedemptionRequest storage req = _requests[redemptionId];
        if (req.status == RedemptionStatus.NONE) revert RequestNotFound(redemptionId);
        if (req.status != RedemptionStatus.ESCROWED) revert InvalidStatus(req.status);
        if (block.timestamp > req.expiryTimestamp) {
            revert RequestExpired(req.expiryTimestamp, block.timestamp);
        }

        // Verify authorization
        _verifySettlementAuth(redemptionId, dematDebitRef, custodianSignature);

        req.status = RedemptionStatus.SETTLED_BURNED;
        req.dematDebitRef = dematDebitRef;

        // Atomic burning of escrowed tokens
        IDigitalSecurityToken(req.tokenAddress).burn(address(this), req.amount, redemptionId);

        emit RedemptionCompleted(redemptionId, dematDebitRef, block.timestamp);

        if (address(redemptionReceiver) != address(0)) {
            try redemptionReceiver.onRedemptionCompleted(
                redemptionId,
                req.investor,
                req.tokenAddress,
                req.amount
            ) {} catch {}
        }
    }

    function executeBatchRedemption(
        bytes32[] calldata redemptionIds,
        bytes32[] calldata dematDebitRefs,
        bytes[] calldata signatures
    ) external override nonReentrant whenNotPaused {
        uint256 len = redemptionIds.length;
        if (len != dematDebitRefs.length || len != signatures.length) {
            revert ArrayLengthMismatch(len, dematDebitRefs.length);
        }

        for (uint256 i = 0; i < len; i++) {
            bytes32 rId = redemptionIds[i];
            RedemptionRequest storage req = _requests[rId];
            if (req.status == RedemptionStatus.NONE) revert RequestNotFound(rId);
            if (req.status != RedemptionStatus.ESCROWED) revert InvalidStatus(req.status);
            if (block.timestamp > req.expiryTimestamp) {
                revert RequestExpired(req.expiryTimestamp, block.timestamp);
            }

            _verifySettlementAuth(rId, dematDebitRefs[i], signatures[i]);

            req.status = RedemptionStatus.SETTLED_BURNED;
            req.dematDebitRef = dematDebitRefs[i];

            IDigitalSecurityToken(req.tokenAddress).burn(address(this), req.amount, rId);

            emit RedemptionCompleted(rId, dematDebitRefs[i], block.timestamp);

            if (address(redemptionReceiver) != address(0)) {
                try redemptionReceiver.onRedemptionCompleted(
                    rId,
                    req.investor,
                    req.tokenAddress,
                    req.amount
                ) {} catch {}
            }
        }
    }

    // =========================================================================
    // Cancellation & SLA Recovery
    // =========================================================================

    function cancelExpiredRedemption(bytes32 redemptionId) external override nonReentrant whenNotPaused {
        RedemptionRequest storage req = _requests[redemptionId];
        if (req.status == RedemptionStatus.NONE) revert RequestNotFound(redemptionId);
        if (req.status != RedemptionStatus.ESCROWED) revert InvalidStatus(req.status);
        if (block.timestamp <= req.expiryTimestamp) {
            revert RequestNotExpired(req.expiryTimestamp, block.timestamp);
        }

        req.status = RedemptionStatus.CANCELLED;

        // Refund escrowed tokens back to investor
        IERC20(req.tokenAddress).safeTransfer(req.investor, req.amount);

        emit RedemptionCancelled(redemptionId, "SLA_EXPIRED");
    }

    function adminCancelRedemption(
        bytes32 redemptionId,
        string calldata reason
    ) external override onlyRole(DEFAULT_ADMIN_ROLE) nonReentrant {
        RedemptionRequest storage req = _requests[redemptionId];
        if (req.status == RedemptionStatus.NONE) revert RequestNotFound(redemptionId);
        if (req.status != RedemptionStatus.ESCROWED) revert InvalidStatus(req.status);

        req.status = RedemptionStatus.CANCELLED;

        // Refund escrowed tokens back to investor
        IERC20(req.tokenAddress).safeTransfer(req.investor, req.amount);

        emit RedemptionCancelled(redemptionId, reason);
    }

    // =========================================================================
    // View Functions
    // =========================================================================

    function getRequest(bytes32 redemptionId) external view override returns (RedemptionRequest memory) {
        RedemptionRequest memory req = _requests[redemptionId];
        if (req.status == RedemptionStatus.NONE) revert RequestNotFound(redemptionId);
        return req;
    }

    // =========================================================================
    // Internal Helpers
    // =========================================================================

    function _verifySettlementAuth(
        bytes32 redemptionId,
        bytes32 dematDebitRef,
        bytes calldata signature
    ) internal view {
        if (signature.length > 0) {
            bytes32 structHash = keccak256(
                abi.encode(REDEMPTION_TYPEHASH, redemptionId, dematDebitRef, block.timestamp)
            );
            bytes32 digest = _hashTypedDataV4(structHash);
            address signer = ECDSA.recover(digest, signature);
            if (!hasRole(CUSTODIAN_ROLE, signer)) revert InvalidSignature();
        } else {
            if (!hasRole(RELAYER_ROLE, msg.sender)) revert InvalidSignature();
        }
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}
}
