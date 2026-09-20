import unittest
from perpetuals_engine import (
    PerpetualsEngine,
    PositionSide,
    ContractType,
    PositionStatus,
    FUNDING_RATE_CLAMP,
)


class TestPerpetualsEngine(unittest.TestCase):
    def setUp(self):
        self.engine = PerpetualsEngine(maintenance_margin_rate=0.005)

    def test_linear_perpetual_pnl_and_margin(self):
        # Open 2 BTC Long at $60,000 with 10x leverage
        # Required margin = (2 * 60,000) / 10 = $12,000
        pos = self.engine.open_position(
            user_id="U1",
            symbol="BTCUSDT",
            side=PositionSide.LONG,
            size_contracts=2.0,
            entry_price=60000.0,
            leverage=10,
            contract_type=ContractType.LINEAR,
        )
        self.assertEqual(pos.margin, 12000.0)

        # Mark price rises to $65,000 (+5,000 per BTC)
        # Unrealised PnL = (65,000 - 60,000) * 2 = +$10,000
        pnl = self.engine.unrealised_pnl(pos, 65000.0)
        self.assertEqual(pnl, 10000.0)

        # Short position: 2 BTC Short at $60,000
        pos_short = self.engine.open_position(
            user_id="U2",
            symbol="BTCUSDT",
            side=PositionSide.SHORT,
            size_contracts=2.0,
            entry_price=60000.0,
            leverage=10,
            contract_type=ContractType.LINEAR,
        )
        pnl_short = self.engine.unrealised_pnl(pos_short, 65000.0)
        self.assertEqual(pnl_short, -10000.0)

    def test_inverse_perpetual_coin_m(self):
        # Open $60,000 notional BTC/USD Inverse Long at $60,000 with 10x leverage
        # 1 BTC notional = $60,000. Initial margin in BTC = (60,000 / 60,000) / 10 = 0.1 BTC
        pos = self.engine.open_position(
            user_id="U3",
            symbol="BTCUSD_INV",
            side=PositionSide.LONG,
            size_contracts=60000.0,
            entry_price=60000.0,
            leverage=10,
            contract_type=ContractType.INVERSE,
        )
        self.assertAlmostEqual(pos.margin, 0.1, places=4)

        # Mark price rises to $120,000
        # PnL in BTC = 60,000 * (1/60,000 - 1/120,000) = 1.0 - 0.5 = +0.5 BTC
        pnl = self.engine.unrealised_pnl(pos, 120000.0)
        self.assertAlmostEqual(pnl, 0.5, places=4)

    def test_funding_rate_calculation_and_clamp(self):
        # Normal premium: Mark 60,100, Index 60,000 -> Premium = 100 / 60,000 = +0.001667
        snapshot = self.engine.compute_funding_rate("BTCUSDT", 60100.0, 60000.0)
        self.assertTrue(-FUNDING_RATE_CLAMP <= snapshot.funding_rate <= FUNDING_RATE_CLAMP)

        # Extreme positive premium: Mark 70,000 vs Index 60,000 -> Premium = +16.6%
        extreme_snap = self.engine.compute_funding_rate("BTCUSDT", 70000.0, 60000.0)
        self.assertEqual(extreme_snap.funding_rate, FUNDING_RATE_CLAMP)  # Clamped at +0.0075 (+75 bps)

        # Extreme negative discount: Mark 50,000 vs Index 60,000
        discount_snap = self.engine.compute_funding_rate("BTCUSDT", 50000.0, 60000.0)
        self.assertEqual(discount_snap.funding_rate, -FUNDING_RATE_CLAMP)  # Clamped at -0.0075 (-75 bps)

    def test_liquidation_and_bankruptcy(self):
        # 10x Long at $60,000, MMR = 0.5% (0.005)
        # Liq = 60,000 * (1 - 0.10 + 0.005) = 60,000 * 0.905 = $54,300
        pos = self.engine.open_position(
            user_id="U4",
            symbol="BTCUSDT",
            side=PositionSide.LONG,
            size_contracts=1.0,
            entry_price=60000.0,
            leverage=10,
            contract_type=ContractType.LINEAR,
        )
        liq = self.engine.liquidation_price(pos)
        self.assertEqual(liq, 54300.0)

        bp = self.engine.bankruptcy_price(pos)
        self.assertEqual(bp, 54000.0)

        # Mark price at 55,000 -> No liquidation
        events = self.engine.check_liquidations(55000.0)
        self.assertEqual(len(events), 0)
        self.assertEqual(pos.status, PositionStatus.ACTIVE)

        # Mark price drops to 54,000 -> Liquidated!
        events = self.engine.check_liquidations(54000.0)
        self.assertEqual(len(events), 1)
        self.assertEqual(pos.status, PositionStatus.LIQUIDATED)
        self.assertEqual(events[0].bankruptcy_price, 54000.0)

    def test_adl_priority_ranking(self):
        # Position 1: 10x leverage, entry $60,000
        pos1 = self.engine.open_position("U5", "BTCUSDT", PositionSide.LONG, 1.0, 60000.0, 10)
        # Position 2: 50x leverage, entry $60,000
        pos2 = self.engine.open_position("U6", "BTCUSDT", PositionSide.LONG, 1.0, 60000.0, 50)

        # At $66,000 (+10% gain):
        # Pos 1: Margin = $6,000, Gain = $6,000 -> PnL Ratio = 1.0 -> Score = 1.0 * 10 = 10
        # Pos 2: Margin = $1,200, Gain = $6,000 -> PnL Ratio = 5.0 -> Score = 5.0 * 50 = 250
        # Pos 2 must be ranked first in ADL queue!
        queue = self.engine.build_adl_queue(66000.0, PositionSide.LONG)
        self.assertEqual(len(queue), 2)
        self.assertEqual(queue[0].position_id, pos2.position_id)
        self.assertEqual(queue[1].position_id, pos1.position_id)


if __name__ == "__main__":
    unittest.main()
