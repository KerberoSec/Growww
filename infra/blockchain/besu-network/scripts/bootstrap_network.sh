#!/usr/bin/env bash
set -euo pipefail

# Bootstrap script for Hyperledger Besu QBFT Consortium Network
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NETWORK_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "=== Initializing Hyperledger Besu QBFT Network ==="
echo "Network Directory: ${NETWORK_DIR}"

GENESIS_FILE="${NETWORK_DIR}/genesis/genesis.json"
if [[ ! -f "${GENESIS_FILE}" ]]; then
    echo "Error: Genesis file not found at ${GENESIS_FILE}"
    exit 1
fi

echo "Verifying Genesis configuration..."
CHAIN_ID=$(grep -o '"chainId": [0-9]*' "${GENESIS_FILE}" | awk '{print $2}')
echo "Configured Chain ID: ${CHAIN_ID}"

echo "Validating Helm charts..."
if command -v helm &> /dev/null; then
    helm lint "${NETWORK_DIR}/helm/besu-node"
    echo "Helm chart validation successful."
else
    echo "Helm not installed locally, skipping local lint."
fi

echo "Network bootstrap prerequisites verified."
exit 0
