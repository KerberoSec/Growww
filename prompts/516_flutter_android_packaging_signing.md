# 516 - Platform-Specific Packaging: Android (Google Play Store) Build & Signing Pipeline

## Purpose
Android represents the dominant retail investment client platform in the Indian financial market. Meeting strict Google Play Store security standards and SEBI digital cybersecurity guidelines requires an automated, reproducible, and tamper-resistant build and release pipeline. 

This prompt defines the production Android packaging specification for the Growww Flutter client. It establishes Gradle Kotlin DSL build scripts, Android App Bundle (.aab) generation, Google Play App Signing with secure upload keystores, ProGuard/R8 code shrinking and symbol obfuscation, native ABI splitting, and a fully automated GitHub Actions CI/CD pipeline ensuring rapid, deterministic, and cryptographically verified releases.

## What You Are Building
A production Android build and release infrastructure in `android/` and `.github/workflows/`:
- `android/app/build.gradle.kts`: Gradle Kotlin DSL configuration with flavors (`dev`, `staging`, `prod`), version codes, signing configs, and R8 optimization rules.
- `android/app/proguard-rules.pro`: Hardened ProGuard/R8 rules preserving Flutter engine symbols, SQLCipher native libraries, cryptographic providers, and serialization DTOs while stripping unused bytecode and obfuscating proprietary trade execution logic.
- `android/key.properties.example` & Secrets Management: Template and CI/CD secret injection mechanism for upload keystore credentials.
- `fastlane/Fastfile` (Android lane): Automated Fastlane pipeline executing unit tests, building signed App Bundles (`.aab`), uploading mapping files, and deploying to Google Play Internal Testing / Production tracks.
- `.github/workflows/android_release.yml`: Continuous deployment workflow orchestrating environment validation, Flutter dependency resolution, Gradle compilation, artifact signing, and automated Play Console dispatch.

## Scope Boundaries
- **In Scope:**
 - Gradle Kotlin DSL configuration for multi-flavor Android builds.
 - Upload keystore generation and cryptographic signature automation.
 - ProGuard/R8 obfuscation, resource shrinking, and native `.so` library stripping.
 - Target SDK 34+ (Android 14) compliance, 64-bit ABI support (`arm64-v8a`, `armeabi-v7a`, `x86_64`).
 - Fastlane and GitHub Actions automation for Google Play Store track deployment.
- **Out of Scope / Handled Elsewhere:**
 - iOS packaging and App Store deployment (Prompt 517).
 - Desktop platform packaging (Prompts 518, 519, 520).
 - App-level Dart business logic and UI screens (Prompts 501-515).

## Technology to Use
- **Primary Toolchain:** Flutter SDK 3.22+, Android Gradle Plugin (AGP 8.4+), Gradle 8.6+, Kotlin 1.9.24+, Java 17 (OpenJDK).
- **Justification:** AGP 8.4+ with Gradle Kotlin DSL provides type-safe build scripts, incremental compilation, advanced R8 optimizations, and native support for modern Android 14 requirements.
- **Dependencies & Tools:**
 - Fastlane (Ruby 3.2+) with `fastlane-plugin-google_play_track_version_codes`
 - Google Play Developer API (Service Account JSON)
 - Android SDK Build-Tools 34.0.0
 - JKS / PKCS12 Keystore toolchain

## Backend / Infra Touchpoints
- **Google Play Developer Console:** Service account authentication for publishing to `internal`, `alpha`, `beta`, and `production` tracks.
- **Google Cloud KMS / GitHub Encrypted Secrets:** Storing `ANDROID_UPLOAD_KEYSTORE_BASE64`, `KEYSTORE_PASSWORD`, `KEY_ALIAS`, and `KEY_PASSWORD`.
- **Sentry / Crashlytics Symbol Upload (Prompt 524):** Automated upload of R8 de-obfuscation mapping files (`mapping.txt`) and native debug symbols (`libapp.so`).

## Blockchain Interaction
- **Cryptographic Library Preservation:** Ensures that native cryptographic bindings utilized for blockchain signature verification, Merkle proof hashing (Keccak-256 / SHA-256), and secure enclave communication are preserved and not mangled by R8 bytecode shrinking.
- **Ledger Verification Compatibility:** Verifies that network security configuration enforces strict TLS 1.3 for API endpoints communicating with Hyperledger Besu JSON-RPC relayers.

## Step-by-Step Build Instructions
1. Navigate to `apps/growww_flutter/android/` and convert `app/build.gradle` to Gradle Kotlin DSL `app/build.gradle.kts`.
2. Configure `compileSdk = 34`, `minSdk = 24` (Android 7.0+), and `targetSdk = 34` in `defaultConfig`.
3. Set up product flavors in `build.gradle.kts`: `dev`, `staging`, and `prod` with distinct `applicationIdSuffix` and app names.
4. Generate a 4096-bit RSA production upload keystore using `keytool` with PKCS12 format.
5. Create `android/key.properties` loading logic that dynamically reads from environment variables in CI/CD or local secure files in developer environments.
6. Configure `signingConfigs` block linking release builds to the upload keystore credentials.
7. Enable `isMinifyEnabled = true`, `isShrinkResources = true`, and configure `proguardFiles` for release build types.
8. Author comprehensive `proguard-rules.pro` containing rules for Flutter, Riverpod, Drift SQLCipher, `local_auth`, and network JSON serializers.
9. Configure `ndk` ABI filters to target `armeabi-v7a`, `arm64-v8a`, and `x86_64`, ensuring 64-bit compliance required by Google Play.
10. Set up `network_security_config.xml` in `res/xml/` enforcing HTTPS domain whitelisting and certificate pinning.
11. Configure Fastlane inside `android/fastlane/` with lanes for `build_aab`, `deploy_internal`, and `deploy_production`.
12. Create GitHub Actions workflow `.github/workflows/android_release.yml` with secrets decoding, Flutter build execution (`flutter build appbundle --flavor prod --release`), and Fastlane deployment.
13. Implement automated native symbol and ProGuard `mapping.txt` archiving to Sentry / Google Play Console.
14. Perform local release verification using `bundletool` to generate and test device-specific APKs on physical Android devices.

## Interfaces / Contracts

```kotlin
// android/app/build.gradle.kts
plugins {
    id("com.android.application")
    id("kotlin-android")
    id("dev.flutter.flutter-gradle-plugin")
}

val keystoreProperties = Properties()
val keystorePropertiesFile = rootProject.file("key.properties")
if (keystorePropertiesFile.exists()) {
    keystoreProperties.load(FileInputStream(keystorePropertiesFile))
}

android {
    namespace = "com.kerberosec.growww"
    compileSdk = 34
    ndkVersion = "26.1.10909125"

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    defaultConfig {
        applicationId = "com.kerberosec.growww"
        minSdk = 24
        targetSdk = 34
        versionCode = flutter.versionCode
        versionName = flutter.versionName
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        ndk {
            abiFilters.addAll(listOf("armeabi-v7a", "arm64-v8a", "x86_64"))
        }
    }

    signingConfigs {
        create("release") {
            keyAlias = System.getenv("ANDROID_KEY_ALIAS") ?: keystoreProperties.getProperty("keyAlias")
            keyPassword = System.getenv("ANDROID_KEY_PASSWORD") ?: keystoreProperties.getProperty("keyPassword")
            storeFile = if (System.getenv("ANDROID_KEYSTORE_PATH") != null) {
                file(System.getenv("ANDROID_KEYSTORE_PATH")!)
            } else {
                file(keystoreProperties.getProperty("storeFile") ?: "upload-keystore.jks")
            }
            storePassword = System.getenv("ANDROID_STORE_PASSWORD") ?: keystoreProperties.getProperty("storePassword")
        }
    }

    buildTypes {
        release {
            signingConfig = signingConfigs.getByName("release")
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    flavorDimensions += "environment"
    productFlavors {
        create("dev") {
            dimension = "environment"
            applicationIdSuffix = ".dev"
            resValue("string", "app_name", "Growww Dev")
        }
        create("staging") {
            dimension = "environment"
            applicationIdSuffix = ".staging"
            resValue("string", "app_name", "Growww Staging")
        }
        create("prod") {
            dimension = "environment"
            resValue("string", "app_name", "Growww")
        }
    }
}
```

```pro
# android/app/proguard-rules.pro
# Keep Flutter wrapper & engine classes
-keep class io.flutter.app.** { *; }
-keep class io.flutter.plugin.**  { *; }
-keep class io.flutter.util.**  { *; }
-keep class io.flutter.view.**  { *; }
-keep class io.flutter.**  { *; }
-keep class io.flutter.plugins.**  { *; }

# Keep SQLCipher and security native libs
-keep class net.sqlcipher.** { *; }
-dontwarn net.sqlcipher.**
-keep class androidx.biometric.** { *; }
-keep class androidx.security.crypto.** { *; }

# Strip logging in production builds
-assumenosideeffects class android.util.Log {
    public static boolean isLoggable(java.lang.String, int);
    public static int v(...);
    public static int d(...);
}
```

## Security & Compliance Notes
- **Play App Signing Integrity:** Use Google Play App Signing so the upload key is solely used to authenticate builds to Google, while Google signs the final distributed binaries with the master app signing key.
- **Keystore Isolation:** Never commit `upload-keystore.jks` or `key.properties` to source control. In CI/CD, generate the keystore file dynamically from base64 GitHub Secrets into a transient runner path and wipe upon pipeline completion.
- **SEBI Mobile Security Directives:** Android builds must strictly disallow `android:allowBackup="false"`, enforce hardware-backed biometric authentication (`androidx.biometric`), and enable `FLAG_SECURE` on window managers during sensitive transaction flows to prevent OS-level screen capture.
- **Root & Tamper Detection:** Release builds include Play Integrity API integration checks to verify device authenticity before allowing financial order routing.

## Acceptance Criteria
- [ ] `flutter build appbundle --flavor prod --release` completes successfully with zero compilation or R8 errors.
- [ ] Output `.aab` file size is optimized (<25 MB compressed download size).
- [ ] App Bundle installs cleanly via `bundletool` onto physical Android 7.0 through Android 14 test devices.
- [ ] ProGuard/R8 obfuscates application classes while preserving runtime serialization, database operations, and biometric authentication.
- [ ] Fastlane script successfully authenticates via Google Play Developer Service Account and deploys `.aab` to the Internal Testing track.
- [ ] `mapping.txt` and native debug symbols are generated and archived as CI build artifacts.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 108 (Environment Strategy).
- **Parallel Tasks:** Prompt 517 (iOS Packaging & Signing), Prompt 524 (Crash Reporting & Analytics).
