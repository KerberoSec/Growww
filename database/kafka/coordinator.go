package kafka

import (
	"sync"
	"time"
)

// ConsumerGroupCoordinator manages group membership, cooperative rebalancing, and offsets.
type ConsumerGroupCoordinator struct {
	mu               sync.RWMutex
	groupID          string
	topics           []string
	assignor         PartitionAssignor
	state            GroupState
	generationID     int
	leaderMemberID   string
	members          map[string]*ConsumerMember
	committedOffsets map[TopicPartition]int64
	highWatermarks   map[TopicPartition]int64
	subscribedParts  []TopicPartition
}

func NewConsumerGroupCoordinator(
	groupID string,
	topics []string,
	partitionsPerTopic map[string]int,
	assignor PartitionAssignor,
) *ConsumerGroupCoordinator {
	if assignor == nil {
		assignor = &StickyAssignor{}
	}

	var parts []TopicPartition
	for _, t := range topics {
		count := partitionsPerTopic[t]
		for i := 0; i < count; i++ {
			parts = append(parts, TopicPartition{Topic: t, Partition: i})
		}
	}

	return &ConsumerGroupCoordinator{
		groupID:          groupID,
		topics:           topics,
		assignor:         assignor,
		state:            GroupStateEmpty,
		generationID:     0,
		members:          make(map[string]*ConsumerMember),
		committedOffsets: make(map[TopicPartition]int64),
		highWatermarks:   make(map[TopicPartition]int64),
		subscribedParts:  parts,
	}
}

// JoinGroup adds a consumer member to the group and triggers a rebalance.
func (c *ConsumerGroupCoordinator) JoinGroup(member *ConsumerMember) (int, []TopicPartition, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if member.SessionTimeout <= 0 {
		member.SessionTimeout = 10 * time.Second
	}
	member.LastHeartbeat = time.Now().UTC()

	c.members[member.MemberID] = member
	if c.leaderMemberID == "" {
		c.leaderMemberID = member.MemberID
	}

	// Trigger rebalance
	c.executeRebalanceLocked()

	return c.generationID, c.members[member.MemberID].AssignedParts, nil
}

// LeaveGroup removes a member from the group and triggers a rebalance.
func (c *ConsumerGroupCoordinator) LeaveGroup(memberID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.members[memberID]; !exists {
		return ErrConsumerNotFound
	}

	delete(c.members, memberID)

	if len(c.members) == 0 {
		c.state = GroupStateEmpty
		c.leaderMemberID = ""
		return nil
	}

	if c.leaderMemberID == memberID {
		// Elect first available member as leader
		for id := range c.members {
			c.leaderMemberID = id
			break
		}
	}

	c.executeRebalanceLocked()
	return nil
}

// Heartbeat refreshes the active liveness timestamp for a consumer.
func (c *ConsumerGroupCoordinator) Heartbeat(memberID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, exists := c.members[memberID]
	if !exists {
		return ErrConsumerNotFound
	}
	m.LastHeartbeat = time.Now().UTC()
	return nil
}

// EvictStaleMembers evicts any member that missed heartbeat beyond session timeout,
// triggering an automatic rebalance.
func (c *ConsumerGroupCoordinator) EvictStaleMembers(now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	evicted := 0
	for id, m := range c.members {
		if now.Sub(m.LastHeartbeat) > m.SessionTimeout {
			delete(c.members, id)
			evicted++
		}
	}

	if evicted > 0 {
		if len(c.members) == 0 {
			c.state = GroupStateEmpty
			c.leaderMemberID = ""
		} else {
			if _, leaderExists := c.members[c.leaderMemberID]; !leaderExists {
				for id := range c.members {
					c.leaderMemberID = id
					break
				}
			}
			c.executeRebalanceLocked()
		}
	}

	return evicted
}

func (c *ConsumerGroupCoordinator) executeRebalanceLocked() {
	c.state = GroupStatePreparingRebalance
	c.generationID++

	memberList := make([]*ConsumerMember, 0, len(c.members))
	for _, m := range c.members {
		memberList = append(memberList, m)
	}

	c.state = GroupStateCompletingRebalance
	assignment := c.assignor.Assign(memberList, c.subscribedParts)

	for memberID, parts := range assignment {
		if m, exists := c.members[memberID]; exists {
			m.AssignedParts = parts
		}
	}

	c.state = GroupStateStable
}

// CommitOffset updates the committed offset for a partition.
func (c *ConsumerGroupCoordinator) CommitOffset(tp TopicPartition, offset int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.committedOffsets[tp] = offset
}

// UpdateHighWatermark updates the latest message offset in the partition log.
func (c *ConsumerGroupCoordinator) UpdateHighWatermark(tp TopicPartition, watermark int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.highWatermarks[tp] = watermark
}

// GetPartitionLag returns the consumer lag for a partition.
func (c *ConsumerGroupCoordinator) GetPartitionLag(tp TopicPartition) PartitionLag {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hw := c.highWatermarks[tp]
	committed := c.committedOffsets[tp]
	lag := hw - committed
	if lag < 0 {
		lag = 0
	}

	return PartitionLag{
		Topic:           tp.Topic,
		Partition:       tp.Partition,
		HighWatermark:   hw,
		CommittedOffset: committed,
		Lag:             lag,
	}
}

// GetTotalGroupLag calculates aggregated lag across all partitions.
func (c *ConsumerGroupCoordinator) GetTotalGroupLag() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var total int64
	for _, tp := range c.subscribedParts {
		hw := c.highWatermarks[tp]
		committed := c.committedOffsets[tp]
		lag := hw - committed
		if lag > 0 {
			total += lag
		}
	}
	return total
}

// GetMemberAssignment returns partitions assigned to a consumer.
func (c *ConsumerGroupCoordinator) GetMemberAssignment(memberID string) ([]TopicPartition, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m, exists := c.members[memberID]
	if !exists {
		return nil, ErrConsumerNotFound
	}
	res := make([]TopicPartition, len(m.AssignedParts))
	copy(res, m.AssignedParts)
	return res, nil
}

// State returns current coordinator state.
func (c *ConsumerGroupCoordinator) State() GroupState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// GenerationID returns current generation counter.
func (c *ConsumerGroupCoordinator) GenerationID() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.generationID
}
