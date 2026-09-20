package src

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RaftRole represents the current consensus state of a node in the Raft cluster.
type RaftRole string

const (
	RoleFollower  RaftRole = "FOLLOWER"
	RoleCandidate RaftRole = "CANDIDATE"
	RoleLeader    RaftRole = "LEADER"
)

// RaftLogEntry represents an ordered transactional operation replicated across nodes.
type RaftLogEntry struct {
	Index   uint64 `json:"index"`
	Term    uint64 `json:"term"`
	Command []byte `json:"command"`
}

// RaftNodeConfig holds clustering topology parameters.
type RaftNodeConfig struct {
	NodeID          string
	Peers           []string
	ElectionTimeout time.Duration
	HeartbeatPeriod time.Duration
}

// RaftClusterNode models a high-availability Raft consensus state machine.
type RaftClusterNode struct {
	mu          sync.RWMutex
	nodeID      string
	peers       []string
	role        RaftRole
	currentTerm uint64
	votedFor    string
	log         []RaftLogEntry
	commitIndex uint64
	lastApplied uint64

	// Cluster quorum calculations
	quorumSize int

	lastHeartbeat time.Time
}

// NewRaftClusterNode initializes a node with follower state.
func NewRaftClusterNode(cfg RaftNodeConfig) *RaftClusterNode {
	totalNodes := len(cfg.Peers) + 1
	quorum := (totalNodes / 2) + 1

	return &RaftClusterNode{
		nodeID:        cfg.NodeID,
		peers:         cfg.Peers,
		role:          RoleFollower,
		currentTerm:   0,
		votedFor:      "",
		log:           make([]RaftLogEntry, 0),
		commitIndex:   0,
		lastApplied:   0,
		quorumSize:    quorum,
		lastHeartbeat: time.Now().UTC(),
	}
}

// GetState returns current role, term, and commit index.
func (r *RaftClusterNode) GetState() (RaftRole, uint64, uint64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.role, r.currentTerm, r.commitIndex
}

// Quorum returns required votes for majority consensus.
func (r *RaftClusterNode) Quorum() int {
	return r.quorumSize
}

// StartElection transitions follower/candidate to candidate, increments term, and votes for self.
func (r *RaftClusterNode) StartElection() (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.role = RoleCandidate
	r.currentTerm++
	r.votedFor = r.nodeID
	r.lastHeartbeat = time.Now().UTC()

	return r.currentTerm, nil
}

// BecomeLeader elevates a candidate to cluster leader upon achieving majority votes.
func (r *RaftClusterNode) BecomeLeader(grantedVotes int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.role != RoleCandidate {
		return errors.New("only candidate can transition to leader")
	}

	if grantedVotes < r.quorumSize {
		return fmt.Errorf("insufficient votes for quorum: got %d, required %d", grantedVotes, r.quorumSize)
	}

	r.role = RoleLeader
	return nil
}

// StepDown transitions leader/candidate back to follower when a higher term is discovered.
func (r *RaftClusterNode) StepDown(higherTerm uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if higherTerm > r.currentTerm {
		r.currentTerm = higherTerm
	}
	r.role = RoleFollower
	r.votedFor = ""
	r.lastHeartbeat = time.Now().UTC()
}

// RequestVoteArgs represents the RPC request to grant an election vote.
type RequestVoteArgs struct {
	Term         uint64
	CandidateID  string
	LastLogIndex uint64
	LastLogTerm  uint64
}

// RequestVoteReply represents the voting outcome.
type RequestVoteReply struct {
	Term        uint64
	VoteGranted bool
}

// HandleRequestVote processes vote solicitation from a candidate.
func (r *RaftClusterNode) HandleRequestVote(args RequestVoteArgs) RequestVoteReply {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If candidate term is smaller, reject
	if args.Term < r.currentTerm {
		return RequestVoteReply{Term: r.currentTerm, VoteGranted: false}
	}

	if args.Term > r.currentTerm {
		r.currentTerm = args.Term
		r.role = RoleFollower
		r.votedFor = ""
	}

	// Check if already voted for someone else in this term
	if r.votedFor == "" || r.votedFor == args.CandidateID {
		r.votedFor = args.CandidateID
		r.lastHeartbeat = time.Now().UTC()
		return RequestVoteReply{Term: r.currentTerm, VoteGranted: true}
	}

	return RequestVoteReply{Term: r.currentTerm, VoteGranted: false}
}

// AppendEntriesArgs represents log replication and heartbeat payload.
type AppendEntriesArgs struct {
	Term         uint64
	LeaderID     string
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []RaftLogEntry
	LeaderCommit uint64
}

// AppendEntriesReply represents the follower acknowledgement.
type AppendEntriesReply struct {
	Term    uint64
	Success bool
}

// HandleAppendEntries handles log replication and leader heartbeat.
func (r *RaftClusterNode) HandleAppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	r.mu.Lock()
	defer r.mu.Unlock()

	if args.Term < r.currentTerm {
		return AppendEntriesReply{Term: r.currentTerm, Success: false}
	}

	// Update term and heartbeat on valid leader contact
	if args.Term > r.currentTerm || r.role == RoleCandidate {
		r.currentTerm = args.Term
		r.role = RoleFollower
		r.votedFor = ""
	}
	r.lastHeartbeat = time.Now().UTC()

	// Append valid entries
	for _, entry := range args.Entries {
		r.log = append(r.log, entry)
	}

	if args.LeaderCommit > r.commitIndex {
		if uint64(len(r.log)) < args.LeaderCommit {
			r.commitIndex = uint64(len(r.log))
		} else {
			r.commitIndex = args.LeaderCommit
		}
	}

	return AppendEntriesReply{Term: r.currentTerm, Success: true}
}

// AppendCommand adds a command to the leader's log and replicates it.
func (r *RaftClusterNode) AppendCommand(command []byte) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.role != RoleLeader {
		return 0, errors.New("cannot append command to non-leader node")
	}

	index := uint64(len(r.log)) + 1
	entry := RaftLogEntry{
		Index:   index,
		Term:    r.currentTerm,
		Command: command,
	}
	r.log = append(r.log, entry)
	return index, nil
}
