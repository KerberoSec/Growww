# 522 - Cross-Platform Push Notification Integration (Mobile FCM/APNs & Desktop System Notifications)

## Purpose
Real-time investor awareness is critical in digital securities trading. Immediate notifications are mandated for trade fills, DvP settlement receipts, margin alerts, KYC verification milestones, and suspicious login detections. Investors utilize diverse devices ranging from mobile smartphones on volatile cellular networks to continuous desktop trading workstations.

This prompt defines the unified push notification architecture for the Growww Flutter client across Android, iOS, Windows, Linux, and macOS. It integrates **Firebase Cloud Messaging (FCM)** for Android, **Apple Push Notification service (APNs)** for iOS and macOS, and **system tray / desktop notification daemons** (Windows Toast Notifications, macOS UserNotifications, Linux `org.freedesktop.Notifications`) for desktop clients. It establishes high-priority notification channels, secure payload decryption, background message handlers, and deep-link routing to relevant order and portfolio views.

## What You Are Building
A cross-platform notification engine in `lib/core/notifications/`:
- `PushNotificationService`: Core abstraction managing device registration, APNs/FCM token extraction, server registration, and topic subscriptions.
- `PlatformNotificationRenderer`: Native UI notification dispatcher leveraging `flutter_local_notifications` and desktop platform channels to display rich system notifications with action buttons.
- `NotificationPayloadRouter`: Intent parser extracting deep links (`growww://trade/order-123`) from background, terminated, and foreground notification payloads and navigating via `GoRouter`.
- `BackgroundMessageHandler`: High-performance background isolate handler processing data-only silent push messages (e.g. invalidating portfolio caches or triggering local sync).
- `NotificationPermissionManager`: Granular permission coordinator explaining the regulatory necessity of trade alert notifications and guiding users through platform permission dialogs.

## Scope Boundaries
- **In Scope:**
 - FCM token lifecycle and APNs device token bridging.
 - Notification channel setup (Android 8.0+ `NotificationChannel` with high importance).
 - Desktop native notification integration (Windows WinRT Toasts, macOS UserNotifications, Linux D-Bus).
 - Background isolate message handling for silent and visible pushes.
 - Foreground HUD banners and in-app sound/vibration haptics.
 - Deep-link payload dispatching to UI screens.
- **Out of Scope / Handled Elsewhere:**
 - Backend notification service and delivery orchestration (Prompt 211).
 - In-app Notification Center UI screen (Prompt 513).
 - Deep linking routing implementation (Prompt 526).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `firebase_core`, `firebase_messaging`, and `flutter_local_notifications`.
- **Justification:** `firebase_messaging` provides battle-tested mobile delivery with APNs translation, while `flutter_local_notifications` provides cross-platform desktop native notification rendering (WinRT, macOS NSUserNotification, Linux D-Bus).
- **Dependencies:**
 - `firebase_core: ^3.1.0`
 - `firebase_messaging: ^15.0.0`
 - `flutter_local_notifications: ^17.1.2`
 - `flutter_riverpod: ^2.5.1`
 - `go_router: ^14.0.0`

## Backend / Infra Touchpoints
- **Notification Service (Prompt 211):** `POST /api/v1/notifications/devices/register` (device token, platform, app version, language), `DELETE /api/v1/notifications/devices/{token}`.
- **Kafka Push Event Consumer (Prompt 104):** Consumes topics `orders.matched`, `settlement.completed`, `auth.security_alert`.

## Blockchain Interaction
- **On-Chain Settlement Alerts:** Push notifications trigger when `SettlementDvP.sol` or `DigitalSecurityToken.sol` emits on-chain event logs (`TokensTransferred`, `SettlementFinalized`) indexed by the blockchain indexer (Prompt 309).
- **Proof-of-Reserve Attestation Alerts:** Delivers automated periodic attestations informing users of daily Merkle root updates published to `ProofOfReserveRegistry.sol`.

## Step-by-Step Build Instructions
1. Scaffold directory `lib/core/notifications/` with `domain/`, `data/`, and `presentation/` subdirectories.
2. Configure Android `AndroidManifest.xml` with permissions: `POST_NOTIFICATIONS`, `VIBRATE`, and `WAKE_LOCK`.
3. Configure iOS `AppDelegate.swift` and macOS `AppDelegate.swift` enabling remote notifications capability and registering for remote notifications with APNs.
4. Set up Android Notification Channels in Dart:
 - `trades_channel`: High Importance, sound enabled, vibration enabled (for order fills and margin alerts).
 - `account_security_channel`: Urgent Importance, bypass DND where permitted (for OTPs and unauthorized login attempts).
 - `market_news_channel`: Default Importance, low sound (for price alerts and corporate action notices).
5. Implement `PushNotificationService` interface and initialize `FirebaseMessaging.instance` on mobile targets.
6. Initialize `flutter_local_notifications` plugin for desktop platforms (Windows, Linux, macOS) with custom icon drawables.
7. Implement `BackgroundMessageHandler` annotated with `@pragma('vm:entry-point')` to handle data-only background notifications safely in a detached Dart isolate.
8. Create device registration mechanism: upon user authentication, retrieve FCM/APNs token, send device metadata to backend Notification Service, and securely store the token locally.
9. Implement foreground message listener: when app is open, display an interactive top overlay banner with sound and haptic feedback.
10. Implement notification tap handler for three lifecycle states:
 - **Foreground:** Tapping on local HUD banner navigates directly.
 - **Background:** `FirebaseMessaging.onMessageOpenedApp` routes payload.
 - **Terminated:** `FirebaseMessaging.instance.getInitialMessage()` retrieves launch payload on cold start.
11. Implement `NotificationPayloadRouter` translating payload types (`ORDER_FILL`, `PRICE_ALERT`, `SECURITY_EVENT`) into `GoRouter` destination routes.
12. Handle token refresh events (`FirebaseMessaging.instance.onTokenRefresh`) by dispatching updated tokens immediately to the backend.
13. Implement token un-registration and topic unsubscription during user logout.
14. Write unit tests for payload parsing and route resolution logic.
15. Verify notification rendering and tap navigation on physical Android, iOS, Windows, Linux, and macOS devices.

## Interfaces / Contracts

```dart
// lib/core/notifications/domain/notification_payload.dart
enum NotificationType {
  tradeExecuted,
  dvpSettlementCompleted,
  marginCall,
  priceAlert,
  securityAlert,
  corporateAction,
  generalInfo
}

class NotificationPayload {
  final String notificationId;
  final NotificationType type;
  final String title;
  final String body;
  final String? deepLink;
  final Map<String, dynamic> data;
  final DateTime timestamp;

  NotificationPayload({
    required this.notificationId,
    required this.type,
    required this.title,
    required this.body,
    this.deepLink,
    required this.data,
    required this.timestamp,
  });

  factory NotificationPayload.fromMap(Map<String, dynamic> map) {
    return NotificationPayload(
      notificationId: map['notification_id'] ?? '',
      type: NotificationType.values.firstWhere(
        (e) => e.name == map['type'],
        orElse: () => NotificationType.generalInfo,
      ),
      title: map['title'] ?? '',
      body: map['body'] ?? '',
      deepLink: map['deep_link'],
      data: Map<String, dynamic>.from(map['data'] ?? {}),
      timestamp: DateTime.tryParse(map['timestamp'] ?? '') ?? DateTime.now(),
    );
  }
}

// Notification Service Contract
abstract class IPushNotificationService {
  Future<void> initialize();
  Future<bool> requestPermissions();
  Future<String?> getDeviceToken();
  Future<void> registerDeviceWithBackend(String token);
  Future<void> unregisterDevice();
  Future<void> subscribeToTopic(String topic);
  Future<void> unsubscribeFromTopic(String topic);
  Stream<NotificationPayload> get onNotificationReceived;
  Stream<NotificationPayload> get onNotificationTapped;
}
```

## Security & Compliance Notes
- **PII & Financial Privacy in Push Payloads:** In adherence to SEBI and RBI cybersecurity mandates, push notification bodies must never contain unmasked financial values (e.g. exact bank account balance, full PAN) or sensitive security tokens. Use localized string templates with generic summaries (e.g., "Your buy order for 0.5 units of RELIANCE has been filled").
- **Token Invalidation on Logout:** When an investor logs out, the client must immediately instruct the backend to dissociate the device token from the user account, preventing subsequent notifications from leaking to shared devices.
- **Silent Push Cryptography:** If silent data pushes are used for remote cache invalidation, payloads must contain a cryptographic server signature to prevent spoofed triggers.

## Acceptance Criteria
- [ ] Push notifications deliver reliably to physical Android and iOS devices in Foreground, Background, and Terminated states.
- [ ] Desktop notifications display natively on Windows (WinRT Toast), macOS (NotificationCenter), and Linux (D-Bus).
- [ ] Tapping a notification correctly launches or resumes the app and deep-links directly to the relevant screen (e.g. Order Details).
- [ ] Notification channels on Android allow granular user control over trade alerts vs. promotional alerts.
- [ ] Device token refreshes are detected and synchronized with the backend.
- [ ] User logout removes device token registration from the backend.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 211 (Notification Service), Prompt 521 (Local Secure Storage), Prompt 526 (Deep Linking).
- **Parallel Tasks:** Prompt 513 (Notifications Center UI), Prompt 524 (Crash Reporting & Analytics).
