// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

import {ICustodialProofRegistry} from "../interfaces/custody/ICustodialProofRegistry.sol";

/**
 * @title CustodialProofRegistry
 * @notice Cryptographic Proof of Solvency and Custodial Proof Verification Registry.
 * @dev Enforces multi-custodian quorum attestations, Merkle proof of reserves and liabilities,
 *      and transparent cryptographic solvency proofs for institution-grade auditing.
 */
contract CustodialProofRegistry is Ownable, ReentrancyGuard, ICustodialProofRegistry {
    using ECDSA for bytes32;

    // --- State Variables ---
    address public override auditor;
    uint256 public override custodianThreshold;
    uint256 public override currentEpoch;
    uint256 public requiredReserveRatioBps; // e.g. 10000 = 100%

    address[] private _custodiansList;
    mapping(address => bool) public override isCustodian;
    mapping(uint256 => SolvencyEpochRecord) public epochs;

    constructor(
        address _auditor,
        address[] memory _initialCustodians,
        uint256 _custodianThreshold,
        uint256 _requiredReserveRatioBps,
        address initialOwner
    ) Ownable(initialOwner) {
        if (_auditor == address(0)) revert InvalidZeroAddress();
        if (_initialCustodians.length == 0 || _custodianThreshold == 0 || _custodianThreshold > _initialCustodians.length) {
            revert InvalidThreshold(_custodianThreshold, _initialCustodians.length);
        }

        auditor = _auditor;
        custodianThreshold = _custodianThreshold;
        requiredReserveRatioBps = _requiredReserveRatioBps == 0 ? 10000 : _requiredReserveRatioBps;

        for (uint256 i = 0; i < _initialCustodians.length; i++) {
            address c = _initialCustodians[i];
            if (c == address(0)) revert InvalidZeroAddress();
            if (isCustodian[c]) revert CustodianAlreadyAdded(c);
            isCustodian[c] = true;
            _custodiansList.push(c);
            emit CustodianAdded(c);
        }
    }

    // =========================================================================
    // View Functions
    // =========================================================================

    function getEpoch(uint256 epochId) external view override returns (SolvencyEpochRecord memory) {
        if (epochId == 0 || epochId > currentEpoch) revert EpochDoesNotExist(epochId);
        return epochs[epochId];
    }

    function getCustodians() external view returns (address[] memory) {
        return _custodiansList;
    }

    function isSolvent(uint256 epochId) public view override returns (bool solvent, uint256 coverageRatioBps) {
        if (epochId == 0 || epochId > currentEpoch) revert EpochDoesNotExist(epochId);
        SolvencyEpochRecord storage rec = epochs[epochId];

        if (rec.totalLiabilitiesE8 == 0) {
            return (true, 10000);
        }

        coverageRatioBps = (rec.totalReservesE8 * 10000) / rec.totalLiabilitiesE8;
        solvent = coverageRatioBps >= requiredReserveRatioBps && !rec.isChallenged;
    }

    function verifyAssetReserveProof(
        uint256 epochId,
        AssetReserveLeaf calldata leaf,
        bytes32[] calldata proof
    ) external view override returns (bool) {
        if (epochId == 0 || epochId > currentEpoch) revert EpochDoesNotExist(epochId);
        if (leaf.custodian == address(0) || leaf.balanceE8 == 0) revert InvalidLeafParameters();

        bytes32 leafHash = hashAssetLeaf(leaf);
        bytes32 root = epochs[epochId].assetsMerkleRoot;
        return verifyMerkleProof(proof, root, leafHash);
    }

    function verifyLiabilityProof(
        uint256 epochId,
        bytes32 leafHash,
        bytes32[] calldata proof
    ) external view override returns (bool) {
        if (epochId == 0 || epochId > currentEpoch) revert EpochDoesNotExist(epochId);
        bytes32 root = epochs[epochId].liabilitiesMerkleRoot;
        return verifyMerkleProof(proof, root, leafHash);
    }

    function hashAssetLeaf(AssetReserveLeaf calldata leaf) public pure returns (bytes32) {
        return keccak256(
            abi.encode(
                leaf.custodian,
                leaf.assetId,
                leaf.balanceE8,
                leaf.depositoryRef,
                leaf.timestamp
            )
        );
    }

    function getEpochDigest(
        bytes32 liabilitiesRoot,
        bytes32 assetsRoot,
        uint256 totalLiabilitiesE8,
        uint256 totalReservesE8,
        uint256 epochId,
        uint256 expiry
    ) public view returns (bytes32) {
        bytes32 structHash = keccak256(
            abi.encode(
                liabilitiesRoot,
                assetsRoot,
                totalLiabilitiesE8,
                totalReservesE8,
                block.chainid,
                epochId,
                expiry
            )
        );
        return keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", structHash));
    }

    // =========================================================================
    // Epoch Publication & Custodial Verification
    // =========================================================================

    function publishAttestedEpoch(
        bytes32 liabilitiesRoot,
        bytes32 assetsRoot,
        uint256 totalLiabilitiesE8,
        uint256 totalReservesE8,
        string calldata reportUri,
        uint256 expiry,
        bytes[] calldata custodianSignatures
    ) external override nonReentrant returns (uint256 epochId) {
        if (msg.sender != auditor && msg.sender != owner()) revert UnauthorizedCaller();
        if (block.timestamp > expiry) revert SignatureExpired(expiry, block.timestamp);

        uint256 requiredMinReserves = (totalLiabilitiesE8 * requiredReserveRatioBps) / 10000;
        if (totalReservesE8 < requiredMinReserves) {
            revert InsufficientReserves(totalReservesE8, requiredMinReserves);
        }

        uint256 nextEpoch = currentEpoch + 1;
        bytes32 digest = getEpochDigest(
            liabilitiesRoot,
            assetsRoot,
            totalLiabilitiesE8,
            totalReservesE8,
            nextEpoch,
            expiry
        );

        // Verify multi-custodian quorum
        uint256 validSignaturesCount = 0;
        address[] memory verifiedSigners = new address[](custodianSignatures.length);

        for (uint256 i = 0; i < custodianSignatures.length; i++) {
            address signer = digest.recover(custodianSignatures[i]);
            if (!isCustodian[signer]) revert InvalidSignature(signer);

            // Ensure no duplicate signers
            for (uint256 j = 0; j < validSignaturesCount; j++) {
                if (verifiedSigners[j] == signer) revert DuplicateSignature(signer);
            }

            verifiedSigners[validSignaturesCount] = signer;
            validSignaturesCount++;
        }

        if (validSignaturesCount < custodianThreshold) {
            revert InsufficientCustodianSignatures(validSignaturesCount, custodianThreshold);
        }

        currentEpoch = nextEpoch;
        epochId = currentEpoch;

        epochs[epochId] = SolvencyEpochRecord({
            liabilitiesMerkleRoot: liabilitiesRoot,
            assetsMerkleRoot: assetsRoot,
            totalLiabilitiesE8: totalLiabilitiesE8,
            totalReservesE8: totalReservesE8,
            timestamp: block.timestamp,
            custodianAttestationsCount: validSignaturesCount,
            auditReportUri: reportUri,
            isChallenged: false
        });

        emit EpochAttested(
            epochId,
            liabilitiesRoot,
            assetsRoot,
            totalLiabilitiesE8,
            totalReservesE8,
            block.timestamp
        );
    }

    function challengeEpoch(uint256 epochId, string calldata reason) external override {
        if (epochId == 0 || epochId > currentEpoch) revert EpochDoesNotExist(epochId);
        SolvencyEpochRecord storage rec = epochs[epochId];
        if (rec.isChallenged) revert EpochAlreadyChallenged(epochId);

        rec.isChallenged = true;
        emit SolvencyChallenged(epochId, msg.sender, reason);
    }

    function resolveChallenge(uint256 epochId, bool upheld) external override {
        if (msg.sender != auditor && msg.sender != owner()) revert UnauthorizedCaller();
        if (epochId == 0 || epochId > currentEpoch) revert EpochDoesNotExist(epochId);

        epochs[epochId].isChallenged = upheld;
        emit SolvencyChallengeResolved(epochId, upheld);
    }

    // =========================================================================
    // Admin Controls
    // =========================================================================

    function addCustodian(address newCustodian) external onlyOwner {
        if (newCustodian == address(0)) revert InvalidZeroAddress();
        if (isCustodian[newCustodian]) revert CustodianAlreadyAdded(newCustodian);

        isCustodian[newCustodian] = true;
        _custodiansList.push(newCustodian);
        emit CustodianAdded(newCustodian);
    }

    function removeCustodian(address custodian) external onlyOwner {
        if (!isCustodian[custodian]) revert CustodianNotFound(custodian);
        if (_custodiansList.length - 1 < custodianThreshold) {
            revert InvalidThreshold(custodianThreshold, _custodiansList.length - 1);
        }

        isCustodian[custodian] = false;
        for (uint256 i = 0; i < _custodiansList.length; i++) {
            if (_custodiansList[i] == custodian) {
                _custodiansList[i] = _custodiansList[_custodiansList.length - 1];
                _custodiansList.pop();
                break;
            }
        }
        emit CustodianRemoved(custodian);
    }

    function setCustodianThreshold(uint256 newThreshold) external onlyOwner {
        if (newThreshold == 0 || newThreshold > _custodiansList.length) {
            revert InvalidThreshold(newThreshold, _custodiansList.length);
        }
        emit CustodianThresholdUpdated(custodianThreshold, newThreshold);
        custodianThreshold = newThreshold;
    }

    function setAuditor(address newAuditor) external onlyOwner {
        if (newAuditor == address(0)) revert InvalidZeroAddress();
        emit AuditorUpdated(auditor, newAuditor);
        auditor = newAuditor;
    }

    function setRequiredReserveRatio(uint256 newRatioBps) external onlyOwner {
        require(newRatioBps >= 10000, "Ratio must be at least 100%");
        requiredReserveRatioBps = newRatioBps;
    }

    // =========================================================================
    // Internal Merkle Helpers
    // =========================================================================

    function verifyMerkleProof(
        bytes32[] calldata proof,
        bytes32 root,
        bytes32 leaf
    ) internal pure returns (bool) {
        bytes32 computedHash = leaf;
        for (uint256 i = 0; i < proof.length; i++) {
            bytes32 proofElement = proof[i];
            if (computedHash <= proofElement) {
                computedHash = keccak256(abi.encodePacked(computedHash, proofElement));
            } else {
                computedHash = keccak256(abi.encodePacked(proofElement, computedHash));
            }
        }
        return computedHash == root;
    }
}
