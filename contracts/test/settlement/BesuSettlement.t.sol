// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../../src/settlement/BesuSettlement.sol";

contract BesuSettlementTest is Test {
    BesuSettlement public settlement;

    address public admin = address(0xAA1);
    address public paymaster = address(0xBB2);
    address public unauthorized = address(0xCC3);
    address public mockAsset = address(0xDD4);
    address public mockSettlementToken = address(0xEE5);

    bytes32 public buyerCommitment = keccak256(abi.encodePacked("PAN_BUYER_001", "SALT_2026"));
    bytes32 public sellerCommitment = keccak256(abi.encodePacked("PAN_SELLER_002", "SALT_2026"));
    bytes32 public executionId = keccak256(abi.encodePacked("EXEC_001_BTC_USDT"));

    function setUp() public {
        vm.prank(admin);
        settlement = new BesuSettlement(paymaster);

        // Set compliance status for both traders
        vm.startPrank(admin);
        settlement.setComplianceStatus(buyerCommitment, true);
        settlement.setComplianceStatus(sellerCommitment, true);
        vm.stopPrank();
    }

    function test_Initialization() public view {
        assertEq(settlement.admin(), admin);
        assertEq(settlement.relayerPaymaster(), paymaster);
        assertTrue(settlement.compliantInvestors(buyerCommitment));
        assertTrue(settlement.compliantInvestors(sellerCommitment));
    }

    function test_SettleTrade_HappyPath() public {
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: executionId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: mockAsset,
            assetAmount: 1e8, // 1 BTC
            settlementToken: mockSettlementToken,
            settlementAmount: 64000e6, // 64,000 USDT
            timestamp: block.timestamp
        });

        // Paymaster relays the gasless settlement
        vm.prank(paymaster);
        bool success = settlement.settleTrade(trade);
        assertTrue(success);
        assertTrue(settlement.executedTrades(executionId));
    }

    function test_SettleTrade_RevertReplay() public {
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: executionId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: mockAsset,
            assetAmount: 1e8,
            settlementToken: mockSettlementToken,
            settlementAmount: 64000e6,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        settlement.settleTrade(trade);

        // Attempt replay
        vm.prank(paymaster);
        vm.expectRevert("Trade already settled");
        settlement.settleTrade(trade);
    }

    function test_SettleTrade_RevertNonCompliant() public {
        bytes32 rogueBuyer = keccak256("ROGUE_BUYER");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: keccak256("EXEC_ROGUE"),
            buyerCommitment: rogueBuyer,
            sellerCommitment: sellerCommitment,
            assetToken: mockAsset,
            assetAmount: 1e8,
            settlementToken: mockSettlementToken,
            settlementAmount: 64000e6,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        vm.expectRevert("Buyer non-compliant");
        settlement.settleTrade(trade);
    }

    function test_SettleTrade_RevertUnauthorizedCaller() public {
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: executionId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: mockAsset,
            assetAmount: 1e8,
            settlementToken: mockSettlementToken,
            settlementAmount: 64000e6,
            timestamp: block.timestamp
        });

        vm.prank(unauthorized);
        vm.expectRevert("Only authorized relayer");
        settlement.settleTrade(trade);
    }

    function test_BatchSettleTrades() public {
        bytes32 exec1 = keccak256("EXEC_BATCH_1");
        bytes32 exec2 = keccak256("EXEC_BATCH_2");

        BesuSettlement.TradeAllocation[] memory trades = new BesuSettlement.TradeAllocation[](2);
        trades[0] = BesuSettlement.TradeAllocation({
            executionId: exec1,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: mockAsset,
            assetAmount: 1e8,
            settlementToken: mockSettlementToken,
            settlementAmount: 64000e6,
            timestamp: block.timestamp
        });
        trades[1] = BesuSettlement.TradeAllocation({
            executionId: exec2,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: mockAsset,
            assetAmount: 2e8,
            settlementToken: mockSettlementToken,
            settlementAmount: 128000e6,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        uint256 count = settlement.batchSettleTrades(trades);
        assertEq(count, 2);
        assertTrue(settlement.executedTrades(exec1));
        assertTrue(settlement.executedTrades(exec2));
    }
}
