use crate::types::BookDelta;
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct ConflationEngine {
    pub symbol: String,
    pub window_ms: u64,
    pub last_flush_ms: u64,
    pub first_update_id: u64,
    pub last_update_id: u64,
    /// Latest quantity per price level: price_e8 -> latest quantity_e8
    pub pending_bids: HashMap<u64, u64>,
    pub pending_asks: HashMap<u64, u64>,
}

impl ConflationEngine {
    pub fn new(symbol: impl Into<String>, window_ms: u64) -> Self {
        Self {
            symbol: symbol.into(),
            window_ms,
            last_flush_ms: 0,
            first_update_id: 0,
            last_update_id: 0,
            pending_bids: HashMap::new(),
            pending_asks: HashMap::new(),
        }
    }

    /// Ingest an incremental delta from the matching engine
    pub fn ingest_delta(&mut self, delta: BookDelta) {
        if self.first_update_id == 0 {
            self.first_update_id = delta.first_update_id;
        }
        self.last_update_id = delta.last_update_id;

        for (price, qty) in delta.bids {
            self.pending_bids.insert(price, qty);
        }
        for (price, qty) in delta.asks {
            self.pending_asks.insert(price, qty);
        }
    }

    /// Check if window elapsed and return conflated delta if updates exist
    pub fn flush_conflated(&mut self, current_time_ms: u64) -> Option<BookDelta> {
        if self.pending_bids.is_empty() && self.pending_asks.is_empty() {
            return None;
        }

        if current_time_ms.saturating_sub(self.last_flush_ms) < self.window_ms && self.last_flush_ms > 0 {
            return None;
        }

        let mut bids_vec: Vec<(u64, u64)> = self.pending_bids.drain().collect();
        let mut asks_vec: Vec<(u64, u64)> = self.pending_asks.drain().collect();

        // Sort bids descending, asks ascending
        bids_vec.sort_by(|a, b| b.0.cmp(&a.0));
        asks_vec.sort_by(|a, b| a.0.cmp(&b.0));

        let conflated = BookDelta {
            symbol: self.symbol.clone(),
            first_update_id: self.first_update_id,
            last_update_id: self.last_update_id,
            timestamp_ns: current_time_ms * 1_000_000,
            bids: bids_vec,
            asks: asks_vec,
        };

        self.last_flush_ms = current_time_ms;
        self.first_update_id = 0;
        self.last_update_id = 0;

        Some(conflated)
    }
}
