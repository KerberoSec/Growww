#!/usr/bin/env python3
"""
Master Test Runner for Growww Infrastructure, Reliability, and Chaos Test Harnesses:
- Prompt 802: Kubernetes Helm production manifests (Gateway, Matching Engine, OME, Ingress)
- Prompt 803: CI/CD GitHub Actions production validation pipeline & coverage gate
- Prompt 814: Envoy proxy rate limiting & DDoS mitigation filter
- Prompt 903: High-throughput matching engine load & performance benchmark harness
- Prompt 904: Chaos Mesh resilience & network partition disaster injection plan
"""

import sys
import unittest
from pathlib import Path

# Ensure repository root is on sys.path
_repo_root = Path(__file__).resolve().parent.parent
if str(_repo_root) not in sys.path:
    sys.path.insert(0, str(_repo_root))

from infra.k8s.scripts import validate_manifests
from tests.ci.test_ci_workflows import TestCIWorkflows
from infra.envoy.scripts.test_envoy_ratelimit import TestEnvoyConfiguration
from tests.benchmarks.test_matching_engine_benchmark import TestMatchingEngineBenchmark
from tests.chaos.test_chaos_resilience import TestChaosResilience


def run_all():
    print("================================================================================")
    print(" GROWWW EXCHANGE: INFRASTRUCTURE, RELIABILITY & CHAOS MASTER TEST HARNESS")
    print("================================================================================")

    print("\n>>> Phase 1: Validating Prompt 802 Kubernetes Helm Production Manifests...")
    k8s_status = validate_manifests.main()
    if k8s_status != 0:
        print("[-] Phase 1 Failed!")
        return 1

    print("\n>>> Phase 2: Running Unit Tests for Prompts 803, 814, 903, and 904...")
    suite = unittest.TestSuite()
    loader = unittest.TestLoader()

    suite.addTests(loader.loadTestsFromTestCase(TestCIWorkflows))
    suite.addTests(loader.loadTestsFromTestCase(TestEnvoyConfiguration))
    suite.addTests(loader.loadTestsFromTestCase(TestMatchingEngineBenchmark))
    suite.addTests(loader.loadTestsFromTestCase(TestChaosResilience))

    runner = unittest.TextTestRunner(verbosity=2)
    test_result = runner.run(suite)

    if not test_result.wasSuccessful():
        print("[-] Phase 2 Unit Tests Failed!")
        return 1

    print("\n================================================================================")
    print(" [ALL PHASES PASSED] All 5 Prompts Validated Cleanly with 100% Success!")
    print("================================================================================")
    return 0


if __name__ == "__main__":
    sys.exit(run_all())
