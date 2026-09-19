# ADR-0045: MPC DKG Ceremony & Multi-Region Hot-Standby Disaster Recovery

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Information Security Officer, Head of Infrastructure  

---

## 1. Context
Generating institutional cryptographic keys without single points of failure requires formal air-gapped Distributed Key Generation (DKG) ceremonies, while exchange uptime requires multi-region hot-standby failover.

---

## 2. Decision
1. **Air-Gapped DKG Ceremony SOP:** Generate threshold MPC key shares inside isolated Faraday cage environments across 3 officers without ever assembling the master private key.
2. **Optical PSBT Signing:** Cold Vault transfers execute via optical dynamic QR streams scanned by offline hardware signers.
3. **Multi-Region Disaster Recovery:** Active-passive multi-region architecture between Mumbai and Frankfurt with RPO = 0 and automated RTO $< 60\text{ seconds}$.

---

## 3. Consequences
- **Positive:** Zero single point of failure in key management; bulletproof disaster resilience.
- **Trade-offs:** Strict operational ceremony protocols require 3-officer physical coordination.
