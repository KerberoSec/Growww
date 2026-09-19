# 224 - Referral & Growth Service (SEBI Advertising Code Compliant)

## Purpose
The Referral & Growth Service powers compliant customer acquisition, referral link generation, affiliate attribution, and promotional incentive distribution across Growww. Operating in the heavily regulated Indian capital markets requires strict adherence to SEBI regulations regarding marketing practices, intermediary advertising codes, and anti-fraud mandates. SEBI explicitly prohibits stockbrokers and investment platforms from offering guaranteed returns, cash kickbacks, or volume-based trading rebates to unregistered influencers.

This service manages compliant referral programs by awarding strictly non-monetary incentives (such as platform fee discount vouchers, educational course access, and reduced profit-fee percentage tiers for limited periods). It incorporates advanced fraud prevention and anti-sybil detection to eliminate self-referrals and synthetic identity rings while maintaining a comprehensive regulatory audit trail.

## What You Are Building
A lightweight, secure Go microservice (`services/growth-service`) delivering:
- **Compliant Referral Link & Attribution Manager:** Generates unique, tamper-resistant referral tokens and tracks multi-touch client attribution.
- **Anti-Sybil & Self-Referral Prevention Engine:** Analyzes IP clusters, device fingerprints, PAN/Bank linkages, and KYC biometric hashes to detect referral fraud.
- **Non-Monetary Reward Distributor:** Manages fee discount credits, educational rewards, and platform tier benefits.
- **SEBI Advertising Compliance Audit Logger:** Automatically logs all promotional campaigns, referral terms, and marketing claims for regulatory inspection.
- **Artifacts Delivered:**
 - `services/growth-service/cmd/server/main.go` - Go microservice entry point.
 - `services/growth-service/internal/referral/attribution.go` - Attribution and click tracking engine.
 - `services/growth-service/internal/antifraud/sybil_detector.go` - Anti-sybil and self-referral heuristics.
 - `services/growth-service/internal/rewards/voucher_engine.go` - Fee voucher credit manager.
 - `proto/growww/growth/v1/growth.proto` - Internal gRPC service contracts.

## Scope Boundaries
- **In Scope:**
 - Dynamic referral code and deep-link generation.
 - Tracking conversion funnels (App Install $\rightarrow$ Registration $\rightarrow$ KYC Verification $\rightarrow$ First Settlement).
 - Graph-based detection of collusion and self-referral rings.
 - Granting fee discount vouchers applied by Fee Engine (Prompt 210).
- **Out of Scope / Handled Elsewhere:**
  - Direct calculation and deduction of Universal Zero-Fee Model (0.00% fee - No fee at all)s (handled by Prompt 210).
  - User identity authentication and device registration (handled by Prompt 201).
 - User push notification dispatch (handled by Prompt 211).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ utilizing `pgx/v5`, `go-redis/v9`, and `confluent-kafka-go`.
- **Justification:** Go offers low-latency execution and efficient concurrency needed to handle viral link clicks, attribution callbacks, and real-time fraud graph checks without adding overhead to core user registration flows.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for storing referral programs, links, attributions, and reward voucher ledgers.
 - Redis 7.2+ for click deduplication, rate limiting, and attribution session caching.
 - Apache Kafka 3.7+ for consuming user lifecycle events and emitting reward grants.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `referral_programs`, `referral_codes`, `referral_conversions`, `fee_discount_vouchers`.
- **Redis 7.2:** Attribution click caches (30-day attribution window) and IP rate limits.
- **Kafka Topics:**
 - Subscribes: `user.registered`, `kyc.verified`, `trade.settled`.
 - Publishes: `growth.reward.granted`, `growth.fraud.flagged`.
- **Notification Service (Prompt 211):** Triggers user notifications upon successful referral milestone completion.
- **Fee Engine (Prompt 210):** Validates and consumes fee discount vouchers during trade fee settlement.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network running QBFT consensus.
- **Zero On-Chain PII & Off-Chain Incentive Management:** In strict accordance with privacy and regulatory principles, referral links, marketing campaigns, and growth rewards operate entirely off-chain.
- **Indirect Blockchain Interaction:** When a referred user executes a trade, any active fee discount voucher is applied off-chain by the Fee Engine (Prompt 210) to reduce the platform fee percentage before the final net profit settlement transaction is submitted to `SettlementDvP.sol`.
- **Zero On-Chain Pollution:** No referral tokens, MLM points, or speculative reward tokens are minted on the permissioned ledger.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/growth-service` following clean architecture guidelines.
2. Define Protobuf definitions in `proto/growww/growth/v1/growth.proto` and generate Go gRPC stubs.
3. Configure PostgreSQL schema migrations for `referral_campaigns`, `referral_codes`, `attributions`, and `vouchers`.
4. Implement the Referral Code Generator producing short, collision-free, base62 alphanumeric codes.
5. Build the Web & Deep-Link Click Handler recording attribution nonces in Redis with 30-day expiration windows.
6. Implement the Kafka consumer ingesting `user.registered` and `kyc.verified` events to match attribution cookies.
7. Build the Anti-Sybil Fraud Engine analyzing device IDs, bank account IFSC/account hashes, PAN hashes, and IP subnets to block self-referrals.
8. Implement the Non-Monetary Voucher Engine issuing 30-day fee discount passes upon verified referee onboarding.
9. Implement the SEBI Compliance Audit logger ensuring all active campaigns have verified disclaimers and zero guaranteed return claims.
10. Integrate with Notification Service (Prompt 211) to alert referrers when milestone conditions are met.
11. Add Prometheus metrics (`referral_clicks_total`, `referral_conversions_total`, `sybil_fraud_blocked_total`).
12. Write unit and integration tests verifying attribution accuracy, fraud graph detection, voucher expiration, and idempotency.

## Interfaces / Contracts

### Protobuf Definition (`growth.proto`)
```protobuf
syntax = "proto3";

package growww.growth.v1;

option go_package = "github.com/growww/services/growth-service/gen/v1;growthv1";

service ReferralGrowthService {
  rpc GenerateReferralCode (GenerateCodeRequest) returns (GenerateCodeResponse);
  rpc TrackClick (TrackClickRequest) returns (TrackClickResponse);
  rpc AttributeUser (AttributeUserRequest) returns (AttributeUserResponse);
  rpc GetUserRewardSummary (UserRewardSummaryRequest) returns (UserRewardSummaryResponse);
  rpc ValidateFeeVoucher (ValidateVoucherRequest) returns (ValidateVoucherResponse);
}

message GenerateCodeRequest {
  string user_id = 1;
}

message GenerateCodeResponse {
  string referral_code = 1;
  string referral_deep_link = 2;
  string share_text = 3; // Pre-approved compliant sharing text
}

message TrackClickRequest {
  string referral_code = 1;
  string click_ip = 2;
  string user_agent = 3;
  string device_fingerprint = 4;
}

message TrackClickResponse {
  string attribution_token = 1;
  bool is_valid = 2;
}

message AttributeUserRequest {
  string user_id = 1;
  string attribution_token = 2;
}

message AttributeUserResponse {
  bool attributed = 1;
  string referrer_user_id = 2;
  string campaign_id = 3;
}

message UserRewardSummaryRequest {
  string user_id = 1;
}

message UserRewardSummaryResponse {
  string user_id = 1;
  int64 total_successful_referrals = 2;
  int64 active_fee_vouchers_count = 3;
  repeated FeeVoucherItem active_vouchers = 4;
}

message FeeVoucherItem {
  string voucher_id = 1;
  string discount_percentage = 2; // e.g. "50%" off platform fee
  int64 max_discount_inr = 3;
  int64 expires_at = 4;
}

message ValidateVoucherRequest {
  string user_id = 1;
  string trade_realized_profit_inr = 2;
}

message ValidateVoucherResponse {
  bool is_valid = 1;
  string voucher_id = 2;
  string discount_amount_inr = 3;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE referral_campaigns (
    campaign_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_name VARCHAR(128) NOT NULL,
    reward_type VARCHAR(32) NOT NULL CHECK (reward_type IN ('FEE_DISCOUNT_VOUCHER', 'EDUCATIONAL_ACCESS')),
    reward_config JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sebi_compliance_notes TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE referral_codes (
    code_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) UNIQUE NOT NULL,
    referral_code VARCHAR(16) UNIQUE NOT NULL,
    total_clicks BIGINT NOT NULL DEFAULT 0,
    total_conversions INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE referral_conversions (
    conversion_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    referrer_id VARCHAR(64) NOT NULL,
    referee_id VARCHAR(64) UNIQUE NOT NULL,
    campaign_id UUID NOT NULL REFERENCES referral_campaigns(campaign_id),
    status VARCHAR(32) NOT NULL DEFAULT 'REGISTERED',
    is_fraud_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    fraud_reason VARCHAR(128),
    reward_issued_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fee_discount_vouchers (
    voucher_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) NOT NULL,
    discount_percentage NUMERIC(5, 2) NOT NULL,
    max_discount_inr NUMERIC(18, 4) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Security & Compliance Notes
- **SEBI Advertising Code Compliance:** All referral sharing copy is strictly templated and system-controlled. Users cannot alter promotional text to make prohibited claims regarding guaranteed investment returns or trading tips.
- **Anti-Sybil Defense:** Strict heuristics analyze shared IP addresses, identical device identifiers, matching bank account details, and mutual referral loops to instantly neutralize fraud rings.
- **Zero Cash Rebates:** In compliance with SEBI intermediary rules, rewards are strictly restricted to service fee discounts and financial educational content; no direct cash payouts or crypto token incentives are permitted.
- **DPDP Act Compliance:** Referrers receive only masked progress notices (e.g. "A friend completed KYC") without revealing full names, email addresses, or financial data of referees.

## Acceptance Criteria
- [ ] Go growth service builds cleanly with zero linter warnings.
- [ ] Referral code generator produces unique base62 codes with collision detection.
- [ ] Multi-touch attribution tracks clicks through 30-day Redis window to completed KYC.
- [ ] Anti-sybil heuristics flag and block simulated self-referrals (matching device/bank hashes).
- [ ] Fee discount vouchers correctly integrate with Fee Engine validation interface.
- [ ] Generated referral messages strictly comply with SEBI advertising restrictions.
- [ ] Test coverage exceeds >=85%.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design), Prompt 201 (User Service), Prompt 202 (KYC Service), Prompt 210 (Fee Engine).
- **Subsequent / Parallel Tasks:** Prompt 211 (Notification Service), Prompt 512 (Flutter Client UI).
