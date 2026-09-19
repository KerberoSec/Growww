# 318 - Shareholder Proxy Voting & Corporate Governance Smart Contract Suite (ERC-1155 / ERC-5805)

## Purpose
In traditional equity markets, retail and institutional investors often face friction when participating in corporate governance: postal ballots are slow, proxy advisory representation is opaque, and fractional share owners are traditionally disenfranchised from exercising voting rights at Annual General Meetings (AGM) and Extra-Ordinary General Meetings (EGM). Under the Indian Companies Act 2013 (Section 108) and SEBI (Listing Obligations and Disclosure Requirements) Regulations (Clause 44), listed entities must provide mandatory electronic voting (e-voting) facilities to all shareholders.

This prompt specifies the design, implementation, formal verification, and deployment of the **On-Chain Shareholder Proxy Voting & Corporate Governance Smart Contract Suite** (`contracts/governance/` and `services/voting-aggregator/`). Utilizing ERC-1155 multi-token governance checkpoints, ERC-5805 / ERC-20Votes balance snapshots, and EIP-712 cryptographic signature aggregations, the system enables fractional and whole-share equity token holders to cast immutable, verifiable, and privacy-preserving proxy votes on corporate resolutions (Ordinary, Special, and Board Elections) with mathematical quorum enforcement and instant result certification.

## What You Are Building
A modular, high-security on-chain shareholder governance infrastructure comprising:
- `ShareholderGovernor.sol`: Core governance orchestrator managing corporate resolution lifecycles (Proposal Creation $\rightarrow$ Voting Period $\rightarrow$ Quorum Evaluation $\rightarrow$ Execution / Certification).
- `ShareholderVotingToken.sol (ERC-1155 / ERC-5805)`: Multi-asset voting token tracking historical checkpointed share balances at specified Record Dates across all listed ISINs (`GROWWW-TCS`, `GROWWW-INFY`, etc.).
- `SecretBallotCommitReveal.sol`: Privacy-preserving voting module allowing shareholders to submit hidden cryptographic commitments (`keccak256(vote, salt)`) during the voting phase and reveal them during the tally phase to eliminate bandwagoning.
- `Gasless Voting Relayer & EIP-712 Aggregator (Go / TypeScript)`: Microservice allowing retail mobile/web users to sign free off-chain EIP-712 vote authorizations, which are batched and posted on-chain by Growww relayers.
- `Corporate Action Resolution Bridge`: Ingests official AGM/EGM resolution notices from BSE/NSE corporate filing feeds and instantiates on-chain ballot proposals with verified resolution metadata.

## Scope Boundaries
- **In Scope:**
 - On-chain proposal creation, quorum threshold enforcement, and multi-choice resolution voting.
 - ERC-1155 / ERC-5805 voting power checkpoints indexed by corporate Record Date block numbers.
 - Proxy voting delegation (allowing shareholders to delegate voting weight to registered Proxy Advisors).
 - Fractional vote tallying down to 18 decimal places of equity ownership.
 - EIP-712 meta-transactions for zero-gas voter participation.
 - Secret ballot commit-reveal mechanisms for sensitive corporate proxy battles.
- **Out of Scope / Handled Elsewhere:**
 - Corporate actions dividend and stock split adjustments (handled in Prompt 222).
 - General multisig administrative control over platform parameters (handled in Prompt 307).
 - Investor identity and KYC tier validation (handled in Prompt 305).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Cancun / Prague).
  *Justification:* Provides native support for custom errors, transient storage for cheap reentrancy locks during tally loops, and deterministic arithmetic for fractional voting power.
- **Contract Frameworks:** OpenZeppelin Contracts v5.0 (Governor, GovernorVotes, GovernorCountingSimple, ERC1155Votes, TimelockController).
- **Standards:** ERC-1155 (Multi-Token), ERC-5805 (Voting with Checkpoints), EIP-712 (Structured Data Signing), EIP-2612 / EIP-1271.
- **Relayer Language:** **Go (v1.22+)** with `go-ethereum` and `gin-gonic`.
- **Testing & Verification:** Foundry (`forge`), Certora Prover / Halmos for formal verification.

## Backend / Infra Touchpoints
- **Corporate Actions Service (Prompt 222):** Notifies governance engine of newly announced AGMs, EGMs, record dates, and resolution notices.
- **Event Indexer (Prompt 309):** Ingests `VoteCast`, `ProposalCreated`, and `ProposalExecuted` events into PostgreSQL for real-time mobile app dashboards.
- **Flutter Multi-Platform Client (Prompt 506 / 508):** Displays interactive corporate ballot cards and collects biometric-authenticated EIP-712 vote signatures.
- **Kafka Cluster (Prompt 403):** Streams aggregated vote events to compliance analytics pipelines.

## Blockchain Interaction
- **Record Date Snapshotting:** When a company announces a corporate meeting with a specified Record Date ($T_{\text{record}}$), the contract records the exact L2 block number corresponding to the cutoff. Only token balances held at that exact block can vote.
- **Fractional Weight Calculation:** If an investor owns $12.4567$ tokens of `GROWWW-RELIANCE`, their vote carries precisely $12,456,700,000,000,000,000$ units of voting weight.
- **Resolution Categories:**
 - *Ordinary Resolution:* Requires simple majority ($> 50\%$) of votes cast.
 - *Special Resolution:* Requires super-majority ($\ge 75\%$) of votes cast (as per Companies Act 2013).
 - *Board Election:* Multi-candidate cumulative or preferential voting tallying.
- **Zero PII Exposure:** On-chain ballots record only investor wallet addresses and voting weights; the official Scrutinizer report maps addresses to demat accounts off-chain via secure compliance keys.

## Step-by-Step Build Instructions
1. Scaffold governance repository under `contracts/governance/`:
 - `src/ShareholderGovernor.sol`, `src/ShareholderVotingToken.sol`, `src/CommitRevealVoting.sol`, `src/interfaces/`.
 - `test/governance/`, `script/governance/`.
2. Implement `ShareholderVotingToken.sol` inheriting `ERC1155Upgradeable` and `ERC1155VotesUpgradeable`:
 - Support multiple equity ISINs within a single contract using `uint256 isinId = uint256(keccak256(abi.encodePacked(isin)))`.
 - Implement checkpointed balance tracking: `getPastVotes(address account, uint256 tokenId, uint256 timepoint)`.
 - Implement delegation: `delegate(uint256 tokenId, address delegatee)`.
3. Implement `ShareholderGovernor.sol`:
 - Inherit from `GovernorUpgradeable`, `GovernorCountingFractionalUpgradeable`, `GovernorVotesUpgradeable`.
 - Define custom proposal parameters: `resolutionType` (Ordinary / Special), `isinId`, `recordDateBlock`, `votingStartBlock`, `votingEndBlock`.
 - Implement `proposeResolution(...)` restricted to verified Corporate Secretary or authorized Admin multi-sig.
4. Implement Fractional Vote Counting Engine:
 - Allow splitting voting weight across `For`, `Against`, and `Abstain` options.
 - Enforce invariant: $\text{Weight}_{\text{For}} + \text{Weight}_{\text{Against}} + \text{Weight}_{\text{Abstain}} \le \text{PastVotes}(\text{Voter}, \text{RecordBlock})$.
5. Implement `CommitRevealVoting.sol`:
 - `commitVote(uint256 proposalId, bytes32 voteCommitment)` during active voting phase.
 - `revealVote(uint256 proposalId, uint8 support, uint256 weight, bytes32 salt)` during 24-hour reveal phase.
 - Validate `keccak256(abi.encodePacked(msg.sender, support, weight, salt)) == voteCommitment`.
6. Implement Gasless EIP-712 Vote Submission:
 - Implement `castVoteBySig(uint256 proposalId, uint8 support, uint8 v, bytes32 r, bytes32 s)`.
 - Implement `castVoteWithReasonAndParamsBySig(...)` for proxy advisors providing institutional rationale.
7. Build Go Gasless Relayer Service (`services/voting-aggregator/`):
 - Expose REST/gRPC endpoints `/api/v1/governance/vote/submit`.
 - Ingest investor EIP-712 signatures, validate signature format and current nonce off-chain.
 - Batch up to 200 signatures into a single `castVoteBatch` transaction to minimize L2 gas costs.
8. Implement Scrutinizer Export & Audit Certification Engine:
 - Generates cryptographically verifiable voting certificates summarizing total votes cast, quorum percentage, and resolution outcome.
 - Emits `ResolutionCertified(proposalId, isin, outcome, totalVotesFor, totalVotesAgainst)`.
9. Write comprehensive Foundry unit tests in `test/governance/ShareholderGovernor.t.sol`:
 - Test vote weighting across fractional balances ($0.001$ to $1,000,000$ shares).
 - Test rejection of votes cast by tokens acquired after the Record Date block.
 - Test Special Resolution 75% threshold calculations with rounding edge cases.
 - Test Commit-Reveal secrecy and expired reveal handling.
10. Write Certora Formal Verification rules asserting that:
 - No account can double-vote or exceed their past checkpointed balance.
 - Quorum calculation cannot overflow under maximum possible token supplies.
11. Run Slither static analyzer and resolve any state variable shadowing or external call risks.
12. Deploy to local staging testnet, simulate AGM of 5 listed companies, and verify automated gasless vote submission for 5,000 simulated shareholders.

## Interfaces / Contracts

### Shareholder Governor Contract Interface (`IShareholderGovernor.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IShareholderGovernor {
    enum ResolutionType {
        Ordinary, // 50% majority required
        Special,  // 75% majority required
        Election  // Board of Directors multi-choice
    }

    enum ProposalState {
        Pending,
        Active,
        CommitPhase,
        RevealPhase,
        Canceled,
        Defeated,
        Succeeded,
        Certified
    }

    struct ProposalDetails {
        uint256 proposalId;
        string isin;
        uint256 isinId;
        ResolutionType resolutionType;
        uint256 recordDateBlock;
        uint256 startTime;
        uint256 endTime;
        uint256 forVotes;
        uint256 againstVotes;
        uint256 abstainVotes;
        uint256 totalEligibleWeight;
        bool isSecretBallot;
        bool certified;
    }

    event ProposalCreated(
        uint256 indexed proposalId,
        string indexed isin,
        ResolutionType resolutionType,
        uint256 recordDateBlock,
        string descriptionUrl
    );

    event VoteCastWithFraction(
        address indexed voter,
        uint256 indexed proposalId,
        uint8 support,
        uint256 weight,
        string reason
    );

    event ResolutionCertified(
        uint256 indexed proposalId,
        string isin,
        bool passed,
        uint256 forVotes,
        uint256 againstVotes,
        uint256 totalVotesCast
    );

    error RecordDateInFuture(uint256 providedBlock, uint256 currentBlock);
    error InsufficientVotingPower(address voter, uint256 available, uint256 requested);
    error ProposalNotActive(uint256 proposalId, ProposalState currentState);
    error InvalidRevealHash(bytes32 expectedHash, bytes32 computedHash);
    error VotingAlreadyEnded(uint256 proposalId);

    function createProposal(
        string calldata isin,
        ResolutionType resolutionType,
        uint256 recordDateBlock,
        uint256 durationSeconds,
        bool isSecretBallot,
        string calldata descriptionUrl
    ) external returns (uint256 proposalId);

    function castVote(uint256 proposalId, uint8 support) external returns (uint256 weight);

    function castVoteBySig(
        uint256 proposalId,
        uint8 support,
        uint256 nonce,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external returns (uint256 weight);

    function state(uint256 proposalId) external view returns (ProposalState);
    function getProposal(uint256 proposalId) external view returns (ProposalDetails memory);
}
```

### EIP-712 Governance Vote Structured Typed Data
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

struct BallotVote {
    uint256 proposalId;
    uint8 support;
    address voter;
    uint256 nonce;
    uint256 deadline;
}

bytes32 constant BALLOT_VOTE_TYPEHASH = keccak256(
    "BallotVote(uint256 proposalId,uint8 support,address voter,uint256 nonce,uint256 deadline)"
);
```

## Security & Compliance Notes
- **Indian Companies Act & SEBI Compliance:** Complies with Section 108 of the Companies Act 2013, Rule 20 of Companies (Management and Administration) Rules 2014, and SEBI LODR Clause 44 for electronic shareholder voting and scrutinizer auditing.
- **Flash Loan & Post-Announcement Sybil Attack Defense:** Voting power is strictly bound to the historical block number of the corporate Record Date. Acquiring tokens after the Record Date grants zero voting rights for the corresponding resolution.
- **Fractional Invariant Preservation:** The fractional voting engine uses 18-decimal fixed-point math with checked arithmetic, preventing rounding vulnerabilities from altering resolution outcomes.
- **Bandwagoning Elimination:** Secret Ballot commit-reveal eliminates early tally disclosure, ensuring institutional proxy advisors and majority stakeholders cannot manipulate retail sentiment mid-election.

## Acceptance Criteria
- [ ] `ShareholderGovernor.sol` and `ShareholderVotingToken.sol` implemented adhering to ERC-1155 and ERC-5805 standards.
- [ ] 100% test coverage across Ordinary (50%) and Special (75%) resolution threshold scenarios.
- [ ] Voting power correctly locked to historical Record Date block; tokens transferred after Record Date cannot vote.
- [ ] Gasless EIP-712 vote submission successfully batches and processes 200+ signatures in a single transaction.
- [ ] Secret ballot commit-reveal successfully verifies valid secrets and rejects invalid revelations.
- [ ] Formal verification with Certora / Halmos mathematically confirms zero double-voting and overflow immunity.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `303` (Token Issuance Smart Contract), Prompt `305` (Transfer Compliance Hooks), Prompt `222` (Corporate Actions Service).
- **Parallel Tasks:** Prompt `316` (Appchain Rollup Sequencer), Prompt `509` (Flutter Order & Voting Flow).
- **Subsequent Prompts Enabled:** Prompt `319` (Institutional Custody Bridge), Prompt `607` (Admin Regulatory Reporting Dashboard).
