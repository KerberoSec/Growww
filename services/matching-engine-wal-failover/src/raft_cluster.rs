use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum NodeState {
    Leader,
    Follower,
    Candidate,
}

#[derive(Debug, Clone)]
pub struct ClusterPeer {
    pub node_id: u32,
    pub address: String,
    pub next_index: u64,
    pub match_index: u64,
}

pub struct RaftClusterTopology {
    pub current_term: u64,
    pub node_id: u32,
    pub state: NodeState,
    pub peers: HashMap<u32, ClusterPeer>,
    pub commit_index: u64,
    pub voted_for: Option<u32>,
}

impl RaftClusterTopology {
    pub fn new(node_id: u32, peer_addresses: Vec<(u32, String)>) -> Self {
        let mut peers = HashMap::new();
        for (id, addr) in peer_addresses {
            peers.insert(
                id,
                ClusterPeer {
                    node_id: id,
                    address: addr,
                    next_index: 1,
                    match_index: 0,
                },
            );
        }

        Self {
            current_term: 1,
            node_id,
            state: NodeState::Follower,
            peers,
            commit_index: 0,
            voted_for: None,
        }
    }

    /// Determines if a quorum of cluster nodes have replicated a sequence index
    pub fn is_quorum_replicated(&self, sequence: u64) -> bool {
        let mut count = 1; // Self count
        for peer in self.peers.values() {
            if peer.match_index >= sequence {
                count += 1;
            }
        }
        let total_nodes = self.peers.len() + 1;
        count > total_nodes / 2
    }

    /// Transition to candidate on election timeout
    pub fn become_candidate(&mut self) {
        self.current_term += 1;
        self.state = NodeState::Candidate;
        self.voted_for = Some(self.node_id);
    }

    /// Transition to leader once majority votes are granted
    pub fn become_leader(&mut self) -> Result<(), String> {
        if self.state != NodeState::Candidate {
            return Err("Must be in candidate state to transition to leader".into());
        }
        self.state = NodeState::Leader;
        // Initialize next_index for peers to commit_index + 1
        for peer in self.peers.values_mut() {
            peer.next_index = self.commit_index + 1;
        }
        Ok(())
    }

    /// Step down to follower upon observing a higher term or valid leader heartbeat
    pub fn step_down(&mut self, term: u64) {
        if term > self.current_term {
            self.current_term = term;
            self.voted_for = None;
        }
        self.state = NodeState::Follower;
    }

    /// Update peer replication match index
    pub fn update_peer_match(&mut self, peer_id: u32, match_index: u64) -> Result<(), String> {
        if let Some(peer) = self.peers.get_mut(&peer_id) {
            peer.match_index = match_index;
            peer.next_index = match_index + 1;
            Ok(())
        } else {
            Err(format!("Peer node {} not found in topology", peer_id))
        }
    }

    /// Recompute commit_index based on median match_index of majority nodes
    pub fn recompute_commit_index(&mut self) -> u64 {
        let mut indices: Vec<u64> = self.peers.values().map(|p| p.match_index).collect();
        // Include self commit/match index
        indices.push(self.commit_index);
        indices.sort_unstable();

        // The median index in sorted order represents the highest quorum-agreed index
        let median_idx = indices.len() / 2;
        let quorum_agreed = indices[median_idx];

        if quorum_agreed > self.commit_index {
            self.commit_index = quorum_agreed;
        }

        self.commit_index
    }
}
