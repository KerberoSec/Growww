#[derive(Debug, Clone)]
pub struct HistoricalTick {
    pub sequence: u64,
    pub symbol: String,
    pub price_e8: u64,
    pub quantity_e8: u64,
    pub timestamp_ns: u64,
}

pub struct MarketReplayDaemon {
    pub recorded_ticks: Vec<HistoricalTick>,
    pub playback_cursor: usize,
    pub speed_multiplier: f64,
}

impl MarketReplayDaemon {
    pub fn new(ticks: Vec<HistoricalTick>, speed: f64) -> Self {
        Self {
            recorded_ticks: ticks,
            playback_cursor: 0,
            speed_multiplier: speed,
        }
    }

    /// Read next batch of historical ticks for deterministic offline exchange simulation and audit
    pub fn step_next(&mut self, batch_size: usize) -> Vec<HistoricalTick> {
        if self.playback_cursor >= self.recorded_ticks.len() {
            return Vec::new();
        }

        let end = std::cmp::min(self.playback_cursor + batch_size, self.recorded_ticks.len());
        let slice = self.recorded_ticks[self.playback_cursor..end].to_vec();
        self.playback_cursor = end;
        slice
    }

    pub fn reset(&mut self) {
        self.playback_cursor = 0;
    }
}
