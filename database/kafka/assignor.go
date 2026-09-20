package kafka

import (
	"sort"
)

// PartitionAssignor defines strategy for allocating partitions among consumer group members.
type PartitionAssignor interface {
	Assign(members []*ConsumerMember, partitions []TopicPartition) map[string][]TopicPartition
}

// RangeAssignor assigns contiguous ranges of partitions per topic.
type RangeAssignor struct{}

func (ra *RangeAssignor) Assign(members []*ConsumerMember, partitions []TopicPartition) map[string][]TopicPartition {
	assignment := make(map[string][]TopicPartition)
	if len(members) == 0 || len(partitions) == 0 {
		return assignment
	}

	sortedMembers := make([]*ConsumerMember, len(members))
	copy(sortedMembers, members)
	sort.Slice(sortedMembers, func(i, j int) bool {
		return sortedMembers[i].MemberID < sortedMembers[j].MemberID
	})

	for _, m := range sortedMembers {
		assignment[m.MemberID] = make([]TopicPartition, 0)
	}

	// Group partitions by topic
	topicMap := make(map[string][]TopicPartition)
	for _, p := range partitions {
		topicMap[p.Topic] = append(topicMap[p.Topic], p)
	}

	for _, parts := range topicMap {
		numParts := len(parts)
		numMembers := len(sortedMembers)

		numPartsPerMember := numParts / numMembers
		extraParts := numParts % numMembers

		partIndex := 0
		for i, m := range sortedMembers {
			take := numPartsPerMember
			if i < extraParts {
				take++
			}
			for j := 0; j < take; j++ {
				assignment[m.MemberID] = append(assignment[m.MemberID], parts[partIndex])
				partIndex++
			}
		}
	}

	return assignment
}

// RoundRobinAssignor distributes partitions in round-robin fashion across sorted members.
type RoundRobinAssignor struct{}

func (rr *RoundRobinAssignor) Assign(members []*ConsumerMember, partitions []TopicPartition) map[string][]TopicPartition {
	assignment := make(map[string][]TopicPartition)
	if len(members) == 0 || len(partitions) == 0 {
		return assignment
	}

	sortedMembers := make([]*ConsumerMember, len(members))
	copy(sortedMembers, members)
	sort.Slice(sortedMembers, func(i, j int) bool {
		return sortedMembers[i].MemberID < sortedMembers[j].MemberID
	})

	for _, m := range sortedMembers {
		assignment[m.MemberID] = make([]TopicPartition, 0)
	}

	for i, p := range partitions {
		member := sortedMembers[i%len(sortedMembers)]
		assignment[member.MemberID] = append(assignment[member.MemberID], p)
	}

	return assignment
}

// StickyAssignor maximizes partition preservation during rebalancing (Cooperative Sticky Rebalance).
type StickyAssignor struct{}

func (sa *StickyAssignor) Assign(members []*ConsumerMember, partitions []TopicPartition) map[string][]TopicPartition {
	assignment := make(map[string][]TopicPartition)
	if len(members) == 0 || len(partitions) == 0 {
		return assignment
	}

	sortedMembers := make([]*ConsumerMember, len(members))
	copy(sortedMembers, members)
	sort.Slice(sortedMembers, func(i, j int) bool {
		return sortedMembers[i].MemberID < sortedMembers[j].MemberID
	})

	memberMap := make(map[string]bool)
	for _, m := range sortedMembers {
		assignment[m.MemberID] = make([]TopicPartition, 0)
		memberMap[m.MemberID] = true
	}

	unassigned := make([]TopicPartition, 0)
	assignedParts := make(map[TopicPartition]bool)

	// Keep existing assignments if member is still active
	targetPerMember := (len(partitions) + len(sortedMembers) - 1) / len(sortedMembers)

	for _, m := range sortedMembers {
		for _, p := range m.AssignedParts {
			if len(assignment[m.MemberID]) < targetPerMember && !assignedParts[p] {
				assignment[m.MemberID] = append(assignment[m.MemberID], p)
				assignedParts[p] = true
			}
		}
	}

	// Find remaining unassigned partitions
	for _, p := range partitions {
		if !assignedParts[p] {
			unassigned = append(unassigned, p)
		}
	}

	// Assign remaining round-robin to members with capacity
	memberIdx := 0
	for _, p := range unassigned {
		// Find member with lowest current count
		minMember := sortedMembers[0]
		minCount := len(assignment[minMember.MemberID])
		for _, m := range sortedMembers {
			if len(assignment[m.MemberID]) < minCount {
				minMember = m
				minCount = len(assignment[m.MemberID])
			}
		}
		assignment[minMember.MemberID] = append(assignment[minMember.MemberID], p)
		memberIdx++
	}

	return assignment
}
