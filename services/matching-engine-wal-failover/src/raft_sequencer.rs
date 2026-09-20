use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RaftRole {
    Leader,
    Follower,
    Candidate,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SequencedLogEntry {
    pub sequence_number: u64,
    pub term: u64,
    pub payload: Vec<u8>,
    pub timestamp_ns: u64,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SequencerSnapshot {
    pub last_included_sequence: u64,
    pub last_included_term: u64,
    pub snapshot_payload: Vec<u8>,
}

pub struct RaftSequencer {
    pub node_id: String,
    pub current_term: u64,
    pub role: RaftRole,
    pub last_sequence: u64,
    pub commit_sequence: u64,
    pub wal_log: Vec<SequencedLogEntry>,
    pub latest_snapshot: Option<SequencerSnapshot>,
}

impl RaftSequencer {
    pub fn new(node_id: String) -> Self {
        Self {
            node_id,
            current_term: 1,
            role: RaftRole::Leader, // Assuming initialized active partition leader
            last_sequence: 0,
            commit_sequence: 0,
            wal_log: Vec::new(),
            latest_snapshot: None,
        }
    }

    /// Append and sequence an incoming order or trade event deterministically.
    /// Only the active Raft leader can assign monotonic sequences.
    pub fn sequence_event(&mut self, payload: Vec<u8>) -> Result<SequencedLogEntry, String> {
        if self.role != RaftRole::Leader {
            return Err("Not current Raft leader; cannot sequence events".into());
        }

        self.last_sequence += 1;
        let now_ns = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_nanos() as u64;

        let entry = SequencedLogEntry {
            sequence_number: self.last_sequence,
            term: self.current_term,
            payload,
            timestamp_ns: now_ns,
        };

        // Durable memory / WAL log
        self.wal_log.push(entry.clone());

        Ok(entry)
    }

    /// Follower append RPC entry from leader
    pub fn append_from_leader(&mut self, entry: SequencedLogEntry) -> Result<(), String> {
        if self.role == RaftRole::Leader {
            return Err("Leader cannot accept append entries as follower".into());
        }

        if entry.sequence_number <= self.last_sequence {
            // Idempotent duplicate check
            return Ok(());
        }

        if entry.sequence_number != self.last_sequence + 1 {
            return Err(format!(
                "Sequence gap detected: expected {}, received {}",
                self.last_sequence + 1,
                entry.sequence_number
            ));
        }

        self.last_sequence = entry.sequence_number;
        self.current_term = self.current_term.max(entry.term);
        self.wal_log.push(entry);

        Ok(())
    }

    /// Mark committed sequence up to commit_seq
    pub fn commit_up_to(&mut self, commit_seq: u64) -> Result<(), String> {
        if commit_seq > self.last_sequence {
            return Err(format!(
                "Cannot commit sequence {} beyond last_sequence {}",
                commit_seq, self.last_sequence
            ));
        }
        if commit_seq > self.commit_sequence {
            self.commit_sequence = commit_seq;
        }
        Ok(())
    }

    /// Failover: Step up to leader for a new term
    pub fn failover_to_leader(&mut self, new_term: u64) -> Result<(), String> {
        if new_term <= self.current_term {
            return Err("New term must be strictly greater than current term".into());
        }
        self.current_term = new_term;
        self.role = RaftRole::Leader;
        Ok(())
    }

    /// Step down to follower
    pub fn step_down_to_follower(&mut self, new_term: u64) {
        if new_term > self.current_term {
            self.current_term = new_term;
        }
        self.role = RaftRole::Follower;
    }

    /// Take a compact snapshot up to committed sequence
    pub fn create_snapshot(&mut self, snapshot_payload: Vec<u8>) -> Result<SequencerSnapshot, String> {
        if self.commit_sequence == 0 {
            return Err("Cannot snapshot empty commit log".into());
        }

        let snapshot = SequencerSnapshot {
            last_included_sequence: self.commit_sequence,
            last_included_term: self.current_term,
            snapshot_payload,
        };

        // Compact WAL log entries up to commit_sequence
        self.wal_log.retain(|e| e.sequence_number > self.commit_sequence);
        self.latest_snapshot = Some(snapshot.clone());

        Ok(snapshot)
    }

    /// Restore state from a leader snapshot
    pub fn restore_from_snapshot(&mut self, snapshot: SequencerSnapshot) {
        self.last_sequence = snapshot.last_included_sequence;
        self.commit_sequence = snapshot.last_included_sequence;
        self.current_term = self.current_term.max(snapshot.last_included_term);
        self.wal_log.clear();
        self.latest_snapshot = Some(snapshot);
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
