package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ConflatedBookUpdate struct {
	Symbol      string      `json:"s"`
	Sequence    uint64      `json:"u"`
	Bids        [][2]string `json:"b"` // [price, size]
	Asks        [][2]string `json:"a"`
	TimestampMs int64       `json:"E"`
}

type WSSClientSession struct {
	SessionID string
	SendChan  chan []byte
	Active    bool
}

type WebSocketStreamingGateway struct {
	mu          sync.RWMutex
	clients     map[string]*WSSClientSession
	lastSeq     uint64
	conflateMS  time.Duration
}

func NewWebSocketStreamingGateway(conflationMs time.Duration) *WebSocketStreamingGateway {
	if conflationMs == 0 {
		conflationMs = 50 * time.Millisecond // 50ms book conflation window
	}
	return &WebSocketStreamingGateway{
		clients:    make(map[string]*WSSClientSession),
		conflateMS: conflationMs,
	}
}

func (g *WebSocketStreamingGateway) RegisterClient(sessionID string) *WSSClientSession {
	g.mu.Lock()
	defer g.mu.Unlock()

	client := &WSSClientSession{
		SessionID: sessionID,
		SendChan:  make(chan []byte, 512),
		Active:    true,
	}
	g.clients[sessionID] = client
	return client
}

func (g *WebSocketStreamingGateway) BroadcastConflatedUpdate(symbol string, bids, asks [][2]string) {
	g.mu.Lock()
	g.lastSeq++
	update := ConflatedBookUpdate{
		Symbol:      symbol,
		Sequence:    g.lastSeq,
		Bids:        bids,
		Asks:        asks,
		TimestampMs: time.Now().UnixMilli(),
	}
	g.mu.Unlock()

	payload, err := json.Marshal(update)
	if err != nil {
		return
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, client := range g.clients {
		if !client.Active {
			continue
		}
		select {
		case client.SendChan <- payload:
		default:
			// Drop old tick under backpressure to maintain low latency
		}
	}
}

func (g *WebSocketStreamingGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "WebSocket Conflated Streaming Gateway running. Active clients: %d\n", len(g.clients))
}
