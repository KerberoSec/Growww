# Growww Platform Environment Strategy, Configuration Management & 12-Factor App Standards

## 1. 5-Tier Environment Taxonomy
Growww operates across 5 discrete environment tiers with strict network, data, and blockchain isolation:

| Environment | Purpose | Blockchain Network | Chain ID | Data Isolation |
| :--- | :--- | :--- | :--- | :--- |
| **Local Dev** | Local developer testing via Docker Compose | Local Besu / Anvil | `1337` | Synthetic mock seed data |
| **CI / Test** | Automated pull request validation & unit/integration tests | Ephemeral Besu | `20241` | Ephemeral test containers |
| **Sandbox / UAT** | Regulatory sandbox testing with SEBI, RBI, and bank partners | Besu Testnet QBFT | `20242` | Sanitized anonymized test data |
| **Staging** | Production replica for pre-release performance and chaos drills | Besu Pre-Prod Consortium | `10085` | Production-mirror synthetic load |
| **Production** | Live institutional trading and settlement | Besu Mainnet Consortium | `10086` | Real institutional assets (HSM secured) |

---

## 2. Configuration Precedence Hierarchy (12-Factor Factor III)
Configuration is loaded following strict precedence order (lowest to highest priority):
1. Code defaults
2. Base configuration (`config/base.yaml`)
3. Environment file (`config/{env}.yaml`)
4. Container environment variables (`GROWWW_*`)
5. Command-line flags (`--config.override`)

---

## 3. Environment Variable Naming Conventions
- All platform environment variables must use uppercase `GROWWW_` prefix with underscore separation:
  - `GROWWW_ENV` (`local`, `test`, `uat`, `staging`, `prod`)
  - `GROWWW_PORT` (e.g. `8080`)
  - `GROWWW_KAFKA_BROKERS` (e.g. `kafka-1:9092,kafka-2:9092`)
  - `GROWWW_DATABASE_URL`
  - `GROWWW_BESU_RPC_URL`
  - `GROWWW_BESU_CHAIN_ID`

---

## 4. Contract Address Registry
Microservices must never hardcode smart contract addresses in source code. Addresses are injected via environment configuration maps:
- `GROWWW_CONTRACT_SETTLEMENT_DVP`
- `GROWWW_CONTRACT_DIGITAL_RUPEE`
- `GROWWW_CONTRACT_SOLVENCY_REGISTRY`
- `GROWWW_CONTRACT_HYBRID_POOL`
- `GROWWW_CONTRACT_FEE_COLLECTOR`

---

## 5. Feature Flags & Kill Switches
- Centralized runtime feature flagging allows progressive percentage rollouts, user whitelisting, and instant emergency circuit-breaker kill-switches.
- Feature flags evaluate in-memory with sub-millisecond latency.
