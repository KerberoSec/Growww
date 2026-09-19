# 108 - Environment Strategy, Configuration Management & 12-Factor App Architecture

## Purpose
Establishes the enterprise environment lifecycle, configuration management architecture, dynamic feature flagging, and 12-Factor App compliance standards for the Growww investment platform. Growww operates across 5 discrete environment tiers (Local Development, Testing/CI, Sandbox/UAT, Staging, and Multi-Region Production), each requiring strict isolation of network boundaries, banking mock adapters, and permissioned blockchain networks.

This prompt provides developers and DevOps engineers with the canonical standards for environment variables, strongly-typed configuration schemas (Pydantic in Python, Viper in Go, config-rs in Rust), Kubernetes ConfigMaps, Helm values hierarchies, feature flag management, and test dataset sanitization.

## What You Are Building
A comprehensive environment and configuration architecture document (`docs/standards/environment_and_config.md`), configuration schema specifications, and Helm values hierarchy:
- 5-Tier Environment Taxonomy: Local Dev (Docker Compose), CI/Test (Ephemeral), Sandbox/UAT (SEBI/RBI regulatory testing sandbox), Staging (Production replica), and Production (Multi-AZ / Multi-Region).
- Strongly Typed Configuration Loaders: Schema-validated configuration models per language preventing application startup with missing or malformed configuration parameters.
- Helm Values Hierarchy: Base configuration (`values.yaml`) overridden cleanly per environment (`values-dev.yaml`, `values-uat.yaml`, `values-prod.yaml`).
- Feature Flag Strategy: Centralized runtime feature toggling (Unleash / LaunchDarkly) for safe progressive rollouts, dark launches, and emergency circuit breakers.

## Scope Boundaries
- **In Scope:**
 - Environment tier definitions, naming conventions, and isolation requirements.
 - Configuration schema design and validation rules (12-Factor App Factor III).
 - Feature flag taxonomies, evaluation lifecycles, and kill-switches.
 - Non-secret environment configuration parameters (ports, endpoints, timeouts, thresholds).
- **Out of Scope / Handled Elsewhere:**
 - Secrets storage, KMS, and HSM key custody (handled in Prompt 109).
 - Docker Compose local development stack (handled in Prompt 801).
 - Kubernetes cluster deployment and Terraform infrastructure (handled in Prompt 802 & 805).
 - Environment promotion and release management (handled in Prompt 810).

## Technology to Use
- **Configuration Parsing & Validation:**
 - *Go Services:* `spf13/viper` with mapstructure decoding and strict struct validation.
 - *Rust Services:* `config-rs` with `serde` deserialization and compile-time validation.
 - *Python Services:* `pydantic-settings` (v2.2+) with automatic `.env` loading and type coercion.
 - *Flutter Client:* Compile-time `--dart-define-from-file` environment JSON files.
- **Runtime Feature Flagging:** Unleash (Open Source) / LaunchDarkly SDK with local in-memory evaluation and sub-second flag synchronization.
- **Kubernetes Orchestration:** Helm 3.14+ and Kustomize for multi-environment manifest templating.

## Backend / Infra Touchpoints
- **Kubernetes ConfigMaps:** Non-sensitive environment parameters mounted as container environment variables or volume files.
- **PostgreSQL Databases:** Isolated database clusters per environment with strict VPC peering rules.
- **Kafka Clusters:** Environment-prefixed topic namespaces (`dev.*`, `uat.*`, `prod.*`) or dedicated physical Kafka clusters per tier.
- **Banking / Depository Mock Gateways:** Automated stub adapters for NSDL/CDSL and UPI in Sandbox/UAT environments.

## Blockchain Interaction
Manages multi-environment ledger node endpoints and smart contract address registries:
- **Ledger Network Segregation:**
 - *Local Dev:* Local single-node or 4-node Besu Docker container (Chain ID: `1337`).
 - *CI / Sandbox:* Ephemeral Besu testnet cluster with automated contract deployment (Chain ID: `20241`).
 - *Staging / Production:* High-availability Hyperledger Besu consortium network with QBFT consensus and HSM signing (Chain ID: `10086`).
- **Contract Address Registry:** Microservices dynamically load deployed smart contract addresses (`DigitalSecurityToken`, `SettlementDvP`, `ComplianceRegistry`) from environment-specific configuration maps without requiring hardcoded Solidity addresses in application code.

## Step-by-Step Build Instructions
1. Author `docs/standards/environment_and_config.md` establishing the 5 environment tiers and isolation boundaries.
2. Define the environment variable naming convention: UPPERCASE with snake_case and service prefixes (e.g., `GROWWW_ORDER_SERVICE_PORT`, `GROWWW_KAFKA_BROKERS`).
3. Define the configuration precedence hierarchy (lowest to highest):
   1. Default code values
   2. Base configuration file (`config/base.yaml`)
   3. Environment-specific configuration file (`config/{env}.yaml`)
   4. Kubernetes ConfigMap environment variables
   5. Local developer `.env` file (local dev only)
   6. Command-line flags
4. Implement strongly typed Pydantic configuration model for Python services (`services/user_service/src/config.py`).
5. Implement strongly typed Viper configuration struct for Go services (`services/order_service/pkg/config/config.go`).
6. Implement strongly typed Serde configuration struct for Rust services (`services/matching_engine/src/config.rs`).
7. Define compile-time environment configuration files for Flutter (`apps/client_flutter/config/env_dev.json`, `env_prod.json`).
8. Establish the Helm values directory structure:
 - `infra/k8s/charts/growww-platform/values.yaml` (default base values)
 - `infra/k8s/charts/growww-platform/values-dev.yaml`
 - `infra/k8s/charts/growww-platform/values-uat.yaml`
 - `infra/k8s/charts/growww-platform/values-prod.yaml`
9. Configure feature flag client integrations (Unleash SDK) across backend and client applications.
10. Define standard feature flag lifecycle: `EXPERIMENT` -> `ROLLOUT` -> `PERMANENT` -> `CLEANUP` (mandatory removal within 60 days).
11. Define test dataset sanitization and synthetic data generation rules for lower environments.
12. Review configuration schemas to ensure zero secrets or passwords are present in non-secret config files.

## Interfaces / Contracts

### Go Configuration Specification (`services/order_service/pkg/config/config.go`)
```go
package config

type Config struct {
	Environment    string         `mapstructure:"environment" validate:"required,oneof=local dev uat staging prod"`
	Server         ServerConfig   `mapstructure:"server" validate:"required"`
	Kafka          KafkaConfig    `mapstructure:"kafka" validate:"required"`
	Database       DatabaseConfig `mapstructure:"database" validate:"required"`
	Blockchain     BesuConfig     `mapstructure:"blockchain" validate:"required"`
	FeatureFlags   FeatureConfig  `mapstructure:"features"`
}

type ServerConfig struct {
	GRPCPort        int `mapstructure:"grpc_port" validate:"required,min=1024,max=65535"`
	HTTPPort        int `mapstructure:"http_port" validate:"required,min=1024,max=65535"`
	ShutdownTimeout int `mapstructure:"shutdown_timeout_sec" validate:"required,min=5"`
}

type BesuConfig struct {
	RPCURL                 string `mapstructure:"rpc_url" validate:"required,url"`
	ChainID                int64  `mapstructure:"chain_id" validate:"required"`
	SettlementContractAddr string `mapstructure:"settlement_contract_address" validate:"required,eth_addr"`
	TokenContractAddr      string `mapstructure:"token_contract_address" validate:"required,eth_addr"`
}
```

### Python Pydantic Configuration Model (`services/user_service/src/config.py`)
```python
from pydantic import Field, PostgresDsn, RedisDsn
from pydantic_settings import BaseSettings, SettingsConfigDict

class AppConfig(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="GROWWW_USER_",
        env_file=".env",
        extra="ignore"
    )

    environment: str = Field(default="dev", pattern="^(local|dev|uat|staging|prod)$")
    http_port: int = Field(default=8000, ge=1024, le=65535)
    database_url: PostgresDsn = Field(...)
    redis_url: RedisDsn = Field(...)
    kafka_brokers: str = Field(default="localhost:9092")
    jwt_issuer: str = Field(default="https://auth.growww.in")
    jwt_expiry_minutes: int = Field(default=15, ge=5, le=60)
    enable_mock_aadhaar: bool = Field(default=False)
```

### Helm Values Environment Matrix
| Parameter | Local Dev | UAT / Sandbox | Production |
|---|---|---|---|
| `replicaCount` | 1 | 2 | 5-20 (HPA Auto) |
| `resources.requests.cpu` | 100m | 250m | 1000m |
| `resources.limits.memory` | 256Mi | 512Mi | 2Gi |
| `blockchain.chain_id` | 1337 | 20241 | 10086 |
| `features.enableMockPayment` | `true` | `true` | `false` |
| `logLevel` | `debug` | `info` | `warn` |

## Security & Compliance Notes
- **Strict Data Isolation:** Production databases and storage buckets must never be reachable or replicated into Dev/UAT environments.
- **No Hardcoded Secrets:** Configuration files checked into Git must contain only structural parameters, endpoints, and timeouts; all credentials, API keys, and private keys must be injected dynamically via Vault (Prompt 109).
- **Audit Logging of Feature Flag Toggles:** Every runtime modification of a feature flag in Unleash must generate a non-repudiable audit event capturing the authorizing engineer, timestamp, and justification.

## Acceptance Criteria
- [ ] Complete Environment and Configuration document (`docs/standards/environment_and_config.md`) published.
- [ ] Strongly typed configuration models implemented and validated for Go, Rust, and Python services.
- [ ] Helm values hierarchy (`values.yaml`, `values-dev.yaml`, `values-prod.yaml`) created with zero hardcoded credentials.
- [ ] Feature flag framework (Unleash SDK) specified with lifecycle governance and circuit-breaker support.
- [ ] Contract address registry and chain ID mappings specified for all 5 environment tiers.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 106 (Monorepo Layout).
- **Parallel Work:** Prompt 107 (Coding Standards), Prompt 109 (Secrets Management).
- **Blocks:** Prompt 801 (Docker Compose Dev Stack), Prompt 802 (K8s Architecture), Prompt 805 (Terraform).
