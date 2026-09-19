"""
Tiered KYC Verification Engine (Prompt 076)
Tier 1: Aadhaar/PAN Instant OTP (₹1 Lakh annual limit)
Tier 2: CKYC & Bank Penny Drop (₹25 Lakh limit)
Tier 3: In-Person Verification / V-CIP & Net Worth Certificate (Unlimited institutional tier)
"""

from typing import Dict, Any

class TieredKYCEngine:
    def __init__(self):
        self.tiers: Dict[str, int] = {} # user -> tier (0, 1, 2, 3)

    def verify_tier1_instant(self, user_id: str, pan: str, aadhaar_otp_verified: bool) -> bool:
        if aadhaar_otp_verified and len(pan) == 10:
            self.tiers[user_id] = 1
            return True
        return False

    def upgrade_to_tier2_ckyc(self, user_id: str, ckyc_number: str, penny_drop_success: bool) -> bool:
        if self.tiers.get(user_id, 0) >= 1 and len(ckyc_number) == 14 and penny_drop_success:
            self.tiers[user_id] = 2
            return True
        return False

    def upgrade_to_tier3_vcip(self, user_id: str, vcip_recording_ref: str, net_worth_cert_ca: bool) -> bool:
        if self.tiers.get(user_id, 0) >= 2 and vcip_recording_ref and net_worth_cert_ca:
            self.tiers[user_id] = 3
            return True
        return False

    def get_withdrawal_limits(self, user_id: str) -> Dict[str, Any]:
        tier = self.tiers.get(user_id, 0)
        limits = {
            0: {"tier": 0, "daily_inr": 0, "annual_inr": 0, "status": "UNVERIFIED"},
            1: {"tier": 1, "daily_inr": 25000, "annual_inr": 100000, "status": "BASIC_OTP"},
            2: {"tier": 2, "daily_inr": 500000, "annual_inr": 2500000, "status": "FULL_CKYC"},
            3: {"tier": 3, "daily_inr": 50000000, "annual_inr": 500000000, "status": "INSTITUTIONAL_VCIP"},
        }
        return limits[tier]
