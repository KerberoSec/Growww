#!/usr/bin/env python3
"""
Automated Test Suite for Envoy Rate Limiting & DDoS Mitigation Architecture (Prompt 814).
Validates configuration schema, rate limit descriptors, circuit breakers, and DDoS mitigation heuristics.
"""

import os
import re
import sys
import unittest
import yaml
from pathlib import Path


class TestEnvoyConfiguration(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.repo_root = Path(__file__).resolve().parent.parent.parent.parent
        cls.envoy_dir = cls.repo_root / "infra" / "envoy"
        cls.envoy_yaml_path = cls.envoy_dir / "envoy.yaml"
        cls.ratelimit_cfg_path = cls.envoy_dir / "config" / "ratelimit-config.yaml"
        cls.lua_filter_path = cls.envoy_dir / "filters" / "ddos_mitigation.lua"

        with open(cls.envoy_yaml_path, "r", encoding="utf-8") as f:
            cls.envoy_config = yaml.safe_load(f)

        with open(cls.ratelimit_cfg_path, "r", encoding="utf-8") as f:
            cls.ratelimit_config = yaml.safe_load(f)

        with open(cls.lua_filter_path, "r", encoding="utf-8") as f:
            cls.lua_code = f.read()

    def test_envoy_yaml_structure(self):
        """Verify top-level static_resources, listeners, clusters, and admin block."""
        self.assertIn("static_resources", self.envoy_config)
        self.assertIn("admin", self.envoy_config)
        static = self.envoy_config["static_resources"]
        self.assertIn("listeners", static)
        self.assertIn("clusters", static)
        self.assertGreaterEqual(len(static["listeners"]), 1)
        self.assertGreaterEqual(len(static["clusters"]), 3)

    def test_connection_limit_and_slowloris_defense(self):
        """Verify TCP connection limit and buffer clamp exist for L4 DDoS defense."""
        listener = self.envoy_config["static_resources"]["listeners"][0]
        self.assertIn("per_connection_buffer_limit_bytes", listener)
        self.assertLessEqual(listener["per_connection_buffer_limit_bytes"], 65536)

        filters = listener["filter_chains"][0]["filters"]
        filter_names = [f.get("name") for f in filters]
        self.assertIn("envoy.filters.network.connection_limit", filter_names)

        conn_limit_filter = next(f for f in filters if f.get("name") == "envoy.filters.network.connection_limit")
        cfg = conn_limit_filter["typed_config"]
        self.assertGreaterEqual(cfg.get("max_connections"), 10000)

    def test_http_filters_chain_ordering(self):
        """Verify the defensive filter chain contains local ratelimit, RBAC, Lua DDoS, global ratelimit, adaptive concurrency."""
        listener = self.envoy_config["static_resources"]["listeners"][0]
        hcm = next(
            f for f in listener["filter_chains"][0]["filters"]
            if f.get("name") == "envoy.filters.network.http_connection_manager"
        )
        http_filters = [f.get("name") for f in hcm["typed_config"]["http_filters"]]

        expected_filters = [
            "envoy.filters.http.local_ratelimit",
            "envoy.filters.http.rbac",
            "envoy.filters.http.lua",
            "envoy.filters.http.ratelimit",
            "envoy.filters.http.adaptive_concurrency",
            "envoy.filters.http.router",
        ]
        for ef in expected_filters:
            self.assertIn(ef, http_filters, f"Filter {ef} is missing from HTTP filter chain")

    def test_ratelimit_descriptors_and_tiers(self):
        """Verify global ratelimit configuration adheres to tiered throughput and brute force defense."""
        self.assertEqual(self.ratelimit_config.get("domain"), "growww-gateway-limits")
        descriptors = self.ratelimit_config.get("descriptors", [])

        # Check Auth Brute Force Limit
        auth_desc = next((d for d in descriptors if d.get("value") == "auth_login"), None)
        self.assertIsNotNone(auth_desc, "Missing auth_login rate limit descriptor")
        remote_addr_limit = auth_desc["descriptors"][0]["rate_limit"]
        self.assertEqual(remote_addr_limit["unit"], "minute")
        self.assertLessEqual(remote_addr_limit["requests_per_unit"], 10)

        # Check Order Tiers
        orders_desc = next((d for d in descriptors if d.get("value") == "orders"), None)
        self.assertIsNotNone(orders_desc, "Missing orders rate limit descriptor")
        tier_descriptors = {sub["value"]: sub["rate_limit"] for sub in orders_desc["descriptors"]}

        self.assertIn("tier0_public", tier_descriptors)
        self.assertIn("tier1_retail", tier_descriptors)
        self.assertIn("tier2_institutional", tier_descriptors)
        self.assertIn("tier3_market_maker", tier_descriptors)

        self.assertEqual(tier_descriptors["tier0_public"]["requests_per_unit"], 10)
        self.assertEqual(tier_descriptors["tier1_retail"]["requests_per_unit"], 100)
        self.assertEqual(tier_descriptors["tier2_institutional"]["requests_per_unit"], 2000)
        self.assertEqual(tier_descriptors["tier3_market_maker"]["requests_per_unit"], 10000)

    def test_upstream_clusters_circuit_breakers_and_outliers(self):
        """Verify upstream matching engine cluster has active circuit breaking and outlier ejection."""
        clusters = self.envoy_config["static_resources"]["clusters"]
        me_cluster = next((c for c in clusters if c.get("name") == "matching_engine_cluster"), None)
        self.assertIsNotNone(me_cluster, "Missing matching_engine_cluster")

        self.assertIn("circuit_breakers", me_cluster)
        cb = me_cluster["circuit_breakers"]["thresholds"][0]
        self.assertGreater(cb["max_connections"], 0)
        self.assertGreater(cb["max_requests"], 0)

        self.assertIn("outlier_detection", me_cluster)
        outlier = me_cluster["outlier_detection"]
        self.assertIn("consecutive_5xx", outlier)
        self.assertIn("base_ejection_time", outlier)

    def test_lua_ddos_script_rules(self):
        """Verify the Lua DDoS filter script blocks automated scanners and enforces clock skew."""
        # Scanner detection
        for bot in ["sqlmap", "nikto", "masscan", "gobuster"]:
            self.assertIn(f'"{bot}"', self.lua_code, f"Bot '{bot}' must be present in blacklist table")

        # Status 403 on scanner match
        self.assertIn('["x-firewall-action"] = "blocked"', self.lua_code)
        self.assertIn('status"] = "403"', self.lua_code)

        # Replay attack clock skew check
        self.assertIn("x-request-timestamp", self.lua_code)
        self.assertIn("300", self.lua_code)


if __name__ == "__main__":
    unittest.main()
