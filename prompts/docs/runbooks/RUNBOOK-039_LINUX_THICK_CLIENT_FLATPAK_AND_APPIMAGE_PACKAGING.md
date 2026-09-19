# SRE Operational Runbook 039: Linux Native Thick Client Flatpak, AppImage & Debian Packaging

**Runbook ID:** RUNBOOK-039-LINUX-PACKAGING  
**Severity Classification:** Tier 2 (Open Source & Developer Distribution)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** Linux Native Thick Client (`apps/growww_flutter/linux`)  
**Target Distributions:** Ubuntu 22.04 / 24.04 LTS, Fedora 39/40, Arch Linux, Debian 12  
**Target Formats:** Flatpak (Flathub), Standalone AppImage, Debian Package (`.deb`)  
**Last Review:** September 2026  

---

## 1. Executive Summary

Linux institutional algorithmic traders, hedge fund developers, and automated trading desks demand zero-friction deployment with reliable display server integration across both X11 and Wayland compositors.

This runbook outlines the build, sandboxing, and distribution pipelines for Flatpak, AppImage, and native Debian packaging formats.

---

## 2. Flatpak Manifest Specification (`org.growww.GrowwwTradingTerminal.yaml`)

Flatpak provides sandboxed isolation while granting explicit hardware access to GPU acceleration (Vulkan/DRI), Wayland sockets, and encrypted credential storage:

```yaml
app-id: org.growww.GrowwwTradingTerminal
runtime: org.freedesktop.Platform
runtime-version: '23.08'
sdk: org.freedesktop.Sdk
command: growww

finish-args:
  # Display server access (Wayland with X11 fallback)
  - --socket=wayland
  - --socket=fallback-x11
  # Hardware GPU acceleration (Vulkan & OpenGL)
  - --device=dri
  # Network connectivity for WebSocket streaming & Hyperledger Besu
  - --share=network
  # Inter-Process Communication (IPC) for multi-window memory sharing
  - --share=ipc
  # Linux Secret Service API access for encrypted key vault
  - --talk-name=org.freedesktop.secrets
  # File access restricted strictly to user downloads and logs
  - --filesystem=xdg-download
```

---

## 3. Standalone AppImage Packaging Protocol

AppImage delivers a single, self-contained binary running on any modern Linux distribution without installation:
1. Bundle Flutter Linux executable (`growww`), shared libraries (`libflutter_linux_glfw.so`, `librust_trading_core.so`), and asset bundles into `AppDir`.
2. Generate `AppRun` trampoline script verifying Vulkan driver availability (`vulkaninfo`) and falling back to software rasterization if hardware GPU is absent.
3. Package with `appimagetool` and sign with GPG release key.

---

## 4. Verification Protocol

1. Test Wayland native execution:
   ```bash
   WAYLAND_DISPLAY=wayland-0 ./Growww-x86_64.AppImage --enable-features=UseOzonePlatform --ozone-platform=wayland
   ```
2. Verify Vulkan hardware pipeline:
   - Check application log: `[INFO] Vulkan Impeller backend initialized successfully (Vulkan 1.3).`
3. Verify Secret Service keyring integration:
   - Save session token, verify entry exists in GNOME Keyring / KWallet via `secret-tool lookup service growww`.
4. Verify Zero-Fee Invariant:
   - Launch application, assert trading order form renders `0.00% Zero-Fee (₹0.00 Comm)` and 0 Gas sponsored paymaster badge.
