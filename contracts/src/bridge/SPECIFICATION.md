# Cross-Chain Bridge & Settlement Adapter Specification

## 1. Executive Overview
The Cross-Chain Bridge coordinates 1:1 backed asset transfers between external Layer-1 networks (Bitcoin, Ethereum, Tron, Polygon, Arbitrum) and sovereign wrapped tokens (`wBTC`, `wUSDT`) on Hyperledger Besu.

## 2. Ingress & Egress Lifecycles
1. **Deposit (Ingress)**:
   - User deposits external asset to designated MPC-controlled vault address.
   - Multi-chain listener confirms block depth (e.g. 19 blocks on Tron, 12 blocks on Ethereum).
   - Ingress adapter submits cryptographic deposit attestation signed by 3-of-5 institutional MPC signers.
   - Bridge contract calls `WrappedToken.mint(userAddress, normalizedAmount)`.
2. **Withdrawal (Egress)**:
   - User initiates withdrawal on Hyperledger Besu via EIP-712 signed request.
   - Bridge contract burns `wBTC` or `wUSDT` on Besu.
   - MPC cluster monitors burn event and executes native transfer to external destination address.

## 3. Decimal Normalization & Precision Security
- When bridging USDT from Ethereum/Tron (6 decimals) to Besu `wUSDT` (6 decimals), 1:1 exact unit equivalence is maintained.
- When bridging Bitcoin (8 decimals) to `wBTC` (8 decimals), 1:1 exact satoshi equivalence is maintained.
- All bridge calculations enforce checked arithmetic with overflow prevention.

## 4. Zero-Fee Invariant
The platform charges 0.00% protocol fee for bridge operations; only standard external miner/network fees are passed through to the user.
