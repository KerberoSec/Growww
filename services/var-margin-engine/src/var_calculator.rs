pub struct VaRMarginEngine {
    pub confidence_z_score: f64, // 2.33 for 99% confidence
    pub holding_period_days: f64, // 1 day
}

impl VaRMarginEngine {
    pub fn new() -> Self {
        Self {
            confidence_z_score: 2.33,
            holding_period_days: 1.0,
        }
    }

    /// Computes Parametric Value-at-Risk: VaR = Portfolio Value * Z * Sigma * sqrt(t)
    pub fn calculate_parametric_var(
        &self,
        portfolio_value_inr: f64,
        daily_volatility_sigma: f64,
    ) -> f64 {
        portfolio_value_inr * self.confidence_z_score * daily_volatility_sigma * self.holding_period_days.sqrt()
    }

    /// Computes Extreme Loss Margin (ELM) covering beyond 99% VaR tail events
    pub fn calculate_extreme_loss_margin(&self, portfolio_value_inr: f64, elm_rate: f64) -> f64 {
        portfolio_value_inr * elm_rate
    }
}
