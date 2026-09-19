//! Demo Matching Execution Engine
//! Sovereign CLOB matching simulator with simulated liquidity bots for paper trading.

pub mod clob;
pub mod demo_analytics;
pub mod demo_engine;
pub mod liquidity_bot;
pub mod types;

pub use clob::OrderBook;
pub use demo_analytics::{DemoAnalyticsEngine, TraderPnLSummary};
pub use demo_engine::{DemoMatchingEngine, UserVirtualPosition, VirtualExecution, VirtualOrder};
pub use liquidity_bot::{BotConfig, LiquidityBot, LiquidityBotManager};
pub use types::{Level2Quote, Level2Snapshot, Match, Order, OrderStatus, OrderType, Side};
