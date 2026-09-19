# 524 - Client-Side Crash Reporting, Performance Monitoring & Privacy-Preserving Analytics

## Purpose
In a mission-critical financial trading terminal handling real INR balances and atomic blockchain settlements, client stability and latency visibility are paramount. A single uncaught exception during order submission or WebSocket market data stream parsing can cause financial losses or regulatory audit failures. However, reporting errors from client devices in a fintech environment subject to India's DPDP Act 2023 and SEBI cybersecurity frameworks requires strict data sanitation to prevent PII, Demat numbers, or bank account details from leaking to third-party telemetry systems.

This prompt defines the client-side crash reporting, performance tracing, and privacy-preserving analytics architecture for the Growww Flutter client. It integrates **Sentry** (and optional Firebase Crashlytics on mobile) with automated PII scrubbers, breadcrumb tracking for UI navigation and network requests, custom performance traces for order-to-fill latency, desktop crash handlers, and an explicit investor privacy opt-out mechanism.

## What You Are Building
A hardened telemetry framework located in `lib/core/telemetry/`:
- `TelemetryManager`: Central initialization hub configuring Sentry Flutter SDK, native error handlers, and global unhandled exception captures.
- `PiiDataScrubber`: High-speed regex and JSON tree sanitizer that automatically redacts PAN numbers, Aadhaar numbers, email addresses, phone numbers, auth tokens, and order payload values from error logs, breadcrumbs, and stack traces before network transmission.
- `AppBreadcrumbTracker`: Automatic lifecycle, navigation (`GoRouterObserver`), and network interceptor logger recording anonymized user paths leading up to a crash.
- `TradingPerformanceTracer`: Custom OpenTelemetry-compatible span tracker measuring end-to-end user operations: Cold Start Time, Order Submission-to-Ack Latency, WebSocket Reconnect Duration, and Impeller/Skia Frame Jank Rate.
- `TelemetryConsentController`: User privacy control allowing investors to toggle anonymous crash and usage telemetry in compliance with India's DPDP Act 2023.

## Scope Boundaries
- **In Scope:**
 - Multi-platform error capture across Android, iOS, Windows, Linux, and macOS.
 - Integration with `SentryFlutter` and `FlutterError.onError`.
 - PII and financial data redaction algorithms.
 - Performance spans and metrics for trading operations.
 - Native symbol and mapping file configuration (Prompts 516, 517).
 - User consent toggle state management.
- **Out of Scope / Handled Elsewhere:**
 - Backend server observability (Prometheus/Grafana/Loki - Prompt 806).
 - Centralized audit log service (Prompt 218).
 - UI Settings screen integration (Prompt 514).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `sentry_flutter` (v8.0+).
- **Justification:** `sentry_flutter` provides out-of-the-box cross-platform error and performance monitoring for all 5 target platforms (Android, iOS, macOS, Windows, Linux) including native crash handling (C++/NDK/Mach-O), automatic breadcrumbs, and granular before-send sanitation callbacks.
- **Dependencies:**
 - `sentry_flutter: ^8.3.0`
 - `flutter_riverpod: ^2.5.1`
 - `go_router: ^14.0.0`
 - `dio: ^5.4.3`

## Backend / Infra Touchpoints
- **Self-Hosted Sentry / Dedicated EU/IN Sentry Instance:** Ingestion endpoint (`DSN`) hosted in a compliant geographical jurisdiction.
- **Backend OpenTelemetry Collector (Prompt 806):** Optional dispatch of client performance spans via HTTP/Protobuf for unified client-to-backend distributed tracing.

## Blockchain Interaction
- **Blockchain Transaction Latency Tracking:** Traces end-to-end latency from client order submission, matching engine execution (Prompt 205), to on-chain `SettlementDvP.sol` event confirmation, logging performance distributions without recording user private wallet keys.

## Step-by-Step Build Instructions
1. Scaffold directory `lib/core/telemetry/` with subdirectories `domain/`, `sanitizers/`, `performance/`, and `presentation/`.
2. Configure `SentryFlutter.init` in `lib/main.dart` wrapping the entire application runner (`runApp()`).
3. Set up global unhandled error handlers: `FlutterError.onError = (details) => Sentry.captureException(...)` and `PlatformDispatcher.instance.onError`.
4. Implement `PiiDataScrubber` with robust regex filters for:
 - Indian PAN: `[A-Z]{5}[0-9]{4}[A-Z]{1}` -> `[REDACTED_PAN]`
 - Aadhaar: `\b\d{4}\s?\d{4}\s?\d{4}\b` -> `[REDACTED_AADHAAR]`
 - Email: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}` -> `[REDACTED_EMAIL]`
 - JWT / Auth Bearer: `Bearer\s+[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*` -> `Bearer [REDACTED_TOKEN]`
5. Attach `PiiDataScrubber` to `options.beforeSend` and `options.beforeBreadcrumb` hooks in Sentry configuration.
6. Create `SentryNavigatorObserver` and register it with `GoRouter` in `lib/core/routing/router.dart` to automatically track view navigation breadcrumbs.
7. Create `DioTelemetryInterceptor` for the HTTP client (Prompt 525) that logs HTTP method, status code, and latency breadcrumbs, while strictly omitting request/response body payloads.
8. Implement `TradingPerformanceTracer` providing helper methods: `startTrace(String name)`, `startSpan(String operation)`, and `finish()`.
9. Instrument critical trading flows: Measure milliseconds taken from "Tap Buy" -> "API Acknowledged" -> "Matching Executed" -> "UI Confirmed".
10. Implement UI frame rate jank monitoring using `options.enableAutoPerformanceTracing = true` and `options.tracesSampleRate = 0.2` (20% sampling in production).
11. Build `TelemetryConsentController` with Riverpod reading the user's telemetry preference from `SecureStorageService` (Prompt 521); if opted-out, set `Sentry.setEnabled(false)`.
12. Write unit tests for `PiiDataScrubber` verifying that all synthetic PII, credit card numbers, Demat BOIDs, and token strings are completely redacted.
13. Write integration tests simulating an intentional asynchronous isolate error and verifying sanitized payload generation.
14. Test native crash capture on Android (JNI/C++ crash) and iOS (mach exception).

## Interfaces / Contracts

```dart
// lib/core/telemetry/sanitizers/pii_scrubber.dart
class PiiDataScrubber {
  static final RegExp _panPattern = RegExp(r'[A-Z]{5}[0-9]{4}[A-Z]{1}', caseSensitive: false);
  static final RegExp _aadhaarPattern = RegExp(r'\b\d{4}\s?\d{4}\s?\d{4}\b');
  static final RegExp _emailPattern = RegExp(r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}');
  static final RegExp _bearerPattern = RegExp(r'Bearer\s+[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*');
  static final RegExp _bankAccPattern = RegExp(r'\b\d{9,18}\b');

  static String scrubString(String input) {
    var sanitized = input;
    sanitized = sanitized.replaceAll(_panPattern, '[REDACTED_PAN]');
    sanitized = sanitized.replaceAll(_aadhaarPattern, '[REDACTED_AADHAAR]');
    sanitized = sanitized.replaceAll(_emailPattern, '[REDACTED_EMAIL]');
    sanitized = sanitized.replaceAll(_bearerPattern, 'Bearer [REDACTED_TOKEN]');
    sanitized = sanitized.replaceAll(_bankAccPattern, '[REDACTED_BANK_ACC]');
    return sanitized;
  }
}

// lib/core/telemetry/performance/trading_performance_tracer.dart
abstract class IPerformanceTracer {
  ISpan startTradeExecutionSpan({required String isin, required String side});
  void recordFrameJank({required double buildDurationMs, required double rasterDurationMs});
  Future<T> traceOperation<T>({
    required String name,
    required String operation,
    required Future<T> Function() block,
  });
}

abstract class ISpan {
  void setTag(String key, String value);
  void finish({String? status});
}
```

```dart
// lib/core/telemetry/telemetry_manager.dart
import 'package:flutter/foundation.dart';
import 'package:sentry_flutter/sentry_flutter.dart';
import 'sanitizers/pii_scrubber.dart';

class TelemetryManager {
  static Future<void> initialize({
    required String dsn,
    required String environment,
    required bool isEnabled,
  }) async {
    await SentryFlutter.init(
      (options) {
        options.dsn = isEnabled ? dsn : '';
        options.environment = environment;
        options.tracesSampleRate = environment == 'prod' ? 0.2 : 1.0;
        options.enableAutoSessionTracking = true;
        options.attachStacktrace = true;
        options.sendDefaultPii = false; // Strictly disabled

        options.beforeSend = (event, {hint}) {
          // Scrub exception messages and stack traces
          final scrubbedExceptions = event.exceptions?.map((ex) {
            return ex.copyWith(
              value: ex.value != null ? PiiDataScrubber.scrubString(ex.value!) : null,
            );
          }).toList();

          return event.copyWith(
            exceptions: scrubbedExceptions,
            message: event.message != null 
                ? SentryMessage(PiiDataScrubber.scrubString(event.message!.formatted)) 
                : null,
          );
        };

        options.beforeBreadcrumb = (breadcrumb, {hint}) {
          if (breadcrumb.message != null) {
            return breadcrumb.copyWith(
              message: PiiDataScrubber.scrubString(breadcrumb.message!),
            );
          }
          return breadcrumb;
        };
      },
    );
  }
}
```

## Security & Compliance Notes
- **DPDP Act (Digital Personal Data Protection Act 2023):** Telemetry collection must be strictly anonymized. Investors must have an accessible toggle in Settings to opt out of crash reporting without degrading app functionality.
- **Zero Financial Payload Logging:** Network breadcrumbs must log only HTTP status code, URL path (without query params containing user IDs), and elapsed milliseconds. Request payloads and response JSON are never attached.
- **Secure DSN Storage:** Sentry DSN must be injected during build time via `--dart-define` and never hardcoded in plaintext public repositories.

## Acceptance Criteria
- [ ] Unhandled Dart and Flutter UI exceptions are captured and reported across all 5 target platforms (Android, iOS, Windows, Linux, macOS).
- [ ] PII scrubber successfully redacts PANs, Aadhaar numbers, phone numbers, emails, and auth tokens from every event and breadcrumb.
- [ ] Network breadcrumbs record HTTP request metadata without exposing JSON request/response bodies.
- [ ] Custom performance spans measure order placement and settlement confirmation latencies accurately.
- [ ] Opting out in the settings menu immediately halts Sentry event transmission.
- [ ] Unit tests for data sanitization achieve 100% test branch coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 108 (Environment Strategy).
- **Parallel Tasks:** Prompt 514 (Settings UI), Prompt 525 (API Client Layer).
