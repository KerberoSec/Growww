# 252 - Gemini Financial Intelligence & Literacy Service (FastAPI / LangChain / RAG / Hyperledger Besu)

## Purpose
Retail and institutional investors navigating modern tokenized capital markets require instantaneous, high-clarity financial education and verifiable real-time transparency. In an exchange ecosystem where fractional equities and real-world assets (RWAs) are 1:1 backed by custodial depository reserves (NSDL/CDSL) and notarized on a permissioned Hyperledger Besu blockchain, users frequently have complex questions regarding asset provenance, fractional ownership mechanics, custody segregation, corporate action distributions, and transaction fees.

At the same time, operating under the strict regulatory framework of the Securities and Exchange Board of India (SEBI Investment Advisers Regulations, 2013) mandates that conversational AI systems must strictly operate as educational and analytical intelligence tools, rather than autonomous trading bots or uncertified financial advisers. Providing buy/sell recommendations, stock tips, or guaranteed return projections is legally prohibited and carries catastrophic liability.

The **Gemini Financial Intelligence & Literacy Service** (`services/gemini-advisor-service`) provides an enterprise-grade, conversational intelligence layer powered by Google Gemini 1.5 Pro and Gemini 1.5 Flash. It enables natural language Q&A, interactive hypothetical scenario modeling, multi-hundred-page regulatory document summarization (SIDs, DRHPs, annual reports), real-time on-chain Proof-of-Reserve verification, and complete fee transparency (explaining the platform's flat 0.00% transaction fee (No fee at all) with its 0.00% fee at launch (governed by FeeController.sol) split). A robust non-advisory guardrail pipeline deterministically prevents unauthorized investment advice while delivering verifiable, cryptographically grounded financial literacy.

## What You Are Building
A high-throughput, asynchronous microservice (`services/gemini-advisor-service`) developed in Python 3.12 with FastAPI and LangChain/LlamaIndex. Key components include:
- **Dual-Engine Gemini Inference Pipeline:** Dynamic orchestration between Gemini 1.5 Flash (for sub-second interactive chat, prompt chip suggestions, and real-time streaming tokens) and Gemini 1.5 Pro (for large-context processing of 1M+ token regulatory filings, DRHPs, and annual reports).
- **Retrieval-Augmented Generation (RAG) Engine:** Hybrid semantic and full-text vector retrieval over PostgreSQL (`pgvector`) storing indexed SEBI circulars, exchange bylaws, equity prospectuses, corporate action schedules, and platform architectural specifications.
- **Real-Time On-Chain & Custody Context Injector:** Live context enricher that queries the Hyperledger Besu RPC node, Proof-of-Reserve Sparse Merkle Tree roots (`ProofOfReserveRegistry.sol`), and Demat pool balances to answer factual queries like "Where is my fractional Reliance share held?" with cryptographic proofs.
- **Hypothetical Scenario Shock Simulator:** Analytical modeling engine that accepts user portfolio parameters and simulates macroeconomic shocks (such as +/- 100 bps RBI repo rate shift, 10% crude oil rally, or broad index drawdowns), producing structured JSON payloads for frontend interactive slider visualizations without offering investment advice.
- **Multi-Layer Non-Advisory Guardrail System:** Real-time semantic filter and classifier pipeline that intercepts advisory intents ("Should I buy HDFC Bank?", "Which stock will double?"), rejecting solicitations with educational context and directing investors to fundamental facts.
- **Regulatory Document Summarizer:** Structured document digestion pipeline that accepts PDF prospectus uploads, extracts key risk factors, capital allocation tables, and promoter holding patterns, outputting executive summaries formatted in clean markdown.
- **Universal Fee Transparency Explainer:** Contextual assistant component that breaks down the 0.00% (No fee at all) platform fee (0 bps (0.00% fee at launch) on turnover) into its statutory and platform components (Treasury reserve, 25% Core Settlement Guarantee Fund, Investor Protection Fund per FeeController governance) and proves zero AUM management fees.
- **Streaming SSE & WebSocket Gateway:** Low-latency server-sent events (SSE) and WebSocket endpoints delivering token-by-token streaming markdown with citations and structured visual widget triggers.

## Scope Boundaries
- **In Scope:**
  - Natural language conversational Q&A regarding platform mechanics, depository custody, fractional shares, and corporate actions.
  - Multi-document RAG over SEBI filings, statutory disclosures, and platform terms.
  - Real-time on-chain context retrieval from Hyperledger Besu and Proof-of-Reserve contracts.
  - Hypothetical portfolio stress testing and shock scenario visualization payloads.
  - Non-advisory guardrails enforcing SEBI compliance and preventing buy/sell advice.
  - Token-by-token streaming responses via Server-Sent Events (SSE) and WebSockets.
  - Audio transcription and text-to-speech integration hooks.
- **Out of Scope / Handled Elsewhere:**
  - Automated order placement or trade execution (handled strictly by Prompt 204 / Prompt 509).
  - Client-side UI rendering and state management (handled in Prompt 515).
  - Direct depository settlement instruction generation (handled in Prompt 208 / Prompt 213).
  - Storage of user PII or trade execution keys (handled in Prompt 201 / Prompt 109).
  - User wallet cash deposits and UPI payment gateways (handled in Prompt 203 / Prompt 212).

## Technology to Use
- **Primary Language & Framework:** Python 3.12 with FastAPI (0.111+) for high-concurrency async I/O, WebSockets, and Server-Sent Events (SSE).
- **LLM Foundation Models & SDK:** Google GenAI SDK (`google-generativeai` and `@google/genai`), LangChain Google GenAI integration (`langchain-google-genai`), and LiteLLM for dynamic routing between `gemini-1.5-pro` and `gemini-1.5-flash`.
- **Vector Database & Embeddings:** PostgreSQL 16+ with `pgvector` extension and Google `text-embedding-004` (768-dimensional embeddings) for hybrid dense and sparse HNSW vector search.
- **Caching & Session State:** Redis 7.2+ for distributed conversational memory (sliding window buffer of recent turns), rate limiting, and cached embeddings of common regulatory queries.
- **Guardrail Enforcement:** NeMo Guardrails combined with custom Pydantic semantic classifiers for real-time non-advisory policy enforcement.
- **Blockchain Connectivity:** `web3.py` with async HTTP/IPC providers to query Hyperledger Besu smart contracts (`ProofOfReserveRegistry.sol`, `SettlementDvP.sol`).
- **Telemetry & Observability:** OpenTelemetry instrumentation, Prometheus metrics exporter, and structured JSON logging with zero PII logging.

## Backend / Infra Touchpoints
- **PostgreSQL 16 with pgvector:** Stores document chunks, semantic embeddings, knowledge base metadata, and anonymized query audit logs across tables `knowledge_documents`, `knowledge_chunks`, `scenario_templates`, and `advisor_audit_logs`.
- **Redis 7.2 Cluster:** Caches user conversation sessions with 24-hour TTL, active rate-limit tokens, and frequently accessed custody reserve snapshots.
- **Apache Kafka:** Publishes audit events to topic `gemini.interactions.v1` and compliance violation events to `gemini.guardrail_alerts.v1`.
- **Portfolio Service (Prompt 209):** Ingests pseudonymized portfolio asset allocation data (percentages by asset class and sector, zero PII) to contextualize scenario shock simulations.
- **Market Data Service (Prompt 207):** Queries real-time index values, sector betas, and historical volatility to parameterize stress-test calculations.
- **Proof-of-Reserve Engine (Prompt 243 / Prompt 308):** Fetches current custodial reserve balances, depository transaction timestamps, and Merkle tree roots.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Verifiable Proof-of-Reserve Grounding:** When an investor queries the custody backing of a specific fractional share or tokenized asset (such as `gRELIANCE` or `gGOLD`), the service directly invokes `ProofOfReserveRegistry.sol` on Hyperledger Besu. It retrieves the latest verified Merkle root, the exact block height, the depository reconciliation timestamp, and the physical backing ratio (100.0% backed).
- **Custodian Demat Pool Verification:** The service formulates natural-language responses referencing verified on-chain batch settlement receipts from `SettlementDvP.sol` without ever exposing internal database keys or customer PII.
- **Zero On-Chain PII Invariant:** In strict compliance with India's DPDP Act 2023, no user identity, chat prompts, or personalized portfolio holdings are ever written to the blockchain. Blockchain queries are purely read-only public verifications of reserve commitments and contract states.

## Financial Intelligence Architecture & Non-Advisory Guardrails

### 1. Dual-Model Intent Router & Query Lifecycle
Incoming queries pass through an intelligent routing layer that determines the execution path based on complexity, required context size, and latency constraints:
1. **Flash Router (< 150ms):** Simple informational questions, platform FAQ queries, fee breakdowns, and glossary lookups are routed to `gemini-1.5-flash` with direct RAG retrieval.
2. **Pro Deep-Reasoning Router:** Complex multi-step queries, DRHP prospectus summarization, balance-sheet extraction, and multi-asset portfolio shock simulations are routed to `gemini-1.5-pro` with 1M+ context window ingestion.
3. **Ledger Context Augmenter:** Queries containing keywords such as "reserve", "custody", "backing", "on-chain", or "blockchain" trigger an automated lookup against the Hyperledger Besu RPC node and the Proof-of-Reserve database. The live cryptographic state is injected directly into the prompt's system context.

```
                    +------------------------------------------+
                    |  Inbound User Prompt (Mobile / Web)      |
                    +--------------------+---------------------+
                                         |
                                         v
                    +------------------------------------------+
                    |  Layer 1: Non-Advisory Guardrail Filter   |
                    |  - Detects Stock Tips / Buy-Sell Advice  |
                    |  - Intercepts Prompt Injections / Jailbreak
                    +--------------------+---------------------+
                                         | (Passed)
                                         v
                    +------------------------------------------+
                    |  Layer 2: Intent Classification & Router  |
                    +---------+----------------------+---------+
                              |                      |
            [RAG / Knowledge] |                      | [Custody / Blockchain]
                              v                      v
        +----------------------------+   +----------------------------+
        | pgvector Semantic Search   |   | Hyperledger Besu RPC       |
        | - SEBI Circulars           |   | - ProofOfReserveRegistry   |
        | - SIDs / DRHPs / Bylaws    |   | - Demat Custody Root Hash  |
        +--------------+-------------+   +--------------+-------------+
                       |                                |
                       +----------------+---------------+
                                        |
                                        v
                    +------------------------------------------+
                    |  Layer 3: System Prompt Synthesis        |
                    |  - SEBI Disclaimers Injected             |
                    |  - 0.00% fee (No fee at all) Invariant Injected          |
                    |  - Grounding Citations Bound             |
                    +--------------------+---------------------+
                                         |
                                         v
                    +------------------------------------------+
                    |  Gemini 1.5 Pro / Flash Model Engine     |
                    +--------------------+---------------------+
                                         |
                                         v
                    +------------------------------------------+
                    |  Layer 4: Output Guardrail & Citation    |
                    |  Verification (Streamed via SSE/WS)      |
                    +------------------------------------------+
```

### 2. Strict Non-Advisory Guardrail Enforcement Matrix
In compliance with SEBI (Investment Advisers) Regulations, 2013 and SEBI (Research Analysts) Regulations, 2014, the service enforces a deterministic three-stage guardrail:

| Intent Category | User Input Example | Guardrail Action | Output Strategy |
| :--- | :--- | :--- | :--- |
| **Direct Stock Advice** | "Should I buy Infosys shares today?" | **BLOCK & REDIRECT** | Declines advisory role. Offers factual historical financial metrics, P/E ratio context, and corporate filing summaries. |
| **Price Target / Forecast** | "Will gold reach 100,000 INR next month?" | **BLOCK & REDIRECT** | Informs user that price predictions cannot be provided. Explains macroeconomic factors influencing commodity prices. |
| **Portfolio Rebalancing Advice** | "Sell my Reliance and buy Tata Motors?" | **BLOCK & REDIRECT** | Intercepts order intent. Directs user to the Order Placement Screen and provides neutral asset class educational guides. |
| **Custody & Reserve Inquiry** | "Is my fractional share really backed?" | **PERMIT & GROUND** | Queries Besu Proof-of-Reserve contract. Returns exact custody ratio (100%), depository partner (NSDL/CDSL), and Merkle root. |
| **Fee Calculation Question** | "Why was a fee deducted from my trade?" | **PERMIT & GROUND** | Transparently itemizes the flat 0.00% transaction fee (No fee at all) (Treasury reserve, Core SGF, Investor Protection Fund per FeeController governance) and proves 0% AUM holding fee. |
| **Scenario Stress Testing** | "What happens to my bonds if rates rise 1%?" | **PERMIT & SIMULATE** | Computes duration-based bond valuation shift, generating interactive slider chart payload with educational annotations. |

### 3. Interactive Scenario Shock Modeling Logic
The service provides structured data for frontend slider exploration without delivering personalized advice:
- **Interest Rate Shock:** $\Delta P_{\text{bond}} \approx -D_{\text{mod}} \times \Delta y \times P_{\text{initial}}$, illustrating duration risk on debt holdings.
- **Equity Market Beta Shock:** $\Delta V_{\text{equity}} \approx \beta_{\text{portfolio}} \times \Delta I_{\text{NIFTY}}$, calculating hypothetical index co-movement.
- **Currency & Commodity Volatility:** Computes cross-asset portfolio impacts using historical correlation matrices.
- All scenario outputs return structured JSON containing labeled data series, upper and lower sensitivity bands, and mandatory educational disclaimers.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Service Scaffolding:** Create `services/gemini-advisor-service` with Poetry/uv, Python 3.12, and strict formatting tools (`ruff`, `mypy`).
2. **Configure Database & pgvector Extension:** Author Alembic migrations establishing tables `knowledge_documents`, `knowledge_chunks` with `vector(768)` HNSW cosine indexing, and `advisor_audit_logs`.
3. **Build Knowledge Ingestion Pipeline:** Create document chunking, semantic parsing, and embedding generator supporting PDF, Markdown, and HTML ingestion using Google `text-embedding-004`.
4. **Implement Hyperledger Besu Ledger Connector:** Build asynchronous Web3 client connecting to Besu RPC to query `ProofOfReserveRegistry.sol` for latest block height, Merkle root, and custody ratios.
5. **Implement Non-Advisory Guardrail Classifier:** Author semantic intent classifier using few-shot classification and regex safeguards to intercept advisory queries and prompt injections.
6. **Implement RAG Context Retriever:** Construct hybrid retriever executing cosine vector similarity search combined with PostgreSQL full-text search (BM25 ranking) over indexed regulatory corpora.
7. **Implement Prompt Synthesis & System Instructions Engine:** Design parameterized prompt templates enforcing non-advisory tone, mandatory SEBI disclaimers, 0.00% fee (No fee at all) breakdowns, and structured citation tags.
8. **Build Gemini 1.5 Flash & Pro Orchestrator:** Implement asynchronous streaming client with automatic model fallback, token usage tracking, and dynamic temperature controls.
9. **Implement Scenario Stress-Testing Engine:** Build mathematical modeling module calculating duration shocks, equity beta sensitivity, and inflation impacts, outputting structured JSON chart contracts.
10. **Build Document Summarizer Endpoint:** Create multi-part file upload handler utilizing Gemini 1.5 Pro's 1M context window to extract risk factors, management discussions, and financials from prospectuses.
11. **Implement Server-Sent Events (SSE) & WebSocket Endpoints:** Build streaming response handlers transmitting token deltas, inline citation chips, and interactive widget action triggers.
12. **Implement Redis Session Memory & Rate Limiting:** Build sliding-window conversational history manager and sliding-token rate limiter per investor account.
13. **Configure Telemetry, Metrics & Audit Logging:** Instrument Prometheus metrics (`gemini_query_latency_seconds`, `guardrail_interceptions_total`, `token_usage_total`) and publish anonymized audit records to Kafka.
14. **Write Comprehensive Automated Test Suite:** Implement unit and integration tests with `pytest-asyncio`, mocking Gemini API responses and verifying guardrail interception across 100+ adversarial test vectors.

## Interfaces / Contracts

### Protobuf Service Definition (`proto/growww/advisor/v1/advisor_service.proto`)
```protobuf
syntax = "proto3";

package growww.advisor.v1;

option go_package = "github.com/growww/proto/gen/go/advisor/v1;advisorv1";

service FinancialAdvisorService {
  rpc StreamFinancialQuery (FinancialQueryRequest) returns (stream FinancialQueryChunk);
  rpc SimulateScenarioShock (ScenarioShockRequest) returns (ScenarioShockResponse);
  rpc SummarizeDocument (SummarizeDocumentRequest) returns (SummarizeDocumentResponse);
  rpc GetProofOfReserveExplanation (ReserveExplanationRequest) returns (ReserveExplanationResponse);
}

message FinancialQueryRequest {
  string session_id = 1;
  string user_id = 2; // Pseudonymized UUID
  string query_text = 3;
  string current_screen_context = 4; // e.g. "SECURITY_DETAIL", "PORTFOLIO", "WALLET"
  string isin_context = 5; // Optional ISIN for asset-specific grounding
  bool enable_voice_response = 6;
}

message CitationMetadata {
  string source_title = 1;
  string source_url = 2;
  string document_type = 3; // "SEBI_CIRCULAR", "PROSPECTUS", "BESU_PROOF_OF_RESERVE"
  string on_chain_tx_hash = 4;
}

message InteractiveWidgetPayload {
  string widget_type = 1; // "CUSTODY_FLOWCHART", "SCENARIO_SLIDER", "FEE_BREAKDOWN_CARD"
  string json_parameters = 2;
}

message FinancialQueryChunk {
  string delta_text = 1;
  bool is_complete = 2;
  bool guardrail_intercepted = 3;
  repeated CitationMetadata citations = 4;
  InteractiveWidgetPayload embedded_widget = 5;
  int64 timestamp_unix_ms = 6;
}

message ScenarioShockRequest {
  string user_id = 1;
  double interest_rate_delta_bps = 2; // e.g. +100.0 bps
  double equity_market_shock_pct = 3; // e.g. -10.0%
  double commodity_shock_pct = 4; // e.g. +15.0%
}

message AssetShockImpact {
  string asset_class = 1; // "EQUITY", "DEBT", "COMMODITY"
  double current_allocation_pct = 2;
  double simulated_pnl_pct = 3;
  double simulated_impact_inr = 4;
  string educational_rationale = 5;
}

message ScenarioShockResponse {
  repeated AssetShockImpact asset_impacts = 1;
  double net_portfolio_simulated_pnl_pct = 2;
  string risk_summary_markdown = 3;
  string mandatory_disclaimer = 4;
}

message SummarizeDocumentRequest {
  string document_id = 1;
  string document_type = 2; // "DRHP", "ANNUAL_REPORT", "SCHEME_INFORMATION_DOCUMENT"
  repeated string focus_areas = 3; // "RISK_FACTORS", "PROMOTER_HOLDING", "USE_OF_PROCEEDS"
}

message SummarizeDocumentResponse {
  string executive_summary = 1;
  repeated string key_risk_factors = 2;
  string promoter_holding_analysis = 3;
  string financial_health_overview = 4;
  string document_sha256 = 5;
}

message ReserveExplanationRequest {
  string isin = 1;
  string asset_symbol = 2;
}

message ReserveExplanationResponse {
  string asset_symbol = 1;
  string depository_name = 2; // "NSDL" or "CDSL"
  string custody_account_reference = 3;
  double backing_ratio_percentage = 4; // Always 100.0%
  string latest_merkle_root = 5;
  uint64 besu_block_height = 6;
  string explanation_markdown = 7;
}
```

### PostgreSQL Database Schema DDL (`pgvector` Knowledge Store)
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

CREATE TYPE document_category_enum AS ENUM (
    'SEBI_REGULATION',
    'EXCHANGE_BYLAW',
    'SCHEME_INFORMATION_DOCUMENT',
    'RED_HERRING_PROSPECTUS',
    'ANNUAL_REPORT',
    'PLATFORM_GOVERNANCE',
    'PROOF_OF_RESERVE_SPEC'
);

CREATE TABLE knowledge_documents (
    document_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    category document_category_enum NOT NULL,
    isin VARCHAR(12),
    issuer_name VARCHAR(255),
    document_url VARCHAR(1000),
    sha256_checksum VARCHAR(64) NOT NULL UNIQUE,
    published_date DATE,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE knowledge_chunks (
    chunk_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(document_id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    chunk_text TEXT NOT NULL,
    token_count INT NOT NULL,
    embedding vector(768) NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_document_chunk UNIQUE(document_id, chunk_index)
);

CREATE INDEX idx_knowledge_chunks_embedding 
ON knowledge_chunks 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

CREATE INDEX idx_knowledge_chunks_text_search 
ON knowledge_chunks 
USING gin (to_tsvector('english', chunk_text));

CREATE TABLE advisor_audit_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(128) NOT NULL,
    user_id UUID NOT NULL,
    prompt_hash VARCHAR(64) NOT NULL,
    intent_detected VARCHAR(100) NOT NULL,
    guardrail_triggered BOOLEAN NOT NULL DEFAULT FALSE,
    guardrail_reason VARCHAR(255),
    model_used VARCHAR(50) NOT NULL,
    tokens_prompt INT NOT NULL,
    tokens_completion INT NOT NULL,
    latency_ms INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_advisor_audit_session ON advisor_audit_logs(session_id);
CREATE INDEX idx_advisor_audit_user ON advisor_audit_logs(user_id);
CREATE INDEX idx_advisor_audit_guardrail ON advisor_audit_logs(guardrail_triggered) WHERE guardrail_triggered = TRUE;
```

### JSON Event Schemas (`gemini.interactions.v1` and `gemini.guardrail_alerts.v1`)
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "GeminiInteractionEvent",
  "type": "object",
  "required": [
    "event_id",
    "session_id",
    "user_id",
    "query_category",
    "model_name",
    "latency_ms",
    "timestamp_unix_ms"
  ],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "session_id": { "type": "string" },
    "user_id": { "type": "string", "format": "uuid" },
    "query_category": { 
      "type": "string", 
      "enum": ["EDUCATION", "PROOF_OF_RESERVE", "FEE_BREAKDOWN", "SCENARIO_SHOCK", "DOCUMENT_SUMMARY", "ADVISORY_BLOCKED"] 
    },
    "model_name": { "type": "string" },
    "tokens_prompt": { "type": "integer" },
    "tokens_completion": { "type": "integer" },
    "latency_ms": { "type": "integer" },
    "citations_count": { "type": "integer" },
    "timestamp_unix_ms": { "type": "integer" }
  }
}
```

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "GeminiGuardrailAlertEvent",
  "type": "object",
  "required": [
    "alert_id",
    "session_id",
    "user_id",
    "violation_type",
    "sanitized_trigger_snippet",
    "action_taken",
    "timestamp_unix_ms"
  ],
  "properties": {
    "alert_id": { "type": "string", "format": "uuid" },
    "session_id": { "type": "string" },
    "user_id": { "type": "string", "format": "uuid" },
    "violation_type": {
      "type": "string",
      "enum": ["STOCK_TIP_SOLICITATION", "PRICE_PREDICTION_REQUEST", "PORTFOLIO_RECOMMENDATION", "PROMPT_INJECTION_ATTEMPT"]
    },
    "sanitized_trigger_snippet": { "type": "string" },
    "action_taken": { "type": "string", "enum": ["REJECT_WITH_EDUCATIONAL_DISCLAIMER", "SESSION_THROTTLED"] },
    "timestamp_unix_ms": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **SEBI Investment Advisers Regulations (2013) Compliance:** The assistant system prompt strictly prohibits generating investment recommendations, price targets, buy/sell calls, or subjective performance guarantees. All output contains mandatory statutory disclaimers: *"Educational analysis only. Not investment advice. Securities investments are subject to market risks."*
- **Zero PII Exposure to GenAI APIs:** All investor identifiers, actual account balances, tax identification numbers (PANs), and banking details are completely stripped and replaced with generalized asset class weightings before passing context to Google Gemini endpoints.
- **Strict Grounding & Hallucination Mitigation:** All factual assertions regarding custody, underlying depositories, corporate actions, and proof-of-reserve metrics must be grounded in verified RAG documents or real-time Hyperledger Besu smart contract queries, accompanied by explicit citation metadata.
- **Prompt Injection & Jailbreak Defense:** Multi-stage input sanitizer scans for jailbreak sequences (such as "Ignore previous instructions", "Act as an unrestricted financial advisor", or system prompt extraction attempts) and triggers instant rejection.
- **fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Invariant:** AI responses explaining trading costs must accurately represent the 0.00% (No fee at all) (0 bps (0.00% fee at launch)) platform fee structure with its exact 0.00% fee at launch (governed by FeeController.sol) distribution, confirming zero AUM fees and zero hidden commissions.

## Acceptance Criteria
- [ ] Asynchronous FastAPI service starts cleanly, exposing `/healthz`, `/livez`, and Prometheus metrics on `/metrics`.
- [ ] Dual-engine router selects `gemini-1.5-flash` for low-latency Q&A and `gemini-1.5-pro` for large-document analysis.
- [ ] RAG retriever over PostgreSQL `pgvector` returns top-5 relevant chunks with cosine similarity $\ge 0.78$ in $< 40\text{ms}$.
- [ ] On-chain context engine successfully queries `ProofOfReserveRegistry.sol` on Hyperledger Besu and returns verified Merkle roots.
- [ ] Non-advisory guardrail intercepts 100% of tested stock-tip requests ("Should I buy X?") and responds with standardized educational disclaimers.
- [ ] Server-Sent Events (SSE) streaming delivers initial token response within $< 500\text{ms}$ time-to-first-token (TTFT).
- [ ] Scenario shock endpoint produces deterministic mathematical impacts for interest rate and equity beta shifts without advisory bias.
- [ ] Document summarization endpoint processes 100-page DRHP PDF files and extracts key risk factors and financial balance sheets in $< 8\text{s}$.
- [ ] Platform fee inquiries accurately state the 0.00% transaction fee (No fee at all) with the 0.00% fee at launch (governed by FeeController.sol) breakdown.
- [ ] Unit and integration test suite achieves $\ge 85\%$ coverage across all routing, guardrail, and retrieval modules.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `103` (API Design Standards), Prompt `105` (Authentication & Authorization Architecture), Prompt `401` (PostgreSQL Database Schema).
- **Parallel Tasks:** Prompt `207` (Market Data Service), Prompt `209` (Portfolio & Holdings Service), Prompt `243` / `308` (Proof-of-Reserve Registry).
- **Downstream Blockers:** Prompt `515` (Flutter Gemini Conversational Assistant Screen), Prompt `601` (Web Trading Application).
