// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {CommitRevealSettlement} from "../../src/settlement/CommitRevealSettlement.sol";
import {ICommitRevealSettlement} from "../../src/interfaces/settlement/ICommitRevealSettlement.sol";

contract MockDexToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract CommitRevealSettlementTest is Test {
    CommitRevealSettlement public implementation;
    CommitRevealSettlement public settlement;

    MockDexToken public bondToken;
    MockDexToken public baseToken;
    MockDexToken public quoteToken;

    address public admin = address(0xAD01);
    address public settlementOperator = address(0x5001);
    address public treasury = address(0x7801);

    address public buyer = address(0x1001);
    address public seller = address(0x1002);
    address public keeper = address(0x1003);

    uint256 public minCommitBond = 100e18;
    uint256 public minBlockDelay = 2;
    uint256 public maxBlockDelay = 50;

    function setUp() public {
        bondToken = new MockDexToken("Growww Bond Token", "BOND");
        baseToken = new MockDexToken("Digital Equity Asset", "DE-ASSET");
        quoteToken = new MockDexToken("Tokenized eINR Cash", "eINR");

        implementation = new CommitRevealSettlement();
        bytes memory initData = abi.encodeWithSelector(
            CommitRevealSettlement.initialize.selector,
            admin,
            address(bondToken),
            minCommitBond,
            minBlockDelay,
            maxBlockDelay,
            treasury
        );

        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), initData);
        settlement = CommitRevealSettlement(address(proxy));

        vm.startPrank(admin);
        settlement.grantRole(settlement.SETTLEMENT_OPERATOR_ROLE(), settlementOperator);
        vm.stopPrank();

        // Fund test users
        bondToken.mint(buyer, 10_000e18);
        bondToken.mint(seller, 10_000e18);
        baseToken.mint(seller, 1000e18);
        quoteToken.mint(buyer, 100_000e18);

        vm.prank(buyer);
        bondToken.approve(address(settlement), type(uint256).max);
        vm.prank(seller);
        bondToken.approve(address(settlement), type(uint256).max);

        vm.prank(buyer);
        quoteToken.approve(address(settlement), type(uint256).max);
        vm.prank(seller);
        baseToken.approve(address(settlement), type(uint256).max);
    }

    function test_Initialization() public {
        assertEq(settlement.bondToken(), address(bondToken));
        assertEq(settlement.minCommitBond(), minCommitBond);
        assertEq(settlement.minBlockDelay(), minBlockDelay);
        assertEq(settlement.maxBlockDelay(), maxBlockDelay);
        assertEq(settlement.treasury(), treasury);
    }

    function test_TwoPhaseCommitAndRevealLifecycle() public {
        ICommitRevealSettlement.OrderParams memory buyOrder = ICommitRevealSettlement.OrderParams({
            baseToken: address(baseToken),
            quoteToken: address(quoteToken),
            side: ICommitRevealSettlement.OrderSide.BUY,
            amount: 10e18,
            limitPrice: 100e18,
            nonce: 1,
            deadline: block.timestamp + 1 hours
        });

        bytes32 salt = keccak256("trader_secret_salt_1");
        bytes32 commitmentHash = settlement.computeCommitmentHash(buyer, buyOrder, salt);

        // Phase 1: Commit
        uint256 buyerBondBefore = bondToken.balanceOf(buyer);
        vm.prank(buyer);
        settlement.commitOrder(commitmentHash, minCommitBond);

        assertEq(bondToken.balanceOf(buyer), buyerBondBefore - minCommitBond);

        ICommitRevealSettlement.OrderCommitment memory commit = settlement.getCommitment(commitmentHash);
        assertEq(commit.trader, buyer);
        assertEq(commit.commitBond, minCommitBond);
        assertEq(uint8(commit.status), uint8(ICommitRevealSettlement.CommitmentStatus.COMMITTED));

        // Attempt reveal too early (same block)
        vm.prank(buyer);
        vm.expectRevert(abi.encodeWithSelector(ICommitRevealSettlement.RevealTooEarly.selector, block.number, block.number + minBlockDelay));
        settlement.revealOrder(buyOrder, salt);

        // Roll forward 3 blocks to enter reveal window
        vm.roll(block.number + 3);

        // Phase 2: Reveal
        vm.prank(buyer);
        bytes32 revealedHash = settlement.revealOrder(buyOrder, salt);

        assertEq(revealedHash, commitmentHash);
        // Bond refunded
        assertEq(bondToken.balanceOf(buyer), buyerBondBefore);

        ICommitRevealSettlement.RevealedOrder memory revOrder = settlement.getRevealedOrder(commitmentHash);
        assertEq(revOrder.trader, buyer);
        assertEq(revOrder.order.amount, 10e18);
        assertFalse(revOrder.isSettled);
    }

    function test_RevertOnRevealTooLate() public {
        ICommitRevealSettlement.OrderParams memory buyOrder = ICommitRevealSettlement.OrderParams({
            baseToken: address(baseToken),
            quoteToken: address(quoteToken),
            side: ICommitRevealSettlement.OrderSide.BUY,
            amount: 5e18,
            limitPrice: 100e18,
            nonce: 2,
            deadline: block.timestamp + 1 hours
        });

        bytes32 salt = keccak256("salt_late");
        bytes32 commitmentHash = settlement.computeCommitmentHash(buyer, buyOrder, salt);

        vm.prank(buyer);
        settlement.commitOrder(commitmentHash, minCommitBond);

        // Roll beyond maxBlockDelay (51 blocks)
        vm.roll(block.number + maxBlockDelay + 1);

        vm.prank(buyer);
        vm.expectRevert();
        settlement.revealOrder(buyOrder, salt);
    }

    function test_SlashExpiredUnrevealedCommitment() public {
        ICommitRevealSettlement.OrderParams memory buyOrder = ICommitRevealSettlement.OrderParams({
            baseToken: address(baseToken),
            quoteToken: address(quoteToken),
            side: ICommitRevealSettlement.OrderSide.BUY,
            amount: 5e18,
            limitPrice: 100e18,
            nonce: 3,
            deadline: block.timestamp + 1 hours
        });

        bytes32 salt = keccak256("salt_slash");
        bytes32 commitmentHash = settlement.computeCommitmentHash(buyer, buyOrder, salt);

        vm.prank(buyer);
        settlement.commitOrder(commitmentHash, minCommitBond);

        // Advance beyond maxBlockDelay
        vm.roll(block.number + maxBlockDelay + 5);

        uint256 keeperBalBefore = bondToken.balanceOf(keeper);
        uint256 treasuryBalBefore = bondToken.balanceOf(treasury);

        vm.prank(keeper);
        settlement.slashExpiredCommitment(commitmentHash);

        // 50% to keeper (50 eINR), 50% to treasury (50 eINR)
        assertEq(bondToken.balanceOf(keeper) - keeperBalBefore, 50e18);
        assertEq(bondToken.balanceOf(treasury) - treasuryBalBefore, 50e18);

        ICommitRevealSettlement.OrderCommitment memory commit = settlement.getCommitment(commitmentHash);
        assertEq(uint8(commit.status), uint8(ICommitRevealSettlement.CommitmentStatus.SLASHED));
    }

    function test_AtomicBatchOrderSettlement() public {
        // Buyer commits and reveals
        ICommitRevealSettlement.OrderParams memory buyOrder = ICommitRevealSettlement.OrderParams({
            baseToken: address(baseToken),
            quoteToken: address(quoteToken),
            side: ICommitRevealSettlement.OrderSide.BUY,
            amount: 10e18,
            limitPrice: 100e18,
            nonce: 10,
            deadline: block.timestamp + 1 hours
        });

        bytes32 buySalt = keccak256("buy_salt_batch");
        bytes32 buyHash = settlement.computeCommitmentHash(buyer, buyOrder, buySalt);

        vm.prank(buyer);
        settlement.commitOrder(buyHash, minCommitBond);

        // Seller commits and reveals
        ICommitRevealSettlement.OrderParams memory sellOrder = ICommitRevealSettlement.OrderParams({
            baseToken: address(baseToken),
            quoteToken: address(quoteToken),
            side: ICommitRevealSettlement.OrderSide.SELL,
            amount: 10e18,
            limitPrice: 95e18,
            nonce: 11,
            deadline: block.timestamp + 1 hours
        });

        bytes32 sellSalt = keccak256("sell_salt_batch");
        bytes32 sellHash = settlement.computeCommitmentHash(seller, sellOrder, sellSalt);

        vm.prank(seller);
        settlement.commitOrder(sellHash, minCommitBond);

        // Advance blocks
        vm.roll(block.number + 5);

        vm.prank(buyer);
        settlement.revealOrder(buyOrder, buySalt);

        vm.prank(seller);
        settlement.revealOrder(sellOrder, sellSalt);

        // Operator settles batch at uniform clearing price = 98 eINR
        uint256 clearingPrice = 98e18;
        bytes32[] memory buyHashes = new bytes32[](1);
        buyHashes[0] = buyHash;
        bytes32[] memory sellHashes = new bytes32[](1);
        sellHashes[0] = sellHash;

        uint256 buyerBaseBefore = baseToken.balanceOf(buyer);
        uint256 sellerQuoteBefore = quoteToken.balanceOf(seller);

        vm.prank(settlementOperator);
        settlement.batchSettleOrders(buyHashes, sellHashes, clearingPrice);

        // Buyer receives 10 base tokens
        assertEq(baseToken.balanceOf(buyer) - buyerBaseBefore, 10e18);
        // Seller receives (10 * 98) = 980 quote tokens
        assertEq(quoteToken.balanceOf(seller) - sellerQuoteBefore, 980e18);

        ICommitRevealSettlement.RevealedOrder memory settledBuy = settlement.getRevealedOrder(buyHash);
        ICommitRevealSettlement.RevealedOrder memory settledSell = settlement.getRevealedOrder(sellHash);
        assertTrue(settledBuy.isSettled);
        assertTrue(settledSell.isSettled);
    }
}
