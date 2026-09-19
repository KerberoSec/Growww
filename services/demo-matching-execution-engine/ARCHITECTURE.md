# Isolated Paper Trading Simulation Engine Architecture

## 1. Architecture Overview
A standalone Rust in-memory matching engine executing simulated paper trades against real-time market data mirrored from Binance and Coinbase Pro.

## 2. Complete Physical & Logical Isolation
- **Ledger Separation**: Uses TigerBeetle `Ledger ID 2` (Demo Paper Trading). All balance reservations, trades, and fees affect virtual balances only.
- **Messaging Separation**: Listens and publishes exclusively to Kafka topics prefixed with `growww.demo.*`.
- **Zero Real Liquidity Impact**: Demo trades do not cross into the production orderbook and cannot influence real spot asset prices.

## 3. Market Mirroring & Latency Arbitrage Protection
- Mirrors real-time BTC/USDT orderbook depth and trades via high-speed WebSocket feeds from external exchanges.
- **Staleness Circuit Breaker**: If external market feeds disconnect or report timestamp staleness $> 2,000\text{ms}$, demo order execution is temporarily paused to prevent stale-quote exploitation.
- **Fill Realism**: Simulated limit orders execute only when real market trades cross the specified limit price with sufficient visible depth.

## 4. Performance & Scale
- Processes up to 500,000 simulated orders/second.
- Zero gas fees and zero on-chain transaction delays for demo users.
