# Kubernetes Production Manifests & Helm Charts

## Purpose & Scope
The `infra/k8s` module manages deployment, networking, and runtime environments for the Growww / NBSE exchange.

## Architectural Responsibilities
Base infrastructure manifests, Envoy API Gateway routes, and environment overlays for local, testnet, and mainnet.

## Local Testing & Scale Profile
- **Local Workstation**: Supports lightweight execution for developers running on Linux/macOS laptops.
- **Enterprise Scale**: Hardened for multi-AZ high-availability supporting up to 1 Crore (10 Million) concurrent connections.
