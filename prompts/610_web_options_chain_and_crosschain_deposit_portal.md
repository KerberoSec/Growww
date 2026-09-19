# 610 - Web Options Chain, Strategy Builder & Multi-Chain Web3 Deposit Portal

## Purpose
Democratizing access to institutional-grade equity derivatives and cross-border capital deployment requires a high-performance, real-time web interface tailored for professional traders and global Web3 investors. Under SEBI and IFSCA GIFT City regulatory frameworks, investors require advanced tooling to analyze options market depth, construct multi-leg hedging strategies, visualize volatility surfaces across strike-expiry dimensions, and seamlessly deposit collateral across heterogeneous blockchain networks.

This prompt specifies the architecture and implementation of the Next.js 14 Web Options Trading Terminal and Multi-Chain Web3 Deposit Portal (`apps/growww_web/app/(app)/options/` and `apps/growww_web/app/(app)/deposit/web3/`). The module provides an interactive options matrix with real-time Black-Scholes Greeks, a multi-leg strategy builder (Straddle, Strangle, Iron Condor, Bull/Bear Spreads) with interactive payoff diagrams, 2D/3D volatility surfaces, and a multi-chain Web3 wallet connect portal supporting MetaMask, Phantom, Leather, and UniSat.

## What You Are Building
A high-performance Next.js 14 web derivatives terminal and multi-chain deposit module featuring:
- `OptionsChainMatrix`: Real-time, virtualized options matrix displaying Calls on the left, Puts on the right, and central Strike ladders, with dynamic expiry selectors, ITM/ATM/OTM heatmap shading, open interest (OI) bars, volume indicators, and live Black-Scholes Greeks (Delta, Gamma, Theta, Vega, Rho, Implied Volatility).
- `OptionsStrategyBuilder`: Interactive multi-leg strategy creation workspace supporting standard templates (Long/Short Straddle, Strangle, Iron Condor, Iron Butterfly, Bull Call Spread, Bear Put Spread, Ratio Spreads, Calendar Spreads, Collars, Covered Calls) and custom multi-leg configurations.
- `PayoffVisualizer`: Interactive HTML5 Canvas / SVG payoff chart rendering theoretical P&L curves at expiration and target dates, identifying break-even points, max profit, max loss, net credit/debit, and aggregate strategy Greeks.
- `AtomicMultiLegOrderTicket`: Execution ticket orchestrating single-click atomic placement of multi-leg derivative orders with customizable limit prices, slippage tolerance, and pre-trade margin preview.
- `VolatilitySurfaceVisualizer`: 2D Volatility Smile/Skew charts and interactive 3D Volatility Surfaces (Strike vs Expiry vs Implied Volatility) rendered with WebGL / Three.js, supporting camera rotation, zoom, cross-section slicing, and historical IV comparison.
- `MultiChainDepositPortal`: Enterprise Web3 wallet connection hub integrating EVM (MetaMask, Rabby, Coinbase Wallet), Solana (Phantom, Solflare), and Bitcoin/Stacks (Leather, UniSat) wallets for cross-chain stablecoin and native collateral deposits into GIFT City settlement vaults.
- `CrossChainDepositTracker`: Real-time transaction tracker visualizing cross-chain bridge progress, block confirmations, AML screening checks, and final atomic minting/credit on the Growww Hyperledger Besu settlement ledger.

## Scope Boundaries
- **In Scope:**
  - Responsive options chain matrix layout with high-frequency WebSocket updates.
  - Client-side Black-Scholes Greeks calculation fallback and live server stream ingestion.
  - Multi-leg strategy builder, payoff curve simulation, and aggregate Greeks aggregation.
  - 2D Volatility Smile charts and hardware-accelerated 3D Volatility Surface rendering.
  - Multi-wallet connector integration across EVM, Solana, and Bitcoin/Stacks ecosystems.
  - Deposit quote aggregation, fee/slippage calculation, and step-by-step transaction orchestration.
  - Pre-deposit address screening state management and compliance warning modals.
- **Out of Scope / Handled Elsewhere:**
  - Core order matching engine for options contracts (Prompt 205).
  - Pre-trade SPAN/VaR margin calculation engine backend (Prompt 206, Prompt 229).
  - Real-time market data WebSocket streaming server (Prompt 207).
  - Foreign investor funding and FX conversion microservice (Prompt 214).
  - Cross-chain custody bridges and deposit verification smart contracts (Prompt 313, Prompt 319).
  - Sanctions screening backend integration (Prompt 703).

## Technology to Use
- **Next.js 14 App Router (React Server Components & Client Components):** Delivers server-rendered initial shell and streaming client components for dynamic options pricing and WebGL canvases.
- **Tailwind CSS 3.4+ & shadcn/ui (Radix UI primitives):** Provides an institutional dark-themed trading interface with precision controls, sliders, steppers, and accessible data grids.
- **`@tanstack/react-virtual`:** Virtualizes large options chain grids (100+ strikes across multiple expiries) maintaining stable 60 FPS under continuous market data bursts.
- **Three.js / React Three Fiber / Drei:** Powers hardware-accelerated 3D Volatility Surface rendering with custom shaders, dynamic lighting, wireframe toggles, and orbit controls.
- **TradingView Lightweight Charts / D3.js:** Renders interactive 2D volatility smile curves and multi-leg P&L payoff graphs with dynamic cursor tracking.
- **Web3 Wallet Adapters & Libraries:**
  - **EVM:** Viem 2.x, Wagmi 2.x, and AppKit (Web3Modal) for MetaMask, Rabby, and WalletConnect v2.
  - **Solana:** `@solana/wallet-adapter-react`, `@solana/wallet-adapter-wallets` for Phantom and Solflare.
  - **Bitcoin / Stacks:** `@leather.io/rpc`, UniSat Wallet API (`window.unisat`), and `@stacks/connect` for Bitcoin/Ordinals/Stacks deposits.
- **Decimal.js / BigNumber.js:** Executes exact multi-leg pricing and fiat conversion without IEEE 754 floating-point errors.
- **Zustand with Immer:** Ephemeral state store managing active strategy legs, custom strikes, slider parameters, and wallet connection states.

## Backend / Infra Touchpoints
- **Market Data Service (Prompt 207):** Connects to `wss://api.growww.in/ws/v1/options` for streaming options quotes, Level-2 depth, implied volatilities, and calculated Greeks.
- **Order Service & Advanced Order Engine (Prompt 204, Prompt 226):** Dispatches multi-leg combo order payloads via `POST /api/v1/orders/multi-leg` with client-generated idempotency keys.
- **Pre-Trade Risk & Margin Engine (Prompt 206, Prompt 229):** Queries real-time SPAN margin requirements and collateral relief for multi-leg strategies via `POST /api/v1/risk/margin-preview`.
- **Foreign Investor Funding & FX Service (Prompt 214):** Requests deposit quotes and cross-chain routing parameters via `POST /api/v1/funding/deposit/quote`.
- **Sanctions Screening & AML Service (Prompt 703):** Validates source wallet addresses against TRM Labs / Chainalysis blocklists via `POST /api/v1/compliance/screen-address`.
- **Audit Log Service (Prompt 218):** Ingests non-repudiable logs of Web3 deposit initiation, signed payloads, and strategy execution intents.

## Blockchain Interaction
- **Target Settlement Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus and deterministic 2-second block finality).
- **Supported Deposit Source Networks:**
  - **Ethereum / Arbitrum / Polygon (EVM):** USDT, USDC, and ETH deposits via institutional vault smart contracts (`GrowwwEvmDepositVault.sol`).
  - **Solana (SVM):** Native SOL and SPL-USDC/USDT transfers to verified deposit custody addresses via Solana Token Program instructions.
  - **Bitcoin / Stacks:** Native BTC transfers with OP_RETURN payload markers and Stacks sBTC/USDC smart contract interactions.
- **On-Chain Settlement Verification:**
  - Deposit transactions across source chains emit cryptographic deposit events ingested by backend custody relayers (Prompt 319).
  - Once required source network confirmations are reached, the custody relayer mints tokenized collateral units on the Growww Hyperledger Besu ledger via `SettlementDvP.sol` or `DigitalSecurityToken.sol`.
  - The web interface monitors Hyperledger Besu WebSocket RPC for `CollateralCredited(address indexed investor, bytes32 indexed sourceTxHash, uint256 amount)` events to provide real-time confirmation.
- **Zero PII Invariant:** Wallet addresses, transaction hashes, and on-chain events contain zero personally identifiable information, adhering strictly to SEBI, DPDP Act 2023, and GDPR mandates.

## Step-by-Step Build Instructions
1. Scaffold options trading and Web3 deposit route architecture under `apps/growww_web/app/(app)/`:
   - `options/[ticker]/page.tsx`: Main options chain terminal and matrix layout.
   - `options/builder/page.tsx`: Standalone multi-leg strategy builder and payoff analyzer.
   - `options/volatility/page.tsx`: 2D Volatility Smile and 3D Volatility Surface explorer.
   - `deposit/web3/page.tsx`: Multi-chain Web3 deposit portal and bridge status tracker.
2. Implement Options Chain Matrix component with three distinct columns: Calls (bid, ask, IV, delta, gamma, theta, OI, volume), central Strike price ladder, and Puts (mirroring Calls data).
3. Integrate `@tanstack/react-virtual` into the options chain matrix to efficiently render dozens of strikes per expiry date without browser DOM degradation.
4. Build Expiry Ribbon selector allowing traders to switch between weekly and monthly expiry cycles, with automated ITM/ATM/OTM background shading based on real-time underlying spot prices.
5. Implement live WebSocket data pipeline subscribing to options tickers, updating matrix cells, bid/ask spreads, and Greeks dynamically with green/red flash indicators on price ticks.
6. Create Strategy Builder component featuring preset dropdowns (Straddle, Strangle, Bull Call Spread, Bear Put Spread, Iron Condor, Iron Butterfly, Calendar Spread, Collar) and dynamic leg table.
7. Build Strategy Leg editor allowing users to toggle Buy/Sell, Call/Put, adjust strike steppers, modify expiration dates, and set individual leg lot ratios (e.g. 1:2 ratio spreads).
8. Build interactive Payoff Visualizer using HTML5 Canvas / D3.js displaying expiration P&L curves, target-date P&L curves, upper/lower break-even thresholds, max potential profit, and max risk values.
9. Connect Pre-Trade Margin API (`/api/v1/risk/margin-preview`) to dynamically calculate and display portfolio margin requirements, margin benefit from spread offsetting, and net debit/credit.
10. Build Atomic Multi-Leg Order Ticket allowing single-click execution with user-defined net price limit, execution time-in-force (IOC, GTC, DAY), and batch idempotency keys.
11. Implement 2D Volatility Smile/Skew chart using Lightweight Charts plotting Implied Volatility across strike prices for selected expiry series.
12. Implement 3D Volatility Surface canvas using Three.js and React Three Fiber:
    - Construct parametric geometry surface mapping Strike (X-axis), Expiry Days (Y-axis), and Implied Volatility (Z-axis).
    - Apply smooth color gradient shaders (cool blue for low IV to warm red/gold for high IV).
    - Add interactive orbit controls, mouse hover tooltip displaying exact (Strike, Expiry, IV) coordinates, and wireframe toggle.
13. Integrate Multi-Chain Web3 Wallet Provider wrapping the deposit portal with Wagmi (EVM), Solana Wallet Adapter, and Leather/UniSat Bitcoin connectors.
14. Build Web3 Deposit Wizard:
    - Chain selector (Ethereum, Arbitrum, Polygon, Solana, Bitcoin, Stacks).
    - Asset selector (USDC, USDT, ETH, SOL, BTC).
    - Dynamic deposit quote calculator displaying exchange rates, bridge fees, network gas estimates, and estimated time to Besu settlement credit.
    - Automated address screening pre-check calling `/api/v1/compliance/screen-address` before unlocking wallet transaction signing.
15. Build Cross-Chain Deposit Tracker monitoring transaction confirmations on the source chain, relayer validation status, and final Hyperledger Besu collateral minting event.
16. Implement comprehensive automated unit and integration tests using Vitest and end-to-end user journey tests with Playwright.

## Interfaces / Contracts

### Options Chain & Greeks Schemas
```typescript
export type OptionType = 'CALL' | 'PUT';
export type OptionMoneyness = 'ITM' | 'ATM' | 'OTM';

export interface OptionGreeks {
  delta: number;
  gamma: number;
  theta: number;
  vega: number;
  rho: number;
  impliedVolatility: number;
}

export interface OptionContractSummary {
  contractId: string;
  ticker: string;
  underlyingIsin: string;
  strikePrice: string; // Fixed-point string e.g. "2500.00"
  expiryDate: string; // ISO date string "2026-09-25"
  optionType: OptionType;
  moneyness: OptionMoneyness;
  bidPrice: string;
  bidSize: number;
  askPrice: string;
  askSize: number;
  lastPrice: string;
  changePercent24h: number;
  volume24h: number;
  openInterest: number;
  openInterestChange24h: number;
  greeks: OptionGreeks;
}

export interface OptionsChainExpiryGroup {
  expiryDate: string;
  daysToExpiry: number;
  underlyingSpotPrice: string;
  strikes: Array<{
    strikePrice: string;
    call: OptionContractSummary | null;
    put: OptionContractSummary | null;
  }>;
}
```

### Multi-Leg Strategy Builder Schemas
```typescript
export type StrategyPreset =
  | 'CUSTOM'
  | 'LONG_CALL'
  | 'LONG_PUT'
  | 'BULL_CALL_SPREAD'
  | 'BEAR_PUT_SPREAD'
  | 'LONG_STRADDLE'
  | 'SHORT_STRADDLE'
  | 'LONG_STRANGLE'
  | 'SHORT_STRANGLE'
  | 'IRON_CONDOR'
  | 'IRON_BUTTERFLY'
  | 'CALENDAR_SPREAD'
  | 'COVERED_CALL'
  | 'COLLAR';

export interface StrategyLeg {
  legId: string;
  contractId: string;
  action: 'BUY' | 'SELL';
  optionType: OptionType;
  strikePrice: string;
  expiryDate: string;
  ratio: number; // e.g. 1, 2
  limitPrice: string;
  impliedVolatility: number;
  greeks: OptionGreeks;
}

export interface OptionsStrategyDefinition {
  preset: StrategyPreset;
  underlyingTicker: string;
  underlyingSpotPrice: string;
  legs: StrategyLeg[];
  netDebitOrCredit: 'NET_DEBIT' | 'NET_CREDIT';
  netPrice: string;
  maxProfit: string | 'UNLIMITED';
  maxLoss: string | 'UNLIMITED';
  breakEvenPoints: string[];
  aggregateGreeks: OptionGreeks;
  requiredMarginInr: string;
  marginBenefitInr: string;
}

export interface StrategyPayoffPoint {
  underlyingPrice: number;
  pnlAtExpiry: number;
  pnlAtTargetDate: number;
}

export interface MultiLegOrderRequest {
  idempotencyKey: string;
  underlyingTicker: string;
  strategyPreset: StrategyPreset;
  netPriceLimit: string;
  timeInForce: 'IOC' | 'GTC' | 'DAY';
  legs: Array<{
    contractId: string;
    action: 'BUY' | 'SELL';
    quantity: number;
    priceLimit?: string;
  }>;
}
```

### 2D/3D Volatility Surface Schemas
```typescript
export interface VolatilitySmilePoint {
  strikePrice: number;
  impliedVolatility: number;
  delta: number;
  openInterest: number;
}

export interface VolatilitySurfacePoint {
  strikePrice: number;
  daysToExpiry: number;
  impliedVolatility: number;
  moneyness: number; // Spot / Strike
}

export interface VolatilitySurfaceDataset {
  underlyingTicker: string;
  spotPrice: number;
  timestamp: string;
  points: VolatilitySurfacePoint[];
  strikes: number[];
  expiries: number[]; // Days to expiry array
  ivMatrix: number[][]; // 2D grid of IV values [expiryIndex][strikeIndex]
}
```

### Multi-Chain Web3 Deposit Schemas
```typescript
export type SupportedWeb3Network =
  | 'ETHEREUM_MAINNET'
  | 'ARBITRUM_ONE'
  | 'POLYGON_POS'
  | 'SOLANA_MAINNET'
  | 'BITCOIN_MAINNET'
  | 'STACKS_MAINNET'
  | 'BESU_CONSORTIUM';

export type Web3WalletProviderType =
  | 'METAMASK'
  | 'RABBY'
  | 'PHANTOM'
  | 'SOLFLARE'
  | 'LEATHER'
  | 'UNISAT'
  | 'WALLET_CONNECT';

export interface CrossChainDepositQuoteRequest {
  sourceNetwork: SupportedWeb3Network;
  depositAsset: 'USDC' | 'USDT' | 'ETH' | 'SOL' | 'BTC' | 'SBTC';
  depositAmount: string;
  investorBesuAddress: `0x${string}`;
}

export interface CrossChainDepositQuoteResponse {
  quoteId: string;
  sourceNetwork: SupportedWeb3Network;
  depositAsset: string;
  depositAmount: string;
  estimatedTargetInrValue: string;
  exchangeRateUsdInr: string;
  relayerFee: string;
  estimatedGasCostUsd: string;
  requiredConfirmations: number;
  estimatedSettlementSeconds: number;
  depositVaultAddress: string;
  quoteExpiresAt: string;
}

export interface Web3DepositTransaction {
  depositId: string;
  sourceNetwork: SupportedWeb3Network;
  walletProvider: Web3WalletProviderType;
  sourceAddress: string;
  sourceTxHash: string;
  asset: string;
  amount: string;
  status:
    | 'WALLET_SIGNING'
    | 'SOURCE_PENDING'
    | 'SOURCE_CONFIRMED'
    | 'AML_SCREENING_PASSED'
    | 'RELAYER_PROCESSING'
    | 'BESU_CREDITED'
    | 'FAILED';
  sourceConfirmations: number;
  requiredConfirmations: number;
  besuCreditTxHash?: `0x${string}`;
  createdAt: string;
  completedAt?: string;
  failureReason?: string;
}
```

## Security & Compliance Notes
- **Atomic Multi-Leg Order Safeguards:** To eliminate leg risk (the risk of one leg executing while other legs fail), the web ticket mandates atomic all-or-none execution semantics verified by the Advanced Order Types Engine.
- **Pre-Deposit AML & Sanctions Screening:** All connected source wallet addresses (EVM, Solana, Bitcoin) are screened against TRM Labs and Chainalysis sanctions databases via the API Gateway before deposit instructions or vault addresses are displayed.
- **Client-Side Signature Safety:** No private keys or seed phrases are ever handled or requested by the web application. All Web3 transactions utilize standard wallet signing protocols (EIP-712 typed data for EVM, Versioned Transactions for Solana, and PSBT / Leather RPC for Bitcoin).
- **SEBI & IFSCA Risk Warnings:** Derivatives trading involves substantial risk of capital loss. In compliance with SEBI and IFSCA regulations, high-prominence statutory risk disclosures and margin call warnings are rendered before entering derivative positions.
- **Idempotency & Replay Protection:** Every multi-leg order submission and Web3 deposit registration generates a unique client UUIDv4 idempotency key to prevent accidental duplicate execution on network retry or double clicks.

## Acceptance Criteria
- [ ] Options Chain Matrix renders calls, puts, and strikes with smooth scrolling and zero lag under high-frequency WebSocket updates.
- [ ] ITM, ATM, and OTM options contracts are clearly distinguished with compliant background shading and live Black-Scholes Greeks.
- [ ] Strategy Builder allows configuring all standard presets (Straddle, Strangle, Iron Condor, Bull/Bear Spreads) and custom legs.
- [ ] Interactive Payoff Visualizer displays accurate P&L curves at expiration and target dates with explicit break-even points.
- [ ] 2D Volatility Smile and 3D Volatility Surface canvases render smoothly with zoom, pan, and rotation controls.
- [ ] Web3 Deposit Portal connects cleanly to MetaMask, Phantom, Leather, and UniSat wallets without wallet library conflicts.
- [ ] Cross-chain deposit quote calculator accurately computes exchange rates, relayer fees, and required block confirmations.
- [ ] Pre-deposit AML checks block sanctioned wallet addresses before generating on-chain deposit transactions.
- [ ] Cross-Chain Deposit Tracker visualizes step-by-step confirmation progress through to Hyperledger Besu collateral credit.
- [ ] Playwright end-to-end tests verify options chain navigation, strategy builder execution, and wallet connection flows.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 204 (Order Service), Prompt 206 (Pre-Trade Risk Engine), Prompt 207 (Market Data Service), Prompt 214 (Foreign Investor Funding), Prompt 601 (Next.js Web Scaffolding).
- **Parallel Tasks:** Prompt 226 (Advanced Order Types Engine), Prompt 319 (Institutional Custody Bridge), Prompt 603 (Web Trading Dashboard).
- **Downstream Blockers:** Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios), Prompt 907 (Regulatory Sandbox Pilot Launch).
