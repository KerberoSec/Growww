# SRE Operational Runbook 040: iOS App Store Financial Exchange Compliance & Release Protocol

**Runbook ID:** RUNBOOK-040-IOS-COMPLIANCE  
**Severity Classification:** Tier 1 (Mobile Distribution & App Store Review)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** iOS Mobile Application (`apps/growww_flutter/ios`)  
**Apple Review Guidelines:** Guideline 3.1.5 (Cryptocurrencies), Guideline 5.1.1 (Data Privacy)  
**Last Review:** September 2026  

---

## 1. Executive Summary

Apple App Store Review imposes stringent regulatory and architectural standards on financial exchange applications under App Store Review Guideline 3.1.5. Submissions must demonstrate appropriate licensing (SEBI/IFSCA), zero in-app purchase (IAP) commission bypass, proof of 1:1 custodial backing, zero on-chain PII, and hardware biometric authorization via FaceID / TouchID.

---

## 2. Statutory Review Documentation Matrix

During Apple App Store submission, the following evidentiary dossier is attached to the App Review Notes:
1. **Regulatory License Filings:** Copy of SEBI Stock Broker Registration & IFSCA GIFT City FinTech Sandbox Approval.
2. **Custodial Proof-of-Reserve Attestation:** Cryptographic audit link pointing to real-time on-chain Merkle root and depository statements.
3. **Zero-PII On-Chain Compliance Certificate:** Proof that PAN numbers, Aadhaar IDs, and bank accounts are never committed to Hyperledger Besu.
4. **Zero-Fee Operating Model Declaration:** Attestation confirming that trading fees are 0.00% (No in-app purchase fees applicable).

---

## 3. iOS Info.plist & Privacy Manifest (`PrivacyInfo.xcprivacy`)

To comply with Apple's mandatory Privacy Manifest enforcement (iOS 17+):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>NSPrivacyTracking</key>
    <false/>
    <key>NSPrivacyCollectedDataTypes</key>
    <array>
        <dict>
            <key>NSPrivacyCollectedDataType</key>
            <string>NSPrivacyCollectedDataTypeUserID</string>
            <key>NSPrivacyCollectedDataTypeLinked</key>
            <true/>
            <key>NSPrivacyCollectedDataTypeTracking</key>
            <false/>
            <key>NSPrivacyCollectedDataTypePurposes</key>
            <array>
                <string>NSPrivacyCollectedDataTypePurposeAppFunctionality</string>
            </array>
        </dict>
    </array>
    <key>NSPrivacyAccessedAPITypes</key>
    <array>
        <dict>
            <key>NSPrivacyAccessedAPIType</key>
            <string>NSPrivacyAccessedAPICategoryUserDefaults</string>
            <key>NSPrivacyAccessedAPITypeReasons</key>
            <array>
                <string>CA92.1</string>
            </array>
        </dict>
    </array>
</dict>
</plist>
```

---

## 4. Verification Protocol

1. Test FaceID biometric trade challenge:
   - Attempt test withdrawal; assert FaceID challenge appears via `LocalAuthentication`.
2. Test Background Socket Suspension:
   - Background app; verify socket closes within 15 seconds, preventing battery drain rejection.
3. Test Zero-Fee Invariant:
   - Verify buy/sell order entry sheet displays: `Trading Fee: ₹0.00 (0.00% Genesis Zero-Fee)`.
