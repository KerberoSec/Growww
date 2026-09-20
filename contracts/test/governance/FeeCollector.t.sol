// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "../../src/governance/FeeCollector.sol";

contract MockFeeAsset is ERC20 {
    constructor() ERC20("Settlement Currency", "eINR") {
        _mint(msg.sender, 100_000_000 * 1e18);
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockBurnableToken is ERC20, IBurnableToken {
    constructor() ERC20("Growww Token", "GROWWW") {
        _mint(msg.sender, 10_000_000 * 1e18);
    }

    function burn(uint256 amount) external override {
        _burn(msg.sender, amount);
    }
}

contract FeeCollectorTest is Test {
    FeeCollector public collector;
    MockFeeAsset public feeAsset;
    MockBurnableToken public utilityToken;

    address public admin = address(0xAA);
    address public settler = address(0xBB);
    address public treasury = address(0x101);
    address public rebatePool = address(0x102);
    address public insurance = address(0x103);

    function setUp() public {
        feeAsset = new MockFeeAsset();
        utilityToken = new MockBurnableToken();

        vm.startPrank(admin);
        collector = new FeeCollector(
            address(feeAsset),
            address(utilityToken),
            treasury,
            rebatePool,
            insurance,
            admin
        );
        collector.setSettlerAuthorization(settler, true);
        vm.stopPrank();

        feeAsset.mint(settler, 100_000 * 1e18);
        vm.prank(settler);
        feeAsset.approve(address(collector), type(uint256).max);

        utilityToken.transfer(address(collector), 500_000 * 1e18);
    }

    function test_70_20_10_RevenueDistribution() public {
        uint256 feeAmount = 10_000 * 1e18; // 10,000 eINR

        vm.prank(settler);
        collector.depositAndDistributeFees(feeAmount);

        assertEq(feeAsset.balanceOf(treasury), 7_000 * 1e18);     // 70%
        assertEq(feeAsset.balanceOf(rebatePool), 2_000 * 1e18);   // 20%
        assertEq(feeAsset.balanceOf(insurance), 1_000 * 1e18);    // 10%
        assertEq(collector.totalFeesCollected(), feeAmount);
    }

    function test_DepositRevertsUnauthorizedSettler() public {
        address hacker = address(0x999);
        feeAsset.mint(hacker, 1_000 * 1e18);

        vm.startPrank(hacker);
        feeAsset.approve(address(collector), type(uint256).max);
        vm.expectRevert(abi.encodeWithSelector(FeeCollector.UnauthorizedSettler.selector, hacker));
        collector.depositAndDistributeFees(1_000 * 1e18);
        vm.stopPrank();
    }

    function test_ExecuteTokenBurn_Success() public {
        uint256 burnAmount = 50_000 * 1e18;

        vm.prank(admin);
        collector.executeTokenBurn(burnAmount);

        assertEq(collector.totalTokensBurned(), burnAmount);
        assertEq(utilityToken.balanceOf(address(collector)), 450_000 * 1e18);
    }

    function test_UpdateWallets() public {
        address newTreasury = address(0x201);

        vm.prank(admin);
        collector.updateWallets(newTreasury, address(0), address(0));

        assertEq(collector.treasuryWallet(), newTreasury);
    }
}
