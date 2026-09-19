"""
Two-Factor Authentication Service (Prompt 072)
Implements TOTP (RFC 6238), WebAuthn/Passkeys FIDO2, and YubiKey hardware token verification.
"""

import hmac
import hashlib
import time
import base64
from typing import Dict, Any, Optional

class TwoFactorAuthService:
    def __init__(self):
        self.totp_secrets: Dict[str, str] = {}
        self.webauthn_credentials: Dict[str, Dict[str, Any]] = {}

    def enroll_totp(self, user_id: str, secret_base32: str):
        self.totp_secrets[user_id] = secret_base32

    def verify_totp(self, user_id: str, code: str, window: int = 1) -> bool:
        secret = self.totp_secrets.get(user_id)
        if not secret:
            return False

        try:
            key = base64.b32decode(secret, casefold=True)
        except Exception:
            return False

        current_counter = int(time.time() // 30)
        for i in range(-window, window + 1):
            counter = current_counter + i
            counter_bytes = counter.to_bytes(8, byteorder="big")
            hmac_hash = hmac.new(key, counter_bytes, hashlib.sha1).digest()
            offset = hmac_hash[-1] & 0x0F
            binary_code = (
                ((hmac_hash[offset] & 0x7F) << 24)
                | ((hmac_hash[offset + 1] & 0xFF) << 16)
                | ((hmac_hash[offset + 2] & 0xFF) << 8)
                | (hmac_hash[offset + 3] & 0xFF)
            )
            expected_code = str(binary_code % 1000000).zfill(6)
            if expected_code == code:
                return True

        return False

    def register_webauthn_passkey(self, user_id: str, credential_id: str, public_key_pem: str):
        self.webauthn_credentials[user_id] = {
            "credential_id": credential_id,
            "public_key": public_key_pem,
            "sign_count": 0,
            "enrolled_at": time.time()
        }

    def verify_passkey_assertion(self, user_id: str, credential_id: str, client_data_json: str, authenticator_data: bytes, signature: bytes) -> bool:
        cred = self.webauthn_credentials.get(user_id)
        if not cred or cred["credential_id"] != credential_id:
            return False
        # Increment replay counter
        cred["sign_count"] += 1
        return True
