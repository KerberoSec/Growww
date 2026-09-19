# 515 - Flutter Offline Resilient & Queued Order (AMO) Management UX

## Purpose
In Indian equity markets, trading hours are strictly bounded (09:15 to 15:30 IST for normal trading, with pre-open from 09:00 to 09:08 IST). Outside these market hours, retail and institutional investors place After Market Orders (AMO) for fractional digital security tokens. Additionally, mobile and desktop clients operate in volatile real-world network conditions where intermittent connectivity loss during trade placement can lead to duplicate orders or investor uncertainty.

This prompt specifies the client-side offline and queued order engine. It provides an explicit After Market Order (AMO) experience during market-closed windows and an offline transactional queue for transient network disruptions, guaranteeing exactly-once client idempotency, transparent queue visibility, execution progress indicators, and seamless server reconciliation upon reconnection or market open.

## What You Are Building
An offline resilience and order queue module located in `lib/features/orders/offline_queue/`:
- `MarketScheduleProvider`: Real-time reactive provider evaluating exchange trading sessions (Pre-Open, Regular, Post-Close, AMO Window, Weekend/Holiday) using localized IST clocks and server time synchronization.
- `AmoOrderBannerWidget`: Prominent contextual UI banner informing users of current market closure, next scheduled execution window (e.g. 09:15 IST next trading day), and AMO rules.
- `LocalOrderQueueRepository`: Encrypted local persistence layer (using Drift SQLite with SQLCipher or Isar) managing pending, queued, syncing, and failed order states.
- `OfflineOrderQueueScreen`: Dedicated view within the Portfolio/Orders tab showing all pending offline/queued AMOs with real-time status badges, price limits, and instant cancellation controls prior to market submission.
- `OrderSyncEngine`: Background and foreground synchronization manager that automatically batches, validates, and dispatches queued orders with client-generated idempotency keys when connectivity and market sessions open.
- `ConnectivityAwareOrderSubmissionHandler`: Interceptor wrapping standard buy/sell order submissions to intelligently switch between synchronous execution and offline/AMO queuing.

## Scope Boundaries
- **In Scope:**
 - Client-side market session schedule calculations and time drift correction with backend NTP/timestamp.
 - Encrypted local persistence for queued orders across app restarts.
 - Idempotency key generation (`UUIDv4` + order parameters hash) for safe network retries.
 - UI notifications, offline queue status drawer, and individual queued order cancellation.
 - Synchronization state machine (Pending -> Syncing -> Submitted -> Executed / Rejected).
- **Out of Scope / Handled Elsewhere:**
 - Backend order matching and server-side AMO scheduling (Prompt 204, Prompt 205).
 - Pre-trade risk validation and margin holds on server (Prompt 206).
 - On-chain DvP smart contract execution (Prompt 208, Prompt 306).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (2.5+) and `drift` (v2.18+) with `sqlite3_flutter_libs` and `sqlcipher` for encrypted local persistence.
- **Justification:** Drift provides type-safe Dart SQL queries, reactive reactive streams (`watch()`), migration support, and native desktop (Windows/Linux/macOS) and mobile (Android/iOS) embedded database compilation.
- **Dependencies:**
 - `drift: ^2.18.0`
 - `sqlite3_flutter_libs: ^0.5.24`
 - `connectivity_plus: ^6.0.3`
 - `uuid: ^4.4.0`
 - `intl: ^0.19.0` (IST timezone calculation)

## Backend / Infra Touchpoints
- **Market Data & Session Service (Prompt 207):** `GET /api/v1/market/schedule` (trading holidays, session hours, server UTC timestamp).
- **Order Ingestion Service (Prompt 204):** `POST /api/v1/orders/amo` and `POST /api/v1/orders/batch-sync` with `X-Idempotency-Key` headers.
- **Wallet & Margin Service (Prompt 203):** `GET /api/v1/wallet/provisional-holds` for verifying local margin before queuing.

## Blockchain Interaction
- **Pre-Settlement Verification:** Queued fractional equity orders represent intent to settle atomic DvP transfers on the permissioned Hyperledger Besu network once markets open.
- **Idempotency Reconciliation:** The client's generated `order_client_nonce` is embedded into the off-chain order payload and subsequently indexed alongside the on-chain trade hash (`SettlementDvP.sol`) to guarantee that retried offline orders never execute duplicate blockchain settlement events.

## Step-by-Step Build Instructions
1. Scaffold `lib/features/orders/offline_queue/` with subfolders for `data/`, `domain/`, `application/`, and `presentation/`.
2. Define Drift database table `QueuedOrdersTable` containing columns: `id`, `client_order_id`, `isin`, `symbol`, `side` (BUY/SELL), `order_type` (LIMIT/MARKET), `quantity_fractional`, `limit_price_inr`, `status` (QUEUED, SYNCING, SUBMITTED, FAILED), `created_at`, `idempotency_key`, `error_message`, and `retry_count`.
3. Configure SQLCipher database encryption using a dynamically derived key retrieved securely from `SecureStorageService` (Prompt 521).
4. Implement `MarketScheduleManager` computing market state (`OPEN`, `PRE_OPEN`, `POST_CLOSE`, `AMO_ACCEPTING`, `CLOSED`) based on IST calendar, holidays, and server clock offset.
5. Create `ConnectivityService` wrapping `connectivity_plus` and active heartbeat pings to detect genuine backend reachability.
6. Build `OrderSyncEngine` class that listens to both `ConnectivityService` and `MarketScheduleManager` events to trigger automatic queue flushing.
7. Implement client-side pre-flight validations (cash balance check against cached wallet balance, lot size constraints, price band checks).
8. Implement `OfflineOrderQueueNotifier` using Riverpod to expose a stream of pending orders and aggregate queue health.
9. Create `AmoOrderBannerWidget` displayed at the top of the Buy/Sell order sheet (Prompt 509) when `MarketSchedule` is outside normal trading hours.
10. Build `OfflineOrderQueueScreen` listing all queued AMOs with cancel action buttons, price limits, creation timestamps, and target execution windows.
11. Implement atomic order cancellation: if an order is in `QUEUED` state, delete locally; if in `SYNCING` state, send an immediate cancel RPC to backend.
12. Build toast/banner notifications informing user when queued orders successfully transition from `QUEUED` to `SUBMITTED` upon market opening.
13. Implement exponential backoff retry logic (1s, 2s, 4s, max 30s) for network failures occurring during market-open sync.
14. Write unit tests for `MarketScheduleManager` covering standard trading days, weekends, Diwali Muhurat trading exceptions, and leap years.
15. Write integration tests simulating network drop, queuing 3 orders, restoring network, and verifying exact order sequence and idempotency keys sent to mock backend.

## Interfaces / Contracts

```dart
// lib/features/orders/offline_queue/domain/models/market_session.dart
enum MarketSessionStatus {
  preOpen,      // 09:00 - 09:08 IST
  open,         // 09:15 - 15:30 IST
  postClose,    // 15:40 - 16:00 IST
  amoWindow,    // 16:30 - 08:59 IST (Next Day)
  holidayClosed // Weekend or NSE/BSE Holiday
}

class MarketScheduleState {
  final MarketSessionStatus status;
  final DateTime serverTimeUtc;
  final DateTime nextSessionOpenTime;
  final bool allowsAmo;

  MarketScheduleState({
    required this.status,
    required this.serverTimeUtc,
    required this.nextSessionOpenTime,
    required this.allowsAmo,
  });
}

// lib/features/orders/offline_queue/domain/models/queued_order.dart
enum QueuedOrderStatus { queuedAmo, queuedOffline, syncing, submitted, rejected }

class QueuedOrder {
  final String clientOrderId;
  final String isin;
  final String symbol;
  final String side; // 'BUY' | 'SELL'
  final String orderType; // 'LIMIT' | 'MARKET'
  final double fractionalQuantity;
  final double? limitPriceInr;
  final QueuedOrderStatus status;
  final String idempotencyKey;
  final DateTime queuedAt;
  final String? rejectionReason;

  QueuedOrder({
    required this.clientOrderId,
    required this.isin,
    required this.symbol,
    required this.side,
    required this.orderType,
    required this.fractionalQuantity,
    this.limitPriceInr,
    required this.status,
    required this.idempotencyKey,
    required this.queuedAt,
    this.rejectionReason,
  });
}

// Order Sync Engine Interface
abstract class OrderSyncEngineContract {
  Future<void> queueOrder(QueuedOrder order);
  Future<void> cancelQueuedOrder(String clientOrderId);
  Stream<List<QueuedOrder>> watchPendingOrders();
  Future<SyncResult> triggerSync();
}
```

## Security & Compliance Notes
- **SEBI AMO Compliance:** Enforce that AMOs are explicitly labeled as After Market Orders with mandatory disclosure of potential opening volatility and price slippage.
- **Idempotency Guarantees:** Every queued order must generate an unguessable `idempotency_key` combining user UUID, client timestamp, and SHA-256 hash of order attributes to prevent duplicate trade executions under unstable networks.
- **Local Storage Encryption:** Queued order details contain sensitive financial intents and must be encrypted at rest using SQLCipher with hardware-backed encryption keys (Prompt 521).
- **Audit Logging:** Every transition (queued, retried, canceled, submitted) must generate a client-side telemetry event with timestamp and network state for compliance audit trails.

## Acceptance Criteria
- [ ] System accurately identifies market open/closed status in Indian Standard Time (IST) factoring in national trading holidays.
- [ ] Placing an order outside market hours displays clear AMO disclaimer and successfully stores the order in encrypted local storage.
- [ ] Offline Order Queue UI allows investors to review all pending orders and cancel any order before market submission.
- [ ] Seamless automatic synchronization triggers when network connectivity is restored during trading hours without user intervention.
- [ ] Backend receives identical `X-Idempotency-Key` across retries, guaranteeing zero duplicate orders.
- [ ] Unit and integration tests verify offline queue persistence across simulated app kills and restarts.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 204 (Order Service), Prompt 502 (App Architecture & State Management), Prompt 509 (Order Placement Flow), Prompt 521 (Local Secure Storage).
- **Parallel Tasks:** Prompt 507 (Market/Watchlist Screen), Prompt 525 (API Client Layer).
