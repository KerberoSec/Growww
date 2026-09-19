# Perpetual Futures & Options Clearing Contracts

## Purpose & Scope
The `derivatives` smart contract module is part of the sovereign Growww / NBSE on-chain clearing infrastructure on Hyperledger Besu.

## Core Architectural Responsibilities
Smart contracts handling margin requirements, mark price evaluations, and futures funding payments.

## Invariants & Security Standards
- **Zero-PII**: Never stores personal identity data on-chain; only stores 32-byte cryptographic hashes.
- **Reentrancy Protection**: All state-mutating functions utilize OpenZeppelin `ReentrancyGuardUpgradeable`.
- **Formal Verification**: All math utilizes strict fixed-point arithmetic with overflow/underflow reverts.

## Local Testing
```bash
forge test --match-path derivatives/** -vvv
```
