# Universal Flutter UI Component & Screen Specification (Desktop & Mobile)

## Desktop Thick Client Screen & Window Catalog (macOS, Windows, Linux)
1. **DesktopMasterShellWindow**: Multi-monitor terminal shell housing window coordinate management, global status indicators, and workspace layout presets.
2. **DetachableDOMDepthLadderWindow**: Native OS window rendering high-frequency DOM price ladder with static price rungs, dynamic bid/ask depth bars, and 1-click execution.
3. **DetachableOrderBookWindow**: Dedicated auxiliary window rendering 50-level Level 2/3 order book with visual depth distributions and microsecond trade tape.
4. **DetachableTradingViewChartWindow**: High-refresh 4K canvas charting window supporting multi-timeframe analysis and customized indicators.
5. **InstitutionalExecutionBlotterWindow**: Auxiliary monitor blotter showing open orders, trade fills, net positions, realized/unrealized PnL, and Besu explorer transaction links.
6. **SettlementAuditMonitorWindow**: Live Hyperledger Besu blockchain monitor with block confirmation progress, validator metrics, and ERC-4337 Paymaster zero-gas sponsorship badges.

## Mobile Client Screen Catalog (iOS & Android)
1. **SpotTradeScreen**: Full-height sliver layout integrating TradingView charts, orderbook ladders, and sticky bottom order drawers.
2. **OrderEntrySheet**: Modal bottom sheet featuring 25%, 50%, 75%, 100% quick chips, fluid percentage slider, Swipe-to-Trade button, and zero-fee badge.
3. **DemoModeHUD**: Persistent golden amber banner with one-click 10,000 vUSDT / 1 vBTC virtual faucet claim modal.
4. **UnifiedPortfolioScreen**: Multi-asset donut chart combining traditional Demat equities with real-time crypto spot holdings.
5. **SecurityCenterScreen**: Biometric authentication toggles (TouchID, Windows Hello, FaceID, BiometricPrompt), TOTP pairing, Passkeys, and active sessions.

## Interactive Gestures & Desktop Hotkeys
- **Desktop DOM Click-to-Trade**: Left-click on bid/ask column submits limit order; right-click cancels; space auto-centers.
- **Desktop Global Shortcuts**: `Ctrl+Alt+Space` (or `Cmd+Opt+Space`) panic cancels all active orders across the exchange.
- **Mobile Swipe-to-Trade**: Thumb-driven horizontal slider requiring full drag gesture to prevent accidental orders.
- **Mobile Tap-to-Fill**: Tapping any orderbook bid or ask level populates the price and quantity fields immediately.
- **Pull-to-Refresh**: Smooth physics-based spring curve refreshing account balances and open order statuses.
