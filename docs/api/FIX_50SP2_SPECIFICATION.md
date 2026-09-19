# Growww / NBSE FIX 5.0 SP2 Protocol Specification

## 1. Overview & Institutional Infrastructure

The Growww / NBSE FIX Gateway (`services/fix-gateway`) provides ultra-low latency, binary-optimized connectivity using the Financial Information eXchange (FIX) 5.0 SP2 protocol. It is engineered for proprietary trading desks, market makers, and institutional prime brokers colocated in Equinix Mumbai and GIFT City datacenters.

### Connection Architecture
- **Primary Sessions:** Colocated cross-connects (10 Gbps fiber) in Equinix Mumbai (`ap-south-1`).
- **Session Types:**
  - **Order Entry Session:** Full duplex for order placement and cancellation.
  - **Drop Copy Session:** Real-time risk management and fill streaming across multiple trader sub-accounts.
- **Transport:** TCP/IP over TLS 1.3 with client mutual certificate verification.

---

## 2. Standard Supported FIX Messages

| Message Type | Tag 35 Value | Description | Direction |
| :--- | :--- | :--- | :--- |
| **Logon** | `A` | Initiates authenticated session with heartbeat interval. | Client <-> Server |
| **Heartbeat** | `0` | Keeps connection alive during inactivity (default 30s). | Client <-> Server |
| **New Order Single** | `D` | Places a new limit or market order. | Client -> Server |
| **Order Cancel Request** | `F` | Requests cancellation of an open resting order. | Client -> Server |
| **Order Cancel/Replace** | `G` | In-place reduction or price modification. | Client -> Server |
| **Execution Report** | `8` | Confirms acceptance, rejection, partial fill, or fill. | Server -> Client |
| **Order Cancel Reject** | `9` | Rejects order cancel/replace request. | Server -> Client |

---

## 3. Mandatory Tags & Field Specifications

### 3.1 New Order Single (`35=D`)
```
8=FIX.5.0SP2 | 9=182 | 35=D | 49=PROP_DESK_MUMBAI | 56=NBSE_PROD | 34=102 | 52=20260919-10:20:00.102 | 11=CL_ORD_88301 | 55=BTC-USDT | 54=1 | 60=20260919-10:20:00.100 | 40=2 | 44=64500.50 | 38=0.50000000 | 59=0 | 10=218 |
```
- `11` (ClOrdID): Unique client order reference.
- `55` (Symbol): Instrument symbol (`BTC-USDT`).
- `54` (Side): `1` for Buy, `2` for Sell.
- `40` (OrdType): `1` for Market, `2` for Limit.
- `44` (Price): Order price in quote currency.
- `38` (OrderQty): Quantity in base currency satoshis.
- `59` (TimeInForce): `0` for Day, `1` for GTC, `3` for IOC, `4` for FOK.

### 3.2 Execution Report (`35=8`)
```
8=FIX.5.0SP2 | 9=245 | 35=8 | 49=NBSE_PROD | 56=PROP_DESK_MUMBAI | 34=103 | 52=20260919-10:20:00.108 | 37=NBSE_ORD_99401 | 11=CL_ORD_88301 | 17=EXEC_8839201 | 150=F | 39=2 | 55=BTC-USDT | 54=1 | 38=0.50000000 | 44=64500.50 | 32=0.50000000 | 31=64500.50 | 151=0.00000000 | 14=0.50000000 | 6=64500.50 | 136=0.0000 | 10=045 |
```
- `37` (OrderID): Exchange assigned persistent ID.
- `150` (ExecType): `0` (New), `F` (Trade / Fill), `4` (Cancelled), `8` (Rejected).
- `39` (OrdStatus): `0` (New), `1` (Partially Filled), `2` (Filled), `4` (Cancelled).
- `32` (LastQty) & `31` (LastPx): Fill size and execution price.
- `136` (FeeRate): Confirms exact 0.00% (Zero Fee) (0.0000) fee rate at launch.

---

## 4. Latency SLA & Asymmetric Speed Bump

1. **Deterministic Execution:** Sub-15 microsecond matching latency from sequencer receipt to execution report dispatch.
2. **500μs Speed Bump:** Aggressive orders received over FIX are subject to the platform's standard 500μs asymmetric speed bump, ensuring market makers and retail limits are protected against predatory latency front-running.
