# 507 - Flutter Market & Watchlist Screen with Real-Time Price Updates

## Purpose
Provides an institutional-grade market discovery and watchlist monitoring interface for fractional securities, equities, and exchange-traded funds. Investors need sub-100ms real-time price updates, customizable watchlists, category filters, and intuitive search to monitor price movements, track tokenized Indian blue-chip equities, and identify investment opportunities across domestic and GIFT City markets.

## What You Are Building
A high-throughput market and watchlist module in `apps/growww_flutter/lib/features/watchlist/` containing:
- **Customizable Watchlist Tabs:** Create, rename, reorder, and delete custom user watchlists (e.g., "Nifty 50 Bluechips", "High Dividend", "GIFT City Exclusives").
- **Sub-100ms Streaming Price Tickers:** Ticker tiles that dynamically flash green (uptick) or red (downtick) on price changes received over WebSocket.
- **Search & Filter Bar:** Instant search with debouncing, filtering by market cap, sector, 1-day performance, and tokenized status.
- **Swipe Actions & Quick Add:** Swipe-to-delete, swipe-to-buy/sell shortcuts on mobile, and right-click context menus on desktop.
- **Market Status Header:** Live indicator of exchange market status (NSE/BSE/GIFT IFSCA: Open, Closed, Pre-Open, Extended Trading).

## Scope Boundaries
- **In Scope:**
 - Watchlist UI tabs, real-time tick animation handling, search indexing, list reordering, and WebSocket stream subscription management.
- **Out of Scope / Handled Elsewhere:**
 - Full candlestick technical chart analysis (handled in Prompt 508).
 - Order execution sheet (handled in Prompt 509).
 - Backend WebSocket market broadcast server (handled in Prompt 207).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` stream controllers.
 - *Justification:* Flutter's fine-grained widget rebuild capabilities allow animating individual stock ticker cells upon price change without invalidating the entire list, ensuring sustained 120 FPS performance during market-open surges.
- **WebSocket Client:** `web_socket_channel` with automated connection pooling and heartbeat ping/pong.
- **Reordering UI:** `ReorderableListView` with smooth drag-and-drop animations across mobile and desktop.

## Backend / Infra Touchpoints
- **Market Data Service:** WebSocket channel (`wss://ws.growww.in/v1/market/stream`) and REST search API (`/api/v1/market/securities/search`) from Prompt 207.
- **User Watchlist API:** REST endpoints (`/api/v1/user/watchlists`) for syncing custom watchlists to the user profile.

## Blockchain Interaction
- **Tokenized Asset Metadata Tagging:** Each security tile displays its underlying digital token contract identifier (`DigitalSecurityToken.sol` ERC-3643 address) alongside its official NSE/BSE ISIN, clearly signaling to the user that the asset is tokenized and backed 1:1 by segregated custodial shares.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/watchlist/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/sources/`.
2. Define domain models: `Watchlist`, `SecurityTickerItem`, `PriceTickEvent`, `MarketStatus`.
3. Implement `WatchlistWebSocketService` managing subscription messages (`{"action": "subscribe", "symbols": [...]}`).
4. Build `WatchlistScreen` with a top `TabBar` displaying user watchlists and an "Add Watchlist" button.
5. Create `SecurityTickerTile` widget rendering security symbol, company name, current price (in INR), daily percentage change, and fractional token badge.
6. Implement `PriceFlashAnimation` widget that flashes a subtle green/red background highlight for 400ms when a new price tick arrives.
7. Build `WatchlistSearchDelegate` / search modal with 250ms debouncing, querying the search microservice and highlighting matching terms.
8. Implement `ReorderableWatchlist` enabling users to long-press and drag securities to organize their watchlists.
9. Add swipe-to-action handlers: swiping left exposes "Remove", swiping right opens "Quick Buy" (Prompt 509).
10. Build desktop-adaptive enhancements: right-click context menu ("View Details", "Buy", "Sell", "Copy ISIN", "Verify On-Chain"), and keyboard up/down navigation.
11. Implement market status badge (e.g., "Market Open - Closes at 3:30 PM IST").
12. Write widget tests verifying price tick rendering, reordering logic, and search debouncing.

## Interfaces / Contracts
```dart
// lib/features/watchlist/domain/models/security_ticker.dart
class SecurityTickerItem {
  final String symbol;
  final String isin;
  final String companyName;
  final double ltp; // Last Traded Price
  final double changeInr;
  final double changePercentage;
  final double dayHigh;
  final double dayLow;
  final double volume;
  final String tokenContractAddress;
  final bool isFractionalEnabled;

  const SecurityTickerItem({
    required this.symbol,
    required this.isin,
    required this.companyName,
    required this.ltp,
    required this.changeInr,
    required this.changePercentage,
    required this.dayHigh,
    required this.dayLow,
    required this.volume,
    required this.tokenContractAddress,
    required this.isFractionalEnabled,
  });
}
```

## Security & Compliance Notes
- **Delayed Quote Indicator:** If market data stream fallback occurs (switching from live WebSocket to polled HTTP), the UI must immediately display a *"Delayed Quotes (15 min)"* warning badge as required by exchange compliance.
- **ISIN Verifiability:** Every security must display its certified SEBI ISIN to prevent deceptive token symbol spoofing.
- **No Unregulated Tokens:** Only securities registered on the platform's compliance master list can be searched or added.

## Acceptance Criteria
- [ ] Watchlists render price updates within 100ms of WebSocket message receipt.
- [ ] Ticker cells trigger subtle green/red flash animations without frame drops.
- [ ] Users can create, rename, reorder, and delete custom watchlists with cloud sync.
- [ ] Search returns results with <250ms latency across 5,000+ listed securities.
- [ ] Desktop right-click context menu opens responsive action options.
- [ ] Market closed / pre-market states are clearly visually indicated.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System).
- **Backend Dependency:** Prompt 207 (Market Data Service), Prompt 221 (Search Service).
- **Enables:** Prompt 508 (Security Detail Screen), Prompt 509 (Order Placement Flow).
