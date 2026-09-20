package keyspace

import (
	"context"
	"strings"
	"sync"
)

// RedisKeyspaceListener receives raw notifications from Redis Pub/Sub channels.
type RedisKeyspaceListener struct {
	mu         sync.Mutex
	dispatcher *ExpiryDispatcher
	poller     *ActiveReconciliationPoller
}

// NewRedisKeyspaceListener creates the listener coordinator.
func NewRedisKeyspaceListener(dispatcher *ExpiryDispatcher, poller *ActiveReconciliationPoller) *RedisKeyspaceListener {
	return &RedisKeyspaceListener{
		dispatcher: dispatcher,
		poller:     poller,
	}
}

// HandleRawPubSubMessage processes an incoming channel message.
// Supported channels:
// - `__keyevent@0__:expired` with message payload being the expired key
// - `__keyspace@0__:<key>` with message payload being "expired" or "del"
func (l *RedisKeyspaceListener) HandleRawPubSubMessage(ctx context.Context, channel, payload string) error {
	var key string
	if strings.HasPrefix(channel, "__keyevent@") {
		// e.g. channel="__keyevent@0__:expired", payload="order:exp:ord-123:BTC-USDT"
		key = payload
	} else if strings.HasPrefix(channel, "__keyspace@") {
		// e.g. channel="__keyspace@0__:order:exp:ord-123:BTC-USDT", payload="expired"
		parts := strings.SplitN(channel, ":", 2)
		if len(parts) == 2 {
			key = parts[1]
		}
	} else {
		key = payload
	}

	event := ParseKey(key)
	event.Channel = channel

	// Remove from active tracking since pub/sub succeeded
	if l.poller != nil {
		l.poller.RemoveTimer(key)
	}

	return l.dispatcher.Dispatch(ctx, event)
}
