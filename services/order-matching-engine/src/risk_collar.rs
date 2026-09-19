use crate::types::{Order, OrderType, Side};

#[derive(Debug, Clone)]
pub struct PriceCollar {
    /// Permitted deviation in basis points (e.g. 1000 bps = 10%)
    pub max_deviation_bps: u32,
    /// Current reference price in 1e8
    pub reference_price_e8: u64,
    pub enabled: bool,
}

impl PriceCollar {
    pub fn new(reference_price_e8: u64, max_deviation_bps: u32) -> Self {
        Self {
            max_deviation_bps,
            reference_price_e8,
            enabled: true,
        }
    }

    pub fn disabled() -> Self {
        Self {
            max_deviation_bps: 0,
            reference_price_e8: 0,
            enabled: false,
        }
    }

    #[inline(always)]
    pub fn update_reference_price(&mut self, new_price_e8: u64) {
        if new_price_e8 > 0 {
            self.reference_price_e8 = new_price_e8;
        }
    }

    /// Check if order price is within the collar boundaries
    pub fn validate_order(&self, order: &Order) -> Result<(), String> {
        if !self.enabled || self.reference_price_e8 == 0 {
            return Ok(());
        }

        // Market orders don't specify price, collar is validated against execution or best price
        if order.order_type == OrderType::Market {
            return Ok(());
        }

        let ref_p = self.reference_price_e8 as u128;
        let max_dev = self.max_deviation_bps as u128;
        let delta = (ref_p * max_dev) / 10_000;

        let lower_bound = ref_p.saturating_sub(delta) as u64;
        let upper_bound = (ref_p + delta) as u64;

        match order.side {
            Side::Buy => {
                if order.price_e8 > upper_bound {
                    return Err(format!(
                        "Price collar breach: Buy order price {} exceeds upper bound {} (ref: {}, max_dev: {} bps)",
                        order.price_e8, upper_bound, self.reference_price_e8, self.max_deviation_bps
                    ));
                }
            }
            Side::Sell => {
                if order.price_e8 < lower_bound {
                    return Err(format!(
                        "Price collar breach: Sell order price {} below lower bound {} (ref: {}, max_dev: {} bps)",
                        order.price_e8, lower_bound, self.reference_price_e8, self.max_deviation_bps
                    ));
                }
            }
        }

        Ok(())
    }
}
