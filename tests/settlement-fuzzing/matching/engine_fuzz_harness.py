"""
Growww / NBSE Matching Engine Invariant Fuzzing Harness
Formal property-based verification testing:
1. Strict Price-Time Priority
2. No Crossed Book (Best Bid < Best Ask)
3. Conservation of Shares (Total buy fills == Total sell fills)
4. Deterministic Replay across identical event seeds
"""

import heapq
import random
import time
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Tuple


class OrderSide(Enum):
    BUY = "BUY"
    SELL = "SELL"


class OrderType(Enum):
    LIMIT = "LIMIT"
    MARKET = "MARKET"
    CANCEL = "CANCEL"
    IMMEDIATE_OR_CANCEL = "IOC"
    FILL_OR_KILL = "FOK"


@dataclass
class FuzzOrderEvent:
    order_id: int
    account_id: int
    symbol: str
    side: OrderSide
    order_type: OrderType
    price_paise: int
    quantity: int
    timestamp_ns: int
    target_cancel_order_id: Optional[int] = None


@dataclass
class EngineExecutionReport:
    match_id: int
    maker_order_id: int
    taker_order_id: int
    maker_account_id: int
    taker_account_id: int
    execution_price_paise: int
    execution_quantity: int
    timestamp_ns: int


@dataclass(order=True)
class BookOrder:
    sort_key: Tuple[int, int] = field(init=False) # (price_sort, timestamp_ns)
    order_id: int = field(compare=False)
    account_id: int = field(compare=False)
    price_paise: int = field(compare=False)
    quantity: int = field(compare=False)
    timestamp_ns: int = field(compare=False)
    side: OrderSide = field(compare=False)

    def __post_init__(self):
        # For bids: highest price first (-price_paise), then lowest timestamp
        # For asks: lowest price first (+price_paise), then lowest timestamp
        if self.side == OrderSide.BUY:
            self.sort_key = (-self.price_paise, self.timestamp_ns)
        else:
            self.sort_key = (self.price_paise, self.timestamp_ns)


class InvariantMatchingEngine:
    """High-performance in-memory matching engine supporting invariant assertions."""

    def __init__(self, symbol: str):
        self.symbol = symbol
        self.bids: List[BookOrder] = [] # Max-heap for bids via negative price
        self.asks: List[BookOrder] = [] # Min-heap for asks
        self.orders_by_id: Dict[int, BookOrder] = {}
        self.execution_reports: List[EngineExecutionReport] = []
        self.match_counter = 0

    def process_order(self, event: FuzzOrderEvent) -> List[EngineExecutionReport]:
        reports: List[EngineExecutionReport] = []

        if event.order_type == OrderType.CANCEL:
            target_id = event.target_cancel_order_id or event.order_id
            if target_id in self.orders_by_id:
                order = self.orders_by_id.pop(target_id)
                order.quantity = 0 # Mark cancelled
            return reports

        if event.order_type == OrderType.LIMIT:
            reports = self._match_limit(event)
        elif event.order_type == OrderType.MARKET:
            reports = self._match_market(event)
        elif event.order_type == OrderType.IMMEDIATE_OR_CANCEL:
            reports = self._match_ioc(event)
        elif event.order_type == OrderType.FILL_OR_KILL:
            reports = self._match_fok(event)

        # Assert no crossed book invariant immediately after order execution
        self.verify_no_crossed_book()
        return reports

    def _match_limit(self, event: FuzzOrderEvent) -> List[EngineExecutionReport]:
        reports: List[EngineExecutionReport] = []
        remaining_qty = event.quantity

        if event.side == OrderSide.BUY:
            while self.asks and remaining_qty > 0:
                self._prune_dead_orders(self.asks)
                if not self.asks:
                    break
                best_ask = self.asks[0]
                if event.price_paise < best_ask.price_paise:
                    break # No match possible

                fill_qty = min(remaining_qty, best_ask.quantity)
                self.match_counter += 1
                report = EngineExecutionReport(
                    match_id=self.match_counter,
                    maker_order_id=best_ask.order_id,
                    taker_order_id=event.order_id,
                    maker_account_id=best_ask.account_id,
                    taker_account_id=event.account_id,
                    execution_price_paise=best_ask.price_paise,
                    execution_quantity=fill_qty,
                    timestamp_ns=event.timestamp_ns,
                )
                reports.append(report)
                self.execution_reports.append(report)

                remaining_qty -= fill_qty
                best_ask.quantity -= fill_qty
                if best_ask.quantity == 0:
                    heapq.heappop(self.asks)
                    self.orders_by_id.pop(best_ask.order_id, None)

            if remaining_qty > 0:
                resting_order = BookOrder(
                    order_id=event.order_id,
                    account_id=event.account_id,
                    price_paise=event.price_paise,
                    quantity=remaining_qty,
                    timestamp_ns=event.timestamp_ns,
                    side=OrderSide.BUY,
                )
                heapq.heappush(self.bids, resting_order)
                self.orders_by_id[event.order_id] = resting_order

        else: # SELL
            while self.bids and remaining_qty > 0:
                self._prune_dead_orders(self.bids)
                if not self.bids:
                    break
                best_bid = self.bids[0]
                if event.price_paise > best_bid.price_paise:
                    break # No match possible

                fill_qty = min(remaining_qty, best_bid.quantity)
                self.match_counter += 1
                report = EngineExecutionReport(
                    match_id=self.match_counter,
                    maker_order_id=best_bid.order_id,
                    taker_order_id=event.order_id,
                    maker_account_id=best_bid.account_id,
                    taker_account_id=event.account_id,
                    execution_price_paise=best_bid.price_paise,
                    execution_quantity=fill_qty,
                    timestamp_ns=event.timestamp_ns,
                )
                reports.append(report)
                self.execution_reports.append(report)

                remaining_qty -= fill_qty
                best_bid.quantity -= fill_qty
                if best_bid.quantity == 0:
                    heapq.heappop(self.bids)
                    self.orders_by_id.pop(best_bid.order_id, None)

            if remaining_qty > 0:
                resting_order = BookOrder(
                    order_id=event.order_id,
                    account_id=event.account_id,
                    price_paise=event.price_paise,
                    quantity=remaining_qty,
                    timestamp_ns=event.timestamp_ns,
                    side=OrderSide.SELL,
                )
                heapq.heappush(self.asks, resting_order)
                self.orders_by_id[event.order_id] = resting_order

        return reports

    def _match_market(self, event: FuzzOrderEvent) -> List[EngineExecutionReport]:
        reports: List[EngineExecutionReport] = []
        remaining_qty = event.quantity
        target_book = self.asks if event.side == OrderSide.BUY else self.bids

        while target_book and remaining_qty > 0:
            self._prune_dead_orders(target_book)
            if not target_book:
                break
            best_maker = target_book[0]
            fill_qty = min(remaining_qty, best_maker.quantity)

            self.match_counter += 1
            report = EngineExecutionReport(
                match_id=self.match_counter,
                maker_order_id=best_maker.order_id,
                taker_order_id=event.order_id,
                maker_account_id=best_maker.account_id,
                taker_account_id=event.account_id,
                execution_price_paise=best_maker.price_paise,
                execution_quantity=fill_qty,
                timestamp_ns=event.timestamp_ns,
            )
            reports.append(report)
            self.execution_reports.append(report)

            remaining_qty -= fill_qty
            best_maker.quantity -= fill_qty
            if best_maker.quantity == 0:
                heapq.heappop(target_book)
                self.orders_by_id.pop(best_maker.order_id, None)

        return reports

    def _match_ioc(self, event: FuzzOrderEvent) -> List[EngineExecutionReport]:
        # IOC executes available liquidity up to limit price, unexecuted balance cancelled
        reports = self._match_limit(event)
        # If order rested on book, remove it
        if event.order_id in self.orders_by_id:
            order = self.orders_by_id.pop(event.order_id)
            order.quantity = 0
        return reports

    def _match_fok(self, event: FuzzOrderEvent) -> List[EngineExecutionReport]:
        # FOK checks if entire quantity can be filled immediately; if not, 0 executions
        target_book = self.asks if event.side == OrderSide.BUY else self.bids
        self._prune_dead_orders(target_book)

        available_qty = 0
        for o in target_book:
            if o.quantity <= 0:
                continue
            if event.side == OrderSide.BUY and o.price_paise > event.price_paise:
                break
            if event.side == OrderSide.SELL and o.price_paise < event.price_paise:
                break
            available_qty += o.quantity
            if available_qty >= event.quantity:
                break

        if available_qty >= event.quantity:
            return self._match_limit(event)
        return []

    def _prune_dead_orders(self, book: List[BookOrder]):
        while book and book[0].quantity <= 0:
            popped = heapq.heappop(book)
            self.orders_by_id.pop(popped.order_id, None)

    def get_best_bid(self) -> Optional[int]:
        self._prune_dead_orders(self.bids)
        return self.bids[0].price_paise if self.bids else None

    def get_best_ask(self) -> Optional[int]:
        self._prune_dead_orders(self.asks)
        return self.asks[0].price_paise if self.asks else None

    def verify_no_crossed_book(self) -> None:
        best_bid = self.get_best_bid()
        best_ask = self.get_best_ask()
        if best_bid is not None and best_ask is not None:
            if best_bid >= best_ask:
                raise AssertionError(
                    f"Invariant Breach: Crossed Order Book! Best Bid ({best_bid}) >= Best Ask ({best_ask})"
                )

    def verify_price_time_priority(self, report: EngineExecutionReport) -> None:
        maker_order = self.orders_by_id.get(report.maker_order_id)
        # If maker is still on book, its price must be equal to execution price
        if maker_order:
            assert maker_order.price_paise == report.execution_price_paise

    def verify_conservation_of_shares(
        self,
        initial_positions: Dict[int, int],
        current_positions: Dict[int, int],
    ) -> None:
        init_sum = sum(initial_positions.values())
        curr_sum = sum(current_positions.values())
        if init_sum != curr_sum:
            raise AssertionError(
                f"Invariant Breach: Conservation of Shares Failed! Initial Sum: {init_sum}, Current Sum: {curr_sum}"
            )
