use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Side {
    Buy,
    Sell,
}

impl Side {
    #[inline(always)]
    pub fn opposite(&self) -> Self {
        match self {
            Side::Buy => Side::Sell,
            Side::Sell => Side::Buy,
        }
    }
}

impl fmt::Display for Side {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Side::Buy => write!(f, "BUY"),
            Side::Sell => write!(f, "SELL"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum OrderType {
    /// Limit order rests on book if unfilled (Good 'Til Cancelled)
    Limit,
    /// Aggressive order matching best available opposing prices
    Market,
    /// Immediate-Or-Cancel: matches available liquidity, cancels remainder
    ImmediateOrCancel,
    /// Fill-Or-Kill: must fill completely immediately or abort with 0 fills
    FillOrKill,
    /// Post-Only: must only provide liquidity; rejected if it would cross
    PostOnly,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum STPMode {
    /// No self-trade prevention
    None,
    /// Incoming taker order is canceled if crossing with own resting maker
    CancelNewest,
    /// Resting maker order is canceled, incoming taker continues matching
    CancelOldest,
    /// Decrement both orders by overlapping size; cancel smaller, keep remainder of larger
    DecrementAndCancel,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum MatchingAlgorithm {
    /// Strict Price-Time Priority (FIFO)
    Fifo,
    /// Pro-Rata allocation based on resting queue sizes with FIFO remainder
    ProRata,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum OrderStatus {
    New,
    PartiallyFilled,
    Filled,
    Canceled,
    Rejected,
    Expired,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Order {
    pub order_id: u64,
    pub user_id: String,
    pub symbol: String,
    pub side: Side,
    pub order_type: OrderType,
    pub stp_mode: STPMode,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub filled_e8: u64,
    pub timestamp_ns: u64,
}

impl Order {
    pub fn new(
        order_id: u64,
        user_id: impl Into<String>,
        symbol: impl Into<String>,
        side: Side,
        order_type: OrderType,
        stp_mode: STPMode,
        price_e8: u64,
        quantity_e8: u64,
        timestamp_ns: u64,
    ) -> Self {
        Self {
            order_id,
            user_id: user_id.into(),
            symbol: symbol.into(),
            side,
            order_type,
            stp_mode,
            price_e8,
            quantity_e8,
            filled_e8: 0,
            timestamp_ns,
        }
    }

    #[inline(always)]
    pub fn remaining_e8(&self) -> u64 {
        self.quantity_e8.saturating_sub(self.filled_e8)
    }

    #[inline(always)]
    pub fn is_filled(&self) -> bool {
        self.filled_e8 >= self.quantity_e8
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Trade {
    pub trade_id: u64,
    pub symbol: String,
    pub maker_order_id: u64,
    pub taker_order_id: u64,
    pub maker_user_id: String,
    pub taker_user_id: String,
    pub maker_side: Side,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub quote_amount_e8: u64,
    pub maker_fee_e8: u64,
    pub taker_fee_e8: u64,
    pub timestamp_ns: u64,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ExecutionResult {
    pub order_id: u64,
    pub status: OrderStatus,
    pub trades: Vec<Trade>,
    pub resting_order: Option<Order>,
    pub stp_canceled_maker_ids: Vec<u64>,
    pub rejected: bool,
    pub rejection_reason: Option<String>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct PriceLevel {
    pub price_e8: u64,
    pub total_quantity_e8: u64,
    pub order_count: u32,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct L2Snapshot {
    pub symbol: String,
    pub last_update_id: u64,
    pub timestamp_ns: u64,
    pub bids: Vec<PriceLevel>,
    pub asks: Vec<PriceLevel>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct L3Snapshot {
    pub symbol: String,
    pub last_update_id: u64,
    pub timestamp_ns: u64,
    pub bids: Vec<(u64, Vec<Order>)>,
    pub asks: Vec<(u64, Vec<Order>)>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct BookDelta {
    pub symbol: String,
    pub first_update_id: u64,
    pub last_update_id: u64,
    pub timestamp_ns: u64,
    /// Vector of (price_e8, new_total_quantity_e8). 0 quantity means level is removed.
    pub bids: Vec<(u64, u64)>,
    pub asks: Vec<(u64, u64)>,
}
