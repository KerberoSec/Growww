"""
Institutional Dark Pool – NBSE Sovereign Exchange
===================================================
Production-grade dark-pool crossing engine with:
  • Midpoint crossing at NBBO midpoint
  • Minimum quantity enforcement (block-size floors)
  • Information leakage prevention (anonymous IDs, delayed reports, anti-gaming)
  • Time-priority FIFO matching
"""

from __future__ import annotations

import hashlib
import time
import uuid
from collections import deque
from dataclasses import dataclass, field
from enum import Enum
from typing import Deque, Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class OrderSide(Enum):
    BUY = "BUY"
    SELL = "SELL"


class OrderStatus(Enum):
    RESTING = "RESTING"
    FILLED = "FILLED"
    PARTIALLY_FILLED = "PARTIALLY_FILLED"
    CANCELLED = "CANCELLED"
    REJECTED = "REJECTED"


@dataclass
class DarkPoolOrder:
    """Institutional dark-pool order."""
    order_id: str                     # internal UUID
    anonymous_id: str                 # hashed ID exposed externally
    participant_id: str
    symbol: str
    side: OrderSide
    total_quantity_e8: int            # total order size (8-decimal fixed point)
    remaining_quantity_e8: int
    min_fill_quantity_e8: int         # minimum acceptable fill size
    max_slippage_bps: int             # max slippage from midpoint in basis points
    status: OrderStatus = OrderStatus.RESTING
    submitted_at: float = field(default_factory=time.time)


@dataclass
class CrossExecution:
    """Record of an executed midpoint cross."""
    execution_id: str
    buy_anonymous_id: str             # never exposes real participant
    sell_anonymous_id: str
    symbol: str
    execution_price_e8: int
    quantity_e8: int
    executed_at: float
    reported_at: float                # delayed report timestamp


@dataclass
class NBBOQuote:
    """National Best Bid and Offer snapshot."""
    symbol: str
    best_bid_e8: int
    best_ask_e8: int
    midpoint_e8: int
    timestamp: float


# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

DEFAULT_MIN_BLOCK_SIZE_E8: int = 100_00000000   # 100 units minimum block
DEFAULT_REPORT_DELAY_SECS: float = 15.0          # 15-second delayed reporting
MAX_ORDER_LIFESPAN_SECS: float = 3600.0           # 1-hour max resting time
ANTI_GAMING_MIN_INTERVAL_SECS: float = 1.0        # min interval between orders per participant


# ---------------------------------------------------------------------------
# Engine
# ---------------------------------------------------------------------------

class DarkPoolCrossingEngine:
    """
    Midpoint crossing facility for institutional block orders.

    Features:
      - Orders matched strictly at NBBO midpoint
      - Minimum block-size enforcement
      - Anti-gaming: rate limiting, anonymous IDs, delayed fill reports
      - FIFO priority within each side
    """

    def __init__(
        self,
        min_block_size_e8: int = DEFAULT_MIN_BLOCK_SIZE_E8,
        report_delay_secs: float = DEFAULT_REPORT_DELAY_SECS,
    ):
        self.min_block_size_e8 = min_block_size_e8
        self.report_delay_secs = report_delay_secs
        self.resting_bids: Deque[DarkPoolOrder] = deque()
        self.resting_asks: Deque[DarkPoolOrder] = deque()
        self.executions: List[CrossExecution] = []
        self._last_order_time: Dict[str, float] = {}  # participant_id -> last submit ts

    # ---- NBBO --------------------------------------------------------------

    @staticmethod
    def compute_nbbo(symbol: str, best_bid_e8: int, best_ask_e8: int) -> NBBOQuote:
        """Compute NBBO midpoint from lit market quotes."""
        if best_bid_e8 <= 0 or best_ask_e8 <= 0:
            raise ValueError("bid and ask must be positive")
        if best_bid_e8 > best_ask_e8:
            raise ValueError("bid cannot exceed ask (crossed market)")
        midpoint = (best_bid_e8 + best_ask_e8) // 2
        return NBBOQuote(
            symbol=symbol,
            best_bid_e8=best_bid_e8,
            best_ask_e8=best_ask_e8,
            midpoint_e8=midpoint,
            timestamp=time.time(),
        )

    # ---- Order submission --------------------------------------------------

    def submit_order(
        self,
        participant_id: str,
        symbol: str,
        side: OrderSide,
        quantity_e8: int,
        min_fill_e8: int = 0,
        max_slippage_bps: int = 0,
    ) -> DarkPoolOrder:
        """
        Submit a dark-pool order.

        Parameters
        ----------
        participant_id : str – institutional participant identifier
        symbol : str – instrument symbol
        side : OrderSide – BUY or SELL
        quantity_e8 : int – order size in 8-decimal fixed point
        min_fill_e8 : int – minimum acceptable fill (0 = use default block minimum)
        max_slippage_bps : int – max basis points slippage from midpoint

        Returns
        -------
        DarkPoolOrder

        Raises
        ------
        ValueError – on validation failures
        """
        # Anti-gaming: rate limiting
        now = time.time()
        last = self._last_order_time.get(participant_id, 0.0)
        if now - last < ANTI_GAMING_MIN_INTERVAL_SECS:
            raise ValueError(
                f"Anti-gaming: minimum {ANTI_GAMING_MIN_INTERVAL_SECS}s between orders"
            )

        # Block-size enforcement
        effective_min = max(min_fill_e8, self.min_block_size_e8)
        if quantity_e8 < effective_min:
            raise ValueError(
                f"Order size {quantity_e8} below minimum block size {effective_min}"
            )

        # Generate anonymous ID (hash of real ID + salt)
        anonymous_id = self._generate_anonymous_id(participant_id)

        order = DarkPoolOrder(
            order_id=str(uuid.uuid4()),
            anonymous_id=anonymous_id,
            participant_id=participant_id,
            symbol=symbol,
            side=side,
            total_quantity_e8=quantity_e8,
            remaining_quantity_e8=quantity_e8,
            min_fill_quantity_e8=effective_min,
            max_slippage_bps=max_slippage_bps,
            submitted_at=now,
        )

        if side == OrderSide.BUY:
            self.resting_bids.append(order)
        else:
            self.resting_asks.append(order)

        self._last_order_time[participant_id] = now
        return order

    # ---- Midpoint crossing -------------------------------------------------

    def run_crossing(self, nbbo: NBBOQuote) -> List[CrossExecution]:
        """
        Run midpoint crossing cycle: match resting bids against asks at midpoint.

        Returns list of executed crosses.
        """
        crosses: List[CrossExecution] = []
        now = time.time()

        # Clean expired orders
        self._purge_expired(now)

        while self.resting_bids and self.resting_asks:
            bid = self.resting_bids[0]
            ask = self.resting_asks[0]

            # Symbol must match
            if bid.symbol != ask.symbol or bid.symbol != nbbo.symbol:
                break

            # Determine matchable quantity
            match_qty = min(bid.remaining_quantity_e8, ask.remaining_quantity_e8)

            # Enforce minimum fill on both sides
            if match_qty < bid.min_fill_quantity_e8 or match_qty < ask.min_fill_quantity_e8:
                break

            # Slippage check (both sides must accept midpoint)
            if not self._slippage_ok(bid, nbbo) or not self._slippage_ok(ask, nbbo):
                break

            # Execute cross
            bid.remaining_quantity_e8 -= match_qty
            ask.remaining_quantity_e8 -= match_qty

            # Update statuses
            if bid.remaining_quantity_e8 == 0:
                bid.status = OrderStatus.FILLED
                self.resting_bids.popleft()
            else:
                bid.status = OrderStatus.PARTIALLY_FILLED

            if ask.remaining_quantity_e8 == 0:
                ask.status = OrderStatus.FILLED
                self.resting_asks.popleft()
            else:
                ask.status = OrderStatus.PARTIALLY_FILLED

            execution = CrossExecution(
                execution_id=str(uuid.uuid4()),
                buy_anonymous_id=bid.anonymous_id,
                sell_anonymous_id=ask.anonymous_id,
                symbol=nbbo.symbol,
                execution_price_e8=nbbo.midpoint_e8,
                quantity_e8=match_qty,
                executed_at=now,
                reported_at=now + self.report_delay_secs,
            )
            self.executions.append(execution)
            crosses.append(execution)

        return crosses

    # ---- Public fill report (delayed) --------------------------------------

    def get_reportable_fills(self) -> List[CrossExecution]:
        """Return fills whose report delay has elapsed (information leakage prevention)."""
        now = time.time()
        return [e for e in self.executions if now >= e.reported_at]

    # ---- Cancel order ------------------------------------------------------

    def cancel_order(self, order_id: str) -> bool:
        """Cancel a resting order. Returns True if found and cancelled."""
        for queue in (self.resting_bids, self.resting_asks):
            for order in queue:
                if order.order_id == order_id and order.status == OrderStatus.RESTING:
                    order.status = OrderStatus.CANCELLED
                    queue.remove(order)
                    return True
        return False

    # ---- Internal helpers --------------------------------------------------

    @staticmethod
    def _generate_anonymous_id(participant_id: str) -> str:
        """Generate a one-way anonymous identifier for the participant."""
        salt = "nbse-dark-pool-v1"
        raw = f"{participant_id}:{salt}:{uuid.uuid4().hex[:8]}"
        return hashlib.sha256(raw.encode()).hexdigest()[:16]

    @staticmethod
    def _slippage_ok(order: DarkPoolOrder, nbbo: NBBOQuote) -> bool:
        """Check if midpoint is within order's max slippage tolerance."""
        if order.max_slippage_bps == 0:
            return True  # no slippage limit
        spread = nbbo.best_ask_e8 - nbbo.best_bid_e8
        half_spread_bps = (spread * 10000) // (2 * nbbo.midpoint_e8) if nbbo.midpoint_e8 > 0 else 0
        return half_spread_bps <= order.max_slippage_bps

    def _purge_expired(self, now: float) -> None:
        """Remove orders that have exceeded max lifespan."""
        for queue in (self.resting_bids, self.resting_asks):
            while queue and (now - queue[0].submitted_at > MAX_ORDER_LIFESPAN_SECS):
                expired = queue.popleft()
                expired.status = OrderStatus.CANCELLED
