# ADR-0011: Fixed-Point Scaled Integer Arithmetic for Financial Math

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Systems Engineer, Lead Financial Architect  

---

## 1. Context
Floating-point arithmetic introduces non-deterministic rounding errors across polyglot architectures (Rust, Go, Python, Solidity) and risks precision loss in high-frequency trading calculations.

---

## 2. Decision
Enforce fixed-point scaled integer representations across all systems:
- Prices: $10^4$ scaling ($0.0001 \text{ INR} = 1 \text{ weINR}$).
- Quantities: $10^6$ scaling ($0.000001 \text{ share} = 1 \text{ micro-unit}$).
- 128-bit intermediate multiplication (`u128` in Rust) with explicit overflow checks before downscaling.

---

## 3. Consequences
- **Positive:** Bit-for-bit determinism across all backend microservices, matching engines, and smart contracts.
- **Trade-offs:** Requires scaling utilities and strict type wrappers across all client SDKs.
