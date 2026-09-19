# 601 - Next.js 14 Investor Web App Scaffolding

## Purpose
Establishes the foundational enterprise web client architecture for Growww (`apps/growww_web/`). Provides an institutional-grade, highly responsive, accessibility-compliant (WCAG 2.1 AA) Next.js 14 web trading application utilizing App Router, React Server Components (RSC), Tailwind CSS, shadcn/ui design system, and Web3 connectivity infrastructure (Viem/Wagmi). This application serves as the primary browser-based portal for domestic Indian and international retail and institutional investors to execute fractional equity trades, monitor holdings backed 1:1 by SEBI-registered custody assets, and verify cryptographic proof-of-reserve records.

## What You Are Building
- Complete Next.js 14 monorepo application structure (`apps/growww_web/`) utilizing App Router (`app/`).
- Core layout architectures: Authenticated Investor Shell (`(app)`), Unauthenticated Auth Shell (`(auth)`), Public Portal Shell (`(marketing)`), and global navigation ribbons.
- Unified design system synchronizing design tokens (colors, typography, spacing, border radii, dark/light theme) with the Flutter multi-platform client design tokens.
- State management and data-fetching architecture using TanStack Query v5, Zustand, and Server Actions.
- Web3 connection and provider layer using Viem, Wagmi, and Web3Modal for Hyperledger Besu JSON-RPC node communication and Merkle proof verification.
- Session and authentication middleware handling JWT tokens, refresh rotation, CSRF protection, and route guarding.

## Scope Boundaries
- **In Scope:**
 - Next.js 14 project scaffolding, App Router structure, root and nested layouts.
 - Tailwind CSS theming matching Flutter design system tokens (Prompt 503).
 - Web3 provider configuration (Viem, Wagmi, Web3Modal) for permissioned Besu chain interaction.
 - Global state management (Zustand) and server state caching (TanStack Query v5).
 - Next.js Edge Middleware for authentication token verification and secure route gating.
 - Typed REST / gRPC-Web API client wrapper with automatic token refresh interceptors.
 - WebSocket client manager for streaming market data with automatic reconnection.
 - Global error boundaries, loading skeletons, and accessibility features.
- **Out of Scope / Handled Elsewhere:**
 - Onboarding and KYC multi-step wizard UI (Prompt 602).
 - Trading terminal and charting integration (Prompt 603).
 - Admin and back-office console applications (Prompts 604-607).
 - Backend API Gateway and BFF microservices (Prompt 219).

## Technology to Use
- **Next.js 14 (App Router, React 18/19 Server Components, TypeScript 5.4+):** Next.js delivers industry-leading server-side rendering (SSR), optimized bundle splitting, server components for minimal client-side JavaScript execution, and enterprise-grade SEO capabilities for public routes.
- **Tailwind CSS 3.4+ & shadcn/ui (Radix UI primitives):** Provides headless, fully accessible UI components styled via utility classes that mirror the Flutter design token palette.
- **TanStack Query v5 (React Query):** Manages asynchronous server state caching, background refetching, and optimistic updates.
- **Zustand 4.5+:** Lightweight client UI state management for modal visibility, draft order tickets, and user preferences.
- **Viem 2.x & Wagmi 2.x with Web3Modal / AppKit:** Powers EVM-compatible Web3 connectivity for permissioned Hyperledger Besu consortium RPC queries and client-side cryptographic proof verification.
- **Axios / Fetch with Ky:** Typed HTTP client configured with interceptors for mTLS-backed BFF gateway communication.

## Backend / Infra Touchpoints
- **API Gateway & BFF Layer (Prompt 219):** Interacts via REST (`/api/v1/*`) and gRPC-Web for user profiles, portfolio queries, and order routing.
- **Market Data WebSocket Service (Prompt 207):** Connects to `wss://api.growww.in/ws/v1/market` for real-time order book and price streaming.
- **User & Identity Service (Prompt 201):** Interacts for session validation, MPIN/WebAuthn authentication, and token refresh.
- **Hosting Infrastructure (Prompt 802):** Containerized Docker deployment behind Cloudflare CDN and Kubernetes Ingress.

## Blockchain Interaction
- **Permissioned Chain RPC:** Configures Viem/Wagmi transport providers connecting to the Growww Hyperledger Besu consortium RPC endpoint (`https://besu-rpc.growww.in`).
- **Read-Only Verification Hooks:** Sets up typed contract hooks for `DigitalSecurityToken.sol` and `ProofOfReserveRegistry.sol` to enable client-side cryptographic verification of holdings and Merkle proofs.
- **RPC Failover & Subscriptions:** Implements fallback RPC transport nodes and WebSocket block header subscriptions for live on-chain event detection.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus, 2-second block finality, Chain ID: 13371).
- **Core Smart Contracts Interfaced:**
 - `DigitalSecurityToken.sol` (ERC-3643 compliant permissioned asset token with automated whitelist gates).
 - `ProofOfReserveRegistry.sol` (Cryptographic Merkle root attestation of 1:1 physical share backing in NSDL/CDSL custody).
- **Client Security:** Read-only client connections for retail web users; private keys are never stored in browser storage for trading actions (fiat-backed DvP settlements are orchestrated via backend HSM relayers).

## Step-by-Step Build Instructions
1. Initialize a modern Next.js 14 monorepo application under `apps/growww_web/` using TypeScript, ESLint (strict config), and Prettier.
2. Configure `tailwind.config.ts` defining Growww brand color palettes (`brand-navy`, `brand-green`, `brand-gold`, `surface-dark`, `surface-light`), typography scales, and animation keyframes matching Flutter design tokens (Prompt 503).
3. Initialize `shadcn/ui` with Radix UI primitives (Button, Dialog, DropdownMenu, Tabs, Toast, Form, Input, Skeleton).
4. Create root layout (`app/layout.tsx`) wrapping the application in `ThemeProvider` (next-themes), `QueryClientProvider` (TanStack Query), and `WagmiProvider` (Viem/Wagmi).
5. Configure custom Wagmi chain definition for Hyperledger Besu (`besuConsortiumChain`) with fallback HTTP/WebSocket transports.
6. Implement Next.js Edge Middleware (`middleware.ts`) to validate JWT session tokens from HttpOnly cookies, redirect unauthenticated users to `/auth/login`, and inject security headers.
7. Build the authenticated application shell layout (`app/(app)/layout.tsx`) featuring a responsive collapsible sidebar, top market ribbon, notifications bell, and user profile drawer.
8. Build the unauthenticated authentication layout (`app/(auth)/layout.tsx`) with split-screen branding, investor risk warnings, and card container.
9. Implement a typed API client (`lib/api-client.ts`) utilizing Fetch/Ky with automatic Authorization header injection, 401 token refresh interception, and structured error parsing.
10. Build a resilient WebSocket connection manager (`lib/websocket-client.ts`) with exponential backoff reconnects, heartbeat ping/pong, and topic multiplexing.
11. Implement Zustand stores (`stores/ui-store.ts`, `stores/market-store.ts`) for managing drawer states, active ticker selections, and notification feeds.
12. Create standardized error boundary (`app/error.tsx`), not-found page (`app/not-found.tsx`), and global loading skeletons (`app/loading.tsx`).
13. Set up unit testing with Vitest / React Testing Library and end-to-end test harness with Playwright (`playwright.config.ts`).

## Interfaces / Contracts
```typescript
import { type Chain } from 'viem';

export const besuConsortiumChain: Chain = {
  id: 13371,
  name: 'Growww Besu Consortium',
  nativeCurrency: { name: 'Growww Gas', symbol: 'GAS', decimals: 18 },
  rpcUrls: {
    default: {
      http: ['https://besu-rpc.growww.in'],
      webSocket: ['wss://besu-ws.growww.in'],
    },
    public: {
      http: ['https://besu-rpc.growww.in'],
      webSocket: ['wss://besu-ws.growww.in'],
    },
  },
  blockExplorers: {
    default: { name: 'Growww Explorer', url: 'https://explorer.growww.in' },
  },
};

export interface UserSession {
  userId: string;
  email: string;
  phone: string;
  kycStatus: 'NOT_STARTED' | 'PENDING' | 'APPROVED' | 'REJECTED';
  kycTier: number;
  walletAddress?: `0x${string}`;
  role: 'INVESTOR' | 'INSTITUTIONAL';
  sessionExpiry: number;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
  timestamp: string;
}

export interface WebSocketMessage<T = unknown> {
  topic: 'market_data' | 'order_updates' | 'por_attestations';
  event: string;
  payload: T;
  sequence: number;
}
```

## Security & Compliance Notes
- **Strict Content Security Policy (CSP):** Enforces `default-src 'self'`, `connect-src 'self' https://api.growww.in wss://api.growww.in https://besu-rpc.growww.in`, `frame-ancestors 'none'`, and `object-src 'none'`.
- **Session Protection:** Auth tokens stored in secure `HttpOnly`, `SameSite=Strict`, `Secure` cookies; short-lived access tokens kept in memory only.
- **XSS & Injection Protection:** Strict React JSX escaping and DOMPurify for any sanitized HTML content.
- **Compliance Disclaimers:** Persistent SEBI-mandated risk warning badge in application footer.

## Acceptance Criteria
- [ ] Next.js 14 web app scaffolding initialized under `apps/growww_web/` with zero TypeScript or ESLint errors.
- [ ] Root layout properly integrates ThemeProvider, WagmiProvider, and TanStack QueryClientProvider.
- [ ] Custom Hyperledger Besu consortium chain configured and successfully connects to RPC endpoint.
- [ ] Edge middleware intercepts unauthenticated requests to `/app/*` and redirects to `/auth/login`.
- [ ] Design tokens perfectly align with Flutter client theme specifications (Prompt 503).
- [ ] API client successfully executes token refresh on simulated 401 response without dropping pending requests.
- [ ] WebSocket client automatically recovers connection after simulated network drop.
- [ ] Lighthouse Performance score >90, Accessibility score >95, and zero high-severity audit warnings.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (System Architecture), 103 (API Standards), 105 (Auth Architecture), 106 (Monorepo Layout), 503 (Design System).
- **Direct Successors / Parallel:** 602 (Web KYC Flow), 603 (Web Trading Dashboard), 608 (Marketing Site).
