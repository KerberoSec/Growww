# 513 - Flutter Notifications Center UI

## Purpose
Provides a centralized, real-time alert and notification hub for investors across mobile and desktop devices. In a high-stakes regulated investment environment, investors require instant notifications for order fills, atomic DvP settlement confirmations, price alerts, corporate dividend distributions, monthly Proof-of-Reserve attestation announcements, and critical security/compliance notices (such as unrecognized login attempts or SEBI regulatory circulars).

## What You Are Building
A responsive, rich notification center in `apps/growww_flutter/lib/features/notifications/` featuring:
- **Notification Inbox View:** Categorized notification list with filtering tabs: All, Orders & Trades, Price Alerts, Blockchain & Custody, Security & Regulatory.
- **Unread Counter & Real-Time Badges:** Live unread notification badge on the app header with instant count updates via WebSocket.
- **Interactive Actionable Notifications:** Swipe-to-dismiss, mark-as-read, clear-all, and one-tap deep-link navigation directly to the relevant order, holding, or security chart.
- **Proof-of-Reserve Broadcast Cards:** Highlighting monthly cryptographic custody attestation releases with direct links to the verification tool.
- **Notification Preferences Modal:** User settings to toggle push/in-app alert channels for trade executions, market volatility, and corporate actions.

## Scope Boundaries
- **In Scope:**
 - In-app notification center UI, category filtering, unread badge counter state, swipe interactions, deep-link routing from notification taps, and notification preferences sheet.
- **Out of Scope / Handled Elsewhere:**
 - Backend notification dispatcher and SMS/Email gateway (handled in Prompt 211).
 - Platform-level push notification integration (FCM, APNs, desktop notifications) (handled in Prompt 522).
 - Deep-link URL router infrastructure (handled in Prompt 526).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` state management.
 - *Justification:* Riverpod state providers enable instant cross-screen badge synchronization (e.g., clearing a notification on desktop immediately updates the badge across the UI).
- **List Animations:** `flutter_slidable` for smooth swipe-to-delete and swipe-to-read actions.
- **Badges:** `badges` package or custom `Badge` widget for responsive unread counters.

## Backend / Infra Touchpoints
- **Notification Service:** REST endpoints (`/api/v1/notifications/inbox`, `/api/v1/notifications/mark-read`, `/api/v1/notifications/preferences`) from Prompt 211.
- **Real-Time Notification Channel:** WebSocket stream (`wss://ws.growww.in/v1/notifications/stream`) for real-time in-app toasts and counter increments.

## Blockchain Interaction
- **On-Chain Event Notifications:**
 - Delivers notifications for ledger events:
 - *DvP Settlement Complete:* "Your order for 0.5 shares of TCS has settled atomically on Hyperledger Besu Block #1,492,100."
 - *Proof-of-Reserve Attestation:* "August 2026 Proof-of-Reserve published: 100% 1:1 backing verified by Independent Custody Auditor."
 - *Dividend Distribution:* "Smart contract distributed ₹14.50 dividend per fractional share to your wallet."

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/notifications/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/`.
2. Define domain models: `NotificationItem`, `NotificationCategory` (orders, priceAlerts, blockchain, regulatory, security), `NotificationPriority`.
3. Implement `NotificationInboxNotifier` managing pagination, unread counts, and optimistic "mark-as-read" state updates.
4. Build `NotificationsCenterScreen` with top `TabBar` (All, Orders, Alerts, Blockchain, Security) and a "Mark All as Read" header button.
5. Create `NotificationCardWidget` rendering category-specific icons, bold title, timestamp, and unread blue dot indicator.
6. Implement `Slidable` swipe actions allowing users to swipe left to delete or swipe right to toggle read/unread state.
7. Build deep-link router integration: tapping an order notification routes to `TradeHistoryDetail` (Prompt 512), tapping a security alert opens `SecurityDetailScreen` (Prompt 508), tapping a Proof-of-Reserve alert opens `PortfolioVerificationSheet` (Prompt 510).
8. Implement pinned security alert cards for critical compliance or unrecognized login notices with high-contrast amber/red styling.
9. Implement real-time WebSocket listener that plays a subtle notification chime and increments the badge counter when a new alert is received.
10. Build `NotificationPreferencesSheet` enabling granular toggles (Trade Fills, Price Alerts, Daily Wrap-up, Reserve Attestation).
11. Add empty-state illustrations for empty tabs ("You're all caught up!").
12. Write widget tests verifying unread badge decrement on tap, swipe-to-dismiss behavior, and deep-link routing.

## Interfaces / Contracts
```dart
// lib/features/notifications/domain/models/notification_item.dart
enum NotificationCategory { orders, priceAlerts, blockchainCustody, regulatory, security }
enum NotificationPriority { normal, high, urgent }

class NotificationItem {
  final String id;
  final String title;
  final String body;
  final NotificationCategory category;
  final NotificationPriority priority;
  final bool isRead;
  final DateTime createdAt;
  final String? deepLinkUri; // e.g., "growww://orders/ORD-12345" or "growww://security/INE002A01018"
  final Map<String, dynamic>? metadata;

  const NotificationItem({
    required this.id,
    required this.title,
    required this.body,
    required this.category,
    required this.priority,
    required this.isRead,
    required this.createdAt,
    this.deepLinkUri,
    this.metadata,
  });
}
```

## Security & Compliance Notes
- **Mandatory Regulatory Notices:** Regulatory alerts and SEBI circulars cannot be swiped away without viewing; they require explicit acknowledgment if marked `urgent`.
- **Security Notice Pinning:** Alerts concerning password changes, MPIN resets, or logins from new devices remain permanently pinned to the top of the inbox until acknowledged.
- **Zero Sensitive PII in Notifications:** Push and in-app notifications must not expose full account numbers or bank account numbers in cleartext.

## Acceptance Criteria
- [ ] Notifications render categorized lists with smooth tab switching.
- [ ] Unread notification count badge updates in real time upon new WebSocket message.
- [ ] Swiping notification triggers smooth delete/mark-as-read transitions.
- [ ] Tapping a notification navigates directly to the target screen via deep link.
- [ ] Blockchain Proof-of-Reserve notifications include verified block number and link to verification modal.
- [ ] Notification preferences sync cleanly with the backend notification service.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System).
- **Backend Dependency:** Prompt 211 (Notification Service), Prompt 522 (Push Notifications).
- **Enables:** Prompt 514 (Settings & Profile UI), Prompt 526 (Deep Linking).
