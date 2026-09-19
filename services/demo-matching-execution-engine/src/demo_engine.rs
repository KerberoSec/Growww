use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

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

pub struct DemoMatchingEngine {
    pub virtual_positions: HashMap<String, HashMap<String, u64>>, // user -> (asset -> balance)
    pub mark_prices: HashMap<String, u64>,
}

impl DemoMatchingEngine {
    pub fn new() -> Self {
        Self {
            virtual_positions: HashMap::new(),
            mark_prices: HashMap::new(),
        }
    }

    pub fn update_mark_price(&mut self, symbol: &str, price_e8: u64) {
        self.mark_prices.insert(symbol.to_string(), price_e8);
    }

    /// Execute virtual paper trade against mirrored real mark price instantly
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

        let now_ms = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_millis() as u64;

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
}
