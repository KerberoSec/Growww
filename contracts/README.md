# Smart Contracts Architecture (Hyperledger Besu EVM)

## Executive Overview
The `contracts/` directory houses the complete smart contract architecture governing the sovereign clearing and settlement of spot cryptocurrency trades (BTC/USDT), tokenized assets, and demo paper balances on a permissioned Hyperledger Besu enterprise blockchain.

## Local Testing Framework
- **Tooling**: Foundry suite (`forge`, `cast`, `anvil`) for compilation, fuzzing, and gas optimization.
- **Local Sandbox**: Deployable directly to a local single-node Besu testnet or Anvil fork with 2-second block intervals.
- **Gasless Paymaster**: Local paymaster contracts sponsor gas for registered test accounts.

## Production Invariants
- **Zero-PII On-Chain**: Only cryptographic hashes `keccak256(buyer, seller, price, quantity, timestamp)` are committed on-chain.
- **Deterministic Finality**: 4-validator QBFT consensus delivers immediate finality without chain forks or reorganizations.
- **Delivery-versus-Payment (DvP)**: Atomic swap guarantees: base asset and quote asset transfer simultaneously or revert entirely.

## Contract Modules
- `src/settlement/`: Atomic DvP settlement contracts and clearing guarantee funds.
- `src/tokens/`: ERC-20, ERC-1400, and ERC-3643 security token implementations (wBTC, wUSDT, eINR).
- `src/demo/`: Testnet virtual faucet and paper trading token contracts.
- `src/custody/`: Multi-signature and MPC-TSS coordinator contracts.
- `src/compliance/`: KYC accreditation registries and statutory transfer restriction hooks.
- `src/bridge/`: HTLC cross-chain atomic swap contracts.
