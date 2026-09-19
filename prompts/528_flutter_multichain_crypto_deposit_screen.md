# 528 - Flutter Multi-Chain Crypto Deposit and Withdrawal Screen (BTC, ETH, SOL, POL, ARB)

## Purpose
Provides an institutional-grade, multi-chain digital asset deposit and withdrawal interface for the Growww multi-platform client (Android, iOS, Windows, macOS, Linux). Operating within regulatory sandboxes, GIFT City IFSC international investment channels, and cross-border tokenized asset settlement rails, investors require a secure, multi-network asset transfer portal supporting major distributed ledgers: Bitcoin (BTC), Ethereum (ETH and ERC-20 tokens), Solana (SOL and SPL tokens), Polygon (POL PoS), and Arbitrum One (ARB Nitro L2).

Cross-chain digital asset operations introduce severe financial risks, including irreversible token loss from network mismatch, address truncation, dusting attacks, and dynamic gas volatility. This specification defines the presentation architecture, Riverpod reactive state management, dynamic vector QR generation (supporting BIP-21, EIP-681, and Solana Pay URI standards), client-side cryptographic address checksum verification (Bech32/Bech32m, EIP-55, Base58Check), dynamic network fee estimators with priority tiers, biometric withdrawal authorization, and real-time block confirmation status trackers streaming over WebSockets.

## What You Are Building
A modular, high-security crypto asset transfer suite located in `apps/growww_flutter/lib/features/crypto_wallet/`:
- `CryptoDepositScreen`: Multi-chain asset and network selector with dynamic address generation, dynamic vector QR code display, one-tap clipboard copy with haptic confirmation, native OS share sheet dispatch, and chain-specific deposit warnings (minimum deposit thresholds, network confirmation requirements).
- `CryptoWithdrawalScreen`: Destination address input with real-time cryptographic checksum validation, saved address whitelist selector, transfer amount input with fiat equivalent display, real-time gas fee estimator, biometric authorization modal, and dual-factor (TOTP/SMS) confirmation.
- `AddressVerificationWidget`: Visual cryptographic checksum indicator validating EIP-55 mixed-case Ethereum/Polygon/Arbitrum addresses, Bech32/Bech32m Bitcoin addresses, and Base58Check Solana addresses, displaying explicit network mismatch warnings if an investor pastes an incompatible address.
- `DynamicFeeEstimatorWidget`: Multi-tier network fee selector (Slow, Standard, Fast) displaying estimated gas units, network base and priority fees, platform transfer fees, estimated time to inclusion, and real-time USD/INR fiat equivalents.
- `TransactionConfirmationTrackerWidget`: Real-time progress tracker monitoring unconfirmed mempool transactions, streaming block confirmation counts against chain finality thresholds (e.g. 3 blocks for BTC, 12 blocks for ETH, 32 blocks for SOL, 64 blocks for POL, L1 batch post for ARB), and providing deep links to public block explorers.
- `Riverpod State Layer`: Immutable state models, asynchronous notifiers (`CryptoDepositNotifier`, `CryptoWithdrawalNotifier`, `GasFeeEstimatorNotifier`, `CryptoTransferHistoryNotifier`), and WebSocket stream providers delivering live deposit detection and confirmation updates.

## Scope Boundaries
- **In Scope:**
  - Multi-platform Flutter UI screens and reusable widgets for crypto deposits, withdrawals, and address verification.
  - Multi-chain support for Bitcoin (BTC), Ethereum (ETH, USDT, USDC), Solana (SOL, USDT, USDC), Polygon (POL, USDT, USDC), and Arbitrum One (ARB, ETH, USDT, USDC).
  - Riverpod state management and domain contracts for deposit address retrieval, fee estimation, withdrawal validation, and live confirmation tracking.
  - Dynamic QR code generation with URI standard encoding (BIP-21, EIP-681, Solana Pay).
  - Client-side address format and checksum verification algorithms.
  - Real-time block confirmation tracking UI with WebSocket status streaming.
  - Biometric authentication challenge integration via `local_auth`.
- **Out of Scope / Handled Elsewhere:**
  - Backend multi-party computation (MPC) key generation and hot/cold wallet custody infrastructure (Prompt 213, Prompt 311).
  - On-chain Delivery-versus-Payment smart contract execution (Prompt 306).
  - Backend blockchain indexer nodes and RPC node infrastructure (Prompt 302, Prompt 309).
  - AML on-chain analytics screening service (Prompt 202, Prompt 703).
  - Fiat bank deposits and UPI payment gateways (Prompt 511).

## Technology to Use
- **Primary Language & Framework:** Flutter 3.22+ with Dart 3.4+ using `flutter_riverpod` (v2.5+) for reactive state management.
- **Justification:** Riverpod provides deterministic asynchronous state handling (`AsyncValue`), modular dependency injection, and clean separation between UI presentation and transfer business logic. Vector QR generation renders crisply across mobile and 4K desktop screens without pixelation.
- **Dependencies:**
  - `flutter_riverpod: ^2.5.1`
  - `qr_flutter: ^4.1.0` (dynamic vector QR code rendering)
  - `crypto: ^3.0.3` (SHA-256 and Keccak hashing for address checksum calculation)
  - `local_auth: ^2.2.0` (biometric authorization for withdrawal dispatch)
  - `share_plus: ^9.0.0` (native OS share dialog for deposit addresses)
  - `url_launcher: ^6.3.0` (block explorer deep-linking)
  - `dio: ^5.4.3+1` (typed REST API communication)
  - `web_socket_channel: ^3.0.0` (real-time transaction confirmation streaming)

## Backend / Infra Touchpoints
- **Custody & Wallet API (`/api/v1/crypto/deposit-address`):** Retrieves or derives chain-specific deposit addresses and public keys.
- **Fee Estimation Service (`/api/v1/crypto/fee-estimate`):** Fetches real-time base fees, priority fees, and gas limits per chain.
- **Withdrawal Dispatch Endpoint (`/api/v1/crypto/withdraw/submit`):** Submits signed withdrawal intents with idempotency keys and biometric authorization tokens.
- **Address Whitelist Service (`/api/v1/crypto/whitelisted-addresses`):** Fetches pre-approved withdrawal destinations and cooling period statuses.
- **Sanctions & Compliance Screening Service (`/api/v1/compliance/screen-address`):** Pre-validates destination addresses against OFAC and AML blocklists before submission.
- **Real-Time Transaction WebSocket (`/ws/v1/crypto/tx-stream`):** Streams real-time mempool detection, block confirmations, and finality events.

## Blockchain Interaction
- **Supported Chains & Finality Standards:**
  - **Bitcoin (BTC):** Native SegWit (`bc1q...`), Taproot (`bc1p...`), and Legacy (`1...`, `3...`) address support. Requires 3 block confirmations for deposit crediting.
  - **Ethereum (ETH / ERC-20):** EIP-55 mixed-case hex addresses (`0x...`). Requires 12 block confirmations (post-Merge beacon chain finality).
  - **Solana (SOL / SPL):** Base58 32-to-44 character public keys. Requires `finalized` commitment level (32 confirmed slots, approx. 12 seconds).
  - **Polygon (POL / ERC-20):** EIP-55 hex addresses (`0x...`). Requires 64 block confirmations for checkpoint finality on Ethereum L1.
  - **Arbitrum One (ARB / ERC-20):** EIP-55 hex addresses (`0x...`). Displays instant Nitro sequencer confirmation followed by L1 batch rollup posting confirmation.
- **QR Code URI Standards:**
  - Bitcoin: `bitcoin:<address>?amount=<value>&label=Growww%20Deposit` (BIP-21).
  - Ethereum, Polygon, Arbitrum: `ethereum:<contract_or_recipient>@<chain_id>?value=<wei_or_tokens>` (EIP-681).
  - Solana: `solana:<address>?amount=<value>&label=Growww%20Deposit` (Solana Pay standard).
- **Block Explorer Deep-Links:**
  - Bitcoin: `https://mempool.space/tx/{txid}`
  - Ethereum: `https://etherscan.io/tx/{txhash}`
  - Solana: `https://solscan.io/tx/{signature}`
  - Polygon: `https://polygonscan.com/tx/{txhash}`
  - Arbitrum: `https://arbiscan.io/tx/{txhash}`

## Step-by-Step Build Instructions
1. Scaffold feature directories in `apps/growww_flutter/lib/features/crypto_wallet/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `domain/interfaces/`, `data/repositories/`, and `data/datasources/`.
2. Define domain models in `domain/models/`: `CryptoAsset`, `CryptoNetwork`, `DepositAddress`, `WithdrawalIntent`, `FeeEstimate`, `WhitelistedAddress`, `CryptoTransaction`, and `ConfirmationStatus`.
3. Implement client-side address verification utility `AddressValidator` supporting:
   - EIP-55 checksum calculation for EVM chains (Ethereum, Polygon, Arbitrum).
   - Bech32 and Bech32m checksum validation for Bitcoin SegWit and Taproot addresses.
   - Base58Check decoding and public key length validation for Solana addresses.
4. Implement URI generator utility `CryptoUriBuilder` creating standardized BIP-21, EIP-681, and Solana Pay URI strings with optional pre-filled deposit amounts.
5. Create `CryptoDepositNotifier` (Riverpod `AutoDisposeAsyncNotifier<DepositAddressState>`):
   - Fetches chain-specific deposit address for selected asset and network.
   - Emits QR payload, formatted address string, network warnings, and minimum deposit limits.
6. Build `CryptoDepositScreen`:
   - Network selector dropdown with visual chain badges (Bitcoin, Ethereum, Solana, Polygon, Arbitrum).
   - Asset selector tab bar (native coin vs supported stablecoins/tokens).
   - Dynamic vector QR card using `QrImageView` with embedded brand watermark.
   - One-tap address copy button with clipboard write, haptic feedback, and toast notification.
   - Native OS share sheet action button using `share_plus`.
   - Dynamic network advisory box highlighting minimum deposit threshold and required block confirmations.
7. Implement `GasFeeEstimatorNotifier` (Riverpod `AutoDisposeAsyncNotifier<FeeEstimateState>`):
   - Periodically polls `/api/v1/crypto/fee-estimate` for selected asset and network.
   - Computes gas units, base fee, max priority fee, total network fee in crypto, and fiat conversion in INR/USD.
8. Build `CryptoWithdrawalScreen`:
   - Recipient address input field with embedded QR scanner trigger and live `AddressVerificationWidget`.
   - Whitelist address book selector sheet with 24-hour cooling period indicators.
   - Amount input field with max-balance button, available balance badge, and fiat conversion preview.
   - Priority fee selector widget (Slow, Standard, Fast) displaying estimated confirmation times and fee totals.
   - Withdrawal summary card showing net receivable amount after network and platform fee deduction.
9. Build `WithdrawalAuthBottomSheet`:
   - Displays transaction destination, asset, amount, fee, and net payout.
   - Triggers biometric authentication challenge via `local_auth`.
   - Prompts for TOTP 2FA code if required by risk engine tier.
   - Dispatches authenticated withdrawal intent with unique idempotency key.
10. Implement `TransactionConfirmationTrackerNotifier` (Riverpod `AutoDisposeAsyncNotifier<List<CryptoTransaction>>`):
    - Establishes WebSocket subscription to `/ws/v1/crypto/tx-stream` filtered by user account.
    - Updates transaction confirmation counts in real time as new blocks are minted.
    - Transitions transaction state from `MEMPOOL_DETECTED` -> `CONFIRMING` -> `FINALIZED` or `FAILED`.
11. Build `TransactionConfirmationTrackerWidget` and `CryptoTransactionHistoryScreen`:
    - Linear progress bar showing active confirmations vs required threshold (e.g. "8/12 Confirmations").
    - Status badge with color-coded states (Pending, Confirming, Completed, Rejected).
    - One-tap link to launch external block explorer in browser via `url_launcher`.
12. Build `ClipboardSanitizer` listener warning the user if the clipboard contents change between copying and pasting, mitigating address-replacement malware.
13. Write unit tests for `AddressValidator` verifying valid and invalid address strings across all 5 networks.
14. Write widget tests for `CryptoDepositScreen`, `CryptoWithdrawalScreen`, and `TransactionConfirmationTrackerWidget` validating UI rendering and state transitions.

## Interfaces / Contracts

### Domain Enums and Identifiers
```dart
// lib/features/crypto_wallet/domain/models/crypto_enums.dart

enum BlockchainNetwork {
  bitcoin,
  ethereum,
  solana,
  polygon,
  arbitrum,
}

enum AddressFormat {
  bech32,
  bech32m,
  legacyP2pkh,
  legacyP2sh,
  eip55Hex,
  base58Solana,
}

enum FeePriorityTier {
  slow,
  standard,
  fast,
}

enum TransferType {
  deposit,
  withdrawal,
}

enum ConfirmationState {
  mempoolDetected,
  confirming,
  finalized,
  failed,
  cancelled,
}
```

### Core Domain Models
```dart
// lib/features/crypto_wallet/domain/models/crypto_models.dart

class CryptoAsset {
  final String symbol;
  final String name;
  final String iconUrl;
  final int decimals;
  final List<BlockchainNetwork> supportedNetworks;

  const CryptoAsset({
    required this.symbol,
    required this.name,
    required this.iconUrl,
    required this.decimals,
    required this.supportedNetworks,
  });
}

class CryptoNetworkConfig {
  final BlockchainNetwork network;
  final String displayName;
  final String nativeSymbol;
  final int requiredConfirmations;
  final String explorerTxUrlPrefix;
  final String explorerAddressUrlPrefix;
  final double minimumDepositAmount;
  final double minimumWithdrawalAmount;

  const CryptoNetworkConfig({
    required this.network,
    required this.displayName,
    required this.nativeSymbol,
    required this.requiredConfirmations,
    required this.explorerTxUrlPrefix,
    required this.explorerAddressUrlPrefix,
    required this.minimumDepositAmount,
    required this.minimumWithdrawalAmount,
  });
}

class DepositAddress {
  final String assetSymbol;
  final BlockchainNetwork network;
  final String address;
  final String? memoOrTag;
  final String qrUriPayload;
  final DateTime generatedAt;
  final DateTime? expiresAt;

  const DepositAddress({
    required this.assetSymbol,
    required this.network,
    required this.address,
    this.memoOrTag,
    required this.qrUriPayload,
    required this.generatedAt,
    this.expiresAt,
  });
}

class NetworkFeeOption {
  final FeePriorityTier tier;
  final double estimatedFeeNative;
  final double estimatedFeeFiatInr;
  final Duration estimatedConfirmationTime;
  final BigInt? gasLimit;
  final BigInt? maxFeePerGasGwei;
  final BigInt? maxPriorityFeePerGasGwei;
  final int? satoshisPerByte;

  const NetworkFeeOption({
    required this.tier,
    required this.estimatedFeeNative,
    required this.estimatedFeeFiatInr,
    required this.estimatedConfirmationTime,
    this.gasLimit,
    this.maxFeePerGasGwei,
    this.maxPriorityFeePerGasGwei,
    this.satoshisPerByte,
  });
}

class FeeEstimate {
  final BlockchainNetwork network;
  final String assetSymbol;
  final NetworkFeeOption slow;
  final NetworkFeeOption standard;
  final NetworkFeeOption fast;
  final double platformFeeInr;
  final DateTime calculatedAt;

  const FeeEstimate({
    required this.network,
    required this.assetSymbol,
    required this.slow,
    required this.standard,
    required this.fast,
    required this.platformFeeInr,
    required this.calculatedAt,
  });
}

class WhitelistedAddress {
  final String id;
  final String label;
  final String address;
  final BlockchainNetwork network;
  final String assetSymbol;
  final bool isCoolingPeriodActive;
  final DateTime coolingPeriodEndsAt;
  final DateTime createdAt;

  const WhitelistedAddress({
    required this.id,
    required this.label,
    required this.address,
    required this.network,
    required this.assetSymbol,
    required this.isCoolingPeriodActive,
    required this.coolingPeriodEndsAt,
    required this.createdAt,
  });
}

class CryptoTransaction {
  final String id;
  final String txHash;
  final TransferType type;
  final String assetSymbol;
  final BlockchainNetwork network;
  final double amount;
  final double networkFee;
  final String fromAddress;
  final String toAddress;
  final int currentConfirmations;
  final int requiredConfirmations;
  final ConfirmationState state;
  final String? blockExplorerUrl;
  final DateTime initiatedAt;
  final DateTime? finalizedAt;

  const CryptoTransaction({
    required this.id,
    required this.txHash,
    required this.type,
    required this.assetSymbol,
    required this.network,
    required this.amount,
    required this.networkFee,
    required this.fromAddress,
    required this.toAddress,
    required this.currentConfirmations,
    required this.requiredConfirmations,
    required this.state,
    this.blockExplorerUrl,
    required this.initiatedAt,
    this.finalizedAt,
  });
}
```

### Abstract Repository and Service Contracts
```dart
// lib/features/crypto_wallet/domain/interfaces/crypto_wallet_repository.dart

abstract class ICryptoWalletRepository {
  Future<DepositAddress> getDepositAddress({
    required String assetSymbol,
    required BlockchainNetwork network,
  });

  Future<FeeEstimate> getFeeEstimate({
    required String assetSymbol,
    required BlockchainNetwork network,
    required String destinationAddress,
    required double amount,
  });

  Future<List<WhitelistedAddress>> getWhitelistedAddresses({
    required BlockchainNetwork network,
  });

  Future<bool> screenDestinationAddress({
    required String destinationAddress,
    required BlockchainNetwork network,
  });

  Future<CryptoTransaction> submitWithdrawal({
    required String assetSymbol,
    required BlockchainNetwork network,
    required String destinationAddress,
    required double amount,
    required FeePriorityTier feeTier,
    required String idempotencyKey,
    required String biometricAuthToken,
    String? totpCode,
  });

  Future<List<CryptoTransaction>> getTransactionHistory({
    int page = 1,
    int pageSize = 20,
    BlockchainNetwork? networkFilter,
    TransferType? typeFilter,
  });

  Stream<CryptoTransaction> streamTransactionUpdates({
    required String userId,
  });
}

abstract class IAddressValidator {
  bool isValidAddress({
    required String address,
    required BlockchainNetwork network,
  });

  AddressFormat detectAddressFormat({
    required String address,
    required BlockchainNetwork network,
  });

  bool verifyEip55Checksum({
    required String hexAddress,
  });

  bool verifyBech32Checksum({
    required String bech32Address,
  });

  bool verifyBase58Solana({
    required String solanaAddress,
  });
}

abstract class ICryptoUriBuilder {
  String buildDepositUri({
    required BlockchainNetwork network,
    required String address,
    required String assetSymbol,
    double? amount,
    String? memo,
  });
}
```

### State Models
```dart
// lib/features/crypto_wallet/presentation/state/crypto_state_models.dart

class DepositScreenState {
  final CryptoAsset selectedAsset;
  final BlockchainNetwork selectedNetwork;
  final DepositAddress? depositAddress;
  final double? customAmount;
  final String qrUri;
  final bool isLoading;
  final String? errorMessage;

  const DepositScreenState({
    required this.selectedAsset,
    required this.selectedNetwork,
    this.depositAddress,
    this.customAmount,
    required this.qrUri,
    required this.isLoading,
    this.errorMessage,
  });
}

class WithdrawalFormState {
  final CryptoAsset selectedAsset;
  final BlockchainNetwork selectedNetwork;
  final String destinationAddress;
  final bool isAddressValid;
  final String? addressValidationError;
  final double amount;
  final double availableBalance;
  final FeePriorityTier selectedFeeTier;
  final FeeEstimate? feeEstimate;
  final WhitelistedAddress? selectedWhitelistedAddress;
  final bool isSanctionsChecking;
  final bool isSanctionsPassed;
  final bool isSubmitting;
  final String? errorMessage;

  const WithdrawalFormState({
    required this.selectedAsset,
    required this.selectedNetwork,
    required this.destinationAddress,
    required this.isAddressValid,
    this.addressValidationError,
    required this.amount,
    required this.availableBalance,
    required this.selectedFeeTier,
    this.feeEstimate,
    this.selectedWhitelistedAddress,
    required this.isSanctionsChecking,
    required this.isSanctionsPassed,
    required this.isSubmitting,
    this.errorMessage,
  });
}
```

## Security & Compliance Notes
- **Strict Address Checksum Validation:** The client must enforce cryptographic checksum verification (EIP-55, Bech32/Bech32m, Base58) before allowing the user to proceed with withdrawal dispatch. If an address fails checksum validation, the UI must block submission and highlight the formatting error.
- **Cross-Chain Network Mismatch Warning:** If an investor inputs an EVM address format (`0x...`) while Bitcoin or Solana is selected, the application must display a prominent warning banner preventing asset loss.
- **Whitelisted Address Cooling Period:** In compliance with regulatory standards, newly added withdrawal addresses are subjected to an immutable 24-hour cooling period during which withdrawals to that address remain disabled.
- **Biometric Challenge & 2FA Enforcement:** All withdrawal dispatches require biometric signature verification (`local_auth`) on mobile/desktop devices and mandatory TOTP/SMS secondary factor authorization.
- **AML and Sanctions Pre-Flight Screening:** Before submitting any on-chain withdrawal intent, the client invokes `/api/v1/compliance/screen-address` to verify the recipient address against global sanctions lists (OFAC, UN, EU).
- **Clipboard Hijacking Protection:** The withdrawal screen monitors clipboard paste events and prompts the user to visually confirm the first 6 and last 6 characters of the pasted address to protect against clipboard-hijacking malware.
- **Dynamic Deposit Expiration & Single-Use Addresses:** For enhanced privacy and accounting precision, deposit addresses carry explicit validity timestamps where applicable, and the UI displays real-time expiry countdowns.

## Acceptance Criteria
- [ ] Multi-chain network selector allows toggling between Bitcoin, Ethereum, Solana, Polygon, and Arbitrum with dynamic UI updates.
- [ ] `CryptoDepositScreen` renders crisp dynamic vector QR codes containing valid BIP-21, EIP-681, and Solana Pay URI payloads.
- [ ] One-tap copy button copies deposit address to clipboard, triggers haptic feedback, and displays confirmation toast.
- [ ] Share button triggers native OS share sheet with formatted address and network instructions.
- [ ] `AddressValidator` correctly validates valid/invalid addresses for Bech32 (BTC), EIP-55 (ETH, POL, ARB), and Base58 (SOL).
- [ ] Inputting an incompatible address format triggers immediate visual validation error and disables the withdrawal button.
- [ ] `DynamicFeeEstimatorWidget` fetches live gas fees and displays accurate crypto and fiat equivalents across Slow, Standard, and Fast tiers.
- [ ] `WithdrawalAuthBottomSheet` challenges the user with biometric authentication and TOTP verification before submitting.
- [ ] `TransactionConfirmationTrackerWidget` streams live block confirmation updates via WebSocket (e.g. 1/3 BTC, 8/12 ETH) and updates the visual progress bar.
- [ ] Tapping the block explorer button opens the corresponding external transaction explorer (Mempool.space, Etherscan, Solscan, PolygonScan, Arbiscan) via `url_launcher`.
- [ ] Whitelisted address selector displays active 24-hour cooling period indicators and disables unconfirmed addresses.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Flutter Scaffolding), Prompt 502 (Architecture & Riverpod State Management), Prompt 503 (Design System), Prompt 505 (Authentication UI), Prompt 521 (Secure Storage), Prompt 525 (API Client Layer).
- **Backend Dependencies:** Prompt 203 (Wallet Service), Prompt 213 (Custodian Depository Integration), Prompt 214 (Foreign Investor Funding), Prompt 309 (Event Indexing Service).
- **Downstream Blockers:** Prompt 512 (Transaction History Screen), Prompt 902 (End-to-End Integration Testing).
