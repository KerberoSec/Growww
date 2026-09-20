"""
Growww / NBSE DvP (Delivery versus Payment) Atomic Settlement Fuzz Coordinator
Validates two-phase atomic settlement, rollback compensation, relayer drops,
and double-entry balance conservation laws under concurrent stress.
"""

import asyncio
import random
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Set, Tuple


class SettlementState(Enum):
    PENDING = "PENDING"
    CASH_HELD = "CASH_HELD"
    CHAIN_SUBMITTED = "CHAIN_SUBMITTED"
    SETTLED = "SETTLED"
    ROLLED_BACK = "ROLLED_BACK"


@dataclass(frozen=True)
class FuzzMatchRecord:
    match_id: str
    symbol: str
    buyer_account_id: str
    seller_account_id: str
    quantity: int
    price_paise: int
    trade_timestamp_ns: int
    fee_paise: int = 0


@dataclass(frozen=True)
class SettlementVerificationResult:
    is_solvent: bool
    cash_delta_sum: int
    token_delta_sum: int
    unresolved_discrepancies: List[str]


@dataclass
class AccountLedger:
    account_id: str
    cash_available_paise: int
    cash_held_paise: int = 0
    tokens_available: Dict[str, int] = field(default_factory=dict)
    tokens_held: Dict[str, int] = field(default_factory=dict)


class DvPSettlementFuzzCoordinator:
    """Async coordinator simulating multi-threaded DvP settlement pipeline with chaos injection."""

    def __init__(self, exchange_fee_bps: int = 5):
        self.exchange_fee_bps = exchange_fee_bps # 0.05%
        self.ledgers: Dict[str, AccountLedger] = {}
        self.exchange_fee_cash_paise = 0
        self.processed_nonces: Set[str] = set()
        self.lock = asyncio.Lock()

    def register_account(self, account_id: str, initial_cash_paise: int, initial_tokens: Dict[str, int]) -> None:
        self.ledgers[account_id] = AccountLedger(
            account_id=account_id,
            cash_available_paise=initial_cash_paise,
            tokens_available=dict(initial_tokens),
        )

    async def execute_settlement(
        self,
        match: FuzzMatchRecord,
        chaos_drop_rpc: bool = False,
        chaos_nonce_race: bool = False,
    ) -> SettlementState:
        """Executes atomic 2-phase settlement: Phase 1 = Hold; Phase 2 = Chain Commit or Rollback."""
        total_cash_required = match.quantity * match.price_paise
        fee_paise = (total_cash_required * self.exchange_fee_bps) // 10000

        async with self.lock:
            buyer = self.ledgers.get(match.buyer_account_id)
            seller = self.ledgers.get(match.seller_account_id)

            if not buyer or not seller:
                return SettlementState.ROLLED_BACK

            # Phase 1: Pre-Execution Hold & Balance Verification
            if buyer.cash_available_paise < (total_cash_required + fee_paise):
                return SettlementState.ROLLED_BACK # Insufficient cash margin

            seller_token_bal = seller.tokens_available.get(match.symbol, 0)
            if seller_token_bal < match.quantity:
                return SettlementState.ROLLED_BACK # Insufficient security tokens

            # Place Holds
            buyer.cash_available_paise -= (total_cash_required + fee_paise)
            buyer.cash_held_paise += (total_cash_required + fee_paise)

            seller.tokens_available[match.symbol] -= match.quantity
            seller.tokens_held[match.symbol] = seller.tokens_held.get(match.symbol, 0) + match.quantity

        # Simulated Asynchronous Blockchain Submission Delay
        await asyncio.sleep(0.001)

        # Chaos Injection: Simulated Relayer Drop / Network Partition / Nonce Race
        if chaos_drop_rpc or (chaos_nonce_race and match.match_id in self.processed_nonces):
            # Phase 2b: Rollback compensation
            async with self.lock:
                buyer.cash_held_paise -= (total_cash_required + fee_paise)
                buyer.cash_available_paise += (total_cash_required + fee_paise)

                seller.tokens_held[match.symbol] -= match.quantity
                seller.tokens_available[match.symbol] += match.quantity
            return SettlementState.ROLLED_BACK

        # Phase 2a: Commit Transfer Atomically
        async with self.lock:
            self.processed_nonces.add(match.match_id)

            # Release buyer hold and debit cash permanently
            buyer.cash_held_paise -= (total_cash_required + fee_paise)
            # Release seller token hold and credit tokens to buyer
            seller.tokens_held[match.symbol] -= match.quantity
            buyer.tokens_available[match.symbol] = buyer.tokens_available.get(match.symbol, 0) + match.quantity

            # Credit cash to seller
            seller.cash_available_paise += total_cash_required
            # Credit exchange fee pool
            self.exchange_fee_cash_paise += fee_paise

        return SettlementState.SETTLED

    async def verify_dvp_solvency_invariants(
        self,
        initial_cash: Dict[str, int],
        initial_tokens: Dict[str, Dict[str, int]],
        symbols: List[str],
    ) -> SettlementVerificationResult:
        """Verifies conservation laws:
        Sum(Final Cash) + Fee Pool == Sum(Initial Cash)
        Sum(Final Tokens per symbol) == Sum(Initial Tokens per symbol)
        Zero orphaned holds.
        """
        discrepancies: List[str] = []

        total_init_cash = sum(initial_cash.values())
        total_curr_cash = self.exchange_fee_cash_paise

        for acc_id, ledger in self.ledgers.items():
            if ledger.cash_held_paise != 0:
                discrepancies.append(f"Orphaned cash hold on {acc_id}: {ledger.cash_held_paise} paise")
            total_curr_cash += ledger.cash_available_paise + ledger.cash_held_paise

        cash_delta = total_curr_cash - total_init_cash
        if cash_delta != 0:
            discrepancies.append(f"Cash conservation broken! Delta: {cash_delta} paise")

        token_delta_sum = 0
        for sym in symbols:
            total_init_tokens = sum(initial_tokens.get(acc, {}).get(sym, 0) for acc in initial_tokens)
            total_curr_tokens = 0
            for acc_id, ledger in self.ledgers.items():
                if ledger.tokens_held.get(sym, 0) != 0:
                    discrepancies.append(f"Orphaned token hold on {acc_id} for {sym}: {ledger.tokens_held[sym]}")
                total_curr_tokens += ledger.tokens_available.get(sym, 0) + ledger.tokens_held.get(sym, 0)

            sym_delta = total_curr_tokens - total_init_tokens
            if sym_delta != 0:
                discrepancies.append(f"Token conservation broken for {sym}! Delta: {sym_delta}")
            token_delta_sum += abs(sym_delta)

        is_solvent = len(discrepancies) == 0
        return SettlementVerificationResult(
            is_solvent=is_solvent,
            cash_delta_sum=cash_delta,
            token_delta_sum=token_delta_sum,
            unresolved_discrepancies=discrepancies,
        )
