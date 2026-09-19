"""
Comprehensive unit tests for Unified Portfolio Holdings & Valuation Service (Prompts 041, 209).
Verifies arbitrary decimal precision, FIFO tax lot depletion, holding reservations,
on-chain 1:1 reconciliation, MTM valuation, and FIU AML surveillance.
"""

import unittest
from decimal import Decimal
from datetime import datetime, timezone, timedelta
import sys
import os

# Add src to path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "src")))

from unified_portfolio import (
    UnifiedPortfolioEngine,
    AssetClass,
    InsufficientBalanceError,
    InvalidReservationError,
    ReconciliationMismatchError,
    EQUITY_QUANTIZE,
    CRYPTO_QUANTIZE,
)


class TestUnifiedPortfolioEngine(unittest.TestCase):
    def setUp(self):
        self.engine = UnifiedPortfolioEngine()
        self.user_id = "USR-TEST-001"

    def test_decimal_precision_and_no_floating_point_drift(self):
        """Verifies exact decimal arithmetic without binary float drift (0.1 + 0.2 == 0.3)."""
        # Sequential fractional crypto buys
        self.engine.record_acquisition(self.user_id, AssetClass.VDA_CRYPTO, "BTC", "0.1", "5000000.00")
        self.engine.record_acquisition(self.user_id, AssetClass.VDA_CRYPTO, "BTC", "0.2", "5000000.00")
        
        total_btc = self.engine.get_total_quantity(self.user_id, "BTC")
        # In IEEE 754 float: 0.1 + 0.2 = 0.30000000000000004
        # In our engine: exactly Decimal('0.30000000')
        self.assertEqual(total_btc, Decimal("0.30000000"))
        self.assertEqual(str(total_btc), "0.30000000")

    def test_fifo_tax_lot_depletion_equity(self):
        """Verifies multi-lot FIFO cost basis tracking and realized P&L calculation."""
        t0 = datetime(2026, 4, 1, tzinfo=timezone.utc)
        t1 = datetime(2026, 4, 15, tzinfo=timezone.utc)
        t2 = datetime(2026, 5, 1, tzinfo=timezone.utc)

        # Lot 1: 100 shares RELIANCE @ ₹2,500 = ₹2,50,000
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "RELIANCE.EQ", "100.0", "2500.00", t0)
        # Lot 2: 50 shares RELIANCE @ ₹3,000 = ₹1,50,000
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "RELIANCE.EQ", "50.0", "3000.00", t1)

        # Sell 120 shares @ ₹3,200 = ₹3,84,000
        slices = self.engine.direct_disposal(self.user_id, "RELIANCE.EQ", "120.0", "3200.00", t2)

        self.assertEqual(len(slices), 2)

        # Slice 1: 100 shares from Lot 1
        s1 = slices[0]
        self.assertEqual(s1.quantity, Decimal("100.000000"))
        self.assertEqual(s1.cost_per_unit_inr, Decimal("2500.0000"))
        self.assertEqual(s1.cost_basis_inr, Decimal("250000.00"))
        self.assertEqual(s1.sale_proceeds_inr, Decimal("320000.00"))
        self.assertEqual(s1.realized_pnl_inr, Decimal("70000.00"))

        # Slice 2: 20 shares from Lot 2
        s2 = slices[1]
        self.assertEqual(s2.quantity, Decimal("20.000000"))
        self.assertEqual(s2.cost_per_unit_inr, Decimal("3000.0000"))
        self.assertEqual(s2.cost_basis_inr, Decimal("60000.00"))
        self.assertEqual(s2.sale_proceeds_inr, Decimal("64000.00"))
        self.assertEqual(s2.realized_pnl_inr, Decimal("4000.00"))

        # Remaining balance should be exactly 30 shares
        remaining_qty = self.engine.get_total_quantity(self.user_id, "RELIANCE.EQ")
        self.assertEqual(remaining_qty, Decimal("30.000000"))

    def test_section_115bbh_vda_taxable_gain_no_loss_setoff(self):
        """Verifies Section 115BBH invariant: losses cannot be set off for VDA assets."""
        t0 = datetime(2026, 4, 1, tzinfo=timezone.utc)
        # Buy 1 BTC @ ₹60,00,000
        self.engine.record_acquisition(self.user_id, AssetClass.VDA_CRYPTO, "BTC", "1.0", "6000000.00", t0)

        # Disposal at loss: Sell 0.5 BTC @ ₹50,00,000 (Loss: -₹5,00,000)
        slices_loss = self.engine.direct_disposal(self.user_id, "BTC", "0.5", "5000000.00")
        self.assertEqual(len(slices_loss), 1)
        self.assertEqual(slices_loss[0].realized_pnl_inr, Decimal("-500000.00"))
        # Invariant: taxable gain under 115BBH is strictly 0.00 (loss cannot be offset)
        self.assertEqual(slices_loss[0].taxable_115bbh_gain_inr, Decimal("0.00"))

        # Disposal at profit: Sell remaining 0.5 BTC @ ₹70,00,000 (Gain: +₹5,00,000)
        slices_gain = self.engine.direct_disposal(self.user_id, "BTC", "0.5", "7000000.00")
        self.assertEqual(len(slices_gain), 1)
        self.assertEqual(slices_gain[0].realized_pnl_inr, Decimal("500000.00"))
        self.assertEqual(slices_gain[0].taxable_115bbh_gain_inr, Decimal("500000.00"))

    def test_holding_reservation_lifecycle(self):
        """Verifies holding reservation, locking, cancellation, and execution commit."""
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "TCS.EQ", "100.0", "3800.00")

        # Initial state: 100 total, 0 reserved, 100 available
        self.assertEqual(self.engine.get_total_quantity(self.user_id, "TCS.EQ"), Decimal("100.000000"))
        self.assertEqual(self.engine.get_available_quantity(self.user_id, "TCS.EQ"), Decimal("100.000000"))

        # Reserve 40 shares for pending limit order
        res1 = self.engine.reserve_holding(self.user_id, "TCS.EQ", "40.0")
        self.assertEqual(self.engine.get_reserved_quantity(self.user_id, "TCS.EQ"), Decimal("40.000000"))
        self.assertEqual(self.engine.get_available_quantity(self.user_id, "TCS.EQ"), Decimal("60.000000"))

        # Over-reservation should raise InsufficientBalanceError
        with self.assertRaises(InsufficientBalanceError):
            self.engine.reserve_holding(self.user_id, "TCS.EQ", "70.0")

        # Release reservation (Order Cancelled)
        self.engine.release_holding(res1.reservation_id)
        self.assertEqual(self.engine.get_reserved_quantity(self.user_id, "TCS.EQ"), Decimal("0.000000"))
        self.assertEqual(self.engine.get_available_quantity(self.user_id, "TCS.EQ"), Decimal("100.000000"))

        # Re-reserve and commit execution (Order Filled)
        res2 = self.engine.reserve_holding(self.user_id, "TCS.EQ", "50.0")
        slices = self.engine.commit_holding(res2.reservation_id, "4000.00")

        self.assertEqual(len(slices), 1)
        self.assertEqual(slices[0].quantity, Decimal("50.000000"))
        self.assertEqual(self.engine.get_total_quantity(self.user_id, "TCS.EQ"), Decimal("50.000000"))
        self.assertEqual(self.engine.get_available_quantity(self.user_id, "TCS.EQ"), Decimal("50.000000"))
        self.assertEqual(self.engine.get_reserved_quantity(self.user_id, "TCS.EQ"), Decimal("0.000000"))

    def test_onchain_reconciliation_invariant(self):
        """Enforces Prompt 209 Invariant: sum(tax_lots) == onchain_token_balance."""
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "INFY.EQ", "25.500000", "1500.00")
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "INFY.EQ", "24.500000", "1550.00")

        # Total expected: 50.000000 shares
        # Matching on-chain balance -> returns True
        is_matched = self.engine.verify_onchain_reconciliation(self.user_id, "INFY.EQ", "50.000000")
        self.assertTrue(is_matched)

        # Divergent on-chain balance (e.g. 49.0 shares) -> raises ReconciliationMismatchError
        with self.assertRaises(ReconciliationMismatchError):
            self.engine.verify_onchain_reconciliation(self.user_id, "INFY.EQ", "49.000000")

    def test_mark_to_market_valuation_and_asset_breakdown(self):
        """Verifies portfolio valuation across Equity, Crypto, and CBDC."""
        # 50 shares of Reliance @ ₹2950
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "RELIANCE.EQ", "50.0", "2950.00")
        # 0.1 BTC @ ₹60,00,000
        self.engine.record_acquisition(self.user_id, AssetClass.VDA_CRYPTO, "BTC", "0.1", "6000000.00")
        # ₹1,00,000 Digital Rupee (CBDC)
        self.engine.record_acquisition(self.user_id, AssetClass.CBDC_INR, "eINR", "100000.0", "1.00")

        market_prices = {
            "RELIANCE.EQ": "3050.00",   # +₹5,000 gain (50 * 100) -> Val: ₹1,52,500
            "BTC": "6500000.00",         # +₹50,000 gain (0.1 * 500,000) -> Val: ₹6,50,000
            "eINR": "1.00"               # Par value -> Val: ₹1,00,000
        }

        portfolio = self.engine.get_consolidated_portfolio(self.user_id, market_prices)

        # Expected Net Worth = 1,52,500 + 6,50,000 + 1,00,000 = ₹9,02,500
        self.assertEqual(portfolio["total_net_worth_inr"], 902500.0)
        # Expected Invested = (50*2950=147500) + (0.1*6000000=600000) + 100000 = ₹8,47,500
        self.assertEqual(portfolio["total_invested_inr"], 847500.0)
        # Total Unrealized PnL = 9,02,500 - 8,47,500 = ₹55,000
        self.assertEqual(portfolio["total_unrealized_pnl_inr"], 55000.0)
        self.assertEqual(portfolio["holdings_count"], 3)

        # Asset class breakdown
        breakdown = portfolio["asset_class_breakdown"]
        self.assertIn("EQUITY", breakdown)
        self.assertIn("VDA_CRYPTO", breakdown)
        self.assertIn("CBDC_INR", breakdown)
        self.assertAlmostEqual(
            breakdown["EQUITY"]["allocation_pct"] + breakdown["VDA_CRYPTO"]["allocation_pct"] + breakdown["CBDC_INR"]["allocation_pct"],
            100.0,
            places=1
        )

    def test_fiu_aml_surveillance_alerts(self):
        """Verifies FIU-IND CTR and structuring surveillance triggers."""
        # 1. Single transfer >= ₹10 Lakhs
        self.engine.record_fiat_transfer(self.user_id, "1200000.00")
        alerts = self.engine.evaluate_fiu_aml_alerts(self.user_id)
        self.assertEqual(len(alerts), 1)
        self.assertEqual(alerts[0]["alert_type"], "FIU_CTR_THRESHOLD")

        # 2. Add structuring transactions (₹9,50,000 and ₹9,20,000)
        self.engine.record_fiat_transfer(self.user_id, "950000.00")
        self.engine.record_fiat_transfer(self.user_id, "920000.00")
        alerts2 = self.engine.evaluate_fiu_aml_alerts(self.user_id)
        alert_types = [a["alert_type"] for a in alerts2]
        self.assertIn("FIU_CTR_THRESHOLD", alert_types)
        self.assertIn("FIU_STR_STRUCTURING", alert_types)

    def test_reservation_and_disposal_edge_cases(self):
        """Tests error boundaries and validation checks."""
        # Zero or negative acquisition quantity
        with self.assertRaises(ValueError):
            self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "TCS.EQ", "0.0", "3000.00")

        # Negative cost
        with self.assertRaises(ValueError):
            self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "TCS.EQ", "10.0", "-100.00")

        # Zero reservation quantity
        self.engine.record_acquisition(self.user_id, AssetClass.EQUITY, "TCS.EQ", "10.0", "3000.00")
        with self.assertRaises(ValueError):
            self.engine.reserve_holding(self.user_id, "TCS.EQ", "0.0")

        # Invalid reservation release
        with self.assertRaises(InvalidReservationError):
            self.engine.release_holding("NON-EXISTENT-RES")

        # Invalid reservation commit
        with self.assertRaises(InvalidReservationError):
            self.engine.commit_holding("NON-EXISTENT-RES", "3200.00")

        # Direct disposal with insufficient balance
        with self.assertRaises(InsufficientBalanceError):
            self.engine.direct_disposal(self.user_id, "TCS.EQ", "50.0", "3200.00")

    def test_default_price_fallback_and_bond_asset_class(self):
        """Verifies weighted average cost fallback when market price is omitted, and bond support."""
        self.engine.record_acquisition(self.user_id, AssetClass.RWA_BOND, "GS2034.BOND", "10.0", "995.50")
        
        # Call get_consolidated_portfolio without providing price for GS2034.BOND
        portfolio = self.engine.get_consolidated_portfolio(self.user_id, {})
        
        bond_holding = next(h for h in portfolio["holdings"] if h["symbol"] == "GS2034.BOND")
        self.assertEqual(bond_holding["asset_class"], "RWA_BOND")
        self.assertEqual(bond_holding["current_market_price_inr"], 995.50)
        self.assertEqual(bond_holding["unrealized_pnl_inr"], 0.0)


if __name__ == "__main__":
    unittest.main()
