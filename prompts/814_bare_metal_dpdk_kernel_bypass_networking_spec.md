# 814 - Bare-Metal DPDK & Solarflare Kernel-Bypass Network Architecture

## Purpose
In modern institutional digital securities exchanges and high-frequency trading (HFT) co-location venues, traditional Linux operating system kernel networking imposes severe latency penalties. The standard Linux networking stack introduces non-deterministic latency jitter through hardware interrupt service routines (ISRs), deferred softirqs, process scheduling context switches, memory copies across kernel/user space boundaries (`sk_buff` allocation and socket buffer copying), and lock contention in the TCP/IP stack. Under sustained order bursts, this architecture exhibits tail latency spikes exceeding 15 to 50 microseconds.

The purpose of this specification is to architect, configure, and implement Growww's bare-metal kernel-bypass network engine (`infra/kernel-bypass`). By leveraging Intel DPDK (Data Plane Development Kit) 23.11 LTS and Solarflare Onload / EF_VI (EtherFabric Virtual Interface) user-space networking, the platform eliminates the Linux kernel network stack entirely from the critical tick-to-trade path. Physical network interface card (NIC) ring buffers are directly mapped into user-space process memory via PCIe Direct Memory Access (DMA), Linux 1GB HugePages, and dedicated poll-mode driver (PMD) CPU cores. This architecture achieves deterministic sub-microsecond packet ingestion and dispatch for institutional FIX/OUCH trading sessions, ITCH market data multicast feeds, and Hyperledger Besu consensus message exchange in exchange co-location environments.

## What You Are Building
A high-performance, bare-metal network capture and DMA packet processor (`infra/kernel-bypass`) directly servicing the FIX / OUCH Protocol Gateway (Prompt 225) and the Order Matching Engine (Prompt 205). Key deliverables include:
- `infra/kernel-bypass/dpdk-core/`: DPDK 23.11 LTS Poll Mode Driver (PMD) rx/tx engine utilizing Intel E810/X710 and NVIDIA Mellanox ConnectX-6 Dx NICs, polling hardware descriptor rings directly in user space without hardware interrupts.
- `infra/kernel-bypass/efvi-solarflare/`: Solarflare Onload and EF_VI (EtherFabric Virtual Interface) user-space network driver interfacing directly with Solarflare XtremeScale X2522/SFN8522 NIC hardware ring buffers for ultra-low-latency Ethernet frame processing.
- `infra/kernel-bypass/dma-ring/`: Lock-free, cache-line aligned single-producer single-consumer (SPSC) and multi-producer single-consumer (MPSC) shared-memory ring buffers mapped into 1GB Linux HugePages (`/dev/hugepages1G`) for zero-copy data handoff.
- `infra/kernel-bypass/packet-filter/`: Hardware-accelerated flow classification utilizing NIC Flow Director (FDIR) and Receive Side Scaling (RSS) to demultiplex inbound UDP ITCH market data multicast feeds and TCP FIX/OUCH trading sessions directly into dedicated CPU core queues.
- `infra/kernel-bypass/shm-ipc/`: Ultra-low-latency zero-copy POSIX shared-memory IPC bridge passing 64-byte aligned packet descriptors between kernel-bypass pollers, the FIX Gateway (`services/fix-gateway`), and the Matching Engine (`services/matching-engine`).
- `infra/kernel-bypass/ptp-clock/`: IEEE 1588v2 PTP (Precision Time Protocol) hardware timestamping engine capturing sub-nanosecond ingress and egress physical layer (PHY) timestamps directly from NIC hardware registers.
- `infra/kernel-bypass/besu-consensus-accelerator/`: High-speed user-space UDP/TCP transport pipe accelerating QBFT consensus rounds, vote broadcasts, and proposed block dissemination between co-located Hyperledger Besu validator nodes.
- `infra/kernel-bypass/monitoring/`: Nanosecond-precision telemetry exporter gathering NIC ring fullness, PCIe bus stalls, packet drop counters, and latency percentiles exposed via Prometheus metrics and shared-memory counters.

## Scope Boundaries
- **In Scope:**
  - Bare-metal OS kernel boot parameter optimization (`isolcpus`, `nohz_full`, `rcu_nocbs`, `intel_idle.max_cstate=0`, `intel_iommu=on`, `iommu=pt`).
  - User-space Poll Mode Driver (PMD) lifecycle management for DPDK 23.11 LTS and Solarflare EF_VI / OpenOnload.
  - 1GB Linux HugePages reservation, allocation, and NUMA-node memory alignment.
  - Dedicated CPU core pinning and lock-free thread polling execution loops (100% spin loop, zero context switches, zero interrupts).
  - Hardware packet classification via NIC Flow Director (FDIR) and RSS hash routing.
  - IEEE 1588v2 PTP sub-nanosecond hardware ingress/egress timestamp extraction.
  - Cache-aligned SPSC and MPSC lock-free ring buffers in pinned shared memory for zero-copy handoff to Matching Engine and FIX Gateway.
  - User-space TCP/IP stack acceleration for FIX 5.0 SP2 and OUCH trading sessions.
  - Accelerated UDP multicast ingestion for ITCH market data and QBFT consensus broadcasts.
  - Ingress MAC and IP validation to prevent network packet spoofing at line rate.
- **Out of Scope / Handled Elsewhere:**
  - Limit Order Book matching logic and price-time priority execution (Prompt 205).
  - FIX protocol tag-value decoding, dictionary validation, and session sequence persistence (Prompt 225).
  - Pre-trade risk rule validation and account balance reservations (Prompt 203, 206).
  - Retail WebSocket client fan-out and REST gateway routing (Prompt 207, 219).
  - Smart contract transaction execution, receipt generation, and state storage (Prompt 208, 301, 305).
  - General corporate LAN / WAN routing and public Internet CDN ingress (Prompt 802, 805).

## Technology to Use
- **Languages & Runtimes:**
  - **C (C11) & C++20:** Primary languages for low-level DPDK PMD wrappers, hardware register manipulation, and Solarflare EF_VI primitives. Justification: Complete control over CPU instruction generation, memory barriers, cache-line alignment, and zero runtime abstraction overhead.
  - **Rust 2021 Edition (stable 1.78+):** Memory-safe interfaces, IPC ring-buffer management, and integration bridges to downstream services via `unsafe` FFI bindings.
- **Kernel-Bypass Frameworks:**
  - **Intel DPDK 23.11 LTS:** `librte_eal` (Environment Abstraction Layer), `librte_ethdev` (Ethernet Device API), `librte_ring` (lock-free multi/single-consumer ring queues), `librte_mempool` (fixed-size object memory pool manager). Supported PMDs: `ice` (Intel 800 Series 25/100GbE), `i40e` (Intel X710 10/40GbE), and `mlx5` (NVIDIA Mellanox ConnectX-5/6 Dx).
  - **Solarflare OpenOnload & EF_VI:** Direct hardware RX/TX ring access via Onload Enterprise stack and `ef_vi` API on Solarflare XtremeScale X2522 and SFN8522 NICs.
- **Memory Subsystem & HugePages:**
  - Linux 1GB HugePages (`hugepagesz=1G hugepages=32`) configured via GRUB bootloader parameters and mounted at `/dev/hugepages1G`.
  - Non-Uniform Memory Access (NUMA) node memory affinity configured via `libnuma` (`numa_alloc_onnode`) ensuring memory pools reside in the same NUMA domain as the PCIe root complex.
- **Kernel Isolation & Core Pinning:**
  - Linux kernel boot parameters: `isolcpus=2-15,18-31 nohz_full=2-15,18-31 rcu_nocbs=2-15,18-31 intel_idle.max_cstate=0 processor.max_cstate=0 intel_pstate=disable default_hugepagesz=1G hugepagesz=1G hugepages=32 iommu=pt intel_iommu=on audit=0 nosoftlockup`.
  - Core pinning via `pthread_setaffinity_np` and thread scheduling class `SCHED_FIFO` with priority 99.
- **Time Synchronization:**
  - IEEE 1588v2 PTP hardware clock synchronization using Linux `ptp4l` and `phc2sys`, synchronized against the exchange co-location Grandmaster Clock.

## Backend / Infra Touchpoints
- **Exchange Co-Location Optical Fiber Links:** Direct cross-connect to exchange co-location 10/25GbE fiber NICs (NSE Co-Location Rack at BKC Mumbai, BSE Co-Location at Fort Mumbai, GIFT City IFSCA Data Centers).
- **PCIe Hardware Bus & DMA:** PCIe Gen4 / Gen5 x16 slots with Direct Memory Access mapping via the Linux `vfio-pci` driver, ensuring safe user-space DMA with hardware IOMMU protection.
- **High-Performance Order Matching Engine (Prompt 205):** Lock-free shared-memory SPSC ring buffer transferring raw pre-parsed order descriptors directly into the Matching Engine cache hierarchy in under 200 nanoseconds.
- **FIX / OUCH Protocol Gateway (Prompt 225):** Bidirectional zero-copy shared-memory queues for inbound institutional FIX/OUCH frames and outbound Execution Reports (`MsgType=8`).
- **Real-Time Pre-Trade Risk Engine (Prompt 206):** Parallel inline packet inspection tap feeding order metadata to pre-trade risk filters in sub-50 nanoseconds before order book insertion.
- **Hyperledger Besu Validator Nodes (Prompt 302, 802, 812):** Dedicated co-located low-latency peer-to-peer network interface accelerating QBFT consensus vote propagation.
- **Observability Stack & Telemetry (Prompt 806):** Hardware NIC ring drop metrics, PCIe bus saturation counters, and nanosecond-level packet processing latency histograms exposed to Prometheus via DPDK telemetry sockets and eBPF probes.

## Blockchain Interaction
The kernel-bypass networking engine directly accelerates consensus communication among co-located Hyperledger Besu validator nodes:
- **Low-Latency QBFT Consensus Acceleration:** Co-located exchange validator nodes run private permissioned Hyperledger Besu instances participating in QBFT (Quorum Byzantine Fault Tolerance) consensus. The kernel-bypass engine intercepts UDP/TCP consensus packets (`consensus.besu.qbft.v1`) on dedicated physical VLANs, routing validator proposals, commit signatures, and round-change votes directly through user-space ring buffers.
- **Consensus Round-Trip Reduction:** By eliminating kernel context switches, socket buffering, and TCP stack traversal, peer-to-peer round-trip consensus messaging latency is cut from ~250 microseconds to under 15 microseconds. This latency reduction enables deterministic sub-100-millisecond block intervals required for continuous Delivery-versus-Payment (DvP) trade finality without introducing backpressure to the high-frequency trading engine.
- **Zero-Copy Block Propagation:** Proposed blocks and batch cryptographic match proofs emitted by the Matching Engine are packetized directly from hugepage memory and dispatched to peer validator NICs via user-space DMA without intervening memory copies.
- **Zero-PII Compliance:** The kernel-bypass consensus accelerator operates strictly on raw cryptographic byte streams (BLS/ECDSA signatures, keccak256 block hashes, Merkle roots, and state diffs). No Personally Identifiable Information (PII) is ever processed or decoded by the networking engine.

## Step-by-Step Build Instructions
1. **Configure Host BIOS & Linux Kernel Parameters:** Configure the bare-metal server BIOS for maximum performance (disable C-states, disable P-states, disable Intel SpeedStep, disable Hyper-Threading if core isolation demands it). Update `/etc/default/grub` with low-latency parameters: `isolcpus=2-15,18-31 nohz_full=2-15,18-31 rcu_nocbs=2-15,18-31 intel_idle.max_cstate=0 processor.max_cstate=0 intel_pstate=disable default_hugepagesz=1G hugepagesz=1G hugepages=32 iommu=pt intel_iommu=on audit=0 nosoftlockup`. Run `update-grub` and reboot.
2. **Mount and Verify 1GB HugePages:** Configure `/etc/fstab` to mount 1GB HugePages on `/dev/hugepages1G`. Verify allocation across NUMA nodes via `cat /sys/devices/system/node/node*/hugepages/hugepages-1048576kB/nr_hugepages`.
3. **Bind Physical NICs to VFIO Driver:** Unbind target 10/25GbE optical fiber NIC ports from default Linux kernel drivers (`ixgbe`, `i40e`, `ice`, `mlx5_core`) and bind them to `vfio-pci` using `dpdk-devbind.py` or driver sysfs interfaces.
4. **Compile and Install Intel DPDK 23.11 LTS:** Build DPDK 23.11 LTS using `meson` and `ninja` with optimization flags: `meson setup build -Dprefix=/opt/dpdk -Dbuildtype=release -Ddefault_library=static -Dc_args="-march=native -O3"`. Install libraries and header files.
5. **Install Solarflare OpenOnload & EF_VI Drivers:** Compile and install Solarflare OpenOnload enterprise kernel modules (`sfc`, `onload`) and user-space libraries (`libefvi`, `libonload`). Configure user-space interface capabilities for Solarflare X2522 NICs.
6. **Implement DPDK Initialization and Port Configuration:** Write `infra/kernel-bypass/dpdk-core/src/init.c`. Initialize DPDK EAL (`rte_eal_init`), configure Ethernet device queues (`rte_eth_dev_configure`), allocate NUMA-local packet memory pools (`rte_pktmbuf_pool_create`), and configure RX/TX ring descriptors (512 descriptors per queue).
7. **Implement Solarflare EF_VI Vi-Allocation Module:** Write `infra/kernel-bypass/efvi-solarflare/src/efvi_rx.c`. Allocate virtual interfaces (`ef_vi_alloc`), allocate DMA-pinned memory buffers, and register memory regions with the Solarflare adapter for hardware DMA packet delivery.
8. **Configure Hardware Flow Director & RSS Filtering:** Implement NIC Flow Director (FDIR) and Receive Side Scaling (RSS) rules using DPDK Flow API (`rte_flow_create`). Direct UDP multicast market data (ITCH) to Core 2, TCP institutional FIX sessions to Cores 3-6, and Hyperledger Besu consensus frames to Core 7.
9. **Implement Core-Pinned Polling Threads:** Author tight, non-blocking polling loops using `rte_eth_rx_burst` and `ef_eventq_poll`. Pin each polling worker to its assigned isolated core via `pthread_setaffinity_np` and set thread priority to `SCHED_FIFO` 99.
10. **Build Lock-Free Shared-Memory Ring Buffers:** Implement `infra/kernel-bypass/dma-ring/` providing cache-line padded (64-byte alignment) single-producer single-consumer (SPSC) ring buffers in POSIX shared memory (`shm_open`, `mmap`) backed by 1GB HugePages.
11. **Implement IEEE 1588 PTP Hardware Timestamping:** Integrate hardware timestamp extraction (`RTE_MBUF_F_RX_IEEE1588_TMST` / `EF_VI_RX_TIMESTAMPS`). Query hardware PHY registers to record ingress nanosecond timestamps into the packet descriptor header before memory handoff.
12. **Connect Shared-Memory IPC to FIX Gateway and Matching Engine:** Author Rust FFI bindings (`infra/kernel-bypass/shm-ipc/bindings/rust`) allowing `services/fix-gateway` (Prompt 225) and `services/matching-engine` (Prompt 205) to read and write packet descriptors directly from/to the shared-memory ring buffers without copying payload data.
13. **Implement Besu Consensus Accelerator:** Author `infra/kernel-bypass/besu-consensus-accelerator/`. Capture incoming QBFT consensus datagrams in user space, extract consensus payloads, and forward them directly to the local Besu validator node via shared memory or local domain sockets.
14. **Enforce Line-Rate Anti-Spoofing Filters:** Implement hardware and software ingress header validation. Drop packets whose source MAC does not match the configured exchange Gateway MAC, or whose source IP deviates from authorized institutional co-location subnets.
15. **Execute Verification and Latency Benchmarking Suite:** Run automated loopback and hardware packet generator tests (using MoonGen or TRex over 25GbE fiber) measuring tick-to-trade ingress processing latency, packet drop rates under burst, and PTP timestamp accuracy.

## Interfaces / Contracts

### Cache-Aligned Packet Descriptor Structure (`infra/kernel-bypass/include/growww_packet_desc.h`)
```c
#ifndef GROWWW_PACKET_DESC_H
#define GROWWW_PACKET_DESC_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define GROWWW_DESC_ALIGN 64
#define GROWWW_MAX_PAYLOAD_SIZE 1500

typedef enum {
    PACKET_TYPE_UNKNOWN       = 0x00,
    PACKET_TYPE_FIX_TCP       = 0x01,
    PACKET_TYPE_OUCH_TCP      = 0x02,
    PACKET_TYPE_ITCH_UDP      = 0x03,
    PACKET_TYPE_BESU_QBFT_UDP = 0x04,
    PACKET_TYPE_HEARTBEAT     = 0x05
} packet_type_t;

typedef enum {
    INGRESS_PORT_EXCHANGE_PRIMARY   = 0x01,
    INGRESS_PORT_EXCHANGE_SECONDARY = 0x02,
    INGRESS_PORT_INTERNAL_CONSENSUS = 0x03
} ingress_port_t;

// Exactly 64 bytes - occupies precisely one L1/L2 cache line
typedef struct __attribute__((aligned(GROWWW_DESC_ALIGN))) {
    uint64_t hw_ingress_timestamp_ns; // IEEE 1588 hardware PHY timestamp
    uint64_t sw_enqueue_timestamp_ns; // TSC timestamp upon ring enqueue
    uint32_t packet_seq_num;          // Hardware or poller monotonic sequence
    uint16_t payload_len;             // Length of payload in hugepage buffer
    uint8_t  packet_type;             // packet_type_t enum
    uint8_t  ingress_port;            // ingress_port_t enum
    uint32_t session_id;              // FIX / OUCH session identifier
    uint32_t client_ip;               // IPv4 address in network byte order
    uint16_t client_port;             // TCP/UDP port
    uint16_t vlan_id;                 // Ingress 802.1Q VLAN tag
    uint64_t hugepage_offset;         // Absolute offset into 1GB hugepage memory pool
    uint8_t  flags;                   // Bit 0: Checksum Valid, Bit 1: PTP Valid, Bit 2: Drop Flag
    uint8_t  reserved[7];             // Zero padding to enforce exact 64-byte size
} growww_packet_desc_t;

_Static_assert(sizeof(growww_packet_desc_t) == 64, "Packet descriptor must be exactly 64 bytes");

#ifdef __cplusplus
}
#endif

#endif // GROWWW_PACKET_DESC_H
```

### Lock-Free SPSC Ring Buffer Interface (`infra/kernel-bypass/include/growww_spsc_ring.h`)
```c
#ifndef GROWWW_SPSC_RING_H
#define GROWWW_SPSC_RING_H

#include <stdint.h>
#include <stdatomic.h>
#include "growww_packet_desc.h"

#define RING_CAPACITY_POWER_OF_TWO 65536 // Must be power of 2
#define RING_MASK (RING_CAPACITY_POWER_OF_TWO - 1)

// Lock-free Single-Producer Single-Consumer (SPSC) Ring Buffer
// Uses cache-line padding (64 bytes) to prevent false sharing between producer and consumer
typedef struct {
    // Producer control block (Cache line 1)
    _Atomic uint64_t head __attribute__((aligned(64)));
    uint64_t head_cache; // Cached tail value for producer to minimize cache-line bounces
    uint8_t  pad1[48];

    // Consumer control block (Cache line 2)
    _Atomic uint64_t tail __attribute__((aligned(64)));
    uint64_t tail_cache; // Cached head value for consumer
    uint8_t  pad2[48];

    // Ring descriptor storage (Contiguous array of 64-byte descriptors)
    growww_packet_desc_t entries[RING_CAPACITY_POWER_OF_TWO] __attribute__((aligned(64)));
} growww_spsc_ring_t;

// Push descriptor into ring (Producer only - Poll Mode Driver)
static inline int growww_spsc_ring_push(growww_spsc_ring_t *ring, const growww_packet_desc_t *desc) {
    uint64_t current_head = atomic_load_explicit(&ring->head, memory_order_relaxed);
    
    // Check available space using cached tail
    if ((current_head - ring->head_cache) >= RING_CAPACITY_POWER_OF_TWO) {
        ring->head_cache = atomic_load_explicit(&ring->tail, memory_order_acquire);
        if ((current_head - ring->head_cache) >= RING_CAPACITY_POWER_OF_TWO) {
            return -1; // Ring is full
        }
    }

    ring->entries[current_head & RING_MASK] = *desc;
    atomic_store_explicit(&ring->head, current_head + 1, memory_order_release);
    return 0;
}

// Pop descriptor from ring (Consumer only - Matching Engine / FIX Gateway)
static inline int growww_spsc_ring_pop(growww_spsc_ring_t *ring, growww_packet_desc_t *out_desc) {
    uint64_t current_tail = atomic_load_explicit(&ring->tail, memory_order_relaxed);

    // Check available items using cached head
    if (current_tail == ring->tail_cache) {
        ring->tail_cache = atomic_load_explicit(&ring->head, memory_order_acquire);
        if (current_tail == ring->tail_cache) {
            return -1; // Ring is empty
        }
    }

    *out_desc = ring->entries[current_tail & RING_MASK];
    atomic_store_explicit(&ring->tail, current_tail + 1, memory_order_release);
    return 0;
}

#endif // GROWWW_SPSC_RING_H
```

### Rust FFI Binding for Packet Descriptor (`infra/kernel-bypass/shm-ipc/bindings/rust/src/descriptor.rs`)
```rust
#[repr(C, align(64))]
#[derive(Debug, Clone, Copy)]
pub struct PacketDescriptor {
    pub hw_ingress_timestamp_ns: u64,
    pub sw_enqueue_timestamp_ns: u64,
    pub packet_seq_num: u32,
    pub payload_len: u16,
    pub packet_type: u8,
    pub ingress_port: u8,
    pub session_id: u32,
    pub client_ip: u32,
    pub client_port: u16,
    pub vlan_id: u16,
    pub hugepage_offset: u64,
    pub flags: u8,
    pub _reserved: [u8; 7],
}

const _: () = assert!(std::mem::size_of::<PacketDescriptor>() == 64);
const _: () = assert!(std::mem::align_of::<PacketDescriptor>() == 64);

impl PacketDescriptor {
    #[inline(always)]
    pub fn is_ptp_valid(&self) -> bool {
        (self.flags & 0x02) != 0
    }

    #[inline(always)]
    pub fn is_checksum_valid(&self) -> bool {
        (self.flags & 0x01) != 0
    }

    #[inline(always)]
    pub fn is_dropped(&self) -> bool {
        (self.flags & 0x04) != 0
    }
}
```

### Linux Kernel Boot & Tuning Manifest (`infra/kernel-bypass/config/sysctl-low-latency.conf`)
```ini
# /etc/sysctl.d/99-growww-low-latency.conf
# Virtual memory and HugePages
vm.nr_hugepages = 32
vm.hugetlb_shm_group = 1001
vm.zone_reclaim_mode = 0
vm.swappiness = 0
vm.dirty_ratio = 10
vm.dirty_background_ratio = 5

# Network socket buffer maximums
net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
net.core.rmem_default = 33554432
net.core.wmem_default = 33554432
net.core.optmem_max = 2048576
net.core.netdev_max_backlog = 100000

# Disable TCP slow start after idle and selective acknowledgments tuning
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_low_latency = 1
net.ipv4.tcp_timestamps = 0
net.ipv4.tcp_sack = 1
net.ipv4.tcp_dsack = 0

# Disable CPU power throttling and IPC bottlenecks
kernel.sched_rt_runtime_us = -1
kernel.sched_migration_cost_ns = 5000000
kernel.numa_balancing = 0
```

## Security & Compliance Notes
- **Line-Rate Anti-Spoofing & Ingress Validation:** The kernel-bypass poller enforces strict layer-2 and layer-3 validation before passing packets to downstream rings. Packets whose source MAC address fails to match authorized exchange routers, or whose IP header source falls outside designated institutional IP subnets, are dropped at line rate and recorded in security counters.
- **Dedicated Isolated Network Namespaces:** Management traffic (SSH, Prometheus scraping, GitOps runners) is strictly isolated to separate physical NICs and Linux network namespaces (`ip netns`). Kernel-bypass trading NICs possess zero Linux kernel IP endpoints and are unreachable from public or corporate networks.
- **SEBI Co-Location Circular Adherence:**
  - Implements the requirements of SEBI Circulars on Algorithmic Trading and Co-Location Facilities (ensuring fair and equitable access to all market participants).
  - All incoming order streams are polled in strict round-robin deterministic order across configured virtual interface queues to eliminate preferential queue processing.
  - Optical fiber link lengths between the exchange demarcation point and Growww's co-location server NICs are calibrated to equal lengths (preventing physical path latency disparities).
  - Hardware PTP IEEE 1588 timestamps are recorded at the physical layer for every order submission, cancellation, and execution report, maintaining continuous clock synchronization within 100 nanoseconds of the National Stock Exchange (NSE) / Bombay Stock Exchange (BSE) master clock.
- **Memory Safety & Process Isolation:** POSIX shared-memory regions mapped into hugepage memory are initialized with strict access permissions (`0600`) owned exclusively by the unprivileged `growww-trader` system user. Hardware IOMMU protection (`intel_iommu=on`, `iommu=pt`) ensures that PCIe DMA operations cannot corrupt memory ranges outside designated packet buffer pools.
- **Failover & Watchdog Protection:** A dedicated watchdog daemon monitors poller heartbeat counters in shared memory. If a kernel-bypass polling worker halts or misses heartbeats for more than 50 microseconds, the system initiates automated failover: resetting hardware queues and triggering Cancel-on-Disconnect (COD) procedures across all active trading sessions.

## Acceptance Criteria
- [ ] Kernel-bypass engine (`infra/kernel-bypass`) successfully initializes DPDK 23.11 LTS and Solarflare EF_VI drivers on target co-location servers.
- [ ] 32 x 1GB Linux HugePages (32GB total) are reserved and verified across designated NUMA nodes without memory fragmentation.
- [ ] Polling threads are pinned to isolated CPU cores (`isolcpus`), demonstrating zero context switches (`voluntary_ctxt_switches = 0`, `nonvoluntary_ctxt_switches = 0`) over continuous 60-minute benchmark runs.
- [ ] Ingress tick-to-trade network processing latency (from physical NIC wire arrival to shared-memory descriptor delivery) remains under 800 nanoseconds at the 99.9th percentile.
- [ ] Sustained line-rate throughput: zero packet drops under continuous 10GbE / 25GbE line rate burst loads (minimum 14.88 million packets per second for 64-byte packets).
- [ ] Hardware Flow Director (FDIR) correctly steers UDP ITCH multicast traffic, TCP FIX sessions, and Besu consensus frames to their designated CPU core queues.
- [ ] IEEE 1588v2 PTP hardware timestamps are extracted and validated with less than 100 nanoseconds offset against the co-location Grandmaster Clock.
- [ ] Lock-free shared-memory SPSC ring buffer transfers packet descriptors to the Order Matching Engine (Prompt 205) and FIX Gateway (Prompt 225) with zero buffer copies.
- [ ] Hyperledger Besu consensus round-trip message latency between validator nodes is reduced by at least 70% compared to standard Linux kernel networking.
- [ ] Anti-spoofing ingress filters drop unauthorized MAC/IP frames at line rate without degrading poller throughput.
- [ ] All audit logs, hardware drop statistics, and latency metrics are exported to the observability stack (Prompt 806).

## Suggested Order / Dependencies
- **Prerequisites:**
  - Bare-metal hardware provisioning: Co-location servers equipped with Intel E810 / Solarflare X2522 NICs and PCIe Gen4/Gen5 slots.
  - Prompt 108 (Environment Strategy and Configuration Management).
  - Prompt 205 (High-Performance Order Matching Engine).
  - Prompt 225 (FIX 5.0 SP2 / ITCH / OUCH Low-Latency Binary Trading Gateway).
  - Prompt 302 (Hyperledger Besu Network Architecture & Consensus Configuration).
- **Parallel Tasks:**
  - Prompt 206 (Real-Time Pre-Trade Risk Engine): Direct inline tap from packet descriptor ring.
  - Prompt 806 (Observability Stack & Low-Latency Telemetry): Integration with DPDK telemetry sockets and PTP offset monitoring.
  - Prompt 812 (Dual-Environment Testnet Sandbox and Mainnet Isolation).
- **Next Steps:**
  - Prompt 901 (Unit, Integration & Hardware-in-the-Loop Latency Testing Strategy): Benchmarking with hardware traffic generators.
  - Prompt 908 (Production Co-Location Launch & Failover Runbook): Operating procedures for live exchange co-location cutover.
