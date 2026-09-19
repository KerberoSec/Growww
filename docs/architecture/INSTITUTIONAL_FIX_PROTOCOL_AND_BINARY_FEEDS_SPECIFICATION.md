# Institutional FIX Protocol & Low-Latency Binary Feeds Specification

This document defines the architecture, message formats, session management, and network topology for institutional access via Financial Information eXchange (FIX 4.4 / 5.0 SP2) and ultra-low latency binary protocols (ITCH/OUCH) on the Growww RWA Exchange.

---

## 1. Network Topology & Colocation Infrastructure

### 1.1 Colocation & Cross-Connects
Institutional desks, High-Frequency Trading (HFT) firms, and market makers connect via:
- **Direct 10Gbps / 25Gbps Fiber Cross-Connects:** In Tier-4 datacenters (Equinix MB1/MB2 Mumbai, GIFT City IFSC Datacenter).
- **Sub-Millisecond Edge Routing:** Dedicated edge gateways bypassing cloud ingress load balancers with Kernel-bypass (Solarflare OpenOnload / DPDK).

---

## 2. FIX 4.4 / 5.0 SP2 Protocol Architecture

### 2.1 Supported FIX Message Specifications

| Message Type | MsgType (Tag 35) | Direction | Description |
|---|---|---|---|
| **Logon** | `A` | Client -> Server | Authenticates with HMAC-SHA256 signature in Tag 96 (`RawData`). |
| **Heartbeat** | `0` | Bi-directional | Periodic keep-alive signal. |
| **New Order Single** | `D` | Client -> Server | Submits Limit, Market, Stop, or Iceberg order. |
| **Order Cancel Request** | `F` | Client -> Server | Requests cancellation of resting order. |
| **Order Cancel/Replace** | `G` | Client -> Server | Modifies order price or size with in-place queue preservation. |
| **Execution Report** | `8` | Server -> Client | Confirms order acceptance, rejection, cancellation, or partial/full fill. |
| **Order Cancel Reject** | `9` | Server -> Client | Rejects cancel request if order is already filled or unknown. |
| **Drop Copy (Execs)** | `8` | Server -> Client | Dedicated read-only stream of all execution reports for back-office risk. |

### 2.2 FIX Authentication & Sequence Recovery
- **Authentication:** Tag 553 (`Username`), Tag 554 (`Password` - API Key), Tag 96 (`RawData` - HMAC-SHA256 of `SendingTime` + `TargetCompID`).
- **Resend Request (Tag 35 = `2`):** In the event of sequence gaps, the FIX engine replays missing messages from the append-only journal in exact sequence.

---

## 3. Ultra-Low Latency Binary Protocol (ITCH / OUCH)

For latency-critical HFT participants, Growww provides direct binary protocols:

### 3.1 Binary OUCH (Order Entry)
- Fixed-length binary message format (no string parsing overhead).
- **Packet Wire Format:**
  - `Header` (1 Byte Packet Type + 4 Bytes Sequence ID + 8 Bytes Timestamp Nanoseconds).
  - `Payload` (8 Bytes Client Order ID, 8 Bytes ISIN Numeric ID, 1 Byte Side, 8 Bytes Scaled Price, 8 Bytes Scaled Quantity, 1 Byte TIF).
- **Median Ingress Latency:** $< 8.5\mu\text{s}$ round-trip.

### 3.2 Binary ITCH (Market Data Feed)
- MoldUDP64 multicast market data distribution.
- Emits atomic Level-3 tick-by-tick order additions, executions, cancellations, and order book state transitions.
