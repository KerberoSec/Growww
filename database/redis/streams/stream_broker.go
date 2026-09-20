package streams

import (
	"fmt"
	"sync"
	"time"
)

// StreamBroker simulates the Redis Streams data structure and commands (XADD, XLEN, XRANGE).
type StreamBroker struct {
	mu      sync.RWMutex
	streams map[string][]*StreamMessage
	lastSeq map[string]int64 // stream -> sequence counter for current ms
}

func NewStreamBroker() *StreamBroker {
	return &StreamBroker{
		streams: make(map[string][]*StreamMessage),
		lastSeq: make(map[string]int64),
	}
}

// XAdd appends a new message to the stream with automatic ID generation and MAXLEN trimming.
func (b *StreamBroker) XAdd(stream string, maxLen int64, payload map[string]interface{}) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now().UTC()
	ms := now.UnixMilli()

	seq := b.lastSeq[stream] + 1
	b.lastSeq[stream] = seq

	msgID := fmt.Sprintf("%d-%d", ms, seq)
	msg := &StreamMessage{
		ID:        msgID,
		Payload:   payload,
		Timestamp: now,
	}

	msgs := b.streams[stream]
	msgs = append(msgs, msg)

	// MAXLEN trimming (FIFO)
	if maxLen > 0 && int64(len(msgs)) > maxLen {
		excess := int64(len(msgs)) - maxLen
		msgs = msgs[excess:]
	}

	b.streams[stream] = msgs
	return msgID, nil
}

// XLen returns the number of messages in a stream.
func (b *StreamBroker) XLen(stream string) int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return int64(len(b.streams[stream]))
}

// XRange returns up to `count` messages from the stream starting at `startID`.
func (b *StreamBroker) XRange(stream string, startID string, count int) []*StreamMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msgs := b.streams[stream]
	if len(msgs) == 0 {
		return nil
	}

	startIdx := 0
	if startID != "-" && startID != "" {
		for i, m := range msgs {
			if m.ID >= startID {
				startIdx = i
				break
			}
		}
	}

	endIdx := startIdx + count
	if count <= 0 || endIdx > len(msgs) {
		endIdx = len(msgs)
	}

	results := make([]*StreamMessage, endIdx-startIdx)
	copy(results, msgs[startIdx:endIdx])
	return results
}
