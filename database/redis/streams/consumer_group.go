package streams

import (
	"context"
	"sync"
	"time"
)

// ConsumerGroup manages message delivery, pending states (PEL), and auto-claiming for a stream.
type ConsumerGroup struct {
	mu           sync.Mutex
	groupName    string
	streamName   string
	broker       *StreamBroker
	lastDeliveredIndex int
	pel          map[string]*PendingMessage // msg_id -> pending info
	dlq          []*StreamMessage
	maxRetries   int
}

func NewConsumerGroup(groupName string, streamName string, broker *StreamBroker, maxRetries int) *ConsumerGroup {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &ConsumerGroup{
		groupName:  groupName,
		streamName: streamName,
		broker:     broker,
		pel:        make(map[string]*PendingMessage),
		dlq:        make([]*StreamMessage, 0),
		maxRetries: maxRetries,
	}
}

// XReadGroup reads new messages for a specific consumer and adds them to the PEL.
func (cg *ConsumerGroup) XReadGroup(
	ctx context.Context,
	consumer string,
	count int,
) ([]*StreamMessage, error) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	cg.broker.mu.RLock()
	allMsgs := cg.broker.streams[cg.streamName]
	cg.broker.mu.RUnlock()

	if cg.lastDeliveredIndex >= len(allMsgs) {
		return nil, nil // No new messages
	}

	start := cg.lastDeliveredIndex
	end := start + count
	if end > len(allMsgs) {
		end = len(allMsgs)
	}

	delivered := make([]*StreamMessage, 0, end-start)
	now := time.Now().UTC()

	for i := start; i < end; i++ {
		msg := allMsgs[i]
		delivered = append(delivered, msg)

		// Record in PEL
		cg.pel[msg.ID] = &PendingMessage{
			ID:               msg.ID,
			Consumer:         consumer,
			DeliveredAt:      now,
			DeliveryAttempts: 1,
		}
	}

	cg.lastDeliveredIndex = end
	return delivered, nil
}

// XAck acknowledges that messages were processed, removing them from the PEL.
func (cg *ConsumerGroup) XAck(ctx context.Context, msgIDs ...string) int {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	ackedCount := 0
	for _, id := range msgIDs {
		if _, exists := cg.pel[id]; exists {
			delete(cg.pel, id)
			ackedCount++
		}
	}
	return ackedCount
}

// XPending returns all currently pending messages in the PEL.
func (cg *ConsumerGroup) XPending() []*PendingMessage {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	res := make([]*PendingMessage, 0, len(cg.pel))
	for _, p := range cg.pel {
		res = append(res, p)
	}
	return res
}

// XAutoClaim reclaims messages idle for longer than `minIdleTime` to a new consumer.
// If delivery attempts exceed `maxRetries`, message is routed to DLQ.
func (cg *ConsumerGroup) XAutoClaim(
	ctx context.Context,
	newConsumer string,
	minIdleTime time.Duration,
	count int,
) ([]*PendingMessage, []*StreamMessage) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	now := time.Now().UTC()
	claimed := make([]*PendingMessage, 0)
	var dlqEmitted []*StreamMessage

	cg.broker.mu.RLock()
	allMsgsMap := make(map[string]*StreamMessage)
	for _, m := range cg.broker.streams[cg.streamName] {
		allMsgsMap[m.ID] = m
	}
	cg.broker.mu.RUnlock()

	for msgID, p := range cg.pel {
		if len(claimed) >= count {
			break
		}

		if now.Sub(p.DeliveredAt) >= minIdleTime {
			p.DeliveryAttempts++
			if p.DeliveryAttempts > cg.maxRetries {
				// Exceeded retry limit -> move to DLQ and remove from PEL
				delete(cg.pel, msgID)
				if msg, exists := allMsgsMap[msgID]; exists {
					cg.dlq = append(cg.dlq, msg)
					dlqEmitted = append(dlqEmitted, msg)
				}
				continue
			}

			// Claim message for new consumer
			p.Consumer = newConsumer
			p.DeliveredAt = now
			claimed = append(claimed, p)
		}
	}

	return claimed, dlqEmitted
}

// DLQ returns dead letter queue messages.
func (cg *ConsumerGroup) DLQ() []*StreamMessage {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	res := make([]*StreamMessage, len(cg.dlq))
	copy(res, cg.dlq)
	return res
}
