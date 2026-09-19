# Sovereign Asset & Stablecoin Token Implementations

## Purpose & Scope
The `tokens` smart contract module is part of the sovereign Growww / NBSE on-chain clearing infrastructure on Hyperledger Besu.

## Core Architectural Responsibilities
ERC-20, ERC-1400, and ERC-3643 token standards for wBTC, wUSDT, and RBI e-Rupee tokenized representations.

## Invariants & Security Standards
- **Zero-PII**: Never stores personal identity data on-chain; only stores 32-byte cryptographic hashes.
- **Reentrancy Protection**: All state-mutating functions utilize OpenZeppelin `ReentrancyGuardUpgradeable`.
- **Formal Verification**: All math utilizes strict fixed-point arithmetic with overflow/underflow reverts.

## Local Testing
```bash
forge test --match-path tokens/** -vvv
```
