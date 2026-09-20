"""
Growww / NBSE Corporate Action Split Rebase Invariant Fuzz Harness
Validates that forward stock splits (1:2, 1:5, 1:10) and reverse splits (2:1, 5:1, 10:1)
execute concurrently with trading while conserving total aggregate position value within 1 paise precision.
"""

from dataclasses import dataclass
from typing import Dict, List, Tuple


@dataclass(frozen=True)
class StockSplitEvent:
    symbol: str
    numerator: int # e.g. 2 for 1:2 split (2 new shares for 1 old)
    denominator: int # e.g. 1
    effective_timestamp_ns: int


@dataclass
class PortfolioPosition:
    account_id: str
    symbol: str
    quantity: int
    average_buy_price_paise: int


@dataclass
class OpenOrder:
    order_id: int
    account_id: str
    symbol: str
    price_paise: int
    quantity: int
    side: str


@dataclass(frozen=True)
class RebaseVerificationReport:
    total_pre_rebase_value_paise: int
    total_post_rebase_value_paise: int
    valuation_delta_paise: int
    rounding_leakage_paise: int
    orders_rebased_count: int
    orders_cancelled_count: int


class CorporateActionFuzzCoordinator:
    """Coordinates corporate action stock splits across positions and in-flight orders."""

    def execute_stock_split(
        self,
        split: StockSplitEvent,
        positions: List[PortfolioPosition],
        open_orders: List[OpenOrder],
    ) -> Tuple[List[PortfolioPosition], List[OpenOrder], RebaseVerificationReport]:
        total_pre_val = 0
        total_post_val = 0
        rounding_leakage = 0

        # Process Portfolio Positions
        rebased_positions: List[PortfolioPosition] = []
        for pos in positions:
            if pos.symbol != split.symbol:
                rebased_positions.append(pos)
                continue

            pre_val = pos.quantity * pos.average_buy_price_paise
            total_pre_val += pre_val

            # Quantity scales by numerator / denominator
            new_qty = (pos.quantity * split.numerator) // split.denominator
            # Price scales inversely: price * denominator / numerator
            new_price = (pos.average_buy_price_paise * split.denominator) // split.numerator

            post_val = new_qty * new_price
            total_post_val += post_val

            rounding_diff = abs(pre_val - post_val)
            rounding_leakage += rounding_diff

            rebased_positions.append(
                PortfolioPosition(
                    account_id=pos.account_id,
                    symbol=pos.symbol,
                    quantity=new_qty,
                    average_buy_price_paise=new_price,
                )
            )

        # Process Open Orders (Rebase resting limit orders)
        rebased_orders: List[OpenOrder] = []
        orders_rebased_count = 0
        orders_cancelled_count = 0

        for ord in open_orders:
            if ord.symbol != split.symbol:
                rebased_orders.append(ord)
                continue

            new_order_qty = (ord.quantity * split.numerator) // split.denominator
            new_order_price = (ord.price_paise * split.denominator) // split.numerator

            if new_order_qty == 0 or new_order_price == 0:
                # Fractional remainder order cancelled
                orders_cancelled_count += 1
            else:
                orders_rebased_count += 1
                rebased_orders.append(
                    OpenOrder(
                        order_id=ord.order_id,
                        account_id=ord.account_id,
                        symbol=ord.symbol,
                        price_paise=new_order_price,
                        quantity=new_order_qty,
                        side=ord.side,
                    )
                )

        val_delta = abs(total_pre_val - total_post_val)
        report = RebaseVerificationReport(
            total_pre_rebase_value_paise=total_pre_val,
            total_post_rebase_value_paise=total_post_val,
            valuation_delta_paise=val_delta,
            rounding_leakage_paise=rounding_leakage,
            orders_rebased_count=orders_rebased_count,
            orders_cancelled_count=orders_cancelled_count,
        )

        return rebased_positions, rebased_orders, report

    def verify_no_valuation_drift(
        self,
        report: RebaseVerificationReport,
        max_allowed_leakage_paise: int = 1,
    ) -> bool:
        """Asserts that total portfolio valuation remains invariant within fractional integer rounding limit."""
        return report.valuation_delta_paise <= max_allowed_leakage_paise
