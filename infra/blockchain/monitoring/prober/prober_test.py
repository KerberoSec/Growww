#!/usr/bin/env python3
"""
Unit and Integration Tests for Hyperledger Besu QBFT Prober Daemon
File: infra/blockchain/monitoring/prober/prober_test.py
"""

import json
import unittest
from unittest.mock import patch, MagicMock
from prober import ConsensusHealthEvaluator, HealthHTTPRequestHandler


class TestBesuProberEvaluator(unittest.TestCase):
    """Tests consensus state evaluation, quorum calculation, and formatting."""

    def setUp(self):
        self.mock_nodes = [
            {"name": "val-growww-core", "url": "http://val-growww-core:8545"},
            {"name": "val-custodian-nsdl", "url": "http://val-custodian-nsdl:8545"},
            {"name": "val-clearing-corp", "url": "http://val-clearing-corp:8545"},
            {"name": "val-gift-city", "url": "http://val-gift-city:8545"},
        ]
        self.evaluator = ConsensusHealthEvaluator(self.mock_nodes)

    def test_healthy_consortium_quorum(self):
        """All 4 validators online, height synchronized, peers >= 3 -> HEALTHY."""
        def mock_probe(node_info):
            return {
                "name": node_info["name"],
                "block": 4829104,
                "peers": 5,
                "status": "SYNCED",
                "latency_seconds": 0.012,
                "error": None,
            }

        with patch.object(self.evaluator, "probe_single_node", side_effect=mock_probe):
            self.evaluator.probe_all()

        state = self.evaluator.latest_state
        self.assertEqual(state["network_status"], "HEALTHY")
        self.assertEqual(state["latest_block_number"], 4829104)
        self.assertEqual(state["validators_total"], 4)
        self.assertEqual(state["validators_online"], 4)
        self.assertEqual(state["inter_node_delta"], 0)
        self.assertTrue(state["quorum_intact"])

    def test_f_equals_one_tolerated_failure(self):
        """1 validator down (3 of 4 online) -> QBFT quorum intact (3 >= 3), status HEALTHY."""
        def mock_probe(node_info):
            if node_info["name"] == "val-gift-city":
                return {
                    "name": node_info["name"],
                    "block": 0,
                    "peers": 0,
                    "status": "UNREACHABLE",
                    "latency_seconds": 0.0,
                    "error": "Connection refused",
                }
            return {
                "name": node_info["name"],
                "block": 4829105,
                "peers": 4,
                "status": "SYNCED",
                "latency_seconds": 0.015,
                "error": None,
            }

        with patch.object(self.evaluator, "probe_single_node", side_effect=mock_probe):
            self.evaluator.probe_all()

        state = self.evaluator.latest_state
        self.assertEqual(state["validators_online"], 3)
        self.assertTrue(state["quorum_intact"])
        self.assertEqual(state["latest_block_number"], 4829105)

    def test_critical_quorum_loss_two_failures(self):
        """2 validators down (only 2 of 4 online) -> Quorum lost, status CRITICAL_QUORUM_LOST."""
        def mock_probe(node_info):
            if node_info["name"] in ("val-clearing-corp", "val-gift-city"):
                return {
                    "name": node_info["name"],
                    "block": 0,
                    "peers": 0,
                    "status": "UNREACHABLE",
                    "latency_seconds": 0.0,
                    "error": "Timeout",
                }
            return {
                "name": node_info["name"],
                "block": 4829100,
                "peers": 2,
                "status": "SYNCED",
                "latency_seconds": 0.02,
                "error": None,
            }

        with patch.object(self.evaluator, "probe_single_node", side_effect=mock_probe):
            self.evaluator.probe_all()

        state = self.evaluator.latest_state
        self.assertEqual(state["validators_online"], 2)
        self.assertFalse(state["quorum_intact"])
        self.assertEqual(state["network_status"], "CRITICAL_QUORUM_LOST")

    def test_block_height_divergence_detection(self):
        """One node lags by 6 blocks -> status DEGRADED_DESYNC."""
        def mock_probe(node_info):
            block = 4829100 if node_info["name"] == "val-gift-city" else 4829107
            return {
                "name": node_info["name"],
                "block": block,
                "peers": 4,
                "status": "SYNCED",
                "latency_seconds": 0.01,
                "error": None,
            }

        with patch.object(self.evaluator, "probe_single_node", side_effect=mock_probe):
            self.evaluator.probe_all()

        state = self.evaluator.latest_state
        self.assertEqual(state["inter_node_delta"], 7)
        self.assertEqual(state["network_status"], "DEGRADED_DESYNC")

    def test_prometheus_metrics_formatting(self):
        """Verify rendered Prometheus exposition syntax."""
        def mock_probe(node_info):
            return {
                "name": node_info["name"],
                "block": 100,
                "peers": 4,
                "status": "SYNCED",
                "latency_seconds": 0.025,
                "error": None,
            }

        with patch.object(self.evaluator, "probe_single_node", side_effect=mock_probe):
            self.evaluator.probe_all()

        metrics_text = self.evaluator.render_prometheus_metrics()
        self.assertIn("growww_consensus_health_status", metrics_text)
        self.assertIn('growww_validators_online{cluster="prod"} 4', metrics_text)
        self.assertIn('growww_node_block_height{cluster="prod",node="val-growww-core",status="SYNCED"} 100', metrics_text)
        self.assertIn('growww_node_peer_count{cluster="prod",node="val-growww-core"} 4', metrics_text)
        self.assertIn('growww_node_rpc_latency_seconds{cluster="prod",node="val-growww-core"} 0.0250', metrics_text)

    def test_json_schema_matches_prompt_310_specification(self):
        """Verify that JSON output schema matches section 8.2 of Prompt 310."""
        def mock_probe(node_info):
            return {
                "name": node_info["name"],
                "block": 4829104,
                "peers": 5,
                "status": "SYNCED",
                "latency_seconds": 0.01,
                "error": None,
            }

        with patch.object(self.evaluator, "probe_single_node", side_effect=mock_probe):
            self.evaluator.probe_all()

        payload = {
            "timestamp": self.evaluator.latest_state["timestamp"],
            "network_status": self.evaluator.latest_state["network_status"],
            "chain_id": self.evaluator.latest_state["chain_id"],
            "consensus_algorithm": self.evaluator.latest_state["consensus_algorithm"],
            "latest_block_number": self.evaluator.latest_state["latest_block_number"],
            "block_time_seconds": self.evaluator.latest_state["block_time_seconds"],
            "validators_total": self.evaluator.latest_state["validators_total"],
            "validators_online": self.evaluator.latest_state["validators_online"],
            "nodes": [
                {
                    "name": n["name"],
                    "block": n["block"],
                    "peers": n["peers"],
                    "status": n["status"],
                }
                for n in self.evaluator.latest_state["nodes"]
            ],
        }

        # Validate presence of all expected keys
        required_keys = [
            "timestamp", "network_status", "chain_id", "consensus_algorithm",
            "latest_block_number", "block_time_seconds", "validators_total",
            "validators_online", "nodes"
        ]
        for key in required_keys:
            self.assertIn(key, payload)

        self.assertEqual(len(payload["nodes"]), 4)
        for node in payload["nodes"]:
            self.assertIn("name", node)
            self.assertIn("block", node)
            self.assertIn("peers", node)
            self.assertIn("status", node)


if __name__ == "__main__":
    unittest.main()
