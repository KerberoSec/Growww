# ADR-0046: SEBI SCORES 2.0 & RBI Ombudsman Automated Grievance Gateway

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Compliance Officer, Head of Customer Support, Legal Lead  

---

## 1. Context
Statutory compliance with SEBI SCORES 2.0 (21-day SLA) and the RBI Integrated Ombudsman Scheme (30-day SLA) requires automated docketing, audit trail extraction, and real-time Action Taken Report (ATR) compilation.

---

## 2. Decision
1. **Direct SEBI SCORES 2.0 Integration:** Automated hourly polling and docketing against investor PAN and trading account UUIDs.
2. **Automated Diagnostic Triage:** Self-service bots resolve deposit tracking and missing memo tickets in $< 5\text{ seconds}$.
3. **ATR Auto-Compiler:** Compiles cryptographically signed audit proofs from Hyperledger Besu into digital Action Taken Reports submitted directly to regulatory portals.

---

## 3. Consequences
- **Positive:** 100% regulatory compliance with zero statutory SLA breaches.
- **Trade-offs:** Requires dedicated compliance escalation monitoring daemons.
