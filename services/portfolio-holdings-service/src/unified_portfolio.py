"""
Unified Portfolio Holdings Service (Prompt 041)
Consolidates NSDL/CDSL equity holdings and on-chain crypto holdings into a unified net worth ledger.
"""

from typing import Dict, List, Any
from dataclasses import dataclass, asdict
from datetime import datetime

@dataclass
class HoldingItem:
    asset_class: str # EQUITY or CRYPTO
    symbol: str
    quantity: float
    average_cost_inr: float
    current_market_price_inr: float
    unrealized_pnl_inr: float
    allocation_pct: float

class UnifiedPortfolioEngine:
    def __init__(self):
        self.portfolios: Dict[str, Dict[str, Dict[str, Any]]] = {}

    def update_holding(self, user_id: str, asset_class: str, symbol: str, quantity: float, avg_cost: float):
        if user_id not in self.portfolios:
            self.portfolios[user_id] = {}
        
        self.portfolios[user_id][symbol] = {
            "asset_class": asset_class,
            "quantity": quantity,
            "avg_cost": avg_cost,
            "updated_at": datetime.utcnow().isoformat()
        }

    def get_consolidated_portfolio(self, user_id: str, market_prices_inr: Dict[str, float]) -> Dict[str, Any]:
        holdings = self.portfolios.get(user_id, {})
        items: List[HoldingItem] = []
        total_portfolio_value = 0.0

        for symbol, data in holdings.items():
            qty = data["quantity"]
            cmp = market_prices_inr.get(symbol, data["avg_cost"])
            val = qty * cmp
            total_portfolio_value += val

        for symbol, data in holdings.items():
            qty = data["quantity"]
            cost = data["avg_cost"]
            cmp = market_prices_inr.get(symbol, cost)
            val = qty * cmp
            unrealized = (cmp - cost) * qty
            alloc = (val / total_portfolio_value * 100.0) if total_portfolio_value > 0 else 0.0

            items.append(HoldingItem(
                asset_class=data["asset_class"],
                symbol=symbol,
                quantity=qty,
                average_cost_inr=round(cost, 2),
                current_market_price_inr=round(cmp, 2),
                unrealized_pnl_inr=round(unrealized, 2),
                allocation_pct=round(alloc, 2)
            ))

        return {
            "user_id": user_id,
            "total_net_worth_inr": round(total_portfolio_value, 2),
            "holdings_count": len(items),
            "holdings": [asdict(i) for i in items],
            "computed_at": datetime.utcnow().isoformat()
        }

if __name__ == "__main__":
    engine = UnifiedPortfolioEngine()
    engine.update_holding("usr-1", "EQUITY", "RELIANCE.EQ", quantity=50, avg_cost=2950.0)
    engine.update_holding("usr-1", "CRYPTO", "BTC", quantity=0.15, avg_cost=5800000.0)
    
    report = engine.get_consolidated_portfolio("usr-1", {
        "RELIANCE.EQ": 3020.0,
        "BTC": 6050000.0
    })
    print("Unified Portfolio:", report)
