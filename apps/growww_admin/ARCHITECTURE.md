# Architecture & Role-Based Access Control Specification

## Executive Overview
`growww_admin` is the internal operations workstation enforcing strict Role-Based Access Control (RBAC) and Maker-Checker dual authorization for compliance officers, support staff, and risk managers.

## Security & Access Invariants
- Mandatory FIDO2 / WebAuthn hardware token authentication for all staff logins.
- Session timeout locked to 10 minutes of inactivity.
- Write-Once-Read-Many (WORM) audit logging capturing every staff query, click, and state modification.

## Operational Desks
1. **KYC Review Desk**: Manual inspection of flagged identity documents and video KYC liveness checks.
2. **P2P Dispute Desk**: Maker-checker arbitration verifying bank statement proofs for contested INR/USDT transfers.
3. **Proof-of-Reserve Desk**: Hourly reconciliation between on-chain Besu balances and internal ledger liabilities.
4. **Market Surveillance Desk**: Neo4j graph viewer visualizing wash trading loops and orderbook layering.
