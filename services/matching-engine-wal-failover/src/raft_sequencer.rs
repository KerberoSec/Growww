use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RaftRole {
    Leader,
    Follower,
    Candidate,
}

#[derive(Debug, Clone)]
pub struct SequencedLogEntry {
    pub sequence_number: u64,
    pub term: u64,
    pub payload: Vec<u8>,
    pub timestamp_ns: u64,
}

pub struct RaftSequencer {
    pub node_id: String,
    pub current_term: u64,
    pub role: RaftRole,
    pub last_sequence: u64,
    pub wal_log: Vec<SequencedLogEntry>,
}

impl RaftSequencer {
    pub fn new(node_id: String) -> Self {
        Self {
            node_id,
            current_term: 1,
            role: RaftRole::Leader, // Assuming active partition leader
            last_sequence: 0,
            wal_log: Vec::new(),
        }
    }

    /// Append and sequence an incoming order or trade event deterministically
    pub fn sequence_event(&mut self, payload: Vec<u8>) -> Result<SequencedLogEntry, String> {
        if self.role != RaftRole::Leader {
            return Err("Not current Raft leader; cannot sequence events".into());
        }

        self.last_sequence += 1;
        let now_ns = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_nanos() as u64;

        let entry = SequencedLogEntry {
            sequence_number: self.last_sequence,
            term: self.current_term,
            payload,
            timestamp_ns: now_ns,
        };

        // In production, mmap zero-copy append to durable disk partition
        self.wal_log.push(entry.clone());

        Ok(entry)
    }

    /// Replay WAL from a specific sequence number during failover / replica recovery
    pub fn replay_from_sequence(&self, start_seq: u64) -> Vec<SequencedLogEntry> {
        self.wal_log
            .iter()
            .filter(|e| e.sequence_number >= start_seq)
            .cloned()
            .collect()
    }
}
