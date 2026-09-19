"""
Growww User Service - Profile, Session, KYC Tiering, and Security
"""

import hashlib
import hmac
import os
import time
import secrets
from typing import Dict, Optional, List, Tuple
from dataclasses import dataclass, field

@dataclass
class UserSession:
    session_id: str
    user_id: str
    created_at: float
    expires_at: float
    ip_address: str
    user_agent: str
    is_active: bool = True

@dataclass
class UserProfile:
    user_id: str
    email: str
    phone: str
    password_hash: str
    salt: str
    kyc_tier: int = 0  # 0: Unverified, 1: Basic, 2: Standard, 3: High-Net-Worth
    pan: Optional[str] = None
    aadhaar_vault_token: Optional[str] = None
    anti_phishing_phrase: Optional[str] = None
    totp_secret: Optional[str] = None
    is_totp_enabled: bool = False
    is_suspended: bool = False
    created_at: float = field(default_factory=time.time)
    sessions: Dict[str, UserSession] = field(default_factory=dict)
    login_attempts: int = 0
    locked_until: float = 0.0

class UserService:
    def __init__(self):
        self.users_by_id: Dict[str, UserProfile] = {}
        self.users_by_email: Dict[str, str] = {}  # email -> user_id
        self.sessions: Dict[str, UserSession] = {} # session_id -> session

    @staticmethod
    def _hash_password(password: str, salt: str) -> str:
        return hashlib.pbkdf2_hmac("sha256", password.encode(), salt.encode(), 100_000).hex()

    def register_user(self, email: str, phone: str, password: str) -> UserProfile:
        email_clean = email.strip().lower()
        if email_clean in self.users_by_email:
            raise ValueError("email already registered")
        if len(password) < 8:
            raise ValueError("password must be at least 8 characters")

        user_id = f"usr_{secrets.token_hex(8)}"
        salt = secrets.token_hex(16)
        pwd_hash = self._hash_password(password, salt)

        user = UserProfile(
            user_id=user_id,
            email=email_clean,
            phone=phone.strip(),
            password_hash=pwd_hash,
            salt=salt,
        )
        self.users_by_id[user_id] = user
        self.users_by_email[email_clean] = user_id
        return user

    def authenticate(self, email: str, password: str, ip: str, user_agent: str) -> Tuple[UserProfile, UserSession]:
        email_clean = email.strip().lower()
        user_id = self.users_by_email.get(email_clean)
        if not user_id:
            raise ValueError("invalid credentials")

        user = self.users_by_id[user_id]

        if user.is_suspended:
            raise PermissionError("account is suspended")

        if time.time() < user.locked_until:
            raise PermissionError(f"account locked until {user.locked_until}")

        expected_hash = self._hash_password(password, user.salt)
        if not hmac.compare_digest(expected_hash, user.password_hash):
            user.login_attempts += 1
            if user.login_attempts >= 5:
                user.locked_until = time.time() + 900  # 15 minutes lockout
            raise ValueError("invalid credentials")

        # Reset login attempts upon successful auth
        user.login_attempts = 0

        # Create session
        session_id = f"ses_{secrets.token_hex(16)}"
        now = time.time()
        session = UserSession(
            session_id=session_id,
            user_id=user_id,
            created_at=now,
            expires_at=now + 86400, # 24h
            ip_address=ip,
            user_agent=user_agent,
            is_active=True,
        )
        self.sessions[session_id] = session
        user.sessions[session_id] = session
        return user, session

    def validate_session(self, session_id: str) -> Optional[UserProfile]:
        session = self.sessions.get(session_id)
        if not session or not session.is_active:
            return None
        if time.time() > session.expires_at:
            session.is_active = False
            return None
        return self.users_by_id.get(session.user_id)

    def logout_session(self, session_id: str):
        session = self.sessions.get(session_id)
        if session:
            session.is_active = False

    def upgrade_kyc_tier(self, user_id: str, new_tier: int, pan: Optional[str] = None, aadhaar_token: Optional[str] = None):
        user = self.users_by_id.get(user_id)
        if not user:
            raise KeyError("user not found")
        if new_tier not in (0, 1, 2, 3):
            raise ValueError("invalid tier")
        if new_tier > user.kyc_tier + 1:
            raise ValueError("cannot skip intermediate KYC tiers")

        user.kyc_tier = new_tier
        if pan:
            user.pan = pan
        if aadhaar_token:
            user.aadhaar_vault_token = aadhaar_token

    def set_anti_phishing_phrase(self, user_id: str, phrase: str):
        user = self.users_by_id.get(user_id)
        if not user:
            raise KeyError("user not found")
        if len(phrase.strip()) < 4:
            raise ValueError("phrase must be at least 4 characters")
        user.anti_phishing_phrase = phrase.strip()
