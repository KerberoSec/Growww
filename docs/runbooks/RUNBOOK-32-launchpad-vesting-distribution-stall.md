# Runbook 32: Launchpad Vesting Distribution Stall & Batch Recovery

**Runbook ID:** RBK-OPS-032  
**Severity Tier:** P2 (High)  
**Authority:** Smart Contracts Lead / Primary Markets Operations Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- Linear vesting streaming unlocks stall or fail to record on Hyperledger Besu.
- Claim transactions fail due to Paymaster nonce desynchronization or gas relayer stalls.

---

## 2. Remediation Workflow

```
[1. Identify Affected Offering & Vesting Vault Contract]
  - Query metrics: launchpad_vesting_claim_failures_total, paymaster_relay_errors
  - Identify project ID, token contract, and LinearVestingVault.sol address
                  |
                  v
[2. Check Paymaster Relayer Health]
  - Query Paymaster gas balance and pending nonce queue on Besu node
  - If nonce blocked, broadcast replacement cancellation transaction with higher gas
  - Top up Paymaster gas allocation from Treasury pool
                  |
                  v
[3. Trigger Batch Vesting Synchronization]
  - Trigger administrative vault sync:
      POST /api/v1/launchpad/admin/sync-vesting-schedule
      {"vault_address": "<VAULT_ADDRESS>"}
  - Update user claimable allowances in frontend cache
                  |
                  v
[4. Verify User Gasless Claims]
  - Test testnet/production claim transaction via EIP-2771 forwarder
  - Verify token balance credits directly to investor wallet
```
