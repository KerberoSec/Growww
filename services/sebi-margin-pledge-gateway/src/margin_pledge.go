// Package main implements the SEBI margin pledge/unpledge gateway for NBSE.
// Handles client collateral pledge lifecycle, NSDL/CDSL depository integration,
// and regulatory margin pledge compliance per SEBI circular SEBI/HO/MRD/DRMNP/CIR/P/2020/127.
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// DepositoryType represents NSDL or CDSL.
type DepositoryType string

const (
	DepositoryNSDL DepositoryType = "NSDL"
	DepositoryCDSL DepositoryType = "CDSL"
)

// PledgeStatus represents the lifecycle state of a pledge request.
type PledgeStatus string

const (
	PledgeStatusInitiated    PledgeStatus = "INITIATED"
	PledgeStatusPending      PledgeStatus = "PENDING_DEPOSITORY"
	PledgeStatusConfirmed    PledgeStatus = "CONFIRMED"
	PledgeStatusRejected     PledgeStatus = "REJECTED"
	PledgeStatusUnpledged    PledgeStatus = "UNPLEDGED"
	PledgeStatusPartialFree  PledgeStatus = "PARTIAL_FREE"
)

// PledgeRequest represents a client collateral pledge or unpledge request.
type PledgeRequest struct {
	PledgeID       string         `json:"pledge_id"`
	ClientID       string         `json:"client_id"`
	ISIN           string         `json:"isin"`
	Symbol         string         `json:"symbol"`
	Quantity       int64          `json:"quantity"`
	HaircutPct     float64        `json:"haircut_pct"`     // SEBI-mandated VaR/ELM haircut
	CollateralINR  float64        `json:"collateral_inr"`  // Post-haircut collateral value
	MarketPriceINR float64        `json:"market_price_inr"`
	Depository     DepositoryType `json:"depository"`
	Status         PledgeStatus   `json:"status"`
	DepositoryRef  string         `json:"depository_ref"` // NSDL/CDSL reference
	CreatedAt      time.Time      `json:"created_at"`
	ConfirmedAt    time.Time      `json:"confirmed_at"`
	UnpledgedAt    time.Time      `json:"unpledged_at"`
}

// PledgeGateway manages the pledge/unpledge lifecycle against NSDL/CDSL.
type PledgeGateway struct {
	mu             sync.RWMutex
	pledges        map[string]*PledgeRequest          // pledgeID -> request
	clientPledges  map[string]map[string]bool          // clientID -> set of pledgeIDs
	totalCollateral map[string]float64                 // clientID -> total post-haircut collateral INR
}

// NewPledgeGateway creates a new PledgeGateway instance.
func NewPledgeGateway() *PledgeGateway {
	return &PledgeGateway{
		pledges:         make(map[string]*PledgeRequest),
		clientPledges:   make(map[string]map[string]bool),
		totalCollateral: make(map[string]float64),
	}
}

// InitiatePledge creates a new margin pledge request and sends it to the depository.
// Applies SEBI-mandated VaR+ELM haircut to compute effective collateral value.
func (g *PledgeGateway) InitiatePledge(
	pledgeID, clientID, isin, symbol string,
	quantity int64,
	marketPriceINR, haircutPct float64,
	depository DepositoryType,
) (*PledgeRequest, error) {
	if quantity <= 0 {
		return nil, errors.New("pledge quantity must be positive")
	}
	if marketPriceINR <= 0 {
		return nil, errors.New("market price must be positive")
	}
	if haircutPct < 0 || haircutPct >= 100 {
		return nil, errors.New("haircut percentage must be in [0, 100)")
	}
	if isin == "" || len(isin) != 12 {
		return nil, errors.New("valid 12-character ISIN required")
	}
	if depository != DepositoryNSDL && depository != DepositoryCDSL {
		return nil, errors.New("depository must be NSDL or CDSL")
	}

	grossValue := float64(quantity) * marketPriceINR
	collateralINR := grossValue * (1.0 - haircutPct/100.0)

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.pledges[pledgeID]; exists {
		return nil, fmt.Errorf("pledge ID %s already exists", pledgeID)
	}

	req := &PledgeRequest{
		PledgeID:       pledgeID,
		ClientID:       clientID,
		ISIN:           isin,
		Symbol:         symbol,
		Quantity:        quantity,
		HaircutPct:     haircutPct,
		CollateralINR:  collateralINR,
		MarketPriceINR: marketPriceINR,
		Depository:     depository,
		Status:         PledgeStatusPending,
		CreatedAt:      time.Now().UTC(),
	}

	g.pledges[pledgeID] = req

	if g.clientPledges[clientID] == nil {
		g.clientPledges[clientID] = make(map[string]bool)
	}
	g.clientPledges[clientID][pledgeID] = true

	return req, nil
}

// ConfirmPledge processes depository confirmation (NSDL/CDSL callback).
func (g *PledgeGateway) ConfirmPledge(pledgeID, depositoryRef string) (*PledgeRequest, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	req, exists := g.pledges[pledgeID]
	if !exists {
		return nil, fmt.Errorf("pledge ID %s not found", pledgeID)
	}
	if req.Status != PledgeStatusPending {
		return nil, fmt.Errorf("pledge %s is in state %s, expected PENDING_DEPOSITORY", pledgeID, req.Status)
	}
	if depositoryRef == "" {
		return nil, errors.New("depository reference must be provided")
	}

	req.Status = PledgeStatusConfirmed
	req.DepositoryRef = depositoryRef
	req.ConfirmedAt = time.Now().UTC()

	g.totalCollateral[req.ClientID] += req.CollateralINR

	return req, nil
}

// RejectPledge handles depository rejection.
func (g *PledgeGateway) RejectPledge(pledgeID, reason string) (*PledgeRequest, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	req, exists := g.pledges[pledgeID]
	if !exists {
		return nil, fmt.Errorf("pledge ID %s not found", pledgeID)
	}
	if req.Status != PledgeStatusPending {
		return nil, fmt.Errorf("pledge %s is in state %s, cannot reject", pledgeID, req.Status)
	}

	req.Status = PledgeStatusRejected
	return req, nil
}

// InitiateUnpledge releases pledged securities back to the client demat account.
func (g *PledgeGateway) InitiateUnpledge(pledgeID string) (*PledgeRequest, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	req, exists := g.pledges[pledgeID]
	if !exists {
		return nil, fmt.Errorf("pledge ID %s not found", pledgeID)
	}
	if req.Status != PledgeStatusConfirmed {
		return nil, fmt.Errorf("only CONFIRMED pledges can be unpledged, current: %s", req.Status)
	}

	req.Status = PledgeStatusUnpledged
	req.UnpledgedAt = time.Now().UTC()

	g.totalCollateral[req.ClientID] -= req.CollateralINR
	if g.totalCollateral[req.ClientID] < 0 {
		g.totalCollateral[req.ClientID] = 0
	}

	return req, nil
}

// GetClientCollateral returns the total post-haircut collateral for a client.
func (g *PledgeGateway) GetClientCollateral(clientID string) float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.totalCollateral[clientID]
}

// GetPledge retrieves a pledge by ID.
func (g *PledgeGateway) GetPledge(pledgeID string) (*PledgeRequest, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	req, exists := g.pledges[pledgeID]
	if !exists {
		return nil, fmt.Errorf("pledge ID %s not found", pledgeID)
	}
	return req, nil
}

// GetClientPledges returns all pledges for a client.
func (g *PledgeGateway) GetClientPledges(clientID string) []*PledgeRequest {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []*PledgeRequest
	for pledgeID := range g.clientPledges[clientID] {
		if p, ok := g.pledges[pledgeID]; ok {
			result = append(result, p)
		}
	}
	return result
}
