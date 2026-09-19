#[derive(Debug, Clone)]
pub struct LiquidationTarget {
    pub position_id: String,
    pub user_id: String,
    pub symbol: String,
    pub size_contracts: u64,
    pub bankruptcy_price_e8: u64,
    pub liquidation_price_e8: u64,
}

pub struct InsuranceFund {
    pub balance_usd: f64,
}

pub struct LiquidationADLEngine {
    pub insurance_fund: InsuranceFund,
}

impl LiquidationADLEngine {
    pub fn new(initial_fund_usd: f64) -> Self {
        Self {
            insurance_fund: InsuranceFund {
                balance_usd: initial_fund_usd,
            },
        }
    }

    /// Resolve liquidation takeover: absorbs bankruptcy deficit via Insurance Fund or triggers ADL
    pub fn resolve_liquidation(
        &mut self,
        target: &LiquidationTarget,
        actual_fill_price_e8: u64,
    ) -> Result<String, String> {
        let bankruptcy = target.bankruptcy_price_e8 as f64 / 1e8;
        let fill = actual_fill_price_e8 as f64 / 1e8;
        let size = target.size_contracts as f64;

        // If filled worse than bankruptcy price, insurance fund must cover deficit
        let deficit_usd = (bankruptcy - fill).abs() * size;

        if fill < bankruptcy {
            if self.insurance_fund.balance_usd >= deficit_usd {
                self.insurance_fund.balance_usd -= deficit_usd;
                Ok(format!(
                    "Deficit of ${:.2} absorbed by Insurance Fund. Remaining fund: ${:.2}",
                    deficit_usd, self.insurance_fund.balance_usd
                ))
            } else {
                // Insurance fund depleted: Auto-Deleveraging (ADL) invoked
                Err(format!(
                    "Insurance Fund depleted! Triggering Auto-Deleveraging (ADL) against highest-profit counterparty positions for ${:.2}",
                    deficit_usd
                ))
            }
        } else {
            // Surplus credited to insurance fund
            self.insurance_fund.balance_usd += deficit_usd;
            Ok(format!("Surplus of ${:.2} credited to Insurance Fund", deficit_usd))
        }
    }
}
