package main

import (
	"fmt"
	"math"
)

type CopyTradeExecution struct {
	LeaderOrderID    string
	LeaderPrice      float64
	FollowerPrice    float64
	MaxSlippageBps   int // e.g. 50 bps = 0.50%
}

// ValidateFollowerExecution ensures follower fill price does not exceed max slippage.
func ValidateFollowerExecution(exec CopyTradeExecution, isBuy bool) error {
	slippage := (exec.FollowerPrice - exec.LeaderPrice) / exec.LeaderPrice
	if !isBuy {
		slippage = -slippage
	}

	maxAllowed := float64(exec.MaxSlippageBps) / 10000.0
	if slippage > maxAllowed {
		return fmt.Errorf("slippage %.4f exceeds maximum allowed threshold %.4f (rejected for follower protection)", slippage, maxAllowed)
	}
	return nil
}

// ScaleQuantity calculates follower order quantity proportionally based on equity ratio.
func ScaleQuantity(leaderQty, leaderEquity, followerEquity float64) float64 {
	if leaderEquity <= 0 || followerEquity <= 0 {
		return 0
	}
	ratio := followerEquity / leaderEquity
	return math.Round(leaderQty*ratio*1000000) / 1000000
}
