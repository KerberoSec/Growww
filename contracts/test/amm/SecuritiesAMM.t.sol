// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import {SecuritiesAMM} from "../../src/amm/SecuritiesAMM.sol";
import {ISecuritiesAMM} from "../../src/interfaces/amm/ISecuritiesAMM.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract MockToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockIdentityRegistry is IIdentityRegistry {
    mapping(address => bool) public verifiedUsers;
    mapping(address => bytes32) public identityIds;
    mapping(bytes32 => bool) public sanctionedIdentities;

    function setVerified(address user, bool status) external {
        verifiedUsers[user] = status;
    }

    function setIdentity(address user, bytes32 id) external {
        identityIds[user] = id;
    }

    function setSanctioned(bytes32 id, bool status) external {
        sanctionedIdentities[id] = status;
    }

    function isVerified(address userAddress) external view override returns (bool) {
        return verifiedUsers[userAddress];
    }

    function getIdentityId(address userAddress) external view override returns (bytes32) {
        return identityIds[userAddress];
    }

    function isSanctioned(bytes32 identityId) external view override returns (bool) {
        return sanctionedIdentities[identityId];
    }

    // Dummy implementations for remaining interface methods
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
    function contains(address) external pure override returns (bool) { return true; }
    function totalIdentities() external pure override returns (uint256) { return 0; }
}

contract SecuritiesAMMTest is Test {
    SecuritiesAMM public amm;
    MockToken public tokenA;
    MockToken public tokenB;
    MockIdentityRegistry public identityRegistry;

    address public admin = address(0xAD01);
    address public feeCollector = address(0xFEE);
    address public alice = address(0x1001);
    address public bob = address(0x1002);
    address public nonKycUser = address(0x9999);
    address public sanctionedUser = address(0x6666);

    uint256 public constant INITIAL_MINT = 1_000_000 ether;

    function setUp() public {
        tokenA = new MockToken("Digital Security Token", "DST");
        tokenB = new MockToken("Digital Rupee", "eINR");
        identityRegistry = new MockIdentityRegistry();

        // Setup verified KYC users
        identityRegistry.setVerified(admin, true);
        identityRegistry.setVerified(alice, true);
        identityRegistry.setVerified(bob, true);
        identityRegistry.setVerified(feeCollector, true);
        identityRegistry.setIdentity(alice, keccak256("alice_identity"));
        identityRegistry.setIdentity(bob, keccak256("bob_identity"));

        // Setup sanctioned user
        identityRegistry.setVerified(sanctionedUser, true);
        bytes32 sanctionedId = keccak256("sanctioned_id");
        identityRegistry.setIdentity(sanctionedUser, sanctionedId);
        identityRegistry.setSanctioned(sanctionedId, true);

        // Deploy AMM (30 bps swap fee, 5 bps protocol fee)
        vm.prank(admin);
        amm = new SecuritiesAMM(
            address(tokenA),
            address(tokenB),
            address(identityRegistry),
            address(0),
            feeCollector,
            30, // 0.30%
            5,  // 0.05%
            admin
        );

        // Distribute funds
        tokenA.mint(alice, INITIAL_MINT);
        tokenB.mint(alice, INITIAL_MINT);
        tokenA.mint(bob, INITIAL_MINT);
        tokenB.mint(bob, INITIAL_MINT);
        tokenA.mint(nonKycUser, INITIAL_MINT);
        tokenB.mint(nonKycUser, INITIAL_MINT);
        tokenA.mint(sanctionedUser, INITIAL_MINT);
        tokenB.mint(sanctionedUser, INITIAL_MINT);

        // Approvals
        vm.prank(alice);
        tokenA.approve(address(amm), type(uint256).max);
        vm.prank(alice);
        tokenB.approve(address(amm), type(uint256).max);

        vm.prank(bob);
        tokenA.approve(address(amm), type(uint256).max);
        vm.prank(bob);
        tokenB.approve(address(amm), type(uint256).max);

        vm.prank(nonKycUser);
        tokenA.approve(address(amm), type(uint256).max);
        vm.prank(nonKycUser);
        tokenB.approve(address(amm), type(uint256).max);

        vm.prank(sanctionedUser);
        tokenA.approve(address(amm), type(uint256).max);
        vm.prank(sanctionedUser);
        tokenB.approve(address(amm), type(uint256).max);
    }

    function test_Initialization() public view {
        assertTrue(amm.token0() != address(0));
        assertTrue(amm.token1() != address(0));
        assertTrue(amm.token0() < amm.token1());
        assertEq(amm.identityRegistry(), address(identityRegistry));
        assertEq(amm.feeCollector(), feeCollector);
        assertEq(amm.feeBps(), 30);
        assertEq(amm.protocolFeeBps(), 5);
        assertEq(amm.owner(), admin);
    }

    function test_RevertIdenticalTokens() public {
        vm.prank(admin);
        vm.expectRevert(ISecuritiesAMM.IdenticalTokens.selector);
        new SecuritiesAMM(
            address(tokenA),
            address(tokenA),
            address(identityRegistry),
            address(0),
            feeCollector,
            30,
            5,
            admin
        );
    }

    function test_RevertExcessiveFee() public {
        vm.prank(admin);
        vm.expectRevert();
        new SecuritiesAMM(
            address(tokenA),
            address(tokenB),
            address(identityRegistry),
            address(0),
            feeCollector,
            350, // > 300 bps
            5,
            admin
        );
    }

    function test_AddLiquidity_RevertNonKYC() public {
        vm.prank(nonKycUser);
        vm.expectRevert(abi.encodeWithSelector(ISecuritiesAMM.CallerNotCompliant.selector, nonKycUser));
        amm.addLiquidity(100 ether, 100 ether, 90 ether, 90 ether, nonKycUser, block.timestamp + 100);
    }

    function test_AddLiquidity_RevertSanctioned() public {
        vm.prank(sanctionedUser);
        vm.expectRevert(abi.encodeWithSelector(ISecuritiesAMM.CallerNotCompliant.selector, sanctionedUser));
        amm.addLiquidity(100 ether, 100 ether, 90 ether, 90 ether, sanctionedUser, block.timestamp + 100);
    }

    function test_AddLiquidity_SuccessAndSubsequentDeposit() public {
        // Alice adds initial liquidity 10,000 t0 and 20,000 t1
        vm.prank(alice);
        (uint256 a0, uint256 a1, uint256 lpShares) = amm.addLiquidity(
            10_000 ether,
            20_000 ether,
            10_000 ether,
            20_000 ether,
            alice,
            block.timestamp + 100
        );

        assertEq(a0, 10_000 ether);
        assertEq(a1, 20_000 ether);
        assertTrue(lpShares > 0);
        assertEq(amm.balanceOf(alice), lpShares);
        assertEq(amm.balanceOf(address(0x000000000000000000000000000000000000dEaD)), amm.MINIMUM_LIQUIDITY());

        (uint256 r0, uint256 r1, ) = amm.getReserves();
        assertEq(r0, 10_000 ether);
        assertEq(r1, 20_000 ether);

        // Bob adds subsequent proportional liquidity
        vm.prank(bob);
        (uint256 bobA0, uint256 bobA1, uint256 bobLp) = amm.addLiquidity(
            5_000 ether,
            10_000 ether,
            4_900 ether,
            9_800 ether,
            bob,
            block.timestamp + 100
        );

        assertEq(bobA0, 5_000 ether);
        assertEq(bobA1, 10_000 ether);
        assertTrue(bobLp > 0);
        assertEq(amm.balanceOf(bob), bobLp);
    }

    function test_RemoveLiquidity_Success() public {
        vm.prank(alice);
        (, , uint256 lpShares) = amm.addLiquidity(
            10_000 ether,
            10_000 ether,
            10_000 ether,
            10_000 ether,
            alice,
            block.timestamp + 100
        );

        uint256 halfShares = lpShares / 2;

        vm.prank(alice);
        (uint256 amount0, uint256 amount1) = amm.removeLiquidity(
            halfShares,
            4_000 ether,
            4_000 ether,
            alice,
            block.timestamp + 100
        );

        assertTrue(amount0 > 4_900 ether);
        assertTrue(amount1 > 4_900 ether);
        assertEq(amm.balanceOf(alice), lpShares - halfShares);
    }

    function test_RemoveLiquidity_RevertSlippage() public {
        vm.prank(alice);
        (, , uint256 lpShares) = amm.addLiquidity(
            10_000 ether,
            10_000 ether,
            10_000 ether,
            10_000 ether,
            alice,
            block.timestamp + 100
        );

        vm.prank(alice);
        vm.expectRevert();
        amm.removeLiquidity(
            lpShares,
            10_001 ether, // Exceeds possible return
            10_000 ether,
            alice,
            block.timestamp + 100
        );
    }

    function test_Swap_Token0ForToken1_Success() public {
        // Alice adds initial liquidity 100,000 each
        vm.prank(alice);
        amm.addLiquidity(
            100_000 ether,
            100_000 ether,
            100_000 ether,
            100_000 ether,
            alice,
            block.timestamp + 100
        );

        address t0 = amm.token0();
        address t1 = amm.token1();

        uint256 swapAmount = 1_000 ether;
        (uint256 expectedOut, ) = amm.getAmountOut(swapAmount, t0);
        uint256 bobBalBefore = ERC20(t1).balanceOf(bob);

        vm.prank(bob);
        uint256 actualOut = amm.swapExactTokensForTokens(
            swapAmount,
            expectedOut,
            t0,
            bob,
            block.timestamp + 100
        );

        assertEq(actualOut, expectedOut);
        assertEq(ERC20(t1).balanceOf(bob), bobBalBefore + actualOut);

        // Verify reserves updated
        (uint256 r0, uint256 r1, ) = amm.getReserves();
        assertEq(r0, 100_000 ether + swapAmount - (swapAmount * amm.protocolFeeBps() / 10000));
        assertEq(r1, 100_000 ether - actualOut);
    }

    function test_Swap_RevertNonKYC() public {
        vm.prank(alice);
        amm.addLiquidity(
            10_000 ether,
            10_000 ether,
            10_000 ether,
            10_000 ether,
            alice,
            block.timestamp + 100
        );

        address t0 = amm.token0();

        vm.prank(nonKycUser);
        vm.expectRevert(abi.encodeWithSelector(ISecuritiesAMM.CallerNotCompliant.selector, nonKycUser));
        amm.swapExactTokensForTokens(
            100 ether,
            10 ether,
            t0,
            nonKycUser,
            block.timestamp + 100
        );
    }

    function test_Swap_RevertSlippage() public {
        vm.prank(alice);
        amm.addLiquidity(
            10_000 ether,
            10_000 ether,
            10_000 ether,
            10_000 ether,
            alice,
            block.timestamp + 100
        );

        address t0 = amm.token0();

        vm.prank(bob);
        vm.expectRevert();
        amm.swapExactTokensForTokens(
            100 ether,
            200 ether, // Unachievable minimum
            t0,
            bob,
            block.timestamp + 100
        );
    }

    function test_Swap_RevertDeadline() public {
        vm.prank(alice);
        amm.addLiquidity(
            10_000 ether,
            10_000 ether,
            10_000 ether,
            10_000 ether,
            alice,
            block.timestamp + 100
        );

        address t0 = amm.token0();

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(ISecuritiesAMM.DeadlineExpired.selector, block.timestamp - 1, block.timestamp));
        amm.swapExactTokensForTokens(
            100 ether,
            10 ether,
            t0,
            bob,
            block.timestamp - 1
        );
    }

    function test_AdminControls_FeeRateAndPause() public {
        vm.prank(admin);
        amm.setFeeRate(50, 10);
        assertEq(amm.feeBps(), 50);
        assertEq(amm.protocolFeeBps(), 10);

        vm.prank(admin);
        amm.pause();

        vm.prank(alice);
        vm.expectRevert();
        amm.addLiquidity(100 ether, 100 ether, 90 ether, 90 ether, alice, block.timestamp + 100);

        vm.prank(admin);
        amm.unpause();
    }
}
