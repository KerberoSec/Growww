package kafka

import (
	"errors"
	"time"
)

var (
	ErrTopicNotFound       = errors.New("kafka topic not found")
	ErrConsumerNotFound    = errors.New("consumer member not found in group")
	ErrRebalanceInProgress = errors.New("group rebalance in progress")
	ErrInvalidPartitionKey = errors.New("partition key is empty or invalid")
	ErrMaxRetriesExceeded  = errors.New("maximum retry limit exceeded, routed to DLQ")
)

type CleanupPolicy string

const (
	CleanupDelete  CleanupPolicy = "delete"
	CleanupCompact CleanupPolicy = "compact"
)

type GroupState string

const (
	GroupStateEmpty               GroupState = "Empty"
	GroupStatePreparingRebalance  GroupState = "PreparingRebalance"
	GroupStateCompletingRebalance GroupState = "CompletingRebalance"
	GroupStateStable              GroupState = "Stable"
	GroupStateDead                GroupState = "Dead"
)

// TopicConfig defines declarative topic governance rules.
type TopicConfig struct {
	Name              string        `json:"name"`
	Partitions        int           `json:"partitions"`
	ReplicationFactor int           `json:"replication_factor"`
	MinISR            int           `json:"min_isr"`
	PartitionKeyField string        `json:"partition_key_field"`
	CleanupPolicy     CleanupPolicy `json:"cleanup_policy"`
	RetentionMs       int64         `json:"retention_ms"`
}

// TopicPartition identifies a specific partition.
type TopicPartition struct {
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
}

// KafkaMessage represents an event flowing through the topic topology.
type KafkaMessage struct {
	Topic       string            `json:"topic"`
	Partition   int               `json:"partition"`
	Offset      int64             `json:"offset"`
	Key         string            `json:"key"`
	Value       []byte            `json:"value"`
	Headers     map[string]string `json:"headers"`
	Timestamp   time.Time         `json:"timestamp"`
	RetryCount  int               `json:"retry_count"`
}

// ConsumerMember represents an active consumer node in a consumer group.
type ConsumerMember struct {
	MemberID       string                    `json:"member_id"`
	ClientID       string                    `json:"client_id"`
	Host           string                    `json:"host"`
	AssignedParts  []TopicPartition          `json:"assigned_partitions"`
	LastHeartbeat  time.Time                 `json:"last_heartbeat"`
	SessionTimeout time.Duration             `json:"session_timeout"`
}

// PartitionLag tracks consumption lag per partition.
type PartitionLag struct {
	Topic           string `json:"topic"`
	Partition       int    `json:"partition"`
	HighWatermark   int64  `json:"high_watermark"`
	CommittedOffset int64  `json:"committed_offset"`
	Lag             int64  `json:"lag"`
}
