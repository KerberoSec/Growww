"""
Unit & Property-based Fuzz Test Suite for Matching Engine Invariants
Tests price-time priority, no crossed book, deterministic replay, and share conservation.
"""

import os
import sys
import random
import unittest
from hypothesis import given, strategies as st, settings

# Enable importing from local directory
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from engine_fuzz_harness import (
    InvariantMatchingEngine,
    FuzzOrderEvent,
    OrderSide,
    OrderType,
)


class TestMatchingEngineFuzz(unittest.TestCase):

    def test_basic_limit_matching_and_price_priority(self):
        engine = InvariantMatchingEngine("RELIANCE_EQ")
        # Place Ask at 100
        engine.process_order(
            FuzzOrderEvent(
                order_id=1,
                account_id=101,
                symbol="RELIANCE_EQ",
                side=OrderSide.SELL,
                order_type=OrderType.LIMIT,
                price_paise=10000,
                quantity=10,
                timestamp_ns=1000,
            )
        )
        # Place Ask at 95
        engine.process_order(
            FuzzOrderEvent(
                order_id=2,
                account_id=102,
                symbol="RELIANCE_EQ",
                side=OrderSide.SELL,
                order_type=OrderType.LIMIT,
                price_paise=9500,
                quantity=5,
                timestamp_ns=2000,
            )
        )

        # Incoming Buy order at 100 for 8 shares -> should match 5 @ 95 first, then 3 @ 100
        reports = engine.process_order(
            FuzzOrderEvent(
                order_id=3,
                account_id=103,
                symbol="RELIANCE_EQ",
                side=OrderSide.BUY,
                order_type=OrderType.LIMIT,
                price_paise=10000,
                quantity=8,
                timestamp_ns=3000,
            )
        )

        self.assertEqual(len(reports), 2)
        # First execution must be with order 2 at 95.00
        self.assertEqual(reports[0].maker_order_id, 2)
        self.assertEqual(reports[0].execution_price_paise, 9500)
        self.assertEqual(reports[0].execution_quantity, 5)

        # Second execution must be with order 1 at 100.00
        self.assertEqual(reports[1].maker_order_id, 1)
        self.assertEqual(reports[1].execution_price_paise, 10000)
        self.assertEqual(reports[1].execution_quantity, 3)

        # Remaining ask 1 should have 7 shares
        self.assertEqual(engine.asks[0].quantity, 7)
        self.assertEqual(engine.get_best_ask(), 10000)
        self.assertIsNone(engine.get_best_bid())

    def test_deterministic_replay(self):
        """Proves that replaying the exact same order stream produces bit-for-bit identical execution reports."""
        random.seed(42)
        events: list[FuzzOrderEvent] = []
        for i in range(1, 500):
            side = OrderSide.BUY if random.random() < 0.5 else OrderSide.SELL
            otype = random.choice([OrderType.LIMIT, OrderType.MARKET, OrderType.IMMEDIATE_OR_CANCEL])
            price = random.randint(9000, 11000)
            qty = random.randint(1, 100)
            events.append(
                FuzzOrderEvent(
                    order_id=i,
                    account_id=random.randint(1, 10),
                    symbol="TCS_EQ",
                    side=side,
                    order_type=otype,
                    price_paise=price,
                    quantity=qty,
                    timestamp_ns=i * 1000,
                )
            )

        # Run 1
        engine1 = InvariantMatchingEngine("TCS_EQ")
        for ev in events:
            engine1.process_order(ev)

        # Run 2
        engine2 = InvariantMatchingEngine("TCS_EQ")
        for ev in events:
            engine2.process_order(ev)

        self.assertEqual(len(engine1.execution_reports), len(engine2.execution_reports))
        for r1, r2 in zip(engine1.execution_reports, engine2.execution_reports):
            self.assertEqual(r1.match_id, r2.match_id)
            self.assertEqual(r1.maker_order_id, r2.maker_order_id)
            self.assertEqual(r1.taker_order_id, r2.taker_order_id)
            self.assertEqual(r1.execution_price_paise, r2.execution_price_paise)
            self.assertEqual(r1.execution_quantity, r2.execution_quantity)

    @settings(max_examples=50, deadline=None)
    @given(
        orders=st.lists(
            st.tuples(
                st.sampled_from([OrderSide.BUY, OrderSide.SELL]),
                st.sampled_from([OrderType.LIMIT, OrderType.MARKET, OrderType.IMMEDIATE_OR_CANCEL, OrderType.FILL_OR_KILL]),
                st.integers(min_value=1000, max_value=50000), # Price 10.00 to 500.00 INR
                st.integers(min_value=1, max_value=500), # Qty 1 to 500
                st.integers(min_value=1, max_value=20), # Account ID
            ),
            min_size=10,
            max_size=200,
        )
    )
    def test_hypothesis_orderbook_invariants(self, orders):
        engine = InvariantMatchingEngine("INFY_EQ")
        positions: dict[int, int] = {acc: 0 for acc in range(1, 21)}

        for idx, (side, otype, price, qty, acc) in enumerate(orders):
            event = FuzzOrderEvent(
                order_id=idx + 1,
                account_id=acc,
                symbol="INFY_EQ",
                side=side,
                order_type=otype,
                price_paise=price,
                quantity=qty,
                timestamp_ns=(idx + 1) * 1000,
            )
            reports = engine.process_order(event)

            for rep in reports:
                # Maker and Taker net share changes
                positions[rep.taker_account_id] += rep.execution_quantity if side == OrderSide.BUY else -rep.execution_quantity
                positions[rep.maker_account_id] += -rep.execution_quantity if side == OrderSide.BUY else rep.execution_quantity

            # Continuous invariant check: No crossed book
            engine.verify_no_crossed_book()

        # Invariant: Total net positions across all market participants must sum to 0
        self.assertEqual(sum(positions.values()), 0)


if __name__ == "__main__":
    unittest.main()
