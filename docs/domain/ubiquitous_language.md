---
title: "Growww Ubiquitous Language Specification & Domain Glossary"
version: "1.0.0-PROD-SPEC"
status: "Approved"
domain_count: 5
term_count: 72
last_updated: "2026-09-19"
governance:
  framework: "Domain-Driven Design (DDD)"
  regulatory_frameworks:
    - "SEBI (Depositories and Participants) Regulations 2018"
    - "SEBI Regulatory Sandbox Framework 2023"
    - "RBI Payment and Settlement Systems Act 2007"
    - "Prevention of Money Laundering Act (PMLA) 2002"
    - "Digital Personal Data Protection Act (DPDP) 2023"
    - "IFSCA (Capital Market Intermediaries) Regulations 2021"
    - "Income Tax Act 1961 (Sections 111A, 112A, 194S)"
approvers:
  lead_software_architect: "Approved (Cryptographic Sign-off)"
  financial_controller: "Approved (Audit Verified)"
  chief_compliance_officer: "Approved (Legal & Statutory Validated)"
---

# Ubiquitous Language & Domain Terms Specification

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Canonical Reference Specification  
**Owner:** Core Financial Engineering, Architecture Governance & Compliance  
**Audience:** Backend Engineers (Go, Python, Rust), Smart Contract Engineers (Solidity), Frontend Engineers (Flutter, Next.js), Risk & Financial Controllers, and Compliance Officers.

---

## 1. Ubiquitous Language Philosophy & DDD Principles

### 1.1 Purpose and Architectural Rationale
In high-consequence financial distributed systems operating at the intersection of capital market regulations, traditional banking rails, and permissioned distributed ledgers, **semantic ambiguity is a primary failure mode**. When financial engineers say "settlement," they may mean T+1 clearing novation; when blockchain developers say "settlement," they may mean block inclusion; when compliance officers say "settlement," they may mean irrevocable fiat funds movement between regulated nodal accounts.

To eliminate semantic divergence, this specification codifies the formal **Domain-Driven Design (DDD) Ubiquitous Language** for Growww. 

```
+--------------------------------------------------------------------------------------------------+
|                                    GROWWW DOMAIN BOUNDARIES                                      |
|                                                                                                  |
|   +--------------------------+   +--------------------------+   +----------------------------+   |
|   |   Depository & Equities  |   |  Permissioned Blockchain |   |   Regulatory & Compliance  |   |
|   | (NSDL, CDSL, Demat, ISIN)|   | (Besu, QBFT, DvP, ERC3643)|  | (SEBI, RBI, PMLA, DPDP Act)|   |
|   +-------------+------------+   +-------------+------------+   +--------------+-------------+   |
|                 \                              |                               /                 |
|                  \                             |                              /                  |
|                   +----------------------------v-----------------------------+                   |
|                   |           Financial, Accounting & Monetization           |                   |
|                   | (FIFO Cost Basis, 0.00% Launch Fee, Multi-Vault Split)  |                   |
|                   +----------------------------+-----------------------------+                   |
|                                                |                                                 |
|                                  +-------------v-------------+                                   |
|                                  |   Banking & Fiat Rails    |                                   |
|                                  | (UPI Intent, Escrow, Nodal|                                   |
|                                  +---------------------------+                                   |
+--------------------------------------------------------------------------------------------------+
```

### 1.2 The 5 Supreme Domain Invariants
Every technical artifact (database table, gRPC proto, REST payload, message broker event, smart contract, and UI model) must align strictly with the following non-negotiable invariants:

1. **Strict 1:1 Physical Custody Backing:** Every Digital Security Token (`DigitalSecurityToken`) issued on-chain maps strictly 1:1 to physical underlying equity shares held in designated SEBI-registered depository accounts (NSDL/CDSL). Unbacked, synthetic, or leveraged fractionalization is strictly prohibited.
2. **Zero PII on Distributed Ledger:** No Personally Identifiable Information (PAN, Aadhaar, full legal name, phone number, physical address, email) may ever enter on-chain transaction payloads, smart contract state, transaction receipts, or execution logs. Identity is represented on-chain solely via salted, zero-knowledge cryptographic `IdentityCommitment` (`bytes32`).
3. **Delivery versus Payment (DvP) Atomicity:** Token ownership transfers execute if and only if fiat payment confirmation has been validated and locked. Off-chain reversible banking reservations occur before irreversible on-chain state mutations.
4. **Universal Zero-Fee Model (0.00% fee - No fee at all):** Platform charges exactly 0.00% (0 bps) fee at launch on trade notional turnover across executed transactions (100% net proceeds credited; future adjustments governed strictly by `FeeController.sol` with a 48-hour timelock and a 50 bps hard ceiling). Holding fees and AUM-based fees are permanently 0.00%. Realized capital gains are computed strictly for user statutory tax compliance (Section 111A/112A).
5. **Deterministic Finality & Zero-Gas Consortium Ledger:** The execution network is a private, permissioned Hyperledger Besu consortium running Quorum Byzantine Fault Tolerance (QBFT). Gas price is set to 0. Probabilistic forks, miner extractable value (MEV), and speculative crypto gas dynamics are eliminated.

---

## 2. Core Domain Terminology Dictionaries

The dictionary is categorized into 5 distinct operational bounded contexts. Each entry establishes the authoritative definition, canonical code symbol, mathematical/accounting implication, regulatory reference, and strictly disallowed synonyms.

### 2.1 Domain 1: Depository & Equities

```
+-----------------------------------------------------------------------------------------------------+
| TERM-001 through TERM-015: Depository & Equities Domain Summary Table                               |
+----------+-----------------------------------+---------+----------------------------+---------------+
| Term ID  | Term Name                         | Acronym | Canonical Symbol           | Regulatory    |
+----------+-----------------------------------+---------+----------------------------+---------------+
| TERM-001 | Delivery versus Payment           | DvP     | DvPSettlement              | SEBI Sett. 18 |
| TERM-002 | Digital Security Token            | DST     | DigitalSecurityToken       | SEBI Sandbox  |
| TERM-003 | National Securities Depository    | NSDL    | NsdlDepositoryGateway      | Depositories  |
| TERM-004 | Central Depository Services Ltd   | CDSL    | CdslDepositoryGateway      | Depositories  |
| TERM-005 | Dematerialized Account            | Demat   | DematAccount               | SEBI DP Reg40 |
| TERM-006 | Int'l Securities Identification   | ISIN    | SecurityIsin               | ISO 6166      |
| TERM-007 | Depository Participant            | DP      | DepositoryParticipant      | SEBI DP Ch.III|
| TERM-008 | Beneficial Ownership              | BO      | BeneficialOwnership        | Comp. Act S89 |
| TERM-009 | Fractional Beneficial Interest    | FBI     | FractionalBeneficialInterest| SEBI Sandbox |
| TERM-010 | Demat Holding Pool                | DHP     | DematHoldingPool           | SEBI Circular |
| TERM-011 | Corporate Action Processing       | CAP     | CorporateActionProcessor   | LODR Reg 42   |
| TERM-012 | Delivery Instruction Slip         | DIS     | DeliveryInstructionSlip    | SEBI e-DIS    |
| TERM-013 | Physical Custody Account          | PCA     | PhysicalCustodyAccount     | SEBI Custodian|
| TERM-014 | Dematerialization Request Form    | DRF     | DematerializationRequest   | Depositories  |
| TERM-015 | Clearing Corporation              | CC      | ClearingCorporation        | SEBI CC Reg18 |
+----------+-----------------------------------+---------+----------------------------+---------------+
```

#### TERM-001: Delivery versus Payment (DvP)
- **Domain:** Depository & Equities / Settlement
- **Acronym:** `DvP`
- **Canonical Code Symbol:** `DvPSettlement`
- **Precise Definition:** An atomic settlement mechanism ensuring that the transfer of fractional digital equity units occurs if and only if the corresponding fiat payment is confirmed.
- **Mathematical / Ledger Implication:**
  $$\Delta \text{Balance}_{\text{Buyer}}(\text{DST}) = +Q \iff \Delta \text{Balance}_{\text{Buyer}}(\text{Escrow Fiat}) = -(P \times Q)$$
- **Regulatory Reference:** SEBI Settlement Regulations 2018; CPSS-IOSCO Principles for Financial Market Infrastructures (Principle 12: Delivery versus Payment).
- **Disallowed Synonyms:** *atomic swap*, *crypto swap*, *token trade*, *p2p swap*.

#### TERM-002: Digital Security Token (DST)
- **Domain:** Depository & Equities / Asset Tokenization
- **Acronym:** `DST`
- **Canonical Code Symbol:** `DigitalSecurityToken`
- **Precise Definition:** A fractional, 1:1 asset-backed digital representation of an underlying equity share held in regulated depository custody (NSDL/CDSL), governed by the ERC-3643 permissioned standard.
- **Mathematical / Ledger Implication:**
  $$\sum_{i=1}^{M} \text{Balance}_i(\text{DST}) \equiv \text{PhysicalSharesHeld}_{\text{Custody}}(\text{ISIN}) \times 10^{18}$$
- **Regulatory Reference:** SEBI Regulatory Sandbox Framework 2023; SEBI (Issue of Capital and Disclosure Requirements) Regulations 2018.
- **Disallowed Synonyms:** *crypto coin*, *synthetic token*, *equity derivative*, *crypto asset*.

#### TERM-003: National Securities Depository Limited (NSDL)
- **Domain:** Depository & Equities
- **Acronym:** `NSDL`
- **Canonical Code Symbol:** `NsdlDepositoryGateway`
- **Precise Definition:** Primary Indian central securities depository facilitating electronic book-entry transfer, dematerialization, and settlement of equity shares.
- **Regulatory Reference:** Depositories Act 1996; SEBI (Depositories and Participants) Regulations 2018.
- **Disallowed Synonyms:** *blockchain vault*, *cold storage depository*, *web3 custodian*.

#### TERM-004: Central Depository Services Limited (CDSL)
- **Domain:** Depository & Equities
- **Acronym:** `CDSL`
- **Canonical Code Symbol:** `CdslDepositoryGateway`
- **Precise Definition:** Secondary Indian central securities depository holding dematerialized securities and facilitating electronic trade clearing and settlements.
- **Regulatory Reference:** Depositories Act 1996; SEBI (Depositories and Participants) Regulations 2018.
- **Disallowed Synonyms:** *on-chain vault*, *crypto custodian*, *decentralized depository*.

#### TERM-005: Dematerialized Account (Demat)
- **Domain:** Depository & Equities
- **Acronym:** `Demat`
- **Canonical Code Symbol:** `DematAccount`
- **Precise Definition:** An electronic book-entry account maintained with an NSDL or CDSL Depository Participant storing physical securities in electronic form identified by a 16-digit Beneficiary Owner Identification (BOID).
- **Regulatory Reference:** SEBI (Depositories and Participants) Regulations 2018, Regulation 40.
- **Disallowed Synonyms:** *crypto wallet*, *token wallet*, *seed phrase account*, *web3 address*.

#### TERM-006: International Securities Identification Number (ISIN)
- **Domain:** Depository & Equities
- **Acronym:** `ISIN`
- **Canonical Code Symbol:** `SecurityIsin`
- **Precise Definition:** A 12-character alphanumeric code established under ISO 6166 uniquely identifying a specific equity or debt security (e.g., `INE002A01018` for Reliance Industries).
- **Structure:** `[Country Code: 2 letters] + [Issuer & Security: 9 alphanumeric] + [Luhn Checksum: 1 digit]`.
- **Regulatory Reference:** ISO 6166 Standard; SEBI Circular CIR/MRD/DP/18/2015.
- **Disallowed Synonyms:** *token address*, *coin contract*, *asset hash*.

#### TERM-007: Depository Participant (DP)
- **Domain:** Depository & Equities
- **Acronym:** `DP`
- **Canonical Code Symbol:** `DepositoryParticipant`
- **Precise Definition:** A registered financial intermediary acting as an authorized agent of the depository (NSDL/CDSL) to provide depository services to investors.
- **Regulatory Reference:** SEBI (Depositories and Participants) Regulations 2018, Chapter III.
- **Disallowed Synonyms:** *crypto exchange node*, *web3 relayer*, *staking validator*.

#### TERM-008: Beneficial Ownership (BO)
- **Domain:** Depository & Equities
- **Acronym:** `BO`
- **Canonical Code Symbol:** `BeneficialOwnership`
- **Precise Definition:** The true legal and equitable right to the economic benefits of equity securities (dividends, voting rights, corporate actions, and capital appreciation) held in depository custody.
- **Regulatory Reference:** Companies Act 2013, Section 89 & 90; PMLA Rules Rule 9(1A).
- **Disallowed Synonyms:** *token holder rights*, *NFT ownership*, *dao governance right*.

#### TERM-009: Fractional Beneficial Interest (FBI)
- **Domain:** Depository & Equities
- **Acronym:** `FBI`
- **Canonical Code Symbol:** `FractionalBeneficialInterest`
- **Precise Definition:** A strictly computed, proportional entitlement to a whole physical share held in the Demat Holding Pool, represented down to 18 decimal places of mathematical precision.
- **Mathematical Precision:**
  $$\text{Units} = \left\lfloor \text{WholeShares} \times 10^{18} \right\rfloor$$
- **Regulatory Reference:** SEBI Regulatory Sandbox Framework 2023; Indian Trust Act 1882.
- **Disallowed Synonyms:** *synthetic fraction*, *derivative slice*, *crypto fractionalization*.

#### TERM-010: Demat Holding Pool (DHP)
- **Domain:** Depository & Equities
- **Acronym:** `DHP`
- **Canonical Code Symbol:** `DematHoldingPool`
- **Precise Definition:** A dedicated, ring-fenced institutional depository account maintained by the licensed broker/custodian holding whole equity shares to back issued fractional digital units 1:1.
- **Regulatory Reference:** SEBI Master Circular for Depositories (Client Asset Protection Guidelines).
- **Disallowed Synonyms:** *liquidity pool*, *amm pool*, *token treasury*, *crypto pool*.

#### TERM-011: Corporate Action Processing (CAP)
- **Domain:** Depository & Equities
- **Acronym:** `CAP`
- **Canonical Code Symbol:** `CorporateActionProcessor`
- **Precise Definition:** The automated lifecycle processing of issuer events including cash dividends, stock splits, bonus shares, rights issues, and mergers across fractional beneficiaries.
- **Mathematical Formula:**
  $$\text{Dividend}_{\text{Beneficiary}} = \text{DividendPerShare} \times \left(\frac{\text{Units}_{\text{Beneficiary}}}{10^{18}}\right)$$
- **Regulatory Reference:** SEBI (Listing Obligations and Disclosure Requirements) Regulations 2015, Regulation 42.
- **Disallowed Synonyms:** *token airdrop*, *staking reward payout*, *yield emission*.

#### TERM-012: Delivery Instruction Slip (DIS)
- **Domain:** Depository & Equities
- **Acronym:** `DIS`
- **Canonical Code Symbol:** `DeliveryInstructionSlip`
- **Precise Definition:** A formal statutory instruction mandate (electronic e-DIS authenticated via OTP/PIN or physical slip) authorizing the debit and transfer of securities from a Demat account.
- **Regulatory Reference:** SEBI Master Circular for Depositories (Section on e-DIS).
- **Disallowed Synonyms:** *gas transfer*, *send transaction*, *private key authorization*.

#### TERM-013: Physical Custody Account (PCA)
- **Domain:** Depository & Equities
- **Acronym:** `PCA`
- **Canonical Code Symbol:** `PhysicalCustodyAccount`
- **Precise Definition:** A segregated custodian depository repository where underlying whole shares are held unencumbered under registered custodian supervision.
- **Regulatory Reference:** SEBI (Custodian of Securities) Regulations 1996.
- **Disallowed Synonyms:** *cold wallet*, *multisig treasury*, *custody smart contract*.

#### TERM-014: Dematerialization Request Form (DRF)
- **Domain:** Depository & Equities
- **Acronym:** `DRF`
- **Canonical Code Symbol:** `DematerializationRequest`
- **Precise Definition:** Statutory documentation submitted to transform physical share certificates into electronic dematerialized credits within NSDL/CDSL.
- **Regulatory Reference:** Depositories Act 1996, Section 6.
- **Disallowed Synonyms:** *token wrapping*, *asset bridging*, *crypto onboarding*.

#### TERM-015: Clearing Corporation (CC)
- **Domain:** Depository & Equities
- **Acronym:** `CC`
- **Canonical Code Symbol:** `ClearingCorporation`
- **Precise Definition:** Regulated central counterparty clearing house (e.g., NSCCL, ICCL, CCIL) guaranteeing settlement and managing counterparty risk for market transactions.
- **Regulatory Reference:** SEBI (Clearing Corporations) Regulations 2018.
- **Disallowed Synonyms:** *dex router*, *amm clearinghouse*, *swap bridge*.

---

### 2.2 Domain 2: Permissioned Blockchain & Distributed Ledger

```
+-----------------------------------------------------------------------------------------------------+
| TERM-016 through TERM-029: Blockchain & Distributed Ledger Domain Summary Table                     |
+----------+-----------------------------------+---------+----------------------------+---------------+
| Term ID  | Term Name                         | Acronym | Canonical Symbol           | Invariant / Ref|
+----------+-----------------------------------+---------+----------------------------+---------------+
| TERM-016 | Hyperledger Besu                  | BESU    | BesuLedgerNode             | Zero-Gas EVM  |
| TERM-017 | QBFT Consensus                    | QBFT    | QbftConsensusEngine        | N >= 3F + 1   |
| TERM-018 | Proof of Reserve                  | PoR     | ProofOfReserve             | Ratio == 1.0  |
| TERM-019 | Merkle Tree Attestation           | MTA     | MerkleTreeAttestation      | FIPS 180-4    |
| TERM-020 | Identity Commitment               | IDC     | IdentityCommitment         | Zero PII Hash |
| TERM-021 | Compliance Registry               | CR      | ComplianceRegistry         | ERC-3643 White|
| TERM-022 | Trusted Validator Node            | TVN     | TrustedValidatorNode       | Consortium Auth|
| TERM-023 | Multi-Signature Authorization     | M-of-N  | MultiSigAuthorization      | FIPS 140-2 L3 |
| TERM-024 | Zero PII Ledger Invariant         | ZPII    | ZeroPiiInvariant           | DPDP Act S8   |
| TERM-025 | Deterministic Finality            | DFIN    | DeterministicFinality      | P(Reorg) == 0 |
| TERM-026 | Security Token Clawback           | STC     | TokenClawbackController    | Disgorgement  |
| TERM-027 | Custody Merkle Root               | CMR     | CustodyMerkleRoot          | Audit Root    |
| TERM-028 | Token Partition                   | TPAR    | TokenPartition             | ERC-1410      |
| TERM-029 | On-Chain Freeze Controller        | OFC     | OnChainFreezeController    | Lawful Freeze |
+----------+-----------------------------------+---------+----------------------------+---------------+
```

#### TERM-016: Hyperledger Besu
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `BESU`
- **Canonical Code Symbol:** `BesuLedgerNode`
- **Precise Definition:** Enterprise-grade Ethereum Virtual Machine (EVM)-compatible client running a private, permissioned consortium network with zero native gas volatility.
- **Invariant:** `baseFeePerGas == 0`; transactions are validated based on validator signature permissioning, not gas bids.
- **Regulatory Reference:** Enterprise Ethereum Alliance Standards; MeitY National Blockchain Strategy.
- **Disallowed Synonyms:** *public blockchain*, *mainnet ethereum*, *unpermissioned chain*, *crypto network*.

#### TERM-017: QBFT Consensus
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `QBFT`
- **Canonical Code Symbol:** `QbftConsensusEngine`
- **Precise Definition:** Quorum Byzantine Fault Tolerance consensus algorithm providing deterministic, non-forking finality with sub-2-second block intervals.
- **Mathematical Invariant:**
  $$N \ge 3F + 1, \quad \text{Quorum} = 2F + 1$$
- **Regulatory Reference:** ISO/IEC 23257 Blockchain Reference Architecture.
- **Disallowed Synonyms:** *proof of work*, *proof of stake*, *mining*, *validator staking reward*.

#### TERM-018: Proof of Reserve (PoR)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `PoR`
- **Canonical Code Symbol:** `ProofOfReserve`
- **Precise Definition:** Cryptographic attestation verifying that the sum of all issued Digital Security Tokens equals exactly the shares held in depository custody.
- **Mathematical Formula:**
  $$\text{ReserveRatio} = \frac{\text{DepositoryHolding}(\text{ISIN})}{\sum_{k} \text{Balance}_k / 10^{18}} \equiv 1.000000000000000000$$
- **Regulatory Reference:** SEBI Master Circular on Cyber Security & Systems Audit.
- **Disallowed Synonyms:** *oracle feed*, *rebase oracle*, *crypto reserve*.

#### TERM-019: Merkle Tree Attestation (MTA)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `MTA`
- **Canonical Code Symbol:** `MerkleTreeAttestation`
- **Precise Definition:** A cryptographic binary tree structure providing succinct, tamper-evident proof of balance inclusion without exposing investor identity or position details.
- **Complexity:** Verification occurs in $O(\log N)$ cryptographic hash operations.
- **Regulatory Reference:** National Blockchain Strategy; FIPS 180-4 Secure Hash Standard.
- **Disallowed Synonyms:** *merkle drop*, *airdrop tree*, *token tree*.

#### TERM-020: Identity Commitment (IDC)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `IDC`
- **Canonical Code Symbol:** `IdentityCommitment`
- **Precise Definition:** A salted cryptographic hash (`bytes32`) binding a user's verified KYC status to an on-chain address while maintaining zero on-chain PII.
- **Formula:**
  $$\text{IDC} = \text{PoseidonHash}(\text{CKYC\_ID}, \text{Salt}_{\text{PAN}}, \text{SecretEntropy})$$
- **Regulatory Reference:** DPDP Act 2023, Section 4; SEBI KYC Master Directions.
- **Disallowed Synonyms:** *wallet address*, *public key identity*, *crypto signature*.

#### TERM-021: Compliance Registry (CR)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `CR`
- **Canonical Code Symbol:** `ComplianceRegistry`
- **Precise Definition:** An on-chain ERC-3643 compliant contract enforcing investor whitelisting, jurisdiction restrictions, and regulatory freeze flags prior to any transfer execution.
- **Invariant:**
  $$\text{canTransfer}(\text{from}, \text{to}, Q) = \text{isWhitelisted}(\text{from}) \land \text{isWhitelisted}(\text{to}) \land \neg \text{isFrozen}(\text{from}) \land \neg \text{isFrozen}(\text{to})$$
- **Regulatory Reference:** ERC-3643 Permissioned Token Standard; SEBI Intermediary Regulations.
- **Disallowed Synonyms:** *blacklist contract*, *tether freeze*, *ban list*.

#### TERM-022: Trusted Validator Node (TVN)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `TVN`
- **Canonical Code Symbol:** `TrustedValidatorNode`
- **Precise Definition:** A designated, permissioned institutional server authorized in the Besu genesis block to propose and validate blocks under QBFT consensus.
- **Regulatory Reference:** RBI Policy Guidelines on Consortium Ledgers; ISO 27001 Controls.
- **Disallowed Synonyms:** *crypto miner*, *staking pool*, *mining farm*.

#### TERM-023: Multi-Signature Authorization (M-of-N)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `M-of-N`
- **Canonical Code Symbol:** `MultiSigAuthorization`
- **Precise Definition:** A threshold cryptographic authorization scheme requiring $M$ distinct private key signatures from $N$ institutional custodians backed by HSMs to execute sensitive ledger actions.
- **Requirement:** Default threshold $M=3, N=5$.
- **Regulatory Reference:** FIPS 140-2 Level 3 HSM Mandate; SEBI Cyber Security Framework.
- **Disallowed Synonyms:** *multisig wallet*, *gnosis safe clone*, *crypto multisig*.

#### TERM-024: Zero PII Ledger Invariant (ZPII)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `ZPII`
- **Canonical Code Symbol:** `ZeroPiiInvariant`
- **Precise Definition:** The inviolable system architecture rule prohibiting any Personally Identifiable Information (Aadhaar, PAN, name, phone) from being committed to ledger state or calldata.
- **Regulatory Reference:** DPDP Act 2023, Section 8(1); Aadhaar Act 2016, Section 29.
- **Disallowed Synonyms:** *gdpr compliant token*, *anonymized crypto*, *shielded address*.

#### TERM-025: Deterministic Finality (DFIN)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `DFIN`
- **Canonical Code Symbol:** `DeterministicFinality`
- **Precise Definition:** State guarantee wherein a block once included in the ledger can never be re-organized, rolled back, or orphaned under QBFT consensus ($P(\text{Reorg}) = 0$).
- **Regulatory Reference:** SEBI Settlement Finality Principles; CPSS-IOSCO Principles.
- **Disallowed Synonyms:** *probabilistic finality*, *nakamoto consensus*, *pow confirmations*.

#### TERM-026: Security Token Clawback (STC)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `STC`
- **Canonical Code Symbol:** `TokenClawbackController`
- **Precise Definition:** Regulatory recovery smart contract function allowing court-ordered or regulator-mandated seizure and reallocation of digital security tokens.
- **Conservation of Supply:**
  $$\text{TotalSupply}_{\text{Post}} \equiv \text{TotalSupply}_{\text{Pre}}$$
- **Regulatory Reference:** SEBI Disgorgement Guidelines; Civil Court Decrees.
- **Disallowed Synonyms:** *token burn hack*, *admin key exploit*, *backdoor transfer*.

#### TERM-027: Custody Merkle Root (CMR)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `CMR`
- **Canonical Code Symbol:** `CustodyMerkleRoot`
- **Precise Definition:** A 32-byte cryptographic root hash recorded periodically on-chain representing depository participant share balances verified by statutory auditors.
- **Regulatory Reference:** SEBI Systems Audit Reporting Framework.
- **Disallowed Synonyms:** *state root*, *world state hash*, *oracle pulse*.

#### TERM-028: Token Partition (TPAR)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `TPAR`
- **Canonical Code Symbol:** `TokenPartition`
- **Precise Definition:** Segregated sub-balance within an investor's token account (e.g., Free, Locked for Settlement, Restricted, Pledged) compliant with ERC-1400/ERC-3643.
- **Balance Invariant:**
  $$\text{TotalBalance}(A) \equiv \sum_{p \in \text{Partitions}} \text{PartitionBalance}(A, p)$$
- **Regulatory Reference:** ERC-1400 / ERC-3643 Partition Standards.
- **Disallowed Synonyms:** *locked token*, *vesting coin*, *staking locked token*.

#### TERM-029: On-Chain Freeze Controller (OFC)
- **Domain:** Permissioned Blockchain & Distributed Ledger
- **Acronym:** `OFC`
- **Canonical Code Symbol:** `OnChainFreezeController`
- **Precise Definition:** Smart contract access-controlled module enabling immediate freeze of specific accounts or token partitions under lawful enforcement directives.
- **Regulatory Reference:** PMLA 2002, Section 17; SEBI Intermediary Regulations.
- **Disallowed Synonyms:** *rug pull killswitch*, *tether blacklist*, *token freeze*.

---

### 2.3 Domain 3: Regulatory & Compliance

```
+-----------------------------------------------------------------------------------------------------+
| TERM-030 through TERM-043: Regulatory & Compliance Domain Summary Table                             |
+----------+-----------------------------------+---------+----------------------------+---------------+
| Term ID  | Term Name                         | Acronym | Canonical Symbol           | Act / Mandate |
+----------+-----------------------------------+---------+----------------------------+---------------+
| TERM-030 | Securities & Exchange Board India | SEBI    | SebiRegulatoryFramework    | SEBI Act 1992 |
| TERM-031 | Reserve Bank of India             | RBI     | RbiBankingFramework        | RBI Act 1934  |
| TERM-032 | Int'l Financial Services Centres  | IFSCA   | IfscaRegulatoryFramework   | IFSCA Act 2019|
| TERM-033 | Prevention of Money Laundering    | PMLA    | PmlaComplianceEngine       | PMLA 2002     |
| TERM-034 | Digital Personal Data Protection  | DPDP    | DpdpPrivacyGuard           | DPDP Act 2023 |
| TERM-035 | Financial Action Task Force       | FATF    | FatfComplianceMonitor      | FATF Rec 16   |
| TERM-036 | Politically Exposed Person        | PEP     | PepScreeningService        | RBI KYC S41   |
| TERM-037 | Central KYC Records Registry      | CKYC    | CkycRegistryGateway        | CERSAI Rules  |
| TERM-038 | Suspicious Transaction Report     | STR     | SuspiciousTransactionReport| FIU-IND 7 Days|
| TERM-039 | Cash Transaction Report           | CTR     | CashTransactionReport      | FIU-IND 10L   |
| TERM-040 | Regulatory Sandbox Framework      | RSF     | RegulatorySandboxFramework | SEBI Sandbox  |
| TERM-041 | Travel Rule Compliance            | TRC     | TravelRuleEngine           | FATF Rec 16   |
| TERM-042 | Aadhaar Offline e-KYC             | OKYC    | AadhaarOfflineEkyc         | UIDAI Reg 2021|
| TERM-043 | Tax Deduction at Source           | TDS     | TdsComplianceService       | IT Act S194   |
+----------+-----------------------------------+---------+----------------------------+---------------+
```

#### TERM-030: Securities and Exchange Board of India (SEBI)
- **Domain:** Regulatory & Compliance
- **Acronym:** `SEBI`
- **Canonical Code Symbol:** `SebiRegulatoryFramework`
- **Precise Definition:** The apex statutory regulatory body overseeing securities and capital markets in the Republic of India.
- **Regulatory Reference:** Securities and Exchange Board of India Act 1992.
- **Disallowed Synonyms:** *crypto regulator*, *dao foundation*, *web3 council*.

#### TERM-031: Reserve Bank of India (RBI)
- **Domain:** Regulatory & Compliance
- **Acronym:** `RBI`
- **Canonical Code Symbol:** `RbiBankingFramework`
- **Precise Definition:** India's central banking institution governing monetary policy, banking rails, payment systems, and foreign exchange regulations.
- **Regulatory Reference:** Reserve Bank of India Act 1934; Payment and Settlement Systems Act 2007.
- **Disallowed Synonyms:** *defi treasury*, *crypto central bank*, *liquidity governor*.

#### TERM-032: International Financial Services Centres Authority (IFSCA)
- **Domain:** Regulatory & Compliance
- **Acronym:** `IFSCA`
- **Canonical Code Symbol:** `IfscaRegulatoryFramework`
- **Precise Definition:** The unified statutory authority regulating financial institutions, cross-border exchanges, and securities operating in IFSC GIFT City, Gujarat.
- **Regulatory Reference:** IFSCA Act 2019; IFSCA (Capital Market Intermediaries) Regulations 2021.
- **Disallowed Synonyms:** *offshore crypto haven*, *tax haven regulator*, *cayman foundation*.

#### TERM-033: Prevention of Money Laundering Act (PMLA)
- **Domain:** Regulatory & Compliance
- **Acronym:** `PMLA`
- **Canonical Code Symbol:** `PmlaComplianceEngine`
- **Precise Definition:** Indian statutory framework designed to prevent money-laundering, terrorist financing, and provide for confiscation of property derived from illegal activities.
- **Regulatory Reference:** Prevention of Money Laundering Act 2002; FIU-IND Guidelines.
- **Disallowed Synonyms:** *crypto aml*, *blockchain forensics*, *tornado screening*.

#### TERM-034: Digital Personal Data Protection Act (DPDP)
- **Domain:** Regulatory & Compliance
- **Acronym:** `DPDP`
- **Canonical Code Symbol:** `DpdpPrivacyGuard`
- **Precise Definition:** Indian data protection statute governing digital processing of personal data, mandating consent, purpose limitation, and the Data Principal's erasure rights.
- **Regulatory Reference:** Digital Personal Data Protection Act 2023 (Act No. 22 of 2023).
- **Disallowed Synonyms:** *crypto privacy*, *zk anonymizer*, *dark pool privacy*.

#### TERM-035: Financial Action Task Force (FATF)
- **Domain:** Regulatory & Compliance
- **Acronym:** `FATF`
- **Canonical Code Symbol:** `FatfComplianceMonitor`
- **Precise Definition:** Global inter-governmental money laundering and terrorist financing watchdog setting international compliance standards.
- **Regulatory Reference:** FATF 40 Recommendations (Recommendation 16 & Guidance on Virtual Assets).
- **Disallowed Synonyms:** *crypto task force*, *web3 sanctions committee*.

#### TERM-036: Politically Exposed Person (PEP)
- **Domain:** Regulatory & Compliance
- **Acronym:** `PEP`
- **Canonical Code Symbol:** `PepScreeningService`
- **Precise Definition:** An individual entrusted with prominent public functions, requiring Enhanced Due Diligence (EDD) and senior management approval prior to onboarding.
- **Regulatory Reference:** RBI Master Direction - Know Your Customer (KYC) Direction 2016, Section 41.
- **Disallowed Synonyms:** *high net worth crypto user*, *vip whale*, *dao council member*.

#### TERM-037: Central KYC Records Registry (CKYC)
- **Domain:** Regulatory & Compliance
- **Acronym:** `CKYC`
- **Canonical Code Symbol:** `CkycRegistryGateway`
- **Precise Definition:** A centralized repository of KYC records of financial sector customers, maintained by CERSAI under Government of India mandate.
- **Regulatory Reference:** PML (Maintenance of Records) Rules 2005, Rule 9A.
- **Disallowed Synonyms:** *decentralized id*, *did credential*, *soulbound kyc token*.

#### TERM-038: Suspicious Transaction Report (STR)
- **Domain:** Regulatory & Compliance
- **Acronym:** `STR`
- **Canonical Code Symbol:** `SuspiciousTransactionReport`
- **Precise Definition:** Statutory report furnished to Financial Intelligence Unit - India (FIU-IND) regarding transactions giving rise to reasonable suspicion of illicit proceeds.
- **Mandate:** Must be filed within 7 working days of arriving at a conclusion of suspicion.
- **Regulatory Reference:** PMLA 2002, Section 12; FIU-IND Reporting Format.
- **Disallowed Synonyms:** *crypto alert*, *scam flag*, *chainalysis hit*.

#### TERM-039: Cash Transaction Report (CTR)
- **Domain:** Regulatory & Compliance
- **Acronym:** `CTR`
- **Canonical Code Symbol:** `CashTransactionReport`
- **Precise Definition:** Mandatory monthly reporting to FIU-IND of all cash transactions exceeding INR 10 Lakhs or its foreign currency equivalent.
- **Regulatory Reference:** PMLA 2002, Section 12; PML Rules Rule 3.
- **Disallowed Synonyms:** *crypto cash out report*, *fiat dump report*.

#### TERM-040: Regulatory Sandbox Framework (RSF)
- **Domain:** Regulatory & Compliance
- **Acronym:** `RSF`
- **Canonical Code Symbol:** `RegulatorySandboxFramework`
- **Precise Definition:** A controlled regulatory environment created by SEBI allowing live testing of innovative fintech products on a limited set of customers.
- **Regulatory Reference:** SEBI Circular SEBI/HO/MIRSD/DOS3/CIR/P/2021/643.
- **Disallowed Synonyms:** *crypto testnet*, *beta launch*, *unlicensed pilot*.

#### TERM-041: Travel Rule Compliance (TRC)
- **Domain:** Regulatory & Compliance
- **Acronym:** `TRC`
- **Canonical Code Symbol:** `TravelRuleEngine`
- **Precise Definition:** Protocol ensuring that verified originator and beneficiary information accompanies cross-border and domestic funds and asset transfers.
- **Regulatory Reference:** FATF Recommendation 16; IFSCA AML Guidelines.
- **Disallowed Synonyms:** *bridge memo*, *tx metadata*, *crypto tag*.

#### TERM-042: Aadhaar Offline e-KYC (OKYC)
- **Domain:** Regulatory & Compliance
- **Acronym:** `OKYC`
- **Canonical Code Symbol:** `AadhaarOfflineEkyc`
- **Precise Definition:** Privacy-preserving XML-based verification mechanism authorized by UIDAI where no Aadhaar number is stored or exposed.
- **Regulatory Reference:** Aadhaar (Authentication and Offline Verification) Regulations 2021.
- **Disallowed Synonyms:** *kyc token*, *biometric hash on-chain*, *aadhaar leak*.

#### TERM-043: Tax Deduction at Source (TDS)
- **Domain:** Regulatory & Compliance
- **Acronym:** `TDS`
- **Canonical Code Symbol:** `TdsComplianceService`
- **Precise Definition:** Statutory deduction of income tax at the source of transaction execution pursuant to Indian Income Tax Act requirements.
- **Regulatory Reference:** Income Tax Act 1961, Section 194S & Section 194Q.
- **Disallowed Synonyms:** *crypto withholding*, *token tax*, *burn fee*.

---

### 2.4 Domain 4: Financial, Accounting & Monetization

```
+-----------------------------------------------------------------------------------------------------+
| TERM-044 through TERM-058: Financial, Accounting & Monetization Domain Summary Table                |
+----------+-----------------------------------+---------+----------------------------+---------------+
| Term ID  | Term Name                         | Acronym | Canonical Symbol           | Formula / Cap |
+----------+-----------------------------------+---------+----------------------------+---------------+
| TERM-044 | Fixed Platform Transaction Fee    | FPTF    | FixedPlatformTransactionFee| 0.00% Launch  |
| TERM-045 | FIFO Cost Basis Engine            | FIFO    | FifoCostBasisEngine        | Earliest Lot  |
| TERM-046 | Realized Capital Gain             | RCG     | RealizedCapitalGain        | S111A / S112A |
| TERM-047 | Unrealized Profit and Loss        | UPNL    | UnrealizedPnL              | MTM Valuation |
| TERM-048 | Tax Lot                           | TLOT    | TaxLot                     | Lot Record    |
| TERM-049 | Trade Notional Turnover           | TNT     | TradeNotionalTurnover      | Price * Qty   |
| TERM-050 | Hold Balance Allocation           | HBAL    | HoldBalanceAllocation      | Escrow Lock   |
| TERM-051 | Multi-Vault Revenue Split         | MVRS    | MultiVaultRevenueSplit     | 10000 bps Sum |
| TERM-052 | Maximum Fee Ceiling               | MFC     | MaxFeeCeiling              | 50 bps Max Cap|
| TERM-053 | Timelocked Gov Adjustment         | TLGA    | TimelockedGovAdjustment    | 48-Hour Delay |
| TERM-054 | Available Withdrawable Balance    | AWB     | AvailableWithdrawableBalance| Settled Cash  |
| TERM-055 | Double-Entry Journal Booking      | DEJB    | DoubleEntryJournal         | Debit == Credit|
| TERM-056 | Net Settled Proceeds              | NSP     | NetSettledProceeds         | Trade Credit  |
| TERM-057 | Corporate Action Cash Distribution| CACD    | CorporateActionCashDist    | Pro-rata Div  |
| TERM-058 | Electronic Contract Note          | ECN     | ElectronicContractNote     | 24-Hour ECN   |
+----------+-----------------------------------+---------+----------------------------+---------------+
```

#### TERM-044: Fixed Platform Transaction Fee (FPTF)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `FPTF`
- **Canonical Code Symbol:** `FixedPlatformTransactionFee`
- **Precise Definition:** A 0.00% fee (No fee at all) assessed on trade notional turnover across executed transactions, operating with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol); FIFO capital gains are computed strictly for user tax compliance (Section 111A/112A).
- **Mathematical Implication:**
  $$\text{Fee} = \text{TradeNotional} \times 0.0000 = 0.00 \implies \text{NetCredit} = \text{TradeNotional} - \text{StatutoryTaxes}$$
- **Regulatory Reference:** Growww Fee Model Mandate; SEBI Brokerage Regulations.
- **Disallowed Synonyms:** *management fee*, *AUM fee*, *holding fee*, *gas fee*, *profit-only fee*, *maker-taker fee*.

#### TERM-045: FIFO Cost Basis Engine (FIFO)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `FIFO`
- **Canonical Code Symbol:** `FifoCostBasisEngine`
- **Precise Definition:** First-In, First-Out tax lot accounting engine tracking exact acquisition timestamps, prices, and quantities to calculate capital gains strictly for user tax compliance.
- **Queue Invariant:** The earliest unsettled tax lot acquired is the first lot matched and closed upon sale.
- **Regulatory Reference:** Income Tax Act 1961, Section 111A, 112A; CBDT Circular No. 768.
- **Disallowed Synonyms:** *lifo engine*, *average cost blender*, *crypto tax estimator*.

#### TERM-046: Realized Capital Gain (RCG)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `RCG`
- **Canonical Code Symbol:** `RealizedCapitalGain`
- **Precise Definition:** The crystallized economic gain or loss resulting from the sale of an equity position against its FIFO cost basis, classified into STCG (held $\le 12$ months) or LTCG (held $> 12$ months).
- **Mathematical Formula:**
  $$\text{Gain}_{\text{Lot } i} = Q_{\text{Matched}, i} \times (P_{\text{Sell}} - P_{\text{Acquisition}, i})$$
  $$\text{RealizedGain}_{\text{Total}} = \sum_{i=1}^{K} \text{Gain}_{\text{Lot } i}$$
- **Regulatory Reference:** Income Tax Act 1961, Sections 45 to 55A.
- **Disallowed Synonyms:** *speculative crypto profit*, *harvested yield*, *token flip gain*.

#### TERM-047: Unrealized Profit and Loss (UPNL)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `UPNL`
- **Canonical Code Symbol:** `UnrealizedPnL`
- **Precise Definition:** Mark-to-Market (MTM) paper profit or loss on currently open positions evaluated against real-time exchange reference prices.
- **Mathematical Formula:**
  $$\text{UPnL} = \sum_{j} Q_{\text{Open}, j} \times (P_{\text{Market}} - P_{\text{Cost}, j})$$
- **Regulatory Reference:** Ind AS 109 Financial Instruments.
- **Disallowed Synonyms:** *paper crypto gains*, *token float*, *synthetic pnl*.

#### TERM-048: Tax Lot (TLOT)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `TLOT`
- **Canonical Code Symbol:** `TaxLot`
- **Precise Definition:** An individual accounting record created upon tokenized equity purchase recording acquisition timestamp, settled quantity, per-unit price, and statutory fees.
- **Schema Tuple:**
  $$\langle \text{LotID: UUID}, \text{ISIN: String}, \text{AcquiredAt: UTC}, \text{Qty: Decimal18}, \text{UnitCost: Decimal18} \rangle$$
- **Regulatory Reference:** Income Tax Rules 1962, Rule 8D.
- **Disallowed Synonyms:** *utxo*, *crypto input*, *token batch*.

#### TERM-049: Trade Notional Turnover (TNT)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `TNT`
- **Canonical Code Symbol:** `TradeNotionalTurnover`
- **Precise Definition:** Total monetary value of an executed trade, calculated as executed fill price multiplied by executed fractional equity quantity.
- **Formula:**
  $$\text{Turnover} = \text{FillPrice} \times \text{FillQuantity}$$
- **Regulatory Reference:** SEBI Turnover Fee Regulations.
- **Disallowed Synonyms:** *crypto volume*, *gas total*, *swap turnover*.

#### TERM-050: Hold Balance Allocation (HBAL)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `HBAL`
- **Canonical Code Symbol:** `HoldBalanceAllocation`
- **Precise Definition:** Segregated ledger state where fiat funds or securities are locked in escrow to guarantee settlement obligations during active orders.
- **Ledger Invariant:**
  $$\text{TotalBalance} \equiv \text{AvailableBalance} + \text{HoldBalance}$$
- **Regulatory Reference:** SEBI Upstreaming of Client Funds Circular.
- **Disallowed Synonyms:** *staked tokens*, *escrowed coins*, *locked liquidity*.

#### TERM-051: Multi-Vault Revenue Split (MVRS)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `MVRS`
- **Canonical Code Symbol:** `MultiVaultRevenueSplit`
- **Precise Definition:** Smart contract governed distribution mechanism routing protocol fees (when active) across statutory reserves, operations, and treasury accounts.
- **Split Constraint:**
  $$\text{Ratio}_{\text{Operations}} + \text{Ratio}_{\text{Reserve}} + \text{Ratio}_{\text{Treasury}} = 10,000 \text{ bps} \ (100.00\%)$$
- **Regulatory Reference:** Companies Act 2013, Section 123 (Dividend & Reserve Rules).
- **Disallowed Synonyms:** *token reward distributor*, *staking fee pool*, *burn address*.

#### TERM-052: Maximum Fee Ceiling (MFC)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `MFC`
- **Canonical Code Symbol:** `MaxFeeCeiling`
- **Precise Definition:** An immutable smart-contract and regulatory hard cap limiting any future platform fee increases to a maximum of 50 basis points (0.50%).
- **Hard Assertion:**
  $$\text{ProposedFeeBps} \le 50 \text{ bps} \ (0.50\%)$$
- **Regulatory Reference:** SEBI (Stock Brokers) Regulations 1992, Brokerage Cap Guidelines.
- **Disallowed Synonyms:** *unlimited gas*, *variable dynamic fee*, *slippage toll*.

#### TERM-053: Timelocked Governance Adjustment (TLGA)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `TLGA`
- **Canonical Code Symbol:** `TimelockedGovAdjustment`
- **Precise Definition:** Mandatory 48-hour time delay before any approved fee schedule or parameter change can become active on-chain, ensuring investor transparency.
- **Time Check:**
  $$\text{BlockTimestamp} \ge \text{ScheduledExecutionTimestamp} + 172,800 \text{ seconds}$$
- **Regulatory Reference:** SEBI Circular on Material System Changes Notification.
- **Disallowed Synonyms:** *instant admin upgrade*, *rugpull window*, *secret fork*.

#### TERM-054: Available Withdrawable Balance (AWB)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `AWB`
- **Canonical Code Symbol:** `AvailableWithdrawableBalance`
- **Precise Definition:** Settled cash funds eligible for immediate banking rail withdrawal, net of unsettled trades, open order holds, and statutory margin requirements.
- **Formula:**
  $$\text{AWB} = \max(0, \text{SettledCash} - \text{ActiveHolds} - \text{UnsettledDebits})$$
- **Regulatory Reference:** SEBI Client Funds Segregation and Payout Mandate.
- **Disallowed Synonyms:** *unstaked balance*, *liquid crypto*, *free tokens*.

#### TERM-055: Double-Entry Journal Booking (DEJB)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `DEJB`
- **Canonical Code Symbol:** `DoubleEntryJournal`
- **Precise Definition:** Immutable ledger transaction entry requiring equal and opposite debit and credit entries, preserving the fundamental accounting equation.
- **Balanced Transaction Constraint:**
  $$\sum \text{Debits} - \sum \text{Credits} = 0, \quad \forall \text{ Transaction UUID}$$
- **Regulatory Reference:** ICAI Accounting Standards; Companies Act 2013, Section 128.
- **Disallowed Synonyms:** *balance mutation*, *token mint/burn delta*, *single entry write*.

#### TERM-056: Net Settled Proceeds (NSP)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `NSP`
- **Canonical Code Symbol:** `NetSettledProceeds`
- **Precise Definition:** The net fiat cash credit resulting from trade execution after deducting 0.00% platform fee at launch and applicable statutory taxes.
- **Calculation:**
  $$\text{NetSettledProceeds} = \text{GrossNotional} - \text{STT} - \text{StampDuty} - \text{TurnoverCharges} - \text{GST}$$
- **Regulatory Reference:** SEBI Settlement Payout Timelines.
- **Disallowed Synonyms:** *crypto payout*, *cash out sum*, *net crypto yield*.

#### TERM-057: Corporate Action Cash Distribution (CACD)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `CACD`
- **Canonical Code Symbol:** `CorporateActionCashDistribution`
- **Precise Definition:** Pro-rata allocation and credit of cash dividends paid by underlying equity issuers directly to fractional beneficial owners.
- **Formula:**
  $$\text{Credit}_i = \text{NetDividendPool} \times \left(\frac{\text{FractionalUnits}_i}{\sum \text{EligibleUnits}}\right)$$
- **Regulatory Reference:** Companies Act 2013, Section 124; SEBI LODR Regulations.
- **Disallowed Synonyms:** *dividend airdrop*, *staking reward*, *token bonus*.

#### TERM-058: Electronic Contract Note (ECN)
- **Domain:** Financial, Accounting & Monetization
- **Acronym:** `ECN`
- **Canonical Code Symbol:** `ElectronicContractNote`
- **Precise Definition:** Cryptographically signed legal document confirming trade execution, statutory taxes, and settlement amounts sent to the user within 24 hours.
- **Regulatory Reference:** SEBI Circular SMDRP/POLICY/CIR-56/2000; Information Technology Act 2000.
- **Disallowed Synonyms:** *tx receipt*, *blockchain explorer link*, *crypto receipt*.

---

### 2.5 Domain 5: Banking, Fiat Rails & Intermediary

```
+-----------------------------------------------------------------------------------------------------+
| TERM-059 through TERM-072: Banking, Fiat Rails & Intermediary Domain Summary Table                  |
+----------+-----------------------------------+---------+----------------------------+---------------+
| Term ID  | Term Name                         | Acronym | Canonical Symbol           | Rail / Spec   |
+----------+-----------------------------------+---------+----------------------------+---------------+
| TERM-059 | UPI Intent Flow                   | UPII    | UpiIntentFlow              | NPCI Intent   |
| TERM-060 | Virtual Payment Address           | VPA     | VirtualPaymentAddress      | user@bank     |
| TERM-061 | Nodal Account                     | NODAL   | NodalAccount               | RBI Guidelines|
| TERM-062 | Regulated Escrow Account          | ESCROW  | RegulatedEscrowAccount     | Tripartite Agr|
| TERM-063 | Real-Time Gross Settlement        | RTGS    | RtgsSettlementRail         | >= INR 200,000|
| TERM-064 | National Electronic Funds Transfer| NEFT    | NeftSettlementRail         | Batched 48x/day|
| TERM-065 | Immediate Payment Service         | IMPS    | ImpsSettlementRail         | <= INR 500,000|
| TERM-066 | Payment Aggregator                | PA      | PaymentAggregatorGateway   | RBI PA/PG 2020|
| TERM-067 | Dynamic QR Code Collection        | DQRC    | DynamicQrCollection        | 5-Min TTL QR  |
| TERM-068 | AutoPay Recurring Mandate         | AUTOPAY | AutoPayMandate             | UPI Mandate   |
| TERM-069 | IFSC Banking Unit                 | IBU     | IfscBankingUnit            | GIFT City FX  |
| TERM-070 | Nostro-Vostro Account Hierarchy   | NOVA    | NostroVostroHierarchy      | Interbank FX  |
| TERM-071 | Liberalised Remittance Scheme     | LRS     | LrsRemittanceGateway       | <= $250k/year |
| TERM-072 | Virtual Account Number            | VAN     | VirtualAccountNumber       | Dynamic Van   |
+----------+-----------------------------------+---------+----------------------------+---------------+
```

#### TERM-059: Unified Payments Interface Intent (UPII)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `UPII`
- **Canonical Code Symbol:** `UpiIntentFlow`
- **Precise Definition:** Mobile app-to-app deep linking protocol facilitating real-time instant debit from user bank accounts to regulated nodal accounts.
- **URI Format:** `upi://pay?pa={vpa}&pn={name}&am={amount}&cu=INR&tr={order_id}`.
- **Regulatory Reference:** NPCI UPI Procedural Guidelines; RBI Master Directions on PPIs.
- **Disallowed Synonyms:** *crypto on-ramp*, *fiat gateway widget*, *moonpay flow*.

#### TERM-060: Virtual Payment Address (VPA)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `VPA`
- **Canonical Code Symbol:** `VirtualPaymentAddress`
- **Precise Definition:** An alphanumeric financial identifier (`user@bank`) mapping to an underlying bank account under NPCI UPI infrastructure.
- **Regex Validation:** `^[a-zA-Z0-9.-]+@[a-zA-Z][a-zA-Z0-9]*$`.
- **Regulatory Reference:** NPCI UPI Core Architecture Specifications.
- **Disallowed Synonyms:** *wallet address*, *ens domain*, *crypto handle*.

#### TERM-061: Nodal Account (NODAL)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `NODAL`
- **Canonical Code Symbol:** `NodalAccount`
- **Precise Definition:** A special internal bank account maintained by an intermediary under RBI guidelines to pool and route funds exclusively between buyers and sellers.
- **Reconciliation Invariant:**
  $$\text{Balance}(\text{Nodal Account}) \equiv \sum \text{Client Unsettled Funds} + \sum \text{Pending Payouts}$$
- **Regulatory Reference:** RBI Guidelines on Nodal Accounts DPSS.CO.PD.No.1102/02.14.08/2009-10.
- **Disallowed Synonyms:** *crypto pool*, *exchange wallet*, *hot wallet account*.

#### TERM-062: Regulated Escrow Account (ESCROW)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `ESCROW`
- **Canonical Code Symbol:** `RegulatedEscrowAccount`
- **Precise Definition:** A ring-fenced bank account held with a scheduled commercial bank governed by a tripartite agreement ensuring conditional fund releases upon settlement triggers.
- **Regulatory Reference:** RBI Master Directions on Escrow Accounts; Indian Contract Act 1872.
- **Disallowed Synonyms:** *smart contract escrow*, *crypto multisig lock*, *defi vault*.

#### TERM-063: Real-Time Gross Settlement (RTGS)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `RTGS`
- **Canonical Code Symbol:** `RtgsSettlementRail`
- **Precise Definition:** Continuous, real-time gross settlement banking rail operated by RBI for high-value financial transfers (INR 2,00,000 and above).
- **Regulatory Reference:** RBI RTGS System Regulations 2013.
- **Disallowed Synonyms:** *high value crypto transfer*, *on-chain whale transfer*.

#### TERM-064: National Electronic Funds Transfer (NEFT)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `NEFT`
- **Canonical Code Symbol:** `NeftSettlementRail`
- **Precise Definition:** Nationwide payment system operated by RBI facilitating half-hourly batched electronic funds transfer between bank accounts.
- **Cycle:** 48 half-hourly batches daily (24x7x365).
- **Regulatory Reference:** RBI NEFT Procedural Guidelines 2019.
- **Disallowed Synonyms:** *batch crypto transfer*, *l2 batch rollup*.

#### TERM-065: Immediate Payment Service (IMPS)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `IMPS`
- **Canonical Code Symbol:** `ImpsSettlementRail`
- **Precise Definition:** 24x7 instant interbank electronic fund transfer service managed by NPCI for domestic transfers up to statutory limits (INR 5,00,000).
- **Regulatory Reference:** NPCI IMPS Operating Guidelines.
- **Disallowed Synonyms:** *instant crypto bridge*, *lightning network rail*.

#### TERM-066: Payment Aggregator (PA)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `PA`
- **Canonical Code Symbol:** `PaymentAggregatorGateway`
- **Precise Definition:** RBI-authorized entity facilitating platforms and merchants to accept payment instruments from customers and settle them into nodal accounts.
- **Regulatory Reference:** RBI Guidelines on Regulation of Payment Aggregators and Payment Gateways (2020).
- **Disallowed Synonyms:** *crypto payment processor*, *bitpay gateway*, *web3 checkout*.

#### TERM-067: Dynamic QR Code Collection (DQRC)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `DQRC`
- **Canonical Code Symbol:** `DynamicQrCollection`
- **Precise Definition:** Real-time generated UPI QR code embedding order reference, exact payable amount, and strict expiry timestamp (300 seconds TTL) for instant reconciliation.
- **Regulatory Reference:** NPCI UPI Specifications for Dynamic QR Codes.
- **Disallowed Synonyms:** *crypto invoice qr*, *bitcoin address qr*.

#### TERM-068: AutoPay Recurring Mandate (AUTOPAY)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `AUTOPAY`
- **Canonical Code Symbol:** `AutoPayMandate`
- **Precise Definition:** Automated standing instruction mandate under NPCI UPI AutoPay facilitating recurring systematic investment plans (SIPs).
- **Requirement:** Push notification to user exactly 24 hours prior to debit.
- **Regulatory Reference:** RBI Framework for Processing of e-Mandates for Recurring Transactions.
- **Disallowed Synonyms:** *smart contract token allowance*, *erc20 approve stream*.

#### TERM-069: IFSC Banking Unit (IBU)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `IBU`
- **Canonical Code Symbol:** `IfscBankingUnit`
- **Precise Definition:** A branch of a scheduled commercial bank licensed by IFSCA to conduct offshore and cross-border banking operations in GIFT City.
- **Regulatory Reference:** IFSCA (Banking) Regulations 2020.
- **Disallowed Synonyms:** *offshore crypto bank*, *crypto custodian branch*.

#### TERM-070: Nostro-Vostro Account Hierarchy (NOVA)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `NOVA`
- **Canonical Code Symbol:** `NostroVostroHierarchy`
- **Precise Definition:** Interbank correspondent banking structure used to settle cross-border foreign currency and INR convertible accounts between entities.
- **Invariant:**
  $$\text{NostroBalance}_{\text{USD}} \equiv \text{VostroLiability}_{\text{USD}}$$
- **Regulatory Reference:** RBI Master Direction - Foreign Exchange Management Act (FEMA).
- **Disallowed Synonyms:** *crosschain liquidity bridge*, *wrapped currency pool*.

#### TERM-071: Liberalised Remittance Scheme (LRS)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `LRS`
- **Canonical Code Symbol:** `LrsRemittanceGateway`
- **Precise Definition:** RBI foreign exchange framework permitting resident individuals to remit up to USD 250,000 per financial year for permissible capital account transactions.
- **Regulatory Reference:** RBI Master Direction on Liberalised Remittance Scheme (LRS).
- **Disallowed Synonyms:** *crypto off-ramp*, *borderless crypto transfer*.

#### TERM-072: Virtual Account Number (VAN)
- **Domain:** Banking, Fiat Rails & Intermediary
- **Acronym:** `VAN`
- **Canonical Code Symbol:** `VirtualAccountNumber`
- **Precise Definition:** A unique dynamic bank account number allocated to each user by an authorized partner bank to automatically match NEFT/RTGS inward deposits.
- **Validation:** Modulo 97-10 checksum validation.
- **Regulatory Reference:** RBI Core Banking Integration Standards.
- **Disallowed Synonyms:** *deposit crypto address*, *hd wallet derivation*.

---

## 3. Cross-Domain Mapping Architecture

Connecting traditional depository and banking constructs to permissioned distributed ledger representations:

```
+-----------------------------------------------------------------------------------------------------------------------------------------+
|                                              CROSS-DOMAIN ENTITY MAPPING MATRIX                                                         |
+------------------------------+-------------------------------+------------------------------+---------------------------+---------------+
| Traditional Depository /     | Permissioned Ledger Entity    | Core Microservice Contract   | Database Representation   | Statutory     |
| Banking Construct            | (Hyperledger Besu)            | (gRPC / Protobuf)            | (PostgreSQL Schema)       | Authority     |
+------------------------------+-------------------------------+------------------------------+---------------------------+---------------+
| NSDL / CDSL Whole Share      | DematHoldingPool (ERC-3643)   | depository.DepositoryGateway | depository_holding_pool   | SEBI          |
| Demat Beneficiary Account    | IdentityCommitment (bytes32)  | identity.IdentityService     | demat_account_registry    | SEBI / CERSAI |
| Delivery Instruction Slip    | DvPSettlement.sol             | settlement.SettlementService | settlement_instruction    | SEBI          |
| Share Transfer Novation      | Atomic Transfer Event         | matching.MatchingEngine      | matched_trade_journal     | SEBI / CCIL   |
| Corporate Action Dividend    | DividendDistributor.sol       | corporate.CorporateService   | dividend_allocation_lot   | MCA / SEBI    |
| Bank Nodal Account           | NodalEscrowRegistry.sol       | banking.NodalManager         | nodal_bank_ledger         | RBI           |
| Escrow Tripartite Hold       | EscrowHoldPartition.sol       | banking.EscrowService        | fiat_escrow_hold          | RBI           |
| UPI Instant Payment          | UpiSettlementAttestation.sol  | payments.UpiGateway          | upi_transaction_log       | NPCI / RBI    |
| CKYC 14-Digit Record         | IdentityRegistry.sol          | kyc.CkycService              | ckyc_profile_store        | CERSAI / PMLA |
| Aadhaar XML e-KYC            | IdentityAttestationVerifier   | kyc.AadhaarService           | kyc_audit_attestation     | UIDAI         |
| FIFO Tax Lot Allocation      | FifoCostLedger.sol            | accounting.TaxLotEngine      | tax_lot_inventory         | CBDT          |
| ECN Contract Note            | ContractNoteAttestation.sol   | compliance.ContractNoteSvc   | electronic_contract_note  | SEBI          |
| Proof of Depository Solvency | CustodyRootStorage.sol        | auditor.ProofOfReserveSvc    | custody_merkle_checkpoint | SEBI          |
+------------------------------+-------------------------------+------------------------------+---------------------------+---------------+
```

---

## 4. Polyglot Canonical Code Symbol Mappings

Every technical entity must use the standard naming conventions across the multi-language repository:

```
+----------------------------------------------------------------------------------------------------------------------------------------------+
|                                              POLYGLOT CODE SYMBOL MAPPING STANDARDS                                                          |
+---------------------------+----------------------------+----------------------------+----------------------------+---------------------------+
| Domain Entity             | Go (1.22)                  | Python (3.11)              | Rust (1.78)                | Dart / Flutter (3.22)     | Solidity (0.8.24)         |
+---------------------------+----------------------------+----------------------------+----------------------------+---------------------------+
| Delivery versus Payment   | DvPSettlementService       | DvPSettlementService       | DvpSettlementEngine        | DvpSettlementService      | SettlementDvP.sol         |
| Digital Security Token    | DigitalSecurityToken       | DigitalSecurityToken       | DigitalSecurityToken       | DigitalSecurityTokenModel | DigitalSecurityToken.sol  |
| Depository Participant    | DepositoryParticipant      | DepositoryParticipant      | DepositoryParticipant      | DepositoryParticipantModel| DepositoryRegistry.sol    |
| Proof of Reserve          | ProofOfReserveService      | ProofOfReserveService      | ProofOfReserveEngine       | ProofOfReserveModel       | ProofOfReserveVerifier.sol|
| Merkle Tree Attestation   | MerkleTreeAttestor         | MerkleTreeAttestor         | MerkleTreeAttestation      | MerkleProofModel          | MerkleProofVerifier.sol   |
| Identity Commitment       | IdentityCommitment         | IdentityCommitment         | IdentityCommitment         | IdentityCommitmentModel   | IdentityRegistry.sol      |
| Compliance Registry       | ComplianceRegistryClient   | ComplianceRegistryClient   | ComplianceRegistryClient   | ComplianceRegistryModel   | ComplianceRegistry.sol    |
| Multi-Signature Auth      | MultiSigAuthorizer         | MultiSigAuthorizer         | MultiSigAuthorizer         | MultiSigRequestModel      | MultiSigGovernance.sol    |
| Token Clawback Controller | TokenClawbackController    | TokenClawbackController    | TokenClawbackController    | ClawbackActionModel       | ERC3643Clawback.sol       |
| Fixed Platform Fee        | FixedPlatformFeeCalculator | FixedPlatformFeeCalculator | FixedPlatformFeeCalculator | TransactionFeeModel       | FeeController.sol         |
| FIFO Cost Basis Engine    | FifoCostBasisEngine        | FifoCostBasisEngine        | FifoCostBasisEngine        | FifoTaxSummaryModel       | FifoCostLedger.sol        |
| Realized Capital Gain     | RealizedCapitalGain        | RealizedCapitalGain        | RealizedCapitalGainEngine  | CapitalGainReportModel    | TaxLotRegistry.sol        |
| Tax Lot                   | TaxLot                     | TaxLot                     | TaxLot                     | TaxLotItem                | TaxLotTracker.sol         |
| Multi-Vault Split         | MultiVaultSplitService     | MultiVaultSplitService     | MultiVaultSplitEngine      | RevenueDistributionModel  | MultiVaultSplitter.sol    |
| Nodal Account             | NodalAccountManager        | NodalAccountManager        | NodalAccountManager        | NodalAccountModel         | NodalEscrowRegistry.sol   |
| Regulated Escrow Account  | RegulatedEscrowService     | RegulatedEscrowService     | RegulatedEscrowEngine      | EscrowAccountModel        | FiatEscrowCoordinator.sol |
| UPI Intent Flow           | UpiIntentGateway           | UpiIntentGateway           | UpiIntentGateway           | UpiIntentHandler          | UpiSettlementLog.sol      |
+---------------------------+----------------------------+----------------------------+----------------------------+---------------------------+
```

---

## 5. Forbidden & Deprecated Terminology Mandate

The use of speculative cryptocurrency jargon or inaccurate financial phrasing creates severe regulatory and legal liabilities under SEBI advertising guidelines, ASCI virtual asset rules, and PMLA frameworks. The following terminology is strictly banned from all documentation, codebases, API schemas, commit messages, and marketing materials:

```
+--------------------------------------------------------------------------------------------------------------------------------------------+
|                                              FORBIDDEN TERMINOLOGY & COMPLIANCE MANDATE                                                    |
+----------------------+---------------------------------+------------------------------------------+----------------------------------------+
| Banned Term          | Risk & Regulatory Implication   | Mandatory Canonical Ubiquitous Term      | Affected Code / System Context         |
+----------------------+---------------------------------+------------------------------------------+----------------------------------------+
| "crypto token"       | Misclassifies regulated equity  | Digital Security Token (DST)             | Smart contracts, APIs, UI models       |
| "cryptocurrency"     | Implies speculative coin asset  | Regulated Equity Unit                    | Public documentation, mobile screens   |
| "gas fee"            | Violates Besu zero-gas model    | Platform Transaction Fee (0.00% launch)  | Blockchain adapters, billing engines   |
| "synthetic asset"    | Prohibited derivative violation | 1:1 Asset-Backed Tokenized Security      | Product collateral, token master       |
| "AUM fee"            | Violates 0.00% fee invariant    | Zero Holding Fee (Universal Model)       | Accounting, fee schedules, contracts   |
| "management fee"     | Conflates broker with AMC/AIF   | Fixed Platform Transaction Fee           | Monetization engine, ledger journals   |
| "profit-only fee"    | Misrepresents fee model         | Fixed Platform Turnover Fee              | FeeController.sol, UI pricing tier     |
| "crypto swap"        | Confuses atomic swap with DvP   | Delivery versus Payment (DvP) Settlement | Settlement saga, order routing         |
| "atomic swap"        | Implies unpermissioned dex trade| Delivery versus Payment (DvP) Settlement | Matching engine, settlement contracts  |
| "unhosted wallet"    | PMLA KYC breach risk            | Custodial Ledger Identity Account        | Key management, authentication gateway |
| "smart contract dex" | Unlicensed exchange operation   | Central Limit Order Book (CLOB) + DvP    | Core trading venue specifications      |
| "liquidity pool"     | Conflates AMM with Demat Pool   | Demat Holding Pool                       | Depository gateways, reserve auditors  |
| "token airdrop"      | Violates corporate action rules | Corporate Action Cash / Stock Dist.      | Corporate actions processor            |
| "yield farming"      | Unregulated unregistered scheme | Equity Dividend Yield                    | Portfolio analytics, client dashboard  |
| "burn to zero"       | Speculative tokenomics slang    | Security Token Redemption and Retirement | Depository burn saga, smart contracts  |
| "mining / staking"   | Conflates PoW/PoS with QBFT     | QBFT Permissioned Block Validation       | Besu infrastructure, consensus monitor |
+----------------------+---------------------------------+------------------------------------------+----------------------------------------+
```

---

## 6. Governance, Enforcement & CI/CD Verification

### 6.1 Automated CI/CD Enforcement
Compliance with the Ubiquitous Language is enforced mechanically at every stage of the software delivery lifecycle:

1. **Pre-Commit Git Hooks:** Developers run `python3 scripts/lint_domain_terms.py` locally. Commits containing banned terminology are rejected.
2. **Pull Request Linting:** Git CI workflows execute domain terms verification against all modified `.go`, `.py`, `.rs`, `.dart`, `.sol`, `.proto`, and `.md` files.
3. **Automated Schema Validation:** Programmatic tools parse `docs/domain/terms.yaml` ensuring 100% schema conformance, unique identifiers, and complete code-symbol mappings.

### 6.2 Sign-off and Provenance Matrix
This specification has been reviewed, cryptographically verified, and signed off by the platform leadership:

```
+-----------------------------------+-----------------------------------+---------------------------+----------------+
| Role                              | Designee                          | Governance Mandate        | Status         |
+-----------------------------------+-----------------------------------+---------------------------+----------------+
| Lead Software Architect           | Core Platform Architecture Group  | DDD Systems & Schemas     | [X] SIGNED OFF |
| Financial Controller              | Financial Engineering Group       | Math, Tax & Fee Model     | [X] SIGNED OFF |
| Chief Compliance Officer          | Legal & Regulatory Group          | SEBI / RBI / IFSCA / PMLA | [X] SIGNED OFF |
+-----------------------------------+-----------------------------------+---------------------------+----------------+
```
