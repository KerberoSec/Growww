# 333 - Tokenized G-Sec Bonds & Coupon Accrual Smart Contracts (Solidity / Besu)

## Purpose
The Indian sovereign debt market, comprising Central Government Securities (G-Secs), Treasury Bills (T-Bills), State Development Loans (SDLs), and Sovereign Green Bonds (SGrBs), represents the largest and most liquid fixed-income asset class in India. Historically managed through Reserve Bank of India (RBI) e-Kuber core banking infrastructure, the Negotiated Dealing System - Order Matching (NDS-OM) platform, and Clearing Corporation of India Limited (CCIL) settlement mechanisms, retail and international participation has remained constrained by institutional ticket sizes, fragmented market infrastructure, and complex bond settlement conventions.

This prompt specifies the smart contract architecture for tokenizing Indian sovereign debt securities on Hyperledger Besu enterprise blockchain under `contracts/debt/`:
1. `TokenizedGSec.sol` (`ITokenizedGSec.sol`): Implements fractionalized, ERC-1400 / ERC-3643 compliant sovereign debt security tokens representing G-Secs, T-Bills, and SDLs with 1:1 underlying custody held in RBI Constituent Subsidiary General Ledger (CSGL) accounts. Enforces Day-Count conventions (Actual/Actual ICMA for coupon-bearing bonds, Actual/365 for money market paper, and 30/360 Indian conventions), Clean vs Dirty price settlement calculations, and maturity redemption with RBI cash escrow payout.
2. `CouponDistributor.sol` (`ICouponDistributor.sol`): Orchestrates automated coupon snapshotting on canonical record dates, accrued interest distribution in e-Rupee CBDC / INR fiat escrow, non-resident tax withholding (TDS) accounting, and uncollected coupon claim registries.
3. `TokenizedGSecSTRIPS.sol` (`ITokenizedGSecSTRIPS.sol`): Implements Separate Trading of Registered Interest and Principal of Securities (STRIPS), enabling institutional and retail investors to split coupon-bearing G-Secs into zero-coupon Principal STRIPS (P-STRIPS) and individual semi-annual Coupon STRIPS (C-STRIPS), or reconstitute them back into fungible parent G-Sec tokens.
4. Universal Flat Fee Integration: Enforces the mandatory 0.00% (Zero Fee) platform transaction fee on all secondary bond turnover, distributed automatically as Platform Treasury, Core Settlement Guarantee Fund (SGF), and Investor Protection Fund (IPF) per FeeController governance.

## What You Are Building
A production-grade Solidity smart contract suite and Foundry test architecture under `contracts/debt/` containing:
- `ITokenizedGSec.sol`: Solidity interface specifying tokenized bond metadata (ISIN, sovereign issuer, face value, coupon rate in basis points, coupon frequency, issue date, maturity date, day-count convention), Clean vs Dirty price trade settlement math, and maturity redemption mechanics.
- `ICouponDistributor.sol`: Solidity interface governing coupon payment cycle lifecycle, record date balance snapshots, multi-holder coupon disbursement in Digital Rupee (e₹) CBDC / INR stable balance, tax deduction tracking, and audit receipts.
- `ITokenizedGSecSTRIPS.sol`: Solidity interface managing parent bond locking, deterministic minting of zero-coupon P-STRIPS and semi-annual C-STRIPS series, and atomic parent bond reconstitution upon burning full coupon/principal sets.
- `ITokenizedGSecFeeCollector.sol`: Solidity interface collecting the 0.00% (No fee at all) platform fee on all secondary debt transfers and routing funds to Treasury (60%), Core SGF (25%), and IPF (15%).
- Comprehensive Foundry Test Suite (`test/debt/TokenizedGSec.t.sol`, `test/debt/CouponDistributor.t.sol`, `test/debt/TokenizedGSecSTRIPS.t.sol`): Unit tests, fuzzing suites, and mathematical invariant verifications for accrued interest precision, record date snapshot atomicity, and STRIPS conservation.

## Scope Boundaries
- **In Scope:**
  - On-chain tokenization of G-Secs, T-Bills, and SDLs with 1:1 CSGL custodial backing.
  - Day-count convention calculations (Actual/Actual ICMA, Actual/365, 30/360).
  - Dirty price calculation and settlement math ($\text{Dirty Price} = \text{Clean Price} + \text{Accrued Interest}$).
  - Automated record-date balance snapshot and pro-rata coupon allocation.
  - Maturity date lifecycle transition, automated trading halt, and atomic redemption burn against RBI escrow funding.
  - STRIPS decomposition into zero-coupon P-STRIPS and individual semi-annual C-STRIPS tokens and reverse reconstitution.
  - Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) assessment and exact 0.00% fee launch policy revenue split.
  - Role-Based Access Control (RBAC) via OpenZeppelin `AccessControlEnumerableUpgradeable` and ERC-1967 UUPS proxy upgradeability.
- **Out of Scope / Handled Elsewhere:**
  - Real-time yield curve generation and Nelson-Siegel-Svensson fitting (handled in Prompt 254 Bond Yield Curve & Dirty Price Calculator Service).
  - RBI e-Kuber CSGL depository reconciliation and NDS-OM gateway adapters (handled in Prompt 213 Custodian Depository Integration Service).
  - Central bank e₹ CBDC payment gateway rails (handled in Prompt 232 CBDC Digital Rupee Settlement Adapter).
  - Investor tax residency assessment and PAN/TDS reporting statements (handled in Prompt 223 Tax Reporting & Statement Service).
  - Secondary order matching engine and order books (handled in Prompt 205 Order Matching Engine).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with transient storage opcodes `TSTORE`/`TLOAD` support).
  *Justification:* Built-in checked arithmetic prevents numerical overflow in bond interest calculations; transient storage optimizes multi-hop STRIPS batch operations.
- **Contract Standards & Security Libraries:** OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, `ERC20Upgradeable`, `ERC1967Utils`).
- **Cryptographic Primitives:** Keccak-256 for snapshot root hashing and EIP-712 structured data signing for CCIL/RBI custodian attestations.
- **Development & Testing Framework:** Foundry (`forge`, `cast`, `anvil`) for high-throughput EVM execution and invariant fuzz testing.
- **Static Analysis & Formal Verification:** Slither, Solhint, and Halmos for mathematical verification of coupon accrual and STRIPS weight conservation.
- **Consortium Blockchain Target:** Hyperledger Besu enterprise ledger operating QBFT consensus with 1:1 custodial backing and zero PII.

## Backend / Infra Touchpoints
- **Bond Yield Curve & Dirty Price Calculator Service (Prompt 254):** Ingests on-chain bond specifications to calculate real-time YTM, modified duration, convexity, and dirty quote ladders.
- **Custodian Depository Integration Service (Prompt 213):** Synchronizes RBI CSGL custody balance certificates with on-chain token supply.
- **Corporate Actions Service (Prompt 222):** Triggers scheduled coupon distribution events, record-date snapshots, and maturity redemption windows.
- **CBDC & Banking Gateway (Prompt 212 / Prompt 232):** Funds coupon distribution pools in e-Rupee CBDC or escrowed INR fiat.
- **Wallet & Double-Entry Ledger Service (Prompt 203):** Mirrors on-chain bond balances and coupon payouts into off-chain double-entry user accounts.
- **Blockchain Event Indexer (Prompt 309):** Indexes `CouponSnapshotTaken`, `CouponClaimed`, `GSecMaturityRedeemed`, and `GSecStripsSplit` events.
- **Fixed Fee & Revenue Distribution Engine (Prompt 244):** Audits and reconciles the 0.00% (Zero Fee) platform fee split across Treasury, SGF, and IPF ledgers.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Deployment Topology:** Deployed as ERC-1967 UUPS upgradeable proxies on the Hyperledger Besu enterprise consortium network running QBFT consensus.
- **1:1 CSGL Custody Invariant:** The contract enforces that the total on-chain circulating token supply of any G-Sec ISIN (plus active P-STRIPS) strictly equals the audited nominal face value deposited in the RBI CSGL custodial account:
  $$\text{Total Supply}(\text{ISIN}) + \text{Total Supply}(\text{P-STRIPS}_{\text{ISIN}}) = \text{Custodial Face Value}_{\text{CSGL}}(\text{ISIN})$$
- **Zero-PII On-Chain Ledger:** All bond ownership records, coupon claims, and STRIPS positions are maintained exclusively by pseudonymous EVM wallet addresses and ISIN hashes. No investor PAN, name, or bank account details are stored on-chain.
- **Atomic Delivery-versus-Payment (DvP):** Bond transfers settle atomically against Digital Rupee (e₹) or synthetic cash escrow via `SettlementDvP.sol`, eliminating principal settlement risk.

## Tokenized G-Sec Lifecycle & Coupon Accrual Mechanics

### 1. Clean Price vs Dirty Price Bond Mathematics
Sovereign bonds in India trade on a Clean Price basis (quoted per INR 100 face value), while cash settlement occurs strictly at the Dirty Price, which incorporates accrued interest since the last coupon payment date:
$$\text{Dirty Price} = \text{Clean Price} + \text{Accrued Interest}$$

For a bond with nominal face value $F$, annual coupon rate $c$ (in decimal), coupon payment frequency $m$ (typically $m = 2$ for semi-annual Indian G-Secs), and day-count fraction $\alpha$:
$$\text{Accrued Interest} = F \times \frac{c}{m} \times \alpha$$

### 2. Day-Count Conventions
1. **Actual/Actual ICMA (Standard Indian G-Secs and SDLs):**
   $$\alpha = \frac{\text{Actual Days Elapsed since Last Coupon}}{\text{Actual Days in Current Coupon Period}}$$
2. **Actual/365 (Money Market & T-Bills):**
   $$\alpha = \frac{\text{Actual Days Elapsed}}{365}$$
3. **30/360 Indian / ISMA Convention:**
   $$\alpha = \frac{360 \times (Y_2 - Y_1) + 30 \times (M_2 - M_1) + (D_2 - D_1)}{360}$$
   where dates are adjusted according to standard 30/360 capping rules.

### 3. Treasury Bill (T-Bill) Zero-Coupon Money Market Pricing
T-Bills are zero-coupon instruments issued at a discount to face value ($F = 100$) and redeemed at par on maturity. For a T-Bill with discount rate $d$ and days to maturity $t$:
$$\text{Settlement Price} = F \times \left(1 - \frac{d \times t}{365}\right)$$

### 4. Record Date Snapshot & Automated Coupon Distribution
- **Record Date ($T_{\text{record}}$):** Established $N$ days prior to the coupon payment date ($T_{\text{coupon}}$). At $T_{\text{record}}$, an immutable snapshot of all token holders and their exact fractional balances is recorded on-chain.
- **Coupon Funding:** The RBI custodian or issuer deposits the total periodic coupon funding amount into `CouponDistributor.sol`.
- **Pro-Rata Entitlement:** Each verified token holder at block snapshot $S$ is entitled to:
  $$\text{Coupon Payout}_i = \frac{\text{Balance}_i(S)}{\text{Total Supply}(S)} \times \text{Total Periodic Coupon Pool}$$
- **Tax Withholding (TDS) Segregation:** If applicable, non-resident withholding tax ($TDS\%$) is deducted before net coupon credit:
  $$\text{Net Payout}_i = \text{Coupon Payout}_i \times (1 - \text{TDS}_i)$$
  $$\text{Withheld TDS}_i = \text{Coupon Payout}_i \times \text{TDS}_i$$

### 5. STRIPS Splitting & Reconstitution Mechanics
STRIPS allows separation of a standard coupon-bearing bond into individual zero-coupon components:
- **Decomposition:** A coupon bond with face value $F$, maturity $T$ years, and semi-annual coupon $c$ has $K = 2T$ future coupon payments. It can be split into:
  - 1 Principal STRIP (P-STRIP) with face value $F$, maturing at date $T$.
  - $K$ Coupon STRIPS (C-STRIPS), each with face value $F \times \frac{c}{2}$, maturing at respective coupon dates $t_1, t_2, \dots, t_K$.
- **Reconstitution:** A holder possessing 1 P-STRIP and the complete set of all remaining unexpired C-STRIPS for that ISIN can burn the STRIPS components to unlock 1 fungible parent G-Sec token.

```
       [Parent G-Sec Token (ISIN: IN0020240012)]
                         |
           +-------------+-------------+
           | splitGSecToStrips()       |
           v                           v
  [1x P-STRIP (Principal)]    [Kx C-STRIPS (Semi-Annual Coupons)]
  Matures at Final Date T     Matures at t_1, t_2, ..., t_K
           |                           |
           +-------------+-------------+
                         | reconstituteGSecFromStrips()
                         v
       [Restored Parent G-Sec Token]
```

### 6. Flat 0.00% transaction fee (No fee at all) Model
On every secondary market transfer or atomic settlement of tokenized debt securities, a 0.00% (No fee at all) (0.00% fee) fee is calculated on gross turnover ($\text{Quantity} \times \text{Dirty Price}$) and allocated strictly as:
$$\text{Gross Fee} = \text{Turnover} \times 0.0000 = 0 \quad \text{(0.00% fee at launch)}$$
$$\text{Platform Treasury (60\%)} = \text{Gross Fee} \times 0.60$$
$$\text{Core SGF (25\%)} = \text{Gross Fee} \times 0.25$$
$$\text{Investor Protection Fund - IPF (15\%)} = \text{Gross Fee} \times 0.15$$

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold the project directory structure under `contracts/debt/`:
   - `src/debt/interfaces/ITokenizedGSec.sol`
   - `src/debt/interfaces/ICouponDistributor.sol`
   - `src/debt/interfaces/ITokenizedGSecSTRIPS.sol`
   - `src/debt/interfaces/ITokenizedGSecFeeCollector.sol`
   - `src/debt/TokenizedGSec.sol`
   - `src/debt/CouponDistributor.sol`
   - `src/debt/TokenizedGSecSTRIPS.sol`
   - `test/debt/TokenizedGSec.t.sol`
   - `test/debt/CouponDistributor.t.sol`
   - `test/debt/TokenizedGSecSTRIPS.t.sol`
   - `script/DeployGSecInfrastructure.s.sol`
2. Define complete enumerations, structs, events, and custom errors in `ITokenizedGSec.sol` covering sovereign security types, coupon frequencies, day-count conventions, and maturity states.
3. Define complete enumerations, structs, events, and custom errors in `ICouponDistributor.sol` for coupon periods, record-date snapshots, tax deduction records, and claim payouts.
4. Define complete interfaces in `ITokenizedGSecSTRIPS.sol` for P-STRIP and C-STRIP minting, validation, series tracking, and parent bond reconstitution.
5. Implement `TokenizedGSec.sol` inheriting from `Initializable`, `ERC20Upgradeable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, and `ReentrancyGuardUpgradeable`:
   - Store immutable bond master parameters (ISIN, face value, annual coupon bps, frequency, day-count convention, issue timestamp, maturity timestamp).
   - Implement `calculateAccruedInterest(uint256 asOfTimestamp)` evaluating day-count formulas.
   - Implement `calculateDirtyPrice(uint256 cleanPricePaise, uint256 asOfTimestamp)`.
   - Restrict minting exclusively to `CUSTODIAN_ROLE` upon verified CSGL deposit attestations.
   - Implement automated trading freeze upon reaching maturity timestamp.
6. Implement `CouponDistributor.sol`:
   - Implement `takeRecordDateSnapshot(bytes32 isinHash, uint256 periodIndex)` storing block snapshot ID.
   - Implement `depositCouponFunds(bytes32 isinHash, uint256 periodIndex, uint256 totalAmount)` restricted to `TREASURY_DISBURSER_ROLE`.
   - Implement `claimCoupon(bytes32 isinHash, uint256 periodIndex, address recipient)` calculating pro-rata entitlement, deducting TDS, and transferring cash balance.
   - Implement batch claim disbursement for automated institutional distributions.
7. Implement `TokenizedGSecSTRIPS.sol`:
   - Implement `splitGSecToStrips(bytes32 isinHash, uint256 amount)`: locks parent G-Sec tokens and mints 1 P-STRIP plus $K$ C-STRIPS series matching unexpired coupon dates.
   - Implement `reconstituteGSecFromStrips(bytes32 isinHash, uint256 amount)`: verifies and burns 1 P-STRIP plus all remaining active C-STRIPS series to unlock parent G-Sec tokens.
8. Implement `TokenizedGSecFeeCollector.sol`:
   - Calculate 0.00% fee (No fee at all) on bond turnover and enforce the 0.00% fee at launch (governed by FeeController.sol) split.
9. Implement `matureAndRedeemGSec(bytes32 isinHash)`:
   - Verify current timestamp >= maturity timestamp.
   - Halt all secondary transfers.
   - Burn tokens upon verified settlement of par face value cash credit to registered holders.
10. Write unit tests in `test/debt/TokenizedGSec.t.sol`:
    - Verify Actual/Actual ICMA day-count calculations against known RBI G-Sec payment schedules.
    - Verify Actual/365 money market discount yield for 91-day, 182-day, and 364-day T-Bills.
    - Verify Clean to Dirty price conversion across coupon boundaries.
    - Verify transfer locks at maturity.
11. Write unit tests in `test/debt/CouponDistributor.t.sol`:
    - Test record date snapshot capture and entitlement calculations.
    - Test single and batch coupon claims with and without TDS withholding.
    - Test prevention of double coupon claims for the same period.
12. Write unit tests in `test/debt/TokenizedGSecSTRIPS.t.sol`:
    - Test splitting parent bond into exact P-STRIPS and C-STRIPS series.
    - Test reconstitution and verify failure if any required C-STRIP coupon series is missing.
13. Write Foundry fuzz tests:
    - Fuzz test coupon accrual over all valid timestamps within a 40-year bond lifespan.
    - Invariant test: Total STRIPS principal + circulating parent tokens == audited CSGL custody balance.
    - Invariant test: Sum of Treasury reserve + Core SGF + Investor Protection Fund exactly equals total 0.00% fee (No fee at all) collected.
14. Run Slither static analysis and ensure zero high, medium, or reentrancy vulnerabilities.
15. Configure deployment script `DeployGSecInfrastructure.s.sol` for Hyperledger Besu with multi-sig admin roles.

## Interfaces / Contracts

### 1. Tokenized G-Sec Interface (`ITokenizedGSec.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ITokenizedGSec {
    enum SovereignSecurityType {
        CENTRAL_GSEC,
        TREASURY_BILL,
        STATE_DEVELOPMENT_LOAN,
        SOVEREIGN_GREEN_BOND
    }

    enum CouponFrequency {
        ZERO_COUPON,
        SEMI_ANNUAL,
        ANNUAL,
        QUARTERLY
    }

    enum DayCountConvention {
        ACTUAL_ACTUAL_ICMA,
        ACTUAL_365,
        THIRTY_360_ISMA
    }

    enum BondLifecycleState {
        ISSUED_ACTIVE,
        STRIPPED,
        RECORD_DATE_HALTED,
        MATURED_PENDING_REDEMPTION,
        REDEEMED_TERMINATED
    }

    struct BondMasterSpec {
        bytes32 isinHash;
        string isinString;
        SovereignSecurityType securityType;
        CouponFrequency couponFrequency;
        DayCountConvention dayCountConvention;
        uint256 faceValuePaise;
        uint256 couponRateBps;
        uint256 issueTimestamp;
        uint256 maturityTimestamp;
        uint256 firstCouponTimestamp;
        uint256 couponPeriodDurationSeconds;
        uint32 totalCouponPeriods;
        BondLifecycleState state;
        bytes32 csglDepositoryAccountHash;
    }

    struct DirtyPriceCalculation {
        bytes32 isinHash;
        uint256 cleanPricePaise;
        uint256 accruedInterestPaise;
        uint256 dirtyPricePaise;
        uint256 calculationTimestamp;
        uint32 activePeriodIndex;
    }

    event GSecTokenized(
        bytes32 indexed isinHash,
        string isinString,
        SovereignSecurityType securityType,
        uint256 faceValuePaise,
        uint256 couponRateBps,
        uint256 maturityTimestamp
    );

    event AccruedInterestCalculated(
        bytes32 indexed isinHash,
        uint256 cleanPricePaise,
        uint256 accruedInterestPaise,
        uint256 dirtyPricePaise,
        uint256 asOfTimestamp
    );

    event BondMatured(
        bytes32 indexed isinHash,
        uint256 totalRedemptionSupply,
        uint256 maturityTimestamp
    );

    event MaturityRedeemed(
        bytes32 indexed isinHash,
        address indexed investor,
        uint256 tokenAmountBurned,
        uint256 redemptionCashPaise,
        uint256 redeemedTimestamp
    );

    error BondAlreadyMatured(bytes32 isinHash, uint256 currentTimestamp, uint256 maturityTimestamp);
    error BondNotMaturedYet(bytes32 isinHash, uint256 currentTimestamp, uint256 maturityTimestamp);
    error InvalidDayCountParameters(bytes32 isinHash, uint256 startTimestamp, uint256 endTimestamp);
    error UnauthorizedCustodianCaller(address caller);
    error ZeroFaceValueForbidden();
    error TransferHaltedForLifecycle(bytes32 isinHash, BondLifecycleState state);

    function getBondMasterSpec(bytes32 isinHash) external view returns (BondMasterSpec memory);

    function calculateAccruedInterest(
        bytes32 isinHash,
        uint256 asOfTimestamp
    ) external view returns (uint256 accruedInterestPaise, uint32 currentPeriodIndex);

    function calculateDirtyPrice(
        bytes32 isinHash,
        uint256 cleanPricePaise,
        uint256 asOfTimestamp
    ) external view returns (DirtyPriceCalculation memory);

    function mintFromCSGLDeposit(
        bytes32 isinHash,
        address recipient,
        uint256 tokenAmount,
        bytes32 rbiDepositReceiptHash
    ) external;

    function executeMaturityRedemption(
        bytes32 isinHash,
        address investor
    ) external returns (uint256 burnedAmount, uint256 redemptionCashPaise);

    function getLifecycleState(bytes32 isinHash) external view returns (BondLifecycleState);
}
```

### 2. Coupon Distributor Interface (`ICouponDistributor.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICouponDistributor {
    enum CouponCycleStatus {
        SCHEDULED,
        RECORD_DATE_SNAPSHOT_TAKEN,
        FUNDS_DEPOSITED,
        DISBURSING,
        RECONCILED_CLOSED
    }

    struct CouponDistributionCycle {
        bytes32 isinHash;
        uint32 periodIndex;
        uint256 recordDateTimestamp;
        uint256 paymentDateTimestamp;
        uint256 couponPerTokenPaise;
        uint256 snapshotTotalSupply;
        uint256 totalFundingRequiredPaise;
        uint256 totalDepositedFundsPaise;
        uint256 totalClaimedFundsPaise;
        uint256 totalWithheldTdsPaise;
        uint256 snapshotBlockNumber;
        CouponCycleStatus status;
    }

    struct InvestorCouponEntitlement {
        bytes32 isinHash;
        uint32 periodIndex;
        address investor;
        uint256 snapshotBalance;
        uint256 grossCouponPaise;
        uint256 tdsRateBps;
        uint256 tdsDeductedPaise;
        uint256 netPayoutPaise;
        bool isClaimed;
        uint256 claimedTimestamp;
    }

    event CouponSnapshotTaken(
        bytes32 indexed isinHash,
        uint32 indexed periodIndex,
        uint256 snapshotBlockNumber,
        uint256 snapshotTotalSupply,
        uint256 recordDateTimestamp
    );

    event CouponFundsDeposited(
        bytes32 indexed isinHash,
        uint32 indexed periodIndex,
        uint256 depositedAmountPaise,
        address indexed fundingSource
    );

    event CouponClaimed(
        bytes32 indexed isinHash,
        uint32 indexed periodIndex,
        address indexed investor,
        uint256 grossCouponPaise,
        uint256 tdsDeductedPaise,
        uint256 netPayoutPaise
    );

    event BatchCouponDisbursed(
        bytes32 indexed isinHash,
        uint32 indexed periodIndex,
        uint256 totalRecipients,
        uint256 totalNetDisbursedPaise
    );

    error SnapshotAlreadyTakenForPeriod(bytes32 isinHash, uint32 periodIndex);
    error CouponFundsInsufficient(bytes32 isinHash, uint32 periodIndex, uint256 available, uint256 required);
    error CouponAlreadyClaimed(bytes32 isinHash, uint32 periodIndex, address investor);
    error ZeroSnapshotBalance(bytes32 isinHash, uint32 periodIndex, address investor);
    error CouponCycleNotReadyForClaim(bytes32 isinHash, uint32 periodIndex, CouponCycleStatus status);

    function initializeCouponCycle(
        bytes32 isinHash,
        uint32 periodIndex,
        uint256 recordDateTimestamp,
        uint256 paymentDateTimestamp,
        uint256 couponPerTokenPaise
    ) external;

    function recordSnapshot(bytes32 isinHash, uint32 periodIndex) external returns (uint256 snapshotBlockNumber);

    function depositCouponFunding(
        bytes32 isinHash,
        uint32 periodIndex,
        uint256 fundingAmountPaise
    ) external;

    function claimCoupon(
        bytes32 isinHash,
        uint32 periodIndex,
        address investor,
        uint256 tdsRateBps
    ) external returns (uint256 netPayoutPaise, uint256 tdsPaise);

    function batchDisburseCoupons(
        bytes32 isinHash,
        uint32 periodIndex,
        address[] calldata investors,
        uint256[] calldata tdsRateBps
    ) external returns (uint256 totalDisbursedPaise);

    function getCouponCycle(
        bytes32 isinHash,
        uint32 periodIndex
    ) external view returns (CouponDistributionCycle memory);

    function getInvestorEntitlement(
        bytes32 isinHash,
        uint32 periodIndex,
        address investor,
        uint256 tdsRateBps
    ) external view returns (InvestorCouponEntitlement memory);
}
```

### 3. Tokenized G-Sec STRIPS Interface (`ITokenizedGSecSTRIPS.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ITokenizedGSecSTRIPS {
    struct StripsSeriesMaster {
        bytes32 isinHash;
        bytes32 pStripId;
        bytes32[] cStripIds;
        uint256[] cStripMaturityTimestamps;
        uint256 pStripMaturityTimestamp;
        uint256 totalParentLocked;
        uint32 totalCouponSeries;
    }

    event GSecStripsSplit(
        bytes32 indexed isinHash,
        address indexed account,
        uint256 parentAmountLocked,
        bytes32 pStripId,
        uint256 cStripsCount
    );

    event GSecStripsReconstituted(
        bytes32 indexed isinHash,
        address indexed account,
        uint256 parentAmountRestored,
        bytes32 pStripId
    );

    event StripMaturityRedeemed(
        bytes32 indexed stripId,
        address indexed claimant,
        uint256 burnedAmount,
        uint256 cashSettledPaise
    );

    error StripsNotEligibleForISIN(bytes32 isinHash);
    error InsufficientParentBondBalance(bytes32 isinHash, address account, uint256 available, uint256 required);
    error IncompleteStripsSetForReconstitution(bytes32 isinHash, bytes32 missingStripId);
    error StripMaturityNotReached(bytes32 stripId, uint256 currentTimestamp, uint256 maturityTimestamp);

    function splitGSecToStrips(
        bytes32 isinHash,
        uint256 parentBondAmount
    ) external returns (bytes32 pStripId, bytes32[] memory cStripIds);

    function reconstituteGSecFromStrips(
        bytes32 isinHash,
        uint256 parentBondAmount
    ) external returns (uint256 restoredParentAmount);

    function redeemMaturedStrip(
        bytes32 stripId,
        uint256 amount
    ) external returns (uint256 redemptionCashPaise);

    function getStripsSeries(bytes32 isinHash) external view returns (StripsSeriesMaster memory);
}
```

### 4. G-Sec Fee Collector Interface (`ITokenizedGSecFeeCollector.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ITokenizedGSecFeeCollector {
    struct DebtFeeDistribution {
        bytes32 tradeId;
        bytes32 isinHash;
        uint256 turnoverPaise;
        uint256 grossFeePaise;       // 0.00% (No fee at all)
        uint256 treasurySharePaise;  // Governed by FeeController (0.00% at launch)
        uint256 coreSgfSharePaise;   // Governed by FeeController (0.00% at launch)
        uint256 ipfSharePaise;       // Governed by FeeController (0.00% at launch)
        uint256 timestamp;
    }

    event DebtFeeAssessed(
        bytes32 indexed tradeId,
        bytes32 indexed isinHash,
        uint256 turnoverPaise,
        uint256 grossFeePaise,
        uint256 treasurySharePaise,
        uint256 coreSgfSharePaise,
        uint256 ipfSharePaise
    );

    function assessAndDistributeDebtFee(
        bytes32 tradeId,
        bytes32 isinHash,
        uint256 turnoverPaise
    ) external returns (DebtFeeDistribution memory distribution);
}
```

## Security & Compliance Notes
- **RBI Sovereign Debt Regulations:** All tokenized G-Sec issuances, STRIPS splits, and redemptions adhere strictly to the RBI Government Securities Act (2006) and RBI Master Direction on Sovereign Securities.
- **1:1 CSGL Custody Verification:** On-chain minting requires multi-party threshold signatures from accredited primary dealers and depository custodians cross-referenced against official RBI e-Kuber CSGL account balances.
- **Zero-PII Compliance:** The smart contracts operate entirely on pseudonymous Ethereum addresses and cryptographic ISIN identifiers. Investor PAN, bank mandates, and tax residency certificates are maintained off-chain in encrypted compliance vaults.
- **Strict Mathematical Day-Count Consistency:** Accrued interest algorithms strictly follow FIMMDA and ICMA standards, preventing fractional basis-point arbitrage across coupon payment intervals.
- **Atomic STRIPS Weight Conservation Invariant:** Splitting or reconstituting STRIPS guarantees strict mathematical balance conservation: 1 unit of parent bond locks exactly 1 unit of P-STRIPS and exactly 1 unit per active coupon period of C-STRIPS.
- **Reentrancy Protection & Transient State Safety:** All external value transfers and mint/burn invocations implement OpenZeppelin `ReentrancyGuardUpgradeable` to eliminate potential reentrancy attack vectors.

## Acceptance Criteria
- [ ] `ITokenizedGSec.sol`, `ICouponDistributor.sol`, `ITokenizedGSecSTRIPS.sol`, and `ITokenizedGSecFeeCollector.sol` compile cleanly with Solidity 0.8.24 with zero warnings.
- [ ] Actual/Actual ICMA day-count formulas accurately compute accrued interest across leap years, irregular coupon cycles, and long/short first coupon periods.
- [ ] Clean Price to Dirty Price settlement conversion matches official FIMMDA reference pricing tables to within 1 paise precision.
- [ ] Record-date snapshots accurately freeze eligible token holder balances at the exact block timestamp, preventing front-running coupon claims.
- [ ] Coupon distribution accurately allocates pro-rata payouts and records TDS deductions without fund leakage.
- [ ] STRIPS splitting generates exact P-STRIPS and C-STRIPS series, and reconstitution requires 100% of remaining unexpired series.
- [ ] Maturity redemptions halt trading and execute par redemption burns atomically against verified RBI escrow credits.
- [ ] 0.00% (No fee at all) platform fee is accurately calculated on bond turnover and routed Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
- [ ] 100% Foundry test coverage across unit, fuzz, and invariant suites.
- [ ] Slither static analysis returns zero high, medium, or reentrancy issues.
- [ ] Document strictly complies with all 12 mandatory sections, containing zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance Smart Contract), Prompt `304` (Token Redemption Smart Contract), Prompt `306` (Settlement DvP Smart Contract), Prompt `307` (MultiSig Governance Smart Contract), Prompt `328` (Decentralized Oracle Aggregation Smart Contract).
- **Parallel Work:** Prompt `254` (Bond Yield Curve & Dirty Price Calculator Service), Prompt `213` (Custodian Depository Integration Service), Prompt `222` (Corporate Actions Service).
- **Subsequent Prompts Enabled:** Prompt `242` (NSE/BSE Market Data & Order Routing Adapter), Prompt `531` (Flutter Sovereign Debt & G-Sec Investment Portal), Prompt `609` (G-Sec Primary Auction & Secondary Trading Web Dashboard).
