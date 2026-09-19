#[derive(Debug, Clone)]
pub enum HSMBackend {
    AWSCloudHSM,
    YubiHSM2,
    SoftwareEnclave,
}

pub struct HSMKeyReference {
    pub key_handle: u32,
    pub backend: HSMBackend,
    pub key_type: String, // e.g. secp256k1 or ed25519
    pub label: String,
}

pub struct HSMOrchestrator {
    pub active_backend: HSMBackend,
    pub keys: Vec<HSMKeyReference>,
}

impl HSMOrchestrator {
    pub fn new(backend: HSMBackend) -> Self {
        Self {
            active_backend: backend,
            keys: Vec::new(),
        }
    }

    pub fn load_root_key(&mut self, handle: u32, label: &str, key_type: &str) -> Result<HSMKeyReference, String> {
        let key_ref = HSMKeyReference {
            key_handle: handle,
            backend: self.active_backend.clone(),
            key_type: key_type.to_string(),
            label: label.to_string(),
        };
        self.keys.push(key_ref.clone());
        Ok(key_ref)
    }

    /// Sign digest using hardware-isolated root private key
    pub fn sign_digest(&self, handle: u32, digest: &[u8; 32]) -> Result<Vec<u8>, String> {
        if !self.keys.iter().any(|k| k.key_handle == handle) {
            return Err("HSM key handle not recognized".into());
        }

        // In production, execute PKCS#11 or YubiHSM connector command:
        // yubihsm_sign_ecdsa(session, handle, digest)
        // Here we simulate hardware signing response
        let mut simulated_signature = Vec::with_capacity(65);
        simulated_signature.extend_from_slice(digest);
        simulated_signature.extend_from_slice(&[1u8; 33]);
        Ok(simulated_signature)
    }
}
