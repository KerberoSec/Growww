package src

import (
	"testing"
)

func TestValidateTopicName_ValidCases(t *testing.T) {
	valid := []string{
		"prod.domestic.trading.order.matched.v1",
		"staging.giftcity.settlement.dvp.executed.v2",
		"testnet.domestic.compliance.investor.whitelisted.v1",
		"dev.global.custody.vault.rebalanced.v3",
	}

	for _, topic := range valid {
		parsed, err := ValidateTopicName(topic)
		if err != nil {
			t.Errorf("expected '%s' to be valid, got error: %v", topic, err)
		}
		if parsed == nil {
			t.Fatalf("expected non-nil parsed topic for '%s'", topic)
		}
	}
}

func TestValidateTopicName_InvalidCases(t *testing.T) {
	invalid := []string{
		"invalid_segments_count",
		"production.domestic.trading.order.matched.v1", // 'production' not in env map
		"prod.mars.trading.order.matched.v1",           // 'mars' not valid entity
		"prod.domestic.trading.order.matched.v0",        // v0 invalid version
		"prod.domestic.trading.order.matched.version1",  // version format invalid
	}

	for _, topic := range invalid {
		_, err := ValidateTopicName(topic)
		if err == nil {
			t.Errorf("expected '%s' to fail validation", topic)
		}
	}
}

func TestSelectPartitionKey(t *testing.T) {
	// 1. Trading order -> user ID partition
	keyUser, err := SelectPartitionKey("trading", "user_12345", "BTC/USDT", "")
	if err != nil || keyUser != "user_12345" {
		t.Errorf("expected user_12345, got %s (err: %v)", keyUser, err)
	}

	// 2. Market matching -> symbol partition
	keySymbol, err := SelectPartitionKey("matching", "user_12345", "BTC/USDT", "")
	if err != nil || keySymbol != "BTC/USDT" {
		t.Errorf("expected BTC/USDT, got %s (err: %v)", keySymbol, err)
	}

	// 3. Settlement clearing -> settlement ID partition
	keyDvp, err := SelectPartitionKey("settlement", "user_12345", "BTC/USDT", "settle_9900")
	if err != nil || keyDvp != "settle_9900" {
		t.Errorf("expected settle_9900, got %s (err: %v)", keyDvp, err)
	}
}
