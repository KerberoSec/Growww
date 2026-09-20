"""
Growww / NBSE Cross-Venue Latency Arbitrage & Smart Order Routing (SOR) Resilience Harness
Simulates routing across multi-exchange venues (NSE, BSE, NBSE, GIFT City) with asymmetric latency jitter,
packet loss, quote freshness validation, and strict slippage bounds.
"""

import random
import time
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple


@dataclass
class VenueQuote:
    venue: str
    symbol: str
    bid_price_paise: int
    ask_price_paise: int
    bid_size: int
    ask_size: int
    quote_timestamp_ns: int


@dataclass
class RoutedExecutionResult:
    order_id: str
    target_venue: str
    intended_price_paise: int
    executed_price_paise: int
    executed_quantity: int
    slippage_bps: float
    is_quote_stale: bool
    status: str # FILLED, REJECTED_STALE_QUOTE, REJECTED_EXCESSIVE_SLIPPAGE, DROPPED_PACKET


class SmartOrderRouter:
    """Institutional SOR engine with quote freshness verification and slippage guard."""

    def __init__(
        self,
        venues: List[str] = None,
        max_allowed_slippage_bps: float = 15.0, # 15 bps (0.15%)
        max_quote_age_ms: int = 100, # 100ms max freshness threshold
    ):
        self.venues = venues or ["NSE", "BSE", "NBSE", "GIFT_CITY"]
        self.max_allowed_slippage_bps = max_allowed_slippage_bps
        self.max_quote_age_ms = max_quote_age_ms
        self.venue_latency_jitter_ms: Dict[str, Tuple[int, int]] = {
            "NSE": (1, 15),
            "BSE": (2, 20),
            "NBSE": (0, 2), # Co-located
            "GIFT_CITY": (10, 80),
        }
        self.venue_packet_loss_rate: Dict[str, float] = {
            "NSE": 0.01,
            "BSE": 0.01,
            "NBSE": 0.00,
            "GIFT_CITY": 0.03,
        }

    def select_best_venue(
        self,
        symbol: str,
        side: str, # BUY or SELL
        quotes: Dict[str, VenueQuote],
        current_time_ns: int,
    ) -> Optional[VenueQuote]:
        """Selects the best available quote considering price and quote freshness."""
        valid_quotes: List[VenueQuote] = []

        for venue, q in quotes.items():
            quote_age_ms = (current_time_ns - q.quote_timestamp_ns) / 1_000_000.0
            if quote_age_ms <= self.max_quote_age_ms:
                valid_quotes.append(q)

        if not valid_quotes:
            return None

        if side == "BUY":
            # Best ask (lowest price)
            return min(valid_quotes, key=lambda q: (q.ask_price_paise, -q.ask_size))
        else:
            # Best bid (highest price)
            return max(valid_quotes, key=lambda q: (q.bid_price_paise, q.bid_size))

    def route_and_execute(
        self,
        order_id: str,
        symbol: str,
        side: str,
        quantity: int,
        quotes: Dict[str, VenueQuote],
        current_time_ns: int,
        adversarial_price_drift_paise: int = 0,
    ) -> RoutedExecutionResult:
        best_quote = self.select_best_venue(symbol, side, quotes, current_time_ns)

        if not best_quote:
            return RoutedExecutionResult(
                order_id=order_id,
                target_venue="NONE",
                intended_price_paise=0,
                executed_price_paise=0,
                executed_quantity=0,
                slippage_bps=0.0,
                is_quote_stale=True,
                status="REJECTED_STALE_QUOTE",
            )

        venue = best_quote.venue
        packet_loss = self.venue_packet_loss_rate.get(venue, 0.0)
        if random.random() < packet_loss:
            return RoutedExecutionResult(
                order_id=order_id,
                target_venue=venue,
                intended_price_paise=best_quote.ask_price_paise if side == "BUY" else best_quote.bid_price_paise,
                executed_price_paise=0,
                executed_quantity=0,
                slippage_bps=0.0,
                is_quote_stale=False,
                status="DROPPED_PACKET",
            )

        intended_price = best_quote.ask_price_paise if side == "BUY" else best_quote.bid_price_paise
        # In-flight latency allows adversarial market price drift
        actual_price = intended_price + (adversarial_price_drift_paise if side == "BUY" else -adversarial_price_drift_paise)

        # Slippage calculation in basis points
        price_diff = abs(actual_price - intended_price)
        slippage_bps = (price_diff / intended_price) * 10000.0 if intended_price > 0 else 0.0

        if slippage_bps > self.max_allowed_slippage_bps:
            return RoutedExecutionResult(
                order_id=order_id,
                target_venue=venue,
                intended_price_paise=intended_price,
                executed_price_paise=0,
                executed_quantity=0,
                slippage_bps=slippage_bps,
                is_quote_stale=False,
                status="REJECTED_EXCESSIVE_SLIPPAGE",
            )

        fill_qty = min(quantity, best_quote.ask_size if side == "BUY" else best_quote.bid_size)
        return RoutedExecutionResult(
            order_id=order_id,
            target_venue=venue,
            intended_price_paise=intended_price,
            executed_price_paise=actual_price,
            executed_quantity=fill_qty,
            slippage_bps=slippage_bps,
            is_quote_stale=False,
            status="FILLED",
        )
