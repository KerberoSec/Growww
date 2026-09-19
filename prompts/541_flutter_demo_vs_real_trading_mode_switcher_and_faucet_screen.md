# 541 - Flutter Demo / Paper Trading Mode Switcher & Faucet Screen

## Purpose
Enabling users to toggle seamlessly between "Demo Paper Trading (Testnet)" and "Real-Money Trading (Mainnet)" in the Flutter app with a persistent environment banner, 1-click 10,000 USDT testnet faucet, and reset portfolio capability.

In high-stakes financial and digital asset trading, novice and experienced traders require an authentic, zero-risk sandbox to master complex order mechanics (such as bracket orders, stop-loss triggers, trailing stops, options chains, and tokenized equity lots), test algorithmic strategies, and familiarize themselves with real-time market data without risking personal capital. Traditional financial platforms often force users to download separate demo applications or maintain disparate login credentials, introducing severe friction that hinders onboarding and reduces platform stickiness.

However, operating paper trading and real-money execution within the same mobile and desktop application introduces severe consumer protection and regulatory risks:
- Mode Confusion Risk: A trader who erroneously believes they are trading in demo mode may unknowingly execute catastrophic real-money trades. Conversely, a user who believes they are trading with real funds may feel misled when discovering their profits cannot be withdrawn.
- Data and Authorization Contamination: Mixing production and testnet authentication tokens or routing paper orders to live matching engines can compromise system integrity and violate regulatory audit standards.

This specification defines the architecture, state management, and user interface for an enterprise-grade trading mode switcher. It establishes a high-visibility, persistent amber/yellow banner across all relevant trading screens, provides a 1-click 10,000 USDT testnet faucet with cooldown enforcement, offers an instant portfolio reset routine, and enforces strict cryptographic and network token segregation between the Hyperledger Besu Testnet and Mainnet environments.

## What You Are Building
A modular, high-performance trading mode and testnet faucet feature located in `lib/screens/trading/` and `lib/features/trading_mode/`:
- `TradingModeSwitcherBanner` (`lib/screens/trading/mode_switcher_banner.dart`): A persistent, high-visibility amber/yellow banner anchored to the top of all market watchlists, order tickets, active position lists, and portfolio overviews whenever Demo mode is active. It displays the "DEMO / PAPER TRADING" badge, live virtual USDT balance, a 1-tap "Claim 10k Faucet" button, a "Reset Portfolio" trigger, and an "Exit Demo" toggle.
- `FaucetDialog` (`lib/screens/trading/faucet_dialog.dart`): An interactive modal bottom sheet offering a 1-click claim button for 10,000 USDT testnet tokens, an animated cooldown progress ring (24-hour rate limit), recent faucet claim receipts, and a one-tap "Reset All Paper Assets" action.
- `TradingModeToggleWidget` (`lib/screens/trading/widgets/trading_mode_toggle_widget.dart`): A fluid, animated segmented pill widget for app headers, navigation drawers, and settings screens, supporting tactile haptic feedback and triggering explicit confirmation modals on mode transitions.
- `ResetPortfolioDialog` (`lib/screens/trading/widgets/reset_portfolio_dialog.dart`): A confirmation dialog warning the user that resetting the demo portfolio permanently wipes all simulated positions, open orders, and paper PnL, re-allocating a pristine 10,000 USDT starting balance.
- `TradingModeController` (`lib/features/trading_mode/controllers/trading_mode_controller.dart`): A Riverpod `AsyncNotifier` managing `TradingModeState`, coordinating mode persistence, faucet cooldown timers, balance updates, and reactive notifications across dependent providers.
- `DemoWalletApiClient` (`lib/features/trading_mode/data/demo_wallet_api_client.dart`): An HTTP API client interfacing with the backend Demo Wallet Service (Prompt 274) for balance queries, faucet claims, and portfolio resets.
- `TradingModeStorage` (`lib/features/trading_mode/data/trading_mode_storage.dart`): A secure local storage repository backed by `flutter_secure_storage` ensuring mode preferences and demo tokens remain isolated from production keys.
- `TradingModeNetworkInterceptor` (`lib/features/trading_mode/data/trading_mode_interceptor.dart`): A Dio HTTP interceptor dynamically injecting demo vs real authorization headers and routing requests to the appropriate gateway endpoints.

## Scope Boundaries
- In Scope:
  - Flutter UI widgets: `TradingModeSwitcherBanner`, `FaucetDialog`, `TradingModeToggleWidget`, and `ResetPortfolioDialog`.
  - Reactive state management with Riverpod: mode switching, virtual balance syncing, faucet cooldown timer countdown, and loading states.
  - Persistent storage of user mode preferences via `flutter_secure_storage` across application launches.
  - 1-click 10,000 USDT testnet faucet claim action with server-side validation and client-side cooldown handling.
  - Portfolio reset capability wiping paper trade history and reinitializing virtual assets to 10,000 USDT.
  - High-visibility persistent amber/yellow banner rendered across all trading views when Demo mode is active.
  - Subtle visual watermarking on charts, order confirmation sheets, and position cards in Demo mode.
  - Mandatory confirmation modal with risk disclosure when switching from Demo mode to Real mode.
  - Dynamic API client reconfiguration and header segregation between Demo and Real modes.
  - Haptic feedback and smooth animations during mode toggling.
- Out of Scope / Handled Elsewhere:
  - Backend Demo Wallet Service and testnet token minting engine (Prompt 274).
  - Real-money double-entry custodial wallet and cash ledger service (Prompt 203).
  - Developer/QA environment switcher, mock KYC injector, and network inspector (Prompt 527).
  - Core order placement flow, order ticket validation, and order types UI (Prompt 509).
  - High-performance order matching engine (Prompt 205).
  - User identity authentication, JWT renewal, and biometric login orchestration (Prompt 505, Prompt 533).
  - Portfolio holdings accounting and capital gains tax computation screens (Prompt 510, Prompt 523).

## Technology to Use
- Primary Framework: Flutter 3.22+ with Dart 3.4+
- State Management: `flutter_riverpod: ^2.5.1` with code generation support (`riverpod_annotation: ^2.3.5`).
- Secure Local Storage: `flutter_secure_storage: ^9.2.2` for encrypted token and mode persistence, alongside `shared_preferences: ^2.2.3` for non-sensitive UI preferences.
- Networking & HTTP Client: `dio: ^5.4.3+1` for REST endpoints, custom interceptors, and dynamic base URL swapping.
- Precision Financial Arithmetic: `decimal: ^2.3.3` for virtual USDT balances to eliminate floating-point rounding errors.
- Micro-Animations & Haptics: Flutter native `AnimationController`, `Ticker`, and `HapticFeedback` (`services.dart`) for responsive tactile feedback.
- Date & Time Formatting: `intl: ^0.19.0` for formatting cooldown timestamps and claim receipt records.

## Backend / Infra Touchpoints
- Demo Wallet Service (Prompt 274):
  - Manages paper-trading virtual accounts, simulated double-entry ledgers, testnet faucet distributions, and portfolio reset operations.
  - Endpoints:
    - `POST /api/v1/demo/faucet/claim`: Validates eligibility and distributes 10,000 USDT testnet tokens to the user testnet smart account.
    - `GET /api/v1/demo/wallet/balance`: Fetches the current virtual balance breakdown (USDT, tokenized equity shares, paper margins).
    - `POST /api/v1/demo/portfolio/reset`: Cancels open paper orders, liquidates simulated positions, and restores virtual cash to 10,000 USDT.
    - `GET /api/v1/demo/faucet/status`: Returns cooldown expiry timestamp, claim limits, and transaction history.
- Real Wallet Service (Prompt 203):
  - Manages real-money fiat INR and custodial USDT balances backed by audited reserves.
  - Endpoints:
    - `GET /api/v1/wallet/balances`: Retrieves real-money cash and asset balances.
    - `POST /api/v1/wallet/transactions`: Initiates real-money deposit and withdrawal actions.
- Environment Switcher & Sandbox Mode (Prompt 527):
  - Provides core `EnvironmentConfig` infrastructure, RPC gateway definitions, and network client foundation that binds application instances to Testnet or Mainnet clusters.
- API Gateway / BFF (Prompt 219):
  - Enforces route segregation, validates bearer tokens, and guarantees that demo requests are forwarded exclusively to paper-trading matching environments while real requests are routed to production execution venues.

## Blockchain Interaction
Interfacing with Besu Testnet for faucet claims and Besu Mainnet for real funds:
- Permissioned Hyperledger Besu Testnet (Chain ID `13371`, RPC `https://testnet-rpc.growww.in`):
  - Active exclusively when Demo mode is engaged.
  - Faucet claims invoke the testnet `MockUSDT` ERC-20 contract to mint 10,000 testnet USDT directly to the user ERC-4337 smart account.
  - Gas fees are 100% sponsored by the testnet Paymaster, allowing frictionless onboarding with zero testnet gas prerequisites.
  - The client provides clickable transaction receipt hashes linking to the Besu Testnet Block Explorer, educating users on on-chain trade settlement verification.
- Permissioned Hyperledger Besu Mainnet (Chain ID `13370`, RPC `https://rpc.growww.in`):
  - Active exclusively when Real mode is engaged.
  - All trades settle via atomic Delivery versus Payment (DvP) smart contracts backed 1:1 by regulated depository custodians and audited bank reserves.
  - Transactions require hardware-backed passkey or secure biometric authorization (Prompt 533).
- Dynamic Web3 Provider Rebinding:
  - Toggling between Demo and Real modes triggers an atomic update of the Web3 RPC client provider (`web3dart`), swapping the RPC endpoint, chain ID, and contract registry mapping (`USDTToken`, `DvPEngine`, `ComplianceRegistry`) so testnet transactions can never leak into mainnet.

## Step-by-Step Build Instructions (10-15 steps)
1. Scaffold directories: Create `lib/features/trading_mode/models/`, `lib/features/trading_mode/controllers/`, `lib/features/trading_mode/data/`, and `lib/screens/trading/widgets/`.
2. Define domain models and enums in `lib/features/trading_mode/models/trading_mode_models.dart`: Define `TradingMode` (`real`, `demo`), `FaucetStatus` (`ready`, `cooldown`, `claiming`, `failed`), `TradingModeState`, `FaucetClaimResponse`, `DemoPortfolioResetResponse`, and `DemoWalletBalance`.
3. Implement `ITradingModeStorage` and `TradingModeStorage` in `lib/features/trading_mode/data/trading_mode_storage.dart`: Securely persist the active mode (`real` vs `demo`) and demo session tokens using `flutter_secure_storage`.
4. Implement `IDemoWalletApiClient` and `DemoWalletApiClient` in `lib/features/trading_mode/data/demo_wallet_api_client.dart`: Implement REST methods for fetching virtual balances, claiming faucet tokens, inspecting cooldown status, and resetting the demo portfolio.
5. Implement `TradingModeNetworkInterceptor` in `lib/features/trading_mode/data/trading_mode_interceptor.dart`: Attach to the primary `Dio` instance to dynamically rewrite target URLs (e.g. routing `/api/v1/orders` to `/api/v1/demo/orders` when in Demo mode) and inject mode-specific authorization headers.
6. Implement `TradingModeNotifier` in `lib/features/trading_mode/controllers/trading_mode_controller.dart`: Build a Riverpod `AsyncNotifier` managing mode transitions, local cooldown tick timers, faucet claiming workflows, and portfolio reset mutations.
7. Build `TradingModeSwitcherBanner` in `lib/screens/trading/mode_switcher_banner.dart`: Construct an amber status banner (`#FFB300` / `#FFF8E1`) featuring a warning icon, "DEMO MODE" indicator, live balance display, "Claim 10k" button, and "Reset" action.
8. Build `FaucetDialog` in `lib/screens/trading/faucet_dialog.dart`: Develop a modal sheet displaying current demo balance, a prominent "Claim 10,000 USDT" button, an active cooldown countdown timer, and claim history.
9. Build `ResetPortfolioDialog` in `lib/screens/trading/widgets/reset_portfolio_dialog.dart`: Implement an explicit confirmation sheet detailing the destruction of open paper orders and restoring a clean 10,000 USDT ledger.
10. Build `TradingModeToggleWidget` in `lib/screens/trading/widgets/trading_mode_toggle_widget.dart`: Create a sliding segmented toggle with haptic feedback, displaying clear "Real Money" and "Demo Mode" states.
11. Implement Real-Money Transition Safety Guardrail: When switching from Demo to Real, display a modal dialog with an explicit risk disclosure ("You are switching to Real-Money Trading with capital at risk") requiring affirmative confirmation before mode activation.
12. Integrate the persistent banner into trading layout shells: Wrap market watchlists, order entry screens, and portfolio dashboards so the banner is visible across all demo trading views.
13. Implement subtle visual watermarking: Render a lightweight "DEMO / PAPER" background watermark on trading charts and order confirmation tickets when `state.isDemoMode` is true.
14. Implement local cooldown timer ticker: Create an in-memory 1-second interval timer that decrements the remaining cooldown seconds without polling the server repeatedly.
15. Write automated unit and widget tests: Verify state transitions, secure storage persistence, Dio routing interception, banner visibility, faucet claim logic, cooldown button disabling, and confirmation flows.

## Interfaces / Contracts

### Domain Enums and State Models
```dart
// lib/features/trading_mode/models/trading_mode_models.dart
import 'package:decimal/decimal.dart';

enum TradingMode {
  real,
  demo,
}

enum FaucetStatus {
  ready,
  cooldown,
  claiming,
  failed,
}

class DemoWalletBalance {
  final Decimal usdtBalance;
  final Decimal totalEquityValue;
  final Decimal unrealizedPnl;
  final Decimal realizedPnl;
  final DateTime lastUpdated;

  const DemoWalletBalance({
    required this.usdtBalance,
    required this.totalEquityValue,
    required this.unrealizedPnl,
    required this.realizedPnl,
    required this.lastUpdated,
  });

  factory DemoWalletBalance.initial() => DemoWalletBalance(
        usdtBalance: Decimal.fromInt(10000),
        totalEquityValue: Decimal.fromInt(10000),
        unrealizedPnl: Decimal.zero,
        realizedPnl: Decimal.zero,
        lastUpdated: DateTime.now(),
      );

  factory DemoWalletBalance.fromJson(Map<String, dynamic> json) {
    return DemoWalletBalance(
      usdtBalance: Decimal.parse(json['usdt_balance'] as String),
      totalEquityValue: Decimal.parse(json['total_equity_value'] as String),
      unrealizedPnl: Decimal.parse(json['unrealized_pnl'] as String),
      realizedPnl: Decimal.parse(json['realized_pnl'] as String),
      lastUpdated: DateTime.parse(json['last_updated'] as String),
    );
  }
}

class TradingModeState {
  final TradingMode mode;
  final DemoWalletBalance demoBalance;
  final FaucetStatus faucetStatus;
  final DateTime? cooldownExpiresAt;
  final String? lastTransactionHash;
  final String? errorMessage;
  final bool isLoading;

  const TradingModeState({
    required this.mode,
    required this.demoBalance,
    required this.faucetStatus,
    this.cooldownExpiresAt,
    this.lastTransactionHash,
    this.errorMessage,
    required this.isLoading,
  });

  bool get isDemoMode => mode == TradingMode.demo;
  bool get isRealMode => mode == TradingMode.real;

  bool get isFaucetOnCooldown {
    if (cooldownExpiresAt == null) return false;
    return DateTime.now().isBefore(cooldownExpiresAt!);
  }

  Duration get remainingCooldown {
    if (cooldownExpiresAt == null) return Duration.zero;
    final remaining = cooldownExpiresAt!.difference(DateTime.now());
    return remaining.isNegative ? Duration.zero : remaining;
  }

  TradingModeState copyWith({
    TradingMode? mode,
    DemoWalletBalance? demoBalance,
    FaucetStatus? faucetStatus,
    DateTime? cooldownExpiresAt,
    String? lastTransactionHash,
    String? errorMessage,
    bool? isLoading,
  }) {
    return TradingModeState(
      mode: mode ?? this.mode,
      demoBalance: demoBalance ?? this.demoBalance,
      faucetStatus: faucetStatus ?? this.faucetStatus,
      cooldownExpiresAt: cooldownExpiresAt ?? this.cooldownExpiresAt,
      lastTransactionHash: lastTransactionHash ?? this.lastTransactionHash,
      errorMessage: errorMessage,
      isLoading: isLoading ?? this.isLoading,
    );
  }
}

class FaucetClaimResponse {
  final bool success;
  final String transactionHash;
  final Decimal creditedAmount;
  final DateTime nextEligibleAt;

  const FaucetClaimResponse({
    required this.success,
    required this.transactionHash,
    required this.creditedAmount,
    required this.nextEligibleAt,
  });

  factory FaucetClaimResponse.fromJson(Map<String, dynamic> json) {
    return FaucetClaimResponse(
      success: json['success'] as bool,
      transactionHash: json['transaction_hash'] as String,
      creditedAmount: Decimal.parse(json['credited_amount'] as String),
      nextEligibleAt: DateTime.parse(json['next_eligible_at'] as String),
    );
  }
}

class DemoPortfolioResetResponse {
  final bool success;
  final Decimal resetBalance;
  final int cancelledOrdersCount;
  final int liquidatedPositionsCount;
  final DateTime timestamp;

  const DemoPortfolioResetResponse({
    required this.success,
    required this.resetBalance,
    required this.cancelledOrdersCount,
    required this.liquidatedPositionsCount,
    required this.timestamp,
  });

  factory DemoPortfolioResetResponse.fromJson(Map<String, dynamic> json) {
    return DemoPortfolioResetResponse(
      success: json['success'] as bool,
      resetBalance: Decimal.parse(json['reset_balance'] as String),
      cancelledOrdersCount: json['cancelled_orders_count'] as int,
      liquidatedPositionsCount: json['liquidated_positions_count'] as int,
      timestamp: DateTime.parse(json['timestamp'] as String),
    );
  }
}
```

### Storage and API Client Contracts
```dart
// lib/features/trading_mode/data/trading_mode_storage.dart
abstract class ITradingModeStorage {
  Future<TradingMode> getSavedTradingMode();
  Future<void> saveTradingMode(TradingMode mode);
  Future<String?> getDemoAuthToken();
  Future<void> saveDemoAuthToken(String token);
  Future<void> clearDemoAuthToken();
  Future<DateTime?> getFaucetCooldownExpiry();
  Future<void> saveFaucetCooldownExpiry(DateTime expiry);
}

// lib/features/trading_mode/data/demo_wallet_api_client.dart
abstract class IDemoWalletApiClient {
  Future<DemoWalletBalance> fetchDemoBalance();
  Future<FaucetClaimResponse> claimFaucet();
  Future<DateTime?> fetchFaucetCooldownStatus();
  Future<DemoPortfolioResetResponse> resetDemoPortfolio();
}
```

### Riverpod State Provider Contracts
```dart
// lib/features/trading_mode/controllers/trading_mode_controller.dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/trading_mode_models.dart';

abstract class TradingModeController extends StateNotifier<AsyncValue<TradingModeState>> {
  TradingModeController(super.state);

  Future<void> setTradingMode(TradingMode mode);
  Future<void> toggleTradingMode();
  Future<void> claimFaucet();
  Future<void> resetDemoPortfolio();
  Future<void> refreshDemoBalance();
  void tickCooldown();
}

// Provider definitions
final tradingModeStorageProvider = Provider<ITradingModeStorage>((ref) {
  throw UnimplementedError();
});

final demoWalletApiClientProvider = Provider<IDemoWalletApiClient>((ref) {
  throw UnimplementedError();
});

final tradingModeProvider = StateNotifierProvider<TradingModeController, AsyncValue<TradingModeState>>((ref) {
  throw UnimplementedError();
});
```

### Network Routing Interceptor Contract
```dart
// lib/features/trading_mode/data/trading_mode_interceptor.dart
import 'package:dio/dio.dart';
import '../models/trading_mode_models.dart';
import 'trading_mode_storage.dart';

class TradingModeNetworkInterceptor extends Interceptor {
  final ITradingModeStorage storage;
  final String mainnetBaseUrl;
  final String testnetDemoBaseUrl;

  TradingModeNetworkInterceptor({
    required this.storage,
    required this.mainnetBaseUrl,
    required this.testnetDemoBaseUrl,
  });

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final currentMode = await storage.getSavedTradingMode();

    if (currentMode == TradingMode.demo) {
      options.headers['X-Trading-Environment'] = 'DEMO_TESTNET';
      final demoToken = await storage.getDemoAuthToken();
      if (demoToken != null && demoToken.isNotEmpty) {
        options.headers['Authorization'] = 'Bearer $demoToken';
      }
      if (options.path.startsWith('/api/v1/orders') ||
          options.path.startsWith('/api/v1/wallet') ||
          options.path.startsWith('/api/v1/portfolio')) {
        options.baseUrl = testnetDemoBaseUrl;
      }
    } else {
      options.headers['X-Trading-Environment'] = 'REAL_MAINNET';
      options.baseUrl = mainnetBaseUrl;
    }

    return handler.next(options);
  }
}
```

## Security & Compliance Notes
- High-Visibility Persistent Banner: The Demo mode banner must remain anchored to the top of all relevant screens with a prominent yellow/amber background (`#FFB300` / `#FFF8E1`), warning icon, and clear typography so users can never confuse paper trading with real money.
- Clear Segregation of API Authorization Tokens: Demo mode must never transmit production JWTs or use production signing keys. Demo authentication tokens must be stored under separate secure storage keys (`key_demo_jwt` vs `key_prod_jwt`) and used exclusively with testnet endpoints.
- Confirmation Safeguard on Real Mode Switch: Transitioning from Demo to Real mode must prompt an unavoidable modal sheet requiring deliberate user confirmation and displaying statutory risk disclosures concerning financial risk.
- UI Watermarking: All order entry tickets, charts, confirmation dialogs, and portfolio summaries in Demo mode must carry explicit "PAPER TRADING - VIRTUAL ASSETS" watermarks to prevent misunderstanding or misrepresentation.
- Rate Limiting and Anti-Abuse: The 10,000 USDT testnet faucet enforces a strict 24-hour cooldown per account and device fingerprint to prevent resource exhaustion on the Besu Testnet.
- Zero PII on Besu Testnet: All testnet smart contracts and faucet claims use pseudonymous addresses with zero Personally Identifiable Information (PII) recorded on-chain.
- Safe Default Mode: Upon fresh app installation or session invalidation, the app must default safely according to platform configuration, preventing unauthorized or unexpected live order placement.

## Acceptance Criteria
- [ ] Users can toggle seamlessly between "Real-Money Trading" and "Demo Paper Trading" without logging out or reinstalling the app.
- [ ] In Demo mode, a high-contrast amber/yellow banner is persistently visible at the top of market watchlists, order entry forms, positions, and portfolio views.
- [ ] In Demo mode, order placement sheets, position tables, and trade tickets display prominent "DEMO / PAPER TRADING" badges and watermarks.
- [ ] Switching from Demo mode to Real mode triggers a modal confirmation dialog with risk disclosures that requires explicit user confirmation.
- [ ] Tapping "Claim Faucet" successfully dispenses 10,000 USDT testnet tokens and updates the virtual balance in real time.
- [ ] The faucet dialog enforces a 24-hour cooldown, displays a countdown timer, and disables the claim button until the cooldown expires.
- [ ] The "Reset Portfolio" action prompts for confirmation, cancels all open paper orders, clears simulated positions, and resets the paper balance to exactly 10,000 USDT.
- [ ] Active trading mode preference is safely persisted in encrypted secure storage across app restarts.
- [ ] API requests are dynamically routed: demo trade operations route to testnet demo endpoints, and real trade operations route to mainnet production endpoints.
- [ ] Production and demo authorization tokens are stored under separate keys and never cross-contaminated.
- [ ] Toggling modes dynamically updates the Web3 RPC provider between Besu Testnet (Chain ID 13371) and Besu Mainnet (Chain ID 13370).
- [ ] Unit and widget test suites achieve greater than 90% code coverage across state controllers, storage adapters, and UI widgets.

## Suggested Order / Dependencies
- Prerequisites:
  - Prompt 501 (Flutter Project Scaffolding)
  - Prompt 502 (Architecture & State Management)
  - Prompt 521 (Local Secure Storage)
  - Prompt 525 (API Client Layer)
  - Prompt 527 (In-App Environment Switcher & Sandbox Mode)
- Parallel Tasks:
  - Prompt 274 (Demo Wallet Service & Testnet Faucet Backend)
  - Prompt 509 (Order Placement Flow & Order Ticket UI)
  - Prompt 510 (Portfolio & Fractional Holdings Screen)
- Downstream Blockers:
  - Prompt 902 (End-to-End Trading Flow Test Suite)
  - Prompt 906 (Regulatory Sandbox & Paper Trading Scenario Verification)
