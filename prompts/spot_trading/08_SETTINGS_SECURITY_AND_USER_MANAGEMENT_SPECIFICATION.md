# Settings, Security & User Management Specification

## 1. Overview & Navigation Anatomy

The User Settings and Security Management subsystem provides traders with granular control over authentication factors, security policies, display preferences, notification channels, and statutory privacy rights.

```
+---------------------------------------------------------------------------------------------------+
|                            USER SETTINGS & NAVIGATION TAXONOMY                                    |
+------------------------------------+--------------------------------------------------------------+
| Category                           | Configurable Parameters & Controls                           |
+------------------------------------+--------------------------------------------------------------+
| 1. Security Center                 | 2FA, FIDO2 Passkeys, Anti-Phishing Code, Withdrawal Whitelist|
| 2. Trading Preferences             | Default Order Type, Slippage Tolerance, Order Confirmations  |
| 3. Display & Appearance            | Dark Mode (Default), Chart Layouts, Monospace Formatting    |
| 4. Notification Center             | Execution Push, Margin Alerts, Liquidation Warnings, SMS     |
| 5. API Key Management              | Read-Only, Trading, Withdrawal Permissions, IP Whitelisting  |
| 6. Privacy & DPDP Rights           | Identity Commitments, Data Download, Crypto-Shredding Erasure|
+------------------------------------+--------------------------------------------------------------+
```

---

## 2. Security Center & Multi-Factor Controls

### 2.1 Multi-Factor Authentication (MFA)
- **FIDO2 / WebAuthn Hardware Passkeys:** Touch ID, Face ID, YubiKey, and Windows Hello. Provides hardware-backed, phishing-resistant authentication.
- **Time-Based One-Time Passwords (TOTP):** Compatible with Google Authenticator and 1Password using SHA-256 HMAC tokens.
- **SMS / WhatsApp Backup:** Emergency secondary recovery channel with mandatory 24-hour withdrawal holds upon SIM change detection.

### 2.2 Anti-Phishing Security Phrase
- Users configure a personalized secret phrase (e.g. `"GrowwwSecure2026"`).
- All official emails and withdrawal confirmation screens display this phrase, protecting users against malicious impersonation campaigns.

### 2.3 Withdrawal Address Whitelisting & 24-Hour Timelock
- Users can enable strict address whitelisting. Withdrawals are only permitted to pre-authorized Bitcoin (SegWit/Taproot) or USDT addresses.
- **24-Hour Cool-Off Invariant:** Adding a new withdrawal address initiates an automatic 24-hour security timelock during which no funds can be sent to that address.

### 2.4 API Key Governance
- Granular permission scopes: `READ_ONLY`, `SPOT_TRADE_EXECUTION`, `WITHDRAWAL`.
- Mandatory IP whitelisting for automated trading bots.
- Automated revocation of inactive API keys after 90 days.

---

## 3. Display, Theme & Trading Experience Preferences

### 3.1 Dark Mode by Default
- Default theme is fixed to **Deep Obsidian Dark Mode** (`#0B0E14`), optimized for reduced eye strain during extended trading sessions and OLED energy efficiency.
- High-contrast color accents: Emerald Green (`#00C087`) for buy/up-ticks; Crimson Red (`#FF3B30`) for sell/down-ticks.

### 3.2 Order Confirmation & Safety Toggles
- **Biometric Order Gate:** Option to require Face ID / Fingerprint confirmation on real-money orders exceeding configured value (e.g. ₹50,000 or 1,000 USDT).
- **One-Tap Market Orders:** Configurable toggle allowing instant execution for high-frequency discretionary traders.

---

## 4. Privacy & Statutory DPDP Act 2023 Data Rights

Under the Digital Personal Data Protection Act 2023 (DPDP Act):
1. **Download Personal Data:** Users can request a cryptographically signed JSON archive of their complete account and trading history.
2. **Right to Erasure (Crypto-Shredding):**
   - User submits an account closure and data erasure request.
   - The platform permanently destroys the user's Data Encryption Key (DEK) inside the CloudHSM.
   - All historical personal data in off-chain databases and logs becomes mathematically unrecoverable.
   - On-chain Poseidon identity commitments are revoked, rendering historical blockchain addresses permanently anonymous.
