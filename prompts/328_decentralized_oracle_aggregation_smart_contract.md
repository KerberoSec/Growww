# 328 - Decentralized Multi-Source Oracle Aggregator Smart Contract Suite (OracleAggregator.sol)

## Purpose
Accurate, low-latency, and tamper-resistant valuation of underlying collateral, fractional tokenized equities, foreign exchange (FX) conversion rates, and cross-border settlement indexes is foundational to the Growww ecosystem. Relying on a single centralized price oracle introduces single-point-of-failure vulnerabilities, manipulation vectors (e.g. flash loan spot distortions), and severe stale-pricing risks during extreme market volatility or exchange circuit halt conditions across Indian domestic exchanges (NSE/BSE) and international trading venues (GIFT City IFSC).

This prompt specifies the design, architecture, interface definitions, and security requirements for the **Decentralized Multi-Source Oracle Aggregator Smart Contract Suite** (`OracleAggregator.sol`, `PythPriceReceiver.sol`, and `ChainlinkPriceReceiver.sol`) deployed on the permissioned Hyperledger Besu consortium ledger. The suite coordinates low-latency push and pull pricing feeds across asset classes (equities, crypto/reserve assets, and FX currency pairs), performs robust median-of-N consensus with statistical outlier filtering, enforces rigorous heartbeat staleness guards, validates confidence intervals, and triggers autonomous circuit breaker halts upon anomalous volatility spikes or inter-oracle divergence.

## What You Are Building
A production-grade, gas-optimized Solidity smart contract suite under `contracts/oracle/` consisting of:
- `OracleAggregator.sol`: Central aggregation and safety gateway calculating composite normalized prices (18-decimal fixed-point precision), filtering statistical outliers, computing median consensus across registered adapters, verifying heartbeat freshness, and tripping automatic circuit breakers.
- `PythPriceReceiver.sol`: Specialized pull-oracle adapter receiver interfacing with the Pyth Network EVM contract, verifying Wormhole cryptographic attestation payloads (VAAs), decoding variable exponent fixed-point values, validating confidence intervals (`conf / price`), and normalizing to 18 decimals.
- `ChainlinkPriceReceiver.sol`: Push-oracle adapter receiver querying Chainlink AggregatorV3 data feeds, verifying round completion integrity (`answeredInRound >= roundId`), checking timestamp staleness against per-asset heartbeats, rejecting negative/zero prices, and scaling 8-decimal feeds to standard precision.
- `IOracleAggregator.sol`, `IPythPriceReceiver.sol`, `IChainlinkPriceReceiver.sol`: Complete interface definitions, custom error declarations, events, and data structures.
- Comprehensive Foundry unit, fuzz, and invariant test specifications (`test/oracle/OracleAggregator.t.sol`) validating median calculations, outlier rejection, confidence boundary rejections, staleness trips, and emergency multisig failovers.

## Scope Boundaries
- **In Scope:**
  - Composite median-of-N aggregation across multiple heterogeneous oracle sources for Equities (e.g. `GROWWW-RELIANCE-INE002A01018`, `GROWWW-TCS-INE467B01029`), FX currency pairs (USD/INR, EUR/INR, GBP/INR, AED/INR), and Digital Reserve Assets (BTC/USD, ETH/USD, USDC/USD).
  - Outlier filtering rejecting price sources that deviate beyond configurable basis points (e.g. >150 bps) from the inter-oracle median.
  - Granular staleness guards enforcing max-staleness heartbeats tailored to asset volatility (e.g. 10s for equities during market hours, 5s for crypto, 10s for FX).
  - Confidence interval threshold validation for Pyth Network feeds, ensuring wide-spread or illiquid feeds are rejected before poisoning composite indexes.
  - Configurable dynamic price collars and volatility circuit breakers halting automated settlements and contract mints/burns when price movements exceed allowable bands within a rolling time window.
  - Multi-party governance authorization hooks for emergency pause, collar adjustments, and feed provider additions (Prompt 307).
  - Decimal normalization converting non-uniform oracle outputs (e.g. Chainlink 8 decimals, Pyth dynamic `expo`) into standard 18-decimal precision (`1e18`).
- **Out of Scope / Handled Elsewhere:**
  - Off-chain market data ingestion and streaming pipeline (handled in Prompt 207).
  - Trade matching engine price-time priority execution (handled in Prompt 205).
  - Realized profit and fee calculation engine (handled in Prompt 210).
  - Off-chain circuit breaker coordination service across microservices (handled in Prompt 713).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Shanghai/Cancun with native transient storage `TSTORE`/`TLOAD` support and checked arithmetic).
  *Justification:* Solidity 0.8.24 guarantees native overflow/underflow protection, custom error definitions for gas efficiency, user-defined value types, and deterministic assembly for fixed-point math operations.
- **Contract Libraries:** OpenZeppelin Contracts Upgradeable v5.0 (AccessControlUpgradeable, PausableUpgradeable, Initializable, ERC1967UpgradeUpgradeable).
- **Oracle SDK Integrations:** Pyth Network SDK contracts (`@pythnetwork/pyth-sdk-solidity`) and Chainlink Aggregator interfaces (`@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol`).
- **Toolchain:** Foundry (`forge` for compilation, fuzzing, and invariant testing; `cast` for Besu RPC contract interactions).
- **Static Analysis & Formal Verification:** Slither, Mythril, and Halmos for mathematical invariant proofs.

## Backend / Infra Touchpoints
- **Market Data Service (Prompt 207):** Microservice relaying off-chain L1/L2 Pyth price update byte arrays (Wormhole VAAs) to `PythPriceReceiver.sol` via batched RPC transactions.
- **Settlement DvP Smart Contract (Prompt 306):** Ingests spot prices and FX conversion rates from `OracleAggregator.sol` to validate trade settlement values and verify statutory INR/USD settlement thresholds.
- **Token Issuance & Redemption Contracts (Prompts 303, 304):** Query `OracleAggregator.sol` to establish accurate Net Asset Value (NAV) per fractional share prior to executing mint or burn operations.
- **Continuous Market Circuit Breaker Coordinator (Prompt 713):** Synchronizes off-chain market halt signals with on-chain oracle circuit breaker state.
- **Multisig Governance Smart Contract (Prompt 307):** Acts as `DEFAULT_ADMIN_ROLE` managing oracle feeds, emergency unpausing, and risk parameters.

## Blockchain Interaction
- **Oracle Price Ingestion (Pull Model):** Relayers invoke `updateAndGetPrice(bytes32 assetId, bytes[] calldata priceUpdateData)` on `OracleAggregator.sol`, forwarding Pyth update blobs with necessary execution fees to the Pyth core contract before resolving composite prices.
- **Oracle Price Ingestion (Push Model):** `ChainlinkPriceReceiver.sol` directly reads latest round data from deployed on-chain Chainlink feeds on the permissioned Besu network.
- **Aggregation & Valuation Queries:** Downstream smart contracts (`SettlementDvP.sol`, `ProofOfReserveRegistry.sol`) invoke read-only `getPrice(bytes32 assetId)` or `getDetailedPrice(bytes32 assetId)` to receive verified, outlier-filtered, normalized prices and timestamps.
- **Circuit Breaker Emission:** If price deviation between consecutive updates exceeds `circuitBreakerBps` or if inter-oracle deviation exceeds `maxDeviationBps`, `OracleAggregator.sol` flips status to `CIRCUIT_BROKEN`, reverts execution for consuming settlement contracts, and emits `CircuitBreakerTriggered`.
- **Zero PII on Chain:** All oracle configurations, assets, and prices use purely cryptographic identifiers (`bytes32 assetId = keccak256("EQUITY:NSE:RELIANCE")`), mathematical values, and public timestamps.

## Step-by-Step Build Instructions
1. Scaffold oracle directory structure under `contracts/oracle/`:
   - `src/OracleAggregator.sol`, `src/receivers/PythPriceReceiver.sol`, `src/receivers/ChainlinkPriceReceiver.sol`.
   - `src/interfaces/IOracleAggregator.sol`, `src/interfaces/IPythPriceReceiver.sol`, `src/interfaces/IChainlinkPriceReceiver.sol`.
   - `test/oracle/OracleAggregator.t.sol`, `test/oracle/PythPriceReceiver.t.sol`, `test/oracle/ChainlinkPriceReceiver.t.sol`.
2. Define custom interfaces, data structures, enumerations, events, and custom error types across all interface files.
3. Implement `PythPriceReceiver.sol` inheriting `Initializable` and `AccessControlUpgradeable`:
   - Store Pyth core contract address (`IPyth`) and asset feed ID mappings.
   - Implement `getLatestPrice(bytes32 assetId)`: call `pyth.getPriceNoOlderThan(feedId, maxStaleness)`, validate confidence interval (`uint64 conf <= (price * maxConfidenceRatioBps) / 10000`), convert negative/positive `expo` to target 18-decimal fixed-point precision, and return normalized price.
   - Implement `updatePriceFeeds(bytes[] calldata updateData)` payable to execute Pyth batch pull updates.
4. Implement `ChainlinkPriceReceiver.sol` inheriting `Initializable` and `AccessControlUpgradeable`:
   - Store Chainlink `AggregatorV3Interface` contract addresses per asset ID.
   - Implement `getLatestPrice(bytes32 assetId)`: call `latestRoundData()`, verify `price > 0`, `updatedAt != 0`, `block.timestamp - updatedAt <= maxStaleness`, and `answeredInRound >= roundId`.
   - Normalize output decimals from feed decimals (e.g. 8) to standard target 18 decimals.
5. Implement `OracleAggregator.sol` inheriting `Initializable`, `PausableUpgradeable`, and `AccessControlUpgradeable`:
   - Maintain `mapping(bytes32 => AssetFeedConfig) public assetConfigs` and `mapping(bytes32 => PriceHistory) internal priceHistories`.
   - Implement asset configuration method `configureAssetFeed(...)` restricted to `ORACLE_MANAGER_ROLE` (governed by multisig).
6. Implement composite median and outlier filtering algorithm in `OracleAggregator.sol`:
   - Query all active registered price receivers for the given `assetId`.
   - Filter out reverted, stale, or unconfident feeds into a memory array of valid prices.
   - Verify count of valid prices is greater than or equal to `config.minimumSources`.
   - Sort valid prices in ascending order in memory using an optimized insertion sort algorithm (for $N \le 7$).
   - Calculate median: if odd, take middle element; if even, take arithmetic average of two middle elements.
   - Filter statistical outliers: reject any individual price where `|price - median| * 10000 / median > config.maxDeviationBps`.
   - Recalculate filtered median from remaining non-outlier sources.
7. Implement price collar and circuit breaker evaluation:
   - Compare filtered median against `lastValidPrice`.
   - If `|median - lastValidPrice| * 10000 / lastValidPrice > config.circuitBreakerBps`, trigger circuit breaker: set asset status to `CIRCUIT_BROKEN`, emit `CircuitBreakerTriggered`, and revert with `CircuitBreakerTripped`.
   - If price violates absolute bounds (`median < minPrice` or `median > maxPrice`), revert with `InvalidPriceRange`.
   - Update `lastValidPrice` and `lastValidTimestamp`.
8. Implement `updateAndGetPrice(bytes32 assetId, bytes[] calldata priceUpdateData)` payable to allow atomic update-and-read workflows for high-frequency settlement contracts.
9. Implement emergency circuit breaker controls:
   - `tripCircuitBreaker(bytes32 assetId, string calldata reason)` callable by `CIRCUIT_BREAKER_ROLE` or automated risk bots.
   - `resumeFeed(bytes32 assetId)` requiring multi-sig governance authorization (`DEFAULT_ADMIN_ROLE`).
   - `setEmergencyFallbackPrice(bytes32 assetId, uint256 fallbackPrice)` guarded by multi-sig timelock.
10. Write Foundry unit tests covering:
    - Normal price ingestion and 18-decimal normalization across Pyth and Chainlink mocks.
    - Single-source failure tolerance where 1 oracle fails or goes stale but remaining sources maintain consensus.
    - Outlier rejection when 1 oracle reports a manipulated or divergent price (>150 bps).
    - Staleness rejection when feed heartbeat exceeds `maxStalenessSeconds`.
    - Pyth confidence ratio rejection when `conf` exceeds allowable limit.
11. Write Foundry fuzz and invariant tests:
    - Invariant: composite median is always bounded between minimum and maximum valid source inputs.
    - Fuzzing random prices, exponents, and confidence intervals to ensure zero math overflows or panics.
12. Run Slither and Mythril static analysis to ensure 0 critical/high/medium security findings.
13. Write deployment scripts in Foundry (`script/DeployOracleAggregator.s.sol`) configured for Hyperledger Besu QBFT RPC endpoint and deploy to local testnet.

## Interfaces / Contracts

### Oracle Aggregator Interface (`IOracleAggregator.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IOracleAggregator {
    enum AssetType {
        EQUITY,
        CRYPTO,
        FX,
        COMMODITY
    }

    enum OracleStatus {
        ACTIVE,
        FROZEN,
        CIRCUIT_BROKEN,
        STALE,
        DEGRADED
    }

    struct AssetFeedConfig {
        bytes32 assetId;
        string symbol;
        AssetType assetType;
        address[] primaryReceivers;
        uint8 minimumSources;
        uint32 maxStalenessSeconds;
        uint32 maxConfidenceRatioBps;
        uint32 maxDeviationBps;
        uint32 circuitBreakerBps;
        uint256 minPrice;
        uint256 maxPrice;
        bool isPaused;
    }

    struct AggregatedPrice {
        uint256 price;
        uint256 confidence;
        uint64 publishTime;
        uint8 sourcesCount;
        OracleStatus status;
    }

    struct PriceHistory {
        uint256 lastValidPrice;
        uint64 lastValidTimestamp;
        OracleStatus status;
    }

    // Events
    event PriceUpdated(
        bytes32 indexed assetId,
        uint256 price,
        uint256 confidence,
        uint64 publishTime,
        uint8 sourcesCount
    );
    event AssetFeedConfigured(
        bytes32 indexed assetId,
        string symbol,
        AssetType indexed assetType,
        uint8 minimumSources,
        uint32 maxStalenessSeconds
    );
    event CircuitBreakerTriggered(
        bytes32 indexed assetId,
        uint256 attemptedPrice,
        uint256 referencePrice,
        uint32 deviationBps,
        string reason
    );
    event OutlierDetected(
        bytes32 indexed assetId,
        address indexed receiver,
        uint256 sourcePrice,
        uint256 medianPrice,
        uint32 deviationBps
    );
    event FeedStatusChanged(bytes32 indexed assetId, OracleStatus oldStatus, OracleStatus newStatus);
    event OracleReceiverAdded(bytes32 indexed assetId, address indexed receiver);
    event OracleReceiverRemoved(bytes32 indexed assetId, address indexed receiver);

    // Errors
    error StalePriceFeed(bytes32 assetId, uint64 publishTime, uint32 maxStaleness);
    error ConfidenceExceeded(bytes32 assetId, uint256 confidence, uint256 maxAllowed);
    error InsufficientValidSources(bytes32 assetId, uint8 validSources, uint8 requiredSources);
    error DeviationExceeded(bytes32 assetId, uint256 sourcePrice, uint256 medianPrice, uint32 deviationBps);
    error CircuitBreakerTripped(bytes32 assetId, uint256 attemptedPrice, uint256 lastPrice, uint32 deviationBps);
    error InvalidPriceRange(bytes32 assetId, uint256 price, uint256 minPrice, uint256 maxPrice);
    error ZeroAddressNotAllowed();
    error UnauthorizedCaller(address caller);
    error InvalidAssetConfiguration(bytes32 assetId);
    error FeedPaused(bytes32 assetId);
    error ZeroPriceReported(bytes32 assetId, address receiver);

    // Read Functions
    function getPrice(bytes32 assetId) external view returns (uint256 price, uint64 publishTime, uint256 confidence);
    function getDetailedPrice(bytes32 assetId) external view returns (AggregatedPrice memory);
    function getAssetConfig(bytes32 assetId) external view returns (AssetFeedConfig memory);
    function getPriceHistory(bytes32 assetId) external view returns (PriceHistory memory);

    // Mutating Functions
    function updateAndGetPrice(
        bytes32 assetId,
        bytes[] calldata priceUpdateData
    ) external payable returns (uint256 price);

    function configureAssetFeed(AssetFeedConfig calldata config) external;
    function addOracleReceiver(bytes32 assetId, address receiver) external;
    function removeOracleReceiver(bytes32 assetId, address receiver) external;
    function tripCircuitBreaker(bytes32 assetId, string calldata reason) external;
    function resumeFeed(bytes32 assetId) external;
    function setEmergencyFallbackPrice(bytes32 assetId, uint256 fallbackPrice) external;
}
```

### Pyth Price Receiver Interface (`IPythPriceReceiver.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IPythPriceReceiver {
    struct PythFeedConfig {
        bytes32 priceFeedId;
        uint32 maxStalenessSeconds;
        uint32 maxConfidenceRatioBps;
        uint8 targetDecimals;
        bool isActive;
    }

    event PythPriceUpdated(bytes32 indexed assetId, bytes32 indexed priceFeedId, uint256 normalizedPrice, uint64 publishTime);
    event PythFeedConfigured(bytes32 indexed assetId, bytes32 indexed priceFeedId, uint32 maxStalenessSeconds);

    error InvalidPythPrice(bytes32 assetId, int64 rawPrice);
    error PythPriceStale(bytes32 assetId, uint256 publishTime, uint32 maxStaleness);
    error PythConfidenceTooWide(bytes32 assetId, uint64 confidence, int64 rawPrice, uint32 ratioBps);
    error PythUpdateFeeFailed();
    error UnauthorizedCaller(address caller);
    error FeedNotActive(bytes32 assetId);

    function getLatestPrice(
        bytes32 assetId
    ) external view returns (uint256 normalizedPrice, uint256 confidence, uint64 publishTime);

    function updatePriceFeeds(bytes[] calldata updateData) external payable;
    function getUpdateFee(bytes[] calldata updateData) external view returns (uint256 fee);
    function configurePythFeed(bytes32 assetId, PythFeedConfig calldata config) external;
    function pythContract() external view returns (address);
}
```

### Chainlink Price Receiver Interface (`IChainlinkPriceReceiver.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IChainlinkPriceReceiver {
    struct ChainlinkFeedConfig {
        address feedAddress;
        uint32 maxStalenessSeconds;
        uint8 feedDecimals;
        uint8 targetDecimals;
        bool isActive;
    }

    event ChainlinkFeedConfigured(bytes32 indexed assetId, address indexed feedAddress, uint32 maxStalenessSeconds);

    error InvalidChainlinkRound(bytes32 assetId, uint80 roundId, uint80 answeredInRound);
    error ChainlinkPriceStale(bytes32 assetId, uint256 updatedAt, uint32 maxStaleness);
    error NegativeOrZeroPrice(bytes32 assetId, int256 price);
    error UnauthorizedCaller(address caller);
    error FeedNotActive(bytes32 assetId);

    function getLatestPrice(
        bytes32 assetId
    ) external view returns (uint256 normalizedPrice, uint256 confidence, uint64 publishTime);

    function configureChainlinkFeed(bytes32 assetId, ChainlinkFeedConfig calldata config) external;
    function getFeedConfig(bytes32 assetId) external view returns (ChainlinkFeedConfig memory);
}
```

## Security & Compliance Notes
- **Outlier Rejection & Flash Loan Manipulation Resistance:** The median-of-N algorithm requires at least $M \ge 3$ independent active oracles for sensitive equities and reserve assets. Any feed deviating beyond `maxDeviationBps` (e.g. 150 bps) is excluded from final consensus, preventing single-exchange or single-node price manipulation.
- **Heartbeat & Staleness Enforcement:** Strict time validation (`block.timestamp - publishTime <= maxStalenessSeconds`) prevents execution against obsolete price data during network congestion or provider outages. Reverts protect users and liquidity providers from stale-arbitrage attacks.
- **Circuit Breakers & Limit Bands:** Price changes exceeding `circuitBreakerBps` (e.g. 1000 bps / 10% in a single update) trigger an immediate on-chain trading halt for the asset, requiring compliance and risk multi-sig authorization (Prompt 307) to reset.
- **Multi-Signature Access Control:** All sensitive parameters (adding/removing oracle receivers, modifying deviation tolerances, updating price collars) are restricted to institutional multisig governance requiring 3-of-5 signatures.
- **Confidence Interval Validation:** Pyth pricing confidence metrics (`conf`) are verified on-chain against `maxConfidenceRatioBps` to protect against illiquid market conditions or disrupted off-chain order books.
- **Mathematical Safety & Determinism:** All decimal scaling and division arithmetic enforce non-zero denominators and checked math under Solidity 0.8.24 to eliminate division-by-zero panics and truncation exploits.

## Acceptance Criteria
- [ ] `OracleAggregator.sol`, `PythPriceReceiver.sol`, and `ChainlinkPriceReceiver.sol` fully implemented adhering to specified interfaces.
- [ ] Decimal normalization converts heterogeneous oracle decimals (8 decimals Chainlink, dynamic `expo` Pyth) to standard 18 decimals (`1e18`) with zero precision loss.
- [ ] Median-of-N aggregation correctly computes mathematical median across odd and even source sets.
- [ ] Outlier filtering reliably rejects feeds exceeding `maxDeviationBps` without reverting if remaining valid sources meet `minimumSources`.
- [ ] Staleness check strictly reverts with `StalePriceFeed` if `block.timestamp - publishTime > maxStalenessSeconds`.
- [ ] Pyth confidence validation reverts with `PythConfidenceTooWide` if `conf / price > maxConfidenceRatioBps`.
- [ ] Circuit breaker trips automatically and emits `CircuitBreakerTriggered` when price movement exceeds `circuitBreakerBps`.
- [ ] 100% test coverage achieved across unit, edge case, and fuzzing test suites in Foundry.
- [ ] Slither and Mythril static analysis passes with 0 high/medium severity findings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Permissioned Blockchain Platform Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `307` (Multi-Party Authorization & Multisig Governance).
- **Parallel Tasks:** Prompt `207` (Market Data Service), Prompt `713` (Continuous Market Circuit Breaker Coordinator).
- **Subsequent Prompts Enabled:** Prompt `303` (Token Issuance Smart Contract), Prompt `306` (Atomic DvP Settlement Smart Contract), Prompt `308` (On-Chain Proof of Reserve Publishing).
