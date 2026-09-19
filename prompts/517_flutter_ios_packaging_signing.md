# 517 - Platform-Specific Packaging: iOS (App Store & TestFlight) Build & Signing Pipeline

## Purpose
The iOS client represents a critical tier for premium domestic and GIFT City international investors. Releasing an enterprise-grade fintech app on Apple iOS requires adherence to strict Apple Human Interface Guidelines, App Store Review Guidelines (Guideline 5.1.1 Data Privacy, Guideline 2.5.4 Cryptography), and automated code signing across developer teams and continuous integration runners.

This prompt specifies the complete iOS build, code signing, and release pipeline for the Growww Flutter client. It configures Xcode schemes and build configurations, implements Fastlane Match for deterministic Git-based certificate and provisioning profile management, configures security entitlements (Keychain Access Groups, Face ID, Associated Domains), and automates TestFlight and App Store distribution via GitHub Actions macOS runners.

## What You Are Building
A production iOS build and distribution framework located in `ios/` and `fastlane/`:
- `ios/Runner.xcodeproj` & `ios/Runner.xcworkspace`: Configured with multi-flavor schemes (`Runner-dev`, `Runner-staging`, `Runner-prod`) and custom xcconfig build settings.
- `ios/Runner/Info.plist`: Compliant privacy strings (`NSFaceIDUsageDescription`, `NSCameraUsageDescription`), ATS network configurations, and deep linking URL schemes.
- `ios/Runner/Runner.entitlements`: Secure entitlements enabling Keychain Sharing, Associated Domains (Universal Links), Push Notifications, and App Groups.
- `ios/ExportOptions.plist`: Automated App Store export configuration specifying bitcode stripping, symbol inclusion, and manual/automatic signing mappings.
- `fastlane/Fastfile` (iOS lane) & `fastlane/Matchfile`: Automated pipeline for certificate synchronization (Fastlane Match with encrypted repository storage), Xcode archiving, `.ipa` generation, and TestFlight / App Store submission.
- `.github/workflows/ios_release.yml`: GitHub Actions macOS CI/CD workflow executing CocoaPods dependency resolution, Flutter release compilation, code signing with ephemeral keychains, and automated App Store Connect upload.

## Scope Boundaries
- **In Scope:**
 - Xcode project schemes, xcconfig configuration files, and target build settings.
 - Code signing automation using Fastlane Match (Apple Distribution Certificates & App Store Provisioning Profiles).
 - Entitlements for FaceID, Keychain, Push Notifications, and Universal Links.
 - Fastlane automation for TestFlight and App Store Connect delivery.
 - dSYM symbol generation, extraction, and automated crash reporting upload.
- **Out of Scope / Handled Elsewhere:**
 - Android packaging and signing (Prompt 516).
 - macOS desktop packaging and notarization (Prompt 520).
 - App-level Dart screens and logic (Prompts 501-515).

## Technology to Use
- **Primary Toolchain:** Flutter SDK 3.22+, Xcode 15.4+, CocoaPods 1.15+, Ruby 3.2+ with Fastlane 2.220+.
- **Justification:** Xcode 15 and Fastlane Match eliminate manual code signing drift, provide reproducible builds in ephemeral CI/CD environments, and support the App Store Connect API v2.
- **Dependencies & Tools:**
 - `fastlane` (Match, Gym, Pilot, Deliver)
 - `local_auth_ios`
 - Apple Developer Program App Store Connect API Key (`.p8`)
 - CocoaPods package manager

## Backend / Infra Touchpoints
- **Apple App Store Connect API:** Automated authentication for TestFlight beta groups, build processing, and App Store submission using JSON Web Tokens (JWT).
- **Encrypted Git / AWS S3 Match Storage:** Secure encrypted storage for distribution certificates and mobile provisioning profiles.
- **Sentry / Firebase dSYM Uploader (Prompt 524):** Automated extraction and upload of debug symbol files (`.dSYM.zip`).

## Blockchain Interaction
- **Apple Cryptography Declaration:** App Store export compliance compliance documentation declaring standard non-exempt symmetric/asymmetric encryption (AES-256, Keccak-256, secp256k1) used for local secure enclave operations and TLS communication with the Hyperledger Besu permissioned network.

## Step-by-Step Build Instructions
1. Open `apps/growww_flutter/ios/` and establish configuration files: `Flutter/Dev.xcconfig`, `Flutter/Staging.xcconfig`, and `Flutter/Prod.xcconfig`.
2. Configure Xcode project configurations (`Debug-prod`, `Profile-prod`, `Release-prod`) mapping to the respective bundle identifiers (`com.kerberosec.growww`, `com.kerberosec.growww.staging`, `com.kerberosec.growww.dev`).
3. Set up Xcode Schemes (`Growww Prod`, `Growww Staging`, `Growww Dev`) with correct build configurations and executable targets.
4. Populate `ios/Runner/Info.plist` with required localized privacy usage keys: `NSFaceIDUsageDescription` ("Growww requires Face ID to securely authenticate trade executions and wallet transactions").
5. Create `ios/Runner/Runner.entitlements` configuring `com.apple.developer.associated-domains` (`applinks:growww.in`), `keychain-access-groups` (`$(AppIdentifierPrefix)com.kerberosec.growww`), and `aps-environment`.
6. Initialize Fastlane Match by creating `fastlane/Matchfile` linked to an encrypted private repository with passphrase encryption.
7. Generate App Store Connect API Key (`.p8`), Key ID, and Issuer ID in Apple Developer Portal for headless CI authentication.
8. Author `fastlane/Fastfile` with lanes: `sync_signing` (running `match(type: "appstore", readonly: true)`), `build_release` (running `gym`), and `upload_testflight` (running `pilot`).
9. Create `ios/ExportOptions.plist` specifying `method: app-store`, `uploadBitcode: false`, and `uploadSymbols: true`.
10. Construct GitHub Actions workflow `.github/workflows/ios_release.yml` using `macos-14` (Apple Silicon M2) runners.
11. In CI workflow, configure steps to create an ephemeral macOS keychain, import Match certificates, install CocoaPods dependencies, and execute `flutter build ipa --flavor prod --release`.
12. Add Fastlane `pilot` step to distribute generated `.ipa` to the Internal TestFlight QA team with dynamic release notes.
13. Configure automatic extraction of `.dSYM` archives and dispatch to Sentry crash monitoring.
14. Validate local builds using real physical iPhone test devices running iOS 16 and iOS 17.

## Interfaces / Contracts

```ruby
# fastlane/Fastfile
default_platform(:ios)

platform :ios do
  desc "Build and upload production IPA to TestFlight"
  lane :beta do
    api_key = app_store_connect_api_key(
      key_id: ENV["ASC_KEY_ID"],
      issuer_id: ENV["ASC_ISSUER_ID"],
      key_content: ENV["ASC_KEY_CONTENT_BASE64"],
      is_key_content_base64: true,
      in_house: false
    )

    # Ephemeral keychain for CI
    if is_ci
      create_keychain(
        name: "ci_keychain",
        password: ENV["CI_KEYCHAIN_PASSWORD"],
        default_keychain: true,
        unlock: true,
        timeout: 3600
      )
    end

    # Sync App Store certificates & profiles
    match(
      type: "appstore",
      app_identifier: ["com.kerberosec.growww"],
      readonly: is_ci,
      keychain_name: is_ci ? "ci_keychain" : nil,
      keychain_password: is_ci ? ENV["CI_KEYCHAIN_PASSWORD"] : nil
    )

    # Build IPA using gym
    build_app(
      workspace: "ios/Runner.xcworkspace",
      scheme: "Growww Prod",
      configuration: "Release-prod",
      export_method: "app-store",
      export_options: "ios/ExportOptions.plist",
      output_directory: "build/ios/ipa",
      output_name: "Growww.ipa"
    )

    # Upload to TestFlight
    upload_to_testflight(
      api_key: api_key,
      skip_waiting_for_build_processing: true,
      distribute_external: false
    )
  end
end
```

```xml
<!-- ios/Runner/Runner.entitlements -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.developer.associated-domains</key>
    <array>
        <string>applinks:growww.in</string>
        <string>applinks:app.growww.in</string>
    </array>
    <key>keychain-access-groups</key>
    <array>
        <string>$(AppIdentifierPrefix)com.kerberosec.growww</string>
    </array>
    <key>aps-environment</key>
    <string>production</string>
</dict>
</plist>
```

## Security & Compliance Notes
- **Hardware Enclave & Keychain Isolation:** Configure Keychain sharing access groups such that only binaries signed with the official team certificate can read stored JWT tokens and encryption seeds.
- **Ephemeral CI Keychains:** On shared or cloud CI runners (GitHub Actions), keychains must be created dynamically with random temporary passwords and explicitly destroyed at the end of the build job.
- **ATS (App Transport Security):** Enforce strict ATS without arbitrary load exceptions; all network calls must use HTTPS with TLS 1.3 and forward secrecy.
- **Data Protection Class:** Apply `NSFileProtectionComplete` to all local application container files to ensure filesystem encryption when the device is locked.

## Acceptance Criteria
- [ ] `flutter build ipa --flavor prod --release` executes without error on Xcode 15+.
- [ ] Fastlane Match downloads and installs valid provisioning profiles and distribution certificates in headless CI.
- [ ] Generated `.ipa` is successfully uploaded to App Store Connect and processed by Apple without rejection warnings.
- [ ] FaceID authentication functions seamlessly on physical iOS hardware with correct localized permission prompts.
- [ ] Universal links (`https://growww.in/trade/INE...`) open the app directly from Safari and Messages.
- [ ] Crashlytics/Sentry dSYM files are generated and uploaded automatically.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 108 (Environment Strategy).
- **Parallel Tasks:** Prompt 516 (Android Packaging & Signing), Prompt 526 (Deep Linking & Universal Links).
