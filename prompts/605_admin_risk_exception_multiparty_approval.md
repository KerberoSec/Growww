# 605 - Admin Console: Risk Exceptions & Multi-Party Approval UI

## Purpose
Provides an institutional multi-party governance and risk management portal (`apps/growww_admin/app/(dashboard)/approvals/`) for operations executives, risk directors, and compliance officers. Enforces cryptographic multi-signature (M-of-N threshold) authorization and maker-checker workflows for high-risk system operations: large INR withdrawals, manual digital security token mint/burn approvals against physical custody receipts, emergency trading halts (circuit breakers), and on-chain contract parameter upgrades. Eliminates single points of failure and prevents rogue administrative actions by requiring multi-party cryptographic co-signing on the permissioned Hyperledger Besu blockchain.

## What You Are Building
- Multi-party approval queue and execution terminal (`apps/growww_admin/app/(dashboard)/approvals/`).
- Web3 hardware wallet integration (Ledger, Trezor, MetaMask Institutional) using Viem, Wagmi, and Web3Modal.
- EIP-712 structured data visualizer converting raw transaction calldata into human-readable, unambiguous parameter summaries prior to cryptographic signing.
- Quorum progress tracker (e.g. 2 of 3 signatures collected) with live verified signer identities, public keys, and timestamp logs.
- Emergency circuit-breaker and market halt trigger console with dual-confirmation safeguards.
- Historical execution archive indexing past multi-sig proposals, cryptographic signature manifests, and on-chain transaction receipts.

## Scope Boundaries
- **In Scope:**
 - Multi-sig dashboard UI, pending approval queue, and proposal detail drawer.
 - Web3 hardware wallet connection and session management (Viem / Wagmi).
 - EIP-712 typed data signing interface and signature aggregation.
 - Smart contract proposal creation, co-signing, and execution submission.
 - Emergency trading pause/unpause controls and real-time quorum visualizer.
- **Out of Scope / Handled Elsewhere:**
 - `MultiSigGovernance.sol` smart contract implementation (Prompt 307).
 - Pre-trade risk calculation service (Prompt 206).
 - Admin backend microservice (Prompt 217).
 - HSM key custody for automated validator nodes (Prompt 311).

## Technology to Use
- **Next.js 14 App Router, React 18/19, TypeScript 5.4+:** Provides secure SSR rendering, server actions for off-chain metadata caching, and reactive client components for Web3 wallet interaction.
- **Viem 2.x & Wagmi 2.x:** High-performance, lightweight TypeScript interfaces for EVM-compatible smart contract calls, typed data encoding, and RPC event subscriptions. Justification: Viem offers superior bundle size, type inference, and native support for custom permissioned EVM chains compared to legacy Web3.js / Ethers.js.
- **Web3Modal v3 / WalletConnect AppKit:** Enables seamless hardware wallet connectivity (Ledger Nano S/X, Trezor) via standard transport protocols.
- **Tailwind CSS & shadcn/ui:** Institutional-grade UI components with real-time status banners and progress gauges.
- **Framer Motion:** Smooth visual state transitions for multi-signature quorum indicators.

## Backend / Infra Touchpoints
- **Admin Backend Service (Prompt 217):** Fetches off-chain proposal metadata, supporting depository attestation attachments, and sends alert notifications (Prompt 211).
- **Hyperledger Besu JSON-RPC Node (Prompt 302):** Directly queries and submits signed transactions to `MultiSigGovernance.sol`.
- **Audit Log Service (Prompt 218):** Records all proposal views, signature attempts, rejections, and execution triggers.

## Blockchain Interaction
- **Direct Multi-Sig Smart Contract Interaction:** Interacts directly with `MultiSigGovernance.sol`, `DigitalSecurityToken.sol`, and `SettlementDvP.sol`.
- **EIP-712 Structured Data Signing:** Officers cryptographically sign structured payloads (`ActionProposal(uint256 proposalId, address targetContract, uint256 value, bytes calldata, uint256 nonce, uint256 deadline)`).
- **On-Chain Proposal Lifecycle:** Invokes `submitTransaction()`, `confirmTransaction()`, and `executeTransaction()` via the connected admin hardware wallet once threshold $M \le \text{signatures}$ is satisfied.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus, Chain ID: 13371).
- **Core Smart Contracts Interfaced:**
 - `MultiSigGovernance.sol` (Tracks registered admin signers, threshold count, proposal states, and signature records).
 - `DigitalSecurityToken.sol` (Executes mint/burn/pause functions gated by the multi-sig governance contract address).
 - `SettlementDvP.sol` (Executes emergency settlement pause and fund release functions).
- **Cryptographic Security:** Administrative private keys remain securely stored on hardware security tokens (Ledger/Trezor); no raw private keys are ever stored in web application memory or browser storage.

## Step-by-Step Build Instructions
1. Scaffold admin approvals route structure (`apps/growww_admin/app/(dashboard)/approvals/`) with sub-views for active proposals, history, and emergency halt.
2. Configure Wagmi client with Web3Modal / AppKit connecting to the Hyperledger Besu permissioned RPC endpoint with custom chain metadata.
3. Build Hardware Wallet Connection Modal supporting Ledger, Trezor, and WalletConnect with active chain verification.
4. Implement Signer Authentication Check verifying that the connected wallet address matches an authorized multi-sig governance owner in `MultiSigGovernance.sol`.
5. Build Proposal Creation Wizard supporting pre-configured templates:
 - Template A: Token Minting against Custody Confirmation (ISIN, Units, Custody Ref).
 - Template B: Token Burning for Custody Release.
 - Template C: Large Fiat Withdrawal Override (>₹10,00,000).
 - Template D: Emergency Trading Halt / Resume.
6. Implement Calldata Encoder / Decoder translating smart contract ABI calls into human-readable action summaries.
7. Build Pending Approvals Queue table displaying Proposal ID, Action Type, Target Contract, Proposer, Expiration Timer, and Current Quorum (e.g. "1 of 3 signed").
8. Build Proposal Detail Drawer displaying full proposal parameters, attached depository PDF attestation documents, and verified EIP-712 payload.
9. Implement Quorum Progress Visualizer bar showing each authorized signer's name, public key, status (Signed / Pending / Rejected), and timestamp.
10. Implement "Sign with Hardware Wallet" button triggering Wagmi `useSignTypedData()` to sign the EIP-712 payload and submit the signature to the governance contract or backend aggregation service.
11. Implement "Execute Transaction" button, dynamically enabled once required signature threshold (e.g. 2 of 3) is satisfied, submitting `executeTransaction()` to Hyperledger Besu.
12. Build Emergency Circuit Breaker Console with one-click "Halt All Trading" trigger requiring immediate dual-officer confirmation.
13. Add real-time Besu event listeners (`ProposalSubmitted`, `ProposalConfirmed`, `ProposalExecuted`) to update UI status instantly across all active admin sessions.
14. Write end-to-end Playwright tests with mock Web3 providers verifying proposal creation, multi-signature aggregation, and execution flow.

## Interfaces / Contracts
```typescript
import { type Address, type Hex } from 'viem';

export interface MultiSigProposal {
  proposalId: string;
  targetContract: Address;
  actionType: 'MINT_TOKEN' | 'BURN_TOKEN' | 'PAUSE_TRADING' | 'LARGE_WITHDRAWAL' | 'UPGRADE_CONTRACT';
  description: string;
  decodedParameters: Record<string, unknown>;
  rawCalldata: Hex;
  proposerAddress: Address;
  proposerName: string;
  requiredQuorum: number; // e.g. 2
  currentSignaturesCount: number;
  signers: Array<{
    address: Address;
    name: string;
    hasSigned: boolean;
    signedAt?: string;
    signature?: Hex;
  }>;
  status: 'PENDING' | 'READY_TO_EXECUTE' | 'EXECUTED' | 'REJECTED' | 'EXPIRED';
  createdAt: string;
  expiresAt: string;
  executionTxHash?: Hex;
}

export const EIP712_PROPOSAL_TYPES = {
  GrowwwProposal: [
    { name: 'proposalId', type: 'uint256' },
    { name: 'targetContract', type: 'address' },
    { name: 'actionType', type: 'string' },
    { name: 'callDataHash', type: 'bytes32' },
    { name: 'nonce', type: 'uint256' },
    { name: 'deadline', type: 'uint256' },
  ],
} as const;

export interface EmergencyHaltState {
  isMarketHalted: boolean;
  haltedAt?: string;
  haltedBy?: Address;
  reason?: string;
  canResume: boolean;
}
```

## Security & Compliance Notes
- **Zero Key Custody on Web Client:** All administrative signatures are produced exclusively on external hardware wallets; private keys never touch the web browser DOM or memory.
- **EIP-712 Anti-Phishing:** Enforces strict domain separator validation (name, version, chainId, verifyingContract) to prevent signature replay attacks across chains or environments.
- **Threshold Guarantees:** Smart contracts enforce minimum $M$-of-$N$ signatures on-chain; the web UI cannot bypass on-chain threshold verification.
- **Dual-Control Circuit Breakers:** Emergency trading halts require instant logging to the immutable audit trail (Prompt 218) and alert dispatch to all compliance officers.

## Acceptance Criteria
- [ ] Admin approvals portal initializes and securely connects to hardware wallets (Ledger/Trezor) via Web3Modal/Wagmi.
- [ ] Non-authorized wallet addresses are immediately blocked from viewing proposal actions.
- [ ] Proposal creation wizard generates valid ABI calldata and displays accurate human-readable parameter previews.
- [ ] EIP-712 typed data signature triggers properly on connected hardware wallet and collects signature.
- [ ] Quorum progress bar dynamically updates in real time as co-signers submit approvals.
- [ ] "Execute Transaction" button is disabled until minimum quorum threshold (e.g. 2 of 3) is verified on-chain.
- [ ] Emergency trading halt trigger sends broadcast event and halts market terminal feeds within 2 seconds.
- [ ] Playwright E2E tests verify proposal creation, multi-party signing, and execution on simulated Besu test node.

## Suggested Order / Dependencies
- **Prerequisites:** 217 (Admin Service), 218 (Audit Log Service), 307 (MultiSig Smart Contract), 601 (Web Scaffolding), 702 (IAM).
- **Direct Successors / Parallel:** 604 (KYC Review Dashboard), 606 (Proof of Reserve Dashboard), 607 (Regulatory Reporting).
