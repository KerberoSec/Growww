// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title MerkleRegistry
 * @notice Cryptographic Proof of Solvency and Proof of Reserves (PoR/PoL) Registry
 * @dev Anchors periodic Merkle tree root hashes for investor liability verification
 */
contract MerkleRegistry {
    address public auditor;
    address public exchangeAdmin;

    struct SolvencyEpoch {
        bytes32 liabilitiesMerkleRoot;
        bytes32 assetsRootHash;
        uint256 totalLiabilitiesE8;
        uint256 totalReservesE8;
        uint256 timestamp;
        string auditReportUri;
    }

    uint256 public currentEpoch;
    mapping(uint256 => SolvencyEpoch) public epochs;

    event EpochPublished(
        uint256 indexed epoch,
        bytes32 liabilitiesMerkleRoot,
        bytes32 assetsRootHash,
        uint256 totalLiabilitiesE8,
        uint256 totalReservesE8,
        uint256 timestamp
    );
    event AuditorUpdated(address indexed newAuditor);

    modifier onlyAuditorOrAdmin() {
        require(msg.sender == auditor || msg.sender == exchangeAdmin, "Unauthorized");
        _;
    }

    constructor(address _auditor) {
        require(_auditor != address(0), "Invalid auditor");
        exchangeAdmin = msg.sender;
        auditor = _auditor;
    }

    function setAuditor(address _newAuditor) external {
        require(msg.sender == exchangeAdmin, "Only admin");
        require(_newAuditor != address(0), "Invalid address");
        auditor = _newAuditor;
        emit AuditorUpdated(_newAuditor);
    }

    /**
     * @notice Publish a new attested solvency epoch
     */
    function publishEpoch(
        bytes32 liabilitiesRoot,
        bytes32 assetsRoot,
        uint256 totalLiabilities,
        uint256 totalReserves,
        string calldata reportUri
    ) external onlyAuditorOrAdmin returns (uint256 epochId) {
        require(totalReserves >= totalLiabilities, "Insolvency: Reserves less than liabilities");
        
        currentEpoch++;
        epochId = currentEpoch;
        
        epochs[epochId] = SolvencyEpoch({
            liabilitiesMerkleRoot: liabilitiesRoot,
            assetsRootHash: assetsRoot,
            totalLiabilitiesE8: totalLiabilities,
            totalReservesE8: totalReserves,
            timestamp: block.timestamp,
            auditReportUri: reportUri
        });

        emit EpochPublished(
            epochId,
            liabilitiesRoot,
            assetsRoot,
            totalLiabilities,
            totalReserves,
            block.timestamp
        );
    }

    /**
     * @notice Verify individual investor account inclusion in solvency tree
     */
    function verifyAccountInclusion(
        uint256 epoch,
        bytes32 leaf,
        bytes32[] calldata proof
    ) external view returns (bool) {
        require(epoch > 0 && epoch <= currentEpoch, "Epoch does not exist");
        bytes32 root = epochs[epoch].liabilitiesMerkleRoot;
        return verifyProof(proof, root, leaf);
    }

    function verifyProof(
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
