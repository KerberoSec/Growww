# 332 - Commodity Scrap Equalizer and In-Transit Escrow Smart Contracts

## Purpose
Physical commodity asset tokenization requires bridging the precision of mathematical digital tokens (18 decimals) with the physical realities of precious metal metallurgy, refining, logistics, and delivery. In physical bullion refining and minting (governed by BIS, LBMA, and WDRA standards for 999/999.9 gold and silver bars), standard minted bars inevitably exhibit miniscule physical weight discrepancies (manufacturing scrap / minting variance) compared to nominal standardized units. Furthermore, tokenized physical commodities require secure institutional transfer across accredited vault facilities, armored logistics carriers (such as Brink's, Sequel, and BVC Logistics), and physical delivery recipients without breaking cryptographic Proof-of-Reserve guarantees or exposing client Personally Identifiable Information (PII).

This prompt specifies the design, architecture, interface definitions, testing strategy, and security model of two core Hyperledger Besu smart contracts:
1. `CommodityScrapEqualizer.sol` (`ICommodityScrapEqualizer.sol`): Programmatically enforces a strict +/- 0.10% (+/- 10 basis points) manufacturing scrap/weight variance tolerance limit for physical bullion bars, executing automatic spot cash (e-Rupee CBDC / INR fiat) or token equalization based on real-time decentralized oracle prices.
2. `InTransitEscrowRegistry.sol` (`IInTransitEscrowRegistry.sol`): Coordinates the lifecycle of physical bullion lots in transit, enforcing cryptographic lot lineage preservation (`splitCommodityLot` in `PhysicalVaultRegistry.sol`), tamper-evident security seals, and atomic token burn execution strictly upon recipient dual-factor biometric proof and time-bounded OTP verification at physical handoff.

## What You Are Building
A production-grade Solidity smart contract suite and test architecture under `contracts/commodities/` containing:
- `ICommodityScrapEqualizer.sol`: Solidity interface defining scrap tolerance evaluation (+/- 0.10% / 1000 ppm), spot price equalization calculation, CBDC/token surcharge debiting, and refund credit mechanics for overweight/underweight bullion bars.
- `IInTransitEscrowRegistry.sol`: Solidity interface orchestrating in-transit lot custody transfers between WDRA vaults, armored carriers, and end-recipients, tracking tamper-evident seal digests, GPS geofence attestations, delivery attempt logs, and dispute escalation states.
- Lineage-Preserving Lot Split Interface extensions for `IPhysicalVaultRegistry.sol`: Extending the canonical vault registry with `splitCommodityLot` functionality that decommissions parent lots, generates verifiable child lots, and links their cryptographic lineage hashes to refinery split assay certificates.
- Biometric & OTP Delivery Verification Interface extensions for `ICommodityDeliveryBurner.sol`: Verifying blinded recipient biometric signature nullifiers and time-bounded OTP preimage proofs before executing irreversible on-chain token burns.
- Comprehensive Foundry Test Suite (`test/commodities/CommodityScrapEqualizer.t.sol`, `test/commodities/InTransitEscrowRegistry.t.sol`): Unit tests, invariant suites, and property-based fuzz tests verifying scrap tolerance bounds, lineage integrity across multi-generation splits, replay-proof biometric/OTP verifications, and atomic escrow state transitions.

## Scope Boundaries
- **In Scope:**
  - Automated manufacturing scrap tolerance enforcement bounded strictly to +/- 0.10% (+/- 10 bps / 1000 ppm) of nominal lot weight.
  - Automatic spot price cash (e-Rupee / stablecoin) or token equalization for physical weight variances during minting, recasting, splitting, and redemption.
  - Lineage-preserving lot splitting (`splitCommodityLot`) in `PhysicalVaultRegistry.sol`, tracking cryptographic ancestry from parent bar serials and NABL assay certificates to child units.
  - Multi-state in-transit escrow registry (`InTransitEscrowRegistry.sol`) tracking armored carrier dispatch, custody handoff, tamper-evident security bag serials, and delivery attempt timeouts.
  - Recipient delivery handoff verification requiring dual cryptographic proofs: zero-PII blinded biometric proof hash and time-bounded OTP hash commitment.
  - Atomic token burn execution (`CommoditySecurityToken.burnForDelivery`) triggered exclusively upon successful recipient biometric + OTP verification.
  - In-transit dispute resolution, return-to-vault workflows, and carrier emergency fail-safe locks.
  - Role-Based Access Control (RBAC) via OpenZeppelin `AccessControlEnumerableUpgradeable` and proxy upgradeability via ERC-1967 UUPS pattern.
- **Out of Scope / Handled Elsewhere:**
  - Raw biometric template capture and Aadhaar / FIDO2 RD hardware SDK interaction (handled off-chain in mobile/carrier terminals; on-chain contracts receive only blinded zero-knowledge / cryptographic proof hashes).
  - Off-chain SMS / TOTP dispatch infrastructure (handled in Prompt 211 Notification Service).
  - Carrier GPS tracking telemetry ingestion and telemetry streaming pipelines (handled in Prompt 243 MCX Commodity & Warehouse Receipt Adapter).
  - Price oracle aggregation and feed publishing (handled in Prompt 328 Oracle Aggregator).
  - Web and mobile user interface for delivery booking and scrap settlement (handled in Prompt 530 and Prompt 608).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with transient storage opcodes `TSTORE`/`TLOAD` support where applicable).
  *Justification:* Provides native checked arithmetic preventing overflow/underflow, user-defined value types for asset precision, and optimized Yul assembly for cryptographic hashing.
- **Contract Standards & Security Libraries:** OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, `ECDSA`, `EIP712Upgradeable`, `ERC1967Utils`).
- **Cryptographic Primitives:** Keccak-256 for lineage Merkle hashing, EIP-712 typed structured data signing for carrier attestations, and SHA-256 / Keccak-256 for blinded biometric and OTP hash commitments.
- **Development & Testing Framework:** Foundry (`forge` for compilation and fuzzing, `cast` for RPC interaction, `anvil` for local node simulation).
- **Static Analysis & Formal Verification:** Slither (Trail of Bits), Solhint, and Halmos / Certora for mathematical invariant validation.
- **Consortium Blockchain Target:** Hyperledger Besu enterprise ledger running QBFT consensus with 1:1 physical custody backing.

## Backend / Infra Touchpoints
- **MCX Commodity & Warehouse Receipt Adapter (Prompt 243):** Ingests physical refinery assay weight certificates, calculates bar weight discrepancies, and submits scrap equalization transactions.
- **Decentralized Oracle Aggregator (Prompt 328):** Provides real-time, volume-weighted spot commodity prices (e.g. `XAU/INR`, `XAG/INR`) for spot cash equalization.
- **Wallet & Double-Entry Ledger Service (Prompt 203):** Adjusts user fiat/e-Rupee balances to reflect spot scrap surcharges or refunds.
- **Blockchain Event Indexer (Prompt 309):** Indexes `ScrapEqualizationExecuted`, `LotSplitRegistered`, `InTransitEscrowCreated`, and `CommodityDeliveryBurnFinalized` events for downstream accounting.
- **Custodian Depository Integration Service (Prompt 213):** Synchronizes WDRA e-NWR subdivision records with on-chain `splitCommodityLot` updates.
- **Notification Service (Prompt 211):** Generates and routes time-bounded delivery OTPs to authenticated recipients upon carrier arrival.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Deployment Model:** Deployed as ERC-1967 UUPS upgradeable proxies on the Hyperledger Besu enterprise consortium network.
- **Strict Scrap Tolerance Invariant:** The contract enforces:
  $$\left|\frac{W_{\text{actual}} - W_{\text{nominal}}}{W_{\text{nominal}}}\right| \le 0.0010 \quad (\pm 0.10\% \text{ or } \pm 1000 \text{ ppm})$$
  Any deposit, recasting, or delivery request exceeding this tolerance is rejected with custom error `ScrapToleranceExceeded`.
- **Zero-PII Biometric & OTP Verification:** Delivery handoff relies entirely on cryptographic commitments:
  $$\text{BiometricProofHash} = \text{keccak256}(\text{abi.encodePacked}(\text{claimantAddress}, \text{biometricNullifier}, \text{deliveryId}))$$
  $$\text{OTPHash} = \text{keccak256}(\text{abi.encodePacked}(\text{otpCode}, \text{otpSalt}, \text{deliveryId}))$$
  No names, phone numbers, Aadhaar numbers, or raw biometric images exist on-chain.
- **Atomic Escrow & Burn Guarantee:** Tokens transferred to `InTransitEscrowRegistry.sol` remain locked in escrow throughout transit. Tokens cannot be returned to circulating supply or redeemed twice. Execution of `confirmHandoffAndExecuteBurn` atomically calls `CommoditySecurityToken.burnForDelivery`, destroys the escrowed balance, marks the in-transit escrow as `DELIVERED`, and permanently decommissions the vaulted lot in `PhysicalVaultRegistry.sol`.

## Scrap Equalization & Lineage Preservation Mechanics

### 1. Mathematical Formulation of Scrap Equalization
When physical bullion is assayed, the actual measured pure weight ($W_{\text{actual}}$) may deviate from the nominal unit weight ($W_{\text{nominal}}$). The scrap weight delta ($\Delta W$) and tolerance ratio ($\tau$) are calculated as:
$$\Delta W = W_{\text{actual}} - W_{\text{nominal}}$$
$$\tau = \frac{|\Delta W|}{W_{\text{nominal}}}$$

If $\tau > 0.0010$ (10 bps), the transaction reverts with `ScrapToleranceExceeded(actualWeight, nominalWeight, maxAllowedDelta)`.

If $\tau \le 0.0010$, the spot cash equalization amount ($E_{\text{cash}}$) is calculated using the oracle spot price ($P_{\text{spot}}$ in 18-decimal e-Rupee per gram):
$$E_{\text{cash}} = |\Delta W| \times P_{\text{spot}}$$

Equalization Settlement Modes:
1. **Underweight Bar ($\Delta W < 0$):**
   - Depositor / Vault receives a spot cash debit or provides additional token balance.
   - For delivery redemptions, the claimant receives a cash equalization refund credit equal to $E_{\text{cash}}$.
2. **Overweight Bar ($\Delta W > 0$):**
   - Depositor / Vault receives a cash equalization credit or mints supplementary fractional tokens.
   - For delivery redemptions, the claimant pays a spot cash surcharge equal to $E_{\text{cash}}$ before dispatch authorization.

### 2. Cryptographic Lineage Preservation in Lot Splitting
When an institutional parent lot ($L_{\text{parent}}$ with weight $W_{\text{parent}}$) is split into $N$ child lots ($L_{\text{child}, i}$ with weights $W_{\text{child}, i}$):
1. **Weight Conservation:**
   $$\sum_{i=1}^{N} W_{\text{child}, i} = W_{\text{parent}} - W_{\text{scrap}}$$
   where $W_{\text{scrap}} \le W_{\text{parent}} \times 0.0010$.
2. **Cryptographic Lineage Hash:**
   Every child lot possesses an immutable cryptographic lineage hash linking it back to the parent lot, the refinery split certification digest, and its child index:
   $$\text{LineageHash}_i = \text{keccak256}(\text{abi.encodePacked}(L_{\text{parent}}, i, W_{\text{child}, i}, \text{childAssayDigest}_i, \text{refinerySplitCertDigest}))$$
3. **Parent Decommissioning:**
   $L_{\text{parent}}$ is marked as `SPLIT_DECOMMISSIONED` in `PhysicalVaultRegistry.sol`. It can never again be delivered, transferred, or split.

### 3. In-Transit Delivery State Machine
```
[DELIVERY_REQUESTED]
         |
         v (Carrier assigned & security pouch seal recorded)
[IN_TRANSIT_DISPATCHED]
         |
    +----+----+
    |         |
    |         v (Recipient unavailable / address issue)
    |    [DELIVERY_ATTEMPT_FAILED]
    |         |
    |         +---> [RETURN_IN_TRANSIT] ---> [RETURNED_TO_VAULT]
    v (Recipient present + Biometric Proof + OTP verified)
[HANDOFF_VERIFIED]
         |
         v (Atomic Token Burn + Lot Decommissioning)
[DELIVERY_FINALIZED]
```

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold directory structure under `contracts/commodities/`:
   - `src/commodities/interfaces/ICommodityScrapEqualizer.sol`
   - `src/commodities/interfaces/IInTransitEscrowRegistry.sol`
   - `src/commodities/interfaces/IPhysicalVaultRegistryLineage.sol`
   - `src/commodities/CommodityScrapEqualizer.sol`
   - `src/commodities/InTransitEscrowRegistry.sol`
   - `test/commodities/CommodityScrapEqualizer.t.sol`
   - `test/commodities/InTransitEscrowRegistry.t.sol`
   - `script/DeployScrapAndEscrow.s.sol`
2. Define complete enumerations, structs, custom errors, and events in `ICommodityScrapEqualizer.sol` covering scrap tolerances, spot cash equalization records, and oracle bindings.
3. Define complete enumerations, structs, custom errors, and events in `IInTransitEscrowRegistry.sol` covering in-transit lifecycle states, carrier assignments, tamper seals, biometric proofs, and OTP hashes.
4. Define interface extensions in `IPhysicalVaultRegistryLineage.sol` specifying `splitCommodityLot`, child lot registration, and cryptographic lineage query functions.
5. Implement `CommodityScrapEqualizer.sol` inheriting from `Initializable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, and `ReentrancyGuardUpgradeable`:
   - Enforce max scrap tolerance constant (`MAX_SCRAP_TOLERANCE_BPS = 10`, representing 0.10%).
   - Implement `calculateScrapDeltaAndEqualization(bytes32 commodityId, uint256 actualWeightGrams, uint256 nominalWeightGrams)` querying `IOracleAggregator` for the latest spot price.
   - Implement `settleMintEqualization` and `settleRedemptionEqualization` supporting atomic cash (ERC-20 CBDC) transfers or token adjustments.
6. Implement `splitCommodityLot` in `PhysicalVaultRegistry.sol`:
   - Restrict execution to `REFINERY_OPERATOR_ROLE` with dual-signature attestation from accredited WDRA vault custodian.
   - Verify parent lot is active (`VAULTED`) and not locked or quarantined.
   - Verify sum of child weights plus scrap loss equals parent weight within +/- 0.10% tolerance.
   - Generate child lot hashes with lineage links and mark parent lot `SPLIT_DECOMMISSIONED`.
7. Implement `InTransitEscrowRegistry.sol` inheriting from `Initializable`, `AccessControlEnumerableUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, and `EIP712Upgradeable`:
   - Implement `createInTransitEscrow(bytes32 deliveryId, bytes32 lotHash, address claimant, uint256 tokenAmount, bytes32 biometricProofHash, bytes32 otpHashCommitment, uint256 validUntil)` locking tokens from claimant or delivery burner.
   - Implement `dispatchWithCarrier(bytes32 deliveryId, address carrierAgent, bytes32 pouchSealDigest, bytes32 airwayBillHash)` restricted to `VAULT_OPERATOR_ROLE`.
   - Implement `recordDeliveryAttempt(bytes32 deliveryId, string calldata failureReason)` to record non-delivery events and increment attempt counter.
   - Implement `confirmHandoffAndExecuteBurn(bytes32 deliveryId, bytes32 biometricNullifier, string calldata otpCode, bytes32 otpSalt, bytes calldata carrierSignature)`:
     - Verify current block timestamp <= `validUntil`.
     - Verify biometric nullifier matches `biometricProofHash`.
     - Verify `keccak256(abi.encodePacked(otpCode, otpSalt, deliveryId)) == otpHashCommitment`.
     - Verify carrier EIP-712 signature over delivery completion digest.
     - Call `ICommoditySecurityToken.burnForDelivery` to execute atomic token burn.
     - Call `IPhysicalVaultRegistry.markLotDelivered` to update lot status.
     - Transition escrow state to `DELIVERY_FINALIZED`.
8. Implement carrier return workflow (`initiateReturnToVault`, `confirmVaultReturn`) for packages that cannot be delivered within max allowed attempts (default: 3 attempts).
9. Implement emergency dispute resolution mechanism (`flagDisputedInTransit`, `resolveDispute`) controlled by multi-sig governance (`GOVERNANCE_ROLE`).
10. Write comprehensive unit tests in `test/commodities/CommodityScrapEqualizer.t.sol`:
    - Test exact nominal weight (zero scrap delta, zero cash equalization).
    - Test underweight within 0.10% tolerance (verifying exact refund cash calculation).
    - Test overweight within 0.10% tolerance (verifying exact surcharge cash calculation).
    - Test scrap variance exceeding 0.10% (verifying immediate revert with `ScrapToleranceExceeded`).
    - Test oracle stale price rejection during equalization.
11. Write comprehensive unit tests in `test/commodities/InTransitEscrowRegistry.t.sol`:
    - Test escrow creation and token locking.
    - Test carrier dispatch and pouch seal recording.
    - Test valid biometric nullifier and OTP matching resulting in successful atomic burn.
    - Test invalid OTP code reverting with `InvalidOTPPreimage`.
    - Test invalid biometric nullifier reverting with `InvalidBiometricProof`.
    - Test expired delivery window reverting with `DeliveryWindowExpired`.
    - Test return-to-vault workflow after 3 failed delivery attempts.
12. Write Foundry property-based fuzz tests:
    - Fuzz test scrap calculations across all valid weight ranges (0.1g to 100,000g) ensuring precision within 1 wei of e-Rupee.
    - Fuzz test lot splitting with random splits (2 to 50 child lots) verifying total weight conservation and lineage hash uniqueness.
    - Invariant test: Total in-transit escrowed tokens + circulating tokens == total active vaulted weight + scrap adjusted reserves.
13. Run Slither static analysis and ensure zero high, medium, or reentrancy findings.
14. Configure deployment script `DeployScrapAndEscrow.s.sol` for Hyperledger Besu with verified proxy admin and role configurations.
15. Document contract ABIs, events, custom errors, and EIP-712 schemas for integration with backend logistics services and Flutter delivery agent apps.

## Interfaces / Contracts

### 1. Commodity Scrap Equalizer Interface (`ICommodityScrapEqualizer.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ICommodityScrapEqualizer {
    enum ScrapAdjustmentType {
        EXACT_MATCH,
        UNDERWEIGHT_REFUND,
        OVERWEIGHT_SURCHARGE
    }

    struct ScrapEqualizationRecord {
        bytes32 recordId;
        bytes32 commodityId;
        bytes32 lotHash;
        uint256 nominalWeightGrams;
        uint256 actualWeightGrams;
        int256 weightDeltaGrams;
        uint256 toleranceBps;
        uint256 spotPricePerGram;
        uint256 equalizationCashAmount;
        ScrapAdjustmentType adjustmentType;
        address settledAccount;
        uint256 settledTimestamp;
    }

    event ScrapEqualizationCalculated(
        bytes32 indexed recordId,
        bytes32 indexed commodityId,
        bytes32 indexed lotHash,
        uint256 nominalWeightGrams,
        uint256 actualWeightGrams,
        int256 weightDeltaGrams,
        uint256 equalizationCashAmount,
        ScrapAdjustmentType adjustmentType
    );

    event ScrapEqualizationSettled(
        bytes32 indexed recordId,
        address indexed settledAccount,
        uint256 cashAmount,
        ScrapAdjustmentType adjustmentType,
        uint256 timestamp
    );

    event ScrapToleranceParametersUpdated(
        bytes32 indexed commodityId,
        uint256 maxToleranceBps,
        address indexed operator
    );

    error ScrapToleranceExceeded(
        uint256 actualWeightGrams,
        uint256 nominalWeightGrams,
        uint256 deltaGrams,
        uint256 maxAllowedToleranceBps
    );

    error InvalidWeightParameters(uint256 nominalWeightGrams, uint256 actualWeightGrams);
    error StaleOraclePrice(bytes32 commodityId, uint256 priceTimestamp, uint256 maxAge);
    error EqualizationAlreadySettled(bytes32 recordId);
    error UnauthorizedEqualizerCaller(address caller);
    error InsufficientCashSettlementBalance(address account, uint256 available, uint256 required);

    function calculateScrapDelta(
        bytes32 commodityId,
        uint256 nominalWeightGrams,
        uint256 actualWeightGrams
    )
        external
        view
        returns (
            int256 weightDeltaGrams,
            uint256 toleranceBps,
            uint256 equalizationCashAmount,
            ScrapAdjustmentType adjustmentType
        );

    function settleMintEqualization(
        bytes32 commodityId,
        bytes32 lotHash,
        uint256 nominalWeightGrams,
        uint256 actualWeightGrams,
        address depositorAccount
    ) external returns (bytes32 recordId, uint256 equalizationCashAmount, ScrapAdjustmentType adjustmentType);

    function settleRedemptionEqualization(
        bytes32 commodityId,
        bytes32 deliveryId,
        bytes32 lotHash,
        uint256 nominalWeightGrams,
        uint256 actualWeightGrams,
        address claimantAccount
    ) external returns (bytes32 recordId, uint256 equalizationCashAmount, ScrapAdjustmentType adjustmentType);

    function getMaxToleranceBps(bytes32 commodityId) external view returns (uint256);

    function getEqualizationRecord(bytes32 recordId) external view returns (ScrapEqualizationRecord memory);
}
```

### 2. In-Transit Escrow Registry Interface (`IInTransitEscrowRegistry.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IInTransitEscrowRegistry {
    enum InTransitState {
        ESCROW_INITIATED,
        CARRIER_ASSIGNED,
        IN_TRANSIT,
        DELIVERY_ATTEMPT_FAILED,
        HANDOFF_VERIFIED,
        DELIVERY_FINALIZED,
        RETURN_IN_TRANSIT,
        RETURNED_TO_VAULT,
        DISPUTED
    }

    struct InTransitEscrowOrder {
        bytes32 deliveryId;
        bytes32 commodityId;
        bytes32 lotHash;
        address claimant;
        address tokenContract;
        uint256 tokenAmount;
        address assignedCarrier;
        address carrierAgent;
        bytes32 securityPouchSealDigest;
        bytes32 airwayBillHash;
        bytes32 biometricProofHash;
        bytes32 otpHashCommitment;
        uint8 deliveryAttempts;
        uint8 maxDeliveryAttempts;
        InTransitState state;
        uint256 creationTimestamp;
        uint256 validUntil;
        uint256 finalizedTimestamp;
    }

    struct DeliveryAttemptRecord {
        bytes32 deliveryId;
        uint8 attemptNumber;
        uint256 attemptTimestamp;
        bytes32 carrierGeoHash;
        string failureReason;
    }

    event InTransitEscrowCreated(
        bytes32 indexed deliveryId,
        bytes32 indexed commodityId,
        bytes32 indexed lotHash,
        address claimant,
        uint256 tokenAmount,
        bytes32 biometricProofHash,
        bytes32 otpHashCommitment,
        uint256 validUntil
    );

    event CarrierDispatched(
        bytes32 indexed deliveryId,
        address indexed carrierContract,
        address indexed carrierAgent,
        bytes32 securityPouchSealDigest,
        bytes32 airwayBillHash
    );

    event DeliveryAttemptLogged(
        bytes32 indexed deliveryId,
        uint8 attemptNumber,
        bytes32 carrierGeoHash,
        string failureReason,
        uint256 timestamp
    );

    event BiometricAndOTPVerified(
        bytes32 indexed deliveryId,
        address indexed claimant,
        bytes32 biometricNullifier,
        uint256 timestamp
    );

    event CommodityDeliveryBurnFinalized(
        bytes32 indexed deliveryId,
        address indexed claimant,
        address indexed tokenContract,
        uint256 burnedTokenAmount,
        bytes32 lotHash,
        uint256 finalizedTimestamp
    );

    event InTransitReturnInitiated(
        bytes32 indexed deliveryId,
        bytes32 returnAirwayBillHash,
        string reason
    );

    event InTransitReturnedToVault(
        bytes32 indexed deliveryId,
        bytes32 indexed vaultId,
        bytes32 indexed lotHash,
        uint256 restoredTimestamp
    );

    event InTransitDisputeOpened(
        bytes32 indexed deliveryId,
        string disputeReason,
        address indexed reporter
    );

    event InTransitDisputeResolved(
        bytes32 indexed deliveryId,
        InTransitState resolvedState,
        string resolutionNotes,
        address indexed resolver
    );

    error EscrowAlreadyExists(bytes32 deliveryId);
    error EscrowNotFound(bytes32 deliveryId);
    error InvalidInTransitState(bytes32 deliveryId, InTransitState current, InTransitState expected);
    error DeliveryWindowExpired(bytes32 deliveryId, uint256 currentTimestamp, uint256 validUntil);
    error DeliveryWindowStillActive(bytes32 deliveryId, uint256 currentTimestamp, uint256 validUntil);
    error MaxDeliveryAttemptsExceeded(bytes32 deliveryId, uint8 attempts, uint8 maxAllowed);
    error InvalidBiometricProof(bytes32 deliveryId, bytes32 providedNullifier);
    error InvalidOTPPreimage(bytes32 deliveryId);
    error InvalidCarrierSignature(bytes32 deliveryId, address expectedSigner, address recoveredSigner);
    error UnauthorizedCarrierOperator(address caller, address authorizedCarrier);
    error UnauthorizedVaultCustodian(address caller);
    error InsufficientEscrowBalance(address token, uint256 available, uint256 required);

    function createInTransitEscrow(
        bytes32 deliveryId,
        bytes32 commodityId,
        bytes32 lotHash,
        address claimant,
        address tokenContract,
        uint256 tokenAmount,
        bytes32 biometricProofHash,
        bytes32 otpHashCommitment,
        uint256 validUntil
    ) external;

    function dispatchWithCarrier(
        bytes32 deliveryId,
        address carrierContract,
        address carrierAgent,
        bytes32 securityPouchSealDigest,
        bytes32 airwayBillHash
    ) external;

    function recordDeliveryAttempt(
        bytes32 deliveryId,
        bytes32 carrierGeoHash,
        string calldata failureReason
    ) external;

    function confirmHandoffAndExecuteBurn(
        bytes32 deliveryId,
        bytes32 biometricNullifier,
        string calldata otpCode,
        bytes32 otpSalt,
        bytes calldata carrierSignature
    ) external;

    function initiateReturnToVault(
        bytes32 deliveryId,
        bytes32 returnAirwayBillHash,
        string calldata reason
    ) external;

    function confirmVaultReturn(
        bytes32 deliveryId,
        bytes32 vaultId,
        bytes32 receivingInspectorDigest
    ) external;

    function flagDisputedInTransit(
        bytes32 deliveryId,
        string calldata disputeReason
    ) external;

    function resolveDispute(
        bytes32 deliveryId,
        InTransitState targetState,
        string calldata resolutionNotes
    ) external;

    function getInTransitEscrowOrder(bytes32 deliveryId) external view returns (InTransitEscrowOrder memory);

    function getDeliveryAttempts(bytes32 deliveryId) external view returns (DeliveryAttemptRecord[] memory);
}
```

### 3. Physical Vault Registry Lineage Extension Interface (`IPhysicalVaultRegistryLineage.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IPhysicalVaultRegistryLineage {
    enum LotLifecycleState {
        UNREGISTERED,
        VAULTED,
        LOCKED_FOR_DELIVERY,
        IN_TRANSIT,
        DELIVERED,
        SPLIT_DECOMMISSIONED,
        QUARANTINED
    }

    struct ChildLotSpec {
        bytes32 childBarSerialHash;
        uint256 netPureWeightGrams;
        bytes32 assayCertificateDigest;
        bytes32 eNwrChildReceiptNumber;
    }

    struct CommodityLotRecord {
        bytes32 lotHash;
        bytes32 commodityId;
        bytes32 vaultId;
        bytes32 barSerialNumberHash;
        bytes32 refineryCode;
        uint16 finenessPurity;
        uint256 grossWeightGrams;
        uint256 netPureWeightGrams;
        bytes32 assayCertificateDigest;
        bytes32 eNwrReceiptNumber;
        bytes32 parentLotHash;
        bytes32 lineageTreeRoot;
        uint32 splitGeneration;
        LotLifecycleState state;
        uint256 registrationTimestamp;
        uint256 lastAuditTimestamp;
    }

    event LotSplitRegistered(
        bytes32 indexed parentLotHash,
        bytes32 indexed lineageTreeRoot,
        uint32 splitGeneration,
        uint256 childLotCount,
        uint256 totalChildWeightGrams,
        uint256 scrapLossGrams,
        bytes32 refinerySplitCertDigest
    );

    event ChildLotCreated(
        bytes32 indexed childLotHash,
        bytes32 indexed parentLotHash,
        uint32 childIndex,
        uint256 netPureWeightGrams,
        bytes32 assayCertificateDigest
    );

    error ParentLotNotVaulted(bytes32 parentLotHash, LotLifecycleState currentState);
    error SplitWeightMismatch(uint256 parentWeight, uint256 totalChildWeight, uint256 scrapLoss);
    error MaxChildLotsExceeded(uint256 providedCount, uint256 maxAllowed);
    error DuplicateChildSerialHash(bytes32 childBarSerialHash);
    error LineageVerificationFailed(bytes32 childLotHash, bytes32 expectedParentLotHash);

    function splitCommodityLot(
        bytes32 parentLotHash,
        ChildLotSpec[] calldata childLots,
        uint256 scrapLossGrams,
        bytes32 refinerySplitCertDigest,
        bytes calldata refinerySignature,
        bytes calldata custodianSignature
    ) external returns (bytes32[] memory childLotHashes, bytes32 lineageTreeRoot);

    function getLotRecord(bytes32 lotHash) external view returns (CommodityLotRecord memory);

    function getChildLots(bytes32 parentLotHash) external view returns (bytes32[] memory childLotHashes);

    function verifyLotLineage(
        bytes32 childLotHash,
        bytes32 expectedRootParentLotHash,
        bytes32[] calldata lineageProof
    ) external view returns (bool isValid);
}
```

## Security & Compliance Notes
- **WDRA & SEBI Electronic Gold Receipt (EGR) Compliance:** All vault deposits, splits, and in-transit movements must comply with WDRA warehousing mandates and SEBI EGR regulations. Splitting an e-NWR on-chain requires verified corresponding subdivision in the repository (NERL/CCRL).
- **Strict Mathematical Scrap Bounding:** The maximum scrap tolerance of +/- 0.10% is enforced immutably on-chain. Contracts reject any transaction attempting to register or settle variance above 10 bps, protecting both investors and vault custodians against metallurgical fraud.
- **Zero-PII Biometric & OTP Verification Architecture:** All recipient verification operations use blinded cryptographic digests (`biometricProofHash`, `otpHashCommitment`). Biometric nullifiers prevent replay attacks across multiple deliveries without exposing biometric vectors or identities on the public ledger.
- **Dual-Control Attestations for In-Transit Handover:** Carrier dispatch and handoff require cryptographic signatures from both the authorized carrier agent key and the recipient, validated against EIP-712 structured data standards.
- **Atomic Permanent Burn Execution:** Tokens are burned permanently via `CommoditySecurityToken.burnForDelivery` within the same transaction execution frame as the verified handoff, ensuring that token supply and physical inventory decrease strictly in unison.
- **Reentrancy and Transient Storage Guards:** All state-changing functions in `CommodityScrapEqualizer.sol` and `InTransitEscrowRegistry.sol` employ OpenZeppelin `ReentrancyGuardUpgradeable` to eliminate cross-function reentrancy risks during token and cash settlements.

## Acceptance Criteria
- [ ] `ICommodityScrapEqualizer.sol`, `IInTransitEscrowRegistry.sol`, and `IPhysicalVaultRegistryLineage.sol` compile cleanly with Solidity 0.8.24 with zero compiler warnings.
- [ ] Automated scrap tolerance enforcement correctly accepts variances <= +/- 0.10% and reverts with `ScrapToleranceExceeded` for any variance > 0.10%.
- [ ] Spot cash equalization calculations accurately compute INR/e-Rupee refund credits for underweight bars and surcharge debits for overweight bars based on real-time oracle prices.
- [ ] `splitCommodityLot` verifies weight conservation, decommissions parent lots with status `SPLIT_DECOMMISSIONED`, registers child lots, and emits deterministic lineage tree roots.
- [ ] `InTransitEscrowRegistry` manages the complete delivery lifecycle from escrow creation through carrier dispatch to final handoff.
- [ ] Delivery handoff and atomic token burn execute strictly when both valid biometric nullifier and time-bounded OTP code preimages are verified on-chain.
- [ ] Replay attacks using previously executed biometric nullifiers or expired OTP codes are rejected with custom errors.
- [ ] Packages exceeding maximum delivery attempts (3) can be securely returned to the vault with `initiateReturnToVault` and `confirmVaultReturn`.
- [ ] 100% test coverage across Foundry unit, fuzz, and invariant suites in `test/commodities/`.
- [ ] Slither static analysis returns zero high, medium, or reentrancy issues across the entire codebase.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance Smart Contract), Prompt `304` (Token Redemption Smart Contract), Prompt `307` (MultiSig Governance Smart Contract), Prompt `328` (Decentralized Oracle Aggregation Smart Contract), Prompt `330` (MCX Commodity Token and WDRA Physical Vault Registry Smart Contracts).
- **Parallel Tasks:** Prompt `243` (MCX Commodity & Warehouse Receipt Adapter), Prompt `213` (Custodian Depository Integration Service), Prompt `203` (Wallet & Double-Entry Ledger Service).
- **Subsequent Prompts Enabled:** Prompt `530` (Flutter Commodity Physical Delivery Flow), Prompt `608` (Commodity Vault & Physical Delivery Web Portal), Prompt `242` (NSE/BSE Market Data & Order Routing Adapter).
