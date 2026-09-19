# SRE Operational Runbook 038: Windows Native Thick Client MSIX Packaging & EV Code Signing Pipeline

**Runbook ID:** RUNBOOK-038-WINDOWS-SIGNING  
**Severity Classification:** Tier 1 (Desktop Distribution & SmartScreen Security)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** Windows Native Thick Client (`apps/growww_flutter/windows`)  
**Target Architectures:** Windows 10/11 x64, Windows on ARM64  
**Last Review:** September 2026  

---

## 1. Executive Summary

Windows SmartScreen blocks unverified binary execution unless an application is signed with an Extended Validation (EV) Code Signing Certificate backed by a FIPS 140-2 Level 2/3 hardware security token (HSM or Azure Key Vault).

This runbook defines the automated packaging, signing, and distribution process for both the Microsoft Store MSIX package and standalone enterprise `.exe` installers.

---

## 2. Automated Windows Build & Signing Architecture

```
+-------------------------------------------------------------------------------------------------------+
|                                  WINDOWS BUILD & SIGNING PIPELINE                                     |
|                                                                                                       |
|  [ GitHub Actions Windows-2022 Runner ]                                                               |
|          |                                                                                            |
|          +---> 1. Compile Rust Socket Core (`cargo build --release --target x86_64-pc-windows-msvc`)  |
|          +---> 2. Build Flutter Desktop Application (`flutter build windows --release`)              |
|          |                                                                                            |
|          v                                                                                            |
|  [ Azure Trusted Signing / Hardware HSM EV Signing ]                                                  |
|  - Sign native binaries: `rust_trading_core.dll`, `flutter_windows.dll`, `growww.exe`                |
|  - Tool: `AzureSignTool` / `SignTool.exe` with RFC 3161 Timestamp (`http://timestamp.digicert.com`)  |
|  - SHA-256 Digest Algorithm with Dual Signing Fallback                                                |
|          |                                                                                            |
|          v                                                                                            |
|  [ MSIX Package Creation & Manifest Generation ]                                                      |
|  - Generate `AppxManifest.xml` with restricted capabilities (`runFullTrust`, `broadFileSystemAccess`) |
|  - Bundle dependencies via `MakeAppx.exe pack`                                                        |
|  - Sign `.msix` container with EV Certificate                                                         |
|          |                                                                                            |
|          v                                                                                            |
|  [ Release Output: Microsoft Store Submission + Standalone Setup.exe (InnoSetup) ]                    |
+-------------------------------------------------------------------------------------------------------+
```

---

## 3. AppxManifest.xml Capabilities Specification

```xml
<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10"
         xmlns:uap="http://schemas.microsoft.com/appx/manifest/uap/windows10"
         xmlns:rescap="http://schemas.microsoft.com/appx/manifest/foundation/windows10/restrictedcapabilities">
  <Identity Name="Growww.NBSE.TradingTerminal"
            Publisher="CN=Growww Technologies Pvt Ltd, O=Growww, L=Bangalore, C=IN"
            Version="1.0.0.0"
            ProcessorArchitecture="x64" />
  
  <Properties>
    <DisplayName>Growww Pro Trading Terminal</DisplayName>
    <PublisherDisplayName>Growww National Blockchain Stock Exchange</PublisherDisplayName>
    <Logo>assets\app_icon_150.png</Logo>
  </Properties>

  <Dependencies>
    <TargetDeviceFamily Name="Windows.Desktop" MinVersion="10.0.19041.0" MaxVersionTested="10.0.22631.0" />
  </Dependencies>

  <Capabilities>
    <rescap:Capability Name="runFullTrust" />
    <Capability Name="internetClient" />
    <Capability Name="internetClientServer" />
    <Capability Name="privateNetworkClientServer" />
  </Capabilities>
</Package>
```

---

## 4. Verification Protocol

1. Verify Authenticode Signature:
   ```cmd
   SignTool.exe verify /pa /v "build\windows\runner\Release\growww.exe"
   ```
2. Verify SmartScreen Trust:
   - Launch executable on a clean Windows 11 installation with SmartScreen enabled.
   - Asserts: Zero SmartScreen warning prompts; app launches directly into Obsidian splash screen.
3. Verify Global Hotkey Registration:
   - Launch application, focus background browser, press `Ctrl+Alt+Space`.
   - Asserts: System tray flashes green and order cancellation toast is delivered.
4. Verify Zero-Fee Invariant:
   - Inspect order ticket header: `0.00% Zero-Fee (₹0.00 Comm)`.
