// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import {LiquidationEngine} from "../../src/liquidation/LiquidationEngine.sol";
import {ILiquidationEngine} from "../../src/interfaces/liquidation/ILiquidationEngine.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract MockLiquidationToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockKYCRegistry is IIdentityRegistry {
    mapping(address => bool) public verified;

    function setVerified(address user, bool status) external {
        verified[user] = status;
    }

    function isVerified(address userAddress) external view override returns (bool) {
        return verified[userAddress];
    }

    function registerIdentity(address, bytes32, uint16, uint8) external override {}
    function batchRegisterIdentity(address[] calldata, bytes32[] calldata, uint16[] calldata, uint8[] calldata) external override {}
    function deleteIdentity(address) external override {}
    function updateSanctionStatus(bytes32, bool) external override {}
    function updateCountry(address, uint16) external override {}
    function updateKycTier(address, uint8) external override {}
    function getInvestorClaim(address) external pure override returns (InvestorClaim memory) {
        return InvestorClaim(bytes32(0), 0, 0, false, 0);
    }
    function getInvestorCountry(address) external pure override returns (uint16) { return 0; }
    function getIdentityId(address) external pure override returns (bytes32) { return bytes32(0); }
    function contains(address) external pure override returns (bool) { return true; }
    function isSanctioned(bytes32) external pure override returns (bool) { return false; }
    function totalIdentities() external pure override returns (uint256) { return 0; }
}

contract LiquidationEngineTest is Test {
    LiquidationEngine public engine;
    MockLiquidationToken public collateralToken;
    MockLiquidationToken public debtToken;
    MockKYCRegistry public identityRegistry;

    address public admin = address(0xAD01);
    address public alice = address(0x1001); // Borrower
    address public bob = address(0x1002);   // Liquidator
    address public nonKycLiquidator = address(0x9999);

    function setUp() public {
        identityRegistry = new MockKYCRegistry();
        identityRegistry.setVerified(admin, true);
        identityRegistry.setVerified(alice, true);
        identityRegistry.setVerified(bob, true);

        vm.prank(admin);
        engine = new LiquidationEngine(address(identityRegistry), admin);

        collateralToken = new MockLiquidationToken("Digital Equity", "DEQ");
        debtToken = new MockLiquidationToken("Digital INR", "eINR");

        // Admin configures fallback prices:
        // Collateral = $100 (100 * 1e8)
        // Debt = $1 (1 * 1e8)
        vm.prank(admin);
        engine.setFallbackPrice(address(collateralToken), 100 * 1e8);
        vm.prank(admin);
        engine.setFallbackPrice(address(debtToken), 1 * 1e8);

        // Configure Collateral: 75% LTV (7500 bps), 80% Liquidation Threshold (8000 bps),
        // 5% Bonus (500 bps), Volatility 0, isSecurityToken = true
        vm.prank(admin);
        engine.configureCollateral(
            address(collateralToken),
            7500, // baseLtvBps
            8000, // liquidationThresholdBps
            500,  // liquidationBonusBps
            0,    // volatilityIndex
            true, // isSecurityToken
            address(0)
        );

        // Configure Debt Asset
        vm.prank(admin);
        engine.setDebtAsset(address(debtToken), true, address(0));

        // Fund users
        collateralToken.mint(alice, 1000 ether);
        debtToken.mint(bob, 100_000 ether);
        debtToken.mint(nonKycLiquidator, 100_000 ether);
        debtToken.mint(address(engine), 100_000 ether); // engine has liquidity to lend

        // Approvals
        vm.prank(alice);
        collateralToken.approve(address(engine), type(uint256).max);

        vm.prank(bob);
        debtToken.approve(address(engine), type(uint256).max);

        vm.prank(nonKycLiquidator);
        debtToken.approve(address(engine), type(uint256).max);
    }

    function test_CollateralDepositAndBorrow() public {
        // Alice deposits 10 DEQ ($1,000 USD)
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);

        assertEq(engine.getUserCollateral(alice, address(collateralToken)), 10 ether);

        // Alice borrows 500 eINR ($500 USD) -> LTV = 50%, Threshold = 80%
        // Adjusted collateral = 1000 * 80% = 800 USD. Debt = 500 USD.
        // HF = 800 / 500 * 10000 = 16000 bps (1.6x)
        vm.prank(alice);
        engine.borrow(address(debtToken), 500 ether);

        assertEq(engine.getUserDebt(alice, address(debtToken)), 500 ether);
        assertEq(engine.getHealthFactor(alice), 16000);
    }

    function test_WithdrawCollateral_RevertsIfUndercollateralized() public {
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);

        vm.prank(alice);
        engine.borrow(address(debtToken), 500 ether);

        // Trying to withdraw 6 DEQ leaves 4 DEQ ($400), which cannot support $500 debt
        vm.prank(alice);
        vm.expectRevert();
        engine.withdrawCollateral(address(collateralToken), 6 ether);

        // Withdrawing 2 DEQ leaves 8 DEQ ($800 * 0.8 = $640 >= $500), should succeed
        vm.prank(alice);
        engine.withdrawCollateral(address(collateralToken), 2 ether);
        assertEq(engine.getUserCollateral(alice, address(collateralToken)), 8 ether);
    }

    function test_DynamicLiquidationThreshold_VolatilityAdjustment() public {
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);
        vm.prank(alice);
        engine.borrow(address(debtToken), 700 ether);

        uint256 initialHf = engine.getHealthFactor(alice);

        // Market volatility surges to 600 bps (6%)
        // Dynamic threshold drops from 8000 to 8000 - 300 = 7700 bps
        vm.prank(admin);
        engine.setVolatilityIndex(address(collateralToken), 600);

        uint256 adjustedThreshold = engine.getDynamicLiquidationThreshold(alice, address(collateralToken));
        assertEq(adjustedThreshold, 7700);

        uint256 updatedHf = engine.getHealthFactor(alice);
        assertTrue(updatedHf < initialHf);
    }

    function test_Liquidate_RevertIfHealthy() public {
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);
        vm.prank(alice);
        engine.borrow(address(debtToken), 500 ether);

        // Position is healthy (HF = 1.6x)
        vm.prank(bob);
        vm.expectRevert();
        engine.liquidatePosition(alice, address(collateralToken), address(debtToken), 250 ether);
    }

    function test_Liquidate_ModeratelyUnderwater_CloseFactorEnforced() public {
        // Alice deposits 10 DEQ ($1000) and borrows 750 eINR ($750)
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);
        vm.prank(alice);
        engine.borrow(address(debtToken), 750 ether);

        // Collateral price drops from $100 to $90
        // Total collateral = 10 * 90 = $900.
        // Adjusted collateral = $900 * 80% = $720.
        // Debt = $750.
        // HF = 720 / 750 * 10000 = 9600 bps (0.96x) -> Moderately underwater!
        vm.prank(admin);
        engine.setFallbackPrice(address(collateralToken), 90 * 1e8);

        uint256 hf = engine.getHealthFactor(alice);
        assertEq(hf, 9600);

        // Liquidator attempts to cover 400 eINR (> 50% max close factor 375 eINR) -> reverts!
        vm.prank(bob);
        vm.expectRevert();
        engine.liquidatePosition(alice, address(collateralToken), address(debtToken), 400 ether);

        // Liquidator covers 300 eINR (<= 50%) -> succeeds!
        uint256 bobCollBalBefore = collateralToken.balanceOf(bob);
        vm.prank(bob);
        uint256 seized = engine.liquidatePosition(
            alice,
            address(collateralToken),
            address(debtToken),
            300 ether
        );

        assertTrue(seized > 0);
        assertEq(collateralToken.balanceOf(bob), bobCollBalBefore + seized);
        assertEq(engine.getUserDebt(alice, address(debtToken)), 450 ether);
    }

    function test_Liquidate_SeverelyUnderwater_FullLiquidationAllowed() public {
        // Alice deposits 10 DEQ ($1000) and borrows 750 eINR ($750)
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);
        vm.prank(alice);
        engine.borrow(address(debtToken), 750 ether);

        // Collateral price crashes from $100 to $70
        // Total collateral = 10 * 70 = $700.
        // Adjusted collateral = $700 * 80% = $560.
        // Debt = $750.
        // HF = 560 / 750 * 10000 = 7466 bps (< 9000 bps) -> Severely underwater!
        vm.prank(admin);
        engine.setFallbackPrice(address(collateralToken), 70 * 1e8);

        // 100% liquidation is allowed
        vm.prank(bob);
        uint256 seized = engine.liquidatePosition(
            alice,
            address(collateralToken),
            address(debtToken),
            750 ether
        );

        assertTrue(seized > 0);
        assertEq(engine.getUserDebt(alice, address(debtToken)), 0);
    }

    function test_Liquidate_RevertsIfLiquidatorNotKYC() public {
        vm.prank(alice);
        engine.depositCollateral(address(collateralToken), 10 ether);
        vm.prank(alice);
        engine.borrow(address(debtToken), 750 ether);

        vm.prank(admin);
        engine.setFallbackPrice(address(collateralToken), 70 * 1e8);

        // Non-KYC liquidator attempts liquidation of security token collateral -> reverts
        vm.prank(nonKycLiquidator);
        vm.expectRevert(
            abi.encodeWithSelector(
                ILiquidationEngine.LiquidatorNotKYCVerified.selector,
                nonKycLiquidator
            )
        );
        engine.liquidatePosition(
            alice,
            address(collateralToken),
            address(debtToken),
            750 ether
        );
    }
}
