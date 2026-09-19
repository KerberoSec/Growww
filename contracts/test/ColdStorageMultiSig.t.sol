// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/custody/ColdStorageMultiSig.sol";

contract MockToken {
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    function mint(address to, uint256 amount) external {
        balanceOf[to] += amount;
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        require(balanceOf[msg.sender] >= amount, "Insufficient balance");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        return true;
    }
}

contract ColdStorageMultiSigTest is TestBase {
    ColdStorageMultiSig vault;
    MockToken token;

    // 5 Institutional Signer Private Keys
    uint256 pk1 = 0x111111;
    uint256 pk2 = 0x222222;
    uint256 pk3 = 0x333333;
    uint256 pk4 = 0x444444;
    uint256 pk5 = 0x555555;

    address signer1;
    address signer2;
    address signer3;
    address signer4;
    address signer5;

    address payable recipient = payable(address(0x8888));

    function setUp() public {
        signer1 = vm.addr(pk1);
        signer2 = vm.addr(pk2);
        signer3 = vm.addr(pk3);
        signer4 = vm.addr(pk4);
        signer5 = vm.addr(pk5);

        // Sort signers in ascending address order for multi-sig verification
        address[] memory rawSigners = new address[](5);
        rawSigners[0] = signer1;
        rawSigners[1] = signer2;
        rawSigners[2] = signer3;
        rawSigners[3] = signer4;
        rawSigners[4] = signer5;

        // Bubble sort to ensure strictly ascending order
        for (uint256 i = 0; i < rawSigners.length; i++) {
            for (uint256 j = i + 1; j < rawSigners.length; j++) {
                if (rawSigners[i] > rawSigners[j]) {
                    address temp = rawSigners[i];
                    rawSigners[i] = rawSigners[j];
                    rawSigners[j] = temp;
                }
            }
        }

        // Initialize 3-of-5 threshold vault
        vault = new ColdStorageMultiSig(rawSigners, 3);
        token = new MockToken();

        // Fund vault with ETH and tokens
        vm.deal(address(vault), 50 ether);
        token.mint(address(vault), 1_000_000 * 1e6);
    }

    function _sortSignatures(
        bytes32 digest,
        uint256[] memory privateKeys
    ) internal returns (bytes[] memory) {
        // Collect recovered addresses & signatures
        uint256 len = privateKeys.length;
        address[] memory addrs = new address[](len);
        bytes[] memory sigs = new bytes[](len);

        for (uint256 i = 0; i < len; i++) {
            (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKeys[i], digest);
            sigs[i] = abi.encodePacked(r, s, v);
            addrs[i] = vm.addr(privateKeys[i]);
        }

        // Sort in ascending address order
        for (uint256 i = 0; i < len; i++) {
            for (uint256 j = i + 1; j < len; j++) {
                if (addrs[i] > addrs[j]) {
                    address tempA = addrs[i];
                    addrs[i] = addrs[j];
                    addrs[j] = tempA;

                    bytes memory tempS = sigs[i];
                    sigs[i] = sigs[j];
                    sigs[j] = tempS;
                }
            }
        }

        return sigs;
    }

    /* -------------------------------------------------------------------------- */
    /*                         1. CONSTRUCTOR & INITIALIZATION                    */
    /* -------------------------------------------------------------------------- */

    function test_Constructor_Success() public view {
        assertEq(vault.threshold(), 3, "Threshold must be 3");
        address[] memory s = vault.getSigners();
        assertEq(s.length, 5, "5 signers initialized");
    }

    function test_Constructor_RevertFewSigners() public {
        address[] memory few = new address[](2);
        few[0] = signer1;
        few[1] = signer2;

        vm.expectRevert("Minimum 3 signers required");
        new ColdStorageMultiSig(few, 2);
    }

    function test_Constructor_RevertInvalidThreshold() public {
        address[] memory five = new address[](5);
        five[0] = signer1;
        five[1] = signer2;
        five[2] = signer3;
        five[3] = signer4;
        five[4] = signer5;

        vm.expectRevert("Invalid quorum threshold");
        new ColdStorageMultiSig(five, 1); // < 2

        vm.expectRevert("Invalid quorum threshold");
        new ColdStorageMultiSig(five, 6); // > 5
    }

    /* -------------------------------------------------------------------------- */
    /*                         2. EIP-712 WITHDRAWAL EXECUTION                    */
    /* -------------------------------------------------------------------------- */

    function test_ExecuteWithdrawal_NativeEthSuccess() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_native_001"),
            recipient: recipient,
            token: address(0),
            amount: 5 ether,
            fee: 0,
            nonce: 1,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);

        // 3 of 5 signers sign
        uint256[] memory signersPks = new uint256[](3);
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        signersPks[2] = pk3;

        bytes[] memory sortedSigs = _sortSignatures(digest, signersPks);

        uint256 recipientBalBefore = recipient.balance;
        bool success = vault.executeWithdrawal(req, sortedSigs);

        assertTrue(success, "Withdrawal execution should succeed");
        assertEq(recipient.balance, recipientBalBefore + 5 ether);
        assertTrue(vault.executedWithdrawals(req.withdrawalId));
        assertTrue(vault.executedNonces(req.nonce));
    }

    function test_ExecuteWithdrawal_ERC20Success() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_erc20_001"),
            recipient: recipient,
            token: address(token),
            amount: 50_000 * 1e6, // 50,000 USDT
            fee: 0,
            nonce: 2,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);

        uint256[] memory signersPks = new uint256[](4); // 4 of 5 signers
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        signersPks[2] = pk4;
        signersPks[3] = pk5;

        bytes[] memory sortedSigs = _sortSignatures(digest, signersPks);

        bool success = vault.executeWithdrawal(req, sortedSigs);
        assertTrue(success);
        assertEq(token.balanceOf(recipient), 50_000 * 1e6);
    }

    /* -------------------------------------------------------------------------- */
    /*                         3. REPLAY & EXPIRY EDGE CASES                      */
    /* -------------------------------------------------------------------------- */

    function test_ExecuteWithdrawal_RevertExpired() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_expired"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 0,
            nonce: 10,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);
        uint256[] memory signersPks = new uint256[](3);
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        signersPks[2] = pk3;
        bytes[] memory sortedSigs = _sortSignatures(digest, signersPks);

        // Fast-forward past validity window
        vm.warp(block.timestamp + 1 hours + 1);

        vm.expectRevert("Withdrawal request expired");
        vault.executeWithdrawal(req, sortedSigs);
    }

    function test_ExecuteWithdrawal_RevertReplay() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_replay"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 0,
            nonce: 11,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);
        uint256[] memory signersPks = new uint256[](3);
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        signersPks[2] = pk3;
        bytes[] memory sortedSigs = _sortSignatures(digest, signersPks);

        vault.executeWithdrawal(req, sortedSigs);

        // Replay attempt
        vm.expectRevert("Withdrawal already executed");
        vault.executeWithdrawal(req, sortedSigs);
    }

    function test_ExecuteWithdrawal_RevertReusedNonce() public {
        ColdStorageMultiSig.WithdrawalRequest memory req1 = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_nonce_test_1"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 0,
            nonce: 25,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest1 = vault.hashWithdrawalRequest(req1);
        uint256[] memory signersPks = new uint256[](3);
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        signersPks[2] = pk3;
        bytes[] memory sortedSigs1 = _sortSignatures(digest1, signersPks);

        vault.executeWithdrawal(req1, sortedSigs1);

        // Different withdrawal ID but SAME nonce
        ColdStorageMultiSig.WithdrawalRequest memory req2 = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_nonce_test_2"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 0,
            nonce: 25, // Reused nonce
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest2 = vault.hashWithdrawalRequest(req2);
        bytes[] memory sortedSigs2 = _sortSignatures(digest2, signersPks);

        vm.expectRevert("Nonce already used");
        vault.executeWithdrawal(req2, sortedSigs2);
    }

    function test_ExecuteWithdrawal_RevertNonZeroFee() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_fee_violation"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 10, // Must be 0 at launch
            nonce: 30,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);
        uint256[] memory signersPks = new uint256[](3);
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        signersPks[2] = pk3;
        bytes[] memory sortedSigs = _sortSignatures(digest, signersPks);

        vm.expectRevert("Fee must be zero at launch");
        vault.executeWithdrawal(req, sortedSigs);
    }

    function test_ExecuteWithdrawal_RevertInsufficientQuorum() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_low_quorum"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 0,
            nonce: 35,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);
        uint256[] memory signersPks = new uint256[](2); // Only 2 signatures (threshold is 3)
        signersPks[0] = pk1;
        signersPks[1] = pk2;
        bytes[] memory sortedSigs = _sortSignatures(digest, signersPks);

        vm.expectRevert("Insufficient signatures for quorum");
        vault.executeWithdrawal(req, sortedSigs);
    }

    function test_ExecuteWithdrawal_RevertDuplicateSignatures() public {
        ColdStorageMultiSig.WithdrawalRequest memory req = ColdStorageMultiSig.WithdrawalRequest({
            withdrawalId: keccak256("withdraw_duplicate_sigs"),
            recipient: recipient,
            token: address(0),
            amount: 1 ether,
            fee: 0,
            nonce: 40,
            validUntil: block.timestamp + 1 hours
        });

        bytes32 digest = vault.hashWithdrawalRequest(req);
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(pk1, digest);
        bytes memory sig1 = abi.encodePacked(r1, s1, v1);

        // Array with duplicate sig1 to try to reach threshold of 3
        bytes[] memory dupeSigs = new bytes[](3);
        dupeSigs[0] = sig1;
        dupeSigs[1] = sig1;
        dupeSigs[2] = sig1;

        vm.expectRevert("Signatures must be strictly ordered without duplicates");
        vault.executeWithdrawal(req, dupeSigs);
    }

    /* -------------------------------------------------------------------------- */
    /*                         4. TIMELOCKED SIGNER UPDATES                       */
    /* -------------------------------------------------------------------------- */

    function test_ScheduleAndExecuteSignerUpdate_Success() public {
        address[] memory newSignerList = new address[](3);
        newSignerList[0] = address(0x9001);
        newSignerList[1] = address(0x9002);
        newSignerList[2] = address(0x9003);

        vm.prank(signer1);
        vault.scheduleSignerUpdate(newSignerList, 2);

        // Cannot execute before 48h
        vm.warp(block.timestamp + 47 hours);
        vm.prank(signer1);
        vm.expectRevert("Timelock active");
        vault.executeSignerUpdate();

        // Warp past 48h
        vm.warp(block.timestamp + 2 hours);
        vm.prank(signer1);
        vault.executeSignerUpdate();

        assertEq(vault.threshold(), 2);
        assertTrue(vault.isSigner(address(0x9001)));
        assertFalse(vault.isSigner(signer1));
    }

    function test_CancelSignerUpdate_Success() public {
        address[] memory newSignerList = new address[](3);
        newSignerList[0] = address(0x9001);
        newSignerList[1] = address(0x9002);
        newSignerList[2] = address(0x9003);

        vm.prank(signer1);
        vault.scheduleSignerUpdate(newSignerList, 2);

        // Cancel update
        vm.prank(signer2);
        vault.cancelSignerUpdate();

        // After cancel, executing update should revert
        vm.warp(block.timestamp + 50 hours);
        vm.prank(signer1);
        vm.expectRevert("No pending update");
        vault.executeSignerUpdate();
    }
}
