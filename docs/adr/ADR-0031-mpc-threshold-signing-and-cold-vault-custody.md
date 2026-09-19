# ADR-0031: Multi-Party Computation (MPC) Threshold Signing & 3-Tier Custody Architecture

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Information Security Officer, Head of Custody, Infrastructure Lead  

---

## 1. Context
Managing institutional crypto reserves and processing high-volume automated user withdrawals requires eliminating single points of failure (private key compromise) while maintaining sub-minute withdrawal execution and 100% reserve protection.

---

## 2. Decision
1. **3-Tier Custodial Reserve Segregation:**
   - **Tier 1: Cold Storage Vaults ($> 95\%$ of Reserves):** Air-gapped, geographic multisig hardware vaults (M-of-N hardware signers, e.g., 3-of-5 threshold) with quorum approval requirements.
   - **Tier 2: Warm Vaults ($3\% - 4\%$ of Reserves):** HSM-backed threshold accounts with automated rebalancing triggers.
   - **Tier 3: Hot Wallet ($\le 1\%$ of Reserves):** Real-time automated withdrawal buffer bounded by per-minute and per-hour circuit breakers.
2. **2-of-3 MPC Threshold Signature Scheme (TSS):**
   - Implements Lindell / GG20 threshold ECDSA and FROST EdDSA.
   - **Key Share 1:** Platform Secure Enclave (AWS Nitro Enclave / confidential compute).
   - **Key Share 2:** Co-signing Policy Engine (validates 2FA, risk scores, whitelist cooldowns, withdrawal limits).
   - **Key Share 3:** Disaster Recovery Backup Custodian (stored in physical bank safe-deposit box).
3. **Zero Full Key Assembly:** A complete private key is never constructed in memory at any point during signing.

---

## 3. Consequences
- **Positive:** Immune to single-node compromise, insider theft, or single-cloud provider outages.
- **Trade-offs:** Requires MPC network latency coordination ($50\text{ms} - 150\text{ms}$) during withdrawal transaction assembly.
