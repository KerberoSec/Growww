# Institutional FIX 5.0 SP2 and ITCH/OUCH Binary Direct Feeds Specification

**Specification ID:** SPEC-ARCH-043-THICK-FIX-ITCH-OUCH  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Desktop Thick Client, Low-Latency Direct Market Access (DMA) & Binary Feeds  
**Target Clients:** Desktop Thick Client (macOS Cocoa/Metal, Windows Win32/DirectX 12, Linux GTK/Vulkan via Flutter Desktop & Rust FFI)  
**Target Protocols:** FIX 5.0 SP2 (with FIXT 1.1 Session Layer), Binary ITCH 5.0 (Market Data), Binary OUCH 5.0 (Order Entry), MoldUDP64  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Architectural Overview

This document specifies the institutional-grade Direct Market Access (DMA) connectivity suite implemented within the Growww/NBSE Desktop Thick Client. Designed for quantitative funds, proprietary trading desks, algorithmic market makers, and institutional asset managers, this architecture bypasses standard retail WebSocket and HTTP gateways in favor of direct, deterministic, ultra-low-latency financial protocols.

```
+-----------------------------------------------------------------------------------------------------------------------+
|                                    DESKTOP THICK CLIENT DMA ARCHITECTURAL PIPELINE                                    |
+-----------------------------------------------------------------------------------------------------------------------+
|                                                                                                                       |
|  +-----------------------------------------------------------------------------------------------------------------+  |
|  | GUI LAYER (Flutter 3.22+ Desktop / Impeller Engine: Metal on macOS, DirectX 12 on Win, Vulkan on Linux)          |  |
|  |  - High-Refresh Order Book (60/120/144/240/360 Hz)     - Zero-Fee Trade Blotters & Order Tickets (0.00% Fee)    |  |
|  |  - Microsecond Execution Latency Displays             - Paymaster Gas Sponsorship Badging (0 Gas Cost)          |  |
|  +-----------------------------------------------------------------------------------------------------------------+  |
|                                        ^                                            |                                 |
|                     Zero-Copy FFI View | (Pointer / Atomic Ring)    FFI Method Call | (C ABI / Dart FFI)              |
|                                        |                                            v                                 |
|  +-----------------------------------------------------------------------------------------------------------------+  |
|  | EMBEDDED RUST ENGINE (librust_trading_core: #[repr(C)] Zero-Allocation Core)                                    |  |
|  |                                                                                                                 |  |
|  |  +------------------------------------+  +---------------------------------+  +-------------------------------+  |
|  |  |  IN-PROCESS ORDER BOOK ENGINE      |  |  BINARY ITCH 5.0 PARSER         |  |  BINARY OUCH 5.0 CLIENT       |  |
|  |  |  - Order-by-Order (L3) Lookup Table|  |  - MoldUDP64 Multicast Receiver |  |  - Zero-Alloc Packet Framer   |  |
|  |  |  - Aggregated Level-2 Ladder Cache |  |  - Sequence Replay Gap Engine   |  |  - Sequence State Machine     |  |
|  |  |  - Atomic Conflation Snapshots     |  |  - Unicast TCP Recovery Client  |  |  - Execution Notification     |  |
|  |  +------------------------------------+  +---------------------------------+  +-------------------------------+  |
|  |                                        ^                                            |                              |
|  |                                        | Packet Dispatch                            | Serialized Outbound          |
|  |  +---------------------------------------------------------------------------------------------------------+      |
|  |  |  INSTITUTIONAL FIX 5.0 SP2 ENGINE (FIXT 1.1 Session Layer)                                              |      |
|  |  |  - In-Memory Circular Journal Buffer      - Microsecond/Nanosecond Clocks (CLOCK_REALTIME / PTP)        |      |
|  |  |  - Automatic Gap-Recovery (ResendRequest)  - Bidirectional 30s Heartbeat & Dual-Homed Session Failover    |      |
|  |  +---------------------------------------------------------------------------------------------------------+      |
|  +-----------------------------------------------------------------------------------------------------------------+  |
|                                        |                                            |                                 |
|             Kernel-Bypass / Standard OS| TCP/IP / TLS 1.3 / UDP Multicast Sockets   | Direct Cross-Connect / Colocation|
|                                        v                                            v                                 |
|  +-----------------------------------------------------------------------------------------------------------------+  |
|  | EXCHANGE INFRASTRUCTURE (Tier-4 Equinix MB1/MB2 Mumbai & GIFT City IFSC Gateways)                              |  |
|  |  - FIX 5.0 SP2 Gateway (Port 9800/TCP TLS)           - OUCH 5.0 Direct Order Ingress (Port 9900/TCP TLS)        |  |
|  |  - MoldUDP64 ITCH Multicast (239.255.0.1:10001/UDP)  - ITCH Unicast TCP Replay Server (Port 10002/TCP)          |  |
|  +-----------------------------------------------------------------------------------------------------------------+  |
+-----------------------------------------------------------------------------------------------------------------------+
```

### Core Architecture Capabilities
1. **Direct Socket Ingress (Bypassing Intermediate Proxies):** The thick client initiates native TCP/TLS 1.3 sockets directly from the host operating system to the exchange edge gateways, eliminating reverse-proxy and WebSocket conflation overhead.
2. **Deterministic Rust FFI Processing:** Network parsing, order book reconstruction, sequence validation, and packet framing are executed in an embedded native Rust library (`librust_trading_core`), exposing thread-safe C ABIs to the Flutter UI thread.
3. **Sub-Microsecond Clock Precision:** All timestamps use nanoseconds elapsed since the Unix epoch (January 1, 1970 00:00:00 UTC), synchronized via IEEE 1588v2 PTP (Precision Time Protocol) or hardware NIC clocks.
4. **Resilient Session Topology:** Automatic sequence number tracking, resend-request negotiation, 30-second heartbeat liveness probes, and zero-data-loss dual-homed failover across active-standby gateway clusters.
5. **Universal Zero-Fee Guarantee:** Unwavering presentation of 0.00% maker fees, 0.00% taker fees, 0 platform brokerage, and 0 gas cost execution badging across all FIX execution reports, OUCH fills, and desktop UI widgets.

---

## 2. Desktop Thick Client FIX 5.0 SP2 Session Gateway Direct Connectivity

### 2.1 Protocol Layering and Standards
The thick client implements the Financial Information eXchange protocol using the dual-layer architecture:
- **Session Layer:** FIXT 1.1 (Financial Information eXchange Transport Layer 1.1).
- **Application Layer:** FIX 5.0 SP2 (Service Pack 2, Extension Packs EP268 compliant).

### 2.2 Transport & Cryptographic Handshake
- **Transport:** Native non-blocking TCP socket (`SO_KEEPALIVE`, `TCP_NODELAY` enabled).
- **Transport Layer Security:** TLS 1.3 mandatory (RFC 8446). Permitted cipher suites:
  - `TLS_AES_256_GCM_SHA384`
  - `TLS_CHACHA20_POLY1305_SHA256`
- **Mutual Authentication (mTLS):** The client provides a client certificate (X.509v3) issued by the Growww Institutional Certificate Authority.
- **Application Authentication (Logon 35=A):**
  - Tag 553 (`Username`): Institutional Client Identifier.
  - Tag 554 (`Password`): Ephemeral API Key.
  - Tag 96 (`RawData`): HMAC-SHA256 signature generated over `SendingTime (Tag 52) + TargetCompID (Tag 56) + MsgSeqNum (Tag 34)` using the institutional API Secret.
  - Tag 98 (`EncryptMethod`): `0` (None, encryption handled at TLS 1.3 transport).
  - Tag 108 (`HeartBtInt`): `30` (30-second heartbeat interval).
  - Tag 1137 (`DefaultApplVerID`): `9` (FIX 5.0 SP2).

### 2.3 FIX 5.0 SP2 Tag Dictionary & Supported Message Schema

The client engine supports the following institutional FIX application messages:

| Message Name | MsgType (Tag 35) | Direction | Application Function |
| :--- | :--- | :--- | :--- |
| **Logon** | `A` | Client <-> Server | Authenticate session and negotiate starting sequence numbers. |
| **Heartbeat** | `0` | Client <-> Server | Verify link integrity during idle intervals. |
| **TestRequest** | `1` | Client <-> Server | Force peer heartbeat response with matched `TestReqID` (Tag 112). |
| **ResendRequest** | `2` | Client <-> Server | Recover missing sequence gap (`BeginSeqNo` Tag 7 to `EndSeqNo` Tag 16). |
| **Reject** | `3` | Client <- Server | Session-level protocol rejection (malformed tag or checksum). |
| **SequenceReset** | `4` | Client <-> Server | Gap-fill or hard reset of sequence numbers (`GapFillFlag` Tag 123). |
| **Logout** | `5` | Client <-> Server | Orderly session termination. |
| **NewOrderSingle** | `D` | Client -> Server | Submit institutional Limit, Market, Stop-Limit, or Post-Only order. |
| **OrderCancelRequest**| `F` | Client -> Server | Request cancellation of an active resting order. |
| **OrderCancelReplace**| `G` | Client -> Server | Modify price or quantity of a resting order. |
| **ExecutionReport** | `8` | Client <- Server | Order acknowledgment, fill, cancel confirmation, or rejection. |
| **OrderCancelReject** | `9` | Client <- Server | Rejection of cancel or cancel/replace request. |

### 2.4 Detailed FIX Field Definitions

```
+-----------------------------------------------------------------------------------------------------------------+
| FIX 5.0 SP2 APPLICATION MESSAGE DEFINITIONS (NEW ORDER SINGLE & EXECUTION REPORT)                              |
+-----------------------------------------------------------------------------------------------------------------+

NEW ORDER SINGLE (MsgType 35=D):
  8=FIXT.1.1 | 9=188 | 35=D | 49=DESK_PROP_01 | 56=GROWWW_DMA | 34=1042 | 52=20260919-12:15:30.123456789 |
  11=ORD_20260919_000101 | 55=BTC/USDT | 48=IN0020260011 | 22=4 | 54=1 | 60=20260919-12:15:30.123400000 |
  38=1.50000000 | 40=2 | 44=68500.00 | 59=1 | 18=M | 10001=Y | 10002=0.0000 | 10=045 |

EXECUTION REPORT (MsgType 35=8):
  8=FIXT.1.1 | 9=242 | 35=8 | 49=GROWWW_DMA | 56=DESK_PROP_01 | 34=2019 | 52=20260919-12:15:30.123891410 |
  37=EX_99812401 | 11=ORD_20260919_000101 | 17=TRD_44102 | 150=F | 39=2 | 55=BTC/USDT | 54=1 |
  38=1.50000000 | 44=68500.00 | 32=1.50000000 | 31=68500.00 | 151=0.00000000 | 14=1.50000000 | 6=68500.00 |
  60=20260919-12:15:30.123850000 | 12=0.00 | 13=3 | 10001=Y | 10002=0.0000 | 10003=SPONSORED_FREE | 10=182 |
```

#### Field Details:
- **Tag 11 (`ClOrdID`):** Client-generated unique order identifier (string, up to 32 ASCII characters).
- **Tag 37 (`OrderID`):** Exchange matching-engine assigned unique order identifier.
- **Tag 17 (`ExecID`):** Unique execution event identifier for auditability.
- **Tag 150 (`ExecType`):** `0` (New), `1` (Partial Fill), `2` (Full Fill), `4` (Canceled), `5` (Replaced), `8` (Rejected).
- **Tag 39 (`OrdStatus`):** Current status of the order matching Tag 150 values.
- **Tag 12 (`Commission`):** Explicitly set to `0.00` in every execution.
- **Tag 13 (`CommType`):** `3` (Absolute amount).
- **Tag 18 (`ExecInst`):** `M` (Participate do not initiate / Post-Only).
- **Tag 60 (`TransactTime`):** Nanosecond UTC timestamp (`YYYYMMDD-HH:MM:SS.sssssssss`).
- **Tag 10001 (`ZeroFeeFlag`):** Custom user-defined tag verifying Growww zero-fee execution (`Y`).
- **Tag 10002 (`TradingFeeRate`):** Rate explicitly set to `0.0000`.
- **Tag 10003 (`GasSponsorshipStatus`):** Confirmation of account abstraction sponsorship (`SPONSORED_FREE`).

---

## 3. Binary ITCH and OUCH Protocol Parser in Rust FFI

For latency-critical workflows, the desktop thick client integrates `librust_trading_core`, an embedded Rust dynamic library linked via Foreign Function Interface (FFI). This subsystem processes binary OUCH for order entry and binary ITCH for tick-by-tick order book data.

```
+-----------------------------------------------------------------------------------------------------------------+
| MEMORY LAYOUT & ZERO-COPY FFI PIPELINE IN LIBRUST_TRADING_CORE                                                 |
+-----------------------------------------------------------------------------------------------------------------+
|                                                                                                                 |
|  [ Inbound MoldUDP64 / TCP Buffer ]                                                                             |
|                     |                                                                                           |
|                     v                                                                                           |
|  [ Rust Kernel Reader ] ---> Stack Allocation / Slice Casting (zero-allocation)                                 |
|                                     |                                                                           |
|                                     v                                                                           |
|                        #[repr(C, packed)] Raw Structs                                                           |
|                                     |                                                                           |
|                                     v                                                                           |
|                  [ In-Process L3 Order Book Mutator ]                                                           |
|                                     |                                                                           |
|                                     v                                                                           |
|            [ Atomic Conflation Buffer (Double-Buffered Top 50 Levels) ]                                         |
|                     |                                       |                                                   |
|                     | Pointer Swap                          | Dart FFI Direct Pointer Read                      |
|                     v                                       v                                                   |
|             [ Write Buffer ]                         [ Flutter UI Render Isolate ]                              |
|                                                                                                                 |
+-----------------------------------------------------------------------------------------------------------------+
```

### 3.1 Binary OUCH 5.0 (Order Entry Protocol) Specification

OUCH is a point-to-point, fixed-width binary protocol operating over a persistent TCP/TLS stream. All integers are encoded in network byte order (Big-Endian).

#### Outbound Client Messages

##### 1. Enter Order Packet (Type 'O')
Length: 48 bytes.
- `Type` (1 byte, char): `'O'` (0x4F).
- `ClientOrderID` (8 bytes, uint64_t): Unique client-assigned numeric identifier.
- `Side` (1 byte, char): `'B'` (Buy) or `'S'` (Sell).
- `Quantity` (8 bytes, uint64_t): Scaled quantity ($10^8$ fixed-point precision, 1.0 = 100,000,000).
- `InstrumentID` (8 bytes, uint64_t): Internal numeric identifier for the traded pair.
- `Price` (8 bytes, uint64_t): Scaled price ($10^4$ fixed-point precision, 1.0000 = 10,000).
- `TimeInForce` (1 byte, uint8_t): `0` = IOC (Immediate-Or-Cancel), `1` = Day/GTC, `2` = Post-Only.
- `FirmID` (4 bytes, alphanumeric): Desk identifier padding with spaces.
- `ZeroFeeToken` (1 byte, uint8_t): Constant `0x01` asserting zero-fee protocol requirement.
- `Reserved` (9 bytes, bytes): Zero-padded padding to align to 64-bit boundaries.

##### 2. Cancel Order Packet (Type 'X')
Length: 17 bytes.
- `Type` (1 byte, char): `'X'` (0x58).
- `ClientOrderID` (8 bytes, uint64_t): ID of the order to cancel.
- `Quantity` (8 bytes, uint64_t): Shares to reduce (`0` = cancel entire remaining quantity).

##### 3. Replace Order Packet (Type 'U')
Length: 33 bytes.
- `Type` (1 byte, char): `'U'` (0x55).
- `ExistingClientOrderID` (8 bytes, uint64_t): Current client order ID.
- `NewClientOrderID` (8 bytes, uint64_t): Replacement client order ID.
- `NewQuantity` (8 bytes, uint64_t): New total open quantity.
- `NewPrice` (8 bytes, uint64_t): New price ($10^4$ fixed point).

#### Inbound Gateway Messages

##### 1. Order Accepted Packet (Type 'A')
Length: 66 bytes.
- `Type` (1 byte, char): `'A'` (0x41).
- `TimestampNanoseconds` (8 bytes, uint64_t): Unix epoch nanoseconds from matching engine.
- `ClientOrderID` (8 bytes, uint64_t): Matched client order identifier.
- `ExchangeOrderID` (8 bytes, uint64_t): Unique order reference assigned by matching engine.
- `Side` (1 byte, char): `'B'` or `'S'`.
- `Quantity` (8 bytes, uint64_t): Accepted quantity ($10^8$).
- `InstrumentID` (8 bytes, uint64_t): Numeric instrument identifier.
- `Price` (8 bytes, uint64_t): Accepted limit price ($10^4$).
- `OrderState` (1 byte, uint8_t): `1` = Active/Resting.
- `FeeBps` (2 bytes, uint16_t): Explicitly `0x0000` (0 basis points).
- `SponsorshipFlag` (1 byte, uint8_t): `0x01` (Gas Sponsored).
- `Padding` (12 bytes, bytes): 64-bit word alignment.

##### 2. Order Executed Packet (Type 'E')
Length: 41 bytes.
- `Type` (1 byte, char): `'E'` (0x45).
- `TimestampNanoseconds` (8 bytes, uint64_t): Execution timestamp.
- `ClientOrderID` (8 bytes, uint64_t): Client order ID.
- `ExecutedQuantity` (8 bytes, uint64_t): Filled volume.
- `ExecutedPrice` (8 bytes, uint64_t): Execution price.
- `MatchID` (8 bytes, uint64_t): Unique trade match identifier.

##### 3. Order Canceled Packet (Type 'C')
Length: 26 bytes.
- `Type` (1 byte, char): `'C'` (0x43).
- `TimestampNanoseconds` (8 bytes, uint64_t): Event timestamp.
- `ClientOrderID` (8 bytes, uint64_t): Canceled order ID.
- `ReasonCode` (1 byte, uint8_t): `1` = User Cancel, `2` = IOC Expiry, `3` = Post-Only Cross.

##### 4. Order Rejected Packet (Type 'J')
Length: 18 bytes.
- `Type` (1 byte, char): `'J'` (0x4A).
- `TimestampNanoseconds` (8 bytes, uint64_t): Timestamp.
- `ClientOrderID` (8 bytes, uint64_t): Rejected order ID.
- `RejectReason` (1 byte, uint8_t): Rejection code (e.g., `1` = Insufficient Balance, `2` = Market Closed).

---

### 3.2 Binary ITCH 5.0 (Market Data Multicast) Specification

ITCH 5.0 distributes every atomic order addition, execution, cancellation, and deletion. Messages are encapsulated inside MoldUDP64 packets over UDP multicast or unicast TCP.

#### MoldUDP64 Packet Framing
Every MoldUDP64 datagram consists of a 20-byte header followed by one or more length-prefixed ITCH messages:
- `Session` (10 bytes, ASCII): Unique session name identifying the multicast stream.
- `SequenceNumber` (8 bytes, uint64_t): The sequence number of the first ITCH message in the packet.
- `MessageCount` (2 bytes, uint16_t): Number of ITCH messages contained in this packet.

Each enclosed ITCH message is prefixed by a 2-byte unsigned integer indicating the `MessageLength`.

#### ITCH Message Types and Layouts

##### 1. Timestamp Message (Type 'T')
Length: 9 bytes.
- `Type` (1 byte, char): `'T'` (0x54).
- `TimestampNanoseconds` (8 bytes, uint64_t): Nanoseconds since Unix epoch.

##### 2. Add Order Message (Type 'A')
Length: 36 bytes.
- `Type` (1 byte, char): `'A'` (0x41).
- `TimestampNanoseconds` (8 bytes, uint64_t): Event timestamp.
- `OrderReferenceNumber` (8 bytes, uint64_t): Unique reference ID for this resting order.
- `Side` (1 byte, char): `'B'` (Buy) or `'S'` (Sell).
- `Shares` (8 bytes, uint64_t): Scaled quantity ($10^8$).
- `InstrumentID` (8 bytes, uint64_t): Numeric identifier for the symbol.
- `Price` (8 bytes, uint64_t): Limit price ($10^4$).

##### 3. Add Order with Participant ID (Type 'F')
Length: 40 bytes.
- Identical to Type 'A', plus:
- `Attribution` (4 bytes, ASCII): Firm / Maker MPID (Market Participant Identifier).

##### 4. Order Executed Message (Type 'E')
Length: 31 bytes.
- `Type` (1 byte, char): `'E'` (0x45).
- `TimestampNanoseconds` (8 bytes, uint64_t): Event timestamp.
- `OrderReferenceNumber` (8 bytes, uint64_t): The resting order that was filled.
- `ExecutedShares` (8 bytes, uint64_t): Number of shares matched.
- `MatchNumber` (8 bytes, uint64_t): Unique execution match ID.

##### 5. Order Executed with Price Message (Type 'C')
Length: 39 bytes.
- Emitted when an order executes at a price differing from its original limit price (e.g., crossing spread).
- `Type` (1 byte, char): `'C'` (0x43).
- Fields identical to 'E', plus:
- `ExecutionPrice` (8 bytes, uint64_t): Actual traded price ($10^4$).

##### 6. Order Cancel / Reduce Message (Type 'X')
Length: 23 bytes.
- `Type` (1 byte, char): `'X'` (0x58).
- `TimestampNanoseconds` (8 bytes, uint64_t): Timestamp.
- `OrderReferenceNumber` (8 bytes, uint64_t): Order reference ID.
- `CanceledShares` (8 bytes, uint64_t): Quantity deducted from resting order.

##### 7. Order Delete Message (Type 'D')
Length: 17 bytes.
- `Type` (1 byte, char): `'D'` (0x44).
- `TimestampNanoseconds` (8 bytes, uint64_t): Timestamp.
- `OrderReferenceNumber` (8 bytes, uint64_t): Order reference ID to remove completely from the book.

##### 8. Order Replace Message (Type 'U')
Length: 33 bytes.
- `Type` (1 byte, char): `'U'` (0x55).
- `TimestampNanoseconds` (8 bytes, uint64_t): Timestamp.
- `OriginalOrderReference` (8 bytes, uint64_t): Existing order to remove.
- `NewOrderReference` (8 bytes, uint64_t): New order reference replacing it.
- `Shares` (8 bytes, uint64_t): Updated remaining shares.
- `Price` (8 bytes, uint64_t): Updated price.

---

### 3.3 Rust FFI Interface Definitions (C-ABI Exposure)

The embedded library `librust_trading_core` exposes high-performance C symbols callable directly from Dart FFI:

```rust
// C-compatible struct representing an aggregated depth level
#[repr(C)]
pub struct FfiBookLevel {
    pub price: u64,       // Fixed-point 10^4
    pub quantity: u64,    // Fixed-point 10^8
    pub order_count: u32, // Number of individual orders resting at level
    pub _padding: u32,
}

// Fixed-capacity 50-level top-of-book snapshot
#[repr(C)]
pub struct FfiBookSnapshot {
    pub instrument_id: u64,
    pub timestamp_ns: u64,
    pub bid_count: u32,
    pub ask_count: u32,
    pub bids: [FfiBookLevel; 50],
    pub asks: [FfiBookLevel; 50],
}

// C-ABI functions exposed by librust_trading_core
extern "C" {
    // Session initialization
    pub fn dma_engine_init(config_json: *const libc::c_char) -> i32;
    pub fn dma_engine_shutdown() -> i32;

    // ITCH Multicast controls
    pub fn itch_subscribe_multicast(
        interface_ip: *const libc::c_char,
        multicast_group: *const libc::c_char,
        port: u16,
    ) -> i32;

    // Zero-copy snapshot acquisition for UI isolate
    pub fn itch_get_book_snapshot(
        instrument_id: u64,
        out_snapshot: *mut FfiBookSnapshot,
    ) -> i32;

    // OUCH Direct Order Submission
    pub fn ouch_send_order(
        client_order_id: u64,
        instrument_id: u64,
        side: u8,        // b'B' or b'S'
        quantity: u64,   // 10^8 fixed-point
        price: u64,      // 10^4 fixed-point
        tif: u8,         // 0=IOC, 1=Day, 2=Post-Only
    ) -> i32;

    pub fn ouch_cancel_order(client_order_id: u64, quantity: u64) -> i32;

    // Register callback for asynchronous execution events
    pub fn ouch_register_exec_callback(
        callback: extern "C" fn(event_ptr: *const libc::c_void, event_len: usize),
    ) -> i32;
}
```

---

## 4. Microsecond and Nanosecond Timestamping and Sequence Replay Gap-Recovery

### 4.1 Clock Architecture and Invariant
- **Reference System:** UTC (Coordinated Universal Time) with leap seconds handled by linear POSIX smearing.
- **Resolution:** Nanoseconds represented as 64-bit unsigned integers (`uint64_t`), denoting nanoseconds elapsed since `1970-01-01 00:00:00.000000000 UTC`.
- **Operating System Hardware Timestamping:**
  - Linux: `CLOCK_REALTIME` via `clock_gettime()`, with kernel socket timestamping via `SO_TIMESTAMPING`.
  - macOS: `clock_gettime_nsec_np(CLOCK_REALTIME)`.
  - Windows: `QueryPerformanceCounter` calibrated against `GetSystemTimePreciseAsFileTime`.
- **Clock Drift Threshold:** Monitored continuously via NTP/PTP daemon. If local client clock skew exceeds $\pm 500\mu\text{s}$ relative to exchange time, the thick client raises an immediate visual warning and suppresses aggressive colocation routing.

```
+-----------------------------------------------------------------------------------------------------------------+
| SEQUENCE RECOVERY STATE MACHINE (GAP DETECTION AND RECONCILIATION)                                              |
+-----------------------------------------------------------------------------------------------------------------+
|                                                                                                                 |
|                Incoming Packet (Seq: N)                                                                         |
|                           |                                                                                     |
|                           v                                                                                     |
|              +-------------------------+                                                                        |
|              | Is Seq == ExpectedSeq?  |                                                                        |
|              +-------------------------+                                                                        |
|               /                       \                                                                         |
|        YES   /                         \   NO (Seq > ExpectedSeq)                                               |
|             v                           v                                                                       |
|   +-------------------+      +-------------------------------------------+                                      |
|   | Process In-Order  |      | Sequence Gap Detected!                    |                                      |
|   | ExpectedSeq += 1  |      | Missing: [ExpectedSeq .. N - 1]           |                                      |
|   +-------------------+      +-------------------------------------------+                                      |
|                                     |                                                                           |
|                                     | 1. Buffer Incoming Packet N in Out-of-Order Cache                         |
|                                     | 2. Emit Recovery Request                                                  |
|                                     v                                                                           |
|                      +----------------------------------+                                                       |
|                      | FIX: Send ResendRequest (35=2)   |                                                       |
|                      | ITCH: Send MoldUDP64 Replay Req  |                                                       |
|                      +----------------------------------+                                                       |
|                                     |                                                                           |
|                                     v                                                                           |
|                      +----------------------------------+                                                       |
|                      | Ingest Missing Packets from Peer |                                                       |
|                      | Apply to Order Book / State      |                                                       |
|                      +----------------------------------+                                                       |
|                                     |                                                                           |
|                                     v                                                                           |
|                      +----------------------------------+                                                       |
|                      | Drain Out-of-Order Cache up to N |                                                       |
|                      | Resume Live Packet Ingestion     |                                                       |
|                      +----------------------------------+                                                       |
+-----------------------------------------------------------------------------------------------------------------+
```

### 4.2 FIX Sequence Recovery Specification
1. **Sequence Maintenance:** The client maintains two persistent sequence integers: `InboundSeqNum` and `OutboundSeqNum`, persisted in a low-latency append-only write-ahead log (WAL) on local NVMe storage.
2. **Gap Detection:** If an incoming message contains `MsgSeqNum > InboundSeqNum + 1`, the client:
   - Queues the arriving message into an in-memory priority queue indexed by `MsgSeqNum`.
   - Emits a `ResendRequest` (35=2) with:
     - `BeginSeqNo (Tag 7) = InboundSeqNum + 1`
     - `EndSeqNo (Tag 16) = 0` (indicating replay up to the most recent message).
3. **Replay Ingestion:**
   - The exchange replays messages with `PossDupFlag (Tag 43) = Y`.
   - Administrative messages (Logon, Logout, Heartbeat, TestRequest, ResendRequest) are skipped by the gateway via `SequenceReset` (35=4) with `GapFillFlag (Tag 123) = Y`.
   - Application messages (ExecutionReport 35=8) are processed, updating local order blotters.
4. **Resumption:** When the replayed stream reaches `MsgSeqNum == InboundSeqNum`, the priority queue is drained and live processing resumes instantaneously.

### 4.3 MoldUDP64 / ITCH Multicast Sequence Gap-Recovery
1. **Primary Feed Monitoring:** Each MoldUDP64 packet carries `SequenceNumber` (first message ID) and `MessageCount`. If `Packet.SequenceNumber > ExpectedItchSeq`, a multicast packet loss event has occurred.
2. **Local Holding Buffer:** The arriving packet is placed into a fixed-size ring buffer (capacity: 65,536 packets).
3. **Unicast Replay Request:** The client dispatches a MoldUDP64 Request Packet over a dedicated TCP replay connection to the Exchange Replay Server:
   - `Session` (10 bytes): Same session identifier.
   - `SequenceNumber` (8 bytes): Missing starting sequence number.
   - `MessageCount` (2 bytes): Number of missed messages (bounded to a maximum of 10,000 per request).
4. **Replay Processing:** The Replay Server transmits the exact binary ITCH packets over TCP. The Rust engine parses and applies them directly to the in-process order book before replaying the buffered multicast packets.
5. **Failover Safeguard:** If gap recovery takes longer than 1,500ms or missed messages exceed 100,000, the client marks the in-process book as STALE, issues a Snapshot Request to acquire a complete Level-2 state dump, and rebuilds the book atomically.

---

## 5. Heartbeat Management, TestRequest Cycles, and Session Failover

### 5.1 Heartbeat State Machine (30-Second Standard)

Both FIX 5.0 SP2 and OUCH sessions enforce a deterministic 30-second heartbeat cycle:

```
+-----------------------------------------------------------------------------------------------------------------+
| FIX/OUCH HEARTBEAT & TESTREQUEST TIMELINE (HEARTBTINT = 30 SECONDS)                                             |
+-----------------------------------------------------------------------------------------------------------------+
|                                                                                                                 |
|  T = 0s               T = 30s                                T = 33s                    T = 60s                 |
|  Last Message Rcvd    Heartbeat Expected                     TestRequest Emitted        Link Dead Threshold     |
|  ---------------------+--------------------------------------+--------------------------+-------------------->  |
|                       |                                      |                          |                       |
|                       | (No packet received in 30s)          | (Grace period +3s expired| (No Heartbeat received|
|                       |                                      |  Emit TestRequest 35=1   |  Force Disconnect     |
|                       |                                      |  with TestReqID="T1001") |  Trigger Failover)    |
|                                                                                                                 |
+-----------------------------------------------------------------------------------------------------------------+
```

#### Operational Rules:
1. **Outbound Heartbeat:** If the client has transmitted no application messages within `HeartBtInt` (30 seconds), it emits a `Heartbeat` (35=0).
2. **Inbound Heartbeat Monitoring:** If no data (packet or heartbeat) is received from the server within `HeartBtInt + 3s` (33 seconds), the client immediately sends a `TestRequest` (35=1) populated with a unique nanosecond-timestamped identifier in Tag 112 (`TestReqID`).
3. **Heartbeat Verification:** The exchange gateway MUST immediately respond with a `Heartbeat` (35=0) echoing the exact `TestReqID` (Tag 112).
4. **Dead Socket Declaration:** If no response echoing Tag 112 is received within $2 \times HeartBtInt$ (60 seconds from the last valid packet), the client kernel socket is forcefully severed via TCP RST/FIN, and the session initiates automated failover.

### 5.2 Dual-Homed Session Failover Topology

Institutional connectivity operates across dual-homed redundant paths:

| Endpoint Role | Primary Network Interface | Secondary Network Interface | Target Datacenter |
| :--- | :--- | :--- | :--- |
| **Primary FIX Gateway** | `10.101.10.50:9800` (Direct Cross-Connect) | `198.51.100.50:9800` (Public IP / WAN) | Equinix MB1 (Mumbai) |
| **Secondary FIX Gateway**| `10.102.10.50:9800` (Direct Cross-Connect) | `198.51.101.50:9800` (Public IP / WAN) | GIFT City IFSC (Gandhinagar) |
| **Primary OUCH Gateway** | `10.101.10.60:9900` (Direct Cross-Connect) | `198.51.100.60:9900` (Public IP / WAN) | Equinix MB1 (Mumbai) |
| **Secondary OUCH Gateway**| `10.102.10.60:9900` (Direct Cross-Connect) | `198.51.101.60:9900` (Public IP / WAN) | GIFT City IFSC (Gandhinagar) |

#### Failover Execution Workflow:
1. **Immediate Socket Severance:** The Rust engine closes the faulty primary socket and cancels all pending I/O timers.
2. **Fast Reconnect:** The client opens an encrypted TLS 1.3 socket to the Secondary Gateway within $< 200\text{ms}$.
3. **Logon Negotiation:**
   - Sends `Logon` (35=A) with `ResetSeqNumFlag (Tag 141) = N`.
   - Advertises the current `OutboundSeqNum`.
4. **State Reconciliation:**
   - The Secondary Gateway reads the peer sequence numbers and returns its `Logon` acknowledgment.
   - If sequence numbers match, normal trading resumes immediately.
   - If the Secondary Gateway detects an inbound gap, it issues a `ResendRequest`; the client engine drains its local WAL journal and replays open state.
   - The client verifies open orders by issuing an `OrderMassStatusRequest` (35=AF) to confirm the live state of all resting orders.

---

## 6. In-Process Order Book Reconstruction from Raw ITCH Packets

The client application reconstructs a full Level-3 (order-by-order) and Level-2 (aggregated price-level) order book in real time directly inside the host machine's memory space, bypassing any server-side aggregation delays.

```
+-----------------------------------------------------------------------------------------------------------------+
| IN-PROCESS ORDER BOOK DATA STRUCTURES (LIBRUST_TRADING_CORE)                                                   |
+-----------------------------------------------------------------------------------------------------------------+
|                                                                                                                 |
|   ITCH TICK STREAM (Add, Exec, Cancel, Delete, Replace)                                                         |
|         |                                                                                                       |
|         v                                                                                                       |
|   +---------------------------------------------------------------------+                                       |
|   | L3 ORDER LOOKUP MAP: Flat HashMap<u64, OrderNode>                   |                                       |
|   | Key: OrderReferenceNumber (u64)                                     |                                       |
|   | Value: { price: u64, shares: u64, side: Side, instrument_id: u64 }  |                                       |
|   +---------------------------------------------------------------------+                                       |
|         |                                                                                                       |
|         | Updates aggregated levels                                                                             |
|         v                                                                                                       |
|   +---------------------------------------------------------------------+                                       |
|   | L2 PRICE LADDER: Contiguous Dense Arrays (Bids: Desc, Asks: Asc)    |                                       |
|   | Bids: [ {Price: 68500.00, Qty: 42.10, Count: 8}, ... Top 50 ]       |                                       |
|   | Asks: [ {Price: 68500.50, Qty: 18.45, Count: 3}, ... Top 50 ]       |                                       |
|   +---------------------------------------------------------------------+                                       |
|         |                                                                                                       |
|         | Atomic Double-Buffer Pointer Swap (Locked at 60/120/144/240/360 Hz)                                   |
|         v                                                                                                       |
|   +---------------------------------------------------------------------+                                       |
|   | FLUTTER DESKTOP RENDER PIPELINE (Zero-Copy Impeller View)           |                                       |
|   +---------------------------------------------------------------------+                                       |
+-----------------------------------------------------------------------------------------------------------------+
```

### 6.1 State Transition Logic for Inbound ITCH Packets

The order book engine maintains absolute consistency by applying the following atomic mutations:

```
+-----------------------------------------------------------------------------------------------------------------+
| ITCH PACKET MUTATION STATE TRANSITIONS                                                                          |
+-----------------------------------------------------------------------------------------------------------------+
|                                                                                                                 |
| 1. Add Order ('A' / 'F'):                                                                                       |
|    - Insert OrderNode(OrderReference, Price, Shares, Side) into L3 Map.                                         |
|    - Locate Price in L2 Ladder (or insert new level in sorted position).                                        |
|    - Level.Quantity += Shares; Level.OrderCount += 1.                                                           |
|                                                                                                                 |
| 2. Order Executed ('E'):                                                                                        |
|    - Lookup OrderNode in L3 Map using OrderReference.                                                           |
|    - OrderNode.Shares -= ExecutedShares.                                                                        |
|    - Level.Quantity -= ExecutedShares.                                                                          |
|    - If OrderNode.Shares == 0: Remove OrderNode from L3 Map; Level.OrderCount -= 1.                             |
|    - If Level.Quantity == 0: Prune Level from L2 Ladder.                                                        |
|                                                                                                                 |
| 3. Order Executed with Price ('C'):                                                                             |
|    - Same as Executed ('E') for quantity deduction; updates last-traded price to ExecutionPrice.                |
|                                                                                                                 |
| 4. Order Cancel / Reduce ('X'):                                                                                 |
|    - Lookup OrderNode in L3 Map.                                                                                |
|    - OrderNode.Shares -= CanceledShares.                                                                        |
|    - Level.Quantity -= CanceledShares.                                                                          |
|    - If OrderNode.Shares == 0: Remove OrderNode from L3 Map; Level.OrderCount -= 1.                             |
|    - If Level.Quantity == 0: Prune Level from L2 Ladder.                                                        |
|                                                                                                                 |
| 5. Order Delete ('D'):                                                                                          |
|    - Lookup OrderNode in L3 Map.                                                                                |
|    - Level.Quantity -= OrderNode.Shares; Level.OrderCount -= 1.                                                 |
|    - Remove OrderNode from L3 Map.                                                                              |
|    - If Level.Quantity == 0: Prune Level from L2 Ladder.                                                        |
|                                                                                                                 |
| 6. Order Replace ('U'):                                                                                         |
|    - Lookup old OrderNode by OriginalOrderReference.                                                            |
|    - Deduct old shares from old price level; prune old level if zero.                                           |
|    - Remove old OrderNode from L3 Map.                                                                          |
|    - Insert new OrderNode(NewOrderReference, NewPrice, NewShares, Side).                                        |
|    - Add NewShares to new price level.                                                                          |
|                                                                                                                 |
+-----------------------------------------------------------------------------------------------------------------+
```

### 6.2 UI Rendering Conflation and Zero-Copy Throttling
- **Problem:** ITCH feeds during market volatility can exceed 500,000 messages per second. Directly repainting the Flutter UI on every tick would saturate the main thread and trigger UI frame drops.
- **Solution:** Atomic double-buffering. The Rust mutator thread continuously updates the active book in memory. At discrete intervals tied to the monitor's physical refresh rate (e.g., every 8.33ms for 120Hz ProMotion or 4.16ms for 240Hz monitors), the background engine performs an atomic pointer swap.
- **Direct Dart Pointer Access:** The Flutter desktop isolate accesses the Top-50 snapshot buffer through direct memory pointers without JSON serialization or message copying, rendering pristine 120+ FPS depth ladders with 0% CPU stutter.

---

## 7. Strict 0.00% Zero-Fee Presentation & 0 Gas Sponsorship Badge

Growww operates under an absolute zero-fee architectural mandate. The thick client enforces this invariant across all communication layers, protocol payloads, and visual displays.

```
+-----------------------------------------------------------------------------------------------------------------+
| DESKTOP BLOTTER ZERO-FEE AND ACCOUNT ABSTRACTION SPONSORSHIP BADGE UI                                          |
+-----------------------------------------------------------------------------------------------------------------+
|                                                                                                                 |
|  +-----------------------------------------------------------------------------------------------------------+  |
|  | ORDER TICKET: BUY BTC/USDT                                                       [ DMA DIRECT: FIX 5.0 ]  |  |
|  +-----------------------------------------------------------------------------------------------------------+  |
|  | Price: 68,500.00 USDT                 Quantity: 1.50000000 BTC                   Order Type: LIMIT        |  |
|  | Total Order Value:                    102,750.00 USDT                                                     |  |
|  | --------------------------------------------------------------------------------------------------------- |  |
|  | Brokerage Fee:                        0.00% (USDT 0.00)                                                  |  |
|  | Exchange Transaction Dues:            0.00% (USDT 0.00)                                                  |  |
|  | Clearing & Settlement Charges:        0.00% (USDT 0.00)                                                  |  |
|  | Statutory Brokerage Platform Charge:  0.00% (USDT 0.00)                                                  |  |
|  | --------------------------------------------------------------------------------------------------------- |  |
|  | NET COMMISSION:                       0.00% (USDT 0.00 EXACT ZERO-FEE)                                    |  |
|  |                                                                                                           |  |
|  | [ BADGE: SPONSORED 0 GAS ]  [ BADGE: ERC-4337 PAYMASTER ACTIVE ]  [ BADGE: ZERO TDS / ZERO SURCHARGE ]   |  |
|  +-----------------------------------------------------------------------------------------------------------+  |
|                                                                                                                 |
|  EXECUTION BLOTTER (LIVE FIX / OUCH DROPS):                                                                     |
|  +--------------+-------------+------+-------+-----------+---------+-----------+-----------------------------+  |
|  | Time (UTC)   | OrderID     | Side | Sym   | Filled    | Price   | Fee (Bps) | Sponsorship State           |  |
|  +--------------+-------------+------+-------+-----------+---------+-----------+-----------------------------+  |
|  | 12:15:30.123 | EX_99812401 | BUY  | BTC   | 1.5000000 | 68500.0 | 0.00 bps  | SPONSORED (0 GAS / FREE)    |  |
|  | 12:15:31.004 | EX_99812402 | SELL | ETH   | 25.000000 | 3640.25 | 0.00 bps  | SPONSORED (0 GAS / FREE)    |  |
|  +--------------+-------------+------+-------+-----------+---------+-----------+-----------------------------+  |
+-----------------------------------------------------------------------------------------------------------------+
```

### 7.1 Protocol Invariant Rules
1. **FIX Protocol Tag 12 (`Commission`):** MUST be formatted as `12=0.00`. If an execution report arrives with a non-zero value in Tag 12, the client logs an invariant violation error and raises an immediate audit alert.
2. **FIX Protocol Tag 13 (`CommType`):** Set to `3` (Absolute / Cash).
3. **Custom Tag 10001 (`ZeroFeeFlag`):** Transmitted as `Y` in all `NewOrderSingle` (35=D) and validated in all `ExecutionReport` (35=8).
4. **Custom Tag 10002 (`TradingFeeRate`):** Set to `0.0000`.
5. **Custom Tag 10003 (`GasSponsorshipStatus`):** Populated with `SPONSORED_FREE`. Confirms that underlying blockchain state transitions (Hyperledger Besu settlement, ERC-4337 UserOperations) are 100% subsidized by the Growww institutional paymaster vault.
6. **OUCH Protocol Invariant:** Field `FeeBps` in Order Accepted ('A') and Order Executed ('E') must strictly equal `0x0000`.

### 7.2 UI/UX Presentation Standards
- **Blotters & Tables:** The Fee column displays `0.00%` formatted in Emerald Green (`#00E676`) or Slate Gray (`#808A9D`), with tooltip confirming: `"Growww Universal Zero-Fee Direct Execution"`.
- **Order Tickets:** The fee breakdown section displays an unalterable summary: `Brokerage: 0.00 USDT | Exchange Fee: 0.00 USDT | Net Cost: 0.00 USDT`.
- **Gas Sponsorship Badge:** Rendered as an emerald badge: `[0 GAS / BESU PAYMASTER SPONSORED]` positioned adjacent to the order confirmation button, guaranteeing that institutional users never incur Layer-1 or Layer-2 gas friction.

---

## 8. Latency Budgets, Edge Case Failure Modes, and Benchmarks

### 8.1 End-to-End Latency Budget (Wire to Pixel)

The thick client is engineered to achieve deterministic sub-millisecond execution and rendering:

| Processing Stage | Implementation Mechanism | Target Latency | P99.9 Latency |
| :--- | :--- | :--- | :--- |
| **Physical Ingress** | Solarflare NIC / Kernel Socket Ingress | $1.2\mu\text{s}$ | $2.8\mu\text{s}$ |
| **TLS 1.3 Decryption** | Ring / Rustls zero-copy buffer | $3.5\mu\text{s}$ | $6.2\mu\text{s}$ |
| **ITCH Packet Parse** | `#[repr(C, packed)]` slice casting | $0.15\mu\text{s}$| $0.35\mu\text{s}$ |
| **L3 Book Mutation** | Fast lookup map + contiguous array | $0.45\mu\text{s}$| $0.95\mu\text{s}$ |
| **Atomic Conflation** | Pointer double-buffering | $0.05\mu\text{s}$| $0.10\mu\text{s}$ |
| **Dart FFI Pointer Read** | Direct memory access via `dart:ffi` | $0.80\mu\text{s}$| $1.50\mu\text{s}$ |
| **Impeller Metal/DX12 GPU Paint** | Hardware rasterization on display frame | $2.5\text{ms}$ | $8.3\text{ms}$ (120Hz sync)|
| **Total In-Process Latency** | Wire to Memory State Ready | $< 6.2\mu\text{s}$| $< 12.0\mu\text{s}$ |

### 8.2 Failure Mode and Remediation Matrix

| Failure Scenario | Immediate Detection Indicator | Automated Remediation Procedure |
| :--- | :--- | :--- |
| **Silent TCP Hang** | Inbound silence $> 33\text{s}$ during active session | Client issues `TestRequest` (35=1). If no reply within 3s, terminates socket and engages secondary gateway. |
| **Multicast Packet Loss** | MoldUDP64 `SequenceNumber > ExpectedSeq` | Instantly buffers live packets in ring cache; requests unicast TCP slice replay; merges back to live stream. |
| **Corrupted Checksum / Tag** | FIX Tag 10 mismatch or corrupt OUCH byte | Discards packet; logs raw frame to diagnostic quarantine buffer; issues session reject; does not corrupt book. |
| **Gateway Mid-Flight Drop** | Socket EOF during in-flight `NewOrderSingle` | Reconnects to secondary gateway; queries `OrderMassStatusRequest` to reconcile resting state before retransmitting. |
| **Clock Desynchronization** | Delta between Tag 60 and local clock $> 500\mu\text{s}$ | Emits warning on GUI status bar; triggers PTP resync request; tags audit log with clock skew flag. |
| **Memory Buffer Saturation** | Inbound queue exceeds 1,000,000 packets | Activates backpressure drain; purges historical tick cache; keeps book top-50 intact; avoids process panic. |

---

## 9. Architectural Sign-Off & Verification Directives

1. **Protocol Compliance:** The gateway integrations specified herein conform strictly to FIXT 1.1, FIX 5.0 SP2, MoldUDP64, and Nasdaq ITCH/OUCH 5.0 standard protocol envelopes.
2. **Zero-Fee Enforcement:** No software layer, client setting, or broker parameter may alter the 0.00% fee invariant. All displays and blotters must present exact zero-fee accounting.
3. **Platform Independence:** The specifications for the embedded Rust engine (`librust_trading_core`) apply identically across macOS (Apple Silicon/Intel), Windows (x64/ARM64), and Linux (x86_64).
