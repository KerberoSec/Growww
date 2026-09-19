package main

import (
	"fmt"
	"sync"
	"time"
)

type ClearingCorporation string

const (
	CC_NSCCL ClearingCorporation = "NSCCL" // National Securities Clearing Corp
	CC_ICCL  ClearingCorporation = "ICCL"  // Indian Clearing Corp
	CC_MCXCC ClearingCorporation = "MCXCCL"
)

type InteropSettlementObligation struct {
	BatchID          string              `json:"batch_id"`
	ClearingCorp     ClearingCorporation `json:"clearing_corp"`
	NetFundsPayable  float64             `json:"net_funds_payable_inr"`
	NetFundsRecv     float64             `json:"net_funds_recv_inr"`
	DeliveryShares   map[string]uint64   `json:"delivery_shares"` // ISIN -> Shares
	ReceiveShares    map[string]uint64   `json:"receive_shares"`
	SettlementStatus string              `json:"settlement_status"`
	Timestamp        time.Time           `json:"timestamp"`
}

type ClearingCorpGateway struct {
	mu          sync.Mutex
	obligations map[string]*InteropSettlementObligation
}

func NewClearingCorpGateway() *ClearingCorpGateway {
	return &ClearingCorpGateway{
		obligations: make(map[string]*InteropSettlementObligation),
	}
}

// GenerateMultilateralNetting computes netting obligations across SEBI clearing corporations
func (g *ClearingCorpGateway) GenerateMultilateralNetting(batchID string, cc ClearingCorporation, payable, recv float64) *InteropSettlementObligation {
	g.mu.Lock()
	defer g.mu.Unlock()

	ob := &InteropSettlementObligation{
		BatchID:          batchID,
		ClearingCorp:     cc,
		NetFundsPayable:  payable,
		NetFundsRecv:     recv,
		DeliveryShares:   make(map[string]uint64),
		ReceiveShares:    make(map[string]uint64),
		SettlementStatus: "NETTED_OBLIGATION_CONFIRMED",
		Timestamp:        time.Now().UTC(),
	}

	g.obligations[batchID] = ob
	fmt.Printf("[Clearing Interop] Multilateral netting generated for %s batch %s\n", cc, batchID)
	return ob
}
