# Demo Faucet & Sandbox Paper Trading Specification

## 1. Executive Overview
Governs virtual testnet tokens (`vUSDT`, `vBTC`) used exclusively within the risk-free demo paper trading environment on Hyperledger Besu Testnet.

## 2. Complete Physical & Logical Isolation
- **Sandbox Chain ID**: Operates on an isolated testnet chain or sandboxed namespace (`IS_DEMO = true`).
- **Ledger Isolation**: Backed by TigerBeetle `Ledger ID 2` (Demo Paper Trading), completely disjoint from `Ledger ID 1` (Real Capital).
- **Zero Asset Value**: Demo tokens possess zero financial value and are programmatically non-withdrawable and non-bridgeable to mainnet.

## 3. Smart Contracts Inventory
- `VirtualFaucet.sol`: Dispenses starting capital for simulated paper trading.
  - Grants: 10,000 vUSDT and 1.0 vBTC per claim.
  - Rate Limit: 1 claim per 24 hours per verified account ID.
  - Sybil Prevention: Enforces signed claim authorization from the backend gateway with CAPTCHA verification.
- `VirtualToken.sol`: Open-mint ERC-20 token simulating real-world transaction mechanics and receipts.
- `SandboxReset.sol`: Allows users to burn existing testnet balances and restore default 10,000 vUSDT starting capital.

## 4. Security Invariants
- Faucet dispensing cannot exceed 100,000 vUSDT total allocation per user lifetime.
- Demo contracts are isolated from real custody and settlement contracts via separate contract deployments and access control roles.
