# Permissioned Blockchain Benchmarking Suite

This directory contains the comparative benchmarking harness evaluating Hyperledger Besu (QBFT), Hyperledger Fabric (Raft), and Polygon CDK for the Growww securities settlement platform.

## Directory Layout

- `besu-qbft/`: 4-node Hyperledger Besu QBFT cluster with genesis, config, and docker-compose.
- `fabric-raft/`: 4-peer Hyperledger Fabric v2.5 cluster with Raft ordering and configtx.
- `polygon-cdk/`: 4-node Polygon CDK / Edge cluster with IBFT PoA consensus.
- `results/`: Recorded benchmark execution metrics (`benchmark_results.json`).

## Running the Benchmarks

### 1. Hyperledger Besu (QBFT)
```bash
cd besu-qbft
docker compose up -d
# Wait for 4-node QBFT network convergence (block production every 2 seconds)
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -H "Content-Type: application/json" http://localhost:8545
```

### 2. Hyperledger Fabric (Raft)
```bash
cd fabric-raft
docker compose up -d
```

### 3. Polygon CDK
```bash
cd polygon-cdk
docker compose up -d
```

## Results Summary

Benchmarking results demonstrate Hyperledger Besu delivering 2,415.6 average TPS and 1,180 ms p95 latency under high-load batch DvP settlement with immediate 1-block deterministic finality. Full evaluation details are documented in [ADR 001](../../docs/architecture/adr_001_blockchain_platform_selection.md).
