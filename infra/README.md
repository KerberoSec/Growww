# Infrastructure, DevOps & Blockchain Orchestration

## Executive Overview
The `infra/` directory contains the complete Infrastructure-as-Code (IaC), container orchestration, and blockchain configuration required to run the Growww / NBSE platform both locally on developer workstations and at massive scale on multi-region cloud infrastructure.

## Infrastructure Domains
1. `blockchain/`: Hyperledger Besu QBFT genesis ceremony, validator node configurations, and peer discovery.
2. `k8s/`: Base Kubernetes manifests, Helm charts, and environment overlays (Local, Testnet, Mainnet).
3. `terraform/`: Multi-region AWS infrastructure (`ap-south-1` Mumbai and `ap-south-2` Hyderabad) and GIFT City PoP.
4. `observability/`: Prometheus alert rules, Grafana dashboards, and OpenTelemetry collector pipelines.
5. `local/`: Docker Compose definitions for running the entire exchange stack locally with a single command.
