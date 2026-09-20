// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "../../src/settlement/PrimaryMarketBridge.sol";

contract MockSettlementToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {
        _mint(msg.sender, 100_000_000 * 1e18);
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract PrimaryMarketBridgeTest is Test {
    PrimaryMarketBridge public bridge;
    MockSettlementToken public cashToken;
    MockSettlementToken public equityToken;

    address public admin = address(0xAA);
    address public oracle = address(0xBB);
    address public relayer = address(0xCC);
    address public buyer = address(0x11);
    address public seller = address(0x22);

    bytes32 public isinReliance = keccak256("INE002A01018");

    function setUp() public {
        cashToken = new MockSettlementToken("e-Rupee Cash", "eINR");
        equityToken = new MockSettlementToken("Tokenized Reliance", "tRELIANCE");

        vm.startPrank(admin);
        bridge = new PrimaryMarketBridge(address(cashToken), address(equityToken), admin);
        bridge.setSessionOracle(oracle);
        bridge.setBrokerRelayer(relayer);
        vm.stopPrank();

        // Fund buyer with cash and approve bridge
        cashToken.mint(buyer, 1_000_000 * 100); // 10 Lakh paise
        vm.startPrank(buyer);
        cashToken.approve(address(bridge), type(uint256).max);
        vm.stopPrank();

        // Fund seller with equity and approve bridge
        equityToken.mint(seller, 1_000); // 1000 shares
        vm.startPrank(seller);
        equityToken.approve(address(bridge), type(uint256).max);
        vm.stopPrank();
    }

    function test_SessionTransitions() public {
        assertEq(uint8(bridge.currentSession()), uint8(PrimaryMarketBridge.MarketSession.CLOSED));

        vm.prank(oracle);
        bridge.setMarketSession(PrimaryMarketBridge.MarketSession.PRE_OPEN);
        assertEq(uint8(bridge.currentSession()), uint8(PrimaryMarketBridge.MarketSession.PRE_OPEN));

        vm.prank(oracle);
        bridge.setMarketSession(PrimaryMarketBridge.MarketSession.REGULAR_OPEN);
        assertEq(uint8(bridge.currentSession()), uint8(PrimaryMarketBridge.MarketSession.REGULAR_OPEN));
    }

    function test_QueueOrder_100PercentEscrowLock() public {
        vm.prank(buyer);
        bytes32 buyId = bridge.queueOrder(isinReliance, PrimaryMarketBridge.OrderSide.BUY, 10, 2500 * 100);

        (,,,,uint256 qty, uint256 price, uint256 escrowed,,) = bridge.orders(buyId);
        assertEq(qty, 10);
        assertEq(price, 2500 * 100);
        assertEq(escrowed, 10 * 2500 * 100);
        assertEq(cashToken.balanceOf(address(bridge)), 10 * 2500 * 100);
    }

    function test_InternalizedCrossing_Success() public {
        // Buyer queues 10 shares @ 2500 INR
        vm.prank(buyer);
        bytes32 buyId = bridge.queueOrder(isinReliance, PrimaryMarketBridge.OrderSide.BUY, 10, 2500 * 100);

        // Seller queues 10 shares @ 2500 INR
        vm.prank(seller);
        bytes32 sellId = bridge.queueOrder(isinReliance, PrimaryMarketBridge.OrderSide.SELL, 10, 2500 * 100);

        // Relayer executes off-market crossing
        vm.prank(relayer);
        bridge.internalizeCrossing(buyId, sellId);

        // Buyer gets equity shares
        assertEq(equityToken.balanceOf(buyer), 10);
        // Seller gets cash
        assertEq(cashToken.balanceOf(seller), 10 * 2500 * 100);
    }

    function test_CancelQueuedOrder_InstantRefund() public {
        vm.prank(buyer);
        bytes32 buyId = bridge.queueOrder(isinReliance, PrimaryMarketBridge.OrderSide.BUY, 10, 2500 * 100);

        assertEq(cashToken.balanceOf(buyer), 1_000_000 * 100 - (10 * 2500 * 100));

        vm.prank(buyer);
        bridge.cancelQueuedOrder(buyId);

        assertEq(cashToken.balanceOf(buyer), 1_000_000 * 100);
    }
}
