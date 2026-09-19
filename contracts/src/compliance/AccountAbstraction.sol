// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title AccountAbstraction
 * @notice ERC-4337 Smart Account with Guardian-Based Social Recovery for Growww / NBSE
 * @dev Compliant with EIP-4337 standard and Zero-PII regulatory requirements
 */
contract AccountAbstraction {
    address public owner;
    address public immutable entryPoint;
    uint256 public nonce;

    // Social Recovery
    address[] public guardians;
    mapping(address => bool) public isGuardian;
    uint256 public guardianQuorum;

    struct RecoveryRequest {
        address proposedOwner;
        uint256 approvalCount;
        uint256 executeAfter;
        bool executed;
    }

    RecoveryRequest public currentRecovery;
    mapping(address => bool) public guardianVoted;
    uint256 public constant RECOVERY_TIMELOCK = 48 hours;

    // Events
    event Executed(address indexed target, uint256 value, bytes data);
    event RecoveryInitiated(address indexed proposedOwner, uint256 executeAfter);
    event RecoveryApproved(address indexed guardian, address indexed proposedOwner);
    event RecoveryExecuted(address indexed oldOwner, address indexed newOwner);
    event GuardianUpdated(address indexed guardian, bool active);

    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner allowed");
        _;
    }

    modifier onlyEntryPointOrOwner() {
        require(msg.sender == entryPoint || msg.sender == owner, "Only EntryPoint or Owner");
        _;
    }

    constructor(address _owner, address _entryPoint, address[] memory _guardians, uint256 _quorum) {
        require(_owner != address(0), "Invalid owner");
        require(_quorum > 0 && _quorum <= _guardians.length, "Invalid quorum");
        
        owner = _owner;
        entryPoint = _entryPoint;
        guardianQuorum = _quorum;

        for (uint256 i = 0; i < _guardians.length; i++) {
            address g = _guardians[i];
            require(g != address(0) && !isGuardian[g], "Invalid guardian");
            isGuardian[g] = true;
            guardians.push(g);
            emit GuardianUpdated(g, true);
        }
    }

    /**
     * @notice Direct execution or execution via EntryPoint
     */
    function execute(address dest, uint256 value, bytes calldata func) external onlyEntryPointOrOwner returns (bytes memory) {
        nonce++;
        (bool success, bytes memory result) = dest.call{value: value}(func);
        require(success, "Execution failed");
        emit Executed(dest, value, func);
        return result;
    }

    /**
     * @notice ERC-4337 UserOperation validation
     */
    function validateUserOp(
        bytes32 userOpHash,
        bytes calldata signature,
        uint256 missingAccountFunds
    ) external returns (uint256 validationData) {
        require(msg.sender == entryPoint, "Only EntryPoint");
        
        // ECDSA signature verification against owner
        bytes32 ethSignedMessageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", userOpHash)
        );
        (bytes32 r, bytes32 s, uint8 v) = splitSignature(signature);
        address recovered = ecrecover(ethSignedMessageHash, v, r, s);
        
        if (recovered != owner) {
            return 1; // SIG_VALIDATION_FAILED
        }

        // Pay missing funds to EntryPoint
        if (missingAccountFunds > 0) {
            (bool success, ) = payable(entryPoint).call{value: missingAccountFunds}("");
            require(success, "Fee reimbursement failed");
        }

        return 0; // Validation success
    }

    /**
     * @notice Initiate Social Recovery by a guardian
     */
    function initiateRecovery(address newOwner) external {
        require(isGuardian[msg.sender], "Only guardian");
        require(newOwner != address(0) && newOwner != owner, "Invalid new owner");

        if (currentRecovery.proposedOwner != newOwner || currentRecovery.executed) {
            currentRecovery = RecoveryRequest({
                proposedOwner: newOwner,
                approvalCount: 1,
                executeAfter: block.timestamp + RECOVERY_TIMELOCK,
                executed: false
            });
            // Reset votes
            for (uint256 i = 0; i < guardians.length; i++) {
                guardianVoted[guardians[i]] = false;
            }
            guardianVoted[msg.sender] = true;
            emit RecoveryInitiated(newOwner, currentRecovery.executeAfter);
        } else {
            require(!guardianVoted[msg.sender], "Already voted");
            guardianVoted[msg.sender] = true;
            currentRecovery.approvalCount++;
            emit RecoveryApproved(msg.sender, newOwner);
        }
    }

    /**
     * @notice Execute recovery after timelock and sufficient quorum
     */
    function executeRecovery() external {
        require(!currentRecovery.executed, "Recovery already executed");
        require(currentRecovery.approvalCount >= guardianQuorum, "Quorum not met");
        require(block.timestamp >= currentRecovery.executeAfter, "Timelock not expired");

        address oldOwner = owner;
        owner = currentRecovery.proposedOwner;
        currentRecovery.executed = true;

        emit RecoveryExecuted(oldOwner, owner);
    }

    /**
     * @notice Cancel recovery if owner regains access
     */
    function cancelRecovery() external onlyOwner {
        currentRecovery.executed = true;
    }

    function splitSignature(bytes memory sig) internal pure returns (bytes32 r, bytes32 s, uint8 v) {
        require(sig.length == 65, "Invalid signature length");
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }
    }

    receive() external payable {}
}
