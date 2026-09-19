package main

import (
	"math"
	"testing"
)

func TestCreatePoolAndQuote(t *testing.T) {
	eng := NewLiquidityEngine()
	pool, err := eng.CreatePool("BTC/USDT", 100.0, 6500000.0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if pool.MidPrice() != 65000.0 {
		t.Errorf("expected mid 65000, got %f", pool.MidPrice())
	}
	cost, err := pool.GetQuote("BUY", 1.0)
	if err != nil {
		t.Fatal(err)
	}
	if cost <= 65000 {
		t.Errorf("buy cost should exceed mid due to slippage+spread, got %f", cost)
	}
}

func TestSellQuote(t *testing.T) {
	eng := NewLiquidityEngine()
	pool, _ := eng.CreatePool("ETH/USDT", 1000.0, 3500000.0, 5)
	proceeds, err := pool.GetQuote("SELL", 10.0)
	if err != nil {
		t.Fatal(err)
	}
	if proceeds <= 0 {
		t.Error("sell proceeds must be positive")
	}
	if proceeds >= 10*3500.0 {
		t.Error("sell proceeds should be less than notional due to slippage+spread")
	}
}

func TestInsufficientLiquidity(t *testing.T) {
	eng := NewLiquidityEngine()
	pool, _ := eng.CreatePool("BTC/USDT", 10.0, 650000.0, 10)
	_, err := pool.GetQuote("BUY", 11.0)
	if err == nil {
		t.Error("expected error for buying more than reserve")
	}
}

func TestKInvariant(t *testing.T) {
	eng := NewLiquidityEngine()
	pool, _ := eng.CreatePool("BTC/USDT", 50.0, 3250000.0, 0)
	expectedK := 50.0 * 3250000.0
	if math.Abs(pool.K-expectedK) > 0.01 {
		t.Errorf("K invariant mismatch: got %f, want %f", pool.K, expectedK)
	}
}
