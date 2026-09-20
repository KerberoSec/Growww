package src

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PlatformEnvironment represents the 5-tier deployment taxonomy
type PlatformEnvironment string

const (
	EnvLocal   PlatformEnvironment = "local"
	EnvTest    PlatformEnvironment = "test"
	EnvUat     PlatformEnvironment = "uat"
	EnvStaging PlatformEnvironment = "staging"
	EnvProd    PlatformEnvironment = "prod"
)

var ValidChainIDs = map[PlatformEnvironment]uint64{
	EnvLocal:   1337,
	EnvTest:    20241,
	EnvUat:     20242,
	EnvStaging: 10085,
	EnvProd:    10086,
}

// ServiceConfig defines strongly typed 12-factor configuration
type ServiceConfig struct {
	Environment        PlatformEnvironment `json:"environment"`
	Port               int                 `json:"port"`
	KafkaBrokers       []string            `json:"kafka_brokers"`
	DatabaseURL        string              `json:"database_url"`
	BesuRPCURL         string              `json:"besu_rpc_url"`
	BesuChainID        uint64              `json:"besu_chain_id"`
	ContractAddresses  map[string]string   `json:"contract_addresses"`
}

// LoadConfigFromEnv builds and validates ServiceConfig from environment variables
func LoadConfigFromEnv() (*ServiceConfig, error) {
	envStr := strings.ToLower(os.Getenv("GROWWW_ENV"))
	if envStr == "" {
		envStr = "local"
	}

	env := PlatformEnvironment(envStr)
	switch env {
	case EnvLocal, EnvTest, EnvUat, EnvStaging, EnvProd:
	default:
		return nil, fmt.Errorf("unsupported environment '%s'", env)
	}

	port := 8080
	if portStr := os.Getenv("GROWWW_PORT"); portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return nil, fmt.Errorf("invalid port '%s'", portStr)
		}
		port = p
	}

	kafkaBrokers := []string{"localhost:9092"}
	if brokers := os.Getenv("GROWWW_KAFKA_BROKERS"); brokers != "" {
		kafkaBrokers = strings.Split(brokers, ",")
	}

	expectedChainID := ValidChainIDs[env]
	chainID := expectedChainID
	if cidStr := os.Getenv("GROWWW_BESU_CHAIN_ID"); cidStr != "" {
		cid, err := strconv.ParseUint(cidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chain ID: %w", err)
		}
		if cid != expectedChainID {
			return nil, fmt.Errorf("chain ID mismatch: environment %s requires %d, got %d", env, expectedChainID, cid)
		}
		chainID = cid
	}

	contracts := make(map[string]string)
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "GROWWW_CONTRACT_") {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				name := strings.ToLower(strings.TrimPrefix(parts[0], "GROWWW_CONTRACT_"))
				contracts[name] = parts[1]
			}
		}
	}

	return &ServiceConfig{
		Environment:       env,
		Port:              port,
		KafkaBrokers:      kafkaBrokers,
		DatabaseURL:       os.Getenv("GROWWW_DATABASE_URL"),
		BesuRPCURL:        os.Getenv("GROWWW_BESU_RPC_URL"),
		BesuChainID:       chainID,
		ContractAddresses: contracts,
	}, nil
}

// FeatureFlagManager manages dynamic runtime toggles and kill-switches
type FeatureFlagManager struct {
	flags map[string]bool
}

func NewFeatureFlagManager() *FeatureFlagManager {
	return &FeatureFlagManager{
		flags: make(map[string]bool),
	}
}

func (f *FeatureFlagManager) SetFlag(key string, enabled bool) {
	f.flags[key] = enabled
}

func (f *FeatureFlagManager) IsEnabled(key string) bool {
	enabled, exists := f.flags[key]
	if !exists {
		return false // safe default: disabled
	}
	return enabled
}

func (f *FeatureFlagManager) RequireFlag(key string) error {
	if !f.IsEnabled(key) {
		return fmt.Errorf("feature '%s' is disabled by kill switch", key)
	}
	return nil
}
