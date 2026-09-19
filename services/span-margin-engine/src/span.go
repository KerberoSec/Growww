package main
import ("math";"sync")
type ScenarioResult struct { PriceShiftPct float64; VolShiftPct float64; PnL float64 }
type PositionRisk struct { Symbol string; Quantity float64; EntryPrice float64; CurrentPrice float64; IsLong bool }
type SPANEngine struct { mu sync.Mutex; scenarios []struct{PriceShift, VolShift float64} }
func NewSPANEngine() *SPANEngine {
	return &SPANEngine{scenarios: []struct{PriceShift, VolShift float64}{
		{-0.15, -0.10}, {-0.10, -0.05}, {-0.05, 0}, {0, 0}, {0.05, 0}, {0.10, 0.05}, {0.15, 0.10},
	}}
}
func (e *SPANEngine) ComputeMargin(positions []PositionRisk) float64 {
	e.mu.Lock(); defer e.mu.Unlock()
	worstLoss := 0.0
	for _, s := range e.scenarios {
		scenarioLoss := 0.0
		for _, p := range positions {
			shiftedPrice := p.CurrentPrice * (1 + s.PriceShift)
			if p.IsLong { scenarioLoss += p.Quantity * (p.CurrentPrice - shiftedPrice) } else { scenarioLoss += p.Quantity * (shiftedPrice - p.CurrentPrice) }
		}
		if scenarioLoss > worstLoss { worstLoss = scenarioLoss }
	}
	return math.Max(worstLoss, 0)
}
