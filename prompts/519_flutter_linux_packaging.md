# 519 - Platform-Specific Packaging: Linux (Flatpak, Snap & AppImage) Build Pipeline

## Purpose
Institutional quantitative trading desks, fintech engineering workstations, and open-source financial operators predominantly run Linux distributions (Ubuntu, Debian, Fedora, Arch Linux). Packaging a native Flutter desktop application on Linux requires handling heterogeneous display servers (X11 and Wayland), diverse packaging ecosystems, shared C/C++ runtime library dependencies (GTK 3, GLib, Libsecret), and sandboxed execution environments.

This prompt specifies the universal Linux build and packaging pipeline for the Growww desktop client. It defines configurations for three standard Linux distribution formats: **Flatpak** (Flathub standard), **Snap** (Ubuntu Snapcraft standard), and standalone **AppImage** (universal binary). It ensures native Wayland/X11 rendering, D-Bus Secret Service integration for token security, and automated CI/CD builds on Ubuntu runners.

## What You Are Building
A complete multi-format Linux desktop build and release suite in `linux/packaging/`:
- `linux/packaging/flatpak/org.kerberosec.growww.yml`: Declarative Flatpak build manifest defining GNOME 46 runtime dependencies, network sandbox holes, and desktop metadata.
- `linux/packaging/snap/snapcraft.yaml`: Canonical Snapcraft definition with strict confinement, plugs (`network`, `password-manager-service`, `x11`, `wayland`, `desktop`), and Flutter extension integration.
- `linux/packaging/appimage/AppRun` & Build Script: Shell automation creating self-contained AppDir directories and invoking `appimagetool` to produce portable `.AppImage` binaries.
- `linux/packaging/org.kerberosec.growww.desktop`: FreeDesktop standard desktop entry file with categories (`Office;Finance;`), icon references, and MIME/protocol bindings (`x-scheme-handler/growww`).
- `linux/packaging/org.kerberosec.growww.metainfo.xml`: AppStream metadata document providing release notes, developer info, screenshots, and licensing metadata.
- `.github/workflows/linux_release.yml`: Continuous deployment matrix building Flatpak bundles, Snaps, and AppImages concurrently on `ubuntu-22.04`.

## Scope Boundaries
- **In Scope:**
 - Native Linux CMake/Ninja compilation targeting GTK 3 and Wayland/X11.
 - Flatpak manifest creation with Flathub runtime and D-Bus portal permissions.
 - Snapcraft manifest creation with strict confinement and plug bindings.
 - Portable AppImage generation with bundled shared libraries (`libsecret`, `libepoxy`).
 - FreeDesktop `.desktop` entry and icon theme asset generation.
 - CI/CD automation and artifact publishing.
- **Out of Scope / Handled Elsewhere:**
 - Windows packaging (Prompt 518).
 - macOS packaging (Prompt 520).
 - Native Linux secure storage plugin implementation (Prompt 521).

## Technology to Use
- **Primary Toolchain:** Flutter SDK 3.22+, Clang/LLVM 16+, CMake 3.28+, Ninja, GTK 3.24+, `pkg-config`, `libsecret-1-dev`, `libjsoncpp-dev`.
- **Justification:** GTK 3 is the standard Flutter desktop embedding runtime on Linux, offering native performance, hardware-accelerated OpenGL/EGL rendering, and full Wayland/X11 compatibility.
- **Dependencies & Tools:**
 - `flatpak-builder`
 - `snapcraft` (LXD / multipass backend)
 - `appimagetool-x86_64.AppImage`
 - FreeDesktop `desktop-file-utils`

## Backend / Infra Touchpoints
- **Flathub & Snap Store:** Automated deployment to Flathub pull-request repositories and Canonical Snapcraft Store edges (`beta`, `stable`).
- **GitHub Releases / S3 Release Bucket:** Direct hosting of standalone `.AppImage` binaries with SHA-256 checksums and GPG detached signatures.

## Blockchain Interaction
- **Libsecret & Keyring Protection:** Communicates with the Linux FreeDesktop Secret Service API (`org.freedesktop.secrets`) to store and retrieve sensitive JWT session tokens and cryptographic signing credentials used to verify Hyperledger Besu proof-of-reserve Merkle roots.
- **High-Throughput WebSocket Streaming:** Leverages native Linux Epoll sockets via Dart async I/O for ultra-low latency market data streaming from the matching engine (Prompt 205).

## Step-by-Step Build Instructions
1. Enable Linux desktop support in Flutter workspace via `flutter config --enable-linux-desktop`.
2. Install system build dependencies: `sudo apt-get install clang cmake ninja-build pkg-config libgtk-3-dev libsecret-1-dev libjsoncpp-dev`.
3. Configure `linux/CMakeLists.txt` ensuring binary output is named `growww` with release optimization flags (`-O3 -DNDEBUG`).
4. Author FreeDesktop compliant `org.kerberosec.growww.desktop` file with valid Exec paths and `MimeType=x-scheme-handler/growww;`.
5. Author AppStream metadata file `org.kerberosec.growww.metainfo.xml` conforming to the FreeDesktop AppStream 0.16 specification.
6. Create Flatpak manifest `linux/packaging/flatpak/org.kerberosec.growww.yml` targeting `org.gnome.Platform//46` and granting D-Bus secret service access (`--talk-name=org.freedesktop.secrets`).
7. Create Snapcraft manifest `linux/packaging/snap/snapcraft.yaml` with `confinement: strict` and `plugs: [network, network-status, password-manager-service, x11, wayland, desktop, desktop-legacy]`.
8. Create AppImage bundling script `scripts/build_appimage.sh` which stages Flutter release binaries, pulls required `libsecret-1.so.0` and dependencies into `AppDir/usr/lib/`, and compiles with `appimagetool`.
9. Build GitHub Actions workflow `.github/workflows/linux_release.yml` with a matrix for `flatpak`, `snap`, and `appimage`.
10. Implement Flatpak CI step using `flatpak/flatpak-github-actions/flatpak-builder` producing an `.isolated-flatpak` single-file bundle.
11. Implement Snapcraft CI step using `canonical/action-snapcraft` to compile `.snap` package.
12. Implement AppImage CI step to produce `Growww-x86_64.AppImage` and generate SHA-256 checksums.
13. Verify that the `.desktop` file passes `desktop-file-validate` and the AppStream XML passes `appstreamcli validate`.
14. Test AppImage and Flatpak installation on both Ubuntu 22.04 (GNOME X11/Wayland) and Fedora 40 (KDE Plasma Wayland).

## Interfaces / Contracts

```yaml
# linux/packaging/flatpak/org.kerberosec.growww.yml
app-id: org.kerberosec.growww
runtime: org.gnome.Platform
runtime-version: '46'
sdk: org.gnome.Sdk
command: growww

finish-args:
 - --share=ipc
 - --share=network
 - --socket=fallback-x11
 - --socket=wayland
 - --device=dri
 - --filesystem=xdg-documents
 - --talk-name=org.freedesktop.secrets
 - --talk-name=org.freedesktop.Notifications

modules:
 - name: growww
    buildsystem: simple
    build-commands:
 - install -D -m 755 growww /app/bin/growww
 - cp -r data /app/bin/
 - cp -r lib /app/bin/
 - install -D -m 644 org.kerberosec.growww.desktop /app/share/applications/org.kerberosec.growww.desktop
 - install -D -m 644 org.kerberosec.growww.metainfo.xml /app/share/metainfo/org.kerberosec.growww.metainfo.xml
 - install -D -m 644 icons/512x512.png /app/share/icons/hicolor/512x512/apps/org.kerberosec.growww.png
    sources:
 - type: dir
        path: ../../../build/linux/x64/release/bundle/
```

```yaml
# linux/packaging/snap/snapcraft.yaml
name: growww
version: '1.0.0'
summary: Compliant Blockchain Stock & Fractional Asset Investment Terminal
description: |
  Growww is an institutional-grade investment terminal providing SEBI-compliant 
  fractional equity trading with on-chain Proof-of-Reserve transparency.
base: core22
confinement: strict
grade: stable

apps:
  growww:
    command: growww
    extensions: [flutter-gtk3]
    plugs:
 - network
 - network-status
 - password-manager-service
 - x11
 - wayland
 - desktop
 - desktop-legacy
 - unity7
    environment:
      LD_LIBRARY_PATH: $SNAP/usr/lib/$CRAFT_ARCH_TRIPLET/pulseaudio:$LD_LIBRARY_PATH

parts:
  growww-app:
    plugin: dump
    source: build/linux/x64/release/bundle/
    stage-packages:
 - libsecret-1-0
 - libjsoncpp25
```

## Security & Compliance Notes
- **Sandboxing & Least Privilege:** In Flatpak and Snap manifests, restrict filesystem access strictly to user-designated document directories (for downloading tax statements and trade receipts). Never request full `--filesystem=host` access.
- **Keyring Storage Security:** Ensure the application utilizes D-Bus Secret Service (`libsecret`) to delegate encryption of auth tokens and cryptographic nonces to GNOME Keyring or KWallet.
- **Wayland Protocol Isolation:** On Wayland sessions, ensure memory buffers are isolated from other desktop processes to prevent background keylogging and unauthorized screen scraping during financial transactions.
- **Reproducible Artifacts:** Produce detached GPG signatures (`.sig`) and SHA-256 checksums for all distributed `.AppImage` binaries to prevent supply chain tampering.

## Acceptance Criteria
- [ ] `flutter build linux --release` compiles with zero linker errors on standard Linux environments.
- [ ] Flatpak bundle builds cleanly, installs, and runs inside a sandboxed environment without security violations.
- [ ] Snapcraft builds a strictly confined `.snap` package that passes `snapcraft review` checks.
- [ ] Portable `.AppImage` runs seamlessly on clean Ubuntu, Fedora, and Arch installations without missing shared library dependencies.
- [ ] Secure storage operations correctly write to and read from GNOME Keyring / KWallet via `libsecret`.
- [ ] Custom URL scheme `growww://` triggers the Linux client when invoked from browser or terminal (`xdg-open`).

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 501 (Project Scaffolding), Prompt 521 (Local Secure Storage), Prompt 108 (Environment Strategy).
- **Parallel Tasks:** Prompt 518 (Windows Packaging), Prompt 520 (macOS Packaging).
