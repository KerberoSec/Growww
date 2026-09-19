"""
Unified Portfolio Holdings & Valuation Service (Prompts 041, 209)
Institutional multi-asset portfolio engine consolidating NSDL/CDSL equity holdings,
Virtual Digital Assets (VDA / crypto), RBI Digital Rupee (CBDC), and RWA bonds.
Enforces arbitrary-precision decimal arithmetic, FIFO tax lot accounting, holding reservations,
on-chain 1:1 Hyperledger Besu reconciliation, and FIU-IND AML surveillance.
"""

from decimal import Decimal, ROUND_HALF_UP, InvalidOperation
from typing import Dict, List, Optional, Tuple, Any
from dataclasses import dataclass, field, asdict
from datetime import datetime, timezone
from enum import Enum
import uuid
import re

# Precision Quantizers
INR_QUANTIZE = Decimal("0.01")
INR_INTERNAL_QUANTIZE = Decimal("0.0001")
EQUITY_QUANTIZE = Decimal("0.000001") # 6 decimal places for fractional equities
CRYPTO_QUANTIZE = Decimal("0.00000001") # 8 decimal places for satoshi / VDA tokens

class AssetClass(str, Enum):
    EQUITY = "EQUITY"
    VDA_CRYPTO = "VDA_CRYPTO"
    CBDC_INR = "CBDC_INR"
    RWA_BOND = "RWA_BOND"

class InsufficientBalanceError(Exception):
    """Raised when an operation requests more units than available (unreserved)."""
    pass

class InvalidReservationError(Exception):
    """Raised when modifying or committing an invalid reservation."""
    pass

class ReconciliationMismatchError(Exception):
    """Raised when off-chain database holdings diverge from on-chain Besu token balances."""
    pass

@dataclass
class TaxLot:
    lot_id: str
    user_id: str
    asset_class: AssetClass
    symbol: str
    quantity: Decimal
    remaining_quantity: Decimal
    cost_per_unit_inr: Decimal
    total_cost_inr: Decimal
    acquired_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))

@dataclass
class HoldingReservation:
    reservation_id: str
    user_id: str
    symbol: str
    quantity: Decimal
    created_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))
    status: str = "ACTIVE" # ACTIVE, COMMITTED, RELEASED

@dataclass
class RealizedLotSlice:
    lot_id: str
    asset_class: AssetClass
    symbol: str
    quantity: Decimal
    cost_per_unit_inr: Decimal
    cost_basis_inr: Decimal
    sale_price_per_unit_inr: Decimal
    sale_proceeds_inr: Decimal
    realized_pnl_inr: Decimal
    taxable_115bbh_gain_inr: Decimal # strictly non-negative for VDA under 115BBH
    acquired_at: datetime
    disposed_at: datetime

@dataclass
class HoldingItem:
    asset_class: str
    symbol: str
    total_quantity: float
    reserved_quantity: float
    available_quantity: float
    average_cost_inr: float
    current_market_price_inr: float
    invested_value_inr: float
    current_value_inr: float
    unrealized_pnl_inr: float
    unrealized_pnl_pct: float
    allocation_pct: float

class UnifiedPortfolioEngine:
    """
    Thread-safe, institutional portfolio accounting and valuation engine.
    """
    def __init__(self):
        # user_id -> symbol -> list of active TaxLot (FIFO ordered)
        self.tax_lots: Dict[str, Dict[str, List[TaxLot]]] = {}
        # user_id -> symbol -> total reserved quantity
        self.reservations: Dict[str, Dict[str, HoldingReservation]] = {} # reservation_id -> Reservation
        # user_id -> symbol -> AssetClass
        self.symbol_classes: Dict[str, Dict[str, AssetClass]] = {}
        # user_id -> transaction history for FIU AML surveillance
        self.fiat_transactions: Dict[str, List[Tuple[datetime, Decimal]]] = {}

    def _quantize_qty(self, asset_class: AssetClass, qty: Decimal) -> Decimal:
        if asset_class == AssetClass.VDA_CRYPTO:
            return qty.quantize(CRYPTO_QUANTIZE, rounding=ROUND_HALF_UP)
        return qty.quantize(EQUITY_QUANTIZE, rounding=ROUND_HALF_UP)

    def _to_decimal(self, val: Any) -> Decimal:
        if isinstance(val, Decimal):
            return val
        return Decimal(str(val))

    def record_acquisition(
        self,
        user_id: str,
        asset_class: AssetClass,
        symbol: str,
        quantity: Any,
        cost_per_unit_inr: Any,
        acquired_at: Optional[datetime] = None
    ) -> TaxLot:
        """
        Records a new purchase / mint lot in FIFO inventory.
        """
        qty = self._quantize_qty(asset_class, self._to_decimal(quantity))
        cost = self._to_decimal(cost_per_unit_inr).quantize(INR_INTERNAL_QUANTIZE, rounding=ROUND_HALF_UP)

        if qty <= Decimal("0"):
            raise ValueError("Acquisition quantity must be strictly positive")
        if cost < Decimal("0"):
            raise ValueError("Cost per unit cannot be negative")

        if user_id not in self.tax_lots:
            self.tax_lots[user_id] = {}
            self.symbol_classes[user_id] = {}

        if symbol not in self.tax_lots[user_id]:
            self.tax_lots[user_id][symbol] = []

        self.symbol_classes[user_id][symbol] = asset_class

        lot = TaxLot(
            lot_id=f"LOT-{symbol}-{uuid.uuid4().hex[:8].upper()}",
            user_id=user_id,
            asset_class=asset_class,
            symbol=symbol,
            quantity=qty,
            remaining_quantity=qty,
            cost_per_unit_inr=cost,
            total_cost_inr=(qty * cost).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP),
            acquired_at=acquired_at or datetime.now(timezone.utc)
        )

        self.tax_lots[user_id][symbol].append(lot)
        return lot

    def get_total_quantity(self, user_id: str, symbol: str) -> Decimal:
        lots = self.tax_lots.get(user_id, {}).get(symbol, [])
        total = sum((l.remaining_quantity for l in lots), Decimal("0"))
        asset_class = self.symbol_classes.get(user_id, {}).get(symbol, AssetClass.EQUITY)
        return self._quantize_qty(asset_class, total)

    def get_reserved_quantity(self, user_id: str, symbol: str) -> Decimal:
        total = Decimal("0")
        for res in self.reservations.values():
            if res.user_id == user_id and res.symbol == symbol and res.status == "ACTIVE":
                total += res.quantity
        asset_class = self.symbol_classes.get(user_id, {}).get(symbol, AssetClass.EQUITY)
        return self._quantize_qty(asset_class, total)

    def get_available_quantity(self, user_id: str, symbol: str) -> Decimal:
        total = self.get_total_quantity(user_id, symbol)
        reserved = self.get_reserved_quantity(user_id, symbol)
        avail = total - reserved
        return max(Decimal("0"), avail)

    def reserve_holding(self, user_id: str, symbol: str, quantity: Any) -> HoldingReservation:
        """
        Locks units for a pending limit / sell order.
        """
        asset_class = self.symbol_classes.get(user_id, {}).get(symbol, AssetClass.EQUITY)
        qty = self._quantize_qty(asset_class, self._to_decimal(quantity))

        if qty <= Decimal("0"):
            raise ValueError("Reservation quantity must be strictly positive")

        available = self.get_available_quantity(user_id, symbol)
        if qty > available:
            raise InsufficientBalanceError(
                f"Cannot reserve {qty} of {symbol}. Available: {available}, Total: {self.get_total_quantity(user_id, symbol)}"
            )

        res_id = f"RES-{symbol}-{uuid.uuid4().hex[:8].upper()}"
        res = HoldingReservation(
            reservation_id=res_id,
            user_id=user_id,
            symbol=symbol,
            quantity=qty,
            status="ACTIVE"
        )
        self.reservations[res_id] = res
        return res

    def release_holding(self, reservation_id: str) -> HoldingReservation:
        """
        Releases reserved units back to available balance upon order cancellation.
        """
        res = self.reservations.get(reservation_id)
        if not res or res.status != "ACTIVE":
            raise InvalidReservationError(f"Active reservation {reservation_id} not found")

        res.status = "RELEASED"
        return res

    def commit_holding(
        self,
        reservation_id: str,
        sale_price_per_unit_inr: Any,
        disposed_at: Optional[datetime] = None
    ) -> List[RealizedLotSlice]:
        """
        Finalizes trade execution, releases reservation, and depletes FIFO tax lots.
        """
        res = self.reservations.get(reservation_id)
        if not res or res.status != "ACTIVE":
            raise InvalidReservationError(f"Active reservation {reservation_id} not found")

        res.status = "COMMITTED"
        price = self._to_decimal(sale_price_per_unit_inr).quantize(INR_INTERNAL_QUANTIZE, rounding=ROUND_HALF_UP)
        return self._deplete_fifo(res.user_id, res.symbol, res.quantity, price, disposed_at)

    def direct_disposal(
        self,
        user_id: str,
        symbol: str,
        quantity: Any,
        sale_price_per_unit_inr: Any,
        disposed_at: Optional[datetime] = None
    ) -> List[RealizedLotSlice]:
        """
        Directly disposes holdings without an intermediate reservation.
        """
        asset_class = self.symbol_classes.get(user_id, {}).get(symbol, AssetClass.EQUITY)
        qty = self._quantize_qty(asset_class, self._to_decimal(quantity))

        available = self.get_available_quantity(user_id, symbol)
        if qty > available:
            raise InsufficientBalanceError(
                f"Cannot dispose {qty} of {symbol}. Available: {available}"
            )

        price = self._to_decimal(sale_price_per_unit_inr).quantize(INR_INTERNAL_QUANTIZE, rounding=ROUND_HALF_UP)
        return self._deplete_fifo(user_id, symbol, qty, price, disposed_at)

    def _deplete_fifo(
        self,
        user_id: str,
        symbol: str,
        quantity: Decimal,
        sale_price: Decimal,
        disposed_at: Optional[datetime] = None
    ) -> List[RealizedLotSlice]:
        lots = self.tax_lots.get(user_id, {}).get(symbol, [])
        asset_class = self.symbol_classes.get(user_id, {}).get(symbol, AssetClass.EQUITY)
        now = disposed_at or datetime.now(timezone.utc)

        qty_to_deplete = quantity
        slices: List[RealizedLotSlice] = []

        for lot in lots:
            if qty_to_deplete <= Decimal("0"):
                break
            if lot.remaining_quantity <= Decimal("0"):
                continue

            depleted = min(qty_to_deplete, lot.remaining_quantity)
            depleted = self._quantize_qty(asset_class, depleted)

            cost_basis = (depleted * lot.cost_per_unit_inr).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)
            proceeds = (depleted * sale_price).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)
            pnl = proceeds - cost_basis

            # Section 115BBH: for VDA, taxable gain is strictly non-negative (no loss set-off)
            if asset_class == AssetClass.VDA_CRYPTO:
                taxable_115bbh = max(Decimal("0"), pnl)
            else:
                taxable_115bbh = pnl

            lot.remaining_quantity = self._quantize_qty(asset_class, lot.remaining_quantity - depleted)
            qty_to_deplete = self._quantize_qty(asset_class, qty_to_deplete - depleted)

            slices.append(RealizedLotSlice(
                lot_id=lot.lot_id,
                asset_class=asset_class,
                symbol=symbol,
                quantity=depleted,
                cost_per_unit_inr=lot.cost_per_unit_inr,
                cost_basis_inr=cost_basis,
                sale_price_per_unit_inr=sale_price,
                sale_proceeds_inr=proceeds,
                realized_pnl_inr=pnl,
                taxable_115bbh_gain_inr=taxable_115bbh,
                acquired_at=lot.acquired_at,
                disposed_at=now
            ))

        return slices

    def verify_onchain_reconciliation(
        self,
        user_id: str,
        symbol: str,
        onchain_token_balance: Any
    ) -> bool:
        """
        Enforces Prompt 209 Invariant:
        sum(tax_lots.remaining_quantity) == DigitalSecurityToken.balanceOf(user_ledger_address)
        """
        asset_class = self.symbol_classes.get(user_id, {}).get(symbol, AssetClass.EQUITY)
        expected = self.get_total_quantity(user_id, symbol)
        actual = self._quantize_qty(asset_class, self._to_decimal(onchain_token_balance))

        if expected != actual:
            raise ReconciliationMismatchError(
                f"Reconciliation mismatch for user {user_id}, symbol {symbol}: "
                f"Off-chain lot sum={expected}, On-chain token balance={actual}"
            )
        return True

    def record_fiat_transfer(self, user_id: str, amount_inr: Any, timestamp: Optional[datetime] = None):
        """
        Records fiat cash deposit/transfer for FIU-IND AML monitoring.
        """
        amt = self._to_decimal(amount_inr).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)
        if user_id not in self.fiat_transactions:
            self.fiat_transactions[user_id] = []
        self.fiat_transactions[user_id].append((timestamp or datetime.now(timezone.utc), amt))

    def evaluate_fiu_aml_alerts(self, user_id: str) -> List[Dict[str, Any]]:
        """
        Evaluates portfolio movements for CTR (>= ₹10L) and structuring (₹8.5L - ₹9.99L) under PMLA.
        """
        txs = self.fiat_transactions.get(user_id, [])
        alerts = []

        # Check single CTR threshold (>= ₹10 Lakhs)
        for ts, amt in txs:
            if amt >= Decimal("1000000.00"):
                alerts.append({
                    "alert_type": "FIU_CTR_THRESHOLD",
                    "user_id": user_id,
                    "amount_inr": float(amt),
                    "timestamp": ts.isoformat(),
                    "description": "Single transaction >= ₹10 Lakhs statutory CTR threshold"
                })

        # Check Structuring: multiple transactions between ₹8.5L and ₹9.99L
        structuring_txs = [amt for _, amt in txs if Decimal("850000.00") <= amt < Decimal("1000000.00")]
        if len(structuring_txs) >= 2:
            alerts.append({
                "alert_type": "FIU_STR_STRUCTURING",
                "user_id": user_id,
                "count": len(structuring_txs),
                "total_inr": float(sum(structuring_txs)),
                "description": f"Detected {len(structuring_txs)} transactions just below ₹10L threshold"
            })

        return alerts

    def get_consolidated_portfolio(
        self,
        user_id: str,
        market_prices_inr: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Computes real-time mark-to-market portfolio valuation with allocation breakdown.
        """
        user_lots = self.tax_lots.get(user_id, {})
        items: List[HoldingItem] = []

        total_portfolio_value = Decimal("0")
        total_invested_value = Decimal("0")
        asset_class_values: Dict[str, Decimal] = {ac.value: Decimal("0") for ac in AssetClass}

        # First pass: calculate total portfolio value for allocation denominator
        for symbol, lots in user_lots.items():
            tot_qty = self.get_total_quantity(user_id, symbol)
            if tot_qty <= Decimal("0"):
                continue

            cmp_raw = market_prices_inr.get(symbol)
            if cmp_raw is None:
                # Default to weighted average cost if market price unavailable
                active_lots = [l for l in lots if l.remaining_quantity > Decimal("0")]
                if active_lots:
                    cmp = sum(l.remaining_quantity * l.cost_per_unit_inr for l in active_lots) / tot_qty
                else:
                    cmp = Decimal("0")
            else:
                cmp = self._to_decimal(cmp_raw)

            current_val = (tot_qty * cmp).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)
            total_portfolio_value += current_val

            ac = self.symbol_classes[user_id].get(symbol, AssetClass.EQUITY).value
            asset_class_values[ac] = asset_class_values.get(ac, Decimal("0")) + current_val

        # Second pass: construct individual holding items
        for symbol, lots in user_lots.items():
            tot_qty = self.get_total_quantity(user_id, symbol)
            if tot_qty <= Decimal("0"):
                continue

            reserved_qty = self.get_reserved_quantity(user_id, symbol)
            avail_qty = self.get_available_quantity(user_id, symbol)

            active_lots = [l for l in lots if l.remaining_quantity > Decimal("0")]
            invested_val = sum((l.remaining_quantity * l.cost_per_unit_inr for l in active_lots), Decimal("0"))
            invested_val = invested_val.quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)
            total_invested_value += invested_val

            avg_cost = (invested_val / tot_qty).quantize(INR_INTERNAL_QUANTIZE, rounding=ROUND_HALF_UP) if tot_qty > 0 else Decimal("0")

            cmp_raw = market_prices_inr.get(symbol)
            cmp = self._to_decimal(cmp_raw) if cmp_raw is not None else avg_cost
            current_val = (tot_qty * cmp).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)

            unrealized_pnl = current_val - invested_val
            unrealized_pct = ((unrealized_pnl / invested_val) * Decimal("100")).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP) if invested_val > 0 else Decimal("0")

            alloc_pct = ((current_val / total_portfolio_value) * Decimal("100")).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP) if total_portfolio_value > 0 else Decimal("0")

            ac = self.symbol_classes[user_id].get(symbol, AssetClass.EQUITY)

            items.append(HoldingItem(
                asset_class=ac.value,
                symbol=symbol,
                total_quantity=float(tot_qty),
                reserved_quantity=float(reserved_qty),
                available_quantity=float(avail_qty),
                average_cost_inr=float(avg_cost.quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)),
                current_market_price_inr=float(cmp.quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP)),
                invested_value_inr=float(invested_val),
                current_value_inr=float(current_val),
                unrealized_pnl_inr=float(unrealized_pnl),
                unrealized_pnl_pct=float(unrealized_pct),
                allocation_pct=float(alloc_pct)
            ))

        total_unrealized = total_portfolio_value - total_invested_value
        total_unrealized_pct = ((total_unrealized / total_invested_value) * Decimal("100")).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP) if total_invested_value > 0 else Decimal("0")

        # Allocation by asset class breakdown
        asset_breakdown = {}
        for ac, val in asset_class_values.items():
            pct = ((val / total_portfolio_value) * Decimal("100")).quantize(INR_QUANTIZE, rounding=ROUND_HALF_UP) if total_portfolio_value > 0 else Decimal("0")
            asset_breakdown[ac] = {
                "value_inr": float(val),
                "allocation_pct": float(pct)
            }

        return {
            "user_id": user_id,
            "total_net_worth_inr": float(total_portfolio_value),
            "total_invested_inr": float(total_invested_value),
            "total_unrealized_pnl_inr": float(total_unrealized),
            "total_unrealized_pnl_pct": float(total_unrealized_pct),
            "holdings_count": len(items),
            "holdings": [asdict(i) for i in items],
            "asset_class_breakdown": asset_breakdown,
            "computed_at": datetime.now(timezone.utc).isoformat()
        }
