# Settlement Smart Contracts

## Overview
Authoritative smart contracts on Hyperledger Besu executing atomic Delivery-versus-Payment (DvP Model 1) clearing with a **Zero-Fee Starting Policy and Dynamic Governance Fee Controller**:
- **0.00% Starting Fee ("0 Means 0 in All")**: Starts at strictly 0.00% fee for all trades (Maker and Taker).
- **Future Fee Expandability**: `FeeController.sol` allows authorized governance to adjust fees later if desired, protected by a 48-hour timelock and immutable ceiling (max 0.50%).
- **0.00% Demo Fee**: Permanent zero-fee paper trading sandbox.
- **0% TDS**: Zero tax withholding on-chain.
- **Zero Gas**: 100% sponsored gas for retail users via ERC-4337 Paymasters.

## Key Contracts
- `DvPAtomicSettlement.sol`: Single-trade and netted-batch atomic settlement.
- `FeeController.sol`: Dynamic governance fee parameter controller.
- `OrderRegistry.sol`: On-chain commitments and cryptographic fill states.
- `SettlementGuaranteeFund.sol`: Central Counterparty guarantee reserves.
- `PaymasterRelayer.sol`: Sponsored gas abstraction for end users.

For complete architectural details, see [SPECIFICATION.md](SPECIFICATION.md).
