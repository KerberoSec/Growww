# 539 - Flutter Fixed-Income Yield & Staking Vault Screen

## Purpose
Provides an intuitive, institutional-grade fixed-income investment and staking interface allowing retail and institutional investors to stake tokenized sovereign Government Securities (gGSEC), Treasury Bills (gTBILL), State Development Loans (gSDL), and AAA-rated corporate debt (gCORP) for automated daily yield compounding under ADR-0040.

Traditional Indian fixed-income markets suffer from high retail entry barriers (standard lot sizes exceeding INR 5,00,000), opaque dirty pricing mechanics, fragmented secondary market liquidity, and irregular semi-annual coupon distribution cycles. This interface democratizes debt investing by fractionalizing sovereign and corporate debt down to INR 100 increments, standardizing complex bond metrics (Yield to Maturity, modified duration, coupon day counts) into accessible Annual Percentage Yield (APY) figures, and delivering continuous daily yield accumulation with sub-paise micro-accrual resolution.

Investors can choose between flexible savings vaults offering anytime unbonding with T+0 liquidity, and locked staking tenors (30-day, 90-day, 180-day, 365-day fixed terms) offering boosted APY rates. Built natively on Flutter for cross-platform deployment (Android, iOS, macOS, Windows, Linux, Web), the application leverages gasless ERC-4337 account abstraction over Hyperledger Besu (ADR-0026) to provide a seamless Web2-like experience where users sign staking, unbonding, and reward-sweeping actions with zero gas overhead and one-tap biometric confirmation.

## What You Are Building
A high-performance, reactive fixed-income yield and staking suite located in `apps/growww_flutter/lib/screens/earn/` (with supporting domain and presentation packages under `apps/growww_flutter/lib/features/earn/`):
- `EarnDashboardScreen`: The primary discovery and portfolio overview screen displaying aggregate active principal, total all-time accrued interest in INR, 24-hour daily yield rate, live APY leaderboard across sovereign and corporate debt pools, and asset classification filter tabs (All, Sovereign G-Secs, Treasury Bills, Corporate Bonds, High Yield).
- `VaultDetailScreen`: Detailed inspection screen for individual fixed-income vaults presenting sovereign issuer credentials, ISIN, credit ratings (SOV / CRISIL AAA), underlying bond maturity dates, current gross and net APY, Total Value Locked (TVL) pool capacity, lockup term options, and on-chain Proof-of-Reserve verification.
- `ApyCalculatorBottomSheet`: Interactive financial calculator allowing users to slide or enter custom principal amounts, select tenors (Flexible, 30D, 90D, 180D, 365D), toggle compounding frequencies (Daily vs At Maturity), preview projected daily/monthly/annual returns, and inspect statutory Section 194A TDS withholding estimates.
- `TermSelectorWidget`: Modular segmented selector component enabling seamless switching between Flexible and Locked terms, rendering dynamic bonus APY badges (+0.25% to +1.85%), computed maturity dates, and early unbonding penalty notices.
- `DailyAccruedInterestTicker`: High-frequency 60 FPS micro-ticker widget that animates continuous fractional interest accumulation in real time down to four decimal places (sub-paise resolution: INR 0.0001 precision), giving users immediate visual feedback of compounding returns.
- `OneTapClaimWidget`: Sticky action widget allowing users to harvest accrued unallocated yield with a single tap, offering a dual-action modal to either sweep earnings to their Demat cash ledger or auto-compound back into the vault, executed gaslessly via ERC-4337 Paymaster sponsorship.
- `StakeConfirmationModal`: High-security confirmation sheet displaying principal deduction, expected daily yield, lockup term, maturity timestamp, TDS advisory, and hardware biometric challenge (`local_auth`) before submitting the on-chain stake transaction.
- `UnstakeRedemptionModal`: Redemption dialog supporting instant unbonding for flexible deposits, scheduled maturity withdrawal for locked tenors, or emergency early exit with transparent 0.25% penalty deduction calculations.
- `YieldTransactionHistoryScreen`: Paginated ledger history rendering all deposit, claim, reinvestment, and withdrawal events with transaction statuses, Besu block explorer links, and downloadable TDS credit notes.
- `Riverpod State Layer`: Asynchronous notifiers (`EarnDashboardNotifier`, `VaultDetailNotifier`, `ApyCalculatorNotifier`, `StakingExecutionNotifier`, `InterestTickerNotifier`) managing reactive state transitions, local caching, and real-time WebSocket yield streams.

## Scope Boundaries
- **In Scope:**
  - Multi-platform Flutter UI screens, dialogs, bottom sheets, and interactive calculators for fixed-income yield staking.
  - Interactive APY calculator supporting flexible and locked tenors with compound interest projections.
  - Term selection matrix (Flexible, 30D, 90D, 180D, 365D) with bonus yield calculation and penalty disclosures.
  - Real-time sub-second micro-accrual interest ticker widget with sub-paise integer arithmetic (`decimal: ^2.3.3`).
  - One-tap yield claiming and auto-reinvestment workflows with gasless ERC-4337 transaction construction.
  - Biometric challenge integration (`local_auth`) for staking and unstaking authorizations.
  - Regulatory disclosure banners for non-guaranteed yield and Section 194A TDS withholding warnings.
  - Form 15G / Form 15H submission status indicator for senior citizens and nil-tax declarations.
  - Error state handling, offline caching, and optimistic UI updates for claim actions.
- **Out of Scope / Handled Elsewhere:**
  - On-chain smart contract deployment and staking pool logic in `BondYieldVault.sol` (Prompt 333, Prompt 329).
  - Bond yield curve fitting, Nelson-Siegel-Svensson calculations, and dirty price valuation engines (Prompt 254).
  - Primary fiat payment gateway processing and UPI/NEFT bank mandate funding (Prompt 212, Prompt 511).
  - Portfolio aggregate balance aggregation and multi-asset net worth reconciliation (Prompt 209).
  - Core double-entry ledger bookkeeping and cash wallet debit/credit mutations (Prompt 203).
  - Custodian depository physical bond settlement (Prompt 213).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (v2.5+) with code generation (`@riverpod`) and immutable state classes (`freezed: ^2.5.2`).
- **Fixed-Point Financial Mathematics:** `decimal: ^2.3.3` and `rational: ^2.2.3` for sub-paise integer calculations (1 INR = 10,000 sub-paise, 0.0001 INR precision) ensuring zero floating-point drift during compound interest micro-ticks.
- **Currency & Number Formatting:** Custom Indian Rupee number formatters supporting the Indian numbering system (Lakhs and Crores, e.g., `INR 12,34,567.8900`) and sub-paise precision display.
- **Hardware Biometrics:** `local_auth: ^2.2.0` for biometric authorization (Fingerprint, FaceID, Touch ID, Windows Hello) on high-value staking and unbonding intents.
- **Micro-Animation & Ticker Engine:** Custom `AnimationController` and `Ticker` driving smooth 60 FPS sub-paise numeric interpolation without rebuilding parent widget trees.
- **Web3 & ERC-4337 Client:** `web3dart: ^2.7.3` combined with custom EIP-4337 UserOperation bundler and Paymaster JSON-RPC clients for gasless meta-transactions (ADR-0026).
- **Networking & Real-Time Streams:** `dio: ^5.4.3+1` for REST API endpoints and `web_socket_channel: ^3.0.0` for live yield accrual and TVL update streams.

## Backend / Infra Touchpoints
- **Bond Yield Curve & Dirty Price Calculator Service (Prompt 254 - Rust / gRPC / SIMD):**
  - Delivers real-time Yield-to-Maturity (YTM), clean/dirty price calculations, accrued coupon days (Actual/Actual ICMA), and daily APY benchmarks for tokenized sovereign debt (gGSEC, gTBILL).
  - Endpoints: `GET /api/v1/yield/calculator/estimate` and gRPC `BondYieldCalculator.ComputeAccruedYield`.
- **Portfolio Service (Prompt 209 - Go / Temporal):**
  - Ingests aggregated user portfolio yield holdings, active stake records, historical earnings summary, and performance metrics across fixed-income vaults.
  - Endpoints: `GET /api/v1/portfolio/yield/holdings`, `GET /api/v1/portfolio/yield/summary`.
- **Wallet & Ledger Service (Prompt 203 - Rust / PostgreSQL Ledger):**
  - Validates available free cash balance and tokenized bond balances (gGSEC, gTBILL, gCORP) before staking; executes idempotent ledger debits/credits for claimed yield payouts and redemption principals.
  - Endpoints: `GET /api/v1/wallet/balances`, `POST /api/v1/wallet/ledger/reserve-staking`.
- **Earn Vault Directory & Metadata API (Prompt 207 / 254):**
  - Retrieves active yield vault products, current APYs, minimum stake thresholds, lockup terms, TVL capacity, and risk tiers.
  - Endpoint: `GET /api/v1/earn/vaults`.
- **Real-Time Yield WebSocket (`wss://ws.growww.in/v1/earn/accrual-stream/{user_id}`):**
  - Delivers real-time delta packets containing second-by-second accrued interest amounts, updated vault APYs, and transaction confirmations.
- **Tax Withholding & TDS Service (Prompt 210 / ADR-0008):**
  - Computes applicable Section 194A TDS (10% standard withholding, 20% for non-PAN/non-compliant, or 0% for valid Form 15G/15H on file) and generates provisional quarterly TDS certificates.
  - Endpoint: `POST /api/v1/tax/tds/simulate-withholding`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **`BondYieldVault.sol` Smart Contract Interface:**
  - Deployed on Hyperledger Besu QBFT network with ERC-3643 permissioned token backing.
  - Key Contract Methods:
    - `stake(address assetToken, uint256 amount, uint8 tenorEnum)`: Locks tokenized bond units into the designated vault pool.
    - `unstake(uint256 stakeId)`: Releases matured principal and unbonding balances back to the user wallet.
    - `emergencyUnstake(uint256 stakeId)`: Executes immediate unbonding prior to maturity with smart-contract-enforced 0.25% penalty burn.
    - `claimYield(uint256 stakeId, bool autoCompound)`: Harvests accrued yield, either minting new vault shares (auto-compound) or disbursing cash tokens.
    - `getAccruedYield(uint256 stakeId) returns (uint256)`: Computes on-chain real-time yield accrual based on block timestamps and cumulative vault index.
- **Gasless Account Abstraction via ERC-4337:**
  - End users sign EIP-712 typed `UserOperation` payloads containing stake, claim, or unbond calldata.
  - The client forwards `UserOperation` to the platform Paymaster service (ADR-0026), which validates compliance identity via `IdentityRegistry.sol` and sponsors gas execution on Besu.
- **Zero On-Chain PII Guarantee:**
  - The smart contract records only pseudonymous `account_id` hashes, `vault_id`, token amounts, lockup timestamps, and accrued yield integer values. No investor names, PAN numbers, or bank account details are stored on the ledger.
- **1:1 Custody Backing & Proof-of-Reserve:**
  - Every tokenized bond unit staked in `BondYieldVault.sol` corresponds 1:1 to physical sovereign debt held in custody at Clearing Corporation of India Limited (CCIL) / Reserve Bank of India (RBI) SGL accounts or NSDL/CDSL depository accounts.
  - The screen displays an on-chain Proof-of-Reserve badge linked to the daily Sparse Merkle Tree root.

## State Management Architecture
- **EarnDashboardNotifier (`AutoDisposeAsyncNotifier<EarnDashboardState>`):**
  - Manages multi-vault discovery directory, category filtering, aggregate portfolio yield summary, and real-time TVL updates.
- **VaultDetailNotifier (`AutoDisposeFamilyAsyncNotifier<VaultDetailState, String>`):**
  - Manages single vault telemetry, underlying bond characteristics, lockup tenor matrices, historical APY charts, and available user balances.
- **ApyCalculatorNotifier (`AutoDisposeFamilyNotifier<ApyCalculatorState, ApyCalculatorArgs>`):**
  - Maintains reactive calculation state as users adjust principal sliders, select tenors, or toggle daily vs maturity compounding frequencies, computing sub-paise projected returns.
- **InterestTickerNotifier (`AutoDisposeFamilyNotifier<InterestTickerState, String>`):**
  - Manages high-frequency micro-ticker state, accumulating accrued interest at 60 FPS using linear interpolation between last known server checkpoint and projected target yield.
- **StakingExecutionNotifier (`AutoDisposeAsyncNotifier<StakingExecutionState>`):**
  - Orchestrates stake, claim, and unbond workflows: validates available balance, challenges user biometrics, constructs EIP-712 payload, dispatches UserOperation to Paymaster, and applies optimistic UI balance adjustments.
- **TdsStatusNotifier (`AutoDisposeAsyncNotifier<TdsStatusState>`):**
  - Tracks user PAN verification status, cumulative annual interest accrual against the INR 40,000 threshold (INR 50,000 for senior citizens), and Form 15G / Form 15H filing status.

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold feature directory structure under `apps/growww_flutter/lib/screens/earn/` and `apps/growww_flutter/lib/features/earn/`: `presentation/screens/`, `presentation/widgets/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `data/repositories/`, `data/datasources/`.
2. Define domain enums and immutable data models in `domain/models/`: `YieldAssetType`, `TenorType`, `StakeStatus`, `RiskTier`, `CompoundingFrequency`, `YieldVaultProduct`, `UserStakeRecord`, `AccruedInterestSchedule`, `SubPaiseAmount`, `ApyCalculationResult`, and `TdsWithholdingEstimate`.
3. Implement `SubPaiseAmount` value object using `decimal: ^2.3.3` to represent financial amounts with four decimal places of precision (1 INR = 10,000 sub-paise) and provide custom Indian Rupee formatters (`formatToInrPaise` and `formatToInrSubPaise`).
4. Implement `IEarnVaultRepository` and `IStakingExecutionRepository` interfaces defining contract methods for fetching vault products, stake records, APY calculations, gasless UserOperation submission, and real-time yield streaming.
5. Implement `EarnDashboardScreen`:
   - Aggregate portfolio yield banner displaying Total Staked Principal, Total Accrued Yield, and 24-Hour Projected Yield.
   - Category filter tabs (All, Sovereign G-Secs, T-Bills, Corporate Bonds, High Yield).
   - Vault product cards displaying asset icon, issuer name, credit rating badge, live APY badge, flexible vs locked indicators, and TVL progress bar.
   - Live APY leaderboard highlighting top-yielding sovereign instruments.
6. Build `VaultDetailScreen`:
   - Vault header with asset ticker (e.g., `gGSEC-718GS2033`), sovereign badge, and verified Proof-of-Reserve shield.
   - Key bond parameters grid: ISIN, Issuer (Government of India / RBI), Coupon Rate, Maturity Date, Modified Duration, and Settlement Frequency.
   - Tenor selection chips (Flexible, 30D, 90D, 180D, 365D) showing APY differentials.
   - Dynamic APY performance sparkline chart across 30D, 90D, and 1Y time horizons.
   - Action dock with primary "Stake Now" button and secondary "Calculate Returns" button.
7. Build `ApyCalculatorBottomSheet`:
   - Principal input field with quick-preset chips (INR 5,000, INR 25,000, INR 1,00,000, INR 5,00,000) and smooth slider control.
   - Tenor selector tabs with automatic APY bonus recalculation.
   - Compounding frequency switch (Daily Auto-Compound vs At Maturity Payout).
   - Real-time return summary table: Daily Earnings, Monthly Earnings, Total Maturity Earnings, Net APY, and Estimated Section 194A TDS deduction.
8. Build `TermSelectorWidget`:
   - Segmented control rendering Flexible (T+0 liquidity) versus Fixed Tenors (30D, 90D, 180D, 365D).
   - Dynamic bonus APY indicators (e.g., `+0.75% APY Boost`).
   - Calculated maturity date banner and early unbonding penalty disclosure (0.25% penalty on principal for emergency exits).
9. Build `DailyAccruedInterestTicker`:
   - High-performance `CustomPainter` or optimized `AnimatedBuilder` driving sub-second micro-ticks at 60 FPS.
   - Smooth numeric interpolation between previous checkpoint balance and calculated instantaneous yield.
   - Display formatting in high-precision Indian Rupee notation (e.g., `INR 1,452.8942`).
   - Visual pulse glow animation triggered whenever a micro-accrual milestone is reached.
10. Build `OneTapClaimWidget`:
    - Floating or embedded action widget showing unallocated accrued yield balance.
    - One-tap "Claim Yield" button opening a bottom sheet with dual actions: "Transfer to Demat Cash" or "Reinvest in Vault".
    - Gasless execution indicator showing zero gas fee badge sponsored by Growww Paymaster.
    - Optimistic balance update with instant toast confirmation while on-chain transaction confirms in background.
11. Build `StakeConfirmationModal` and `UnstakeRedemptionModal`:
    - Detailed confirmation card showing deposit principal, token deductions, projected returns, lockup expiry date, and TDS disclosure.
    - Biometric verification trigger (`local_auth`) verifying device fingerprint or FaceID.
    - Unstake modal displaying standard maturity redemption options versus emergency early exit with explicit 0.25% penalty deduction.
12. Build `YieldTransactionHistoryScreen`:
    - Filterable history list (All, Deposits, Claims, Reinvestments, Redemptions).
    - Transaction card showing date, transaction type, amount in sub-paise precision, status badge (Confirmed, Pending, Failed), and Besu transaction hash with copy/share actions.
    - TDS credit certificate download button linking to generated Form 16A provisional summaries.
13. Implement comprehensive unit tests verifying sub-paise precision arithmetic, APY compounding formulas, Section 194A TDS thresholds, and tenor bonus calculations.
14. Implement widget and integration tests for `EarnDashboardScreen`, `VaultDetailScreen`, `ApyCalculatorBottomSheet`, `DailyAccruedInterestTicker`, and `OneTapClaimWidget`.

## Interfaces / Contracts

```dart
// lib/features/earn/domain/models/earn_enums.dart

enum YieldAssetType {
  gGsec,
  gTbill,
  gSdl,
  gCorp,
}

enum TenorType {
  flexible,
  locked30Days,
  locked90Days,
  locked180Days,
  locked365Days,
}

enum StakeStatus {
  active,
  pendingClaim,
  unbonding,
  matured,
  completed,
  cancelled,
}

enum RiskTier {
  sovereignRiskFree,
  institutionalLowRisk,
  corporateModerateRisk,
}

enum CompoundingFrequency {
  daily,
  atMaturity,
}

enum ClaimDestination {
  dematCashLedger,
  autoCompoundReinvest,
}

// lib/features/earn/domain/models/sub_paise_amount.dart

class SubPaiseAmount {
  /// Internal integer representation where 1 INR = 10,000 sub-paise
  final int subPaiseValue;

  const SubPaiseAmount(this.subPaiseValue);

  factory SubPaiseAmount.fromInr(double inr) {
    return SubPaiseAmount((inr * 10000).round());
  }

  double toInr() => subPaiseValue / 10000.0;

  String formatInr({bool showSubPaise = false}) {
    final inrValue = toInr();
    if (showSubPaise) {
      return 'INR ${inrValue.toStringAsFixed(4)}';
    }
    return 'INR ${inrValue.toStringAsFixed(2)}';
  }
}

// lib/features/earn/domain/models/yield_vault_product.dart

class YieldVaultProduct {
  final String vaultId;
  final String vaultAddress; // Smart contract address on Besu
  final String assetSymbol; // e.g. "gGSEC-718GS2033"
  final String displayName; // e.g. "7.18% GS 2033 Sovereign Vault"
  final YieldAssetType assetType;
  final String isin;
  final String issuerName; // e.g. "Government of India / RBI"
  final String creditRating; // e.g. "SOV", "CRISIL AAA"
  final double baseApy; // e.g. 7.18 for 7.18%
  final Map<TenorType, double> tenorBonusApy;
  final SubPaiseAmount minStakeAmount;
  final SubPaiseAmount totalValueLocked;
  final SubPaiseAmount maxPoolCapacity;
  final RiskTier riskTier;
  final DateTime bondMaturityDate;
  final double modifiedDurationYears;
  final bool isAcceptingStakes;
  final String proofOfReserveRootHash;

  const YieldVaultProduct({
    required this.vaultId,
    required this.vaultAddress,
    required this.assetSymbol,
    required this.displayName,
    required this.assetType,
    required this.isin,
    required this.issuerName,
    required this.creditRating,
    required this.baseApy,
    required this.tenorBonusApy,
    required this.minStakeAmount,
    required this.totalValueLocked,
    required this.maxPoolCapacity,
    required this.riskTier,
    required this.bondMaturityDate,
    required this.modifiedDurationYears,
    required this.isAcceptingStakes,
    required this.proofOfReserveRootHash,
  });
}

// lib/features/earn/domain/models/user_stake_record.dart

class UserStakeRecord {
  final String stakeId;
  final String vaultId;
  final String assetSymbol;
  final SubPaiseAmount principalAmount;
  final TenorType tenorType;
  final double agreedApy;
  final CompoundingFrequency compoundingFrequency;
  final StakeStatus status;
  final DateTime stakedAt;
  final DateTime? lockedUntil;
  final SubPaiseAmount totalAccruedYield;
  final SubPaiseAmount unallocatedYield;
  final SubPaiseAmount alreadyClaimedYield;
  final DateTime lastYieldCheckpoint;
  final String onChainTxHash;

  const UserStakeRecord({
    required this.stakeId,
    required this.vaultId,
    required this.assetSymbol,
    required this.principalAmount,
    required this.tenorType,
    required this.agreedApy,
    required this.compoundingFrequency,
    required this.status,
    required this.stakedAt,
    this.lockedUntil,
    required this.totalAccruedYield,
    required this.unallocatedYield,
    required this.alreadyClaimedYield,
    required this.lastYieldCheckpoint,
    required this.onChainTxHash,
  });
}

// lib/features/earn/domain/models/apy_calculation_result.dart

class ApyCalculationResult {
  final SubPaiseAmount principal;
  final TenorType tenor;
  final double effectiveApy;
  final SubPaiseAmount projectedDailyYield;
  final SubPaiseAmount projectedMonthlyYield;
  final SubPaiseAmount projectedTotalYield;
  final SubPaiseAmount estimatedTdsDeduction;
  final SubPaiseAmount netMaturityAmount;
  final DateTime projectedMaturityDate;

  const ApyCalculationResult({
    required this.principal,
    required this.tenor,
    required this.effectiveApy,
    required this.projectedDailyYield,
    required this.projectedMonthlyYield,
    required this.projectedTotalYield,
    required this.estimatedTdsDeduction,
    required this.netMaturityAmount,
    required this.projectedMaturityDate,
  });
}

// lib/features/earn/domain/models/tds_withholding_estimate.dart

class TdsWithholdingEstimate {
  final double tdsRatePercentage; // 10% standard, 20% non-PAN, 0% Form 15G/15H
  final SubPaiseAmount grossInterestEarned;
  final SubPaiseAmount tdsAmountWithheld;
  final SubPaiseAmount netInterestPayable;
  final bool isForm15G15HApplied;
  final String panMasked;

  const TdsWithholdingEstimate({
    required this.tdsRatePercentage,
    required this.grossInterestEarned,
    required this.tdsAmountWithheld,
    required this.netInterestPayable,
    required this.isForm15G15HApplied,
    required this.panMasked,
  });
}

// lib/features/earn/domain/interfaces/i_earn_vault_repository.dart

abstract class IEarnVaultRepository {
  Future<List<YieldVaultProduct>> fetchYieldVaults({
    YieldAssetType? filterType,
  });

  Future<YieldVaultProduct> fetchVaultDetails({
    required String vaultId,
  });

  Future<List<UserStakeRecord>> fetchUserActiveStakes();

  Future<ApyCalculationResult> calculateApyReturns({
    required String vaultId,
    required SubPaiseAmount principal,
    required TenorType tenor,
    required CompoundingFrequency compounding,
  });

  Stream<SubPaiseAmount> streamUserLiveAccrual({
    required String userId,
  });

  Future<TdsWithholdingEstimate> estimateTdsWithholding({
    required SubPaiseAmount interestAmount,
  });
}

// lib/features/earn/domain/interfaces/i_staking_execution_repository.dart

abstract class IStakingExecutionRepository {
  Future<UserStakeRecord> initiateStake({
    required String vaultId,
    required SubPaiseAmount amount,
    required TenorType tenor,
    required CompoundingFrequency compounding,
    required String biometricAuthToken,
    required String idempotencyKey,
  });

  Future<String> claimAccruedYield({
    required String stakeId,
    required ClaimDestination destination,
  });

  Future<String> initiateUnstake({
    required String stakeId,
    required bool isEmergencyExit,
    required String biometricAuthToken,
  });
}
```

## Security & Compliance Notes
- **SEBI Non-Guaranteed Return Disclosure:** Explicit statutory warning banner stating that historical APYs reflect underlying bond coupon rates and market YTM, not guaranteed bank fixed deposit rates. Fixed-income debt instruments carry sovereign interest rate risk, duration volatility, and secondary market liquidity variations.
- **Section 194A TDS Compliance:** Clear notification and inline calculation of 10% TDS withholding on cumulative annual interest exceeding INR 40,000 (INR 50,000 for senior citizens). Displays Form 15G / Form 15H filing status to indicate zero-TDS applicability for qualifying investors.
- **Sub-Paise Precision & Deterministic Rounding:** All financial computations use integer sub-paise precision (1 INR = 10,000 sub-paise) backed by the `decimal` package to eliminate IEEE 754 floating-point rounding errors during continuous micro-accrual compounding.
- **Biometric Dual-Factor Authorization:** High-value stake operations (exceeding INR 50,000) and all unbonding requests mandate hardware-backed biometric verification (`local_auth`) to prevent unauthorized capital movement on compromised devices.
- **Anti-Front-Running Lockup Cooldown:** Unstaking requests enforce a 24-hour unbonding cooldown window or a smart-contract-enforced 0.25% emergency exit fee to prevent yield arbitrage around coupon record dates.
- **Gasless Meta-Transactions via ERC-4337:** Staking, claiming, and unbonding actions use EIP-712 typed data signatures sponsored by the Growww Paymaster (ADR-0026). The smart contracts authenticate the caller via `IdentityRegistry.sol` to guarantee KYC-verified participation.
- **Zero On-Chain PII Policy:** All blockchain transactions via ERC-4337 store pseudonymous hashes only. No investor names, PAN numbers, bank accounts, or residential addresses are written to the Besu ledger.

## Acceptance Criteria
- [ ] `EarnDashboardScreen` renders aggregate staked principal, all-time accrued interest, and 24-hour daily yield rate accurately.
- [ ] Vault directory displays active tokenized debt pools (gGSEC, gTBILL, gCORP, gSDL) with accurate APY, credit ratings, and TVL bars.
- [ ] `ApyCalculatorBottomSheet` allows smooth principal sliding and computes daily, monthly, and maturity yields in real time.
- [ ] `TermSelectorWidget` switches between Flexible and Locked terms, dynamically applying bonus APYs and computing exact maturity dates.
- [ ] `DailyAccruedInterestTicker` runs at 60 FPS with smooth micro-accrual interpolation down to four decimal places (INR 0.0001 precision).
- [ ] `OneTapClaimWidget` allows single-tap yield harvesting with options to sweep to Demat cash or auto-reinvest into the vault.
- [ ] All blockchain interactions with `BondYieldVault.sol` execute gaslessly via ERC-4337 Paymaster sponsorship.
- [ ] High-value stakes and unstaking actions enforce hardware-backed biometric verification via `local_auth`.
- [ ] Statutory non-guaranteed yield warning banner and Section 194A TDS deduction warnings render prominently.
- [ ] Sub-paise integer arithmetic ensures zero floating-point calculation discrepancies across all screens and calculators.
- [ ] Specification contains zero application implementation code (only domain contracts, interfaces, and architecture).
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Scaffolding), Prompt 502 (Riverpod Architecture), Prompt 503 (Design System & Theming), Prompt 504 (Onboarding & KYC Flow), Prompt 505 (Authentication UI), Prompt 521 (Secure Storage), Prompt 525 (API Client Layer).
- **Backend Dependencies:** Prompt 203 (Wallet & Ledger Service), Prompt 209 (Portfolio Service), Prompt 210 (Fee & PnL Engine), Prompt 254 (Bond Yield Curve & Dirty Price Calculator Service).
- **Blockchain Dependencies:** Prompt 304 (Token Redemption Service), Prompt 329 (NBSE Settlement DvP & Fee Collector), Prompt 333 (Tokenized G-Sec Bonds & Coupon Accrual Contract).
- **Downstream Blockers:** Prompt 510 (Portfolio Holdings Screen), Prompt 512 (Transaction History Screen), Prompt 902 (End-to-End Integration Testing).
