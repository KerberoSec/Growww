"""
Property-based Fuzz Test Suite for Cross-Venue Latency Arbitrage & SOR Resilience
Tests smart order routing, quote staleness filtering, and slippage bounding under simulated network jitter.
"""

import os
import sys
import unittest
from hypothesis import given, strategies as st, settings

current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from latency_arbitrage_harness import (
    SmartOrderRouter,
    VenueQuote,
    RoutedExecutionResult,
)


class TestLatencyArbitrageFuzz(unittest.TestCase):

    def test_routes_to_cheapest_fresh_venue(self):
        sor = SmartOrderRouter(max_allowed_slippage_bps=15.0, max_quote_age_ms=100)
        current_time_ns = 1_000_000_000 # 1s

        quotes = {
            "NSE": VenueQuote("NSE", "SBIN", bid_price_paise=79900, ask_price_paise=80000, bid_size=100, ask_size=100, quote_timestamp_ns=950_000_000), # 50ms old
            "BSE": VenueQuote("BSE", "SBIN", bid_price_paise=79950, ask_price_paise=79980, bid_size=100, ask_size=100, quote_timestamp_ns=920_000_000), # 80ms old - Cheapest Ask 799.80
            "NBSE": VenueQuote("NBSE", "SBIN", bid_price_paise=79920, ask_price_paise=80010, bid_size=100, ask_size=100, quote_timestamp_ns=995_000_000), # 5ms old
            "GIFT_CITY": VenueQuote("GIFT_CITY", "SBIN", bid_price_paise=79800, ask_price_paise=79900, bid_size=100, ask_size=100, quote_timestamp_ns=800_000_000), # 200ms old (STALE!)
        }

        # BUY order should pick BSE (cheapest among fresh quotes: 799.80 vs 800.00 vs 800.10; GIFT_CITY is stale)
        res = sor.route_and_execute("ORD_1", "SBIN", "BUY", 50, quotes, current_time_ns)

        if res.status != "DROPPED_PACKET":
            self.assertEqual(res.status, "FILLED")
            self.assertEqual(res.target_venue, "BSE")
            self.assertEqual(res.intended_price_paise, 79980)

    def test_rejects_excessive_slippage_under_toxic_drift(self):
        sor = SmartOrderRouter(max_allowed_slippage_bps=10.0, max_quote_age_ms=100) # 10 bps = 0.10%
        current_time_ns = 1_000_000_000

        quotes = {
            "NBSE": VenueQuote("NBSE", "RELIANCE", bid_price_paise=299900, ask_price_paise=300000, bid_size=100, ask_size=100, quote_timestamp_ns=990_000_000)
        }

        # Inject 100 Rs (10000 paise) toxic price drift on 3000 Rs stock (33.3 bps > 10 bps max)
        res = sor.route_and_execute("ORD_2", "RELIANCE", "BUY", 10, quotes, current_time_ns, adversarial_price_drift_paise=10000)

        self.assertEqual(res.status, "REJECTED_EXCESSIVE_SLIPPAGE")
        self.assertGreater(res.slippage_bps, 10.0)

    @settings(max_examples=50, deadline=None)
    @given(
        base_price=st.integers(min_value=10000, max_value=500000),
        drift_paise=st.integers(min_value=0, max_value=500),
        quote_age_ms=st.integers(min_value=1, max_value=200),
    )
    def test_hypothesis_sor_invariants(self, base_price, drift_paise, quote_age_ms):
        sor = SmartOrderRouter(max_allowed_slippage_bps=20.0, max_quote_age_ms=100)
        current_time_ns = 1_000_000_000
        quote_time_ns = current_time_ns - (quote_age_ms * 1_000_000)

        quotes = {
            "NBSE": VenueQuote("NBSE", "STOCK", bid_price_paise=base_price, ask_price_paise=base_price + 10, bid_size=50, ask_size=50, quote_timestamp_ns=quote_time_ns)
        }

        res = sor.route_and_execute("ORD_TEST", "STOCK", "BUY", 10, quotes, current_time_ns, adversarial_price_drift_paise=drift_paise)

        if quote_age_ms > 100:
            self.assertEqual(res.status, "REJECTED_STALE_QUOTE")
        elif res.status == "FILLED":
            self.assertLessEqual(res.slippage_bps, 20.0)


if __name__ == "__main__":
    unittest.main()
