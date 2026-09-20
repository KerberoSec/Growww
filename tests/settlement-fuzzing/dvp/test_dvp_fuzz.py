"""
Property-based Fuzz Test Suite for DvP Settlement Invariants
Tests atomic 2-phase settlement, network dropouts, rollback compensation, and double-entry solvency.
"""

import os
import sys
import asyncio
import random
import unittest
from hypothesis import given, strategies as st, settings

current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from dvp_fuzz_coordinator import (
    DvPSettlementFuzzCoordinator,
    FuzzMatchRecord,
    SettlementState,
)


class TestDvPSettlementFuzz(unittest.IsolatedAsyncioTestCase):

    async def test_concurrent_settlement_with_chaos_drops(self):
        coordinator = DvPSettlementFuzzCoordinator(exchange_fee_bps=5)
        symbols = ["TATA_MOTORS_EQ", "HDFC_BANK_EQ"]

        init_cash = {
            "ACC_BUYER_1": 100000000, # 10 Lakh INR (in paise)
            "ACC_BUYER_2": 100000000,
            "ACC_SELLER_1": 50000000,
            "ACC_SELLER_2": 50000000,
        }
        init_tokens = {
            "ACC_BUYER_1": {"TATA_MOTORS_EQ": 0, "HDFC_BANK_EQ": 0},
            "ACC_BUYER_2": {"TATA_MOTORS_EQ": 0, "HDFC_BANK_EQ": 0},
            "ACC_SELLER_1": {"TATA_MOTORS_EQ": 10000, "HDFC_BANK_EQ": 5000},
            "ACC_SELLER_2": {"TATA_MOTORS_EQ": 10000, "HDFC_BANK_EQ": 5000},
        }

        for acc in init_cash:
            coordinator.register_account(acc, init_cash[acc], init_tokens[acc])

        # Generate 200 random matches
        random.seed(12345)
        matches: list[FuzzMatchRecord] = []
        for i in range(200):
            buyer = random.choice(["ACC_BUYER_1", "ACC_BUYER_2"])
            seller = random.choice(["ACC_SELLER_1", "ACC_SELLER_2"])
            sym = random.choice(symbols)
            qty = random.randint(5, 50)
            price = random.randint(10000, 50000) # 100 to 500 INR
            matches.append(
                FuzzMatchRecord(
                    match_id=f"MATCH_{i+1}",
                    symbol=sym,
                    buyer_account_id=buyer,
                    seller_account_id=seller,
                    quantity=qty,
                    price_paise=price,
                    trade_timestamp_ns=i * 1000000,
                )
            )

        # Execute concurrently with 20% random RPC drops
        tasks = []
        for m in matches:
            drop_rpc = random.random() < 0.20
            tasks.append(coordinator.execute_settlement(m, chaos_drop_rpc=drop_rpc))

        results = await asyncio.gather(*tasks)

        settled_count = sum(1 for r in results if r == SettlementState.SETTLED)
        rolled_back_count = sum(1 for r in results if r == SettlementState.ROLLED_BACK)

        self.assertEqual(len(results), 200)
        self.assertGreater(settled_count, 0)
        self.assertGreater(rolled_back_count, 0)

        # Assert full double-entry solvency invariant
        res = await coordinator.verify_dvp_solvency_invariants(init_cash, init_tokens, symbols)
        self.assertTrue(res.is_solvent, f"Solvency Invariant Failed: {res.unresolved_discrepancies}")
        self.assertEqual(res.cash_delta_sum, 0)
        self.assertEqual(res.token_delta_sum, 0)

    async def test_duplicate_nonce_replay_rejection(self):
        coordinator = DvPSettlementFuzzCoordinator()
        init_cash = {"BUYER": 5000000, "SELLER": 0}
        init_tokens = {"BUYER": {"STOCK": 0}, "SELLER": {"STOCK": 100}}

        coordinator.register_account("BUYER", 5000000, init_tokens["BUYER"])
        coordinator.register_account("SELLER", 0, init_tokens["SELLER"])

        match = FuzzMatchRecord(
            match_id="DUP_MATCH_1",
            symbol="STOCK",
            buyer_account_id="BUYER",
            seller_account_id="SELLER",
            quantity=10,
            price_paise=10000,
            trade_timestamp_ns=1000,
        )

        res1 = await coordinator.execute_settlement(match)
        self.assertEqual(res1, SettlementState.SETTLED)

        # Attempt to replay the same match_id
        res2 = await coordinator.execute_settlement(match, chaos_nonce_race=True)
        self.assertEqual(res2, SettlementState.ROLLED_BACK)

        res = await coordinator.verify_dvp_solvency_invariants(init_cash, init_tokens, ["STOCK"])
        self.assertTrue(res.is_solvent)


if __name__ == "__main__":
    unittest.main()
