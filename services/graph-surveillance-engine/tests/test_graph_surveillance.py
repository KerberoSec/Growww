import unittest
import time
import os
import sys

# Add src to sys.path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'src')))

from graph_models import TraderNode, TradeEvent
from cycle_detector import StreamingTradeCycleDetector
from sybil_cluster_detector import SybilClusterDetector

class TestGraphSurveillance(unittest.TestCase):

    def test_circular_trading_ring_detection(self):
        detector = StreamingTradeCycleDetector(time_window_ms=10_000, min_cycle_len=3, max_cycle_len=5)
        now = int(time.time() * 1000)

        # Circular trades: Alice -> Bob -> Charlie -> Alice
        t1 = TradeEvent(trade_id="TR-1", symbol="INFY", seller_id="Alice", buyer_id="Bob", quantity=100.0, price=1500.0, timestamp_ms=now)
        t2 = TradeEvent(trade_id="TR-2", symbol="INFY", seller_id="Bob", buyer_id="Charlie", quantity=100.0, price=1502.0, timestamp_ms=now + 50)
        t3 = TradeEvent(trade_id="TR-3", symbol="INFY", seller_id="Charlie", buyer_id="Alice", quantity=100.0, price=1501.0, timestamp_ms=now + 100)

        detector.record_trade(t1)
        detector.record_trade(t2)
        detector.record_trade(t3)

        rings = detector.detect_collusion_rings("INFY")
        self.assertEqual(len(rings), 1)
        ring = rings[0]
        self.assertEqual(ring.cycle_length, 3)
        self.assertEqual(set(ring.participant_ids), {"Alice", "Bob", "Charlie"})
        self.assertEqual(ring.total_volume, 300.0)
        self.assertGreaterEqual(ring.confidence_score, 0.8)

    def test_no_cycle_in_linear_trading(self):
        detector = StreamingTradeCycleDetector(time_window_ms=10_000, min_cycle_len=3)
        now = int(time.time() * 1000)

        # Linear trading: Alice -> Bob -> Charlie -> David
        t1 = TradeEvent(trade_id="TR-1", symbol="TCS", seller_id="Alice", buyer_id="Bob", quantity=50.0, price=3800.0, timestamp_ms=now)
        t2 = TradeEvent(trade_id="TR-2", symbol="TCS", seller_id="Bob", buyer_id="Charlie", quantity=50.0, price=3805.0, timestamp_ms=now + 50)
        t3 = TradeEvent(trade_id="TR-3", symbol="TCS", seller_id="Charlie", buyer_id="David", quantity=50.0, price=3810.0, timestamp_ms=now + 100)

        detector.record_trade(t1)
        detector.record_trade(t2)
        detector.record_trade(t3)

        rings = detector.detect_collusion_rings("TCS")
        self.assertEqual(len(rings), 0)

    def test_sybil_clustering_by_shared_device_and_bank(self):
        detector = SybilClusterDetector(min_cluster_size=2)

        trader_a = TraderNode(trader_id="USR-101", device_uuid="DEV-IPHONE-ABC", bank_account_hash="BANK-HDFC-999")
        trader_b = TraderNode(trader_id="USR-102", device_uuid="DEV-IPHONE-ABC", bank_account_hash="BANK-ICICI-111")
        trader_c = TraderNode(trader_id="USR-103", device_uuid="DEV-ANDROID-XYZ", bank_account_hash="BANK-HDFC-999")
        trader_d = TraderNode(trader_id="USR-104", device_uuid="DEV-PIXEL-777", bank_account_hash="BANK-SBI-555")

        detector.register_trader(trader_a)
        detector.register_trader(trader_b)
        detector.register_trader(trader_c)
        detector.register_trader(trader_d)

        clusters = detector.detect_clusters()
        self.assertEqual(len(clusters), 2)

        # Check shared device cluster
        dev_clusters = [c for c in clusters if c.shared_attribute_type == "DEVICE_UUID"]
        self.assertEqual(len(dev_clusters), 1)
        self.assertEqual(dev_clusters[0].trader_ids, ["USR-101", "USR-102"])

        # Check shared bank account cluster
        bank_clusters = [c for c in clusters if c.shared_attribute_type == "BANK_ACCOUNT_HASH"]
        self.assertEqual(len(bank_clusters), 1)
        self.assertEqual(bank_clusters[0].trader_ids, ["USR-101", "USR-103"])

if __name__ == '__main__':
    unittest.main()
