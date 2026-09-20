// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

/**
 * @title SolvencyRegistry
 * @notice Authoritative on-chain registry for Proof-of-Reserve and Merkle liability audits (Prompt 097).
 * @dev Anchors published Merkle liability roots, total reserve balances, auditor signatures, and solvency ratios on Hyperledger Besu.
 */
contract SolvencyRegistry is AccessControl {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    bytes32 public constant AUDITOR_ROLE = keccak256("AUDITOR_ROLE");
    bytes32 public constant PUBLISHER_ROLE = keccak256("PUBLISHER_ROLE");

    uint256 public constant BPS_SCALE = 10000;

    struct SolvencySnapshot {
        uint256 snapshotId;
        bytes32 merkleRoot;
        uint256 totalLiability;
        uint256 totalReserve;
        uint256 solvencyRatioBps; // e.g. 10500 = 105.00%
        uint256 timestamp;
        address auditor;
    }

    uint256 public nextSnapshotId = 1;
    mapping(uint256 => SolvencySnapshot) public snapshots;
    bytes32 public latestMerkleRoot;
    uint256 public latestSnapshotId;

    event SolvencySnapshotPublished(
        uint256 indexed snapshotId,
        bytes32 indexed merkleRoot,
        uint256 totalLiability,
        uint256 totalReserve,
        uint256 solvencyRatioBps,
        address indexed auditor
    );

    error InvalidSnapshotData();
    error InsolventExchange();
    error InvalidAuditorSignature();

    constructor(address admin) {
        require(admin != address(0), "Invalid admin");
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(PUBLISHER_ROLE, admin);
        _grantRole(AUDITOR_ROLE, admin);
    }

    /**
     * @notice Publishes a new solvency audit snapshot with auditor ECDSA signature (v, r, s)
     */
    function publishSolvencySnapshot(
        bytes32 merkleRoot,
        uint256 totalLiability,
        uint256 totalReserve,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external onlyRole(PUBLISHER_ROLE) returns (uint256) {
        if (merkleRoot == bytes32(0) || totalLiability == 0) revert InvalidSnapshotData();
        if (totalReserve < totalLiability) revert InsolventExchange();

        bytes32 digest = keccak256(abi.encodePacked(merkleRoot, totalLiability, totalReserve, block.chainid));
        bytes32 ethHash = digest.toEthSignedMessageHash();
        address recoveredAuditor = ecrecover(ethHash, v, r, s);
        if (!hasRole(AUDITOR_ROLE, recoveredAuditor)) revert InvalidAuditorSignature();

        uint256 ratioBps = (totalReserve * BPS_SCALE) / totalLiability;
        uint256 sid = nextSnapshotId++;

        SolvencySnapshot storage snap = snapshots[sid];
        snap.snapshotId = sid;
        snap.merkleRoot = merkleRoot;
        snap.totalLiability = totalLiability;
        snap.totalReserve = totalReserve;
        snap.solvencyRatioBps = ratioBps;
        snap.timestamp = block.timestamp;
        snap.auditor = recoveredAuditor;

        latestSnapshotId = sid;
        latestMerkleRoot = merkleRoot;

        emit SolvencySnapshotPublished(sid, merkleRoot, totalLiability, totalReserve, ratioBps, recoveredAuditor);
        return sid;
    }

    /**
     * @notice Cryptographically verifies a user's leaf inclusion in a Merkle root
     */
    function verifyLeafInclusion(
        bytes32 root,
        bytes32 leaf,
        bytes32[] calldata proof
    ) external pure returns (bool) {
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

    function getLatestSnapshot() external view returns (SolvencySnapshot memory) {
        return snapshots[latestSnapshotId];
    }
}
