package main

import (
	"sync"
	"time"
)

// ConflationConfig defines parameters for WebSocket depth conflation
type ConflationConfig struct {
	IntervalMs    int64 // Minimum interval between broadcasts in milliseconds
	MaxDepthLevel int   // Maximum number of price levels to broadcast per side
}

// ConflatedDepthMessage is the final message sent over WebSocket
type ConflatedDepthMessage struct {
	Symbol      string       `json:"symbol"`
	Sequence    uint64       `json:"sequence"`
	Bids        [][2]uint64  `json:"bids"` // [priceE8, qtyE8] sorted best first
	Asks        [][2]uint64  `json:"asks"` // [priceE8, qtyE8] sorted best first
	Timestamp   int64        `json:"timestamp_ms"`
	Conflated   bool         `json:"conflated"` // True if intermediate snapshots were merged
}

// DepthConflationEngine manages rate-limiting and depth truncation for WebSocket broadcasting
type DepthConflationEngine struct {
	mu             sync.Mutex
	broadcaster    *DepthBroadcaster
	config         ConflationConfig
	lastBroadcast  time.Time
	lastSeqSent    uint64
	pendingMsg     *ConflatedDepthMessage
	subscribers    map[string]chan ConflatedDepthMessage
}

// NewDepthConflationEngine creates a new conflation engine wrapping the ring buffer
func NewDepthConflationEngine(broadcaster *DepthBroadcaster, cfg ConflationConfig) *DepthConflationEngine {
	if cfg.IntervalMs <= 0 {
		cfg.IntervalMs = 100 // Default 100ms conflation window
	}
	if cfg.MaxDepthLevel <= 0 {
		cfg.MaxDepthLevel = 20 // Default 20 levels
	}
	return &DepthConflationEngine{
		broadcaster: broadcaster,
		config:      cfg,
		subscribers: make(map[string]chan ConflatedDepthMessage),
	}
}

// Subscribe registers a WebSocket subscriber
func (e *DepthConflationEngine) Subscribe(subscriberID string) <-chan ConflatedDepthMessage {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := make(chan ConflatedDepthMessage, 128)
	e.subscribers[subscriberID] = ch
	return ch
}

// Unsubscribe removes a WebSocket subscriber
func (e *DepthConflationEngine) Unsubscribe(subscriberID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ch, ok := e.subscribers[subscriberID]; ok {
		close(ch)
		delete(e.subscribers, subscriberID)
	}
}

// TryBroadcast checks if it's time to send a conflated depth update.
// Returns the message if broadcast, nil if conflation window hasn't elapsed.
func (e *DepthConflationEngine) TryBroadcast() *ConflatedDepthMessage {
	e.mu.Lock()
	defer e.mu.Unlock()

	snapshot, ok := e.broadcaster.GetLatest()
	if !ok || snapshot.Sequence <= e.lastSeqSent {
		return nil
	}

	now := time.Now()
	elapsed := now.Sub(e.lastBroadcast)
	if elapsed.Milliseconds() < e.config.IntervalMs {
		// Within conflation window, store as pending
		e.pendingMsg = e.buildConflatedMessage(snapshot, true)
		return nil
	}

	// Broadcast now
	msg := e.buildConflatedMessage(snapshot, e.pendingMsg != nil)
	e.lastBroadcast = now
	e.lastSeqSent = snapshot.Sequence
	e.pendingMsg = nil

	// Fan out to subscribers
	for id, ch := range e.subscribers {
		select {
		case ch <- *msg:
		default:
			// Slow consumer, skip
			_ = id
		}
	}

	return msg
}

// truncateLevels limits depth to MaxDepthLevel
func (e *DepthConflationEngine) truncateLevels(levels [][2]uint64) [][2]uint64 {
	if len(levels) > e.config.MaxDepthLevel {
		return levels[:e.config.MaxDepthLevel]
	}
	return levels
}

func (e *DepthConflationEngine) buildConflatedMessage(snap RingBufferDepthSnapshot, conflated bool) *ConflatedDepthMessage {
	return &ConflatedDepthMessage{
		Symbol:    "BTC/USDT",
		Sequence:  snap.Sequence,
		Bids:      e.truncateLevels(snap.Bids),
		Asks:      e.truncateLevels(snap.Asks),
		Timestamp: snap.Timestamp.UnixMilli(),
		Conflated: conflated,
	}
}

// SubscriberCount returns current active subscriber count
func (e *DepthConflationEngine) SubscriberCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.subscribers)
}
