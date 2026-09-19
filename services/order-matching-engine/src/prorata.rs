use crate::types::{Order, Trade};
use crate::fees::FeeEngine;
use std::collections::VecDeque;

#[derive(Debug, Clone, Copy, Default)]
pub struct ProRataMatcher;

impl ProRataMatcher {
    /// Matches incoming order against a price queue using Pro-Rata allocation
    /// with residual distribution based on time priority.
    pub fn match_queue_prorata(
        queue: &mut VecDeque<Order>,
        taker: &mut Order,
        price_e8: u64,
        next_trade_id: &mut u64,
        maker_fee_bps: u16,
        taker_fee_bps: u16,
    ) -> Vec<Trade> {
        let mut trades = Vec::new();
        if queue.is_empty() || taker.remaining_e8() == 0 {
            return trades;
        }

        let total_available: u64 = queue.iter().map(|o| o.remaining_e8()).sum();
        if total_available == 0 {
            return trades;
        }

        let taker_needed = taker.remaining_e8();

        if taker_needed >= total_available {
            // Taker takes out the entire queue
            while let Some(mut maker) = queue.pop_front() {
                let match_qty = maker.remaining_e8();
                if match_qty > 0 {
                    taker.filled_e8 += match_qty;
                    maker.filled_e8 += match_qty;

                    let (quote_amt, maker_fee, taker_fee) = FeeEngine::calculate_trade_financials(
                        price_e8,
                        match_qty,
                        maker_fee_bps,
                        taker_fee_bps,
                    );

                    trades.push(Trade {
                        trade_id: *next_trade_id,
                        symbol: taker.symbol.clone(),
                        maker_order_id: maker.order_id,
                        taker_order_id: taker.order_id,
                        maker_user_id: maker.user_id.clone(),
                        taker_user_id: taker.user_id.clone(),
                        maker_side: maker.side,
                        price_e8,
                        quantity_e8: match_qty,
                        quote_amount_e8: quote_amt,
                        maker_fee_e8: maker_fee,
                        taker_fee_e8: taker_fee,
                        timestamp_ns: taker.timestamp_ns,
                    });
                    *next_trade_id += 1;
                }
            }
            return trades;
        }

        // Taker needed < total_available: Pro-rata allocation
        let mut allocations: Vec<u64> = Vec::with_capacity(queue.len());
        let mut allocated_sum: u64 = 0;

        for maker in queue.iter() {
            let maker_rem = maker.remaining_e8() as u128;
            let needed = taker_needed as u128;
            let tot = total_available as u128;
            let alloc = ((needed * maker_rem) / tot) as u64;
            let safe_alloc = std::cmp::min(alloc, maker.remaining_e8());
            allocations.push(safe_alloc);
            allocated_sum += safe_alloc;
        }

        let mut residual = taker_needed.saturating_sub(allocated_sum);

        // Distribute residual based on FIFO (time-priority) among makers with capacity
        if residual > 0 {
            for (i, maker) in queue.iter().enumerate() {
                let current_alloc = allocations[i];
                let maker_rem = maker.remaining_e8();
                if current_alloc < maker_rem {
                    let available_capacity = maker_rem - current_alloc;
                    let extra = std::cmp::min(available_capacity, residual);
                    allocations[i] += extra;
                    residual -= extra;
                    if residual == 0 {
                        break;
                    }
                }
            }
        }

        // Apply allocations and generate trades
        let mut i = 0;
        while i < queue.len() {
            let alloc = allocations[i];
            if alloc > 0 {
                let maker = &mut queue[i];
                maker.filled_e8 += alloc;
                taker.filled_e8 += alloc;

                let (quote_amt, maker_fee, taker_fee) = FeeEngine::calculate_trade_financials(
                    price_e8,
                    alloc,
                    maker_fee_bps,
                    taker_fee_bps,
                );

                trades.push(Trade {
                    trade_id: *next_trade_id,
                    symbol: taker.symbol.clone(),
                    maker_order_id: maker.order_id,
                    taker_order_id: taker.order_id,
                    maker_user_id: maker.user_id.clone(),
                    taker_user_id: taker.user_id.clone(),
                    maker_side: maker.side,
                    price_e8,
                    quantity_e8: alloc,
                    quote_amount_e8: quote_amt,
                    maker_fee_e8: maker_fee,
                    taker_fee_e8: taker_fee,
                    timestamp_ns: taker.timestamp_ns,
                });
                *next_trade_id += 1;
            }

            if queue[i].remaining_e8() == 0 {
                queue.remove(i);
                allocations.remove(i);
            } else {
                i += 1;
            }
        }

        trades
    }
}
