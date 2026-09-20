package kafka

import (
	"testing"
	"time"
)

func TestStandardTopicTopology(t *testing.T) {
	topology := NewStandardTopicTopology()

	// Verify trading.orders.placed
	ordersTopic, err := topology.GetTopic("trading.orders.placed")
	if err != nil {
		t.Fatalf("expected trading.orders.placed topic: %v", err)
	}
	if ordersTopic.Partitions != 32 || ordersTopic.ReplicationFactor != 3 || ordersTopic.MinISR != 2 {
		t.Errorf("orders topic specs mismatch: %+v", ordersTopic)
	}
	if ordersTopic.PartitionKeyField != "isin" {
		t.Errorf("expected isin partition key, got %s", ordersTopic.PartitionKeyField)
	}

	// Verify ledger.journal.postings
	ledgerTopic, err := topology.GetTopic("ledger.journal.postings")
	if err != nil {
		t.Fatalf("expected ledger.journal.postings topic: %v", err)
	}
	if ledgerTopic.Partitions != 64 || ledgerTopic.PartitionKeyField != "user_id" {
		t.Errorf("ledger topic specs mismatch: %+v", ledgerTopic)
	}

	// Verify compact topic
	compTopic, _ := topology.GetTopic("compliance.investor.status")
	if compTopic.CleanupPolicy != CleanupCompact {
		t.Errorf("expected compact cleanup policy, got %v", compTopic.CleanupPolicy)
	}
}

func TestPartitionKeyDeterminism(t *testing.T) {
	p := NewPartitioner()
	isin := "INE002A01018"
	partitions := 32

	part1, err := p.RoutePartition(isin, partitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must consistently map to identical partition across multiple calls
	for i := 0; i < 100; i++ {
		part2, _ := p.RoutePartition(isin, partitions)
		if part1 != part2 {
			t.Fatalf("determinism violated: got %d then %d", part1, part2)
		}
	}

	// Empty key should return error
	if _, err := p.RoutePartition("", partitions); err != ErrInvalidPartitionKey {
		t.Errorf("expected ErrInvalidPartitionKey on empty key, got %v", err)
	}
}

func TestPartitionAssignors(t *testing.T) {
	members := []*ConsumerMember{
		{MemberID: "c-1"},
		{MemberID: "c-2"},
	}

	partitions := []TopicPartition{
		{Topic: "trading.orders.placed", Partition: 0},
		{Topic: "trading.orders.placed", Partition: 1},
		{Topic: "trading.orders.placed", Partition: 2},
		{Topic: "trading.orders.placed", Partition: 3},
	}

	// 1. Range Assignor
	rangeAssignor := &RangeAssignor{}
	rangeAssign := rangeAssignor.Assign(members, partitions)
	if len(rangeAssign["c-1"]) != 2 || len(rangeAssign["c-2"]) != 2 {
		t.Errorf("expected 2 partitions each in RangeAssignor")
	}

	// 2. RoundRobin Assignor
	rrAssignor := &RoundRobinAssignor{}
	rrAssign := rrAssignor.Assign(members, partitions)
	if len(rrAssign["c-1"]) != 2 || len(rrAssign["c-2"]) != 2 {
		t.Errorf("expected 2 partitions each in RoundRobinAssignor")
	}

	// 3. Sticky Assignor
	stickyAssignor := &StickyAssignor{}
	stickyAssign := stickyAssignor.Assign(members, partitions)
	if len(stickyAssign["c-1"]) != 2 || len(stickyAssign["c-2"]) != 2 {
		t.Errorf("expected 2 partitions each in StickyAssignor")
	}
}

func TestConsumerGroupCoordinatorRebalanceAndEviction(t *testing.T) {
	topicName := "trading.orders.placed"
	partsPerTopic := map[string]int{topicName: 4}

	coord := NewConsumerGroupCoordinator(
		"order-matching-group",
		[]string{topicName},
		partsPerTopic,
		&StickyAssignor{},
	)

	// Member 1 joins
	m1 := &ConsumerMember{MemberID: "member-1", SessionTimeout: 50 * time.Millisecond}
	gen1, parts1, err := coord.JoinGroup(m1)
	if err != nil || gen1 != 1 || len(parts1) != 4 {
		t.Fatalf("expected member 1 to receive all 4 partitions in gen 1, got gen %d, parts %d", gen1, len(parts1))
	}

	// Member 2 joins -> triggers cooperative rebalance to generation 2
	m2 := &ConsumerMember{MemberID: "member-2", SessionTimeout: 50 * time.Millisecond}
	gen2, parts2, err := coord.JoinGroup(m2)
	if err != nil || gen2 != 2 || len(parts2) != 2 {
		t.Fatalf("expected member 2 to receive 2 partitions in gen 2, got gen %d, parts %d", gen2, len(parts2))
	}

	assignment1, _ := coord.GetMemberAssignment("member-1")
	if len(assignment1) != 2 {
		t.Errorf("expected member 1 to retain 2 partitions after rebalance, got %d", len(assignment1))
	}

	// Member 1 maintains regular heartbeat, Member 2 stops heartbeating
	time.Sleep(30 * time.Millisecond)
	_ = coord.Heartbeat("member-1")
	time.Sleep(30 * time.Millisecond)

	// Evict stale member 2 (missed session timeout of 50ms, while member 1 refreshed 30ms ago)
	evicted := coord.EvictStaleMembers(time.Now().UTC())
	if evicted != 1 {
		t.Fatalf("expected 1 evicted member, got %d", evicted)
	}

	// Generation should now be 3, member 1 re-assigned all 4 partitions
	if coord.GenerationID() != 3 {
		t.Errorf("expected generation 3 after eviction rebalance, got %d", coord.GenerationID())
	}
	assignmentAfterEvict, _ := coord.GetMemberAssignment("member-1")
	if len(assignmentAfterEvict) != 4 {
		t.Errorf("expected member 1 to reclaim all 4 partitions, got %d", len(assignmentAfterEvict))
	}
}

func TestOffsetCommitAndLagTelemetry(t *testing.T) {
	topicName := "ledger.journal.postings"
	partsPerTopic := map[string]int{topicName: 2}

	coord := NewConsumerGroupCoordinator(
		"ledger-writer-group",
		[]string{topicName},
		partsPerTopic,
		&RangeAssignor{},
	)

	tp0 := TopicPartition{Topic: topicName, Partition: 0}
	tp1 := TopicPartition{Topic: topicName, Partition: 1}

	// Set high watermarks (produced messages)
	coord.UpdateHighWatermark(tp0, 1500)
	coord.UpdateHighWatermark(tp1, 3000)

	// Commit consumer offsets
	coord.CommitOffset(tp0, 1400) // 100 lag
	coord.CommitOffset(tp1, 2800) // 200 lag

	lag0 := coord.GetPartitionLag(tp0)
	if lag0.Lag != 100 {
		t.Errorf("expected lag 100 on tp0, got %d", lag0.Lag)
	}

	lag1 := coord.GetPartitionLag(tp1)
	if lag1.Lag != 200 {
		t.Errorf("expected lag 200 on tp1, got %d", lag1.Lag)
	}

	totalLag := coord.GetTotalGroupLag()
	if totalLag != 300 {
		t.Errorf("expected total group lag 300, got %d", totalLag)
	}
}

func TestDLQRetryRouting(t *testing.T) {
	router := NewDLQRetryRouter(2, []string{"retry.1m", "retry.5m"})

	msg := &KafkaMessage{
		Topic:      "trading.orders.placed",
		Partition:  3,
		Offset:     1204,
		Key:        "INE002A01018",
		Value:      []byte(`{"order_id": "ord-fail-1"}`),
		RetryCount: 0,
	}

	// 1st failure: route to .retry.1m
	target1, err := router.RouteFailure(msg, "Transient matching engine timeout")
	if err != nil || target1 != "trading.orders.placed.retry.1m" {
		t.Fatalf("expected retry.1m topic, got %s (err: %v)", target1, err)
	}
	if msg.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", msg.RetryCount)
	}

	// 2nd failure: route to .retry.5m
	target2, err := router.RouteFailure(msg, "Transient matching engine timeout")
	if err != nil || target2 != "trading.orders.placed.retry.5m" {
		t.Fatalf("expected retry.5m topic, got %s (err: %v)", target2, err)
	}
	if msg.RetryCount != 2 {
		t.Errorf("expected retry count 2, got %d", msg.RetryCount)
	}

	// 3rd failure: maxRetries (2) exceeded -> route to .dlq!
	target3, err := router.RouteFailure(msg, "Permanent deserialization failure")
	if err != ErrMaxRetriesExceeded || target3 != "trading.orders.placed.dlq" {
		t.Fatalf("expected dlq topic with ErrMaxRetriesExceeded, got %s (err: %v)", target3, err)
	}

	dlqMsgs := router.GetDLQMessages()
	if len(dlqMsgs) != 1 {
		t.Fatalf("expected 1 DLQ message, got %d", len(dlqMsgs))
	}
	if dlqMsgs[0].Headers["x-terminal-dlq"] != "true" {
		t.Errorf("expected terminal DLQ header")
	}
}
