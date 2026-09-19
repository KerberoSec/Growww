# 543 - Flutter BTC & USDT Deposit & Withdrawal Management Modal

## Purpose
Provides a production-grade, highly secure, and responsive modal interface for depositing and withdrawing real Bitcoin (BTC) and Tether USD (USDT) across mobile (Android, iOS) and desktop (Windows, macOS, Linux) platforms within the Growww NBSE trading client. Operating under regulated GIFT City IFSC international digital asset rails and sandbox environments, capital funding and liquidation demand frictionless crypto asset transfer flows without compromising cryptographic security.

Cross-chain crypto deposits and withdrawals present critical operational hazards: irreparable token loss from network mismatches (e.g. sending ERC-20 USDT to a TRC-20 address), clipboard-hijacking malware that mutates destination strings, dusting attacks, and unconfirmed mempool congestion. This specification governs the complete modal architecture located at `lib/screens/wallet/crypto_transfer_modal.dart`, orchestrating Riverpod reactive state management, dynamic vector QR generation (`qr_flutter`), cryptographic address format validation, two-factor authentication (2FA) and local biometric authorization (`local_auth`), dynamic fee calculations, and real-time transaction status tracking linked to public block explorers.

## What You Are Building
A modular, cross-platform modal management system situated at `lib/screens/wallet/crypto_transfer_modal.dart` and its supporting widget hierarchy in `lib/screens/wallet/widgets/`:
- `CryptoTransferModal`: Root container providing a responsive bottom sheet on mobile devices and an adaptive centered modal dialog on desktop environments, with segmented tabs for Deposit and Withdrawal.
- Asset & Network Selector: Dynamic asset toggle between Bitcoin (BTC) and Tether (USDT), with chain-specific network pickers:
  - BTC Networks: Native SegWit (Bech32 - `bc1q...`) and Taproot (Bech32m - `bc1p...`).
  - USDT Networks: Tron (TRC-20), Ethereum (ERC-20), and Polygon (Polygon PoS ERC-20).
- Dynamic QR Code Generator: Integrated vector QR renderer using `qr_flutter` encoding BIP-21 URIs for Bitcoin and standard payment URIs for USDT networks, with tap-to-expand and share capability.
- Copy-to-Clipboard & Address Reveal: One-touch copy action with tactile haptic feedback, visual confirmation snackbar, address truncation formatting (`first 8 ... last 8`), and complete address toggle.
- Deposit Advisory & Minimum Threshold Banner: Prominently formatted warning cards detailing asset-specific minimum deposit requirements, required on-chain block confirmations, and token contract confirmations.
- Withdrawal Input & Validation Suite: Destination address input equipped with real-time regex and checksum validation, address book whitelist integration, maximum balance toggle, and fiat value preview.
- Dynamic Fee & Net Calculation Engine: Dynamic network fee estimation across priority tiers (Slow, Standard, Fast), platform fee deduction, and net receivable calculations.
- Two-Stage Withdrawal Confirmation Dialog: High-friction review modal explicitly surfacing destination address, network, raw network fee, and net amount with a mandatory 5-second review timer before enabling authorization.
- Biometric & 2FA Sign-Off: Secure withdrawal execution requiring device biometric verification (`local_auth`) and Time-Based One-Time Password (TOTP) verification.
- In-Modal Transaction History Tracker: Embedded real-time transaction ledger displaying pending mempool transfers, live block confirmation counters, and deep links to public block explorers.

## Scope Boundaries
- **In Scope:**
  - Responsive Flutter UI modal (`lib/screens/wallet/crypto_transfer_modal.dart`) and sub-components supporting mobile and desktop form factors.
  - Dedicated flows for Bitcoin (Native SegWit, Taproot) and USDT (TRC-20, ERC-20, Polygon).
  - Dynamic QR code generation via `qr_flutter` matching BIP-21 and EVM/Tron URI conventions.
  - Client-side cryptographic address format verification and anti-phishing validation.
  - Biometric sign-off (`local_auth`) and TOTP verification flows for withdrawal requests.
  - Dynamic fee display, minimum limit warnings, and balance deduction logic.
  - Transaction state visualization with real-time block confirmation counts and block explorer launching (`url_launcher`).
  - Riverpod state providers, domain models, and API repository interfaces.
- **Out of Scope / Handled Elsewhere:**
  - Backend blockchain listener and indexing services: BTC Deposit Listener (Prompt 277), USDT Deposit Service (Prompt 278).
  - Backend withdrawal execution, MPC signing, and fee bumping: Withdrawal Engine (Prompt 279), Multi-Chain MPC-TSS Vault (Prompt 237).
  - Main wallet funds overview screen and fiat bank deposit rails: Wallet Funds Screen (Prompt 511), Payment Gateway Service (Prompt 212).
  - Full-screen multi-asset transaction reporting: Transaction History Screen (Prompt 512).
  - Core authentication, MPIN, and session lifecycle management: Authentication UI (Prompt 505).

## Technology to Use
- **Framework & Language:** Flutter 3.22+ and Dart 3.4+ targeting Android, iOS, Windows, macOS, and Linux.
- **State Management:** `flutter_riverpod: ^2.5.1` (`AutoDisposeAsyncNotifier` and immutable state models).
- **QR Code Rendering:** `qr_flutter: ^4.1.0` (high-fidelity vector QR rendering with error correction level M).
- **Biometric Security:** `local_auth: ^2.2.0` (Face ID, Touch ID, BiometricPrompt, and Windows Hello integration).
- **Cryptography & Hashing:** `crypto: ^3.0.3` (SHA-256 and Keccak-256 checksum validations).
- **OS Integration & Feedback:**
  - `flutter/services.dart` (`Clipboard` and `HapticFeedback`).
  - `share_plus: ^9.0.0` (native operating system share sheet integration).
  - `url_launcher: ^6.3.0` (deep-linking to external block explorers).
- **HTTP & Networking:** `dio: ^5.4.3+1` (typed REST API client) and `web_socket_channel: ^3.0.0` (real-time transaction confirmation push).

## Backend / Infra Touchpoints
- **BTC Deposit Listener (Prompt 277):**
  - `GET /api/v1/crypto/btc/deposit-address`: Fetches or derives user-dedicated Native SegWit (`bc1q`) or Taproot (`bc1p`) deposit address.
  - `WS /ws/v1/crypto/btc/deposits`: WebSocket channel streaming live unconfirmed mempool transactions and block confirmation increments for Bitcoin.
- **USDT Deposit Service (Prompt 278):**
  - `GET /api/v1/crypto/usdt/deposit-address?network={network}`: Retrieves network-specific deposit address for TRC-20, ERC-20, or Polygon.
  - `WS /ws/v1/crypto/usdt/deposits`: WebSocket channel streaming inbound USDT transfer events and finality confirmations.
- **Withdrawal Engine (Prompt 279):**
  - `POST /api/v1/crypto/withdraw/estimate-fee`: Requests current network fee, gas price, and recommended priority tiers.
  - `POST /api/v1/crypto/withdraw/submit`: Submits signed withdrawal intent with unique idempotency key, TOTP code, and biometric token.
  - `GET /api/v1/crypto/withdraw/status/{withdrawal_id}`: Polls real-time withdrawal processing status.
- **Wallet Account Service (Prompt 203):**
  - `GET /api/v1/wallet/balances`: Fetches real-time available and locked balances for BTC and USDT.
- **Compliance & Sanctions Screening (Prompt 202):**
  - `POST /api/v1/compliance/screen-address`: Validates destination address against AML sanctions lists prior to withdrawal submission.

## Blockchain Interaction
- **Supported Networks and Address Standards:**
  - **Bitcoin (BTC):**
    - Native SegWit: Bech32 format starting with `bc1q`, 42 characters, lowercase alphanumeric. Minimum 2 block confirmations for credit.
    - Taproot: Bech32m format starting with `bc1p`, 62 characters, lowercase alphanumeric. Minimum 2 block confirmations for credit.
  - **Tether (USDT):**
    - TRC-20 (Tron): Base58Check format starting with `T`, 34 characters. Requires 19-20 block confirmations.
    - ERC-20 (Ethereum): EIP-55 mixed-case hexadecimal format starting with `0x`, 42 characters. Requires 12 block confirmations (post-Merge beacon chain finality).
    - Polygon (Polygon PoS): EIP-55 mixed-case hexadecimal format starting with `0x`, 42 characters. Requires 64 block confirmations.
- **QR Code URI Payloads:**
  - BTC: `bitcoin:{address}?amount={optional_amount}&label=Growww%20Deposit`
  - USDT (TRC-20): `tron:{address}?token=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t&amount={optional_amount}`
  - USDT (ERC-20): `ethereum:0xdac17f958d2ee523a2206206994597c13d831ec7/transfer?address={address}&uint256={optional_amount_in_units}`
  - USDT (Polygon): `ethereum:0xc2132d05d31c914a87c6611c10748aeb04b58e8f/transfer?address={address}&uint256={optional_amount_in_units}`
- **Block Explorer Visualizers:**
  - Bitcoin: `https://mempool.space/tx/{tx_hash}` (Fallback: `https://blockstream.info/tx/{tx_hash}`)
  - USDT TRC-20: `https://tronscan.org/#/transaction/{tx_hash}`
  - USDT ERC-20: `https://etherscan.io/tx/{tx_hash}`
  - USDT Polygon: `https://polygonscan.com/tx/{tx_hash}`

## Step-by-Step Build Instructions
1. **Scaffold Directory Structure:** Create feature folders under `lib/screens/wallet/` including `crypto_transfer_modal.dart`, `widgets/deposit/`, `widgets/withdraw/`, `widgets/history/`, `models/`, and `controllers/`.
2. **Define Dart Models and Enums:** Implement immutable domain models in `lib/screens/wallet/models/`: `DepositAddress`, `WithdrawalRequest`, `TransactionStatus`, `CryptoFeeEstimate`, and relevant enums (`CryptoAssetType`, `CryptoTransferNetwork`, `TransferDirection`, `TransferStatusStage`, `FeePriorityLevel`).
3. **Build Address Format and Anti-Phishing Validator:** Create `CryptoAddressValidator` implementing:
   - Bech32 and Bech32m checksum decoding for Bitcoin `bc1q` and `bc1p` strings.
   - Base58Check decoding and byte length checks for Tron `T...` addresses.
   - EIP-55 Keccak-256 mixed-case checksum verification for Ethereum and Polygon `0x...` addresses.
   - Incompatible network detection (e.g. flagging an `0x` address pasted into a TRC-20 input).
4. **Implement URI Builder Utility:** Build `CryptoPaymentUriBuilder` that constructs standardized BIP-21, Tron, and EIP-681 URI strings with proper URL encoding for QR display.
5. **Implement Riverpod Notifiers:**
   - `CryptoDepositNotifier`: Fetches active deposit address from Prompt 277 or Prompt 278 based on selected asset and network, handling loading and error states.
   - `CryptoWithdrawalNotifier`: Manages recipient address input, fee tier selection, balance checks, pre-flight AML screening, and withdrawal dispatch to Prompt 279.
   - `CryptoTransferHistoryNotifier`: Manages a real-time list of recent transfers, updating block confirmation counts via WebSocket connection.
6. **Construct Responsive Modal Root Container:** In `crypto_transfer_modal.dart`, implement `CryptoTransferModal` with a responsive adaptive layout:
   - On mobile (`MediaQuery.of(context).size.width < 600`): Render as a modal bottom sheet with drag handle, constrained to 90% screen height.
   - On desktop: Render as a floating dialog with fixed width (540px) and max height (780px).
   - Top navigation: Asset switcher tabs (BTC / USDT) and direction tabs (Deposit / Withdraw).
7. **Develop Deposit Tab View (`CryptoDepositTabView`):**
   - Network selection chips or dropdown (BTC: SegWit vs Taproot; USDT: TRC-20 vs ERC-20 vs Polygon).
   - Display warning pill showing the minimum deposit threshold (e.g. 0.0005 BTC, 10 USDT) and confirmation count requirements.
   - Dynamic QR code card with `QrImageView`, rounded border, and center asset logo overlay.
   - Display full address with copy button, share button, and clipboard feedback.
8. **Develop Withdrawal Tab View (`CryptoWithdrawalTabView`):**
   - Network selector matching available withdrawal chains.
   - Destination address text field with integrated paste button, QR scanner trigger, and real-time checksum validation banner.
   - Withdrawal amount input field with percentage presets (25%, 50%, 75%, 100%), available balance indicator, and fiat conversion.
   - Network fee tier selector (Slow, Standard, Fast) displaying estimated cost in crypto and fiat.
   - Summary breakdown displaying: Amount Entered, Network Fee, Platform Fee, and Net Amount Deducted.
9. **Build Mandatory Confirmation Dialog (`WithdrawalConfirmationDialog`):**
   - Render a high-friction modal displaying destination address (split into highlighted first 8 and last 8 characters), target network, fee, and net payout.
   - Enforce a 5-second unskippable countdown on the "Confirm & Authorize" button to prevent hasty approvals.
10. **Implement Biometric and 2FA Sign-Off Flow:**
    - Invoke `local_auth` (`authenticate` with `biometricOnly: true`) when user clicks confirm.
    - If biometric succeeds, present a 6-digit TOTP input modal (or pass cached session token if biometric challenge proves possession).
    - Dispatch signed payload to `/api/v1/crypto/withdraw/submit` with UUIDv4 idempotency key.
11. **Develop In-Modal Transaction History Section (`TransferHistorySheetWidget`):**
    - Show the last 5 transactions for the active asset and network.
    - Render a linear progress indicator displaying live confirmations against target threshold (e.g. "1/2 Blocks" for BTC, "8/12 Blocks" for ERC-20).
    - Tap to open transaction hash in system browser via `url_launcher`.
12. **Implement Clipboard Hijacking Guard:**
    - When user taps "Paste", inspect clipboard content and compare checksum against network rules.
    - If address looks valid but does not match recent user clipboard history, display a brief visual alert: "Please verify destination address characters carefully".
13. **Unit and Widget Testing:**
    - Write unit tests for `CryptoAddressValidator` covering valid/invalid Bech32, Bech32m, Base58Check, and EIP-55 addresses.
    - Write widget tests for `CryptoTransferModal` verifying tab switching, QR code rendering, error banner appearances, and withdrawal validation button states.

## Interfaces / Contracts

### Domain Models
```dart
// lib/screens/wallet/models/crypto_transfer_models.dart

enum CryptoAssetType {
  btc,
  usdt,
}

enum CryptoTransferNetwork {
  btcNativeSegwit, // Bitcoin Bech32 (P2WPKH)
  btcTaproot,      // Bitcoin Bech32m (P2TR)
  usdtTrc20,       // Tron TRC-20
  usdtErc20,       // Ethereum ERC-20
  usdtPolygon,     // Polygon PoS
}

enum TransferDirection {
  deposit,
  withdrawal,
}

enum TransferStatusStage {
  pending,
  mempoolDetected,
  confirming,
  finalized,
  failed,
  rejected,
}

enum FeePriorityLevel {
  economic,
  standard,
  priority,
}

class DepositAddress {
  final String address;
  final CryptoAssetType asset;
  final CryptoTransferNetwork network;
  final String qrUriPayload;
  final double minimumDepositAmount;
  final int requiredConfirmations;
  final DateTime generatedAt;
  final DateTime? expiresAt;

  const DepositAddress({
    required this.address,
    required this.asset,
    required this.network,
    required this.qrUriPayload,
    required this.minimumDepositAmount,
    required this.requiredConfirmations,
    required this.generatedAt,
    this.expiresAt,
  });

  factory DepositAddress.fromJson(Map<String, dynamic> json) {
    return DepositAddress(
      address: json['address'] as String,
      asset: CryptoAssetType.values.byName(json['asset'] as String),
      network: CryptoTransferNetwork.values.byName(json['network'] as String),
      qrUriPayload: json['qr_uri_payload'] as String,
      minimumDepositAmount: (json['minimum_deposit_amount'] as num).toDouble(),
      requiredConfirmations: json['required_confirmations'] as int,
      generatedAt: DateTime.parse(json['generated_at'] as String),
      expiresAt: json['expires_at'] != null ? DateTime.parse(json['expires_at'] as String) : null,
    );
  }
}

class WithdrawalRequest {
  final String requestId;
  final String idempotencyKey;
  final CryptoAssetType asset;
  final CryptoTransferNetwork network;
  final String destinationAddress;
  final double amount;
  final double networkFee;
  final double platformFee;
  final double netAmount;
  final String totpCode;
  final String biometricAuthToken;
  final DateTime timestamp;

  const WithdrawalRequest({
    required this.requestId,
    required this.idempotencyKey,
    required this.asset,
    required this.network,
    required this.destinationAddress,
    required this.amount,
    required this.networkFee,
    required this.platformFee,
    required this.netAmount,
    required this.totpCode,
    required this.biometricAuthToken,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() {
    return {
      'request_id': requestId,
      'idempotency_key': idempotencyKey,
      'asset': asset.name,
      'network': network.name,
      'destination_address': destinationAddress,
      'amount': amount,
      'network_fee': networkFee,
      'platform_fee': platformFee,
      'net_amount': netAmount,
      'totp_code': totpCode,
      'biometric_auth_token': biometricAuthToken,
      'timestamp': timestamp.toIso8601String(),
    };
  }
}

class TransactionStatus {
  final String transactionId;
  final String txHash;
  final CryptoAssetType asset;
  final CryptoTransferNetwork network;
  final TransferDirection direction;
  final double amount;
  final TransferStatusStage stage;
  final int currentConfirmations;
  final int requiredConfirmations;
  final String? blockExplorerUrl;
  final DateTime createdAt;
  final DateTime? updatedAt;
  final String? failureReason;

  const TransactionStatus({
    required this.transactionId,
    required this.txHash,
    required this.asset,
    required this.network,
    required this.direction,
    required this.amount,
    required this.stage,
    required this.currentConfirmations,
    required this.requiredConfirmations,
    this.blockExplorerUrl,
    required this.createdAt,
    this.updatedAt,
    this.failureReason,
  });

  factory TransactionStatus.fromJson(Map<String, dynamic> json) {
    return TransactionStatus(
      transactionId: json['transaction_id'] as String,
      txHash: json['tx_hash'] as String,
      asset: CryptoAssetType.values.byName(json['asset'] as String),
      network: CryptoTransferNetwork.values.byName(json['network'] as String),
      direction: TransferDirection.values.byName(json['direction'] as String),
      amount: (json['amount'] as num).toDouble(),
      stage: TransferStatusStage.values.byName(json['stage'] as String),
      currentConfirmations: json['current_confirmations'] as int,
      requiredConfirmations: json['required_confirmations'] as int,
      blockExplorerUrl: json['block_explorer_url'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: json['updated_at'] != null ? DateTime.parse(json['updated_at'] as String) : null,
      failureReason: json['failure_reason'] as String?,
    );
  }
}

class CryptoFeeEstimate {
  final CryptoAssetType asset;
  final CryptoTransferNetwork network;
  final FeePriorityLevel priority;
  final double estimatedFeeCrypto;
  final double estimatedFeeFiat;
  final Duration estimatedConfirmationTime;

  const CryptoFeeEstimate({
    required this.asset,
    required this.network,
    required this.priority,
    required this.estimatedFeeCrypto,
    required this.estimatedFeeFiat,
    required this.estimatedConfirmationTime,
  });
}
```

### Service Contracts
```dart
// lib/screens/wallet/controllers/crypto_transfer_contracts.dart

abstract class ICryptoTransferRepository {
  Future<DepositAddress> getDepositAddress({
    required CryptoAssetType asset,
    required CryptoTransferNetwork network,
  });

  Future<CryptoFeeEstimate> estimateWithdrawalFee({
    required CryptoAssetType asset,
    required CryptoTransferNetwork network,
    required FeePriorityLevel priority,
    required double amount,
  });

  Future<TransactionStatus> submitWithdrawal({
    required WithdrawalRequest request,
  });

  Future<List<TransactionStatus>> getRecentTransfers({
    required CryptoAssetType asset,
    int limit = 10,
  });

  Stream<TransactionStatus> subscribeTransactionUpdates({
    required String accountId,
  });
}

abstract class ICryptoAddressValidator {
  bool isValidAddress({
    required String address,
    required CryptoTransferNetwork network,
  });

  String? getNetworkMismatchWarning({
    required String address,
    required CryptoTransferNetwork selectedNetwork,
  });
}

abstract class IBiometricAuthService {
  Future<bool> isBiometricAvailable();
  Future<String?> authenticateForWithdrawal({
    required String promptMessage,
  });
}
```

### State Models
```dart
// lib/screens/wallet/models/modal_state_models.dart

class DepositModalState {
  final CryptoAssetType selectedAsset;
  final CryptoTransferNetwork selectedNetwork;
  final DepositAddress? depositAddress;
  final bool isLoading;
  final String? errorMessage;
  final bool isAddressRevealed;

  const DepositModalState({
    required this.selectedAsset,
    required this.selectedNetwork,
    this.depositAddress,
    required this.isLoading,
    this.errorMessage,
    this.isAddressRevealed = false,
  });
}

class WithdrawalModalState {
  final CryptoAssetType selectedAsset;
  final CryptoTransferNetwork selectedNetwork;
  final String destinationAddress;
  final bool isAddressValid;
  final String? addressValidationError;
  final double amount;
  final double availableBalance;
  final FeePriorityLevel feePriority;
  final CryptoFeeEstimate? feeEstimate;
  final bool isAmlScreening;
  final bool isSubmitting;
  final String? submissionError;

  const WithdrawalModalState({
    required this.selectedAsset,
    required this.selectedNetwork,
    required this.destinationAddress,
    required this.isAddressValid,
    this.addressValidationError,
    required this.amount,
    required this.availableBalance,
    required this.feePriority,
    this.feeEstimate,
    required this.isAmlScreening,
    required this.isSubmitting,
    this.submissionError,
  });
}
```

## Security & Compliance Notes
- **Strict Anti-Phishing Address Validation:** Real-time client verification prevents users from submitting invalid formats. Addresses are checked against strict checksum standards: Bech32/Bech32m for Bitcoin, Base58Check for Tron, and EIP-55 for Ethereum/Polygon. If a user pastes an EVM address (`0x...`) while TRC-20 or Bitcoin is selected, a prominent error prevents form submission.
- **Clipboard Poisoning & Address Hijacking Defense:** Client software inspects clipboard paste events. The modal displays a split-address verification banner prompting the user to visually confirm the first 8 and last 8 characters of the destination address before submitting.
- **Mandatory Dual-Stage Confirmation Dialog:** Withdrawals cannot be submitted in a single click. A secondary confirmation sheet displays the destination address, network name, network fee, and net receivable amount with an unskippable 5-second countdown timer.
- **Biometric Sign-Off Enforcement:** All withdrawal submissions require hardware-backed biometric authentication (`local_auth`) on devices with Face ID, Touch ID, or biometric capabilities. The local authentication token is signed and bundled with the withdrawal request.
- **Secondary Factor (TOTP) Authorization:** In addition to biometrics, users must provide a valid 6-digit TOTP code generated by their authenticator app.
- **Pre-Flight AML Sanctions Screening:** The client triggers an asynchronous check against `/api/v1/compliance/screen-address` upon destination address entry, ensuring the address is not flagged on OFAC or international sanctions lists before allowing the user to initiate 2FA.
- **Idempotency Protection:** Every withdrawal submission creates a unique UUIDv4 idempotency key to prevent double-spending or duplicate debit requests in unstable network conditions.
- **Zero Local Private Keys:** The Flutter client acts purely as a presentation and authorization gateway. Private keys and signing logic reside entirely within backend MPC-TSS vaults (Prompt 237).

## Acceptance Criteria
- [ ] `CryptoTransferModal` renders correctly as a modal bottom sheet on mobile screens and as an adaptive centered dialog on desktop.
- [ ] Asset switcher allows seamless toggling between BTC and USDT, updating network selector options dynamically.
- [ ] Network selector supports Bitcoin Native SegWit (`bc1q`) and Taproot (`bc1p`) for BTC.
- [ ] Network selector supports Tron (TRC-20), Ethereum (ERC-20), and Polygon (Polygon PoS) for USDT.
- [ ] `CryptoDepositTabView` displays crisp vector QR code via `qr_flutter` with accurate BIP-21 or EVM/Tron URI payload.
- [ ] Tapping "Copy Address" copies the exact string to the system clipboard, triggers haptic feedback, and shows a confirmation toast.
- [ ] Prominent advisory banner indicates the minimum deposit threshold and required block confirmations for the selected network.
- [ ] `CryptoWithdrawalTabView` validates destination address format in real-time, blocking submission and showing red warning text if invalid or mismatched.
- [ ] Percentage balance buttons (25%, 50%, 75%, 100%) correctly populate the withdrawal amount based on available wallet balance minus fees.
- [ ] Dynamic fee estimator displays real-time network fees across Economic, Standard, and Priority tiers with fiat conversions.
- [ ] Secondary confirmation dialog enforces a 5-second review countdown and highlights first 8 and last 8 characters of the destination address.
- [ ] Biometric prompt (`local_auth`) successfully triggers and blocks withdrawal if authentication fails or is cancelled.
- [ ] Valid TOTP 6-digit entry is enforced before final withdrawal submission to the backend API.
- [ ] Embedded transaction history tracker streams live confirmation counts via WebSocket (e.g. 1/2 for BTC, 7/12 for ERC-20).
- [ ] Clicking transaction hash launches the correct external block explorer (Mempool.space, Tronscan, Etherscan, Polygonscan) in the system browser via `url_launcher`.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 501: Flutter Multi-Platform Project Scaffolding
  - Prompt 502: Flutter App Architecture & State Management (Riverpod)
  - Prompt 503: Flutter Design System & Theming
  - Prompt 505: Flutter Authentication UI (Biometrics, MPIN, OTP, Session)
  - Prompt 521: Local Secure Storage for Credentials
  - Prompt 525: Flutter API Client Layer (Typed Dio)
- **Backend Touchpoints & Service Dependencies:**
  - Prompt 203: Wallet & Double-Entry Account Ledger Service
  - Prompt 277: Bitcoin (BTC) Deposit Listener & Address Derivation Service
  - Prompt 278: Multi-Chain USDT Deposit & Processing Service
  - Prompt 279: High-Security Multi-Chain Crypto Withdrawal Engine
  - Prompt 202: KYC & Sanctions Screening Service
- **Downstream Blockers & Integrations:**
  - Prompt 511: Flutter Wallet & Funds Screen (Modal launcher integration)
  - Prompt 512: Flutter Transaction & Trade History Screen
  - Prompt 902: End-to-End Integration Testing Suite
