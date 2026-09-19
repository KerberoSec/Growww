package main

import (
	"strings"
	"testing"
)

func TestClassifyQueryIntent(t *testing.T) {
	if ClassifyQueryIntent("What is my Section 194S TDS liability?") != IntentTaxInquiry {
		t.Error("expected TAX_INQUIRY")
	}
	if ClassifyQueryIntent("Show my current portfolio balance and PnL") != IntentPortfolioStatus {
		t.Error("expected PORTFOLIO_STATUS")
	}
	if ClassifyQueryIntent("Place a buy limit order for 0.5 BTC") != IntentOrderPlacement {
		t.Error("expected ORDER_PLACEMENT")
	}
}

func TestBuildPromptContext(t *testing.T) {
	ctx := PortfolioContext{UserID: "u1", EquitiesINR: 50000, VDACryptoINR: 100000, CBDCBalanceINR: 25000, KYCTier: 2}
	prompt := BuildPromptContext(ctx, "Calculate my tax")
	if !strings.Contains(prompt, "Equities: INR 50000.00") {
		t.Error("context missing equities balance")
	}
}
