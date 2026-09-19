package main

import (
	"sync"
	"time"
)

type ClearingMember struct {
	MemberID          string
	BaseCapitalINR    float64
	RequiredMarginINR float64
	CollateralINR     float64
	MarginDeficitINR  float64
	Status            string // NORMAL, MARGIN_CALL, DEFAULTED
	LastCallTimestamp time.Time
}

type SGFMarginCallWorker struct {
	mu      sync.RWMutex
	members map[string]*ClearingMember
}

func NewSGFMarginCallWorker() *SGFMarginCallWorker {
	return &SGFMarginCallWorker{members: make(map[string]*ClearingMember)}
}

func (w *SGFMarginCallWorker) RegisterMember(m *ClearingMember) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.members[m.MemberID] = m
}

func (w *SGFMarginCallWorker) EvaluateMarginCalls() []*ClearingMember {
	w.mu.Lock()
	defer w.mu.Unlock()
	var alerted []*ClearingMember
	for _, m := range w.members {
		if m.CollateralINR < m.RequiredMarginINR {
			m.MarginDeficitINR = m.RequiredMarginINR - m.CollateralINR
			m.Status = "MARGIN_CALL"
			m.LastCallTimestamp = time.Now().UTC()
			alerted = append(alerted, m)
		} else {
			m.MarginDeficitINR = 0
			m.Status = "NORMAL"
		}
	}
	return alerted
}
