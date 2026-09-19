use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum NodeState {
    Leader,
    Follower,
    Candidate,
}

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
}
