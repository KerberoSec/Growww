# 211 - Transactional Notification Service (Go)

## Purpose
The Transactional Notification Service is the primary communication bridge keeping Growww investors informed in real time. It delivers mission-critical transactional alerts - such as OTP authentication challenges, KYC state updates, INR deposit/withdrawal confirmations, order fill notifications, trade settlement receipts, and market volatility warnings - across multiple delivery channels: Mobile Push Notifications (FCM / APNs), SMS (via TRAI DLT-registered Indian telecom gateways), and Email (via AWS SES).

In financial trading, timely and reliable delivery of execution alerts is essential for regulatory compliance (SEBI trade confirmation mandates) and user trust. This service guarantees fast dispatch, automated vendor failover, localized multi-language templating, and strict compliance with telecom and privacy regulations.

## What You Are Building
A high-throughput, asynchronous Go microservice (`services/notification-service`). Concrete deliverables include:
- High-concurrency notification dispatch engine utilizing worker pools for non-blocking multi-channel delivery.
- Multi-provider gateway adapters for SMS (Gupshup, ValueFirst, AWS SNS with automated failover), Push (Firebase Cloud Messaging / Apple Push Notification service), and Email (AWS SES).
- Localized message templating engine supporting English, Hindi, and major Indian regional languages with strict variable substitution.
- Telecom Regulatory Authority of India (TRAI) Distributed Ledger Technology (DLT) template ID mapper for SMS delivery.
- Kafka consumers for domain events (`user.events.v1`, `kyc.events.v1`, `payment.events.v1`, `order.events.v1`, `trade.settled.v1`).
- Notification history repository in PostgreSQL with delivery status tracking and user preference filters.

## Scope Boundaries
- **In Scope:**
 - Transactional and regulatory notification dispatch (OTP, orders, settlements, funds).
 - Multi-channel routing and vendor failover.
 - Template rendering with parameter sanitization and PII masking.
 - User notification preference enforcement (channel opt-ins/opt-outs).
 - Delivery receipt logging and retry with exponential backoff.
- **Out of Scope / Handled Elsewhere:**
 - Bulk promotional marketing campaigns (Prompt 224).
 - Client-side notification center UI rendering (Prompt 513).
 - Generating OTP secrets or validating MPINs (Prompt 201).
 - Trade execution matching (Prompt 205).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+. Selected for its lightweight goroutine worker pools, minimal memory footprint under high concurrency, fast JSON/string templating, and robust HTTP/2 client connections to push/SMS providers.
- **Push & Cloud SDKs:** `firebase.google.com/go/v4` for FCM push; `github.com/aws/aws-sdk-go-v2/service/ses` for email; `github.com/aws/aws-sdk-go-v2/service/sns` for SMS fallback.
- **Database & Storage:** PostgreSQL 16+ using `pgx/v5` and `sqlc` for logging delivery attempts and user preferences; Redis 7.2 for rate limiting and deduplication.
- **Streaming & Messaging:** `segmentio/kafka-go` consuming domain event streams.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `notification_templates`, `notification_logs`, `user_notification_preferences`.
- **Redis 7.2:** Deduplication keys `notif:dedup:{user_id}:{event_hash}` (10-minute TTL) to prevent spamming users with duplicate alerts.
- **Apache Kafka:** Consumes `user.events.v1`, `kyc.events.v1`, `payment.events.v1`, `order.events.v1`, `trade.settled.v1`.
- **Telecom / Cloud Gateways:** FCM, APNs, Gupshup / ValueFirst SMS APIs, AWS SES.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Settlement Confirmation Alerts:** The service consumes events emitted by the Blockchain Indexing Service (Prompt 309) for finalized smart contract events (`SettlementDvPCompleted`, `ProofOfReserveUpdated`).
- **Transparency Receipts:** When an equity settlement confirms on Hyperledger Besu, the service generates and dispatches a trade confirmation message containing the public transaction hash (`tx_hash`) and block number, allowing the investor to verify their 1:1 asset backing on the public explorer.
- **Zero On-Chain PII:** Blockchain events contain only token contract addresses, quantities, and pseudonymous public addresses. The Notification Service securely maps the address to the registered user profile in PostgreSQL to deliver the personalized alert.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/notification-service` with strict linter rules, Makefile, and standard layout (`cmd/`, `internal/dispatcher/`, `internal/providers/`, `internal/templates/`).
2. **Define Protobuf Contracts:** Create `proto/growww/notification/v1/notification_service.proto` defining `SendDirectNotification` (for OTPs) and `UpdateUserPreferences`.
3. **Generate Go Stubs:** Compile protobuf schemas to Go gRPC client and server stubs.
4. **Design PostgreSQL Schema:** Write database migrations for `notification_templates`, `notification_logs`, and `user_notification_preferences`.
5. **Implement Template Engine:** Build template manager supporting localized string formatting (`text/template`) mapped to mandatory TRAI DLT template IDs for domestic SMS compliance.
6. **Implement SMS Provider Adapters:** Build primary adapter for Gupshup/ValueFirst and secondary fallback adapter for AWS SNS with automated failover on HTTP 5xx or timeout ($> 3\text{s}$).
7. **Implement FCM Push Notification Adapter:** Build FCM v1 HTTP/2 client with connection pooling, message batching, and handling of stale device tokens.
8. **Implement Email Adapter:** Build AWS SES client supporting HTML/plaintext multipart emails with DKIM/SPF verification headers.
9. **Implement Redis Deduplication & Rate Limiter:** Enforce maximum 3 OTPs per 5 minutes per user and prevent duplicate notifications for identical event IDs.
10. **Build Asynchronous Worker Pool:** Implement bounded goroutine worker pool (e.g. 50 concurrent dispatch workers) with channel queues to prevent resource exhaustion during market volatility spikes.
11. **Implement Kafka Event Consumers:** Build consumer groups subscribing to `order.events.v1`, `trade.settled.v1`, `payment.events.v1`, translating domain events into localized user notifications.
12. **Implement User Preference Checker:** Filter notifications against user opt-in settings (e.g., allow disabling market alerts while keeping mandatory regulatory/OTP alerts enabled).
13. **Expose gRPC Service for Direct Alerts:** Implement `SendDirectNotification` for high-priority synchronous OTP dispatching.
14. **Configure Telemetry & Observability:** Expose Prometheus metrics for delivery success/failure rates, vendor latency histograms, and channel queue depths.
15. **Write Comprehensive Test Suite:** Implement unit tests with mock SMS/Push gateways, verifying template variable substitutions, vendor failover, and rate-limiting behavior.

## Interfaces / Contracts

### Protobuf Definition (`notification_service.proto`)
```protobuf
syntax = "proto3";

package growww.notification.v1;

option go_package = "growww/notification/v1;notificationv1";

service NotificationService {
  rpc SendDirectNotification (SendDirectNotificationRequest) returns (SendDirectNotificationResponse);
  rpc GetUserPreferences (GetUserPreferencesRequest) returns (GetUserPreferencesResponse);
  rpc UpdateUserPreferences (UpdateUserPreferencesRequest) returns (UpdateUserPreferencesResponse);
}

enum ChannelType {
  CHANNEL_TYPE_UNSPECIFIED = 0;
  CHANNEL_TYPE_PUSH = 1;
  CHANNEL_TYPE_SMS = 2;
  CHANNEL_TYPE_EMAIL = 3;
  CHANNEL_TYPE_IN_APP = 4;
}

enum NotificationPriority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1; // Market updates
  PRIORITY_MEDIUM = 2; // Order updates
  PRIORITY_HIGH = 3; // Trade settlement, funds deposit
  PRIORITY_CRITICAL = 4; // OTP, security alerts
}

message SendDirectNotificationRequest {
  string idempotency_key = 1;
  string user_id = 2;
  ChannelType channel = 3;
  NotificationPriority priority = 4;
  string template_id = 5; // e.g. "AUTH_OTP_LOGIN_V1"
  map<string, string> template_variables = 6;
  string recipient_override = 7; // Optional direct phone/email for pre-registered users
}

message SendDirectNotificationResponse {
  string notification_id = 1;
  bool queued = 2;
  string status = 3; // "SENT", "QUEUED", "SUPPRESSED"
}

message GetUserPreferencesRequest {
  string user_id = 1;
}

message GetUserPreferencesResponse {
  string user_id = 1;
  bool allow_push = 2;
  bool allow_sms = 3;
  bool allow_email = 4;
  bool allow_market_alerts = 5;
}

message UpdateUserPreferencesRequest {
  string user_id = 1;
  bool allow_push = 2;
  bool allow_sms = 3;
  bool allow_email = 4;
  bool allow_market_alerts = 5;
}

message UpdateUserPreferencesResponse {
  bool success = 1;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE delivery_channel_enum AS ENUM ('PUSH', 'SMS', 'EMAIL', 'IN_APP');
CREATE TYPE delivery_status_enum AS ENUM ('QUEUED', 'DISPATCHED', 'DELIVERED', 'FAILED', 'SUPPRESSED');

CREATE TABLE notification_templates (
    template_id VARCHAR(50) PRIMARY KEY, -- e.g., 'TRADE_SETTLED_V1'
    dlt_template_id VARCHAR(50), -- Telecom TRAI DLT ID for SMS compliance
    channel delivery_channel_enum NOT NULL,
    language_code VARCHAR(10) NOT NULL DEFAULT 'en', -- en, hi, gu, mr, ta
    subject_template TEXT, -- Used for Email / Push Title
    body_template TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_notification_preferences (
    user_id UUID PRIMARY KEY,
    allow_push BOOLEAN NOT NULL DEFAULT TRUE,
    allow_sms BOOLEAN NOT NULL DEFAULT TRUE,
    allow_email BOOLEAN NOT NULL DEFAULT TRUE,
    allow_market_alerts BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notification_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    template_id VARCHAR(50) NOT NULL,
    channel delivery_channel_enum NOT NULL,
    provider_name VARCHAR(50) NOT NULL, -- GUPSHUP, AWS_SES, FCM, AWS_SNS
    provider_message_id VARCHAR(255),
    recipient_masked VARCHAR(100) NOT NULL,
    status delivery_status_enum NOT NULL DEFAULT 'QUEUED',
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    dispatched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notif_user_created ON notification_logs(user_id, created_at DESC);
CREATE INDEX idx_notif_status ON notification_logs(status) WHERE status = 'QUEUED';
```

## Security & Compliance Notes
- **TRAI DLT Registration Mandate:** Under Indian telecom regulations, every commercial SMS sent in India must be registered on a DLT platform with a registered Entity ID, Header ID, and Template ID. Unregistered SMS templates are rejected by telecom operators.
- **Data Protection & Privacy (DPDP Act 2023):** Never include full bank account numbers, unmasked PANs, or complete Aadhaar numbers in push notifications or SMS messages. All sensitive identifiers must be masked (e.g. `Account ending in *4012`).
- **OTP Protection:** OTP SMS messages must expire within 5 minutes and must explicitly include fraud prevention disclaimers ("Do not share this OTP with anyone, including Growww staff").

## Acceptance Criteria
- [ ] Notification service successfully connects to FCM, Gupshup/ValueFirst, and AWS SES.
- [ ] OTP dispatch requests are processed and sent to SMS gateway in $< 500\text{ms}$.
- [ ] Primary SMS provider failures automatically failover to secondary provider without dropped messages.
- [ ] Templates strictly adhere to registered TRAI DLT formats and support English and Hindi.
- [ ] All sensitive financial account numbers and phone numbers are masked in logs and notification texts.
- [ ] User notification preference opt-outs are respected for non-critical notifications while mandatory security/OTP alerts are always delivered.
- [ ] Worker pools handle 1,000 notifications/sec without memory leaks or queue stalls.

## Suggested Order / Dependencies
- **Prerequisites:** 104 (Kafka Standards), 201 (User Service), 522 (Flutter Push Notifications).
- **Parallel Tasks:** 204 (Order Service), 208 (Trade Settlement Service).
- **Downstream Blockers:** 513 (Flutter Notification Center UI).
