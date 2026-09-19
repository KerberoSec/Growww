# SRE Operational Runbook 041: Android Google Play Financial Services Policy Compliance & Release Protocol

**Runbook ID:** RUNBOOK-041-ANDROID-COMPLIANCE  
**Severity Classification:** Tier 1 (Mobile Distribution & Google Play Console)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** Android Mobile Application (`apps/growww_flutter/android`)  
**Target SDK Versions:** Target SDK 34 (Android 14), Minimum SDK 26 (Android 8.0)  
**Google Play Policies:** Financial Services Policy, Cryptographic Asset Guidelines, Data Safety Section  
**Last Review:** September 2026  

---

## 1. Executive Summary

Google Play Console enforces rigorous compliance verifications for financial exchanges operating in India, including submission of FIU-IND (Financial Intelligence Unit - India) reporting entity certificates, strict Data Safety declarations (DPDP Act 2023 alignment), and Android `BiometricPrompt` Class 3 (Strong) biometric authentication.

---

## 2. Android App Bundle (AAB) & Play App Signing

1. The Android application is packaged exclusively as an Android App Bundle (`.aab`) with Play Feature Delivery and dynamic language splits.
2. Binary signature is protected via Google Play App Signing with upload keys stored in an encrypted PKCS#12 keystore.
3. R8 / ProGuard code shrinking and obfuscation rules protect cryptographic primitives and API routes from reverse engineering.

---

## 3. AndroidManifest.xml Security Configuration

```xml
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="in.growww.nbse.trading">

    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
    <uses-permission android:name="android.permission.USE_BIOMETRIC" />
    <uses-permission android:name="android.permission.POST_NOTIFICATIONS" />

    <application
        android:label="Growww NBSE"
        android:icon="@mipmap/ic_launcher"
        android:allowBackup="false"
        android:fullBackupContent="false"
        android:networkSecurityConfig="@xml/network_security_config"
        android:hardwareAccelerated="true">

        <activity
            android:name=".MainActivity"
            android:exported="true"
            android:windowSoftInputMode="adjustResize"
            android:configChanges="orientation|keyboardHidden|keyboard|screenSize|smallestScreenSize|locale|layoutDirection|fontScale|screenLayout|density|uiMode">
            <intent-filter>
                <action android:name="android.intent.action.MAIN"/>
                <category android:name="android.intent.category.LAUNCHER"/>
            </intent-filter>
        </activity>
    </application>
</manifest>
```

---

## 4. Verification Protocol

1. Verify `FLAG_SECURE` screen privacy:
   - When app transitions to background or app switcher, the screen displays a blank Obsidian shield, preventing Android OS from caching sensitive balance screenshots.
2. Verify BiometricPrompt Class 3 challenge:
   - Perform high-value transfer; assert native Android Biometric prompt renders with subtitle and cancellation hooks.
3. Verify Zero-Fee Invariant:
   - Confirm order entry ticket renders `0.00% Zero-Fee (₹0.00 Comm)` and 0 Gas Paymaster sponsored badge.
