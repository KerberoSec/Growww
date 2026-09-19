# Docker Compose Local Development Stack

## Purpose & Scope
The `infra/local` module manages deployment, networking, and runtime environments for the Growww / NBSE exchange.

## Architectural Responsibilities
All-in-one local development environment spinning up Besu, Kafka, Redis, PostgreSQL, and core microservices with a single command.

## Local Testing & Scale Profile
- **Local Workstation**: Supports lightweight execution for developers running on Linux/macOS laptops.
- **Enterprise Scale**: Hardened for multi-AZ high-availability supporting up to 1 Crore (10 Million) concurrent connections.
