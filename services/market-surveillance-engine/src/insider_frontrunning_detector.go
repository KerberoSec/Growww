package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// SEBITypology represents statutory market abuse classifications
type SEBITypology string

const (
	TypologyInsiderTradingUPSI SEBITypology = "SEBI_PIT_INSIDER_TRADING_UPSI"
	TypologyFrontrunningPFUTP  SEBITypology = "SEBI_PFUTP_FRONTRUNNING_PARENT_ORDER"
)

// DesignatedPerson represents an insider under SEBI (PIT) Regulations
type DesignatedPerson struct {
	PersonID            string    `json:"person_id"`
	Name                string    `json:"name"`
	PAN                 string    `json:"pan"`
	CompanySymbol       string    `json:"company_symbol"`
	Role                string    `json:"role"` // DIRECTOR, KMP, AUDITOR, PROMOTER, CONNECTED_PERSON
	TradingWindowClosed bool      `json:"trading_window_closed"`
	ConnectedAccounts   []string  `json:"connected_accounts"` // Family, associates
}

// UPSIEvent represents Unpublished Price Sensitive Information under SEBI PIT
type UPSIEvent struct {
	EventID                  string    `json:"event_id"`
	Symbol                   string    `json:"symbol"`
	EventType                string    `json:"event_type"` // EARNINGS, MERGER_ACQUISITION, DIVIDEND, REGULATORY_APPROVAL
	AnnouncementTimestamp    time.Time `json:"announcement_timestamp"`
	ExpectedImpact           string    `json:"expected_impact"` // BULLISH, BEARISH
	PostPriceChangePercent   float64   `json:"post_price_change_percent"`
	PreAnnouncementPriceE8   uint64    `json:"pre_announcement_price_e8"`
	PostAnnouncementPriceE8  uint64    `json:"post_announcement_price_e8"`
}

// TradeExecution represents matched order executions
type TradeExecution struct {
	TradeID   string    `json:"trade_id"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"` // BUY, SELL
	PriceE8   uint64    `json:"price_e8"`
	QtyE8     uint64    `json:"qty_e8"`
	Timestamp time.Time `json:"timestamp"`
}

// ParentOrder represents large institutional block orders subject to frontrunning surveillance
type ParentOrder struct {
	ParentOrderID     string    `json:"parent_order_id"`
	ClientUserID      string    `json:"client_user_id"`
	Symbol            string    `json:"symbol"`
	Side              string    `json:"side"` // BUY, SELL
	PriceE8           uint64    `json:"price_e8"`
	QtyE8             uint64    `json:"qty_e8"`
	NotionalINR       float64   `json:"notional_inr"`
	ArrivalTimestamp  time.Time `json:"arrival_timestamp"`
	ExecutedTimestamp time.Time `json:"executed_timestamp"`
}

// SEBIInsiderTradingAlert represents a formal insider trading surveillance alert
type SEBIInsiderTradingAlert struct {
	AlertID                 string       `json:"alert_id"`
	Typology                SEBITypology `json:"typology"`
	Symbol                  string       `json:"symbol"`
	AccusedUserID           string       `json:"accused_user_id"`
	AccusedName             string       `json:"accused_name"`
	InsiderRole             string       `json:"insider_role"`
	UPSIEventID             string       `json:"upsi_event_id"`
	PreAnnouncementTrades   int          `json:"pre_announcement_trades"`
	TotalVolumeE8           uint64       `json:"total_volume_e8"`
	VolumeZScore            float64      `json:"volume_z_score"`
	DirectionalAlignment    bool         `json:"directional_alignment"`
	ConfidenceRate          float64      `json:"confidence_rate"`
	EstimatedIllicitGainINR float64      `json:"estimated_illicit_gain_inr"`
	LegalCitation           string       `json:"legal_citation"`
	TriggeredAt             time.Time    `json:"triggered_at"`
}

// SEBIFrontrunningAlert represents a frontrunning / ahead-of-client trading alert
type SEBIFrontrunningAlert struct {
	AlertID                 string       `json:"alert_id"`
	Typology                SEBITypology `json:"typology"`
	Symbol                  string       `json:"symbol"`
	SuspectUserID           string       `json:"suspect_user_id"`
	ParentOrderID           string       `json:"parent_order_id"`
	VictimClientUserID      string       `json:"victim_client_user_id"`
	LeadTimeMs              int64        `json:"lead_time_ms"`
	FrontrunnerQtyE8        uint64       `json:"frontrunner_qty_e8"`
	FrontrunnerEntryPriceE8 uint64       `json:"frontrunner_entry_price_e8"`
	FrontrunnerExitPriceE8  uint64       `json:"frontrunner_exit_price_e8"`
	EstimatedIllicitGainINR float64      `json:"estimated_illicit_gain_inr"`
	ConfidenceRate          float64      `json:"confidence_rate"`
	LegalCitation           string       `json:"legal_citation"`
	TriggeredAt             time.Time    `json:"triggered_at"`
}

// SEBIDossier represents the formal regulatory filing payload for SEBI IMSS
type SEBIDossier struct {
	DossierID               string       `json:"dossier_id"`
	CaseTitle               string       `json:"case_title"`
	Typology                SEBITypology `json:"typology"`
	AccusedEntityID         string       `json:"accused_entity_id"`
	ConfidenceScore         float64      `json:"confidence_score"`
	EstimatedIllicitGainINR float64      `json:"estimated_illicit_gain_inr"`
	LegalCitations          []string     `json:"legal_citations"`
	EvidenceTimeline        []string     `json:"evidence_timeline"`
	RecommendedAction       string       `json:"recommended_action"`
	GeneratedAt             time.Time    `json:"generated_at"`
	TamperProofHash         string       `json:"tamper_proof_hash"`
}

// SEBIMarketSurveillanceEngine implements institutional SEBI PIT & PFUTP detection
type SEBIMarketSurveillanceEngine struct {
	mu                sync.RWMutex
	sddRegistry       map[string]map[string]*DesignatedPerson // Symbol -> PersonID/UserID -> DesignatedPerson
	upsiEvents        map[string]*UPSIEvent                   // EventID -> UPSIEvent
	insiderAlerts     []*SEBIInsiderTradingAlert
	frontrunAlerts    []*SEBIFrontrunningAlert
	frozenAccounts    map[string]string                       // UserID -> FreezeReason
	dossiers          map[string]*SEBIDossier
}

func NewSEBIMarketSurveillanceEngine() *SEBIMarketSurveillanceEngine {
	return &SEBIMarketSurveillanceEngine{
		sddRegistry:    make(map[string]map[string]*DesignatedPerson),
		upsiEvents:     make(map[string]*UPSIEvent),
		insiderAlerts:  make([]*SEBIInsiderTradingAlert, 0),
		frontrunAlerts: make([]*SEBIFrontrunningAlert, 0),
		frozenAccounts: make(map[string]string),
		dossiers:       make(map[string]*SEBIDossier),
	}
}

// RegisterDesignatedPerson records an insider or connected person into the Structured Digital Database (SDD)
func (e *SEBIMarketSurveillanceEngine) RegisterDesignatedPerson(dp DesignatedPerson) {
	e.mu.Lock()
	defer e.mu.Unlock()

	sym := strings.ToUpper(dp.CompanySymbol)
	if e.sddRegistry[sym] == nil {
		e.sddRegistry[sym] = make(map[string]*DesignatedPerson)
	}
	e.sddRegistry[sym][dp.PersonID] = &dp
	for _, connID := range dp.ConnectedAccounts {
		e.sddRegistry[sym][connID] = &dp
	}
}

// RegisterUPSIEvent logs an unpublished price sensitive announcement event
func (e *SEBIMarketSurveillanceEngine) RegisterUPSIEvent(upsi UPSIEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.upsiEvents[upsi.EventID] = &upsi
}

// UserBaselineStats holds historical trading volume statistics for Z-score calculation
type UserBaselineStats struct {
	MeanDailyVolumeE8 float64
	StdDevVolumeE8    float64
}

// DetectInsiderTrading identifies trading by connected persons ahead of material UPSI disclosures
func (e *SEBIMarketSurveillanceEngine) DetectInsiderTrading(
	upsi UPSIEvent,
	trades []TradeExecution,
	baselines map[string]UserBaselineStats,
	lookbackHours int,
) []*SEBIInsiderTradingAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	sym := strings.ToUpper(upsi.Symbol)
	sddMap := e.sddRegistry[sym]
	lookbackDuration := time.Duration(lookbackHours) * time.Hour
	windowStart := upsi.AnnouncementTimestamp.Add(-lookbackDuration)

	// Group trades by UserID in the pre-announcement window
	userTrades := make(map[string][]TradeExecution)
	for _, tr := range trades {
		if strings.ToUpper(tr.Symbol) != sym {
			continue
		}
		if tr.Timestamp.After(windowStart) && tr.Timestamp.Before(upsi.AnnouncementTimestamp) {
			userTrades[tr.UserID] = append(userTrades[tr.UserID], tr)
		}
	}

	alerts := make([]*SEBIInsiderTradingAlert, 0)

	for userID, txList := range userTrades {
		dp, isInsider := sddMap[userID]

		var totalVolE8 uint64
		buyVolE8 := uint64(0)
		sellVolE8 := uint64(0)
		var totalCostINR float64

		for _, tr := range txList {
			totalVolE8 += tr.QtyE8
			priceINR := float64(tr.PriceE8) / 1e8
			qty := float64(tr.QtyE8) / 1e8
			totalCostINR += priceINR * qty

			if strings.EqualFold(tr.Side, "BUY") {
				buyVolE8 += tr.QtyE8
			} else {
				sellVolE8 += tr.QtyE8
			}
		}

		// Calculate statistical volume Z-Score
		baseline, hasBaseline := baselines[userID]
		zScore := 0.0
		if hasBaseline && baseline.StdDevVolumeE8 > 0 {
			zScore = (float64(totalVolE8) - baseline.MeanDailyVolumeE8) / baseline.StdDevVolumeE8
		} else if totalVolE8 > 0 {
			zScore = 3.0 // Arbitrary high default if anomalous without baseline
		}

		// Directional alignment check
		// Bullish news: BUYing is suspect
		// Bearish news: SELLing / Shorting is suspect
		directionalMatch := false
		if strings.EqualFold(upsi.ExpectedImpact, "BULLISH") && buyVolE8 > sellVolE8 {
			directionalMatch = true
		} else if strings.EqualFold(upsi.ExpectedImpact, "BEARISH") && sellVolE8 > buyVolE8 {
			directionalMatch = true
		}

		// Scoring Heuristic:
		// 1. Insider/Connected Person: +0.40
		// 2. High Z-Score (>= 2.5): +0.30
		// 3. Directional Alignment: +0.25
		confidence := 0.0
		if isInsider {
			confidence += 0.40
		}
		if zScore >= 2.5 {
			confidence += 0.35
		} else if zScore >= 1.5 {
			confidence += 0.20
		}
		if directionalMatch {
			confidence += 0.25
		}

		// Flag alert if (isInsider and directionalMatch) OR (confidence >= 0.75)
		if (isInsider && directionalMatch) || confidence >= 0.75 {
			// Calculate estimated illicit profit
			illicitGain := 0.0
			qty := float64(totalVolE8) / 1e8
			prePrice := float64(upsi.PreAnnouncementPriceE8) / 1e8
			postPrice := float64(upsi.PostAnnouncementPriceE8) / 1e8

			if strings.EqualFold(upsi.ExpectedImpact, "BULLISH") && postPrice > prePrice {
				illicitGain = (postPrice - prePrice) * qty
			} else if strings.EqualFold(upsi.ExpectedImpact, "BEARISH") && prePrice > postPrice {
				illicitGain = (prePrice - postPrice) * qty
			}

			insiderRole := "EXTERNAL_ANOMALOUS_TRADER"
			insiderName := "UNREGISTERED_ENTITY"
			if isInsider {
				insiderRole = dp.Role
				insiderName = dp.Name
			}

			alert := &SEBIInsiderTradingAlert{
				AlertID:                 fmt.Sprintf("SEBI-PIT-%d", time.Now().UnixNano()),
				Typology:                TypologyInsiderTradingUPSI,
				Symbol:                  sym,
				AccusedUserID:           userID,
				AccusedName:             insiderName,
				InsiderRole:             insiderRole,
				UPSIEventID:             upsi.EventID,
				PreAnnouncementTrades:   len(txList),
				TotalVolumeE8:           totalVolE8,
				VolumeZScore:            math.Round(zScore*100) / 100,
				DirectionalAlignment:    directionalMatch,
				ConfidenceRate:          math.Min(1.0, math.Round(confidence*100)/100),
				EstimatedIllicitGainINR: math.Round(illicitGain*100) / 100,
				LegalCitation:           "SEBI (Prohibition of Insider Trading) Regulations, 2015 - Regulation 3(1) and 4(1)",
				TriggeredAt:             time.Now().UTC(),
			}

			alerts = append(alerts, alert)
			e.insiderAlerts = append(e.insiderAlerts, alert)
		}
	}

	return alerts
}

// DetectFrontrunning identifies ahead-of-client trading before institutional block orders
func (e *SEBIMarketSurveillanceEngine) DetectFrontrunning(
	parentOrders []ParentOrder,
	marketOrders []OrderEvent,
	trades []TradeExecution,
	maxLeadWindowMs int64,
) []*SEBIFrontrunningAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	alerts := make([]*SEBIFrontrunningAlert, 0)

	for _, parent := range parentOrders {
		sym := strings.ToUpper(parent.Symbol)
		parentTime := parent.ArrivalTimestamp
		parentSide := strings.ToUpper(parent.Side)

		// Look for suspect orders placed strictly before parent order arrival within maxLeadWindowMs
		for _, ord := range marketOrders {
			if strings.ToUpper(ord.Symbol) != sym {
				continue
			}
			if ord.UserID == parent.ClientUserID {
				continue // Parent order issuer cannot front-run themselves
			}
			if ord.IsCancel {
				continue
			}
			if strings.ToUpper(ord.Side) != parentSide {
				continue // Must be on the same side (BUY ahead of BUY)
			}

			leadTime := parentTime.Sub(ord.Timestamp)
			leadMs := leadTime.Milliseconds()

			// Must be placed strictly before parent order and within lead window (e.g. 10ms to 5000ms)
			if leadMs > 0 && leadMs <= maxLeadWindowMs {
				// Now check if this suspect order was executed and subsequently exited (reversal) after parent executed
				var entryPriceE8 uint64
				var exitPriceE8 uint64
				var tradeQtyE8 uint64
				hasEntry := false
				hasExit := false

				// Opposite side for exit
				exitSide := "SELL"
				if parentSide == "SELL" {
					exitSide = "BUY"
				}

				for _, tr := range trades {
					if tr.UserID == ord.UserID && strings.ToUpper(tr.Symbol) == sym {
						// Entry trade
						if strings.ToUpper(tr.Side) == parentSide && tr.Timestamp.Before(parent.ExecutedTimestamp) {
							hasEntry = true
							entryPriceE8 = tr.PriceE8
							tradeQtyE8 = tr.QtyE8
						}
						// Exit trade (within 60 seconds after parent execution)
						if strings.ToUpper(tr.Side) == exitSide && tr.Timestamp.After(parent.ExecutedTimestamp) &&
							tr.Timestamp.Sub(parent.ExecutedTimestamp) <= 60*time.Second {
							hasExit = true
							exitPriceE8 = tr.PriceE8
						}
					}
				}

				if hasEntry {
					// Compute estimated illicit gain
					qty := float64(tradeQtyE8) / 1e8
					entryP := float64(entryPriceE8) / 1e8
					exitP := float64(exitPriceE8) / 1e8

					illicitGain := 0.0
					if hasExit {
						if parentSide == "BUY" && exitP > entryP {
							illicitGain = (exitP - entryP) * qty
						} else if parentSide == "SELL" && entryP > exitP {
							illicitGain = (entryP - exitP) * qty
						}
					} else {
						// Theoretical gain based on parent execution price impact
						parentP := float64(parent.PriceE8) / 1e8
						if parentSide == "BUY" && parentP > entryP {
							illicitGain = (parentP - entryP) * qty
						}
					}

					confidence := 0.90
					if hasExit {
						confidence = 0.97
					}

					alert := &SEBIFrontrunningAlert{
						AlertID:                 fmt.Sprintf("SEBI-PFUTP-%d", time.Now().UnixNano()),
						Typology:                TypologyFrontrunningPFUTP,
						Symbol:                  sym,
						SuspectUserID:           ord.UserID,
						ParentOrderID:           parent.ParentOrderID,
						VictimClientUserID:      parent.ClientUserID,
						LeadTimeMs:              leadMs,
						FrontrunnerQtyE8:        tradeQtyE8,
						FrontrunnerEntryPriceE8: entryPriceE8,
						FrontrunnerExitPriceE8:  exitPriceE8,
						EstimatedIllicitGainINR: math.Round(illicitGain*100) / 100,
						ConfidenceRate:          confidence,
						LegalCitation:           "SEBI (PFUTP) Regulations, 2003 - Regulation 3 and 4(2)(q) Front-Running",
						TriggeredAt:             time.Now().UTC(),
					}

					alerts = append(alerts, alert)
					e.frontrunAlerts = append(e.frontrunAlerts, alert)
				}
			}
		}
	}

	return alerts
}

// GenerateSEBIDossier compiles statutory regulatory evidence for SEBI IMSS submission
func (e *SEBIMarketSurveillanceEngine) GenerateSEBIDossier(alertID string) (*SEBIDossier, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check insider alerts
	for _, ia := range e.insiderAlerts {
		if ia.AlertID == alertID {
			dossierID := fmt.Sprintf("DOSSIER-PIT-%s", ia.AlertID)
			timeline := []string{
				fmt.Sprintf("T-0: UPSI Registered ID %s on symbol %s", ia.UPSIEventID, ia.Symbol),
				fmt.Sprintf("T-Pre: Accused %s (%s) executed %d trades for %d units with Volume Z-score %.2f", ia.AccusedUserID, ia.AccusedName, ia.PreAnnouncementTrades, ia.TotalVolumeE8, ia.VolumeZScore),
				fmt.Sprintf("T-Post: Directional bet confirmed; estimated illicit profit ₹%.2f", ia.EstimatedIllicitGainINR),
			}

			payload := fmt.Sprintf("%s:%s:%s:%.2f", dossierID, ia.AccusedUserID, ia.Typology, ia.ConfidenceRate)
			hash := sha256.Sum256([]byte(payload))

			d := &SEBIDossier{
				DossierID:               dossierID,
				CaseTitle:               fmt.Sprintf("Investigation into Insider Trading in %s by %s", ia.Symbol, ia.AccusedName),
				Typology:                ia.Typology,
				AccusedEntityID:         ia.AccusedUserID,
				ConfidenceScore:         ia.ConfidenceRate,
				EstimatedIllicitGainINR: ia.EstimatedIllicitGainINR,
				LegalCitations: []string{
					"SEBI (Prohibition of Insider Trading) Regulations, 2015 - Regulation 3",
					"SEBI (Prohibition of Insider Trading) Regulations, 2015 - Regulation 4",
				},
				EvidenceTimeline:  timeline,
				RecommendedAction: "IMMEDIATE_ACCOUNT_FREEZE_AND_SEBI_DISPATCH",
				GeneratedAt:       time.Now().UTC(),
				TamperProofHash:   hex.EncodeToString(hash[:]),
			}
			e.dossiers[dossierID] = d
			return d, nil
		}
	}

	// Check frontrunning alerts
	for _, fa := range e.frontrunAlerts {
		if fa.AlertID == alertID {
			dossierID := fmt.Sprintf("DOSSIER-PFUTP-%s", fa.AlertID)
			timeline := []string{
				fmt.Sprintf("T-Lead: Suspect %s placed order %d ms ahead of Parent Order %s", fa.SuspectUserID, fa.LeadTimeMs, fa.ParentOrderID),
				fmt.Sprintf("T-Entry: Executed entry fill at price %d for qty %d", fa.FrontrunnerEntryPriceE8, fa.FrontrunnerQtyE8),
				fmt.Sprintf("T-Exit: Unwound position at %d; extracted profit ₹%.2f", fa.FrontrunnerExitPriceE8, fa.EstimatedIllicitGainINR),
			}

			payload := fmt.Sprintf("%s:%s:%s:%.2f", dossierID, fa.SuspectUserID, fa.Typology, fa.ConfidenceRate)
			hash := sha256.Sum256([]byte(payload))

			d := &SEBIDossier{
				DossierID:               dossierID,
				CaseTitle:               fmt.Sprintf("Investigation into Front-Running of Institutional Client %s in %s", fa.VictimClientUserID, fa.Symbol),
				Typology:                fa.Typology,
				AccusedEntityID:         fa.SuspectUserID,
				ConfidenceScore:         fa.ConfidenceRate,
				EstimatedIllicitGainINR: fa.EstimatedIllicitGainINR,
				LegalCitations: []string{
					"SEBI (PFUTP) Regulations, 2003 - Regulation 3",
					"SEBI (PFUTP) Regulations, 2003 - Regulation 4(2)(q)",
				},
				EvidenceTimeline:  timeline,
				RecommendedAction: "QUOTE_PURGE_TRADING_SUSPENSION_AND_SEBI_FILING",
				GeneratedAt:       time.Now().UTC(),
				TamperProofHash:   hex.EncodeToString(hash[:]),
			}
			e.dossiers[dossierID] = d
			return d, nil
		}
	}

	return nil, fmt.Errorf("alert %s not found", alertID)
}

// TriggerProtectiveFreeze enforces immediate trading suspension on accounts flagged with critical violations
func (e *SEBIMarketSurveillanceEngine) TriggerProtectiveFreeze(userID, reason string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.frozenAccounts[userID] = reason
	return true
}

// IsAccountFrozen checks if an account has an active surveillance freeze
func (e *SEBIMarketSurveillanceEngine) IsAccountFrozen(userID string) (bool, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	reason, frozen := e.frozenAccounts[userID]
	return frozen, reason
}
