"""
Unit & Integration Tests for Chaos Mesh Resilience & Disaster Recovery Harness (Prompt 904).
Validates CRD schemas, scenario definitions, runbook standards, and orchestrator execution.
"""

import json
import os
import sys
import unittest
import yaml
from pathlib import Path

# Ensure repository root is on sys.path
_repo_root = Path(__file__).resolve().parent.parent.parent
if str(_repo_root) not in sys.path:
    sys.path.insert(0, str(_repo_root))

from tests.chaos.scripts.run_chaos_suite import ChaosSuiteOrchestrator


class TestChaosResilience(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.repo_root = Path(__file__).resolve().parent.parent.parent
        cls.chaos_dir = cls.repo_root / "tests" / "chaos"
        cls.experiments_dir = cls.chaos_dir / "experiments"
        cls.workflows_dir = cls.chaos_dir / "workflows"
        cls.runbooks_dir = cls.chaos_dir / "runbooks"
        cls.reports_dir = cls.chaos_dir / "reports"

    def test_chaos_mesh_experiments_crd_syntax_and_kinds(self):
        """All Chaos Mesh manifests must parse as valid YAML with correct apiVersion and kinds."""
        expected_kinds = {"PodChaos", "NetworkChaos", "StressChaos", "IOChaos", "DNSChaos"}
        found_kinds = set()

        for f in self.experiments_dir.glob("*.yaml"):
            with open(f, "r", encoding="utf-8") as fp:
                docs = list(yaml.safe_load_all(fp))
                for doc in docs:
                    if not doc:
                        continue
                    self.assertEqual(doc.get("apiVersion"), "chaos-mesh.org/v1alpha1")
                    kind = doc.get("kind")
                    self.assertIn(kind, expected_kinds)
                    found_kinds.add(kind)
                    self.assertIn("metadata", doc)
                    self.assertIn("spec", doc)
                    # Must specify a target duration
                    self.assertIn("duration", doc["spec"])

        self.assertIn("PodChaos", found_kinds)
        self.assertIn("NetworkChaos", found_kinds)

    def test_chaos_workflow_crd_structure(self):
        """Chaos Mesh Workflow must define a valid serial disaster injection pipeline."""
        wf_file = self.workflows_dir / "chaos_game_day_workflow.yaml"
        self.assertTrue(wf_file.is_file(), "Missing chaos_game_day_workflow.yaml")

        with open(wf_file, "r", encoding="utf-8") as fp:
            wf = yaml.safe_load(fp)

        self.assertEqual(wf.get("apiVersion"), "chaos-mesh.org/v1alpha1")
        self.assertEqual(wf.get("kind"), "Workflow")
        spec = wf.get("spec", {})
        self.assertIn("entry", spec)
        self.assertIn("templates", spec)

    def test_disaster_recovery_runbook_content(self):
        """Game day runbook must document SEBI/RBI compliance, RTO < 30s, and RPO = 0."""
        runbook = self.runbooks_dir / "GAME_DAY_RUNBOOK.md"
        self.assertTrue(runbook.is_file(), "Missing GAME_DAY_RUNBOOK.md")

        with open(runbook, "r", encoding="utf-8") as fp:
            content = fp.read()

        self.assertIn("RTO < 30 seconds", content)
        self.assertIn("RPO = 0", content)
        self.assertIn("Abort & Rollback Triggers", content)
        self.assertIn("SEBI", content)

    def test_chaos_orchestrator_execution_and_metrics(self):
        """Orchestrator must execute all 7 disaster scenarios with max RTO < 30s and RPO = 0."""
        orchestrator = ChaosSuiteOrchestrator(mode="simulation")
        summary = orchestrator.run_all()

        self.assertEqual(summary.total_scenarios, 7)
        self.assertEqual(summary.failed_scenarios, 0)
        self.assertTrue(summary.zero_state_loss_verified)
        self.assertTrue(summary.overall_resilience_passed)
        self.assertLess(summary.max_observed_rto_sec, 30.0)

        # Verify Besu 1-node partition had zero downtime (uninterrupted consensus)
        besu_single = next(s for s in summary.scenarios if s.scenario_id == "DR-SCEN-05")
        self.assertEqual(besu_single.observed_rto_sec, 0.0)
        self.assertEqual(besu_single.observed_rpo_state_loss, 0)

        # Verify Besu quorum loss recovered cleanly
        besu_quorum = next(s for s in summary.scenarios if s.scenario_id == "DR-SCEN-06")
        self.assertTrue(besu_quorum.rto_passed)
        self.assertEqual(besu_quorum.observed_rpo_state_loss, 0)

        # Export report and verify JSON
        out_report = self.reports_dir / "test_verification_report.json"
        orchestrator.export_report(summary, out_report)
        self.assertTrue(out_report.is_file())

        with open(out_report, "r", encoding="utf-8") as fp:
            data = json.load(fp)
            self.assertEqual(data["total_scenarios"], 7)
            self.assertEqual(data["zero_state_loss_verified"], True)


if __name__ == "__main__":
    unittest.main()
