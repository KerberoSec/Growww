pub mod conflation;
pub mod engine;
pub mod fees;
pub mod orderbook;
pub mod partition_manager;
pub mod prorata;
pub mod risk_collar;
pub mod snapshot_delta;
pub mod types;

// Re-export key types for ergonomic usage
pub use engine::MatchingEngine;
pub use fees::FeeEngine;
pub use orderbook::OrderBook;
pub use partition_manager::SymbolPartitionManager;
pub use prorata::ProRataMatcher;
pub use risk_collar::PriceCollar;
pub use snapshot_delta::{ClientBookSync, SyncStatus};
pub use types::{
    BookDelta, ExecutionResult, L2Snapshot, L3Snapshot, MatchingAlgorithm, Order, OrderStatus,
    OrderType, PriceLevel, STPMode, Side, Trade,
};
