#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PositionSide {
    Long,
    Short,
}

#[derive(Debug, Clone)]
pub struct PerpetualPosition {
    pub position_id: String,
    pub user_id: String,
    pub symbol: String,
    pub side: PositionSide,
    pub size_contracts: u64,
    pub entry_price_e8: u64,
    pub leverage: u32,
    pub isolated_margin_e8: u64,
    pub liquidation_price_e8: u64,
}

pub struct PerpetualsEngine {
    pub maintenance_margin_rate: f64, // e.g. 0.005 for 0.5% (200x max leverage)
}

impl PerpetualsEngine {
    pub fn new(mmr: f64) -> Self {
        Self {
            maintenance_margin_rate: mmr,
        }
    }

    /// Calculate unrealized PnL for linear USD-M perpetual futures
    pub fn calculate_unrealized_pnl(&self, pos: &PerpetualPosition, mark_price_e8: u64) -> i64 {
        let entry = pos.entry_price_e8 as i64;
        let mark = mark_price_e8 as i64;
        let size = pos.size_contracts as i64;

        match pos.side {
            PositionSide::Long => (mark - entry) * size / 100_000_000,
            PositionSide::Short => (entry - mark) * size / 100_000_000,
        }
    }

    /// Calculate deterministic liquidation price for isolated position
    pub fn calculate_liquidation_price(&self, pos: &PerpetualPosition) -> u64 {
        let entry = pos.entry_price_e8 as f64;
        let mmr = self.maintenance_margin_rate;
        let lev = pos.leverage as f64;

        let liq = match pos.side {
            PositionSide::Long => entry * (1.0 - (1.0 / lev) + mmr),
            PositionSide::Short => entry * (1.0 + (1.0 / lev) - mmr),
        };

        liq.max(0.0) as u64
    }
}
