# Demo Trading Faucet & Virtual Wallet Service Architecture

## 1. Architecture Overview
The Demo Faucet and Virtual Wallet Service manages testnet asset distributions, demo balance tracking, and one-click balance resets for the risk-free paper trading platform.

## 2. Virtual Asset Allocation & Rate Limiting
- **Default Starting Balance**: 10,000 vUSDT and 1.0 vBTC credited to TigerBeetle `Ledger ID 2` upon user onboarding.
- **Refill Policy**: Users can claim additional test funds once every 24 hours if their balance drops below 1,000 vUSDT.
- **Anti-Abuse Controls**: Faucet claims require authenticated session tokens, device fingerprint validation, and reCAPTCHA v3 verification to eliminate Sybil spam.

## 3. One-Click Reset Mechanics
Users can reset their paper trading account at any time:
1. Existing open demo orders are canceled in the `demo-matching-execution-engine`.
2. TigerBeetle demo accounts are cleared and reset to initial allocation (10,000 vUSDT and 1.0 vBTC).
3. Portfolio performance history, PnL analytics, and virtual trade logs are archived for user reference.
