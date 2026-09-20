# Growww Platform API Design Standards & Guidelines

## 1. Scope & Architecture Principles
This document defines the canonical API design standards for synchronous communications across the Growww / NBSE institutional trading platform.

### Architectural Principles:
1. **External Gateway (RESTful HTTP/JSON & WebSocket):** High developer ergonomic standards, OpenAPI 3.1 compliant schemas, RFC 7807 structured error responses.
2. **Internal Inter-Service (gRPC over HTTP/2):** Protocol Buffers v3 binary encoding, microsecond serialization, deadline propagation, and streaming multiplexing.
3. **Ledger Interface (Hyperledger Besu JSON-RPC 2.0):** EIP-712 structured typed data signing for non-repudiation and zero-leakage auditable execution.

---

## 2. REST API Standards

### 2.1 URI Naming Taxonomy
- Resource paths must use lowercase kebab-case plural nouns:
  - `/api/v1/orders`
  - `/api/v1/trade-executions`
  - `/api/v1/wallet-balances`
- Never use verbs in URIs. Actions are represented strictly by HTTP verbs:
  - `GET /api/v1/orders` (List orders)
  - `POST /api/v1/orders` (Submit order)
  - `GET /api/v1/orders/{order_id}` (Retrieve single order)
  - `DELETE /api/v1/orders/{order_id}` (Cancel order)

### 2.2 Standard HTTP Request Headers
| Header | Type | Description |
| :--- | :--- | :--- |
| `X-Request-ID` | UUIDv7 | Unique request identifier generated at edge gateway for end-to-end tracing. |
| `X-Correlation-ID` | UUIDv7 | Root transaction tracing identifier propagated across microservices. |
| `Idempotency-Key` | UUIDv4 | Client-provided token ensuring exactly-once execution for mutating requests (`POST`, `PUT`, `DELETE`). |
| `X-Growww-Entity` | String | Jurisdiction entity routing: `DOMESTIC` (SEBI/India) or `GIFT_CITY` (IFSCA/Offshore). |
| `Authorization` | String | `Bearer <JWT>` session token signed with RS256 / Ed25519. |

### 2.3 Cursor-Based Pagination
All collections returning lists of records must support opaque base64 cursor pagination:
```json
{
  "data": [ ... ],
  "pagination": {
    "limit": 50,
    "has_more": true,
    "next_cursor": "ZXlKaGJHY2lPaUpTVXpVeE1p..."
  }
}
```

### 2.4 RFC 7807 Error Envelope Format
All error responses (`4xx` and `5xx`) return standard `application/problem+json` envelopes:
```json
{
  "type": "https://api.growww.trade/errors/insufficient-margin",
  "title": "Insufficient Margin Available",
  "status": 422,
  "detail": "Account available margin 45,000 USDT is less than initial margin requirement 60,000 USDT.",
  "instance": "/api/v1/orders/ord_98741",
  "error_code": "ERROR_CODE_INSUFFICIENT_FUNDS",
  "request_id": "018d9f45-728b-7000-848f-39589dfd0001",
  "timestamp": 1774088000,
  "invalid_params": [
    {
      "name": "amount_e8",
      "reason": "Exceeds maximum allowable leverage position"
    }
  ]
}
```

---

## 3. gRPC & Protocol Buffers Standards

### 3.1 Package Hierarchy & Versioning
- All protobuf definitions follow the hierarchy:
  - `growww.<domain>.<subdomain>.v<major>`
  - Example: `growww.order.matching.v1`, `growww.common.v1`
- Breaking changes require incrementing the major version (`v1` -> `v2`).

### 3.2 Canonical gRPC Error Codes Mapping
| HTTP Status | gRPC Code | Platform Error Code |
| :--- | :--- | :--- |
| `400 Bad Request` | `INVALID_ARGUMENT` | `ERROR_CODE_VALIDATION_FAILED` |
| `401 Unauthorized` | `UNAUTHENTICATED` | `ERROR_CODE_UNAUTHENTICATED` |
| `403 Forbidden` | `PERMISSION_DENIED` | `ERROR_CODE_PERMISSION_DENIED` |
| `404 Not Found` | `NOT_FOUND` | `ERROR_CODE_NOT_FOUND` |
| `409 Conflict` | `ALREADY_EXISTS` | `ERROR_CODE_ALREADY_EXISTS` |
| `429 Too Many Requests` | `RESOURCE_EXHAUSTED` | `ERROR_CODE_RATE_LIMITED` |
| `500 Internal Server Error` | `INTERNAL` | `ERROR_CODE_INTERNAL_ERROR` |

---

## 4. Hyperledger Besu JSON-RPC & EIP-712 Compliance
- Smart contract invocations for trading, settlement, and launchpads require EIP-712 domain separation:
  - `EIP712Domain(string name, string version, uint256 chainId, address verifyingContract)`
- Ensures zero signature replay across Testnet and Mainnet environments.
