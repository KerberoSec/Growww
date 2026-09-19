# Growww Pro Web Trading Terminal (Next.js 14)

## Purpose & Scope
`growww_web` is an institutional-grade web trading terminal built on Next.js 14 (App Router), Tailwind CSS, and TypeScript. It offers a customizable multi-window workspace designed for high-frequency day traders and algorithmic market makers.

## Key Architectural Features
- **Multi-Monitor Window Popping**: Detachable charts, orderbooks, and watchlists for multi-monitor desktop setups.
- **Microsecond Tabular Layouts**: Zero-layout-shift tabular number rendering using JetBrains Mono.
- **Keyboard Hotkeys**: Shift+B (Buy Market), Shift+S (Sell Market), Escape (Cancel All Orders).
- **Conflated WebSocket Stream**: Smooth 20 FPS Level 2 depth ladder updates without browser thread locking.

## Directory Structure
- `src/app/trade/btc-usdt/`: Main BTC/USDT pro trading workstation.
- `src/app/demo/`: Interactive paper trading simulator and strategy backtester.
- `src/app/wallet/`: Non-custodial Web3 wallet connection and fiat INR deposit management.
- `src/app/settings/`: Security preferences, API key management with IP whitelisting, and DPDP privacy export.
- `src/components/tradingview/`: Custom TradingView Advanced Charts integration.

## Local Testing
```bash
pnpm install
pnpm dev # Serves at http://localhost:3000
```
