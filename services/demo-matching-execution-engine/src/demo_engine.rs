//! Isolated Demo CLOB Matching and Virtual Execution Simulator.
//! Simulates institutional-grade Level 2/Level 3 matching for risk-free paper trading.

use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

use crate::clob::OrderBook;
use crate::liquidity_bot::{BotConfig, LiquidityBotManager};
use crate::types::{Level2Snapshot, Match, Order, OrderType};

#[derive(Debug, Clone)]
pub struct VirtualOrder {
    pub order_id: String,
    pub user_id: String,
    pub symbol: String,
    pub side: String, // BUY or SELL
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub order_type: String, // MARKET or LIMIT
}

#[derive(Debug, Clone)]
pub struct VirtualExecution {
    pub execution_id: String,
    pub order_id: String,
    pub user_id: String,
    pub fill_price_e8: u64,
    pub fill_qty_e8: u64,
    pub timestamp_ms: u64,
}

#[derive(Debug, Clone, Default)]
pub struct UserVirtualPosition {
    pub user_id: String,
    pub symbol: String,
    pub base_qty_e8: i64,      // Positive for long, negative for short
    pub quote_balance_e8: i64, // Virtual USDT cash
    pub avg_entry_price_e8: u64,
    pub realized_pnl_e8: i64,
}

pub struct DemoMatchingEngine {
    pub books: HashMap<String, OrderBook>,
    pub mark_prices: HashMap<String, u64>,
    pub last_mark_update_ms: HashMap<String, u64>,
    pub bot_manager: LiquidityBotManager,
    pub user_positions: HashMap<String, HashMap<String, UserVirtualPosition>>, // user -> (symbol -> position)
    pub max_staleness_ms: u64,
}

impl DemoMatchingEngine {
    pub fn new() -> Self {
        Self {
            books: HashMap::new(),
            mark_prices: HashMap::new(),
            last_mark_update_ms: HashMap::new(),
            bot_manager: LiquidityBotManager::new(),
            user_positions: HashMap::new(),
            max_staleness_ms: 2000, // 2000 ms staleness circuit breaker
        }
    }

    /// Register a trading pair symbol with simulated liquidity bots.
    pub fn init_symbol(&mut self, symbol: &str, initial_mark_price_e8: u64) {
        let sym = symbol.to_string();
        let mut book = OrderBook::new(sym.clone());

        let bot_cfg = BotConfig {
            bot_id: format!("bot-mm-{}", symbol.to_lowercase().replace('/', "-")),
            symbol: sym.clone(),
            num_levels: 5,
            spread_bps: 10,       // 0.10%
            level_step_bps: 5,    // 0.05%
            base_qty_e8: 100_000_000, // 1.0 unit
            qty_multiplier: 1.2,
        };

        self.bot_manager.register_bot(bot_cfg);
        let now_ms = current_time_ms();
        self.bot_manager.on_price_update(&mut book, initial_mark_price_e8, now_ms);

        self.books.insert(sym.clone(), book);
        self.mark_prices.insert(sym.clone(), initial_mark_price_e8);
        self.last_mark_update_ms.insert(sym, now_ms);
    }

    /// Update external mark price feed and refresh simulated liquidity quotes.
    pub fn update_mark_price(&mut self, symbol: &str, price_e8: u64) {
        let now_ms = current_time_ms();
        self.mark_prices.insert(symbol.to_string(), price_e8);
        self.last_mark_update_ms.insert(symbol.to_string(), now_ms);

        if let Some(book) = self.books.get_mut(symbol) {
            self.bot_manager.on_price_update(book, price_e8, now_ms);
        }
    }

    /// Staleness check: checks if mirrored market data is current.
    pub fn is_price_stale(&self, symbol: &str) -> bool {
        let now_ms = current_time_ms();
        match self.last_mark_update_ms.get(symbol) {
            Some(&last) => now_ms.saturating_sub(last) > self.max_staleness_ms,
            None => true,
        }
    }

    /// Submit an Order to the simulated CLOB.
    /// Matches against resting quotes (including bot market maker liquidity).
    pub fn submit_order(&mut self, order: Order) -> Result<Vec<Match>, String> {
        if self.is_price_stale(&order.symbol) {
            return Err(format!(
                "Order rejected: Market feed for {} is stale (> {}ms)",
                order.symbol, self.max_staleness_ms
            ));
        }

        let book = self
            .books
            .get_mut(&order.symbol)
            .ok_or_else(|| format!("Symbol {} not found in demo matching engine", order.symbol))?;

        let matches = match order.order_type {
            OrderType::Limit => {
                let (matches, _) = book.place_limit_order(order);
                matches
            }
            OrderType::Market => {
                let (matches, _) = book.place_market_order(order);
                matches
            }
        };

        // Update virtual position balances for retail trader
        for trade in &matches {
            self.update_virtual_position(trade);
        }

        Ok(matches)
    }

    /// Cancel a resting order in the book.
    pub fn cancel_order(&mut self, symbol: &str, order_id: &str) -> Result<Order, String> {
        let book = self
            .books
            .get_mut(symbol)
            .ok_or_else(|| format!("Symbol {} not found in engine", symbol))?;
        book.cancel_order(order_id)
    }

    /// Retrieve Level 2 depth snapshot for a symbol.
    pub fn get_l2_depth(&self, symbol: &str, depth: usize) -> Result<Level2Snapshot, String> {
        let book = self
            .books
            .get(symbol)
            .ok_or_else(|| format!("Symbol {} not found", symbol))?;
        Ok(book.get_l2_depth(depth, current_time_ms()))
    }

    /// Legacy / Instant virtual order execution method (for backward compatibility).
    pub fn execute_virtual_order(&mut self, order: VirtualOrder) -> Result<VirtualExecution, String> {
        let mark_price = match self.mark_prices.get(&order.symbol) {
            Some(&p) => p,
            None => return Err("No mirrored mark price available for symbol".into()),
        };

        let fill_price = if order.order_type == "MARKET" {
            mark_price
        } else {
            // For virtual limit order, check crossing condition
            if order.side == "BUY" && mark_price <= order.price_e8 {
                order.price_e8
            } else if order.side == "SELL" && mark_price >= order.price_e8 {
                order.price_e8
            } else {
                return Err("Resting virtual limit order placed in queue".into());
            }
        };

        let now_ms = current_time_ms();
        let exec = VirtualExecution {
            execution_id: format!("demo-exec-{}", now_ms),
            order_id: order.order_id,
            user_id: order.user_id,
            fill_price_e8: fill_price,
            fill_qty_e8: order.quantity_e8,
            timestamp_ms: now_ms,
        };

        Ok(exec)
    }

    /// Update user position and calculate realized PnL on fills.
    fn update_virtual_position(&mut self, trade: &Match) {
        // Buyer side position update
        if !trade.buyer_id.starts_with("bot-") {
            let pos = self
                .user_positions
                .entry(trade.buyer_id.clone())
                .or_default()
                .entry(trade.symbol.clone())
                .or_insert_with(|| UserVirtualPosition {
                    user_id: trade.buyer_id.clone(),
                    symbol: trade.symbol.clone(),
                    base_qty_e8: 0,
                    quote_balance_e8: 0,
                    avg_entry_price_e8: 0,
                    realized_pnl_e8: 0,
                });

            let cost_e8 = ((trade.price_e8 as u128 * trade.quantity_e8 as u128) / 100_000_000) as i64;
            pos.base_qty_e8 += trade.quantity_e8 as i64;
            pos.quote_balance_e8 -= cost_e8;
            pos.avg_entry_price_e8 = trade.price_e8;
        }

        // Seller side position update
        if !trade.seller_id.starts_with("bot-") {
            let pos = self
                .user_positions
                .entry(trade.seller_id.clone())
                .or_default()
                .entry(trade.symbol.clone())
                .or_insert_with(|| UserVirtualPosition {
                    user_id: trade.seller_id.clone(),
                    symbol: trade.symbol.clone(),
                    base_qty_e8: 0,
                    quote_balance_e8: 0,
                    avg_entry_price_e8: 0,
                    realized_pnl_e8: 0,
                });

            let revenue_e8 = ((trade.price_e8 as u128 * trade.quantity_e8 as u128) / 100_000_000) as i64;
            pos.base_qty_e8 -= trade.quantity_e8 as i64;
            pos.quote_balance_e8 += revenue_e8;
            pos.avg_entry_price_e8 = trade.price_e8;
        }
    }
}

pub fn current_time_ms() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis() as u64
}
