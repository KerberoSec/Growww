# 320 - Permissioned Testnet Cluster, Developer Faucet & Mock Depository Minting Gateway

## Purpose
Before executing institutional tokenized equity settlement and cross-border transactions on production mainnet, developers, institutional partner nodes, liquidity providers, and regulatory sandbox observers require a high-fidelity, deterministic, and isolated staging environment. This environment must faithfully mirror production consensus (Hyperledger Besu QBFT), permissioning contracts, settlement Delivery-versus-Payment (DvP) workflows, and cross-chain bridges (Sepolia L1 tethering) without exposing real capital or live demat securities.

This prompt specifies the architecture, infrastructure manifests, cryptographic configuration, developer faucet, and mock depository gateway for the **Permissioned Testnet Cluster & Developer Faucet Infrastructure** (`infra/blockchain/testnet/` and `services/testnet-faucet/`). It establishes an enterprise-grade Besu QBFT testnet (Chain ID `13371`), an automated anti-abuse developer faucet for mock native gas and test stablecoins (mINR/mUSD), a high-performance **Mock NSDL/CDSL Depository Gateway** that simulates Indian electronic depository interfaces (SPEED-e / CDSL Easiest) for automated minting and demat synchronization, and full integration with public Sepolia L1 testnets for cross-chain interoperability validation.

## What You Are Building
A production-grade testnet environment and developer tooling ecosystem comprising:
- **Hyperledger Besu QBFT Testnet Cluster (Chain ID: 13371):** Multi-node permissioned network with 4 validator nodes, 2 load-balanced public RPC relayers, 1 archival node, and 1 dedicated regulatory observer node.
- **Developer Faucet Service (`services/testnet-faucet/`):** High-throughput, rate-limited backend (Go v1.22+) and frontend interface providing authenticated developers with testnet gas (tETH/tGAS), mock fiat settlement tokens (`mINR`, `mUSD`), and sandbox equity tokens with Cloudflare Turnstile anti-abuse protection.
- **Mock NSDL/CDSL Share Minting Gateway (`services/mock-depository/`):** A high-fidelity microservice simulating National Securities Depository Limited (NSDL) and Central Depository Services Limited (CDSL) APIs, handling demat credit/debit instructions, beneficial owner (BENPOS) snapshots, pledge allocations, and automated token minting triggers.
- **Sepolia L1 Testnet Interoperability Tether:** Deployed Chainlink CCIP and LayerZero v2 testnet endpoint configurations bridging the Growww Besu Testnet with Ethereum Sepolia and Arbitrum Sepolia.
- **Blockscout EVM Testnet Explorer & Telemetry:** Dedicated open-source block explorer and Grafana dashboard exposing real-time block progression, validator health, gas usage, and contract verification.

## Scope Boundaries
- **In Scope:**
  - `genesis.json` configuration for Testnet Cluster (Chain ID `13371`) with 2-second QBFT block periods and zero base fee.
  - Terraform, Helm, and Kubernetes StatefulSet manifests for 4 testnet validators, bootnodes, RPC nodes, and explorer.
  - Developer Faucet REST API, Redis sliding-window rate limiters, and `TestnetTokenFaucet.sol` smart contract.
  - Mock NSDL/CDSL REST/JSON API gateway simulating ISO 20022 clearing messages and SPEED-e demat operations.
  - Synthetic test data generation scripts producing realistic Indian equity ISINs (e.g., `INE002A01018` for mock Reliance, `INE467B01029` for mock TCS) with 1:1 simulated vault backing.
  - Sepolia cross-chain bridge testnet contract deployments and mock verifiers.
- **Out of Scope / Handled Elsewhere:**
  - Production mainnet validator key ceremony and genesis deployment (handled in Prompt 321).
  - Production HSM and CloudHSM key management integration (handled in Prompt 311).
  - Production live NSDL/CDSL depository VPN connections (handled in Prompt 213).
  - Production cross-chain bridge contracts (handled in Prompt 319).

## Technology to Use
- **Blockchain Node Engine:** **Hyperledger Besu v24.x+** running QBFT consensus in testnet mode.
- **Smart Contract Language:** **Solidity 0.8.24** with EVM target Prague / Cancun.
- **Faucet Backend & Mock Depository Language:** **Go (v1.22+)** with `go-ethereum/ethclient`, `gin-gonic/gin`, and `redis-go`.
- **Database & Cache:** **Redis 7.2** (for faucet rate-limiting and nonce management) and **PostgreSQL 16** (for mock depository ledger state).
- **Explorer:** **Blockscout v6.x** containerized with PostgreSQL backend.
- **Orchestration & IaC:** Kubernetes (EKS / GKE), Helm v3, Terraform, and Docker Compose for local development clusters.
- **Anti-Abuse Protection:** Cloudflare Turnstile, GitHub OAuth verification, and IP/Wallet sliding window throttles.

## Backend / Infra Touchpoints
- **Custodian Depository Service (Prompt 213):** Switches seamlessly between the `Mock Depository Gateway` in testnet/staging environments and live NSDL/CDSL interfaces in production.
- **Trade Settlement Service (Prompt 208):** Executes DvP settlement tests against the Besu testnet nodes via JSON-RPC endpoint `https://rpc.testnet.growww.internal`.
- **Event Indexing Service (Prompt 309):** Indexes testnet blocks, token mints, and DvP settlements into the staging PostgreSQL database.
- **CI/CD Integration Pipeline (Prompt 804):** Automated end-to-end integration test runners spin up ephemeral dockerized testnet instances or run tests directly against the staging testnet cluster.

## Blockchain Interaction
- **Network ID & Chain ID:** Chain ID `13371`, Network ID `13371`.
- **Consensus Timing:** QBFT block period configured to 2.0 seconds with round-robin leader election and 4-second request timeouts.
- **Gas Economics:** Zero base fee (`zeroBaseFee: true`) with optional fixed minimum gas price ($1\text{ Gwei}$) to prevent memory pool transaction flooding while keeping testing frictionless.
- **Mock Token Issuance & KYC Registry:** Deploys testnet instances of `IdentityRegistry.sol` (ERC-3643) pre-seeded with test institutional and retail identities. The Mock Depository Gateway invokes `TokenIssuance.sol` to mint testnet equity tokens upon receiving simulated demat credit requests.
- **Zero PII Transmission:** Faucet claims and mock depository records use synthetic account identifiers (e.g., `MOCK-BOID-001928471029`) and public Ethereum addresses. No real user PAN, Aadhaar, bank details, or contact data are stored or transmitted.

## Step-by-Step Build Instructions
1. Scaffold directory structure under `infra/blockchain/testnet/` and `services/testnet-faucet/`:
   - `infra/blockchain/testnet/genesis/` (Besu `genesis.json` and static nodes configuration).
   - `infra/blockchain/testnet/helm/` (Helm charts for validators, RPCs, and Blockscout).
   - `infra/blockchain/testnet/docker-compose/` (Local multi-node cluster for local dev).
   - `services/testnet-faucet/` (Go faucet service with REST API and web UI).
   - `services/mock-depository/` (Mock NSDL/CDSL gateway).
2. Generate testnet validator keys for 4 nodes (`val-1`, `val-2`, `val-3`, `val-4`) using `besu operator generate-blockchain-config` with deterministic test mnemonics.
3. Configure `genesis.json` for Chain ID `13371`:
   - Define QBFT consensus parameters: `blockperiodseconds: 2`, `epochlength: 30000`, `requesttimeoutseconds: 4`.
   - Allocate initial testnet gas tokens to the `TestnetTokenFaucet` smart contract and deployment addresses.
   - Pre-allocate permissioning contracts (`NodeRules.sol`, `AccountRules.sol`) at deterministic genesis addresses.
4. Deploy the 4-node Besu QBFT Testnet Cluster on Kubernetes across multi-AZ staging infrastructure.
5. Deploy 2 public/internal RPC Relayer nodes with WebSocket and JSON-RPC enabled (`eth`, `net`, `web3`, `txpool` APIs) behind an Envoy ingress controller.
6. Deploy Blockscout v6 instance connected to the testnet RPC endpoint with automatic smart contract verification service.
7. Implement `TestnetTokenFaucet.sol` smart contract:
   - Maintains pools of native test gas, mock INR settlement tokens (`mINR`), and mock USD tokens (`mUSD`).
   - Supports owner-funded faucet reserves and rate-limited claim disbursement.
   - Integrates authorized backend relayer signature verification (EIP-712) for claimed disbursements.
8. Build the Developer Faucet Backend Service (`services/testnet-faucet/`):
   - Implement claim endpoints: `POST /api/v1/faucet/claim-gas` and `POST /api/v1/faucet/claim-tokens`.
   - Configure Redis sliding-window rate limiting: max 1 claim per wallet address per 24 hours, max 5 claims per IP per 24 hours, optional GitHub OAuth tier (higher claim quota for verified contributors).
   - Integrate Cloudflare Turnstile token validation on all inbound requests.
9. Build the Mock NSDL/CDSL Depository Gateway (`services/mock-depository/`):
   - Implement demat share credit simulation endpoint: `POST /api/v1/depository/mock/demat-credit`.
   - Implement settlement instruction simulation endpoint: `POST /api/v1/depository/mock/settlement-instruction`.
   - Implement corporate action trigger endpoint: `POST /api/v1/depository/mock/corporate-action` (stock split, bonus, dividend payout).
   - Maintain internal PostgreSQL state tracking simulated BOID demat balances, pledging locks, and automated Web3 bridge triggers.
10. Connect the Mock Depository Gateway to the testnet `TokenIssuance.sol` and `TokenRedemption.sol` smart contracts:
    - On simulated demat credit: gateway automatically signs and submits an on-chain `mint()` transaction to credit the corresponding testnet ERC-3643 equity token.
    - On testnet `burn()` event: gateway catches on-chain event via WebSocket and updates mock depository demat records.
11. Configure Sepolia L1 and Arbitrum Sepolia cross-chain bridge endpoints:
    - Deploy testnet instances of `InstitutionalBridgeHub.sol` configured with Chainlink CCIP Sepolia Router (`0x0BF3dE8c5D3e8A2B34D2BEeB17ABfCeBaf363A59`) and LayerZero v2 Sepolia Endpoint.
12. Build automated synthetic data seeding script (`scripts/seed_testnet_data.sh`):
    - Seeds mock Indian equities: `mRELIANCE` (ISIN: `IN9002A01018`), `mTCS` (ISIN: `IN9467B01029`), `mHDFCBANK` (ISIN: `IN9040A01034`), `mINFY` (ISIN: `IN9009A01012`).
    - Seeds test accounts with realistic mock compliance claims and KYC credentials.
13. Execute automated end-to-end integration tests:
    - Test faucet claim throttle rejection and successful disbursement.
    - Test mock demat credit -> on-chain token mint -> atomic DvP trade -> token burn -> mock demat debit.
14. Configure Prometheus alerts and Grafana dashboards for testnet block height progression, peer count, transaction mempool backlog, and faucet reserve balance monitoring.

## Interfaces / Contracts

### Besu Testnet Genesis Configuration (`genesis.json`)
```json
{
  "config": {
    "chainId": 13371,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "zeroBaseFee": true,
    "fixedBaseFee": 0,
    "qbft": {
      "blockperiodseconds": 2,
      "epochlength": 30000,
      "requesttimeoutseconds": 4,
      "validatorcontractaddress": "0x0000000000000000000000000000000000007777",
      "validators": [
        "0xFE3B557E8Fb62b89F4916B721be55cEb828dBd73",
        "0x627306090abaB3A6e1400e9345bC60c78a8BEf57",
        "0xf17f52151EbEF6C7334FAD080c5704D77216b732",
        "0xC5fdf4076b8F3A5357c5E395ab970B5B54098Fef"
      ]
    },
    "permissions": {
      "node": "0x0000000000000000000000000000000000008888",
      "account": "0x0000000000000000000000000000000000009999"
    }
  },
  "nonce": "0x0",
  "timestamp": "0x66E5C000",
  "extraData": "0xf87aa0000000000000000000000000000000000000000000000000000000000000000094fe3b557e8fb62b89f4916b721be55ceb828dbd7394627306090abab3a6e1400e9345bc60c78a8bef5794f17f52151ebef6c7334fad080c5704d77216b73294c5fdf4076b8f3a5357c5e395ab970b5b54098fefc880c0",
  "gasLimit": "0x1C9C380",
  "difficulty": "0x1",
  "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
  "alloc": {
    "0xFE3B557E8Fb62b89F4916B721be55cEb828dBd73": {
      "balance": "0x1000000000000000000000000000"
    },
    "0x1111111111111111111111111111111111111111": {
      "comment": "Testnet Faucet Vault Contract",
      "balance": "0x5000000000000000000000000000"
    }
  }
}
```

### Testnet Token Faucet Smart Contract Interface (`ITestnetTokenFaucet.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface ITestnetTokenFaucet {
    struct ClaimAllowance {
        uint256 nativeGasAmount;
        uint256 mockInrAmount;
        uint256 mockUsdAmount;
        uint256 equityTokenAmount;
    }

    event TokensClaimed(
        address indexed recipient,
        uint256 nativeGasAmount,
        uint256 mockInrAmount,
        uint256 mockUsdAmount,
        address indexed equityToken,
        uint256 equityTokenAmount
    );

    event FaucetConfigUpdated(
        uint256 nativeGasAmount,
        uint256 mockInrAmount,
        uint256 cooldownPeriodSeconds
    );

    event FaucetEmergencyPaused(address indexed triggeredBy, string reason);

    error ClaimCooldownActive(address recipient, uint256 timeRemaining);
    error InvalidClaimSignature(address signer);
    error FaucetReserveInsufficient(address token, uint256 requested, uint256 available);
    error UnauthorizedRelayer(address caller);

    function claimTestTokens(
        address recipient,
        address equityToken,
        uint256 nonce,
        uint256 expiry,
        bytes calldata signature
    ) external;

    function getCooldownRemaining(address recipient) external view returns (uint256);
    function getClaimAllowance() external view returns (ClaimAllowance memory);
}
```

### Mock NSDL/CDSL Depository Gateway API Specification (OpenAPI 3.1 Excerpt)
```yaml
openapi: 3.1.0
info:
  title: Mock NSDL/CDSL Depository Gateway API
  version: 1.0.0
  description: Simulates Indian Demat Depository operations for testnet clearing, settlement, and 1:1 token issuance.
paths:
  /api/v1/depository/mock/demat-credit:
    post:
      summary: Simulate Demat Share Credit & Trigger Token Minting
      operationId: simulateDematCredit
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - mockBoid
                - isin
                - quantity
                - beneficiaryPanHash
                - transactionRef
              properties:
                mockBoid:
                  type: string
                  example: "IN300012-10029384"
                isin:
                  type: string
                  example: "IN9002A01018"
                quantity:
                  type: integer
                  minimum: 1
                  example: 500
                beneficiaryPanHash:
                  type: string
                  example: "0x8f2d9c4e1a3b5c7e9f0a2d4b6c8e0a2f4c6e8a0b2d4f6a8c0e2a4b6c8d0e2f4a"
                targetWalletAddress:
                  type: string
                  example: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
                transactionRef:
                  type: string
                  example: "TX-NSDL-MOCK-20260918-009182"
      responses:
        '201':
          description: Demat credit recorded and on-chain token mint transaction dispatched.
          content:
            application/json:
              schema:
                type: object
                properties:
                  depositoryReference:
                    type: string
                  onChainTxHash:
                    type: string
                  isin:
                    type: string
                  mintedQuantity:
                    type: integer
                  status:
                    type: string
                    enum: [CONFIRMED, PENDING_MINT]

  /api/v1/depository/mock/settlement-instruction:
    post:
      summary: Process Simulated DvP Settlement Depository Lock/Release
      operationId: processSettlementInstruction
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - tradeId
                - sellerBoid
                - buyerBoid
                - isin
                - quantity
                - settlementAmountInr
              properties:
                tradeId:
                  type: string
                sellerBoid:
                  type: string
                buyerBoid:
                  type: string
                isin:
                  type: string
                quantity:
                  type: integer
                settlementAmountInr:
                  type: number
      responses:
        '200':
          description: Depository settlement instruction processed.
```

## Security & Compliance Notes
- **Strict Network & Capital Isolation:** Testnet tokens possess zero monetary value. Under no circumstances may testnet private keys, validator seeds, or RPC ingress credentials overlap with production mainnet infrastructure.
- **Anti-Abuse & Sybil Resistance:** The Developer Faucet utilizes multi-tiered rate limiting (Redis token bucket algorithm per IP and wallet address) paired with Cloudflare Turnstile bot detection and optional GitHub OAuth account verification (requiring an account age $>30$ days) to prevent testnet gas exhaustion.
- **Zero PII & Synthetic Data Compliance:** All depository accounts and participant records generated in testnet workflows strictly utilize synthetic identifiers (e.g. `MOCK-BOID-*`, `IN9*` synthetic ISINs). Real investor Aadhaar numbers, Permanent Account Numbers (PAN), or residential addresses are strictly forbidden from testnet configurations and logs.
- **1:1 Depository Backing Simulation:** The Mock Depository Gateway enforces exact 1:1 mathematical invariance between mock demat share records in PostgreSQL and on-chain circulating testnet tokens, providing an accurate testing ground for production Proof of Reserve (PoR) publishing (Prompt 308).

## Acceptance Criteria
- [ ] Besu QBFT Testnet Cluster (Chain ID `13371`) operational with 4 validators achieving steady 2.0s block finality.
- [ ] Blockscout EVM Explorer successfully deployed, indexing testnet blocks, transactions, and verifying contracts.
- [ ] Developer Faucet backend service dispatches native gas and mock settlement tokens (`mINR`, `mUSD`) within $<2\text{ seconds}$ per valid request.
- [ ] Redis rate limiter correctly enforces cooldown restrictions (max 1 claim/wallet/24h) and rejects automated bot requests.
- [ ] Mock NSDL/CDSL Depository Gateway successfully processes `demat-credit` API calls and invokes on-chain `TokenIssuance.sol` to mint mock equity tokens.
- [ ] Dual-chain test transfer verified between Growww Besu Testnet and Ethereum Sepolia using testnet Chainlink CCIP router.
- [ ] Comprehensive automated test suite passes with 100% success rate across faucet, mock depository, and cluster consensus tests.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Platform Selection), Prompt `302` (Network Topology & Validator Setup), Prompt `303` (Token Issuance Smart Contract).
- **Parallel Tasks:** Prompt `309` (Event Indexing Service), Prompt `310` (Chain Node Monitoring & Alerting).
- **Subsequent Prompts Enabled:** Prompt `321` (Mainnet Genesis Ceremony & Validator Onboarding), Prompt `804` (CI/CD Automated Test Pipeline).
