# Go Common Utilities & Infrastructure Kit

## Purpose & Scope
`packages/go_common` provides foundational Go packages shared across all 79 backend microservices.

## Package Modules
- `httpserver/`: Standardized HTTP/REST server setup with graceful shutdown and rate limiting.
- `config/`: 12-factor environment configuration loader with secret masking.
- `logger/`: Zero-allocation structured JSON logging (uber-go/zap) with trace ID injection.
- `health/`: Kubernetes liveness, readiness, and startup health check probes.
