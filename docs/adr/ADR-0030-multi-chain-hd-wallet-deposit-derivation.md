# ADR-0030: Hierarchical Deterministic (HD) Multi-Chain Wallet Address Derivation

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Technology Officer, Lead Blockchain Custody Architect, Security Lead  

---

## 1. Context
Users depositing external cryptocurrencies (BTC, ETH, USDT, USDC, SOL, etc.) require unique, deterministic deposit addresses per blockchain that can be generated dynamically without storing private keys in hot database instances.

---

## 2. Decision
1. **Hierarchical Deterministic Derivation Standards:**
   - **Bitcoin (BTC Native SegWit Bech32):** BIP-84 derivation path: `m/84'/0'/0'/0/{user_id_index}`.
   - **Ethereum & EVM L2s (Arbitrum, Base, Polygon, Optimism):** BIP-44 / EIP-2334 derivation path: `m/44'/60'/0'/0/{user_id_index}` (single unified EVM address per user).
   - **Solana (SOL & SPL Tokens):** BIP-44 Ed25519 derivation path: `m/44'/501'/0'/{user_id_index}'`.
   - **TRON (TRC-20 USDT):** BIP-44 derivation path: `m/44'/195'/0'/0/{user_id_index}`.
2. **Master Extended Public Key (xpub / ypub / zpub) Storage:**
   - Master private seeds are generated and isolated inside FIPS 140-3 Level 4 hardware security modules (AWS CloudHSM / HashiCorp Vault with HSM backend).
   - Only the hardened Account Master Extended Public Key (`zpub` / `xpub`) is provisioned to the `wallet-service` and `custody-adapter` to derive child public addresses on demand ($O(1)$ derivation).
3. **Memo / Tag Handling for Shared Address Chains:**
   - Chains with high account creation costs or tag standards (XRP, TON) utilize a pooled exchange master address with a mandatory 32-bit unique integer `Deposit Tag / Memo` tied to the user account ID.

---

## 3. Consequences
- **Positive:** Hot microservices cannot leak private keys; addresses are deterministically regenerable; sub-millisecond address generation.
- **Trade-offs:** Requires automated sweep relayers to consolidate funds from individual deposit addresses to hot/cold vaults.
