"""
Unit test suite for tiered_kyc_engine and KYC/AML compliance workflows.
"""

import unittest
from tiered_kyc_engine import TieredKYCEngine

class TestTieredKYCEngine(unittest.TestCase):
    def setUp(self):
        self.engine = TieredKYCEngine()

    def test_tier1_instant_verification(self):
        # Valid Tier 1
        res = self.engine.verify_tier1_instant("user-01", "ABCDE1234F", True)
        self.assertTrue(res)
        limits = self.engine.get_withdrawal_limits("user-01")
        self.assertEqual(limits["tier"], 1)
        self.assertEqual(limits["annual_inr"], 100000)

        # Invalid PAN
        res_invalid = self.engine.verify_tier1_instant("user-02", "ABC", True)
        self.assertFalse(res_invalid)

    def test_tier2_ckyc_upgrade(self):
        self.engine.verify_tier1_instant("user-01", "ABCDE1234F", True)
        # Upgrade to Tier 2
        res = self.engine.upgrade_to_tier2_ckyc("user-01", "12345678901234", True)
        self.assertTrue(res)
        limits = self.engine.get_withdrawal_limits("user-01")
        self.assertEqual(limits["tier"], 2)
        self.assertEqual(limits["daily_inr"], 500000)

    def test_tier3_vcip_upgrade(self):
        self.engine.verify_tier1_instant("user-01", "ABCDE1234F", True)
        self.engine.upgrade_to_tier2_ckyc("user-01", "12345678901234", True)
        # Upgrade to Tier 3
        res = self.engine.upgrade_to_tier3_vcip("user-01", "s3://recordings/vcip-user-01.mp4", True)
        self.assertTrue(res)
        limits = self.engine.get_withdrawal_limits("user-01")
        self.assertEqual(limits["tier"], 3)
        self.assertEqual(limits["daily_inr"], 50000000)

    def test_unverified_limits(self):
        limits = self.engine.get_withdrawal_limits("user-unverified")
        self.assertEqual(limits["tier"], 0)
        self.assertEqual(limits["status"], "UNVERIFIED")

if __name__ == "__main__":
    unittest.main()
