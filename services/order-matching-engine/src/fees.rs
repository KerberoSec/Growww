use std::sync::atomic::{AtomicU16, Ordering};

pub static MAKER_FEE_BPS: AtomicU16 = AtomicU16::new(0);
pub static TAKER_FEE_BPS: AtomicU16 = AtomicU16::new(0);

pub const BASIS_POINTS_DIVISOR: u128 = 10_000;
pub const BASE_DECIMALS_FACTOR: u128 = 100_000_000; // 1e8 precision

#[derive(Debug, Clone, Default)]
pub struct FeeEngine;

impl FeeEngine {
    pub fn new() -> Self {
        Self
    }

    #[inline(always)]
    pub fn get_maker_fee_bps() -> u16 {
        MAKER_FEE_BPS.load(Ordering::Relaxed)
    }

    #[inline(always)]
    pub fn get_taker_fee_bps() -> u16 {
        TAKER_FEE_BPS.load(Ordering::Relaxed)
    }

    #[inline(always)]
    pub fn set_maker_fee_bps(bps: u16) {
        MAKER_FEE_BPS.store(bps, Ordering::SeqCst);
    }

    #[inline(always)]
    pub fn set_taker_fee_bps(bps: u16) {
        TAKER_FEE_BPS.store(bps, Ordering::SeqCst);
    }

    /// Calculate quote amount and fees using checked fixed-point arithmetic
    /// QuoteAmount = ((quantity_e8 * price_e8) / 10^8)
    /// Fee = (QuoteAmount * fee_bps) / 10,000
    #[inline(always)]
    pub fn calculate_trade_financials(
        price_e8: u64,
        quantity_e8: u64,
        maker_bps: u16,
        taker_bps: u16,
    ) -> (u64, u64, u64) {
        let p = price_e8 as u128;
        let q = quantity_e8 as u128;

        // Quote amount in e8 fixed-point
        let quote_amount = (p.saturating_mul(q)) / BASE_DECIMALS_FACTOR;

        let maker_fee = (quote_amount.saturating_mul(maker_bps as u128)) / BASIS_POINTS_DIVISOR;
        let taker_fee = (quote_amount.saturating_mul(taker_bps as u128)) / BASIS_POINTS_DIVISOR;

        (quote_amount as u64, maker_fee as u64, taker_fee as u64)
    }
}
