// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {HybridPool} from "../../src/amm/HybridPool.sol";
import {DigitalRupee} from "../../src/tokens/DigitalRupee.sol";

contract HybridPoolTest is Test {
    HybridPool internal pool;
    DigitalRupee internal tokenA; // e.g. WBTC
    DigitalRupee internal tokenB; // e.g. USDT

    address internal admin = address(0xAA1);
    address internal operator = address(0xBB1);
    address internal router = address(0xCC1);
    address internal trader = address(0xDD1);

    function setUp() public {
        vm.startPrank(admin);
        tokenA = new DigitalRupee(admin);
        tokenB = new DigitalRupee(admin);

        pool = new HybridPool(admin, address(tokenA), address(tokenB), 15); // 0.15% fee
        pool.grantRole(pool.OPERATOR_ROLE(), operator);
        pool.grantRole(pool.ROUTER_ROLE(), router);

        // Mint tokens to operator for pool liquidity
        tokenA.mint(operator, 100 * 1e18, keccak256("OP-A"), "op@rbi.edr");
        tokenB.mint(operator, 6_000_000 * 1e18, keccak256("OP-B"), "op@rbi.edr");

        // Mint tokens to trader
        tokenA.mint(trader, 10 * 1e18, keccak256("TR-A"), "trader@rbi.edr");
        tokenB.mint(trader, 500_000 * 1e18, keccak256("TR-B"), "trader@rbi.edr");
        vm.stopPrank();

        // Operator adds initial liquidity: 10 WBTC and 600,000 USDT (1 WBTC = 60,000 USDT)
        vm.startPrank(operator);
        tokenA.approve(address(pool), 10 * 1e18);
        tokenB.approve(address(pool), 600_000 * 1e18);
        pool.addLiquidity(10 * 1e18, 600_000 * 1e18);
        vm.stopPrank();
    }

    function test_InitialReservesAndPricing() public {
        assertEq(pool.reserveA(), 10 * 1e18);
        assertEq(pool.reserveB(), 600_000 * 1e18);

        // Check quote for 1 WBTC in -> USDT out
        (uint256 amountOut, uint256 fee) = pool.getAmountOut(1 * 1e18, address(tokenA));
        assertGt(amountOut, 0);
        assertGt(fee, 0);
        // Approximately 600,000 / 11 = ~54,545 USDT
        assertGt(amountOut, 50_000 * 1e18);
        assertLt(amountOut, 60_000 * 1e18);
    }

    function test_DirectSwap_Success() public {
        uint256 tradeAmountA = 1 * 1e18; // 1 WBTC
        (uint256 expectedOut, ) = pool.getAmountOut(tradeAmountA, address(tokenA));

        vm.startPrank(trader);
        tokenA.approve(address(pool), tradeAmountA);

        uint256 balBBefore = tokenB.balanceOf(trader);
        uint256 actualOut = pool.swap(address(tokenA), tradeAmountA, expectedOut, trader);
        vm.stopPrank();

        assertEq(actualOut, expectedOut);
        assertEq(tokenB.balanceOf(trader), balBBefore + actualOut);
        assertEq(pool.reserveA(), 11 * 1e18);
        assertEq(pool.reserveB(), 600_000 * 1e18 - actualOut);
    }

    function test_RouterSwap_AuthorizedExecution() public {
        uint256 tradeAmountB = 60_000 * 1e18; // 60,000 USDT
        (uint256 expectedOutA, ) = pool.getAmountOut(tradeAmountB, address(tokenB));

        // Trader approves pool directly for Smart Order Router
        vm.prank(trader);
        tokenB.approve(address(pool), tradeAmountB);

        uint256 balABefore = tokenA.balanceOf(trader);

        // Router initiates atomic routerSwap on behalf of trader
        vm.prank(router);
        uint256 actualOutA = pool.routerSwap(trader, address(tokenB), tradeAmountB, expectedOutA, trader);

        assertEq(actualOutA, expectedOutA);
        assertEq(tokenA.balanceOf(trader), balABefore + actualOutA);
    }

    function test_SlippageProtection_Revert() public {
        uint256 tradeAmountA = 1 * 1e18;
        (uint256 expectedOut, ) = pool.getAmountOut(tradeAmountA, address(tokenA));

        vm.startPrank(trader);
        tokenA.approve(address(pool), tradeAmountA);

        // Demanding 1 token more than mathematically possible should revert
        vm.expectRevert(HybridPool.SlippageExceeded.selector);
        pool.swap(address(tokenA), tradeAmountA, expectedOut + 1e18, trader);
        vm.stopPrank();
    }
}
