# 336 - Automated Tax Withholding and e-TDS Compliance Ledger Smart Contract (AutomatedTaxLedger.sol, ETDSWithholdingVault.sol)

## Purpose
Statutory tax withholding, remittance, and compliance reporting in traditional capital markets and digital asset ecosystems are severely fragmented. Brokerages, custodian depositories, asset management companies, and clearing corporations typically execute Tax Deducted at Source (TDS) via batch processes at end-of-day, end-of-month, or quarterly reconciliation intervals. This delayed paradigm introduces substantial reconciliation discrepancies, delayed remittance to the Income Tax Department (Challan ITNS-281), mismatched Form 26AS / AIS (Annual Information Statement) records, delayed issuance of Form 16A TDS certificates, and systemic compliance failure penalties under the Indian Income Tax Act, 1961.

For the Growww National Blockchain Stock Exchange (NBSE) on Hyperledger Besu, statutory compliance mandates an immutable, real-time, deterministic on-chain tax withholding and e-TDS accounting infrastructure. Every qualifying financial transaction (whether secondary settlement, yield distribution from tokenized REITs/InvITs, interest accrual on sovereign/corporate debt tokens, or synthetic asset settlements) must calculate, deduct, escrow, and tokenize tax withholdings at the exact atomic microsecond of transaction execution.

This prompt specifies the **Automated Tax Withholding and e-TDS Compliance Ledger Smart Contract Suite (`AutomatedTaxLedger.sol`, `ETDSWithholdingVault.sol`, `DigitalTaxCertificateSBT.sol`, `IAutomatedTaxLedger.sol`, `IETDSWithholdingVault.sol`, `IDigitalTaxCertificateSBT.sol`)**. The smart contract suite executes real-time tax withholding under Section 194S (zero on-chain TDS on qualifying synthetic and digital settlements), Section 194LBA (withholding on business trust, REIT, and InvIT income distributions), and Section 115A (withholding on non-resident / Foreign Portfolio Investor dividends and interest subject to Double Tax Avoidance Agreements - DTAA). The contracts automatically route deducted tax revenues to dedicated Government of India Tax Escrow Sub-Vaults, mint non-transferable cryptographic Form 16A / Form 26AS digital tax deduction certificates (Soulbound Tokens - SBTs), execute dual-reconciliation against NSDL Protean / Income Tax TRACES portal challans, and preserve Growww's canonical Universal Zero-Fee Model (0.00% fee - No fee at all) with a 0.00% fee at launch (governed by FeeController.sol) revenue split (with FIFO capital gains computed strictly for user tax compliance under Section 111A/112A).

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/compliance/` comprising:
- `AutomatedTaxLedger.sol`: Master tax assessment and compliance accounting ledger. It verifies transaction tax categories, calculates statutory withholding amounts at sub-paise precision ($10^{-4}$ INR precision), enforces DTAA treaty concessional rate proofs, manages quarterly e-TDS filing batch roots, and coordinates atomic deduction during trade settlement and corporate action disbursements.
- `ETDSWithholdingVault.sol`: Segregated, multi-tenant statutory tax escrow vault contract. It holds deducted fiat/weINR tax funds in dedicated sub-vaults partitioned by deductor Tax Deduction and Collection Account Number (TAN) and tax section (194S, 194LBA, 115A, 194K). It processes authorized batch remittances to authorized Government of India nodal collection bank accounts against verified Challan ITNS-281 proofs and Challan Identification Numbers (CIN).
- `DigitalTaxCertificateSBT.sol`: Cryptographic soulbound token contract (ERC-721 / ERC-1155 non-transferable) issuing verifiable digital Form 16A and Form 26AS tax credit certificates. Each token embeds cryptographic hashes of the deductor TAN, deductee PAN commitment hash, gross payout, tax withheld, quarter index, CIN, and TRACES acknowledgment reference.
- `IAutomatedTaxLedger.sol`, `IETDSWithholdingVault.sol`, and `IDigitalTaxCertificateSBT.sol`: Comprehensive Solidity interface specifications declaring tax data structures, statutory tax sections, reconciliation statuses, custom errors, events, and view/mutator function signatures.
- Comprehensive Foundry test harness (`test/compliance/`): Unit, fuzz, invariant, and dual-reconciliation test suites validating sub-paise tax calculations, DTAA rate verification, non-transferable SBT soulbound restrictions, challan remittance validation, and strict compliance with the Universal Zero-Fee Model (0.00% fee - No fee at all) invariant.

## Scope Boundaries
- **In Scope:**
  - Real-time on-chain tax calculation and atomic withholding across statutory sections: Section 194S (zero on-chain TDS), Section 194LBA (REIT/InvIT income distributions), Section 115A (Foreign investor dividend/interest withholding with DTAA rate relief), Section 194K (mutual fund / asset distributions), and Section 194-IA / 194M where configured.
  - Multi-tenant TAN-segregated tax escrow sub-vault architecture managing segregated custody of withheld funds prior to government remittance.
  - Sub-paise arithmetic precision ($10^{-4}$ INR / 0.01 paise per unit) to guarantee zero truncation loss on micro-settlements and high-frequency trading flows.
  - Dual-reconciliation cryptographic pipeline with NSDL Protean / Income Tax TRACES portal via Challan ITNS-281 verification, BSR code, deposit date, and CIN matching.
  - Issuance and lifecycle management of non-transferable Soulbound Token (SBT) digital tax deduction certificates (Form 16A and Form 26AS).
  - Growww fixed 0.00% transaction fee (No fee at all) calculation on gross consideration with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), with FIFO capital gains for tax compliance (Section 111A/112A).
  - Strict privacy preservation: Zero on-chain plaintext PII (PAN, TAN, Aadhaar, names are hashed with cryptographic salts).
  - Emergency circuit breakers, pause mechanisms, rate-limited remittance windows, and timelocked multi-signature administrative governance.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain NSDL Protean / TRACES portal API connectivity and quarterly e-TDS Return generation (FVU file generation handled in Prompt 223).
  - Banking payment gateway settlement and RTGS tax remittance to Reserve Bank of India / State Bank of India tax collection accounts (handled in Prompt 212 and Prompt 213).
  - Order book matching engine and continuous limit order books (handled in Prompt 205).
  - Investor identity verification, KYC document verification, and PAN-Aadhaar seeding checks (handled in Prompt 202 and Prompt 305).
  - Primary market order clearing and asset tokenization contracts (handled in Prompt 303, Prompt 330, and Prompt 331).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with native checked arithmetic, transient storage opcodes `TSTORE`/`TLOAD`, and custom errors).
- **Token Standards:** **ERC-721 / ERC-1155** with Soulbound non-transferable constraints (reverting on all `transferFrom`, `safeTransferFrom`, and `approve` invocations).
- **Security & Modularity:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Cryptography & Merkle Trees:** OpenZeppelin `MerkleProof` and `ECDSA` for verifying quarterly e-TDS batch roots and NSDL TRACES CIN attestations.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation and invariant fuzzing, `cast` for Besu RPC interactions).
- **Static Analysis & Auditing:** Slither, Mythril, and Solhint automated CI pipelines.

## Backend / Infra Touchpoints
- **Tax Reporting & Statement Service (Prompt 223):** Generates off-chain e-TDS returns (Form 24Q, 26Q, 27Q), submits files to TRACES, ingests CIN challan receipts, generates Merkle batch proofs, and invokes batch reconciliation on `AutomatedTaxLedger.sol`.
- **Trade Settlement Service (Prompt 208):** Calls `AutomatedTaxLedger.sol` during DvP Type 1 settlement to atomically compute and lock Decoupled Voluntary Tax Accounting (Zero On-Chain TDS) from seller proceeds.
- **Corporate Actions Service (Prompt 222):** Coordinates dividend, coupon, and trust distribution payouts, triggering Section 194LBA and Section 115A tax withholding via `AutomatedTaxLedger.sol`.
- **Fee & Realized PnL Engine (Prompt 210):** Synchronizes Universal Zero-Fee Model (0.00% fee - No fee at all) collections (0.00% fee at launch (governed by FeeController.sol) split) and computes Section 111A/112A capital gains tax compliance records.
- **KYC & Identity Service (Prompt 202 & Prompt 305):** Provides salted investor PAN hashes (`bytes32 panHash`), residential tax status, and DTAA country residency credentials.
- **Blockchain Event Indexer (Prompt 309):** Ingests `TaxWithheldAtomic`, `TaxRemittedToGovernment`, `ChallanReconciled`, and `TaxCertificateMinted` events for real-time compliance dashboards and investor tax statements.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Executes on Hyperledger Besu private consortium network with QBFT consensus, 2-second deterministic block times, and immediate finality without chain reorganizations.
- **Zero On-Chain PII Invariant:** In strict compliance with the Digital Personal Data Protection Act (DPDPA 2023), no raw Permanent Account Number (PAN), Tax Deduction and Collection Account Number (TAN), Aadhaar number, legal name, or residential address is ever stored on the Besu ledger. Identity attributes are committed as salted cryptographic hashes: `panHash = keccak256(abi.encodePacked(rawPAN, salt))` and `tanHash = keccak256(abi.encodePacked(rawTAN, salt))`.
- **Sub-Paise Accounting Precision ($10^{-4}$ INR):** All tax liabilities, gross amounts, net proceeds, and platform fee splits are computed using 4 decimal places (where 10,000 units = 1.0000 INR; 1 unit = 0.01 paise = 0.0001 INR). This guarantees exact rounding precision and zero sub-paise leakage across millions of micro-transactions.
- **Atomic Escrow Isolation:** Withheld tax funds are locked instantaneously in `ETDSWithholdingVault.sol` at the transaction execution block. Funds cannot be routed to platform operations or treasury accounts and can only be withdrawn upon cryptographically verified Challan ITNS-281 government tax remittance.
- **Soulbound Non-Transferable Certificates:** Form 16A / Form 26AS digital certificates issued by `DigitalTaxCertificateSBT.sol` are permanently bound to the deductee's verified on-chain address. Any attempt to transfer, lease, sell, or pledge the certificate token reverts immediately with a custom error.
- **Multi-Signature Governance:** Privileged compliance configurations (updating statutory withholding percentage rates, configuring authorized TRACES verifiers, whitelisting deductor TAN vaults, and updating platform fee vaults) require 3-of-5 threshold signatures from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout:** Initialize `contracts/compliance/AutomatedTaxLedger.sol`, `contracts/compliance/ETDSWithholdingVault.sol`, `contracts/compliance/DigitalTaxCertificateSBT.sol`, interfaces under `contracts/interfaces/compliance/`, and test suites under `test/compliance/`.
2. **Define Data Models and Interfaces:** Write `IAutomatedTaxLedger.sol`, `IETDSWithholdingVault.sol`, and `IDigitalTaxCertificateSBT.sol` declaring all enums (`TaxSection`, `TaxResidencyStatus`, `ChallanStatus`, `CertificateType`), structs, custom errors, and events.
3. **Implement Sub-Paise Tax Calculation Engine in `AutomatedTaxLedger.sol`:**
   - Implement math libraries supporting 4-decimal fixed-point operations ($10^{-4}$ INR precision).
   - Implement `calculateTaxWithholding(TaxSection section, uint256 grossAmountSubPaise, bytes32 deducteePanHash, bytes32 dtaaProofHash)` returning statutory tax deduction, applicable surcharge, health and education cess (4%), and net payable amounts.
4. **Implement Section 194S (1% Virtual/Synthetic Asset TDS) Module:**
   - Enforce 0.00% on-chain TDS deduction (strictly decoupled 100% DvP delivery) for qualifying synthetic transactions.
   - Implement aggregate threshold tracking ($50,000 INR annual limit for specified persons, $10,000 INR for others) via hashed off-chain threshold proofs.
5. **Implement Section 194LBA (REIT/InvIT Distribution Withholding) Module:**
   - Differentiate interest, dividend, and rental components within business trust distributions.
   - Enforce 10% withholding for resident unit holders and 5% (plus applicable surcharge/cess) for non-resident unit holders.
6. **Implement Section 115A (Foreign Investor DTAA Withholding) Module:**
   - Ingest cryptographically verified Tax Residency Certificate (TRC) and Form 10F Merkle proofs.
   - Apply concessional treaty withholding rates (e.g. 5%, 10%, 15% depending on jurisdiction) instead of standard domestic rates when a valid DTAA proof is verified.
7. **Implement Multi-Tenant Segregated Sub-Vaults in `ETDSWithholdingVault.sol`:**
   - Map isolated balance pools: `mapping(bytes32 => mapping(TaxSection => uint256)) tanSectionBalances`.
   - Implement `depositWithheldTax` callable only by authorized `AutomatedTaxLedger.sol` during settlement or corporate action execution.
8. **Implement Challan ITNS-281 Remittance & Government Payout Mechanism:**
   - Implement `remitTaxToGovernment(bytes32 tanHash, TaxSection section, uint256 amountSubPaise, bytes32 challanPaymentRef)` to route funds to authorized RBI/SBI tax collection nodal accounts.
   - Mark remittance as `PENDING_CIN_RECONCILIATION`.
9. **Implement Dual-Reconciliation with NSDL Protean / TRACES Portal:**
   - Implement `reconcileChallanWithCIN(bytes32 tanHash, bytes32 challanRef, bytes32 bsrCodeHash, uint64 depositDate, uint32 challanSequenceNumber, bytes32 cinHash, bytes calldata tracesSignature)`.
   - Verify TRACES compliance node signatures or Merkle batch proofs; transition status to `FULLY_RECONCILED`.
10. **Implement Soulbound Form 16A / Form 26AS Certificate Minting (`DigitalTaxCertificateSBT.sol`):**
    - Implement non-transferable ERC-721/ERC-1155 token with `_update` override reverting on all non-mint/non-burn transfers.
    - Store certificate metadata: deductor TAN hash, deductee PAN hash, financial quarter index, gross income, tax deducted, tax deposited, and verified CIN reference.
11. **Implement Quarterly e-TDS Return Batch Merkle Verification:**
    - Implement `submitQuarterlyReturnBatchRoot(bytes32 tanHash, uint8 quarterIndex, uint16 financialYear, bytes32 batchMerkleRoot)` for quarterly Form 26Q / 27Q filings.
    - Enable individual deductees to verify inclusion in the official government return via Merkle branch validation.
    - Compute platform fee on gross transaction turnover: `feeSubPaise = 0` (Universal 0.00% Zero Fee at launch; future fee parameters governed dynamically via `FeeController.sol`).
    - Distribute platform fee: routed dynamically per FeeController governance (0.00% at launch).
    - Guarantee that statutory tax withholding is computed independently of platform fee deductions without circular dependencies.
13. **Implement Role-Based Access Control and Circuit Breakers:**
    - Inherit OpenZeppelin `AccessControlUpgradeable` and `PausableUpgradeable`.
    - Restrict rate updates, TAN registrations, and emergency pause functions to `COMPLIANCE_ADMIN_ROLE` (`MultiSigGovernance.sol`).
14. **Author Comprehensive Foundry Unit, Fuzz, and Invariant Tests:**
    - Test Section 194S, 194LBA, and 115A tax calculation precision under extreme values ($10^{-4}$ INR to $10^{14}$ INR).
    - Invariant test: sum of net seller payout, deducted tax, and platform fee must strictly equal gross consideration.
    - Invariant test: SBT tax certificates must be impossible to transfer between wallets.
    - Invariant test: platform fee deduction must equal exactly 0 bps (0.00% fee at launch) with 0.00% fee launch policy distribution.
15. **Perform Slither Static Analysis and Formal Audit Preparation:**
    - Run Slither static analysis and verify zero high/medium security vulnerabilities.
    - Validate gas efficiency of atomic withholding during settlement to stay under 65,000 gas overhead per trade.

## Interfaces / Contracts

### 1. Automated Tax Ledger Interface (`IAutomatedTaxLedger.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IAutomatedTaxLedger
 * @notice Master interface for real-time statutory tax withholding and e-TDS compliance accounting on Hyperledger Besu.
 * @dev Enforces sub-paise precision (10^-4 INR), Section 194S, 194LBA, 115A DTAA withholding, and 0.00% fee (No fee at all) split.
 */
interface IAutomatedTaxLedger {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum TaxSection {
        SECTION_194S,   // zero on-chain TDS on Virtual/Synthetic Digital Assets
        SECTION_194LBA, // TDS on Business Trust / REIT / InvIT Distributions
        SECTION_115A,   // TDS on Non-Resident Foreign Investors (Dividends/Interest/DTAA)
        SECTION_194K,   // TDS on Mutual Fund Units / Asset Yields
        SECTION_194IA   // TDS on Transfer of Specified Immovable / Infrastructure Assets
    }

    enum TaxResidencyStatus {
        RESIDENT_INDIVIDUAL,
        RESIDENT_CORPORATE,
        NON_RESIDENT_INDIVIDUAL,
        FOREIGN_PORTFOLIO_INVESTOR,
        DTAA_BENEFICIARY
    }

    enum ChallanStatus {
        UNREMITTED,
        REMITTED_PENDING_CIN,
        FULLY_RECONCILED,
        RECONCILIATION_DISCREPANCY
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct TaxCalculationResult {
        uint256 grossConsiderationPaise; // Gross amount before tax and fees (4 decimals)
        uint256 baseTaxPaise;            // Base statutory withholding amount
        uint256 surchargePaise;          // Statutory surcharge amount
        uint256 cessPaise;               // 4% Health and Education cess
        uint256 totalTaxWithheldPaise;   // Total tax deducted at source
        uint256 platformFeePaise;        // Canonical 0.00% (Zero Fee) platform fee
        uint256 netDisbursementPaise;    // Net payable to recipient
        uint32 effectiveRateBps;         // Effective withholding rate in basis points
    }

    struct TaxDeductionRecord {
        bytes32 deductionId;             // Unique cryptographic deduction hash
        bytes32 tanHash;                 // Deductor TAN salted hash
        bytes32 deducteePanHash;         // Deductee PAN salted hash
        address deducteeWallet;          // Deductee on-chain settlement wallet
        TaxSection section;              // Statutory Income Tax section
        uint256 grossAmountPaise;        // Gross consideration in sub-paise
        uint256 taxWithheldPaise;        // Total tax deducted
        uint64 transactionTimestamp;     // Block timestamp of transaction
        uint64 quarterIndex;             // Financial quarter (1: Q1, 2: Q2, 3: Q3, 4: Q4)
        uint16 financialYear;            // Assessment financial year (e.g. 2026)
        bool remittedToGovt;             // True if remitted via Challan ITNS-281
        bytes32 challanRef;              // Associated challan reference hash
    }

    struct DTAAReliefProof {
        bytes32 countryCodeHash;         // Country of tax residency hash
        uint32 treatyRateBps;            // Treaty tax withholding rate in bps
        uint64 trcExpiryTimestamp;       // Tax Residency Certificate expiry
        bytes32 form10FDocumentHash;     // Salted hash of Form 10F declaration
        bytes32 trcCertificateHash;      // Salted hash of TRC document
        bytes authoritySignature;        // Signature of authorized compliance verifier
    }

    struct QuarterlyBatchRoot {
        bytes32 tanHash;                 // Deductor TAN hash
        uint8 quarterIndex;              // Financial quarter index (1 to 4)
        uint16 financialYear;            // Financial year (e.g. 2026)
        bytes32 batchMerkleRoot;         // Merkle root of all quarterly deduction records
        uint256 totalGrossAmountPaise;   // Aggregated quarterly turnover
        uint256 totalTaxDepositedPaise;  // Aggregated tax deposited
        bytes32 tracesAckNumberHash;     // TRACES acknowledgment receipt hash
        uint64 submittedTimestamp;       // Filing timestamp
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event TaxWithheldAtomic(
        bytes32 indexed deductionId,
        bytes32 indexed tanHash,
        bytes32 indexed deducteePanHash,
        address deducteeWallet,
        TaxSection section,
        uint256 grossAmountPaise,
        uint256 taxWithheldPaise,
        uint256 netDisbursementPaise
    );

    event DTAAReliefApplied(
        bytes32 indexed deducteePanHash,
        bytes32 indexed countryCodeHash,
        uint32 standardRateBps,
        uint32 treatyRateBps,
        uint256 taxSavedPaise
    );

    event QuarterlyBatchRootSubmitted(
        bytes32 indexed tanHash,
        uint8 indexed quarterIndex,
        uint16 indexed financialYear,
        bytes32 batchMerkleRoot,
        bytes32 tracesAckNumberHash
    );

    event PlatformFeeDistributed(
        bytes32 indexed transactionRef,
        uint256 totalFeePaise,
        uint256 treasurySharePaise,
        uint256 coreSgfSharePaise,
        uint256 ipfSharePaise
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error InvalidTaxRate(TaxSection section, uint32 rateBps);
    error InvalidDTAAProof(bytes32 deducteePanHash, string reason);
    error TRCExpired(uint64 expiryTimestamp, uint64 currentTimestamp);
    error DeductionAlreadyExists(bytes32 deductionId);
    error DeductionNotFound(bytes32 deductionId);
    error ZeroGrossConsideration();
    error UnauthorizedTaxSettler(address caller);
    error UnauthorizedComplianceAdmin(address caller);
    error MathOverflowOrUnderflow();

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function executeAtomicTaxWithholding(
        bytes32 tanHash,
        bytes32 deducteePanHash,
        address deducteeWallet,
        TaxSection section,
        TaxResidencyStatus residencyStatus,
        uint256 grossConsiderationPaise,
        bytes32 transactionRef,
        bytes calldata dtaaProofBytes
    ) external returns (TaxCalculationResult memory result, bytes32 deductionId);

    function submitQuarterlyReturnBatch(
        bytes32 tanHash,
        uint8 quarterIndex,
        uint16 financialYear,
        bytes32 batchMerkleRoot,
        uint256 totalGrossPaise,
        uint256 totalTaxPaise,
        bytes32 tracesAckHash
    ) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function calculateEstimatedTax(
        TaxSection section,
        TaxResidencyStatus residencyStatus,
        uint256 grossConsiderationPaise,
        bytes32 deducteePanHash,
        bytes calldata dtaaProofBytes
    ) external view returns (TaxCalculationResult memory result);

    function getDeduction(bytes32 deductionId) external view returns (TaxDeductionRecord memory);
    function getQuarterlyBatchRoot(bytes32 tanHash, uint16 financialYear, uint8 quarterIndex) external view returns (QuarterlyBatchRoot memory);
    function withholdingVault() external view returns (address);
    function certificateSBT() external view returns (address);
    function getPlatformFeeVaults() external view returns (address treasury, address coreSgf, address ipf);
}
```

### 2. e-TDS Withholding Vault Interface (`IETDSWithholdingVault.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IAutomatedTaxLedger} from "./IAutomatedTaxLedger.sol";

/**
 * @title IETDSWithholdingVault
 * @notice Interface for segregated TAN-based statutory tax escrow sub-vaults and government remittances.
 * @dev Manages Challan ITNS-281 remittances and dual-reconciliation against NSDL Protean / TRACES CIN records.
 */
interface IETDSWithholdingVault {
    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct ChallanRemittanceRecord {
        bytes32 challanRef;              // Unique internal challan reference hash
        bytes32 tanHash;                 // Salted Deductor TAN hash
        IAutomatedTaxLedger.TaxSection section; // Tax section
        uint256 remittanceAmountPaise;   // Remitted amount in sub-paise
        uint64 remittanceTimestamp;      // Block timestamp of remittance
        IAutomatedTaxLedger.ChallanStatus status; // Reconciliation status
        bytes32 bsrCodeHash;             // Bank Branch BSR code hash
        uint64 depositDate;              // Challan deposit date (YYYYMMDD)
        uint32 challanSequenceNumber;    // Challan serial sequence number
        bytes32 cinHash;                 // Challan Identification Number (CIN) hash
        bytes32 nodalBankTxRef;          // RBI/SBI nodal bank transfer UTR hash
    }

    struct TANVaultBalance {
        bytes32 tanHash;                 // Salted Deductor TAN hash
        uint256 totalWithheldPaise;      // Total cumulative tax withheld
        uint256 totalRemittedPaise;      // Total cumulative tax remitted to government
        uint256 currentEscrowBalancePaise;// Current pending unremitted escrow balance
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event TaxDepositedToVault(
        bytes32 indexed tanHash,
        IAutomatedTaxLedger.TaxSection indexed section,
        bytes32 indexed deductionId,
        uint256 amountPaise
    );

    event TaxRemittedToGovernment(
        bytes32 indexed challanRef,
        bytes32 indexed tanHash,
        IAutomatedTaxLedger.TaxSection indexed section,
        uint256 amountPaise,
        bytes32 nodalBankTxRef
    );

    event ChallanReconciledWithCIN(
        bytes32 indexed challanRef,
        bytes32 indexed tanHash,
        bytes32 indexed cinHash,
        bytes32 bsrCodeHash,
        uint64 depositDate,
        uint32 challanSequenceNumber
    );

    event ChallanDiscrepancyFlagged(
        bytes32 indexed challanRef,
        bytes32 indexed tanHash,
        string discrepancyReason
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error InsufficientVaultBalance(bytes32 tanHash, uint256 requested, uint256 available);
    error ChallanAlreadyExists(bytes32 challanRef);
    error ChallanNotFound(bytes32 challanRef);
    error InvalidCINProof(bytes32 challanRef, string reason);
    error ChallanAlreadyReconciled(bytes32 challanRef);
    error UnauthorizedVaultCaller(address caller);
    error ZeroRemittanceAmount();

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function depositWithheldTax(
        bytes32 tanHash,
        IAutomatedTaxLedger.TaxSection section,
        bytes32 deductionId,
        uint256 amountPaise
    ) external;

    function remitTaxToGovernment(
        bytes32 tanHash,
        IAutomatedTaxLedger.TaxSection section,
        uint256 amountPaise,
        bytes32 nodalBankTxRef
    ) external returns (bytes32 challanRef);

    function reconcileChallanWithCIN(
        bytes32 challanRef,
        bytes32 bsrCodeHash,
        uint64 depositDate,
        uint32 challanSequenceNumber,
        bytes32 cinHash,
        bytes calldata tracesAttestationSignature
    ) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function getTANVaultBalance(bytes32 tanHash, IAutomatedTaxLedger.TaxSection section) external view returns (TANVaultBalance memory);
    function getChallanRecord(bytes32 challanRef) external view returns (ChallanRemittanceRecord memory);
    function totalSystemEscrowBalance() external view returns (uint256);
    function authorizedLedger() external view returns (address);
    function authorizedComplianceNode() external view returns (address);
}
```

### 3. Digital Tax Certificate Soulbound Token Interface (`IDigitalTaxCertificateSBT.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IAutomatedTaxLedger} from "./IAutomatedTaxLedger.sol";

/**
 * @title IDigitalTaxCertificateSBT
 * @notice Interface for non-transferable Soulbound Form 16A and Form 26AS digital tax deduction certificates.
 * @dev Enforces strict non-transferability (SBT) and cryptographic linkage to verified TRACES / NSDL challans.
 */
interface IDigitalTaxCertificateSBT {
    // -------------------------------------------------------------------------
    // Enums
    // -------------------------------------------------------------------------

    enum CertificateType {
        FORM_16A, // Quarterly certificate for non-salary tax deduction at source (TDS)
        FORM_26AS // Annual Tax Credit Statement summary token
    }

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    struct TaxCertificateMetadata {
        uint256 tokenId;                 // Unique Soulbound Token identifier
        CertificateType certificateType; // Form 16A or Form 26AS
        bytes32 tanHash;                 // Deductor TAN salted hash
        bytes32 deducteePanHash;         // Deductee PAN salted hash
        address deducteeWallet;          // Bound recipient wallet address
        IAutomatedTaxLedger.TaxSection section; // Governing statutory section
        uint256 totalGrossAmountPaise;   // Total consideration covered
        uint256 totalTaxDepositedPaise;  // Total tax deposited under certificate
        uint8 quarterIndex;              // Financial quarter index (1 to 4)
        uint16 financialYear;            // Assessment financial year
        bytes32 cinHash;                 // Associated CIN proof hash
        bytes32 tracesAckNumberHash;     // TRACES verification reference hash
        uint64 issuanceTimestamp;        // Block timestamp of certificate issuance
        bytes32 merkleInclusionProofHash;// Inclusion proof hash in official quarterly filing
    }

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event TaxCertificateMinted(
        uint256 indexed tokenId,
        CertificateType indexed certificateType,
        bytes32 indexed deducteePanHash,
        address deducteeWallet,
        bytes32 tanHash,
        uint256 totalTaxDepositedPaise,
        uint8 quarterIndex,
        uint16 financialYear
    );

    event TaxCertificateRevoked(
        uint256 indexed tokenId,
        bytes32 indexed deducteePanHash,
        string revocationReason
    );

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error SoulboundTokenNonTransferable();
    error CertificateAlreadyMinted(bytes32 certificateUniqueHash);
    error CertificateNotFound(uint256 tokenId);
    error UnauthorizedCertificateIssuer(address caller);
    error InvalidCertificateData();

    // -------------------------------------------------------------------------
    // State-Modifying Functions
    // -------------------------------------------------------------------------

    function mintTaxCertificate(
        address recipientDeductee,
        CertificateType certificateType,
        bytes32 tanHash,
        bytes32 deducteePanHash,
        IAutomatedTaxLedger.TaxSection section,
        uint256 totalGrossPaise,
        uint256 totalTaxPaise,
        uint8 quarterIndex,
        uint16 financialYear,
        bytes32 cinHash,
        bytes32 tracesAckHash,
        bytes32 merkleInclusionProofHash
    ) external returns (uint256 tokenId);

    function revokeCertificate(uint256 tokenId, string calldata reason) external;

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function getCertificate(uint256 tokenId) external view returns (TaxCertificateMetadata memory);
    function verifyCertificateAuthenticity(
        uint256 tokenId,
        bytes32 deducteePanHash,
        bytes32 tanHash,
        bytes32 expectedCinHash
    ) external view returns (bool isValid);
    function getTokensByDeductee(bytes32 deducteePanHash) external view returns (uint256[] memory);
}
```

## Security & Compliance Notes
- **Zero On-Chain PII Compliance:** In compliance with the Digital Personal Data Protection Act (DPDPA 2023) and SEBI cybersecurity frameworks, sensitive fiscal identifiers (PAN, TAN, Aadhaar, legal names) are never recorded in plaintext on the Hyperledger Besu blockchain. Every identifier is committed via high-entropy salted hashes (`keccak256(abi.encodePacked(rawPAN, salt))`).
- **Fixed Platform Fee Invariant:** In strict accordance with the Growww Core Fee Model (Prompt 006), the platform fee is assessed at exactly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) on trade notional turnover with an automated 0.00% fee at launch (governed by FeeController.sol) revenue split. Capital gains are computed strictly for off-chain user tax compliance (Section 111A/112A). Zero custody, holding, or certificate issuance fees are charged.
- **Atomic Escrow & Fund Segregation:** Withheld tax funds are locked instantaneously in dedicated `ETDSWithholdingVault.sol` sub-vaults. Smart contract logic strictly prevents combining statutory tax escrow balances with exchange operational funds, clearing margins, or treasury assets. Funds can exit the vault only upon cryptographically verified Challan ITNS-281 remittance.
- **Double Tax Avoidance Agreement (DTAA) Safeguards:** For foreign investors (Section 115A), concessional withholding rates require valid, unexpired Tax Residency Certificates (TRC) and Form 10F Merkle proofs signed by authorized compliance officers. The contract automatically falls back to maximum statutory domestic rates if the TRC is expired or if the cryptographic signature fails.
- **Soulbound Immutability (SBT):** Form 16A and Form 26AS digital certificates are implemented as non-transferable Soulbound Tokens. All standard transfer and approval hooks (`transferFrom`, `safeTransferFrom`, `approve`, `setApprovalForAll`) revert unconditionally, preventing unauthorized tax credit transfers, fraudulent secondary trading, or certificate rehypothecation.
- **Dual-Reconciliation & CIN Verification:** Remitted taxes are reconciled against the NSDL Protean / TRACES portal using Challan Identification Numbers (CIN) consisting of the 7-digit BSR code, deposit date (YYYYMMDD), and 5-digit challan sequence number. A discrepancy flag is raised immediately if remitted sums fail to match CIN cryptographic attestations.
- **Multi-Signature Governance:** Privileged actions (configuring statutory tax rates, updating authorized TRACES relayer nodes, whitelisting deductor TAN sub-vaults, and pausing contracts) require 3-of-5 multi-signature authorization from `MultiSigGovernance.sol` (Prompt 307) backed by FIPS 140-2 Level 3 HSM keys (Prompt 311).

## Acceptance Criteria
- [ ] `IAutomatedTaxLedger.sol`, `IETDSWithholdingVault.sol`, and `IDigitalTaxCertificateSBT.sol` compile under Solidity 0.8.24 with zero compiler warnings.
- [ ] Sub-paise precision arithmetic ($10^{-4}$ INR / 0.01 paise per unit) executes with zero rounding leakage across Section 194S, Section 194LBA, and Section 115A tax withholdings.
- [ ] Section 194S accurately enforces zero on-chain TDS on DvP settlements, routing trade events to off-chain tax service and routes tax atomically to `ETDSWithholdingVault.sol`.
- [ ] Section 194LBA accurately differentiates REIT/InvIT distribution components and applies 10% (resident) or 5% (non-resident) withholding.
- [ ] Section 115A accurately applies concessional DTAA treaty withholding rates upon cryptographic TRC and Form 10F proof validation.
- [ ] Withheld tax balances remain strictly segregated in multi-tenant TAN sub-vaults and can only be withdrawn via Challan ITNS-281 remittance.
- [ ] Dual-reconciliation mechanism verifies Challan Identification Numbers (CIN), BSR codes, and TRACES attestation signatures.
- [ ] Digital Form 16A and Form 26AS certificates mint as Soulbound Tokens (SBTs) that revert on any transfer attempt.
- [ ] Quarterly e-TDS batch Merkle roots (Form 26Q / 27Q) are verified and stored for individual deductee inclusion verification.
- [ ] Platform fee computation mathematically enforces exact 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) deduction on gross turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol).
- [ ] Invariant fuzz tests in Foundry execute 10,000 runs confirming zero balance discrepancies, zero negative balances, and zero unauthorized withdrawals.
- [ ] Slither and Mythril static analysis suites pass with zero high or medium severity vulnerabilities.
- [ ] Zero em dashes and zero en dashes present across entire specification documentation.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `303` (Token Issuance Smart Contract - ERC-3643), Prompt `305` (Transfer Compliance Hooks), Prompt `306` (Settlement DvP Smart Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `208` (Trade Settlement Service), Prompt `210` (Fee & Realized PnL Engine), Prompt `222` (Corporate Actions Service), Prompt `223` (Tax Reporting & Statement Service), Prompt `334` (RBI CBDC eINR Bridge Contract).
- **Subsequent Prompts Enabled:** Prompt `308` (Proof of Reserve Registry), Prompt `309` (Event Indexing Service), Prompt `335` (ZK Light Client Cross-Chain Verifier Contract), Prompt `509` (Flutter Order Placement Flow).
