#!/usr/bin/env python3
"""
Comprehensive Regulatory & Invariant Verification Suite for Prompts 000-049
Growww / NBSE Sovereign Exchange - Arun Branch

Validates:
1. docs/architecture/schema/invariants.yaml against Project North Star (Prompt 000)
2. docs/compliance/traceability_matrix.yaml against Regulatory Sandbox (Prompt 003, 047)
3. Statutory KYC/AML Domestic Policy rules (Prompt 004)
4. International Foreign Investor KYC/AML Policy (Prompt 005)
5. Zero-PII Ledger Invariant & Fee Invariants (Prompt 000, 006)
6. FIU-IND AML/CFT Surveillance & Travel Rule (Prompt 049)
7. Protocol Buffers compliance schemas
"""

import os
import sys
import yaml
import unittest

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
sys.path.insert(0, os.path.join(REPO_ROOT, "services", "compliance-service", "src"))

from compliance_engine import (
    ComplianceEngine,
    KYCTier,
    RiskTier,
    SanctionsSource,
)


class TestPrompts000To049Invariants(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.invariants_path = os.path.join(REPO_ROOT, "docs", "architecture", "schema", "invariants.yaml")
        cls.traceability_path = os.path.join(REPO_ROOT, "docs", "compliance", "traceability_matrix.yaml")
        
        with open(cls.invariants_path, "r", encoding="utf-8") as f:
            cls.invariants = yaml.safe_load(f)["system_invariants"]

        with open(cls.traceability_path, "r", encoding="utf-8") as f:
            cls.framework = yaml.safe_load(f)["regulatory_framework"]

        cls.engine = ComplianceEngine(secret_salt="SOVEREIGN_SYSTEM_SALT_VERIFICATION_TEST")

    # -------------------------------------------------------------
    # 1. Project North Star (Prompt 000) Invariants Verification
    # -------------------------------------------------------------
    def test_000_1_to_1_asset_backing_invariant(self):
        backing = self.invariants["asset_backing"]
        self.assertEqual(backing["type"], "1:1_physical_custody")
        self.assertIn(backing["depository"], ["NSDL_CDSL", "NSDL", "CDSL"])
        self.assertFalse(backing["synthetic_allowed"])

    def test_000_fiat_banking_rails_invariant(self):
        fiat = self.invariants["fiat_rails"]
        self.assertIn("UPI", fiat["domestic"])
        self.assertIn("IMPS", fiat["domestic"])
        self.assertIn("NEFT", fiat["domestic"])
        self.assertIn("RTGS", fiat["domestic"])
        self.assertFalse(fiat["unregulated_stablecoins_allowed"])

    def test_000_ledger_and_zero_pii_invariant(self):
        ledger = self.invariants["ledger"]
        self.assertEqual(ledger["type"], "permissioned_consortium")
        self.assertEqual(ledger["node_engine"], "Hyperledger_Besu")
        self.assertEqual(ledger["consensus"], "QBFT")
        self.assertEqual(ledger["block_time_seconds"], 2)
        self.assertFalse(ledger["on_chain_pii_allowed"])

    def test_000_universal_zero_fee_model_invariant(self):
        monetization = self.invariants["monetization"]
        self.assertEqual(monetization["platform_fee_rate"], 0.0)
        self.assertEqual(monetization["revenue_split"]["maker_fee_bps"], 0)
        self.assertEqual(monetization["revenue_split"]["taker_fee_bps"], 0)
        self.assertEqual(monetization["holding_fee_rate"], 0.0)
        self.assertEqual(monetization["aum_fee_rate"], 0.0)

    def test_000_hsm_and_security_invariant(self):
        security = self.invariants["security"]
        self.assertEqual(security["key_custody"], "FIPS_140_2_L3_HSM")
        self.assertEqual(security["inter_service_auth"], "mTLS_SPIFFE_SPIRE")

    # -------------------------------------------------------------
    # 2. Regulatory Sandbox Pathway (Prompt 003, 047) Verification
    # -------------------------------------------------------------
    def test_003_sebi_sandbox_constraints(self):
        sebi = self.framework["authorities"]["sebi"]
        self.assertEqual(sebi["sandbox_program"], "SEBI_INNOVATION_SANDBOX")
        self.assertEqual(sebi["constraints"]["max_domestic_users"], 10000)
        self.assertEqual(sebi["constraints"]["max_portfolio_inr_per_user"], 50000.00)
        self.assertEqual(sebi["constraints"]["settlement_cycle"], "T_PLUS_ZERO_DVP")

    def test_003_rbi_innovation_hub_constraints(self):
        rbi = self.framework["authorities"]["rbi"]
        self.assertEqual(rbi["sandbox_program"], "RBI_INNOVATION_HUB")
        self.assertTrue(rbi["constraints"]["unregulated_stablecoins_prohibited"])

    def test_003_ifsca_gift_city_constraints(self):
        ifsca = self.framework["authorities"]["ifsca"]
        self.assertEqual(ifsca["sandbox_program"], "IFSCA_FINTECH_SANDBOX")
        self.assertEqual(ifsca["constraints"]["max_foreign_users"], 5000)
        self.assertEqual(ifsca["constraints"]["max_portfolio_usd_per_user"], 10000.00)
        self.assertTrue(ifsca["constraints"]["sanctions_screening_mandatory"])

    # -------------------------------------------------------------
    # 3. Protocol Buffers & Dossiers Integrity
    # -------------------------------------------------------------
    def test_required_compliance_proto_files_exist(self):
        proto_files = [
            os.path.join(REPO_ROOT, "packages", "proto", "growww", "compliance", "engine", "v1", "compliance_engine.proto"),
            os.path.join(REPO_ROOT, "packages", "proto", "growww", "compliance", "sandbox", "v1", "sandbox.proto"),
            os.path.join(REPO_ROOT, "packages", "proto", "growww", "compliance", "fiu", "v1", "fiu.proto"),
            os.path.join(REPO_ROOT, "packages", "proto", "growww", "fiu", "v1", "fiu_ind_str.proto"),
        ]
        for p in proto_files:
            self.assertTrue(os.path.isfile(p), f"Required proto file missing: {p}")

    def test_required_compliance_dossiers_exist(self):
        dossiers = [
            os.path.join(REPO_ROOT, "docs", "compliance", "kyc_aml_domestic_policy.md"),
            os.path.join(REPO_ROOT, "docs", "compliance", "kyc_aml_foreign_policy.md"),
            os.path.join(REPO_ROOT, "docs", "compliance", "regulatory_pathway.md"),
            os.path.join(REPO_ROOT, "docs", "compliance", "FIU_IND_AML_CFT_AND_PMLA_RULEBOOK.md"),
            os.path.join(REPO_ROOT, "docs", "architecture", "two_entity_structure.md"),
        ]
        for d in dossiers:
            self.assertTrue(os.path.isfile(d), f"Required dossier missing: {d}")

    # -------------------------------------------------------------
    # 4. Engine End-to-End Compliance Verification
    # -------------------------------------------------------------
    def test_engine_domestic_kyc_and_zero_pii_commitment(self):
        profile = {
            "investor_uuid": "usr-inv-2026-audit",
            "pan": "ABCPS9876P",
            "pan_status": "OPERATIVE",
            "aadhaar_linked": True,
            "aadhaar_masked": "XXXX-XXXX-8888",
            "ckyc_number": "20098765432100",
            "ifsc": "HDFC0001234",
            "declared_name": "Kavita Rao",
            "beneficiary_name": "Kavita Rao",
            "face_liveness_score": 0.98,
            "anti_spoof_passed": True,
            "annual_income": 800000.0,
        }
        ok, res, msg = self.engine.onboard_domestic_investor(profile)
        self.assertTrue(ok, f"Verification failed: {msg}")
        self.assertEqual(res["assigned_tier"], "TIER_2_FULL_CKYC")
        self.assertEqual(res["risk_tier"], "LOW")
        self.assertTrue(res["commitment_hash"].startswith("0x"))
        self.assertNotIn("Kavita Rao", res["commitment_hash"])
        self.assertNotIn("ABCPS9876P", res["commitment_hash"])

    def test_engine_sanctioned_individual_blocked(self):
        res = self.engine.screen_sanctions_and_pep("Dawood Ibrahim Kaskar", "IND")
        self.assertTrue(res["sanctioned"])
        self.assertTrue(res["frozen"])
        self.assertEqual(res["source"], SanctionsSource.UN_CONSOLIDATED)


if __name__ == "__main__":
    unittest.main()
