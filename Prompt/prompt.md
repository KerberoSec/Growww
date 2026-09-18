# MASTER PROMPT — Generate the "prompts/" Directory for the Growww Project

## ROLE

You are a senior staff-level software architect and technical writer. Your job in this
task is **not** to write the Growww application itself. Your job is to **generate a
complete `prompts/` directory** containing 100+ standalone, numbered Markdown prompt
files. Each file is a self-contained instruction sheet that a developer (or another AI
coding agent, e.g. Claude Code) can hand to a coding assistant later to actually build
one specific piece of the system. Think of this as writing the full "build ticket
backlog" for a large fintech + blockchain + multi-platform client program, where every
ticket is detailed enough that no additional context is required to start work.

## PROJECT CONTEXT (read this fully before generating anything)

**Project name:** Growww
**One-line purpose:** A blockchain-based, SEBI/RBI-compliant investment infrastructure
that lets Indian and (subject to regulatory approval) international investors buy
fractional, real-asset-backed digital representations of Indian equities, using INR,
through licensed brokers/custodians/depositories, with a permissioned blockchain ledger
providing transparency, auditability, and proof-of-reserve — NOT an unregulated crypto
exchange, and NOT synthetic/derivative tokens disconnected from real securities.

**Non-negotiable design principles every generated prompt must respect:**
1. Every digital token/unit issued must map 1:1 (or fractionally) to a real underlying
   Indian security held in custody by a SEBI-registered custodian/depository (NSDL/CDSL)
   — never a synthetic or purely speculative token.
2. All INR movement happens through RBI-licensed banking rails / payment aggregators —
   never through unregulated stablecoins or crypto on-ramps for the domestic leg.
3. The blockchain layer is **permissioned** (e.g. Hyperledger Fabric, Besu, or a
   permissioned Ethereum/Polygon Supernet — the prompts should explore/compare these,
   not assume one), used for issuance, transfer, redemption, and proof-of-reserve
   records — it is a transparency/settlement-record layer, not a public/speculative
   crypto asset layer.
4. Compliance (KYC/AML, sanctions screening, transaction monitoring, SEBI/RBI/GIFT-City
   IFSCA reporting) is built into the architecture from day one, not bolted on later.
5. Multi-party authorization (no single admin key) for all sensitive operations
   (freezes, minting, redemption approval, permission changes).
6. Fee model: no holding fee; platform fee only on realized profit (reference figure:
   0.1%), so prompts dealing with the fee/ledger engine must implement profit-based
   fee calculation, not AUM-based fees.
7. Two coordinated legal/technical entities: (a) a domestic regulated entity handling
   custody of real Indian securities, and (b) an international gateway entity
   (potentially GIFT City / IFSCA) handling foreign investor onboarding — architecture
   and services must reflect this two-entity separation with controlled, audited
   inter-entity communication.
8. Client platforms required: **Android, iOS, Windows, Linux, macOS** — all as a single
   **Flutter** codebase (thick/native client, not a wrapped web view), plus a
   Next.js/React web app and backend/admin consoles.
9. Use the widest reasonable set of modern, production-grade technologies (polyglot
   backend: Python/FastAPI, Go, Rust; PostgreSQL, Redis, Kafka; Docker + Kubernetes;
   permissioned blockchain node infra; CI/CD; observability stack) — favor showing the
   *right* tool per component over forcing one language everywhere, and every prompt
   must justify its technology choice in one paragraph.
10. Security-first: HSM/KMS-backed key custody, mTLS between services, zero-trust
    internal network, continuous audit logging, independent security review checkpoints.

**Reference materials already produced for this project (use them as background
knowledge; do not just copy them — expand, structure, and convert them into the
numbered prompt backlog described below):**
- `indian_blockchain_stock_exchange_plan.md` — regulatory/legal architecture plan
  (SEBI/RBI/IFSCA pathway, entity structure, licensing).
- `web3_architecture_plan.md` — blockchain/Web3 layer design.
- `SYSTEM_DESIGN.md` — overall system design reference.
- `SettlementDvP.sol` — delivery-vs-payment settlement smart contract reference.
- `matching_engine.py` — order matching engine reference implementation.
- `kyc_service.py` — KYC service reference implementation.
- `docker-compose.yml` — reference local dev stack.
- `kerberosec_coin_build_plan.md` — related token/coin build plan for cross-reference
  on token-engineering patterns (mint/burn/reserve/security practices) where relevant.

Treat these as the seed architecture. The prompts you generate should be consistent
with them, filling gaps, adding the Flutter multi-platform client layer (which is not
yet covered), adding DevOps/observability/testing prompts, and breaking the whole
system down into buildable units.

## WHAT YOU MUST PRODUCE

Create a directory named `prompts/` containing **at least 110 Markdown files**. Do not
just describe them — actually write each file's full content per the template below.
Organize them into the numbered categories and file-naming scheme specified below.
Every prompt file must be independently usable: someone should be able to open
`037_backend_order_matching_engine.md` alone, with zero other context, and know exactly
what to build.

### Directory & naming scheme

```
prompts/
  00_INDEX.md                          <- master table of contents, one line per prompt
  000_project_north_star.md            <- vision/purpose prompt (see category 0)
  0xx_...                              <- Category 0: Vision, Purpose & Regulatory Frame (00x)
  1xx_...                              <- Category 1: Architecture & Standards (10x-1xx)
  2xx_...                              <- Category 2: Backend & Microservices (2xx)
  3xx_...                              <- Category 3: Blockchain / Ledger Layer (3xx)
  4xx_...                              <- Category 4: Data, Storage & Messaging (4xx)
  5xx_...                              <- Category 5: Flutter Multi-Platform Client (5xx)
  6xx_...                              <- Category 6: Web App & Admin Console (6xx)
  7xx_...                              <- Category 7: Security, Compliance & Identity (7xx)
  8xx_...                              <- Category 8: DevOps, Infra & Observability (8xx)
  9xx_...                              <- Category 9: Testing, QA, Launch & Ops (9xx)
```

Each filename = `<number>_<short_snake_case_title>.md`, e.g.:
`101_system_architecture_overview.md`, `214_order_matching_engine_service.md`,
`307_permissioned_blockchain_choice_and_setup.md`, `512_flutter_app_shell_and_navigation.md`.

Numbers do not need to be contiguous but must be strictly increasing within a category
and must never repeat. Pad to 3 digits.

### `00_INDEX.md` requirements

A single table with columns: `#`, `Filename`, `Category`, `One-line purpose`,
`Depends on (#s)`. This lets a reader or an agent pick a build order.

## TEMPLATE — every individual prompt file (e.g. `2xx_*.md`) must contain these
exact sections, in this order:

```markdown
# <Prompt Number> — <Title>

## Purpose
Why this piece exists, what user or business problem it solves, and how it fits into
the Growww vision (1-2 short paragraphs, plug into the north-star vision, not generic).

## What You Are Building
A precise, concrete description of the deliverable (a service, a screen, a contract,
a pipeline, a doc). Bullet the concrete outputs (files/repos/endpoints/binaries).

## Scope Boundaries
Explicitly state what is IN scope and what is OUT of scope / handled by another prompt
(cross-reference other prompt numbers).

## Technology to Use
- Primary language(s)/framework(s), named explicitly, with a one-paragraph justification
  ("why this and not X").
- Any library/service dependencies (with version-class guidance, e.g. "Postgres 16+").
- How it fits the polyglot stack (which other services it talks to and how — REST,
  gRPC, Kafka topic, on-chain call, etc).

## Backend / Infra Touchpoints
What databases, queues, blockchain contracts, or third-party APIs (e.g. NSDL/CDSL,
UPI/payment aggregator, KYC/AML vendor, custodian API) this interacts with.

## Blockchain Interaction (write "N/A" only if genuinely not applicable)
Exactly how this component reads or writes to the permissioned ledger: which
contract/chaincode, which events it emits/subscribes to, how INR-denominated trades or
token mint/burn/transfer/redemption events reconcile with off-chain custody records.

## Step-by-Step Build Instructions
A numbered, actionable sequence (5-15 steps) an engineer or coding agent should follow,
from scaffolding to a runnable/testable state.

## Interfaces / Contracts
Concrete API shapes, gRPC/proto sketches, DB schema sketches, or smart-contract
function signatures as relevant — enough to be implementable without guessing.

## Security & Compliance Notes
Specific controls relevant to this piece (key custody, PII handling, audit logging,
rate limiting, multi-party approval, regulatory reporting hook, etc).

## Acceptance Criteria
A checklist of concrete, testable "done" conditions.

## Suggested Order / Dependencies
Which other prompt numbers should be completed first, and which can happen in parallel.
```

Keep each file focused — roughly 400-900 words of substance, not padding. Do not
repeat entire paragraphs across files; cross-reference by number instead.

## FULL LIST OF PROMPTS TO GENERATE (minimum set — you may add more, never fewer)

Generate one file per line below, following the template. Use these as your outline;
flesh out the actual detailed content yourself using the project context above.

### Category 0 — Vision, Purpose & Regulatory Frame (000–099)
000 Project north star: what Growww is, is not, and why it must exist
001 Glossary of domain terms (DvP, custodian, depository, IFSCA, proof-of-reserve, etc.)
002 Two-entity legal/technical structure (domestic regulated entity + GIFT City gateway)
003 Regulatory pathway overview (SEBI, RBI, IFSCA, sandbox strategy)
004 KYC/AML policy definition (domestic investors)
005 KYC/AML policy definition (foreign investors via GIFT City)
006 Fee model specification (profit-only 0.1% realized-gain fee engine spec)
007 Proof-of-reserve public disclosure design
008 Data protection & privacy policy (DPDP Act / GDPR-equivalent for foreign users)
009 Risk disclosure & investor protection content requirements
010 Non-functional requirements master doc (latency, uptime, RTO/RPO targets)

### Category 1 — Architecture & Standards (100–199)
101 System architecture overview (C4-style context + container diagrams as text/mermaid)
102 Service boundary map & bounded contexts
103 API design standards (REST + gRPC conventions, versioning, error format)
104 Event schema & Kafka topic naming standards
105 Authentication & authorization architecture (OAuth2/OIDC, mTLS service mesh)
106 Monorepo vs polyrepo decision + repo layout convention
107 Coding standards & linting per language (Python, Go, Rust, Dart, TS)
108 Environment strategy (local/dev/staging/prod) and config management
109 Secrets management architecture (Vault/KMS/HSM)
110 Inter-entity secure communication design (domestic <-> GIFT City gateway)
111 Domain model: users, accounts, holdings, orders, trades, tokens, settlements
112 Idempotency & exactly-once processing strategy across services

### Category 2 — Backend & Microservices (200–299)
201 User service (registration, profile, auth) — FastAPI
202 KYC/AML service — document capture, liveness check, sanctions screening
203 Wallet/account service (INR ledger, balances, holds)
204 Order service (order intake, validation, lifecycle state machine)
205 Order matching engine (Rust or Go) — spec building on matching_engine.py reference
206 Risk & margin pre-trade checks service
207 Market data service (WebSocket price/order-book streaming)
208 Trade settlement service (DvP orchestration, links to Category 3)
209 Portfolio & holdings service (fractional unit accounting)
210 Fee & realized-P&L calculation engine
211 Notification service (push/email/SMS)
212 Payment gateway integration service (UPI/NEFT/RTGS on-ramp & off-ramp)
213 Custodian/depository integration service (NSDL/CDSL API adapter)
214 Foreign-investor funding & FX service (GIFT City leg)
215 Reconciliation service (on-chain vs off-chain ledger reconciliation)
216 Reporting service (SEBI/RBI/IFSCA regulatory report generation)
217 Admin/back-office service (account actions, freezes, exception handling)
218 Audit log service (immutable append-only operational audit trail)
219 API gateway & BFF (backend-for-frontend) layer
220 Rate limiting & abuse-prevention service
221 Search/discovery service (browse securities, company info)
222 Corporate actions service (dividends, splits, bonuses reflected in tokens)
223 Tax reporting/statement generation service (capital gains statements)
224 Referral / growth service (optional, compliant with SEBI advertising rules)

### Category 3 — Blockchain / Ledger Layer (300–399)
301 Permissioned blockchain platform evaluation & selection (Fabric vs Besu vs Polygon Supernet)
302 Network topology & node/validator setup for the permissioned chain
303 Smart contract / chaincode: token issuance (mint against custody confirmation)
304 Smart contract / chaincode: token redemption (burn against custody release)
305 Smart contract / chaincode: transfer with compliance hooks (whitelist/KYC gate)
306 Smart contract / chaincode: settlement DvP logic (build on SettlementDvP.sol reference)
307 Multi-party authorization / multisig governance contract for sensitive ops
308 On-chain proof-of-reserve publishing mechanism
309 Event indexing service (chain -> queryable off-chain index)
310 Chain node monitoring & alerting
311 Key management for validators/signers (HSM integration)
312 Chain upgrade & governance process (protocol/contract versioning)
313 Cross-entity ledger bridge (domestic ledger <-> GIFT City ledger visibility)
314 Chain disaster recovery & backup strategy

### Category 4 — Data, Storage & Messaging (400–449)
401 PostgreSQL schema design (core transactional data)
402 Redis usage patterns (session, cache, real-time order book state)
403 Kafka cluster design & topic partitioning strategy
404 Data warehouse / analytics pipeline (for compliance & BI)
405 Data retention & archival policy implementation
406 Backup & restore automation for all datastores
407 Master data management (securities master, corporate actions master)

### Category 5 — Flutter Multi-Platform Client (500–599)
501 Flutter project scaffolding for Android/iOS/Windows/Linux/macOS from one codebase
502 App architecture (state management choice: Riverpod/Bloc — justify), folder structure
503 Design system & theming (light/dark, platform-adaptive widgets)
504 Onboarding & KYC flow UI (camera capture, liveness, document upload)
505 Authentication UI (biometric login, MPIN, OTP, session handling)
506 Home/dashboard screen (portfolio summary, market movers)
507 Market/watchlist screen with real-time price updates (WebSocket client)
508 Security detail screen (charts via TradingView-style lightweight charts in Flutter)
509 Order placement flow (buy/sell, order types, confirmation)
510 Portfolio & holdings screen (fractional units, P&L, proof-of-reserve display)
511 Wallet/funds screen (deposit/withdraw, UPI/bank linking)
512 Transaction & trade history screen
513 Notifications center UI
514 Settings & profile management UI
515 Offline/queued-order UX for market-closed hours
516 Platform-specific packaging: Android (Play Store) build & signing pipeline
517 Platform-specific packaging: iOS (App Store) build & signing pipeline
518 Platform-specific packaging: Windows (MSIX) build pipeline
519 Platform-specific packaging: Linux (Snap/Flatpak/AppImage) build pipeline
520 Platform-specific packaging: macOS (notarized .app/.dmg) build pipeline
521 Local secure storage (Keychain/Keystore/DPAPI abstraction) for tokens/secrets
522 Push notification integration across platforms (FCM/APNs/desktop equivalents)
523 Accessibility & localization (i18n for English/Hindi + regional languages)
524 Client-side crash reporting & analytics integration
525 Flutter <-> backend API client layer (typed client, retry/backoff, auth refresh)
526 Deep linking & universal links across platforms

### Category 6 — Web App & Admin Console (600–649)
601 Next.js/React investor web app scaffolding
602 Web onboarding/KYC flow
603 Web trading dashboard (parity with Flutter app)
604 Admin console: user & KYC review dashboard
605 Admin console: risk/exception queue & multi-party approval UI
606 Admin console: proof-of-reserve & reconciliation dashboard
607 Admin console: regulatory reporting dashboard
608 Public marketing/landing site with investor disclosures

### Category 7 — Security, Compliance & Identity (700–749)
701 Threat model for the full system (STRIDE-based)
702 Identity & access management for internal staff (RBAC, least privilege)
703 Sanctions/PEP screening integration
704 Transaction monitoring & suspicious-activity alerting
705 Penetration testing & bug bounty program plan
706 Incident response runbook
707 Data encryption standards (at rest, in transit, key rotation)
708 Secure SDLC & code review policy
709 Third-party vendor security assessment checklist
710 Business continuity / disaster recovery plan

### Category 8 — DevOps, Infra & Observability (800–849)
801 Docker Compose local dev environment (extend reference docker-compose.yml)
802 Kubernetes cluster architecture (namespaces, network policies)
803 CI pipeline design (build/test/lint per language)
804 CD pipeline design (progressive delivery, canary/blue-green)
805 Infrastructure as Code (Terraform modules for cloud + blockchain nodes)
806 Observability stack (Prometheus/Grafana/Loki/Tempo or equivalent)
807 Centralized logging & audit trail pipeline
808 Alerting & on-call runbooks
809 Cost monitoring & optimization plan
810 Environment promotion & release management process

### Category 9 — Testing, QA, Launch & Ops (900–950)
901 Unit/integration testing strategy per service
902 End-to-end testing strategy (Flutter integration tests + backend)
903 Load & performance testing plan (order matching engine focus)
904 Chaos engineering / resilience testing plan
905 Security testing automation (SAST/DAST/dependency scanning in CI)
906 UAT plan with real (sandboxed) regulatory scenarios
907 Regulatory sandbox pilot launch plan
908 Production launch runbook & rollback plan
909 Post-launch monitoring, SLOs & error budgets
910 Customer support & grievance-redressal process (SEBI-mandated)

## OUTPUT FORMAT RULES

- Actually create every file listed above (110 files) plus `00_INDEX.md`, using
  `create_file`/equivalent — do not summarize them inline in chat instead of writing
  files.
- Write in clear, direct English. No marketing fluff. Assume the reader is a competent
  engineer who wants to start building immediately.
- Be consistent: if prompt 205 says the matching engine is in Rust, no other prompt may
  contradict that without explicitly noting the trade-off.
- After generating all files, present the `prompts/` directory to the user and share
  `00_INDEX.md` first so they can navigate.
