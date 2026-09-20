#!/usr/bin/env bash
set -euo pipefail

# Health verification script for Hyperledger Besu QBFT Consortium Network
RPC_URL="${1:-http://localhost:8545}"

echo "=== Verifying Besu QBFT Network Health at ${RPC_URL} ==="

# Check block number
BLOCK_HEX=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -H "Content-Type: application/json" "${RPC_URL}" | grep -o '"result":"[^"]*"' | cut -d'"' -f4 || echo "")

if [[ -n "${BLOCK_HEX}" ]]; then
    BLOCK_NUM=$((16#${BLOCK_HEX#0x}))
    echo "Current Block Height: ${BLOCK_NUM}"
else
    echo "Warning: Unable to reach RPC endpoint at ${RPC_URL}. Target may be offline or starting."
fi

# Check peer count
PEERS_HEX=$(curl -s -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":2}' -H "Content-Type: application/json" "${RPC_URL}" | grep -o '"result":"[^"]*"' | cut -d'"' -f4 || echo "")

if [[ -n "${PEERS_HEX}" ]]; then
    PEER_COUNT=$((16#${PEERS_HEX#0x}))
    echo "Connected Peers: ${PEER_COUNT}"
fi

echo "Verification check completed."
