# ADR-0026: Gasless Meta-Transactions via EIP-2771 Paymaster Sponsorship

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Technology Officer, Lead Blockchain Architect, Head of Product  

---

## 1. Context
To deliver a frictionless zero-fee trading experience where users pay strictly a 0.00% (No fee at all) platform fee with zero native gas overhead, users must not be required to acquire or hold native gas tokens (ETH / Besu gas) to sign and execute smart contract transactions.

---

## 2. Decision
Implement an EIP-2771 meta-transaction Paymaster architecture:
1. **Trusted Forwarder Integration:** Smart contracts (`SettlementDvP.sol`, `EquityToken.sol`, `IdentityRegistry.sol`) inherit `ERC2771Context` and trust the platform's canonical `Forwarder` contract.
2. **Off-Chain EIP-712 Signatures:** End users sign typed data (`ForwardRequest`) off-chain using their secure enclave or custody keys.
3. **Platform Paymaster Relayer:** The platform's automated relayer submits the transaction on-chain, paying gas from the Treasury reserve fee pool allocation.
4. **Sender Verification:** Contracts extract the authentic user address via `_msgSender()` instead of `msg.sender`.

---

## 3. Consequences
- **Positive:** Users experience a 100% web2-like seamless experience with zero gas fees.
- **Trade-offs:** Relayer infrastructure must maintain funded gas accounts and monitor relayer nonce queues to prevent transaction stalls.
