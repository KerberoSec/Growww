"""
Unit & Integration Tests for Matching Engine Benchmark Harness (Prompt 903).
Verifies generator distributions, order book invariants, Price-Time priority matching,
and sub-millisecond p99 latency threshold enforcement.
"""

import json
import os
import sys
import unittest
from pathlib import Path

# Ensure repository root is on sys.path
_repo_root = Path(__file__).resolve().parent.parent.parent
if str(_repo_root) not in sys.path:
    sys.path.insert(0, str(_repo_root))

from tests.benchmarks.engine.market_data_generator import (
    OrderEvent,
    OrderSide,
    OrderType,
    SyntheticMarketGenerator,
)
from tests.benchmarks.engine.matching_engine_simulator import BookOrder, OrderBook
from tests.benchmarks.engine.load_harness import MatchingEngineBenchmarkHarness


class TestMatchingEngineBenchmark(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.repo_root = Path(__file__).resolve().parent.parent.parent
        cls.k6_dir = cls.repo_root / "tests" / "benchmarks" / "k6"
        cls.reports_dir = cls.repo_root / "tests" / "benchmarks" / "reports"

    def test_synthetic_market_generator_ticks_and_types(self):
        """Verify generator respects discrete tick size 0.05 and produces diverse order types."""
        gen = SyntheticMarketGenerator(symbol="RELIANCE", base_price=2850.0, tick_size=0.05, seed=123)
        events = gen.generate_batch(500)

        self.assertEqual(len(events), 500)
        types_found = {e.order_type for e in events}
        self.assertIn(OrderType.LIMIT, types_found)
        self.assertIn(OrderType.MARKET, types_found)

        for event in events:
            if event.order_type == OrderType.LIMIT:
                # Check 5 paise tick size adherence: price * 100 % 5 == 0
                cents = round(event.price * 100)
                self.assertEqual(cents % 5, 0, f"Price {event.price} violates 0.05 tick size")
                self.assertGreater(event.quantity, 0)

    def test_order_book_price_time_priority(self):
        """Verify orders at the same price execute strictly in FIFO arrival sequence."""
        book = OrderBook(symbol="TCS")
        now_ns = 1000000000

        # Place 2 resting Sell Limit orders at same price 3500.00
        order1 = OrderEvent(101, "TCS", OrderSide.SELL, OrderType.LIMIT, 3500.00, 10.0, now_ns, "tier1_retail")
        order2 = OrderEvent(102, "TCS", OrderSide.SELL, OrderType.LIMIT, 3500.00, 15.0, now_ns + 100, "tier1_retail")
        book.process_event(order1)
        book.process_event(order2)

        # Place incoming aggressive Buy Limit order for 15.0 @ 3500.00
        taker = OrderEvent(103, "TCS", OrderSide.BUY, OrderType.LIMIT, 3500.00, 15.0, now_ns + 200, "tier2_institutional")
        matches, _ = book.process_event(taker)

        self.assertEqual(len(matches), 2)
        # First match must be Order 101 for 10 units (FIFO priority)
        self.assertEqual(matches[0].sell_order_id, 101)
        self.assertEqual(matches[0].quantity, 10.0)
        # Second match must be Order 102 for 5 units (partial fill)
        self.assertEqual(matches[1].sell_order_id, 102)
        self.assertEqual(matches[1].quantity, 5.0)

        # Invariant check
        valid, msg = book.verify_invariants()
        self.assertTrue(valid, msg)

    def test_order_book_invariants_and_cancellation(self):
        """Verify cancellation removes order and keeps book uncrossed."""
        book = OrderBook(symbol="HDFCBANK")
        now_ns = 2000000000

        # Resting Buy at 1600.00 and Sell at 1605.00
        buy = OrderEvent(201, "HDFCBANK", OrderSide.BUY, OrderType.LIMIT, 1600.00, 20.0, now_ns, "tier1_retail")
        sell = OrderEvent(202, "HDFCBANK", OrderSide.SELL, OrderType.LIMIT, 1605.00, 20.0, now_ns + 10, "tier1_retail")
        book.process_event(buy)
        book.process_event(sell)

        self.assertEqual(book.get_best_bid(), 1600.00)
        self.assertEqual(book.get_best_ask(), 1605.00)

        # Cancel Buy Order
        cancel = OrderEvent(203, "HDFCBANK", OrderSide.BUY, OrderType.CANCEL, 0.0, 0.0, now_ns + 20, "tier1_retail", target_order_id=201)
        book.process_event(cancel)

        self.assertIsNone(book.get_best_bid())
        self.assertEqual(book.get_best_ask(), 1605.00)

        valid, msg = book.verify_invariants()
        self.assertTrue(valid, msg)

    def test_benchmark_harness_performance_and_sla(self):
        """Run benchmark harness with 20,000 orders and verify sub-millisecond p99 SLA."""
        harness = MatchingEngineBenchmarkHarness(
            symbols=["RELIANCE", "TCS", "HDFCBANK"],
            target_orders=20000,
        )
        harness.warmup(2000)
        result = harness.run_benchmark()

        self.assertEqual(result.total_orders, 20000)
        self.assertGreater(result.matched_trades, 0)
        self.assertGreater(result.throughput_ops, 10000.0)

        # Prompt 903 SLA: p99 matching latency must be < 1,000 microseconds (< 1ms)
        self.assertTrue(result.sla_p99_passed, f"p99 latency failed SLA: {result.p99_latency_us} us")
        self.assertLess(result.p99_latency_us, 1000.0)
        self.assertTrue(result.invariants_passed, f"Invariants failed: {result.error_message}")

        # Test report export
        out_file = self.reports_dir / "test_run_report.json"
        harness.export_report(result, out_file)
        self.assertTrue(out_file.is_file())

        with open(out_file, "r", encoding="utf-8") as f:
            data = json.load(f)
            self.assertEqual(data["total_orders"], 20000)
            self.assertIn("p99_latency_us", data)

    def test_k6_scripts_structure_and_thresholds(self):
        """Verify k6 test scripts exist and define proper SLA thresholds."""
        expected_scripts = [
            "order_placement_burst.js",
            "market_data_fanout.js",
            "dvp_settlement_batch.js",
        ]
        for script in expected_scripts:
            path = self.k6_dir / script
            self.assertTrue(path.is_file(), f"Missing k6 script: {script}")
            with open(path, "r", encoding="utf-8") as f:
                content = f.read()
                self.assertIn("export const options", content)
                self.assertIn("thresholds", content)


if __name__ == "__main__":
    unittest.main()
