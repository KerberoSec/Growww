"""
Trade Settlement Service – NBSE Sovereign Exchange
====================================================
Production-grade settlement engine with:
  • T+0 / T+1 settlement cycles
  • Bilateral netting (net obligations between counterparties)
  • Clearing obligation tracking
  • Settlement confirmation with cryptographic fingerprints
  • Batch settlement processing
"""

from __future__ import annotations

import hashlib
import time
import uuid
from collections import defaultdict
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class SettlementCycle(Enum):
    T_PLUS_0 = "T+0"  # Same-day settlement (crypto)
    T_PLUS_1 = "T+1"  # Next business day (equities)


class SettlementStatus(Enum):
    PENDING = "PENDING"
    NETTED = "NETTED"
    CLEARING = "CLEARING"
    SETTLED = "SETTLED"
    FAILED = "FAILED"


class ObligationType(Enum):
    DELIVER_ASSET = "DELIVER_ASSET"
    DELIVER_CASH = "DELIVER_CASH"


@dataclass
class TradeExecution:
    """A matched trade awaiting settlement."""
    execution_id: str
    symbol: str
    buyer_id: str
    seller_id: str
    quantity_e8: int              # asset quantity (8-decimal)
    price_e8: int                 # price per unit (8-decimal)
    trade_timestamp: float
    settlement_cycle: SettlementCycle
    settlement_date: float        # epoch timestamp of settlement due date


@dataclass
class NettingResult:
    """Result of bilateral netting between two parties."""
    netting_id: str
    party_a: str
    party_b: str
    symbol: str
    net_quantity_e8: int          # positive = A delivers to B, negative = B delivers to A
    net_cash_e8: int              # positive = A pays B, negative = B pays A
    trade_count: int
    gross_quantity_e8: int
    netting_efficiency_pct: float # % reduction from netting


@dataclass
class ClearingObligation:
    """A clearing obligation resulting from netting."""
    obligation_id: str
    obligor_id: str               # party who must deliver
    beneficiary_id: str           # party who receives
    obligation_type: ObligationType
    symbol: str
    amount_e8: int
    due_date: float
    status: SettlementStatus = SettlementStatus.PENDING
    fulfilled_at: Optional[float] = None


@dataclass
class SettlementConfirmation:
    """Final settlement confirmation record."""
    confirmation_id: str
    execution_ids: List[str]
    netting_id: Optional[str]
    settlement_hash: str          # SHA-256 fingerprint
    obligations_fulfilled: int
    settled_at: float
    settlement_cycle: SettlementCycle


# ---------------------------------------------------------------------------
# Engine
# ---------------------------------------------------------------------------

class SettlementEngine:
    """
    Multi-cycle settlement engine with netting and clearing.

    Processes trade executions through:
    1. Ingestion → 2. Netting → 3. Clearing obligations → 4. Settlement confirmation
    """

    def __init__(self):
        self.pending_trades: List[TradeExecution] = []
        self.netting_results: List[NettingResult] = []
        self.obligations: List[ClearingObligation] = []
        self.confirmations: List[SettlementConfirmation] = []

    # ---- Trade Ingestion ---------------------------------------------------

    def ingest_trade(
        self,
        symbol: str,
        buyer_id: str,
        seller_id: str,
        quantity_e8: int,
        price_e8: int,
        settlement_cycle: SettlementCycle = SettlementCycle.T_PLUS_0,
    ) -> TradeExecution:
        """Ingest a matched trade for settlement processing."""
        if quantity_e8 <= 0 or price_e8 <= 0:
            raise ValueError("quantity and price must be positive")
        if buyer_id == seller_id:
            raise ValueError("buyer and seller cannot be the same")

        now = time.time()
        if settlement_cycle == SettlementCycle.T_PLUS_0:
            settlement_date = now  # immediate
        else:
            settlement_date = now + 86400  # T+1 = 24 hours

        trade = TradeExecution(
            execution_id=str(uuid.uuid4()),
            symbol=symbol,
            buyer_id=buyer_id,
            seller_id=seller_id,
            quantity_e8=quantity_e8,
            price_e8=price_e8,
            trade_timestamp=now,
            settlement_cycle=settlement_cycle,
            settlement_date=settlement_date,
        )
        self.pending_trades.append(trade)
        return trade

    # ---- Bilateral Netting -------------------------------------------------

    def compute_bilateral_netting(self, symbol: Optional[str] = None) -> List[NettingResult]:
        """
        Compute bilateral netting across all pending trades.

        Groups trades by (party_pair, symbol) and nets quantities and cash.
        Returns list of netting results.
        """
        # Group by (sorted party pair, symbol)
        groups: Dict[Tuple[str, str, str], List[TradeExecution]] = defaultdict(list)
        for trade in self.pending_trades:
            if symbol and trade.symbol != symbol:
                continue
            pair = tuple(sorted([trade.buyer_id, trade.seller_id]))
            key = (pair[0], pair[1], trade.symbol)
            groups[key].append(trade)

        results: List[NettingResult] = []
        for (party_a, party_b, sym), trades in groups.items():
            net_qty = 0    # positive = A delivers to B
            net_cash = 0   # positive = A pays B
            gross_qty = 0

            for t in trades:
                cash = t.quantity_e8 * t.price_e8 // 100_000_000  # cash in e8
                if t.seller_id == party_a:
                    # A sells to B: A delivers asset, B pays cash
                    net_qty -= t.quantity_e8
                    net_cash += cash
                else:
                    # B sells to A: B delivers asset, A pays cash
                    net_qty += t.quantity_e8
                    net_cash -= cash
                gross_qty += t.quantity_e8

            net_qty_abs = abs(net_qty)
            efficiency = (1.0 - net_qty_abs / gross_qty) * 100.0 if gross_qty > 0 else 0.0

            netting = NettingResult(
                netting_id=str(uuid.uuid4()),
                party_a=party_a,
                party_b=party_b,
                symbol=sym,
                net_quantity_e8=net_qty,
                net_cash_e8=net_cash,
                trade_count=len(trades),
                gross_quantity_e8=gross_qty,
                netting_efficiency_pct=efficiency,
            )
            results.append(netting)

        self.netting_results.extend(results)
        return results

    # ---- Clearing Obligations ----------------------------------------------

    def generate_clearing_obligations(
        self,
        netting_results: Optional[List[NettingResult]] = None,
    ) -> List[ClearingObligation]:
        """
        Generate clearing obligations from netting results.

        Each netting result produces up to two obligations:
        1. Asset delivery obligation
        2. Cash delivery obligation
        """
        if netting_results is None:
            netting_results = self.netting_results

        obligations: List[ClearingObligation] = []
        now = time.time()

        for nr in netting_results:
            # Asset obligation
            if nr.net_quantity_e8 != 0:
                if nr.net_quantity_e8 > 0:
                    obligor, beneficiary = nr.party_a, nr.party_b
                    amount = nr.net_quantity_e8
                else:
                    obligor, beneficiary = nr.party_b, nr.party_a
                    amount = abs(nr.net_quantity_e8)

                obligations.append(ClearingObligation(
                    obligation_id=str(uuid.uuid4()),
                    obligor_id=obligor,
                    beneficiary_id=beneficiary,
                    obligation_type=ObligationType.DELIVER_ASSET,
                    symbol=nr.symbol,
                    amount_e8=amount,
                    due_date=now + 86400,
                ))

            # Cash obligation
            if nr.net_cash_e8 != 0:
                if nr.net_cash_e8 > 0:
                    obligor, beneficiary = nr.party_b, nr.party_a
                    amount = nr.net_cash_e8
                else:
                    obligor, beneficiary = nr.party_a, nr.party_b
                    amount = abs(nr.net_cash_e8)

                obligations.append(ClearingObligation(
                    obligation_id=str(uuid.uuid4()),
                    obligor_id=obligor,
                    beneficiary_id=beneficiary,
                    obligation_type=ObligationType.DELIVER_CASH,
                    symbol=nr.symbol,
                    amount_e8=amount,
                    due_date=now + 86400,
                ))

        self.obligations.extend(obligations)
        return obligations

    # ---- Settlement Confirmation -------------------------------------------

    def settle_obligations(
        self,
        obligation_ids: Optional[List[str]] = None,
    ) -> SettlementConfirmation:
        """
        Settle a batch of clearing obligations and produce a confirmation.

        If obligation_ids is None, settles all pending obligations.
        """
        now = time.time()
        target_obligations = [
            o for o in self.obligations
            if o.status == SettlementStatus.PENDING
            and (obligation_ids is None or o.obligation_id in obligation_ids)
        ]

        if not target_obligations:
            raise ValueError("no pending obligations to settle")

        # Mark as settled
        for o in target_obligations:
            o.status = SettlementStatus.SETTLED
            o.fulfilled_at = now

        # Compute settlement hash
        hash_input = ":".join(
            f"{o.obligation_id}:{o.obligor_id}:{o.beneficiary_id}:{o.amount_e8}"
            for o in sorted(target_obligations, key=lambda x: x.obligation_id)
        )
        settlement_hash = hashlib.sha256(hash_input.encode()).hexdigest()

        # Determine execution IDs involved
        execution_ids = [t.execution_id for t in self.pending_trades]

        # Determine netting IDs
        netting_id = self.netting_results[-1].netting_id if self.netting_results else None

        confirmation = SettlementConfirmation(
            confirmation_id=str(uuid.uuid4()),
            execution_ids=execution_ids,
            netting_id=netting_id,
            settlement_hash=settlement_hash,
            obligations_fulfilled=len(target_obligations),
            settled_at=now,
            settlement_cycle=SettlementCycle.T_PLUS_0,
        )
        self.confirmations.append(confirmation)

        # Clear settled trades
        self.pending_trades.clear()

        return confirmation

    # ---- Queries -----------------------------------------------------------

    def get_pending_count(self) -> int:
        """Return count of pending unsettled trades."""
        return len(self.pending_trades)

    def get_obligation_summary(self) -> Dict[str, int]:
        """Return count of obligations by status."""
        summary: Dict[str, int] = defaultdict(int)
        for o in self.obligations:
            summary[o.status.value] += 1
        return dict(summary)
