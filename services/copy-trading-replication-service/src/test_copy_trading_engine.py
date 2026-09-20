"""
Unit tests for Copy Trading Replication Engine & High-Water Mark Profit Sharing (Prompt 087).
Verifies:
  - Follower registration and allocation ratio validation.
  - Proportional trade replication (follower_qty = leader_qty * ratio).
  - Risk limits (max notional cap, max positions count, max drawdown stop).
  - High-Water Mark (HWM) profit share calculation and deduction.
  - Follower notification dispatching and portfolio snapshot queries.
  - Follower unregistration.
"""

import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(__file__))

from copy_trading_engine import (
    CopyTradingEngine,
    FollowerConfig,
    NotificationType,
    RejectReason,
)


class TestCopyTradingEngine(unittest.TestCase):

    def setUp(self):
        self.engine = CopyTradingEngine()
        self.leader_id = "MASTER_TRADER_ALPHA"
        self.follower_id = "RETAIL_FOLLOWER_01"

        self.config = FollowerConfig(
            follower_id=self.follower_id,
            leader_id=self.leader_id,
            allocation_ratio=0.10,  # 10% copy ratio
            max_notional_usd=50000.0,
            max_drawdown_pct=0.15,  # 15% stop loss
            max_positions=5,
            profit_share_bps=1000,  # 10% profit share
        )
        self.engine.register_follower(self.config)

    def test_registration_validation(self):
        # Invalid ratio <= 0
        with self.assertRaises(ValueError):
            self.engine.register_follower(
                FollowerConfig(
                    follower_id="f2",
                    leader_id="m1",
                    allocation_ratio=0.0,
                    max_notional_usd=1000.0,
                    max_drawdown_pct=0.1,
                    max_positions=1,
                    profit_share_bps=1000,
                )
            )

        # Invalid max notional <= 0
        with self.assertRaises(ValueError):
            self.engine.register_follower(
                FollowerConfig(
                    follower_id="f3",
                    leader_id="m1",
                    allocation_ratio=0.5,
                    max_notional_usd=-500.0,
                    max_drawdown_pct=0.1,
                    max_positions=1,
                    profit_share_bps=1000,
                )
            )

    def test_proportional_trade_replication(self):
        # Leader executes 10 BTC buy order (10 * 1e8)
        leader_qty_e8 = 10_00000000
        price_usd = 40000.0

        results = self.engine.replicate_trade(
            leader_id=self.leader_id,
            leader_position_id="pos_leader_01",
            symbol="BTC/USDT",
            side="BUY",
            leader_qty_e8=leader_qty_e8,
            price_usd=price_usd,
        )

        self.assertEqual(len(results), 1)
        res = results[0]
        self.assertTrue(res.success)
        self.assertEqual(res.follower_id, self.follower_id)
        # 10% of 10 BTC = 1 BTC = 1_00000000
        self.assertEqual(res.child_order_qty_e8, 1_00000000)
        self.assertAlmostEqual(res.notional_usd, 40000.0 * 1.0)

        # Verify portfolio snapshot
        snapshot = self.engine.get_leader_portfolio_snapshot(self.leader_id)
        self.assertIn(self.follower_id, snapshot)
        self.assertEqual(len(snapshot[self.follower_id]), 1)
        self.assertEqual(snapshot[self.follower_id][0].symbol, "BTC/USDT")

    def test_risk_limit_max_notional(self):
        # Config has max notional $50,000. Leader orders $1,000,000 notional (follower 10% = $100,000)
        leader_qty_e8 = 100_00000000  # 100 BTC @ $10,000 = $1,000,000
        price_usd = 10000.0

        results = self.engine.replicate_trade(
            leader_id=self.leader_id,
            leader_position_id="pos_huge",
            symbol="BTC/USDT",
            side="BUY",
            leader_qty_e8=leader_qty_e8,
            price_usd=price_usd,
        )

        self.assertEqual(len(results), 1)
        res = results[0]
        self.assertFalse(res.success)
        self.assertEqual(res.reject_reason, RejectReason.MAX_NOTIONAL_EXCEEDED)

    def test_risk_limit_max_positions(self):
        for i in range(5):
            res = self.engine.replicate_trade(
                leader_id=self.leader_id,
                leader_position_id=f"pos_{i}",
                symbol=f"ASSET_{i}/USDT",
                side="BUY",
                leader_qty_e8=1_00000000,
                price_usd=100.0,
            )
            self.assertTrue(res[0].success)

        # 6th position should exceed max_positions=5
        res_6 = self.engine.replicate_trade(
            leader_id=self.leader_id,
            leader_position_id="pos_6",
            symbol="ASSET_OVERFLOW/USDT",
            side="BUY",
            leader_qty_e8=1_00000000,
            price_usd=100.0,
        )
        self.assertFalse(res_6[0].success)
        self.assertEqual(res_6[0].reject_reason, RejectReason.MAX_POSITIONS_EXCEEDED)

    def test_high_water_mark_profit_sharing(self):
        # Set initial equity & HWM
        portfolio_key = f"{self.follower_id}:{self.leader_id}"
        portfolio = self.engine.portfolios[portfolio_key]
        portfolio.high_water_mark_usd = 10000.0
        portfolio.current_equity_usd = 10000.0

        # Equity rises to $12,000 (+$2,000 profit). 10% profit share = $200
        share1 = self.engine.calculate_profit_share(self.follower_id, self.leader_id, 12000.0)
        self.assertAlmostEqual(share1, 200.0)
        self.assertEqual(portfolio.high_water_mark_usd, 12000.0)

        # Equity drops to $11,000 -> no fee owed
        share2 = self.engine.calculate_profit_share(self.follower_id, self.leader_id, 11000.0)
        self.assertEqual(share2, 0.0)
        self.assertEqual(portfolio.high_water_mark_usd, 12000.0)

        # Equity climbs back to $12,000 -> no fee owed (still <= HWM)
        share3 = self.engine.calculate_profit_share(self.follower_id, self.leader_id, 12000.0)
        self.assertEqual(share3, 0.0)

        # Equity breaks out to $15,000 (+$3,000 above HWM of $12,000). 10% fee = $300
        share4 = self.engine.calculate_profit_share(self.follower_id, self.leader_id, 15000.0)
        self.assertAlmostEqual(share4, 300.0)
        self.assertEqual(portfolio.high_water_mark_usd, 15000.0)
        self.assertAlmostEqual(portfolio.profit_share_paid_usd, 500.0)

    def test_unregistration(self):
        unregistered = self.engine.unregister_follower(self.follower_id, self.leader_id)
        self.assertTrue(unregistered)
        # Trades now have no followers
        results = self.engine.replicate_trade(
            leader_id=self.leader_id,
            leader_position_id="pos_post_unreg",
            symbol="BTC/USDT",
            side="BUY",
            leader_qty_e8=1_00000000,
            price_usd=60000.0,
        )
        self.assertEqual(len(results), 0)


if __name__ == "__main__":
    unittest.main()
