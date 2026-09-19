#!/usr/bin/env python3
"""
Growww / NBSE Sovereign Exchange - Proof of Solvency & Merkle Verification Utility
Author: Agent 5 (Prompts 200-249 Infrastructure)
Description:
    Production-grade cryptographic utility for generating Proof of Liabilities (PoL)
    Merkle Trees, individual user inclusion proofs, and verifying cryptographic solvency
    matching the on-chain MerkleRegistry.sol smart contract.
Zero-Dependency: Implements pure Python Keccak-256 (FIPS-202 Keccak-f[1600])
"""

import sys
import os
import json
import argparse
from typing import List, Dict, Tuple, Optional

# --- Pure Python Keccak-256 Implementation ---

RC = [
    0x0000000000000001, 0x0000000000008082, 0x800000000000808A, 0x8000000080008000,
    0x000000000000808B, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
    0x000000000000008A, 0x0000000000000088, 0x0000000080008009, 0x000000008000000A,
    0x000000008000808B, 0x800000000000008B, 0x8000000000008089, 0x8000000000008003,
    0x8000000000008002, 0x8000000000000080, 0x000000000000800A, 0x800000008000000A,
    0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008
]

RHO = [
    [0, 36, 3, 41, 18],
    [1, 44, 10, 45, 2],
    [62, 6, 43, 15, 61],
    [28, 55, 25, 21, 56],
    [27, 20, 39, 8, 14]
]


def _rot(x: int, n: int) -> int:
    return ((x << (n % 64)) & 0xFFFFFFFFFFFFFFFF) | (x >> (64 - (n % 64)))


def _keccak_f1600(state: List[int]) -> None:
    lanes = [[state[x + 5 * y] for y in range(5)] for x in range(5)]
    for r in range(24):
        # Theta
        c = [lanes[x][0] ^ lanes[x][1] ^ lanes[x][2] ^ lanes[x][3] ^ lanes[x][4] for x in range(5)]
        d = [c[(x - 1) % 5] ^ _rot(c[(x + 1) % 5], 1) for x in range(5)]
        for x in range(5):
            for y in range(5):
                lanes[x][y] ^= d[x]
        # Rho and Pi
        b = [[0] * 5 for _ in range(5)]
        for x in range(5):
            for y in range(5):
                b[y][(2 * x + 3 * y) % 5] = _rot(lanes[x][y], RHO[x][y])
        # Chi
        for x in range(5):
            for y in range(5):
                lanes[x][y] = b[x][y] ^ ((~b[(x + 1) % 5][y]) & b[(x + 2) % 5][y])
        # Iota
        lanes[0][0] ^= RC[r]
    for x in range(5):
        for y in range(5):
            state[x + 5 * y] = lanes[x][y]


def keccak256(data: bytes) -> bytes:
    """Computes the Keccak-256 hash of arbitrary bytes, identical to Ethereum's keccak256."""
    rate = 136  # 1088 bits / 8
    state = [0] * 25
    pad_len = rate - (len(data) % rate)
    if pad_len == 1:
        padded = data + b'\x81'
    else:
        padded = data + b'\x01' + b'\x00' * (pad_len - 2) + b'\x80'

    for offset in range(0, len(padded), rate):
        block = padded[offset:offset + rate]
        for i in range(len(block) // 8):
            lane = int.from_bytes(block[i * 8:(i + 1) * 8], 'little')
            state[i] ^= lane
        _keccak_f1600(state)

    out = bytearray()
    for i in range(4):
        out.extend(state[i].to_bytes(8, 'little'))
    return bytes(out)


# --- Merkle Solvency Functions Matching MerkleRegistry.sol ---

def compute_leaf(user_id: str, balance_e8: int) -> bytes:
    """
    Computes leaf hash for user account liability.
    Leaf formula: keccak256(abi.encodePacked(keccak256(user_id), uint256(balance_e8)))
    Matches MerkleRegistry.sol and zero-PII regulatory standard.
    """
    user_hash = keccak256(user_id.encode('utf-8'))
    balance_bytes = balance_e8.to_bytes(32, byteorder='big')
    return keccak256(user_hash + balance_bytes)


def hash_pair(left: bytes, right: bytes) -> bytes:
    """
    Sorted pair hashing matching OpenZeppelin and MerkleRegistry.sol:
    if (a <= b) keccak256(abi.encodePacked(a, b)) else keccak256(abi.encodePacked(b, a))
    """
    if left <= right:
        return keccak256(left + right)
    else:
        return keccak256(right + left)


def verify_merkle_proof(leaf: bytes, proof: List[bytes], root: bytes) -> bool:
    """
    Cryptographically verifies a Merkle proof against a root hash.
    Direct implementation of MerkleRegistry.sol verifyProof() logic.
    """
    computed = leaf
    for sibling in proof:
        computed = hash_pair(computed, sibling)
    return computed == root


class SolvencyMerkleTree:
    """
    Solvency Merkle Tree builder and proof generator for Growww / NBSE liabilities.
    """
    def __init__(self):
        self.accounts: List[Dict] = []
        self.leaves: List[bytes] = []
        self.levels: List[List[bytes]] = []
        self.root: Optional[bytes] = None

    def add_account(self, user_id: str, balance_e8: int) -> bytes:
        leaf = compute_leaf(user_id, balance_e8)
        self.accounts.append({
            "user_id": user_id,
            "balance_e8": balance_e8,
            "leaf_hex": "0x" + leaf.hex()
        })
        self.leaves.append(leaf)
        return leaf

    def build_tree(self) -> bytes:
        if not self.leaves:
            raise ValueError("Cannot build tree with zero accounts")

        # Sort leaves deterministically by their hash
        indexed_leaves = sorted(enumerate(self.leaves), key=lambda item: item[1])
        # Re-index accounts to match sorted order
        self.accounts = [self.accounts[i] for i, _ in indexed_leaves]
        current_level = [leaf for _, leaf in indexed_leaves]
        self.leaves = list(current_level)
        self.levels = [current_level]

        while len(current_level) > 1:
            next_level = []
            for i in range(0, len(current_level), 2):
                if i + 1 < len(current_level):
                    parent = hash_pair(current_level[i], current_level[i + 1])
                else:
                    # Odd number of leaves: elevate the odd node or pair with itself
                    parent = hash_pair(current_level[i], current_level[i])
                next_level.append(parent)
            self.levels.append(next_level)
            current_level = next_level

        self.root = current_level[0]
        return self.root

    def get_proof(self, account_index: int) -> List[bytes]:
        """Generates the Merkle inclusion proof for the leaf at account_index."""
        if not self.levels:
            self.build_tree()

        proof = []
        idx = account_index

        for level in self.levels[:-1]:
            is_odd = (len(level) % 2 != 0)
            if idx % 2 == 0:
                # Sibling is on the right
                if idx + 1 < len(level):
                    proof.append(level[idx + 1])
                elif is_odd and idx == len(level) - 1:
                    proof.append(level[idx])
            else:
                # Sibling is on the left
                proof.append(level[idx - 1])
            idx //= 2

        return proof

    def generate_manifest(self, epoch_id: int, total_reserves_e8: int, report_uri: str) -> Dict:
        """Generates a complete auditable solvency epoch manifest."""
        if self.root is None:
            self.build_tree()

        total_liabilities = sum(acc["balance_e8"] for acc in self.accounts)
        if total_reserves_e8 < total_liabilities:
            raise ValueError(
                f"Insolvency detected: Reserves ({total_reserves_e8}) < Liabilities ({total_liabilities})"
            )

        solvency_ratio = (total_reserves_e8 / total_liabilities) if total_liabilities > 0 else 1.0

        user_proofs = {}
        for idx, acc in enumerate(self.accounts):
            proof_bytes = self.get_proof(idx)
            proof_hex = ["0x" + p.hex() for p in proof_bytes]
            # Self-verification check
            leaf = bytes.fromhex(acc["leaf_hex"][2:])
            assert verify_merkle_proof(leaf, proof_bytes, self.root), "Internal proof verification failure"
            user_proofs[acc["user_id"]] = {
                "balance_e8": acc["balance_e8"],
                "leaf_hash": acc["leaf_hex"],
                "merkle_proof": proof_hex
            }

        return {
            "epoch_id": epoch_id,
            "liabilities_merkle_root": "0x" + self.root.hex(),
            "total_liabilities_e8": total_liabilities,
            "total_reserves_e8": total_reserves_e8,
            "solvency_ratio": round(solvency_ratio, 6),
            "solvency_status": "SOLVENT" if total_reserves_e8 >= total_liabilities else "INSOLVENT",
            "audit_report_uri": report_uri,
            "account_count": len(self.accounts),
            "user_proofs": user_proofs
        }


# --- CLI Interface ---

def main():
    parser = argparse.ArgumentParser(description="Growww Proof of Solvency Verification Utility")
    subparsers = parser.add_subparsers(dest="command")

    # Generate command
    gen_parser = subparsers.add_parser("generate", help="Build Merkle tree and generate epoch manifest")
    gen_parser.add_argument("--accounts", type=str, required=True, help="Path to JSON file with accounts")
    gen_parser.add_argument("--reserves", type=int, required=True, help="Total reserves in e8 satoshis")
    gen_parser.add_argument("--epoch", type=int, default=1, help="Epoch ID")
    gen_parser.add_argument("--report-uri", type=str, default="ipfs://QmGrowwwSolvency", help="Audit report URI")
    gen_parser.add_argument("--output", type=str, default="solvency_manifest.json", help="Output file path")

    # Verify command
    verify_parser = subparsers.add_parser("verify", help="Verify single user inclusion proof")
    verify_parser.add_argument("--user-id", type=str, required=True, help="User ID / Anonymized ID")
    verify_parser.add_argument("--balance", type=int, required=True, help="Account balance in e8 satoshis")
    verify_parser.add_argument("--root", type=str, required=True, help="Liabilities root hex (0x...)")
    verify_parser.add_argument("--proof", type=str, required=True, help="Comma-separated proof hashes or JSON file")

    args = parser.parse_args()

    if args.command == "generate":
        with open(args.accounts, "r") as f:
            accounts_data = json.load(f)

        tree = SolvencyMerkleTree()
        for item in accounts_data:
            tree.add_account(item["user_id"], int(item["balance_e8"]))

        manifest = tree.generate_manifest(args.epoch, args.reserves, args.report_uri)
        with open(args.output, "w") as f:
            json.dump(manifest, f, indent=2)

        print(f"[SUCCESS] Epoch {args.epoch} manifest generated successfully!")
        print(f"Liabilities Root: {manifest['liabilities_merkle_root']}")
        print(f"Total Liabilities: {manifest['total_liabilities_e8']} e8")
        print(f"Total Reserves:    {manifest['total_reserves_e8']} e8")
        print(f"Solvency Ratio:    {manifest['solvency_ratio'] * 100:.2f}%")
        print(f"Manifest written to: {args.output}")

    elif args.command == "verify":
        leaf = compute_leaf(args.user_id, args.balance)
        root_bytes = bytes.fromhex(args.root.replace("0x", ""))

        if os.path.exists(args.proof):
            with open(args.proof, "r") as f:
                proof_list = json.load(f)
        else:
            proof_list = [p.strip() for p in args.proof.split(",") if p.strip()]

        proof_bytes = [bytes.fromhex(p.replace("0x", "")) for p in proof_list]
        is_valid = verify_merkle_proof(leaf, proof_bytes, root_bytes)

        if is_valid:
            print(f"[VERIFIED] User '{args.user_id}' with balance {args.balance} e8 is provably included in root {args.root}!")
            sys.exit(0)
        else:
            print(f"[REJECTED] Proof verification failed for user '{args.user_id}' against root {args.root}!")
            sys.exit(1)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
