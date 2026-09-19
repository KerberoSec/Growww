package main

import (
	"sync"
	"time"
)

type VaultRebalanceProposal struct {
	ProposalID      string
	AssetSymbol     string
	Amount          float64
	SourceType      string // HOT_WALLET, COLD_VAULT
	DestinationType string // COLD_VAULT, HOT_WALLET
	Reason          string // THRESHOLD_BREACH, DEFICIT_REPLENISHMENT
	ProposedAt      time.Time
	Signatures      []string
	RequiredSigners int
	Executed        bool
}

type HotColdRebalancer struct {
	mu           sync.Mutex
	hotCapacity  map[string]float64
	hotBalances  map[string]float64
	coldBalances map[string]float64
	proposals    []*VaultRebalanceProposal
}

func NewHotColdRebalancer() *HotColdRebalancer {
	return &HotColdRebalancer{
		hotCapacity:  map[string]float64{"BTC": 10.0, "USDT": 500000.0, "ETH": 100.0},
		hotBalances:  make(map[string]float64),
		coldBalances: make(map[string]float64),
	}
}

func (r *HotColdRebalancer) SetBalances(asset string, hot, cold float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hotBalances[asset] = hot
	r.coldBalances[asset] = cold
}

func (r *HotColdRebalancer) EvaluateSweepToCold(asset string) *VaultRebalanceProposal {
	r.mu.Lock()
	defer r.mu.Unlock()
	hot := r.hotBalances[asset]
	cap := r.hotCapacity[asset]

	if hot > cap {
		excess := hot - (cap * 0.8) // Sweep down to 80% of capacity
		prop := &VaultRebalanceProposal{
			ProposalID:      time.Now().Format("20060102150405-") + asset,
			AssetSymbol:     asset,
			Amount:          excess,
			SourceType:      "HOT_WALLET",
			DestinationType: "COLD_VAULT",
			Reason:          "THRESHOLD_BREACH",
			ProposedAt:      time.Now().UTC(),
			RequiredSigners: 3,
		}
		r.proposals = append(r.proposals, prop)
		return prop
	}
	return nil
}
