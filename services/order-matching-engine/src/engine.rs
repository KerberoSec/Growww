use crate::fees::FeeEngine;
use crate::orderbook::OrderBook;
use crate::partition_manager::SymbolPartitionManager;
use crate::types::{BookDelta, ExecutionResult, L2Snapshot, L3Snapshot, MatchingAlgorithm, Order};
use std::collections::HashMap;

pub struct MatchingEngine {
    pub partition_manager: SymbolPartitionManager,
    pub books: HashMap<String, OrderBook>,
}

impl MatchingEngine {
    pub fn new(num_partitions: u32) -> Self {
        Self {
            partition_manager: SymbolPartitionManager::new(num_partitions),
            books: HashMap::new(),
        }
    }

    pub fn register_symbol(
        &mut self,
        symbol: &str,
        algorithm: MatchingAlgorithm,
        reference_price_e8: u64,
        collar_bps: u32,
    ) {
        let book = OrderBook::with_config(
            symbol.to_string(),
            algorithm,
            reference_price_e8,
            collar_bps,
        );
        self.books.insert(symbol.to_string(), book);
    }

    pub fn get_or_create_book(&mut self, symbol: &str) -> &mut OrderBook {
        self.books
            .entry(symbol.to_string())
            .or_insert_with(|| OrderBook::new(symbol.to_string()))
    }

    pub fn submit_order(&mut self, order: Order) -> ExecutionResult {
        let symbol = order.symbol.clone();
        let book = self.get_or_create_book(&symbol);
        book.process_order(order)
    }

    pub fn cancel_order(&mut self, symbol: &str, order_id: u64) -> Option<Order> {
        self.books.get_mut(symbol).and_then(|b| b.cancel_order(order_id))
    }

    pub fn cancel_replace(
        &mut self,
        symbol: &str,
        order_id: u64,
        new_price_e8: u64,
        new_quantity_e8: u64,
        timestamp_ns: u64,
    ) -> Result<(Option<Order>, ExecutionResult), String> {
        let book = self
            .books
            .get_mut(symbol)
            .ok_or_else(|| format!("Symbol {} not found", symbol))?;
        book.cancel_replace_order(order_id, new_price_e8, new_quantity_e8, timestamp_ns)
    }

    pub fn get_l2_snapshot(&self, symbol: &str, depth: usize) -> Option<L2Snapshot> {
        self.books.get(symbol).map(|b| b.generate_l2_snapshot(depth))
    }

    pub fn get_l3_snapshot(&self, symbol: &str) -> Option<L3Snapshot> {
        self.books.get(symbol).map(|b| b.generate_l3_snapshot())
    }

    pub fn drain_deltas(&mut self, symbol: &str) -> Vec<BookDelta> {
        self.books
            .get_mut(symbol)
            .map(|b| b.drain_deltas())
            .unwrap_or_default()
    }

    pub fn hot_reload_fees(maker_bps: u16, taker_bps: u16) {
        FeeEngine::set_maker_fee_bps(maker_bps);
        FeeEngine::set_taker_fee_bps(taker_bps);
    }
}
