# Hyperledger Besu Enterprise Consortium Cluster

## Purpose & Scope
The `infra/blockchain` module manages deployment, networking, and runtime environments for the Growww / NBSE exchange.

## Architectural Responsibilities
Configurations, QBFT genesis templates, and validator node deployment scripts for the private settlement blockchain.

## Local Testing & Scale Profile
- **Local Workstation**: Supports lightweight execution for developers running on Linux/macOS laptops.
- **Enterprise Scale**: Hardened for multi-AZ high-availability supporting up to 1 Crore (10 Million) concurrent connections.
