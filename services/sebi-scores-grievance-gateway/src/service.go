package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// GrievanceService is the unified facade combining gateway, ATR generator, and escalation timer
type GrievanceService struct {
	Gateway         *SCORESGatewayConnector
	ATRGenerator    *ATRGenerator
	EscalationTimer *EscalationTimer
	UserStore       UserMappingStore
}

// NewGrievanceService initializes the unified grievance service
func NewGrievanceService(cfg ScoresGatewayConfig) *GrievanceService {
	userStore := NewMemoryUserMappingStore()
	gw := NewSCORESGatewayConnector(cfg, userStore)
	atrGen := NewATRGenerator(gw)
	timer := NewEscalationTimer(gw)

	return &GrievanceService{
		Gateway:         gw,
		ATRGenerator:    atrGen,
		EscalationTimer: timer,
		UserStore:       userStore,
	}
}

// RegisterHTTPHandlers mounts REST endpoints on the provided http.ServeMux
func (s *GrievanceService) RegisterHTTPHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/grievance/scores/webhook", s.handleWebhook)
	mux.HandleFunc("/api/v1/grievance/dockets", s.handleListDockets)
	mux.HandleFunc("/api/v1/grievance/dockets/", s.handleGetDocket)
	mux.HandleFunc("/api/v1/grievance/atr/draft", s.handleDraftATR)
	mux.HandleFunc("/api/v1/grievance/atr/approve", s.handleApproveATR)
	mux.HandleFunc("/api/v1/grievance/atr/sign", s.handleSignATR)
	mux.HandleFunc("/api/v1/grievance/atr/submit", s.handleSubmitATR)
	mux.HandleFunc("/api/v1/grievance/sla/dashboard", s.handleSLADashboard)
}

func (s *GrievanceService) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload ScoresWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Scores-Signature-256")
	docket, err := s.Gateway.IngestSCORESComplaint(payload, sig)
	if err != nil {
		if err == ErrDuplicateDocket {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(docket)
}

func (s *GrievanceService) handleListDockets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dockets := s.Gateway.ListDockets()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dockets)
}

func (s *GrievanceService) handleGetDocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/grievance/dockets/")
	docket, err := s.Gateway.GetDocket(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(docket)
}

func (s *GrievanceService) handleDraftATR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DraftATRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	atr, err := s.ATRGenerator.DraftATR(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(atr)
}

type ApproveATRRequest struct {
	ATRID         string `json:"atr_id"`
	CheckerUserID string `json:"checker_user_id"`
	Notes         string `json:"notes"`
}

func (s *GrievanceService) handleApproveATR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ApproveATRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	atr, err := s.ATRGenerator.ApproveATR(req.ATRID, req.CheckerUserID, req.Notes)
	if err != nil {
		if err == ErrDualControlViolation {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(atr)
}

type SignATRRequest struct {
	ATRID    string `json:"atr_id"`
	SignerDN string `json:"signer_dn"`
}

func (s *GrievanceService) handleSignATR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignATRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	atr, err := s.ATRGenerator.SignATRWithDSC(req.ATRID, req.SignerDN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(atr)
}

type SubmitATRRequest struct {
	ATRID string `json:"atr_id"`
}

func (s *GrievanceService) handleSubmitATR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitATRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	atr, err := s.ATRGenerator.GetATR(req.ATRID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	res, err := s.Gateway.SubmitATRToRegulator(atr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *GrievanceService) handleSLADashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dash := s.EscalationTimer.GetSLADashboard(time.Now().UTC())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dash)
}

func main() {
	cfg := ScoresGatewayConfig{
		ScoresAPIEndpoint: "https://scores.sebi.gov.in/api/v2",
		EntitySEBICode:    "INZ000301032",
		WebhookSecret:     "SCORES_2026_HMAC_SECRET_PROD",
		UseMTLS:           true,
	}
	svc := NewGrievanceService(cfg)
	mux := http.NewServeMux()
	svc.RegisterHTTPHandlers(mux)
	_ = http.ListenAndServe(":8080", mux)
}
