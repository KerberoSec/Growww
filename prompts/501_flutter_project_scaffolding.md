# 501 - Flutter Multi-Platform Project Scaffolding (Mobile + Desktop)

## Purpose
Establishes the enterprise-grade, multi-platform Flutter client codebase for Growww, supporting Android, iOS, Windows, Linux, and macOS from a single unified repository. In alignment with the Growww North Star vision, this codebase provides a high-performance, native thick-client experience enabling Indian and international investors to seamlessly onboard, inspect real-time market data, execute fractional equity trades, and verify on-chain Proof-of-Reserve attestations without web-view latency or platform fragmentation.

## What You Are Building
A production-ready Flutter 3.22+ workspace skeleton located at `apps/growww_flutter/` with:
- Configured platform runners for Android, iOS, Windows, Linux, and macOS.
- Unified Flutter Version Management (`.fvmrc`) locking the engine version.
- Multi-flavor environment configuration (`dev`, `staging`, `prod`) using Dart environment defines.
- Platform-native build scripts, CMake/C++ desktop runners, Xcode workspace configurations, and Android Gradle Kotlin DSL configurations.
- Core asset pipeline setup (SVGs, vector icons, custom fonts, blockchain ABI JSON definitions).
- Automated CI compile verification targets for all 5 operating systems.

## Scope Boundaries
- **In Scope:**
 - Workspace scaffolding, multi-platform runners, FVM configuration, dependency specifications in `pubspec.yaml`, flavor configurations, root `main_<flavor>.dart` entrypoints, and CI validation scripts.
- **Out of Scope / Handled Elsewhere:**
 - Layered state management architecture (handled in Prompt 502).
 - Design system tokens and atomic widgets (handled in Prompt 503).
 - Concrete feature screens (handled in Prompts 504-515).
 - Store deployment and platform code-signing pipelines (handled in Prompts 516-520).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+.
 - *Justification:* Flutter compiles to native ARM64 and x86_64 machine code across Android, iOS, Windows, macOS, and Linux. The Impeller rendering engine ensures predictable 60/120 FPS UI performance free from runtime shader compilation jank, critical for high-frequency price tick charts and low-latency order execution.
- **Environment & Build Management:** Flutter Version Management (`fvm`), `flutter_flavorizr` / custom compile-time defines (`--dart-define-from-file`), Melos for monorepo workspace orchestration.
- **Native Platform Runners:**
 - Android: Gradle Kotlin DSL, compileSdk 34, minSdk 26 (Android 8.0 Oreo).
 - iOS: Xcode 15+, CocoaPods/Swift Package Manager, iOS Deployment Target 15.0+.
 - macOS: App Sandbox enabled, Hardened Runtime, macOS 12.0+.
 - Windows: Visual Studio C++ 2022 toolchain, CMake 3.20+, Windows 10/11 x64/ARM64.
 - Linux: CMake, Ninja, GTK 3.0+, Clang/GCC on Ubuntu 22.04 LTS+.

## Backend / Infra Touchpoints
- **Environment Endpoints:** Configuration targets for REST API Gateway (Prompt 219), Market Data WebSocket (Prompt 207), and Blockchain JSON-RPC endpoints.
- **Local Dev Mock Server:** Local Docker Compose endpoint (`http://localhost:8000`) for offline client testing.
- **Asset Bundles:** Local assets directory containing SEBI mandatory statutory risk disclosures, token ABI schemas, and regulatory iconography.

## Blockchain Interaction
- **Client Web3 Connectivity:** Prepares the asset pipeline and dependency layer to include client-side Ethereum/Besu ABI contracts (`DigitalSecurityToken.json`, `SettlementDvP.json`, `ProofOfReserveRegistry.json`).
- **Ledger Transparency Hooks:** Prepares client-side configuration for querying the permissioned Hyperledger Besu consortium network (RPC nodes) directly for public Proof-of-Reserve root hashes and transaction settlement verification badges.

## Step-by-Step Build Instructions
1. Initialize the Flutter application directory under `apps/growww_flutter/` targeting all 5 platforms (`flutter create --org in.growww.client --platforms=android,ios,windows,linux,macos growww_flutter`).
2. Configure `.fvmrc` pinning Flutter SDK version to `3.22.x` stable channel and verify with `fvm use 3.22.x`.
3. Configure `pubspec.yaml` with core production dependencies (Riverpod, Dio, Freezed, Web3dart, Local Auth, Flutter SVG, Google Fonts).
4. Organize the folder hierarchy into `lib/app/`, `lib/core/`, `lib/features/`, `lib/shared/`, `assets/`, and `config/`.
5. Create environment configuration files in `config/` (`env.dev.json`, `env.staging.json`, `env.prod.json`) containing API URLs, Besu RPC endpoints, and feature flags.
6. Create flavor entry points: `lib/main_dev.dart`, `lib/main_staging.dart`, and `lib/main_prod.dart` reading configurations via `String.fromEnvironment`.
7. Configure Android runner in `android/app/build.gradle.kts` setting `compileSdk = 34`, `minSdk = 26`, and multidex support.
8. Configure iOS runner in `ios/Runner.xcodeproj` with custom build configurations matching the 3 flavors and camera/biometric usage descriptions in `Info.plist`.
9. Configure macOS runner in `macos/Runner/` with entitlement configs allowing outgoing network connections and sandbox storage access.
10. Configure Windows runner in `windows/runner/` with modern DPI awareness, application title, and window dimension bounds (minimum 1024x768).
11. Configure Linux runner in `linux/my_application.cc` setting application ID and GTK window dimensions.
12. Establish asset directory mapping in `pubspec.yaml` for images, icons, fonts (Inter, JetBrains Mono), and smart contract ABI definitions in `assets/contracts/`.
13. Create a verification script `scripts/verify_builds.sh` that sequentially runs `flutter analyze` and dry-run compiles across all 5 target platforms.

## Interfaces / Contracts
```dart
// lib/core/config/app_config.dart
enum AppFlavor { dev, staging, prod }

class AppConfig {
  final AppFlavor flavor;
  final String apiBaseUrl;
  final String wsBaseUrl;
  final String blockchainRpcUrl;
  final String proofOfReserveContractAddress;
  final bool enableProofOfReserveBadge;
  final Duration networkTimeout;

  const AppConfig({
    required this.flavor,
    required this.apiBaseUrl,
    required this.wsBaseUrl,
    required this.blockchainRpcUrl,
    required this.proofOfReserveContractAddress,
    required this.enableProofOfReserveBadge,
    this.networkTimeout = const Duration(seconds: 15),
  });

  static AppConfig fromEnvironment() {
    const flavorStr = String.fromEnvironment('GROWWW_FLAVOR', defaultValue: 'dev');
    final flavor = AppFlavor.values.firstWhere(
      (e) => e.name == flavorStr,
      orElse: () => AppFlavor.dev,
    );

    return AppConfig(
      flavor: flavor,
      apiBaseUrl: const String.fromEnvironment('API_BASE_URL', defaultValue: 'https://api.dev.growww.in'),
      wsBaseUrl: const String.fromEnvironment('WS_BASE_URL', defaultValue: 'wss://ws.dev.growww.in'),
      blockchainRpcUrl: const String.fromEnvironment('BESU_RPC_URL', defaultValue: 'https://rpc.besu.dev.growww.in'),
      proofOfReserveContractAddress: const String.fromEnvironment('POR_CONTRACT', defaultValue: '0x0000000000000000000000000000000000000000'),
      enableProofOfReserveBadge: const bool.fromEnvironment('ENABLE_POR_BADGE', defaultValue: true),
    );
  }
}
```

## Security & Compliance Notes
- **Zero Hardcoded Secrets:** No API keys, JWT secrets, or private keys shall exist in source code or committed config JSONs. All secrets are injected at build/runtime.
- **Obfuscation Ready:** Release build commands must enforce Dart code obfuscation (`--obfuscate --split-debug-info=symbols/`).
- **SEBI Data Locality:** Network configurations must strictly route through domestic Indian infrastructure endpoints for domestic flavor builds.
- **Platform Sandbox:** Desktop builds (macOS/Windows/Linux) must operate within platform sandbox boundaries with minimal required system privileges.

## Acceptance Criteria
- [ ] Project scaffolds cleanly and runs `flutter pub get` without dependency conflicts.
- [ ] `fvm flutter analyze` returns 0 issues with strict lint rules enabled (`very_good_analysis` or equivalent).
- [ ] Application compiles and launches successfully on Android emulator / physical device.
- [ ] Application compiles and launches successfully on iOS simulator / physical device.
- [ ] Application compiles and launches natively on Windows (x64), Linux (GTK), and macOS.
- [ ] Flavor configurations correctly switch API and RPC URLs without code edits.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 106 (Monorepo & Repo Layout), Prompt 107 (Coding Standards & Linting).
- **Parallel Tasks:** Can proceed alongside Prompt 502 (App Architecture) and Prompt 503 (Design System).
- **Enables:** Prompts 504-526.
