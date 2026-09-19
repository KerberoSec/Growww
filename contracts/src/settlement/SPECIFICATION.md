# Settlement Contracts Architecture, DvP & Dynamic Fee Controller Specification

## 1. Executive Summary & Starting Invariants
The settlement smart contract suite on Hyperledger Besu executes atomic Delivery-versus-Payment (DvP) clearing for spot crypto trades (BTC/USDT). The protocol enforces a **Zero-Fee Initial Launch with Future Parameter Governance**:
- **0.00% Starting Fee Invariant ("0 Means 0 in All")**:
  - `makerFeeBps = 0` (0.00% Maker Fee at launch).
  - `takerFeeBps = 0` (0.00% Taker Fee at launch).
  - Zero gas fees for users via ERC-4337 Paymasters.
  - Zero on-chain TDS (0% tax withholding).
  - 100% of gross trade proceeds delivered to the seller at launch.
- **Dynamic Future Fee Expandability**:
  - Governed by `FeeController.sol`, allowing the exchange governance to increase or adjust fees in the future if required.
  - Protected by a **48-Hour Timelock** and **3-of-5 Multi-Signature Governance**.
  - Enforces a hard-coded immutable ceiling: `MAX_FEE_CEILING = 50 bps` (0.50%), preventing arbitrary or unbounded fee spikes.

## 2. Smart Contract Inventory
1. `DvPAtomicSettlement.sol`: Core contract executing atomic swaps for single trades and aggregated netted batches.
2. `FeeController.sol`: Dynamic governance contract managing active `makerFeeBps` and `takerFeeBps` with timelock protection.
3. `OrderRegistry.sol`: On-chain registry tracking order commitments, cancellations, and fill nonces.
4. `SettlementGuaranteeFund.sol`: Collateral pool providing liquidity buffer to absorb settlement defaults.
5. `PaymasterRelayer.sol`: ERC-4337 gas sponsorship contract enabling zero-gas trading for retail users.

## 3. Dynamic Fee Engine & Calculation Logic
```solidity
// In FeeController.sol
contract FeeController is Initializable, AccessControlUpgradeable, UUPSUpgradeable {
    bytes32 public constant GOVERNANCE_ROLE = keccak256("GOVERNANCE_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    uint16 public makerFeeBps; // Initialized to 0 (0.00%)
    uint16 public takerFeeBps; // Initialized to 0 (0.00%)
    uint16 public constant MAX_FEE_CEILING = 50; // Hard ceiling 0.50% (50 bps)
    
    uint256 public constant TIMELOCK_DELAY = 48 hours;
    
    struct PendingFeeChange {
        uint16 newMakerBps;
        uint16 newTakerBps;
        uint256 effectiveTimestamp;
    }
    PendingFeeChange public pendingChange;

    event FeeChangeProposed(uint16 newMakerBps, uint16 newTakerBps, uint256 effectiveTimestamp);
    event FeeScheduleUpdated(uint16 newMakerBps, uint16 newTakerBps);
    event FeeChangeCancelled(uint16 discardedMakerBps, uint16 discardedTakerBps);

    function initialize(address multisigAdmin) external initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        _grantRole(DEFAULT_ADMIN_ROLE, multisigAdmin);
        _grantRole(GOVERNANCE_ROLE, multisigAdmin);
        _grantRole(UPGRADER_ROLE, multisigAdmin);
        makerFeeBps = 0;
        takerFeeBps = 0;
    }

    function scheduleFeeUpdate(uint16 _makerBps, uint16 _takerBps) external onlyRole(GOVERNANCE_ROLE) {
        require(_makerBps <= MAX_FEE_CEILING && _takerBps <= MAX_FEE_CEILING, "EXCEEDS_CEILING");
        pendingChange = PendingFeeChange({
            newMakerBps: _makerBps,
            newTakerBps: _takerBps,
            effectiveTimestamp: block.timestamp + TIMELOCK_DELAY
        });
        emit FeeChangeProposed(_makerBps, _takerBps, pendingChange.effectiveTimestamp);
    }

    function executeFeeUpdate() external onlyRole(GOVERNANCE_ROLE) {
        require(pendingChange.effectiveTimestamp != 0, "NO_PENDING_CHANGE");
        require(block.timestamp >= pendingChange.effectiveTimestamp, "TIMELOCK_ACTIVE");
        makerFeeBps = pendingChange.newMakerBps;
        takerFeeBps = pendingChange.newTakerBps;
        delete pendingChange; // Safely resets effectiveTimestamp to 0
        emit FeeScheduleUpdated(makerFeeBps, takerFeeBps);
    }

    function cancelFeeUpdate() external onlyRole(GOVERNANCE_ROLE) {
        require(pendingChange.effectiveTimestamp != 0, "NO_PENDING_CHANGE");
        uint16 discardedMaker = pendingChange.newMakerBps;
        uint16 discardedTaker = pendingChange.newTakerBps;
        delete pendingChange;
        emit FeeChangeCancelled(discardedMaker, discardedTaker);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[47] private __gap;
}
```

### Mathematical Formulation:
- **At Launch (`makerFeeBps == 0`, `takerFeeBps == 0`)**:
  $$\text{MakerFee} = \frac{\text{GrossQuote} \times 0}{10000} = 0$$
  $$\text{TakerFee} = \frac{\text{GrossQuote} \times 0}{10000} = 0$$
  $$\text{NetSellerProceeds} = \text{GrossQuote} - \text{TakerFee} = \text{GrossQuote} \quad (100\% \text{ Net Proceeds})$$
- **In Future (If Fees are Activated)**:
  $$\text{MakerFee} = \left\lfloor \frac{\text{GrossQuote} \times \text{makerFeeBps}}{10000} \right\rfloor$$
  $$\text{TakerFee} = \left\lfloor \frac{\text{GrossQuote} \times \text{takerFeeBps}}{10000} \right\rfloor$$
  $$\text{NetSellerProceeds} = \text{GrossQuote} - \text{TakerFee}$$

## 4. EIP-712 Typed Structured Data Verification
Every settlement execution requires EIP-712 typed order authorization:
```solidity
struct OrderCommitment {
    bytes32 orderId;
    address trader;
    address baseToken;
    address quoteToken;
    uint8 side; // 0 = Buy, 1 = Sell
    uint256 price; // 6 decimals (USDT)
    uint256 quantity; // 8 decimals (Satoshis)
    uint256 nonce;
    uint256 expiry;
}
```

## 5. Security & Upgrade Invariants
- **ReentrancyGuard**: OpenZeppelin `ReentrancyGuardUpgradeable` on all entrypoints.
- **Checks-Effects-Interactions (CEI)**: State variables mutated prior to token transfers.
- **UUPS Storage Safety**: Explicit `uint256[50] private __gap;` at contract bottom.
