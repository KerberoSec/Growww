//! Core data types for the Demo CLOB Matching and Execution Engine.

use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Side {
    Buy,
    Sell,
}

impl fmt::Display for Side {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Side::Buy => write!(f, "BUY"),
            Side::Sell => write!(f, "SELL"),
        }
    }
}

impl Side {
    pub fn opposite(&self) -> Self {
        match self {
            Side::Buy => Side::Sell,
            Side::Sell => Side::Buy,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OrderType {
    Limit,
    Market,
}

impl fmt::Display for OrderType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            OrderType::Limit => write!(f, "LIMIT"),
            OrderType::Market => write!(f, "MARKET"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OrderStatus {
    Pending,
    Open,
    PartiallyFilled,
    Filled,
    Cancelled,
    Rejected,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Order {
    pub order_id: String,
    pub user_id: String,
    pub symbol: String,
    pub side: Side,
    pub order_type: OrderType,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub filled_qty_e8: u64,
    pub status: OrderStatus,
    pub timestamp_ms: u64,
    pub is_bot: bool,
}

impl Order {
    pub fn new(
        order_id: String,
        user_id: String,
        symbol: String,
        side: Side,
        order_type: OrderType,
        price_e8: u64,
        quantity_e8: u64,
        timestamp_ms: u64,
        is_bot: bool,
    ) -> Self {
        Self {
            order_id,
            user_id,
            symbol,
            side,
            order_type,
            price_e8,
            quantity_e8,
            filled_qty_e8: 0,
            status: OrderStatus::Open,
            timestamp_ms,
            is_bot,
        }
    }

    pub fn remaining_qty_e8(&self) -> u64 {
        self.quantity_e8.saturating_sub(self.filled_qty_e8)
    }

    pub fn is_filled(&self) -> bool {
        self.filled_qty_e8 >= self.quantity_e8
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Match {
    pub trade_id: String,
    pub buy_order_id: String,
    pub sell_order_id: String,
    pub buyer_id: String,
    pub seller_id: String,
    pub symbol: String,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub timestamp_ms: u64,
    pub taker_side: Side,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Level2Quote {
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub order_count: usize,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Level2Snapshot {
    pub symbol: String,
    pub timestamp_ms: u64,
    pub bids: Vec<Level2Quote>,
    pub asks: Vec<Level2Quote>,
}
