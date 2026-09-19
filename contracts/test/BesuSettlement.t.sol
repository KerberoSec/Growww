// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/settlement/BesuSettlement.sol";

contract BesuSettlementTest is TestBase {
    BesuSettlement settlement;

    address admin = address(0xAA1);
    address paymaster = address(0xBB2);
    address unauthorized = address(0xCC3);
    address newPaymaster = address(0xDD4);

    bytes32 buyerCommitment = keccak256("buyer_subaccount_001");
    bytes32 sellerCommitment = keccak256("seller_subaccount_002");
    bytes32 nonCompliantCommitment = keccak256("blacklisted_003");

    address wBtcToken = address(0x1111);
    address wUsdtToken = address(0x2222);

    function setUp() public {
        vm.prank(admin);
        settlement = new BesuSettlement(paymaster);
    }

    /* -------------------------------------------------------------------------- */
    /*                         1. CONSTRUCTOR & INITIAL STATE                     */
    /* -------------------------------------------------------------------------- */

    function test_Constructor_Success() public view {
        assertEq(settlement.admin(), admin, "Admin should be deployer");
        assertEq(settlement.relayerPaymaster(), paymaster, "Paymaster should match constructor param");
        assertEq(settlement.PLATFORM_FEE_BPS(), 0, "Platform fee must be 0 bps at launch");
    }

    function test_Constructor_RevertZeroPaymaster() public {
        vm.expectRevert("Invalid paymaster");
        new BesuSettlement(address(0));
    }

    /* -------------------------------------------------------------------------- */
    /*                         2. ACCESS CONTROL & PAYMASTER                      */
    /* -------------------------------------------------------------------------- */

    function test_UpdatePaymaster_Success() public {
        vm.prank(admin);
        settlement.updatePaymaster(newPaymaster);
        assertEq(settlement.relayerPaymaster(), newPaymaster);
    }

    function test_UpdatePaymaster_RevertUnauthorized() public {
        vm.prank(unauthorized);
        vm.expectRevert("Only admin");
        settlement.updatePaymaster(newPaymaster);
    }

    function test_UpdatePaymaster_RevertZeroAddress() public {
        vm.prank(admin);
        vm.expectRevert("Invalid address");
        settlement.updatePaymaster(address(0));
    }

    /* -------------------------------------------------------------------------- */
    /*                         3. COMPLIANCE REGISTRY                             */
    /* -------------------------------------------------------------------------- */

    function test_SetComplianceStatus_Success() public {
        assertFalse(settlement.compliantInvestors(buyerCommitment));

        vm.prank(admin);
        settlement.setComplianceStatus(buyerCommitment, true);
        assertTrue(settlement.compliantInvestors(buyerCommitment));

        vm.prank(admin);
        settlement.setComplianceStatus(buyerCommitment, false);
        assertFalse(settlement.compliantInvestors(buyerCommitment));
    }

    function test_SetComplianceStatus_RevertUnauthorized() public {
        vm.prank(unauthorized);
        vm.expectRevert("Only admin");
        settlement.setComplianceStatus(buyerCommitment, true);
    }

    function test_BatchSetComplianceStatus_Success() public {
        bytes32[] memory commitments = new bytes32[](2);
        commitments[0] = buyerCommitment;
        commitments[1] = sellerCommitment;

        bool[] memory statuses = new bool[](2);
        statuses[0] = true;
        statuses[1] = true;

        vm.prank(admin);
        settlement.batchSetComplianceStatus(commitments, statuses);

        assertTrue(settlement.compliantInvestors(buyerCommitment));
        assertTrue(settlement.compliantInvestors(sellerCommitment));
    }

    function test_BatchSetComplianceStatus_RevertLengthMismatch() public {
        bytes32[] memory commitments = new bytes32[](2);
        commitments[0] = buyerCommitment;
        commitments[1] = sellerCommitment;

        bool[] memory statuses = new bool[](1);
        statuses[0] = true;

        vm.prank(admin);
        vm.expectRevert("Array length mismatch");
        settlement.batchSetComplianceStatus(commitments, statuses);
    }

    /* -------------------------------------------------------------------------- */
    /*                         4. ATOMIC SINGLE DVP SETTLEMENT                    */
    /* -------------------------------------------------------------------------- */

    function _setBothCompliant() internal {
        vm.startPrank(admin);
        settlement.setComplianceStatus(buyerCommitment, true);
        settlement.setComplianceStatus(sellerCommitment, true);
        vm.stopPrank();
    }

    function test_SettleTrade_SuccessByRelayer() public {
        _setBothCompliant();

        bytes32 execId = keccak256("trade_execution_001");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: execId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 100000000, // 1 BTC in sats
            settlementToken: wUsdtToken,
            settlementAmount: 65000000000, // 65,000 USDT in cents/micros
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        bool success = settlement.settleTrade(trade);

        assertTrue(success, "Settlement should return true");
        assertTrue(settlement.executedTrades(execId), "Trade should be marked executed");
    }

    function test_SettleTrade_SuccessByAdmin() public {
        _setBothCompliant();

        bytes32 execId = keccak256("trade_execution_admin");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: execId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 50000000,
            settlementToken: wUsdtToken,
            settlementAmount: 32500000000,
            timestamp: block.timestamp
        });

        vm.prank(admin);
        bool success = settlement.settleTrade(trade);
        assertTrue(success);
        assertTrue(settlement.executedTrades(execId));
    }

    function test_SettleTrade_RevertUnauthorizedCaller() public {
        _setBothCompliant();

        bytes32 execId = keccak256("trade_execution_unauth");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: execId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 100000000,
            settlementToken: wUsdtToken,
            settlementAmount: 65000000000,
            timestamp: block.timestamp
        });

        vm.prank(unauthorized);
        vm.expectRevert("Only authorized relayer");
        settlement.settleTrade(trade);
    }

    function test_SettleTrade_RevertDuplicateExecution() public {
        _setBothCompliant();

        bytes32 execId = keccak256("trade_execution_dup");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: execId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 100000000,
            settlementToken: wUsdtToken,
            settlementAmount: 65000000000,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        settlement.settleTrade(trade);

        // Replay attempt
        vm.prank(paymaster);
        vm.expectRevert("Trade already settled");
        settlement.settleTrade(trade);
    }

    function test_SettleTrade_RevertBuyerNonCompliant() public {
        vm.prank(admin);
        settlement.setComplianceStatus(sellerCommitment, true);
        // buyer is NOT compliant

        bytes32 execId = keccak256("trade_noncompliant_buyer");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: execId,
            buyerCommitment: nonCompliantCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 100000000,
            settlementToken: wUsdtToken,
            settlementAmount: 65000000000,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        vm.expectRevert("Buyer non-compliant");
        settlement.settleTrade(trade);
    }

    function test_SettleTrade_RevertSellerNonCompliant() public {
        vm.prank(admin);
        settlement.setComplianceStatus(buyerCommitment, true);
        // seller is NOT compliant

        bytes32 execId = keccak256("trade_noncompliant_seller");
        BesuSettlement.TradeAllocation memory trade = BesuSettlement.TradeAllocation({
            executionId: execId,
            buyerCommitment: buyerCommitment,
            sellerCommitment: nonCompliantCommitment,
            assetToken: wBtcToken,
            assetAmount: 100000000,
            settlementToken: wUsdtToken,
            settlementAmount: 65000000000,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        vm.expectRevert("Seller non-compliant");
        settlement.settleTrade(trade);
    }

    function test_SettleTrade_RevertZeroAmounts() public {
        _setBothCompliant();

        bytes32 execIdZeroAsset = keccak256("trade_zero_asset");
        BesuSettlement.TradeAllocation memory tradeZeroAsset = BesuSettlement.TradeAllocation({
            executionId: execIdZeroAsset,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 0,
            settlementToken: wUsdtToken,
            settlementAmount: 65000000000,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        vm.expectRevert("Invalid amounts");
        settlement.settleTrade(tradeZeroAsset);

        bytes32 execIdZeroSettlement = keccak256("trade_zero_settlement");
        BesuSettlement.TradeAllocation memory tradeZeroSettlement = BesuSettlement.TradeAllocation({
            executionId: execIdZeroSettlement,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 100000000,
            settlementToken: wUsdtToken,
            settlementAmount: 0,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        vm.expectRevert("Invalid amounts");
        settlement.settleTrade(tradeZeroSettlement);
    }

    /* -------------------------------------------------------------------------- */
    /*                         5. BATCH DVP SETTLEMENT                            */
    /* -------------------------------------------------------------------------- */

    function test_BatchSettleTrades_AllValid() public {
        _setBothCompliant();

        BesuSettlement.TradeAllocation[] memory trades = new BesuSettlement.TradeAllocation[](3);
        for (uint256 i = 0; i < 3; i++) {
            trades[i] = BesuSettlement.TradeAllocation({
                executionId: keccak256(abi.encodePacked("batch_exec_", i)),
                buyerCommitment: buyerCommitment,
                sellerCommitment: sellerCommitment,
                assetToken: wBtcToken,
                assetAmount: (i + 1) * 10000000,
                settlementToken: wUsdtToken,
                settlementAmount: (i + 1) * 6500000000,
                timestamp: block.timestamp
            });
        }

        vm.prank(paymaster);
        uint256 settledCount = settlement.batchSettleTrades(trades);
        assertEq(settledCount, 3, "All 3 trades should be settled");

        for (uint256 i = 0; i < 3; i++) {
            assertTrue(settlement.executedTrades(trades[i].executionId));
        }
    }

    function test_BatchSettleTrades_PartialFiltering() public {
        _setBothCompliant();

        BesuSettlement.TradeAllocation[] memory trades = new BesuSettlement.TradeAllocation[](4);

        // Trade 0: Valid
        trades[0] = BesuSettlement.TradeAllocation({
            executionId: keccak256("batch_valid_1"),
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 10000000,
            settlementToken: wUsdtToken,
            settlementAmount: 6500000000,
            timestamp: block.timestamp
        });

        // Trade 1: Non-compliant seller
        trades[1] = BesuSettlement.TradeAllocation({
            executionId: keccak256("batch_bad_seller"),
            buyerCommitment: buyerCommitment,
            sellerCommitment: nonCompliantCommitment,
            assetToken: wBtcToken,
            assetAmount: 20000000,
            settlementToken: wUsdtToken,
            settlementAmount: 13000000000,
            timestamp: block.timestamp
        });

        // Trade 2: Non-compliant buyer
        trades[2] = BesuSettlement.TradeAllocation({
            executionId: keccak256("batch_bad_buyer"),
            buyerCommitment: nonCompliantCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 30000000,
            settlementToken: wUsdtToken,
            settlementAmount: 19500000000,
            timestamp: block.timestamp
        });

        // Trade 3: Valid
        trades[3] = BesuSettlement.TradeAllocation({
            executionId: keccak256("batch_valid_2"),
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            assetToken: wBtcToken,
            assetAmount: 40000000,
            settlementToken: wUsdtToken,
            settlementAmount: 26000000000,
            timestamp: block.timestamp
        });

        vm.prank(paymaster);
        uint256 settledCount = settlement.batchSettleTrades(trades);

        // Only Trade 0 and Trade 3 should settle
        assertEq(settledCount, 2, "Only 2 valid trades should settle");
        assertTrue(settlement.executedTrades(trades[0].executionId));
        assertFalse(settlement.executedTrades(trades[1].executionId));
        assertFalse(settlement.executedTrades(trades[2].executionId));
        assertTrue(settlement.executedTrades(trades[3].executionId));

        // Re-executing same batch: duplicates skipped, returns 0
        vm.prank(paymaster);
        uint256 reSettledCount = settlement.batchSettleTrades(trades);
        assertEq(reSettledCount, 0, "No duplicates should be settled");
    }
}
