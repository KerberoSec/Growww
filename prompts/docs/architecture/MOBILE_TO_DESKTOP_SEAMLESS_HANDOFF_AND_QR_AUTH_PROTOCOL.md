# Mobile-to-Desktop Seamless Handoff & QR Biometric Authentication Protocol

**Specification ID:** SPEC-ARCH-044-HANDOFF-QR  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Authentication, Cross-Device Session Handoff & Cryptography  
**Target Environments:** Web Pro Terminal, Desktop Thick Clients (macOS, Windows, Linux), Mobile (iOS & Android)  
**Last Updated:** September 2026  

---

## 1. Executive Summary

Traders frequently move between mobile phones (iOS / Android) and multi-monitor desktop workstations (Web Pro Terminal or macOS/Windows Native Thick Clients). Requiring passwords, complex 2FA manual typing, or hardware YubiKeys on every desktop session login slows rapid trade entry.

The **Mobile-to-Desktop Seamless Handoff Protocol** allows an active user to log in instantly by scanning a dynamic QR code on the desktop screen using their mobile device. The mobile app authorizes the desktop session via hardware biometric validation (TouchID / FaceID / Android BiometricPrompt) and an EIP-712 structured cryptographic signature, initializing an ephemeral desktop session token in under 800 milliseconds without ever exposing master private keys.

---

## 2. Cryptographic Protocol Architecture

```
+-------------------------------------------------------------------------------------------------------+
|                                    QR BIOMETRIC HANDOFF PROTOCOL                                      |
|                                                                                                       |
|  [ Desktop Terminal ]                      [ Auth Gateway Service ]               [ Mobile App ]      |
|  (Web / macOS / Win)                       (`auth-service`)                       (iOS / Android)     |
|          |                                         |                                     |            |
|          | --- 1. Request Session QR Challenge --> |                                     |            |
|          | <-- 2. Ephemeral Challenge & UUID ----- |                                     |            |
|          |                                         |                                     |            |
|     [ Render QR ]                                  |                                     |            |
|     (60s Expiry)                                   |                                     |            |
|          |                                         |                                     |            |
|          | ................ 3. User Scans QR with Camera ...............................> |            |
|          |                                         |                                     |            |
|          |                                         |                            [ Biometric Prompt ]  |
|          |                                         |                            (TouchID / FaceID)    |
|          |                                         |                                     |            |
|          |                                         | <-- 4. EIP-712 Signature + Push --- |            |
|          |                                         |     (Signed by Secure Enclave)      |            |
|          |                                         |                                                  |
|          |                                    [ Verify Sig ]                                          |
|          |                                    [ Mint Token ]                                          |
|          |                                         |                                                  |
|          | <-- 5. Session Active via WSS --------- |                                                  |
|          |                                                                                            |
|   [ Desktop Logged In ]                                                                               |
|   (Watchlist & Layout Synced)                                                                         |
+-------------------------------------------------------------------------------------------------------+
```

---

## 3. Protocol Message Contracts

### 3.1 EIP-712 Typed Structured Challenge
The desktop terminal requests a dynamic challenge payload formatted according to EIP-712:

```json
{
  "types": {
    "EIP712Domain": [
      { "name": "name", "type": "string" },
      { "name": "version", "type": "string" },
      { "name": "chainId", "type": "uint256" },
      { "name": "verifyingContract", "type": "address" }
    ],
    "DesktopSessionAuth": [
      { "name": "sessionNonce", "type": "bytes32" },
      { "name": "desktopClientType", "type": "string" },
      { "name": "ipAddressHash", "type": "bytes32" },
      { "name": "timestamp", "type": "uint64" },
      { "name": "feeTierInvariant", "type": "string" }
    ]
  },
  "primaryType": "DesktopSessionAuth",
  "domain": {
    "name": "Growww NBSE Institutional Exchange",
    "version": "1.0",
    "chainId": 2026,
    "verifyingContract": "0x0000000000000000000000000000000000002026"
  },
  "message": {
    "sessionNonce": "0x9f83a4b7...21c4",
    "desktopClientType": "DESKTOP_MACOS_PRO",
    "ipAddressHash": "0xe2b4c1...8910",
    "timestamp": 1792454400,
    "feeTierInvariant": "0.00% ZERO-FEE VIP"
  }
}
```

---

## 4. Security Controls & Anti-Phishing Guardrails

1. **60-Second Dynamic Challenge Expiry:**
   - The QR challenge is single-use and invalidates automatically after 60 seconds.
   - Replay attempts trigger immediate IP rate-limiting.
2. **Proximity & IP Correlation:**
   - The auth gateway compares the IP subnet and ASN of the desktop terminal and scanning mobile phone.
   - If disparate geolocations are detected (e.g. desktop in Mumbai, mobile in London), the handoff is rejected with a security alert toast.
3. **Session Revocation & Kill Switch:**
   - Active desktop sessions can be terminated with one tap from the mobile application's Security Center (`/settings/security/active-sessions`).
4. **Strict Zero-Fee Enforcement:**
   - The authentication token binds the user session to the zero-fee invariant (`0.00% maker / 0.00% taker / 0 gas / 0 TDS`), ensuring that no platform fees can ever be injected into orders dispatched from the authorized workstation.

---

## 5. Live State Handoff (Watchlists & Draft Orders)

Upon successful QR authentication, the central synchronization service automatically pushes the user's active state:
- **Active Watchlist:** Pushes pinned tickers and real-time custom tags.
- **Chart Layouts & Indicators:** Rehydrates saved TradingView indicator templates.
- **Draft Order Form:** If the user had begun composing an order on their mobile phone, the draft unsubmitted parameters (symbol, lot size, limit price) populate the desktop order entry ticket instantaneously.
