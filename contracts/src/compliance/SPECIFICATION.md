# Compliance & Identity Claims Registry Specification

## 1. Executive Overview
Implements the ERC-3643 (T-REX) compliance suite on Hyperledger Besu to enforce on-chain regulatory compliance, KYC verification, and statutory transfer restrictions without storing Personally Identifiable Information (PII).

## 2. Core Architecture & Zero-PII Invariant
- **Zero-PII On-Chain**: No names, PAN numbers, Aadhaar numbers, or email addresses are stored on the blockchain.
- **Cryptographic Identity Claims**: Traders are identified by their wallet address linked to an `ONCHAINID` contract holding 32-byte cryptographic claim hashes issued by authorized KYC verification authorities.
- **Claim Topics**:
  - Topic 1: Identity Verified (KYC/AML verified).
  - Topic 2: Jurisdiction Accreditation (FATF / India resident compliance).
  - Topic 3: Investor Classification (Retail vs Institutional).

## 3. Smart Contracts Inventory
- `IdentityRegistry.sol`: Maps wallet addresses to their corresponding ONCHAINID identity contracts.
- `ClaimTopicsRegistry.sol`: Maintains trusted claim topics required for asset trading.
- `TrustedIssuersRegistry.sol`: Manages public keys of authorized KYC verification services.
- `ComplianceRules.sol`: Modular rule engine evaluating whether a proposed token transfer satisfies:
  - Sanction list screening.
  - Maximum daily transfer limits.
  - Geographic jurisdiction restrictions.

## 4. Transfer Hook Execution Flow
Before any token transfer executes on Besu:
1. Token contract calls `compliance.canTransfer(sender, recipient, amount)`.
2. Compliance contract queries `IdentityRegistry` for both sender and recipient.
3. If both parties hold valid, unexpired claims from a trusted issuer and comply with velocity rules, transfer returns `true`.
4. Otherwise, transaction reverts with a descriptive error code (e.g., `COMPLIANCE_RECIPIENT_NOT_VERIFIED`).
