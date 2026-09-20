package src

import (
	"testing"
	"time"
)

func TestRaftClusterConsensusAndLeadership(t *testing.T) {
	cfgNode1 := RaftNodeConfig{
		NodeID:          "engine-node-01",
		Peers:           []string{"engine-node-02", "engine-node-03"},
		ElectionTimeout: 150 * time.Millisecond,
		HeartbeatPeriod: 50 * time.Millisecond,
	}

	cfgNode2 := RaftNodeConfig{
		NodeID:          "engine-node-02",
		Peers:           []string{"engine-node-01", "engine-node-03"},
		ElectionTimeout: 150 * time.Millisecond,
		HeartbeatPeriod: 50 * time.Millisecond,
	}

	node1 := NewRaftClusterNode(cfgNode1)
	node2 := NewRaftClusterNode(cfgNode2)

	// In 3-node cluster, Quorum should be (3/2) + 1 = 2
	if node1.Quorum() != 2 {
		t.Fatalf("expected quorum of 2, got %d", node1.Quorum())
	}

	// Node 1 starts election
	term, err := node1.StartElection()
	if err != nil || term != 1 {
		t.Fatalf("start election failed: %v", err)
	}

	role, currentTerm, _ := node1.GetState()
	if role != RoleCandidate || currentTerm != 1 {
		t.Fatalf("expected candidate role with term 1, got %s term %d", role, currentTerm)
	}

	// Solicit vote from node 2
	reply := node2.HandleRequestVote(RequestVoteArgs{
		Term:         term,
		CandidateID:  "engine-node-01",
		LastLogIndex: 0,
		LastLogTerm:  0,
	})

	if !reply.VoteGranted {
		t.Fatalf("expected vote to be granted from node 2")
	}

	// Node 1 receives self-vote + node2 vote = 2 votes (Quorum achieved)
	if err := node1.BecomeLeader(2); err != nil {
		t.Fatalf("failed to transition to leader: %v", err)
	}

	role, _, _ = node1.GetState()
	if role != RoleLeader {
		t.Fatalf("expected leader role, got %s", role)
	}

	// Node 1 replicates a trade log command
	cmd := []byte(`{"order_id":"ORD-1001","price_e8":250000000}`)
	idx, err := node1.AppendCommand(cmd)
	if err != nil || idx != 1 {
		t.Fatalf("failed to append command: %v", err)
	}

	// Send AppendEntries heartbeat/log to node 2
	appendReply := node2.HandleAppendEntries(AppendEntriesArgs{
		Term:         term,
		LeaderID:     "engine-node-01",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries: []RaftLogEntry{
			{Index: 1, Term: term, Command: cmd},
		},
		LeaderCommit: 1,
	})

	if !appendReply.Success {
		t.Fatalf("expected append entries to succeed")
	}

	// Step-down check: higher term discovered
	node1.StepDown(5)
	roleAfterStepDown, newTerm, _ := node1.GetState()
	if roleAfterStepDown != RoleFollower || newTerm != 5 {
		t.Fatalf("expected follower after stepdown with term 5, got %s term %d", roleAfterStepDown, newTerm)
	}
}
