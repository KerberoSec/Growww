# Sovereign Token Standards & Decimal Specification

## 1. Executive Overview
Implements compliant ERC-20 and ERC-3643 tokens representing spot crypto assets (wBTC, wUSDT) and sovereign digital currency on Hyperledger Besu.

## 2. Token Inventory & Decimal Standards
1. **Wrapped Bitcoin (`wBTC`)**:
   - Standard: ERC-20 with mint/burn hooks restricted to MPC Custody.
   - Decimals: **8** (`1 wBTC = 100,000,000 Satoshis`).
2. **Wrapped Tether (`wUSDT`)**:
   - Standard: ERC-20 with multichain bridge mint/burn authorization.
   - Decimals: **6** (`1 wUSDT = 1,000,000 Micro-USDT`).
   - Matches Ethereum, Tron, and Arbitrum USDT standards.
3. **Digital Rupee (`w-eINR`)**:
   - Standard: ERC-3643 permissioned token backed 1:1 by Reserve Bank of India e-Rupee.
   - Decimals: **2** external display (`1 w-eINR = 100 Paise`), backed by **6-decimal internal accounting** ($10^{-6}$ micro-eINR) to guarantee zero sub-paise rounding leakage.

## 3. Decimal Normalization & Arithmetic Invariants
When calculating trading pair interactions (BTC/USDT):
- **Price Representation**: 6 decimal places (matching USDT quote precision).
- **Quantity Representation**: 8 decimal places (matching BTC satoshi precision).
- **Notional Calculation**:
  `Notional_USDT_Micro = (Quantity_Satoshis * Price_Micro) / 10^8`
- **Rounding Rule**: Floor rounding towards zero for payouts; fractional remainder dust below 1 micro-unit is directed to the Settlement Guarantee Fund reserve.

## 4. Fee & Frictionless Economics
- **Zero Transfer Tax**: 0% transfer tax or burn on token transfers.
- **Gasless User Experience**: Compatible with EIP-2612 permit and ERC-4337 Paymaster sponsorship.
