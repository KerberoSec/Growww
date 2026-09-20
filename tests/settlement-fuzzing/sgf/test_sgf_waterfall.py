"""
Property-based Fuzz Test Suite for SGF Waterfall Invariants
Tests multi-member cascading defaults, collateral haircuts, statutory floor preservation, and mutualized pool absorption.
"""

import os
import sys
import unittest
from hypothesis import given, strategies as st, settings

current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from sgf_waterfall_simulator import (
    SGFWaterfallSimulator,
    MemberCollateralState,
)


class TestSGFWaterfallFuzz(unittest.TestCase):

    def test_single_member_collateral_absorbs_fully(self):
        sim = SGFWaterfallSimulator(
            core_exchange_sgf_paise=50000000000, # 500 Cr Paise
            statutory_core_sgf_floor_paise=10000000000, # 100 Cr Paise
        )
        sim.add_member(
            MemberCollateralState(
                member_id="MEMBER_01",
                cash_margin_paise=200000000, # 2 Cr
                approved_securities_value_paise=100000000, # 1 Cr
                sgf_contribution_paise=50000000, # 50 Lakh
            )
        )

        # Deficit 1.5 Cr (Haircut 20% on securities = 80 Lakh available, total collateral = 2.8 Cr)
        dist = sim.execute_member_default(
            defaulter_id="MEMBER_01",
            settlement_deficit_paise=150000000,
            haircut_percentage=0.20,
        )

        self.assertEqual(dist.defaulter_collateral_used, 150000000)
        self.assertEqual(dist.defaulter_sgf_used, 0)
        self.assertEqual(dist.core_sgf_used, 0)
        self.assertEqual(dist.mutualized_sgf_used, 0)
        self.assertEqual(dist.uncovered_loss, 0)
        self.assertTrue(sim.verify_waterfall_invariants(dist))

    def test_multi_tier_exhaustion_up_to_mutualization(self):
        # Core SGF = 102000000 (Floor is 100000000 -> Only 2000000 available in Tier 3)
        sim = SGFWaterfallSimulator(
            core_exchange_sgf_paise=102000000,
            statutory_core_sgf_floor_paise=100000000,
        )
        sim.add_member(
            MemberCollateralState(
                member_id="DEFAULTER_A",
                cash_margin_paise=10000000, # 10 Lakh (10000000)
                approved_securities_value_paise=0,
                sgf_contribution_paise=5000000, # 5 Lakh (5000000)
            )
        )
        sim.add_member(
            MemberCollateralState(
                member_id="MEMBER_B",
                cash_margin_paise=50000000,
                approved_securities_value_paise=0,
                sgf_contribution_paise=50000000, # 50 Lakh (50000000)
            )
        )

        # Deficit = 100000000 (1 Cr Paise)
        # Tier 1 (Defaulter Collateral) = 10000000 (10 Lakh) -> Rem: 90000000
        # Tier 2 (Defaulter SGF) = 5000000 (5 Lakh) -> Rem: 85000000
        # Tier 3 (Core SGF above floor) = 2000000 (20 Thousand) -> Rem: 83000000
        # Tier 4 (Non-Defaulter SGF) = 50000000 (50 Lakh) -> Rem: 33000000
        # Uncovered loss = 33000000
        dist = sim.execute_member_default(
            defaulter_id="DEFAULTER_A",
            settlement_deficit_paise=100000000,
            haircut_percentage=0.0,
        )

        self.assertEqual(dist.defaulter_collateral_used, 10000000)
        self.assertEqual(dist.defaulter_sgf_used, 5000000)
        self.assertEqual(dist.core_sgf_used, 2000000)
        self.assertEqual(dist.mutualized_sgf_used, 50000000)
        self.assertEqual(dist.uncovered_loss, 33000000)
        self.assertTrue(sim.verify_waterfall_invariants(dist))
        self.assertEqual(sim.core_exchange_sgf_paise, 100000000) # Floor strictly protected

    @settings(max_examples=50, deadline=None)
    @given(
        num_members=st.integers(min_value=5, max_value=30),
        deficit_paise=st.integers(min_value=1000000, max_value=5000000000),
        haircut_pct=st.floats(min_value=0.0, max_value=0.5),
    )
    def test_hypothesis_sgf_invariants(self, num_members, deficit_paise, haircut_pct):
        sim = SGFWaterfallSimulator(
            core_exchange_sgf_paise=15000000000,
            statutory_core_sgf_floor_paise=10000000000,
        )
        for i in range(num_members):
            sim.add_member(
                MemberCollateralState(
                    member_id=f"CM_{i}",
                    cash_margin_paise=50000000,
                    approved_securities_value_paise=100000000,
                    sgf_contribution_paise=20000000,
                )
            )

        dist = sim.execute_member_default("CM_0", deficit_paise, haircut_percentage=haircut_pct)
        self.assertTrue(sim.verify_waterfall_invariants(dist))
        self.assertGreaterEqual(sim.core_exchange_sgf_paise, 10000000000)


if __name__ == "__main__":
    unittest.main()
