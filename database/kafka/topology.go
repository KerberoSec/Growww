package kafka

import (
	"fmt"
	"sync"
)

// Standard institutional retention periods (in milliseconds).
const (
	Retention7Days  = int64(7 * 24 * 3600 * 1000)
	Retention30Days = int64(30 * 24 * 3600 * 1000)
)

// TopicTopologyRegistry maintains declarative topic configurations across all bounded contexts.
type TopicTopologyRegistry struct {
	mu     sync.RWMutex
	topics map[string]*TopicConfig
}

// NewStandardTopicTopology initializes the standard Growww / NBSE topic layout.
func NewStandardTopicTopology() *TopicTopologyRegistry {
	reg := &TopicTopologyRegistry{
		topics: make(map[string]*TopicConfig),
	}

	standardTopics := []*TopicConfig{
		{
			Name:              "trading.orders.placed",
			Partitions:        32,
			ReplicationFactor: 3,
			MinISR:            2,
			PartitionKeyField: "isin",
			CleanupPolicy:     CleanupDelete,
			RetentionMs:       Retention7Days,
		},
		{
			Name:              "trading.trades.executed",
			Partitions:        32,
			ReplicationFactor: 3,
			MinISR:            2,
			PartitionKeyField: "isin",
			CleanupPolicy:     CleanupDelete,
			RetentionMs:       Retention30Days,
		},
		{
			Name:              "ledger.journal.postings",
			Partitions:        64,
			ReplicationFactor: 3,
			MinISR:            2,
			PartitionKeyField: "user_id",
			CleanupPolicy:     CleanupDelete,
			RetentionMs:       Retention30Days,
		},
		{
			Name:              "blockchain.besu.events",
			Partitions:        16,
			ReplicationFactor: 3,
			MinISR:            2,
			PartitionKeyField: "contract_address",
			CleanupPolicy:     CleanupDelete,
			RetentionMs:       Retention30Days,
		},
		{
			Name:              "compliance.investor.status",
			Partitions:        16,
			ReplicationFactor: 3,
			MinISR:            2,
			PartitionKeyField: "blockchain_address",
			CleanupPolicy:     CleanupCompact,
			RetentionMs:       -1, // compact topics retain forever
		},
	}

	for _, t := range standardTopics {
		_ = reg.RegisterTopic(t)
	}

	return reg
}

// RegisterTopic validates and registers a new topic definition.
func (r *TopicTopologyRegistry) RegisterTopic(cfg *TopicConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cfg.Name == "" {
		return fmt.Errorf("topic name cannot be empty")
	}
	if cfg.Partitions <= 0 {
		return fmt.Errorf("partitions must be strictly positive (got %d)", cfg.Partitions)
	}
	if cfg.ReplicationFactor < cfg.MinISR {
		return fmt.Errorf("replication factor (%d) cannot be less than min.insync.replicas (%d)",
			cfg.ReplicationFactor, cfg.MinISR)
	}

	r.topics[cfg.Name] = cfg
	return nil
}

// GetTopic retrieves a topic configuration.
func (r *TopicTopologyRegistry) GetTopic(name string) (*TopicConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.topics[name]
	if !exists {
		return nil, ErrTopicNotFound
	}
	copyCfg := *t
	return &copyCfg, nil
}

// ListTopics returns all registered topic configs.
func (r *TopicTopologyRegistry) ListTopics() []*TopicConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*TopicConfig, 0, len(r.topics))
	for _, t := range r.topics {
		copyCfg := *t
		result = append(result, &copyCfg)
	}
	return result
}
