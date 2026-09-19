use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct FollowerConfig {
    pub follower_user_id: String,
    pub allocation_ratio: f64, // e.g. 0.10 means copy with 10% of master's position size
    pub max_notional_usd: f64,
    pub profit_share_bps: u32, // e.g. 1000 = 10% high-water mark profit sharing
}

pub struct CopyTradingReplicationEngine {
    pub followers: HashMap<String, Vec<FollowerConfig>>, // master_id -> list of followers
}

impl CopyTradingReplicationEngine {
    pub fn new() -> Self {
        Self {
            followers: HashMap::new(),
        }
    }

    pub fn register_follower(&mut self, master_id: &str, config: FollowerConfig) {
        self.followers
            .entry(master_id.to_string())
            .or_default()
            .push(config);
    }

    /// Calculate replicated child orders when master executes a spot or futures trade
    pub fn replicate_order(
        &self,
        master_id: &str,
        symbol: &str,
        master_qty_e8: u64,
        price_usd: f64,
    ) -> Vec<(String, u64)> {
        let mut child_orders = Vec::new();

        if let Some(follower_list) = self.followers.get(master_id) {
            for f in follower_list {
                let scaled_qty = (master_qty_e8 as f64 * f.allocation_ratio) as u64;
                let notional = (scaled_qty as f64 / 1e8) * price_usd;

                if notional <= f.max_notional_usd && scaled_qty > 0 {
                    child_orders.push((f.follower_user_id.clone(), scaled_qty));
                }
            }
        }

        child_orders
    }
}
