# 510 - Flutter Portfolio & Holdings Screen (Fractional Units, P&L, Proof-of-Reserve)

## Purpose
Provides a transparent, comprehensive portfolio management interface displaying an investor's fractional share holdings, asset allocation breakdown, unrealized and realized P&L, historical return metrics (XIRR), and independent cryptographic verification of each asset against on-chain Proof-of-Reserve records. This screen embodies the Growww core promise: transparent, fractional real-security ownership backed 1:1 in SEBI-regulated custody with a zero-holding-fee policy and canonical Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover (0.00% fee at launch (governed by FeeController.sol) split), with FIFO capital gains computed strictly for user tax compliance (Section 111A/112A).

## What You Are Building
A feature-packed portfolio and holdings view located in `apps/growww_flutter/lib/features/portfolio/` containing:
- **Portfolio Valuation Header:** Total Current Value, Invested Value, Total Overall P&L (in INR and %), and Today's P&L.
- **Transaction Fee & Tax Summary:** Breakdown card displaying cumulative 0.00% (Zero Fee) platform transaction fees on turnover (with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol)), FIFO capital gains computed strictly for user tax compliance (Section 111A/112A), and a prominent "₹0 Holding Fees / Zero AUM Charges" badge.
- **Asset Allocation Chart:** Interactive donut chart displaying sector and asset class distribution (Equities, ETFs, Government Bonds, GIFT City listings).
- **Holdings List with Fractional Precision:** Security list displaying fractional quantities (e.g., 1.4520 shares), average buy price, current market price (CMP), and individual P&L.
- **On-Chain Proof-of-Reserve Verification Modal:** One-tap audit tool per holding that validates the user's tokenized balance against the smart contract Merkle root.
- **Export Statements Sheet:** Quick generator for capital gains statements, holding certificates, and depository audit reports.

## Scope Boundaries
- **In Scope:**
 - Portfolio summary UI, asset allocation donut chart, fractional holdings list, client-side Merkle proof verification visualizer, and filter/sorting controls.
- **Out of Scope / Handled Elsewhere:**
 - Full tax statement PDF generation engine (handled in Prompt 223).
 - Corporate action processing (handled in Prompt 222).
 - Off-chain depository reconciliation pipeline (handled in Prompt 215).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with `flutter_riverpod` state management.
 - *Justification:* Riverpod `StateNotifier` facilitates smooth local filtering and sorting (by Value, P&L, Alphabetical) without re-fetching data from the backend.
- **Charts:** `fl_chart` for responsive, animated Pie/Donut asset allocation charts.
- **Cryptography / Web3:** `web3dart` and `crypto` for client-side SHA-256 Merkle proof verification.

## Backend / Infra Touchpoints
- **Portfolio Service:** REST endpoint (`/api/v1/portfolio/holdings`) and analytics endpoint (`/api/v1/portfolio/analytics`) from Prompt 209.
- **Fee Engine:** Realized P&L fee audit endpoint (`/api/v1/fees/realized-summary`) from Prompt 210.
- **Proof-of-Reserve Registry:** Merkle root query endpoint (`/api/v1/blockchain/por/proof/:holdingId`) from Prompt 308.

## Blockchain Interaction
- **Client-Side Merkle Proof Verification:**
 - For each holding, the user can tap "Verify on Ledger".
 - The client fetches the Merkle leaf node for the user's holding and the Merkle branch path from the indexer.
 - The client independently computes the SHA-256 root hash in Dart and checks if it matches the current root stored in `ProofOfReserveRegistry.sol`.
 - Displays a green cryptographic verification seal: *"Holding Verified on Hyperledger Besu Block #1,492,012"*.

## Step-by-Step Build Instructions
1. Scaffold feature directories in `lib/features/portfolio/`: `presentation/screens/`, `presentation/widgets/`, `application/`, `domain/models/`, `data/`.
2. Define domain models: `PortfolioSummary`, `HoldingItem`, `AssetAllocationSlice`, `MerkleProofVerificationResult`.
3. Implement `PortfolioController` with pull-to-refresh and auto-refresh on order completion events.
4. Build `PortfolioSummaryCard` displaying Current Value, Invested Value, Total Return, and Today's P&L.
5. Build `PlatformFeeSummaryWidget` detailing fixed 0.00% transaction fee (No fee at all)s paid on trade turnover (with 0.00% fee at launch (governed by FeeController.sol) allocation), FIFO tax compliance gains (Section 111A/112A), and celebrating zero holding/AUM charges.
6. Build `AssetAllocationDonutChart` using `fl_chart`, supporting touch highlights on individual sectors.
7. Implement `HoldingsListView` with sorting chips (Value, Today's Return, Total Return, Name) and search filtering.
8. Create `HoldingItemCard` displaying company logo, fractional quantity (up to 4 decimal places), invested vs current value, and Proof-of-Reserve badge.
9. Implement `MerkleProofVerificationSheet` that executes the client-side cryptographic hashing algorithm and displays visual step-by-step verification nodes.
10. Implement quick swipe actions on holding cards: Swipe Right for "Add More (Buy)", Swipe Left for "Sell / Exit".
11. Build `ExportStatementModal` allowing date-range selection for holding statements.
12. Write unit tests for Merkle root computation and widget tests for portfolio sorting.

## Interfaces / Contracts
```dart
// lib/features/portfolio/domain/models/holding_item.dart
class HoldingItem {
  final String symbol;
  final String isin;
  final String companyName;
  final double fractionalQuantity; // e.g., 2.3450 shares
  final double averageBuyPrice;
  final double currentMarketPrice;
  final double investedAmount;
  final double currentValue;
  final double unrealizedPnlInr;
  final double unrealizedPnlPercentage;
  final String tokenContractAddress;
  final String merkleLeafHash;

  const HoldingItem({
    required this.symbol,
    required this.isin,
    required this.companyName,
    required this.fractionalQuantity,
    required this.averageBuyPrice,
    required this.currentMarketPrice,
    required this.investedAmount,
    required this.currentValue,
    required this.unrealizedPnlInr,
    required this.unrealizedPnlPercentage,
    required this.tokenContractAddress,
    required this.merkleLeafHash,
  });
}
```

## Security & Compliance Notes
- **PAN & Demat Masking:** Any exported screenshots or PDF certificates must mask user PAN and Demat account numbers per DPDP Act guidelines.
- **Accurate Tax Lot Accounting:** Realized capital gains must use FIFO (First-In, First-Out) accounting strictly for user tax compliance in accordance with Section 111A/112A of the Indian Income Tax Act.
- **Zero Hidden Charges:** The fee transparency card must accurately reflect that holding fees and AUM charges are ₹0.00.

## Acceptance Criteria
- [ ] Portfolio balances and fractional shares display accurately to 4 decimal places.
- [ ] Asset allocation donut chart updates dynamically and supports interactive slice touch.
- [ ] Client-side Merkle proof verification validates successfully against on-chain root.
- [ ] Sorting and filtering of holdings execute instantly in memory.
- [ ] Quick swipe actions trigger order placement flows with pre-filled holding details.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Scaffolding), Prompt 502 (Architecture), Prompt 503 (Design System).
- **Backend Dependency:** Prompt 209 (Portfolio Service), Prompt 210 (Fee Engine), Prompt 308 (Proof of Reserve).
- **Enables:** Prompt 511 (Wallet & Funds), Prompt 512 (Transaction History).
