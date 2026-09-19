"""
Unit tests for Growww Compliance Engine (Prompts 000-049)
Testing Indian KYC/AML onboarding, PEP/sanctions screening, KYC tier limits,
FIU-IND structuring/CTR, Travel Rule, and regulatory sandbox boundaries.
"""

import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from compliance_engine import (
    ComplianceEngine,
    KYCTier,
    RiskTier,
    PEPClassification,
    SanctionsSource,
)


class TestComplianceEngine(unittest.TestCase):
    def setUp(self):
        self.engine = ComplianceEngine(secret_salt="TEST_SALT_12345")

    # -------------------------------------------------------------
    # 1. PAN Validation Tests
    # -------------------------------------------------------------
    def test_valid_pan_structures(self):
        valid_pans = [
            ("ABCPE1234F", "INDIVIDUAL"),
            ("XYZCA9876K", "COMPANY"),
            ("AAAHM5555L", "HINDU_UNDIVIDED_FAMILY"),
            ("BBEFP1111Q", "PARTNERSHIP_FIRM"),
        ]
        for pan, expected_entity in valid_pans:
            ok, entity = self.engine.validate_pan(pan)
            self.assertTrue(ok, f"PAN {pan} should be valid")
            self.assertEqual(entity, expected_entity)

    def test_invalid_pan_structures(self):
        invalid_pans = [
            "ABC1234F",    # Too short
            "ABCDE12345",  # Ends with digit
            "12345ABCDE",  # Starts with digits
            "ABCZE1234F",  # Invalid 4th character 'Z'
        ]
        for pan in invalid_pans:
            ok, _ = self.engine.validate_pan(pan)
            self.assertFalse(ok, f"PAN {pan} should be invalid")

    def test_pan_operative_and_aadhaar_linkage(self):
        # Inoperative PAN
        ok, msg = self.engine.validate_pan("ABCPE1234F", status="INOPERATIVE")
        self.assertFalse(ok)
        self.assertIn("INOPERATIVE", msg)

        # Unlinked Aadhaar
        ok, msg = self.engine.validate_pan("ABCPE1234F", aadhaar_linked=False)
        self.assertFalse(ok)
        self.assertIn("Section 139AA", msg)

    # -------------------------------------------------------------
    # 2. Aadhaar Masking & Data Vault Tests
    # -------------------------------------------------------------
    def test_aadhaar_masking_enforcement(self):
        valid_masked = ["XXXX-XXXX-1234", "********5678", "XXXXXXXX9999"]
        for m in valid_masked:
            ok, digits = self.engine.validate_aadhaar_masking(m)
            self.assertTrue(ok)
            self.assertEqual(len(digits), 4)

        raw_aadhaar = ["123456789012", "1234-5678-9012"]
        for r in raw_aadhaar:
            ok, msg = self.engine.validate_aadhaar_masking(r)
            self.assertFalse(ok)
            self.assertIn("CRITICAL COMPLIANCE VIOLATION", msg)

    # -------------------------------------------------------------
    # 3. Bank Penny Drop Matching Tests
    # -------------------------------------------------------------
    def test_bank_penny_drop_exact_and_honorifics(self):
        ok, score, _ = self.engine.validate_bank_penny_drop("HDFC0001234", "Rahul Sharma", "Rahul Sharma")
        self.assertTrue(ok)
        self.assertGreaterEqual(score, 0.85)

        # Honorific stripping
        ok, score, _ = self.engine.validate_bank_penny_drop("SBIN0001234", "Mr. Rajesh Kumar", "Rajesh Kumar")
        self.assertTrue(ok)
        self.assertGreaterEqual(score, 0.85)

        # Token reordering
        ok, score, _ = self.engine.validate_bank_penny_drop("ICIC0000001", "Verma Amit", "Amit Verma")
        self.assertTrue(ok)
        self.assertEqual(score, 1.0)

    def test_bank_penny_drop_mismatch_rejected(self):
        ok, score, msg = self.engine.validate_bank_penny_drop("HDFC0001234", "Rahul Sharma", "Sunil Gupta")
        self.assertFalse(ok)
        self.assertIn("failed", msg)

    # -------------------------------------------------------------
    # 4. Sanctions and PEP Screening Tests
    # -------------------------------------------------------------
    def test_sanctions_hit_and_freeze(self):
        res = self.engine.screen_sanctions_and_pep("Dawood Ibrahim Kaskar", "IND")
        self.assertTrue(res["sanctioned"])
        self.assertTrue(res["frozen"])
        self.assertEqual(res["source"], SanctionsSource.UN_CONSOLIDATED)

        # FATF Blacklist country
        res = self.engine.screen_sanctions_and_pep("Kim Jung", "PRK")
        self.assertTrue(res["sanctioned"])
        self.assertTrue(res["frozen"])

    def test_pep_screening_edd_trigger(self):
        res = self.engine.screen_sanctions_and_pep("Arun Kumar Ministerial", "IND")
        self.assertFalse(res["sanctioned"])
        self.assertFalse(res["frozen"])
        self.assertTrue(res["pep"])
        self.assertTrue(res["edd_required"])
        self.assertEqual(res["pep_category"], PEPClassification.DOMESTIC)

    # -------------------------------------------------------------
    # 5. KYC Tier Limit Enforcement Tests
    # -------------------------------------------------------------
    def test_tier1_limit_enforcement(self):
        # Permitted withdrawal
        ok, _ = self.engine.check_withdrawal_limit(KYCTier.TIER_1_BASIC_OTP, 20000.0, 0.0, 0.0)
        self.assertTrue(ok)

        # Daily limit exceeded (₹25k quota)
        ok, msg = self.engine.check_withdrawal_limit(KYCTier.TIER_1_BASIC_OTP, 10000.0, 20000.0, 20000.0)
        self.assertFalse(ok)
        self.assertIn("Daily limit exceeded", msg)

        # Annual limit exceeded (₹1L quota)
        ok, msg = self.engine.check_withdrawal_limit(KYCTier.TIER_1_BASIC_OTP, 5000.0, 0.0, 98000.0)
        self.assertFalse(ok)
        self.assertIn("Annual limit exceeded", msg)

    def test_tier0_unverified_blocked(self):
        ok, msg = self.engine.check_withdrawal_limit(KYCTier.TIER_0_UNVERIFIED, 100.0, 0.0, 0.0)
        self.assertFalse(ok)
        self.assertIn("Unverified user", msg)

    # -------------------------------------------------------------
    # 6. FIU-IND AML Surveillance Tests
    # -------------------------------------------------------------
    def test_structuring_evasion_str_alert(self):
        recent = [900000.0, 950000.0]
        res = self.engine.monitor_transaction_aml("USER-99", 50000.0, recent)
        self.assertTrue(res["str_triggered"])
        self.assertEqual(res["alerts"][0]["report_type"], "STR")

    def test_ctr_mandatory_alert(self):
        res = self.engine.monitor_transaction_aml("USER-99", 1200000.0, [])
        self.assertTrue(res["ctr_triggered"])
        self.assertEqual(res["alerts"][0]["report_type"], "CTR")

    def test_travel_rule_vda_enforcement(self):
        # Missing travel rule payload for > ₹50k crypto transfer
        res = self.engine.monitor_transaction_aml("USER-99", 75000.0, [], is_crypto=True, travel_rule=None)
        self.assertFalse(res["allowed"])
        self.assertFalse(res["travel_rule_compliant"])

        # Valid travel rule payload
        valid_tr = {
            "originator_name": "Vikram Malhotra",
            "originator_account_id": "ACC-123",
            "beneficiary_name": "Alice Smith",
            "beneficiary_wallet": "0x123abc",
        }
        res = self.engine.monitor_transaction_aml("USER-99", 75000.0, [], is_crypto=True, travel_rule=valid_tr)
        self.assertTrue(res["allowed"])
        self.assertTrue(res["travel_rule_compliant"])

    # -------------------------------------------------------------
    # 7. Regulatory Sandbox Bounds Tests
    # -------------------------------------------------------------
    def test_sebi_and_ifsca_sandbox_caps(self):
        # Domestic within bounds
        ok, _ = self.engine.validate_sandbox_bounds(is_domestic=True, current_users=5000, portfolio_fiat=45000.0)
        self.assertTrue(ok)

        # Domestic exceeding portfolio cap (₹50k)
        ok, msg = self.engine.validate_sandbox_bounds(is_domestic=True, current_users=5000, portfolio_fiat=60000.0)
        self.assertFalse(ok)
        self.assertIn("exceeds SEBI Sandbox", msg)

        # Foreign exceeding user cap (5k)
        ok, msg = self.engine.validate_sandbox_bounds(is_domestic=False, current_users=5000, portfolio_fiat=5000.0)
        self.assertFalse(ok)
        self.assertIn("limit (5,000) reached", msg)

    # -------------------------------------------------------------
    # 8. Full Domestic Onboarding E2E
    # -------------------------------------------------------------
    def test_full_domestic_onboarding(self):
        profile = {
            "investor_uuid": "usr-in-101",
            "pan": "ABCPS1234P",
            "aadhaar_masked": "XXXX-XXXX-4321",
            "ifsc": "HDFC0001234",
            "declared_name": "Pooja Sharma",
            "beneficiary_name": "Pooja Sharma",
            "face_liveness_score": 0.95,
            "anti_spoof_passed": True,
            "ckyc_number": "20012345678901",
            "annual_income": 1500000.0,
        }
        ok, record, msg = self.engine.onboard_domestic_investor(profile)
        self.assertTrue(ok, f"Onboarding failed: {msg}")
        self.assertEqual(record["assigned_tier"], "TIER_2_FULL_CKYC")
        self.assertEqual(record["risk_tier"], "LOW")
        self.assertTrue(record["commitment_hash"].startswith("0x"))
        self.assertEqual(len(record["commitment_hash"]), 66)


if __name__ == "__main__":
    unittest.main()
