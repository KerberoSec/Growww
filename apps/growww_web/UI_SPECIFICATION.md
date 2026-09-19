# Pro Web Trading Terminal UI & Component Specification

## Component Architecture
- `DockviewWorkspace`: Primary multi-split grid container managing draggable, dockable, and floating panels.
- `DetachablePopoutWindow`: Autonomous browser window wrapper synchronizing state via `BroadcastChannel` and `SharedWorker`.
- `TradingViewBridge`: Canvas wrapper embedding TradingView Advanced Charts with custom binary WebSocket datafeed adapters.
- `WebGLDepthLadder`: Hardware-accelerated Depth of Market (DOM) click-to-trade vertical ladder rendering at 60/120 FPS.
- `OrderbookTable`: Zero-layout-shift tabular numbers with microsecond green/red depth bar visual distribution.
- `ExecutionBlotterTable`: Live positions blotter with real-time PnL, TWAP/VWAP slicing status, and Besu transaction hash links.
- `QuickOrderBar`: Top-bar keyboard shortcut target allowing instant limit/market order entry with zero-fee badges (`0.00%`).
- `LinkGroupSelector`: Color-coded dropdown (Red, Blue, Green, Yellow) locking symbol synchronization across detached windows.

## Keyboard Hotkey Map
- `Shift + B`: Instant Buy Market Order form execution.
- `Shift + S`: Instant Sell Market Order form execution.
- `Escape`: Cancel all active open orders in current market.
- `Shift + Escape`: Global Panic Cancel across all pairs.
- `Space`: Focus search bar to switch trading pairs.
- `1` through `6`: Switch chart timeframes (1m, 5m, 15m, 1h, 4h, 1D).
- `Ctrl + D`: Reset workspace layout to default preset.
- `F11`: Toggle native browser full-screen trading mode.
