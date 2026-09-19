# Cross-Chain HTLC Atomic Swap Contracts

## Purpose & Scope
The `bridge` smart contract module is part of the sovereign Growww / NBSE on-chain clearing infrastructure on Hyperledger Besu.

## Core Architectural Responsibilities
Hashed Timelock Contracts (HTLC) for trust-minimized swaps between Bitcoin, Ethereum, and Besu.

## Invariants & Security Standards
- **Zero-PII**: Never stores personal identity data on-chain; only stores 32-byte cryptographic hashes.
- **Reentrancy Protection**: All state-mutating functions utilize OpenZeppelin `ReentrancyGuardUpgradeable`.
- **Formal Verification**: All math utilizes strict fixed-point arithmetic with overflow/underflow reverts.

## Local Testing
```bash
forge test --match-path bridge/** -vvv
```
