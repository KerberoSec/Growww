"""
Synthetic Indian Equity & Digital Asset Market Data Generator.
Replicates historical NSE/BSE tick-by-tick order flow characteristics:
- Price clustering around discrete tick sizes (0.05 INR / 5 paise)
- Realistic bid-ask spread dynamics
- Power-law order volume distribution
- Limit orders (75%), Market orders (15%), Cancellations (8%), Modifications (2%)
"""

import math
import random
import time
from dataclasses import dataclass
from enum import Enum
from typing import List, Optional


class OrderSide(Enum):
    BUY = "BUY"
    SELL = "SELL"


class OrderType(Enum):
    LIMIT = "LIMIT"
    MARKET = "MARKET"
    CANCEL = "CANCEL"
    MODIFY = "MODIFY"


@dataclass
class OrderEvent:
    order_id: int
    symbol: str
    side: OrderSide
    order_type: OrderType
    price: float
    quantity: float
    timestamp_ns: int
    user_tier: str
    target_order_id: Optional[int] = None


class SyntheticMarketGenerator:
    def __init__(self, symbol: str = "RELIANCE", base_price: float = 2850.0, tick_size: float = 0.05, seed: int = 42):
        self.symbol = symbol
        self.base_price = base_price
        self.tick_size = tick_size
        self.current_mid_price = base_price
        self.order_counter = 0
        self.active_order_ids: List[int] = []
        random.seed(seed)

    def round_to_tick(self, price: float) -> float:
        """Round price to the nearest exchange tick size."""
        return round(round(price / self.tick_size) * self.tick_size, 2)

    def generate_event(self) -> OrderEvent:
        """Generate a single realistic order event."""
        self.order_counter += 1
        now_ns = time.perf_counter_ns()

        # Drift mid-price slightly via geometric Brownian motion step
        drift = random.gauss(0, 0.2)
        self.current_mid_price = max(self.base_price * 0.8, min(self.base_price * 1.2, self.current_mid_price + drift))

        # Event distribution: 75% Limit, 15% Market, 8% Cancel, 2% Modify
        dice = random.random()
        side = OrderSide.BUY if random.random() > 0.5 else OrderSide.SELL

        tier_weights = ["tier1_retail"] * 70 + ["tier2_institutional"] * 25 + ["tier3_market_maker"] * 5
        user_tier = random.choice(tier_weights)

        # Quantity follows Pareto / Power-law distribution
        quantity = round(min(500.0, max(1.0, random.paretovariate(1.5))), 2)

        if dice < 0.75:
            # Limit order placed near the spread
            spread_offset = abs(random.gauss(0, 0.5)) + self.tick_size
            if side == OrderSide.BUY:
                price = self.round_to_tick(self.current_mid_price - spread_offset)
            else:
                price = self.round_to_tick(self.current_mid_price + spread_offset)

            event = OrderEvent(
                order_id=self.order_counter,
                symbol=self.symbol,
                side=side,
                order_type=OrderType.LIMIT,
                price=max(self.tick_size, price),
                quantity=quantity,
                timestamp_ns=now_ns,
                user_tier=user_tier,
            )
            self.active_order_ids.append(self.order_counter)
            if len(self.active_order_ids) > 10000:
                self.active_order_ids.pop(0)
            return event

        elif dice < 0.90:
            # Aggressive Market Order
            return OrderEvent(
                order_id=self.order_counter,
                symbol=self.symbol,
                side=side,
                order_type=OrderType.MARKET,
                price=0.0,
                quantity=min(20.0, quantity),
                timestamp_ns=now_ns,
                user_tier=user_tier,
            )

        elif dice < 0.98 and self.active_order_ids:
            # Cancellation
            target_id = random.choice(self.active_order_ids)
            return OrderEvent(
                order_id=self.order_counter,
                symbol=self.symbol,
                side=side,
                order_type=OrderType.CANCEL,
                price=0.0,
                quantity=0.0,
                timestamp_ns=now_ns,
                user_tier=user_tier,
                target_order_id=target_id,
            )

        else:
            # Modification
            target_id = random.choice(self.active_order_ids) if self.active_order_ids else self.order_counter
            return OrderEvent(
                order_id=self.order_counter,
                symbol=self.symbol,
                side=side,
                order_type=OrderType.MODIFY,
                price=self.round_to_tick(self.current_mid_price),
                quantity=quantity,
                timestamp_ns=now_ns,
                user_tier=user_tier,
                target_order_id=target_id,
            )

    def generate_batch(self, count: int) -> List[OrderEvent]:
        """Generate a batch of pre-allocated order events."""
        return [self.generate_event() for _ in range(count)]
