# Custody Vault & Threshold Coordinator Specification

## 1. Executive Overview
The Custody Vault smart contract suite coordinates cold, warm, and hot cryptocurrency storage on Hyperledger Besu, interfacing with the off-chain MPC-TSS (Threshold Signature Scheme) infrastructure.

## 2. Vault Topology & Security Tiers
- **Hot Settlement Relayer Vault**: Holds operational float (1-2 days of settlement volume) for automated trading liquidity.
- **Warm Rebalancing Vault**: 2-of-3 MPC-TSS controlled vault for processing larger withdrawals and treasury rebalancing.
- **Cold Reserve Vault**: 3-of-5 institutional multi-signature with hardware security modules (HSM) and 48-hour timelock for cold storage.

## 3. Smart Contract Inventory
- `CustodyVault.sol`: Primary vault holding reserve assets, managing deposits, and dispatching authorized withdrawals.
- `TimelockCoordinator.sol`: Enforces mandatory 24-hour time delay on critical parameter changes and high-value transfers.
- `CircuitBreaker.sol`: Automated anomaly detection contract that pauses vault outflows if withdrawals exceed volumetric velocity thresholds.

## 4. Withdrawal Verification & EIP-712 Signing
Withdrawals must be authorized by a cryptographic multi-signature or MPC-TSS threshold signature:
```solidity
struct WithdrawalRequest {
    bytes32 withdrawalId;
    address recipient;
    address token;
    uint256 amount;
    uint256 fee; // Fixed at 0 (Zero withdrawal fee at launch)
    uint256 nonce;
    uint256 validUntil;
}
```
Signatures are validated against the authorized MPC signer set. Nonces are tracked in a bitmap to ensure each withdrawal can only be executed exactly once.

## 5. Storage Layout & Upgrade Safety
- Implements UUPS (Universal Upgradeable Proxy Standard) with OpenZeppelin contracts.
- Explicit storage gap: `uint256[50] private __gap;` reserved for future variable expansions.
- Employs `ReentrancyGuardUpgradeable` on all withdrawal endpoints.
