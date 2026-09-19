# Growww Flutter Multi-Platform Client Application

## Purpose & Scope
`growww_flutter` is the sovereign, multi-platform client application serving iOS, Android, macOS, Windows, and Linux. Built with Flutter 3.19+, it provides a unified codebase for both mobile retail trading and multi-window desktop institutional terminals.

## Key Architectural Features
- **Obsidian Dark Theme**: Deep Obsidian (`#0B0E14`) theme with neon green (`#00F0A0`) and neon red (`#FF3B56`) accents.
- **TradingView Integration**: Low-latency candlestick charting engine with 100+ technical indicators.
- **Demo / Real Switcher**: Instant one-tap toggle between Virtual Paper Trading and Real-Money Blockchain Trading.
- **Biometric Quick-Auth**: Apple FaceID / TouchID and Android BiometricPrompt for cryptographic trade signing.
- **Offline Queued Orders**: Resilient local SQLite queue holding orders during intermittent network disconnections.

## Directory Structure
- `lib/screens/trading/`: BTC/USDT spot trade screens, buy/sell sheets, and orderbook widgets.
- `lib/screens/demo/`: Paper trading dashboard, PnL analytics, and one-click faucet claim button.
- `lib/screens/wallet/`: Multichain crypto deposit addresses (Taproot BTC, USDT) and instant INR on-ramp.
- `lib/screens/settings/`: User security center, TOTP, Passkeys, anti-phishing codes, and active sessions.
- `lib/controllers/`: Riverpod state management controllers for market data, orders, and portfolio.

## Local Testing
```bash
flutter pub get
flutter run -d chrome  # or -d linux / macos / windows
```
