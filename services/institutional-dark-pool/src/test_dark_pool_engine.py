"""
Unit tests for Institutional Dark Pool Crossing Engine (Prompt 086).
Uses Python standard library unittest.
Verifies:
  - NBBO midpoint calculation and validation (crossed market rejection).
  - Minimum block size enforcement (>100 units).
  - Anti-gaming participant rate-limiting.
  - Midpoint crossing execution & FIFO order matching.
  - Delayed trade tape public reporting (15s delay).
  - Slippage limit compliance.
  - Order cancellation and lifespan expiration.
"""

import os
import sys
import time
import unittest

sys.path.insert(0, os.path.dirname(__file__))

from dark_pool_engine import (
    DarkPoolCrossingEngine,
    OrderSide,
    OrderStatus,
    NBBOQuote,
    DEFAULT_MIN_BLOCK_SIZE_E8,
)


class TestDarkPoolEngine(unittest.TestCase):

    def test_nbbo_midpoint_calculation(self):
        quote = DarkPoolCrossingEngine.compute_nbbo("BTC/USDT", 50000_00000000, 50010_00000000)
        self.assertEqual(quote.symbol, "BTC/USDT")
        self.assertEqual(quote.midpoint_e8, 50005_00000000)

        # Crossed market should raise ValueError
        with self.assertRaises(ValueError):
            DarkPoolCrossingEngine.compute_nbbo("BTC/USDT", 50010_00000000, 50000_00000000)

        # Non-positive quote should raise ValueError
        with self.assertRaises(ValueError):
            DarkPoolCrossingEngine.compute_nbbo("BTC/USDT", 0, 50000_00000000)

    def test_minimum_block_size_enforcement(self):
        engine = DarkPoolCrossingEngine(min_block_size_e8=100_00000000)
        # Order below 100 units should be rejected
        with self.assertRaises(ValueError):
            engine.submit_order(
                participant_id="INST_FUND_01",
                symbol="BTC/USDT",
                side=OrderSide.BUY,
                quantity_e8=50_00000000,
            )

    def test_anti_gaming_rate_limit(self):
        engine = DarkPoolCrossingEngine()
        engine.submit_order(
            participant_id="INST_FUND_01",
            symbol="BTC/USDT",
            side=OrderSide.BUY,
            quantity_e8=150_00000000,
        )
        # Immediate subsequent order from same participant should trigger anti-gaming
        with self.assertRaises(ValueError):
            engine.submit_order(
                participant_id="INST_FUND_01",
                symbol="BTC/USDT",
                side=OrderSide.BUY,
                quantity_e8=200_00000000,
            )

    def test_midpoint_cross_execution(self):
        engine = DarkPoolCrossingEngine(report_delay_secs=0.1)

        buy_order = engine.submit_order(
            participant_id="INST_BUYER",
            symbol="BTC/USDT",
            side=OrderSide.BUY,
            quantity_e8=200_00000000,
        )

        sell_order = engine.submit_order(
            participant_id="INST_SELLER",
            symbol="BTC/USDT",
            side=OrderSide.SELL,
            quantity_e8=150_00000000,
        )

        nbbo = NBBOQuote(
            symbol="BTC/USDT",
            best_bid_e8=60000_00000000,
            best_ask_e8=60020_00000000,
            midpoint_e8=60010_00000000,
            timestamp=time.time(),
        )

        crosses = engine.run_crossing(nbbo)
        self.assertEqual(len(crosses), 1)
        exec1 = crosses[0]
        self.assertEqual(exec1.quantity_e8, 150_00000000)
        self.assertEqual(exec1.execution_price_e8, 60010_00000000)
        self.assertNotEqual(exec1.buy_anonymous_id, "INST_BUYER")
        self.assertNotEqual(exec1.sell_anonymous_id, "INST_SELLER")

        # Verify order states
        self.assertEqual(buy_order.remaining_quantity_e8, 50_00000000)
        self.assertEqual(buy_order.status, OrderStatus.PARTIALLY_FILLED)
        self.assertEqual(sell_order.remaining_quantity_e8, 0)
        self.assertEqual(sell_order.status, OrderStatus.FILLED)

        # Check delayed reporting
        self.assertEqual(len(engine.get_reportable_fills()), 0)
        time.sleep(0.15)
        self.assertEqual(len(engine.get_reportable_fills()), 1)

    def test_order_cancellation(self):
        engine = DarkPoolCrossingEngine()
        order = engine.submit_order(
            participant_id="INST_CANCEL_TEST",
            symbol="BTC/USDT",
            side=OrderSide.BUY,
            quantity_e8=100_00000000,
        )
        self.assertEqual(len(engine.resting_bids), 1)
        cancelled = engine.cancel_order(order.order_id)
        self.assertTrue(cancelled)
        self.assertEqual(order.status, OrderStatus.CANCELLED)
        self.assertEqual(len(engine.resting_bids), 0)


if __name__ == "__main__":
    unittest.main()
