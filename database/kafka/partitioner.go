package kafka

import (
	"hash/fnv"
)

// Partitioner routes a message to a partition based on its partition key.
type Partitioner struct{}

func NewPartitioner() *Partitioner {
	return &Partitioner{}
}

// PartitionKeyHash computes a 32-bit deterministic FNV-1a hash.
func (p *Partitioner) PartitionKeyHash(key string) uint32 {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	return hasher.Sum32()
}

// RoutePartition computes the target partition index for a key across partitionCount partitions.
// Guarantees that any identical key will deterministically land in the same partition.
func (p *Partitioner) RoutePartition(key string, partitionCount int) (int, error) {
	if key == "" {
		return 0, ErrInvalidPartitionKey
	}
	if partitionCount <= 0 {
		return 0, ErrTopicNotFound
	}

	hash := p.PartitionKeyHash(key)
	// Clear the sign bit to guarantee non-negative modulo
	partition := int(hash & 0x7fffffff) % partitionCount
	return partition, nil
}
