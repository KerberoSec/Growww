package main

import (
	"fmt"
	"sync"
	"time"
)

type ActionPhase string
const (
	Announced ActionPhase = "ANNOUNCED"
	RecordSet ActionPhase = "RECORD_DATE_SET"
	ExDate    ActionPhase = "EX_DATE"
	Paid      ActionPhase = "PAID"
)

type CorpAction struct {
	ID            string
	Symbol        string
	Type          string // DIVIDEND, SPLIT, BONUS, BUYBACK
	AnnouncedDate time.Time
	RecordDate    time.Time
	ExDate        time.Time
	PaymentDate   time.Time
	Phase         ActionPhase
	Ratio         float64
	DividendPerShare float64
}

type CorpActionsService struct {
	mu      sync.Mutex
	actions map[string]*CorpAction
}

func NewCorpActionsService() *CorpActionsService {
	return &CorpActionsService{actions: make(map[string]*CorpAction)}
}

func (s *CorpActionsService) Announce(ca *CorpAction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.actions[ca.ID]; exists {
		return fmt.Errorf("action %s already exists", ca.ID)
	}
	ca.Phase = Announced
	s.actions[ca.ID] = ca
	return nil
}

func (s *CorpActionsService) AdvancePhase(id string, newPhase ActionPhase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ca, ok := s.actions[id]
	if !ok { return fmt.Errorf("action not found") }
	ca.Phase = newPhase
	return nil
}

func (s *CorpActionsService) ComputeSplitAdjustment(id string, qty float64, price float64) (newQty, newPrice float64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ca, ok := s.actions[id]
	if !ok { return 0, 0, fmt.Errorf("action not found") }
	if ca.Type != "SPLIT" || ca.Ratio <= 0 { return 0, 0, fmt.Errorf("not a split action") }
	return qty * ca.Ratio, price / ca.Ratio, nil
}

func (s *CorpActionsService) GetAction(id string) (*CorpAction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ca, ok := s.actions[id]
	if !ok { return nil, fmt.Errorf("not found") }
	return ca, nil
}
