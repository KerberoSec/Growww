# 520 - Platform-Specific Packaging: macOS (Notarized .app & .dmg) Build Pipeline

## Purpose
Institutional asset managers, professional wealth advisors, and high-net-worth individual (HNWI) investors frequently operate on macOS workstations (Mac Studio, MacBook Pro). Distributing a commercial desktop application outside the Mac App Store requires compliance with Apple Gatekeeper, code signing with an Apple Developer ID Application certificate, enabling Apple Hardened Runtime, and undergoing automated cloud notarization via Apple Notary Service (`notarytool`).

This prompt establishes the production macOS build, signing, notarization, and DMG packaging pipeline for the Growww Flutter client. It defines universal binary compilation (`arm64` Apple Silicon + `x86_64` Intel), sandboxing and hardened runtime entitlements, automated notarization ticket stapling, branded disk image (`.dmg`) generation using `create-dmg`, and continuous integration automation via GitHub Actions.

## What You Are Building
A production macOS desktop release framework in `macos/` and `scripts/`:
- `macos/Runner/Release.entitlements`: Hardened runtime security entitlements granting outbound network access (`com.apple.security.network.client`), keychain access, and JIT/unsigned memory flags required by the Flutter Dart AOT runtime.
- `macos/Runner/Info.plist`: macOS bundle configuration, bundle identifier (`com.kerberosec.growww.desktop`), high-DPI scaling, and URL scheme handlers (`growww://`).
- `scripts/build_macos_dmg.sh`: Comprehensive Bash script executing universal build compilation, codesign with Developer ID certificate and secure timestamping, Apple notarization submission, notarization stapling, and `.dmg` packaging.
- `assets/macos/dmg_background.png`: Retina graphical background for the installer DMG window with drag-and-drop pointer to `/Applications`.
- `.github/workflows/macos_release.yml`: GitHub Actions workflow on `macos-14` (M2 runner) automating the end-to-end build, codesign, notarization, and DMG release attachment.

## Scope Boundaries
- **In Scope:**
 - Flutter macOS universal binary build (`arm64` and `x86_64`).
 - Code signing with Apple Developer ID Application certificate.
 - Hardened runtime configuration and entitlements.
 - Cloud notarization using `xcrun notarytool` with App Store Connect credentials.
 - Ticket stapling using `xcrun stapler`.
 - Drag-and-drop installer `.dmg` creation with custom branding.
 - GitHub Actions CI/CD automation.
- **Out of Scope / Handled Elsewhere:**
 - iOS mobile build and TestFlight distribution (Prompt 517).
 - Windows MSIX packaging (Prompt 518).
 - Linux packaging (Prompt 519).

## Technology to Use
- **Primary Toolchain:** Flutter SDK 3.22+, Xcode 15.4+, Apple Developer ID Application Certificate, macOS SDK 14.0+, `create-dmg` v1.1+.
- **Justification:** Xcode 15 and `notarytool` provide deterministic Apple notarization in CI/CD without relying on legacy ALTOOL or deprecated Java tools, ensuring full compatibility with macOS Sonoma and macOS Sequoia Gatekeeper.
- **Dependencies & Tools:**
 - `create-dmg` (Homebrew)
 - `xcrun notarytool` & `xcrun stapler`
 - `codesign` utility
 - Fastlane Match (macOS certificates)

## Backend / Infra Touchpoints
- **Apple Notary Service:** REST API endpoint for automated binary upload, static security analysis, and notarization receipt generation.
- **S3 / CloudFront Release CDN:** Hosting signed and notarized `.dmg` installer files with SHA-256 checksums.
- **Sparkle Framework / GitHub Releases API (Prompt 525):** Desktop auto-update feed channel for macOS desktop clients.

## Blockchain Interaction
- **macOS Keychain Enclave Access:** Client utilizes the macOS native Data Protection Keychain (`SecItemAdd`, `SecItemCopyMatching`) to securely store cryptographic authorization tokens and on-chain identity assertions without exposing raw secrets to unauthorized desktop processes.

## Step-by-Step Build Instructions
1. Enable macOS desktop target in Flutter workspace via `flutter config --enable-macos-desktop`.
2. Configure `macos/Runner.xcodeproj` schemes and build settings for `Release` configuration targeting macOS 12.0+ deployment target.
3. Configure `macos/Runner/Release.entitlements` with required flags: `com.apple.security.network.client`, `com.apple.security.files.user-selected.read-write`, and `com.apple.security.cs.allow-jit`.
4. Configure Apple Developer ID Application certificate in macOS Keychain or decode dynamically from base64 in CI/CD runner.
5. Author build script `scripts/build_macos_dmg.sh` to orchestrate universal binary compilation: `flutter build macos --release`.
6. Apply recursive code signing on all embedded frameworks and native `.dylib` binaries using `codesign --force --verify --verbose --timestamp --options runtime --sign "Developer ID Application: KerberoSec Technologies India Pvt Ltd (TEAMID)"`.
7. Sign the outer `Growww.app` bundle with the Developer ID certificate and custom entitlements.
8. Package the signed application into a temporary ZIP archive: `ditto -c -k --keepParent build/macos/Build/Products/Release/Growww.app build/Growww.zip`.
9. Submit the archive to Apple Notary Service: `xcrun notarytool submit build/Growww.zip --key /path/to/AuthKey.p8 --key-id $KEY_ID --issuer $ISSUER_ID --wait`.
10. Verify notarization success; retrieve and log diagnostic audit log upon any rejection.
11. Staple the cryptographic notarization ticket to the application bundle: `xcrun stapler staple build/macos/Build/Products/Release/Growww.app`.
12. Generate branded installer DMG using `create-dmg` specifying window size (660x400), icon size (128px), and `/Applications` link coordinates.
13. Code-sign the final `.dmg` container with the Developer ID Application certificate and staple the notarization ticket to the `.dmg`.
14. Construct `.github/workflows/macos_release.yml` executing the full pipeline and verifying Gatekeeper compliance using `spctl --assess -vv --type install Growww.dmg`.
15. Test clean installation on Apple Silicon (M1/M2/M3) and Intel Macs with Gatekeeper enabled.

## Interfaces / Contracts

```xml
<!-- macos/Runner/Release.entitlements -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <false/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <false/>
</dict>
</plist>
```

```bash
#!/usr/bin/env bash
# scripts/build_macos_dmg.sh
set -euo pipefail

APP_NAME="Growww"
BUNDLE_ID="com.kerberosec.growww.desktop"
DEVELOPER_ID="Developer ID Application: KerberoSec Technologies India Pvt Ltd (${APPLE_TEAM_ID})"
APP_PATH="build/macos/Build/Products/Release/${APP_NAME}.app"
DMG_PATH="build/macos/Build/Products/Release/${APP_NAME}.dmg"

echo "==> Building Flutter macOS Release Binary..."
flutter build macos --release

echo "==> Code Signing Embedded Frameworks and App Bundle with Hardened Runtime..."
find "${APP_PATH}/Contents/Frameworks" -type f -name "*.dylib" -o -name "*.framework" | while read -r lib; do
    codesign --force --verify --verbose --timestamp --options runtime --sign "${DEVELOPER_ID}" "${lib}"
done

codesign --force --verify --verbose --timestamp --options runtime \
    --entitlements macos/Runner/Release.entitlements \
    --sign "${DEVELOPER_ID}" "${APP_PATH}"

echo "==> Submitting to Apple Notary Service..."
ditto -c -k --keepParent "${APP_PATH}" "${APP_NAME}_to_notarize.zip"
xcrun notarytool submit "${APP_NAME}_to_notarize.zip" \
    --key "${NOTARY_KEY_PATH}" \
    --key-id "${NOTARY_KEY_ID}" \
    --issuer "${NOTARY_ISSUER_ID}" \
    --wait

echo "==> Stapling Notarization Ticket to App..."
xcrun stapler staple "${APP_PATH}"

echo "==> Creating Branded DMG..."
create-dmg \
  --volname "${APP_NAME} Installer" \
  --volicon "assets/icons/macos/app.icns" \
  --background "assets/macos/dmg_background.png" \
  --window-pos 200 120 \
  --window-size 660 400 \
  --icon-size 128 \
  --icon "${APP_NAME}.app" 180 170 \
  --hide-extension "${APP_NAME}.app" \
  --app-drop-link 480 170 \
  "${DMG_PATH}" \
  "${APP_PATH}"

echo "==> Signing and Stapling Final DMG..."
codesign --force --verify --verbose --timestamp --sign "${DEVELOPER_ID}" "${DMG_PATH}"
xcrun stapler staple "${DMG_PATH}"

echo "==> Verifying Gatekeeper Assessment..."
spctl --assess -vv --type install "${DMG_PATH}"
echo "==> Successfully created signed and notarized ${DMG_PATH}"
```

## Security & Compliance Notes
- **Apple Hardened Runtime Enforcement:** Binary compilation strictly enables Hardened Runtime to prevent code injection, DLL hijacking, and memory tampering on macOS workstations.
- **Notarization & Gatekeeper Compliance:** Un-notarized applications trigger Gatekeeper blocking dialogs on macOS Sonoma/Sequoia. All distributed releases must pass automated notarization and have tickets stapled offline.
- **Keychain Access Control:** Ensure the macOS client registers explicit keychain access groups, preventing rogue desktop processes from querying Growww API session tokens or encrypted database keys.
- **Cryptographic Export Compliance:** Maintain App Store export compliance documentation confirming compliance with US BIS and Indian CERT-In guidelines regarding standard commercial cryptography.

## Acceptance Criteria
- [ ] `flutter build macos --release` compiles cleanly without architecture mismatch errors.
- [ ] Codesign verifies with zero warnings using `codesign -vvv --deep --strict`.
- [ ] Apple Notary Service returns `Status: Accepted` with zero security policy violations.
- [ ] `xcrun stapler staple` successfully embeds the ticket into both `.app` and `.dmg`.
- [ ] Final `.dmg` mounts with correct custom background graphics, icon placement, and `/Applications` link.
- [ ] `spctl --assess` confirms the DMG is accepted by Gatekeeper without developer warnings.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 108 (Environment Strategy).
- **Parallel Tasks:** Prompt 518 (Windows Packaging), Prompt 519 (Linux Packaging).
