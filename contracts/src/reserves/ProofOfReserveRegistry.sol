// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title ProofOfReserveRegistry
 * @dev Prompt 007 - Public Proof-of-Reserve & Custody Verification Registry for Growww / NBSE
 * Anchors multi-party signed daily attestations from SEBI-registered custodians (NSDL/CDSL)
 * and independent statutory auditors to prove 1:1 physical share backing of RWA tokens on Besu.
 */
contract ProofOfReserveRegistry is Ownable {
    using ECDSA for bytes32;

    struct Attestation {
        string isin;
        uint256 depositoryShareCount;
        uint256 onChainTokenSupply;
        bytes32 merkleRoot;
        uint256 timestamp;
        bytes custodianSignature;
        bytes auditorSignature;
        bool isSolvent;
    }

    address public custodianSigner;
    address public auditorSigner;
    address public circuitBreakerAddress;

    // ISIN => Epoch => Attestation
    mapping(string => mapping(uint256 => Attestation)) public attestations;
    // ISIN => Current Epoch
    mapping(string => uint256) public latestEpoch;
    // Global pause flag if any ISIN fails 1:1 solvency backing
    bool public globalMintPaused;

    event AttestationPublished(
        string indexed isin,
        uint256 indexed epoch,
        uint256 depositoryShareCount,
        uint256 onChainTokenSupply,
        bytes32 merkleRoot,
        bool isSolvent,
        uint256 timestamp
    );

    event SolvencyBreachDetected(
        string indexed isin,
        uint256 depositoryShareCount,
        uint256 onChainTokenSupply,
        uint256 deficit,
        uint256 timestamp
    );

    event SignersUpdated(address indexed custodian, address indexed auditor);
    event CircuitBreakerUpdated(address indexed circuitBreaker);

    error InvalidSignerAddress();
    error InsolventBackingDeficit(uint256 depository, uint256 supply);
    error InvalidCustodianSignature();
    error InvalidAuditorSignature();
    error EpochDoesNotExist();

    constructor(
        address _custodianSigner,
        address _auditorSigner,
        address _circuitBreaker
    ) Ownable(msg.sender) {
        if (_custodianSigner == address(0) || _auditorSigner == address(0)) {
            revert InvalidSignerAddress();
        }
        custodianSigner = _custodianSigner;
        auditorSigner = _auditorSigner;
        circuitBreakerAddress = _circuitBreaker;
    }

    function setSigners(address _custodian, address _auditor) external onlyOwner {
        if (_custodian == address(0) || _auditor == address(0)) {
            revert InvalidSignerAddress();
        }
        custodianSigner = _custodian;
        auditorSigner = _auditor;
        emit SignersUpdated(_custodian, _auditor);
    }

    function setCircuitBreaker(address _circuitBreaker) external onlyOwner {
        circuitBreakerAddress = _circuitBreaker;
        emit CircuitBreakerUpdated(_circuitBreaker);
    }

    /**
     * @notice Publish daily attested custody holdings against circulating token supply
     * @dev Enforces multi-party signatures from both custodian and auditor
     */
    function publishAttestation(
        string calldata isin,
        uint256 depositoryShareCount,
        uint256 onChainTokenSupply,
        bytes32 merkleRoot,
        bytes calldata custodianSig,
        bytes calldata auditorSig
    ) external returns (uint256 epochId) {
        // Construct canonical attestation message digest
        bytes32 messageHash = keccak256(
            abi.encodePacked(
                isin,
                depositoryShareCount,
                onChainTokenSupply,
                merkleRoot,
                block.chainid,
                address(this)
            )
        );
        bytes32 ethSignedDigest = MessageHashUtils.toEthSignedMessageHash(messageHash);

        // Verify signatures
        address recoveredCustodian = ethSignedDigest.recover(custodianSig);
        if (recoveredCustodian != custodianSigner) {
            revert InvalidCustodianSignature();
        }

        address recoveredAuditor = ethSignedDigest.recover(auditorSig);
        if (recoveredAuditor != auditorSigner) {
            revert InvalidAuditorSignature();
        }

        bool solvent = (depositoryShareCount >= onChainTokenSupply);
        if (!solvent) {
            globalMintPaused = true;
            emit SolvencyBreachDetected(
                isin,
                depositoryShareCount,
                onChainTokenSupply,
                onChainTokenSupply - depositoryShareCount,
                block.timestamp
            );
        }

        uint256 epoch = latestEpoch[isin] + 1;
        latestEpoch[isin] = epoch;
        epochId = epoch;

        attestations[isin][epoch] = Attestation({
            isin: isin,
            depositoryShareCount: depositoryShareCount,
            onChainTokenSupply: onChainTokenSupply,
            merkleRoot: merkleRoot,
            timestamp: block.timestamp,
            custodianSignature: custodianSig,
            auditorSignature: auditorSig,
            isSolvent: solvent
        });

        emit AttestationPublished(
            isin,
            epoch,
            depositoryShareCount,
            onChainTokenSupply,
            merkleRoot,
            solvent,
            block.timestamp
        );
    }

    /**
     * @notice Self-service investor verification verifying salted leaf inclusion
     * @dev leaf = keccak256(investorCommitment, salt, balance)
     */
    function verifyInvestorInclusion(
        string calldata isin,
        uint256 epoch,
        bytes32 investorCommitment,
        bytes32 salt,
        uint256 balance,
        bytes32[] calldata proof
    ) external view returns (bool) {
        if (epoch == 0 || epoch > latestEpoch[isin]) {
            revert EpochDoesNotExist();
        }

        bytes32 leaf = keccak256(abi.encodePacked(investorCommitment, salt, balance));
        bytes32 root = attestations[isin][epoch].merkleRoot;

        return verifyMerkleProof(proof, root, leaf);
    }

    /**
     * @dev Standard sorted pair Merkle proof verifier
     */
    function verifyMerkleProof(
        bytes32[] calldata proof,
        bytes32 root,
        bytes32 leaf
    ) public pure returns (bool) {
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
