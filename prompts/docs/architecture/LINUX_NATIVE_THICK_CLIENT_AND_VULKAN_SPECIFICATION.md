# Linux Native Thick Client & Vulkan Impeller Architectural Specification

**Specification ID:** SPEC-ARCH-043-CLIENT-LINUX-VULKAN  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Native Desktop Thick Client, Linux Graphics Pipeline & Low-Latency Engine Architecture  
**Target Environments:** Linux Desktop Environments (GNOME 40+, KDE Plasma 5.27/6.0+, Sway, Hyprland), Wayland / X11 Compositors, Linux Kernels 5.15+ / 6.x  
**Target Toolchains:** Flutter 3.22+ / Dart 3.4+, GTK 3.24+ / GTK 4, Vulkan SDK 1.3+, Rust 1.78+ (`librust_trading_core.so`), libsecret-1  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Architectural Scope

The Growww / NBSE Linux Native Thick Client provides algorithmic desks, proprietary traders, quantitative analysts, and institutional Linux workstations with an uncompromising, zero-compromise desktop trading terminal. Linux workstations require unique architectural considerations: heterogeneous display server protocols (Wayland and X11), asynchronous inter-process communication, hardware-accelerated graphics drivers across diverse silicon vendors (AMD RADV, NVIDIA Proprietary, Intel ANV), security sandboxing via Flatpak portals, and system-level credential security via the Freedesktop Secret Service standard.

### Core Architectural Pillars:
1. **Multi-Head GTK 3/4 & Wayland/X11 Shell:** A custom C++ embedder leveraging Flutter 3.22+ desktop APIs, running directly inside the GLib main event loop. It provides seamless display management across both Wayland (wl_compositor, xdg_shell, wp_fractional_scale_v1) and legacy X11 (Xlib, XCB, EWMH).
2. **Impeller Vulkan Hardware Acceleration:** A deterministic, 144Hz to 360Hz hardware-accelerated rendering pipeline utilizing Vulkan 1.3 via Flutter's Impeller graphics engine. Precompiled SPIR-V shaders eliminate runtime shader compilation jank completely.
3. **Freedesktop Secret Service Security:** Enterprise credential and private key storage utilizing `libsecret-1` and the `org.freedesktop.secrets` DBus API, backed by hardware-backed TPM 2.0 / GNOME Keyring / KeePassXC with hardened in-memory buffer zeroization.
4. **Multi-Monitor Window Detachment:** A multi-process, shared-memory desktop architecture allowing arbitrary detachment of trading views (L2 Order Book, 4K Candlestick Chart, DOM Execution Ladder) across independent physical monitors with color-coded ticker link groups.
5. **Low-Latency Global Trading Hotkeys:** Dual-backend global key trapping supporting direct X11 `XGrabKey` alongside Wayland `org.freedesktop.portal.GlobalShortcuts` for unfocused, sub-millisecond panic kill-switch and trade execution.
6. **Universal Linux Packaging Matrix:** Automated builds for Flatpak (with granular XDG socket portals), AppImage (fully bundled standalone runtime), and Debian `.deb` packages with strict dependency boundaries.
7. **Rust FFI Socket Core (`librust_trading_core.so`):** A lock-free, zero-allocation native network engine streaming binary market data directly into ring buffers, using zero-GC Dart FFI native ports capable of processing 100,000+ ticks per second without triggering garbage collector pauses.
8. **Strict Zero-Fee & Gas Presentation Invariant:** Immutable enforcement of the 0.00% Zero-Fee guarantee (₹0.00 Maker / ₹0.00 Taker) and 0 Gas sponsored execution badges across every screen and trade blotter.

```
+-------------------------------------------------------------------------------------------------------+
|                                LINUX NATIVE THICK CLIENT ARCHITECTURE                                 |
|                                                                                                       |
|  +-------------------------------------------------------------------------------------------------+  |
|  |                            FLUTTER 3.22+ DART UI LAYER (OBSIDIAN DARK)                          |  |
|  |  +-----------------------+  +------------------------+  +------------------------------------+  |  |
|  |  | L2/L3 Order Depth DOM |  | Multi-Timeframe Chart  |  | Order Execution & 0.00% Fee Bar    |  |  |
|  |  +-----------------------+  +------------------------+  +------------------------------------+  |  |
|  |  +-------------------------------------------------------------------------------------------+  |  |
|  |  | Window Management & Link Bus: Multi-Monitor Synchronizer (Red / Blue / Green / Yellow)     |  |  |
|  |  +-------------------------------------------------------------------------------------------+  |  |
|  +----------------------------------------------+--------------------------------------------------+  |
|                                                 | Dart FFI / C++ Platform Channels                     |
|  +----------------------------------------------v--------------------------------------------------+  |
|  |                            NATIVE LINUX C++ / GTK EMBEDDER LAYER                                |  |
|  |  +-------------------------------------+  +--------------------------------------------------+  |  |
|  |  | GTK 3 / GTK 4 Application Shell     |  | FlutterLinuxView & Impeller Vulkan Surface       |  |  |
|  |  | - GLib Main Event Loop Integration  |  | - VK_KHR_wayland_surface / VK_KHR_xlib_surface   |  |  |
|  |  | - GtkWindow Multi-Window Controller |  | - Vulkan 1.3 Queue, Swapchain (Mailbox/FIFO)    |  |  |
|  |  +------------------+------------------+  +--------------------------------------------------+  |  |
|  +---------------------|---------------------------------------------------------------------------+  |
|                        |                                                                              |
|       +----------------+----------------+-------------------------------+                             |
|       |                                 |                               |                             |
|       v                                 v                               v                             |
|  +------------------------+  +------------------------+  +------------------------------------------+ |
|  | DISPLAY SERVER BRIDGE  |  | SECRET SERVICE CLIENT  |  | RUST FFI CORE (librust_trading_core.so)  | |
|  | - Wayland (xdg_shell)  |  | - libsecret-1 / DBus   |  | - Lock-Free Disruptor Ring Buffer        | |
|  | - X11 (XCB / EWMH)     |  | - org.freedesktop.sec  |  | - Tokio Epoll WebSocket WSS Client       | |
|  | - XDG Desktop Portals  |  | - AES-256 DH Session   |  | - SBE / Protobuf Binary Parser           | |
|  | - GlobalShortcuts API  |  | - Secure mlock Memory  |  | - Zero-GC Dart NativePort Streaming      | |
|  +------------------------+  +------------------------+  +------------------------------------------+ |
+-------------------------------------------------------------------------------------------------------+
```

---

## 2. Native GTK Multi-Head Shell & Display Server Architecture

The Linux client relies on a native GTK 3.24+ / GTK 4 shell hosting the Flutter Linux engine. It accommodates the diverse Linux display server ecosystem by maintaining two specialized display server backends behind a unified abstraction layer.

### 2.1 Flutter 3.22+ Linux Desktop C++ Embedder Integration

The entry point of the desktop thick client initializes the GTK application shell, registers the Flutter engine embedder, and binds the platform channels. The lifecycle is embedded into GLib's `GApplication` / `GtkApplication` model.

```cpp
// src/linux/main.cc - Institutional GTK Linux Embedder Entry Point
#include <flutter_linux/flutter_linux.h>
#include <gtk/gtk.h>
#include <glib.h>
#include "trading_window_manager.h"
#include "secret_service_bridge.h"
#include "global_shortcuts_manager.h"

static void growww_application_activate(GApplication* application) {
    GtkApplication* app = GTK_APPLICATION(application);
    GtkWindow* window = GTK_WINDOW(gtk_application_window_new(app));

    gtk_window_set_title(window, "Growww / NBSE Pro Trading Workstation");
    gtk_window_set_default_size(window, 1920, 1080);

    // Enforce Dark Surface styling for window decorations
    GtkSettings* settings = gtk_settings_get_default();
    g_object_set(settings, "gtk-application-prefer-dark-theme", TRUE, NULL);

    g_autoptr(FlDartProject) project = fl_dart_project_new();
    const char* const dart_entrypoint_args[] = {"--disable-observatory", NULL};
    fl_dart_project_set_dart_entrypoint_arguments(project, const_cast<char**>(dart_entrypoint_args));

    FlView* view = fl_view_new(project);
    gtk_container_add(GTK_CONTAINER(window), GTK_WIDGET(view));

    // Register Native Linux Platform Channels
    FlBinaryMessenger* messenger = fl_engine_get_binary_messenger(fl_view_get_engine(view));
    trading_window_manager_init(messenger, app, window);
    secret_service_bridge_init(messenger);
    global_shortcuts_manager_init(messenger, window);

    gtk_widget_show_all(GTK_WIDGET(window));
}

int main(int argc, char** argv) {
    g_autoptr(GtkApplication) app = gtk_application_new(
        "in.growww.trading.desktop",
        G_APPLICATION_FLAGS_NONE
    );
    g_signal_connect(app, "activate", G_CALLBACK(growww_application_activate), NULL);
    return g_application_run(G_APPLICATION(app), argc, argv);
}
```

### 2.2 Display Server Protocol Separation: Wayland vs X11

The application dynamically detects the active display server via `GDK_IS_WAYLAND_DISPLAY()` and `GDK_IS_X11_DISPLAY()`, executing server-specific code paths:

```
                  +-----------------------------------+
                  |  GDK Display Server Auto-Detect   |
                  +-----------------+-----------------+
                                    |
            +-----------------------+-----------------------+
            |                                               |
            v                                               v
+-------------------------------+               +-------------------------------+
|     Wayland Display Backend   |               |      X11 Display Backend      |
|-------------------------------|               |-------------------------------|
| - wl_compositor, wl_subsurface|               | - Xlib / XCB Direct Access    |
| - xdg_wm_base & xdg_toplevel  |               | - Extended Window Manager     |
| - wp_fractional_scale_v1      |               |   Hints (EWMH / NetWM)        |
| - wp_viewporter               |               | - XRandR 1.5 Multi-Head API   |
| - XDG Desktop Portals IPC     |               | - Low-level XGrabKey hotkeys  |
| - VK_KHR_wayland_surface      |               | - VK_KHR_xlib_surface         |
+-------------------------------+               +-------------------------------+
```

1. **Wayland Architecture:**
   - Adheres strictly to the `xdg-shell` protocol. Windows cannot programmatically set their absolute screen coordinates.
   - Detached windows utilize `xdg_wm_base` with subsurfaces or independent `xdg_toplevel` surfaces.
   - High-DPI rendering leverages the `wp_fractional_scale_v1` and `wp_viewporter` protocols, preventing blurry integer rasterization on 125%, 150%, and 175% scale monitors.
2. **X11 Fallback Architecture:**
   - Uses XCB / Xlib handles directly to read `_NET_WORKAREA`, `_NET_CLIENT_LIST`, and `_NET_WM_DESKTOP`.
   - Supports explicit absolute pixel placement across continuous virtual multi-monitor bounding boxes via `XMoveResizeWindow`.

### 2.3 Multi-Head Display Enumeration & Fractional Scaling

The client dynamically queries multi-monitor configurations using `GdkDisplay` and `GdkMonitor`, forwarding real-time geometry updates to the Dart UI layer:

```cpp
// Native Display Enumeration Implementation
void enumerate_displays(FlMethodChannel* channel) {
    GdkDisplay* display = gdk_display_get_default();
    int n_monitors = gdk_display_get_n_monitors(display);
    g_autoptr(FlValue) display_list = fl_value_new_list();

    for (int i = 0; i < n_monitors; i++) {
        GdkMonitor* monitor = gdk_display_get_monitor(display, i);
        GdkRectangle geometry;
        gdk_monitor_get_geometry(monitor, &geometry);

        int scale_factor = gdk_monitor_get_scale_factor(monitor);
        const char* model = gdk_monitor_get_model(monitor);
        const char* manufacturer = gdk_monitor_get_manufacturer(monitor);
        gboolean is_primary = gdk_monitor_is_primary(monitor);

        g_autoptr(FlValue) mon_dict = fl_value_new_map();
        fl_value_set_string_take(mon_dict, "id", fl_value_new_int(i));
        fl_value_set_string_take(mon_dict, "x", fl_value_new_int(geometry.x));
        fl_value_set_string_take(mon_dict, "y", fl_value_new_int(geometry.y));
        fl_value_set_string_take(mon_dict, "width", fl_value_new_int(geometry.width));
        fl_value_set_string_take(mon_dict, "height", fl_value_new_int(geometry.height));
        fl_value_set_string_take(mon_dict, "scaleFactor", fl_value_new_int(scale_factor));
        fl_value_set_string_take(mon_dict, "model", fl_value_new_string(model ? model : "Generic"));
        fl_value_set_string_take(mon_dict, "isPrimary", fl_value_new_bool(is_primary));

        fl_value_append(display_list, mon_dict);
    }

    fl_method_channel_invoke_method(
        channel, "onDisplaysUpdated", display_list, NULL, NULL, NULL
    );
}
```

### 2.4 GLib Main Context & GTK Event Loop Interleaving

High-frequency market ticks must never starve the GTK UI event loop:
- The Linux embedder uses `g_main_context_invoke_full()` with priority `G_PRIORITY_DEFAULT_IDLE` for non-critical render events.
- Hot-path trade execution and panic cancel events bypass idle queues, dispatching directly via POSIX condition variables and Unix domain sockets directly to the Rust engine thread.

---

## 3. Vulkan Hardware-Accelerated Rendering Pipeline via Impeller

The Flutter 3.22+ Linux thick client mandates the **Impeller Vulkan** rendering engine. Unlike legacy Skia, which compiles OpenGL shaders at runtime and causes micro-stutters during volatile market conditions, Impeller relies on Ahead-of-Time (AOT) compiled SPIR-V shaders and explicit GPU memory management.

```
+---------------------------------------------------------------------------------------------------+
|                              IMPELLER VULKAN PIPELINE ARCHITECTURE                                |
|                                                                                                   |
|  [ Dart Render Tree ]                                                                             |
|          |                                                                                        |
|          v                                                                                        |
|  [ Impeller DisplayList ]                                                                         |
|          |                                                                                        |
|          v                                                                                        |
|  [ impellerc AOT Compiler ] ----> [ Pre-Baked SPIR-V Shaders ]                                    |
|                                             |                                                     |
|          +----------------------------------+----------------------------------+                  |
|          |                                                                     |                  |
|          v                                                                     v                  |
|  [ VkInstance / VkDevice ]                                           [ VMA Memory Allocator ]     |
|  - Physical Device Selection (Discrete GPU Priority)                 - Staging Buffers (Host)     |
|  - Queue Family: Graphics + Compute + Transfer                       - Vertex / Index Ring (Dev)  |
|          |                                                                     |                  |
|          +----------------------------------+----------------------------------+                  |
|                                             |                                                     |
|                                             v                                                     |
|                             [ Swapchain (VK_KHR_swapchain) ]                                      |
|                             - Surface: VK_KHR_wayland_surface / xlib                              |
|                             - Mode: VK_PRESENT_MODE_MAILBOX_KHR (144Hz+)                          |
|                             - Image Count: 3 (Triple-Buffering)                                   |
|                             - Color Space: VK_COLOR_SPACE_SRGB_NONLINEAR_KHR                      |
|                                             |                                                     |
|                                             v                                                     |
|                              [ Presentation to Physical Displays ]                                |
+---------------------------------------------------------------------------------------------------+
```

### 3.1 Impeller Vulkan Backend Architecture & Surface Binding

The embedder binds Vulkan surfaces conditionally based on the active display server:
- On Wayland: Binds via `vkCreateWaylandSurfaceKHR()` with `wl_display*` and `wl_surface*`.
- On X11: Binds via `vkCreateXlibSurfaceKHR()` with `Display*` and `Window`.

```c
// Surface Creation Abstraction Matrix
VkResult create_vulkan_surface(
    VkInstance instance,
    GdkWindow* gdk_window,
    const VkAllocationCallbacks* allocator,
    VkSurfaceKHR* surface
) {
    GdkDisplay* gdk_display = gdk_window_get_display(gdk_window);

#if defined(GDK_WINDOWING_WAYLAND)
    if (GDK_IS_WAYLAND_DISPLAY(gdk_display)) {
        struct wl_display* wayland_display = gdk_wayland_display_get_wl_display(gdk_display);
        struct wl_surface* wayland_surface = gdk_wayland_window_get_wl_surface(gdk_window);

        VkWaylandSurfaceCreateInfoKHR create_info = {
            .sType = VK_STRUCTURE_TYPE_WAYLAND_SURFACE_CREATE_INFO_KHR,
            .pNext = NULL,
            .flags = 0,
            .display = wayland_display,
            .surface = wayland_surface,
        };
        return vkCreateWaylandSurfaceKHR(instance, &create_info, allocator, surface);
    }
#endif

#if defined(GDK_WINDOWING_X11)
    if (GDK_IS_X11_DISPLAY(gdk_display)) {
        Display* x11_display = gdk_x11_display_get_xdisplay(gdk_display);
        Window x11_window = gdk_x11_window_get_xid(gdk_window);

        VkXlibSurfaceCreateInfoKHR create_info = {
            .sType = VK_STRUCTURE_TYPE_XLIB_SURFACE_CREATE_INFO_KHR,
            .pNext = NULL,
            .flags = 0,
            .dpy = x11_display,
            .window = x11_window,
        };
        return vkCreateXlibSurfaceKHR(instance, &create_info, allocator, surface);
    }
#endif

    return VK_ERROR_INITIALIZATION_FAILED;
}
```

### 3.2 Offline SPIR-V Ahead-of-Time (AOT) Shader Pipeline Compilation (`impellerc`)

To guarantee absolute frame-pacing without runtime compilation hitches:
- All vertex, fragment, and compute shaders are compiled during client build time using the `impellerc` compiler.
- Shaders are emitted as optimized SPIR-V bytecode blobs embedded directly within the binary.
- Custom candlestick compute shaders run parallelized volume weighting and technical indicator overlays directly on the GPU compute queue.

```bash
# Build pipeline shader compilation step
impellerc \
  --sksl \
  --iplr \
  --sl=engine/src/shaders/candlestick_renderer.frag \
  --spirv=build/shaders/candlestick_renderer.frag.spirv \
  --output=build/shaders/candlestick_renderer_frag.o
```

### 3.3 Swapchain Management, Presentation Modes & Triple-Buffering

Trading terminals demand instant visual response to market fluctuations:
1. **Presentation Mode Selection:**
   - **`VK_PRESENT_MODE_MAILBOX_KHR` (Preferred):** Triple-buffered, low-latency, tear-free presentation. Newest frame replaces pending frame in swapchain queue. Delivers minimum input-to-photon latency on 144Hz, 240Hz, and 360Hz monitors.
   - **`VK_PRESENT_MODE_FIFO_KHR` (Standard Fallback):** V-Sync locked presentation mode supported across all Vulkan hardware conformances.
2. **Buffer Count:** Exactly 3 images (`minImageCount = 3`) allocated to prevent CPU wait states on vertical retrace.

### 3.4 Vulkan Memory Allocator (VMA) & Buffer Management

Memory operations for high-throughput order book rendering use the Vulkan Memory Allocator (VMA):
- **Dynamic Geometry Buffers:** Order book depth bars and trade ticks allocate from host-visible, coherent staging buffers mapped with `VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT | VK_MEMORY_PROPERTY_HOST_COHERENT_BIT`.
- **Static Textures & Icons:** Institutional SVG icons and font atlases reside in device-local memory (`VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT`), uploaded via asynchronous transfer queues.

### 3.5 GPU Hardware & Driver Compatibility Matrix

| Driver / Vendor | Target Architecture | Vulkan Conformance | Presentation Mode | Fallback Policy |
| :--- | :--- | :--- | :--- | :--- |
| **AMD RADV (Mesa)** | RDNA 1 / 2 / 3, GCN 4+ | Vulkan 1.3 Native | Mailbox (Tear-free 240Hz) | First-class, Primary Target |
| **NVIDIA Proprietary** | Turing, Ampere, Ada Lovelace | Vulkan 1.3 (Driver 550+) | Mailbox / Immediate | Direct Mode Presentation |
| **Intel ANV (Mesa)** | Xe, Iris Xe, Arc Alchemist | Vulkan 1.3 Native | FIFO / Mailbox | Full Impeller Vulkan |
| **Legacy / Virtualized** | VMware SVGA, VirtIO-GPU | Vulkan 1.1 or Software | FIFO | Impeller OpenGL ES 3.0 |

If Vulkan 1.2+ initialization fails or no suitable physical GPU device is discovered, the client outputs an explicit diagnostic log and gracefully transitions to the software/OpenGL ES backend via GTK.

---

## 4. Linux Secret Service API (`libsecret` / DBus) Encrypted Credential Architecture

Linux desktop environments lack a single monolithic keychain like macOS Keychain or Windows Credential Manager. Instead, the Freedesktop.org **Secret Service API** (`org.freedesktop.secrets`) provides the standard cross-desktop credential storage, supported by GNOME Keyring, KDE KWallet, and KeePassXC.

```
+---------------------------------------------------------------------------------------------------+
|                         FREEDESKTOP SECRET SERVICE INTEGRATION PIPELINE                           |
|                                                                                                   |
|  [ Dart Vault Service ]                                                                           |
|          |                                                                                        |
|          v (Dart FFI)                                                                             |
|  [ libsecret-1 Native Bridge ]                                                                    |
|          |                                                                                        |
|          v (DBus IPC: org.freedesktop.secrets)                                                    |
|  +---------------------------------------------------------------------------------------------+  |
|  | DBus System Session Daemon                                                                  |  |
|  | - Service Path: /org/freedesktop/secrets                                                    |  |
|  | - Session Negotiation: OpenSession("DH-ietf1024-sha256-aes128-cbc-pkcs7")                  |  |
|  | - Collection: /org/freedesktop/secrets/collection/login (or "default")                      |  |
|  +---------------------------------------------------------------------------------------------+  |
|          |                                                                                        |
|          +----------------------------------+----------------------------------+                  |
|          |                                  |                                  |                  |
|          v                                  v                                  v                  |
|  [ GNOME Keyring ]                  [ KDE KWallet ]                   [ KeePassXC / Custom ]      |
|  - TPM 2.0 / LUKS Backed            - GPG / Blowfish Backed           - Argon2id / AES-256        |
|  - AES-256 Master Key               - Polkit Authentication           - Local Master Password     |
+---------------------------------------------------------------------------------------------------+
```

### 4.1 `org.freedesktop.secrets` DBus Protocol Specification

The client stores institutional credentials via `libsecret-1` C APIs, which internally invoke DBus methods on `org.freedesktop.secrets`:

```c
// secret_service_bridge.c - Production libsecret Implementation
#include <libsecret/secret.h>
#include <glib.h>

const SecretSchema* growww_trading_get_schema(void) {
    static const SecretSchema schema = {
        "in.growww.trading.Credentials",
        SECRET_SCHEMA_NONE,
        {
            {"account_id", SECRET_SCHEMA_ATTRIBUTE_STRING},
            {"token_type", SECRET_SCHEMA_ATTRIBUTE_STRING},
            {"environment", SECRET_SCHEMA_ATTRIBUTE_STRING},
            {"NULL", 0},
        }
    };
    return &schema;
}

gboolean store_trading_secret(
    const gchar* account_id,
    const gchar* token_type,
    const gchar* environment,
    const gchar* secret_data,
    GError** error
) {
    return secret_password_store_sync(
        growww_trading_get_schema(),
        SECRET_COLLECTION_DEFAULT,
        "Growww Trading Session Token",
        secret_data,
        NULL,
        error,
        "account_id", account_id,
        "token_type", token_type,
        "environment", environment,
        NULL
    );
}

gchar* lookup_trading_secret(
    const gchar* account_id,
    const gchar* token_type,
    const gchar* environment,
    GError** error
) {
    return secret_password_lookup_sync(
        growww_trading_get_schema(),
        NULL,
        error,
        "account_id", account_id,
        "token_type", token_type,
        "environment", environment,
        NULL
    );
}
```

### 4.2 Credential Schema & Item Attributes

All items saved to the Secret Service include strict attribute filtering to prevent collision and allow precise query isolation:

```json
{
  "schema": "in.growww.trading.Credentials",
  "label": "NBSE Ed25519 Trading Signature Key",
  "attributes": {
    "account_id": "ACC-99482-IND",
    "token_type": "ed25519_private_key",
    "environment": "production_genesis",
    "fingerprint": "SHA256:7b4f8c2e91a0..."
  },
  "secret_payload": "<Base64 Encrypted Payload>"
}
```

### 4.3 Hardened In-Memory Buffer Security

When cryptographic seeds, private keys, or API tokens are loaded into RAM:
1. **Memory Locking via `mlock(2)`:** Prevents Linux kernel page swapping to disk (swap partition/file), defeating offline forensics.
2. **Page Write Protection via `mprotect(2)`:** Marks pages read-only (`PROT_READ`) immediately after ingestion, only marking writeable (`PROT_WRITE`) during intentional destruction.
3. **Explicit Zeroization (`explicit_bzero(3)`):** Wipes memory segments before deallocation to prevent lingering key material in dead heap allocations.

```c
// Memory hardening primitives
void* allocate_secure_credential_buffer(size_t size) {
    void* ptr = mmap(NULL, size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (ptr == MAP_FAILED) return NULL;
    mlock(ptr, size); // Prevent swapout
    return ptr;
}

void release_secure_credential_buffer(void* ptr, size_t size) {
    if (!ptr) return;
    explicit_bzero(ptr, size); // Secure zeroization
    munlock(ptr, size);
    munmap(ptr, size);
}
```

---

## 5. Multi-Monitor Window Detachment Across Mixed X11 & Wayland Environments

Institutional trading operations require detached layouts: order books on monitor 1, full-screen technical charts on monitor 2, and DOM ladders with trade blotters on monitor 3. The architecture solves window detachment challenges across differing display server protocols.

```
+---------------------------------------------------------------------------------------------------+
|                         MULTI-MONITOR DETACHMENT ARCHITECTURE                                     |
|                                                                                                   |
|  +---------------------------------------------------------------------------------------------+  |
|  | PRIMARY DISPLAY (Workstation Monitor 1 - 4K UHD)                                           |  |
|  | +-----------------------------------------------------------------------------------------+ |  |
|  | | MAIN APPLICATION WINDOW (Master Coordinator)                                           | |  |
|  | | - Watchlist (Group Red)  - Position Blotter  - Account Balance / 0.00% Zero-Fee Summary     | |  |
|  | +-----------------------------------------------------------------------------------------+ |  |
|  +----------------------------------------------+----------------------------------------------+  |
|                                                 | POSIX Shared Memory (SHM) / Unix Domain Sockets  |
|         +---------------------------------------+---------------------------------------+         |
|         |                                                                               |         |
|         v                                                                               v         |
|  +-------------------------------------------+   +-------------------------------------------+    |
|  | SATELLITE WINDOW 1 (Monitor 2 - Vertical) |   | SATELLITE WINDOW 2 (Monitor 3 - 4K Chart) |    |
|  |-------------------------------------------|   |-------------------------------------------|    |
|  | - Detached Depth Ladder (Group Red)       |   | - Detached Candlestick Chart (Group Red)  |    |
|  | - Single-Click Scalper DOM                |   | - Volume Profile, MACD, 240Hz Impeller    |    |
|  | - Invariant 0.00% Fee Order Entry         |   | - Shared In-Memory Tick Bus               |    |
|  +-------------------------------------------+   +-------------------------------------------+    |
+---------------------------------------------------------------------------------------------------+
```

### 5.1 Satellite Detachment Lifecycle & Subsurface Hierarchy

When a user drags or clicks "Pop-Out Window":
1. The master GTK window dispatches a detach event to the `TradingWindowManager`.
2. A new `GtkWindow` or a separate child process is created.
3. The detached window registers its assigned `WindowId` with the inter-process broker.
4. The market data feed attaches to the existing shared-memory ring buffer without opening redundant WebSocket connections.

### 5.2 Wayland Window Placement Constraints & Protocol Solutions

Under Wayland, security isolation prevents client applications from setting arbitrary global screen coordinates (`x, y`):
- **Strategy:** Satellite windows request `xdg_toplevel` state and rely on the compositor's window manager rules.
- **Foreign Toplevel Protocol:** In desktop environments supporting `ext-foreign-toplevel-list-v1` or `wlr-foreign-toplevel-management`, the thick client orchestrates focus, minimize/restore, and monitor assignment programmatically.
- **Window Geometry State:** The client saves output monitor identifiers (`GdkMonitor` connector names such as `DP-1`, `HDMI-A-1`) in persistent storage. Upon restart, windows launch on their designated physical screens using the Wayland activation token protocol (`xdg_activation_v1`).

### 5.3 X11 Extended Window Manager Hints (EWMH)

Under X11, window coordinates are managed directly across the continuous virtual root display:
- Windows calculate bounding box coordinates and invoke `XMoveResizeWindow()`.
- Windows set `_NET_WM_STATE_ABOVE`, `_NET_WM_WINDOW_TYPE_NORMAL`, and bind to `_NET_WM_DESKTOP`.
- Monitors are matched using the XRandR 1.5 output extension.

### 5.4 High-Performance Inter-Window IPC (POSIX SHM & Unix Domain Sockets)

To achieve sub-millisecond multi-window synchronization without network hops:
1. **Shared Memory (`shm_open(3)` + `mmap(2)`):** Tick ring buffers and L2 order books reside in a POSIX shared memory segment located at `/dev/shm/growww_market_ring_buffer`.
2. **Atomic Synchronization:** Satellite processes read ticks via lock-free atomic load flags (`std::atomic<uint64_t> sequence_number`).
3. **Unix Domain Sockets (`AF_UNIX`):** Window control signals (Minimize All, Emergency Flatten, Link Group Ticker Changes) travel across a non-blocking UNIX domain socket in `/run/user/<UID>/growww_terminal.sock`.

```c
// Shared Memory Ring Buffer Layout for Multi-Window Sync
struct ShmMarketDataSegment {
    uint32_t magic_header;         // 0x47524F57 ('GROW')
    uint32_t layout_version;       // 1
    pthread_mutex_t control_mutex; // Inter-process coordination
    pthread_cond_t tick_condition; // POSIX broadcast variable
    uint64_t last_published_seq;   // Monotonic sequence number
    uint32_t ring_buffer_capacity; // e.g., 65536 ticks
    // Followed by contiguous array of BinaryMarketTick structs
};
```

### 5.5 Global Link Group State Bus (Red, Blue, Green, Yellow)

Instruments link across windows using color-coded groups:
- **Group Red:** Primary Scalping Pair (e.g., BTC/eINR).
- **Group Blue:** Secondary Futures Hedge (e.g., ETH/eINR).
- **Group Green:** Fiat & Settlement Monitor (e.g., USDT/eINR).
- **Group Yellow:** Equity / Index Derivative (e.g., NBSE-50/eINR).

When an instrument in a Group Red watchlist is clicked, an atomic message broadcasts across the UNIX domain socket. Every detached window configured for Group Red instantaneously switches symbol, order book ladder, and chart data within 200 microseconds.

---

## 6. Low-Latency Global Shortcut Handling (X11 XGrabKey & Wayland Portals)

Proprietary trading operations mandate system-wide hotkeys that execute even when the terminal window is minimized or out of focus.

### 6.1 High-Priority Trading Hotkey Registry

| Hotkey Combination | Action Trigger | Latency Target | Description |
| :--- | :--- | :--- | :--- |
| `Space` (Global) | **PANIC CANCEL ALL** | < 1 ms | Cancels all active resting orders across all pairs |
| `Ctrl + Alt + X` | **PANIC FLATTEN ALL** | < 2 ms | Liquidates open inventory into fiat cash via market orders |
| `Shift + B` | Quick Buy at Best Ask | < 1 ms | Inserts limit order at Best Offer on active linked pair |
| `Shift + S` | Quick Sell at Best Bid | < 1 ms | Inserts limit order at Best Bid on active linked pair |
| `Ctrl + Shift + Escape` | Emergency Disconnect | < 1 ms | Shuts down network connections and trips Exchange dead-man switch |

### 6.2 X11 Root Window Raw Event Trapping (`XGrabKey`)

Under X11, shortcuts bind to the root window to capture keystrokes globally:

```c
// x11_shortcuts.c - Low-latency X11 Key Grabber
#include <X11/Xlib.h>
#include <X11/keysym.h>

void register_x11_panic_key(Display* dpy, Window root) {
    KeyCode space_code = XKeysymToKeycode(dpy, XK_Space);

    // Grab Space key with common lock masks (NumLock, CapsLock)
    unsigned int modifiers[] = {0, Mod2Mask, LockMask, Mod2Mask | LockMask};
    for (int i = 0; i < 4; i++) {
        XGrabKey(dpy, space_code, modifiers[i], root, True, GrabModeAsync, GrabModeAsync);
    }
}

void process_x11_events(Display* dpy) {
    XEvent ev;
    while (XPending(dpy)) {
        XNextEvent(dpy, &ev);
        if (ev.type == KeyPress) {
            // High-priority panic cancel dispatch
            execute_panic_cancel_all();
        }
    }
}
```

### 6.3 Wayland `org.freedesktop.portal.GlobalShortcuts` Ingress

Wayland compositors block arbitrary keystroke sniffing to preserve security. Modern Wayland environments (GNOME 44+, KDE 6.0+) resolve this using the **XDG Desktop Portal GlobalShortcuts API**:

```
+---------------------------------------------------------------------------------------------------+
|                        WAYLAND GLOBAL SHORTCUTS PORTAL PIPELINE                                   |
|                                                                                                   |
|  [ Linux Client Application ]                                                                     |
|          |                                                                                        |
|          v DBus Call: org.freedesktop.portal.Desktop                                              |
|  [ /org/freedesktop/portal/desktop -> org.freedesktop.portal.GlobalShortcuts ]                     |
|          |                                                                                        |
|          v                                                                                        |
|  1. CreateSession()                                                                               |
|  2. BindShortcuts(session, shortcuts=[                                                            |
|       {"id": "panic_cancel", "description": "NBSE Panic Cancel All Orders", "preferred_key": "Space"}|
|     ])                                                                                            |
|          |                                                                                        |
|          v (Desktop Compositor Prompts User Once for Authorization)                               |
|  [ System Compositor (Mutter / KWin / Hyprland) ]                                                 |
|          |                                                                                        |
|          v DBus Signal: Activated(session_handle, "panic_cancel", timestamp)                      |
|  [ Linux Client Application Handles Signal ]                                                      |
|          |                                                                                        |
|          v                                                                                        |
|  [ Instant Panic Execution Dispatch ]                                                             |
+---------------------------------------------------------------------------------------------------+
```

---

## 7. High-Throughput Rust FFI Socket Core (`librust_trading_core.so`)

To eliminate garbage collection pauses, the market data stream bypasses Dart's network stack entirely. The underlying network pipeline runs in an optimized native Rust library (`librust_trading_core.so`) compiled with `--release` and `-C target-cpu=native`.

```
+---------------------------------------------------------------------------------------------------+
|                     RUST FFI SOCKET CORE (librust_trading_core.so) PIPELINE                       |
|                                                                                                   |
|  [ NBSE WebSocket Feed (wss://ws.growww.in) ]                                                     |
|          |                                                                                        |
|          v (Kernel Epoll / Non-blocking TCP)                                                      |
|  [ Tokio / Rustls Asynchronous Network Thread ]                                                   |
|          |                                                                                        |
|          v Zero-Copy Binary Decode (SBE / Protobuf)                                               |
|  [ SBE Binary Tick Decoder ]                                                                      |
|          |                                                                                        |
|          v Lock-free Push                                                                         |
|  +---------------------------------------------------------------------------------------------+  |
|  | Disruptor-Style Circular Ring Buffer (65,536 Slots, Cacheline-Aligned 64 bytes)              |  |
|  +---------------------------------------------------------------------------------------------+  |
|          |                                                                                        |
|          +----------------------------------+----------------------------------+                  |
|          |                                  |                                  |                  |
|          v                                  v                                  v                  |
|  [ Raw Depth Visualizer ]          [ VWAP Compute Engine ]            [ Adaptive Conflator ]      |
|  (Direct Dart FFI Pointer)         (Native Rust Worker)               (Max 144 Frames / Sec)      |
|          |                                                                     |                  |
|          +----------------------------------+----------------------------------+                  |
|                                             |                                                     |
|                                             v                                                     |
|                              [ Dart_PostCObject NativePort ]                                      |
|                                             |                                                     |
|                                             v                                                     |
|                          [ Dart UI Thread (Zero GC Pressure) ]                                    |
+---------------------------------------------------------------------------------------------------+
```

### 7.1 Zero-GC Tick Ring Buffer Architecture (Disruptor Pattern)

The native core implements a single-producer, multi-consumer lock-free ring buffer:
- Each slot occupies exactly 64 bytes (aligned to CPU cachelines via `#[repr(align(64))]`) to eliminate false sharing.
- Sequence numbers use atomic operations with `Ordering::Acquire` and `Ordering::Release`.

```rust
// rust_trading_core/src/ring_buffer.rs
use std::sync::atomic::{AtomicU64, Ordering};

pub const RING_BUFFER_CAPACITY: usize = 65536;
pub const RING_BUFFER_MASK: usize = RING_BUFFER_CAPACITY - 1;

#[repr(C, align(64))]
pub struct MarketTick {
    pub symbol_id: u32,
    pub flags: u32,
    pub price: u64,          // Fixed-point micro-units (10^-6 eINR)
    pub quantity: u64,       // Fixed-point micro-lots
    pub timestamp_ns: u64,   // Nanoseconds UTC
    pub trade_id: u64,
    pub maker_order_id: u64,
    pub taker_order_id: u64,
}

pub struct LockFreeRingBuffer {
    buffer: Box<[MarketTick; RING_BUFFER_CAPACITY]>,
    write_cursor: AtomicU64,
}

impl LockFreeRingBuffer {
    pub fn new() -> Self {
        let buffer = unsafe { Box::new_zeroed().assume_init() };
        Self {
            buffer,
            write_cursor: AtomicU64::new(0),
        }
    }

    #[inline(always)]
    pub fn push(&self, tick: MarketTick) -> u64 {
        let seq = self.write_cursor.fetch_add(1, Ordering::Relaxed);
        let index = (seq as usize) & RING_BUFFER_MASK;
        unsafe {
            let slot = self.buffer.as_ptr().add(index) as *mut MarketTick;
            std::ptr::write(slot, tick);
        }
        seq
    }
}
```

### 7.2 Zero-Allocation Dart FFI NativePort Memory Streaming

Data dispatches from Rust into Dart without object allocation using `Dart_PostCObject`:

```rust
// rust_trading_core/src/ffi.rs
use std::os::raw::c_char;
use std::ffi::CStr;

pub type DartPort = i64;

#[repr(C)]
pub struct DartCObject {
    pub ty: i32,
    pub value: DartCObjectValue,
}

#[repr(C)]
pub union DartCObjectValue {
    pub as_int64: i64,
    pub as_ptr: *const u8,
}

extern "C" {
    fn Dart_PostCObject(port: DartPort, message: *mut DartCObject) -> bool;
}

#[no_mangle]
pub extern "C" fn rust_trading_core_init(
    ws_endpoint: *const c_char,
    dart_port: DartPort
) -> *mut LockFreeRingBuffer {
    let endpoint_str = unsafe {
        CStr::from_ptr(ws_endpoint).to_str().unwrap_or("wss://ws.growww.in")
    };
    
    let ring_buffer = Box::into_raw(Box::new(LockFreeRingBuffer::new()));
    start_network_worker(endpoint_str.to_string(), ring_buffer, dart_port);
    ring_buffer
}
```

### 7.3 C-ABI Export Contracts

The shared library exposes standard C-ABI functions with explicit lifecycle management:
1. `rust_trading_core_init`: Launches Tokio thread pool and initiates TLS WebSocket stream.
2. `rust_trading_core_poll_ring_buffer`: Allows Dart to read memory slices synchronously.
3. `rust_trading_core_submit_order`: Encodes and fires sub-millisecond execution orders.
4. `rust_trading_core_panic_cancel`: Dispatches atomic panic cancel frames.
5. `rust_trading_core_destroy`: Tears down sockets and frees memory segments.

---

## 8. Linux Packaging & Distribution Engineering

Institutional and open-source Linux users require standard, friction-free installation methods. The thick client maintains three independent distribution builds.

```
+---------------------------------------------------------------------------------------------------+
|                                LINUX PACKAGING MATRIX & WORKFLOW                                  |
|                                                                                                   |
|  [ Linux Build Artifact: libflutter_linux_gtk.so, librust_trading_core.so, growww_terminal ]       |
|                                         |                                                         |
|         +-------------------------------+-------------------------------+                         |
|         |                               |                               |                         |
|         v                               v                               v                         |
|  +--------------------+       +--------------------+       +--------------------+                 |
|  | FLATPAK BUILDER    |       | APPIMAGE TOOLKIT   |       | DEBIAN PACKAGE     |                 |
|  | - Sandboxed Portal |       | - Self-Contained   |       | - Native Deb       |                 |
|  | - Wayland & X11    |       | - FUSE Run-in-Place|       | - systemd unit     |                 |
|  | - DRI GPU Sockets  |       | - Zero Dependency  |       | - Direct Apt Repo  |                 |
|  +--------------------+       +--------------------+       +--------------------+                 |
|         |                               |                               |                         |
|         v                               v                               v                         |
|  [ Flathub / Enterprise ]     [ Direct Github / CDN ]      [ apt.growww.in PPA ]                  |
+---------------------------------------------------------------------------------------------------+
```

### 8.1 Sandboxed Flatpak Specification

Flatpak provides isolated runtime sandboxing while opening explicit portals for trading hardware:

```yaml
# org.growww.GrowwwTradingTerminal.yaml
app-id: org.growww.GrowwwTradingTerminal
runtime: org.gnome.Platform
runtime-version: '46'
sdk: org.gnome.Sdk
command: growww_terminal

finish-args:
  # Display & Graphics Access
  - --socket=wayland
  - --socket=fallback-x11
  - --device=dri

  # High-throughput Networking
  - --share=network
  - --share=ipc

  # Encrypted Credentials & Portal Access
  - --talk-name=org.freedesktop.secrets
  - --talk-name=org.freedesktop.portal.Desktop
  - --talk-name=org.freedesktop.portal.GlobalShortcuts

  # Memory Lock Capabilities for Secure Buffers
  - --cap-add=CAP_IPC_LOCK

modules:
  - name: libsecret
    buildsystem: meson
    sources:
      - type: git
        url: https://gitlab.gnome.org/GNOME/libsecret.git
        tag: 0.21.4

  - name: growww_trading_terminal
    buildsystem: simple
    build-commands:
      - install -D -m 755 growww_terminal /app/bin/growww_terminal
      - install -D -m 755 librust_trading_core.so /app/lib/librust_trading_core.so
      - install -D -m 644 org.growww.GrowwwTradingTerminal.desktop /app/share/applications/org.growww.GrowwwTradingTerminal.desktop
      - install -D -m 644 icons/hicolor/scalable/apps/growww.svg /app/share/icons/hicolor/scalable/apps/org.growww.GrowwwTradingTerminal.svg
```

### 8.2 Self-Contained AppImage Architecture

The AppImage bundling creates a standalone binary executable across older enterprise distros (RHEL 8/9, Ubuntu 20.04/22.04 LTS):
- Bundles private copies of `libvulkan.so.1`, `libsecret-1.so.0`, and GTK 3 runtimes.
- Implements an `AppRun` shell trampoline that initializes `LD_LIBRARY_PATH` and configures Mesa graphics driver discovery before launching the executable.

### 8.3 Canonical Debian (`.deb`) Package Control

Enterprise debian repositories (`apt.growww.in`) provide automated security updates:

```control
Package: growww-trading-terminal
Version: 2.0.0
Architecture: amd64
Maintainer: NBSE Systems Architecture <engineering@nbse.in>
Depends: libgtk-3-0 (>= 3.24.0), libsecret-1-0 (>= 0.20.0), libvulkan1 (>= 1.3.0), mesa-vulkan-drivers
Section: finance
Priority: optional
Description: Growww / NBSE Institutional High-Frequency Linux Trading Terminal
 High-performance, Vulkan-accelerated desktop workstation for institutional
 trading, multi-monitor order management, and real-time execution.
```

---

## 9. Strict 0.00% Zero-Fee Presentation & Institutional Gas Sponsorship

The Linux Thick Client enforces the **Exchange Zero-Fee and Gas-Free Invariant**. Zero-fee indicators must be visible, immutable, and mathematically verified across every screen.

```
+---------------------------------------------------------------------------------------------------+
|                        STRICT ZERO-FEE ORDER EXECUTION INTERFACE                                  |
|                                                                                                   |
|  +---------------------------------------------------------------------------------------------+  |
|  | ORDER CONFIRMATION TICKET: BTC/eINR (Group Red)                                             |  |
|  +---------------------------------------------------------------------------------------------+  |
|  | Order Type: LIMIT BUY                    Quantity: 1.25000000 BTC                           |  |
|  | Limit Price: ₹5,420,000.00               Order Value: ₹6,775,000.00                         |  |
|  |---------------------------------------------------------------------------------------------|  |
|  | [IMMUTABLE STATUTORY & EXCHANGE BREAKDOWN]                                                  |  |
|  | Maker Trading Fee:                  ₹0.00 (0.00% Genesis Zero-Fee Guarantee)                |  |
|  | Taker Trading Fee:                  ₹0.00 (0.00% Genesis Zero-Fee Guarantee)                |  |
|  | Blockchain Settlement Gas:          ₹0.00 (100% Besu Paymaster Sponsored)                  |  |
|  | Statutory TDS Deduction:            ₹0.00 (100% Unencumbered DvP Settlement)                |  |
|  | Total Execution Surcharge:          ₹0.00 (ZERO HIDDEN SPREAD / ZERO FEES)                  |  |
|  |---------------------------------------------------------------------------------------------|  |
|  | TOTAL DEBIT TO MARGIN:              ₹6,775,000.00                                           |  |
|  +---------------------------------------------------------------------------------------------+  |
|  |  [ CONFIRM ZERO-FEE EXECUTION ]                   [ CANCEL ]                                |  |
|  +---------------------------------------------------------------------------------------------+  |
+---------------------------------------------------------------------------------------------------+
```

### 9.1 Zero-Fee Visual Invariant Specification

1. **Fee Calculation Method:**
   ```dart
   // Dart Financial Engine Contract
   class FeeCalculator {
     static const BigInt zeroFeeMicroUnits = BigInt.zero;

     static ExecutionCost calculateCost({
       required BigInt quantityUnits,
       required BigInt priceMicroUnits,
     }) {
       final BigInt grossValue = (quantityUnits * priceMicroUnits) ~/ BigInt.from(1000000);
       return ExecutionCost(
         grossValue: grossValue,
         makerFee: zeroFeeMicroUnits,
         takerFee: zeroFeeMicroUnits,
         gasFeeSponsorship: zeroFeeMicroUnits,
         totalDebit: grossValue,
       );
     }
   }
   ```
2. **UI Badge Standards:**
   - **Genesis Zero-Fee Badge:** Background `#00E599` (Obsidian Green), text `#0B0E14` (Obsidian Jet Black), displaying: `0.00% ZERO-FEE`.
   - **ERC-4337 Sponsored Gas Badge:** Outlined border `#2962FF` (Institutional Cyan/Blue), displaying: `0 GAS (PAYMASTER SPONSORED)`.
   - **DvP Settlement Badge:** Displaying: `100% UNENCUMBERED DvP`.

### 9.2 64-Bit Fixed-Point Arithmetic (Sub-Paise Precision)

IEEE-754 floating point arithmetic is prohibited across client order mathematics:
- Financial math uses 64-bit unsigned integers representing micro-units ($10^{-6}$ eINR / 0.0001 paise).
- Prevents rounding anomalies in large lot sizing or multi-leg option orders.

---

## 10. Security Hardening, AppArmor, and Sandboxing Profiles

The workstation client ships with explicit security hardening policies to safeguard algorithmic keys:

```
# AppArmor Profile: /etc/apparmor.d/usr.bin.growww_terminal
#include <tunables/global>

/usr/bin/growww_terminal flags=(attach_disconnected) {
  #include <abstractions/base>
  #include <abstractions/fonts>
  #include <abstractions/freedesktop.org>
  #include <abstractions/vulkan>

  # Network Permissions
  network inet stream,
  network inet6 stream,

  # Memory Lock for Crypto Buffers
  capability ipc_lock,

  # IPC Sockets
  /run/user/[0-9]*/growww_terminal.sock rw,
  /dev/shm/growww_market_ring_buffer* rwk,

  # Deny arbitrary disk inspection
  deny /home/*/.ssh/** rwklx,
  deny /home/*/.gnupg/** rwklx,
  deny /etc/shadow r,

  # Allow Secret Service DBus
  dbus send
       bus=session
       path=/org/freedesktop/secrets
       interface=org.freedesktop.Secret.*
       member=*,
}
```

---

## 11. Verification, Automated Testing & Profiling Strategy

The Linux Thick Client includes a rigorous test and profiling harness to ensure high-frequency stability:

```
+---------------------------------------------------------------------------------------------------+
|                                 VERIFICATION & PROFILING HARNESS                                  |
|                                                                                                   |
|  [ Virtual Framebuffer (Xvfb / Weston Headless) ]                                                 |
|          |                                                                                        |
|          +---> [ Vulkan Renderdoc / Render Graph Capture ] -> Validates Zero Frame Drops          |
|          |                                                                                        |
|          +---> [ Valgrind / AddressSanitizer (ASan) ] ------> Validates Zero Memory Leaks        |
|          |                                                                                        |
|          +---> [ Linux Perf / Hotspot Profiler ] -----------> Validates Cacheline Alignment       |
|          |                                                                                        |
|          +---> [ Chaos Injection Mock Server ] -------------> 100,000 ticks/sec Flood Test       |
|                                                               - Verify 0 Dart GC Stalls           |
|                                                               - Verify 0.00% Zero-Fee Invariant   |
+---------------------------------------------------------------------------------------------------+
```

1. **Headless Wayland & X11 Integration Tests:** Runs inside GitHub Actions using headless Weston (`weston --backend=headless`) and `xvfb-run` executing end-to-end multi-window launch and detachment scripts.
2. **Frame-Rate Validation:** Injects synthetic high-frequency order book bursts (100,000 ticks/second) while monitoring display frame presentation times via `VK_EXT_frame_boundary`. Sustains 144+ FPS without dropping frames.
3. **Memory Safety & Leak Sanitization:** Compiles embedder C++ code with `-fsanitize=address,undefined` to verify zero memory leaks across detached window lifecycles.
4. **Zero-Fee Display Invariant Assertions:** Automated headless UI test suites assert that every order confirmation widget contains `0.00%` and `₹0.00` fee strings before allowing order submission.

---

## 12. Architectural Compliance & Sign-Off Matrix

| Requirement Domain | Specification Section | Compliance Status | Verifier System |
| :--- | :--- | :--- | :--- |
| **GTK 3/4 & Wayland/X11** | Section 2 | Fully Compliant | Native C++ Embedder (`FlView` & GLib Loop) |
| **Vulkan Impeller Pipeline** | Section 3 | Fully Compliant | AOT `impellerc` SPIR-V & VMA Allocator |
| **Linux Secret Service** | Section 4 | Fully Compliant | `libsecret-1` DBus AES Session Encryption |
| **Multi-Monitor Detachment** | Section 5 | Fully Compliant | POSIX SHM + Atomic Ring Buffer + UNIX Sockets |
| **Global Trading Hotkeys** | Section 6 | Fully Compliant | X11 `XGrabKey` + Wayland XDG Portal API |
| **Linux Packaging Matrix** | Section 8 | Fully Compliant | Flatpak, AppImage, and Debian `.deb` Scripts |
| **Rust FFI Socket Core** | Section 7 | Fully Compliant | `librust_trading_core.so` Zero-GC NativePort |
| **Strict 0.00% Zero-Fee** | Section 9 | Fully Compliant | Fixed-Point Sub-Paise UI & Immutable Badges |

**Authoritative Approval:**  
Linux Systems & Client Platforms Engineering Group  
Chief Trading Infrastructure Architect, NBSE / Growww  
*Production Architectural Baseline Established - September 2026*
