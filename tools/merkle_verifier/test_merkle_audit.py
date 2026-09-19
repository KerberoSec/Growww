#!/usr/bin/env python3
"""
Unit tests for Python Merkle Solvency & Proof of Reserves verification utility.
Tests Keccak-256 accuracy, sorted-pair Merkle tree hashing, and user inclusion proofs.
"""

import unittest
import json
import os
import tempfile
from merkle_audit import (
    keccak256,
    compute_leaf,
    hash_pair,
    verify_merkle_proof,
    SolvencyMerkleTree,
)


class TestMerkleAudit(unittest.TestCase):

    def test_keccak256_standard_vectors(self):
        # Empty string
        self.assertEqual(
            keccak256(b"").hex(),
            "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
        )
        # Standard Ethereum message "hello"
        self.assertEqual(
            keccak256(b"hello").hex(),
            "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8"
        )

    def test_hash_pair_sorting_invariance(self):
        a = keccak256(b"nodeA")
        b = keccak256(b"nodeB")

        hash_ab = hash_pair(a, b)
        hash_ba = hash_pair(b, a)

        # Hash pair must be canonically sorted and identical regardless of order
        self.assertEqual(hash_ab, hash_ba)

    def test_single_account_tree(self):
        tree = SolvencyMerkleTree()
        tree.add_account("user_single", 500000000)
        root = tree.build_tree()

        proof = tree.get_proof(0)
        leaf = compute_leaf("user_single", 500000000)

        # For single-node tree, proof is empty and leaf equals root
        self.assertEqual(proof, [])
        self.assertTrue(verify_merkle_proof(leaf, proof, root))

    def test_two_accounts_tree(self):
        tree = SolvencyMerkleTree()
        tree.add_account("alice", 1000)
        tree.add_account("bob", 2000)
        root = tree.build_tree()

        # Both accounts must verify
        for idx, acc in enumerate(tree.accounts):
            proof = tree.get_proof(idx)
            leaf = bytes.fromhex(acc["leaf_hex"][2:])
            self.assertTrue(verify_merkle_proof(leaf, proof, root))

    def test_four_accounts_tree(self):
        tree = SolvencyMerkleTree()
        for i in range(4):
            tree.add_account(f"investor_{i}", (i + 1) * 10_00000000)

        root = tree.build_tree()

        for idx, acc in enumerate(tree.accounts):
            proof = tree.get_proof(idx)
            leaf = bytes.fromhex(acc["leaf_hex"][2:])
            self.assertTrue(
                verify_merkle_proof(leaf, proof, root),
                f"Failed to verify account {acc['user_id']}"
            )

    def test_odd_number_of_accounts(self):
        tree = SolvencyMerkleTree()
        for i in range(5):
            tree.add_account(f"trader_{i}", (i + 1) * 5_00000000)

        root = tree.build_tree()

        for idx, acc in enumerate(tree.accounts):
            proof = tree.get_proof(idx)
            leaf = bytes.fromhex(acc["leaf_hex"][2:])
            self.assertTrue(
                verify_merkle_proof(leaf, proof, root),
                f"Failed to verify account {acc['user_id']} in odd-sized tree"
            )

    def test_tamper_detection_fails_verification(self):
        tree = SolvencyMerkleTree()
        tree.add_account("charlie", 5000)
        tree.add_account("dave", 8000)
        root = tree.build_tree()

        proof = tree.get_proof(0)
        acc = tree.accounts[0]

        # 1. Tamper with balance
        forged_leaf = compute_leaf(acc["user_id"], 999999)
        self.assertFalse(verify_merkle_proof(forged_leaf, proof, root))

        # 2. Tamper with proof
        tampered_proof = [keccak256(b"malicious_hash")]
        leaf = bytes.fromhex(acc["leaf_hex"][2:])
        self.assertFalse(verify_merkle_proof(leaf, tampered_proof, root))

    def test_insolvency_check(self):
        tree = SolvencyMerkleTree()
        tree.add_account("user_1", 100_000)
        tree.add_account("user_2", 200_000)
        # Total liabilities = 300,000. Reserves = 250,000 (insolvent)
        with self.assertRaises(ValueError) as ctx:
            tree.generate_manifest(1, 250_000, "ipfs://insolvent")
        self.assertIn("Insolvency detected", str(ctx.exception))

    def test_manifest_export_and_audit(self):
        tree = SolvencyMerkleTree()
        tree.add_account("userA", 1000)
        tree.add_account("userB", 2000)
        tree.add_account("userC", 3000)

        manifest = tree.generate_manifest(epoch_id=42, total_reserves_e8=7000, report_uri="ipfs://audit-epoch-42")

        self.assertEqual(manifest["epoch_id"], 42)
        self.assertEqual(manifest["total_liabilities_e8"], 6000)
        self.assertEqual(manifest["total_reserves_e8"], 7000)
        self.assertGreater(manifest["solvency_ratio"], 1.0)
        self.assertEqual(manifest["solvency_status"], "SOLVENT")
        self.assertEqual(manifest["account_count"], 3)
        self.assertEqual(len(manifest["user_proofs"]), 3)


if __name__ == "__main__":
    unittest.main()
