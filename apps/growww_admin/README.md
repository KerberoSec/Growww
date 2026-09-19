# Growww Back-Office Admin & Compliance Portal

## Purpose & Scope
`growww_admin` is the internal back-office workstation for compliance officers, dispute arbitrators, and executive risk managers. It strictly enforces Role-Based Access Control (RBAC) and Maker-Checker dual authorization.

## Key Operational Modules
- **KYC Verification Review**: Tier 1 to Tier 3 manual exception reviews and Aadhaar/PAN audit verification.
- **P2P Dispute Arbitration**: Maker-checker desk for resolving contested INR-to-USDT bank escrow transfers.
- **Proof-of-Reserve Auditor**: Hourly Merkle liability tree reconciliations against on-chain Besu reserve balances.
- **Regulatory Reporting**: Direct SEBI, RBI, and FIU-IND report export (STR, CTR, Form 26Q e-TDS).

## Security Standards
- Hardware FIDO2 / WebAuthn token mandatory for all administrative staff logins.
- Immutable WORM audit logging capturing every staff action, click, and query diff.
