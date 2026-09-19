"""
Anti-Phishing Code & Email/SMS Verification Shield (Prompt 073)
Injects secret user-configured anti-phishing codes into all official notifications to prevent spoofing.
"""

import hmac
import hashlib
from typing import Dict, Optional

class AntiPhishingShield:
    def __init__(self):
        self.user_phishing_codes: Dict[str, str] = {}

    def set_anti_phishing_code(self, user_id: str, code: str):
        if len(code) < 4 or len(code) > 20:
            raise ValueError("Anti-phishing code must be between 4 and 20 characters")
        self.user_phishing_codes[user_id] = code

    def stamp_outbound_email_payload(self, user_id: str, email_body: str) -> str:
        code = self.user_phishing_codes.get(user_id, "DEFAULT-SECURE-KEY")
        banner = f"\n[GROWWW OFFICIAL SECURITY BANNER: Anti-Phishing Code: {code}]\n\n"
        return banner + email_body

    def verify_inbound_code(self, user_id: str, presented_code: str) -> bool:
        stored = self.user_phishing_codes.get(user_id)
        if not stored:
            return False
        return hmac.compare_digest(stored, presented_code)
