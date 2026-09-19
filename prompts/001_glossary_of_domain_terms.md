# 001 - Glossary of Domain Terms & Ubiquitous Language

## Purpose
To prevent semantic divergence and communication breakdown across multi-disciplinary teams (financial engineers, distributed systems developers, compliance officers, and Flutter UI developers), this document establishes the formal Domain-Driven Design (DDD) Ubiquitous Language for Growww.

Standardizing terminology such as Delivery versus Payment (DvP), Demat Holding Pool, Tokenized Equity Unit, Beneficial Ownership, Fixed Platform Fee, Realized Capital Gain Engine, and Whitelist Attestation ensures zero ambiguity across codebase naming conventions, database schemas, gRPC service contracts, and API payloads. Every technical artifact produced across Categories 1 through 9 must adhere strictly to the definitions established here.

## What You Are Building
An exhaustive domain dictionary and ubiquitous language specification (`docs/domain/ubiquitous_language.md`) containing:
- Depository and Equities terms (NSDL, CDSL, Demat, ISIN, Corporate Actions, Fractional Beneficial Interest).
- Permissioned Blockchain terms (Hyperledger Besu, QBFT Consensus, DvP Settlement, Proof-of-Reserve, Merkle Tree Attestation, Identity Commitment).
- Regulatory & Compliance terms (SEBI, RBI, IFSCA, PMLA, DPDP Act 2023, FATF, PEP, CKYC, STR, CTR).
- Financial & Accounting terms (FIFO Cost Basis, Realized P&L, Unrealized P&L, Universal Zero-Fee Model (0.00% fee - No fee at all), Multi-Vault Revenue Split, Hold/Escrow, Tax Lot).
- Banking & Fiat Rail terms (UPI Intent, VPA, Nodal Account, Escrow, RTGS, NEFT, IMPS, Payment Aggregator).
- Standardized Code Symbol Mappings (naming conventions for classes, database fields, proto messages, and smart contracts).

## Scope Boundaries
- **In Scope:**
 - Precise definitions, domain context, mathematical implications, and code-symbol naming standards for all core terms.
 - Cross-domain mapping tables connecting traditional banking/depository constructs to permissioned ledger entities.
 - Forbidden/deprecated terminology list (e.g., banning "crypto token", "gas fee", "synthetic asset", "AUM fee").
- **Out of Scope / Handled Elsewhere:**
 - Full relational database DDL schemas (covered in Prompt 401).
 - Detailed Protobuf message definitions (covered in Prompt 103 and Category 2).
 - Smart contract Solidity code (covered in Category 3).

## Technology to Use
- Primary Format: Markdown specification (`docs/domain/ubiquitous_language.md`) structured with machine-readable YAML frontmatter and dictionary tables.
- Tooling: Custom pre-commit linter script (`scripts/lint_domain_terms.py`) and Vale prose linter configured with custom domain vocabulary rules.
- Justification: Automated prose and codebase linting enforces terminology compliance at PR time, catching domain drift before code is merged into main branches.

## Backend / Infra Touchpoints
- Internal Developer Documentation Portal (searchable glossary).
- Git CI/CD workflow (linter executing on all documentation and code PRs).
- Code generators (proto and OpenAPI codegen templates referencing canonical terms).

## Blockchain Interaction
Defines canonical nomenclature and semantic guarantees for on-chain entities on the Hyperledger Besu QBFT network:
- `DigitalSecurityToken`: Fractional, 1:1 asset-backed ERC-3643 compliant digital equity unit.
- `DvPSettlementContract`: Atomic smart contract executing simultaneous equity token delivery against off-chain fiat settlement confirmation.
- `ComplianceRegistry`: On-chain whitelist storing investor identity commitments and regulatory freeze statuses.
- `ReserveAttestation`: Cryptographic Merkle root attestation of 1:1 physical share custody in NSDL/CDSL depositories.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review domain invariants and operational concepts from Prompt 000.
2. Compile and standardize Depository and Equities terminology from SEBI and NSDL/CDSL regulatory manuals.
3. Compile and standardize Indian Banking rail terminology (RBI Master Directions for UPI, Escrow, and Nodal Accounts).
4. Compile and standardize International IFSC terminology from IFSCA (Capital Market Intermediaries) Regulations.
5. Compile and standardize Permissioned Blockchain terms specific to Hyperledger Besu and QBFT consensus.
6. Compile and standardize Accounting & Tax terms (FIFO tax lot matching for user tax compliance under Section 111A/112A, Fixed 0.00% (Zero Fee) [0.00% fee / 0 bps at launch] Transaction Fee Model mechanics on trade notional turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol)).
7. Create a forbidden terminology table explicitly prohibiting misleading terms (e.g., "gas fee", "crypto coin", "synthetic token", "profit-only fee").
8. Define canonical code-symbol mappings (e.g., mapping "Delivery versus Payment" to `DvPSettlementService` in Go/Python and `SettlementDvP.sol` in Solidity).
9. Construct a searchable Markdown table with columns: Term, Domain, Canonical Code Symbol, Precise Definition, Regulatory Reference, and Disallowed Synonyms.
10. Implement a YAML-based term dictionary (`docs/domain/terms.yaml`) for programmatic consumption by CI linters.
11. Write the Vale / Python linting rule enforcing approved term usage across documentation and code comments.
12. Review with Financial Engineering, Legal Compliance, and Lead Architects to achieve unanimous sign-off.

## Interfaces / Contracts
```yaml
# Terminology Definition Schema (docs/domain/terms.yaml)
terms:
 - id: "TERM-001"
    name: "Delivery versus Payment"
    acronym: "DvP"
    canonical_symbol: "DvPSettlement"
    domain: "Settlement & Ledger"
    definition: "An atomic settlement mechanism ensuring that the transfer of fractional digital equity units occurs if and only if the corresponding fiat payment is confirmed."
    disallowed_synonyms: ["atomic swap", "crypto swap", "token trade"]
    regulatory_reference: "SEBI Settlement Regulations 2018"

 - id: "TERM-002"
    name: "Digital Security Token"
    acronym: "DST"
    canonical_symbol: "DigitalSecurityToken"
    domain: "Asset Tokenization"
    definition: "A fractional, 1:1 asset-backed digital representation of an underlying equity share held in regulated depository custody (NSDL/CDSL)."
    disallowed_synonyms: ["crypto coin", "synthetic token", "equity derivative"]
    regulatory_reference: "SEBI Sandbox Framework 2023"

 - id: "TERM-003"
    name: "Fixed Platform Transaction Fee"
    acronym: "FPTF"
    canonical_symbol: "FixedPlatformTransactionFee"
    domain: "Accounting & Monetization"
    definition: "A 0.00% fee (No fee at all) assessed on trade notional turnover across executed transactions, operating with 0.00% fees at launch (100% net proceeds credited); FIFO capital gains are computed strictly for user tax compliance (Section 111A/112A)."
    disallowed_synonyms: ["management fee", "AUM fee", "holding fee", "gas fee", "profit-only fee"]
    regulatory_reference: "Growww Fee Model Mandate"
```

## Security & Compliance Notes
- Terminology must precisely map to definitions in the Prevention of Money Laundering Act (PMLA) 2002, DPDP Act 2023, and SEBI (Stock Brokers and Sub-Brokers) Regulations.
- Misleading marketing or technical language (such as describing tokens as "crypto" rather than "regulated digital securities") violates SEBI advertising codes and must be blocked at CI linting.

## Acceptance Criteria
- [ ] `docs/domain/ubiquitous_language.md` contains over 60 fully defined terms spanning all 5 operational domains.
- [ ] Canonical code-symbol mappings are established for every core entity across Python, Go, Rust, Dart, and Solidity.
- [ ] A forbidden/disallowed synonyms table is documented and integrated into CI prose linters.
- [ ] Machine-readable `terms.yaml` is generated and validated against schema.
- [ ] Sign-off obtained from Lead Software Architect, Financial Controller, and Compliance Officer.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star).
- Parallel Tasks: 002 (Two-Entity Structure), 003 (Regulatory Pathway).
- Downstream Blockers: Blocks domain modeling (111), API design (103), and service boundary definitions (102).
