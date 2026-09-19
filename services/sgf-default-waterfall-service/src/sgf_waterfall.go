package main

import (
	"fmt"
	
)

type SGFDefaultWaterfall struct {
	DefaulterMarginINR   float64
	ExchangeDedicatedCap float64 // Core SGF equity
	NonDefaultersSGF     float64
	RecoveryFromDefault  float64
}

func NewSGFDefaultWaterfall(defaulterMargin, coreSGF, nonDefaulterSGF float64) *SGFDefaultWaterfall {
	return &SGFDefaultWaterfall{
		DefaulterMarginINR:   defaulterMargin,
		ExchangeDedicatedCap: coreSGF,
		NonDefaultersSGF:     nonDefaulterSGF,
	}
}

// ResolveDefaultWaterfall applies SEBI Core SGF default waterfall order
func (w *SGFDefaultWaterfall) ResolveDefaultWaterfall(defaultLossINR float64) (remainingLoss float64, steps []string) {
	loss := defaultLossINR
	steps = append(steps, fmt.Sprintf("Initial settlement default loss: ₹%.2f", loss))

	// Step 1: Defaulter's own margin and deposits
	if loss <= w.DefaulterMarginINR {
		w.DefaulterMarginINR -= loss
		steps = append(steps, fmt.Sprintf("Fully absorbed by defaulter's initial margin: ₹%.2f", loss))
		return 0, steps
	}
	loss -= w.DefaulterMarginINR
	steps = append(steps, fmt.Sprintf("Exhausted defaulter margin (₹%.2f). Uncovered: ₹%.2f", w.DefaulterMarginINR, loss))
	w.DefaulterMarginINR = 0

	// Step 2: Clearing Corporation dedicated Core SGF capital (25% platform revenue reserve)
	if loss <= w.ExchangeDedicatedCap {
		w.ExchangeDedicatedCap -= loss
		steps = append(steps, fmt.Sprintf("Absorbed by Exchange Core SGF dedicated tranche: ₹%.2f", loss))
		return 0, steps
	}
	loss -= w.ExchangeDedicatedCap
	steps = append(steps, fmt.Sprintf("Exhausted Exchange Core SGF (₹%.2f). Uncovered: ₹%.2f", w.ExchangeDedicatedCap, loss))
	w.ExchangeDedicatedCap = 0

	// Step 3: Non-defaulters' mutualized pool
	if loss <= w.NonDefaultersSGF {
		w.NonDefaultersSGF -= loss
		steps = append(steps, fmt.Sprintf("Absorbed by mutualized non-defaulters SGF: ₹%.2f", loss))
		return 0, steps
	}
	loss -= w.NonDefaultersSGF
	steps = append(steps, fmt.Sprintf("Critical systemic stress: Uncovered loss ₹%.2f", loss))
	w.NonDefaultersSGF = 0

	return loss, steps
}
