//! Simulated Liquidity Provisioning Bots for Demo Trading CLOB.
//! Generates synthetic depth, tight bid-ask spreads, and realistic order fills
//! against retail demo accounts without accessing production capital.

use std::collections::HashMap;
use crate::clob::OrderBook;
use crate::types::{Order, OrderType, Side};

#[derive(Debug, Clone)]
pub struct BotConfig {
    pub bot_id: String,
    pub symbol: String,
    pub num_levels: usize,       // Number of bid/ask levels (e.g. 5)
    pub spread_bps: u64,         // Initial half spread in basis points (e.g. 10 = 0.10%)
    pub level_step_bps: u64,     // Increment between subsequent levels (e.g. 5 bps)
    pub base_qty_e8: u64,        // Base quantity per level (e.g. 1 BTC = 1e8)
    pub qty_multiplier: f64,     // Multiplier for deeper levels
}

impl Default for BotConfig {
    fn default() -> Self {
        Self {
            bot_id: "demo-liquidity-bot-system".to_string(),
            symbol: "BTC/USDT".to_string(),
            num_levels: 5,
            spread_bps: 10,       // 0.10%
            level_step_bps: 5,    // 0.05% per step
            base_qty_e8: 50_000_000, // 0.50 BTC
            qty_multiplier: 1.25,
        }
    }
}

pub struct LiquidityBot {
    pub config: BotConfig,
    order_seq: u64,
}

impl LiquidityBot {
    pub fn new(config: BotConfig) -> Self {
        Self {
            config,
            order_seq: 0,
        }
    }

    /// Refresh all quotes around a new mark price.
    /// Cancels existing resting orders of this bot and places fresh tiered bids and asks.
    pub fn refresh_quotes(&mut self, book: &mut OrderBook, mark_price_e8: u64, timestamp_ms: u64) {
        // First cancel existing orders placed by this bot
        book.cancel_bot_orders(&self.config.bot_id);

        if mark_price_e8 == 0 {
            return;
        }

        // Generate Asks (above mark price)
        for i in 0..self.config.num_levels {
            let offset_bps = self.config.spread_bps + (i as u64 * self.config.level_step_bps);
            let price_delta = (mark_price_e8 * offset_bps) / 10_000;
            let ask_price = mark_price_e8.saturating_add(price_delta);

            let qty_scale = self.config.qty_multiplier.powi(i as i32);
            let qty = (self.config.base_qty_e8 as f64 * qty_scale) as u64;

            self.order_seq += 1;
            let order = Order::new(
                format!("{}-ask-{}-{}", self.config.bot_id, self.order_seq, i),
                self.config.bot_id.clone(),
                self.config.symbol.clone(),
                Side::Sell,
                OrderType::Limit,
                ask_price,
                qty,
                timestamp_ms,
                true,
            );

            book.place_limit_order(order);
        }

        // Generate Bids (below mark price)
        for i in 0..self.config.num_levels {
            let offset_bps = self.config.spread_bps + (i as u64 * self.config.level_step_bps);
            let price_delta = (mark_price_e8 * offset_bps) / 10_000;
            let bid_price = mark_price_e8.saturating_sub(price_delta);

            if bid_price == 0 {
                continue;
            }

            let qty_scale = self.config.qty_multiplier.powi(i as i32);
            let qty = (self.config.base_qty_e8 as f64 * qty_scale) as u64;

            self.order_seq += 1;
            let order = Order::new(
                format!("{}-bid-{}-{}", self.config.bot_id, self.order_seq, i),
                self.config.bot_id.clone(),
                self.config.symbol.clone(),
                Side::Buy,
                OrderType::Limit,
                bid_price,
                qty,
                timestamp_ms,
                true,
            );

            book.place_limit_order(order);
        }
    }

    /// Simulate a noise trader placing an aggressive market order to simulate real tick activity.
    pub fn inject_noise_order(
        &mut self,
        book: &mut OrderBook,
        side: Side,
        qty_e8: u64,
        timestamp_ms: u64,
    ) -> Vec<crate::types::Match> {
        self.order_seq += 1;
        let order = Order::new(
            format!("{}-noise-{}", self.config.bot_id, self.order_seq),
            format!("{}-noise", self.config.bot_id),
            self.config.symbol.clone(),
            side,
            OrderType::Market,
            0,
            qty_e8,
            timestamp_ms,
            true,
        );

        let (matches, _) = book.place_market_order(order);
        matches
    }
}

pub struct LiquidityBotManager {
    bots: HashMap<String, LiquidityBot>,
}

impl LiquidityBotManager {
    pub fn new() -> Self {
        Self {
            bots: HashMap::new(),
        }
    }

    pub fn register_bot(&mut self, config: BotConfig) {
        let symbol = config.symbol.clone();
        self.bots.insert(symbol, LiquidityBot::new(config));
    }

    pub fn on_price_update(&mut self, book: &mut OrderBook, mark_price_e8: u64, timestamp_ms: u64) {
        if let Some(bot) = self.bots.get_mut(&book.symbol) {
            bot.refresh_quotes(book, mark_price_e8, timestamp_ms);
        }
    }
}
