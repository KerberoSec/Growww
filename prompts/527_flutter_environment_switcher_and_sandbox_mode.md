# 527 - Flutter In-App Environment Switcher, Mock KYC Toggle, Sandbox Mode Banner & Network Inspector

## Purpose
During mobile and desktop application development, QA validation, UAT testing, and regulatory sandbox evaluations (SEBI Innovation Sandbox / IFSCA Regulatory Sandbox), engineers, QA analysts, and regulatory auditors must dynamically inspect network traffic, toggle between backend environments (Local Docker, Dev, UAT/Sandbox, Staging, Production), simulate complex investor KYC states, and trigger mock portfolio funding without rebuilding or reinstalling application binaries.

In a regulated financial client, developer tools and environment switches must be strictly guarded. Under no circumstances should test credentials, mock KYC bypasses, or insecure proxy configurations leak into production release builds. This prompt defines the architecture and implementation for an enterprise-grade In-App Environment Switcher, persistent Sandbox Mode Banner, Mock KYC State Switcher, Demat/Cash Faucet Trigger, and real-time In-App Network Telemetry Inspector for the Growww Flutter client across Android, iOS, Windows, Linux, and macOS.

## What You Are Building
A modular, compile-time-guarded developer and QA debugging suite located in `lib/core/debug/` and `lib/core/environment/`:
- `EnvironmentConfig`: Immutable environment descriptor defining API Gateway URLs, WebSocket streaming endpoints, Hyperledger Besu RPC relay URLs, and environment identifiers (`LOCAL`, `DEV`, `UAT_SANDBOX`, `STAGING`, `PROD`).
- `EnvironmentController`: Riverpod-managed state notifier responsible for persisting active environment selections in secure storage, dynamically invalidating the Dio HTTP client and WebSocket connection pools, and orchestrating zero-downtime hot-reconnection.
- `SecretGestureDetector`: Multi-tap (7 consecutive taps on the version number in Settings) and shake gesture detector with biometric authentication challenge to prevent accidental activation.
- `SandboxModeBanner`: Persistent, non-intrusive, high-contrast overlay banner and floating badge rendered when running in non-production environments (`UAT_SANDBOX`, `DEV`, `LOCAL`), displaying the active environment, active user tier, and a one-tap trigger to open the Developer Drawer.
- `MockKycSwitcher`: Dynamic KYC persona switcher enabling instant toggling between compliance states (`KYC_VERIFIED_TIER1_RETAIL`, `KYC_VERIFIED_TIER2_HNI`, `KYC_PENDING_AADHAAR`, `KYC_REJECTED_PEP`, `GIFT_CITY_NRI_VERIFIED`) by communicating with the backend Sandbox KYC mock provider.
- `DematFaucetActionSheet`: In-app faucet action sheet invoking backend sandbox endpoints to instantly seed mock INR cash balance (e.g. ₹10,00,000) and mock fractional blue-chip demat shares (e.g. 10.0 shares of Reliance, 5.0 shares of TCS) for instantaneous trading testing.
- `InAppNetworkInspector`: Integrated network telemetry interceptor (powered by `TalkerFlutter` / custom Dio inspector) capturing request/response headers, encrypted payload previews, latency metrics, HTTP status codes, gRPC frames, and cURL export capabilities.
- `CompileTimeGuards`: Tree-shaking flags using `kReleaseMode` and `const bool.fromEnvironment('ENABLE_DEV_MENU')` ensuring all debug menus, inspectors, and mock toggles are completely excised from production builds.

## Scope Boundaries
- **In Scope:**
  - Multi-environment switching with persistent local storage and live API client reconfiguration.
  - Custom endpoint entry (overriding base URL and Besu RPC URL for local developer machines).
  - Sandbox mode banner overlay with visual state indication.
  - In-app mock KYC state toggling and test persona switching.
  - In-app testnet cash and fractional demat equity faucet triggers.
  - In-app HTTP/WebSocket/gRPC network inspector with search, filtering, and cURL export.
  - Secret gesture activation and compile-time security stripping.
- **Out of Scope / Handled Elsewhere:**
  - Backend API Gateway routing and rate limiting (Prompt 219).
  - Backend Mock KYC and sandbox faucet engine (Prompt 202, Prompt 203, Prompt 609).
  - Production release signing and store packaging (Prompts 516, 517, 518, 519, 520).
  - Local secure storage implementation (Prompt 521).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (v2.5+), `dio` (v5.4+), and `talker_flutter` (v4.3+).
- **Justification:** `talker_flutter` provides lightweight, memory-efficient in-app logging and HTTP traffic inspection with zero native dependency overhead. Riverpod enables reactive dependency re-instantiation when the active environment changes.
- **Dependencies:**
  - `flutter_riverpod: ^2.5.1`
  - `dio: ^5.4.3+1`
  - `talker_flutter: ^4.3.0`
  - `talker_dio_logger: ^4.3.0`
  - `flutter_secure_storage: ^9.2.2`
  - `shared_preferences: ^2.2.3`
  - `sensors_plus: ^5.0.1` (for shake gesture detection on mobile)

## Backend / Infra Touchpoints
- **Sandbox API Gateway (`https://sandbox-api.growww.in`):** Receives API traffic when in `UAT_SANDBOX` mode.
- **Mock KYC Provider (`/api/v1/sandbox/kyc/override`):** Injects mock KYC statuses into the investor session for testing onboarding gates.
- **Sandbox Faucet Service (`/api/v1/sandbox/faucet/fund`):** Seeds testnet INR wallet balances and NSDL/CDSL mock demat balances.
- **Permissioned Blockchain Node (`https://sandbox-rpc.growww.in`):** Connects to the Hyperledger Besu UAT network for on-chain proof-of-reserve and DvP settlement inspection.

## Blockchain Interaction
- **Dynamic RPC Node Switching:** When switching between environments (e.g. Local Besu node `http://127.0.0.1:8545` to Sandbox Besu node `https://sandbox-rpc.growww.in`), the client dynamically updates the read-only Web3 RPC client provider (`viem` / `web3dart`), ensuring proof-of-reserve queries and on-chain transaction receipt lookups point to the corresponding consortium network.
- **Testnet Contract Address Mapping:** Loads distinct contract registry configurations (`ComplianceRegistry`, `DvPSettlementEngine`, `ProofOfReserveRegistry`) per environment flavor.

## Step-by-Step Build Instructions
1. Scaffold directory `lib/core/debug/` with `views/`, `controllers/`, `widgets/`, `network/`, and `models/`.
2. Define the `AppEnvironment` enum (`local`, `dev`, `uatSandbox`, `staging`, `production`) and `EnvironmentConfig` data class in `lib/core/environment/environment_config.dart`.
3. Implement `EnvironmentController` as a `StateNotifier<EnvironmentConfig>` backed by `ISecureStorageService` to persist environment selections across app restarts.
4. Integrate `EnvironmentController` into `ApiClient` (Prompt 525) so that updating the environment immediately reconfigures Dio `BaseOptions.baseUrl` and flushes existing connection sockets.
5. Configure `Talker` logger and `TalkerDioLogger` interceptor in `lib/core/debug/network/network_inspector.dart` to capture HTTP requests, responses, status codes, query parameters, and latency.
6. Build `SandboxModeBanner`:
   - Display a persistent, non-intrusive 24px top status banner when `config.isProduction == false`.
   - Render environment name (e.g. "SANDBOX - MOCK TRADING"), network indicator dot, and quick-access tap target.
7. Implement `SecretGestureDetector`:
   - Wrap the version tile in `SettingsScreen` (Prompt 514) with a 7-tap counter resetting after a 2-second timeout.
   - On mobile devices, listen to accelerometer shake events via `sensors_plus` when `kDebugMode` or `ENABLE_DEV_MENU` is true.
   - Challenge the user with local MPIN or biometric authentication before displaying the Developer Menu sheet.
8. Build `DeveloperDrawer` / `DeveloperMenuModal`:
   - Environment Selector dropdown (`Local`, `Dev`, `UAT Sandbox`, `Staging`, `Production`).
   - Custom Base URL and Custom Besu RPC URL input fields for local microservice debugging.
   - Mock KYC Persona Selector with instant state apply button.
   - Demat Portfolio & INR Wallet Faucet trigger button.
   - Launch Network Inspector button (`TalkerScreen`).
   - Clear All App Cache & Local Storage button.
9. Implement `MockKycController`:
   - Send `POST /api/v1/sandbox/kyc/override` with selected KYC status.
   - Refresh `userProfileProvider` and invalidate `kycStatusProvider` to test UI state transitions without app restart.
10. Implement `FaucetController`:
    - Send `POST /api/v1/sandbox/faucet/fund` with payload specifying desired test INR and stock quantities.
    - Trigger haptic feedback and invalidate `walletBalanceProvider` and `portfolioHoldingsProvider`.
11. Enforce compile-time tree shaking:
    - Wrap developer menu routes, widgets, and gestures with `if (kReleaseMode && !const bool.fromEnvironment('ENABLE_DEV_MENU'))`.
    - Ensure zero developer utilities or test endpoints are compiled into public Google Play Store / Apple App Store production binaries.
12. Write unit tests for `EnvironmentController` verifying storage persistence, URL resolution, and Dio instance mutation.
13. Write widget tests for `SandboxModeBanner` verifying visibility in non-production flavors and absence in production.
14. Test network inspection on Android, iOS, Windows, Linux, and macOS platforms to verify cURL export and UI responsiveness.

## Interfaces / Contracts

### Environment Configuration Model
```dart
// lib/core/environment/environment_config.dart
enum AppEnvironment {
  local,
  dev,
  uatSandbox,
  staging,
  production,
}

class EnvironmentConfig {
  final AppEnvironment environment;
  final String apiBaseUrl;
  final String wsMarketDataBaseUrl;
  final String besuRpcUrl;
  final String complianceRegistryAddress;
  final String dvpEngineAddress;
  final bool isProduction;
  final bool enableNetworkInspector;

  const EnvironmentConfig({
    required this.environment,
    required this.apiBaseUrl,
    required this.wsMarketDataBaseUrl,
    required this.besuRpcUrl,
    required this.complianceRegistryAddress,
    required this.dvpEngineAddress,
    required this.isProduction,
    required this.enableNetworkInspector,
  });

  factory EnvironmentConfig.fromEnvironment(AppEnvironment env, {String? customApiUrl, String? customRpcUrl}) {
    switch (env) {
      case AppEnvironment.local:
        return EnvironmentConfig(
          environment: AppEnvironment.local,
          apiBaseUrl: customApiUrl ?? 'http://10.0.2.2:8000/api/v1',
          wsMarketDataBaseUrl: 'ws://10.0.2.2:8001/ws',
          besuRpcUrl: customRpcUrl ?? 'http://10.0.2.2:8545',
          complianceRegistryAddress: '0x5FbDB2315678afecb367f032d93F642f64180aa3',
          dvpEngineAddress: '0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512',
          isProduction: false,
          enableNetworkInspector: true,
        );
      case AppEnvironment.dev:
        return const EnvironmentConfig(
          environment: AppEnvironment.dev,
          apiBaseUrl: 'https://dev-api.growww.in/api/v1',
          wsMarketDataBaseUrl: 'wss://dev-ws.growww.in/ws',
          besuRpcUrl: 'https://dev-rpc.growww.in',
          complianceRegistryAddress: '0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0',
          dvpEngineAddress: '0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9',
          isProduction: false,
          enableNetworkInspector: true,
        );
      case AppEnvironment.uatSandbox:
        return const EnvironmentConfig(
          environment: AppEnvironment.uatSandbox,
          apiBaseUrl: 'https://sandbox-api.growww.in/api/v1',
          wsMarketDataBaseUrl: 'wss://sandbox-ws.growww.in/ws',
          besuRpcUrl: 'https://sandbox-rpc.growww.in',
          complianceRegistryAddress: '0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9',
          dvpEngineAddress: '0x5FC8d32690cc91D4c39d9d3abcBD16989F875707',
          isProduction: false,
          enableNetworkInspector: true,
        );
      case AppEnvironment.staging:
        return const EnvironmentConfig(
          environment: AppEnvironment.staging,
          apiBaseUrl: 'https://staging-api.growww.in/api/v1',
          wsMarketDataBaseUrl: 'wss://staging-ws.growww.in/ws',
          besuRpcUrl: 'https://staging-rpc.growww.in',
          complianceRegistryAddress: '0x0165878A594ca255338adfa4d48449f69242Eb8F',
          dvpEngineAddress: '0xa513E6E4b8f2a923D98304ec87F64353C4D5C853',
          isProduction: false,
          enableNetworkInspector: true,
        );
      case AppEnvironment.production:
        return const EnvironmentConfig(
          environment: AppEnvironment.production,
          apiBaseUrl: 'https://api.growww.in/api/v1',
          wsMarketDataBaseUrl: 'wss://ws.growww.in/ws',
          besuRpcUrl: 'https://rpc.growww.in',
          complianceRegistryAddress: '0x2279B7A0a67E14288777993e08610cB73f1b91EC',
          dvpEngineAddress: '0x8A791620dd6260079BF849Dc5567aDC3F2FdC318',
          isProduction: true,
          enableNetworkInspector: false,
        );
    }
  }
}
```

### Mock KYC and Faucet API Contracts
```dart
// lib/core/debug/models/sandbox_controls_dto.dart
enum MockKycStatus {
  tier1RetailVerified,
  tier2HniVerified,
  giftCityNriVerified,
  pendingAadhaarOtp,
  rejectedPepSanctioned,
}

class MockKycOverrideRequest {
  final MockKycStatus status;
  final String? customPan;
  final String? jurisdiction;

  MockKycOverrideRequest({
    required this.status,
    this.customPan,
    this.jurisdiction,
  });

  Map<String, dynamic> toJson() => {
    'status': status.name,
    'custom_pan': customPan,
    'jurisdiction': jurisdiction ?? 'DOMESTIC_RESIDENT',
  };
}

class SandboxFaucetRequest {
  final double inrAmount;
  final Map<String, double> equityHoldings; // ISIN -> Fractional Quantity

  SandboxFaucetRequest({
    required this.inrAmount,
    required this.equityHoldings,
  });

  Map<String, dynamic> toJson() => {
    'inr_amount': inrAmount,
    'equity_holdings': equityHoldings,
  };
}
```

## Security & Compliance Notes
- **Strict Build Stripping:** In release builds distributed to public app stores (`kReleaseMode == true` without explicit compile-time flag `ENABLE_DEV_MENU`), all developer drawers, network inspector routes, shake gesture handlers, and mock KYC switchers MUST be tree-shaken and completely inaccessible.
- **Biometric Challenge on Debug Menu:** Opening the developer menu requires the active device MPIN or biometric challenge to prevent unauthorized testing overrides on physical QA devices.
- **Zero Real Funds in Sandbox:** The Sandbox Mode Banner must remain permanently visible in non-production builds so users, auditors, and QA analysts never confuse testnet assets with real capital.
- **Network Inspector PII Redaction:** The In-App Network Inspector must automatically mask authorization tokens (`Bearer eyJ...`), Aadhaar numbers, PAN numbers, and bank account numbers in its visual request/response inspector.

## Acceptance Criteria
- [ ] 7 consecutive taps on the version number opens the Developer Menu when running in debug or QA builds.
- [ ] Switching between Local, Dev, Sandbox, Staging, and Production instantly updates the API client base URL and invalidates WebSocket connections.
- [ ] Sandbox Mode Banner renders visibly at the top of the app in all non-production environments and displays active environment details.
- [ ] Toggling Mock KYC state immediately updates investor verification status without requiring manual Aadhaar/PAN upload.
- [ ] Demat & Cash Faucet trigger successfully seeds sandbox INR balances and fractional shares.
- [ ] In-App Network Inspector records all HTTP requests, responses, latency, and status codes, and allows copying request details as cURL.
- [ ] Release production builds completely strip developer menus and network inspection utilities with zero PII leakage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Scaffolding), Prompt 502 (Architecture & State Management), Prompt 521 (Local Secure Storage), Prompt 525 (API Client Layer).
- **Parallel Tasks:** Prompt 514 (Settings & Profile UI), Prompt 609 (Developer Portal & Testnet Faucet UI).
- **Downstream Blockers:** Prompt 902 (End-to-End Testing), Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios).
