use crate::types::{BookDelta, L2Snapshot, PriceLevel};
use std::collections::{BTreeMap, VecDeque};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SyncStatus {
    BufferingDeltas,
    Synchronized,
    Desynchronized,
}

#[derive(Debug, Clone)]
pub struct ClientBookSync {
    pub symbol: String,
    pub status: SyncStatus,
    pub last_update_id: u64,
    // Bids descending (highest price first)
    pub bids: BTreeMap<u64, u64>,
    // Asks ascending (lowest price first)
    pub asks: BTreeMap<u64, u64>,
    pub delta_buffer: VecDeque<BookDelta>,
    pub max_buffer_size: usize,
}

impl ClientBookSync {
    pub fn new(symbol: impl Into<String>) -> Self {
        Self {
            symbol: symbol.into(),
            status: SyncStatus::BufferingDeltas,
            last_update_id: 0,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            delta_buffer: VecDeque::new(),
            max_buffer_size: 10_000,
        }
    }

    /// Step 1: Buffer incoming deltas before/during snapshot fetch
    pub fn buffer_delta(&mut self, delta: BookDelta) {
        if self.status == SyncStatus::Synchronized {
            let _ = self.apply_delta(&delta);
        } else {
            if self.delta_buffer.len() >= self.max_buffer_size {
                self.delta_buffer.pop_front();
            }
            self.delta_buffer.push_back(delta);
        }
    }

    /// Step 2: Apply authoritative snapshot and replay buffered deltas
    pub fn apply_snapshot(&mut self, snapshot: L2Snapshot) -> Result<(), String> {
        if snapshot.symbol != self.symbol {
            return Err("Symbol mismatch in snapshot".into());
        }

        self.bids.clear();
        self.asks.clear();

        for level in snapshot.bids {
            if level.total_quantity_e8 > 0 {
                self.bids.insert(level.price_e8, level.total_quantity_e8);
            }
        }

        for level in snapshot.asks {
            if level.total_quantity_e8 > 0 {
                self.asks.insert(level.price_e8, level.total_quantity_e8);
            }
        }

        self.last_update_id = snapshot.last_update_id;

        // Drain buffered deltas, discarding any whose last_update_id <= snapshot.last_update_id
        let mut first_applicable_found = false;
        while let Some(delta) = self.delta_buffer.pop_front() {
            if delta.last_update_id <= self.last_update_id {
                continue;
            }

            if !first_applicable_found {
                // First delta must bridge or cover snapshot.last_update_id + 1
                if delta.first_update_id > self.last_update_id + 1 {
                    self.status = SyncStatus::Desynchronized;
                    return Err(format!(
                        "Gap detected after snapshot: snapshot lastUpdateId={}, delta firstUpdateId={}",
                        self.last_update_id, delta.first_update_id
                    ));
                }
                first_applicable_found = true;
            }

            self.apply_delta_internal(&delta)?;
        }

        self.status = SyncStatus::Synchronized;
        Ok(())
    }

    /// Step 3: Apply single delta in Synchronized mode
    pub fn apply_delta(&mut self, delta: &BookDelta) -> Result<(), String> {
        if self.status != SyncStatus::Synchronized {
            return Err("Cannot apply delta directly: client not synchronized".into());
        }

        // Sequence validation: delta.first_update_id must equal last_update_id + 1
        if delta.first_update_id != self.last_update_id + 1 {
            self.status = SyncStatus::Desynchronized;
            return Err(format!(
                "Sequence gap detected: expected {}, got {}. Triggering resync.",
                self.last_update_id + 1,
                delta.first_update_id
            ));
        }

        self.apply_delta_internal(delta)?;
        Ok(())
    }

    fn apply_delta_internal(&mut self, delta: &BookDelta) -> Result<(), String> {
        // Apply bids
        for &(price, qty) in &delta.bids {
            if qty == 0 {
                // Explicit zero clears price level (zero ghost liquidity)
                self.bids.remove(&price);
            } else {
                self.bids.insert(price, qty);
            }
        }

        // Apply asks
        for &(price, qty) in &delta.asks {
            if qty == 0 {
                // Explicit zero clears price level (zero ghost liquidity)
                self.asks.remove(&price);
            } else {
                self.asks.insert(price, qty);
            }
        }

        self.last_update_id = delta.last_update_id;
        Ok(())
    }

    pub fn get_top_levels(&self, depth: usize) -> (Vec<PriceLevel>, Vec<PriceLevel>) {
        let bids: Vec<PriceLevel> = self
            .bids
            .iter()
            .rev()
            .take(depth)
            .map(|(&price, &qty)| PriceLevel {
                price_e8: price,
                total_quantity_e8: qty,
                order_count: 1,
            })
            .collect();

        let asks: Vec<PriceLevel> = self
            .asks
            .iter()
            .take(depth)
            .map(|(&price, &qty)| PriceLevel {
                price_e8: price,
                total_quantity_e8: qty,
                order_count: 1,
            })
            .collect();

        (bids, asks)
    }
}
