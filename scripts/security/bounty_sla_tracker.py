#!/usr/bin/env python3
"""
Automated Vulnerability Triage, CVSS Scoring, and SLA Tracker (Prompt 705)
Manages bug bounty report lifecycle, validates scopes, and computes remediation SLAs.
"""

from datetime import datetime, timezone
from typing import Dict, Any, Optional

SLA_HOURS = {
    "CRITICAL": 24,
    "HIGH": 168,      # 7 days
    "MEDIUM": 720,    # 30 days
    "LOW": 2160       # 90 days
}

BOUNTY_BANDS_INR = {
    "CRITICAL": (500000, 1000000),
    "HIGH": (150000, 300000),
    "MEDIUM": (30000, 75000),
    "LOW": (5000, 15000),
}

AUTHORIZED_ASSET_TYPES = {"API", "WEB", "MOBILE", "SMART_CONTRACT"}

class BountySLATracker:
    def __init__(self, manifest_config: Optional[Dict[str, Any]] = None):
        self.config = manifest_config or {}

    def classify_severity(self, cvss_score: float) -> str:
        if not (0.0 <= cvss_score <= 10.0):
            raise ValueError(f"Invalid CVSS score: {cvss_score}. Must be between 0.0 and 10.0.")
        if cvss_score >= 9.0:
            return "CRITICAL"
        elif cvss_score >= 7.0:
            return "HIGH"
        elif cvss_score >= 4.0:
            return "MEDIUM"
        elif cvss_score > 0.0:
            return "LOW"
        return "INFORMATIONAL"

    def calculate_sla_expiry(self, submitted_at_iso: str, severity: str) -> str:
        submitted_dt = datetime.fromisoformat(submitted_at_iso.replace("Z", "+00:00"))
        sla_hours = SLA_HOURS.get(severity, 2160)
        from datetime import timedelta
        expiry_dt = submitted_dt + timedelta(hours=sla_hours)
        return expiry_dt.isoformat()

    def evaluate_sla_status(self, submitted_at_iso: str, severity: str, current_time_iso: Optional[str] = None) -> Dict[str, Any]:
        submitted_dt = datetime.fromisoformat(submitted_at_iso.replace("Z", "+00:00"))
        now_dt = (
            datetime.fromisoformat(current_time_iso.replace("Z", "+00:00"))
            if current_time_iso
            else datetime.now(timezone.utc)
        )
        sla_hours = SLA_HOURS.get(severity, 2160)
        from datetime import timedelta
        total_sla_seconds = sla_hours * 3600
        elapsed_seconds = (now_dt - submitted_dt).total_seconds()
        remaining_seconds = total_sla_seconds - elapsed_seconds

        if remaining_seconds <= 0:
            status = "SLA_BREACHED"
        elif remaining_seconds <= (total_sla_seconds * 0.25):
            status = "URGENT_ESCALATION"
        else:
            status = "ON_TRACK"

        return {
            "severity": severity,
            "sla_hours": sla_hours,
            "elapsed_hours": round(elapsed_seconds / 3600, 2),
            "remaining_hours": max(0.0, round(remaining_seconds / 3600, 2)),
            "status": status,
            "escalate_pagerduty": status in ("URGENT_ESCALATION", "SLA_BREACHED") and severity == "CRITICAL"
        }

    def compute_bounty_recommendation(self, severity: str, quality_multiplier: float = 1.0) -> Dict[str, Any]:
        if severity not in BOUNTY_BANDS_INR:
            return {"eligible": False, "recommended_inr": 0}
        
        min_inr, max_inr = BOUNTY_BANDS_INR[severity]
        multiplier = max(0.0, min(1.0, quality_multiplier))
        calculated = min_inr + (max_inr - min_inr) * multiplier
        return {
            "eligible": True,
            "currency": "INR",
            "min_band_inr": min_inr,
            "max_band_inr": max_inr,
            "recommended_inr": int(calculated)
        }

if __name__ == "__main__":
    tracker = BountySLATracker()
    sev = tracker.classify_severity(9.8)
    res = tracker.evaluate_sla_status("2026-09-20T00:00:00Z", sev)
    print("Vulnerability Triage Evaluation:", res)
