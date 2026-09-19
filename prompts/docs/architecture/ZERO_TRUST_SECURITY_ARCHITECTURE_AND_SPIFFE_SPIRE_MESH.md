# Zero-Trust Security Architecture & SPIFFE/SPIRE Identity Mesh

**Specification ID:** SPEC-ARCH-052-ZT  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Enterprise Cyber Security & Cryptographic Identity Mesh  
**Owner:** Information Security & Cryptographic Platform Group  

---

## 1. Executive Summary & Security Invariants
The Growww / NBSE infrastructure enforces strict **Zero-Trust Architecture (NIST SP 800-207)** across all 24 distributed microservices and Hyperledger Besu consortium nodes:
- **Zero Implicit Trust**: No service, pod, or network interface is trusted based on IP address, network boundary, or cloud subnet.
- **Cryptographic Workload Identity**: Every pod receives a cryptographically verifiable SPIFFE ID issued by an in-cluster SPIRE agent.
- **Microsecond Mutual TLS (mTLS)**: 100% of inter-service gRPC and HTTP/2 traffic is encrypted using short-lived X.509 SVID certificates rotating every 60 minutes.
- **Hardware-Rooted Attestation**: SPIRE server keys are backed by AWS CloudHSM / Nitro Enclaves, preventing unauthorized node impersonation.

---

## 2. SPIFFE ID Taxonomy & Identity Attestation

```
+----------------------------------------------------------------------------------------------------+
| CRYPTOGRAPHIC WORKLOAD ATTESTATION & SHORT-LIVED SVID ROTATION                                     |
|                                                                                                    |
|  [ Kubernetes Pod Starts ] ---> [ Node SPIRE Agent Attests Pod cgroup & UID ]                      |
|                                                |                                                   |
|                                                v                                                   |
|                               [ Query SPIRE Server via mTLS ]                                      |
|                                                |                                                   |
|                                                v                                                   |
|                               [ Issue 60-Minute X.509 SVID ]                                       |
|                               - SPIFFE ID: spiffe://growww.internal/ns/settlement/sa/relayer       |
|                                                |                                                   |
|                                                v                                                   |
|                      [ Envoy Sidecar Ingests SVID via Secret Discovery API ]                       |
|                                                |                                                   |
|                                                v                                                   |
|          [ mTLS Handshake with Peer Service Validating Cryptographic SAN Identity ]                |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Micro-Segmentation & Cilium eBPF Network Policies:
1. **Default-Deny All**: Network policies drop all cross-namespace traffic by default.
2. **Explicit Cryptographic Whitelisting**:
   - `order-service` can connect strictly to `matching-engine` on port 50051.
   - `matching-engine` can emit strictly to Kafka cluster on port 9092.
   - External internet ingress is permitted only through cloud WAF and Envoy Edge Gateways.
