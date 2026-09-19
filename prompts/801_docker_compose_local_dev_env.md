# 801 - Full-Stack Docker Compose Local Development Environment

## Purpose
The local development environment serves as the foundational sandbox for all Growww engineering teams - spanning backend microservice engineers, Flutter client developers, frontend web engineers, and smart contract auditors. It solves the friction of multi-service orchestration by spinning up the complete, interconnected Growww ecosystem with a single command (`make dev-up`). 

By packaging all databases, streaming message brokers, mock domestic and international financial rails (NSDL/CDSL, UPI payment aggregators, UIDAI KYC gateways), and a local permissioned Hyperledger Besu blockchain node with pre-deployed regulatory smart contracts, developers can test complex end-to-end investment flows (KYC -> INR deposit -> limit order matching -> atomic DvP settlement -> token minting/custody confirmation) entirely offline with sub-second feedback loops.

## What You Are Building
A fully automated, zero-configuration local development stack and tooling harness:
- `deployments/local/docker-compose.yml`: Master multi-container compose file orchestrating the infrastructure, mock servers, and core service runtimes.
- `deployments/local/docker-compose.override.yml.example`: Developer customization template for mounting local source code directories and enabling debug ports.
- `deployments/local/genesis/besu-qbft-dev.json`: Besu Genesis configuration for a single-node QBFT permissioned ledger with 2-second block times and pre-funded test accounts.
- `deployments/local/mocks/`: Lightweight HTTP/gRPC mock servers for NSDL/CDSL Depository APIs, UPI / Bank IMPS/NEFT rails, and UIDAI Aadhaar/PAN verification.
- `deployments/local/scripts/init-contracts.sh`: Automated smart contract compilation and deployment harness using Foundry/Cast that deploys `DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, and `ProofOfReserveRegistry.sol`, outputting deployed contract addresses to a shared JSON volume.
- `deployments/local/scripts/seed-data.sh`: Database and Kafka event seeding script populating domestic equity master records (ISINs for Reliance, TCS, HDFC, Infosys), KYC tiers, and dummy investor accounts.
- `Makefile`: Convenient developer shortcuts (`make dev-up`, `make dev-down`, `make dev-clean`, `make dev-logs`, `make dev-seed`, `make dev-contracts-deploy`).

## Scope Boundaries
- **In Scope:**
 - Containerization and orchestration of local datastores: PostgreSQL 16, Redis 7, Apache Kafka with KRaft mode.
 - Hyperledger Besu dev node running QBFT consensus with JSON-RPC (8545) and WebSocket (8546) ports exposed.
 - Mock third-party financial API servers simulating domestic depositories, banking payment gateways, and KYC providers.
 - Automated smart contract deployment and contract address/ABI propagation to shared mount volumes.
 - Healthcheck orchestration and dependency startup ordering (e.g., PostgreSQL & Kafka ready before backend services start).
- **Out of Scope / Handled Elsewhere:**
 - Multi-node Kubernetes production orchestration (Prompt 802).
 - Cloud infrastructure provisioning on AWS/GCP (Prompt 805).
 - Production Hardware Security Module (HSM) key custody integrations (Prompt 311, 707).

## Technology to Use
- **Docker Engine 24.x+ & Docker Compose v2.24+**: Industry standard for multi-container orchestration. Justification: Docker Compose v2 provides fast, reproducible container lifecycle management, native compose profiles (e.g. `--profile core`, `--profile full`), and robust healthcheck dependency management (`condition: service_healthy`) across Linux, macOS, and Windows workstations.
- **Hyperledger Besu (v24.x)**: Official permissioned Ethereum-compatible enterprise node running in private QBFT mode.
- **PostgreSQL 16 (Alpine)**: Core relational transactional datastore with pre-loaded UUID and pg_stat_statements extensions.
- **Redis 7 (Alpine)**: In-memory cache, session store, and pub/sub message broker.
- **Apache Kafka 3.7+ (KRaft mode via Confluent Community Image)**: Distributed event bus running without Zookeeper overhead to minimize local memory footprints.
- **Foundry (Forge/Cast)**: Blazing-fast Ethereum development toolkit used in a containerized bootstrap container to deploy contracts and generate ABIs.
- **WireMock / Prism / Go Mocks**: Ultra-fast mock API containers for simulating banking and depository partner systems.

## Backend / Infra Touchpoints
- **PostgreSQL**: Port `5432` (`growww_core_dev`, `growww_ledger_dev`, `growww_auth_dev`).
- **Redis**: Port `6379` (DB 0: Sessions/Rate Limits, DB 1: Market Data Cache).
- **Kafka Broker**: Port `9092` (Internal Docker network), Port `29092` (Host workstation access).
- **Kafka Schema Registry**: Port `8081` (Avro/Protobuf schema definitions).
- **Hyperledger Besu RPC / WS**: Ports `8545` (HTTP JSON-RPC) and `8546` (WebSocket).
- **Mock Depository API (NSDL/CDSL)**: Port `9101` (REST simulation of demat credit/debit & pledge verification).
- **Mock Banking Gateway (UPI/IMPS)**: Port `9102` (REST & Webhook simulation of INR payment capture & payout).
- **Mock Identity Service (UIDAI/PAN)**: Port `9103` (REST simulation of OTP generation, e-KYC payload, and PAN-Aadhaar seeding checks).

## Blockchain Interaction
Spins up a single-node Hyperledger Besu container configured with QBFT consensus mechanism, instant 2-second block finality, and zero gas price for deterministic local testing.
- **Genesis & Bootstrapping**: Mounts `besu-qbft-dev.json` defining chain ID `13370`, initial validator addresses, and pre-allocating 10,000 test ETH to five well-known developer relayer addresses.
- **Automated Contract Deployment**: A transient bootstrap container (`growww-contract-deployer`) boots Foundry, executes Solidity migrations for `DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`, and `MultiSigGovernance.sol`, registers initial securities (e.g. `INFY`, `RELIANCE`), and writes the contract addresses to `/shared/contracts/addresses.json`.
- **Event Streaming**: Besu emits block and log events over WebSocket port `8546`, allowing the local event indexer service (Prompt 309) and backend settlement orchestrator (Prompt 208) to subscribe to `SettlementExecuted`, `TokensMinted`, and `TokensBurned` events.

## Step-by-Step Build Instructions
1. Create directory structure: `deployments/local/{genesis,mocks,scripts,volumes,seeds}`.
2. Construct the Besu genesis block `besu-qbft-dev.json` specifying QBFT consensus parameters, block period of 2 seconds, and chain ID `13370`.
3. Create the mock NSDL/CDSL Depository server using WireMock or lightweight Go/FastAPI container exposing endpoints `/v1/depository/holdings`, `/v1/depository/lock`, and `/v1/depository/settle`.
4. Create the mock UPI payment aggregator container exposing `/v1/upi/collect`, `/v1/upi/payout`, and webhook simulator `/v1/upi/simulate-callback`.
5. Create the mock UIDAI/NSDL PAN KYC service exposing `/v1/kyc/aadhaar/generate-otp`, `/v1/kyc/aadhaar/verify-otp`, and `/v1/kyc/pan/verify`.
6. Write the master `deployments/local/docker-compose.yml` declaring network `growww-local-net`, volumes, services, and healthcheck contracts.
7. Configure PostgreSQL container with initialization SQL scripts creating databases and roles (`growww_user`, `growww_settlement`, `growww_kyc`).
8. Configure Kafka in KRaft mode with environment variables configuring automatic topic creation (`trades.executed`, `orders.placed`, `compliance.kyc.events`, `ledger.blockchain.events`).
9. Build the `growww-contract-deployer` container with Foundry that compiles Solidity contracts, runs migrations against `http://besu-node:8545`, and writes artifacts to a mounted docker volume.
10. Write `deployments/local/scripts/seed-data.sh` to inject initial equity master records, standard user tiers, and test market data tickers via REST/Kafka.
11. Implement the `Makefile` with targets: `dev-up`, `dev-down`, `dev-restart`, `dev-logs`, `dev-clean`, and `dev-seed`.
12. Create `.env.example` with dummy developer credentials and instructions.
13. Execute `make dev-up` to verify cold-start launch finishes cleanly under 90 seconds.
14. Validate cross-container communication (e.g. backend service connecting to Kafka, PostgreSQL, Redis, and Besu RPC).

## Interfaces / Contracts
```yaml
# deployments/local/docker-compose.yml (Excerpt)
version: '3.8'

networks:
  growww-local-net:
    driver: bridge
    ipam:
      config:
 - subnet: 172.28.0.0/16

volumes:
  postgres_data:
  redis_data:
  besu_data:
  contract_artifacts:

services:
  besu-node:
    image: hyperledger/besu:24.1.0
    container_name: growww-besu-dev
    command:
 - --config-file=/config/besu-config.toml
 - --network=dev
 - --genesis-file=/config/genesis.json
 - --data-path=/data
 - --rpc-http-enabled=true
 - --rpc-http-host=0.0.0.0
 - --rpc-http-port=8545
 - --rpc-http-api=ETH,NET,WEB3,QBFT,TXPOOL
 - --rpc-http-cors-origins=*
 - --rpc-ws-enabled=true
 - --rpc-ws-host=0.0.0.0
 - --rpc-ws-port=8546
 - --rpc-ws-api=ETH,NET,WEB3,QBFT
    volumes:
 - ./genesis/besu-qbft-dev.json:/config/genesis.json:ro
 - ./genesis/besu-config.toml:/config/besu-config.toml:ro
 - besu_data:/data
    networks:
 - growww-local-net
    ports:
 - "8545:8545"
 - "8546:8546"
    healthcheck:
      test: ["CMD-SHELL", "curl -sf -X POST --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}' -H 'Content-Type: application/json' http://localhost:8545 || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 10

  contract-deployer:
    image: ghcr.io/foundry-rs/foundry:latest
    container_name: growww-contract-deployer
    depends_on:
      besu-node:
        condition: service_healthy
    volumes:
 - ../../contracts:/workspace
 - contract_artifacts:/workspace/out
 - ./scripts/init-contracts.sh:/init-contracts.sh:ro
    entrypoint: ["/bin/sh", "/init-contracts.sh"]
    networks:
 - growww-local-net
```

## Security & Compliance Notes
- All private keys, JWT signing secrets, and database passwords provided in `.env.example` must use obvious dummy values (e.g. `0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80`) and must fail startup if used in a staging/production profile.
- Mock KYC and Aadhaar servers must never store real Indian citizen PII or real Aadhaar/PAN formats that match production databases.
- Container boundaries enforce non-root users (`USER 10001:10001`) where applicable to mirror production Kubernetes security contexts.

## Acceptance Criteria
- [ ] Running `make dev-up` initiates all infrastructure containers and reaches a healthy, steady state in <90 seconds.
- [ ] Besu dev node produces blocks every 2 seconds and responds to `eth_blockNumber` via JSON-RPC.
- [ ] Smart contracts are automatically deployed upon startup, and `/shared/contracts/addresses.json` contains valid contract addresses for `DigitalSecurityToken` and `SettlementDvP`.
- [ ] PostgreSQL, Redis, and Kafka pass container healthchecks and are accessible from host development tools on designated localhost ports.
- [ ] Mock NSDL, UPI, and UIDAI APIs correctly handle positive and negative simulated responses (e.g. simulated Insufficient Demat Balance error via header trigger).
- [ ] `make dev-clean` completely resets local volumes and state to a pristine condition.

## Suggested Order / Dependencies
- Prerequisites: Prompt 106 (Repository Layout), Prompt 108 (Environment Strategy), Prompt 303/306 (Smart Contract Definitions).
- Parallel Tasks: Prompt 803 (CI Pipeline), Prompt 805 (Terraform Cloud Infra).
