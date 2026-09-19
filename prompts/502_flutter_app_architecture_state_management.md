# 502 - Flutter App Architecture & State Management (Riverpod)

## Purpose
Establishes the enterprise Clean Architecture and reactive state management foundation for the Growww Flutter client. In a mission-critical fintech application handling real-time financial market data, high-speed order execution, and cryptographic blockchain verifications, the state management layer must guarantee compile-time safety, unidirectional data flow, immutability, deterministic error handling, and testability without UI context dependencies.

## What You Are Building
A four-layer Clean Architecture framework (`apps/growww_flutter/lib/`) utilizing `flutter_riverpod` (with code generation) and functional error handling (`fpdart`), including:
- **Presentation Layer:** Platform-adaptive UI widgets and Riverpod `AsyncNotifier` view-models.
- **Application Layer:** Orchestration use cases, workflow coordinators (e.g., DvP order placement, KYC submission).
- **Domain Layer:** Pure Dart business logic, immutable entities (`Freezed`), failure unions, and repository interfaces.
- **Data Layer:** Remote REST/gRPC data sources, WebSocket price stream sources, local secure storage caches, and concrete repository implementations.
- Global reactive state providers for authentication status, active portfolio, live market tickers, and blockchain Proof-of-Reserve validation.

## Scope Boundaries
- **In Scope:**
 - Layer contracts, base controller classes, code generation setup with `riverpod_generator` and `freezed`, repository patterns, AsyncValue extension helpers, and dependency injection wiring.
- **Out of Scope / Handled Elsewhere:**
 - Scaffolding & platform runners (handled in Prompt 501).
 - Concrete screen designs and layout widgets (handled in Prompts 504-515).
 - Low-level network transport, interceptors, and SSL pinning (handled in Prompt 525).

## Technology to Use
- **Primary Framework:** `flutter_riverpod` (v2.5+) with `riverpod_annotation` and `riverpod_generator`.
 - *Justification:* Unlike traditional `Bloc` or `Provider`, Riverpod offers compile-time provider safety, seamless async state caching (`AsyncValue`), automatic lifecycle management (`autoDispose`), family scoping, and completely decouples business logic from the Flutter `BuildContext`.
- **Data Modeling & Immutability:** `freezed` and `json_serializable` for robust immutable domain models, copy-with semantics, and JSON serialization.
- **Functional Programming & Error Handling:** `fpdart` (`Either<Failure, T>`, `Option<T>`) for explicit, type-safe failure handling without uncaught runtime exceptions.
- **Code Generation Tooling:** `build_runner` with optimized build filters.

## Backend / Infra Touchpoints
- **API Client Gateway:** Consumes typed API clients generated for REST/gRPC microservices (Prompt 525).
- **WebSocket Streams:** Binds low-latency order-book and price streams from Market Data Service (Prompt 207).
- **Secure Storage Cache:** Interacts with hardware-backed secure storage (Prompt 521) for persisting encrypted session tokens.

## Blockchain Interaction
- **Reactive On-Chain State:** Exposes reactive Riverpod stream providers that subscribe to permissioned Hyperledger Besu JSON-RPC filter logs (e.g., DvP settlement confirmations from `SettlementDvP.sol`, Proof-of-Reserve updates from `ProofOfReserveRegistry.sol`).
- **Cryptographic Attestation State:** Caches and exposes Merkle root verification states to provide instantaneous UI updates for Proof-of-Reserve badges across portfolio and security detail screens.

## Step-by-Step Build Instructions
1. Add state management dependencies to `pubspec.yaml` (`flutter_riverpod`, `riverpod_annotation`, `freezed_annotation`, `fpdart`, `build_runner`, `riverpod_generator`, `freezed`, `json_serializable`).
2. Define the architectural directory structure under `lib/`: `core/`, `features/<feature_name>/presentation/`, `application/`, `domain/`, `data/`.
3. Create core domain failure union classes in `lib/core/errors/failures.dart` using Freezed (e.g., `ServerFailure`, `NetworkFailure`, `AuthFailure`, `BlockchainVerificationFailure`).
4. Implement generic base AsyncNotifier extensions and UI helper widgets (`AsyncValueWidget`) to standardise loading, error, and data states.
5. Setup the global `ProviderScope` at the root widget in `lib/app/app.dart` with custom observer logging for development debugging.
6. Implement core auth state provider (`auth_state_provider.dart`) managing unauthenticated, authenticated, and biometric-locked states.
7. Implement user session coordinator ensuring that user logout triggers a global provider cache invalidation (`ref.invalidate()`) to prevent state leaks.
8. Create data layer repository base templates handling conversion of network exceptions into domain `Either<Failure, T>` types.
9. Implement WebSocket stream provider wrappers with automatic reconnect and exponential backoff state emission.
10. Implement caching policies (TTL-based cache invalidation for market static data and portfolio snapshots).
11. Configure `build.yaml` with build targets and cache options for high-speed code generation via `dart run build_runner build`.
12. Write unit tests for sample Riverpod notifiers verifying state transitions using `ProviderContainer`.

## Interfaces / Contracts
```dart
// lib/core/errors/failures.dart
import 'package:freezed_annotation/freezed_annotation.dart';
part 'failures.freezed.dart';

@freezed
class Failure with _$Failure {
  const factory Failure.server({required String message, int? statusCode}) = ServerFailure;
  const factory Failure.network({required String message}) = NetworkFailure;
  const factory Failure.unauthorized({required String message}) = UnauthorizedFailure;
  const factory Failure.blockchainAttestationFailed({required String reason}) = BlockchainAttestationFailure;
  const factory Failure.validation({required String message, String? field}) = ValidationFailure;
}

// lib/core/architecture/base_controller.dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fpdart/fpdart.dart';
import '../errors/failures.dart';

abstract class BaseAsyncController<T> extends AutoDisposeAsyncNotifier<T> {
  Future<void> executeTask(Future<Either<Failure, T>> Function() task) async {
    state = const AsyncValue.loading();
    final result = await task();
    state = result.fold(
      (failure) => AsyncValue.error(failure, StackTrace.current),
      (data) => AsyncValue.data(data),
    );
  }
}
```

## Security & Compliance Notes
- **State Sanitization on Session Expiry:** All user PII, bank account details, portfolio holdings, and active orders stored in Riverpod providers must be wiped from memory immediately upon logout or token expiry.
- **No Sensitive State Logging in Production:** Riverpod state observer must disable state mutation logging when `AppConfig.flavor == AppFlavor.prod`.
- **Deterministic Audit Trail:** State mutation events for critical actions (order placement, fund withdrawal) must be correlated with trace IDs for forensic auditability.

## Acceptance Criteria
- [ ] Clean four-layer architecture is strictly established and documented in `apps/growww_flutter/docs/architecture.md`.
- [ ] Code generation with `riverpod_generator` and `freezed` executes without errors.
- [ ] Base controller correctly maps `fpdart` `Either<Failure, T>` into Riverpod `AsyncValue<T>`.
- [ ] Session invalidation properly resets all scoped and global provider states.
- [ ] Unit tests for state notifiers pass with >90% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Project Scaffolding).
- **Parallel Tasks:** Prompt 503 (Design System & Theming).
- **Enables:** Prompts 504-515 (All feature UI screens) and Prompt 525 (API Client Layer).
