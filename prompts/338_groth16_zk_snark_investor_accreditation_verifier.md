# 338 - Groth16 ZK-SNARK Investor Accreditation & Jurisdictional Verifier Contract (Solidity / Circom)

## Purpose
In institutional digital asset ecosystems, primary token issuance, and secondary market trading of private placement securities, tokenized Alternative Investment Funds (SEBI AIF Category I/II/III), Real Estate Investment Trusts (REITs), Infrastructure Investment Trusts (InvITs), and GIFT City / IFSCA cross-border instruments, statutory regulations mandate strict accredited investor verification and jurisdictional gating. 

Under the Securities and Exchange Board of India (SEBI) Accredited Investor framework (SEBI/HO/IMD/IMD-I/DF9/P/CIR/2021/620) and the International Financial Services Centres Authority (IFSCA Fund Management) Regulations 2022, participants must satisfy stringent eligibility criteria:
- **SEBI Accredited Investor (Domestic India):** Individual / HUF / Family Trust possessing an annual income $\ge \text{INR 2 Crore}$, or net worth $\ge \text{INR 7.5 Crore}$ (with at least $\text{INR 3.75 Crore}$ in financial assets, excluding primary residence), or annual income $\ge \text{INR 1 Crore}$ combined with net worth $\ge \text{INR 5 Crore}$ (with at least $\text{INR 2.5 Crore}$ in financial assets); Body Corporates possessing a minimum net worth $\ge \text{INR 50 Crore}$; Large Value Accredited Investors (LVAI) committing $\ge \text{INR 25 Crore}$ to AIF funds.
- **IFSCA Accredited Investor (GIFT City International Financial Services Centre):** Individual accredited investors possessing a liquid net worth $\ge \text{USD 1,000,000}$ (excluding primary residence); Institutional entities possessing net investable assets $\ge \text{USD 5,000,000}$.
- **FATF Jurisdictional Compliance:** Strict exclusion of investors originating from Financial Action Task Force (FATF) High-Risk ("Blacklist") or Monitored ("Greylist") jurisdictions, as well as OFAC/UN/RBI sanctions lists.

Conventional accreditation workflows necessitate transmitting unencrypted financial records (chartered accountant net-worth certificates, audited balance sheets, IT returns, Demat account holdings, and bank statements) to multiple issuers, brokerages, and on-chain relayers. This model creates severe data privacy vulnerabilities, exposes investors to identity theft, violates India's **Digital Personal Data Protection Act (DPDP Act 2023)**, and exposes trading desks to competitive front-running when net-worth tiers are leaked on public or consortium ledgers.

This prompt specifies the architecture, cryptographic circuit implementation, trusted setup ceremony pipeline, and smart contract verification suite for the **Groth16 Zero-Knowledge SNARK Investor Accreditation & Jurisdictional Verifier**. By leveraging rank-1 constraint systems (R1CS) implemented in **Circom 2.1+** and deployed on the **BN254 (alt_bn128)** elliptic curve, the Growww platform enables investors to cryptographically prove that their certified net worth exceeds statutory thresholds and that their residency is non-sanctioned and within authorized jurisdictions, with mathematical zero-knowledge privacy guarantees: zero disclosure of underlying net-worth figures, income values, or Personally Identifiable Information (PII).

---

## What You Are Building
A production-grade, privacy-preserving zero-knowledge proof generation and on-chain verification infrastructure comprising:
- **`circuits/accreditation.circom`**: A production Circom 2.1+ arithmetic circuit that validates:
  1. Investor's net-worth attribute issued by a licensed Accreditation Agency (e.g., CDSL Ventures Limited, NSDL Database Management Limited) satisfies the required tier threshold ($W \ge T_k$).
  2. The credential carries a valid BabyJubjub EdDSA digital signature issued by an authorized SEBI / IFSCA Accreditation Agency authority public key.
  3. The investor's ISO 3166-1 numeric country code belongs to the allowed jurisdiction bitmask and does not collide with FATF blacklisted jurisdictions.
  4. The accreditation certificate is valid, unexpired, and verifiable against an on-chain Credential Merkle Tree root.
  5. Deterministic derivation of an anonymized nullifier hash to prevent proof double-spending and replay attacks across trading epochs.
- **`contracts/src/compliance/ZKAccreditationVerifier.sol`**: An ultra-optimized EVM Groth16 smart contract verifier executing the pairing equation checks over the BN254 curve ($e(A, B) = e(\alpha, \beta) \cdot e(x \cdot \gamma, \delta) \cdot e(C, \delta)$) utilizing EVM precompiles (`0x06`, `0x07`, `0x08`) at sub-220,000 gas consumption.
- **`contracts/src/compliance/IZKAccreditationVerifier.sol`**: Formal Solidity interface declaring the verification method, attestation state structs, nullifier registries, custom error codes, and audit events.
- **`contracts/src/compliance/modules/ZKAccreditationModule.sol`**: Pluggable ERC-3643 / ONCHAINID modular compliance rule hooked into `ComplianceRegistry.sol` (Prompt 305) and `DigitalSecurityToken.sol` (Prompt 303) to atomically gate token minting, transfers, and secondary market settlement based on verified zero-knowledge accreditation proofs.
- **`services/zk-prover-daemon/`**: Client-side / edge daemon utilizing `snarkjs` and WebAssembly witness generators to produce Groth16 proofs without leaking private inputs outside the investor's sovereign custody.
- **Comprehensive Foundry & SnarkJS Test Suite (`test/compliance/ZKAccreditationVerifier.t.sol`)**: Unit, differential, and fuzz tests confirming mathematical soundness, zero-knowledge leakage resistance, under-constrained signal prevention, and gas bounds.

---

## Scope Boundaries
- **In Scope:**
  - Design, constraint optimization, and compilation of `circuits/accreditation.circom` with Circom 2.1+.
  - Implementation of BabyJubjub EdDSA signature verification, Poseidon hashing, bitwise range comparators, and Merkle membership checks in R1CS.
  - Multi-party computation (MPC) Powers of Tau Phase 1 and Circuit-specific Phase 2 trusted setup orchestration.
  - Development of `contracts/src/compliance/ZKAccreditationVerifier.sol` supporting BN254 pairing precompiles.
  - Integration with ERC-3643 transfer hooks (`canTransfer`) and `IdentityRegistry.sol`.
  - Nullifier registry management preventing cross-epoch proof replay.
  - Integration with KYC/AML Service (Prompt 202) for issuing signed credential attestations.
  - Foundry automated test coverage including negative proofs, expired certificates, invalid signatures, tampered public inputs, and replay vectors.
- **Out of Scope / Handled Elsewhere:**
  - Physical inspection and manual Chartered Accountant auditing of investor income/tax returns (handled off-chain by SEBI-registered Accreditation Agencies).
  - Off-chain KYC document OCR and government database API connectivity (handled in Prompt 202).
  - High-level investor portfolio tracking and trade execution (handled in Prompt 208 and Prompt 209).
  - Custodian depository physical share immobilisation (handled in Prompt 213).
  - Platform-wide automated tax withholding (handled in Prompt 336).

---

## Technology to Use
- **Zero-Knowledge Circuit DSL:** **Circom 2.1+** (utilizing `circomlib` v2.0+ standard libraries for `Poseidon`, `BabyJubjub`, `EdDSABabyJubjubVerifier`, `Bitify`, and `Comparators`).
  *Justification:* Circom offers strict compile-time constraint enforcement, minimal R1CS constraint generation, and high ecosystem maturity for Groth16 proving on EVM-compatible chains.
- **Prover & Setup Tooling:** **SnarkJS (v0.7+)** and **RapidSNARK (C++/CUDA)** for accelerated witness generation, proving key generation, and proof serialization.
- **Cryptographic Elliptic Curve:** **BN254 (alt_bn128 / bn128)** curve with order $r = 21888242871839275222246405745257275088548364400416034343698204186575808495617$.
  *Justification:* Supported natively by EVM precompiles at fixed gas costs (`ecAdd` = 150 gas, `ecMul` = 6,000 gas, `ecPairing` = 34,000 + 34,000 per pair).
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai / Cancun with Yul assembly optimization for precompile invocations).
- **Smart Contract Standards:** **ERC-3643 (T-REX Modular Compliance)**, **ERC-173 (Contract Ownership)**, **OpenZeppelin Contracts v5.0** (`AccessControl`, `ReentrancyGuard`, `Pausable`).
- **Testing & Verification:** **Foundry (`forge`, `cast`)**, **Ecne** (constraint analyzer), and **Picus** (under-constrained signal detector).

---

## Backend / Infra Touchpoints
- **KYC/AML Service (Prompt 202):** Ingests accreditation certificates issued by SEBI-approved Accreditation Agencies (CDSL/NSDL), extracts certified net-worth and jurisdiction attributes, signs the cryptographic claim using a BabyJubjub private key held in an HSM, and delivers the private witness package securely to the investor's client application.
- **ZK Proof Generator Daemon:** Runs client-side (in the investor's browser or mobile application using WebAssembly / `snarkjs`) or within a secure hardware enclave (TEE) to compile the private witness and compute the Groth16 proof $(A, B, C)$ without exposing plain financial values to Growww servers.
- **PostgreSQL Database:** Maintains an audit trail of public attestation transactions, credential Merkle roots, revocation lists, and anonymous nullifier commitments. Explicitly stores zero raw financial figures, net worth values, or plaintext PII in adherence to the DPDP Act 2023.
- **Identity Registry & Compliance Router (Prompt 305):** Ingests successful on-chain accreditation proofs and updates investor status flags in `IdentityRegistry.sol` to unlock trading permissions for private placement and AIF security tokens.
- **Chain Event Indexer (Prompt 309):** Indexes `AccreditationVerified`, `AccreditationRevoked`, and `NullifierSpent` events for compliance telemetry and administrative oversight.

---

## Blockchain Interaction
- **Network Execution:** Executes on the Hyperledger Besu permissioned consortium ledger operating with QBFT consensus, 2-second block finality, and zero gas volatility.
- **On-Chain Proof Verification Function:**
  ```solidity
  function verifyProof(
      uint256[2] calldata a,
      uint256[2][2] calldata b,
      uint256[2] calldata c,
      uint256[4] calldata input
  ) external view returns (bool r);
  ```
- **Public Inputs Array (`uint256[4]`):**
  1. `input[0]`: `merkleRoot` - Merkle root of the active Accreditation Credential Tree maintained by authorized agencies.
  2. `input[1]`: `nullifierHash` - Unique deterministic hash derived from investor secret and epoch scope: $\text{nullifierHash} = \text{Poseidon}(\text{identitySecret}, \text{scopeId}, \text{epoch})$, guaranteeing proof uniqueness without identity correlation.
  3. `input[2]`: `accreditationTier` - Target accreditation category being asserted (e.g., Tier 1 = Domestic INR 7.5 Cr, Tier 2 = Domestic INR 25 Cr LVAI, Tier 3 = GIFT City USD 1M Individual, Tier 4 = GIFT City USD 5M Corporate).
  4. `input[3]`: `jurisdictionBitmask` - Bitmask or hash commitment specifying authorized ISO 3166-1 numeric country codes for the target security offering.
- **ERC-3643 Integration Hook:**
  `ZKAccreditationModule.sol` implements `IModularCompliance`. During token transfer execution:
  ```solidity
  function canTransfer(address from, address to, uint256 value) external view returns (bool) {
      // Validates that the recipient 'to' holds an active, unexpired, non-revoked ZK accreditation attestation
      return zkVerifier.hasValidAccreditation(to, requiredTierForToken);
  }
  ```
- **Nullifier Tracking & Replay Prevention:**
  The verifier contract checks `mapping(bytes32 => bool) public isNullifierSpent;`. If the submitted `nullifierHash` has already been recorded for the current accreditation epoch, the transaction immediately reverts with `NullifierAlreadySpent(bytes32)`.

---

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Repository Structure:**
   - Create directories: `circuits/`, `circuits/test/`, `contracts/src/compliance/`, `contracts/src/interfaces/compliance/`, `test/compliance/`, and `services/zk-prover-daemon/`.
   - Install dependencies: `npm install -D circomlib snarkjs`, configure `foundry.toml` with EVM version `cancun` and Yul optimizer settings.
2. **Implement R1CS Accreditation Circuit (`circuits/accreditation.circom`):**
   - Declare private signals: `identitySecret`, `netWorthAmount`, `netWorthCurrency`, `accreditationExpiryTimestamp`, `investorCountryCode`, `issuerPubKeyX`, `issuerPubKeyY`, `sigR8x`, `sigR8y`, `sigS`, and `merklePathElements[DEPTH]`.
   - Declare public signals: `merkleRoot`, `nullifierHash`, `accreditationTier`, and `jurisdictionBitmask`.
   - Enforce range check: Prove $W \ge T_k$ using `GreaterEqThan(64)` where threshold $T_k$ is derived from `accreditationTier`.
   - Enforce credential validity: Prove $\text{accreditationExpiryTimestamp} > \text{currentTimestamp}$ using `GreaterThan(64)`.
   - Enforce BabyJubjub EdDSA signature: Verify signature over `Poseidon(identityCommitment, netWorthAmount, netWorthCurrency, accreditationExpiryTimestamp, investorCountryCode)` against known authorized issuer public keys.
   - Enforce Merkle tree inclusion: Verify leaf commitment against `merkleRoot` using `Poseidon` hash tree of depth 20.
   - Enforce nullifier derivation: Verify `nullifierHash == Poseidon(identitySecret, scopeId, epoch)`.
3. **Analyze and Audit Constraints:**
   - Compile circuit using `circom --r1cs --wasm --sym circuits/accreditation.circom -o build/`.
   - Run **Ecne** and **Picus** formal constraint analysis to ensure all signals are strictly constrained with zero degree-of-freedom flaws.
4. **Execute Multi-Party Trusted Setup Ceremony:**
   - Download or generate Perpetual Powers of Tau Phase 1 transcript (`powersOfTau28_hez_final_16.ptau` for $\le 65,536$ constraints).
   - Perform Circuit-Specific Phase 2 setup using `snarkjs groth16 setup` to generate `accreditation_0000.zkey`.
   - Contribute entropy with multiple independent key holders; export final verification key (`accreditation_final.zkey` and `verification_key.json`).
5. **Generate Solidity Verifier Contract:**
   - Execute `snarkjs zkey export solidityverifier build/accreditation_final.zkey contracts/src/compliance/ZKAccreditationVerifierGenerated.sol`.
   - Refactor generated contract into production-grade `contracts/src/compliance/ZKAccreditationVerifier.sol` supporting modular interfaces, custom errors, access control, and stateful nullifier tracking.
6. **Define Interface Specifications:**
   - Create `contracts/src/interfaces/compliance/IZKAccreditationVerifier.sol` defining the `AccreditationAttestation` struct, `verifyProof`, `registerAccreditation`, `isAccredited`, and event definitions.
7. **Implement State Management & Attestation Registry:**
   - In `ZKAccreditationVerifier.sol`, implement state variables:
     - `mapping(address => AccreditationAttestation) private _attestations;`
     - `mapping(bytes32 => bool) public isNullifierSpent;`
     - `mapping(bytes32 => bool) public isMerkleRootValid;`
     - `mapping(address => bool) public isAuthorizedAgencyKey;`
   - Implement `addCredentialRoot(bytes32 root, uint256 validityWindow)` restricted to `COMPLIANCE_ADMIN_ROLE`.
8. **Implement Public Attestation Function (`registerAccreditation`):**
   - Verify submitted Groth16 proof coordinates $(a, b, c)$ against public inputs `[merkleRoot, nullifierHash, tier, jurisdictionBitmask]`.
   - Verify `isMerkleRootValid[merkleRoot] == true`.
   - Verify `!isNullifierSpent[nullifierHash]`.
   - Mark `isNullifierSpent[nullifierHash] = true`.
   - Store attestation bound to `msg.sender` (or target investor address) with valid expiration timestamp.
   - Emit `AccreditationRegistered(address indexed investor, uint8 tier, bytes32 nullifierHash, uint64 expiryTimestamp)`.
9. **Implement ERC-3643 Compliance Module (`ZKAccreditationModule.sol`):**
   - Implement `IModularCompliance` interface.
   - Bind `ZKAccreditationModule` to `ComplianceRegistry.sol` for tokenized securities requiring accredited investor status.
   - In `moduleCheck(address from, address to, uint256 value)`, verify `zkVerifier.isAccredited(to, requiredTier)`.
10. **Develop Prover Service Daemon (`services/zk-prover-daemon/`):**
    - Implement a lightweight Node.js / Rust daemon wrapping `snarkjs` and native witness calculation.
    - Expose REST / gRPC endpoint for the investor client to submit encrypted credentials, compute witness, and generate proof payload $(a, b, c, \text{inputs})$.
11. **Write Unit & Functional Tests in Foundry (`test/compliance/ZKAccreditationVerifier.t.sol`):**
    - Test valid Groth16 proof submission and attestation storage.
    - Test invalid proof rejection (mutated point coordinates, malformed $b$ point).
    - Test tampered public inputs (e.g. attempting to claim Tier 2 with Tier 1 proof).
    - Test nullifier double-spending reverts with `NullifierAlreadySpent`.
    - Test expired credential root rejection.
12. **Write Fuzzing & Invariant Tests:**
    - Perform invariant fuzzing on pairing checks with random elliptic curve points.
    - Assert that no unverified address can attain accredited status under any state transition.
13. **Optimize Gas Consumption:**
    - Benchmark pairing checks; ensure verification gas remains strictly under 220,000 gas.
    - Utilize calldata slicing and minimal memory allocation during precompile calls.
14. **Perform End-to-End Transfer Simulation:**
    - Deploy `ZKAccreditationVerifier`, `ComplianceRegistry`, and `DigitalSecurityToken` on a local Besu devnet.
    - Simulate an AIF token transfer; verify transfer succeeds when recipient submits valid ZK proof and reverts when unaccredited.
15. **Document Verification Workflows & Security Playbooks:**
    - Publish circuit specifications, ceremony transcripts, verification key hashes, and operational procedures for quarterly credential root rollovers.

---

## Interfaces / Contracts

### Circom Signal Declarations (`circuits/accreditation.circom`)
```circom
pragma circom 2.1.6;

include "poseidon.circom";
include "eddsaposeidonverifier.circom";
include "comparators.circom";
include "bitify.circom";

template AccreditationVerifier(TREE_DEPTH) {
    // -------------------------------------------------------------
    // 1. PUBLIC SIGNALS
    // -------------------------------------------------------------
    signal input merkleRoot;                     // Root of certified investor credential tree
    signal input nullifierHash;                   // Deterministic anti-replay nullifier
    signal input accreditationTier;              // Required tier (1=Domestic Retail, 2=LVAI, 3=GIFT Indiv, 4=GIFT Corp)
    signal input jurisdictionBitmask;            // Allowed ISO 3166-1 country bitmask

    // -------------------------------------------------------------
    // 2. PRIVATE SIGNALS
    // -------------------------------------------------------------
    signal input identitySecret;                 // Investor private secret / nullifier preimage
    signal input netWorthAmount;                 // Certified net worth in basic units (paise / cents)
    signal input netWorthCurrency;               // ISO 4217 numeric (356 = INR, 840 = USD)
    signal input accreditationExpiryTimestamp;   // Expiration UNIX timestamp
    signal input investorCountryCode;            // Investor residency ISO 3166-1 numeric
    signal input currentTimestamp;               // Epoch timestamp at proof generation
    signal input scopeId;                        // Application / token compliance scope ID
    signal input epochId;                        // Epoch identifier for nullifier scope

    // Agency Signature Verification Signals (BabyJubjub EdDSA)
    signal input issuerPubKeyAx;                 // Agency BabyJubjub Public Key X
    signal input issuerPubKeyAy;                 // Agency BabyJubjub Public Key Y
    signal input sigR8x;                         // EdDSA Signature R8 Point X
    signal input sigR8y;                         // EdDSA Signature R8 Point Y
    signal input sigS;                           // EdDSA Signature S scalar

    // Merkle Inclusion Signals
    signal input merklePathElements[TREE_DEPTH]; // Sibling hashes in credential tree
    signal input merklePathIndices[TREE_DEPTH];  // Path directions (0 = left, 1 = right)

    // -------------------------------------------------------------
    // 3. CONSTRAINTS & VERIFICATION LOGIC
    // -------------------------------------------------------------

    // A. Verify Accreditation Not Expired
    component compExpiry = GreaterThan(64);
    compExpiry.in[0] <== accreditationExpiryTimestamp;
    compExpiry.in[1] <== currentTimestamp;
    compExpiry.out === 1;

    // B. Verify Net Worth Threshold by Tier
    // Tier 1 (Domestic INR 7.5 Cr = 750,000,000,000 paise)
    // Tier 2 (Domestic LVAI INR 25 Cr = 2,500,000,000,000 paise)
    // Tier 3 (GIFT City Individual USD 1M = 100,000,000 cents)
    // Tier 4 (GIFT City Institutional USD 5M = 500,000,000 cents)
    signal requiredThreshold;
    component tierSelector = EscalarMulAny(4);
    // [Threshold resolution logic dynamically mapping accreditationTier to requiredThreshold]
    component compThreshold = GreaterEqThan(64);
    compThreshold.in[0] <== netWorthAmount;
    compThreshold.in[1] <== requiredThreshold;
    compThreshold.out === 1;

    // C. Verify Jurisdictional Compliance
    // Ensures investorCountryCode bit is active within jurisdictionBitmask
    component countryBits = Num2Bits(256);
    countryBits.in <== jurisdictionBitmask;
    // [Bitwise selection verifying countryBits[investorCountryCode] === 1]

    // D. Verify Credential Attestation Hash & EdDSA Signature
    component credentialHasher = Poseidon(5);
    credentialHasher.inputs[0] <== identitySecret;
    credentialHasher.inputs[1] <== netWorthAmount;
    credentialHasher.inputs[2] <== netWorthCurrency;
    credentialHasher.inputs[3] <== accreditationExpiryTimestamp;
    credentialHasher.inputs[4] <== investorCountryCode;

    component eddsaVerifier = EdDSAPoseidonVerifier();
    eddsaVerifier.enabled <== 1;
    eddsaVerifier.Ax <== issuerPubKeyAx;
    eddsaVerifier.Ay <== issuerPubKeyAy;
    eddsaVerifier.R8x <== sigR8x;
    eddsaVerifier.R8y <== sigR8y;
    eddsaVerifier.S <== sigS;
    eddsaVerifier.M <== credentialHasher.out;

    // E. Verify Merkle Tree Membership of Credential
    // [Sequential Poseidon tree verification from leaf commitment to merkleRoot]

    // F. Enforce Nullifier Derivation
    component nullifierHasher = Poseidon(3);
    nullifierHasher.inputs[0] <== identitySecret;
    nullifierHasher.inputs[1] <== scopeId;
    nullifierHasher.inputs[2] <== epochId;
    nullifierHasher.out === nullifierHash;
}

component main {public [merkleRoot, nullifierHash, accreditationTier, jurisdictionBitmask]} = AccreditationVerifier(20);
```

---

### Solidity Verifier Interface (`IZKAccreditationVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IZKAccreditationVerifier
 * @notice Interface for Groth16 Zero-Knowledge Verification of Investor Accreditation & Jurisdiction
 * @dev Enforces SEBI & IFSCA accredited investor rules on-chain with zero-knowledge data privacy
 */
interface IZKAccreditationVerifier {

    enum AccreditationTier {
        NONE,
        SEBI_RETAIL_ACCREDITED,       // Net worth >= INR 7.5 Cr / Income >= INR 2 Cr
        SEBI_LARGE_VALUE_ACCREDITED,  // Minimum commitment >= INR 25 Cr (AIF LVAI)
        IFSCA_INDIVIDUAL_ACCREDITED,  // Net worth >= USD 1,000,000 (excluding primary residence)
        IFSCA_INSTITUTIONAL_ACCREDITED // Net assets >= USD 5,000,000
    }

    struct AccreditationAttestation {
        AccreditationTier tier;
        uint64 verifiedAt;
        uint64 expiresAt;
        bytes32 nullifierHash;
        bytes32 merkleRootUsed;
        bool isValid;
    }

    struct ProofPoints {
        uint256[2] a;
        uint256[2][2] b;
        uint256[2] c;
    }

    // Events
    event AccreditationRegistered(
        address indexed investor,
        AccreditationTier indexed tier,
        bytes32 indexed nullifierHash,
        uint64 expiresAt,
        bytes32 merkleRoot
    );

    event AccreditationRevoked(address indexed investor, bytes32 indexed nullifierHash, string reason);
    event CredentialRootAdded(bytes32 indexed root, uint256 validUntil);
    event CredentialRootRevoked(bytes32 indexed root);
    event AgencyKeyWhitelisted(address indexed agencyPubKey, string agencyName);
    event AgencyKeyRemoved(address indexed agencyPubKey);

    // Custom Errors
    error InvalidProof();
    error NullifierAlreadySpent(bytes32 nullifierHash);
    error InvalidCredentialRoot(bytes32 merkleRoot);
    error RootExpired(bytes32 merkleRoot, uint256 expirationTime);
    error TierMismatch(uint8 requestedTier, uint8 provedTier);
    error AttestationExpired(address investor, uint64 expiresAt);
    error AttestationNotFound(address investor);
    error UnauthorizedCaller(address caller);
    error ZeroAddress();

    /**
     * @notice Verifies a raw Groth16 zero-knowledge proof against the BN254 pairing engine
     * @param a Point A in G1
     * @param b Point B in G2
     * @param c Point C in G1
     * @param input Public inputs array [merkleRoot, nullifierHash, accreditationTier, jurisdictionBitmask]
     * @return success True if the proof is cryptographically sound
     */
    function verifyProof(
        uint256[2] calldata a,
        uint256[2][2] calldata b,
        uint256[2] calldata c,
        uint256[4] calldata input
    ) external view returns (bool success);

    /**
     * @notice Verifies a ZK proof and atomically registers an accreditation attestation for the sender
     * @param proof Cryptographic Groth16 proof points (A, B, C)
     * @param publicInputs Public signals [merkleRoot, nullifierHash, tier, jurisdictionBitmask]
     * @param attestationDuration Duration in seconds for which the on-chain attestation remains valid
     * @return success True if verification and attestation succeed
     */
    function registerAccreditation(
        ProofPoints calldata proof,
        uint256[4] calldata publicInputs,
        uint64 attestationDuration
    ) external returns (bool success);

    /**
     * @notice Checks whether an investor possesses a valid, unexpired accreditation of at least the required tier
     * @param investor Address of the investor to check
     * @param requiredTier Minimum accreditation tier demanded by the security token
     * @return isEligible True if investor meets or exceeds the required tier
     */
    function isAccredited(address investor, AccreditationTier requiredTier) external view returns (bool isEligible);

    /**
     * @notice Retrieves the active attestation details for an investor
     * @param investor Address of the investor
     * @return attestation Full attestation data struct
     */
    function getAttestation(address investor) external view returns (AccreditationAttestation memory attestation);

    /**
     * @notice Checks if a nullifier hash has already been spent
     * @param nullifierHash The unique hash to query
     * @return spent True if already consumed
     */
    function isNullifierSpent(bytes32 nullifierHash) external view returns (bool spent);
}
```

---

## Security & Compliance Notes
- **Zero Financial Data Disclosure (DPDP Act 2023):** Under no circumstance is raw investor net worth, taxable income, demat balance, bank statement, or PAN/Aadhaar number written to the blockchain or transmitted through node mempools. All values remain encapsulated within the private witness locally generated in the user's secure runtime.
- **Trusted Setup & Toxic Waste Mitigation:**
  - Groth16 requires a two-phase trusted setup. Phase 1 utilizes the publicly audited, multi-thousand-participant Hermez Perpetual Powers of Tau ceremony.
  - Phase 2 (circuit-specific) must be conducted across a minimum of 10 independent participating institutions (custodians, brokerages, legal counsel, and SEBI/IFSCA auditors) using air-gapped cryptographic hardware modules (Prompt 311).
  - The toxic waste ($\tau, \alpha, \beta, \gamma, \delta$) is mathematically discarded upon ceremony completion, preventing any party from forging accreditation proofs.
- **Proof Replay & Nullifier Integrity:**
  - A unique `nullifierHash = Poseidon(identitySecret, scopeId, epoch)` is derived inside the arithmetic circuit and exposed as a public input.
  - Smart contracts permanently store used nullifiers. A single zero-knowledge proof cannot be submitted twice, and proof payloads intercepted from node RPC channels cannot be re-executed by adversarial front-runners.
- **BN254 Subgroup Checks & Pairing Precompiles:**
  - To defend against small-subgroup attacks, the verifier explicitly validates that points $A \in G_1$, $B \in G_2$, and $C \in G_1$ reside on the prime-order subgroup of BN254.
  - Elliptic curve operations strictly invoke native EVM precompiles (`0x06` for point addition, `0x07` for scalar multiplication, `0x08` for bilinear pairing checks) using memory-safe Yul assembly.
- **Accreditation Agency Revocation Window:**
  - Accredited investor status certificates under SEBI guidelines are typically valid for a maximum of 3 years (or 1 year depending on category).
  - The circuit enforces $\text{accreditationExpiryTimestamp} > \text{currentTimestamp}$.
  - The contract maintains an active `merkleRoot` whitelist with an expiry window; invalidated or revoked credential trees can be purged instantly by compliance administrators.
- **Under-Constrained Signal Vulnerability Prevention:**
  - Every circuit variable must be tied to an active R1CS linear combination.
  - Circuit compilation requires strict verification with `picus` and `ecne` to guarantee zero unconstrained output signals or under-constrained rank regressions.

---

## Acceptance Criteria
- [ ] **Circuit Compilation:** `circuits/accreditation.circom` compiles cleanly under Circom 2.1+ producing fewer than 45,000 R1CS constraints with zero unconstrained signal warnings.
- [ ] **Constraint Soundness Audit:** Automated static verification with **Ecne** and **Picus** reports zero degree-of-freedom flaws and confirms complete input-output determinism.
- [ ] **Precompile Verification Efficiency:** `ZKAccreditationVerifier.sol` deploys cleanly to Hyperledger Besu with EVM Cancun compatibility and executes `verifyProof` in $\le 220,000$ gas.
- [ ] **Accreditation Gating:** Valid Groth16 proofs for SEBI Tier 1, SEBI LVAI Tier 2, IFSCA Tier 3, and IFSCA Tier 4 pass verification and record attestations accurately.
- [ ] **Tamper Resistance:** Any mutation of public inputs (e.g. altering `accreditationTier` from 1 to 2 without altering private net-worth witness) or mutation of proof curve coordinates results in immediate revert (`InvalidProof()`).
- [ ] **Replay Resistance:** Submitting a previously used `nullifierHash` immediately reverts with `NullifierAlreadySpent`.
- [ ] **Expiration Enforcement:** Submitting a proof containing an expired `accreditationExpiryTimestamp` fails circuit witness generation or reverts during contract attestation.
- [ ] **ERC-3643 Transfer Gating:** `ZKAccreditationModule.sol` successfully prevents non-accredited addresses from receiving restricted security tokens and permits transfers immediately upon on-chain attestation registration.
- [ ] **Zero-Knowledge Privacy Verification:** Memory traces, RPC transaction logs, and on-chain storage contain zero plaintext PII, bank balances, or exact net-worth figures.

---

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `202` (KYC/AML Service - manages identity onboarding and credential issuance).
  - Prompt `303` (Token Issuance Smart Contract - defines security token properties).
  - Prompt `305` (Transfer Compliance Hooks Smart Contract - provides ERC-3643 `ComplianceRegistry` and `IdentityRegistry` base architecture).
- **Parallel Tasks:**
  - Prompt `317` (ZK Proof-of-Solvency Verifier - shares BN254 curve utilities and trusted setup ceremony infrastructure).
  - Prompt `311` (Validator Key Management HSM - manages ceremony entropy generation and authorized agency keys).
- **Subsequent Prompts Enabled:**
  - Prompt `331` (Primary Market Order Routing & Clearing Bridge - routes accredited investor orders for private placement offerings).
  - Prompt `333` (Tokenized G-Sec Bonds & Corporate Debt Contract - enforces accredited investor thresholds for institutional tranches).
  - Prompt `604` (Compliance Officer Portal - displays anonymized accreditation telemetry and root lifecycle management).
