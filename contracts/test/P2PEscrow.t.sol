// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/settlement/P2PEscrow.sol";

/**
 * @notice Malicious buyer attempting reentrancy attacks on P2PEscrow
 */
contract MaliciousEscrowBuyer {
    P2PEscrow public escrow;
    bytes32 public targetTradeId;
    uint256 public attackAttempts;
    bool public attackSucceeded;
    string public attackFunction;

    constructor(address _escrow) {
        escrow = P2PEscrow(_escrow);
    }

    function setTarget(bytes32 _tradeId, string calldata _func) external {
        targetTradeId = _tradeId;
        attackFunction = _func;
        attackAttempts = 0;
        attackSucceeded = false;
    }

    receive() external payable {
        attackAttempts++;
        if (attackAttempts == 1) {
            if (keccak256(bytes(attackFunction)) == keccak256(bytes("releaseEscrow"))) {
                // Attempt to reenter releaseEscrow
                try escrow.releaseEscrow(targetTradeId) {
                    attackSucceeded = true;
                } catch {
                    // Expected: revert due to CEI state guard
                }
            } else if (keccak256(bytes(attackFunction)) == keccak256(bytes("raiseDispute"))) {
                try escrow.raiseDispute(targetTradeId) {
                    attackSucceeded = true;
                } catch {
                    // Expected: revert due to CEI state guard
                }
            }
        }
    }
}

/**
 * @notice Malicious seller attempting reentrancy attacks on P2PEscrow
 */
contract MaliciousEscrowSeller {
    P2PEscrow public escrow;
    bytes32 public targetTradeId;
    uint256 public attackAttempts;
    bool public attackSucceeded;

    constructor(address _escrow) {
        escrow = P2PEscrow(_escrow);
    }

    function createTrade(bytes32 _tradeId, address payable _buyer, uint256 _duration) external payable {
        targetTradeId = _tradeId;
        attackAttempts = 0;
        attackSucceeded = false;
        escrow.createEscrow{value: msg.value}(_tradeId, _buyer, _duration);
    }

    receive() external payable {
        attackAttempts++;
        if (attackAttempts == 1) {
            // Attempt to reenter cancelExpired
            try escrow.cancelExpired(targetTradeId) {
                attackSucceeded = true;
            } catch {
                // Expected: revert
            }
        }
    }
}

contract RevertingReceiver {
    receive() external payable {
        revert("Always reverts on payment");
    }
}

contract P2PEscrowTest is TestBase {
    P2PEscrow escrow;

    address arbitrator = address(0xAAAA);
    address payable seller = payable(address(0x1111));
    address payable buyer = payable(address(0x2222));
    address stranger = address(0x9999);

    bytes32 defaultTradeId = keccak256("p2p_trade_001");
    uint256 defaultAmount = 1 ether;
    uint256 defaultWindow = 1 hours; // 3600s >= 15 mins

    function setUp() public {
        escrow = new P2PEscrow(arbitrator);
        vm.deal(seller, 100 ether);
        vm.deal(buyer, 100 ether);
    }

    /* -------------------------------------------------------------------------- */
    /*                         1. CONSTRUCTOR & INITIAL STATE                     */
    /* -------------------------------------------------------------------------- */

    function test_Constructor_Success() public view {
        assertEq(escrow.arbitrator(), arbitrator);
    }

    function test_Constructor_RevertZeroArbitrator() public {
        vm.expectRevert("Invalid arbitrator");
        new P2PEscrow(address(0));
    }

    /* -------------------------------------------------------------------------- */
    /*                         2. ESCROW CREATION VALIDATIONS                     */
    /* -------------------------------------------------------------------------- */

    function test_CreateEscrow_Success() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        (
            bytes32 tradeId,
            address payable s,
            address payable b,
            uint256 amount,
            uint256 expiry,
            P2PEscrow.EscrowStatus status,
            string memory fiatRef
        ) = escrow.escrows(defaultTradeId);

        assertEq(tradeId, defaultTradeId);
        assertEq(s, seller);
        assertEq(b, buyer);
        assertEq(amount, defaultAmount);
        assertEq(expiry, block.timestamp + defaultWindow);
        assertEq(uint8(status), uint8(P2PEscrow.EscrowStatus.FUNDED));
        assertEq(bytes(fiatRef).length, 0);
        assertEq(address(escrow).balance, defaultAmount);
    }

    function test_CreateEscrow_RevertZeroValue() public {
        vm.prank(seller);
        vm.expectRevert("No funds locked");
        escrow.createEscrow{value: 0}(defaultTradeId, buyer, defaultWindow);
    }

    function test_CreateEscrow_RevertZeroBuyer() public {
        vm.prank(seller);
        vm.expectRevert("Invalid buyer");
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, payable(address(0)), defaultWindow);
    }

    function test_CreateEscrow_RevertSelfBuyer() public {
        vm.prank(seller);
        vm.expectRevert("Invalid buyer");
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, seller, defaultWindow);
    }

    function test_CreateEscrow_RevertDuplicateTradeId() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(seller);
        vm.expectRevert("Trade ID already exists");
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);
    }

    function test_CreateEscrow_RevertShortWindow() public {
        vm.prank(seller);
        vm.expectRevert("Window too short");
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, 14 minutes);
    }

    /* -------------------------------------------------------------------------- */
    /*                         3. HAPPY PATH FIAT PAID & RELEASE                  */
    /* -------------------------------------------------------------------------- */

    function test_MarkPaidAndRelease_Success() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Buyer marks paid with UPI reference
        vm.prank(buyer);
        escrow.markFiatPaid(defaultTradeId, "UPI/2026/09/GROWWW12345");

        (, , , , , P2PEscrow.EscrowStatus statusPaid, string memory fiatRef) = escrow.escrows(defaultTradeId);
        assertEq(uint8(statusPaid), uint8(P2PEscrow.EscrowStatus.FIAT_PAID));
        assertEq(fiatRef, "UPI/2026/09/GROWWW12345");

        // Seller confirms fiat receipt and releases crypto
        uint256 buyerBalanceBefore = buyer.balance;
        vm.prank(seller);
        escrow.releaseEscrow(defaultTradeId);

        (, , , , , P2PEscrow.EscrowStatus statusCompleted, ) = escrow.escrows(defaultTradeId);
        assertEq(uint8(statusCompleted), uint8(P2PEscrow.EscrowStatus.COMPLETED));
        assertEq(buyer.balance, buyerBalanceBefore + defaultAmount);
        assertEq(address(escrow).balance, 0);
    }

    function test_ReleaseDirectlyFromFunded_Success() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Seller can release even before buyer marks paid
        uint256 buyerBalanceBefore = buyer.balance;
        vm.prank(seller);
        escrow.releaseEscrow(defaultTradeId);

        (, , , , , P2PEscrow.EscrowStatus status, ) = escrow.escrows(defaultTradeId);
        assertEq(uint8(status), uint8(P2PEscrow.EscrowStatus.COMPLETED));
        assertEq(buyer.balance, buyerBalanceBefore + defaultAmount);
    }

    /* -------------------------------------------------------------------------- */
    /*                         4. INVALID TRANSITIONS & TIMEOUTS                  */
    /* -------------------------------------------------------------------------- */

    function test_MarkFiatPaid_RevertNonBuyer() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(stranger);
        vm.expectRevert("Only buyer can mark paid");
        escrow.markFiatPaid(defaultTradeId, "FAKE_REF");
    }

    function test_MarkFiatPaid_RevertExpired() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Advance time past expiry
        vm.warp(block.timestamp + defaultWindow + 1);

        vm.prank(buyer);
        vm.expectRevert("Payment window expired");
        escrow.markFiatPaid(defaultTradeId, "LATE_REF");
    }

    function test_MarkFiatPaid_RevertWrongStatus() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(buyer);
        escrow.markFiatPaid(defaultTradeId, "REF_1");

        // Attempting to mark paid again while already FIAT_PAID
        vm.prank(buyer);
        vm.expectRevert("Invalid status");
        escrow.markFiatPaid(defaultTradeId, "REF_2");
    }

    function test_ReleaseEscrow_RevertNonSeller() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(buyer);
        vm.expectRevert("Only seller can release");
        escrow.releaseEscrow(defaultTradeId);
    }

    function test_ReleaseEscrow_RevertDoubleRelease() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(seller);
        escrow.releaseEscrow(defaultTradeId);

        vm.prank(seller);
        vm.expectRevert("Cannot release");
        escrow.releaseEscrow(defaultTradeId);
    }

    /* -------------------------------------------------------------------------- */
    /*                         5. EXPIRY CANCELLATION                             */
    /* -------------------------------------------------------------------------- */

    function test_CancelExpired_Success() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Advance time past payment window
        vm.warp(block.timestamp + defaultWindow + 1);

        uint256 sellerBalanceBefore = seller.balance;
        vm.prank(seller);
        escrow.cancelExpired(defaultTradeId);

        (, , , , , P2PEscrow.EscrowStatus status, ) = escrow.escrows(defaultTradeId);
        assertEq(uint8(status), uint8(P2PEscrow.EscrowStatus.CANCELLED));
        assertEq(seller.balance, sellerBalanceBefore + defaultAmount);
        assertEq(address(escrow).balance, 0);
    }

    function test_CancelExpired_RevertBeforeExpiry() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Try to cancel while window is still active
        vm.warp(block.timestamp + defaultWindow - 10);

        vm.prank(seller);
        vm.expectRevert("Payment window still active");
        escrow.cancelExpired(defaultTradeId);
    }

    function test_CancelExpired_RevertIfFiatPaid() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(buyer);
        escrow.markFiatPaid(defaultTradeId, "BANK_REF_PAID");

        // Advance time past expiry
        vm.warp(block.timestamp + defaultWindow + 100);

        // Seller CANNOT cancel because buyer marked fiat paid
        vm.prank(seller);
        vm.expectRevert("Cannot cancel in current state");
        escrow.cancelExpired(defaultTradeId);
    }

    function test_CancelExpired_RevertNonSeller() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.warp(block.timestamp + defaultWindow + 1);

        vm.prank(stranger);
        vm.expectRevert("Only seller can cancel");
        escrow.cancelExpired(defaultTradeId);
    }

    /* -------------------------------------------------------------------------- */
    /*                         6. DISPUTE RESOLUTION                              */
    /* -------------------------------------------------------------------------- */

    function test_DisputeAndResolveToBuyer_Success() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(buyer);
        escrow.markFiatPaid(defaultTradeId, "DISPUTED_IMPS_REF");

        // Buyer opens dispute because seller refuses to release
        vm.prank(buyer);
        escrow.raiseDispute(defaultTradeId);

        (, , , , , P2PEscrow.EscrowStatus statusDisputed, ) = escrow.escrows(defaultTradeId);
        assertEq(uint8(statusDisputed), uint8(P2PEscrow.EscrowStatus.DISPUTED));

        // Arbitrator resolves to buyer
        uint256 buyerBalanceBefore = buyer.balance;
        vm.prank(arbitrator);
        escrow.resolveDispute(defaultTradeId, true);

        (, , , , , P2PEscrow.EscrowStatus statusResolved, ) = escrow.escrows(defaultTradeId);
        assertEq(uint8(statusResolved), uint8(P2PEscrow.EscrowStatus.RESOLVED_TO_BUYER));
        assertEq(buyer.balance, buyerBalanceBefore + defaultAmount);
    }

    function test_DisputeAndResolveToSeller_Success() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Seller opens dispute (e.g. fraudulent fiat reference)
        vm.prank(seller);
        escrow.raiseDispute(defaultTradeId);

        // Arbitrator resolves to seller
        uint256 sellerBalanceBefore = seller.balance;
        vm.prank(arbitrator);
        escrow.resolveDispute(defaultTradeId, false);

        (, , , , , P2PEscrow.EscrowStatus statusResolved, ) = escrow.escrows(defaultTradeId);
        assertEq(uint8(statusResolved), uint8(P2PEscrow.EscrowStatus.RESOLVED_TO_SELLER));
        assertEq(seller.balance, sellerBalanceBefore + defaultAmount);
    }

    function test_RaiseDispute_RevertUnauthorized() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(stranger);
        vm.expectRevert("Unauthorized party");
        escrow.raiseDispute(defaultTradeId);
    }

    function test_ResolveDispute_RevertNonArbitrator() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        vm.prank(buyer);
        escrow.raiseDispute(defaultTradeId);

        vm.prank(stranger);
        vm.expectRevert("Only arbitrator allowed");
        escrow.resolveDispute(defaultTradeId, true);
    }

    function test_ResolveDispute_RevertNotDisputed() public {
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(defaultTradeId, buyer, defaultWindow);

        // Not yet disputed
        vm.prank(arbitrator);
        vm.expectRevert("Not disputed");
        escrow.resolveDispute(defaultTradeId, true);
    }

    /* -------------------------------------------------------------------------- */
    /*                         7. REENTRANCY ATTACK RESISTANCE                    */
    /* -------------------------------------------------------------------------- */

    function test_Reentrancy_BuyerCannotDrainOnRelease() public {
        MaliciousEscrowBuyer maliciousBuyer = new MaliciousEscrowBuyer(address(escrow));
        bytes32 attackTradeId = keccak256("attack_trade_001");

        // Fund escrow with malicious buyer
        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(attackTradeId, payable(address(maliciousBuyer)), defaultWindow);

        // Target reentrancy into releaseEscrow
        maliciousBuyer.setTarget(attackTradeId, "releaseEscrow");

        // Seller releases
        vm.prank(seller);
        escrow.releaseEscrow(attackTradeId);

        // Verify attack failed: reentrancy was prevented by state change prior to call
        assertFalse(maliciousBuyer.attackSucceeded(), "Reentrant call must not succeed");
        assertEq(maliciousBuyer.attackAttempts(), 1, "Only one payment received");
        assertEq(address(maliciousBuyer).balance, defaultAmount, "Buyer only gets its exact escrow amount");
    }

    function test_Reentrancy_SellerCannotDrainOnCancel() public {
        MaliciousEscrowSeller maliciousSeller = new MaliciousEscrowSeller(address(escrow));
        bytes32 attackTradeId = keccak256("attack_trade_seller_cancel");

        // Seller creates trade with defaultAmount sent
        maliciousSeller.createTrade{value: defaultAmount}(attackTradeId, buyer, defaultWindow);

        // Advance time past expiry
        vm.warp(block.timestamp + defaultWindow + 1);

        // Seller calls cancelExpired
        vm.prank(address(maliciousSeller));
        escrow.cancelExpired(attackTradeId);

        assertFalse(maliciousSeller.attackSucceeded(), "Reentrant cancel must fail");
        assertEq(maliciousSeller.attackAttempts(), 1, "Only one refund received");
        assertEq(address(maliciousSeller).balance, defaultAmount, "Seller only gets its exact refund amount");
    }

    /* -------------------------------------------------------------------------- */
    /*                         8. REVERTING RECEIVER HANDLING                     */
    /* -------------------------------------------------------------------------- */

    function test_ReleaseEscrow_RevertOnTransferFailure() public {
        RevertingReceiver revBuyer = new RevertingReceiver();
        bytes32 tradeId = keccak256("reverting_buyer_trade");

        vm.prank(seller);
        escrow.createEscrow{value: defaultAmount}(tradeId, payable(address(revBuyer)), defaultWindow);

        vm.prank(seller);
        vm.expectRevert("Payout failed");
        escrow.releaseEscrow(tradeId);
    }
}
