# SRE Operational Runbook 037: macOS Native Thick Client Code Signing, Notarization & Release Pipeline

**Runbook ID:** RUNBOOK-037-MACOS-SIGNING  
**Severity Classification:** Tier 1 (Desktop Distribution & Security Compliance)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** macOS Native Thick Client (`apps/growww_flutter/macos`)  
**Target Architectures:** Apple Silicon (`arm64`), Intel (`x86_64`), Universal Mach-O Binary  
**Last Review:** September 2026  

---

## 1. Executive Summary

Apple macOS mandates strict Gatekeeper security controls. Any production thick client distributed outside the Mac App Store must be cryptographically signed with a valid Apple Developer ID Application certificate, compiled with Hardened Runtime, and notarized by Apple Ticket servers via `xcrun notarytool`. Unnotarized builds trigger scary malware warning dialogs that block institutional and retail traders from launching the application.

---

## 2. Automated Notarization & Release Architecture

```
+-------------------------------------------------------------------------------------------------------+
|                                    MACOS BUILD & NOTARIZATION PIPELINE                                |
|                                                                                                       |
|  [ GitHub Actions macOS-14 M1 Runner ]                                                                |
|          |                                                                                            |
|          +---> 1. Compile Rust FFI Core (`cargo build --release --target aarch64 & x86_64`)           |
|          +---> 2. Merge Universal Binary via `lipo -create -output librust_trading_core.dylib`       |
|          +---> 3. Build Flutter Desktop Application (`flutter build macos --release`)                |
|          |                                                                                            |
|          v                                                                                            |
|  [ Code Signing with Apple Developer ID ]                                                             |
|  - Embed Entitlements (`entitlements.plist`) with Hardened Runtime                                   |
|  - Sign nested dynamic libraries (`codesign --sign "Developer ID Application: ..."` )                |
|  - Sign main App Bundle (`Growww.app`) with secure timestamp (`--timestamp`)                         |
|          |                                                                                            |
|          v                                                                                            |
|  [ DMG Packaging & Apple Notarization ]                                                               |
|  - Package into `Growww-macOS.dmg` via `create-dmg`                                                   |
|  - Submit to Apple Notarization Service (`xcrun notarytool submit --keychain-profile "AC_PASSWORD"`) |
|  - Staple Notarization Ticket (`xcrun stapler staple "Growww-macOS.dmg"`)                             |
|          |                                                                                            |
|          v                                                                                            |
|  [ Production Distribution (AWS S3 / Cloudflare R2 / Sparkle Auto-Update Feed) ]                     |
+-------------------------------------------------------------------------------------------------------+
```

---

## 3. Entitlements Configuration (`Entitlements.plist`)

To enable multi-window IPC, Rust FFI unmanaged memory buffers, Apple Secure Enclave Keychain access, and audio synthesis:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <false/>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.network.server</key>
    <true/>
    <key>keychain-access-groups</key>
    <array>
        <string>$(AppIdentifierPrefix)in.growww.desktop</string>
    </array>
</dict>
</plist>
```

---

## 4. Verification & Gatekeeper Acceptance Protocol

Prior to publishing any macOS desktop release:
1. Verify signature validity:
   ```bash
   codesign --verify --deep --strict --verbose=2 /Applications/Growww.app
   ```
2. Verify Apple Notarization staple:
   ```bash
   spctl --assess --type execute --verbose /Applications/Growww.app
   # Must output: /Applications/Growww.app: accepted, source=Notarized Developer ID
   ```
3. Verify Universal Mach-O Architecture:
   ```bash
   file /Applications/Growww.app/Contents/MacOS/Growww
   # Must output: Mach-O universal binary with 2 architectures: [x86_64] [arm64]
   ```
4. Verify Zero-Fee Invariant: Launch application, place test paper trade, assert that fee confirmation renders `0.00% ZERO-FEE (₹0.00)`.
