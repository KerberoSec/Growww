package src

import (
	"os"
	"testing"
)

func TestConfigLoader_DefaultLocalEnvironment(t *testing.T) {
	os.Unsetenv("GROWWW_ENV")
	os.Unsetenv("GROWWW_PORT")
	os.Unsetenv("GROWWW_BESU_CHAIN_ID")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Environment != EnvLocal {
		t.Errorf("expected EnvLocal, got %s", cfg.Environment)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.BesuChainID != 1337 {
		t.Errorf("expected default local chain ID 1337, got %d", cfg.BesuChainID)
	}
}

func TestConfigLoader_ProductionValidation(t *testing.T) {
	os.Setenv("GROWWW_ENV", "prod")
	os.Setenv("GROWWW_PORT", "9000")
	os.Setenv("GROWWW_BESU_CHAIN_ID", "10086")
	os.Setenv("GROWWW_CONTRACT_DVP", "0x1234567890abcdef1234567890abcdef12345678")
	defer func() {
		os.Unsetenv("GROWWW_ENV")
		os.Unsetenv("GROWWW_PORT")
		os.Unsetenv("GROWWW_BESU_CHAIN_ID")
		os.Unsetenv("GROWWW_CONTRACT_DVP")
	}()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Environment != EnvProd {
		t.Errorf("expected prod, got %s", cfg.Environment)
	}
	if cfg.Port != 9000 {
		t.Errorf("expected port 9000, got %d", cfg.Port)
	}
	if cfg.BesuChainID != 10086 {
		t.Errorf("expected prod chain ID 10086, got %d", cfg.BesuChainID)
	}
	if cfg.ContractAddresses["dvp"] != "0x1234567890abcdef1234567890abcdef12345678" {
		t.Errorf("expected dvp contract address mapped, got %+v", cfg.ContractAddresses)
	}
}

func TestConfigLoader_ChainIDMismatchError(t *testing.T) {
	os.Setenv("GROWWW_ENV", "prod")
	os.Setenv("GROWWW_BESU_CHAIN_ID", "1337") // Mismatch: local chain ID in prod!
	defer func() {
		os.Unsetenv("GROWWW_ENV")
		os.Unsetenv("GROWWW_BESU_CHAIN_ID")
	}()

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Errorf("expected chain ID mismatch error, but passed")
	}
}

func TestFeatureFlags_EvaluationAndKillSwitch(t *testing.T) {
	ff := NewFeatureFlagManager()

	// Default unconfigured flag should be false
	if ff.IsEnabled("crypto_earn_v2") {
		t.Errorf("expected default disabled")
	}

	err := ff.RequireFlag("crypto_earn_v2")
	if err == nil {
		t.Errorf("expected error when required disabled flag")
	}

	// Enable flag
	ff.SetFlag("crypto_earn_v2", true)
	if !ff.IsEnabled("crypto_earn_v2") {
		t.Errorf("expected flag enabled")
	}
	if err := ff.RequireFlag("crypto_earn_v2"); err != nil {
		t.Errorf("expected RequireFlag to pass when enabled: %v", err)
	}

	// Emergency kill switch
	ff.SetFlag("crypto_earn_v2", false)
	if ff.IsEnabled("crypto_earn_v2") {
		t.Errorf("expected flag killed")
	}
}
