# Runbook 21: Hot Wallet Liquidity Rebalancing & Cold Vault Sweeping

**Runbook ID:** RBK-OPS-021  
**Severity Tier:** P2 (High)  
**Authority:** Treasury Lead / Head of Custody Operations  

---

## 1. Description & Trigger Conditions
Triggered when:
- Hot Wallet balance for an asset (BTC, ETH, USDT, SOL) breaches operational bounds:
  - Low Liquidity Threshold ($< 0.5\%$ of asset reserves): Automated withdrawals risk queuing or stalling.
  - High Excess Liquidity Threshold ($> 1.5\%$ of asset reserves): Hot wallet holds excessive funds, violating security policy.
- Deposit addresses hold accumulated un-swept token balances requiring batch gas sponsorship consolidation.

---

## 2. Remediation Workflow

```
[1. Evaluate Balance & Rebalancing Direction]
  - Query custody-service: GET /api/v1/custody/reserves/liquidity-status
  - Check Hot Wallet balance vs Target Reserve Ratio (1.0% target)
                  |
                  +-----------------------------------+
                  |                                   |
                  v                                   v
       [A. Low Hot Wallet Balance]         [B. High Hot Wallet Balance / Sweeps]
       - Calculate required top-up amount   - Calculate excess surplus amount
       - Initiate Warm Vault to Hot Wallet  - Initiate Hot Wallet to Warm/Cold sweep
         rebalance transfer via MPC TSS     - For deposit addresses: broadcast batch
       - Sign with 2-of-3 MPC shares          sweep with gas relayer sponsorship
                  |                                   |
                  +-----------------+-----------------+
                                    |
                                    v
[2. Multi-Signature Policy Validation]
  - Automated policy engine checks: daily withdrawal caps, 2FA, risk scores
  - Broadcast transaction to blockchain mempool with dynamic EIP-1559 priority fee
                  |
                  v
[3. Confirm On-Chain Finality & Reconcile Ledger]
  - Await required block confirmations (e.g. 12 blocks for Ethereum, 3 blocks for Bitcoin)
  - Post internal double-entry rebalance journal:
      Debit: Account 1210 (Hot Wallet Vault)
      Credit: Account 1220 (Warm Vault)
  - Verify hot wallet reserve ratio returns to target range (0.8% - 1.2%)
```
