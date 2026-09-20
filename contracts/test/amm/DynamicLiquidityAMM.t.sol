// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "../../src/amm/DynamicLiquidityAMM.sol";

contract MockERC20Token is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {
        _mint(msg.sender, 100_000_000 * 1e18);
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract DynamicLiquidityAMMTest is Test {
    DynamicLiquidityAMM public amm;
    MockERC20Token public tokenA;
    MockERC20Token public tokenB;

    address public admin = address(0xAA);
    address public provider = address(0xBB);
    address public trader = address(0xCC);

    function setUp() public {
        tokenA = new MockERC20Token("Token A", "TKA");
        tokenB = new MockERC20Token("Token B", "TKB");

        vm.prank(admin);
        amm = new DynamicLiquidityAMM(address(tokenA), address(tokenB), admin);

        // Fund provider
        tokenA.mint(provider, 100_000 * 1e18);
        tokenB.mint(provider, 100_000 * 1e18);

        // Fund trader
        tokenA.mint(trader, 10_000 * 1e18);
        tokenB.mint(trader, 10_000 * 1e18);

        vm.startPrank(provider);
        tokenA.approve(address(amm), type(uint256).max);
        tokenB.approve(address(amm), type(uint256).max);
        vm.stopPrank();

        vm.startPrank(trader);
        tokenA.approve(address(amm), type(uint256).max);
        tokenB.approve(address(amm), type(uint256).max);
        vm.stopPrank();
    }

    function test_InitialLiquidityProvision() public {
        vm.prank(provider);
        uint256 shares = amm.addLiquidity(10_000 * 1e18, 10_000 * 1e18, 1000);

        assertTrue(shares > 0);
        assertEq(amm.reserve0(), 10_000 * 1e18);
        assertEq(amm.reserve1(), 10_000 * 1e18);
        assertEq(amm.totalLiquidityShares(), shares);
    }

    function test_SwapZeroForOne_Success() public {
        vm.prank(provider);
        amm.addLiquidity(10_000 * 1e18, 10_000 * 1e18, 1000);

        uint256 amountIn = 100 * 1e18;
        vm.prank(trader);
        uint256 amountOut = amm.swap(true, amountIn, 90 * 1e18);

        assertTrue(amountOut > 0);
        assertEq(amm.reserve0(), 10_100 * 1e18);
        assertEq(amm.reserve1(), 10_000 * 1e18 - amountOut);
    }

    function test_Swap_RevertsOnExcessSlippage() public {
        vm.prank(provider);
        amm.addLiquidity(10_000 * 1e18, 10_000 * 1e18, 1000);

        uint256 amountIn = 100 * 1e18;
        vm.prank(trader);
        vm.expectRevert(abi.encodeWithSelector(DynamicLiquidityAMM.SlippageExceeded.selector, 98862854398827705361, 100 * 1e18));
        amm.swap(true, amountIn, 100 * 1e18);
    }

    function test_DynamicFeeScalesWithVolatility() public {
        vm.prank(provider);
        amm.addLiquidity(10_000 * 1e18, 10_000 * 1e18, 1000);

        uint32 initialFee = amm.currentDynamicFeeBps();
        assertEq(initialFee, 1500);

        // Substantial swap to shift price ratio
        vm.prank(trader);
        amm.swap(true, 5000 * 1e18, 1);

        uint32 postSwapFee = amm.currentDynamicFeeBps();
        assertTrue(postSwapFee > initialFee);
    }

    function test_RemoveLiquidity_Success() public {
        vm.prank(provider);
        uint256 shares = amm.addLiquidity(10_000 * 1e18, 10_000 * 1e18, 1000);

        vm.prank(provider);
        (uint256 out0, uint256 out1) = amm.removeLiquidity(shares / 2, 4900 * 1e18, 4900 * 1e18);

        assertEq(out0, 5_000 * 1e18);
        assertEq(out1, 5_000 * 1e18);
        assertEq(amm.reserve0(), 5_000 * 1e18);
        assertEq(amm.reserve1(), 5_000 * 1e18);
    }
}
