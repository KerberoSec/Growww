package main

import (
	"fmt"
	"sync"
	"time"
)

type Report struct {
	ID          string
	Type        string
	Period      string
	Target      string
	GeneratedAt time.Time
	Status      string
	Payload     string
}

type ReportingService struct {
	mu      sync.Mutex
	reports []*Report
}

func NewReportingService() *ReportingService { return &ReportingService{} }

func (s *ReportingService) GenerateReport(reportType, period, target string) (*Report, error) {
	if reportType == "" || target == "" {
		return nil, fmt.Errorf("type and target required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r := &Report{
		ID:          fmt.Sprintf("RPT-%d", len(s.reports)+1),
		Type:        reportType,
		Period:      period,
		Target:      target,
		GeneratedAt: time.Now().UTC(),
		Status:      "GENERATED",
		Payload:     fmt.Sprintf(`{"type":"%s","period":"%s"}`, reportType, period),
	}
	s.reports = append(s.reports, r)
	return r, nil
}

func (s *ReportingService) SubmitReport(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.reports {
		if r.ID == id {
			r.Status = "SUBMITTED"
			return nil
		}
	}
	return fmt.Errorf("report not found")
}

func (s *ReportingService) GetReports() []*Report {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reports
}
