# Observability Stack: Prometheus, Grafana & OpenTelemetry

## Purpose & Scope
The `infra/observability` module manages deployment, networking, and runtime environments for the Growww / NBSE exchange.

## Architectural Responsibilities
Grafana dashboards, Prometheus alerting rules, and Jaeger distributed tracing topologies for microsecond observability.

## Local Testing & Scale Profile
- **Local Workstation**: Supports lightweight execution for developers running on Linux/macOS laptops.
- **Enterprise Scale**: Hardened for multi-AZ high-availability supporting up to 1 Crore (10 Million) concurrent connections.
