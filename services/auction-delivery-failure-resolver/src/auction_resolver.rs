pub struct AuctionDeliveryResolver {
    pub penalty_rate_pct: f64, // e.g. 20% standard valuation penalty on short delivery
}

impl AuctionDeliveryResolver {
    pub fn new() -> Self {
        Self {
            penalty_rate_pct: 20.0,
        }
    }

    /// Resolve delivery default on T+1: buys shares in auction market or executes cash closeout
    pub fn resolve_short_delivery(
        &self,
        short_quantity: u64,
        closing_price_inr: f64,
        highest_trade_price_inr: f64,
    ) -> (f64, String) {
        // Closeout price is maximum of (highest price in trading period, or closing price + 20%)
        let penalty_price = closing_price_inr * (1.0 + self.penalty_rate_pct / 100.0);
        let closeout_price = penalty_price.max(highest_trade_price_inr);
        let total_debit = (short_quantity as f64) * closeout_price;

        let details = format!(
            "Short delivery of {} shares resolved via cash closeout @ ₹{:.2} (Valuation: ₹{:.2})",
            short_quantity, closeout_price, total_debit
        );

        (total_debit, details)
    }
}
