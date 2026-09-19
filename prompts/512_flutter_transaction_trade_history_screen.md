# 512 - Flutter Transaction & Trade History Screen

## Purpose
Provides an immutable, tamper-evident transaction journal and trade history ledger for investors, auditors, and regulators. In a compliant financial system where fractional asset trades are settled atomically via permissioned smart contracts, every trade, deposit, withdrawal, corporate dividend, and Universal Zero-Fee Model (0.00% fee - No fee at all) deduction must be searchable, filterable, and backed by dual verifiable receipts: an official SEBI-compliant Electronic Contract Note (ECN) and an immutable on-chain transaction hash.

## What You Are Building
A high-performance transaction and trade history module in `apps/growww_flutter/lib/features/history/` featuring:
- **Comprehensive Ledger View:** Infinite scrolling list of all historical account events (Trades, Cash Deposits, Withdrawals, Dividend Credits, Platform Fees (0.00% (Zero Fee)), Corporate Actions).
- **Multi-Dimensional Filtering & Search:** Filter by date range, transaction category, specific security ISIN/symbol, and settlement status (Completed, Pending, Failed).
- **Trade Detail Bottom Sheet / Inspector:** Displays execution price, fractional quantity, trade execution timestamp, statutory charge breakdown, and SEBI Order ID.
- **Dual Verification Receipt:**
 - One-tap download of SEBI-compliant PDF Contract Note.
 - Blockchain Verification Card displaying Besu Transaction Hash (`0x...`), Block Number, and DvP settlement timestamp with block explorer launch button.
- **Statement & Tax Export:** Export filtered records as CSV or digitally signed PDF statements for tax filing and accounting.

## Scope Boundaries
- **In Scope:**
 - Transaction history UI, infinite scroll pagination, filter chips, trade receipt drawer, contract note PDF download trigger, and on-chain tx hash inspection.
- **Out of Scope / Handled Elsewhere:**
 - Backend contract note PDF generation service (handled in Prompt 216).
 - On-chain event indexing and transaction retrieval backend (handled in Prompt 309).
 - Tax calculation engine (handled in Prompt 223).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` infinite scroll pagination.
 - *Justification:* Riverpod combined with `SliverList` ensures smooth 120 FPS scrolling over thousands of historical ledger records with minimal memory footprint.
- **Document Viewing & Sharing:** `open_file` and `share_plus` for opening and exporting downloaded Contract Note PDFs.
- **Date Formatting:** `intl` package for localized Indian Standard Time (IST) formatting.

## Backend / Infra Touchpoints
- **Reporting Service:** REST endpoints (`/api/v1/history/transactions`, `/api/v1/history/contract-note/:tradeId`) from Prompt 216.
- **Trade Settlement Service:** Settlement details query (`/api/v1/settlement/details/:orderId`) from Prompt 208.
- **Event Indexing Service:** On-chain transaction receipts from Prompt 309.

## Blockchain Interaction
- **Immutable On-Chain Transaction Verification:**
 - Every executed trade displays its corresponding Hyperledger Besu transaction hash (`SettlementDvP.sol` execution).
 - Users can tap "Verify on Blockchain Explorer" to view the smart contract log event proving atomic DvP delivery:
 - Escrow locked -> Token transferred to investor wallet -> Fiat settled -> Timestamp recorded on ledger.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/history/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/`.
2. Define domain models: `TransactionRecord`, `TransactionType` (TradeBuy, TradeSell, Deposit, Withdrawal, Dividend, PlatformFee), `TradeDetails`, `BlockchainReceipt`.
3. Implement `TransactionHistoryNotifier` with pagination state (page number, hasMore, isFetchingNextPage).
4. Build `TransactionHistoryScreen` with a top search bar and horizontal filter chips (All, Trades, Funds, Dividends, Fees).
5. Implement `DateRangeFilterSheet` allowing quick ranges (Today, 7D, 30D, FY 2025-26, Custom Range).
6. Build `TransactionListTile` with semantic icons (Green arrow for buy/deposit, Red arrow for sell/withdrawal, Gold coin for dividend, Blue lock for fee).
7. Build `TradeDetailModalSheet` showing breakdown of traded security, fractional units, executed price, STT, GST, and Universal Zero-Fee Model (0.00% fee - No fee at all) (0.00% fee at launch (future fee parameters governed by FeeController.sol)).
8. Implement `BlockchainVerificationCard` inside the detail sheet with one-tap copy for `tx_hash` and button to launch the Besu block explorer.
9. Implement Contract Note downloader: fetches signed PDF byte stream from backend and opens via native PDF viewer (`open_file`).
10. Build `ExportStatementDialog` allowing CSV/PDF export with date filters.
11. Implement sticky date headers (e.g., "September 2026", "August 2026") using `SliverStickyHeader`.
12. Write widget tests for pagination loading triggers, filter combinations, and detail sheet rendering.

## Interfaces / Contracts
```dart
// lib/features/history/domain/models/transaction_record.dart
enum TransactionType { tradeBuy, tradeSell, deposit, withdrawal, dividendCredit, platformFeeDeduction }
enum TransactionStatus { completed, pending, failed }

class TransactionRecord {
  final String id;
  final String title;
  final String? subtitle;
  final TransactionType type;
  final TransactionStatus status;
  final double amountInr;
  final double? fractionalQuantity;
  final double? unitPriceInr;
  final DateTime timestamp;
  final String? sebiContractNoteId;
  final String? blockchainTxHash;
  final int? blockchainBlockNumber;

  const TransactionRecord({
    required this.id,
    required this.title,
    this.subtitle,
    required this.type,
    required this.status,
    required this.amountInr,
    this.fractionalQuantity,
    this.unitPriceInr,
    required this.timestamp,
    this.sebiContractNoteId,
    this.blockchainTxHash,
    this.blockchainBlockNumber,
  });
}
```

## Security & Compliance Notes
- **SEBI 8-Year Record Mandate:** The history screen must allow querying up to 8 years of historical records as required by SEBI record retention regulations.
- **Cryptographic Receipt Integrity:** Displayed transaction hashes are verified against the local Besu network RPC to guarantee that local records have not been altered or tampered with.
- **PII Protection on Exports:** Exported statements must redact sensitive banking account details, displaying only masked account numbers (`XXXX-XXXX-1234`).

## Acceptance Criteria
- [ ] Transaction history paginates smoothly without duplicate entries or dropped frames.
- [ ] Category chips and date filters instantly filter ledger records.
- [ ] Tapping a trade opens the detail sheet showing full breakdown and on-chain tx hash.
- [ ] Contract Note PDF downloads and opens in native viewer across mobile and desktop.
- [ ] Blockchain Explorer link opens the correct on-chain transaction receipt.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System).
- **Backend Dependency:** Prompt 208 (Settlement), Prompt 216 (Reporting Service), Prompt 309 (Event Indexing).
- **Enables:** Prompts 513 (Notifications Center), 514 (Settings & Profile).
