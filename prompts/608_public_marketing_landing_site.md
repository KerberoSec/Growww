# 608 - Public Landing Page, Education Hub & Risk Disclosures

## Purpose
Creates the public-facing, SEO-optimized marketing web portal, investor education academy, and statutory regulatory disclosure center (`apps/growww_web/app/(marketing)/` or `apps/growww_marketing/`). Educates Indian and international investors on the revolutionary benefits of 1:1 real-security-backed fractional equity investing on a transparent permissioned blockchain ledger. Clearly communicates the transparent Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol) (with zero holding/AUM fees and FIFO capital gains computed strictly for user tax compliance under Section 111A/112A), showcases live proof-of-reserve transparency, and prominently displays all SEBI-mandated investor risk disclosures and grievance redressal mechanisms.

## What You Are Building
- High-converting, responsive marketing landing pages (`apps/growww_web/app/(marketing)/page.tsx`).
- Hero Section with interactive visual explaining fractional tokenization backed 1:1 by NSDL/CDSL custody shares.
- Interactive Fee Comparison Calculator: visually contrasting traditional broker high turnover/AUM charges vs Growww's 0.00% platform fee (No fee at all) model with zero custody or holding charges.
- "How It Works" interactive step-by-step explainer (1. Bank Rails $\to$ 2. Custodian Demat Lock $\to$ 3. Besu DvP Issuance $\to$ 4. Instant Settlement).
- Live Proof-of-Reserve Trust Widget displaying real-time 1:1 custody backing counter.
- Supported Blue-Chip Equities Showcase carousel with live indicative prices.
- Investor Education Academy & Glossary Hub (`/education`, `/glossary`) explaining fractional shares, DvP, custody vs tokens, and regulatory safeguards.
- SEBI & RBI Mandatory Risk Disclosure Banners, Disclaimer Modals, and Grievance Redressal / SCORES Escalation Matrix (`/grievances`).

## Scope Boundaries
- **In Scope:**
 - Public marketing UI, landing page hero, feature sections, and interactive calculators.
 - SEO metadata optimization, OpenGraph cards, and schema.org JSON-LD structured data.
 - Interactive "How It Works" animated diagrams and fee sliders.
 - Educational glossary directory with search and alphabetical filtering.
 - Mandatory SEBI statutory risk disclosure banners, modals, and grievance redressal pages.
- **Out of Scope / Handled Elsewhere:**
 - Authenticated investor trading application (Prompt 603).
 - Web onboarding and KYC wizard (Prompt 602).
 - Marketing tracking and client-side crash analytics integrations (Prompt 524).
 - Fee calculation backend microservice (Prompt 210).

## Technology to Use
- **Next.js 14 App Router (Static Site Generation [SSG] & Incremental Static Regeneration [ISR]):** Generates pre-rendered HTML at build time with revalidation for sub-second page loads, near-perfect Lighthouse scores (100/100), and optimal Google search indexing.
- **Tailwind CSS 3.4+ & Framer Motion:** Delivers fluid scroll-triggered animations, interactive slider transitions, and responsive mobile-first layouts matching the Growww design tokens.
- **Next.js Metadata API & `schema-dts`:** Manages dynamic OpenGraph/Twitter social cards and rich JSON-LD structured data (FinancialProduct, Organization, FAQPage).
- **Lucide React & Canvas Confetti:** Modern financial iconography and interactive celebration effects on fee savings calculator.
- **Viem 2.x (Read-Only):** Lightweight public RPC client to fetch real-time custody backing stats directly from `ProofOfReserveRegistry.sol`.

## Backend / Infra Touchpoints
- **Public Platform Stats API (`/api/v1/public/stats`):** Fetches live aggregate 24h trading volume, total tokenized equities, and active investor count.
- **Public Proof-of-Reserve API (Prompt 215):** Fetches current custody backing ratio.
- **Cloudflare Edge CDN / Vercel:** Distributes globally cached static assets with edge caching and SSL termination.

## Blockchain Interaction
- **Live Custody Reserve Feed:** Embeds a lightweight read-only RPC call to `ProofOfReserveRegistry.sol` on the permissioned Hyperledger Besu network to query total physical shares in custody and display a live "100.00% Backed" badge on the landing page hero without requiring user wallet connection.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
 - `ProofOfReserveRegistry.sol` (Queries `getGlobalReserveHealth()` returning total shares vs total tokens).
- **Security:** Public read-only queries are routed through rate-limited public RPC endpoints or cached edge endpoints; no client-side transaction signing occurs on marketing routes.

## Step-by-Step Build Instructions
1. Scaffold public marketing route structure (`app/(marketing)/layout.tsx`, `app/(marketing)/page.tsx`, `app/(marketing)/education/page.tsx`, `app/(marketing)/glossary/page.tsx`, `app/(marketing)/grievances/page.tsx`).
2. Build Public Navigation Bar with Growww logo, navigation links (How It Works, Proof-of-Reserve, Fee Model, Education), Theme Switcher, and CTA buttons ("Log In", "Get Started").
3. Implement Hero Section featuring high-impact headline, animated value propositions, and dynamic call-to-action button linking directly to onboarding (Prompt 602).
4. Build Interactive Ultra-Low Fee Calculator:
 - Slider 1: Investment / Trade Turnover Amount (₹500 to ₹10,00,000).
 - Slider 2: Investment Duration (1 month to 5 years).
 - Slider 3: Expected Annual Return % (-20% to +100%).
 - Dynamic Display: Highlights exact 0.00% platform fee (No fee at all) on trade turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), ₹0 recurring holding/AUM charges, and explains FIFO capital gains computed strictly for user tax compliance (Section 111A/112A), contrasting with conventional brokerage charges.
5. Implement Animated "How It Works" 4-Step Interactive Visualizer using Framer Motion tabs:
 - Step 1: Fund in INR via UPI / NetBanking.
 - Step 2: Physical shares purchased and locked in SEBI-registered Depository (NSDL/CDSL).
 - Step 3: Fractional Digital Security Tokens issued on Hyperledger Besu.
 - Step 4: Instant Atomic Delivery-versus-Payment (DvP) trade settlement.
6. Build Live Proof-of-Reserve Trust Card embedding live on-chain stats (Total Custody Backing, Active ISINs, Last Depository Attestation time).
7. Build Featured Equities Carousel showcasing top Indian stocks (e.g. Reliance, TCS, HDFC Bank, Infosys) with fractional unit price examples (e.g. "Buy ₹500 of Reliance").
8. Build Investor Education Academy (`/education`) with categorized articles on fractional shares, regulatory protections, and blockchain security.
9. Implement Interactive Domain Glossary (`/glossary`) derived from Prompt 001 with search bar and A-Z alphabetical index.
10. Build SEBI Mandatory Risk Disclosure Banner and persistent footer containing statutory disclaimers regarding market risks and sandbox parameters.
11. Build Grievance Redressal & Investor Charter Page (`/grievances`) detailing Level 1 (Customer Support), Level 2 (Head of Compliance), Level 3 (SEBI SCORES / SMART ODR Portal) escalation pathways.
12. Add JSON-LD Structured Data for SEO (Organization, WebSite, FAQPage, FinancialProduct) in `layout.tsx`.
13. Run Google Lighthouse audit to verify Performance score 100, Accessibility 100, Best Practices 100, and SEO 100.

## Interfaces / Contracts
```typescript
export interface FeeCalculationInput {
  tradeTurnoverInr: number;
  holdingPeriodMonths: number;
  estimatedCapitalGainInr: number;
}

export interface FeeCalculationBreakdown {
  grossTurnoverInr: number;
  growwwPlatformFeeInr: number; // Fixed 0.00% (Zero Fee) on turnover
  treasurySplitInr: number;     // Governed by FeeController (0.00% at launch)
  coreSgfSplitInr: number;      // Governed by FeeController (0.00% at launch)
  ipfSplitInr: number;          // Governed by FeeController (0.00% at launch)
  conventionalBrokerEstimateInr: number; // High AUM or brokerage fees
  investorSavingsInr: number;
  taxComplianceSTCG_LTCG_Inr: number; // FIFO tax compliance Section 111A/112A
}

export interface PlatformLiveStats {
  totalReserveRatioPercent: number; // e.g. 100.00
  totalCustodyValueInr: string;
  totalTokenizedSecuritiesCount: number;
  lastAttestationTime: string;
  besuBlockHeight: number;
}

export interface GlossaryEntry {
  term: string;
  slug: string;
  category: 'REGULATORY' | 'BLOCKCHAIN' | 'TRADING' | 'CUSTODY';
  shortDefinition: string;
  fullExplanation: string;
  relatedTerms: string[];
}

export interface GrievanceEscalationTier {
  level: number;
  designation: string;
  contactPerson: string;
  email: string;
  phone: string;
  turnaroundTimeDays: number;
  externalPortalUrl?: string;
}
```

## Security & Compliance Notes
- **SEBI Advertising Code Compliance:** All marketing content strictly complies with SEBI circulars: no guarantees of assured returns, prominent equal-size font for risk disclaimers ("Investments in securities market are subject to market risks; read all scheme related documents carefully before investing").
- **Privacy & Cookies:** Cookie banner fully compliant with Digital Personal Data Protection Act (DPDP Act 2023).
- **Strict Content Security Policy (CSP):** No unauthorized third-party tracking pixels or non-whitelisted CDN scripts allowed.

## Acceptance Criteria
- [ ] Marketing landing page renders responsively across mobile, tablet, and ultra-wide desktop viewports.
- [ ] Interactive fee calculator accurately computes Universal Zero-Fee Model (0.00% fee - No fee at all)s on turnover, demonstrates zero holding/AUM fees, and clarifies FIFO capital gains for tax compliance.
- [ ] "How It Works" interactive component transitions smoothly with clear visual diagrams.
- [ ] Live proof-of-reserve trust card displays live custody statistics fetched via public API or Besu RPC.
- [ ] Glossary directory allows searching and filtering all financial and blockchain domain terms.
- [ ] Grievance redressal page provides accurate multi-level escalation contacts including SEBI SCORES portal links.
- [ ] Structured JSON-LD metadata validates without errors on Google Rich Results Test tool.
- [ ] Google Lighthouse achieves 95+ across Performance, Accessibility, Best Practices, and SEO.

## Suggested Order / Dependencies
- **Prerequisites:** 000 (North Star), 001 (Glossary), 006 (Fee Model), 007 (Proof of Reserve), 009 (Risk Disclosures), 601 (Web Scaffolding).
- **Direct Successors / Parallel:** 602 (Web KYC), 603 (Web Trading Dashboard), 606 (PoR Dashboard).
