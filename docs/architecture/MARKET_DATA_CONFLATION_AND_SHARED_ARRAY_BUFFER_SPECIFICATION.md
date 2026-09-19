# Market Data Conflation & Shared Memory Ring Buffer Architecture Specification

**Specification ID:** SPEC-ARCH-052-MKT-CONFLATION-SHMEM  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Ultra-Low-Latency Market Data Architecture, Conflation & Shared Memory IPC  
**Target Environments:** Web Pro Terminal (Web Workers, SharedArrayBuffer, Atomics), Native Desktop Thick Client (macOS, Windows, Linux via Rust FFI and OS Shared Memory)  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Problem Statement

High-throughput cryptocurrency and derivatives trading engines frequently generate extreme bursts of market data during periods of market volatility, liquidations, and cascade order fills. Under peak conditions, the NBSE matching engine produces between 50,000 and 200,000 order book updates and trade execution reports per second per active trading pair (e.g., BTC/USDT).

### 1.1 The Client-Side Rendering Bottleneck
Modern display technology operates at fixed refresh frequencies:
- Standard 60Hz displays: 16.67ms per frame.
- High-performance 120Hz ProMotion displays: 8.33ms per frame.
- Pro trading and gaming monitors (144Hz to 240Hz): 6.94ms down to 4.17ms per frame.

Human cognitive visual processing cannot track more than 30 to 60 distinct visual state transitions per second in a complex tabular depth ladder. Ingesting raw, un-conflated tick-by-tick market data directly into the client user interface thread triggers catastrophic performance degradation:
1. **JavaScript Main-Thread Saturation:** De-serializing high-frequency JSON or Protobuf payloads on the React main thread monopolizes the event loop, causing dropped frames (jank), high input latency for order submission, and frozen user controls.
2. **Garbage Collection (GC) Pressure:** Instantiating temporary JavaScript objects (such as ladder arrays, price-level objects, and trade rows) at 100,000 allocations per second triggers frequent V8 Scavenger (Minor GC) and Full Mark-Sweep-Compact (Major GC) cycles, causing recurring 10ms to 80ms UI freezes.
3. **DOM Thrashing & Reconciliation Cascades:** Rapid React reconciliation passes over thousands of DOM elements or Canvas redraw calls overwhelm the browser composition pipeline.

### 1.2 The Sovereign Dual-Domain Architecture
To achieve zero frame drops, locked 60/120 FPS rendering, and deterministic sub-millisecond data propagation, the platform establishes two zero-copy, shared-memory processing pipelines:
- **Web Browser Domain (Web Pro Terminal):** A dedicated background Web Worker manages the inbound WebSocket binary stream, performs price-level conflation, updates a persistent Level 2 order book snapshot, and writes state directly into an off-heap `SharedArrayBuffer` using atomic memory operations (`Atomics`). The React 19 main thread and WebGL canvas pipelines read from this shared memory block with zero object allocations and zero serialization overhead.
- **Desktop Thick Client Domain (macOS, Windows, Linux):** A background daemon written in Rust (`rust_trading_core`) connects directly to raw multicast/WebSocket feeds, executes SIMD-accelerated delta reconstruction, and manages an operating-system-level memory-mapped ring buffer (`mmap` via POSIX SHM on macOS/Linux; `CreateFileMappingW` on Windows). Detached child windows, floating charts, and Depth of Market (DOM) panels attach to this shared memory block simultaneously across native OS processes.

```
+----------------------------------------------------------------------------------------------------+
| NBSE ZERO-COPY DUAL-DOMAIN MARKET DATA ARCHITECTURE                                                |
|                                                                                                    |
|  [ Matching Engine Core ] (200,000 updates/sec)                                                    |
|             |                                                                                      |
|             v                                                                                      |
|  [ Edge WebSocket Gateway Swarm ] (Protobuf Binary Framing)                                        |
|             |                                                                                      |
|             +--------------------------------------+                                               |
|             | (TLS WSS Stream)                     | (Direct Colocation / TCP)                     |
|             v                                      v                                               |
|  +------------------------------------+  +------------------------------------------------------+  |
|  | BROWSER / WEB PRO TERMINAL         |  | DESKTOP THICK CLIENT (macOS / Win / Linux)           |  |
|  |                                    |  |                                                      |  |
|  |  [ Dedicated Web Worker ]          |  |  [ Native Rust FFI Core Daemon ]                     |  |
|  |  - Binary Protobuf Ingestion       |  |  - Zero-Copy SBE / Protobuf Ingestion                |  |
|  |  - 50ms Conflation Engine          |  |  - 50ms SIMD Delta Conflation Engine                 |  |
|  |  - L2/L3 Delta Reconstruction      |  |  - Lock-Free SPMC Circular Ring Buffer               |  |
|  |             |                      |  |             |                                        |  |
|  |             v                      |  |             v                                        |  |
|  |   SharedArrayBuffer (SAB)          |  |   OS Memory-Mapped Shared Memory (POSIX SHM / Win32) |  |
|  |   [Atomics SPSC Lock-Free Queues]  |  |   [Aligned 64-byte Cache-Line Ring & Seqlocks]       |  |
|  |             |                      |  |             +-------------------+--------------------+  |
|  |             v                      |  |             |                   |                    |  |
|  |  [ React 19 UI Main Thread ]       |  |             v                   v                    v  |
|  |  - Zero-GC TypedArray Views        |  |     [ Main Window ]    [ Pop-Out Chart ]   [ DOM Ladder] |  |
|  |  - 20 FPS Conflated DOM Render     |  |     (Flutter Metal)    (DirectX 12 / Vulkan)(Native Cocoa|  |
|  |  - WebGL / Canvas Order Book       |  |     (120Hz ProMotion)  (144Hz - 360Hz)     (Win32 Child) |  |
|  +------------------------------------+  +------------------------------------------------------+  |
+----------------------------------------------------------------------------------------------------+
```

---

## 2. High-Frequency Tick Conflation Mechanics

Conflation is the deterministic process of merging multiple discrete market data state modifications occurring within a predefined time window into a single consolidated update. 

### 2.1 The 50ms Conflation Interval & 20 Hz Frame Budget
The system enforces a strict conflation window of **50 milliseconds**, which caps UI updates at exactly **20 dispatches per second** per instrument.

```
Incoming Stream (Raw Fills & Deltas):
T00ms    T07ms    T14ms    T22ms    T31ms    T44ms    T49ms   | T50ms Conflation Boundary
 [D1] ---- [D2] ---- [D3] ---- [D4] ---- [D5] ---- [D6] ---- [D7]   |
                                                               |
  --> Accumulated and coalesced in Worker/Rust Memory           |
                                                               v
Emitted Conflated Frame to UI Ring Buffer:                    [ FRAME_01 ] (Net delta state)
```

#### Mathematical Properties of Conflation:
1. **Volume Conservation:** The cumulative volume of trades executed within interval $[T, T + \Delta t)$ is strictly preserved:
   $$V_{\text{interval}} = \sum_{i=1}^{N} v_i$$
2. **High/Low Extrema Preservation:**
   $$P_{\text{high}} = \max(p_1, p_2, \dots, p_N), \quad P_{\text{low}} = \min(p_1, p_2, \dots, p_N)$$
3. **Last Traded Price (LTP) Monotonicity:** The conflated LTP reflects the price $p_N$ of the latest valid chronological fill within the window.
4. **Idempotent Ladder Coalescing:** Multiple price-level replacements at price $P_k$ resolve to the final quantity state $Q_k(T + \Delta t)$. If $Q_k(T + \Delta t) = 0$, a single `DELETE` operation is emitted.

### 2.2 Micro-Batching vs. Time-Sliced Dirty Flagging
The conflation engine avoids maintaining dynamic queues or allocations during active trading. Instead, it utilizes a pre-allocated fixed-capacity Sparse Set and Dirty Bitfield:

```
+-------------------------------------------------------------------------------------------+
| CONFLATION DIRTY BITFIELD AND FIXED ARRAY COALESCING                                      |
|                                                                                           |
|  Dirty Bitmask (Uint32Array): [ 0b00000000000000000000000000000101 ] (Levels 0 and 2)    |
|                                                                                           |
|  Price Index Mapping:                                                                     |
|  Slot 0: Price = 65420.50 | Net Qty = 14.852 | Action = UPDATE (Dirty Bit 0 set)          |
|  Slot 1: Price = 65420.00 | Net Qty = 02.100 | Action = UNCHANGED (Dirty Bit 1 clear)      |
|  Slot 2: Price = 65419.50 | Net Qty = 00.000 | Action = DELETE (Dirty Bit 2 set)          |
|  ...                                                                                      |
|  Slot 49: Price = 65395.00| Net Qty = 50.120 | Action = UNCHANGED                         |
+-------------------------------------------------------------------------------------------+
```

#### Coalescing Execution Flow:
1. When a delta arrives at $T_{\text{tick}}$, the engine computes the ladder index for the affected price level.
2. The quantity is updated in-place inside the pre-allocated staging table.
3. The corresponding bit in the 64-bit dirty mask is flagged (`dirty_mask |= (1ULL << index)`).
4. A high-resolution monotonic timer tracks the elapsed interval:
   - In Web Workers: Checked via `performance.now()`.
   - In Rust Core: Evaluated via CPU clock cycles or `clock_gettime(CLOCK_MONOTONIC_RAW)`.
5. Upon reaching the 50ms boundary, the engine scans the dirty bitfield using hardware-accelerated Count Leading Zeros (`clz`) or Find First Set (`ffs`) instructions.
6. Only modified levels are copied into the outgoing shared ring buffer slot.
7. The dirty bitfield is reset to zero (`0x0`) in a single CPU instruction.

### 2.3 Monotonic Sequence Integrity and Anti-Drift Timers
Standard browser `setInterval` and `setTimeout` timers are subject to event-loop latency and thread jitter, causing intervals to drift from 50ms to 75ms or 110ms under load.

To eliminate timer drift:
- In the Web Worker, the conflation loop utilizes a self-correcting recursive `requestAnimationFrame`-style scheduling model or microsecond sleep loop built on top of high-resolution timestamps:
  $$\Delta t_{\text{next}} = \max(0, 50.0 - (\text{now}() - T_{\text{target}}))$$
- In the Desktop Rust Core, the thread sleeps using high-precision POSIX timer file descriptors (`timerfd_create` with `CLOCK_MONOTONIC`) or Windows Multimedia High-Resolution Timers (`timeBeginPeriod(1)` with `CreateWaitableTimerExW`).

---

## 3. SharedArrayBuffer Layout & Atomics Synchronization (Web Pro Terminal)

In the browser, the Web Worker and the React main thread share a single, non-transferable, fixed-size `SharedArrayBuffer` (SAB). This design circumvents the structured clone algorithm used by `postMessage`, which incurs high memory serialization and deserialization costs on every frame.

### 3.1 Cross-Origin Isolation Prerequisites
To instantiate a `SharedArrayBuffer` in modern Chromium, WebKit, and Gecko browsers, the web application must be delivered with strict Cross-Origin Isolation HTTP headers:
```http
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Resource-Policy: same-origin
```

### 3.2 Exact Byte-Level Memory Layout
The shared memory segment is allocated as a contiguous 4,194,304-byte (4 MB) buffer, aligned to 64-byte CPU cache line boundaries.

```
+----------------------------------------------------------------------------------------------------+
| SHARED ARRAY BUFFER PHYSICAL MEMORY MAP (4 MB TOTAL CAPACITY)                                      |
|                                                                                                    |
| Offset (Hex)     Offset (Dec)    Size (Bytes)   Section Description                                |
| -------------------------------------------------------------------------------------------------  |
| 0x00000000       0               64 B           Control Header & Atomic Sync Block (Cache Line 0)  |
| 0x00000040       64              64 B           Worker Heartbeat & Metrics Block (Cache Line 1)    |
| 0x00000080       128             1920 B         Reserved Control Metadata & Checksums              |
| 0x00000800       2048            16,384 B       L2 Snapshot Cache: Double-Buffered Bids Ladder     |
| 0x00004800       18432           16,384 B       L2 Snapshot Cache: Double-Buffered Asks Ladder     |
| 0x00008800       34816           4096 B         24-Hour Ticker & Instrument Statistics Block       |
| 0x00009800       38912           3,145,728 B    Conflated SPSC Ring Buffer Slots (1024 Slots x 3KB)|
| 0x00319800       3250176         944,128 B      Trade Blotter Circular Queue & Unallocated Slack   |
+----------------------------------------------------------------------------------------------------+
```

#### Cache-Line 0: Atomic Control Block Layout (Offsets 0x00 - 0x3F)
All variables are strictly aligned to prevent false sharing across CPU cores:

| Offset | Type | Identifier | Alignment | Functional Description |
| :--- | :--- | :--- | :--- | :--- |
| `0x00` | `Uint32` | `MAGIC_HEADER` | 4 bytes | Constant `0x47524F57` ("GROW" signature verification). |
| `0x04` | `Uint32` | `SCHEMA_VERSION` | 4 bytes | Binary protocol version (e.g., `0x00010000` = v1.0.0). |
| `0x08` | `Uint32` | `RING_HEAD` | 4 bytes | Atomic write index updated by Web Worker producer. |
| `0x0C` | `Uint32` | `RING_TAIL` | 4 bytes | Atomic read index updated by React UI consumer. |
| `0x10` | `BigUint64`| `GLOBAL_SEQ_NO` | 8 bytes | Monotonically increasing sequence number from matching engine. |
| `0x18` | `Uint32` | `ACTIVE_BUFFER_L2`| 4 bytes | Double-buffer selector for L2 snapshot (`0` = Buffer A, `1` = Buffer B). |
| `0x1C` | `Uint32` | `DROPPED_FRAMES` | 4 bytes | Counter incremented when ring buffer overflows (slow UI consumer). |
| `0x20` | `Int32` | `FUTEX_LOCK` | 4 bytes | Futex word used for `Atomics.wait` / `Atomics.notify`. |
| `0x24` | `Uint32` | `CONFLATION_COUNT`| 4 bytes | Running total of conflation frames emitted. |
| `0x28` | `Uint8[24]`| `PADDING_CACHE_0`| 1 byte | Alignment padding to round to exactly 64 bytes. |

#### Order Book Double-Buffered Ladder Layout (Offsets 0x0800 - 0x87FF)
To enable zero-copy snapshot reads by the UI without tearing, the Level 2 depth table uses double buffering (Buffer A and Buffer B) governed by the atomic pointer `ACTIVE_BUFFER_L2`:
- Each price level occupies **32 bytes**:
  - `price`: `Float64` (8 bytes, IEEE 754 double precision).
  - `quantity`: `Float64` (8 bytes, scaled cumulative base asset volume).
  - `order_count`: `Uint32` (4 bytes, number of discrete orders resting at this price).
  - `flags`: `Uint32` (4 bytes, bit 0: is_new, bit 1: updated, bit 2: liquidated).
  - `reserved`: `Uint64` (8 bytes, alignment and future synthetic spread data).
- Depth capacity: 50 Bids + 50 Asks = 100 levels per buffer.
- Memory size per buffer: $100 \times 32\text{ bytes} = 3,200\text{ bytes}$.

### 3.3 Atomics SPSC Synchronization Architecture
The communication protocol between the Web Worker (Producer) and the React UI Thread (Consumer) is structured as a lock-free Single-Producer Single-Consumer (SPSC) circular queue.

```
Web Worker Thread (Producer):
1. Ingest WebSocket Binary Packet.
2. Accumulate deltas within 50ms interval.
3. Obtain target ring slot: index = Atomics.load(RING_HEAD) & (RING_CAPACITY - 1).
4. Check capacity: (index - Atomics.load(RING_TAIL)) < RING_CAPACITY.
   - If full: increment DROPPED_FRAMES, drop frame or overwrite oldest according to policy.
5. Populate slot memory with conflated updates.
6. Memory Barrier: Atomics.store(RING_HEAD, head + 1) with Release semantics.
7. Optional notification: Atomics.notify(FUTEX_LOCK, 1) (if secondary consumer thread is waiting).

React Main Thread (Consumer):
1. Execution tied to requestAnimationFrame loop (locked to display refresh rate).
2. Load producer head: const head = Atomics.load(RING_HEAD).
3. Load consumer tail: const tail = Atomics.load(RING_TAIL).
4. If head === tail: No new conflated frames; skip rendering.
5. While tail < head:
   - Read frame data directly using pre-allocated Float64Array view.
   - Update WebGL / Canvas vertex buffers in-place.
   - Increment tail.
6. Commit read pointer: Atomics.store(RING_TAIL, head) with Release semantics.
```

> [!IMPORTANT]
> The browser runtime explicitly disallows calling `Atomics.wait()` on the main browser thread to prevent locking the UI. Consequently, the React main thread **never calls `Atomics.wait()`**. Instead, the main thread reads the ring buffer via atomic loads during the standard `requestAnimationFrame` tick. Non-main worker threads (such as auxiliary export or logging workers) are permitted to use `Atomics.wait()`.

### 3.4 Production Architectural Implementation Patterns

#### Web Worker Producer Engine (TypeScript / JavaScript):
```typescript
// worker_conflation_producer.ts
export const SAB_BYTE_LENGTH = 4 * 1024 * 1024; // 4 MB
export const RING_CAPACITY = 1024;
export const SLOT_SIZE = 3072; // 3 KB per conflated frame

// Offset Constants
export const OFFSET_MAGIC = 0;
export const OFFSET_RING_HEAD = 2; // Int32 index (offset 8 bytes)
export const OFFSET_RING_TAIL = 3; // Int32 index (offset 12 bytes)
export const OFFSET_GLOBAL_SEQ = 2; // BigInt64 index (offset 16 bytes)
export const OFFSET_ACTIVE_L2 = 7; // Int32 index (offset 28 bytes)
export const OFFSET_DROPPED = 8; // Int32 index (offset 32 bytes)
export const OFFSET_RING_PAYLOAD = 38912;

export class MarketDataWorkerProducer {
  private sab: SharedArrayBuffer;
  private int32View: Int32Array;
  private bigIntView: BigInt64Array;
  private float64View: Float64Array;
  private conflationIntervalMs: number = 50.0;
  private lastFlushTimestamp: number = 0;
  private pendingDeltasCount: number = 0;

  constructor(sharedBuffer: SharedArrayBuffer) {
    this.sab = sharedBuffer;
    this.int32View = new Int32Array(this.sab);
    this.bigIntView = new BigInt64Array(this.sab);
    this.float64View = new Float64Array(this.sab);
    this.lastFlushTimestamp = performance.now();
  }

  public onDirectWebSocketMessage(binaryFrame: ArrayBuffer): void {
    // 1. Zero-copy binary parse of inbound Protobuf / FlatBuffers frame
    this.accumulateDeltaInStaging(binaryFrame);
    this.pendingDeltasCount++;

    const now = performance.now();
    if (now - this.lastFlushTimestamp >= this.conflationIntervalMs) {
      this.flushConflatedFrame(now);
      this.lastFlushTimestamp = now;
      this.pendingDeltasCount = 0;
    }
  }

  private flushConflatedFrame(timestamp: number): void {
    const head = Atomics.load(this.int32View, OFFSET_RING_HEAD);
    const tail = Atomics.load(this.int32View, OFFSET_RING_TAIL);

    // Bounded Ring Buffer Overflow Check
    if (head - tail >= RING_CAPACITY) {
      Atomics.add(this.int32View, OFFSET_DROPPED, 1);
      // Policy: Advance tail to discard oldest unconsumed frame
      Atomics.store(this.int32View, OFFSET_RING_TAIL, tail + 1);
    }

    const slotIndex = head & (RING_CAPACITY - 1);
    const slotByteOffset = OFFSET_RING_PAYLOAD + slotIndex * SLOT_SIZE;

    // Write conflated L2 levels directly into SharedArrayBuffer slot
    this.writeSlotPayload(slotByteOffset, timestamp);

    // Release memory barrier: ensure all writes are visible before advancing head
    Atomics.store(this.int32View, OFFSET_RING_HEAD, head + 1);
  }

  private accumulateDeltaInStaging(frame: ArrayBuffer): void {
    // In-place coalescing inside pre-allocated staging arrays
  }

  private writeSlotPayload(byteOffset: number, timestamp: number): void {
    // Direct Float64Array and Int32Array writes into SAB slot
  }
}
```

#### React UI Thread Consumer (Hook & WebGL Integration):
```typescript
// useOrderBookSharedBuffer.ts
import { useEffect, useRef } from "react";

export function useOrderBookSharedBuffer(sab: SharedArrayBuffer) {
  const int32ViewRef = useRef<Int32Array | null>(null);
  const float64ViewRef = useRef<Float64Array | null>(null);
  const lastProcessedTailRef = useRef<number>(0);

  useEffect(() => {
    int32ViewRef.current = new Int32Array(sab);
    float64ViewRef.current = new Float64Array(sab);

    let animationFrameId: number;

    const renderTick = () => {
      const int32 = int32ViewRef.current;
      const float64 = float64ViewRef.current;

      if (int32 && float64) {
        const head = Atomics.load(int32, OFFSET_RING_HEAD);
        const tail = lastProcessedTailRef.current;

        if (head !== tail) {
          // Process latest conflated frame (skip intermediate if multiple frames arrived)
          const latestSlotIndex = (head - 1) & (RING_CAPACITY - 1);
          const slotByteOffset = OFFSET_RING_PAYLOAD + latestSlotIndex * SLOT_SIZE;

          // ZERO-GC RENDER: Pass TypedArray view slice directly to WebGL vertex buffer
          updateCanvasDepthLadder(sab, slotByteOffset);

          // Update tail to latest head
          lastProcessedTailRef.current = head;
          Atomics.store(int32, OFFSET_RING_TAIL, head);
        }
      }

      animationFrameId = requestAnimationFrame(renderTick);
    };

    animationFrameId = requestAnimationFrame(renderTick);
    return () => cancelAnimationFrame(animationFrameId);
  }, [sab]);
}

function updateCanvasDepthLadder(buffer: SharedArrayBuffer, offset: number): void {
  // WebGL gl.bufferSubData fast path (zero JS object allocations)
}
```

---

## 4. Rust FFI Memory-Mapped Ring Buffer on Desktop (macOS, Windows, Linux)

On desktop platforms (packaged via Flutter Desktop with native AppKit, Win32, and GTK wrappers), performance requirements are heightened. Traders operate multi-monitor arrays with detached child windows (Floating 4K Charts, Order Book Ladders, Click-to-Trade DOMs, and Execution Blotters).

Using standard inter-process communication (like JSON-RPC over TCP sockets or named pipes) causes serialization bottlenecks when broadcasting 200,000 updates/second to multiple detached windows. The desktop platform solves this by running a shared Rust core daemon that creates an operating-system-level **Memory-Mapped Circular Ring Buffer**.

```
+----------------------------------------------------------------------------------------------------+
| CROSS-WINDOW OS SHARED MEMORY ARCHITECTURE (DESKTOP)                                              |
|                                                                                                    |
|  [ Inbound WebSocket / Multicast Stream ]                                                          |
|                    |                                                                               |
|                    v                                                                               |
|  +----------------------------------------------------------------------------------------------+  |
|  | BACKGROUND RUST ENGINE (PRODUCER PROCESS / ISOLATE)                                          |  |
|  | - SBE / Protobuf Binary Unpack                                                               |  |
|  | - SIMD Price-Level Sorting & Conflation (50ms)                                               |  |
|  | - Write to Memory-Mapped Segment with Release Barrier                                        |  |
|  +----------------------------------------------------------------------------------------------+  |
|                    |                                                                               |
|                    v                                                                               |
|  +----------------------------------------------------------------------------------------------+  |
|  | OPERATING SYSTEM SHARED MEMORY SEGMENT (`/nbse_market_data_btc_usdt`)                        |  |
|  | - POSIX: `shm_open()` + `mmap()` (macOS / Linux)                                             |  |
|  | - Windows: `CreateFileMappingW()` + `MapViewOfFile()`                                       |  |
|  | - Seqlock-Protected L2 Depth Snapshot Tables (100% Lock-Free)                                |  |
|  | - Aligned 64-byte Cache-Line Architecture (Zero False Sharing)                               |  |
|  +----------------------------------------------------------------------------------------------+  |
|         |                                      |                                      |            |
|         v                                      v                                      v            |
|  +-----------------------------+  +-----------------------------+  +-----------------------------+ |
|  | OS WINDOW 1: MAIN TERMINAL  |  | OS WINDOW 2: DETACHED CHART |  | OS WINDOW 3: DETACHED DOM   | |
|  | Flutter Impeller / Metal    |  | Flutter / DirectX 12        |  | Native AppKit / Win32 Child | |
|  | Zero-Copy C-ABI Dart FFI    |  | Zero-Copy C-ABI Dart FFI    |  | Zero-Copy C-ABI Dart FFI    | |
|  +-----------------------------+  +-----------------------------+  +-----------------------------+ |
+----------------------------------------------------------------------------------------------------+
```

### 4.1 Operating System Inter-Process Primitives
The implementation abstracts platform-specific OS APIs under a unified Rust interface:
- **macOS (Darwin) & Linux (POSIX):**
  - File descriptor creation: `shm_open(name, O_CREAT | O_RDWR, 0666)`.
  - Allocation sizing: `ftruncate(fd, size)`.
  - Memory mapping: `mmap(ptr::null_mut(), size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0)`.
  - Cleanup: `munmap()` and `shm_unlink()`.
- **Windows (Win32 API):**
  - Mapping object creation: `CreateFileMappingW(INVALID_HANDLE_VALUE, ... PAGE_READWRITE, 0, size, name)`.
  - View mapping: `MapViewOfFile(handle, FILE_MAP_ALL_ACCESS, 0, 0, size)`.
  - Cleanup: `UnmapViewOfFile()` and `CloseHandle()`.

### 4.2 Cache-Line Alignment & NUMA Optimization
To achieve microsecond delivery without cache contention across CPU cores:
1. Every critical atomic index (`head`, `tail`, `sequence_number`, `seqlock_version`) is aligned to **64 bytes** using Rust's `#[repr(align(64))]`.
2. This eliminates **False Sharing**, which occurs when two threads running on different physical cores modify adjacent variables residing on the same 64-byte L1/L2 cache line, forcing bus invalidation cycles.
3. Memory accesses use explicit compiler intrinsics with `Ordering::Acquire`, `Ordering::Release`, and `Ordering::Relaxed` semantics.

### 4.3 Lock-Free Single-Producer Multi-Consumer (SPMC) Seqlock Algorithm
When multiple independent OS windows consume the same Level 2 depth table, traditional mutexes induce lock contention and thread priority inversion. The platform implements an atomic **Seqlock** (Sequence Lock) for the Level 2 ladder snapshot:

```rust
// Architectural Representation of Seqlock Synchronized Depth Table
#[repr(C, align(64))]
pub struct SeqlockOrderBook {
    pub sequence_version: AtomicU64, // Odd = Producer Writing; Even = Consistent
    pub timestamp_ns: AtomicU64,
    pub bids: [MarketLevelRaw; 50],
    pub asks: [MarketLevelRaw; 50],
}

#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct MarketLevelRaw {
    pub price_raw: i64,      // Fixed-point scaled by 1e8 (sub-satoshi / sub-paise)
    pub quantity_raw: i64,   // Fixed-point scaled by 1e8
    pub order_count: u32,
    pub flags: u32,
}
```

#### Seqlock Read/Write Mechanics:
- **Writer (Rust Producer Daemon):**
  1. Increment `sequence_version` by 1 using `Ordering::Release` (transitions even $\to$ odd, signaling a write is in progress).
  2. Issue a write memory fence: `std::sync::atomic::fence(Ordering::Release)`.
  3. Overwrite price ladder arrays in shared memory.
  4. Issue another write memory fence: `std::sync::atomic::fence(Ordering::Release)`.
  5. Increment `sequence_version` by 1 again (transitions odd $\to$ even, signaling write complete).
  6. The writer is **100% wait-free** and is never blocked by reader processes.
- **Reader (Detached Windows / Flutter FFI):**
  1. Read `sequence_version` with `Ordering::Acquire`.
  2. If `sequence_version` is odd, a write is in progress: yield CPU (`std::hint::spin_loop()`) and retry.
  3. Read the order book data directly into local render buffers.
  4. Issue an acquire memory fence: `std::sync::atomic::fence(Ordering::Acquire)`.
  5. Re-read `sequence_version`. If it does not equal the initial version, torn read occurred: discard local buffer and retry.
  6. Because the conflation interval is 50ms and memory copying takes sub-microsecond time, collisions occur in less than 0.001% of read attempts.

### 4.4 Rust Core FFI Export Interface (C-ABI)
The compiled shared library (`librust_trading_core.so` on Linux, `.dylib` on macOS, `.dll` on Windows) exports a stable, zero-allocation C-ABI for direct integration with Dart FFI:

```rust
// librust_trading_core/src/ffi.rs
use std::ffi::c_char;

#[repr(C)]
pub struct L2SnapshotView {
    pub sequence_number: u64,
    pub timestamp_ns: u64,
    pub bids_ptr: *const MarketLevelRaw,
    pub bids_len: u32,
    pub asks_ptr: *const MarketLevelRaw,
    pub asks_len: u32,
}

#[no_mangle]
pub unsafe extern "C" fn nbse_shm_attach(
    symbol_ptr: *const c_char,
    is_writable: bool,
) -> *mut std::ffi::c_void {
    // Platform-specific shm_open / CreateFileMappingW
    // Returns direct pointer to mapped virtual memory address
    std::ptr::null_mut()
}

#[no_mangle]
pub unsafe extern "C" fn nbse_shm_poll_l2_seqlock(
    shm_handle: *mut std::ffi::c_void,
    out_snapshot: *mut L2SnapshotView,
    retry_max: u32,
) -> bool {
    // Executes wait-free Seqlock read loop without memory allocations
    true
}

#[no_mangle]
pub unsafe extern "C" fn nbse_shm_detach(shm_handle: *mut std::ffi::c_void) -> bool {
    // munmap / UnmapViewOfFile and close handles cleanly
    true
}
```

---

## 5. Level 2 and Level 3 Order Book Reconstruction Algorithms

The client platform supports dual market data fidelities:
1. **Level 2 (Aggregated Price-Point Depth):** 20-depth for compact mobile and dashboard ladders; 50-depth for Pro DOM ladder charting.
2. **Level 3 (Order-by-Order Granularity):** Individual resting order tracking for institutional algorithmic traders and depth-delta visualizers.

```
+----------------------------------------------------------------------------------------------------+
| ORDER BOOK RECONSTRUCTION STATE MACHINE                                                            |
|                                                                                                    |
|    +-------------------------+                                                                     |
|    |      DISCONNECTED       |                                                                     |
|    +-------------------------+                                                                     |
|                 | (WebSocket Connect)                                                              |
|                 v                                                                                  |
|    +-------------------------+                                                                     |
|    |  BUFFERING_DELTAS       | <--------------------------------------------+                      |
|    |  Queue live ticks       |                                              |                      |
|    +-------------------------+                                              |                      |
|                 | (Request REST / Dedicated WS Snapshot)                    |                      |
|                 v                                                           |                      |
|    +-------------------------+                                              |                      |
|    |  PROCESSING_SNAPSHOT    |                                              |                      |
|    |  Base Seq: S_base       |                                              |                      |
|    +-------------------------+                                              |                      |
|                 |                                                           |                      |
|                 |-- Gap Detected: First Delta Seq > S_base + 1 ------------>| (Resync Waterfall)   |
|                 |                                                           |                      |
|                 v (Sequence Alignment Validated)                            |                      |
|    +-------------------------+                                              |                      |
|    |  APPLYING_BACKLOG       |                                              |                      |
|    |  Discard Seq <= S_base  |                                              |                      |
|    +-------------------------+                                              |                      |
|                 | (Backlog Exhausted)                                       |                      |
|                 v                                                           |                      |
|    +-------------------------+                                              |                      |
|    |  SYNCHRONIZED_STREAMING |                                              |                      |
|    |  Continuous Conflation  |                                              |                      |
|    +-------------------------+                                              |                      |
|                 |                                                           |                      |
|                 +-- Sequence Break: delta.prev_seq != local_seq ----------->+                      |
|                 +-- Checksum Mismatch: CRC32(L2) != delta.checksum -------->+                      |
+----------------------------------------------------------------------------------------------------+
```

### 5.1 Level 2 (20-Depth & 50-Depth) Deterministic Synchronization Protocol
To reconstruct an authoritative Level 2 order book from an asynchronous WebSocket stream, the client strictly executes the following sequence:

#### Step 1: Ingestion & Buffer Queuing
Upon channel subscription, the client immediately begins buffering incoming delta messages into a pre-allocated staging ring without applying them:
```json
{
  "type": "depth_delta",
  "symbol": "BTC_USDT",
  "first_sequence_id": 10842001,
  "last_sequence_id": 10842004,
  "prev_sequence_id": 10842000,
  "timestamp_ns": 1789804800050000000,
  "bids": [["65420.50", "1.25000000"]],
  "asks": [["65422.00", "0.00000000"]]
}
```

#### Step 2: Snapshot Retrieval
The client simultaneously issues an expedited request for a full Level 2 snapshot across a dedicated low-latency channel:
```json
{
  "type": "depth_snapshot",
  "symbol": "BTC_USDT",
  "sequence_id": 10842002,
  "timestamp_ns": 1789804800025000000,
  "bids": [
    ["65420.50", "2.50000000"],
    ["65420.00", "5.10000000"]
  ],
  "asks": [
    ["65421.50", "0.85000000"],
    ["65422.00", "1.10000000"]
  ],
  "checksum": 3294821045
}
```

#### Step 3: Sequence Number Verification & Pruning
When the snapshot arrives with base sequence $S_{\text{base}}$:
1. Scan buffered deltas: Discard any delta where $\text{last\_sequence\_id} \le S_{\text{base}}$.
2. Validate alignment: Find the first delta where $\text{first\_sequence\_id} \le S_{\text{base}} + 1 \le \text{last\_sequence\_id}$.
3. If no buffered delta covers $S_{\text{base}} + 1$, an unrecoverable gap exists. Discard all state and restart the synchronization waterfall.

#### Step 4: Continuous Delta Application & In-Place Array Mutation
For every confirmed delta packet:
1. Verify contiguous sequence monotonicity: $\text{delta.prev\_sequence\_id} == \text{current\_local\_sequence\_id}$.
2. Iterate through `bids` and `asks`:
   - If `quantity > 0`: Search for price level. If found, update quantity; if not found and within depth limits, insert level maintaining descending price order for Bids, ascending for Asks.
   - If `quantity == 0`: Delete price level and shift subsequent entries left.
3. Truncate ladder arrays to exact depth limit (20 or 50 levels).

### 5.2 Real-Time Checksum Verification Algorithm
To protect against subtle memory corruption, out-of-order execution, or dropped packets, the matching engine transmits a CRC32 / xxHash64 checksum with every snapshot and conflated delta packet.

#### Checksum Construction Standard:
1. Extract the top 20 or 50 bids and asks.
2. Format price and quantity as string literals without trailing zeros: `"<price>:<qty>"`.
3. Concatenate alternating bid and ask levels into an evaluation string:
   $$\text{CRC\_INPUT} = \text{bid\_p}_1 + ":" + \text{bid\_q}_1 + ":" + \text{ask\_p}_1 + ":" + \text{ask\_q}_1 + \dots$$
4. Compute IEEE 802.3 CRC32 integer hash.
5. If $\text{CRC32}_{\text{computed}} \ne \text{CRC32}_{\text{packet}}$, the order book state is corrupt. The engine drops the ladder, flags UI telemetry, and automatically re-fetches a full snapshot.

### 5.3 Level 3 (Order-by-Order) Reconstruction Engine
Level 3 streams emit individual order lifecycle events directly from the matching engine's order book:
- `ORDER_ADD`: Insert new order at price level `(order_id, side, price, quantity)`.
- `ORDER_MODIFY`: In-place size reduction or priority-losing price update `(order_id, new_quantity)`.
- `ORDER_EXECUTE`: Trade execution against resting order `(order_id, fill_quantity, match_id)`.
- `ORDER_CANCEL`: Order cancellation `(order_id, remaining_quantity)`.

#### Data Structures for Zero-Allocation Level 3 Reconstruction:
1. **Order Repository:** Pre-allocated Flat Hash Map (e.g., Google `swiss_table` in Rust or typed index pool in Web Workers) mapping `order_id` (Uint64) to `OrderNode`:
   ```rust
   #[repr(C)]
   pub struct OrderNode {
       pub order_id: u64,
       pub price: i64,
       pub quantity: i64,
       pub side: u8, // 0 = Bid, 1 = Ask
       pub next_order_idx: u32,
       pub prev_order_idx: u32,
   }
   ```
2. **Aggregated Price Level Bucketing:** Price levels maintain a doubly-linked list of resting orders. Price-level aggregate volume is updated in $O(1)$ time upon each order modification.
3. **Conflated Level 3 Emission:** Rather than rendering 50,000 order additions/cancellations directly, the engine computes Level 2 depth projections at each 50ms interval, ensuring complete mathematical consistency while sparing the rendering pipeline.

---

## 6. Memory Benchmarks & Zero-GC Execution Hot Paths

The fundamental engineering objective of this architecture is **Zero Garbage Collection (0 KB/sec GC allocation)** on hot ingestion paths across both Web and Desktop platforms.

### 6.1 JavaScript Engine (V8) Zero-Allocation Directives
To achieve zero memory allocations in the Web Pro Terminal:
1. **Strict Prohibition of Dynamic Object Creation:** No object literals (`{}`), arrays (`[]`), or closures are allocated inside message handlers.
2. **No Array Transformation Methods:** Functions such as `.map()`, `.filter()`, `.reduce()`, and `.slice()` are banned on the hot path because they allocate intermediate Array instances.
3. **Pre-Allocated Static Pools:** All working structures (staging tables, binary decoding views, parsing cursors) are instantiated once during worker startup.
4. **Direct TypedArray Buffer Manipulation:** Inbound Protobuf data is parsed using zero-copy cursor readers directly over the incoming `Uint8Array` buffer.

```
+----------------------------------------------------------------------------------------------------+
| ZERO-GC PIPELINE BENCHMARKS & ALLOCATION PROFILE                                                  |
|                                                                                                    |
| Ingestion Phase             Standard Implementation        NBSE Zero-Copy SAB Engine               |
| -------------------------------------------------------------------------------------------------  |
| JSON / Binary Deserialization: 24,000 KB / sec (GC Eden)    0 KB / sec (Direct SAB TypedArray)     |
| Order Book Delta Processing:  18,500 KB / sec (GC Arrays)  0 KB / sec (In-Place Array Updates)    |
| UI Thread Cross-Post:         12,000 KB / sec (Structured)  0 KB / sec (Shared Memory Pointer)     |
| React Component State Update:  8,500 KB / sec (VirtualDOM)  0 KB / sec (Canvas gl.bufferSubData)   |
| -------------------------------------------------------------------------------------------------  |
| Total Heap Allocation Rate:   63,000 KB / sec               0.000 KB / sec (LOCKED ZERO-GC)        |
| V8 Scavenger GC Pause Time:   12ms - 45ms per 2 seconds     0.000 ms (Zero GC Pauses)              |
| Frame Time Consistency:       Erratic (Drops to 15 FPS)     Locked 60 / 120 FPS Flatline           |
+----------------------------------------------------------------------------------------------------+
```

### 6.2 Latency Distribution Profile & Performance Targets
Benchmarked under simulated market cascade loads of **250,000 ticks/sec sustained** and **500,000 ticks/sec burst**:

| Benchmark Metric | Traditional React/WebSocket | Desktop Rust Core (POSIX/Win32 SHM) | Web Pro Terminal (Worker + SAB) |
| :--- | :--- | :--- | :--- |
| **Ingestion Heap Allocation** | 63 MB / second | **0.00 bytes** | **0.00 bytes** |
| **P50 Latency (Wire to Shm)** | 8.42 ms | **0.04 ms (40 microseconds)** | **0.42 ms (420 microseconds)** |
| **P90 Latency** | 24.10 ms | **0.12 ms** | **0.95 ms** |
| **P99 Latency** | 78.50 ms | **0.38 ms** | **1.85 ms** |
| **P99.9 Latency** | 240.00 ms (GC Spikes) | **0.85 ms** | **3.10 ms** |
| **Max UI Conflation Frequency**| Unlimited (DOM crash) | **20.0 FPS (50.0ms lock)** | **20.0 FPS (50.0ms lock)** |
| **Display Render Frame Rate** | 22 - 45 FPS (Jank) | **120 / 144 / 240 FPS (Locked)**| **60 / 120 FPS (Locked)** |

---

## 7. Strict 0.00% Zero-Fee Presentation & 0 Gas Sponsorship Badges

The Growww / NBSE platform operates on a cryptographically enforced zero-cost transaction model. The Market Data UI and order book interfaces must communicate this zero-fee structure cleanly to the user.

### 7.1 Visual Placement in Market Data Ladder & Order Entry DOM
1. **Spread & Fee Header Indicator:**
   - Displayed immediately adjacent to the live Top-of-Book Spread display.
   - Text presentation: `FEES: 0.00% MAKER | 0.00% TAKER`.
   - Styling: Fixed-width monospace typeface, styled in neon green (`#00C087`) or bright cyan (`#00F0FF`).
2. **Click-to-Trade DOM Ladder (Depth of Market):**
   - In dynamic PnL projection columns on the ladder, projected trading fees are locked at `0.0000 USDT` / `0.00 INR`.
   - Fee impact tooltips must explicitly state: `"NBSE Sovereign Exchange: 0.00% Spot & Futures Trading Fees - Zero Platform Commission"`.
3. **Execution Trade Blotter:**
   - The fee column for executed public fills and private user trades must display `0.00` with an active tooltip displaying the blockchain sponsorship transaction identifier.

### 7.2 Zero-Gas Sponsorship Badge Specification
For all decentralized and hybrid trade settlements executed via Hyperledger Besu and Account Abstraction (EIP-4337):
- **Visual Badge Component:** An interactive badge placed prominently in the market data header.
  - Label: `"0 GAS SPONSORED"`
  - Icon: Fuel pump icon with a slash or lightning badge.
  - Colorway: Emerald gradient background (`#00C087` to `#00A872`) with pure white text (`#FFFFFF`).
- **Telemetry Popover:** Clicking the badge opens a telemetry inspection modal providing live verification:
  - Account Abstraction Paymaster Address: `0x71C...0000` (Besu Paymaster Contract).
  - Cumulative Gas Sponsored: Dynamic counter of sponsored transactions.
  - User Gas Incurred: Permanently `0.00 Gwei` ($0.00 USD).
- **Client-Layer Immutability Invariant:** Fee display values are hardcoded as immutable mathematical constants in the UI layer. They are decoupled from dynamic server overrides to prevent any unauthorized fee introductions.

```
+----------------------------------------------------------------------------------------------------+
| ZERO-FEE & ZERO-GAS UI COMPONENT SPECIFICATION                                                     |
|                                                                                                    |
|  +----------------------------------------------------------------------------------------------+  |
|  | BTC/USDT PERP  65,420.50  +4.82%  | SPREAD: 0.50 (0.0008%) |  [ 0.00% ZERO-FEE GUARANTEED ]  |  |
|  | 24h Vol: 1.48B USDT               | HIGH: 66,200.00        |  [ 0 GAS SPONSORED BY PAYMASTER]|  |
|  +----------------------------------------------------------------------------------------------+  |
|                                                                                                    |
|  Depth of Market (DOM) Click-to-Trade Ladder:                                                      |
|  [ BID SIZE ]  [ PRICE ]   [ ASK SIZE ]  |  [ ORDER ENTRY PANEL: BUY 1.00 BTC @ 65,420.50 ]        |
|  14.285        65,420.50                 |  - Maker Fee:      0.00 USDT (0.00%)                    |
|  08.120        65,420.00                 |  - Taker Fee:      0.00 USDT (0.00%)                    |
|                65,421.00   02.450        |  - Blockchain Gas: 0.00 Gwei (100% Paymaster Sponsored) |
|                65,421.50   11.890        |  ----------------------------------------------------   |
|                                          |  [ PLACE ZERO-FEE ORDER (EST. TOTAL: 65,420.50 USDT) ]  |
+----------------------------------------------------------------------------------------------------+
```

---

## 8. Implementation Verification & Soak-Testing Criteria

Before deploying the conflation and shared-memory subsystem to production, the engineering group must verify compliance against the following automated test suites:

### 8.1 72-Hour Soak Test Under Synthetic Stress Load
- **Condition:** 100,000 ticks/sec sustained traffic injected into WebSocket gateways for 72 consecutive hours.
- **Pass Criteria:**
  - V8 Heap Growth: Total memory variance $\le 2.0\text{ MB}$ over 72 hours (no memory leaks).
  - SharedArrayBuffer Integrity: Zero torn reads detected across 5,184,000 conflated UI frames.
  - OS Memory Mapping: Memory resident set size (RSS) of Rust desktop core stays flat at $\le 32\text{ MB}$.

### 8.2 Sequence Gap Injection & Recovery Test
- **Condition:** Randomly drop 5% of incoming delta packets at the network emulation proxy.
- **Pass Criteria:**
  - Client state machine detects gap within $\le 1$ delta packet.
  - Ingestion buffer transitions to `BUFFERING_DELTAS` immediately.
  - Resync snapshot acquired and applied within $\le 150\text{ ms}$.
  - Top-50 order book CRC32 checksum returns to 100% agreement with matching engine.

### 8.3 Multi-Window Desktop Stress Test
- **Condition:** Open 8 simultaneous detached Flutter thick-client windows (3 charts, 2 DOM ladders, 2 blotters, 1 main dashboard) on macOS and Windows workstations.
- **Pass Criteria:**
  - All 8 windows render at target display refresh rate (120Hz on macOS ProMotion, 144Hz on Windows).
  - CPU utilization remains below 8% across all combined rendering processes.
  - Cross-window state disparity (time delta between Window A and Window B reading the same tick): $\le 50\text{ microseconds}$.
