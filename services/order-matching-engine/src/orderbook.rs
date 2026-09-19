use std::collections::{BTreeMap, HashMap, VecDeque};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Side {
    Buy,
    Sell,
}

#[derive(Debug, Clone)]
pub struct Order {
    pub order_id: u64,
    pub user_id: String,
    pub side: Side,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub filled_e8: u64,
    pub timestamp_ns: u64,
}

impl Order {
    pub fn remaining_e8(&self) -> u64 {
        self.quantity_e8.saturating_sub(self.filled_e8)
    }
}

#[derive(Debug, Clone)]
pub struct Trade {
    pub trade_id: u64,
    pub maker_order_id: u64,
    pub taker_order_id: u64,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub timestamp_ns: u64,
}

pub struct OrderBook {
    pub symbol: String,
    // Bids sorted descending (highest price first)
    pub bids: BTreeMap<u64, VecDeque<Order>>,
    // Asks sorted ascending (lowest price first)
    pub asks: BTreeMap<u64, VecDeque<Order>>,
    // Quick lookup for cancellation: order_id -> (side, price)
    pub order_index: HashMap<u64, (Side, u64)>,
    pub next_trade_id: u64,
}

impl OrderBook {
    pub fn new(symbol: String) -> Self {
        Self {
            symbol,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            order_index: HashMap::new(),
            next_trade_id: 1,
        }
    }

    pub fn place_order(&mut self, mut order: Order) -> (Vec<Trade>, Option<Order>) {
        let mut trades = Vec::new();

        match order.side {
            Side::Buy => {
                while order.remaining_e8() > 0 {
                    let best_ask_price = match self.asks.keys().next() {
                        Some(&p) => p,
                        None => break,
                    };

                    if order.price_e8 < best_ask_price {
                        break;
                    }

                    if let Some(queue) = self.asks.get_mut(&best_ask_price) {
                        while let Some(mut maker) = queue.pop_front() {
                            let match_qty = std::cmp::min(order.remaining_e8(), maker.remaining_e8());
                            order.filled_e8 += match_qty;
                            maker.filled_e8 += match_qty;

                            let trade = Trade {
                                trade_id: self.next_trade_id,
                                maker_order_id: maker.order_id,
                                taker_order_id: order.order_id,
                                price_e8: best_ask_price,
                                quantity_e8: match_qty,
                                timestamp_ns: order.timestamp_ns,
                            };
                            self.next_trade_id += 1;
                            trades.push(trade);

                            if maker.remaining_e8() > 0 {
                                queue.push_front(maker);
                                break;
                            } else {
                                self.order_index.remove(&maker.order_id);
                            }

                            if order.remaining_e8() == 0 {
                                break;
                            }
                        }

                        if queue.is_empty() {
                            self.asks.remove(&best_ask_price);
                        }
                    }
                }

                // If unfilled resting volume remains, place into book
                if order.remaining_e8() > 0 {
                    let price = order.price_e8;
                    self.order_index.insert(order.order_id, (Side::Buy, price));
                    self.bids.entry(price).or_default().push_back(order.clone());
                    (trades, Some(order))
                } else {
                    (trades, None)
                }
            }
            Side::Sell => {
                while order.remaining_e8() > 0 {
                    let best_bid_price = match self.bids.keys().next_back() {
                        Some(&p) => p,
                        None => break,
                    };

                    if order.price_e8 > best_bid_price {
                        break;
                    }

                    if let Some(queue) = self.bids.get_mut(&best_bid_price) {
                        while let Some(mut maker) = queue.pop_front() {
                            let match_qty = std::cmp::min(order.remaining_e8(), maker.remaining_e8());
                            order.filled_e8 += match_qty;
                            maker.filled_e8 += match_qty;

                            let trade = Trade {
                                trade_id: self.next_trade_id,
                                maker_order_id: maker.order_id,
                                taker_order_id: order.order_id,
                                price_e8: best_bid_price,
                                quantity_e8: match_qty,
                                timestamp_ns: order.timestamp_ns,
                            };
                            self.next_trade_id += 1;
                            trades.push(trade);

                            if maker.remaining_e8() > 0 {
                                queue.push_front(maker);
                                break;
                            } else {
                                self.order_index.remove(&maker.order_id);
                            }

                            if order.remaining_e8() == 0 {
                                break;
                            }
                        }

                        if queue.is_empty() {
                            self.bids.remove(&best_bid_price);
                        }
                    }
                }

                if order.remaining_e8() > 0 {
                    let price = order.price_e8;
                    self.order_index.insert(order.order_id, (Side::Sell, price));
                    self.asks.entry(price).or_default().push_back(order.clone());
                    (trades, Some(order))
                } else {
                    (trades, None)
                }
            }
        }
    }

    pub fn cancel_order(&mut self, order_id: u64) -> Option<Order> {
        let (side, price) = self.order_index.remove(&order_id)?;
        match side {
            Side::Buy => {
                if let Some(queue) = self.bids.get_mut(&price) {
                    if let Some(pos) = queue.iter().position(|o| o.order_id == order_id) {
                        let removed = queue.remove(pos);
                        if queue.is_empty() {
                            self.bids.remove(&price);
                        }
                        return removed;
                    }
                }
            }
            Side::Sell => {
                if let Some(queue) = self.asks.get_mut(&price) {
                    if let Some(pos) = queue.iter().position(|o| o.order_id == order_id) {
                        let removed = queue.remove(pos);
                        if queue.is_empty() {
                            self.asks.remove(&price);
                        }
                        return removed;
                    }
                }
            }
        }
        None
    }

    pub fn get_l2_depth(&self, depth: usize) -> (Vec<(u64, u64)>, Vec<(u64, u64)>) {
        let bids: Vec<(u64, u64)> = self.bids
            .iter()
            .rev()
            .take(depth)
            .map(|(&price, queue)| (price, queue.iter().map(|o| o.remaining_e8()).sum()))
            .collect();

        let asks: Vec<(u64, u64)> = self.asks
            .iter()
            .take(depth)
            .map(|(&price, queue)| (price, queue.iter().map(|o| o.remaining_e8()).sum()))
            .collect();

        (bids, asks)
    }
}
