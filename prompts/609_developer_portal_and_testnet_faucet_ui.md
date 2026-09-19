# 609 - Public Developer Portal, Interactive API Explorer, Web3 Testnet Faucet & Mock Demat Faucet UI

## Purpose
Third-party fintech integrators, institutional algorithms, market makers, and ecosystem developers require a dedicated, self-service developer portal to build against the Growww platform. To accelerate partner onboarding and third-party algorithmic trading integrations within the SEBI Innovation Sandbox and IFSCA Regulatory Sandbox, developers must be able to explore OpenAPI 3.1 specifications, execute live sandbox API requests, register webhooks with HMAC-SHA256 signature verification, request testnet Gas and sandbox e-Rupee tokens, and airdrop fractional blue-chip demat equities into test portfolios without manual administrative intervention.

This prompt specifies the architecture and implementation of the Public Developer Portal (`apps/growww_web/app/(devportal)/`), Interactive API & FIX Explorer, Self-Service API Key & Webhook Manager, Permissioned Web3 Testnet Faucet, and Mock Demat Portfolio Faucet UI.

## What You Are Building
A Next.js 14-based developer platform featuring:
- `DeveloperPortalHome`: Comprehensive documentation hub with quickstart guides for Python, TypeScript, Go, and Dart SDKs, architectural overviews, FIX 4.4/5.0 SP2 specifications, and rate-limiting tier tables.
- `InteractiveApiExplorer`: Interactive OpenAPI 3.1 playground (custom-styled with Tailwind CSS and Scalar / Stoplight Elements) supporting sandbox Bearer authentication, header customization, live request execution against `https://sandbox-api.growww.in`, and multi-language code generation (cURL, Python `requests`/`httpx`, TypeScript `axios`/`fetch`, Go `net/http`).
- `ApiKeyAndWebhookDashboard`: Self-service dashboard for authenticated developers to generate sandbox API keys (`gw_test_...`), manage allowed IP CIDRs, configure webhook callback endpoints (`https://partner.example.com/webhooks/growww`), test webhook delivery with sample payloads, and inspect cryptographic signature verification headers (`X-Growww-Signature: sha256=...`).
- `Web3TestnetFaucetUI`: Web3-enabled faucet interface connected to the permissioned Hyperledger Besu UAT network. Integrators can input their Ethereum-compatible 0x address or connect a Web3 wallet (MetaMask, Rabby) to request testnet Gas tokens and Sandbox CBDC e-Rupee tokens (ERC-20 test tokens) capped at a fixed rate limit.
- `MockDematPortfolioFaucetUI`: Sandbox portfolio provisioning tool allowing developers to select blue-chip equity ISINs (Reliance `INE002A01018`, TCS `INE467B01029`, HDFC Bank `INE040A01034`), specify fractional quantities (e.g. 10.50000000 tokens), and trigger atomic depository demat locking and on-chain DvP token issuance in the regulatory sandbox.
- `FixProtocolGuide`: Interactive FIX protocol dictionary documenting FIX tag mappings (Tag 35 MsgType, Tag 55 Symbol, Tag 38 OrderQty, Tag 44 Price) and sandbox FIX Gateway connection parameters (Prompt 225).

## Scope Boundaries
- **In Scope:**
  - Developer portal UI, documentation layout, and interactive code snippet generation.
  - Interactive OpenAPI 3.1 explorer and sandbox API request runner.
  - API key generation and webhook registration dashboard for sandbox environments.
  - Web3 testnet gas and CBDC faucet interface for Hyperledger Besu.
  - Mock Demat fractional equity airdrop and portfolio seeder interface.
  - Rate limiting, CAPTCHA verification, and faucet quota management.
- **Out of Scope / Handled Elsewhere:**
  - Production institutional API key issuance and manual approval workflows (Prompt 605, Prompt 702).
  - Backend API Gateway and rate limiting enforcement (Prompt 219, Prompt 220).
  - Backend Sandbox Faucet microservice implementation (Prompt 203, Prompt 214).
  - FIX Gateway server implementation (Prompt 225).

## Technology to Use
- **Next.js 14 App Router (React Server Components & Client Components):** Renders high-performance documentation pages with SSG/ISR, dynamic API exploration, and client-side Web3 wallet connectivity.
- **Tailwind CSS 3.4+ & Radix UI / Shadcn UI:** Provides clean developer-centric UI components with high readability, syntax highlighting, and dark mode support.
- **Scalar / Stoplight Elements React:** Modern, accessible OpenAPI 3.1 interactive console embedded into the Next.js portal.
- **Viem 2.x & Wagmi 2.x:** Lightweight Web3 wallet connection and RPC client for Hyperledger Besu testnet faucet transactions.
- **Prism.js / Shiki:** Server-side and client-side code block syntax highlighting for Python, TypeScript, Go, Dart, and cURL snippets.
- **Lucide React & Sonner:** Clean developer icons and toast notifications for faucet confirmations and API key copy events.

## Backend / Infra Touchpoints
- **Sandbox API Gateway (`https://sandbox-api.growww.in`):** Destination for live requests executed from the Interactive API Explorer.
- **Developer Service (`/api/v1/developer/keys`, `/api/v1/developer/webhooks`):** Manages API key lifecycles and webhook subscription endpoints.
- **Sandbox Faucet Microservice (`/api/v1/sandbox/faucet/`):** Processes faucet claims for testnet INR cash balances and mock demat shares.
- **Cloudflare Turnstile:** Invisible CAPTCHA protecting public faucet and API key generation endpoints against automated abuse.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu (Permissioned Consortium UAT Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
  - `TestnetFaucetRelayer.sol`: Dispenses testnet Gas tokens and Sandbox CBDC e-Rupee tokens to the requested 0x developer address upon verifying backend authorization signatures.
  - `DvPSettlementEngine.sol`: Interfaced during mock demat airdrops to simulate atomic delivery-versus-payment issuance of fractional equity tokens.
  - `ProofOfReserveRegistry.sol`: Displays live sandbox proof-of-reserve metrics on the faucet interface.
- **Security & Rate Limiting:** Faucet claims are rate-limited to 5 requests per hour per 0x address and IP address, signed by a backend relayer key to prevent testnet gas drainage.

## Step-by-Step Build Instructions
1. Scaffold developer portal route hierarchy under `apps/growww_web/app/(devportal)/`:
   - `layout.tsx`: Developer navigation header, search bar, documentation sidebar, and wallet connect button.
   - `page.tsx`: Developer portal homepage with quickstart cards, SDK links, and architectural guides.
   - `docs/[[...slug]]/page.tsx`: Dynamic markdown documentation reader rendering API architecture, auth guides, and FIX specs.
   - `api-explorer/page.tsx`: Interactive OpenAPI 3.1 console powered by Scalar.
   - `dashboard/keys/page.tsx`: Self-service Sandbox API key manager.
   - `dashboard/webhooks/page.tsx`: Webhook endpoint manager and delivery log viewer.
   - `faucet/web3/page.tsx`: Web3 Gas & CBDC Testnet Faucet.
   - `faucet/demat/page.tsx`: Mock Demat fractional portfolio seeder UI.
2. Build Developer Portal Navigation Header with quick links (Docs, API Explorer, Faucet, Dashboard, GitHub SDKs) and Dark/Light theme toggle.
3. Integrate OpenAPI 3.1 Explorer using Scalar:
   - Ingest OpenAPI JSON schema from `/api/v1/openapi.json`.
   - Configure sandbox base URL `https://sandbox-api.growww.in`.
   - Provide interactive headers editor (`Authorization: Bearer <token>`, `X-Idempotency-Key`).
   - Implement multi-language code generator tabs (cURL, Python httpx, TypeScript axios, Go).
4. Implement Sandbox API Key Management UI:
   - Create new key modal with name, environment tag, and IP whitelist input.
   - Display newly generated API Key secret once with high-visibility copy button and security warning.
   - List active keys with creation date, last used timestamp, IP restrictions, and one-click Revoke action.
5. Implement Webhook Subscription Dashboard:
   - Form to register webhook URLs (e.g. `https://api.partner.com/events`) and select event topics (`order.filled`, `settlement.completed`, `margin.alert`, `corporate_action.announced`).
   - Webhook signing secret generator with HMAC-SHA256 signature verification code samples.
   - "Send Test Event" trigger and live webhook delivery logs showing HTTP status codes, latency, and payload payloads.
6. Build Web3 Testnet Faucet Interface:
   - Connect Web3 wallet (MetaMask, Rabby) or manual 0x address input field.
   - Display Besu UAT network status, current block height, and faucet contract health.
   - Asset selector: Testnet Gas Tokens (0.1 BESU) or Sandbox CBDC e-Rupee (₹50,000 test tokens).
   - Integrate Cloudflare Turnstile CAPTCHA challenge before submission.
   - Execute claim via `POST /api/v1/sandbox/faucet/web3` and display transaction hash link to the Besu block explorer.
7. Build Mock Demat Portfolio Faucet UI:
   - Investor account ID or ledger address selector.
   - Equity ISIN selector dropdown (Reliance, TCS, HDFC Bank, Infosys, ICICI Bank, Bharti Airtel).
   - Fractional share quantity input with predefined chips (+1.0, +5.0, +10.0, +50.0 shares).
   - Test INR Cash Balance slider (up to ₹10,00,000).
   - "Seed Mock Portfolio" CTA button dispatching `POST /api/v1/sandbox/faucet/demat`.
   - Render animated confirmation card showing simulated CDSL depository holding receipt and Besu DvP token mint event.
8. Build Interactive FIX Protocol Guide:
   - Document FIX 4.4 and 5.0 SP2 logon sequence, Tag definitions, order placement messages (`NewOrderSingle - 35=D`), and execution reports (`ExecutionReport - 35=8`).
   - Interactive FIX message parser highlighting tag meanings.
9. Implement Rate Limiting and Quota Indicators:
   - Display real-time sandbox API rate limit usage (e.g. 45/100 requests per second used).
   - Display remaining daily faucet quota for the connected developer account.
10. Write end-to-end Playwright tests verifying documentation navigation, API explorer code snippet rendering, API key generation flow, and faucet claim execution.

## Interfaces / Contracts

### Developer Portal DTO Schemas
```typescript
export interface SandboxApiKey {
  id: string;
  name: string;
  keyPrefix: string; // e.g. "gw_test_8f2a..."
  createdAt: string;
  lastUsedAt: string | null;
  allowedIpCidrs: string[];
  status: 'ACTIVE' | 'REVOKED';
}

export interface CreateApiKeyRequest {
  name: string;
  allowedIpCidrs?: string[];
}

export interface CreateApiKeyResponse {
  apiKey: SandboxApiKey;
  secretKey: string; // Displayed only once upon creation
}

export interface WebhookEndpoint {
  id: string;
  url: string;
  subscribedEvents: string[]; // e.g. ['order.filled', 'settlement.completed']
  secret: string;
  status: 'ACTIVE' | 'DISABLED';
  createdAt: string;
}

export interface WebhookDeliveryLog {
  id: string;
  webhookId: string;
  event: string;
  statusCode: number;
  requestPayload: Record<string, any>;
  responseBody: string;
  durationMs: number;
  timestamp: string;
  signatureHeader: string;
}

export interface Web3FaucetClaimRequest {
  recipientAddress: string; // 0x Ethereum-compatible Besu address
  assetType: 'GAS_TOKEN' | 'SANDBOX_CBDC_INR';
  turnstileToken: string;
}

export interface MockDematFaucetClaimRequest {
  accountNumber: string;
  inrCashAmount: number;
  equities: Array<{
    isin: string;
    symbol: string;
    quantity: number;
  }>;
  turnstileToken: string;
}
```

### Webhook Verification Code Sample (TypeScript / Node.js)
```typescript
import crypto from 'crypto';

export function verifyGrowwwWebhookSignature(
  payload: string,
  signatureHeader: string, // "sha256=<hex_digest>"
  webhookSecret: string
): boolean {
  const [algorithm, signature] = signatureHeader.split('=');
  if (algorithm !== 'sha256' || !signature) {
    return false;
  }

  const expectedSignature = crypto
    .createHmac('sha256', webhookSecret)
    .update(payload)
    .digest('hex');

  return crypto.timingSafeEqual(
    Buffer.from(signature, 'utf8'),
    Buffer.from(expectedSignature, 'utf8')
  );
}
```

## Security & Compliance Notes
- **Strict Sandbox Isolation:** The developer portal and all interactive faucets operate exclusively against the isolated UAT / Sandbox environment (`sandbox-api.growww.in`, `sandbox-rpc.growww.in`). Sandbox API keys and mock demat tokens have zero validity or value on production trading rails.
- **One-Time Secret Visibility:** API keys and webhook signing secrets are displayed exactly once upon creation. Only cryptographic hashes are stored in the database.
- **Anti-Abuse & Rate Limiting:** Public faucet endpoints enforce strict rate limiting (Cloudflare Turnstile CAPTCHA + 5 requests per hour per IP/address) to prevent Denial of Service on sandbox nodes.
- **Timing-Safe Webhook Signatures:** All documentation and sample code must emphasize timing-safe byte comparisons (`crypto.timingSafeEqual`) to protect partner systems against timing attacks.

## Acceptance Criteria
- [ ] Developer Portal homepage loads with sub-second response times and complete SDK documentation.
- [ ] Interactive API Explorer parses the OpenAPI 3.1 specification and executes live test requests against sandbox endpoints.
- [ ] Developers can generate, name, inspect, and revoke Sandbox API keys with IP whitelisting.
- [ ] Webhook manager allows configuring endpoints, selecting event types, testing delivery, and inspecting signatures.
- [ ] Web3 Testnet Faucet successfully dispenses testnet Gas and sandbox e-Rupee tokens with clear Besu transaction hash links.
- [ ] Mock Demat Faucet seeds fractional blue-chip equity tokens and testnet INR balances into the developer's sandbox portfolio.
- [ ] Prominent sandbox disclaimer banners prevent any confusion between testnet assets and real capital.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 000 (North Star), Prompt 103 (API Standards), Prompt 219 (API Gateway), Prompt 601 (Web App Scaffolding).
- **Parallel Tasks:** Prompt 225 (FIX Gateway), Prompt 527 (Flutter Environment Switcher & Sandbox Mode).
- **Downstream Blockers:** Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios), Prompt 907 (Regulatory Sandbox Pilot Launch).
