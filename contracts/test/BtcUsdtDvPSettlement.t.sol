// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/settlement/BtcUsdtDvPSettlement.sol";

contract BtcUsdtDvPSettlementTest is TestBase {
    BtcUsdtDvPSettlement dvp;

    address operator = address(0x1001);
    address relayer = address(0x2002);
    address attacker = address(0x6666);

    bytes32 buyerCommitment = keccak256("buyer_besu_wallet_1");
    bytes32 sellerCommitment = keccak256("seller_besu_wallet_2");

    function setUp() public {
        dvp = new BtcUsdtDvPSettlement(operator, relayer);
    }

    /* -------------------------------------------------------------------------- */
    /*                         1. CONSTRUCTOR & INITIAL STATE                     */
    /* -------------------------------------------------------------------------- */

    function test_Constructor_Success() public view {
        assertEq(dvp.exchangeOperator(), operator);
        assertEq(dvp.settlementRelayer(), relayer);
    }

    function test_Constructor_RevertZeroOperator() public {
        vm.expectRevert("Invalid addresses");
        new BtcUsdtDvPSettlement(address(0), relayer);
    }

    function test_Constructor_RevertZeroRelayer() public {
        vm.expectRevert("Invalid addresses");
        new BtcUsdtDvPSettlement(operator, address(0));
    }

    /* -------------------------------------------------------------------------- */
    /*                         2. ACCESS CONTROL                                  */
    /* -------------------------------------------------------------------------- */

    function test_ExecuteDvP_RevertUnauthorized() public {
        bytes32 tradeHash = keccak256("trade_hash_unauthorized");
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(attacker);
        vm.expectRevert("Unauthorized relayer");
        dvp.executeDvP(trade);
    }

    function test_VoidTrade_RevertUnauthorized() public {
        bytes32 tradeHash = keccak256("trade_hash_void_unauth");
        vm.prank(attacker);
        vm.expectRevert("Unauthorized relayer");
        dvp.voidTrade(tradeHash, "Malicious void attempt");
    }

    /* -------------------------------------------------------------------------- */
    /*                         3. DVP EXECUTION SUCCESS                           */
    /* -------------------------------------------------------------------------- */

    function test_ExecuteDvP_SuccessByRelayer() public {
        bytes32 tradeHash = keccak256("trade_hash_success_relayer");
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000, // 1 BTC
            usdtAmountCents: 65000000000, // 65,000 USDT
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        bool success = dvp.executeDvP(trade);

        assertTrue(success, "DvP execution must succeed");
        assertTrue(dvp.settledTrades(tradeHash), "Trade must be marked as settled");
        assertEq(dvp.accountNonces(buyerCommitment), 1, "Buyer nonce must update");
        assertEq(dvp.accountNonces(sellerCommitment), 1, "Seller nonce must update");
    }

    function test_ExecuteDvP_SuccessByOperator() public {
        bytes32 tradeHash = keccak256("trade_hash_success_operator");
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 250000000,
            usdtAmountCents: 162500000000,
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(operator);
        bool success = dvp.executeDvP(trade);

        assertTrue(success);
        assertTrue(dvp.settledTrades(tradeHash));
    }

    /* -------------------------------------------------------------------------- */
    /*                         4. NONCE & REPLAY VALIDATIONS                      */
    /* -------------------------------------------------------------------------- */

    function test_ExecuteDvP_RevertDuplicateTradeHash() public {
        bytes32 tradeHash = keccak256("trade_hash_dup");
        BtcUsdtDvPSettlement.DvPAllocation memory trade1 = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        dvp.executeDvP(trade1);

        // Attempt replay with higher nonce but same tradeHash
        BtcUsdtDvPSettlement.DvPAllocation memory trade2 = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 2,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        vm.expectRevert("Trade already settled");
        dvp.executeDvP(trade2);
    }

    function test_ExecuteDvP_RevertStaleBuyerNonce() public {
        bytes32 tradeHash1 = keccak256("trade_hash_buyer_1");
        BtcUsdtDvPSettlement.DvPAllocation memory trade1 = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash1,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 5,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        dvp.executeDvP(trade1);

        // Trade 2 with stale or equal nonce for buyer
        bytes32 tradeHash2 = keccak256("trade_hash_buyer_2");
        bytes32 otherSeller = keccak256("seller_other");
        BtcUsdtDvPSettlement.DvPAllocation memory trade2 = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash2,
            buyerCommitment: buyerCommitment,
            sellerCommitment: otherSeller,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 5, // Equal to current nonce 5 -> must revert
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        vm.expectRevert("Stale buyer nonce");
        dvp.executeDvP(trade2);
    }

    function test_ExecuteDvP_RevertStaleSellerNonce() public {
        bytes32 tradeHash1 = keccak256("trade_hash_seller_1");
        BtcUsdtDvPSettlement.DvPAllocation memory trade1 = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash1,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 10,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        dvp.executeDvP(trade1);

        // Trade 2 with seller nonce lower than 10
        bytes32 tradeHash2 = keccak256("trade_hash_seller_2");
        bytes32 otherBuyer = keccak256("buyer_other");
        BtcUsdtDvPSettlement.DvPAllocation memory trade2 = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash2,
            buyerCommitment: otherBuyer,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 50000000,
            usdtAmountCents: 32500000000,
            executionNonce: 9, // lower than 10 -> must revert
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        vm.expectRevert("Stale seller nonce");
        dvp.executeDvP(trade2);
    }

    function test_ExecuteDvP_MonotonicNonceProgression() public {
        for (uint64 nonce = 1; nonce <= 5; nonce++) {
            bytes32 tradeHash = keccak256(abi.encodePacked("trade_seq_", nonce));
            BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
                tradeHash: tradeHash,
                buyerCommitment: buyerCommitment,
                sellerCommitment: sellerCommitment,
                btcAmountSat: 10000000 * nonce,
                usdtAmountCents: 6500000000 * nonce,
                executionNonce: nonce,
                matchedTimestamp: block.timestamp
            });

            vm.prank(relayer);
            dvp.executeDvP(trade);
            assertEq(dvp.accountNonces(buyerCommitment), nonce);
            assertEq(dvp.accountNonces(sellerCommitment), nonce);
        }
    }

    /* -------------------------------------------------------------------------- */
    /*                         5. AMOUNT VALIDATIONS                              */
    /* -------------------------------------------------------------------------- */

    function test_ExecuteDvP_RevertZeroBtc() public {
        bytes32 tradeHash = keccak256("trade_zero_btc");
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 0, // 0 BTC
            usdtAmountCents: 65000000000,
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        vm.expectRevert("Invalid BTC amount");
        dvp.executeDvP(trade);
    }

    function test_ExecuteDvP_RevertZeroUsdt() public {
        bytes32 tradeHash = keccak256("trade_zero_usdt");
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 0, // 0 USDT
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        vm.expectRevert("Invalid USDT amount");
        dvp.executeDvP(trade);
    }

    /* -------------------------------------------------------------------------- */
    /*                         6. VOID TRADE LIFECYCLE                            */
    /* -------------------------------------------------------------------------- */

    function test_VoidTrade_Success() public {
        bytes32 tradeHash = keccak256("trade_to_void");

        vm.prank(relayer);
        dvp.voidTrade(tradeHash, "Counterparty bank API timeout");

        assertTrue(dvp.settledTrades(tradeHash), "Voided trade marked settled to prevent future execution");

        // Attempting to execute voided trade must revert
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        vm.expectRevert("Trade already settled");
        dvp.executeDvP(trade);
    }

    function test_VoidTrade_RevertAlreadySettled() public {
        bytes32 tradeHash = keccak256("trade_settled_then_void");
        BtcUsdtDvPSettlement.DvPAllocation memory trade = BtcUsdtDvPSettlement.DvPAllocation({
            tradeHash: tradeHash,
            buyerCommitment: buyerCommitment,
            sellerCommitment: sellerCommitment,
            btcAmountSat: 100000000,
            usdtAmountCents: 65000000000,
            executionNonce: 1,
            matchedTimestamp: block.timestamp
        });

        vm.prank(relayer);
        dvp.executeDvP(trade);

        // Cannot void already settled trade
        vm.prank(operator);
        vm.expectRevert("Cannot void already settled trade");
        dvp.voidTrade(tradeHash, "Attempted retro-void");
    }

    function test_VoidTrade_RevertDoubleVoid() public {
        bytes32 tradeHash = keccak256("trade_double_void");

        vm.prank(relayer);
        dvp.voidTrade(tradeHash, "Reason 1");

        vm.prank(relayer);
        vm.expectRevert("Cannot void already settled trade");
        dvp.voidTrade(tradeHash, "Reason 2");
    }
}
