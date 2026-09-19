# 538 - Flutter P2P Fiat Trading & Encrypted Chat Escrow Screen

## Purpose
Provides an institutional-grade, peer-to-peer (P2P) fiat-to-crypto trading and escrow interface for the Growww multi-platform client (Android, iOS, macOS, Windows, Linux). Operating within the regulatory framework of ADR-0039 (Peer-to-Peer Automated Escrow Lockbox & Dispute Arbitration), direct peer-to-peer bank transfers (IMPS, NEFT, RTGS) and Unified Payments Interface (UPI) rails require trustless digital asset collateralization, end-to-end encrypted (E2EE) communications, and automated dispute arbitration workflows.

P2P fiat trading exposes retail and institutional participants to critical counterparty risks, including payment fraud, unauthorized third-party bank transfers, fraudulent chargeback claims, triangle arbitration scams, and off-platform social engineering. This specification defines the presentation architecture, Riverpod reactive state management, dynamic vector UPI QR code generation, client-side cryptographic chat session management (Signal Double Ratchet protocol), tamper-resistant payment screenshot watermarking, dynamic escrow countdown timers, biometric release authorization, on-chain multisig escrow lock visualization, and maker-checker dispute escalation.

## What You Are Building
A responsive, high-security P2P fiat trading and escrow suite located in `apps/growww_flutter/lib/screens/p2p/` and supporting modules in `apps/growww_flutter/lib/features/p2p_trading/`:
- `P2POrderWizardScreen`: Role-adaptive trading workflow wizard dynamically presenting customized views based on whether the authenticated user is the Buyer or Seller, and synchronizing order states (`ESCROW_LOCK_PENDING`, `ESCROW_LOCKED`, `BUYER_PAID`, `DISPUTED`, `SETTLED`, `CANCELLED`).
- `P2PEscrowBuyerStepView`: Buyer workflow interface displaying seller bank/UPI credentials, dynamic vector UPI QR codes with pre-filled amounts, one-tap clipboard copy helpers with haptic feedback, a 15-minute escrow countdown timer, tamper-resistant payment screenshot uploader, and "Payment Sent" confirmation action.
- `P2PEscrowSellerStepView`: Seller monitoring interface displaying on-chain collateral lock status, verified buyer KYC credentials, real-time incoming fiat checklist, anti-fraud warnings, and a biometric-gated "Release Crypto" one-tap release action.
- `P2PEncryptedChatView`: Embedded end-to-end encrypted messaging interface implementing the Signal Double Ratchet protocol over WebSockets, featuring delivery/read receipts, encrypted media transfers, anti-fraud keyword heuristics, and visual anti-phishing security phrase banners.
- `UpiQrGeneratorWidget`: Dynamic vector QR code generator adhering to the National Payments Corporation of India (NPCI) UPI specification (`upi://pay?pa=...&pn=...&am=...&cu=INR&tn=...`), enabling direct scanning from external banking apps.
- `WatermarkedScreenshotUploaderWidget`: Camera and photo picker interface applying immutable client-side watermarks (Order ID, UTC timestamp, Buyer User ID, and fraud disclaimer) across payment receipts prior to S3 upload.
- `DisputeEscalationModal`: Maker-checker dispute escalation sheet allowing counterparties to trigger formal arbitration upon timer expiration or payment discrepancy, capturing bank Unique Transaction Reference (UTR) numbers and supporting proof.
- `EscrowStatusCard`: Real-time on-chain collateral status widget rendering Hyperledger Besu / EVM escrow smart contract lock confirmations, block inclusion depth, and 2-of-3 multisig resolution state.
- `Riverpod State Layer`: Immutable state models, asynchronous notifiers (`P2POrderNotifier`, `P2PChatNotifier`, `P2PEscrowTimerNotifier`, `P2PDisputeNotifier`), and WebSocket stream providers delivering live order updates and encrypted messages.

## Scope Boundaries
- **In Scope:**
  - Multi-platform Flutter UI screens, widgets, and step wizards in `lib/screens/p2p/` and `lib/features/p2p_trading/`.
  - Buyer payment flow: Seller payment method presentation, UPI vector QR generation, 15-minute countdown timer, watermarked receipt upload, and payment declaration.
  - Seller release flow: Buyer KYC verification badge, payment receipt inspection, biometric authorization challenge, and one-tap cryptographic release dispatch.
  - End-to-end encrypted chat widget: Local Signal Double Ratchet session handling, key derivation, encrypted message framing, typing indicators, and media attachments.
  - Client-side payment screenshot watermarking using pure Dart canvas processing before network upload.
  - Visual anti-phishing security phrase validation matching user profile settings.
  - Real-time on-chain escrow lock confirmation and multisig state tracking via WebSockets.
  - Formal dispute escalation modal with UTR entry and proof submission.
  - Strict KYC Name Invariant enforcement: Highlighting discrepancies between bank account holder names and KYC identity.
- **Out of Scope / Handled Elsewhere:**
  - Backend P2P Escrow microservice business logic, matching engine, and order book persistence (Prompt 269).
  - Automated Open Banking / Account Aggregator API integrations for direct bank statement polling (Prompt 212, Prompt 215).
  - Smart contract implementation for on-chain escrow lockbox and multi-party custody (Prompt 305, Prompt 311).
  - Push notification dispatch infrastructure via FCM and APNs (Notification Service, Prompt 211).
  - Centralized KYC verification, document OCR, and biometric face matching (Prompt 202).
  - Cloud storage bucket provisioning, IAM policies, and lifecycle management (Prompt 805).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (v2.5+) for deterministic state management and dependency injection.
- **Justification:** Riverpod provides clean lifecycle handling (`AutoDisposeAsyncNotifier`), testable domain decoupling, and native handling of real-time WebSocket streams without UI rebuild leaks.
- **Dependencies:**
  - `flutter_riverpod: ^2.5.1` (reactive state management)
  - `web_socket_channel: ^3.0.0` (bi-directional real-time order and chat streaming)
  - `dio: ^5.4.3+1` (REST communication with S3 pre-signed upload support)
  - `qr_flutter: ^4.1.0` (dynamic vector UPI QR code rendering)
  - `libsignal_protocol_dart: ^0.6.4` (Signal Double Ratchet cryptographic session management)
  - `cryptography: ^2.7.0` (AES-256-GCM and HMAC primitives for media encryption)
  - `image_picker: ^1.1.2` (camera and gallery payment receipt capture)
  - `image: ^4.2.0` (pure Dart pixel processing for canvas watermarking without native platform dependencies)
  - `local_auth: ^2.2.0` (biometric fingerprint / Face ID authorization for fund release)
  - `flutter_secure_storage: ^9.2.2` (hardware-backed secure storage for Double Ratchet identity keys)
  - `vibration: ^2.0.1` (haptic tactile feedback for copy actions and timer milestones)
  - `intl: ^0.19.0` (fiat currency and countdown formatting)

## Backend / Infra Touchpoints
- **P2P Escrow Service (Prompt 269):**
  - `GET /api/v1/p2p/orders/{order_id}`: Retrieves comprehensive order details, escrow lock parameters, and counterparty KYC verification metadata.
  - `POST /api/v1/p2p/orders/{order_id}/mark-paid`: Buyer marks payment as transferred, submitting the bank UTR reference and receipt URL.
  - `POST /api/v1/p2p/orders/{order_id}/release`: Seller authorizes the release of escrowed crypto following biometric authorization.
  - `POST /api/v1/p2p/orders/{order_id}/cancel`: Buyer cancels the order prior to payment submission.
  - `POST /api/v1/p2p/orders/{order_id}/dispute`: Either counterparty files a formal dispute, locking escrow for maker-checker arbitration.
  - `GET /ws/v1/p2p/orders/{order_id}/stream`: WebSocket channel delivering real-time state machine transitions and on-chain lock confirmations.
- **Encrypted Chat Signaling Gateway (Prompt 269 / Prompt 219):**
  - `GET /ws/v1/p2p/chat/{order_id}`: Ephemeral WebSocket tunnel transporting encrypted Signal message frames between buyer and seller.
  - `GET /api/v1/p2p/chat/{order_id}/prekey-bundle`: Fetches the counterparty pre-key bundle for initial Signal handshake.
- **S3 Storage Gateway (Prompt 805):**
  - `POST /api/v1/p2p/orders/{order_id}/proof-upload-url`: Obtains pre-signed S3 upload URLs with mandatory client-side encrypted metadata headers.
- **Notification Service (Prompt 211):**
  - Subscribes client devices to high-priority FCM/APNs push notifications for timer alerts ("5 minutes remaining before automatic cancellation"), payment transferred alerts, and crypto release notices.
- **User Profile & Security Service (Prompt 201):**
  - Fetches the user anti-phishing security phrase to render in the chat header, confirming the authentic Growww client session.

## Blockchain Interaction
- **Hyperledger Besu & EVM Escrow Settlement:**
  - P2P crypto collateral is locked inside an on-chain `P2PEscrowLockbox.sol` smart contract deployed on Hyperledger Besu (or designated EVM L2 settlement rails).
  - The seller funds are locked atomically upon order creation via a 2-of-3 multisig arrangement (Seller Key, Buyer Key, Growww Arbiter Key).
- **Client-Side Escrow Visualization:**
  - `EscrowStatusCard` tracks the lifecycle of the smart contract lock:
    - `AWAITING_COLLATERAL`: Seller has initiated the trade; awaiting on-chain lock transaction submission.
    - `CONFIRMING_LOCK`: Collateral deposit transaction submitted to the network; tracking block depth (e.g. 4/12 confirmations).
    - `COLLATERAL_LOCKED`: Smart contract event `EscrowLocked(bytes32 orderId, address seller, uint256 amount)` finalized.
    - `RELEASE_PENDING`: Seller signed release; smart contract executing `releaseFunds(orderId)`.
    - `SETTLED`: Crypto transferred to buyer custody wallet; verified by transaction receipt.
    - `ARBITRATION_LOCKED`: Dispute active; release gated behind Growww Compliance Arbiter signature.
  - Displays public block explorer links, transaction hashes, smart contract addresses, and token decimals (e.g. USDT, USDC, BTC, ETH).

## Step-by-Step Build Instructions
1. Scaffold feature and screen directories under `apps/growww_flutter/lib/screens/p2p/` and `apps/growww_flutter/lib/features/p2p_trading/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `domain/interfaces/`, `data/repositories/`, and `data/datasources/`.
2. Define domain enums and models in `domain/models/`: `P2POrderStatus`, `P2PUserRole`, `PaymentMethodType`, `EscrowOnChainState`, `P2POrder`, `PaymentAccountDetails`, `P2PEscrowStatus`, `EncryptedChatMessage`, `ChatPayloadType`, and `DisputeCase`.
3. Implement secure storage key management in `data/datasources/p2p_keystore.dart` using `flutter_secure_storage` to persist Signal Identity Keys, Pre-Keys, and Signed Pre-Keys for E2EE trade chat.
4. Implement `SignalEncryptionService` implementing `IP2PChatEncryptionService`:
   - Generates local identity keys and initializes sessions using counterparty pre-key bundles.
   - Encrypts outgoing text and image media payloads using Double Ratchet sessions.
   - Decrypts incoming ciphertext frames and handles ratchet key updates.
5. Create `WatermarkProcessor` utility in `domain/services/watermark_processor.dart`:
   - Ingests raw screenshot byte data via `image_picker`.
   - Renders semi-transparent, diagonal, multi-line watermark text containing: `GROWWW P2P PROOF | ORDER: {order_id} | USER: {buyer_id} | {utc_timestamp} | FOR ESCROW PURPOSES ONLY`.
   - Compresses the resulting image into optimized JPEG format to minimize bandwidth consumption.
6. Create `UpiUriBuilder` in `domain/services/upi_uri_builder.dart`:
   - Builds NPCI-compliant UPI URI strings: `upi://pay?pa={vpa}&pn={payee_name}&am={amount}&cu=INR&tn=Growww_P2P_{order_id}`.
   - Validates Virtual Payment Address (VPA) format before building QR payloads.
7. Implement `P2PEscrowTimerNotifier` (Riverpod `AutoDisposeAsyncNotifier<Duration>`):
   - Computes remaining time based on order expiration timestamp from backend.
   - Ticks once per second, triggering haptic alerts at critical thresholds (5 minutes, 2 minutes, 30 seconds).
   - Emits expiration events triggering UI state transitions to `EXPIRED`.
8. Implement `P2POrderNotifier` (Riverpod `AutoDisposeAsyncNotifier<P2POrder>`):
   - Fetches active order state from `P2PEscrowRepository`.
   - Connects to `/ws/v1/p2p/orders/{order_id}/stream` for real-time WebSocket push updates.
   - Dispatches `markAsPaid`, `releaseCrypto`, `cancelOrder`, and `escalateDispute` commands.
9. Implement `P2PChatNotifier` (Riverpod `AutoDisposeAsyncNotifier<List<EncryptedChatMessage>>`):
   - Connects to `/ws/v1/p2p/chat/{order_id}`.
   - Maintains chat message history, manages unread message counts, handles outgoing message queueing, and processes inbound ciphertext streams.
10. Build `P2PEscrowBuyerStepView`:
    - Top order header with fiat amount, crypto quantity, unit price, and counterparty KYC verification badge.
    - Prominent dynamic countdown timer card with color-coded status (Green > 5m, Amber 2m-5m, Red < 2m).
    - Payment details card showing seller bank name, account number, IFSC code, and UPI ID with one-tap copy buttons.
    - Vector QR code widget (`UpiQrGeneratorWidget`) rendered using `QrImageView`.
    - Payment proof uploader (`WatermarkedScreenshotUploaderWidget`) with receipt preview and watermark verification badge.
    - Mandatory UTR number text field.
    - Primary CTA: "Transferred, Notify Seller" with secondary "Cancel Order" action.
11. Build `P2PEscrowSellerStepView`:
    - On-chain escrow status banner (`EscrowStatusCard`) showing locked crypto collateral on Hyperledger Besu.
    - Buyer identity card with verified KYC name and anti-fraud checklist (e.g. "Ensure bank account sender name exactly matches KYC name").
    - Payment receipt viewer displaying buyer watermarked screenshot and reported UTR number.
    - Countdown timer indicating remaining buyer payment window.
    - Primary CTA: "Release Crypto" gated behind `LocalAuthService` biometric challenge and secondary confirmation modal.
12. Build `P2PEncryptedChatView`:
    - Security header displaying user anti-phishing security phrase and E2EE lock indicator.
    - Bubble list with delivery indicators (Sent, Delivered, Read) and timestamp.
    - Image attachment viewer rendering watermarked receipts with zoom/pan capabilities.
    - Anti-fraud warning banner intercepting phone numbers, external email addresses, and keywords (e.g. "telegram", "whatsapp", "call me").
13. Build `DisputeEscalationModal`:
    - Triggerable only after payment timer expiration or when payment discrepancy is declared.
    - Radio selection for dispute grounds: "Payment not received", "Incorrect amount received", "Third-party bank account used", "Buyer marked paid without paying".
    - Evidence submission form requiring UTR number, bank statement PDF/JPEG attachment, and written description.
14. Build `P2POrderWizardScreen`:
    - Main parent scaffold hosting dynamic tab navigation between Trade Steps and Encrypted Chat.
    - Displays persistent unread chat message badge over chat tab icon.
    - Implements Android `FLAG_SECURE` window protections on sensitive payment detail views.
15. Write comprehensive unit and widget tests:
    - Unit tests for `WatermarkProcessor` ensuring watermark text inclusion and pixel integrity.
    - Unit tests for `UpiUriBuilder` verifying NPCI formatting.
    - Unit tests for `SignalEncryptionService` verifying ciphertext roundtrip encryption and decryption.
    - Widget tests for `P2PEscrowBuyerStepView` and `P2PEscrowSellerStepView` validating buyer and seller conditional rendering.

## Interfaces / Contracts

### Domain Enums and Data Models
```dart
// apps/growww_flutter/lib/features/p2p_trading/domain/models/p2p_enums.dart

enum P2POrderStatus {
  created,
  escrowLockPending,
  escrowLocked,
  buyerPaid,
  disputed,
  settled,
  cancelled,
  expired,
}

enum P2PUserRole {
  buyer,
  seller,
}

enum PaymentMethodType {
  upi,
  imps,
  neft,
  rtgs,
}

enum EscrowOnChainState {
  awaitingCollateral,
  confirmingLock,
  collateralLocked,
  releasePending,
  settledOnChain,
  disputeArbitration,
  failed,
}

enum ChatPayloadType {
  text,
  imageProof,
  systemNotice,
}

enum DisputeReason {
  paymentNotReceived,
  partialPaymentReceived,
  thirdPartyAccountUsed,
  buyerNotResponding,
  fraudulentProofUploaded,
}
```

```dart
// apps/growww_flutter/lib/features/p2p_trading/domain/models/p2p_order_models.dart

class CounterpartyKYC {
  final String userId;
  final String verifiedFullName;
  final String maskedPanNumber;
  final bool isKycVerified;
  final double completionRatePercent;
  final int totalCompletedTrades;

  const CounterpartyKYC({
    required this.userId,
    required this.verifiedFullName,
    required this.maskedPanNumber,
    required this.isKycVerified,
    required this.completionRatePercent,
    required this.totalCompletedTrades,
  });
}

class PaymentAccountDetails {
  final PaymentMethodType methodType;
  final String accountHolderName;
  final String? bankName;
  final String? accountNumber;
  final String? ifscCode;
  final String? upiId;
  final String? qrCodePayload;

  const PaymentAccountDetails({
    required this.methodType,
    required this.accountHolderName,
    this.bankName,
    this.accountNumber,
    this.ifscCode,
    this.upiId,
    this.qrCodePayload,
  });
}

class P2PEscrowStatus {
  final EscrowOnChainState state;
  final String escrowContractAddress;
  final String? lockTransactionHash;
  final String? releaseTransactionHash;
  final int currentConfirmations;
  final int requiredConfirmations;
  final int multisigThreshold;
  final int multisigSignaturesCount;

  const P2PEscrowStatus({
    required this.state,
    required this.escrowContractAddress,
    this.lockTransactionHash,
    this.releaseTransactionHash,
    required this.currentConfirmations,
    required this.requiredConfirmations,
    required this.multisigThreshold,
    required this.multisigSignaturesCount,
  });
}

class P2POrder {
  final String orderId;
  final P2PUserRole currentUserRole;
  final P2POrderStatus status;
  final String assetSymbol;
  final double cryptoAmount;
  final double fiatAmountInr;
  final double unitPriceInr;
  final CounterpartyKYC counterparty;
  final PaymentAccountDetails sellerPaymentDetails;
  final P2PEscrowStatus escrowStatus;
  final DateTime createdAt;
  final DateTime paymentWindowExpiresAt;
  final String? paymentUtr;
  final String? paymentProofUrl;
  final String antiPhishingPhrase;

  const P2POrder({
    required this.orderId,
    required this.currentUserRole,
    required this.status,
    required this.assetSymbol,
    required this.cryptoAmount,
    required this.fiatAmountInr,
    required this.unitPriceInr,
    required this.counterparty,
    required this.sellerPaymentDetails,
    required this.escrowStatus,
    required this.createdAt,
    required this.paymentWindowExpiresAt,
    this.paymentUtr,
    this.paymentProofUrl,
    required this.antiPhishingPhrase,
  });
}

class EncryptedChatMessage {
  final String messageId;
  final String orderId;
  final String senderId;
  final ChatPayloadType payloadType;
  final String plaintextContent;
  final String? mediaUrl;
  final DateTime timestamp;
  final bool isDelivered;
  final bool isRead;
  final bool isSentByMe;

  const EncryptedChatMessage({
    required this.messageId,
    required this.orderId,
    required this.senderId,
    required this.payloadType,
    required this.plaintextContent,
    this.mediaUrl,
    required this.timestamp,
    required this.isDelivered,
    required this.isRead,
    required this.isSentByMe,
  });
}

class DisputeCase {
  final String disputeId;
  final String orderId;
  final String initiatorUserId;
  final DisputeReason reason;
  final String description;
  final String utrNumber;
  final List<String> evidenceUrls;
  final DateTime filedAt;

  const DisputeCase({
    required this.disputeId,
    required this.orderId,
    required this.initiatorUserId,
    required this.reason,
    required this.description,
    required this.utrNumber,
    required this.evidenceUrls,
    required this.filedAt,
  });
}
```

### Repository and Service Contracts
```dart
// apps/growww_flutter/lib/features/p2p_trading/domain/interfaces/p2p_contracts.dart

abstract class IP2PEscrowRepository {
  Future<P2POrder> getOrderDetails(String orderId);
  Stream<P2POrder> subscribeToOrderUpdates(String orderId);
  Future<void> markPaymentAsSent({
    required String orderId,
    required String utrNumber,
    required String paymentProofUrl,
  });
  Future<void> releaseCryptoEscrow({
    required String orderId,
    required String biometricAuthToken,
  });
  Future<void> cancelOrder(String orderId);
  Future<void> fileDispute({
    required String orderId,
    required DisputeReason reason,
    required String description,
    required String utrNumber,
    required List<String> evidenceUrls,
  });
  Future<String> getProofUploadPresignedUrl(String orderId);
}

abstract class IP2PChatEncryptionService {
  Future<void> initializeSession({
    required String orderId,
    required String counterpartyUserId,
    required Map<String, dynamic> counterpartyPreKeyBundle,
  });
  Future<String> encryptMessage({
    required String orderId,
    required String plaintext,
  });
  Future<String> decryptMessage({
    required String orderId,
    required String ciphertext,
  });
}

abstract class IWatermarkService {
  Future<List<int>> applySecurityWatermark({
    required List<int> rawImageBytes,
    required String orderId,
    required String buyerUserId,
    required DateTime timestamp,
  });
}
```

### Presentation State Models
```dart
// apps/growww_flutter/lib/screens/p2p/state/p2p_presentation_state.dart

class P2POrderScreenState {
  final P2POrder? order;
  final bool isLoading;
  final String? errorMessage;
  final Duration remainingPaymentDuration;
  final bool isSubmittingPayment;
  final bool isReleasingEscrow;
  final int unreadChatMessagesCount;

  const P2POrderScreenState({
    this.order,
    required this.isLoading,
    this.errorMessage,
    required this.remainingPaymentDuration,
    required this.isSubmittingPayment,
    required this.isReleasingEscrow,
    required this.unreadChatMessagesCount,
  });

  P2POrderScreenState copyWith({
    P2POrder? order,
    bool? isLoading,
    String? errorMessage,
    Duration? remainingPaymentDuration,
    bool? isSubmittingPayment,
    bool? isReleasingEscrow,
    int? unreadChatMessagesCount,
  });
}

class P2PChatScreenState {
  final List<EncryptedChatMessage> messages;
  final bool isConnecting;
  final bool isSessionInitialized;
  final String? errorMessage;
  final bool isSendingMedia;

  const P2PChatScreenState({
    required this.messages,
    required this.isConnecting,
    required this.isSessionInitialized,
    this.errorMessage,
    required this.isSendingMedia,
  });
}
```

## Security & Compliance Notes
- **Strict KYC Name Invariant Enforcement (ADR-0039):** The fiat sender bank account name must match the verified user KYC record with 100% precision. The UI displays prominent warning banners on both buyer and seller views. Third-party bank transfers are explicitly prohibited; violation results in permanent account suspension and escrow forfeiture.
- **Client-Side Tamper-Resistant Screenshot Watermarking:** To prevent fraudulent recycling of payment receipts across multiple orders or external exchanges, `WatermarkProcessor` burns indelible semi-transparent text across uploaded screenshots before transmission. The watermark includes the Order ID, Buyer User ID, UTC timestamp, and statutory fraud disclaimer (`FOR GROWWW P2P ESCROW ONLY - NOT TRANSFERABLE`).
- **Anti-Phishing Security Phrase Verification:** The user custom anti-phishing phrase (configured during onboarding) is rendered permanently in the navigation app bar and encrypted chat banner. This allows the user to immediately verify that the screen and system prompts originate from the authentic Growww client and have not been hijacked by phishing proxies.
- **End-to-End Encrypted Trade Chat (Signal Double Ratchet):** All trade communications are encrypted on-device before transmission across the WebSocket tunnel using `libsignal_protocol_dart`. Intermediate WebSocket relays and database brokers have zero plaintext visibility. Ephemeral keys are wiped upon order finalization or dispute resolution.
- **Biometric Challenge for Escrow Release:** Crypto collateral release constitutes an irreversible financial event. Sellers must pass a mandatory biometric authentication challenge (`local_auth` Face ID / Fingerprint) before the client transmits the signed release payload to the backend.
- **Screen Capture Prevention (`FLAG_SECURE`):** Sensitive payment account numbers, IFSC codes, and UPI virtual payment addresses are protected on Android devices by enabling `FLAG_SECURE` in window manager parameters to block background screenshotting and screen recording malware.
- **Chat Heuristic Anti-Fraud Filters:** The client-side chat controller analyzes draft messages and inbound texts with regular expressions detecting off-platform communication solicitations (e.g. phone numbers, email patterns, "telegram", "tg", "wa", "whatsapp"). If detected, the client displays an immediate high-severity security alert advising the user that leaving the platform voids escrow protection.

## Acceptance Criteria
- [ ] `P2POrderWizardScreen` dynamically adapts between `P2PEscrowBuyerStepView` and `P2PEscrowSellerStepView` based on `P2PUserRole`.
- [ ] `P2PEscrowBuyerStepView` renders seller bank details and dynamic vector UPI QR code using `QrImageView`.
- [ ] Tapping any payment credential (account number, IFSC, UPI ID) copies the value to the clipboard and triggers haptic tactile feedback.
- [ ] 15-minute escrow countdown timer decrements once per second, updates color codes dynamically, and triggers vibration alerts at 5-minute and 2-minute milestones.
- [ ] `WatermarkedScreenshotUploaderWidget` burns indelible semi-transparent watermark text onto captured images prior to S3 upload.
- [ ] `P2PEscrowSellerStepView` renders verified buyer KYC name, masked PAN, and historical completion statistics.
- [ ] "Release Crypto" CTA is disabled until buyer marks order as paid, and prompts for biometric authentication (`local_auth`) before submitting release intent.
- [ ] `P2PEncryptedChatView` establishes an E2EE session using `libsignal_protocol_dart` and sends/receives encrypted messages over WebSockets.
- [ ] Anti-phishing security phrase appears prominently in both the order header and encrypted chat banner.
- [ ] Inbound and outbound chat text containing off-platform contact handles triggers an immediate security alert banner.
- [ ] `EscrowStatusCard` visualizes on-chain lock states (`AWAITING_COLLATERAL`, `COLLATERAL_LOCKED`, `SETTLED`) and displays block confirmations.
- [ ] Dispute escalation modal allows submitting UTR numbers and evidence attachments upon timer expiry.
- [ ] Android window parameters enable `FLAG_SECURE` to prevent unauthorized screen captures during payment inspection.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Project Scaffolding), Prompt 502 (Architecture & Riverpod State Management), Prompt 503 (Design System & Theming), Prompt 505 (Authentication UI), Prompt 521 (Local Secure Storage), Prompt 525 (API Client Layer).
- **Backend Dependencies:** Prompt 202 (KYC & AML Service), Prompt 211 (Notification Service), Prompt 269 (P2P Escrow & Dispute Resolution Service), Prompt 805 (S3 Bucket Infrastructure).
- **Downstream Blockers:** Prompt 512 (Transaction & Trade History Screen), Prompt 902 (End-to-End P2P Integration Testing).
