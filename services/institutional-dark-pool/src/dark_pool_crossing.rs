use std::collections::VecDeque;

#[derive(Debug, Clone)]
pub struct DarkPoolOrder {
    pub order_id: String,
    pub participant_id: String,
    pub symbol: String,
    pub side: String, // BUY or SELL
    pub min_quantity_e8: u64,
    pub total_quantity_e8: u64,
    pub max_slippage_bps: u32,
}

pub struct DarkPoolCrossingFacility {
    pub resting_bids: VecDeque<DarkPoolOrder>,
    pub resting_asks: VecDeque<DarkPoolOrder>,
}

impl DarkPoolCrossingFacility {
    pub fn new() -> Self {
        Self {
            resting_bids: VecDeque::new(),
            resting_asks: VecDeque::new(),
        }
    }

    /// Midpoint cross: matches large block orders strictly at midpoint of NBBO to avoid market impact
    pub fn match_midpoint_cross(
        &mut self,
        nbbo_midpoint_e8: u64,
    ) -> Vec<(String, String, u64, u64)> {
        let mut executed_crosses = Vec::new();

        while let (Some(bid), Some(ask)) = (self.resting_bids.front(), self.resting_asks.front()) {
            let matched_qty = std::cmp::min(bid.total_quantity_e8, ask.total_quantity_e8);

            // Verify minimum block quantity requirement
            if matched_qty >= bid.min_quantity_e8 && matched_qty >= ask.min_quantity_e8 {
                let bid_order = self.resting_bids.pop_front().unwrap();
                let ask_order = self.resting_asks.pop_front().unwrap();

                executed_crosses.push((
                    bid_order.order_id,
                    ask_order.order_id,
                    nbbo_midpoint_e8,
                    matched_qty,
                ));
            } else {
                break;
            }
        }

        executed_crosses
    }
}
