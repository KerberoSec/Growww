# 526 - Deep Linking & Universal Links Across Mobile and Desktop Platforms

## Purpose
Seamless navigation is vital for retail trading conversions, bank payment callbacks (UPI/NetBanking), push notification interactions, and sharing stock research over messaging channels. In a regulated financial application, deep links must not only route users directly to specific security detail screens, order receipts, or corporate action ballots, but must also enforce ironclad cryptographic link verification (Android App Links, iOS Universal Links) to prevent phishing or transaction interception by rogue third-party apps.

This prompt specifies the universal deep linking and URL scheme routing architecture for the Growww Flutter client across Android, iOS, Windows, Linux, and macOS. It configures **Android App Links** with `assetlinks.json`, **Apple Universal Links** with `apple-app-site-association` (AASA), custom URL schemes (`growww://`), `GoRouter` route mapping with authentication state guards, UPI payment callback handling, and link spoofing defenses.

## What You Are Building
A cross-platform deep linking and routing infrastructure located in `lib/core/routing/`:
- `AppRouter`: Centralized `GoRouter` definition with declarative routes, route transition animations, query parameter parsing, and authentication state redirection (`redirect:` guard).
- `DeepLinkHandler`: Multi-platform intent listener capturing incoming URLs from cold start (`getInitialUri()`) and background resumes (`uriLinkStream`), normalizing both HTTPS universal links and `growww://` protocol schemes.
- `PaymentCallbackRouter`: Specialized routing handler parsing UPI and payment gateway return intents (`growww://payments/upi/callback?txnId=...&status=SUCCESS`), verifying payment signatures, and transitioning the wallet UI.
- `SecurityVerificationFiles`: Standardized server-side verification templates:
 - `.well-known/assetlinks.json` (Android Digital Asset Links).
 - `.well-known/apple-app-site-association` (Apple Universal Links AASA).
- `DeepLinkAuthGuard`: Stateful interceptor that preserves the deep link destination when an unauthenticated or session-expired user opens a link, prompts for biometric/MPIN login, and resumes navigation upon authentication.

## Scope Boundaries
- **In Scope:**
 - Configuration of Android App Links (`android:autoVerify="true"`).
 - Configuration of iOS Universal Links and entitlements.
 - Windows/Linux/macOS custom URL scheme registration (`growww://`).
 - Integration with `go_router` and `app_links` package.
 - Authentication redirection guards and query parameter validation.
 - UPI payment gateway callback parsing.
- **Out of Scope / Handled Elsewhere:**
 - Push notification token management (Prompt 522).
 - Web application routing (Prompt 601).
 - Payment gateway backend processing (Prompt 212).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `go_router` (v14.0+) and `app_links` (v6.0+).
- **Justification:** `app_links` provides modern multi-platform link listening across Android, iOS, Windows, Linux, and macOS in a single unified API, replacing fragmented legacy plugins. `go_router` offers first-class declarative URL routing, nested route shells, and reactive redirect guards.
- **Dependencies:**
 - `go_router: ^14.1.4`
 - `app_links: ^6.1.1`
 - `flutter_riverpod: ^2.5.1`

## Backend / Infra Touchpoints
- **Web Infrastructure CDN (`growww.in`):** Static hosting of `https://growww.in/.well-known/assetlinks.json` and `https://growww.in/.well-known/apple-app-site-association` with `application/json` MIME type and zero redirects.
- **Payment Gateway / Bank Webhook (Prompt 212):** Redirects mobile browsers and UPI apps to `growww://payments/upi/callback`.

## Blockchain Interaction
- **On-Chain Settlement Receipt Deep Linking:** Supports URLs structured as `https://growww.in/ledger/tx/:txHash` or `growww://ledger/tx/:txHash`, allowing investors to click transaction hashes in notifications and inspect immutable on-chain proof-of-reserve Merkle branches and DvP settlement blocks.

## Step-by-Step Build Instructions
1. Scaffold directory `lib/core/routing/` with `router.dart`, `deep_link_service.dart`, `route_names.dart`, and `guards/`.
2. Configure Android `AndroidManifest.xml` with intent filters:
 - HTTPS Universal Links: `<data android:scheme="https" android:host="growww.in" android:pathPrefix="/trade" />` with `android:autoVerify="true"`.
 - Custom Scheme: `<data android:scheme="growww" />`.
3. Configure iOS `Runner.entitlements` with `applinks:growww.in` and `Info.plist` with `CFBundleURLSchemes = ["growww"]`.
4. Configure Windows MSIX protocol activation in `pubspec.yaml` and Linux `.desktop` file with `x-scheme-handler/growww`.
5. Author server verification files:
 - Create `assetlinks.json` with package name `com.kerberosec.growww` and SHA-256 certificate fingerprints for debug and release keys.
 - Create `apple-app-site-association` with `appID: "TEAMID.com.kerberosec.growww"` and components mapping `/trade/*`, `/orders/*`, `/kyc/*`.
6. Implement `DeepLinkService` using `AppLinks`:
 - Listen to `appLinks.uriLinkStream` for active application lifecycle events.
 - Call `appLinks.getInitialLink()` during cold boot initialization.
7. Define declarative routes in `AppRouter` using `GoRouter`:
 - `/trade/:isin` -> `SecurityDetailScreen` (Prompt 508)
 - `/orders/:orderId` -> `OrderDetailScreen` (Prompt 512)
 - `/wallet/deposit/callback` -> `PaymentCallbackScreen` (Prompt 511)
 - `/kyc/status` -> `KycStatusScreen` (Prompt 504)
 - `/ledger/tx/:txHash` -> `LedgerVerificationScreen` (Prompt 510)
8. Implement `DeepLinkAuthGuard`: when an unauthenticated user opens `https://growww.in/trade/INE002A01018`, save the target path in `redirect` state, route user to `LoginScreen`, and upon successful authentication, execute `context.go(savedPath)`.
9. Implement `PaymentCallbackRouter` to parse query parameters (`txnId`, `status`, `signature`) from UPI gateway redirects, ensuring strict cryptographic signature validation before displaying confirmation.
10. Add sanitization logic to reject malformed, non-whitelisted, or suspicious URL schemes (preventing open redirect vulnerabilities).
11. Implement desktop command-line argument parser in `lib/main_desktop.dart` to capture URL arguments passed during executable invocation on Windows and Linux (`growww "growww://trade/INE002A01018"`).
12. Write unit tests for `DeepLinkHandler` asserting URI parsing across all supported path formats.
13. Write integration tests verifying that `GoRouter` correctly redirects unauthenticated deep link attempts to login and preserves destination state.
14. Test App Links verification on physical Android 14 device using `adb shell pm get-app-links com.kerberosec.growww`.
15. Test Universal Links on physical iOS device using Safari URL entry and Apple Notes links.

## Interfaces / Contracts

```json
// .well-known/assetlinks.json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.kerberosec.growww",
      "sha256_cert_fingerprints": [
        "14:6D:E9:01:0F:D7:58:65:42:01:E2:9B:F3:33:65:68:5A:F5:1B:32:00:1E:56:67:89:12:34:56:78:90:AB:CD"
      ]
    }
  }
]
```

```json
// .well-known/apple-app-site-association
{
  "applinks": {
    "apps": [],
    "details": [
      {
        "appID": "ABCDE12345.com.kerberosec.growww",
        "paths": [
          "/trade/*",
          "/orders/*",
          "/kyc/*",
          "/ledger/*",
          "/wallet/*"
        ]
      }
    ]
  }
}
```

```dart
// lib/core/routing/deep_link_service.dart
import 'package:app_links/app_links.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

class DeepLinkService {
  final AppLinks _appLinks = AppLinks();
  final GoRouter _router;

  DeepLinkService(this._router);

  void initialize() {
    // Cold start deep link
    _appLinks.getInitialLink().then((uri) {
      if (uri != null) _handleIncomingUri(uri);
    });

    // Foreground / Background deep link stream
    _appLinks.uriLinkStream.listen((uri) {
      _handleIncomingUri(uri);
    });
  }

  void _handleIncomingUri(Uri uri) {
    // Normalize HTTPS and custom scheme paths
    String targetPath = uri.path;
    if (uri.query.isNotEmpty) {
      targetPath = '$targetPath?${uri.query}';
    }

    // Validate path whitelist to prevent open redirection
    if (_isValidRoute(uri.path)) {
      _router.go(targetPath);
    }
  }

  bool _isValidRoute(String path) {
    final validPrefixes = ['/trade', '/orders', '/kyc', '/ledger', '/wallet', '/payments'];
    return validPrefixes.any((prefix) => path.startsWith(prefix));
  }
}
```

## Security & Compliance Notes
- **App Link Verification Integrity:** Never rely solely on custom schemes (`growww://`) for sensitive operations, as unverified custom schemes can be claimed by rogue applications on Android. All critical entry points must use cryptographically verified HTTPS App Links / Universal Links.
- **Open Redirect Prevention:** Strict path whitelisting must reject any incoming link containing unexpected host redirects, javascript schemes, or un-sanitized external URLs.
- **Payment Signature Validation:** UPI and bank payment return callbacks must verify that `signature` parameters match the backend's HMAC-SHA256 signature before transitioning UI state or assuming payment completion.
- **Authentication State Isolation:** Deep links to private financial data (holdings, order receipts, bank accounts) MUST never bypass authentication. If the session is unauthenticated, the app must require full biometrics/MPIN before presenting the requested view.

## Acceptance Criteria
- [ ] Clicking a `https://growww.in/trade/INE...` link in browser or messages opens the app directly without showing browser disambiguation dialogs.
- [ ] Custom URL scheme `growww://` functions across Android, iOS, Windows, Linux, and macOS.
- [ ] Unauthenticated users attempting to access protected deep links are directed to login, and automatically navigate to the deep-linked screen upon successful authentication.
- [ ] Payment gateway callbacks from external UPI apps correctly resume the app and process transaction status parameters.
- [ ] Server verification files (`assetlinks.json` and `apple-app-site-association`) pass Google Digital Asset Links and Apple CDN validation checks.
- [ ] Deep link intent handling is covered by comprehensive unit and router integration tests.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 502 (App Architecture & State Management), Prompt 505 (Authentication UI), Prompt 521 (Local Secure Storage).
- **Parallel Tasks:** Prompt 522 (Push Notifications), Prompt 525 (API Client Layer).
