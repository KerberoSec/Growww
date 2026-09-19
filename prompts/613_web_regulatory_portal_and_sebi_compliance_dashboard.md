# 613 - Next.js 14 Regulatory Audit & SEBI/IFSCA Supervisory Dashboard

## Purpose
Operating an equity and tokenized financial market infrastructure under the jurisdiction of the Securities and Exchange Board of India (SEBI) and the International Financial Services Centres Authority (IFSCA / GIFT City) requires an independent, tamper-proof supervisory gateway. Market surveillance officers, statutory auditors, and designated enforcement authorities cannot rely on standard operational back-office dashboards or third-party reporting delays when investigating market manipulation, sudden volatility cascades, or systemic risk events. Under the SEBI Integrated Market Surveillance System (IMSS) regulations, SEBI Master Circular for Stock Exchanges, and IFSCA Market Infrastructure Institutions (MII) Regulations 2021, regulatory bodies demand direct, real-time supervisory access to inspect order books down to nanosecond timestamps, conduct forensic trade reconstructions, inspect automated volatility circuit halts, and issue instant judicial freeze mandates.

This prompt specifies the architecture, engineering design, and implementation of the **Next.js 14 Regulatory Audit & SEBI/IFSCA Supervisory Dashboard** (`apps/growww_admin/regulator`). The platform serves as an air-gapped, high-assurance web portal granting regulatory officers read-only real-time market depth telemetry, historical tick-by-tick order book replay with sub-millisecond scrubbing, dynamic circuit breaker halt diagnostics, and cryptographic verification of compliance bhavcopies against the Hyperledger Besu consortium blockchain. Furthermore, it incorporates a dual-officer maker-checker judicial enforcement terminal to execute instant account freezes, trading suspensions, and on-chain contract blacklisting, all while strictly upholding the Digital Personal Data Protection (DPDP) Act 2023 through zero PII decryption without authenticated judicial search warrants.

## What You Are Building
A high-security, auditable Next.js 14 regulatory web application (`apps/growww_admin/regulator`) featuring:
- `RegulatoryDashboardShell`: Air-gapped supervisory portal with mandatory Mutual TLS (mTLS) client certificate verification, hardware-token FIDO2/WebAuthn multi-factor authentication, active regulatory jurisdiction toggle (SEBI Domestic vs IFSCA International), and dynamic forensic screen watermarking (displaying officer badge ID, workstation IP, session nonce, and timestamp).
- `LiveOrderBookInspector`: High-resolution, real-time Level-2 and Level-3 order book visualizer streaming aggregated market depth, resting order queues, top-of-book spread dynamics, and participant attribution using anonymized entity codes with nanosecond sequencing.
- `TradeReconstructionPlayer`: Tick-by-tick forensic trade sequence player featuring an interactive scrubbing timeline (play, pause, step forward/backward, slow motion 0.1x to 10x), capable of reconstructing the exact state of the order book and tape at any historical microsecond.
- `CircuitHaltAnalyzer`: Volatility investigation terminal analyzing Limit-Up / Limit-Down (LULD) dynamic corridor breaches, index-wide circuit breaker halts, pre-halt quote stuffing volumes, order cancellation cascades, and call auction equilibrium price discovery curves.
- `SurveillanceAlertDesk`: Triaged supervisory alert inbox ingesting real-time algorithmic market abuse alerts from the Surveillance Engine (Prompt 228), including spoofing, layering, wash trading, circular trading networks, and Order-to-Trade Ratio (OTR) collar violations with anomaly heatmaps and participant clustering graphs.
- `ComplianceBhavcopyExporter`: Cryptographic export engine generating official compliance bhavcopies, daily closing settlement records, and audited trade blotters in XML, XBRL, and CSV formats, complete with SHA-256 integrity checksums and Hyperledger Besu block receipt proofs.
- `JudicialFreezeTerminal`: Statutory enforcement console allowing authorized regulatory officers to issue instant participant trading freezes, symbol suspensions, and on-chain smart contract transfer blocks (`ComplianceRegistry.sol`), enforced via dual-officer maker-checker approvals and WebAuthn hardware signatures.
- `AuditLogVerifyingInspector`: Interactive cryptographic audit verifier querying the Immutable Audit Log Service (Prompt 218) and Hyperledger Besu ledger nodes to mathematically validate RFC 6962 Merkle inclusion proofs, continuous hash chains, and on-chain hourly epoch roots.
- `WarrantBackedPiiVault`: DPDP Act 2023 compliant zero-knowledge identity unmasking terminal that enforces zero PII exposure by default, allowing decrypted inspection of beneficial owner details (PAN, Demat, KYC records) exclusively upon cryptographic verification of a valid judicial search warrant.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router application structure located at `apps/growww_admin/regulator`.
  - Mutual TLS (mTLS) client certificate validation and FIDO2/WebAuthn hardware token MFA gate.
  - Live Level-2 and Level-3 order book inspection with nanosecond timestamp sequencing.
  - Interactive historical trade reconstruction scrubber with dynamic book replay.
  - Circuit halt and LULD volatility corridor investigation panel with call auction price discovery graphs.
  - Real-time surveillance alert inbox with OTR calculation displays and participant cluster graphs.
  - Maker-checker judicial freeze terminal issuing instant freeze commands to matching engine and smart contracts.
  - Cryptographic verification tool validating off-chain audit logs against Hyperledger Besu on-chain Merkle roots.
  - Compliance bhavcopy exporter generating cryptographically sealed files (XML, XBRL, CSV) with SHA-256 manifests.
  - Warrant-backed PII decryption module requiring digital court warrant upload and dual approval.
  - Dynamic forensic watermarking on all supervisory views to deter unauthorized external screen capture or leakage.
  - Comprehensive immutable access logging recording every click, query, scrub, and export.
- **Out of Scope / Handled Elsewhere:**
  - Real-time algorithmic market abuse pattern detection algorithms (handled in Go by Surveillance Engine - Prompt 228).
  - High-throughput continuous automated statutory reporting generation and dispatch pipelines (handled in Python/Go by Prompt 233).
  - Internal compliance officer daily operational admin reporting (handled in Prompt 607).
  - Low-latency order matching engine and write-ahead log execution (handled in Prompt 205).
  - Clearing member capital adequacy and base minimum capital tracking (handled in Prompt 614).
  - Smart contract consensus deployment and core ledger node operations (handled in Prompt 302, 305, 306).

## Technology to Use
- **Next.js 14 App Router (React Server Components + strict 'use client' supervisory interactive components):** Provides a secure, server-rendered application shell with strict isolation between administrative back-office and supervisory regulatory routes.
- **TypeScript 5.4+:** Enforces complete type safety across regulatory schemas, Protobuf contracts, audit receipts, and market depth definitions.
- **Tailwind CSS 3.4+ & shadcn/ui:** Institutional dark-mode design system with high-contrast data visualization, accessible modals, and dense tabular financial layouts.
- **TanStack Table (v8) & TanStack Virtual:** High-performance virtualized tables rendering 100,000+ trade ticks and order events at 60 FPS without browser memory degradation or DOM thrashing.
- **Recharts & Lightweight Charts:** High-performance charting libraries visualizing LULD volatility corridors, depth charts, volume histograms, and call auction equilibrium curves.
- **WebSockets (`native ws` or `@stomp/stompjs`):** Low-latency full-duplex communication streaming live Level-2/Level-3 book deltas, alert broadcasts, and circuit halt status transitions.
- **WebAuthn / FIDO2 API:** Hardware token MFA authentication (YubiKey 5 Series, Nitrokey) required for session bootstrap and dual-officer judicial freeze authorizations.
- **Viem 2.x:** High-assurance Web3 client querying Hyperledger Besu permissioned RPC nodes for block headers, transaction receipts, and smart contract state reads.
- **Web Crypto API:** In-browser cryptographic verification of SHA-256 hash chains, Merkle inclusion proofs, and digital signature attestation.
- **jszip & file-saver:** Client-side generation and verification of signed compliance bhavcopy packages containing SHA-256 manifests and on-chain block receipts.
- **mTLS Client Certificates:** Enforced via reverse proxy (Nginx / Envoy) and Next.js middleware, requiring regulatory officers to present government-issued X.509 client certificates.

## Backend / Infra Touchpoints
- **Continuous Regulatory Reporting Service (Prompt 233):** Interfaces via REST (`/api/v1/regulatory/telemetry`) to query real-time statutory filing status, automated transmission acknowledgments, and daily compliance bhavcopy datasets.
- **Real-Time Market Surveillance Engine (Prompt 228):** Subscribes via WebSocket (`wss://surveillance.growww.in/ws/v1/alerts`) and gRPC-Web to receive live market abuse alerts, OTR threshold breaches, dynamic price band status, and circuit breaker trip events.
- **Immutable Audit Log Service (Prompt 218):** Interfaces via gRPC and REST (`https://audit.growww.in/api/v1/audit/verify`) to fetch historical audit event trees, RFC 6962 Merkle inclusion proofs, and on-chain epoch commitment hashes.
- **Order Matching Engine (Prompt 205):** Dispatches emergency symbol trading halts, resume commands, and call auction initiation signals via dedicated supervisory gRPC control plane.
- **Order Service (Prompt 204):** Dispatches instant participant account freeze and active order cancellation commands (`POST /api/v1/internal/judicial/freeze`).
- **Hyperledger Besu Consortium Nodes (Prompt 302):** Directly connects via JSON-RPC (`https://rpc.besu.growww.in`) to read state from `ComplianceRegistry.sol`, `SettlementDvP.sol`, and `AuditAnchorRegistry.sol`.
- **Ingress Gateway & Reverse Proxy:** Enforces mutual TLS (mTLS) with Certificate Revocation List (CRL) validation and OCSP stapling for all regulatory endpoints.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu permissioned consortium network operating QBFT (Quorum Byzantine Fault Tolerant) consensus with deterministic 2-second block intervals and 1:1 asset backing.
- **Direct RPC Queries:**
  - `ComplianceRegistry.sol`: The dashboard calls `isAccountFrozen(address target) external view returns (bool)` and executes judicial freeze transactions `freezeParticipant(address target, bytes32 judicialWarrantHash, bytes officerSignature)` via authorized regulatory relayer.
  - `SettlementDvP.sol`: Inspects atomic Delivery-versus-Payment trade settlement receipts, block finality timestamps, and clearing token transfers for any selected execution ID.
  - `AuditAnchorRegistry.sol`: Queries hourly epoch commitments via `getEpochRoot(uint256 epochId) external view returns (bytes32 merkleRoot, uint64 eventCount, uint256 blockTimestamp)` to verify that off-chain audit records match the immutable ledger state byte-for-byte.
  - `TransferComplianceHooks.sol`: Inspects exchange-wide emergency pause flags, circuit breaker states, and jurisdiction-specific regulatory rulesets.
- **Cryptographic Proof Verification:** The dashboard downloads Merkle inclusion proofs from the Audit Log Service (Prompt 218) and independently hashes the leaf node across the proof siblings in WebAssembly / Web Crypto, verifying equality against the on-chain root stored in `AuditAnchorRegistry.sol`.
- **Zero PII on Ledger:** All personal identification details (PAN, Aadhaar Vault references, Passport numbers, bank account numbers) remain strictly off-chain. On-chain accounts utilize cryptographic participant hashes and whitelisted Ethereum addresses (`0x...`).

## Step-by-Step Build Instructions

1. **Scaffold Regulatory Portal Route Structure & Layout:**
   - Create route directory `apps/growww_admin/regulator/` with root layout `layout.tsx`, page router `page.tsx`, and loading boundary `loading.tsx`.
   - Configure sub-routes:
     - `order-book/page.tsx`: Live Level-2 and Level-3 order book depth inspector.
     - `trade-replay/page.tsx`: Historical tick-by-tick trade reconstruction scrubber.
     - `circuit-halts/page.tsx`: Volatility corridor and circuit breaker forensic analyzer.
     - `surveillance-alerts/page.tsx`: Market abuse alerts, OTR monitors, and entity clustering.
     - `bhavcopy-export/page.tsx`: Cryptographic compliance bhavcopy generator and exporter.
     - `judicial-freeze/page.tsx`: Maker-checker participant freeze and symbol halt terminal.
     - `audit-verifier/page.tsx`: On-chain Merkle proof and audit log cryptographic inspector.
     - `warrant-vault/page.tsx`: Warrant-backed PII decryption and KYC unmasking console.

2. **Implement Mutual TLS (mTLS) & WebAuthn Regulatory Authentication Gate:**
   - Implement Next.js edge middleware (`middleware.ts`) inspecting `x-client-cert-sha256`, `x-client-cert-dn`, and `x-client-cert-issuer` forwarded by the reverse proxy.
   - Restrict access to approved SEBI and IFSCA root Certificate Authorities (CAs).
   - Build a WebAuthn / FIDO2 authentication challenge component (`WebAuthnGate.tsx`) demanding a hardware token touch (YubiKey) before initializing the supervisory session.
   - Enforce an automatic 15-minute inactivity screen lock with hardware token re-authentication.

3. **Construct Dynamic Forensic Watermarking System (`RegulatoryWatermark.tsx`):**
   - Implement a non-intrusive, tamper-resistant Canvas and SVG background overlay rendering tiled micro-watermarks across all supervisory viewports.
   - Watermark tiles display: Regulatory Authority (SEBI / IFSCA), Officer Badge ID, Workstation Public/Private IP, Current UTC/IST Timestamp, and Ephemeral Session Hash.
   - Implement CSS protection rules (`user-select: none; pointer-events: none;`) and DOM mutation observers that automatically blank the screen if the watermark DOM element is tampered with or hidden via browser developer tools.

4. **Build Live Order Book Depth Inspector (`LiveOrderBookInspector.tsx`):**
   - Connect to Market Data Service (Prompt 207) via WebSocket (`wss://marketdata.growww.in/ws/v1/l2-l3`).
   - Implement real-time Level-2 aggregated depth ladder (Bids/Asks) and Level-3 individual resting order queue.
   - Display nanosecond order arrival sequence, order size, limit price, time-in-force, and anonymized participant code (e.g., `MEMBER-7821-PROP`).
   - Implement depth visualizer using Recharts area charts displaying cumulative volume curves and bid-ask spread ribbons.

5. **Construct Historical Trade Reconstruction Scrubber (`TradeReconstructionPlayer.tsx`):**
   - Build an interactive timeline scrubber component powered by an event-driven player state machine.
   - Load historical order matching events from ClickHouse / Matching Engine replay API (`/api/v1/surveillance/replay`).
   - Implement playback controls: Play, Pause, Step Forward (1 tick), Step Backward (1 tick), Jump to Timestamp, and Playback Speed selector (0.1x, 0.5x, 1x, 2x, 5x, 10x).
   - Dynamically recompute and render the exact state of the order book and the time-and-sales tape at the active scrub position.

6. **Implement Circuit Breaker & LULD Forensic Analyzer (`CircuitHaltAnalyzer.tsx`):**
   - Connect to Surveillance Engine (Prompt 228) and Risk Service (Prompt 206) to ingest historical volatility halt logs.
   - Render multi-axis price chart with historical trade prints overlaid against:
     - Static Daily Price Bands (+/- 10%, +/- 20%).
     - Rolling 5-Minute LULD Dynamic Corridors (+/- 3%, +/- 5%).
   - Highlight exact trigger points where trades breached dynamic corridors, displaying pre-halt quote stuffing volumes, order cancellation ratios, and matching engine halt ack latencies.
   - Visualize the subsequent Call Auction phase, displaying order accumulation, indicative equilibrium price, and uncrossing volume.

7. **Build Surveillance Alert Desk & OTR Anomaly Heatmap (`SurveillanceAlertDesk.tsx`):**
   - Construct a virtualized data table using TanStack Table displaying alerts emitted by Prompt 228: `SPOOFING`, `LAYERING`, `WASH_TRADING`, `CIRCULAR_TRADING`, `MOMENTUM_IGNITION`, and `OTR_VIOLATION`.
   - Implement interactive detail modal showing order cancellation speeds (<200ms), contra-side execution fills, and participant trade graphs.
   - Build an OTR (Order-to-Trade Ratio) monitor component displaying rolling hourly participant OTR metrics:
     $$\text{OTR} = \frac{\text{Submitted} + \text{Modified} + \text{Cancelled}}{\text{Executed Trades} + 1}$$
     Highlighting participants exceeding SEBI regulatory thresholds (OTR > 100:1).

8. **Implement Maker-Checker Judicial Freeze Terminal (`JudicialFreezeTerminal.tsx`):**
   - Construct a multi-step statutory enforcement workflow:
     - **Initiation (Maker):** Officer inputs target Participant Code / Wallet Address / Symbol, selects Freeze Scope (`FULL_ACCOUNT_FREEZE`, `ORDER_SUBMISSION_BLOCK`, `COLLATERAL_HOLD`, `SYMBOL_TRADING_HALT`), enters Statutory Mandate Reference (e.g., SEBI Order Ref / Court Case ID), and attaches the signed judicial warrant PDF.
     - **Verification (Checker):** A second supervisory officer reviews the dossier, validates warrant authenticity, and co-signs using their WebAuthn hardware token.
   - Upon dual-approval, execute atomic dispatch:
     - Send gRPC freeze command to Order Matching Engine (Prompt 205) to cancel all resting orders immediately.
     - Call `freezeParticipant` on `ComplianceRegistry.sol` via Hyperledger Besu relayer.
     - Emit high-priority audit event to Audit Log Service (Prompt 218).

9. **Develop On-Chain Cryptographic Audit Verifier (`AuditLogVerifier.tsx`):**
   - Provide an interactive verification terminal where officers input an Audit Event ID, Transaction Hash, or Reporting Batch ID.
   - Fetch the off-chain structured audit event from Audit Log Service (Prompt 218) and the corresponding on-chain Merkle root from `AuditAnchorRegistry.sol` using Viem.
   - Recompute the leaf hash and climb the RFC 6962 Merkle proof tree client-side using Web Crypto API.
   - Display a step-by-step cryptographic verification certificate with visual status indicators: Leaf Hash Match, Proof Node Path Valid, and Besu On-Chain Root Sealed.

10. **Build Compliance Bhavcopy & Trade Blotter Exporter (`ComplianceBhavcopyExporter.tsx`):**
    - Interface with Continuous Regulatory Reporting Service (Prompt 233) to fetch daily official compliance bhavcopy datasets.
    - Render tabular preview of symbol bhavcopies: Open, High, Low, Close, VWAP, Total Traded Volume, Total Traded Value, Number of Trades, and On-Chain Settlement Hashes.
    - Implement export generator bundling the dataset into SEBI/IFSCA standardized formats (XML, XBRL, CSV).
    - Bundle the report file, on-chain Besu block receipts, and an HSM-generated SHA-256 manifest into a cryptographically sealed ZIP file using `jszip` and `file-saver`.

11. **Implement Warrant-Backed PII Decryption Vault (`WarrantBackedPiiVault.tsx`):**
    - Enforce default masking on all supervisory screens (e.g., PAN: `XXXXX1234X`, Demat: `************5678`, Name: `J*** D**`).
    - Build a judicial unmasking modal requiring:
      - Upload of digitally signed Court / Judicial Search Warrant (PDF).
      - Extraction and validation of the warrant digital signature (Class-3 PKI).
      - Dual-officer hardware token MFA authorization.
    - Upon validation, request temporary (30-minute) single-record decryption token from the Key Management Service (KMS).
    - Automatically log the unmasking event with the warrant document hash and officer badge IDs into Kafka topic `audit.pii.unmask.v1`.

12. **Configure Security Hardening, Content Security Policy, and Anti-Exfiltration:**
    - Configure strict HTTP headers in Next.js configuration:
      - `Content-Security-Policy`: Disallow inline scripts (enforce nonces), restrict `connect-src` strictly to regulatory APIs and Besu RPC nodes, restrict `frame-ancestors 'none'`.
      - `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`.
    - Implement clipboard and print screen interception: suppress printing (`@media print { body { display: none !important; } }`) and log any copy/clipboard export events to the audit trail.

13. **Write Comprehensive Automated Unit, Cryptographic, and E2E Tests:**
    - Vitest unit tests for client-side Merkle proof verification, OTR ratio calculations, and bhavcopy schema validation.
    - Playwright end-to-end tests validating mTLS certificate authentication, trade replay scrubber operations, dual-officer maker-checker freeze workflows, and watermark tamper resilience.

## Interfaces / Contracts

### Regulatory Session & Authentication Schemas
```typescript
export type RegulatoryJurisdiction = 'SEBI_DOMESTIC' | 'IFSCA_GIFT_CITY';

export interface RegulatoryOfficerSession {
  sessionId: string;
  badgeId: string;
  officerName: string;
  department: 'MARKET_SURVEILLANCE' | 'INVESTIGATION_AND_ENFORCEMENT' | 'STATUTORY_AUDIT';
  jurisdiction: RegulatoryJurisdiction;
  clientCertThumbprint: string;
  workstationIp: string;
  sessionStartedAt: string;
  expiresAt: string;
  hardwareTokenVerified: boolean;
  activeWarrantIds: string[];
}

export interface WatermarkPayload {
  badgeId: string;
  jurisdiction: RegulatoryJurisdiction;
  ipAddress: string;
  timestampUtc: string;
  sessionNonce: string;
}
```

### Live Order Book & Trade Replay Schemas
```typescript
export interface OrderBookLevel {
  price: string; // Fixed-point string representation
  quantity: number;
  orderCount: number;
}

export interface OrderBookRestingOrder {
  orderId: string;
  sequenceNo: bigint;
  participantCode: string; // Pseudonymized, e.g. "MEMBER-4912"
  side: 'BUY' | 'SELL';
  price: string;
  originalQty: number;
  remainingQty: number;
  timeInForce: 'DAY' | 'IOC' | 'FOK';
  timestampNs: bigint;
}

export interface LiveOrderBookSnapshot {
  isin: string;
  symbol: string;
  sequenceNo: bigint;
  timestampNs: bigint;
  bids: OrderBookLevel[];
  asks: OrderBookLevel[];
  restingOrders: OrderBookRestingOrder[];
}

export interface TradeReconstructionEvent {
  eventId: string;
  sequenceNo: bigint;
  timestampNs: bigint;
  eventType: 'ORDER_PLACED' | 'ORDER_MODIFIED' | 'ORDER_CANCELLED' | 'TRADE_MATCHED' | 'CIRCUIT_HALT';
  isin: string;
  price: string;
  quantity: number;
  buyerParticipantCode?: string;
  sellerParticipantCode?: string;
  executionId?: string;
  besuTxHash?: `0x${string}`;
}

export interface TradeReplayState {
  currentTimestampNs: bigint;
  isPlaying: boolean;
  playbackSpeed: 0.1 | 0.5 | 1.0 | 2.0 | 5.0 | 10.0;
  totalEvents: number;
  currentEventIndex: number;
  activeBook: LiveOrderBookSnapshot;
}
```

### Circuit Breaker & Volatility Schemas
```typescript
export type VolatilityHaltType = 'DYNAMIC_LULD_BREACH' | 'INDEX_CIRCUIT_BREAKER_LEVEL_1' | 'INDEX_CIRCUIT_BREAKER_LEVEL_2' | 'INDEX_CIRCUIT_BREAKER_LEVEL_3';

export interface CircuitHaltRecord {
  haltId: string;
  isin: string;
  symbol: string;
  haltType: VolatilityHaltType;
  haltTriggeredAtNs: bigint;
  scheduledResumeAtNs: bigint;
  actualResumeAtNs?: bigint;
  triggerPrice: string;
  referencePrice: string;
  upperCollarPrice: string;
  lowerCollarPrice: string;
  preHaltQuotesPerSec: number;
  preHaltCancellationsPerSec: number;
  callAuctionEquilibriumPrice?: string;
  callAuctionMatchedVolume?: number;
  status: 'HALTED' | 'CALL_AUCTION' | 'RESUMED';
}
```

### Surveillance Alerts & Judicial Freeze Schemas
```typescript
export type AbusePatternType = 'SPOOFING' | 'LAYERING' | 'WASH_TRADING' | 'CIRCULAR_TRADING' | 'MOMENTUM_IGNITION' | 'OTR_VIOLATION';

export interface SurveillanceAlertRecord {
  alertId: string;
  alertType: AbusePatternType;
  isin: string;
  symbol: string;
  severity: 'INFO' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  detectedAt: string;
  involvedParticipantCodes: string[];
  confidenceScore: number; // 0.0 to 1.0
  evidencePayload: {
    orderToTradeRatio?: number;
    quoteLifespanMs?: number;
    washTradeMatchIds?: string[];
    circularGraphNodes?: string[];
  };
  investigationStatus: 'PENDING_REVIEW' | 'INVESTIGATING' | 'ESCALATED_JUDICIAL' | 'DISMISSED';
}

export type FreezeScope = 'FULL_ACCOUNT_FREEZE' | 'ORDER_SUBMISSION_BLOCK' | 'COLLATERAL_HOLD' | 'SYMBOL_TRADING_HALT';

export interface JudicialFreezeRequest {
  targetType: 'PARTICIPANT' | 'WALLET_ADDRESS' | 'SYMBOL';
  targetIdentifier: string; // Participant Code, Ethereum Address, or ISIN
  freezeScope: FreezeScope;
  statutoryAuthority: RegulatoryJurisdiction;
  courtOrderReference: string;
  warrantDocumentHash: string; // SHA-256 of uploaded court warrant PDF
  makerOfficerBadgeId: string;
  makerNotes: string;
  makerSignature: string; // WebAuthn assertion signature
}

export interface JudicialFreezeApproval {
  freezeRequestId: string;
  checkerOfficerBadgeId: string;
  checkerSignature: string; // WebAuthn assertion signature
  approvalTimestamp: string;
}

export interface FreezeExecutionResponse {
  freezeRequestId: string;
  status: 'EXECUTED' | 'FAILED' | 'REJECTED';
  executedAt: string;
  matchingEngineCancelledOrdersCount: number;
  besuContractTxHash?: `0x${string}`;
  auditLogReceiptId: string;
}
```

### On-Chain Cryptographic Audit & Bhavcopy Schemas
```typescript
export interface MerkleInclusionProofRequest {
  eventId: string;
  epochId: number;
}

export interface MerkleInclusionProofResponse {
  eventId: string;
  epochId: number;
  leafHash: `0x${string}`;
  leafIndex: number;
  proofSiblings: `0x${string}`[];
  besuBlockNumber: number;
  besuEpochRoot: `0x${string}`;
  besuAnchorTxHash: `0x${string}`;
  isVerifiedClientSide: boolean;
}

export interface ComplianceBhavcopyRecord {
  reportingDate: string;
  isin: string;
  symbol: string;
  instrumentType: string;
  openPrice: string;
  highPrice: string;
  lowPrice: string;
  closePrice: string;
  vwapPrice: string;
  totalQuantityTraded: number;
  totalValueTradedFiat: string;
  totalTradesCount: number;
  onChainSettlementBatchId: string;
  besuBlockHash: `0x${string}`;
  sha256Hash: string;
}
```

### REST / GraphQL Supervisory Endpoints
```text
# Real-Time Telemetry & Order Book Inspection
GET  /api/v1/regulator/book/live?isin={isin}
GET  /api/v1/regulator/book/replay?isin={isin}&fromNs={from}&toNs={to}
GET  /api/v1/regulator/circuit-halts?jurisdiction={jurisdiction}&status={status}
GET  /api/v1/regulator/surveillance/alerts?severity={severity}&limit={limit}

# Judicial Freeze & Statutory Enforcement
POST /api/v1/regulator/judicial/freeze/propose      (Maker: Initiate Freeze Request)
POST /api/v1/regulator/judicial/freeze/approve      (Checker: WebAuthn Dual Approval)
POST /api/v1/regulator/judicial/freeze/symbol-halt  (Emergency Instant Symbol Halt)

# Cryptographic Audit Verification & Bhavcopy Export
GET  /api/v1/regulator/audit/proof?eventId={eventId}&epochId={epochId}
GET  /api/v1/regulator/bhavcopy/daily?date={YYYY-MM-DD}&jurisdiction={jurisdiction}
POST /api/v1/regulator/bhavcopy/export-package      (Generate Signed Compliance ZIP)

# Warrant-Backed Identity Unmasking (DPDP Act 2023)
POST /api/v1/regulator/pii/unmask-request           (Upload Warrant & Request Token)
POST /api/v1/regulator/pii/unmask-approve           (Dual Officer Approval)
GET  /api/v1/regulator/pii/beneficial-owner/{code}  (Decrypted KYC Profile with Lease)
```

## Security & Compliance Notes
- **Mutual TLS (mTLS) Client Certificate Authentication:** Access to the supervisory portal is strictly restricted at the edge proxy layer. Officers must provide valid, government-issued X.509 client certificates issued by designated root CAs (SEBI PKI / IFSCA Root CA). Handshakes without authorized certificates are dropped immediately.
- **Hardware Token Multi-Factor Authentication (FIDO2 / WebAuthn):** High-privilege enforcement operations (judicial freeze orders, symbol trading halts, and beneficial owner unmasking) mandate cryptographic signature generation using FIPS 140-2 Level 3 hardware security keys (e.g., YubiKey 5 Series). Passwords or SMS/TOTP tokens are explicitly prohibited for regulatory enforcement actions.
- **Digital Personal Data Protection (DPDP) Act 2023 Compliance:** Retail investor personal identifiable information (PAN, Aadhaar Vault reference, mobile number, bank account details) is encrypted at rest using envelope encryption (AES-256-GCM) with keys managed in AWS KMS. Regulatory officers view pseudonymized participant codes (`MEMBER-XXXX`, `CLIENT-YYYY`) by default. Decryption is technically barred without:
  1. Uploading a cryptographically verified Judicial Search Warrant issued by a competent court of law.
  2. Dual-officer maker-checker WebAuthn digital signatures.
  3. Time-bounded single-record decryption lease (maximum 30 minutes), after which client memory is purged.
- **Dynamic Anti-Exfiltration Forensic Watermarking:** To prevent unauthorized leaks via physical camera photography or screen capture software, every rendered supervisory viewport features an immutable, semi-transparent SVG watermark containing the active officer badge ID, workstation IP address, timestamp, and ephemeral session nonce. Attempted DOM alteration triggers an immediate session lock.
- **Immutable Regulatory Access Logging:** Every user action within the supervisory portal (page accesses, query parameters, order book scrub interactions, export downloads, and freeze proposals) is dispatched synchronously to Kafka topic `regulatory.portal.access.v1` and committed to the Immutable Audit Log Service (Prompt 218) and S3 WORM storage with an 8-year statutory retention hold.
- **Content Security Policy (CSP) & Client Hardening:**
  ```http
  Content-Security-Policy: default-src 'none'; script-src 'self' 'nonce-{RANDOM}'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss://surveillance.growww.in wss://marketdata.growww.in https://audit.growww.in https://rpc.besu.growww.in; font-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'none'; form-action 'self';
  ```
- **Air-Gapped Operation & Anti-Clickjacking:** Frame ancestors are set to `'none'` and `X-Frame-Options: DENY` is strictly enforced. The application blocks external third-party tracking scripts, analytics SDKs, and external CDNs.

## Acceptance Criteria
- [ ] Next.js 14 regulatory supervisory portal initializes under route `apps/growww_admin/regulator` with strict server-side mTLS certificate validation.
- [ ] Non-mTLS connections or invalid client certificates are rejected at the edge gateway with HTTP 403 Forbidden.
- [ ] FIDO2 / WebAuthn hardware token authentication challenge succeeds on session bootstrap and enforces automatic 15-minute inactivity lock.
- [ ] Dynamic forensic watermark renders across all supervisory viewports with officer badge ID, IP address, and timestamp; DOM tampering triggers immediate screen blanking.
- [ ] Live Order Book Inspector streams real-time Level-2 aggregated depth and Level-3 individual resting orders with nanosecond sequencing.
- [ ] Order book depth chart visualizes cumulative bids and asks curves with real-time spread ribbon updates.
- [ ] Trade Reconstruction Player scrubs historical tick-by-tick events with play, pause, step-forward, step-backward, and speed controls (0.1x to 10x).
- [ ] Order book state and time-and-sales tape accurately re-render at any selected historical microsecond scrub point.
- [ ] Circuit Halt Analyzer displays historical price charts overlaid with static daily bands (+/- 10%, +/- 20%) and rolling 5-minute LULD corridors (+/- 3%, +/- 5%).
- [ ] Volatility halts display pre-halt quote stuffing metrics, cancellation spikes, and subsequent call auction equilibrium price discovery curves.
- [ ] Surveillance Alert Desk displays triaged market abuse alerts (spoofing, layering, wash trading, circular trading) with confidence scores and participant trade graphs.
- [ ] Order-to-Trade Ratio (OTR) monitor computes rolling hourly participant metrics and flags accounts exceeding the 100:1 collar.
- [ ] Maker-Checker Judicial Freeze Terminal requires dual-officer WebAuthn signatures and warrant hash to execute participant freezes or symbol halts.
- [ ] Executed freeze commands immediately cancel resting orders via Matching Engine gRPC and call `freezeParticipant` on `ComplianceRegistry.sol`.
- [ ] Audit Log Verifier fetches off-chain audit logs and mathematically verifies RFC 6962 Merkle inclusion proofs against Hyperledger Besu on-chain roots.
- [ ] Compliance Bhavcopy Exporter generates standardized XML, XBRL, and CSV reports bundled with SHA-256 manifests into cryptographically sealed ZIP files.
- [ ] Warrant-Backed PII Vault enforces zero PII exposure by default; temporary unmasking requires court warrant PDF upload and dual WebAuthn authorization.
- [ ] All supervisory clicks, queries, replays, and exports are immutably logged to Kafka topic `regulatory.portal.access.v1` and persisted in WORM storage.
- [ ] Vitest unit tests achieve >90% code coverage across Merkle proof verification, OTR calculations, and bhavcopy schema validation.
- [ ] Playwright E2E tests validate complete supervisory workflows including mTLS gate, trade replay scrubber, and maker-checker freeze execution.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 204 (Order Service & Lifecycle Management).
  - Prompt 205 (Low-Latency Order Matching Engine Core).
  - Prompt 207 (High-Throughput Market Data Service).
  - Prompt 218 (Immutable Audit Log Service).
  - Prompt 228 (Real-Time Market Surveillance & Dynamic Volatility Engine).
  - Prompt 233 (Continuous 24x7 Regulatory Statutory Reporting & Compliance Dispatcher).
  - Prompt 302 (Hyperledger Besu Private Consortium Network Architecture).
  - Prompt 305 (ERC-1400 Security Token Suite & Transfer Compliance Hooks).
  - Prompt 306 (Settlement DvP & Atomic Clearing Smart Contracts).
  - Prompt 601 (Next.js Investor Web App Scaffolding & Foundation).
- **Parallel Tasks:**
  - Prompt 607 (Admin Console: Regulatory Reporting & Audit Export Portal).
  - Prompt 614 (Next.js 14 Clearing Member & Broker Capital Adequacy Portal).
  - Prompt 806 (Observability Stack Telemetry & Metrics).
- **Downstream Blockers:**
  - Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios).
  - Prompt 907 (Regulatory Sandbox Pilot Launch & Supervisory Handover).
