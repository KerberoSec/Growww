package main

import (
	"fmt"
	"strings"
)

type PortfolioContext struct {
	UserID         string
	EquitiesINR    float64
	VDACryptoINR   float64
	CBDCBalanceINR float64
	UnrealizedPnL  float64
	KYCTier        int
}

type QueryIntent string

const (
	IntentTaxInquiry      QueryIntent = "TAX_INQUIRY"
	IntentPortfolioStatus QueryIntent = "PORTFOLIO_STATUS"
	IntentOrderPlacement  QueryIntent = "ORDER_PLACEMENT"
	IntentGeneralFAQ      QueryIntent = "GENERAL_FAQ"
)

// ClassifyQueryIntent routes natural language queries to the appropriate sub-engine.
func ClassifyQueryIntent(query string) QueryIntent {
	q := strings.ToLower(query)
	if strings.Contains(q, "tax") || strings.Contains(q, "tds") || strings.Contains(q, "115bbh") || strings.Contains(q, "194s") {
		return IntentTaxInquiry
	}
	if strings.Contains(q, "portfolio") || strings.Contains(q, "holding") || strings.Contains(q, "balance") || strings.Contains(q, "pnl") {
		return IntentPortfolioStatus
	}
	if strings.Contains(q, "buy") || strings.Contains(q, "sell") || strings.Contains(q, "order") || strings.Contains(q, "trade") {
		return IntentOrderPlacement
	}
	return IntentGeneralFAQ
}

// BuildPromptContext injects real user financial state securely into LLM context window.
func BuildPromptContext(ctx PortfolioContext, userQuery string) string {
	return fmt.Sprintf(
		"USER_FINANCIAL_CONTEXT:\n- Equities: INR %.2f\n- VDA Crypto: INR %.2f\n- CBDC Cash: INR %.2f\n- Unrealized PnL: INR %.2f\n- KYC Tier: %d\n\nUSER_QUERY: %s\n\nINSTRUCTION: Provide strict, non-advisory, factual calculations complying with Indian SEBI/IT Act rules.",
		ctx.EquitiesINR, ctx.VDACryptoINR, ctx.CBDCBalanceINR, ctx.UnrealizedPnL, ctx.KYCTier, userQuery,
	)
}
