use matching_engine_wal_failover::{
    MmapWALSegment, NodeState, RaftClusterTopology, RaftRole, RaftSequencer,
    SequencedLogEntry,
};

#[test]
fn test_mmap_wal_segment_append_and_read() {
    let mut segment = MmapWALSegment::new(1, 1024);
    assert_eq!(segment.written_bytes, 0);
    assert_eq!(segment.remaining_capacity(), 1024);
    assert!(!segment.is_full());

    let rec1 = b"ORDER:BTC/USDT:BUY:65000:100";
    let rec2 = b"ORDER:ETH/USDT:SELL:3500:200";

    let written1 = segment.append_record(rec1).expect("rec1 append failed");
    assert_eq!(written1, rec1.len() + 4);

    let written2 = segment.append_record(rec2).expect("rec2 append failed");
    assert_eq!(written2, rec1.len() + 4 + rec2.len() + 4);

    let records = segment.read_records().expect("read records failed");
    assert_eq!(records.len(), 2);
    assert_eq!(records[0], rec1);
    assert_eq!(records[1], rec2);
}

#[test]
fn test_mmap_wal_segment_capacity_overflow() {
    let mut segment = MmapWALSegment::new(2, 20); // small capacity
    let big_rec = vec![0u8; 25];
    let err = segment.append_record(&big_rec).unwrap_err();
    assert!(err.contains("Segment capacity exceeded"));
}

#[test]
fn test_raft_cluster_topology_quorum_and_commit() {
    let peers = vec![
        (2, "10.0.0.2:9092".to_string()),
        (3, "10.0.0.3:9092".to_string()),
    ];
    let mut topo = RaftClusterTopology::new(1, peers);
    assert_eq!(topo.state, NodeState::Follower);
    assert_eq!(topo.commit_index, 0);

    // Initial check: only node 1 at seq 10, no quorum
    assert!(!topo.is_quorum_replicated(10));

    // Update peer 2 to seq 10 (now 2 out of 3 agree: node 1 + node 2)
    topo.update_peer_match(2, 10).expect("update peer 2");
    assert!(topo.is_quorum_replicated(10));

    // Sequence 15: only peer 2 is not at 15
    assert!(!topo.is_quorum_replicated(15));

    // Peer 3 catches up to 15
    topo.update_peer_match(3, 15).expect("update peer 3");
    assert!(topo.is_quorum_replicated(15));

    // Election transition
    topo.become_candidate();
    assert_eq!(topo.state, NodeState::Candidate);
    assert_eq!(topo.current_term, 2);
    assert_eq!(topo.voted_for, Some(1));

    topo.become_leader().expect("become leader");
    assert_eq!(topo.state, NodeState::Leader);

    // Step down when higher term arrives
    topo.step_down(3);
    assert_eq!(topo.state, NodeState::Follower);
    assert_eq!(topo.current_term, 3);
    assert_eq!(topo.voted_for, None);
}

#[test]
fn test_raft_sequencer_monotonic_events() {
    let mut seq = RaftSequencer::new("node-alpha".into());
    assert_eq!(seq.role, RaftRole::Leader);
    assert_eq!(seq.last_sequence, 0);

    let e1 = seq.sequence_event(b"TRADE:1".to_vec()).unwrap();
    assert_eq!(e1.sequence_number, 1);
    assert_eq!(e1.term, 1);

    let e2 = seq.sequence_event(b"TRADE:2".to_vec()).unwrap();
    assert_eq!(e2.sequence_number, 2);

    let e3 = seq.sequence_event(b"TRADE:3".to_vec()).unwrap();
    assert_eq!(e3.sequence_number, 3);

    assert_eq!(seq.last_sequence, 3);
    assert_eq!(seq.wal_log.len(), 3);

    // Replay from seq 2
    let replay = seq.replay_from_sequence(2);
    assert_eq!(replay.len(), 2);
    assert_eq!(replay[0].sequence_number, 2);
    assert_eq!(replay[1].sequence_number, 3);
}

#[test]
fn test_raft_sequencer_follower_append_and_gaps() {
    let mut follower = RaftSequencer::new("node-beta".into());
    follower.role = RaftRole::Follower;

    let entry1 = SequencedLogEntry {
        sequence_number: 1,
        term: 1,
        payload: b"EVENT:1".to_vec(),
        timestamp_ns: 1000,
    };
    follower.append_from_leader(entry1.clone()).expect("append entry 1");
    assert_eq!(follower.last_sequence, 1);

    // Duplicate append is idempotent
    follower.append_from_leader(entry1).expect("idempotent duplicate");
    assert_eq!(follower.last_sequence, 1);

    // Out of order / gap entry should be rejected
    let gap_entry = SequencedLogEntry {
        sequence_number: 5,
        term: 1,
        payload: b"EVENT:5".to_vec(),
        timestamp_ns: 2000,
    };
    let err = follower.append_from_leader(gap_entry).unwrap_err();
    assert!(err.contains("Sequence gap detected"));
}

#[test]
fn test_raft_sequencer_snapshot_and_restore() {
    let mut sequencer = RaftSequencer::new("node-gamma".into());

    for i in 1..=5 {
        sequencer.sequence_event(format!("ORDER:{}", i).into_bytes()).unwrap();
    }
    assert_eq!(sequencer.wal_log.len(), 5);

    // Commit up to 3
    sequencer.commit_up_to(3).expect("commit up to 3");
    assert_eq!(sequencer.commit_sequence, 3);

    // Snapshot state
    let snapshot = sequencer
        .create_snapshot(b"SNAPSHOT_STATE_DATA".to_vec())
        .expect("snapshot creation");
    assert_eq!(snapshot.last_included_sequence, 3);

    // Log should be compacted, keeping only sequences > 3
    assert_eq!(sequencer.wal_log.len(), 2);
    assert_eq!(sequencer.wal_log[0].sequence_number, 4);
    assert_eq!(sequencer.wal_log[1].sequence_number, 5);

    // New node restoring from snapshot
    let mut fresh_node = RaftSequencer::new("node-delta".into());
    fresh_node.restore_from_snapshot(snapshot);
    assert_eq!(fresh_node.last_sequence, 3);
    assert_eq!(fresh_node.commit_sequence, 3);
}

#[test]
fn test_raft_sequencer_failover_and_rejection_when_not_leader() {
    let mut seq = RaftSequencer::new("node-omega".into());
    seq.step_down_to_follower(2);
    assert_eq!(seq.role, RaftRole::Follower);

    let err = seq.sequence_event(b"CANNOT_SEQUENCE".to_vec()).unwrap_err();
    assert!(err.contains("Not current Raft leader"));

    // Step up via failover
    seq.failover_to_leader(3).expect("failover to leader");
    assert_eq!(seq.role, RaftRole::Leader);
    assert_eq!(seq.current_term, 3);

    let ok_entry = seq.sequence_event(b"LEADER_SEQUENCE".to_vec()).expect("sequence as leader");
    assert_eq!(ok_entry.term, 3);
    assert_eq!(ok_entry.sequence_number, 1);
}
