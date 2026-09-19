//! Central Limit Order Book (CLOB) implementation for simulated demo trading.
//! Enforces Price-Time Priority (FIFO), Self-Trade Prevention, and full L2 depth aggregation.

use std::collections::{BTreeMap, HashMap, VecDeque};
use crate::types::{Level2Quote, Level2Snapshot, Match, Order, OrderStatus, OrderType, Side};

pub struct OrderBook {
    pub symbol: String,
    // Bids stored by price descending. In BTreeMap, keys are sorted ascending, so we can iterate in reverse.
    bids: BTreeMap<u64, VecDeque<Order>>,
    // Asks stored by price ascending.
    asks: BTreeMap<u64, VecDeque<Order>>,
    // Index: order_id -> (Side, price_e8)
    order_index: HashMap<String, (Side, u64)>,
    match_sequence: u64,
}

impl OrderBook {
    pub fn new(symbol: String) -> Self {
        Self {
            symbol,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            order_index: HashMap::new(),
            match_sequence: 0,
        }
    }

    /// Returns the highest bid price if available.
    pub fn best_bid(&self) -> Option<u64> {
        self.bids.keys().next_back().copied()
    }

    /// Returns the lowest ask price if available.
    pub fn best_ask(&self) -> Option<u64> {
        self.asks.keys().next().copied()
    }

    /// Returns the mid price in e8 units.
    pub fn mid_price(&self) -> Option<u64> {
        match (self.best_bid(), self.best_ask()) {
            (Some(bid), Some(ask)) => Some((bid + ask) / 2),
            (Some(bid), None) => Some(bid),
            (None, Some(ask)) => Some(ask),
            (None, None) => None,
        }
    }

    /// Returns current spread in e8 units.
    pub fn spread(&self) -> Option<u64> {
        match (self.best_bid(), self.best_ask()) {
            (Some(bid), Some(ask)) if ask >= bid => Some(ask - bid),
            _ => None,
        }
    }

    /// Place a Limit Order into the CLOB.
    /// Matches crossing prices immediately (FIFO at each price level).
    /// If quantity remains, places the remainder in the resting queue.
    pub fn place_limit_order(&mut self, mut order: Order) -> (Vec<Match>, Option<Order>) {
        if order.quantity_e8 == 0 || order.price_e8 == 0 {
            order.status = OrderStatus::Rejected;
            return (vec![], Some(order));
        }

        let mut matches = Vec::new();

        match order.side {
            Side::Buy => {
                // Match against asks where ask_price <= order.price_e8
                let mut prices_to_remove = Vec::new();

                for (&ask_price, queue) in self.asks.iter_mut() {
                    if ask_price > order.price_e8 || order.is_filled() {
                        break;
                    }

                    let mut filled_in_level = 0;
                    for resting_order in queue.iter_mut() {
                        if order.is_filled() {
                            break;
                        }

                        // Self-Trade Prevention: Skip if same user
                        if resting_order.user_id == order.user_id {
                            continue;
                        }

                        let match_qty = std::cmp::min(order.remaining_qty_e8(), resting_order.remaining_qty_e8());
                        if match_qty == 0 {
                            continue;
                        }

                        order.filled_qty_e8 += match_qty;
                        resting_order.filled_qty_e8 += match_qty;

                        self.match_sequence += 1;
                        let trade = Match {
                            trade_id: format!("TRD-{}-{:08}", self.symbol, self.match_sequence),
                            buy_order_id: order.order_id.clone(),
                            sell_order_id: resting_order.order_id.clone(),
                            buyer_id: order.user_id.clone(),
                            seller_id: resting_order.user_id.clone(),
                            symbol: self.symbol.clone(),
                            price_e8: ask_price, // Passive maker price
                            quantity_e8: match_qty,
                            timestamp_ms: order.timestamp_ms,
                            taker_side: Side::Buy,
                        };
                        matches.push(trade);

                        if resting_order.is_filled() {
                            resting_order.status = OrderStatus::Filled;
                            self.order_index.remove(&resting_order.order_id);
                            filled_in_level += 1;
                        } else {
                            resting_order.status = OrderStatus::PartiallyFilled;
                        }
                    }

                    // Clean up filled orders from queue
                    if filled_in_level > 0 {
                        queue.retain(|o| !o.is_filled());
                    }
                    if queue.is_empty() {
                        prices_to_remove.push(ask_price);
                    }
                }

                for p in prices_to_remove {
                    self.asks.remove(&p);
                }

                if order.is_filled() {
                    order.status = OrderStatus::Filled;
                    (matches, None)
                } else {
                    if order.filled_qty_e8 > 0 {
                        order.status = OrderStatus::PartiallyFilled;
                    } else {
                        order.status = OrderStatus::Open;
                    }
                    self.order_index.insert(order.order_id.clone(), (Side::Buy, order.price_e8));
                    let resting = order.clone();
                    self.bids.entry(order.price_e8).or_default().push_back(order);
                    (matches, Some(resting))
                }
            }
            Side::Sell => {
                // Match against bids where bid_price >= order.price_e8 (highest first)
                let mut prices_to_remove = Vec::new();
                let bid_prices: Vec<u64> = self.bids.keys().rev().copied().collect();

                for bid_price in bid_prices {
                    if bid_price < order.price_e8 || order.is_filled() {
                        break;
                    }

                    if let Some(queue) = self.bids.get_mut(&bid_price) {
                        let mut filled_in_level = 0;
                        for resting_order in queue.iter_mut() {
                            if order.is_filled() {
                                break;
                            }

                            // Self-trade prevention
                            if resting_order.user_id == order.user_id {
                                continue;
                            }

                            let match_qty = std::cmp::min(order.remaining_qty_e8(), resting_order.remaining_qty_e8());
                            if match_qty == 0 {
                                continue;
                            }

                            order.filled_qty_e8 += match_qty;
                            resting_order.filled_qty_e8 += match_qty;

                            self.match_sequence += 1;
                            let trade = Match {
                                trade_id: format!("TRD-{}-{:08}", self.symbol, self.match_sequence),
                                buy_order_id: resting_order.order_id.clone(),
                                sell_order_id: order.order_id.clone(),
                                buyer_id: resting_order.user_id.clone(),
                                seller_id: order.user_id.clone(),
                                symbol: self.symbol.clone(),
                                price_e8: bid_price, // Passive maker price
                                quantity_e8: match_qty,
                                timestamp_ms: order.timestamp_ms,
                                taker_side: Side::Sell,
                            };
                            matches.push(trade);

                            if resting_order.is_filled() {
                                resting_order.status = OrderStatus::Filled;
                                self.order_index.remove(&resting_order.order_id);
                                filled_in_level += 1;
                            } else {
                                resting_order.status = OrderStatus::PartiallyFilled;
                            }
                        }

                        if filled_in_level > 0 {
                            queue.retain(|o| !o.is_filled());
                        }
                        if queue.is_empty() {
                            prices_to_remove.push(bid_price);
                        }
                    }
                }

                for p in prices_to_remove {
                    self.bids.remove(&p);
                }

                if order.is_filled() {
                    order.status = OrderStatus::Filled;
                    (matches, None)
                } else {
                    if order.filled_qty_e8 > 0 {
                        order.status = OrderStatus::PartiallyFilled;
                    } else {
                        order.status = OrderStatus::Open;
                    }
                    self.order_index.insert(order.order_id.clone(), (Side::Sell, order.price_e8));
                    let resting = order.clone();
                    self.asks.entry(order.price_e8).or_default().push_back(order);
                    (matches, Some(resting))
                }
            }
        }
    }

    /// Place a Market Order into the CLOB.
    /// Matches aggressively against resting opposite orders.
    /// Returns all generated matches and unfilled quantity (if book ran out of depth).
    pub fn place_market_order(&mut self, mut order: Order) -> (Vec<Match>, u64) {
        order.order_type = OrderType::Market;
        let mut matches = Vec::new();

        match order.side {
            Side::Buy => {
                let mut prices_to_remove = Vec::new();

                for (&ask_price, queue) in self.asks.iter_mut() {
                    if order.is_filled() {
                        break;
                    }

                    let mut filled_in_level = 0;
                    for resting_order in queue.iter_mut() {
                        if order.is_filled() {
                            break;
                        }

                        if resting_order.user_id == order.user_id {
                            continue;
                        }

                        let match_qty = std::cmp::min(order.remaining_qty_e8(), resting_order.remaining_qty_e8());
                        if match_qty == 0 {
                            continue;
                        }

                        order.filled_qty_e8 += match_qty;
                        resting_order.filled_qty_e8 += match_qty;

                        self.match_sequence += 1;
                        let trade = Match {
                            trade_id: format!("TRD-{}-{:08}", self.symbol, self.match_sequence),
                            buy_order_id: order.order_id.clone(),
                            sell_order_id: resting_order.order_id.clone(),
                            buyer_id: order.user_id.clone(),
                            seller_id: resting_order.user_id.clone(),
                            symbol: self.symbol.clone(),
                            price_e8: ask_price,
                            quantity_e8: match_qty,
                            timestamp_ms: order.timestamp_ms,
                            taker_side: Side::Buy,
                        };
                        matches.push(trade);

                        if resting_order.is_filled() {
                            resting_order.status = OrderStatus::Filled;
                            self.order_index.remove(&resting_order.order_id);
                            filled_in_level += 1;
                        } else {
                            resting_order.status = OrderStatus::PartiallyFilled;
                        }
                    }

                    if filled_in_level > 0 {
                        queue.retain(|o| !o.is_filled());
                    }
                    if queue.is_empty() {
                        prices_to_remove.push(ask_price);
                    }
                }

                for p in prices_to_remove {
                    self.asks.remove(&p);
                }
            }
            Side::Sell => {
                let mut prices_to_remove = Vec::new();
                let bid_prices: Vec<u64> = self.bids.keys().rev().copied().collect();

                for bid_price in bid_prices {
                    if order.is_filled() {
                        break;
                    }

                    if let Some(queue) = self.bids.get_mut(&bid_price) {
                        let mut filled_in_level = 0;
                        for resting_order in queue.iter_mut() {
                            if order.is_filled() {
                                break;
                            }

                            if resting_order.user_id == order.user_id {
                                continue;
                            }

                            let match_qty = std::cmp::min(order.remaining_qty_e8(), resting_order.remaining_qty_e8());
                            if match_qty == 0 {
                                continue;
                            }

                            order.filled_qty_e8 += match_qty;
                            resting_order.filled_qty_e8 += match_qty;

                            self.match_sequence += 1;
                            let trade = Match {
                                trade_id: format!("TRD-{}-{:08}", self.symbol, self.match_sequence),
                                buy_order_id: resting_order.order_id.clone(),
                                sell_order_id: order.order_id.clone(),
                                buyer_id: resting_order.user_id.clone(),
                                seller_id: order.user_id.clone(),
                                symbol: self.symbol.clone(),
                                price_e8: bid_price,
                                quantity_e8: match_qty,
                                timestamp_ms: order.timestamp_ms,
                                taker_side: Side::Sell,
                            };
                            matches.push(trade);

                            if resting_order.is_filled() {
                                resting_order.status = OrderStatus::Filled;
                                self.order_index.remove(&resting_order.order_id);
                                filled_in_level += 1;
                            } else {
                                resting_order.status = OrderStatus::PartiallyFilled;
                            }
                        }

                        if filled_in_level > 0 {
                            queue.retain(|o| !o.is_filled());
                        }
                        if queue.is_empty() {
                            prices_to_remove.push(bid_price);
                        }
                    }
                }

                for p in prices_to_remove {
                    self.bids.remove(&p);
                }
            }
        }

        let unfilled = order.remaining_qty_e8();
        (matches, unfilled)
    }

    /// Cancel an active resting order by ID.
    pub fn cancel_order(&mut self, order_id: &str) -> Result<Order, String> {
        let (side, price_e8) = self
            .order_index
            .remove(order_id)
            .ok_or_else(|| format!("Order {} not found in active book", order_id))?;

        let queue_opt = match side {
            Side::Buy => self.bids.get_mut(&price_e8),
            Side::Sell => self.asks.get_mut(&price_e8),
        };

        if let Some(queue) = queue_opt {
            if let Some(pos) = queue.iter().position(|o| o.order_id == order_id) {
                let mut order = queue.remove(pos).unwrap();
                order.status = OrderStatus::Cancelled;

                if queue.is_empty() {
                    match side {
                        Side::Buy => {
                            self.bids.remove(&price_e8);
                        }
                        Side::Sell => {
                            self.asks.remove(&price_e8);
                        }
                    }
                }
                return Ok(order);
            }
        }

        Err(format!("Order {} internal index mismatch", order_id))
    }

    /// Aggregate depth snapshot up to `depth_limit` levels.
    pub fn get_l2_depth(&self, depth_limit: usize, timestamp_ms: u64) -> Level2Snapshot {
        let mut bids = Vec::with_capacity(depth_limit);
        for (&price_e8, queue) in self.bids.iter().rev().take(depth_limit) {
            let total_qty: u64 = queue.iter().map(|o| o.remaining_qty_e8()).sum();
            bids.push(Level2Quote {
                price_e8,
                quantity_e8: total_qty,
                order_count: queue.len(),
            });
        }

        let mut asks = Vec::with_capacity(depth_limit);
        for (&price_e8, queue) in self.asks.iter().take(depth_limit) {
            let total_qty: u64 = queue.iter().map(|o| o.remaining_qty_e8()).sum();
            asks.push(Level2Quote {
                price_e8,
                quantity_e8: total_qty,
                order_count: queue.len(),
            });
        }

        Level2Snapshot {
            symbol: self.symbol.clone(),
            timestamp_ms,
            bids,
            asks,
        }
    }

    /// Cancel all bot orders in the book (useful during bot quote refresh).
    pub fn cancel_bot_orders(&mut self, bot_id: &str) -> Vec<Order> {
        let mut cancelled = Vec::new();
        let bot_orders: Vec<String> = self
            .order_index
            .iter()
            .filter_map(|(id, (side, price))| {
                let queue = match side {
                    Side::Buy => self.bids.get(price),
                    Side::Sell => self.asks.get(price),
                };
                if let Some(q) = queue {
                    if q.iter().any(|o| o.order_id == *id && o.user_id == bot_id) {
                        return Some(id.clone());
                    }
                }
                None
            })
            .collect();

        for id in bot_orders {
            if let Ok(order) = self.cancel_order(&id) {
                cancelled.push(order);
            }
        }

        cancelled
    }

    pub fn total_orders(&self) -> usize {
        self.order_index.len()
    }
}
