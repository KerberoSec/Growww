// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "../crypto/IPostQuantumVerifier.sol";

/**
 * @title DualSignedSettlementHook
 * @notice Enforces dual-signature clearance (ECDSA + ML-DSA-44) on batch settlement transactions before DvP execution.
 */
contract DualSignedSettlementHook is Ownable {

    IPostQuantumVerifier public pqcVerifier;
    mapping(bytes32 => bool) public settledBatches;

    event SettlementBatchCleared(bytes32 indexed batchId, uint256 grossNotionalPaise, uint256 timestamp);

    error BatchAlreadySettled(bytes32 batchId);
    error DualSignatureVerificationFailed(bytes32 batchId);

    constructor(address _verifier, address initialOwner) Ownable(initialOwner) {
        require(_verifier != address(0), "Invalid verifier");
        pqcVerifier = IPostQuantumVerifier(_verifier);
    }

    function setVerifier(address _verifier) external onlyOwner {
        require(_verifier != address(0), "Invalid verifier");
        pqcVerifier = IPostQuantumVerifier(_verifier);
    }

    function processBatchSettlement(IPostQuantumVerifier.DualSignatureBatch calldata batch) external returns (bool) {
        if (settledBatches[batch.batchId]) {
            revert BatchAlreadySettled(batch.batchId);
        }

        IPostQuantumVerifier.VerificationResult memory result = pqcVerifier.verifyDualSignatureBatch(batch);
        if (!result.isValid) {
            revert DualSignatureVerificationFailed(batch.batchId);
        }

        settledBatches[batch.batchId] = true;
        emit SettlementBatchCleared(batch.batchId, batch.grossNotionalPaise, block.timestamp);
        return true;
    }
}
