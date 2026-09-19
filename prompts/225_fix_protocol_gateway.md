# 225 - FIX 5.0 SP2 / ITCH / OUCH Low-Latency Binary Trading Gateway (Rust)

## Purpose
To operate as a continuous 24/7 blockchain-powered stock exchange accommodating domestic and international institutional participants, hedge funds, and market makers, Growww requires industry-standard low-latency exchange connectivity protocols alongside retail REST/WebSocket APIs. 

The FIX / ITCH / OUCH Gateway provides high-throughput, deterministic institutional ingress and egress:
1. **FIX 5.0 SP2 (Financial Information eXchange):** Standard tag-value / FAST protocol session management for institutional order intake, cancellation, execution reports, and drop-copy feeds.
2. **OUCH Protocol:** Lightweight, point-to-point binary order entry and cancellation protocol designed for microsecond-sensitive algorithmic trading desks and market makers.
3. **ITCH 5.0 Protocol:** High-performance direct market data binary feed broadcasting full depth-of-book order additions, modifications, executions, cancellations, and administrative state changes via TCP and UDP multicast.

The gateway maps institutional integer/fractional equity quantities (down to 6 decimal places, $10^{-6}$ micro-shares) and INR prices ($10^{-4}$ paise precision) between binary/FIX wire encodings and Growww's internal microsecond ringbuffers and Kafka command streams.

## What You Are Building
A bare-metal optimized, asynchronous Rust microservice (`services/fix-gateway`). Concrete deliverables include:
- **FIX 5.0 SP2 Session & Application Engine:** Custom zero-allocation TCP server handling Logon (`MsgType=A`), Heartbeat (`MsgType=0`), Test Request (`MsgType=1`), Resend Request (`MsgType=2`), Sequence Reset (`MsgType=4`), New Order Single (`MsgType=D`), Order Cancel Request (`MsgType=F`), Order Cancel/Replace (`MsgType=G`), Execution Report (`MsgType=8`), and Order Cancel Reject (`MsgType=9`).
- **OUCH 4.2+ Binary Order Gateway:** Zero-copy binary parser and serializer for inbound Enter Order (`'O'`), Cancel Order (`'X'`), and outbound Order Accepted (`'A'`), Order Executed (`'E'`), Order Canceled (`'C'`), and Broken Trade (`'B'`) messages.
- **ITCH 5.0 Binary Market Data Multicast Streamer:** Binary encoder streaming System Event (`'S'`), Stock Directory (`'R'`), Order Book Directory (`'d'`), Add Order (`'A'`), Order Executed (`'E'`), Order Cancel (`'X'`), Order Delete (`'D'`), and Trade Message (`'P'`) packets.
- **Cancel-on-Disconnect (COD) Watchdog:** Heartbeat monitor terminating all resting open orders for institutional sessions within 50ms of TCP socket disconnection.
- **Drop-Copy Streamer:** Real-time FIX drop-copy stream multiplexing all execution reports to designated institutional compliance clearing desks and SEBI/IFSCA monitoring endpoints.
- **Internal IPC / Kafka Bridge:** Direct low-latency gRPC/ringbuffer connection to Order Matching Engine (Prompt 205) and pre-trade Risk Engine (Prompt 206).

## Scope Boundaries
- **In Scope:**
 - TCP session management, sequence number tracking, and automated replay/resend logic.
 - Parsing tag-value FIX 5.0 SP2 and custom Growww FIX tags for fractional shares.
 - Native binary parsing of OUCH order entry frames and ITCH market data frames.
 - Pre-trade session-level authorization (CompID, SenderSubID, IP whitelisting, mTLS certificate validation).
 - Cancel-on-Disconnect (COD) safety automation.
 - Translating wire messages to internal protobuf messages routed to matching engine and risk engine.
 - Real-time Drop-Copy session management.
- **Out of Scope / Handled Elsewhere:**
 - In-memory order book matching algorithm (Prompt 205).
 - User wallet cash holds and ledger balances (Prompt 203).
 - On-chain DvP smart contract settlement (Prompt 208, 306).
 - Retail REST/WebSocket fan-out (Prompt 207, 219).

## Technology to Use
- **Primary Language & Runtime:** Rust 2021 Edition (stable 1.78+) for zero-cost abstractions, deterministic latency without garbage collection pauses, memory safety, and high-performance socket I/O.
- **Networking & Async I/O:** `tokio` with custom `io_uring` integration or `mio` non-blocking sockets; `bytes` crate for zero-copy memory slice parsing.
- **Concurrency Primitives:** `crossbeam-channel` and `parking_lot` for lock-free intra-process ringbuffers.
- **FIX / Binary Parsing:** Custom zero-allocation byte parser; `nom` parser combinator for binary OUCH/ITCH packet framing.
- **Persistence & Session Storage:** `rocksdb` for persistent inbound/outbound sequence numbers and message replay stores.
- **Messaging & RPC:** `tonic` / `prost` for gRPC inter-service communication; `rdkafka` for asynchronous drop-copy event streaming.
- **Security & TLS:** `rustls` (or `native-tls`) for mTLS 1.3 with client certificate validation and strict cipher suites.

## Backend / Infra Touchpoints
- **Order Matching Engine (Prompt 205):** Direct microsecond gRPC/IPC channel for immediate order dispatch and execution event ingestion.
- **Risk & Margin Checks Service (Prompt 206):** In-line pre-trade institutional margin/credit limit validation.
- **RocksDB Local Storage:** `/var/data/growww/fix-gateway/sessions/` for fast sequence number WAL and message replay stores.
- **Apache Kafka:** Publishes drop-copy streams to `fix.dropcopy.v1` and ITCH data updates to `engine.itch.events.v1`.
- **Admin Back Office (Prompt 217):** Administration API for session provisioning, session resetting, and dynamic CompID enabling/disabling.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Pseudonymous Institutional Identifiers:** Institutional participant CompIDs (e.g., `PROP_LP_01`, `CITADEL_INDIA_01`) map to KYC-verified custodian demat account IDs and public ledger addresses (`0x...`) registered in `TransferComplianceHooks.sol` (Prompt 305).
- **Execution Receipt Tracking:** Outbound FIX Execution Reports (`MsgType=8`) include custom tag `10084` (`OnChainTxHash`) and tag `10085` (`BlockNumber`) once the atomic trade settlement is queued or confirmed on Hyperledger Besu via QBFT consensus.
- **Zero PII on Wire:** FIX/OUCH/ITCH protocol feeds transmit only cryptographic public addresses, pseudonymous account IDs, ISIN codes, prices, and quantities.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Rust Service:** Initialize Cargo workspace `services/fix-gateway` with aggressive performance release flags (`opt-level = 3`, `lto = "fat"`, `codegen-units = 1`).
2. **Implement Binary Frame Layouts:** Define packed byte structures for ITCH 5.0 and OUCH 4.2 messages with big-endian wire encodings and fixed-point price/quantity converters.
3. **Build Zero-Allocation FIX Lexer:** Implement tag-value FIX byte scanner capable of extracting tags, integers, timestamps, and fixed-point decimals without heap allocations.
4. **Implement Session State Machine:** Build FIX session layer managing sequence numbers, Heartbeat/TestRequest timers, gap detection, ResendRequest range processing, and sequence synchronization across restarts.
5. **Implement RocksDB Session Store:** Persist incoming and outgoing sequence numbers and raw message log in RocksDB with fsync guarantees for crash recovery.
6. **Implement OUCH Gateway Listener:** Create dedicated async TCP listener for OUCH binary connections with sub-microsecond binary deserialization.
7. **Implement ITCH Multicast Broadcaster:** Build UDP multicast and TCP replay server streaming ITCH 5.0 binary order book state updates to co-located subscribers.
8. **Build Pre-Trade Risk & Validation Interceptor:** Validate institutional permissions, maximum order value, tick size (0.05 INR), price collars, and credit line against Risk Engine before matching engine dispatch.
9. **Integrate Cancel-on-Disconnect (COD):** Implement async socket heartbeat monitor. If TCP FIN/RST or heartbeat timeout is detected, immediately synthesize bulk cancel request for all resting orders belonging to that session.
10. **Implement Drop-Copy Service:** Build FIX Drop-Copy endpoint streaming duplicate execution reports (`MsgType=8`) and order status events to compliance participant CompIDs.
11. **Implement Fractional Share FIX Tag Mapping:** Map fractional equity lots ($10^{-6}$ shares) to standard FIX tag `38` (`OrderQty`) using 6 decimal float formatting and custom tag `20038` (`MicroOrderQty`) for integer micro-shares.
12. **Connect to Matching Engine via gRPC/IPC:** Establish persistent bidirectional streaming gRPC connection to `services/matching-engine` for deterministic order intake and match receipt dispatch.
13. **Expose Prometheus Telemetry:** Export real-time metrics including parsing latency histograms (p50, p99, p99.9), active TCP sessions, messages/sec per CompID, and COD triggers.
14. **Build Protocol Fuzzing & Conformance Test Suite:** Write integration tests verifying FIX 5.0 SP2 spec compliance, sequence gap replay, simultaneous message bursts, and corrupted packet handling.
15. **Execute Benchmark & Stress Testing:** Run load generator streaming 50,000 FIX/OUCH orders per second, confirming sub-100 microsecond gateway processing latency.

## Interfaces / Contracts

### FIX Protocol Tag Specification (Custom Growww Dictionary)
```text
Standard Tags Supported:
  Tag 8    : BeginString (FIX.5.0SP2)
  Tag 9    : BodyLength
  Tag 35   : MsgType (A, 0, 1, 2, 4, 5, D, F, G, 8, 9)
  Tag 34   : MsgSeqNum
  Tag 49   : SenderCompID
  Tag 56   : TargetCompID (GROWWW_EXCHANGE)
  Tag 52   : SendingTime (UTC timestamp: YYYYMMDD-HH:MM:SS.sss)
  Tag 11   : ClOrdID (Client Order ID, max 36 chars)
  Tag 37   : OrderID (Exchange Order ID, UUID)
  Tag 48   : SecurityID (ISIN, e.g., INE002A01018)
  Tag 22   : SecurityIDSource (4 = ISIN)
  Tag 54   : Side (1 = Buy, 2 = Sell)
  Tag 38   : OrderQty (Fractional shares formatted to 6 decimal places, e.g. 10.500000)
  Tag 44   : Price (Limit Price in INR formatted to 4 decimal places, e.g. 2450.5000)
  Tag 40   : OrdType (1 = Market, 2 = Limit, 3 = Stop, 4 = Stop Limit, 9 = Pegged)
  Tag 59   : TimeInForce (0 = Day, 1 = GTC, 3 = IOC, 4 = FOK)
  Tag 150  : ExecType (0 = New, 1 = Partial Fill, 2 = Fill, 4 = Canceled, 8 = Rejected)
  Tag 39   : OrdStatus (0 = New, 1 = Partially Filled, 2 = Filled, 4 = Canceled, 8 = Rejected)
  Tag 10   : CheckSum (3-digit modulo 256 checksum)

Growww Custom Extension Tags:
  Tag 10038: MicroOrderQty (Integer representation of quantity in 10^-6 micro-shares)
  Tag 10044: PaisePrice (Integer representation of price in 10^-4 INR sub-paise)
  Tag 10084: OnChainTxHash (Settlement transaction hash on Hyperledger Besu)
  Tag 10085: SettlementBlockNumber (Ledger block number)
  Tag 10090: SelfTradePreventionMode (1 = Cancel Incoming, 2 = Cancel Resting, 3 = Decrement)
```

### OUCH Binary Packet Layouts (C-Style Structs in Rust)
```rust
#[repr(C, packed)]
pub struct OuchEnterOrder {
    pub message_type: u8,       // 'O' (0x4F)
    pub client_order_id: [u8; 16], // Alphanumeric / UUID
    pub side: u8,               // 'B' = Buy, 'S' = Sell
    pub micro_quantity: u64,    // Quantity * 1,000,000 (micro-shares)
    pub isin: [u8; 12],         // e.g., "INE002A01018"
    pub sub_paise_price: u64,   // Price in INR * 10,000 (paise * 100)
    pub time_in_force: u8,      // '0' = Day, '3' = IOC, '4' = FOK
    pub firm_id: [u8; 8],       // Institutional Member ID
    pub display: u8,            // 'Y' = Displayed, 'N' = Non-displayed (Dark)
    pub post_only: u8,          // 'Y' = Post Only, 'N' = Aggressive OK
}

#[repr(C, packed)]
pub struct OuchOrderAccepted {
    pub message_type: u8,       // 'A' (0x41)
    pub timestamp_ns: u64,      // Nanoseconds since UTC midnight
    pub client_order_id: [u8; 16],
    pub exchange_order_id: u64, // Monotonic engine order ID
    pub side: u8,
    pub micro_quantity: u64,
    pub isin: [u8; 12],
    pub sub_paise_price: u64,
    pub time_in_force: u8,
    pub order_state: u8,        // 'L' = Live on book
}
```

### ITCH 5.0 Market Data Message Layout
```rust
#[repr(C, packed)]
pub struct ItchAddOrder {
    pub message_type: u8,       // 'A' (0x41)
    pub stock_locate: u16,      // Internal stock index ID
    pub tracking_number: u16,
    pub timestamp_ns: u64,      // Nanoseconds since UTC midnight
    pub order_reference_id: u64,// Unique order reference
    pub side: u8,               // 'B' = Buy, 'S' = Sell
    pub micro_quantity: u64,    // Micro-shares (10^-6)
    pub isin: [u8; 12],
    pub sub_paise_price: u64,   // Price * 10,000
}
```

## Security & Compliance Notes
- **Strict Mutual TLS (mTLS):** All FIX and OUCH TCP connections require mutual TLS 1.3 with certificate fingerprint binding against pre-registered SEBI/IFSCA broker-dealer credentials.
- **Cancel-on-Disconnect (COD):** The gateway guarantees that in the event of network partition, socket drop, or heartbeat failure, all open limit orders for that session are cancelled within $< 50\text{ms}$ to eliminate unmonitored risk exposure.
- **Drop-Copy Auditing:** Unmodifiable drop-copy streams are mirrored real-time to exchange regulatory monitoring services complying with SEBI algorithmic trading circulars.
- **Deterministic Sequence Recovery:** Resend requests (`MsgType=2`) reconstruct historical session events directly from local RocksDB WAL without touching main transaction databases.

## Acceptance Criteria
- [ ] FIX 5.0 SP2 parser successfully passes full session lifecycle (Logon, Heartbeat, TestRequest, NewOrderSingle, Cancel, Logout).
- [ ] Supports both standard floating-point FIX tags (Tag 38/44) and fixed-point micro-tags (Tag 10038/10044) with zero precision loss.
- [ ] OUCH binary order entry processes new orders with end-to-end gateway parsing latency $< 5\mu\text{s}$.
- [ ] ITCH market data stream broadcasts live Level-3 order events over UDP multicast and TCP recovery ports.
- [ ] Cancel-on-Disconnect cancels 100% of resting orders within 50ms upon abrupt TCP termination.
- [ ] Session sequence numbers survive unexpected gateway process restarts without message duplication or sequence corruption.
- [ ] Drop-copy feed broadcasts identical execution reports to registered compliance audit endpoints.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (Architecture Overview), 103 (API Standards), 105 (Auth & mTLS), 205 (Order Matching Engine), 206 (Risk Engine).
- **Parallel Tasks:** 226 (Advanced Order Types Engine), 227 (Institutional Dark Pool Service).
- **Downstream Blockers:** 228 (Real-Time Market Surveillance Engine), 704 (SEBI Regulatory Reporting).
