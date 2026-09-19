use std::collections::hash_map::DefaultHasher;
use std::hash::{Hash, Hasher};

pub struct SymbolPartitionManager {
    pub total_partitions: u32,
}

impl SymbolPartitionManager {
    pub fn new(num_partitions: u32) -> Self {
        Self {
            total_partitions: num_partitions,
        }
    }

    /// Computes deterministic partition ID for symbol to ensure zero-lock cross-symbol matching
    pub fn route_symbol(&self, symbol: &str) -> u32 {
        let mut hasher = DefaultHasher::new();
        symbol.hash(&mut hasher);
        (hasher.finish() % self.total_partitions as u64) as u32
    }
}
