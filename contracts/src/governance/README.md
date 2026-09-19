# Consortium Validator Governance & Timelock Controllers

## Purpose & Scope
The `governance` smart contract module is part of the sovereign Growww / NBSE on-chain clearing infrastructure on Hyperledger Besu.

## Core Architectural Responsibilities
Multi-sig timelock contracts managing validator set updates, contract upgrades, and emergency circuit breakers.

## Invariants & Security Standards
- **Zero-PII**: Never stores personal identity data on-chain; only stores 32-byte cryptographic hashes.
- **Reentrancy Protection**: All state-mutating functions utilize OpenZeppelin `ReentrancyGuardUpgradeable`.
- **Formal Verification**: All math utilizes strict fixed-point arithmetic with overflow/underflow reverts.

## Local Testing
```bash
forge test --match-path governance/** -vvv
```
