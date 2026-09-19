# 518 - Platform-Specific Packaging: Windows (MSIX & Microsoft Store) Build Pipeline

## Purpose
High-frequency traders, family offices, and institutional desk operators require high-performance desktop terminals with multi-monitor support, advanced charting, and robust desktop OS integration. On Windows 10 and 11, the modern, secure distribution format is MSIX (Microsoft Installer / AppX successor), which provides declarative packaging, clean installation/uninstallation, AppContainer isolation, and integration with the Microsoft Store and WinGet package manager.

This prompt defines the Windows desktop build and MSIX packaging pipeline for the Growww Flutter client. It configures the native Visual Studio C++ CMake build environment, integrates the `msix` packaging toolchain, establishes digital code signing with EV/OV certificates, configures Windows DPAPI local security, implements `.appinstaller` continuous auto-updates, and automates CI/CD releases on GitHub Actions Windows runners.

## What You Are Building
A production Windows desktop release and MSIX packaging pipeline in `windows/` and `pubspec.yaml`:
- `windows/runner/main.cpp` & `CMakeLists.txt`: C++ Flutter desktop host configured for minimum window constraints (e.g. 1024x768), DPI awareness (Per-Monitor V2), and single-instance process mutual exclusion (Mutex).
- `pubspec.yaml` (`msix_config` block): Comprehensive MSIX metadata, publisher identities, capabilities (`internetClient`), protocol associations (`growww://`), and icon assets.
- `windows/packaging/AppInstaller.xml`: Declarative XML file enabling automatic background update checks from an Azure CDN / S3 bucket without requiring Microsoft Store hosting.
- `scripts/build_windows_msix.ps1`: Automated PowerShell build script executing Flutter compilation, C++ native artifact linking, MSIX bundling, and `SignTool.exe` digital signing.
- `.github/workflows/windows_release.yml`: Continuous deployment workflow on `windows-2022` runners building signed `.msix` and generating WinGet manifest submissions.

## Scope Boundaries
- **In Scope:**
 - Flutter Windows native C++ runner setup and CMake compilation.
 - MSIX package generation using `msix` package and Windows SDK tools (`MakeAppx.exe`).
 - Code signing with digital certificates using `SignTool.exe` with RFC 3161 timestamping.
 - Sideloading and Microsoft Store packaging profiles.
 - Window manager customization (title bar styling, min/max dimensions) and single-instance locks.
 - Auto-update distribution via `.appinstaller` and WinGet manifests.
- **Out of Scope / Handled Elsewhere:**
 - Linux desktop packaging (Prompt 519).
 - macOS desktop packaging (Prompt 520).
 - Web application deployment (Prompt 601).

## Technology to Use
- **Primary Toolchain:** Flutter SDK 3.22+, Visual Studio 2022 (with "Desktop development with C++" workload), CMake 3.28+, Windows 11 SDK (10.0.22621.0).
- **Justification:** Flutter Windows compiles directly to native x64/ARM64 machine code leveraging DirectX/ANGLE rendering, delivering 120 FPS desktop charting with zero web-engine overhead.
- **Dependencies & Tools:**
 - `msix: ^3.16.7` (Dart build package)
 - `window_manager: ^0.3.9` (desktop window resizing and frame control)
 - `bitsdojo_window: ^0.1.6` (custom title bar styling)
 - Windows SDK `signtool.exe` and `MakeAppx.exe`

## Backend / Infra Touchpoints
- **Windows Update CDN / S3 Bucket:** Static hosting of `Growww.msix` and `Growww.appinstaller` with Cache-Control headers for continuous client auto-updating.
- **Microsoft Partner Center:** API integration for automated Microsoft Store app ingestion (optional corporate track).
- **WinGet Community Repository:** Automated pull request submission for `winget install KerberoSec.Growww`.

## Blockchain Interaction
- **Desktop RPC & Node Verification:** The Windows client interfaces directly with permissioned Hyperledger Besu validator nodes via TLS 1.3 JSON-RPC / WebSocket endpoints for ultra-low latency L2 order book streaming and DvP settlement hash inspection.
- **Local Keystore Protection:** Windows client leverages Windows Data Protection API (DPAPI / CNG Cryptography Next Generation) to encrypt on-chain session tokens and locally cached Merkle proof receipts.

## Step-by-Step Build Instructions
1. In `apps/growww_flutter/`, ensure Flutter Windows desktop support is enabled via `flutter config --enable-windows-desktop`.
2. Inspect `windows/runner/CMakeLists.txt` and verify target architecture is configured for `x64` with modern C++20 standard flags.
3. Modify `windows/runner/main.cpp` to implement a Windows Named Mutex (`CreateMutexW`) preventing accidental concurrent execution of multiple trading terminal instances.
4. Integrate `window_manager` in Dart entrypoint (`lib/main_desktop.dart`) to enforce minimum window dimensions (1280x800) and preserve window position across app restarts.
5. Create high-resolution Windows asset icons in `windows/runner/resources/` (Square44x44Logo, Square150x150Logo, Wide310x150Logo, StoreLogo).
6. Configure the `msix_config` section in `pubspec.yaml` with package name, display name, publisher ID, identity name, protocol schemas, and store capabilities.
7. Generate a code signing PFX certificate for staging/internal testing or configure EV Hardware Token signing via CI secrets.
8. Create PowerShell build script `scripts/build_windows_msix.ps1` that runs `flutter build windows --release` followed by `dart run msix:create`.
9. Add code signing step to `scripts/build_windows_msix.ps1` utilizing `signtool.exe sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f certificate.pfx`.
10. Author `windows/packaging/Growww.appinstaller` XML specifying the remote URL where newer `.msix` releases are polled upon app launch.
11. Build GitHub Actions workflow `.github/workflows/windows_release.yml` running on `windows-latest`.
12. In the GitHub Actions workflow, restore Visual Studio build cache, run Flutter build, invoke MSIX generation, execute SignTool signing, and publish the artifact.
13. Generate WinGet package manifest (`KerberoSec.Growww.yaml`, `KerberoSec.Growww.installer.yaml`) with accurate SHA-256 installer hashes.
14. Test clean installation, update via AppInstaller, and clean uninstallation on a clean Windows 11 VM sandbox.

## Interfaces / Contracts

```yaml
# pubspec.yaml (MSIX Configuration)
msix_config:
  display_name: Growww Terminal
  publisher_display_name: KerberoSec Technologies India Pvt Ltd
  identity_name: KerberoSec.GrowwwTerminal
  publisher: CN=KerberoSec Technologies, O=KerberoSec, L=Mumbai, S=Maharashtra, C=IN
  msix_version: 1.0.0.0
  logo_path: assets/icons/windows/StoreLogo.png
  start_menu_icon_path: assets/icons/windows/Square44x44Logo.png
  tile_icon_path: assets/icons/windows/Square150x150Logo.png
  capabilities:
 - internetClient
  protocol_activations:
 - growww
  execution_alias: growww
  install_sub_folder: Growww
  architecture: x64
```

```cpp
// windows/runner/main.cpp (Single Instance Mutex Snippet)
#include <windows.h>
#include <flutter/dart_project.h>
#include <flutter/flutter_view_controller.h>
#include "flutter_window.h"
#include "utils.h"

int APIENTRY wWinMain(_In_ HINSTANCE hInstance, _In_opt_ HINSTANCE hPrevInstance,
                      _In_ LPWSTR lpCmdLine, _In_ int nCmdShow) {
    // Single instance mutex to prevent duplicate trading sessions
    HANDLE hMutex = CreateMutexW(NULL, TRUE, L"Global\\GrowwwTradingTerminalSingleInstanceMutex");
    if (GetLastError() == ERROR_ALREADY_EXISTS) {
        HWND hWnd = FindWindowW(NULL, L"Growww Terminal");
        if (hWnd != NULL) {
            ShowWindow(hWnd, SW_RESTORE);
            SetForegroundWindow(hWnd);
        }
        return 0;
    }

    if (!::AttachConsole(ATTACH_PARENT_PROCESS) && ::IsDebuggerPresent()) {
        CreateAndAttachConsole();
    }

    ::CoInitializeEx(nullptr, COINIT_APARTMENTTHREADED);

    flutter::DartProject project(L"data");
    FlutterWindow window(project);
    Win32Window::Point origin(10, 10);
    Win32Window::Size size(1440, 900);
    if (!window.Create(L"Growww Terminal", origin, size)) {
        return 0;
    }
    window.SetQuitOnClose(true);

    ::MSG msg;
    while (::GetMessage(&msg, nullptr, 0, 0)) {
        ::TranslateMessage(&msg);
        ::DispatchMessage(&msg);
    }

    ::CoUninitialize();
    if (hMutex) ReleaseMutex(hMutex);
    return 0;
}
```

## Security & Compliance Notes
- **Windows AppContainer Isolation:** MSIX packages run within lightweight AppContainer security boundaries, preventing unauthorized file system tampering and isolating application registry entries.
- **SmartScreen Reputation & EV Signing:** Production MSIX binaries must be signed using an Extended Validation (EV) Code Signing certificate with RFC 3161 timestamping to prevent Windows Defender SmartScreen untrusted binary warnings.
- **Memory & Process Integrity:** Enforce Data Execution Prevention (DEP), Address Space Layout Randomization (ASLR / High Entropy VA), and Control Flow Guard (CFG) in Visual Studio C++ linker flags.
- **Anti-Tamper & Credential Isolation:** Local API tokens and cached credentials must be stored via Windows DPAPI (`CryptProtectData`) tied strictly to the current user login session (Prompt 521).

## Acceptance Criteria
- [ ] `flutter build windows --release` and `dart run msix:create` produce a valid `.msix` bundle without errors.
- [ ] Output `.msix` installs cleanly on Windows 10 (21H2+) and Windows 11 without administrative elevation.
- [ ] Windows Defender SmartScreen recognizes valid code signature and does not block installation.
- [ ] Launching a second instance of the app automatically focuses the existing running window and exits cleanly.
- [ ] Deep links via `growww://` protocol launch or focus the desktop client and route to the specified security view.
- [ ] Windows DPAPI successfully secures persistent login state across app restarts.
- [ ] AppInstaller XML successfully identifies and prompts for updates when a new version is published to the CDN.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 108 (Environment Strategy).
- **Parallel Tasks:** Prompt 519 (Linux Packaging), Prompt 520 (macOS Packaging).
