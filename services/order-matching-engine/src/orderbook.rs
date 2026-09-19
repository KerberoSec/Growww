use std::collections::{BTreeMap, HashMap, VecDeque};

use crate::fees::FeeEngine;
use crate::prorata::ProRataMatcher;
use crate::risk_collar::PriceCollar;
use crate::types::{
    BookDelta, ExecutionResult, L2Snapshot, L3Snapshot, MatchingAlgorithm, Order, OrderStatus,
    OrderType, PriceLevel, STPMode, Side, Trade,
};

pub struct OrderBook {
    pub symbol: String,
    // Bids sorted descending (highest price first: BTreeMap key is price_e8)
    pub bids: BTreeMap<u64, VecDeque<Order>>,
    // Asks sorted ascending (lowest price first: BTreeMap key is price_e8)
    pub asks: BTreeMap<u64, VecDeque<Order>>,
    // Quick lookup for cancellation: order_id -> (side, price)
    pub order_index: HashMap<u64, (Side, u64)>,
    pub next_trade_id: u64,
    pub last_update_id: u64,
    pub matching_algorithm: MatchingAlgorithm,
    pub price_collar: PriceCollar,
    pub pending_deltas: Vec<BookDelta>,
}

impl OrderBook {
    pub fn new(symbol: String) -> Self {
        Self {
            symbol,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            order_index: HashMap::new(),
            next_trade_id: 1,
            last_update_id: 0,
            matching_algorithm: MatchingAlgorithm::Fifo,
            price_collar: PriceCollar::disabled(),
            pending_deltas: Vec::new(),
        }
    }

    pub fn with_config(
        symbol: String,
        matching_algorithm: MatchingAlgorithm,
        reference_price_e8: u64,
        max_deviation_bps: u32,
    ) -> Self {
        let price_collar = if reference_price_e8 > 0 && max_deviation_bps > 0 {
            PriceCollar::new(reference_price_e8, max_deviation_bps)
        } else {
            PriceCollar::disabled()
        };

        Self {
            symbol,
            bids: BTreeMap::new(),
            asks: BTreeMap::new(),
            order_index: HashMap::new(),
            next_trade_id: 1,
            last_update_id: 0,
            matching_algorithm,
            price_collar,
            pending_deltas: Vec::new(),
        }
    }

    /// Backwards compatible simple place_order
    pub fn place_order(&mut self, order: Order) -> (Vec<Trade>, Option<Order>) {
        let res = self.process_order(order);
        (res.trades, res.resting_order)
    }

    /// Full-featured order processor handling STP, order types, collars, and deltas
    pub fn process_order(&mut self, mut order: Order) -> ExecutionResult {
        // 1. Validate Price Collar
        if let Err(err) = self.price_collar.validate_order(&order) {
            return ExecutionResult {
                order_id: order.order_id,
                status: OrderStatus::Rejected,
                trades: Vec::new(),
                resting_order: None,
                stp_canceled_maker_ids: Vec::new(),
                rejected: true,
                rejection_reason: Some(err),
            };
        }

        // 2. Pre-flight check for Post-Only (Maker-Or-Cancel)
        if order.order_type == OrderType::PostOnly {
            let would_cross = match order.side {
                Side::Buy => {
                    if let Some(&best_ask) = self.asks.keys().next() {
                        order.price_e8 >= best_ask
                    } else {
                        false
                    }
                }
                Side::Sell => {
                    if let Some(&best_bid) = self.bids.keys().next_back() {
                        order.price_e8 <= best_bid
                    } else {
                        false
                    }
                }
            };

            if would_cross {
                return ExecutionResult {
                    order_id: order.order_id,
                    status: OrderStatus::Rejected,
                    trades: Vec::new(),
                    resting_order: None,
                    stp_canceled_maker_ids: Vec::new(),
                    rejected: true,
                    rejection_reason: Some("Post-Only order would cross opposing book".into()),
                };
            }
        }

        // 3. Pre-flight check for Fill-Or-Kill (FOK)
        if order.order_type == OrderType::FillOrKill {
            let matchable_qty = self.calculate_matchable_quantity(&order);
            if matchable_qty < order.quantity_e8 {
                return ExecutionResult {
                    order_id: order.order_id,
                    status: OrderStatus::Canceled,
                    trades: Vec::new(),
                    resting_order: None,
                    stp_canceled_maker_ids: Vec::new(),
                    rejected: false,
                    rejection_reason: Some("Fill-Or-Kill order cannot be fully filled".into()),
                };
            }
        }

        let mut trades = Vec::new();
        let mut stp_canceled_makers = Vec::new();
        let mut affected_bid_prices: HashMap<u64, ()> = HashMap::new();
        let mut affected_ask_prices: HashMap<u64, ()> = HashMap::new();

        let maker_fee_bps = FeeEngine::get_maker_fee_bps();
        let taker_fee_bps = FeeEngine::get_taker_fee_bps();

        match order.side {
            Side::Buy => {
                let mut taker_cancelled_by_stp = false;
                while order.remaining_e8() > 0 {
                    let best_ask_price = match self.asks.keys().next() {
                        Some(&p) => p,
                        None => break,
                    };

                    if order.order_type != OrderType::Market && order.price_e8 < best_ask_price {
                        break;
                    }

                    affected_ask_prices.insert(best_ask_price, ());

                    if self.matching_algorithm == MatchingAlgorithm::ProRata {
                        let queue = self.asks.get_mut(&best_ask_price).unwrap();
                        let pt_trades = ProRataMatcher::match_queue_prorata(
                            queue,
                            &mut order,
                            best_ask_price,
                            &mut self.next_trade_id,
                            maker_fee_bps,
                            taker_fee_bps,
                        );
                        trades.extend(pt_trades);

                        if queue.is_empty() {
                            self.asks.remove(&best_ask_price);
                        }
                    } else {
                        // Strict FIFO
                        if let Some(queue) = self.asks.get_mut(&best_ask_price) {
                            while let Some(mut maker) = queue.pop_front() {
                                // Self-Trade Prevention check
                                if maker.user_id == order.user_id && order.stp_mode != STPMode::None {
                                    match order.stp_mode {
                                        STPMode::CancelNewest => {
                                            // Put maker back
                                            queue.push_front(maker);
                                            taker_cancelled_by_stp = true;
                                            break;
                                        }
                                        STPMode::CancelOldest => {
                                            // Cancel resting maker, taker continues
                                            stp_canceled_makers.push(maker.order_id);
                                            self.order_index.remove(&maker.order_id);
                                            continue;
                                        }
                                        STPMode::DecrementAndCancel => {
                                            let taker_rem = order.remaining_e8();
                                            let maker_rem = maker.remaining_e8();
                                            if taker_rem <= maker_rem {
                                                // Taker fully cancelled
                                                maker.filled_e8 += taker_rem;
                                                order.filled_e8 += taker_rem;
                                                if maker.remaining_e8() > 0 {
                                                    queue.push_front(maker);
                                                } else {
                                                    self.order_index.remove(&maker.order_id);
                                                }
                                                taker_cancelled_by_stp = true;
                                                break;
                                            } else {
                                                // Maker fully cancelled, taker reduced
                                                order.filled_e8 += maker_rem;
                                                stp_canceled_makers.push(maker.order_id);
                                                self.order_index.remove(&maker.order_id);
                                                continue;
                                            }
                                        }
                                        STPMode::None => unreachable!(),
                                    }
                                }

                                let match_qty = std::cmp::min(order.remaining_e8(), maker.remaining_e8());
                                order.filled_e8 += match_qty;
                                maker.filled_e8 += match_qty;

                                let (quote_amt, maker_fee, taker_fee) =
                                    FeeEngine::calculate_trade_financials(
                                        best_ask_price,
                                        match_qty,
                                        maker_fee_bps,
                                        taker_fee_bps,
                                    );

                                trades.push(Trade {
                                    trade_id: self.next_trade_id,
                                    symbol: self.symbol.clone(),
                                    maker_order_id: maker.order_id,
                                    taker_order_id: order.order_id,
                                    maker_user_id: maker.user_id.clone(),
                                    taker_user_id: order.user_id.clone(),
                                    maker_side: Side::Sell,
                                    price_e8: best_ask_price,
                                    quantity_e8: match_qty,
                                    quote_amount_e8: quote_amt,
                                    maker_fee_e8: maker_fee,
                                    taker_fee_e8: taker_fee,
                                    timestamp_ns: order.timestamp_ns,
                                });
                                self.next_trade_id += 1;

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

                        if taker_cancelled_by_stp {
                            break;
                        }
                    }
                }

                // If unfilled resting volume remains, place into book if Limit or PostOnly
                let resting_order = if !taker_cancelled_by_stp
                    && order.remaining_e8() > 0
                    && (order.order_type == OrderType::Limit || order.order_type == OrderType::PostOnly)
                {
                    let price = order.price_e8;
                    self.order_index.insert(order.order_id, (Side::Buy, price));
                    self.bids.entry(price).or_default().push_back(order.clone());
                    affected_bid_prices.insert(price, ());
                    Some(order.clone())
                } else {
                    None
                };

                // Update collar reference price
                if let Some(last_trade) = trades.last() {
                    self.price_collar.update_reference_price(last_trade.price_e8);
                }

                self.record_delta(&affected_bid_prices, &affected_ask_prices, order.timestamp_ns);

                let status = if taker_cancelled_by_stp {
                    OrderStatus::Canceled
                } else if order.is_filled() {
                    OrderStatus::Filled
                } else if order.filled_e8 > 0 {
                    OrderStatus::PartiallyFilled
                } else if order.order_type == OrderType::ImmediateOrCancel
                    || order.order_type == OrderType::Market
                {
                    OrderStatus::Canceled
                } else {
                    OrderStatus::New
                };

                ExecutionResult {
                    order_id: order.order_id,
                    status,
                    trades,
                    resting_order,
                    stp_canceled_maker_ids: stp_canceled_makers,
                    rejected: false,
                    rejection_reason: None,
                }
            }
            Side::Sell => {
                let mut taker_cancelled_by_stp = false;
                while order.remaining_e8() > 0 {
                    let best_bid_price = match self.bids.keys().next_back() {
                        Some(&p) => p,
                        None => break,
                    };

                    if order.order_type != OrderType::Market && order.price_e8 > best_bid_price {
                        break;
                    }

                    affected_bid_prices.insert(best_bid_price, ());

                    if self.matching_algorithm == MatchingAlgorithm::ProRata {
                        let queue = self.bids.get_mut(&best_bid_price).unwrap();
                        let pt_trades = ProRataMatcher::match_queue_prorata(
                            queue,
                            &mut order,
                            best_bid_price,
                            &mut self.next_trade_id,
                            maker_fee_bps,
                            taker_fee_bps,
                        );
                        trades.extend(pt_trades);

                        if queue.is_empty() {
                            self.bids.remove(&best_bid_price);
                        }
                    } else {
                        // Strict FIFO
                        if let Some(queue) = self.bids.get_mut(&best_bid_price) {
                            while let Some(mut maker) = queue.pop_front() {
                                // Self-Trade Prevention check
                                if maker.user_id == order.user_id && order.stp_mode != STPMode::None {
                                    match order.stp_mode {
                                        STPMode::CancelNewest => {
                                            queue.push_front(maker);
                                            taker_cancelled_by_stp = true;
                                            break;
                                        }
                                        STPMode::CancelOldest => {
                                            stp_canceled_makers.push(maker.order_id);
                                            self.order_index.remove(&maker.order_id);
                                            continue;
                                        }
                                        STPMode::DecrementAndCancel => {
                                            let taker_rem = order.remaining_e8();
                                            let maker_rem = maker.remaining_e8();
                                            if taker_rem <= maker_rem {
                                                maker.filled_e8 += taker_rem;
                                                order.filled_e8 += taker_rem;
                                                if maker.remaining_e8() > 0 {
                                                    queue.push_front(maker);
                                                } else {
                                                    self.order_index.remove(&maker.order_id);
                                                }
                                                taker_cancelled_by_stp = true;
                                                break;
                                            } else {
                                                order.filled_e8 += maker_rem;
                                                stp_canceled_makers.push(maker.order_id);
                                                self.order_index.remove(&maker.order_id);
                                                continue;
                                            }
                                        }
                                        STPMode::None => unreachable!(),
                                    }
                                }

                                let match_qty = std::cmp::min(order.remaining_e8(), maker.remaining_e8());
                                order.filled_e8 += match_qty;
                                maker.filled_e8 += match_qty;

                                let (quote_amt, maker_fee, taker_fee) =
                                    FeeEngine::calculate_trade_financials(
                                        best_bid_price,
                                        match_qty,
                                        maker_fee_bps,
                                        taker_fee_bps,
                                    );

                                trades.push(Trade {
                                    trade_id: self.next_trade_id,
                                    symbol: self.symbol.clone(),
                                    maker_order_id: maker.order_id,
                                    taker_order_id: order.order_id,
                                    maker_user_id: maker.user_id.clone(),
                                    taker_user_id: order.user_id.clone(),
                                    maker_side: Side::Buy,
                                    price_e8: best_bid_price,
                                    quantity_e8: match_qty,
                                    quote_amount_e8: quote_amt,
                                    maker_fee_e8: maker_fee,
                                    taker_fee_e8: taker_fee,
                                    timestamp_ns: order.timestamp_ns,
                                });
                                self.next_trade_id += 1;

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

                        if taker_cancelled_by_stp {
                            break;
                        }
                    }
                }

                let resting_order = if !taker_cancelled_by_stp
                    && order.remaining_e8() > 0
                    && (order.order_type == OrderType::Limit || order.order_type == OrderType::PostOnly)
                {
                    let price = order.price_e8;
                    self.order_index.insert(order.order_id, (Side::Sell, price));
                    self.asks.entry(price).or_default().push_back(order.clone());
                    affected_ask_prices.insert(price, ());
                    Some(order.clone())
                } else {
                    None
                };

                if let Some(last_trade) = trades.last() {
                    self.price_collar.update_reference_price(last_trade.price_e8);
                }

                self.record_delta(&affected_bid_prices, &affected_ask_prices, order.timestamp_ns);

                let status = if taker_cancelled_by_stp {
                    OrderStatus::Canceled
                } else if order.is_filled() {
                    OrderStatus::Filled
                } else if order.filled_e8 > 0 {
                    OrderStatus::PartiallyFilled
                } else if order.order_type == OrderType::ImmediateOrCancel
                    || order.order_type == OrderType::Market
                {
                    OrderStatus::Canceled
                } else {
                    OrderStatus::New
                };

                ExecutionResult {
                    order_id: order.order_id,
                    status,
                    trades,
                    resting_order,
                    stp_canceled_maker_ids: stp_canceled_makers,
                    rejected: false,
                    rejection_reason: None,
                }
            }
        }
    }

    /// Cancel order by ID
    pub fn cancel_order(&mut self, order_id: u64) -> Option<Order> {
        let (side, price) = self.order_index.remove(&order_id)?;
        let mut removed_order = None;

        match side {
            Side::Buy => {
                if let Some(queue) = self.bids.get_mut(&price) {
                    if let Some(pos) = queue.iter().position(|o| o.order_id == order_id) {
                        removed_order = queue.remove(pos);
                        if queue.is_empty() {
                            self.bids.remove(&price);
                        }
                    }
                }
                let mut affected_bids = HashMap::new();
                affected_bids.insert(price, ());
                self.record_delta(&affected_bids, &HashMap::new(), 0);
            }
            Side::Sell => {
                if let Some(queue) = self.asks.get_mut(&price) {
                    if let Some(pos) = queue.iter().position(|o| o.order_id == order_id) {
                        removed_order = queue.remove(pos);
                        if queue.is_empty() {
                            self.asks.remove(&price);
                        }
                    }
                }
                let mut affected_asks = HashMap::new();
                affected_asks.insert(price, ());
                self.record_delta(&HashMap::new(), &affected_asks, 0);
            }
        }

        removed_order
    }

    /// Atomic Cancel/Replace (Order Amend)
    /// Rule: If price is unchanged and new quantity <= remaining quantity, retain priority!
    /// If price changed OR new quantity > remaining quantity, lose priority and place at queue end.
    pub fn cancel_replace_order(
        &mut self,
        order_id: u64,
        new_price_e8: u64,
        new_quantity_e8: u64,
        timestamp_ns: u64,
    ) -> Result<(Option<Order>, ExecutionResult), String> {
        let &(side, old_price) = self
            .order_index
            .get(&order_id)
            .ok_or_else(|| "Order not found".to_string())?;

        // Check if eligible for priority retention
        let can_retain_priority = if new_price_e8 == old_price {
            let queue = match side {
                Side::Buy => self.bids.get(&old_price),
                Side::Sell => self.asks.get(&old_price),
            };
            if let Some(q) = queue {
                if let Some(existing) = q.iter().find(|o| o.order_id == order_id) {
                    new_quantity_e8 <= existing.remaining_e8()
                } else {
                    false
                }
            } else {
                false
            }
        } else {
            false
        };

        if can_retain_priority {
            // Modify in place! Priority preserved.
            let queue = match side {
                Side::Buy => self.bids.get_mut(&old_price).unwrap(),
                Side::Sell => self.asks.get_mut(&old_price).unwrap(),
            };
            let existing = queue.iter_mut().find(|o| o.order_id == order_id).unwrap();
            existing.quantity_e8 = existing.filled_e8 + new_quantity_e8;

            let modified = existing.clone();

            let mut aff_bids = HashMap::new();
            let mut aff_asks = HashMap::new();
            match side {
                Side::Buy => aff_bids.insert(old_price, ()),
                Side::Sell => aff_asks.insert(old_price, ()),
            };
            self.record_delta(&aff_bids, &aff_asks, timestamp_ns);

            let res = ExecutionResult {
                order_id,
                status: OrderStatus::New,
                trades: Vec::new(),
                resting_order: Some(modified.clone()),
                stp_canceled_maker_ids: Vec::new(),
                rejected: false,
                rejection_reason: None,
            };

            return Ok((Some(modified), res));
        }

        // Priority lost: Cancel old order, place new order
        let canceled = self
            .cancel_order(order_id)
            .ok_or_else(|| "Failed to cancel order during replace".to_string())?;

        let new_order = Order {
            order_id,
            user_id: canceled.user_id.clone(),
            symbol: canceled.symbol.clone(),
            side: canceled.side,
            order_type: canceled.order_type,
            stp_mode: canceled.stp_mode,
            price_e8: new_price_e8,
            quantity_e8: new_quantity_e8,
            filled_e8: 0,
            timestamp_ns,
        };

        let exec_result = self.process_order(new_order);
        Ok((Some(canceled), exec_result))
    }

    /// Pre-flight quantity calculation for FOK orders
    fn calculate_matchable_quantity(&self, order: &Order) -> u64 {
        let mut matchable = 0u64;
        match order.side {
            Side::Buy => {
                for (&ask_price, queue) in self.asks.iter() {
                    if order.order_type != OrderType::Market && order.price_e8 < ask_price {
                        break;
                    }
                    for maker in queue.iter() {
                        if maker.user_id != order.user_id || order.stp_mode == STPMode::None {
                            matchable += maker.remaining_e8();
                            if matchable >= order.quantity_e8 {
                                return matchable;
                            }
                        }
                    }
                }
            }
            Side::Sell => {
                for (&bid_price, queue) in self.bids.iter().rev() {
                    if order.order_type != OrderType::Market && order.price_e8 > bid_price {
                        break;
                    }
                    for maker in queue.iter() {
                        if maker.user_id != order.user_id || order.stp_mode == STPMode::None {
                            matchable += maker.remaining_e8();
                            if matchable >= order.quantity_e8 {
                                return matchable;
                            }
                        }
                    }
                }
            }
        }
        matchable
    }

    fn record_delta(
        &mut self,
        affected_bids: &HashMap<u64, ()>,
        affected_asks: &HashMap<u64, ()>,
        timestamp_ns: u64,
    ) {
        if affected_bids.is_empty() && affected_asks.is_empty() {
            return;
        }

        let first_id = self.last_update_id + 1;
        self.last_update_id += 1;
        let last_id = self.last_update_id;

        let mut bid_deltas = Vec::with_capacity(affected_bids.len());
        for &p in affected_bids.keys() {
            let total_qty = self
                .bids
                .get(&p)
                .map(|q| q.iter().map(|o| o.remaining_e8()).sum())
                .unwrap_or(0);
            bid_deltas.push((p, total_qty));
        }

        let mut ask_deltas = Vec::with_capacity(affected_asks.len());
        for &p in affected_asks.keys() {
            let total_qty = self
                .asks
                .get(&p)
                .map(|q| q.iter().map(|o| o.remaining_e8()).sum())
                .unwrap_or(0);
            ask_deltas.push((p, total_qty));
        }

        if self.pending_deltas.len() >= 10_000 {
            self.pending_deltas.drain(0..5_000);
        }

        self.pending_deltas.push(BookDelta {
            symbol: self.symbol.clone(),
            first_update_id: first_id,
            last_update_id: last_id,
            timestamp_ns,
            bids: bid_deltas,
            asks: ask_deltas,
        });
    }

    pub fn drain_deltas(&mut self) -> Vec<BookDelta> {
        std::mem::take(&mut self.pending_deltas)
    }

    pub fn get_l2_depth(&self, depth: usize) -> (Vec<(u64, u64)>, Vec<(u64, u64)>) {
        let bids: Vec<(u64, u64)> = self
            .bids
            .iter()
            .rev()
            .take(depth)
            .map(|(&price, queue)| (price, queue.iter().map(|o| o.remaining_e8()).sum()))
            .collect();

        let asks: Vec<(u64, u64)> = self
            .asks
            .iter()
            .take(depth)
            .map(|(&price, queue)| (price, queue.iter().map(|o| o.remaining_e8()).sum()))
            .collect();

        (bids, asks)
    }

    pub fn generate_l2_snapshot(&self, depth: usize) -> L2Snapshot {
        let bids: Vec<PriceLevel> = self
            .bids
            .iter()
            .rev()
            .take(depth)
            .map(|(&price, queue)| PriceLevel {
                price_e8: price,
                total_quantity_e8: queue.iter().map(|o| o.remaining_e8()).sum(),
                order_count: queue.len() as u32,
            })
            .collect();

        let asks: Vec<PriceLevel> = self
            .asks
            .iter()
            .take(depth)
            .map(|(&price, queue)| PriceLevel {
                price_e8: price,
                total_quantity_e8: queue.iter().map(|o| o.remaining_e8()).sum(),
                order_count: queue.len() as u32,
            })
            .collect();

        L2Snapshot {
            symbol: self.symbol.clone(),
            last_update_id: self.last_update_id,
            timestamp_ns: 0,
            bids,
            asks,
        }
    }

    pub fn generate_l3_snapshot(&self) -> L3Snapshot {
        let bids: Vec<(u64, Vec<Order>)> = self
            .bids
            .iter()
            .rev()
            .map(|(&p, queue)| (p, queue.iter().cloned().collect()))
            .collect();

        let asks: Vec<(u64, Vec<Order>)> = self
            .asks
            .iter()
            .map(|(&p, queue)| (p, queue.iter().cloned().collect()))
            .collect();

        L3Snapshot {
            symbol: self.symbol.clone(),
            last_update_id: self.last_update_id,
            timestamp_ns: 0,
            bids,
            asks,
        }
    }
}
