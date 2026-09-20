// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "../../src/liquidity/OffHoursCLMM.sol";
import "../../src/liquidity/ICLMMExchange.sol";

contract MockToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {
        _mint(msg.sender, 100_000_000 * 1e18);
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract OffHoursCLMMTest is Test {
    OffHoursCLMM public clmm;
    MockToken public token0;
    MockToken public token1;

    address public admin = address(0xAA);
    address public oracle = address(0xBB);
    address public trader = address(0xCC);

    uint160 public initialPrice = 1000 * 1e18; // base price

    function setUp() public {
        token0 = new MockToken("Token0", "TK0");
        token1 = new MockToken("Token1", "TK1");

        // Ensure token0 address < token1 address
        if (address(token0) > address(token1)) {
            MockToken temp = token0;
            token0 = token1;
            token1 = temp;
        }

        vm.startPrank(admin);
        clmm = new OffHoursCLMM(
            address(token0),
            address(token1),
            60,
            initialPrice,
            admin
        );
        clmm.setOracleFeeder(oracle);
        clmm.addLiquidity(100_000_000 * 1e18);
        vm.stopPrank();

        // Fund pool
        token0.mint(address(clmm), 10_000_000 * 1e18);
        token1.mint(address(clmm), 10_000_000 * 1e18);

        // Fund trader
        token0.mint(trader, 1_000_000 * 1e18);
        token1.mint(trader, 1_000_000 * 1e18);

        vm.startPrank(trader);
        token0.approve(address(clmm), type(uint256).max);
        token1.approve(address(clmm), type(uint256).max);
        vm.stopPrank();
    }

    function test_InitialCollarParameters() public {
        ICLMMExchange.CollarConfig memory cfg = clmm.getCollarConfig();
        assertEq(cfg.collarBandBps, 500); // 5%
        assertEq(cfg.primaryClosingPriceX96, initialPrice);
        assertEq(cfg.minCollarPriceX96, (initialPrice * 95) / 100);
        assertEq(cfg.maxCollarPriceX96, (initialPrice * 105) / 100);
        assertTrue(cfg.isCollarEnforced);
    }

    function test_Swap_SuccessWithinCollar() public {
        vm.startPrank(trader);
        ICLMMExchange.SwapParams memory params = ICLMMExchange.SwapParams({
            recipient: trader,
            zeroForOne: true,
            amountSpecified: 1000 * 1e18,
            sqrtPriceLimitX96: 0,
            data: ""
        });

        (int256 a0, int256 a1) = clmm.swap(params);
        vm.stopPrank();

        assertEq(a0, 1000 * 1e18);
        assertTrue(a1 < 0); // Trader received token1
    }

    function test_Swap_RevertsWhenOutsideTradingHours() public {
        vm.prank(admin);
        clmm.setOffHoursActive(false, false);

        vm.startPrank(trader);
        ICLMMExchange.SwapParams memory params = ICLMMExchange.SwapParams({
            recipient: trader,
            zeroForOne: true,
            amountSpecified: 100 * 1e18,
            sqrtPriceLimitX96: 0,
            data: ""
        });

        vm.expectRevert(abi.encodeWithSelector(ICLMMExchange.OutsideTradingHours.selector));
        clmm.swap(params);
        vm.stopPrank();
    }

    function test_OracleUpdatesPrimaryClosingPrice() public {
        uint160 newPrice = 1100 * 1e18;
        vm.prank(oracle);
        clmm.updatePrimaryClosingPrice(newPrice);

        ICLMMExchange.CollarConfig memory cfg = clmm.getCollarConfig();
        assertEq(cfg.primaryClosingPriceX96, newPrice);
        assertEq(cfg.minCollarPriceX96, (newPrice * 95) / 100);
        assertEq(cfg.maxCollarPriceX96, (newPrice * 105) / 100);
    }

    function test_UpdateCollarBand() public {
        vm.prank(admin);
        clmm.updateCollarBand(1000); // 10%

        ICLMMExchange.CollarConfig memory cfg = clmm.getCollarConfig();
        assertEq(cfg.collarBandBps, 1000);
        assertEq(cfg.minCollarPriceX96, (initialPrice * 90) / 100);
        assertEq(cfg.maxCollarPriceX96, (initialPrice * 110) / 100);
    }
}
