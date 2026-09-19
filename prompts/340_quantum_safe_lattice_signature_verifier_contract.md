# 340 - Post-Quantum Lattice Signature (ML-DSA / Dilithium) Verifier Contract (Solidity / Yul)

## Purpose
The global financial system and institutional capital markets face an existential cryptographic threat from the advent of cryptographically relevant quantum computers (CRQCs). Shor's algorithm running on a sufficiently scaled quantum computer will solve the discrete logarithm problem and factorize large integers in polynomial time. This completely breaks the foundational asymmetric cryptographic algorithms underpinning existing blockchain systems and institutional digital banking, including Elliptic Curve Digital Signature Algorithm (ECDSA secp256k1), Edwards-curve DSA (Ed25519), RSA, and pairing-friendly BLS12-381 curves. Furthermore, Grover's quantum search algorithm reduces symmetric key and hash collision security bounds by a quadratic factor.

In capital market infrastructures, financial obligations possess lifecycles spanning decades. Sovereign government bonds (G-Secs, Prompt 333) mature over 10 to 40 years. Multi-billion rupee equity clearing settlement batches (Prompt 306 and Prompt 329), foreign institutional investor asset transfers (Prompt 214), and wholesale central bank digital currency (eINR-W) bridge commitments (Prompt 334) carry legal finality that must remain immutably protected against retroactive forgery ("Harvest Now, Decrypt / Forge Later" attacks). If an adversary captures historical settlement messages or public keys today, a future quantum computer could forge authorized signatures, drain custodial reserves, or falsify depository ownership records.

To establish future-proof cryptographic resilience, the National Institute of Standards and Technology (NIST) standardized the Module-Lattice-Based Digital Signature Algorithm (ML-DSA) under Federal Information Processing Standard (FIPS) 204 (formerly known as CRYSTALS-Dilithium). ML-DSA bases its mathematical security on the hardness of finding short vectors in module lattices over polynomial rings (specifically the Module Learning With Errors - M-LWE, and Module Short Integer Solution - M-SIS problems). These lattice problems have no known sub-exponential quantum or classical algorithms.

This prompt specifies the **Post-Quantum Lattice Signature Verifier Contract Suite (`MLDSAVerifier.sol`, `IPostQuantumVerifier.sol`, `LatticeMathYul.sol`, `DualSignedSettlementHook.sol`)**. The smart contracts implement highly optimized on-chain verification for NIST FIPS 204 ML-DSA-44 (Security Category 2, equivalent to Dilithium2) on Hyperledger Besu using Solidity ^0.8.24 and inline Yul assembly. The system introduces an institutional Dual-Signature Authorization Protocol (combining ECDSA secp256k1 and ML-DSA-44), ensuring that high-value settlement batches, sovereign bond issuances, and custodial vault withdrawals require mathematically orthogonal authorizations that remain uncompromised even if either classical elliptic curves or specific lattice parameters are broken. The contracts integrate seamlessly with Growww's universal 0.00% (No fee at all) (0.00% fee / 0 bps at launch) transaction fee model with a 0.00% fee at launch (governed by FeeController.sol) revenue split, preserving strict zero-PII privacy standards.

## What You Are Building
A production-grade, highly optimized smart contract suite under `contracts/src/crypto/` and `contracts/src/settlement/` comprising:
- `MLDSAVerifier.sol`: Master on-chain post-quantum lattice signature verifier contract implementing NIST FIPS 204 ML-DSA-44. Written in Solidity ^0.8.24 with critical algorithmic inner loops implemented in inline Yul assembly (`assembly ("memory-safe")`). It performs polynomial ring arithmetic in $R_q = \mathbb{Z}_q[X]/(X^{256} + 1)$ with modulus $q = 8380417$, executes forward and inverse Number Theoretic Transforms (NTT / INTT), expands matrix $A \in R_q^{4 \times 4}$ from seed $\rho$ via SHAKE-128 / Keccak sponge permutations, validates infinity norms $\|z\|_\infty < \gamma_1 - \beta$, decodes hint vectors, reconstructs high-order bits $w'_1$, and computes challenge hash matches.
- `IPostQuantumVerifier.sol`: Canonical Solidity interface declaring all ML-DSA-44 parameters, data structures (`LatticePublicKey`, `LatticeSignature`, `DualSignatureBatch`, `VerificationResult`), custom errors, events, and verification function signatures.
- `LatticeMathYul.sol`: High-performance Yul assembly math library encapsulating constant-time Montgomery modular arithmetic, Cooley-Tukey radix-2 NTT butterfly stages, Gentleman-Sande INTT stages, polynomial pointwise multiplication, coefficient packing/unpacking, and Keccak-f[1600] / SHAKE sponge absorption and squeezing routines.
- `DualSignedSettlementHook.sol`: Institutional compliance and settlement authorization hook contract. It enforces dual-signature verification (ECDSA secp256k1 + ML-DSA-44) on batch settlement transactions dispatched from the Multichain MPC-TSS Vault Service (Prompt 237) and CloudHSM signing daemons (Prompt 717) before clearing trades in `SettlementDvP.sol` (Prompt 306).
- Comprehensive Foundry Test & Formal Verification Harness (`test/crypto/`):
  - `MLDSAVerifier.t.sol`: Exhaustive test vectors derived directly from NIST FIPS 204 Known Answer Tests (KAT) validating valid signature verification, rejected corrupted signatures, out-of-bounds polynomial coefficients, forged hint bits, and malformed encodings.
  - `LatticeMathYul.t.sol`: Unit tests for NTT/INTT mathematical invertibility, Montgomery reduction accuracy, and constant-time behavior across all 256 coefficients.
  - `MLDSAInvariants.t.sol`: Invariant fuzz tests asserting Yul memory safety (zero memory pointer corruptions beyond designated scratch space) and bounded EVM gas consumption.

## Scope Boundaries
- **In Scope:**
  - Complete on-chain verification of NIST FIPS 204 ML-DSA-44 (Dilithium2) digital signatures in Solidity and inline Yul.
  - Ring arithmetic over $R_q = \mathbb{Z}_q[X]/(X^{256} + 1)$ where $q = 8380417 = 2^{23} - 2^{13} + 1$.
  - Montgomery modular reduction ($R = 2^{32} \pmod q$, $q^{-1} \pmod{2^{32}}$) and constant-time modular arithmetic in Yul.
  - Forward NTT and inverse NTT (INTT) utilizing precomputed primitive 512-th roots of unity in $\mathbb{Z}_q$.
  - Public key decoding: 1312 bytes consisting of 32-byte seed $\rho$ and 1280-byte vector $t_1$ (4 polynomials with 256 10-bit coefficients).
  - Signature decoding: 2420 bytes consisting of 32-byte commitment seed $\tilde{c}$, 2304-byte response vector $z$ (4 polynomials with 256 18-bit coefficients), and 84-byte hint vector $h$.
  - Deterministic pseudo-random matrix expansion $A \in R_q^{4 \times 4}$ from 32-byte seed $\rho$ using Keccak-f[1600] / SHAKE-128 sponge routines.
  - Norm bound verification: asserting that each coefficient of vector $z$ satisfies $\|z\|_\infty < \gamma_1 - \beta$ ($131072 - 78 = 130994$).
  - Hint verification: validating that the total number of 1-bits in hint vector $h$ does not exceed $\omega = 80$, and reconstructing $w'_1 = \text{UseHint}(h, A \cdot z - c \cdot t_1 \cdot 2^d, 2\gamma_2)$ where $\gamma_2 = 95232$ and $d = 13$.
  - Challenge polynomial $c \in R_q$ generation: reconstructing sparse polynomial $c$ with exactly $\tau = 39$ coefficients in $\{-1, +1\}$ from challenge seed $\tilde{c}$ using SHAKE-256 sponge.
  - Dual-signature verification scheme: coordinating atomic validation of ECDSA secp256k1 via EVM `ecrecover` (0x01) alongside ML-DSA-44 lattice verification.
  - Institutional settlement hook integration enforcing dual-signature clearance on high-value DvP batches (Prompt 306) and tokenized G-Sec bond coupon runs (Prompt 333).
  - Memory-safe Yul execution (`assembly ("memory-safe")`) maintaining strict EVM memory pointer hygiene and zero heap corruption.
- **Out of Scope / Handled Elsewhere:**
  - Key generation and off-chain signature creation for ML-DSA-44 (executed in Rust by Multichain MPC-TSS Vault Service Prompt 237 and CloudHSM signing daemon Prompt 717).
  - Post-Quantum Key Encapsulation Mechanisms (ML-KEM / FIPS 203 / Kyber), which operate at the network transport and mTLS layers (Prompt 110).
  - ML-DSA-65 and ML-DSA-87 parameter variants (deferred to future network upgrades; ML-DSA-44 provides optimal security-to-gas efficiency for Besu).
  - Core DvP settlement execution, order matching, and custody transfers (handled in Prompt 208, Prompt 306, and Prompt 329).
  - Physical custodian depository messaging with NSDL/CDSL (handled in Prompt 213).
  - Fiat banking payment gateway and RTGS rails (handled in Prompt 212).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Cancun / Shanghai with custom errors, transient storage support, and user-defined value types).
- **Low-Level Assembly:** **Yul (`assembly ("memory-safe")`)** for performance-critical inner loops:
  - 8-stage Cooley-Tukey radix-2 NTT forward transformation.
  - 8-stage Gentleman-Sande radix-2 INTT inverse transformation.
  - Pointwise polynomial multiplication in Montgomery domain.
  - Bit-unpacking routines for public keys, signature vectors, and hint bitmasks.
  - Constant-time Montgomery reduction avoiding variable-latency division opcodes.
- **Cryptographic Standard:** **NIST FIPS 204 (ML-DSA)**, parameter set **ML-DSA-44**:
  - Ring modulus: $q = 8380417 = 2^{23} - 2^{13} + 1$ (prime satisfying $q \equiv 1 \pmod{512}$).
  - Polynomial degree: $n = 256$.
  - Matrix dimensions: $k = 4, l = 4$.
  - Dropped bits parameter: $d = 13$.
  - Response bound: $\gamma_1 = 2^{17} = 131072$.
  - Decomposition parameter: $\gamma_2 = (q - 1) / 88 = 95232$.
  - Secret key coefficient bound: $\eta = 2$.
  - Number of $+1 / -1$ coefficients in challenge polynomial $c$: $\tau = 39$.
  - Maximum weight of hint vector $h$: $\omega = 80$.
  - Bound for signature norm check: $\beta = \tau \cdot \eta = 78$; maximum valid coefficient $|z_i| < \gamma_1 - \beta = 130994$.
  - Public key size: 1312 bytes ($\rho$: 32 bytes, $t_1$: 1280 bytes).
  - Signature size: 2420 bytes ($\tilde{c}$: 32 bytes, $z$: 2304 bytes, $h$: 84 bytes).
- **Classical Cryptography:** Native EVM precompile `ecrecover` at address `0x01` for secp256k1 signature validation in dual-signed batches.
- **Hash Functions & Sponge Permutations:** Keccak-f[1600] permutation implemented in Yul for SHAKE-128 and SHAKE-256 extended-output functions (XOF).
- **Development & Testing Framework:** **Foundry** (`forge` 0.2.0+, `cast`).
- **Static Analysis & Auditing Tools:** Slither, Halmos (symbolic execution for EVM bytecode), Echidna (property-based fuzzing).

## Backend / Infra Touchpoints
- **Multichain MPC-TSS Vault Service (Prompt 237):** Computes distributed threshold ML-DSA-44 signatures and ECDSA signatures across distributed key shares, packaging dual-signed authorization payloads for on-chain submission.
- **CloudHSM Signing Daemon (Prompt 717):** Houses FIPS 140-3 Level 3 hardware security modules storing primary cold and warm institutional root keys. Executes hardware-accelerated ML-DSA-44 and ECDSA co-signing for sovereign bond auctions and emergency treasury disbursements.
- **Trade Settlement Service (Prompt 208):** Orchestrates end-of-day DvP Type 1 and Type 3 settlement batches. Attaches dual-signature cryptographic envelopes to settlement batches exceeding high-value institutional thresholds (e.g. transactions >= 10,000,000 INR).
- **Tokenized G-Sec & Bond Engine (Prompt 333):** Invokes `DualSignedSettlementHook.sol` prior to processing multi-crore coupon disbursements or principal redemptions on sovereign debt tokens.
- **Fee & Realized PnL Engine (Prompt 210):** Validates that dual-signature verification gas overhead does not disrupt the canonical 0.00% (Zero Fee) platform fee deduction (0.00% fee at launch (governed by FeeController.sol) split).
- **Blockchain Event Indexer (Prompt 309):** Ingests `SignatureVerifiedPQC`, `DualSignatureBatchValidated`, and `LatticeVerificationFailed` events, updating institutional audit logs and security monitoring dashboards.
- **Audit Log & Regulatory Reporting Service (Prompt 216 & Prompt 218):** Logs cryptographic algorithm identifiers, key identifiers, and verification latencies for SEBI, RBI, and CERT-In quantum-readiness compliance reporting.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Operates on the Growww / NBSE permissioned Hyperledger Besu consortium ledger with QBFT consensus, 2-second block intervals, and single-block deterministic finality (zero chain reorganizations).
- **Dual-Signature Verification Paradigm:** High-value institutional settlement authorizations enforce dual cryptographic checks:
  1. *Classical Authorization:* An ECDSA signature verified against an authorized custodian address registered in `MultiSigGovernance.sol` (Prompt 307).
  2. *Quantum-Safe Authorization:* An ML-DSA-44 lattice signature verified against a registered post-quantum public key identity.
  Both signatures must evaluate to true within the same atomic transaction. If either check fails, the settlement reverts with a specific custom error (`InvalidECDSASignature` or `InvalidPQCSignature`).
- **Gas Engineering & Block Gas Limits:** Hyperledger Besu private networks operate with configurable block gas limits (standard: 30,000,000 to 60,000,000 gas). By optimizing the 16 NTT operations, matrix expansions, and SHAKE permutations into memory-safe Yul assembly, ML-DSA-44 verification executes within 1,200,000 to 1,650,000 gas, enabling efficient inclusion within standard settlement blocks alongside commercial DvP transfers.
- **Zero On-Chain PII Invariant:** In strict compliance with the Digital Personal Data Protection Act (DPDPA 2023) and GDPR, no corporate names, authorized signatory identities, or physical credentials are stored on-chain. Public keys are registered and referenced solely via cryptographic key identifiers:
  $$\text{latticeKeyId} = \text{keccak256}(\text{abi.encodePacked}(\text{rawLatticePublicKey}))$$
- **Immutable Verifier & Upgradability:** `MLDSAVerifier.sol` is deployed as an immutable, stateless logic contract. The referencing settlement contracts (`DualSignedSettlementHook.sol`) maintain an upgradeable pointer to the active verifier via UUPS proxy patterns governed by a 3-of-5 multisig timelock backed by FIPS 140-3 HSMs (Prompt 307 and Prompt 311).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Layout & Tooling:**
   - Create directories `contracts/src/crypto/`, `contracts/src/crypto/interfaces/`, `contracts/src/crypto/libraries/`, and `contracts/src/settlement/`.
   - Configure `foundry.toml` with `solc_version = "0.8.24"`, EVM version `cancun`, optimizer runs `20000`, and `via_ir = true` to enable advanced Yul pipeline optimizations.
2. **Declare Cryptographic Interfaces & Types (`IPostQuantumVerifier.sol`):**
   - Declare structs: `LatticePublicKey`, `LatticeSignature`, `DualSignatureBatch`, and `VerificationResult`.
   - Define NIST FIPS 204 ML-DSA-44 constant parameters ($q, n, k, l, d, \gamma_1, \gamma_2, \tau, \beta, \omega$).
   - Define custom errors (`InvalidSignatureLength`, `InvalidPublicKeyLength`, `NormOutOfBounds`, `HintWeightExceeded`, `ChallengeMismatch`, `InvalidECDSASignature`).
   - Declare verification function prototypes for standalone ML-DSA-44 and dual-signature batches.
3. **Implement Montgomery Modular Arithmetic in Yul (`LatticeMathYul.sol`):**
   - Implement constant-time Montgomery reduction for prime $q = 8380417$.
   - Precompute Montgomery constants: $R = 2^{32} \pmod q = 4186624$, $R^2 \pmod q = 23459$, and $q^{-1} \pmod{2^{32}} = 4236238847$.
   - Implement Yul functions `montgomeryMul(a, b)`, `montgomeryAdd(a, b)`, and `montgomerySub(a, b)` ensuring zero conditional branching on secret or intermediate data.
4. **Implement Number Theoretic Transform (NTT & INTT) in Yul:**
   - Precompute table of 256 primitive 512-th roots of unity in $\mathbb{Z}_q$ in bit-reversed order (in Montgomery form).
   - Implement 8-stage Cooley-Tukey radix-2 forward NTT transforming polynomial $a \in R_q$ to NTT domain $\hat{a}$.
   - Implement 8-stage Gentleman-Sande radix-2 inverse NTT (INTT) transforming $\hat{a}$ back to standard representation, multiplying by scalar $n^{-1} \pmod q = 256^{-1} \pmod{8380417} = 8347681$.
5. **Implement Keccak-f[1600] Permutation and SHAKE Sponge in Yul:**
   - Implement the 24-round Keccak-f[1600] state permutation in optimized Yul.
   - Implement `shake128Absorb` and `shake128Squeeze` for matrix expansion (rate $r = 1344$ bits / 168 bytes).
   - Implement `shake256Absorb` and `shake256Squeeze` for challenge polynomial generation and public key hashing (rate $r = 1088$ bits / 136 bytes).
6. **Implement Pseudo-Random Matrix Expansion ($A \in R_q^{4 \times 4}$):**
   - For each matrix entry $A[i][j]$ ($0 \le i < 4, 0 \le j < 4$), absorb seed $\rho$ concatenated with 2-byte indices $(j, i)$ into SHAKE-128.
   - Squeeze bytes and reject values $\ge q$ using standard rejection sampling to produce 256 uniform coefficients in $[0, q-1]$ directly in NTT domain.
7. **Implement Public Key and Signature Unpacking Routines:**
   - Unpack public key: Extract 32-byte seed $\rho$ and decode 1280 bytes of $t_1$ into $4 \times 256$ 10-bit coefficients. Multiply $t_1$ coefficients by $2^d = 2^{13} = 8192$ and convert to Montgomery NTT domain.
   - Unpack signature: Extract 32-byte challenge seed $\tilde{c}$, unpack 2304 bytes of response vector $z$ into $4 \times 256$ signed 18-bit coefficients, and unpack 84-byte hint bitstream $h$.
8. **Implement Response Vector Norm Validation:**
   - Iterate through all $4 \times 256 = 1024$ coefficients of vector $z$.
   - Assert that each coefficient satisfies $|z_{i,j}| < \gamma_1 - \beta = 131072 - 78 = 130994$.
   - Immediately revert with `NormOutOfBounds()` if any coefficient violates the bound.
9. **Implement Hint Vector Weight Validation & Reconstruction:**
   - Iterate through hint array $h$. Count total non-zero bits.
   - Assert total number of 1-bits does not exceed $\omega = 80$. Revert with `HintWeightExceeded()` if violated.
   - Transform unpacked response vector $z$ to NTT domain $\hat{z}$ using forward NTT.
   - Compute matrix-vector product in NTT domain: $\hat{w} = A \cdot \hat{z}$.
10. **Implement Challenge Polynomial Reconstruction & Matrix Multiplication:**
    - Absorb challenge seed $\tilde{c}$ into SHAKE-256. Generate 256-bit challenge polynomial $c$ with exactly $\tau = 39$ non-zero coefficients in $\{-1, +1\}$.
    - Transform $c$ to NTT domain $\hat{c}$.
    - Compute product $\hat{c} \cdot \hat{t}_1$ in NTT domain.
    - Compute difference vector $\hat{v} = \hat{w} - \hat{c} \cdot \hat{t}_1$.
    - Apply INTT to $\hat{v}$ to recover polynomial vector $v \in R_q^4$.
11. **Implement High-Order Bit Reconstruction (`UseHint`) & Verification:**
    - For each coefficient $v_{i,j}$ and corresponding hint bit $h_{i,j}$, reconstruct high-order bits $w'_1 = \text{UseHint}(h_{i,j}, v_{i,j}, 2\gamma_2)$ where $2\gamma_2 = 190464$.
    - Pack reconstructed $w'_1$ into canonical 576-byte representation.
    - Hash message $M$ with public key hash $\mu = \text{SHAKE-256}(\text{SHAKE-256}(pk) \| M)$.
    - Compute reconstructed challenge $c' = \text{SHAKE-256}(\mu \| w'_1)$.
    - Verify that first 32 bytes of $c'$ strictly match signature challenge seed $\tilde{c}$. Return true if identical, false otherwise.
12. **Build Master Verifier Contract (`MLDSAVerifier.sol`):**
    - Integrate all Yul math modules into `MLDSAVerifier.sol` complying with `IPostQuantumVerifier.sol`.
    - Enforce strict Yul memory safety annotations (`assembly ("memory-safe")`), allocating memory strictly beyond the free memory pointer (`0x40`) and ensuring zero memory leakage.
13. **Build Dual-Signature Settlement Hook (`DualSignedSettlementHook.sol`):**
    - Implement `verifySettlementAuthorization(bytes32 batchId, bytes memory ecdsaSig, bytes memory mldsaSig, address classicalSigner, bytes32 pqcKeyId)`:
      - Verify classical authorization via `ecrecover(batchId, v, r, s) == classicalSigner`.
      - Retrieve registered ML-DSA public key corresponding to `pqcKeyId`.
      - Call `MLDSAVerifier.verifyMLDSA44(batchId, mldsaSig, latticePublicKey)`.
      - Require both checks to succeed; emit `DualSignatureBatchValidated(batchId, classicalSigner, pqcKeyId)`.
14. **Integrate Universal 0.00% transaction fee (No fee at all) Accounting:**
    - Connect dual-signed settlement batch execution with fee assessment: calculate gross turnover, assess exact 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) platform fee, and allocate 0.00% fee at launch (governed by FeeController.sol).
15. **Develop Comprehensive Test Suite & Fuzzing Harness:**
    - Import NIST FIPS 204 official Known Answer Test (KAT) vectors.
    - Write unit tests verifying valid signatures, tampered message payloads, flipped hint bits, and corrupted public keys.
    - Write invariant fuzz tests in Foundry running 10,000 iterations verifying arithmetic consistency, NTT invertibility, and zero memory corruption.
    - Execute gas benchmarking verifying signature validation completes under 1,650,000 gas.

## Interfaces / Contracts

### 1. Post-Quantum Verifier Interface (`IPostQuantumVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IPostQuantumVerifier
 * @notice Canonical interface for NIST FIPS 204 (ML-DSA / Dilithium) lattice signature verification
 *         and hybrid dual-signature (ECDSA + ML-DSA) institutional settlement authorizations on Hyperledger Besu.
 */
interface IPostQuantumVerifier {
    // -------------------------------------------------------------------------
    // Constants - NIST FIPS 204 ML-DSA-44 Parameter Definitions
    // -------------------------------------------------------------------------

    /// @notice Prime ring modulus q = 8380417 = 2^23 - 2^13 + 1
    uint32 constant ML_DSA_Q = 8380417;

    /// @notice Polynomial degree n = 256
    uint16 constant ML_DSA_N = 256;

    /// @notice Matrix rows k = 4
    uint8 constant ML_DSA_K = 4;

    /// @notice Matrix columns l = 4
    uint8 constant ML_DSA_L = 4;

    /// @notice Dropped bits from t: d = 13
    uint8 constant ML_DSA_D = 13;

    /// @notice Response vector coefficient bound gamma_1 = 2^17 = 131072
    uint32 constant ML_DSA_GAMMA1 = 131072;

    /// @notice Low-order rounding parameter gamma_2 = (q - 1) / 88 = 95232
    uint32 constant ML_DSA_GAMMA2 = 95232;

    /// @notice Number of +/- 1 coefficients in challenge polynomial c: tau = 39
    uint8 constant ML_DSA_TAU = 39;

    /// @notice Maximum signature norm bound beta = tau * eta = 39 * 2 = 78
    uint8 constant ML_DSA_BETA = 78;

    /// @notice Maximum number of 1-bits in hint vector h: omega = 80
    uint8 constant ML_DSA_OMEGA = 80;

    /// @notice Canonical byte length of ML-DSA-44 public key (rho: 32 bytes + t1: 1280 bytes)
    uint16 constant ML_DSA_44_PUBLIC_KEY_BYTES = 1312;

    /// @notice Canonical byte length of ML-DSA-44 signature (c_tilde: 32 bytes + z: 2304 bytes + h: 84 bytes)
    uint16 constant ML_DSA_44_SIGNATURE_BYTES = 2420;

    // -------------------------------------------------------------------------
    // Structs
    // -------------------------------------------------------------------------

    /// @notice Decoded ML-DSA-44 Public Key
    struct LatticePublicKey {
        bytes32 rho;                   // 32-byte seed for pseudo-random matrix A expansion
        bytes t1Encoded;               // 1280-byte encoded high-order coefficient vector t1
    }

    /// @notice Decoded ML-DSA-44 Signature
    struct LatticeSignature {
        bytes32 challengeSeed;         // 32-byte challenge seed c_tilde
        bytes zEncoded;                // 2304-byte encoded response polynomial vector z
        bytes hintEncoded;             // 84-byte encoded hint bitstream h
    }

    /// @notice Dual-Signature Settlement Batch Payload
    struct DualSignatureBatch {
        bytes32 batchId;               // Unique cryptographic settlement batch identifier
        bytes32 stateRoot;             // Merkle state root of trades settled in this batch
        uint256 grossNotionalPaise;    // Gross trade consideration in paise (for 0.00% fee (No fee at all) assessment)
        uint64 timestamp;              // Batch execution timestamp
        bytes ecdsaSignature;          // 65-byte secp256k1 classical signature (r, s, v)
        bytes mldsaSignature;          // 2420-byte NIST FIPS 204 ML-DSA-44 signature
        address classicalSigner;       // Expected authorized ECDSA signer address
        bytes32 pqcKeyId;              // Registered cryptographic identifier of ML-DSA public key
    }

    /// @notice Cryptographic Verification Result Details
    struct VerificationResult {
        bool isValid;                  // Overall verification boolean
        bool classicalValid;           // Outcome of ECDSA secp256k1 check
        bool pqcValid;                 // Outcome of ML-DSA-44 lattice check
        uint256 gasConsumed;           // Gas consumed during verification execution
        uint64 verifiedTimestamp;      // Timestamp of verification execution
    }

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error InvalidPublicKeyLength(uint256 actualLength, uint256 expectedLength);
    error InvalidSignatureLength(uint256 actualLength, uint256 expectedLength);
    error NormOutOfBounds(uint32 coefficientValue, uint32 allowedBound);
    error HintWeightExceeded(uint256 actualWeight, uint256 maximumWeight);
    error ChallengeMismatch(bytes32 expectedChallenge, bytes32 actualChallenge);
    error ClassicalSignatureFailed(address recoveredSigner, address expectedSigner);
    error LatticeKeyNotRegistered(bytes32 pqcKeyId);
    error LatticeKeyRevoked(bytes32 pqcKeyId);
    error InvalidBatchTimestamp(uint64 batchTimestamp, uint64 currentBlockTimestamp);
    error ZeroGrossConsideration();

    // -------------------------------------------------------------------------
    // Events
    // -------------------------------------------------------------------------

    event LatticePublicKeyRegistered(
        bytes32 indexed pqcKeyId,
        address indexed authority,
        bytes32 indexed rho,
        uint64 registeredAt
    );

    event LatticePublicKeyRevoked(
        bytes32 indexed pqcKeyId,
        address indexed revokedBy,
        uint64 revokedAt
    );

    event SignatureVerifiedPQC(
        bytes32 indexed messageDigest,
        bytes32 indexed pqcKeyId,
        bool success,
        uint256 gasUsed
    );

    event DualSignatureBatchValidated(
        bytes32 indexed batchId,
        address indexed classicalSigner,
        bytes32 indexed pqcKeyId,
        uint256 grossNotionalPaise,
        uint256 platformFeePaise
    );

    // -------------------------------------------------------------------------
    // Core Verification Functions
    // -------------------------------------------------------------------------

    /**
     * @notice Verifies an arbitrary message digest against a NIST FIPS 204 ML-DSA-44 signature.
     * @param messageDigest The 32-byte cryptographic digest of the message payload being verified.
     * @param signatureBytes The raw 2420-byte encoded ML-DSA-44 signature.
     * @param publicKeyBytes The raw 1312-byte encoded ML-DSA-44 public key.
     * @return isValid True if the signature is mathematically valid, false otherwise.
     */
    function verifyMLDSA44(
        bytes32 messageDigest,
        bytes calldata signatureBytes,
        bytes calldata publicKeyBytes
    ) external view returns (bool isValid);

    /**
     * @notice Verifies an ML-DSA-44 signature using a registered on-chain public key identifier.
     * @param messageDigest The 32-byte cryptographic digest of the message payload.
     * @param signatureBytes The raw 2420-byte encoded ML-DSA-44 signature.
     * @param pqcKeyId The unique registered hash identifier of the post-quantum public key.
     * @return isValid True if key exists, is active, and signature is valid.
     */
    function verifyWithRegisteredKey(
        bytes32 messageDigest,
        bytes calldata signatureBytes,
        bytes32 pqcKeyId
    ) external view returns (bool isValid);

    /**
     * @notice Verifies a hybrid dual-signature batch (ECDSA secp256k1 + ML-DSA-44).
     * @param batch The complete dual-signature batch struct containing payloads and both signatures.
     * @return result Detailed verification outcome struct including classical and PQC breakdown.
     */
    function verifyDualSignatureBatch(
        DualSignatureBatch calldata batch
    ) external view returns (VerificationResult memory result);

    /**
     * @notice Computes the registered cryptographic identifier for an ML-DSA-44 public key.
     * @param publicKeyBytes Raw 1312-byte public key payload.
     * @return pqcKeyId Keccak-256 commitment hash of the public key.
     */
    function computeKeyId(bytes calldata publicKeyBytes) external pure returns (bytes32 pqcKeyId);
}
```

### 2. Post-Quantum Lattice Signature Verifier Implementation (`MLDSAVerifier.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {IPostQuantumVerifier} from "./interfaces/IPostQuantumVerifier.sol";
import {LatticeMathYul} from "./libraries/LatticeMathYul.sol";

/**
 * @title MLDSAVerifier
 * @notice Highly optimized on-chain verifier for NIST FIPS 204 (ML-DSA-44 / Dilithium2) lattice signatures.
 * @dev Implements ring operations, Number Theoretic Transforms, and SHAKE permutations in inline Yul.
 */
contract MLDSAVerifier is IPostQuantumVerifier {
    // -------------------------------------------------------------------------
    // Storage - Key Registry
    // -------------------------------------------------------------------------

    struct RegisteredKey {
        bytes publicKey;
        address authority;
        uint64 registeredAt;
        bool isActive;
    }

    /// @notice Mapping from unique public key identifier to registered key details
    mapping(bytes32 => RegisteredKey) public registeredKeys;

    /// @notice Authorized administrative registry manager (governed by MultiSigGovernance.sol)
    address public immutable governanceContract;

    // -------------------------------------------------------------------------
    // Modifiers
    // -------------------------------------------------------------------------

    modifier onlyGovernance() {
        if (msg.sender != governanceContract) revert ClassicalSignatureFailed(msg.sender, governanceContract);
        _;
    }

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    constructor(address _governanceContract) {
        require(_governanceContract != address(0), "Invalid governance address");
        governanceContract = _governanceContract;
    }

    // -------------------------------------------------------------------------
    // Public Key Registry Management
    // -------------------------------------------------------------------------

    function registerLatticePublicKey(bytes calldata publicKeyBytes) external onlyGovernance returns (bytes32 pqcKeyId) {
        if (publicKeyBytes.length != ML_DSA_44_PUBLIC_KEY_BYTES) {
            revert InvalidPublicKeyLength(publicKeyBytes.length, ML_DSA_44_PUBLIC_KEY_BYTES);
        }

        pqcKeyId = keccak256(publicKeyBytes);
        bytes32 rho = bytes32(publicKeyBytes[0:32]);

        registeredKeys[pqcKeyId] = RegisteredKey({
            publicKey: publicKeyBytes,
            authority: msg.sender,
            registeredAt: uint64(block.timestamp),
            isActive: true
        });

        emit LatticePublicKeyRegistered(pqcKeyId, msg.sender, rho, uint64(block.timestamp));
    }

    function revokeLatticePublicKey(bytes32 pqcKeyId) external onlyGovernance {
        if (!registeredKeys[pqcKeyId].isActive) revert LatticeKeyRevoked(pqcKeyId);
        registeredKeys[pqcKeyId].isActive = false;
        emit LatticePublicKeyRevoked(pqcKeyId, msg.sender, uint64(block.timestamp));
    }

    // -------------------------------------------------------------------------
    // View Functions
    // -------------------------------------------------------------------------

    function computeKeyId(bytes calldata publicKeyBytes) external pure override returns (bytes32) {
        return keccak256(publicKeyBytes);
    }

    function verifyWithRegisteredKey(
        bytes32 messageDigest,
        bytes calldata signatureBytes,
        bytes32 pqcKeyId
    ) external view override returns (bool) {
        RegisteredKey storage keyRecord = registeredKeys[pqcKeyId];
        if (keyRecord.registeredAt == 0) revert LatticeKeyNotRegistered(pqcKeyId);
        if (!keyRecord.isActive) revert LatticeKeyRevoked(pqcKeyId);

        return this.verifyMLDSA44(messageDigest, signatureBytes, keyRecord.publicKey);
    }

    function verifyMLDSA44(
        bytes32 messageDigest,
        bytes calldata signatureBytes,
        bytes calldata publicKeyBytes
    ) external view override returns (bool isValid) {
        uint256 gasStart = gasleft();

        if (publicKeyBytes.length != ML_DSA_44_PUBLIC_KEY_BYTES) {
            revert InvalidPublicKeyLength(publicKeyBytes.length, ML_DSA_44_PUBLIC_KEY_BYTES);
        }
        if (signatureBytes.length != ML_DSA_44_SIGNATURE_BYTES) {
            revert InvalidSignatureLength(signatureBytes.length, ML_DSA_44_SIGNATURE_BYTES);
        }

        // Delegate execution to low-level Yul assembly verification pipeline
        isValid = LatticeMathYul.verifyMLDSA44Internal(
            messageDigest,
            signatureBytes,
            publicKeyBytes
        );

        uint256 gasUsed = gasStart - gasleft();
        // Event emission via staticcall wrapper if invoked in stateful context or internal tracking
        return isValid;
    }

    function verifyDualSignatureBatch(
        DualSignatureBatch calldata batch
    ) external view override returns (VerificationResult memory result) {
        uint256 startGas = gasleft();

        if (batch.grossNotionalPaise == 0) revert ZeroGrossConsideration();

        // 1. Classical ECDSA secp256k1 Verification
        bytes32 ecdsaDigest = keccak256(abi.encodePacked(
            batch.batchId,
            batch.stateRoot,
            batch.grossNotionalPaise,
            batch.timestamp
        ));

        address recoveredSigner = address(0);
        if (batch.ecdsaSignature.length == 65) {
            bytes32 r;
            bytes32 s;
            uint8 v;
            bytes memory sig = batch.ecdsaSignature;
            assembly ("memory-safe") {
                r := mload(add(sig, 32))
                s := mload(add(sig, 64))
                v := byte(0, mload(add(sig, 96)))
            }
            if (v < 27) v += 27;
            if (v == 27 || v == 28) {
                recoveredSigner = ecrecover(ecdsaDigest, v, r, s);
            }
        }

        result.classicalValid = (recoveredSigner != address(0) && recoveredSigner == batch.classicalSigner);

        // 2. Post-Quantum ML-DSA-44 Verification
        RegisteredKey storage keyRecord = registeredKeys[batch.pqcKeyId];
        if (keyRecord.registeredAt == 0 || !keyRecord.isActive) {
            result.pqcValid = false;
        } else {
            result.pqcValid = LatticeMathYul.verifyMLDSA44Internal(
                ecdsaDigest,
                batch.mldsaSignature,
                keyRecord.publicKey
            );
        }

        result.isValid = result.classicalValid && result.pqcValid;
        result.gasConsumed = startGas - gasleft();
        result.verifiedTimestamp = uint64(block.timestamp);
    }
}
```

### 3. Dual-Signed Settlement Hook (`DualSignedSettlementHook.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {IPostQuantumVerifier} from "./interfaces/IPostQuantumVerifier.sol";

/**
 * @title DualSignedSettlementHook
 * @notice Settlement compliance hook validating dual-signature authorizations (ECDSA + ML-DSA-44)
 *         for high-value trade clearance batches and sovereign bond coupon distributions.
 */
contract DualSignedSettlementHook {
    // -------------------------------------------------------------------------
    // Immutable Dependencies & Parameters
    // -------------------------------------------------------------------------

    IPostQuantumVerifier public immutable pqcVerifier;
    address public immutable settlementDvP;
    address public immutable treasuryVault;
    address public immutable coreSgfVault;
    address public immutable ipfVault;

    /// @notice Minimum threshold in paise (e.g. 10,000,000 INR = 1,000,000,000 paise) requiring dual PQC signing
    uint256 public immutable dualSignThresholdPaise;

    // -------------------------------------------------------------------------
    // Custom Errors
    // -------------------------------------------------------------------------

    error UnauthorizedSettlementCaller(address caller);
    error DualSignatureAuthorizationFailed(bytes32 batchId, bool classicalValid, bool pqcValid);
    error InvalidFeeSplitDistribution();

    // -------------------------------------------------------------------------
    // Constructor
    // -------------------------------------------------------------------------

    constructor(
        address _pqcVerifier,
        address _settlementDvP,
        address _treasuryVault,
        address _coreSgfVault,
        address _ipfVault,
        uint256 _dualSignThresholdPaise
    ) {
        require(_pqcVerifier != address(0), "Invalid verifier");
        require(_settlementDvP != address(0), "Invalid DvP");
        require(_treasuryVault != address(0), "Invalid Treasury");
        require(_coreSgfVault != address(0), "Invalid Core SGF");
        require(_ipfVault != address(0), "Invalid IPF");

        pqcVerifier = IPostQuantumVerifier(_pqcVerifier);
        settlementDvP = _settlementDvP;
        treasuryVault = _treasuryVault;
        coreSgfVault = _coreSgfVault;
        ipfVault = _ipfVault;
        dualSignThresholdPaise = _dualSignThresholdPaise;
    }

    // -------------------------------------------------------------------------
    // Authorization Verification Hook
    // -------------------------------------------------------------------------

    /**
     * @notice Enforces dual-signature verification on settlement batches.
     * @param batch Dual-signature batch payload with classical and post-quantum cryptographic proofs.
     * @return authorized True if the batch satisfies all classical and quantum-safe authorization conditions.
     */
    function authorizeBatchSettlement(
        IPostQuantumVerifier.DualSignatureBatch calldata batch
    ) external returns (bool authorized) {
        if (msg.sender != settlementDvP) {
            revert UnauthorizedSettlementCaller(msg.sender);
        }

        // Transactions below the institutional threshold require only standard classical authorization
        if (batch.grossNotionalPaise < dualSignThresholdPaise) {
            return true;
        }

        // Execute full dual-signature cryptographic verification
        IPostQuantumVerifier.VerificationResult memory result = pqcVerifier.verifyDualSignatureBatch(batch);
        if (!result.isValid) {
            revert DualSignatureAuthorizationFailed(batch.batchId, result.classicalValid, result.pqcValid);
        }

        // Assess and verify universal 0.00% (No fee at all) platform fee allocation (0 bps (0.00% fee at launch))
        uint256 feeBps = feeController.takerFeeBps(); // 0 bps at launch
        uint256 totalFeePaise = (batch.grossNotionalPaise * feeBps) / 10000;
        uint256 treasuryShare = totalFeePaise > 0 ? (totalFeePaise * 60) / 100 : 0;
        uint256 coreSgfShare = totalFeePaise > 0 ? (totalFeePaise * 25) / 100 : 0;
        uint256 ipfShare = totalFeePaise - (treasuryShare + coreSgfShare);

        if (treasuryShare + coreSgfShare + ipfShare != totalFeePaise) {
            revert InvalidFeeSplitDistribution();
        }

        return true;
    }
}
```

## Security & Compliance Notes
- **Lattice Hardness & Shor Algorithm Immunity:** ML-DSA-44 derives its mathematical security from the worst-case hardness of the Module Learning With Errors (M-LWE) and Module Short Integer Solution (M-SIS) problems over the polynomial ring $R_q = \mathbb{Z}_q[X]/(X^{256} + 1)$. There are no known quantum algorithms (including Shor's or Grover's algorithms) capable of solving lattice vector problems in polynomial time. ML-DSA-44 provides 128 bits of post-quantum security (NIST Security Level 2), matching the post-quantum collision resistance of SHA-256.
- **Dual-Signature Defense-in-Depth:** While lattice cryptography has undergone extensive peer review, combining ECDSA secp256k1 and ML-DSA-44 in a hybrid dual-signature construction provides defense-in-depth against unforeseen mathematical vulnerabilities in either classical elliptic curve assumptions or lattice parameter heuristics. A settlement batch requires both signatures to be cryptographically valid.
- **Constant-Time Yul Arithmetic & Timing Side-Channel Elimination:** All polynomial inner loops, Montgomery modular multiplications, and coefficient comparisons implemented in `LatticeMathYul.sol` execute in constant time. No conditional jumps (`if`, `switch`) occur based on intermediate polynomial coefficient values or signature component bits, preventing cache and timing side-channel attacks on validator nodes.
- **Strict EVM Memory Safety in Yul (`assembly ("memory-safe")`):** Yul blocks operate with explicit memory safety annotations. The contract calculates required polynomial scratch memory upfront, checks memory bounds against Solidity's free memory pointer at offset `0x40`, and restores memory pointers upon completion. This prevents stack/memory collisions, heap corruption, and arbitrary memory overwrite vulnerabilities.
- **NIST FIPS 204 Parameter Enforcement:**
  - Norm Bound: The response vector $z$ is strictly verified against $\|z\|_\infty < \gamma_1 - \beta$ ($131072 - 78 = 130994$). Any coefficient exceeding this bound results in immediate transaction rejection (`NormOutOfBounds`).
  - Hint Bit Weight: The hint vector $h$ is strictly verified to contain no more than $\omega = 80$ non-zero bits. Any signature containing more than 80 ones is rejected (`HintWeightExceeded`), preventing hint manipulation and forgery attacks.
  - Modulus & Reduction: All arithmetic is strictly reduced modulo prime $q = 8380417$.
- **Zero On-Chain PII Invariant:** In compliance with India's Digital Personal Data Protection Act (DPDPA 2023), SEBI cybersecurity frameworks, and GDPR:
  - Smart contracts store only cryptographic public key identifiers (`bytes32 pqcKeyId`), public key seeds (`bytes32 rho`), and compressed coefficient vectors.
  - Zero institutional names, personal account numbers (PAN), tax identifiers (TAN), IP addresses, or authorized representative credentials exist on the Besu ledger.
- **Universal Zero-Fee Model (0.00% fee across all trading - No fee at all):** In strict compliance with Prompt 006, dual-signed high-value settlement authorizations calculate the exact 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) platform fee on gross turnover. Future fee parameter adjustments are governed dynamically via FeeController.sol (0.00% at launch). Zero holding, custody, or post-quantum security surcharge fees are assessed.
- **Multi-Signature Administrative Governance:** Registration, revocation, or upgrade of post-quantum public keys and verifier contracts requires 3-of-5 threshold signatures from `MultiSigGovernance.sol` (Prompt 307) executed via FIPS 140-3 Level 3 HSM keys (Prompt 311 and Prompt 717).

## Acceptance Criteria
- [ ] `IPostQuantumVerifier.sol`, `MLDSAVerifier.sol`, `LatticeMathYul.sol`, and `DualSignedSettlementHook.sol` compile under Solidity ^0.8.24 with zero warnings and `via_ir = true`.
- [ ] Number Theoretic Transform (NTT) and Inverse NTT (INTT) algorithms in Yul satisfy strict mathematical invertibility: $\text{INTT}(\text{NTT}(a)) \equiv a \pmod q$ for all 256 coefficients across 10,000 randomized polynomial fuzzing inputs.
- [ ] Montgomery multiplication in Yul produces bit-exact modular arithmetic outputs identical to reference Python and Rust implementations across all corner cases ($0, 1, q-1, q$).
- [ ] Official NIST FIPS 204 ML-DSA-44 Known Answer Tests (KAT) vectors pass with 100% precision: valid signatures evaluate to true; single-bit flips in message digest, challenge seed $\tilde{c}$, response vector $z$, or public key $t_1$ revert with expected custom errors.
- [ ] Norm bound verification strictly rejects any signature where any coefficient $|z_i| \ge 130994$.
- [ ] Hint weight verification strictly rejects any signature with $> 80$ non-zero bits in $h$.
- [ ] Single ML-DSA-44 signature verification gas consumption executes reliably within 1,200,000 to 1,650,000 gas on Hyperledger Besu.
- [ ] Dual-signature verification protocol correctly validates atomic execution: passes when both ECDSA and ML-DSA-44 signatures are valid; reverts if either classical or lattice signature is invalid or forged.
- [ ] Yul assembly routines pass Echidna and Halmos invariant tests confirming zero memory pointer corruption beyond allocated scratch space.
- [ ] Slither and Mythril static analysis pipelines complete with zero high, medium, or reentrancy security findings.
- [ ] Settlement hook correctly calculates the exact 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) platform fee on gross notional consideration with the 0.00% fee at launch (governed by FeeController.sol) split.
- [ ] Zero em dashes and zero en dashes present across the entire specification document (strict adherence to standard ASCII hyphens).

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `307` (MultiSig Governance Contract), Prompt `311` (Validator Key Management HSM).
- **Parallel Tasks:** Prompt `237` (Multichain MPC-TSS Vault Service), Prompt `717` (CloudHSM Signing Daemon), Prompt `333` (Tokenized G-Sec Bonds & Coupon Accrual Contract).
- **Downstream Prompts Enabled:** Prompt `306` (Settlement DvP Smart Contract), Prompt `329` (NBSE Settlement DvP & Fee Collector), Prompt `334` (RBI CBDC eINR Wholesale & Retail Bridge Contract), Prompt `308` (On-Chain Proof of Reserve Publishing).
