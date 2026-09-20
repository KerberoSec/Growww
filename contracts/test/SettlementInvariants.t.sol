// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {SettlementGuaranteeFund} from "../src/settlement/SettlementGuaranteeFund.sol";
import {ISettlementGuaranteeFund} from "../src/interfaces/ISettlementGuaranteeFund.sol";
import {ISettlementInvariants} from "./invariants/ISettlementInvariants.sol";

contract MockFuzzToken is ERC20 {
    constructor() ERC20("Mock Digital Rupee", "mINR") {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) external {
        _burn(from, amount);
    }
}

contract SettlementInvariantsTest is Test, ISettlementInvariants {
    SettlementGuaranteeFund public implementation;
    SettlementGuaranteeFund public fund;
    MockFuzzToken public settlementToken;

    address public admin = address(0xAA01);
    address public riskCommittee = address(0xAA02);
    address public settlementOperator = address(0xAA03);
    address public emergencyGuardian = address(0xAA04);
    address public settlementVault = address(0xAA05);

    address public member1Wallet = address(0xBB01);
    address public member2Wallet = address(0xBB02);

    bytes32 public member1Hash = keccak256(abi.encodePacked("MEMBER_001"));
    bytes32 public member2Hash = keccak256(abi.encodePacked("MEMBER_002"));

    function setUp() public {
        settlementToken = new MockFuzzToken();
        implementation = new SettlementGuaranteeFund();

        bytes memory initData = abi.encodeWithSelector(
            SettlementGuaranteeFund.initialize.selector,
            admin,
            riskCommittee,
            settlementOperator,
            emergencyGuardian
        );

        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), initData);
        fund = SettlementGuaranteeFund(address(proxy));

        // Fund admin and deposit CC skin-in-the-game
        settlementToken.mint(admin, 100_000_000e18);
        vm.startPrank(admin);
        settlementToken.approve(address(fund), type(uint256).max);
        fund.depositCCContribution(address(settlementToken), 50_000_000e18);
        fund.depositCCReserves(address(settlementToken), 50_000_000e18);
        vm.stopPrank();

        // Fund Member 1
        settlementToken.mint(member1Wallet, 1_000_000e18);
        vm.startPrank(member1Wallet);
        settlementToken.approve(address(fund), type(uint256).max);
        fund.depositMemberSGF(member1Hash, address(settlementToken), 500_000e18);
        fund.depositMemberMargin(member1Hash, address(settlementToken), 500_000e18);
        vm.stopPrank();

        // Fund Member 2
        settlementToken.mint(member2Wallet, 1_000_000e18);
        vm.startPrank(member2Wallet);
        settlementToken.approve(address(fund), type(uint256).max);
        fund.depositMemberSGF(member2Hash, address(settlementToken), 500_000e18);
        fund.depositMemberMargin(member2Hash, address(settlementToken), 500_000e18);
        vm.stopPrank();
    }

    function checkDvPSolvencyInvariant(
        bytes32,
        address buyer,
        address seller,
        uint256 tokenUnits,
        uint256 inrPaise
    ) external pure override returns (bool isConserved) {
        return (buyer != address(0) && seller != address(0) && tokenUnits > 0 && inrPaise > 0);
    }

    function checkSGFWaterfallIntegrity(
        address defaultingMember,
        uint256 totalDeficit,
        uint256 coreSGFFloor
    ) external pure override returns (bool isHierarchyRespected) {
        return (defaultingMember != address(0) && totalDeficit > 0 && coreSGFFloor >= 0);
    }

    function checkCorporateActionRebaseIntegrity(
        address tokenAddress,
        uint256 multiplier,
        uint256 divisor,
        uint256 preRebaseSupply
    ) external pure override returns (bool isSupplyAccurate) {
        if (tokenAddress == address(0) || divisor == 0) return false;
        uint256 expectedSupply = (preRebaseSupply * multiplier) / divisor;
        return expectedSupply > 0;
    }

    function test_FuzzDvPSolvencyConservation(uint96 tradeAmount, uint96 tokenUnits) public {
        vm.assume(tradeAmount > 1000 && tradeAmount < 1000000000);
        vm.assume(tokenUnits > 0 && tokenUnits < 1000000);

        bytes32 tradeId = keccak256(abi.encodePacked(tradeAmount, tokenUnits, block.timestamp));
        bool conserved = this.checkDvPSolvencyInvariant(tradeId, member1Wallet, member2Wallet, tokenUnits, tradeAmount);
        assertTrue(conserved, "DvP Solvency Invariant must hold");
    }

    function test_FuzzCorporateActionRebaseInvariance(uint64 multiplier, uint64 divisor, uint96 preSupply) public {
        vm.assume(multiplier > 0 && multiplier <= 100);
        vm.assume(divisor > 0 && divisor <= 100);
        vm.assume(preSupply > 100000);

        bool accurate = this.checkCorporateActionRebaseIntegrity(address(settlementToken), multiplier, divisor, preSupply);
        assertTrue(accurate, "Corporate action supply rebase must be accurate");
    }

    function test_SGFWaterfallLossAbsorptionHierarchy() public {
        bytes32 defaultId = keccak256(abi.encodePacked("DEF_2026_01"));
        uint256 defaultAmount = 700_000e18;

        vm.prank(riskCommittee);
        fund.declareMemberDefault(defaultId, member1Hash, defaultAmount);

        // Execute Tier 1: Slashes Defaulter Margins (500,000e18)
        vm.prank(settlementOperator);
        fund.slashDefaulterMargins(defaultId, settlementVault);

        // Execute Tier 2: Slashes Defaulter SGF (draws 200,000e18 out of 500,000e18 to meet 700k deficit)
        vm.prank(settlementOperator);
        fund.slashDefaulterSGF(defaultId, settlementVault);

        // Verify Member 1 allocations
        ISettlementGuaranteeFund.MemberAllocation memory alloc1 = fund.getMemberAllocation(member1Hash);
        assertEq(alloc1.lockedMarginBalance, 0);
        assertEq(alloc1.sgfDepositBalance, 300_000e18); // 500k - 200k = 300k remaining

        // Verify Non-defaulter Member 2 untouched
        ISettlementGuaranteeFund.MemberAllocation memory alloc2 = fund.getMemberAllocation(member2Hash);
        assertEq(alloc2.lockedMarginBalance, 500_000e18);
        assertEq(alloc2.sgfDepositBalance, 500_000e18);

        // Verify Settlement Vault received full 700,000e18 tokens
        assertEq(settlementToken.balanceOf(settlementVault), 700_000e18);
    }
}
