use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct TraderPnLSummary {
    pub user_id: String,
    pub total_trades: u32,
    pub winning_trades: u32,
    pub total_realized_pnl_usd: f64,
    pub win_rate_pct: f64,
    pub current_streak: u32,
    pub leaderboard_rank: u32,
}

pub struct DemoAnalyticsEngine {
    pub profiles: HashMap<String, TraderPnLSummary>,
}

impl DemoAnalyticsEngine {
    pub fn new() -> Self {
        Self {
            profiles: HashMap::new(),
        }
    }

    pub fn record_closed_trade(&mut self, user_id: &str, realized_pnl: f64) {
        let profile = self.profiles.entry(user_id.to_string()).or_insert_with(|| TraderPnLSummary {
            user_id: user_id.to_string(),
            total_trades: 0,
            winning_trades: 0,
            total_realized_pnl_usd: 0.0,
            win_rate_pct: 0.0,
            current_streak: 0,
            leaderboard_rank: 0,
        });

        profile.total_trades += 1;
        profile.total_realized_pnl_usd += realized_pnl;

        if realized_pnl > 0.0 {
            profile.winning_trades += 1;
            profile.current_streak += 1;
        } else {
            profile.current_streak = 0;
        }

        profile.win_rate_pct = (profile.winning_trades as f64 / profile.total_trades as f64) * 100.0;
    }

    pub fn get_leaderboard(&self, limit: usize) -> Vec<TraderPnLSummary> {
        let mut board: Vec<TraderPnLSummary> = self.profiles.values().cloned().collect();
        board.sort_by(|a, b| b.total_realized_pnl_usd.partial_cmp(&a.total_realized_pnl_usd).unwrap());

        for (idx, trader) in board.iter_mut().enumerate() {
            trader.leaderboard_rank = (idx + 1) as u32;
        }

        board.into_iter().take(limit).collect()
    }
}
