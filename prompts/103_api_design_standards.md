# 103 - API Design Standards (REST & gRPC Conventions, Versioning & Error Formats)

## Purpose
Establishes enterprise-grade API design standards and conventions across all synchronous communication channels in the Growww platform. Growww operates a high-throughput, polyglot microservice ecosystem interfacing with client applications across 5 native platforms, web apps, admin consoles, and external depository/banking rails.

This prompt provides developers with the canonical standards for public client-facing RESTful APIs (OpenAPI 3.1) and high-performance internal microservice communication (gRPC / Protocol Buffers v3). It standardizes URI taxonomy, HTTP methods/status codes, RFC 7807 structured error envelopes, cursor-based pagination, request tracing headers, and semantic versioning rules.

## What You Are Building
A comprehensive API standards specification (`docs/standards/api_guidelines.md`), linting configuration suite, and foundational Protobuf definitions:
- REST API Guidelines: Resource naming, HTTP verbs, status code mapping, query parameter conventions, cursor pagination, and OpenAPI 3.1 schema generation.
- gRPC API Guidelines: Protobuf v3 package structure, RPC naming conventions, streaming vs unary patterns, deadline propagation, and metadata forwarding.
- Standard Envelope & Error Format: RFC 7807 compliant JSON error structures and canonical Protobuf status envelopes (`proto/common/v1/error.proto`, `proto/common/v1/pagination.proto`).
- Linting & Governance: Spectral ruleset (`.spectral.yaml`) for OpenAPI linting and Buf configuration (`buf.yaml`, `buf.gen.yaml`) for gRPC validation.

## Scope Boundaries
- **In Scope:**
 - REST & gRPC API design rules, URI schemas, and header specifications.
 - Standard error handling, error codes taxonomy, and localization envelopes.
 - Pagination, filtering, sorting, and field-mask standards.
 - Protobuf package structure, versioning, and breaking change detection rules.
 - API linter configurations (Spectral and Buf CLI).
- **Out of Scope / Handled Elsewhere:**
 - Event schemas and Kafka topic naming (handled in Prompt 104).
 - Authentication tokens, OAuth2 flows, and mTLS configuration (handled in Prompt 105).
 - Business-specific API endpoints and service logic (handled in Category 2 & 5).

## Technology to Use
- **Protocols & Serialization:**
 - *Internal RPCs:* gRPC over HTTP/2 using Protocol Buffers v3 for binary serialization, multiplexing, and microsecond-level serialization efficiency.
 - *Client/External Gateway:* RESTful HTTPS / JSON and WebSocket streaming for real-time market data.
- **Specification & Tooling:**
 - *Buf CLI (v1.30+):* Linting, formatting, breaking change detection, and multi-language stub generation (Python, Go, Rust, Dart, TypeScript).
 - *OpenAPI 3.1 / Swagger:* Automated schema generation via FastAPI (Python) and `protoc-gen-openapiv2` (Go/gRPC).
 - *Spectral CLI:* Automated linting for OpenAPI documentation in CI/CD.
 - *RFC 7807 (Problem Details for HTTP APIs):* Standardized JSON error envelope format.

## Backend / Infra Touchpoints
- **API Gateway (Envoy):** Performs protocol translation (gRPC-JSON transcoding), request ID injection, rate limiting, and CORS enforcement.
- **Service Mesh (Istio):** Routes internal gRPC traffic with load balancing, mTLS, circuit breaking, and distributed tracing.
- **Observability (OpenTelemetry):** Propagates W3C Trace Context (`traceparent`, `tracestate`) across all HTTP and gRPC boundaries.

## Blockchain Interaction
Standardizes API interactions with the permissioned Hyperledger Besu blockchain:
- **JSON-RPC 2.0 Gateway Conventions:** Internal relayer services communicate with Besu nodes via standardized JSON-RPC 2.0 over HTTPS/mTLS.
- **EIP-712 Typed Signing Schemas:** API endpoints accepting investor transaction authorizations define strict EIP-712 structured data schemas, preventing signature tampering and replay attacks.
- **On-Chain Transaction Status Mapping:** On-chain transaction receipts (`tx_hash`, `block_number`, `status`, `revert_reason`) map to standardized API response envelopes (`LedgerTransactionStatusResponse`).

## Step-by-Step Build Instructions
1. Initialize repository directory `docs/standards/` and `proto/common/v1/`.
2. Author `docs/standards/api_guidelines.md` establishing the core REST and gRPC conventions.
3. Define REST URI naming conventions: plural nouns (`/api/v1/orders`), kebab-case URL segments, no verbs in URIs (use standard HTTP methods: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`).
4. Define standard HTTP headers:
 - `X-Request-ID`: Unique UUIDv7 generated at edge for request lifecycle tracking.
 - `X-Correlation-ID`: End-to-end transaction tracing identifier.
 - `Idempotency-Key`: Client-provided UUID for state-changing operations (Prompt 112).
 - `X-Growww-Entity`: Entity designation (`DOMESTIC` or `GIFT_CITY`).
5. Define cursor-based pagination standards using base64-encoded cursor tokens (`cursor`, `limit`, `has_more`, `next_cursor`).
6. Define RFC 7807 compliant error format with standard fields (`type`, `title`, `status`, `detail`, `instance`, `error_code`, `invalid_params`).
7. Create foundational Protobuf definitions under `proto/common/v1/`:
 - `error.proto`: Standard error message with domain-specific error codes.
 - `pagination.proto`: Reusable `PaginationRequest` and `PaginationResponse` messages.
 - `money.proto`: Safe monetary representation (`currency_code`, `units`, `nanos`) preventing floating-point errors.
 - `fractional_share.proto`: Fixed-point 6-decimal representation for fractional equities.
8. Configure `buf.yaml` with strict linting rules (`DEFAULT`, `COMMENTS`, `FILE_LOWER_SNAKE_CASE`, `FIELD_LOWER_SNAKE_CASE`).
9. Configure `buf.gen.yaml` to generate typed client and server stubs for Go, Rust, Python, Dart, and TypeScript.
10. Create `.spectral.yaml` ruleset for linting OpenAPI specifications with rules enforcing camelCase parameters, documented status codes (400, 401, 403, 404, 429, 500), and RFC 7807 error references.
11. Define semantic versioning policy: backwards-compatible additions within major versions (`/v1/`), breaking changes require new major version (`/v2/`) with 180-day deprecation notice.
12. Integrate Buf linting and Spectral checks into the CI pipeline (Prompt 803) to automatically fail PRs with breaking API changes.

## Interfaces / Contracts

### Common Error Protocol Buffer (`proto/common/v1/error.proto`)
```protobuf
syntax = "proto3";

package growww.common.v1;

option go_package = "github.com/growww/proto/gen/go/common/v1;commonv1";
option java_package = "com.growww.proto.common.v1";

enum ErrorCode {
  ERROR_CODE_UNSPECIFIED = 0;
  ERROR_CODE_VALIDATION_FAILED = 1001;
  ERROR_CODE_UNAUTHENTICATED = 1002;
  ERROR_CODE_PERMISSION_DENIED = 1003;
  ERROR_CODE_NOT_FOUND = 1004;
  ERROR_CODE_ALREADY_EXISTS = 1005;
  ERROR_CODE_INSUFFICIENT_FUNDS = 2001;
  ERROR_CODE_KYC_REQUIRED = 2002;
  ERROR_CODE_MARKET_CLOSED = 3001;
  ERROR_CODE_ORDER_REJECTED = 3002;
  ERROR_CODE_CUSTODY_LOCK_FAILED = 4001;
  ERROR_CODE_LEDGER_REVERTED = 5001;
  ERROR_CODE_RATE_LIMITED = 9001;
  ERROR_CODE_INTERNAL_ERROR = 9999;
}

message FieldViolation {
  string field = 1;
  string description = 2;
  string constraint = 3;
}

message ErrorResponse {
  ErrorCode error_code = 1;
  string message = 2;
  string request_id = 3;
  int64 timestamp = 4;
  repeated FieldViolation field_violations = 5;
  map<string, string> metadata = 6;
}
```

### Common Pagination Protocol Buffer (`proto/common/v1/pagination.proto`)
```protobuf
syntax = "proto3";

package growww.common.v1;

option go_package = "github.com/growww/proto/gen/go/common/v1;commonv1";

message PaginationRequest {
  int32 page_size = 1; // Default: 20, Max: 100
  string page_token = 2; // Opaque base64 cursor
}

message PaginationResponse {
  string next_page_token = 1;
  int64 total_count = 2; // Optional, populated when explicitly requested
  bool has_more = 3;
}
```

### RFC 7807 REST JSON Error Envelope Schema
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "ProblemDetails",
  "type": "object",
  "properties": {
    "type": { "type": "string", "format": "uri" },
    "title": { "type": "string" },
    "status": { "type": "integer" },
    "detail": { "type": "string" },
    "instance": { "type": "string" },
    "errorCode": { "type": "string" },
    "requestId": { "type": "string", "format": "uuid" },
    "timestamp": { "type": "string", "format": "date-time" },
    "invalidParams": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "reason": { "type": "string" }
        },
        "required": ["name", "reason"]
      }
    }
  },
  "required": ["type", "title", "status", "detail", "errorCode", "requestId", "timestamp"]
}
```

## Security & Compliance Notes
- **Sanitized Error Output:** Internal stack traces, raw SQL queries, and internal microservice IP addresses must never be returned in API error responses.
- **Request Tracing for Regulatory Audits:** Every incoming API call must bind to an immutable `X-Request-ID` and `X-Correlation-ID` logged across all downstream services and audit trails.
- **Strict Input Validation:** All API inputs must be strictly validated against JSON Schema / Protobuf constraints before processing, rejecting unknown fields and preventing injection attacks.
- **Rate Limiting Response Headers:** All public REST endpoints must return standard rate-limiting headers (`RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset`).

## Acceptance Criteria
- [ ] Complete API guidelines document (`docs/standards/api_guidelines.md`) published and approved.
- [ ] Reusable common Protobuf schemas (`error.proto`, `pagination.proto`, `money.proto`, `fractional_share.proto`) created under `proto/common/v1/`.
- [ ] `buf.yaml` and `buf.gen.yaml` configured and generating valid typed stubs across all 5 languages.
- [ ] Spectral ruleset (`.spectral.yaml`) validating OpenAPI 3.1 specs with zero syntax warnings.
- [ ] RFC 7807 JSON error envelope schema implemented and verified against standard error scenarios.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture Overview), Prompt 102 (Service Boundary Map).
- **Parallel Work:** Prompt 104 (Event Schema Standards), Prompt 105 (Auth Architecture).
- **Blocks:** Prompt 111 (Canonical Domain Model), Prompt 112 (Idempotency), Category 2 (Microservices), Category 5 (Flutter API Client).
