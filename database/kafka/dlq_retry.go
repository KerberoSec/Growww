package kafka

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// DLQRetryRouter coordinates exponential retry topic hops and dead letter queuing.
type DLQRetryRouter struct {
	mu           sync.Mutex
	maxRetries   int
	retryTopics  []string // e.g. ["retry.1m", "retry.5m"]
	dlqMessages  []*KafkaMessage
	retryQueues  map[string][]*KafkaMessage
}

func NewDLQRetryRouter(maxRetries int, retryIntervals []string) *DLQRetryRouter {
	if maxRetries <= 0 {
		maxRetries = 2
	}
	if len(retryIntervals) == 0 {
		retryIntervals = []string{"retry.1m", "retry.5m"}
	}

	return &DLQRetryRouter{
		maxRetries:  maxRetries,
		retryTopics: retryIntervals,
		dlqMessages: make([]*KafkaMessage, 0),
		retryQueues: make(map[string][]*KafkaMessage),
	}
}

// RouteFailure inspects message retry count and dispatches to retry topic or terminal DLQ.
func (r *DLQRetryRouter) RouteFailure(msg *KafkaMessage, failureReason string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if msg.Headers == nil {
		msg.Headers = make(map[string]string)
	}

	msg.Headers["x-original-topic"] = msg.Topic
	msg.Headers["x-failure-reason"] = failureReason
	msg.Headers["x-failed-at"] = time.Now().UTC().Format(time.RFC3339)

	if msg.RetryCount < r.maxRetries {
		// Route to appropriate retry interval topic
		retryIdx := msg.RetryCount
		if retryIdx >= len(r.retryTopics) {
			retryIdx = len(r.retryTopics) - 1
		}
		retrySuffix := r.retryTopics[retryIdx]
		targetTopic := fmt.Sprintf("%s.%s", msg.Topic, retrySuffix)

		msg.RetryCount++
		msg.Headers["x-retry-count"] = strconv.Itoa(msg.RetryCount)
		r.retryQueues[targetTopic] = append(r.retryQueues[targetTopic], msg)

		return targetTopic, nil
	}

	// Max retries exceeded: Route to terminal DLQ
	dlqTopic := fmt.Sprintf("%s.dlq", msg.Topic)
	msg.Headers["x-terminal-dlq"] = "true"
	r.dlqMessages = append(r.dlqMessages, msg)
	return dlqTopic, ErrMaxRetriesExceeded
}

// GetDLQMessages returns all dead-lettered messages.
func (r *DLQRetryRouter) GetDLQMessages() []*KafkaMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]*KafkaMessage, len(r.dlqMessages))
	copy(res, r.dlqMessages)
	return res
}

// GetRetryMessages returns messages queued in a retry topic.
func (r *DLQRetryRouter) GetRetryMessages(retryTopic string) []*KafkaMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.retryQueues[retryTopic]
}
