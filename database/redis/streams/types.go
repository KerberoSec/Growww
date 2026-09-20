package streams

import (
	"errors"
	"time"
)

var (
	ErrConsumerGroupExists  = errors.New("consumer group already exists")
	ErrConsumerGroupNotFound= errors.New("consumer group not found")
	ErrStreamNotFound       = errors.New("redis stream not found")
	ErrMessageNotFound      = errors.New("stream message ID not found")
	ErrDeadLetterLimit      = errors.New("message exceeded max processing attempts, routed to DLQ")
)

// MarketTick represents an incoming sub-millisecond price/trade event.
type MarketTick struct {
	ISIN       string    `json:"isin"`
	Symbol     string    `json:"symbol"`
	Price      float64   `json:"price"`
	Quantity   float64   `json:"quantity"`
	SequenceNo uint64    `json:"sequence_no"`
	Timestamp  time.Time `json:"timestamp"`
}

// ConflatedTicker represents a collapsed ticker snapshot emitted at interval boundaries.
type ConflatedTicker struct {
	ISIN         string    `json:"isin"`
	Symbol       string    `json:"symbol"`
	SequenceFrom uint64    `json:"sequence_from"`
	SequenceTo   uint64    `json:"sequence_to"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Close        float64   `json:"close"`
	Volume       float64   `json:"volume"`
	Turnover     float64   `json:"turnover"`
	VWAP         float64   `json:"vwap"` // Volume-Weighted Average Price
	TickCount    int       `json:"tick_count"`
	EmittedAt    time.Time `json:"emitted_at"`
}

// StreamMessage represents a message stored in a Redis Stream.
type StreamMessage struct {
	ID        string                 `json:"id"` // e.g. "1726830000000-0"
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// PendingMessage tracks unacknowledged messages in the Pending Entries List (PEL).
type PendingMessage struct {
	ID              string
	Consumer        string
	DeliveredAt     time.Time
	DeliveryAttempts int
}

// Prompt 419 Protobuf Service Contract Representations
type RedisstreamsconflatedmarketfeedRequest struct {
	RequestID   string            `json:"request_id"`
	EntityID    string            `json:"entity_id"`
	AmountE8    uint64            `json:"amount_e8"`
	TimestampMs uint64            `json:"timestamp_ms"`
	Metadata    map[string]string `json:"metadata"`
}

type RedisstreamsconflatedmarketfeedResponse struct {
	RequestID       string `json:"request_id"`
	Success         bool   `json:"success"`
	TransactionHash string `json:"transaction_hash"`
	BlockNumber     uint64 `json:"block_number"`
	ErrorMessage    string `json:"error_message,omitempty"`
}
