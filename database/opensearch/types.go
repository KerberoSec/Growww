package opensearch

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidDocument       = errors.New("invalid document payload")
	ErrBufferFull            = errors.New("bulk ingestion buffer full")
	ErrIndexNotFound         = errors.New("specified opensearch index not found")
	ErrTamperedAuditTrail    = errors.New("audit trail cryptographic checksum verification failed")
	ErrLifecycleReconstitute = errors.New("failed to reconstitute trade lifecycle")
)

type AlertType string

const (
	AlertWashTrading     AlertType = "WASH_TRADING"
	AlertSpoofing        AlertType = "SPOOFING"
	AlertLayering        AlertType = "LAYERING"
	AlertCircularTrading AlertType = "CIRCULAR_TRADING"
)

type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "LOW"
	SeverityMedium   AlertSeverity = "MEDIUM"
	SeverityHigh     AlertSeverity = "HIGH"
	SeverityCritical AlertSeverity = "CRITICAL"
)

// TradeDocument matches 001_trades_mapping.json.
type TradeDocument struct {
	TradeID          string    `json:"trade_id"`
	ISIN             string    `json:"isin"`
	Symbol           string    `json:"symbol"`
	BuyerOrderID     string    `json:"buyer_order_id"`
	SellerOrderID    string    `json:"seller_order_id"`
	BuyerAddress     string    `json:"buyer_address"`
	SellerAddress    string    `json:"seller_address"`
	FractionalUnits  float64   `json:"fractional_units"`
	PricePerUnitINR  float64   `json:"price_per_unit_inr"`
	GrossAmountINR   float64   `json:"gross_amount_inr"`
	FeeAmountINR     float64   `json:"fee_amount_inr"`
	SettlementStatus string    `json:"settlement_status"`
	OnChainTxHash    string    `json:"on_chain_tx_hash"`
	OnChainBlockNum  uint64    `json:"on_chain_block_num"`
	ExecutedAt       time.Time `json:"executed_at"`
}

// AuditTrailDocument matches 002_audit_trail_mapping.json.
type AuditTrailDocument struct {
	EventID       string                 `json:"event_id"`
	Action        string                 `json:"action"`
	ActorID       string                 `json:"actor_id"`
	ActorRole     string                 `json:"actor_role"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	CorrelationID string                 `json:"correlation_id"`
	ClientIP      string                 `json:"client_ip"`
	BeforeState   map[string]interface{} `json:"before_state,omitempty"`
	AfterState    map[string]interface{} `json:"after_state,omitempty"`
	TamperHash    string                 `json:"tamper_hash"`
	Timestamp     time.Time              `json:"timestamp"`
}

// CalculateChecksum computes SHA-256 tamper hash.
func (d *AuditTrailDocument) CalculateChecksum() string {
	raw, _ := json.Marshal(struct {
		EventID       string
		Action        string
		ActorID       string
		EntityID      string
		CorrelationID string
		Timestamp     int64
	}{
		EventID:       d.EventID,
		Action:        d.Action,
		ActorID:       d.ActorID,
		EntityID:      d.EntityID,
		CorrelationID: d.CorrelationID,
		Timestamp:     d.Timestamp.UnixNano(),
	})
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}

// SurveillanceAlert matches 003_surveillance_alerts_mapping.json.
type SurveillanceAlert struct {
	AlertID         string                 `json:"alert_id"`
	AlertType       AlertType              `json:"alert_type"`
	Severity        AlertSeverity          `json:"severity"`
	ISIN            string                 `json:"isin"`
	Symbol          string                 `json:"symbol"`
	Participants    []string               `json:"participants"`
	ConfidenceScore float64                `json:"confidence_score"`
	Description     string                 `json:"description"`
	Evidence        map[string]interface{} `json:"evidence"`
	Status          string                 `json:"status"` // OPEN, REVIEWED, ESCALATED_SEBI
	DetectedAt      time.Time              `json:"detected_at"`
}

// ReconstitutedLifecycle represents the chronological lifecycle of a matched order and trade.
type ReconstitutedLifecycle struct {
	TradeID        string                `json:"trade_id"`
	BuyerOrderID   string                `json:"buyer_order_id"`
	SellerOrderID  string                `json:"seller_order_id"`
	AuditEvents    []*AuditTrailDocument `json:"audit_events"`
	Trade          *TradeDocument        `json:"trade"`
	IntegrityValid bool                  `json:"integrity_valid"`
}

// Candlestick represents an aggregated OHLCV bar.
type Candlestick struct {
	ISIN        string    `json:"isin"`
	WindowStart time.Time `json:"window_start"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Close       float64   `json:"close"`
	Volume      float64   `json:"volume"`
	TradeCount  int       `json:"trade_count"`
}
