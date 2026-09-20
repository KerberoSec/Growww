"""
Property-based Fuzz Test Suite for Corporate Action Stock Split Rebase Invariants
Tests 1:2, 1:5, 1:10, 10:1, and 5:1 splits under active orders and positions, asserting valuation conservation.
"""

import os
import sys
import unittest
from hypothesis import given, strategies as st, settings

current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from rebase_fuzzer import (
    CorporateActionFuzzCoordinator,
    StockSplitEvent,
    PortfolioPosition,
    OpenOrder,
)


class TestCorporateActionsFuzz(unittest.TestCase):

    def test_forward_split_1_to_10_valuation_invariance(self):
        coordinator = CorporateActionFuzzCoordinator()
        # 1 share of 1000.00 INR (100000 paise) splits 1:10 -> 10 shares of 100.00 INR (10000 paise)
        positions = [
            PortfolioPosition(account_id="USER_1", symbol="MRF_EQ", quantity=10, average_buy_price_paise=100000),
            PortfolioPosition(account_id="USER_2", symbol="MRF_EQ", quantity=50, average_buy_price_paise=100000),
        ]
        orders = [
            OpenOrder(order_id=1, account_id="USER_1", symbol="MRF_EQ", price_paise=100000, quantity=5, side="BUY"),
            OpenOrder(order_id=2, account_id="USER_2", symbol="MRF_EQ", price_paise=100000, quantity=10, side="SELL"),
        ]

        split = StockSplitEvent(symbol="MRF_EQ", numerator=10, denominator=1, effective_timestamp_ns=1000)
        rebased_pos, rebased_ord, report = coordinator.execute_stock_split(split, positions, orders)

        self.assertEqual(report.total_pre_rebase_value_paise, 6000000) # 60 * 100000 = 60 Lakh
        self.assertEqual(report.total_post_rebase_value_paise, 6000000)
        self.assertEqual(report.valuation_delta_paise, 0)
        self.assertTrue(coordinator.verify_no_valuation_drift(report))

        # Check rebased orders
        self.assertEqual(len(rebased_ord), 2)
        self.assertEqual(rebased_ord[0].quantity, 50)
        self.assertEqual(rebased_ord[0].price_paise, 10000)

    def test_reverse_split_5_to_1(self):
        coordinator = CorporateActionFuzzCoordinator()
        positions = [
            PortfolioPosition(account_id="USER_1", symbol="PENNY_STOCK", quantity=100, average_buy_price_paise=2000), # 100 @ 20 Rs = 2000 Rs
        ]
        orders = []

        # 5 old shares become 1 new share (price becomes 100 Rs)
        split = StockSplitEvent(symbol="PENNY_STOCK", numerator=1, denominator=5, effective_timestamp_ns=1000)
        rebased_pos, _, report = coordinator.execute_stock_split(split, positions, orders)

        self.assertEqual(rebased_pos[0].quantity, 20)
        self.assertEqual(rebased_pos[0].average_buy_price_paise, 10000)
        self.assertEqual(report.total_post_rebase_value_paise, 200000)
        self.assertEqual(report.valuation_delta_paise, 0)

    @settings(max_examples=50, deadline=None)
    @given(
        qty=st.integers(min_value=10, max_value=10000),
        price_paise=st.integers(min_value=1000, max_value=100000),
        ratio=st.sampled_from([(2, 1), (5, 1), (10, 1), (1, 2), (1, 5), (1, 10)]),
    )
    def test_hypothesis_split_valuation_conservation(self, qty, price_paise, ratio):
        coordinator = CorporateActionFuzzCoordinator()
        # Round quantity to be multiple of denominator to test clean integer division
        clean_qty = (qty // ratio[1]) * ratio[1]
        if clean_qty == 0:
            clean_qty = ratio[1]

        # Round price to be multiple of numerator
        clean_price = (price_paise // ratio[0]) * ratio[0]
        if clean_price == 0:
            clean_price = ratio[0]

        positions = [
            PortfolioPosition(account_id="ACC", symbol="SYM", quantity=clean_qty, average_buy_price_paise=clean_price)
        ]
        split = StockSplitEvent(symbol="SYM", numerator=ratio[0], denominator=ratio[1], effective_timestamp_ns=1000)
        _, _, report = coordinator.execute_stock_split(split, positions, [])

        self.assertEqual(report.valuation_delta_paise, 0)
        self.assertTrue(coordinator.verify_no_valuation_drift(report))


if __name__ == "__main__":
    unittest.main()
