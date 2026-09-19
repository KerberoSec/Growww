# Product Definition & Institutional Entity Architecture

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Head of Legal, Compliance & Institutional Architecture  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Executive Summary & Legal Entity Segmentation

To eliminate architectural and regulatory ambiguity, the Growww platform maintains strict structural separation between two distinct legal and operational entities:

1. **Phase 1: Growww Retail Broker & Depository Participant (Growww Technologies India Pvt. Ltd.)**
2. **Phase 2: National Blockchain Stock Exchange (NBSE Ltd.) - Sovereign FMI & Recognized Exchange**

```
+----------------------------------------------------------------------------------------------------+
|                                    INSTITUTIONAL ENTITY TOPOLOGY                                   |
|                                                                                                    |
|  +-------------------------------------------------+  +-----------------------------------------+  |
|  |             PHASE 1: RETAIL BROKER / DP         |  |       PHASE 2: RECOGNIZED EXCHANGE      |  |
|  |            (Growww Technologies India)          |  |                 (NBSE Ltd.)             |  |
|  |                                                 |  |                                         |  |
|  |  * SEBI Registered Stock Broker                 |  |  * Recognized Stock Exchange (SCRA s.4) |  |
|  |  * Depository Participant (NSDL/CDSL)           |  |  * Clearing Corporation (SEBI SECC Regs)|  |
|  |  * Client Margin & Internal Ledger              |  |  * Public Order Book & Multilateral CLOB|  |
|  |  * Direct Market Access (DMA) to NSE/BSE        |  |  * Atomic DvP on Hyperledger Besu Ledger|  |
|  |  * Strict 1:1 Demat Pool Vaulting               |  |  * Statutory SGF & IPF Administration   |  |
|  +-------------------------------------------------+  +-----------------------------------------+  |
+----------------------------------------------------------------------------------------------------+
```

---

## 2. Comparative Matrix: Broker vs Recognized Exchange

| Dimension | Phase 1: Growww Broker & DP | Phase 2: NBSE Recognized Exchange |
|---|---|---|
| **Legal Entity** | Growww Technologies India Pvt. Ltd. | NBSE Financial Market Infrastructure Ltd. |
| **Primary Regulators** | SEBI, RBI, FIU-IND | SEBI (Market Regulation Dept), IFSCA, RBI |
| **Licensing Framework** | SEBI (Stock Brokers) Regs 1992, SEBI (DP) Regs 2018 | Securities Contracts (Regulation) Act 1956 (s. 4), SEBI (SECC) Regs 2018 |
| **Capital Requirement** | Base Minimum Capital: INR 50 Crores | Net Worth: INR 100 Crores + SGF Core Buffer |
| **Order Book Role** | Internal client routing & aggregation; DMA to NSE/BSE | Central Limit Order Book (CLOB); multilateral price discovery |
| **Custody & Settlement** | Segregated client pool accounts at NSDL/CDSL | Atomic on-chain DvP on permissioned Besu ledger + Depository Gateway |
| **Revenue Model** | 0.00% (Zero Fee) platform fee + statutory levies (STT, Stamp, GST) | Exchange turnover charge (0.00% (Zero Fee) flat) + Clearing & Settlement fees |
| **Service Scope** | `201_user_service` through `224_referral_growth_service` | `244_nbse_fee_distribution`, `329_nbse_settlement_dvp`, `715_nbse_surveillance` |

---

## 3. Phased Implementation Roadmap

### Phase 1: Regulated Retail Broker & Depository Bridge
- **Timeline:** Current Active Phase.
- **Operating Scope:** Custody of fractional Indian equities and MCX commodities in regulated depository accounts (NSDL/CDSL). 
- **Execution Mechanism:** Orders matched against internal pre-funded inventory or routed via DMA to primary exchanges.
- **Ledger Invariant:** Every token represents a strict 1:1 claim against verified physical assets custodied with licensed depository participants.

### Phase 2: Recognized Sovereign Stock Exchange (NBSE)
- **Timeline:** Regulatory Sandbox Promotion -> Full Commercial Licensing.
- **Operating Scope:** Independent multi-lateral trading venue, sovereign clearing corporation, and atomic settlement infrastructure.
- **Corporate Ring-Fencing:** Managed under an independent board with public interest directors under SEBI SECC Regulations. Phase 1 broker entities connect strictly as Trading Members (TM) with zero preferential routing or order internalisation.

---

## 4. Prompt Catalog Tagging & Scope Mapping

Every prompt in `Prompt/prompts/` is categorized by target phase:

- **Phase 1 (Broker / DP):** Prompts `101` to `224`, `401` to `615`, `701` to `710`, `801` to `910`.
- **Phase 2 (NBSE Exchange / FMI):** Prompts `225` to `244`, `301` to `330`, `711` to `718`, `811` to `813`, `911` to `914`.
- **Cross-Phase Foundations:** Protocols, domain types, cryptography, CI/CD, and infrastructure configs.
