"""
Perpetuals Engine – NBSE Sovereign Exchange
============================================
Production-grade USD-M perpetual futures engine with:
  • Funding rate calculation (premium index + interest rate)
  • Mark price computation (EMA-smoothed index price)
  • Liquidation engine with bankruptcy price detection
  • Auto-Deleveraging (ADL) priority queue
"""

from __future__ import annotations

import time
import math
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class PositionSide(Enum):
    LONG = "LONG"
    SHORT = "SHORT"


class PositionStatus(Enum):
    ACTIVE = "ACTIVE"
    LIQUIDATED = "LIQUIDATED"
    ADL_CLOSED = "ADL_CLOSED"


@dataclass
class PerpetualPosition:
    """Single isolated-margin perpetual position."""
    position_id: str
    user_id: str
    symbol: str
    side: PositionSide
    size_contracts: float       # number of contracts
    entry_price: float          # in USD
    leverage: int
    margin: float               # isolated margin in USD
    status: PositionStatus = PositionStatus.ACTIVE
    realised_pnl: float = 0.0
    accumulated_funding: float = 0.0


@dataclass
class FundingRateSnapshot:
    """Single funding rate epoch result."""
    symbol: str
    timestamp: float
    premium_index: float        # (mark - index) / index
    interest_rate: float        # fixed per-epoch interest (e.g. 0.0001)
    funding_rate: float         # clamped premium + interest
    next_funding_ts: float


@dataclass
class LiquidationEvent:
    """Record of a liquidation."""
    position_id: str
    user_id: str
    symbol: str
    side: PositionSide
    size_contracts: float
    mark_price: float
    bankruptcy_price: float
    timestamp: float


@dataclass
class ADLCandidate:
    """Candidate for auto-deleveraging, ranked by profit ratio × leverage."""
    position_id: str
    user_id: str
    side: PositionSide
    pnl_ratio: float   # unrealised PnL / margin
    leverage: int
    priority_score: float


# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

FUNDING_INTERVAL_HOURS: float = 8.0
FUNDING_RATE_CLAMP: float = 0.0005  # ±5 bps clamp
DEFAULT_INTEREST_RATE: float = 0.0001  # 1 bp per epoch
DEFAULT_MAINTENANCE_MARGIN_RATE: float = 0.005  # 0.5 %


# ---------------------------------------------------------------------------
# Engine
# ---------------------------------------------------------------------------

class PerpetualsEngine:
    """
    Core perpetual-futures engine managing positions, funding, mark-price,
    liquidation, and auto-deleveraging.

    Parameters
    ----------
    maintenance_margin_rate : float
        MMR as a fraction (e.g. 0.005 for 0.5 %).
    interest_rate_per_epoch : float
        Fixed interest component of funding rate.
    """

    def __init__(
        self,
        maintenance_margin_rate: float = DEFAULT_MAINTENANCE_MARGIN_RATE,
        interest_rate_per_epoch: float = DEFAULT_INTEREST_RATE,
    ):
        self.mmr = maintenance_margin_rate
        self.interest_rate = interest_rate_per_epoch
        self.positions: Dict[str, PerpetualPosition] = {}
        self.liquidation_log: List[LiquidationEvent] = []
        self._last_funding_ts: float = time.time()

    # ---- Funding Rate ------------------------------------------------------

    def compute_funding_rate(
        self,
        symbol: str,
        mark_price: float,
        index_price: float,
    ) -> FundingRateSnapshot:
        """
        Compute the funding rate for the current epoch.

        funding_rate = clamp(premium_index + interest_rate, -CLAMP, +CLAMP)
        where premium_index = (mark_price - index_price) / index_price
        """
        if index_price <= 0:
            raise ValueError("index_price must be positive")

        premium = (mark_price - index_price) / index_price
        raw = premium + self.interest_rate
        clamped = max(-FUNDING_RATE_CLAMP, min(FUNDING_RATE_CLAMP, raw))

        now = time.time()
        return FundingRateSnapshot(
            symbol=symbol,
            timestamp=now,
            premium_index=premium,
            interest_rate=self.interest_rate,
            funding_rate=clamped,
            next_funding_ts=now + FUNDING_INTERVAL_HOURS * 3600,
        )

    def apply_funding(
        self,
        funding_rate: float,
        mark_price: float,
    ) -> List[Tuple[str, float]]:
        """
        Apply funding payments to all active positions.

        Longs pay shorts when funding_rate > 0.
        funding_payment = position_value * funding_rate
        Returns list of (position_id, payment) where positive = paid, negative = received.
        """
        payments: List[Tuple[str, float]] = []
        for pos in self.positions.values():
            if pos.status != PositionStatus.ACTIVE:
                continue
            notional = pos.size_contracts * mark_price
            payment = notional * funding_rate
            if pos.side == PositionSide.SHORT:
                payment = -payment
            pos.accumulated_funding += payment
            pos.margin -= payment  # deducted from margin
            payments.append((pos.position_id, payment))
        return payments

    # ---- Mark Price --------------------------------------------------------

    @staticmethod
    def compute_mark_price(
        index_price: float,
        last_mark: float,
        ema_weight: float = 0.1,
    ) -> float:
        """
        Compute EMA-smoothed mark price.

        mark = ema_weight * index + (1 - ema_weight) * last_mark
        """
        if index_price <= 0:
            raise ValueError("index_price must be positive")
        return ema_weight * index_price + (1.0 - ema_weight) * last_mark

    # ---- Position management -----------------------------------------------

    def open_position(
        self,
        user_id: str,
        symbol: str,
        side: PositionSide,
        size_contracts: float,
        entry_price: float,
        leverage: int,
    ) -> PerpetualPosition:
        """Open a new isolated-margin perpetual position."""
        if leverage < 1:
            raise ValueError("leverage must be >= 1")
        if size_contracts <= 0 or entry_price <= 0:
            raise ValueError("size and entry_price must be positive")

        margin = (size_contracts * entry_price) / leverage
        pos = PerpetualPosition(
            position_id=str(uuid.uuid4()),
            user_id=user_id,
            symbol=symbol,
            side=side,
            size_contracts=size_contracts,
            entry_price=entry_price,
            leverage=leverage,
            margin=margin,
        )
        self.positions[pos.position_id] = pos
        return pos

    def unrealised_pnl(self, pos: PerpetualPosition, mark_price: float) -> float:
        """Calculate unrealised PnL for a position."""
        if pos.side == PositionSide.LONG:
            return (mark_price - pos.entry_price) * pos.size_contracts
        else:
            return (pos.entry_price - mark_price) * pos.size_contracts

    # ---- Liquidation -------------------------------------------------------

    def liquidation_price(self, pos: PerpetualPosition) -> float:
        """
        Compute the deterministic liquidation price for an isolated position.

        Long:  liq = entry * (1 - 1/lev + mmr)
        Short: liq = entry * (1 + 1/lev - mmr)
        """
        e = pos.entry_price
        lev = pos.leverage
        if pos.side == PositionSide.LONG:
            return max(0.0, e * (1.0 - 1.0 / lev + self.mmr))
        else:
            return e * (1.0 + 1.0 / lev - self.mmr)

    def bankruptcy_price(self, pos: PerpetualPosition) -> float:
        """
        Compute the bankruptcy price (margin fully consumed).

        Long:  bp = entry * (1 - 1/leverage)
        Short: bp = entry * (1 + 1/leverage)
        """
        e = pos.entry_price
        lev = pos.leverage
        if pos.side == PositionSide.LONG:
            return max(0.0, e * (1.0 - 1.0 / lev))
        else:
            return e * (1.0 + 1.0 / lev)

    def check_liquidations(self, mark_price: float) -> List[LiquidationEvent]:
        """
        Scan all active positions; liquidate those breaching maintenance margin.

        Returns list of newly liquidated positions.
        """
        events: List[LiquidationEvent] = []
        for pos in list(self.positions.values()):
            if pos.status != PositionStatus.ACTIVE:
                continue
            liq_px = self.liquidation_price(pos)

            should_liquidate = False
            if pos.side == PositionSide.LONG and mark_price <= liq_px:
                should_liquidate = True
            elif pos.side == PositionSide.SHORT and mark_price >= liq_px:
                should_liquidate = True

            if should_liquidate:
                pos.status = PositionStatus.LIQUIDATED
                bp = self.bankruptcy_price(pos)
                evt = LiquidationEvent(
                    position_id=pos.position_id,
                    user_id=pos.user_id,
                    symbol=pos.symbol,
                    side=pos.side,
                    size_contracts=pos.size_contracts,
                    mark_price=mark_price,
                    bankruptcy_price=bp,
                    timestamp=time.time(),
                )
                self.liquidation_log.append(evt)
                events.append(evt)
        return events

    # ---- Auto-Deleveraging Queue -------------------------------------------

    def build_adl_queue(self, mark_price: float, target_side: PositionSide) -> List[ADLCandidate]:
        """
        Build the ADL priority queue for a given side.

        Priority = (unrealised_pnl_ratio) × leverage
        Higher priority = gets auto-deleveraged first.
        Only profitable positions on the target side are candidates.
        """
        candidates: List[ADLCandidate] = []
        for pos in self.positions.values():
            if pos.status != PositionStatus.ACTIVE or pos.side != target_side:
                continue
            upnl = self.unrealised_pnl(pos, mark_price)
            if upnl <= 0:
                continue
            pnl_ratio = upnl / pos.margin if pos.margin > 0 else 0.0
            score = pnl_ratio * pos.leverage
            candidates.append(ADLCandidate(
                position_id=pos.position_id,
                user_id=pos.user_id,
                side=pos.side,
                pnl_ratio=pnl_ratio,
                leverage=pos.leverage,
                priority_score=score,
            ))
        candidates.sort(key=lambda c: c.priority_score, reverse=True)
        return candidates

    def execute_adl(
        self,
        liquidated_size: float,
        mark_price: float,
        target_side: PositionSide,
    ) -> List[Tuple[str, float]]:
        """
        Execute auto-deleveraging against the ADL queue.

        Returns list of (position_id, size_reduced).
        """
        queue = self.build_adl_queue(mark_price, target_side)
        remaining = liquidated_size
        reductions: List[Tuple[str, float]] = []

        for candidate in queue:
            if remaining <= 0:
                break
            pos = self.positions[candidate.position_id]
            reduce = min(pos.size_contracts, remaining)
            pos.size_contracts -= reduce
            remaining -= reduce
            if pos.size_contracts <= 0:
                pos.status = PositionStatus.ADL_CLOSED
            reductions.append((pos.position_id, reduce))

        return reductions
