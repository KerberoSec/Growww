# 348 - Virtual Faucet & Demo Token Minting Smart Contract (Solidity)

## Purpose
In electronic securities trading and digital asset ecosystems, paper trading (simulated mock trading) serves as a critical sandbox for retail investor education, institutional algo backtesting, pre-trade risk testing, and developer integration. On the Growww National Blockchain Stock Exchange (NBSE), production settlement operates under an atomic Delivery-versus-Payment (DvP Model 1) architecture on a permissioned Hyperledger Besu consortium ledger. Real production trades involve real fiat-backed wrapped electronic Indian Rupees (weINR) held in scheduled commercial bank escrow accounts, regulated stablecoins (USDT) held in IFSCA GIFT City vaults, and fractional equity security tokens backed 1:1 by depository receipts (NSDL/CDSL).

To allow prospective retail traders, financial engineering students, algo trading firms, and API developers to practice order routing, margin management, algorithmic execution, and portfolio optimization without risking real capital, the platform provides a sandboxed paper trading environment. However, deploying demo tokens on a blockchain introduces unique architectural and economic challenges:
1. **Grey-Market & Secondary Speculation Risk:** If virtual tokens can be transferred freely between arbitrary external wallets, malicious actors can create secondary off-market peer-to-peer (P2P) trading venues, scam inexperienced users by masquerading demo tokens as real assets, or run unauthorized OTC pools.
2. **Mainnet Contamination Risk:** If a virtual faucet or mock token contract is accidentally or maliciously deployed or executed on the production Growww Mainnet (Chain ID 2026), it could compromise accounting integrity or confuse financial reconciliation pipelines.
3. **Sybil & Node Resource Exhaustion:** Unchecked faucet dispensing can lead to testnet state bloat, storage spam, and denial-of-service (DoS) on Besu testnet nodes.
4. **Gas Friction for Novice Users:** Forcing novice paper traders to first acquire testnet native gas tokens before they can claim paper trading balances creates a major onboarding bottleneck.

This prompt specifies the **Virtual Faucet & Demo Token Minting Smart Contract (`VirtualFaucet.sol`, `VirtualTokens.sol`, `IVirtualFaucet.sol`, `IVirtualToken.sol`)**. Deployed under `contracts/src/demo/` exclusively on the Hyperledger Besu Testnet (Chain ID 13371), this contract suite issues non-transferable, soulbound mock assets (`vUSDT`, `vBTC`, `vINR`, `vETH`), enforces strict rate limiting and daily dispensing quotas, restricts token transfers exclusively to registered Demo Exchange smart contracts (such as the Demo Matching Engine and Demo Settlement DvP), provides administrative and user-initiated portfolio reset hooks, and supports gasless claiming via ERC-4337 Account Abstraction Paymasters.

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/src/demo/` comprising:
- `VirtualFaucet.sol`: The core automated faucet contract controlling token dispensing, rate-limiting windows, tiered claim quotas, gasless claim permits, administrative emergency halts, and user portfolio resets.
- `IVirtualFaucet.sol`: The master Solidity interface defining all faucet data structures, quota tiers, claim permit types, events, and external function signatures.
- `VirtualTokens.sol`: Standardized, upgradeable mock token implementations (`vUSDT`, `vBTC`, `vINR`, `vETH`) featuring soulbound transfer restrictions that allow token balance movement exclusively between user wallets and authorized Demo Exchange contracts.
- `IVirtualToken.sol`: The interface specifying custom demo token functionality including authorized exchange whitelisting, privileged faucet minting, privileged burning, and full balance resets.
- Multi-Asset Dispensing Engine: Capability to dispense coordinated bundles of virtual currencies (e.g., 100,000 vINR cash balance, 1,000 vUSDT margin balance, 0.5 vBTC asset balance) in a single atomic transaction.
- Soulbound Transfer Barrier: A runtime transfer hook overriding `_update` (OpenZeppelin v5) that intercepts all ERC-20 transfers. Transfers between arbitrary retail wallets revert immediately, while transfers to or from authorized Demo Matching Engine and Demo Settlement contracts execute seamlessly.
- Strict Hardcoded Chain-ID Guardrail: Both the faucet and virtual tokens enforce an immutable runtime and deployment check asserting `block.chainid == 13371` (Besu Testnet). Any attempt to deploy or execute on Mainnet (Chain ID 2026) immediately reverts with `InvalidNetworkEnvironment()`.
- Gasless ERC-4337 & EIP-712 Permitted Minting: Native integration with the Testnet Paymaster and EIP-712 structured claim signatures (`ClaimPermit`), enabling zero-gas token claims for paper traders.
- Portfolio Reset Hooks: Administrative and self-service reset functions that allow traders to burn all current virtual holdings and reset their demo balances back to clean baseline figures.
- Comprehensive Foundry Test Suite (`test/demo/VirtualFaucet.t.sol`, `test/demo/VirtualTokens.t.sol`): Exhaustive unit, fuzz, invariant, and integration tests verifying rate-limiting, anti-transfer restrictions, chain ID assertions, and gasless claim flows.
- Deployment Script (`script/DeployVirtualFaucet.s.sol`): Hardened Foundry deployment and initialization script targeting Hyperledger Besu Testnet.

## Scope Boundaries
- **In Scope:**
  - Deployment of non-transferable, isolated demo tokens (`vUSDT`, `vBTC`, `vINR`, `vETH`) on Hyperledger Besu Testnet.
  - Automated faucet dispensing with configurable per-user cooldown periods and tier-based dispensing quotas (Retail, Algo Trader, Institutional Sandbox).
  - Soulbound transfer restrictions: transfers permitted strictly between user wallets and approved Demo Exchange contracts (e.g., Demo Settlement DvP, Demo Matching Engine escrow). Arbitrary P2P transfers (`userA -> userB`) are strictly rejected.
  - Strict network isolation: assertion that contract execution occurs exclusively on Testnet (`block.chainid == 13371`).
  - Self-service and administrative portfolio reset mechanics (burn existing demo balances and re-mint baseline balances).
  - EIP-712 meta-transaction permit support (`claimWithPermit`) for gasless user onboarding via testnet relayers and ERC-4337 Paymasters.
  - Global dispensing velocity limits (maximum total tokens dispensed per asset per 24-hour epoch) to prevent testnet ledger bloat.
  - Foundry test harness covering invariant fuzzing, transfer restriction violations, and cooldown enforcement.
- **Out of Scope / Handled Elsewhere:**
  - Real-money fiat bank onboarding and UPI/IMPS payment gateways (handled in Prompt 212 and Prompt 203).
  - Off-chain demo order book matching and matching engine execution (handled in Prompt 273).
  - Demo wallet balance visualization and off-chain portfolio tracking services (handled in Prompt 274).
  - Production Delivery-versus-Payment settlement contracts (handled in Prompt 306 and Prompt 329).
  - Production KYC verification and Aadhaar/PAN compliance pipelines (handled in Prompt 202).
  - Mainnet ERC-4337 Paymaster and UserOperation sponsorship (handled in Prompt 259).
  - Real-world asset custody attestation and Proof of Reserves (handled in Prompt 327 and Prompt 606).

| Out-of-Scope Component | Responsible System | Relevant Prompt |
| :--- | :--- | :--- |
| Demo Matching Engine & Simulated Order Book | Demo Matching Engine | Prompt 273 |
| Demo Wallet Service & Portfolio Cache | Demo Wallet Service | Prompt 274 |
| Production DvP Settlement Contract | SettlementDvP.sol / NBSEFeeCollector.sol | Prompt 306 / Prompt 329 |
| Production KYC/AML Verifier & Registry | KYC/AML Service / ERC-3643 Identity | Prompt 202 / Prompt 304 |
| Production Custodial Fiat Reserves (weINR) | Custodian Depository / Bank Escrow | Prompt 203 / Prompt 213 |
| Production Paymaster & Gas Sponsorship | Mainnet Paymaster Service | Prompt 259 / Prompt 337 |
| Mobile Sandbox Toggle & Paper Trading UI | Flutter Sandbox Mode Switcher | Prompt 527 |

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with transient storage opcodes `TSTORE`/`TLOAD`, custom errors, and native checked arithmetic).
- **Base Frameworks & Libraries:**
  - OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `PausableUpgradeable`, `ReentrancyGuardUpgradeable`, `ERC20Upgradeable`).
  - OpenZeppelin Contracts v5.0 `SafeERC20` for robust token transfers.
  - OpenZeppelin Contracts v5.0 `ECDSA` and `EIP712Upgradeable` for typed structured message verification on gasless claim permits.
- **Target Network & Environment:**
  - **Hyperledger Besu Testnet:** Permissioned consortium testnet running QBFT consensus, 2-second block period, deterministic single-block finality, Chain ID `13371`.
  - **Mainnet Isolation Assertion:** Hardcoded check against Mainnet Chain ID `2026`.
- **Token Decimals & Specifications:**
  - `vINR` (Virtual Indian Rupee): 18 decimals, baseline paper trading currency.
  - `vUSDT` (Virtual USD Tether): 6 decimals, GIFT City simulated derivative and cross-border currency.
  - `vBTC` (Virtual Bitcoin): 8 decimals, crypto derivative paper trading underlying.
  - `vETH` (Virtual Ethereum): 18 decimals, crypto derivative paper trading underlying.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, gas tracking, and invariant fuzzing; `cast` for RPC testing on Besu testnet nodes).
- **Static Analysis & Auditing:** Slither (Trail of Bits) and Solhint CI linting.

## Backend / Infra Touchpoints
- **Demo Wallet Service (Prompt 274):** Coordinates user paper trading balances, listens to faucet mint events, submits gasless claim permits to testnet relayers on user behalf, and manages demo portfolio snapshots.
- **Demo Matching Engine (Prompt 273):** High-speed in-memory simulated order book executing limit, market, and stop orders with virtual tokens. Acts as the registered authorized exchange contract allowed to transfer soulbound virtual tokens between accounts for trade settlement and margin locking.
- **ERC-4337 Paymaster Service (Prompt 259 / Prompt 337):** Sponsoring testnet gas fees for retail users invoking `claimTokens` or `claimWithPermit`, ensuring paper traders never need testnet ETH/native gas.
- **API Gateway & BFF (Prompt 219):** Exposes `/api/v1/demo/faucet/claim` and `/api/v1/demo/portfolio/reset` endpoints, enforcing IP-based and user-session rate limits before forwarding transactions to the blockchain relayer.
- **Flutter Environment Switcher & Sandbox Mode (Prompt 527):** Client application toggle allowing users to switch from live trading to paper trading mode, automatically binding the UI to the testnet RPC endpoint and virtual token contracts.
- **Blockchain Event Indexer (Prompt 309):** Ingests `TokensDispensed`, `PortfolioReset`, and `ExchangeAuthorizationUpdated` events from the Besu Testnet, maintaining sub-second balance caches in Redis and PostgreSQL read-replicas.

## Blockchain Interaction (permissioned Hyperledger Besu Testnet with isolated demo tokens, zero PII, QBFT)
- **Consensus & Block Finality:** Deployed on Hyperledger Besu Testnet operating under QBFT consensus with 2-second block intervals and instant single-block transaction finality (zero reorganizations).
- **Chain ID Guardrail:** During deployment and in every state-altering execution, contracts assert:
  ```solidity
  if (block.chainid != TESTNET_CHAIN_ID) revert InvalidNetworkEnvironment(block.chainid, TESTNET_CHAIN_ID);
  ```
  Where `TESTNET_CHAIN_ID = 13371`. This guarantees that if bytecode is mistakenly deployed on Mainnet (`block.chainid == 2026`), constructor initialization fails instantly, and any calls revert.
- **Soulbound Exchange Isolation:** In `VirtualTokens.sol`, the internal ERC-20 `_update(address from, address to, uint256 amount)` method enforces:
  1. Minting (`from == address(0)`): Allowed only if `msg.sender == virtualFaucetAddress` or caller has `MINTER_ROLE`.
  2. Burning (`to == address(0)`): Allowed if `msg.sender == virtualFaucetAddress`, `from == msg.sender`, or caller has `BURNER_ROLE`.
  3. Transfer (`from != address(0) && to != address(0)`): Allowed **only** if either `from` or `to` is a whitelisted Demo Exchange contract (e.g., Demo Settlement DvP, Demo Margin Vault). Direct P2P transfers between user wallets revert with `TokenTransferRestricted()`.
- **Zero On-Chain PII Guarantee:** No user identities, PANs, Aadhaar numbers, legal names, or email addresses are stored on-chain. Faucet claims track beneficiary EVM addresses (`address user`) and optional salted user identity hashes (`bytes32 userIdHash = keccak256(abi.encodePacked(userId, salt))`).
- **Gasless Claim Architecture:** 
  1. Off-chain: Trader clicks "Claim 100,000 vINR Demo Funds" in the Growww app.
  2. The Demo Wallet Service constructs an EIP-712 `ClaimPermit` digest and requests user signature, or formats an ERC-4337 `UserOperation`.
  3. The transaction is submitted to the Besu Testnet via the Growww Testnet Paymaster, which sponsors 100% of the gas.
  4. `VirtualFaucet.sol` verifies the permit or caller, checks rate limits, and invokes privileged `mint` on `vINR`, `vUSDT`, etc.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Directory & Dependency Structure:** Set up Foundry workspace under `contracts/`:
   - `contracts/src/demo/VirtualFaucet.sol`
   - `contracts/src/demo/VirtualTokens.sol`
   - `contracts/src/demo/interfaces/IVirtualFaucet.sol`
   - `contracts/src/demo/interfaces/IVirtualToken.sol`
   - `test/demo/VirtualFaucet.t.sol`
   - `test/demo/VirtualTokens.t.sol`
   - `script/DeployVirtualFaucet.s.sol`
2. **Define Master Interfaces & Custom Errors:** Create `IVirtualFaucet.sol` and `IVirtualToken.sol` declaring data structures (`QuotaTier`, `ClaimRequest`, `ClaimPermit`), events (`TokensDispensed`, `PortfolioReset`, `ExchangeAuthorized`, `RateLimitUpdated`), and custom error codes (`CooldownActive`, `QuotaExceeded`, `TransferRestricted`, `InvalidNetworkEnvironment`, `UnauthorizedExchange`).
3. **Implement Virtual Token Base Contract (`VirtualTokens.sol`):** Build `VirtualToken` inheriting OpenZeppelin's `ERC20Upgradeable`, `AccessControlUpgradeable`, and `UUPSUpgradeable`. Implement decimal configurability (`uint8 private immutable _decimals`).
4. **Implement Soulbound Transfer Guardrails in `VirtualToken.sol`:** Override OpenZeppelin v5's internal `_update(address from, address to, uint256 value)` function:
   - Check if `from == address(0)` (mint) -> assert caller is registered faucet or has `MINTER_ROLE`.
   - Check if `to == address(0)` (burn) -> assert caller is faucet, token owner, or has `BURNER_ROLE`.
   - For regular transfers (`from != address(0) && to != address(0)`) -> assert `isApprovedExchange[from] || isApprovedExchange[to]`. If neither party is an approved demo exchange, revert with `TokenTransferRestricted(from, to)`.
5. **Implement Exchange Whitelisting & Admin Hooks in `VirtualToken.sol`:** Add functions `setApprovedExchange(address exchange, bool status)` and `resetBalance(address account)` restricted to `DEMO_ADMIN_ROLE`.
6. **Implement Strict Chain-ID Assertions:** Add `onlyTestnet` modifier and constructor assertion in both `VirtualToken.sol` and `VirtualFaucet.sol`:
   ```solidity
   uint256 public constant TESTNET_CHAIN_ID = 13371;
   modifier onlyTestnet() {
       if (block.chainid != TESTNET_CHAIN_ID) revert InvalidNetworkEnvironment(block.chainid, TESTNET_CHAIN_ID);
       _;
   }
   ```
7. **Configure Storage Layout & ERC-7201 Namespaces:** Implement upgrade-safe storage layout using ERC-7201 standard for `VirtualFaucet` to guarantee storage layout stability across upgrades.
8. **Implement Role-Based Access Control in `VirtualFaucet.sol`:** Configure OpenZeppelin `AccessControlUpgradeable` with roles:
   - `DEFAULT_ADMIN_ROLE`: Governance multi-sig for contract upgrades and parameter modifications.
   - `FAUCET_OPERATOR_ROLE`: Operational backend service allowed to adjust quotas, configure supported tokens, and process batch portfolio resets.
   - `PAUSER_ROLE`: Emergency monitoring role capable of toggling faucet pause state.
9. **Implement Tiered Quota & Cooldown Mechanics in `VirtualFaucet.sol`:** Implement quota configuration per tier (`RETAIL`, `ALGO_TRADER`, `INSTITUTIONAL`):
   - Track `lastClaimTimestamp[user][token]` and `userTier[user]`.
   - Enforce cooldown window: `block.timestamp >= lastClaimTimestamp[user][token] + cooldownPeriod`.
   - Enforce 24-hour global dispensing cap per asset to prevent testnet node memory pool congestion.
10. **Implement Atomic Multi-Asset Claiming:** Build `claimDemoBundle(uint8 tier)` allowing a trader to mint a standard testing package (e.g., 100,000 vINR + 1,000 vUSDT + 0.1 vBTC) in a single atomic transaction.
11. **Implement Gasless EIP-712 Permit Minting:** Implement `claimWithPermit(...)`:
    - Define EIP-712 domain separator: `EIP712("NBSE_VirtualFaucet", "1.0")`.
    - Validate signature over `ClaimPermit(address user, uint8 tier, uint256 nonce, uint256 deadline)`.
    - Verify `deadline >= block.timestamp` and enforce sequential nonces (`userNonces[user]++`).
    - Mint tokens directly to the signing user's address, regardless of who submitted the transaction (e.g., relayer/paymaster).
12. **Implement Portfolio Reset & Rebalance Mechanics:** Implement `resetDemoPortfolio(address user)`:
    - Burn all existing virtual token balances for the specified user across all registered testnet assets.
    - Re-mint clean default baseline balances corresponding to the user's tier.
    - Emit `PortfolioReset(address indexed user, uint256 timestamp)`.
13. **Build Comprehensive Foundry Mock & Test Harness:** Create `test/demo/VirtualFaucet.t.sol` and `test/demo/VirtualTokens.t.sol`:
    - Unit tests for single and bundle claims, cooldown enforcement, and tier transitions.
    - Soulbound verification: assert that direct `transfer(alice, bob)` always reverts, while `transfer(alice, demoExchange)` succeeds.
    - Mainnet rejection test: mock `block.chainid = 2026` and assert that deployment and function execution revert.
14. **Implement Invariant Fuzzing Suite:** Write property-based tests verifying:
    - Global daily dispensing cap cannot be breached under any sequence of claims.
    - Tokens can never exist outside of user wallets and approved exchange contracts.
    - Rate-limit cooldown cannot be bypassed via zero-value calls or rapid replay.
15. **Compile, Profile Gas, and Run Static Analysis:** Run `forge test`, verify zero compiler warnings, profile gas costs (ensure bundle claim consumes $< 120,000$ gas), and execute Slither to verify zero high or medium findings.

## Interfaces / Contracts

### 1. Virtual Token Master Interface (`IVirtualToken.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title IVirtualToken
 * @author Growww National Blockchain Stock Exchange (NBSE)
 * @notice Interface for non-transferable, soulbound testnet demo tokens used in paper trading.
 */
interface IVirtualToken is IERC20 {
    // --- Custom Errors ---
    error TokenTransferRestricted(address from, address to);
    error UnauthorizedMinter(address caller);
    error UnauthorizedBurner(address caller);
    error UnauthorizedAdmin(address caller);
    error InvalidNetworkEnvironment(uint256 currentChainId, uint256 expectedChainId);
    error ZeroAddressDetected();

    // --- Events ---
    event ExchangeAuthorizationUpdated(address indexed exchange, bool isAuthorized);
    event FaucetMinterUpdated(address indexed previousFaucet, address indexed newFaucet);
    event UserBalanceReset(address indexed user, uint256 burnedAmount, uint256 timestamp);

    // --- External Functions ---

    /// @notice Returns the decimals of the virtual token
    function decimals() external view returns (uint8);

    /// @notice Checks if an address is an authorized demo exchange or settlement contract
    function isApprovedExchange(address exchange) external view returns (bool);

    /// @notice Sets whether an address is an authorized demo exchange contract
    function setApprovedExchange(address exchange, bool isAuthorized) external;

    /// @notice Mints virtual tokens for paper trading (restricted to faucet or minter role)
    function mint(address to, uint256 amount) external;

    /// @notice Burns virtual tokens from a specific account (restricted to faucet, owner, or burner role)
    function burn(address from, uint256 amount) external;

    /// @notice Burns the entire virtual token balance of an account during portfolio reset
    function resetBalance(address account) external returns (uint256 burnedAmount);
}
```

### 2. Virtual Faucet Master Interface (`IVirtualFaucet.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IVirtualFaucet
 * @author Growww National Blockchain Stock Exchange (NBSE)
 * @notice Interface for the automated testnet faucet dispensing demo tokens for paper trading.
 */
interface IVirtualFaucet {
    // --- Enums ---

    /// @notice User classification tier determining faucet dispensing quotas
    enum QuotaTier {
        RETAIL,          // Standard retail paper trader
        ALGO_TRADER,     // High-frequency or algorithmic strategy tester
        INSTITUTIONAL    // Institutional market maker sandbox account
    }

    // --- Structs ---

    /// @notice Dispensing quota and cooldown configuration for a specific token and tier
    struct TokenQuotaConfig {
        uint256 claimAmount;         // Tokens dispensed per valid claim request
        uint32 cooldownDuration;     // Cooldown duration in seconds (e.g., 86,400 = 24 hours)
        uint256 dailyGlobalCap;      // Maximum total tokens dispensed globally per 24h epoch
        uint256 dailyDispensed;      // Total tokens dispensed in the current epoch
        uint64 currentEpochDay;      // Epoch day number (block.timestamp / 86400)
        bool isActive;               // Whether faucet dispensing is enabled for this token
    }

    /// @notice Standard multi-asset bundle allocation per tier
    struct BundleAllocation {
        address tokenAddress;
        uint256 amount;
    }

    /// @notice EIP-712 gasless claim permit structure
    struct ClaimPermit {
        address user;
        QuotaTier tier;
        uint256 nonce;
        uint256 deadline;
    }

    // --- Custom Errors ---
    error CooldownActive(address user, address token, uint256 availableAt);
    error GlobalDailyCapExceeded(address token, uint256 requested, uint256 remaining);
    error TokenNotSupported(address token);
    error FaucetPaused();
    error InvalidSignature();
    error SignatureExpired(uint256 deadline, uint256 currentTimestamp);
    error InvalidNonce(uint256 expected, uint256 provided);
    error InvalidNetworkEnvironment(uint256 currentChainId, uint256 expectedChainId);
    error ZeroAddressDetected();
    error ArrayLengthMismatch();

    // --- Events ---
    event TokensDispensed(
        address indexed user,
        address indexed token,
        uint256 amount,
        QuotaTier tier,
        uint256 timestamp
    );

    event BundleDispensed(
        address indexed user,
        QuotaTier tier,
        uint256 tokenCount,
        uint256 timestamp
    );

    event PortfolioReset(
        address indexed user,
        QuotaTier tier,
        uint256 timestamp
    );

    event QuotaConfigUpdated(
        address indexed token,
        QuotaTier tier,
        uint256 claimAmount,
        uint32 cooldownDuration,
        uint256 dailyGlobalCap
    );

    event UserTierUpdated(address indexed user, QuotaTier tier);

    // --- External Functions ---

    /// @notice Directly claims demo tokens for a single supported token
    function claimToken(address token) external;

    /// @notice Directly claims the standard multi-asset paper trading bundle
    function claimBundle() external;

    /// @notice Gasless claim using an EIP-712 permit submitted by a relayer or paymaster
    function claimWithPermit(
        ClaimPermit calldata permit,
        bytes calldata signature
    ) external;

    /// @notice Resets the user's paper trading portfolio: burns current balances and re-mints bundle
    function resetPortfolio(address user) external;

    /// @notice Returns the timestamp when the user can next claim a specific token
    function getNextClaimTime(address user, address token) external view returns (uint256);

    /// @notice Returns remaining daily global quota for a given token
    function getRemainingDailyQuota(address token) external view returns (uint256);

    /// @notice Checks if an address is eligible to claim a token immediately
    function isEligibleToClaim(address user, address token) external view returns (bool, string memory reason);
}
```

### 3. Virtual Token Implementation Snippet (`VirtualToken.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {ERC20Upgradeable} from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {IVirtualToken} from "./interfaces/IVirtualToken.sol";

/**
 * @title VirtualToken
 * @notice Production implementation of soulbound virtual tokens for paper trading on Besu Testnet.
 */
contract VirtualToken is IVirtualToken, ERC20Upgradeable, AccessControlUpgradeable, UUPSUpgradeable {
    bytes32 public constant DEMO_ADMIN_ROLE = keccak256("DEMO_ADMIN_ROLE");
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

    uint256 public constant TESTNET_CHAIN_ID = 13371;

    uint8 private _customDecimals;
    address public faucetAddress;
    mapping(address => bool) private _approvedExchanges;

    modifier onlyTestnet() {
        if (block.chainid != TESTNET_CHAIN_ID) {
            revert InvalidNetworkEnvironment(block.chainid, TESTNET_CHAIN_ID);
        }
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        string memory name_,
        string memory symbol_,
        uint8 decimals_,
        address admin_,
        address faucet_
    ) external initializer onlyTestnet {
        if (admin_ == address(0) || faucet_ == address(0)) revert ZeroAddressDetected();

        __ERC20_init(name_, symbol_);
        __AccessControl_init();
        __UUPSUpgradeable_init();

        _customDecimals = decimals_;
        faucetAddress = faucet_;

        _grantRole(DEFAULT_ADMIN_ROLE, admin_);
        _grantRole(DEMO_ADMIN_ROLE, admin_);
        _grantRole(MINTER_ROLE, faucet_);
        _grantRole(BURNER_ROLE, faucet_);
    }

    function decimals() public view override(ERC20Upgradeable, IVirtualToken) returns (uint8) {
        return _customDecimals;
    }

    function isApprovedExchange(address exchange) external view override returns (bool) {
        return _approvedExchanges[exchange];
    }

    function setApprovedExchange(address exchange, bool isAuthorized) external override onlyRole(DEMO_ADMIN_ROLE) onlyTestnet {
        if (exchange == address(0)) revert ZeroAddressDetected();
        _approvedExchanges[exchange] = isAuthorized;
        emit ExchangeAuthorizationUpdated(exchange, isAuthorized);
    }

    function setFaucetAddress(address newFaucet) external onlyRole(DEMO_ADMIN_ROLE) onlyTestnet {
        if (newFaucet == address(0)) revert ZeroAddressDetected();
        address oldFaucet = faucetAddress;
        _revokeRole(MINTER_ROLE, oldFaucet);
        _revokeRole(BURNER_ROLE, oldFaucet);

        faucetAddress = newFaucet;
        _grantRole(MINTER_ROLE, newFaucet);
        _grantRole(BURNER_ROLE, newFaucet);

        emit FaucetMinterUpdated(oldFaucet, newFaucet);
    }

    function mint(address to, uint256 amount) external override onlyRole(MINTER_ROLE) onlyTestnet {
        if (to == address(0)) revert ZeroAddressDetected();
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) external override onlyTestnet {
        if (!hasRole(BURNER_ROLE, msg.sender) && msg.sender != from) {
            revert UnauthorizedBurner(msg.sender);
        }
        _burn(from, amount);
    }

    function resetBalance(address account) external override onlyRole(BURNER_ROLE) onlyTestnet returns (uint256) {
        if (account == address(0)) revert ZeroAddressDetected();
        uint256 balance = balanceOf(account);
        if (balance > 0) {
            _burn(account, balance);
            emit UserBalanceReset(account, balance, block.timestamp);
        }
        return balance;
    }

    /**
     * @dev Overrides ERC-20 token transfer logic to enforce soulbound restriction.
     * Tokens can ONLY move if minting, burning, or interacting with an approved Demo Exchange.
     */
    function _update(address from, address to, uint256 value) internal override onlyTestnet {
        // Allow minting (from == address(0)) and burning (to == address(0))
        if (from != address(0) && to != address(0)) {
            // Strict soulbound guardrail: either sender or recipient MUST be an approved exchange contract
            if (!_approvedExchanges[from] && !_approvedExchanges[to]) {
                revert TokenTransferRestricted(from, to);
            }
        }
        super._update(from, to, value);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(DEFAULT_ADMIN_ROLE) onlyTestnet {}
}
```

### 4. Virtual Faucet Implementation Snippet (`VirtualFaucet.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import {EIP712Upgradeable} from "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {IVirtualFaucet} from "./interfaces/IVirtualFaucet.sol";
import {IVirtualToken} from "./interfaces/IVirtualToken.sol";

/**
 * @title VirtualFaucet
 * @notice Production automated faucet dispensing soulbound paper trading tokens on Hyperledger Besu Testnet.
 */
contract VirtualFaucet is
    IVirtualFaucet,
    AccessControlUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable,
    EIP712Upgradeable,
    UUPSUpgradeable
{
    using ECDSA for bytes32;

    bytes32 public constant FAUCET_OPERATOR_ROLE = keccak256("FAUCET_OPERATOR_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");

    uint256 public constant TESTNET_CHAIN_ID = 13371;
    bytes32 public constant CLAIM_PERMIT_TYPEHASH =
        keccak256("ClaimPermit(address user,uint8 tier,uint256 nonce,uint256 deadline)");

    // Token => QuotaTier => TokenQuotaConfig
    mapping(address => mapping(QuotaTier => TokenQuotaConfig)) public tokenQuotas;
    // User => Token => LastClaimTimestamp
    mapping(address => mapping(address => uint256)) public lastClaimTimestamp;
    // User => Assigned QuotaTier
    mapping(address => QuotaTier) public userTiers;
    // User => Nonce for EIP-712 gasless claims
    mapping(address => uint256) public userNonces;

    // Supported active tokens list
    address[] public supportedTokens;
    mapping(address => bool) public isTokenSupported;

    modifier onlyTestnet() {
        if (block.chainid != TESTNET_CHAIN_ID) {
            revert InvalidNetworkEnvironment(block.chainid, TESTNET_CHAIN_ID);
        }
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address admin, address operator) external initializer onlyTestnet {
        if (admin == address(0) || operator == address(0)) revert ZeroAddressDetected();

        __AccessControl_init();
        __Pausable_init();
        __ReentrancyGuard_init();
        __EIP712_init("NBSE_VirtualFaucet", "1.0");
        __UUPSUpgradeable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(FAUCET_OPERATOR_ROLE, operator);
        _grantRole(PAUSER_ROLE, admin);
        _grantRole(PAUSER_ROLE, operator);
    }

    function configureToken(
        address token,
        QuotaTier tier,
        uint256 claimAmount,
        uint32 cooldownDuration,
        uint256 dailyGlobalCap,
        bool isActive
    ) external onlyRole(FAUCET_OPERATOR_ROLE) onlyTestnet {
        if (token == address(0)) revert ZeroAddressDetected();

        if (!isTokenSupported[token]) {
            supportedTokens.push(token);
            isTokenSupported[token] = true;
        }

        tokenQuotas[token][tier] = TokenQuotaConfig({
            claimAmount: claimAmount,
            cooldownDuration: cooldownDuration,
            dailyGlobalCap: dailyGlobalCap,
            dailyDispensed: 0,
            currentEpochDay: uint64(block.timestamp / 86400),
            isActive: isActive
        });

        emit QuotaConfigUpdated(token, tier, claimAmount, cooldownDuration, dailyGlobalCap);
    }

    function setUserTier(address user, QuotaTier tier) external onlyRole(FAUCET_OPERATOR_ROLE) onlyTestnet {
        if (user == address(0)) revert ZeroAddressDetected();
        userTiers[user] = tier;
        emit UserTierUpdated(user, tier);
    }

    function claimToken(address token) external override nonReentrant whenNotPaused onlyTestnet {
        _dispenseToken(msg.sender, token, userTiers[msg.sender]);
    }

    function claimBundle() external override nonReentrant whenNotPaused onlyTestnet {
        _dispenseBundle(msg.sender, userTiers[msg.sender]);
    }

    function claimWithPermit(
        ClaimPermit calldata permit,
        bytes calldata signature
    ) external override nonReentrant whenNotPaused onlyTestnet {
        if (permit.deadline < block.timestamp) {
            revert SignatureExpired(permit.deadline, block.timestamp);
        }
        if (permit.nonce != userNonces[permit.user]) {
            revert InvalidNonce(userNonces[permit.user], permit.nonce);
        }

        bytes32 structHash = keccak256(
            abi.encode(CLAIM_PERMIT_TYPEHASH, permit.user, uint8(permit.tier), permit.nonce, permit.deadline)
        );
        bytes32 digest = _hashTypedDataV4(structHash);
        address recoveredSigner = digest.recover(signature);

        if (recoveredSigner != permit.user) revert InvalidSignature();

        userNonces[permit.user]++;
        _dispenseBundle(permit.user, permit.tier);
    }

    function resetPortfolio(address user) external override nonReentrant onlyTestnet {
        if (msg.sender != user && !hasRole(FAUCET_OPERATOR_ROLE, msg.sender)) {
            revert UnauthorizedAdmin(msg.sender);
        }

        // 1. Burn existing virtual balances across all supported tokens
        for (uint256 i = 0; i < supportedTokens.length; i++) {
            address token = supportedTokens[i];
            IVirtualToken(token).resetBalance(user);
        }

        // 2. Re-mint baseline bundle balances for the user's tier
        QuotaTier tier = userTiers[user];
        _dispenseBundle(user, tier);

        emit PortfolioReset(user, tier, block.timestamp);
    }

    function _dispenseToken(address user, address token, QuotaTier tier) internal {
        if (!isTokenSupported[token]) revert TokenNotSupported(token);

        TokenQuotaConfig storage config = tokenQuotas[token][tier];
        if (!config.isActive) revert TokenNotSupported(token);

        // Check Cooldown
        uint256 nextAvailable = lastClaimTimestamp[user][token] + config.cooldownDuration;
        if (block.timestamp < nextAvailable) {
            revert CooldownActive(user, token, nextAvailable);
        }

        // Epoch daily cap rollover check
        uint64 currentDay = uint64(block.timestamp / 86400);
        if (currentDay > config.currentEpochDay) {
            config.currentEpochDay = currentDay;
            config.dailyDispensed = 0;
        }

        // Check global daily cap
        if (config.dailyDispensed + config.claimAmount > config.dailyGlobalCap) {
            revert GlobalDailyCapExceeded(
                token,
                config.claimAmount,
                config.dailyGlobalCap - config.dailyDispensed
            );
        }

        // Update accounting
        config.dailyDispensed += config.claimAmount;
        lastClaimTimestamp[user][token] = block.timestamp;

        // Execute privileged mint on virtual token
        IVirtualToken(token).mint(user, config.claimAmount);

        emit TokensDispensed(user, token, config.claimAmount, tier, block.timestamp);
    }

    function _dispenseBundle(address user, QuotaTier tier) internal {
        uint256 count = supportedTokens.length;
        for (uint256 i = 0; i < count; i++) {
            address token = supportedTokens[i];
            if (tokenQuotas[token][tier].isActive) {
                _dispenseToken(user, token, tier);
            }
        }
        emit BundleDispensed(user, tier, count, block.timestamp);
    }

    function getNextClaimTime(address user, address token) external view override returns (uint256) {
        QuotaTier tier = userTiers[user];
        TokenQuotaConfig storage config = tokenQuotas[token][tier];
        return lastClaimTimestamp[user][token] + config.cooldownDuration;
    }

    function getRemainingDailyQuota(address token) external view override returns (uint256) {
        QuotaTier tier = QuotaTier.RETAIL; // Default reference tier
        TokenQuotaConfig storage config = tokenQuotas[token][tier];
        uint64 currentDay = uint64(block.timestamp / 86400);
        if (currentDay > config.currentEpochDay) {
            return config.dailyGlobalCap;
        }
        if (config.dailyDispensed >= config.dailyGlobalCap) {
            return 0;
        }
        return config.dailyGlobalCap - config.dailyDispensed;
    }

    function isEligibleToClaim(address user, address token) external view override returns (bool, string memory) {
        if (!isTokenSupported[token]) return (false, "Token not supported");
        QuotaTier tier = userTiers[user];
        TokenQuotaConfig storage config = tokenQuotas[token][tier];
        if (!config.isActive) return (false, "Token tier inactive");
        if (block.timestamp < lastClaimTimestamp[user][token] + config.cooldownDuration) {
            return (false, "Cooldown active");
        }
        uint64 currentDay = uint64(block.timestamp / 86400);
        uint256 dispensed = (currentDay > config.currentEpochDay) ? 0 : config.dailyDispensed;
        if (dispensed + config.claimAmount > config.dailyGlobalCap) {
            return (false, "Daily global cap reached");
        }
        return (true, "Eligible");
    }

    function pause() external onlyRole(PAUSER_ROLE) onlyTestnet {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) onlyTestnet {
        _unpause();
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(DEFAULT_ADMIN_ROLE) onlyTestnet {}
}
```

## Security & Compliance Notes
- **Strict Chain-ID Guardrail (`block.chainid == 13371`):**
  - The single most critical security invariant of this contract suite is that it **cannot exist, initialize, or execute on the Growww Mainnet (Chain ID 2026)**.
  - To enforce this guarantee, every state-mutating function and initialization lifecycle hook includes the `onlyTestnet` modifier:
    $$\text{require}(\text{block.chainid} == 13371, \text{"VirtualFaucet: Testnet only"})$$
  - If an automated deployment pipeline or rogue admin mistakenly attempts to deploy these contracts to Mainnet, constructor execution and proxy initialization immediately revert.
- **Soulbound Non-Transferability & Grey-Market Prevention:**
  - Standard ERC-20 tokens can be transferred freely between arbitrary accounts (`alice -> bob`). In a simulated exchange ecosystem, free transferability creates severe systemic hazards:
    1. Fraudulent OTC secondary sales where scammers sell "cheap USDT" to retail users who do not understand it is testnet `vUSDT`.
    2. Off-chain gaming or bribery markets settled in paper-trading tokens.
    3. Distortion of paper-trading leaderboards and algorithmic competitions via collusive token dumping.
  - `VirtualToken.sol` solves this by overriding `_update`. P2P transfers between user wallets revert with `TokenTransferRestricted`. 
  - Token transfers are permitted strictly to or from approved contracts registered in `_approvedExchanges`. These approved contracts are solely the `DemoMatchingEngine`, `DemoSettlementDvP`, and `VirtualEscrow` contracts.
- **Rate-Limiting & Anti-Sybil Defense:**
  - Individual users are restricted by a configurable cooldown window (`cooldownDuration = 86,400 seconds` / 24 hours per token).
  - To prevent automated botnets from rotating EVM addresses and draining testnet node memory pools with spam transactions, a 24-hour global dispensing cap (`dailyGlobalCap`) is enforced per asset across all users.
  - The API Gateway (Prompt 219) and Demo Wallet Service (Prompt 274) enforce layer-7 IP and user session rate limits before generating EIP-712 permits.
- **Gasless Claim Security & Replay Protection:**
  - Gasless minting via `claimWithPermit` utilizes EIP-712 structured hashing.
  - Each permit includes a monotonic sequential nonce (`userNonces[user]++`) and an absolute deadline timestamp (`deadline >= block.timestamp`), preventing transaction replay across time, relayers, or fork environments.
  - Tokens are minted strictly to `permit.user`, preventing third-party relayers or paymasters from diverting virtual tokens to their own accounts.
- **Reentrancy & Pull-Over-Push Accounting:**
  - All external claiming and portfolio reset operations enforce OpenZeppelin `ReentrancyGuardUpgradeable`.
  - State variables (cooldown timestamps, daily dispensed counters, nonces) are updated **prior** to external minting calls, conforming strictly to the Checks-Effects-Interactions pattern.
- **SEBI & Regulatory Sandbox Alignment:**
  - Under SEBI Circular guidelines on Innovation Sandboxes and Retail Investor Education, mock trading platforms must maintain strict separation between real investor capital and paper trading instruments.
  - No real rupee or foreign currency transactions may be linked to demo tokens.
  - The contract does not store any Personally Identifiable Information (PII), adhering strictly to the Digital Personal Data Protection Act (DPDPA 2023).

## Acceptance Criteria
- [ ] Production Solidity contracts `VirtualFaucet.sol`, `VirtualTokens.sol`, `IVirtualFaucet.sol`, and `IVirtualToken.sol` compiled under Solidity ^0.8.24 with zero compiler warnings and ERC-7201 upgradeable storage layout adherence.
- [ ] Mainnet Chain ID Guardrail verified: Deployment and execution on any network where `block.chainid != 13371` reverts with `InvalidNetworkEnvironment`. Specifically, tests simulating Mainnet (`block.chainid == 2026`) revert across all entrypoints.
- [ ] Soulbound Transfer Restriction verified:
  - Direct token transfers between non-exchange addresses (`transfer(bob, amount)`) revert with `TokenTransferRestricted`.
  - Transfers where `to` or `from` is an approved demo exchange (`setApprovedExchange(exchange, true)`) execute successfully.
- [ ] Tiered Quota & Cooldown Mechanics verified:
  - Single token claims (`claimToken`) and bundle claims (`claimBundle`) dispense exact configured amounts.
  - Subsequent claim attempts within the cooldown window revert with `CooldownActive`.
  - Dispensing attempts that would exceed the 24-hour `dailyGlobalCap` revert with `GlobalDailyCapExceeded`.
- [ ] Gasless Claim Permit (EIP-712) verified:
  - Valid EIP-712 signature over `ClaimPermit` successfully mints bundle to the signer when submitted by an arbitrary relayer.
  - Submitting an expired permit reverts with `SignatureExpired`.
  - Replaying a signature or submitting an invalid nonce reverts with `InvalidNonce`.
- [ ] Portfolio Reset Hook verified:
  - Invoking `resetPortfolio(user)` burns all current balances of `vINR`, `vUSDT`, `vBTC`, and `vETH` held by `user`.
  - Re-mints standard initial baseline balances corresponding to the user's tier.
  - Unauthorized callers attempting to reset another user's portfolio revert with `UnauthorizedAdmin`.
- [ ] Slither static analysis runs with zero high or medium severity findings, verifying absence of unchecked external calls, reentrancy vulnerabilities, or improper access control.
- [ ] Invariant Fuzz Testing: Foundry invariant test suite runs $\ge 10,000$ iterations asserting that no soulbound virtual token can ever be transferred directly between two non-whitelisted retail addresses.
- [ ] Gas Benchmark: Single token claim consumes $< 60,000$ gas; multi-asset bundle claim consumes $< 120,000$ gas on Besu Testnet.
- [ ] Foundry test suite achieves $\ge 95\%$ line and branch coverage across all execution flows, role permissions, custom error reverts, and boundary conditions.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `301` (Permissioned Blockchain Platform Evaluation & Selection - Hyperledger Besu)
  - Prompt `302` (Network Topology & QBFT Validator Infrastructure Setup)
  - Prompt `307` (Multi-Party Authorization & Multisig Governance Smart Contract)
  - Prompt `311` (Validator & Relayer Key Management via HSM)
- **Parallel Tasks:**
  - Prompt `273` (Demo Matching Engine & Simulated Order Book)
  - Prompt `274` (Demo Wallet Service & Paper Trading Balance Ledger)
  - Prompt `259` (ERC-4337 Account Abstraction Bundler & Gasless Paymaster Service)
  - Prompt `337` (ERC-4337 Modular Smart Account & Ephemeral Session Keys Contract)
  - Prompt `527` (Flutter Environment Switcher & Sandbox Mode Toggle)
  - Prompt `309` (Blockchain Event Indexing & Query Service)
- **Subsequent Prompts Enabled:**
  - Prompt `518` (Flutter Mobile Paper Trading Terminal & Analytics)
  - Prompt `618` (Web Algorithmic Trading Backtesting & Sandbox Console)
  - Prompt `906` (Regulatory Sandbox Scenarios & Automated UAT Test Suites)
  - Prompt `915` (Deterministic Market Replay & Flash Crash Simulator)
