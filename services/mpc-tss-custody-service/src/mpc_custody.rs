use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct KeyShare {
    pub party_id: u32,
    pub share_commitment: Vec<u8>,
    pub threshold: u32,
    pub total_parties: u32,
}

#[derive(Debug, Clone)]
pub struct SigningSession {
    pub session_id: String,
    pub message_digest: [u8; 32],
    pub participating_parties: Vec<u32>,
    pub partial_signatures: HashMap<u32, Vec<u8>>,
    pub threshold: u32,
}

pub struct MPCCustodyCoordinator {
    pub key_shares: HashMap<u32, KeyShare>,
    pub sessions: HashMap<String, SigningSession>,
    pub threshold: u32,
    pub total_parties: u32,
}

impl MPCCustodyCoordinator {
    pub fn new(threshold: u32, total_parties: u32) -> Self {
        Self {
            key_shares: HashMap::new(),
            sessions: HashMap::new(),
            threshold,
            total_parties,
        }
    }

    pub fn start_signing_session(&mut self, session_id: String, digest: [u8; 32], parties: Vec<u32>) -> Result<(), String> {
        if parties.len() < self.threshold as usize {
            return Err("Insufficient parties to satisfy threshold quorum".into());
        }

        let session = SigningSession {
            session_id: session_id.clone(),
            message_digest: digest,
            participating_parties: parties,
            partial_signatures: HashMap::new(),
            threshold: self.threshold,
        };

        self.sessions.insert(session_id, session);
        Ok(())
    }

    pub fn submit_partial_signature(&mut self, session_id: &str, party_id: u32, partial_sig: Vec<u8>) -> Result<bool, String> {
        let session = match self.sessions.get_mut(session_id) {
            Some(s) => s,
            None => return Err("Signing session not found".into()),
        };

        session.partial_signatures.insert(party_id, partial_sig);

        // Check if threshold is reached
        Ok(session.partial_signatures.len() >= session.threshold as usize)
    }
}
