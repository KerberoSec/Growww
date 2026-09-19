"""
Copy Trading Replication Service – NBSE Sovereign Exchange
==========================================================
Production-grade copy-trading engine with:
  • Leader portfolio mirroring (full position replication)
  • Proportional sizing (allocation ratio)
  • Risk limits (max notional, max drawdown, max positions)
  • High-water-mark profit sharing
  • Follower notification system
"""

from __future__ import annotations

import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Callable, Dict, List, Optional, Tuple


# ---------------------------------------------------------------------------
# Enums & Data Structures
# ---------------------------------------------------------------------------

class NotificationType(Enum):
    ORDER_REPLICATED = "ORDER_REPLICATED"
    RISK_LIMIT_HIT = "RISK_LIMIT_HIT"
    PROFIT_SHARE_SETTLED = "PROFIT_SHARE_SETTLED"
    LEADER_TRADE = "LEADER_TRADE"
    STOP_LOSS_TRIGGERED = "STOP_LOSS_TRIGGERED"


class RejectReason(Enum):
    NONE = "NONE"
    MAX_NOTIONAL_EXCEEDED = "MAX_NOTIONAL_EXCEEDED"
    MAX_DRAWDOWN_EXCEEDED = "MAX_DRAWDOWN_EXCEEDED"
    MAX_POSITIONS_EXCEEDED = "MAX_POSITIONS_EXCEEDED"
    ZERO_QUANTITY = "ZERO_QUANTITY"


@dataclass
class FollowerConfig:
    """Configuration for a follower's copy-trading relationship."""
    follower_id: str
    leader_id: str
    allocation_ratio: float       # e.g. 0.10 → 10% of leader's position
    max_notional_usd: float       # maximum USD exposure
    max_drawdown_pct: float       # max drawdown before auto-stop (e.g. 0.20 = 20%)
    max_positions: int            # max simultaneous replicated positions
    profit_share_bps: int         # e.g. 1000 = 10% HWM profit share to leader
    active: bool = True


@dataclass
class ReplicatedPosition:
    """A position replicated from leader to follower."""
    position_id: str
    leader_position_id: str
    follower_id: str
    leader_id: str
    symbol: str
    side: str                     # "BUY" or "SELL"
    quantity_e8: int
    entry_price_usd: float
    current_price_usd: float
    notional_usd: float
    created_at: float


@dataclass
class FollowerPortfolioState:
    """Tracks a follower's portfolio state for risk management."""
    follower_id: str
    leader_id: str
    positions: Dict[str, ReplicatedPosition] = field(default_factory=dict)
    total_notional_usd: float = 0.0
    high_water_mark_usd: float = 0.0
    current_equity_usd: float = 0.0
    initial_equity_usd: float = 0.0
    profit_share_paid_usd: float = 0.0


@dataclass
class Notification:
    """Notification sent to a follower."""
    notification_id: str
    follower_id: str
    notification_type: NotificationType
    message: str
    timestamp: float
    metadata: Dict = field(default_factory=dict)


@dataclass
class ReplicationResult:
    """Result of attempting to replicate a leader's trade."""
    success: bool
    follower_id: str
    child_order_qty_e8: int = 0
    notional_usd: float = 0.0
    reject_reason: RejectReason = RejectReason.NONE
    position_id: str = ""


# ---------------------------------------------------------------------------
# Engine
# ---------------------------------------------------------------------------

class CopyTradingEngine:
    """
    Production-grade copy-trading replication engine.

    Manages follower subscriptions, proportional order sizing,
    risk limits, profit sharing, and notifications.
    """

    def __init__(self):
        self.followers: Dict[str, List[FollowerConfig]] = {}        # leader_id -> configs
        self.portfolios: Dict[str, FollowerPortfolioState] = {}     # follower_id -> state
        self.notifications: List[Notification] = []
        self._notification_callback: Optional[Callable[[Notification], None]] = None

    # ---- Registration ------------------------------------------------------

    def register_follower(self, config: FollowerConfig) -> None:
        """Register a follower to copy a leader's trades."""
        if config.allocation_ratio <= 0 or config.allocation_ratio > 1.0:
            raise ValueError("allocation_ratio must be in (0, 1.0]")
        if config.max_notional_usd <= 0:
            raise ValueError("max_notional_usd must be positive")

        self.followers.setdefault(config.leader_id, []).append(config)
        portfolio_key = f"{config.follower_id}:{config.leader_id}"
        self.portfolios[portfolio_key] = FollowerPortfolioState(
            follower_id=config.follower_id,
            leader_id=config.leader_id,
        )

    def unregister_follower(self, follower_id: str, leader_id: str) -> bool:
        """Unregister a follower from a leader."""
        if leader_id not in self.followers:
            return False
        original_len = len(self.followers[leader_id])
        self.followers[leader_id] = [
            f for f in self.followers[leader_id] if f.follower_id != follower_id
        ]
        return len(self.followers[leader_id]) < original_len

    def set_notification_callback(self, callback: Callable[[Notification], None]) -> None:
        """Set a callback for real-time follower notifications."""
        self._notification_callback = callback

    # ---- Replication -------------------------------------------------------

    def replicate_trade(
        self,
        leader_id: str,
        leader_position_id: str,
        symbol: str,
        side: str,
        leader_qty_e8: int,
        price_usd: float,
    ) -> List[ReplicationResult]:
        """
        Replicate a leader's trade to all active followers.

        Returns a list of replication results (one per follower).
        """
        results: List[ReplicationResult] = []

        follower_configs = self.followers.get(leader_id, [])
        for config in follower_configs:
            if not config.active:
                continue
            result = self._replicate_for_follower(
                config, leader_position_id, symbol, side, leader_qty_e8, price_usd,
            )
            results.append(result)

        return results

    def _replicate_for_follower(
        self,
        config: FollowerConfig,
        leader_position_id: str,
        symbol: str,
        side: str,
        leader_qty_e8: int,
        price_usd: float,
    ) -> ReplicationResult:
        """Replicate a single trade for one follower with risk checks."""
        portfolio_key = f"{config.follower_id}:{config.leader_id}"
        portfolio = self.portfolios.get(portfolio_key)
        if portfolio is None:
            portfolio = FollowerPortfolioState(
                follower_id=config.follower_id, leader_id=config.leader_id,
            )
            self.portfolios[portfolio_key] = portfolio

        # Proportional sizing
        scaled_qty = int(leader_qty_e8 * config.allocation_ratio)
        if scaled_qty <= 0:
            self._notify(config.follower_id, NotificationType.RISK_LIMIT_HIT,
                         f"Scaled quantity is zero for {symbol}")
            return ReplicationResult(
                success=False, follower_id=config.follower_id,
                reject_reason=RejectReason.ZERO_QUANTITY,
            )

        notional = (scaled_qty / 1e8) * price_usd

        # Risk check: max notional
        if portfolio.total_notional_usd + notional > config.max_notional_usd:
            self._notify(config.follower_id, NotificationType.RISK_LIMIT_HIT,
                         f"Max notional ${config.max_notional_usd:.2f} would be exceeded")
            return ReplicationResult(
                success=False, follower_id=config.follower_id,
                notional_usd=notional,
                reject_reason=RejectReason.MAX_NOTIONAL_EXCEEDED,
            )

        # Risk check: max positions
        if len(portfolio.positions) >= config.max_positions:
            self._notify(config.follower_id, NotificationType.RISK_LIMIT_HIT,
                         f"Max positions ({config.max_positions}) reached")
            return ReplicationResult(
                success=False, follower_id=config.follower_id,
                reject_reason=RejectReason.MAX_POSITIONS_EXCEEDED,
            )

        # Risk check: max drawdown
        if portfolio.initial_equity_usd > 0:
            drawdown = (portfolio.high_water_mark_usd - portfolio.current_equity_usd) / portfolio.high_water_mark_usd
            if drawdown >= config.max_drawdown_pct:
                config.active = False
                self._notify(config.follower_id, NotificationType.STOP_LOSS_TRIGGERED,
                             f"Max drawdown {config.max_drawdown_pct*100:.1f}% hit, auto-stopped")
                return ReplicationResult(
                    success=False, follower_id=config.follower_id,
                    reject_reason=RejectReason.MAX_DRAWDOWN_EXCEEDED,
                )

        # Create replicated position
        pos_id = str(uuid.uuid4())
        position = ReplicatedPosition(
            position_id=pos_id,
            leader_position_id=leader_position_id,
            follower_id=config.follower_id,
            leader_id=config.leader_id,
            symbol=symbol,
            side=side,
            quantity_e8=scaled_qty,
            entry_price_usd=price_usd,
            current_price_usd=price_usd,
            notional_usd=notional,
            created_at=time.time(),
        )
        portfolio.positions[pos_id] = position
        portfolio.total_notional_usd += notional

        self._notify(config.follower_id, NotificationType.ORDER_REPLICATED,
                     f"Replicated {side} {scaled_qty/1e8:.4f} {symbol} @ ${price_usd:.2f}")

        return ReplicationResult(
            success=True,
            follower_id=config.follower_id,
            child_order_qty_e8=scaled_qty,
            notional_usd=notional,
            position_id=pos_id,
        )

    # ---- Profit Sharing (High-Water Mark) ----------------------------------

    def calculate_profit_share(
        self,
        follower_id: str,
        leader_id: str,
        current_equity_usd: float,
    ) -> float:
        """
        Calculate and deduct high-water-mark profit share.

        Returns the profit share amount in USD owed to the leader.
        """
        portfolio_key = f"{follower_id}:{leader_id}"
        portfolio = self.portfolios.get(portfolio_key)
        if portfolio is None:
            return 0.0

        config = self._get_follower_config(follower_id, leader_id)
        if config is None:
            return 0.0

        portfolio.current_equity_usd = current_equity_usd

        # Only charge on new high-water mark
        if current_equity_usd <= portfolio.high_water_mark_usd:
            return 0.0

        profit_above_hwm = current_equity_usd - portfolio.high_water_mark_usd
        share_rate = config.profit_share_bps / 10_000.0
        share_amount = profit_above_hwm * share_rate

        portfolio.high_water_mark_usd = current_equity_usd
        portfolio.profit_share_paid_usd += share_amount

        self._notify(follower_id, NotificationType.PROFIT_SHARE_SETTLED,
                     f"Profit share ${share_amount:.2f} settled to leader {leader_id}",
                     metadata={"amount_usd": share_amount})

        return share_amount

    # ---- Portfolio Mirroring -----------------------------------------------

    def get_leader_portfolio_snapshot(
        self,
        leader_id: str,
    ) -> Dict[str, List[ReplicatedPosition]]:
        """
        Get a snapshot of all followers' replicated positions for a leader.

        Returns follower_id -> list of positions.
        """
        snapshot: Dict[str, List[ReplicatedPosition]] = {}
        for key, portfolio in self.portfolios.items():
            if portfolio.leader_id == leader_id:
                snapshot[portfolio.follower_id] = list(portfolio.positions.values())
        return snapshot

    # ---- Notifications -----------------------------------------------------

    def _notify(
        self,
        follower_id: str,
        ntype: NotificationType,
        message: str,
        metadata: Optional[Dict] = None,
    ) -> Notification:
        n = Notification(
            notification_id=str(uuid.uuid4()),
            follower_id=follower_id,
            notification_type=ntype,
            message=message,
            timestamp=time.time(),
            metadata=metadata or {},
        )
        self.notifications.append(n)
        if self._notification_callback:
            self._notification_callback(n)
        return n

    def get_notifications(self, follower_id: str) -> List[Notification]:
        """Get all notifications for a specific follower."""
        return [n for n in self.notifications if n.follower_id == follower_id]

    # ---- Helpers -----------------------------------------------------------

    def _get_follower_config(self, follower_id: str, leader_id: str) -> Optional[FollowerConfig]:
        for config in self.followers.get(leader_id, []):
            if config.follower_id == follower_id:
                return config
        return None
