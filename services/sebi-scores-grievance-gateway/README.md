# sebi-scores-grievance-gateway

## Executive Overview
`sebi-scores-grievance-gateway` is an institutional-grade microservice written in **Go** within the Growww / NBSE trading platform architecture.

## Primary Responsibility
Direct gateway to SEBI SCORES (SEBI Complaints Redress System) with automated SLA escalations.

## Interface & Communication Boundaries
- **Primary gRPC Service**: `SEBISCORESGatewayService` (defined in `packages/proto/`)
- **Inbound Event Stream**: `grievance.tickets.v1`
- **Outbound Event Stream**: `scores.status.v1`
- **Persistence Layer**: Distributed Redis Cluster for in-memory caching and PostgreSQL for ACID durability.
- **Blockchain Touchpoint**: Integrates with the Hyperledger Besu enterprise settlement layer where applicable.

## Local Testing Setup
```bash
# Run unit tests locally
go test ./...

# Run service in local development mode
go run main.go
```

## High-Scale Production Topology
- **Horizontal Autoscaling**: Kubernetes HPA dynamically scales from 3 to 30 replicas based on CPU/memory utilization and Kafka consumer lag.
- **Latency Target**: p99 < 5ms processing time per message.
- **Telemetry**: OpenTelemetry distributed tracing with Prometheus metrics exported on port `:9090`.
- **Zero-Trust**: Mutual TLS (mTLS) with SPIFFE/SPIRE dynamic X.509 certificate authentication.
