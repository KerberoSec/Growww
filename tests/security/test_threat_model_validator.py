import unittest
import os
import sys

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "scripts", "security")))

from validate_threat_model import parse_threat_model, validate_threats

class TestThreatModelValidator(unittest.TestCase):

    def setUp(self):
        self.repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
        self.threat_doc = os.path.join(self.repo_root, "docs", "security", "threat_model.md")

    def test_parse_real_threat_model_file(self):
        threats = parse_threat_model(self.threat_doc)
        self.assertGreaterEqual(len(threats), 8)
        ids = [t["id"] for t in threats]
        self.assertIn("TH-001", ids)
        self.assertIn("TH-006", ids)
        self.assertIn("TH-007", ids)

    def test_all_real_threats_pass_validation(self):
        threats = parse_threat_model(self.threat_doc)
        errors = validate_threats(threats)
        self.assertEqual(len(errors), 0, f"Expected 0 errors, got: {errors}")

    def test_validator_catches_unmitigated_threat(self):
        bad_threats = [{
            "id": "TH-999",
            "category": "Spoofing",
            "component": "API",
            "description": "Unauthenticated API access",
            "cvss": "9.0",
            "dread": "High",
            "mitigation": "None planned yet",
            "prompt": "Prompt 001",
            "status": "Open" # Not Mitigated!
        }]
        errors = validate_threats(bad_threats)
        self.assertTrue(any("expected 'Mitigated'" in e for e in errors))

    def test_validator_catches_invalid_stride_category(self):
        bad_threats = [{
            "id": "TH-998",
            "category": "BadCategory",
            "component": "Database",
            "description": "Data corruption",
            "cvss": "7.0",
            "dread": "Med",
            "mitigation": "Automated replication backup",
            "prompt": "Prompt 401",
            "status": "Mitigated"
        }]
        errors = validate_threats(bad_threats)
        self.assertTrue(any("Invalid STRIDE category" in e for e in errors))

if __name__ == '__main__':
    unittest.main()
