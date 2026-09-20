// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {SettlementGuaranteeFund} from "../../src/settlement/SettlementGuaranteeFund.sol";
import {ISettlementGuaranteeFund} from "../../src/interfaces/ISettlementGuaranteeFund.sol";

contract MockSettlementToken is ERC20 {
    constructor() ERC20("Mock e-Rupee", "meINR") {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract SettlementGuaranteeFundTest is Test {
    SettlementGuaranteeFund public implementation;
    SettlementGuaranteeFund public fund;
    MockSettlementToken public token;

    address public admin = address(0xAD01);
    address public riskCommittee = address(0xAC01);
    address public settlementOperator = address(0x5001);
    address public emergencyGuardian = address(0xEA01);
    address public settlementVault = address(0xFA01);

    address public member1Wallet = address(0x1001);
    address public member2Wallet = address(0x1002);
    address public member3Wallet = address(0x1003);
    address public member4Wallet = address(0x1004);

    bytes32 public member1Hash = keccak256(abi.encodePacked("MEMBER_001"));
    bytes32 public member2Hash = keccak256(abi.encodePacked("MEMBER_002"));
    bytes32 public member3Hash = keccak256(abi.encodePacked("MEMBER_003"));
    bytes32 public member4Hash = keccak256(abi.encodePacked("MEMBER_004"));

    function setUp() public {
        token = new MockSettlementToken();
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

        // Fund test members and CC accounts
        token.mint(member1Wallet, 1_000_000e18);
        token.mint(member2Wallet, 1_000_000e18);
        token.mint(member3Wallet, 1_000_000e18);
        token.mint(member4Wallet, 1_000_000e18);
        token.mint(admin, 10_000_000e18);

        vm.prank(member1Wallet);
        token.approve(address(fund), type(uint256).max);

        vm.prank(member2Wallet);
        token.approve(address(fund), type(uint256).max);

        vm.prank(member3Wallet);
        token.approve(address(fund), type(uint256).max);

        vm.prank(member4Wallet);
        token.approve(address(fund), type(uint256).max);

        vm.prank(admin);
        token.approve(address(fund), type(uint256).max);
    }

    function test_Initialization() public {
        assertTrue(fund.hasRole(fund.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(fund.hasRole(fund.RISK_COMMITTEE_ROLE(), riskCommittee));
        assertTrue(fund.hasRole(fund.SETTLEMENT_OPERATOR_ROLE(), settlementOperator));
        assertTrue(fund.hasRole(fund.EMERGENCY_GUARDIAN_ROLE(), emergencyGuardian));
        assertEq(fund.defaultAssessmentCapMultiplierBps(), 20000);
    }

    function test_InitializeRevertsOnZeroAddress() public {
        bytes memory initData = abi.encodeWithSelector(
            SettlementGuaranteeFund.initialize.selector,
            address(0),
            riskCommittee,
            settlementOperator,
            emergencyGuardian
        );
        vm.expectRevert(ISettlementGuaranteeFund.InvalidAddress.selector);
        new ERC1967Proxy(address(implementation), initData);
    }

    function test_MemberDepositSGFAndMargin() public {
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 10_000e18);

        vm.prank(member1Wallet);
        fund.depositMemberMargin(member1Hash, address(token), 5_000e18);

        ISettlementGuaranteeFund.MemberAllocation memory alloc = fund.getMemberAllocation(member1Hash);
        assertEq(alloc.sgfDepositBalance, 10_000e18);
        assertEq(alloc.lockedMarginBalance, 5_000e18);
        assertEq(alloc.depositToken, address(token));
        assertFalse(alloc.isDefaulted);

        (uint256 totalSgf, uint256 totalMargin, uint256 ccContrib, uint256 ccReserves, uint256 totalTracked) =
            fund.getTrackedAllocations(address(token));
        assertEq(totalMargin, 5_000e18);
        assertEq(totalSgf, 10_000e18);
        assertEq(ccContrib, 0);
        assertEq(ccReserves, 0);
        assertEq(totalTracked, 15_000e18);
        assertEq(token.balanceOf(address(fund)), 15_000e18);
    }

    function test_DepositCCContributionsAndReserves() public {
        vm.prank(admin);
        fund.depositCCContribution(address(token), 50_000e18);

        vm.prank(admin);
        fund.depositCCReserves(address(token), 25_000e18);

        (, , uint256 ccContrib, uint256 ccReserves, uint256 totalTracked) = fund.getTrackedAllocations(address(token));
        assertEq(ccContrib, 50_000e18);
        assertEq(ccReserves, 25_000e18);
        assertEq(totalTracked, 75_000e18);
        assertEq(token.balanceOf(address(fund)), 75_000e18);
    }

    function test_DeclareMemberDefault() public {
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 10_000e18);

        bytes32 defaultId = keccak256(abi.encodePacked("DEF_001"));

        vm.prank(riskCommittee);
        fund.declareMemberDefault(defaultId, member1Hash, 25_000e18);

        ISettlementGuaranteeFund.DefaultRecord memory record = fund.getDefaultRecord(defaultId);
        assertEq(record.defaultId, defaultId);
        assertEq(record.defaulterIdHash, member1Hash);
        assertEq(record.totalDefaultAmount, 25_000e18);
        assertEq(uint8(record.status), uint8(ISettlementGuaranteeFund.DefaultStatus.DECLARED));

        ISettlementGuaranteeFund.MemberAllocation memory alloc = fund.getMemberAllocation(member1Hash);
        assertTrue(alloc.isDefaulted);

        // Cannot declare default twice
        vm.prank(riskCommittee);
        vm.expectRevert(abi.encodeWithSelector(ISettlementGuaranteeFund.DefaultAlreadyDeclared.selector, defaultId));
        fund.declareMemberDefault(defaultId, member1Hash, 25_000e18);

        // Defaulter cannot deposit more
        vm.prank(member1Wallet);
        vm.expectRevert(abi.encodeWithSelector(ISettlementGuaranteeFund.MemberAlreadyDefaulted.selector, member1Hash));
        fund.depositMemberSGF(member1Hash, address(token), 1_000e18);
    }

    function test_Waterfall_FullSequentialExecution() public {
        // Setup initial pool
        // Member 1 (Defaulter): 100 margin, 200 SGF
        vm.prank(member1Wallet);
        fund.depositMemberMargin(member1Hash, address(token), 100e18);
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 200e18);

        // Solvent members 2, 3, 4: 200 SGF each (total 600 SGF)
        vm.prank(member2Wallet);
        fund.depositMemberSGF(member2Hash, address(token), 200e18);
        vm.prank(member3Wallet);
        fund.depositMemberSGF(member3Hash, address(token), 200e18);
        vm.prank(member4Wallet);
        fund.depositMemberSGF(member4Hash, address(token), 200e18);

        // Clearing Corporation: 300 skin-in-the-game, 500 emergency reserves
        vm.prank(admin);
        fund.depositCCContribution(address(token), 300e18);
        vm.prank(admin);
        fund.depositCCReserves(address(token), 500e18);

        // Declare default for Member 1 for total 1000 tokens
        bytes32 defaultId = keccak256(abi.encodePacked("DEFAULT_1000"));
        vm.prank(riskCommittee);
        fund.declareMemberDefault(defaultId, member1Hash, 1000e18);

        // Step 1: Slash Defaulter Margins (100e18 available)
        vm.prank(settlementOperator);
        fund.slashDefaulterMargins(defaultId, settlementVault);
        assertEq(token.balanceOf(settlementVault), 100e18);

        // Step 2: Slash Defaulter SGF (200e18 available)
        vm.prank(settlementOperator);
        fund.slashDefaulterSGF(defaultId, settlementVault);
        assertEq(token.balanceOf(settlementVault), 300e18);

        // Step 3: Slash CC Skin-in-the-Game (300e18 available)
        vm.prank(settlementOperator);
        fund.slashCCContribution(defaultId, settlementVault, 300e18);
        assertEq(token.balanceOf(settlementVault), 600e18);

        // Step 4: Slash Pooled SGF from solvent members (300e18 required out of 600e18 available)
        // Each of the 3 solvent members should have 100e18 deducted pro-rata
        vm.prank(settlementOperator);
        fund.slashPooledSGF(defaultId, settlementVault, 300e18);
        assertEq(token.balanceOf(settlementVault), 900e18);

        ISettlementGuaranteeFund.MemberAllocation memory alloc2 = fund.getMemberAllocation(member2Hash);
        ISettlementGuaranteeFund.MemberAllocation memory alloc3 = fund.getMemberAllocation(member3Hash);
        ISettlementGuaranteeFund.MemberAllocation memory alloc4 = fund.getMemberAllocation(member4Hash);
        assertEq(alloc2.sgfDepositBalance, 100e18);
        assertEq(alloc3.sgfDepositBalance, 100e18);
        assertEq(alloc4.sgfDepositBalance, 100e18);

        // Step 5: Slash CC Emergency Reserves (100e18 required to complete 1000e18 default)
        vm.prank(settlementOperator);
        fund.slashCCReserves(defaultId, settlementVault, 100e18);
        assertEq(token.balanceOf(settlementVault), 1000e18);

        ISettlementGuaranteeFund.DefaultRecord memory resolvedRecord = fund.getDefaultRecord(defaultId);
        assertEq(uint8(resolvedRecord.status), uint8(ISettlementGuaranteeFund.DefaultStatus.RESOLVED));
        assertEq(resolvedRecord.totalRecoveredAmount, 1000e18);

        // Invariant check: tracked allocations equal contract balance
        (, , , , uint256 expectedRemaining) = fund.getTrackedAllocations(address(token));
        assertEq(token.balanceOf(address(fund)), expectedRemaining);
    }

    function test_Waterfall_StrictSequenceEnforcement() public {
        vm.prank(member1Wallet);
        fund.depositMemberMargin(member1Hash, address(token), 100e18);
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 200e18);

        bytes32 defaultId = keccak256(abi.encodePacked("SEQ_TEST"));
        vm.prank(riskCommittee);
        fund.declareMemberDefault(defaultId, member1Hash, 500e18);

        // Attempting Tranche 2 (Defaulter SGF) before Tranche 1 (Margins) must revert
        vm.prank(settlementOperator);
        vm.expectRevert(
            abi.encodeWithSelector(
                ISettlementGuaranteeFund.InvalidTrancheSequence.selector,
                ISettlementGuaranteeFund.TrancheType.DEFAULTER_MARGINS,
                ISettlementGuaranteeFund.TrancheType.DEFAULTER_SGF
            )
        );
        fund.slashDefaulterSGF(defaultId, settlementVault);
    }

    function test_AssessmentCallAndReplenishment() public {
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 10_000e18);

        vm.prank(member2Wallet);
        fund.depositMemberSGF(member2Hash, address(token), 10_000e18);

        bytes32 defaultId = keccak256(abi.encodePacked("DEF_ASSESS"));

        vm.prank(riskCommittee);
        fund.declareMemberDefault(defaultId, member1Hash, 5_000e18);

        uint256 deadline = block.timestamp + 2 days;
        vm.prank(riskCommittee);
        fund.issueAssessmentCall(defaultId, member2Hash, 5_000e18, deadline);

        ISettlementGuaranteeFund.AssessmentCall memory call = fund.getAssessmentCall(defaultId, member2Hash);
        assertEq(call.assessedAmount, 5_000e18);
        assertEq(call.deadline, deadline);
        assertFalse(call.isFulfilled);

        // Replenish assessment call
        vm.prank(member2Wallet);
        fund.replenishAssessment(defaultId, member2Hash, 5_000e18);

        ISettlementGuaranteeFund.AssessmentCall memory replenished = fund.getAssessmentCall(defaultId, member2Hash);
        assertTrue(replenished.isFulfilled);
    }

    function test_MemberWithdrawals() public {
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 10_000e18);
        vm.prank(member1Wallet);
        fund.depositMemberMargin(member1Hash, address(token), 5_000e18);

        // Withdraw SGF
        vm.prank(member1Wallet);
        fund.withdrawMemberSGF(member1Hash, 3_000e18, member1Wallet);

        // Withdraw Margin
        vm.prank(member1Wallet);
        fund.withdrawMemberMargin(member1Hash, 2_000e18, member1Wallet);

        ISettlementGuaranteeFund.MemberAllocation memory alloc = fund.getMemberAllocation(member1Hash);
        assertEq(alloc.sgfDepositBalance, 7_000e18);
        assertEq(alloc.lockedMarginBalance, 3_000e18);
    }

    function test_EmergencyPausePreventsOperations() public {
        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 10_000e18);

        vm.prank(emergencyGuardian);
        fund.pause();

        vm.prank(member1Wallet);
        vm.expectRevert();
        fund.depositMemberSGF(member1Hash, address(token), 1_000e18);

        vm.prank(emergencyGuardian);
        fund.unpause();

        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), 1_000e18);
        ISettlementGuaranteeFund.MemberAllocation memory alloc = fund.getMemberAllocation(member1Hash);
        assertEq(alloc.sgfDepositBalance, 11_000e18);
    }

    function testFuzz_InvariantTrackedAllocationsEqualBalance(uint96 m1Sgf, uint96 m1Margin, uint96 ccContrib) public {
        vm.assume(m1Sgf > 0 && m1Sgf < 100_000e18);
        vm.assume(m1Margin > 0 && m1Margin < 100_000e18);
        vm.assume(ccContrib > 0 && ccContrib < 100_000e18);

        token.mint(member1Wallet, uint256(m1Sgf) + uint256(m1Margin));
        token.mint(admin, uint256(ccContrib));

        vm.prank(member1Wallet);
        fund.depositMemberSGF(member1Hash, address(token), uint256(m1Sgf));

        vm.prank(member1Wallet);
        fund.depositMemberMargin(member1Hash, address(token), uint256(m1Margin));

        vm.prank(admin);
        fund.depositCCContribution(address(token), uint256(ccContrib));

        (, , , , uint256 sum) = fund.getTrackedAllocations(address(token));
        assertEq(token.balanceOf(address(fund)), sum);
    }
}
