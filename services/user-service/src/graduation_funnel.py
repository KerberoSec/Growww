"""
Demo-to-Real Trading Graduation Funnel Service (Prompt 037)
Evaluates demo paper trading performance, risk profile, and suitability before unlocking real money trading.
"""

from typing import Dict, Any
from dataclasses import dataclass
from datetime import datetime

@dataclass
class GraduationCriteria:
    min_demo_trades: int = 10
    min_win_rate_pct: float = 40.0
    risk_assessment_completed: bool = True
    kyc_status_approved: bool = True

class GraduationFunnelService:
    def __init__(self, criteria: GraduationCriteria = None):
        self.criteria = criteria or GraduationCriteria()
        self.user_records: Dict[str, Dict[str, Any]] = {}

    def register_trader_stats(self, user_id: str, demo_trades: int, win_rate: float, pnl_usd: float):
        self.user_records[user_id] = {
            "demo_trades": demo_trades,
            "win_rate": win_rate,
            "pnl_usd": pnl_usd,
            "risk_ack": False,
            "graduated": False,
            "updated_at": datetime.utcnow().isoformat()
        }

    def submit_risk_acknowledgement(self, user_id: str) -> bool:
        if user_id in self.user_records:
            self.user_records[user_id]["risk_ack"] = True
            return True
        return False

    def evaluate_graduation(self, user_id: str, kyc_approved: bool) -> Dict[str, Any]:
        record = self.user_records.get(user_id)
        if not record:
            return {"eligible": False, "reason": "No demo trading history found"}

        trades_ok = record["demo_trades"] >= self.criteria.min_demo_trades
        win_rate_ok = record["win_rate"] >= self.criteria.min_win_rate_pct
        risk_ok = record["risk_ack"]
        kyc_ok = kyc_approved

        eligible = trades_ok and risk_ok and kyc_ok

        if eligible:
            record["graduated"] = True

        return {
            "eligible": eligible,
            "checks": {
                "trades_threshold": f"{record['demo_trades']}/{self.criteria.min_demo_trades} ({trades_ok})",
                "win_rate_threshold": f"{record['win_rate']:.1f}%/{self.criteria.min_win_rate_pct}% ({win_rate_ok})",
                "risk_acknowledgement": risk_ok,
                "kyc_compliance": kyc_ok
            },
            "status": "GRADUATED_TO_REAL" if eligible else "IN_PROGRESS"
        }

if __name__ == "__main__":
    funnel = GraduationFunnelService()
    funnel.register_trader_stats("usr-123", demo_trades=15, win_rate=55.0, pnl_usd=1250.0)
    funnel.submit_risk_acknowledgement("usr-123")
    result = funnel.evaluate_graduation("usr-123", kyc_approved=True)
    print("Graduation Result:", result)
