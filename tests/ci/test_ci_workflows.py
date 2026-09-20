"""
Unit tests for CI/CD GitHub Actions production validation pipelines.
Validates workflow syntax, required security gates, and coverage threshold enforcement.
"""

import os
import subprocess
import unittest
import yaml
from pathlib import Path


class TestCIWorkflows(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.repo_root = Path(__file__).resolve().parent.parent.parent
        cls.workflows_dir = cls.repo_root / ".github" / "workflows"
        cls.scripts_dir = cls.repo_root / "scripts" / "ci"

    def test_workflow_files_exist_and_are_valid_yaml(self):
        """All workflow files must exist and parse as valid YAML."""
        expected_workflows = [
            "ci-pr-gate.yml",
            "ci-smart-contracts.yml",
            "ci-backend-matrix.yml",
            "ci-container-publish.yml",
        ]
        for wf in expected_workflows:
            wf_path = self.workflows_dir / wf
            self.assertTrue(wf_path.is_file(), f"Missing workflow file: {wf}")
            with open(wf_path, "r", encoding="utf-8") as f:
                content = yaml.safe_load(f)
                self.assertIsInstance(content, dict, f"Invalid YAML structure in {wf}")
                self.assertIn("name", content, f"Missing 'name' in {wf}")
                self.assertIn("jobs", content, f"Missing 'jobs' in {wf}")

    def test_pr_gate_security_jobs(self):
        """ci-pr-gate.yml must contain secret scan, SAST, and coverage gate."""
        pr_gate_path = self.workflows_dir / "ci-pr-gate.yml"
        with open(pr_gate_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
        jobs = data.get("jobs", {})
        self.assertIn("secret-scan", jobs, "Missing secret-scan job in ci-pr-gate.yml")
        self.assertIn("sast-analysis", jobs, "Missing sast-analysis job in ci-pr-gate.yml")
        self.assertIn("coverage-gate", jobs, "Missing coverage-gate job in ci-pr-gate.yml")

    def test_smart_contracts_verification_pipeline(self):
        """ci-smart-contracts.yml must contain forge, slither, and determinism checks."""
        sc_path = self.workflows_dir / "ci-smart-contracts.yml"
        with open(sc_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
        jobs = data.get("jobs", {})
        self.assertIn("forge-verify", jobs)
        self.assertIn("slither-analysis", jobs)
        self.assertIn("bytecode-determinism", jobs)

    def test_container_publish_pipeline(self):
        """ci-container-publish.yml must contain build, trivy CVE scan, and SBOM generation."""
        cont_path = self.workflows_dir / "ci-container-publish.yml"
        with open(cont_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
        jobs = data.get("jobs", {})
        self.assertIn("build-scan-sign", jobs)
        steps = jobs["build-scan-sign"].get("steps", [])
        step_names = [s.get("name", "") for s in steps]
        self.assertTrue(any("Trivy" in name for name in step_names), "Missing Trivy scan step")
        self.assertTrue(any("SBOM" in name or "Syft" in name for name in step_names), "Missing SBOM step")
        self.assertTrue(any("Cosign" in name for name in step_names), "Missing Cosign step")

    def test_coverage_script_execution(self):
        """check-coverage.sh must be executable and return 0 on threshold pass and 1 on failure."""
        script_path = self.scripts_dir / "check-coverage.sh"
        self.assertTrue(script_path.is_file())
        self.assertTrue(os.access(script_path, os.X_OK))

        # Test passing threshold (80%)
        res_pass = subprocess.run(
            [str(script_path), "--threshold", "80"],
            cwd=str(self.repo_root),
            capture_output=True,
            text=True,
        )
        self.assertEqual(res_pass.returncode, 0)
        self.assertIn("[SUCCESS]", res_pass.stdout)

        # Test failing threshold (99.9%)
        res_fail = subprocess.run(
            [str(script_path), "--threshold", "99.9"],
            cwd=str(self.repo_root),
            capture_output=True,
            text=True,
        )
        self.assertEqual(res_fail.returncode, 1)
        self.assertIn("[FAILURE]", res_fail.stdout)


if __name__ == "__main__":
    unittest.main()
