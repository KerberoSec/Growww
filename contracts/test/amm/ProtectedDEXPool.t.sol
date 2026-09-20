// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import {ProtectedDEXPool} from "../../src/amm/ProtectedDEXPool.sol";
import {IProtectedDEXPool, IFlashBorrower} from "../../src/interfaces/amm/IProtectedDEXPool.sol";

contract MockDEXToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockFlashBorrower is IFlashBorrower {
    bool public attemptReentrancy;
    bool public failRepayment;

    function setAttemptReentrancy(bool status) external {
        attemptReentrancy = status;
    }

    function setFailRepayment(bool status) external {
        failRepayment = status;
    }

    function onFlashLoan(
        address /*initiator*/,
        address token,
        uint256 amount,
        uint256 fee,
        bytes calldata /*data*/
    ) external override returns (bytes32) {
        if (attemptReentrancy) {
            // Attempt forbidden reentrancy
            ProtectedDEXPool(msg.sender).flashLoan(address(this), token, amount / 2, "");
        }

        if (!failRepayment) {
            // Repay amount + fee back to pool
            ERC20(token).transfer(msg.sender, amount + fee);
        }

        return keccak256("ERC3156FlashBorrower.onFlashLoan");
    }
}

contract ProtectedDEXPoolTest is Test {
    ProtectedDEXPool public pool;
    MockDEXToken public tokenA;
    MockDEXToken public tokenB;
    MockFlashBorrower public borrower;

    address public admin = address(0xAD01);
    address public alice = address(0x1001);
    address public bob = address(0x1002);

    uint256 public constant INITIAL_MINT = 1_000_000 ether;

    function setUp() public {
        tokenA = new MockDEXToken("Token A", "TKNA");
        tokenB = new MockDEXToken("Token B", "TKNB");
        borrower = new MockFlashBorrower();

        // 30 bps swap fee, 9 bps flash loan fee, 500 bps (5%) max TWAP deviation,
        // 300 bps (3%) max block price impact, 300 seconds TWAP window
        vm.prank(admin);
        pool = new ProtectedDEXPool(
            address(tokenA),
            address(tokenB),
            30,   // swapFeeBps
            9,    // flashLoanFeeBps
            500,  // maxTwapDeviationBps
            300,  // maxBlockPriceImpactBps
            300,  // twapWindow
            admin
        );

        tokenA.mint(alice, INITIAL_MINT);
        tokenB.mint(alice, INITIAL_MINT);
        tokenA.mint(bob, INITIAL_MINT);
        tokenB.mint(bob, INITIAL_MINT);
        tokenA.mint(address(borrower), 10_000 ether);
        tokenB.mint(address(borrower), 10_000 ether);

        vm.prank(alice);
        tokenA.approve(address(pool), type(uint256).max);
        vm.prank(alice);
        tokenB.approve(address(pool), type(uint256).max);

        vm.prank(bob);
        tokenA.approve(address(pool), type(uint256).max);
        vm.prank(bob);
        tokenB.approve(address(pool), type(uint256).max);
    }

    function test_Deployment_Parameters() public view {
        assertTrue(pool.token0() != address(0));
        assertTrue(pool.token1() != address(0));
        assertEq(pool.swapFeeBps(), 30);
        assertEq(pool.flashLoanFeeBps(), 9);
        assertEq(pool.maxTwapDeviationBps(), 500);
        assertEq(pool.maxBlockPriceImpactBps(), 300);
        assertEq(pool.twapWindow(), 300);
        assertTrue(pool.flashLoanProtectionEnabled());
    }

    function test_AddAndRemoveLiquidity() public {
        vm.prank(alice);
        (uint256 a0, uint256 a1, uint256 lpShares) = pool.addLiquidity(
            10_000 ether,
            10_000 ether,
            alice
        );

        assertEq(a0, 10_000 ether);
        assertEq(a1, 10_000 ether);
        assertTrue(lpShares > 0);
        assertEq(pool.balanceOf(alice), lpShares);

        vm.prank(alice);
        (uint256 r0, uint256 r1) = pool.removeLiquidity(lpShares / 2, alice);
        assertTrue(r0 > 4_900 ether);
        assertTrue(r1 > 4_900 ether);
    }

    function test_SameBlockTradeRestricted_FlashLoanDefense() public {
        vm.prank(alice);
        pool.addLiquidity(100_000 ether, 100_000 ether, alice);

        address t0 = pool.token0();

        // 1st swap in block N
        vm.prank(bob);
        pool.swap(100 ether, 90 ether, t0, bob);

        // 2nd swap by same user in same block reverts
        vm.prank(bob);
        vm.expectRevert(
            abi.encodeWithSelector(
                IProtectedDEXPool.SameBlockTradeRestricted.selector,
                bob,
                block.number
            )
        );
        pool.swap(50 ether, 40 ether, t0, bob);

        // Moving to next block clears restriction
        vm.roll(block.number + 1);
        vm.prank(bob);
        pool.swap(50 ether, 40 ether, t0, bob);
    }

    function test_MaxBlockPriceImpactExceeded() public {
        // Pool with 10,000 liquidity
        vm.prank(alice);
        pool.addLiquidity(10_000 ether, 10_000 ether, alice);

        address t0 = pool.token0();

        // Roll to block 10
        vm.roll(10);
        // Small initial swap to fix block start price
        vm.prank(alice);
        pool.swap(10 ether, 9 ether, t0, alice);

        // Bob attempts massive swap that would shift price > 300 bps (3%) in same block
        vm.prank(bob);
        vm.expectRevert();
        pool.swap(800 ether, 500 ether, t0, bob);
    }

    function test_TWAPConsultationAndAccumulation() public {
        vm.prank(alice);
        pool.addLiquidity(50_000 ether, 50_000 ether, alice);

        address t0 = pool.token0();

        // Advance time 500 seconds and roll block
        vm.warp(block.timestamp + 500);
        vm.roll(block.number + 50);

        uint256 twapOut = pool.consultTwap(t0, 100 ether, 500);
        assertEq(twapOut, 100 ether); // 1:1 price
    }

    function test_FlashLoan_SuccessAndFeeRepayment() public {
        vm.prank(alice);
        pool.addLiquidity(50_000 ether, 50_000 ether, alice);

        address t0 = pool.token0();
        uint256 loanAmount = 1_000 ether;
        uint256 expectedFee = (loanAmount * 9) / 10000; // 0.9 ether

        uint256 poolBalBefore = ERC20(t0).balanceOf(address(pool));

        bool success = pool.flashLoan(address(borrower), t0, loanAmount, "");
        assertTrue(success);

        // Verify pool balance increased by flash loan fee
        assertEq(ERC20(t0).balanceOf(address(pool)), poolBalBefore + expectedFee);
    }

    function test_FlashLoan_RevertsIfUnrepaid() public {
        vm.prank(alice);
        pool.addLiquidity(50_000 ether, 50_000 ether, alice);

        address t0 = pool.token0();
        borrower.setFailRepayment(true);

        vm.expectRevert();
        pool.flashLoan(address(borrower), t0, 1_000 ether, "");
    }

    function test_FlashLoan_ReentrancyBlocked() public {
        vm.prank(alice);
        pool.addLiquidity(50_000 ether, 50_000 ether, alice);

        address t0 = pool.token0();
        borrower.setAttemptReentrancy(true);

        vm.expectRevert();
        pool.flashLoan(address(borrower), t0, 1_000 ether, "");
    }

    function test_AdminControls() public {
        vm.prank(admin);
        pool.setMaxTwapDeviation(600);
        assertEq(pool.maxTwapDeviationBps(), 600);

        vm.prank(admin);
        pool.setMaxBlockPriceImpact(400);
        assertEq(pool.maxBlockPriceImpactBps(), 400);

        vm.prank(admin);
        pool.setFlashLoanFee(15);
        assertEq(pool.flashLoanFeeBps(), 15);

        vm.prank(admin);
        pool.toggleFlashLoanProtection(false);
        assertFalse(pool.flashLoanProtectionEnabled());
    }
}
