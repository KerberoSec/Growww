// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/**
 * @title ColdStorageMultiSig
 * @notice 3-of-5 Institutional Cold Storage Vault with EIP-712 Multi-Sig & 48h Timelock
 * @dev Compliant with Prompts 150, 198 and Custody Specification
 */
contract ColdStorageMultiSig {
    // EIP-712 Domain Separator & Typehashes
    bytes32 public immutable DOMAIN_SEPARATOR;
    bytes32 public constant WITHDRAWAL_TYPEHASH = keccak256(
        "WithdrawalRequest(bytes32 withdrawalId,address recipient,address token,uint256 amount,uint256 fee,uint256 nonce,uint256 validUntil)"
    );

    uint256 public constant TIMELOCK_DELAY = 48 hours;
    uint256 public constant ZERO_FEE = 0; // Zero withdrawal fee invariant

    uint256 public threshold;
    address[] public signers;
    mapping(address => bool) public isSigner;

    struct WithdrawalRequest {
        bytes32 withdrawalId;
        address recipient;
        address token; // address(0) for native ETH / eINR gas coin
        uint256 amount;
        uint256 fee; // Must be 0 at launch
        uint256 nonce;
        uint256 validUntil;
    }

    mapping(bytes32 => bool) public executedWithdrawals;
    mapping(uint256 => bool) public executedNonces;

    // Timelocked Signer Set Updates
    struct PendingSignerUpdate {
        address[] newSigners;
        uint256 newThreshold;
        uint256 effectiveTimestamp;
        bool executed;
    }
    PendingSignerUpdate public pendingUpdate;

    event WithdrawalExecuted(
        bytes32 indexed withdrawalId,
        address indexed recipient,
        address indexed token,
        uint256 amount,
        uint256 nonce
    );
    event SignerUpdateScheduled(address[] newSigners, uint256 newThreshold, uint256 effectiveTimestamp);
    event SignerUpdateExecuted(address[] newSigners, uint256 newThreshold);
    event SignerUpdateCancelled();
    event DepositReceived(address indexed sender, uint256 amount);

    modifier onlySigner() {
        require(isSigner[msg.sender], "Only authorized signer");
        _;
    }

    constructor(address[] memory _signers, uint256 _threshold) {
        require(_signers.length >= 3, "Minimum 3 signers required");
        require(_threshold >= 2 && _threshold <= _signers.length, "Invalid quorum threshold");

        for (uint256 i = 0; i < _signers.length; i++) {
            address s = _signers[i];
            require(s != address(0), "Invalid zero address signer");
            require(!isSigner[s], "Duplicate signer in init");
            isSigner[s] = true;
            signers.push(s);
        }
        threshold = _threshold;

        DOMAIN_SEPARATOR = keccak256(
            abi.encode(
                keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
                keccak256(bytes("GrowwwColdStorage")),
                keccak256(bytes("1")),
                block.chainid,
                address(this)
            )
        );
    }

    /**
     * @notice Get all active signers
     */
    function getSigners() external view returns (address[] memory) {
        return signers;
    }

    /**
     * @notice Compute EIP-712 structured hash for a withdrawal request
     */
    function hashWithdrawalRequest(WithdrawalRequest calldata req) public view returns (bytes32) {
        bytes32 structHash = keccak256(
            abi.encode(
                WITHDRAWAL_TYPEHASH,
                req.withdrawalId,
                req.recipient,
                req.token,
                req.amount,
                req.fee,
                req.nonce,
                req.validUntil
            )
        );

        return keccak256(abi.encodePacked("\x19\x01", DOMAIN_SEPARATOR, structHash));
    }

    /**
     * @notice Execute cold vault withdrawal with M-of-N threshold signatures
     */
    function executeWithdrawal(
        WithdrawalRequest calldata req,
        bytes[] calldata signatures
    ) external returns (bool) {
        require(!executedWithdrawals[req.withdrawalId], "Withdrawal already executed");
        require(!executedNonces[req.nonce], "Nonce already used");
        require(block.timestamp <= req.validUntil, "Withdrawal request expired");
        require(req.recipient != address(0), "Invalid recipient");
        require(req.amount > 0, "Invalid withdrawal amount");
        require(req.fee == ZERO_FEE, "Fee must be zero at launch");
        require(signatures.length >= threshold, "Insufficient signatures for quorum");

        bytes32 digest = hashWithdrawalRequest(req);

        // Verify distinct signatures from authorized signers
        address lastSigner = address(0);
        for (uint256 i = 0; i < signatures.length; i++) {
            address recovered = recoverSigner(digest, signatures[i]);
            require(isSigner[recovered], "Unauthorized signer signature");
            require(recovered > lastSigner, "Signatures must be strictly ordered without duplicates");
            lastSigner = recovered;
        }

        // Checks-Effects: update state before external transfers
        executedWithdrawals[req.withdrawalId] = true;
        executedNonces[req.nonce] = true;

        if (req.token == address(0)) {
            // Native currency transfer
            require(address(this).balance >= req.amount, "Insufficient vault balance");
            (bool sent, ) = req.recipient.call{value: req.amount}("");
            require(sent, "Native transfer failed");
        } else {
            // ERC-20 transfer
            (bool success, bytes memory data) = req.token.call(
                abi.encodeWithSignature("transfer(address,uint256)", req.recipient, req.amount)
            );
            require(success && (data.length == 0 || abi.decode(data, (bool))), "ERC20 transfer failed");
        }

        emit WithdrawalExecuted(req.withdrawalId, req.recipient, req.token, req.amount, req.nonce);
        return true;
    }

    /**
     * @notice Schedule 48h timelocked signer set update
     */
    function scheduleSignerUpdate(
        address[] calldata newSigners,
        uint256 newThreshold
    ) external onlySigner {
        require(newSigners.length >= 3, "Minimum 3 signers required");
        require(newThreshold >= 2 && newThreshold <= newSigners.length, "Invalid threshold");

        pendingUpdate = PendingSignerUpdate({
            newSigners: newSigners,
            newThreshold: newThreshold,
            effectiveTimestamp: block.timestamp + TIMELOCK_DELAY,
            executed: false
        });

        emit SignerUpdateScheduled(newSigners, newThreshold, pendingUpdate.effectiveTimestamp);
    }

    /**
     * @notice Execute timelocked signer update after 48h delay
     */
    function executeSignerUpdate() external onlySigner {
        require(pendingUpdate.effectiveTimestamp != 0, "No pending update");
        require(!pendingUpdate.executed, "Update already executed");
        require(block.timestamp >= pendingUpdate.effectiveTimestamp, "Timelock active");

        // Clear previous signers
        for (uint256 i = 0; i < signers.length; i++) {
            isSigner[signers[i]] = false;
        }
        delete signers;

        // Apply new signers
        for (uint256 i = 0; i < pendingUpdate.newSigners.length; i++) {
            address s = pendingUpdate.newSigners[i];
            require(s != address(0), "Invalid zero address");
            require(!isSigner[s], "Duplicate signer");
            isSigner[s] = true;
            signers.push(s);
        }
        threshold = pendingUpdate.newThreshold;
        pendingUpdate.executed = true;

        emit SignerUpdateExecuted(pendingUpdate.newSigners, pendingUpdate.newThreshold);
    }

    /**
     * @notice Cancel pending signer update
     */
    function cancelSignerUpdate() external onlySigner {
        require(pendingUpdate.effectiveTimestamp != 0 && !pendingUpdate.executed, "No active pending update");
        delete pendingUpdate;
        emit SignerUpdateCancelled();
    }

    /**
     * @notice Safely recover signer from signature, rejecting malleable signatures (secp256k1)
     */
    function recoverSigner(bytes32 digest, bytes memory sig) public pure returns (address) {
        require(sig.length == 65, "Invalid signature length");
        bytes32 r;
        bytes32 s;
        uint8 v;
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }

        // Enforce canonical s-values to prevent signature malleability
        require(
            uint256(s) <= 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D57617F32D5926182E4727FEE104A38,
            "Invalid malleable signature s-value"
        );
        require(v == 27 || v == 28, "Invalid signature v-value");

        address signer = ecrecover(digest, v, r, s);
        require(signer != address(0), "Invalid recovered signature");
        return signer;
    }

    receive() external payable {
        emit DepositReceived(msg.sender, msg.value);
    }
}
