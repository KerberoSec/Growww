"""
Growww / NBSE Sovereign Exchange - Compliance Engine
Covers Prompts 000 - 049:
- Indian KYC/AML Onboarding (PAN format & entity validation, Aadhaar 8-digit masking, CKYCR, Bank Penny Drop Jaro-Winkler >= 0.85)
- PEP & Sanctions Screening (OFAC SDN, UN 1267, MHA UAPA, FATF Blacklist, Account Freezing, EDD triggers)
- Tiered KYC Limits & Upgrades (Tier 0..3 quotas, Re-KYC periodicity)
- FIU-IND AML Surveillance (Cash structuring evasion, CTR >= ₹10L, FATF IVMS-101 Travel Rule > ₹50,000)
- Regulatory Sandbox Guardrails (SEBI 10k users/₹50k portfolio, IFSCA 5k users/$10k portfolio, Universal 0.00% Zero Fee)
- Zero-PII Cryptographic Identity Commitment Generation for Hyperledger Besu ComplianceRegistry.sol
"""

import hashlib
import re
import secrets
from datetime import datetime, timedelta, timezone
from enum import Enum
from typing import Dict, List, Optional, Tuple, Any


class KYCTier(Enum):
    TIER_0_UNVERIFIED = 0
    TIER_1_BASIC_OTP = 1
    TIER_2_FULL_CKYC = 2
    TIER_3_INSTITUTIONAL_VCIP = 3


class RiskTier(Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"


class PEPClassification(Enum):
    NONE = "NONE"
    DOMESTIC = "DOMESTIC_PEP"
    FOREIGN = "FOREIGN_PEP"
    CLOSE_ASSOCIATE = "CLOSE_ASSOCIATE"


class SanctionsSource(Enum):
    NONE = "NONE"
    OFAC_SDN = "OFAC_SDN"
    UN_CONSOLIDATED = "UN_CONSOLIDATED"
    EU_FINANCIAL = "EU_FINANCIAL"
    UK_HM_TREASURY = "UK_HM_TREASURY"
    MHA_UAPA_INDIA = "MHA_UAPA_INDIA"


class ComplianceEngine:
    # 4th character in PAN defines the entity type under Indian Income Tax Department rules
    PAN_ENTITY_TYPES = {
        'P': "INDIVIDUAL",
        'C': "COMPANY",
        'H': "HINDU_UNDIVIDED_FAMILY",
        'F': "PARTNERSHIP_FIRM",
        'A': "ASSOCIATION_OF_PERSONS",
        'T': "TRUST",
        'B': "BODY_OF_INDIVIDUALS",
        'L': "LOCAL_AUTHORITY",
        'J': "ARTIFICIAL_JURIDICAL_PERSON",
        'G': "GOVERNMENT",
    }

    FATF_BLACKLIST = {"PRK", "IRN", "MMR"}

    KNOWN_SANCTIONS = [
        {"name": "AL-QAIDA", "source": SanctionsSource.UN_CONSOLIDATED, "type": "TERRORIST_ORG"},
        {"name": "LASHKAR-E-TAIBA", "source": SanctionsSource.MHA_UAPA_INDIA, "type": "BANNED_ORG"},
        {"name": "JAISH-E-MOHAMMED", "source": SanctionsSource.MHA_UAPA_INDIA, "type": "BANNED_ORG"},
        {"name": "DAWOOD IBRAHIM KASKAR", "source": SanctionsSource.UN_CONSOLIDATED, "type": "INDIVIDUAL"},
        {"name": "HAFIZ MUHAMMAD SAEED", "source": SanctionsSource.UN_CONSOLIDATED, "type": "INDIVIDUAL"},
        {"name": "OFAC BLOCKED ENTITY LTD", "source": SanctionsSource.OFAC_SDN, "type": "ENTITY"},
    ]

    KNOWN_PEPS = [
        {"name": "ARUN KUMAR MINISTERIAL", "category": PEPClassification.DOMESTIC, "office": "Union Minister"},
        {"name": "VIKRAMADITYA SINGH POLITICIAN", "category": PEPClassification.DOMESTIC, "office": "State Politician"},
        {"name": "FOREIGN DIPLOMAT JOHN DOE", "category": PEPClassification.FOREIGN, "office": "Ambassador"},
    ]

    TIER_LIMITS = {
        KYCTier.TIER_0_UNVERIFIED: {"daily_inr": 0.0, "annual_inr": 0.0, "desc": "UNVERIFIED"},
        KYCTier.TIER_1_BASIC_OTP: {"daily_inr": 25000.0, "annual_inr": 100000.0, "desc": "BASIC_OTP"},
        KYCTier.TIER_2_FULL_CKYC: {"daily_inr": 500000.0, "annual_inr": 2500000.0, "desc": "FULL_CKYC"},
        KYCTier.TIER_3_INSTITUTIONAL_VCIP: {"daily_inr": 50000000.0, "annual_inr": 500000000.0, "desc": "INSTITUTIONAL_VCIP"},
    }

    RE_KYC_PERIODS = {
        RiskTier.LOW: timedelta(days=365 * 10),     # 10 years
        RiskTier.MEDIUM: timedelta(days=365 * 8),   # 8 years
        RiskTier.HIGH: timedelta(days=365 * 2),     # 2 years
    }

    def __init__(self, secret_salt: str = "GROWWW_SYSTEM_SALT_DEFAULT"):
        self.salt = secret_salt
        self.investors: Dict[str, Dict[str, Any]] = {}
        self.alerts: List[Dict[str, Any]] = []

    # -------------------------------------------------------------
    # 1. PAN Validation (Prompt 004)
    # -------------------------------------------------------------
    def validate_pan(self, pan: str, is_individual: bool = True, aadhaar_linked: bool = True, status: str = "OPERATIVE") -> Tuple[bool, str]:
        pan = pan.strip().upper()
        if not re.match(r"^[A-Z]{5}[0-9]{4}[A-Z]$", pan):
            return False, "PAN must follow statutory 10-character structure (5 letters, 4 digits, 1 letter)"

        entity_char = pan[3]
        if entity_char not in self.PAN_ENTITY_TYPES:
            return False, f"Invalid 4th character '{entity_char}' in PAN: unknown entity classification"

        if status != "OPERATIVE":
            return False, f"PAN status is {status}; must be OPERATIVE under NSDL database"

        if is_individual and not aadhaar_linked:
            return False, "Aadhaar-PAN linking is mandatory for individual domestic investors under Section 139AA"

        return True, self.PAN_ENTITY_TYPES[entity_char]

    def hash_pan(self, pan: str) -> str:
        clean = pan.strip().upper()
        return hashlib.sha256(f"{clean}:{self.salt}".encode("utf-8")).hexdigest()

    # -------------------------------------------------------------
    # 2. Aadhaar Data Vault & 8-Digit Masking (Prompt 004, 008)
    # -------------------------------------------------------------
    def validate_aadhaar_masking(self, aadhaar_input: str) -> Tuple[bool, str]:
        clean = aadhaar_input.strip()
        # Raw 12 digit unmasked Aadhaar violation
        if re.match(r"^[0-9]{4}[-\s]?[0-9]{4}[-\s]?[0-9]{4}$|^[0-9]{12}$", clean):
            return False, "CRITICAL COMPLIANCE VIOLATION: Raw 12-digit Aadhaar detected; UIDAI regulations require 8-digit masking"

        # Valid masked pattern: first 8 masked with X or *, last 4 digits exposed
        if not re.match(r"^(?:[X*]{4}[-\s]?[X*]{4}[-\s]?[0-9]{4}|[X*]{8}[0-9]{4})$", clean):
            return False, "Masked Aadhaar must mask first 8 digits and expose only last 4 digits (e.g. XXXX-XXXX-1234)"

        digits = re.findall(r"[0-9]", clean)
        if len(digits) != 4:
            return False, "Masked Aadhaar must expose exactly 4 digits"

        return True, "".join(digits)

    def generate_vault_token(self) -> str:
        return f"ADV-TOK-{secrets.token_hex(16)}"

    # -------------------------------------------------------------
    # 3. Bank Account Penny Drop & Jaro-Winkler Matching (Prompt 004)
    # -------------------------------------------------------------
    @staticmethod
    def jaro_winkler(s1: str, s2: str) -> float:
        s1 = s1.upper().strip()
        s2 = s2.upper().strip()
        if s1 == s2:
            return 1.0
        if not s1 or not s2:
            return 0.0

        max_dist = max(len(s1), len(s2)) // 2 - 1
        s1_matches = [False] * len(s1)
        s2_matches = [False] * len(s2)

        matches = 0
        for i, c1 in enumerate(s1):
            start = max(0, i - max_dist)
            end = min(i + max_dist + 1, len(s2))
            for j in range(start, end):
                if not s2_matches[j] and c1 == s2[j]:
                    s1_matches[i] = True
                    s2_matches[j] = True
                    matches += 1
                    break

        if matches == 0:
            return 0.0

        transpositions = 0
        k = 0
        for i, c1 in enumerate(s1):
            if s1_matches[i]:
                while not s2_matches[k]:
                    k += 1
                if c1 != s2[k]:
                    transpositions += 1
                k += 1

        jaro = (matches / len(s1) + matches / len(s2) + (matches - transpositions / 2) / matches) / 3.0
        if jaro < 0.7:
            return jaro

        prefix = 0
        for c1, c2 in zip(s1[:4], s2[:4]):
            if c1 == c2:
                prefix += 1
            else:
                break
        return jaro + prefix * 0.1 * (1.0 - jaro)

    def validate_bank_penny_drop(self, ifsc: str, declared_name: str, beneficiary_name: str) -> Tuple[bool, float, str]:
        ifsc = ifsc.strip().upper()
        if not re.match(r"^[A-Z]{4}0[A-Z0-9]{6}$", ifsc):
            return False, 0.0, f"Invalid IFSC code format: {ifsc}"

        # Clean honorifics
        honorifics = {"MR", "MRS", "MS", "DR", "SHRI", "SMT", "SH", "KUMAR"}
        def clean(n: str) -> str:
            tokens = [t for t in re.sub(r"[^A-Z\s]", " ", n.upper()).split() if t not in honorifics]
            return " ".join(tokens)

        d_clean = clean(declared_name)
        b_clean = clean(beneficiary_name)

        # Check token sets for permutation
        d_tokens = set(d_clean.split())
        b_tokens = set(b_clean.split())
        if d_tokens == b_tokens and len(d_tokens) > 0:
            score = 1.0
        else:
            score = self.jaro_winkler(d_clean, b_clean)

        if score < 0.85:
            return False, score, f"Bank penny drop name match failed: score {score:.3f} < 0.85 threshold"

        return True, score, "SUCCESS"

    # -------------------------------------------------------------
    # 4. Sanctions & PEP Screening (Prompt 004, 005)
    # -------------------------------------------------------------
    def screen_sanctions_and_pep(self, name: str, country_iso3: str) -> Dict[str, Any]:
        norm = name.strip().upper()
        country = country_iso3.strip().upper()

        # Check FATF Blacklist
        if country in self.FATF_BLACKLIST:
            return {
                "sanctioned": True,
                "frozen": True,
                "source": SanctionsSource.UN_CONSOLIDATED,
                "pep": False,
                "pep_category": PEPClassification.NONE,
                "edd_required": True,
                "reason": f"Nationality/Country {country} on FATF Blacklist",
            }

        # Check Sanctions lists
        for s in self.KNOWN_SANCTIONS:
            s_name = s["name"]
            if s_name in norm or norm in s_name:
                return {
                    "sanctioned": True,
                    "frozen": True,
                    "source": s["source"],
                    "pep": False,
                    "pep_category": PEPClassification.NONE,
                    "edd_required": True,
                    "reason": f"Matches sanctions list: {s['name']} ({s['source'].value})",
                }

        # Check PEP lists
        for p in self.KNOWN_PEPS:
            p_name = p["name"]
            if p_name in norm or norm in p_name:
                return {
                    "sanctioned": False,
                    "frozen": False,
                    "source": SanctionsSource.NONE,
                    "pep": True,
                    "pep_category": p["category"],
                    "edd_required": True,
                    "reason": f"Politically Exposed Person: {p['office']}",
                }

        return {
            "sanctioned": False,
            "frozen": False,
            "source": SanctionsSource.NONE,
            "pep": False,
            "pep_category": PEPClassification.NONE,
            "edd_required": False,
            "reason": "Clear",
        }

    # -------------------------------------------------------------
    # 5. KYC Tier Limits & Upgrades (Prompt 004, 076)
    # -------------------------------------------------------------
    def check_withdrawal_limit(self, tier: KYCTier, amount_inr: float, daily_spent_inr: float, annual_spent_inr: float) -> Tuple[bool, str]:
        if tier == KYCTier.TIER_0_UNVERIFIED:
            return False, "Unverified user (Tier 0) cannot transact or withdraw"

        limits = self.TIER_LIMITS[tier]
        if daily_spent_inr + amount_inr > limits["daily_inr"]:
            return False, f"Daily limit exceeded: ₹{daily_spent_inr + amount_inr:.2f} > quota ₹{limits['daily_inr']:.2f}"

        if annual_spent_inr + amount_inr > limits["annual_inr"]:
            return False, f"Annual limit exceeded: ₹{annual_spent_inr + amount_inr:.2f} > quota ₹{limits['annual_inr']:.2f}"

        return True, "Within limits"

    # -------------------------------------------------------------
    # 6. FIU-IND AML Surveillance & Travel Rule (Prompt 004, 049)
    # -------------------------------------------------------------
    def monitor_transaction_aml(self, user_id: str, amount_inr: float, recent_txs: List[float], is_crypto: bool = False, travel_rule: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        result = {
            "allowed": True,
            "str_triggered": False,
            "ctr_triggered": False,
            "travel_rule_compliant": True,
            "alerts": [],
        }

        # Structuring check (multiple txs in [8.5L, 10L))
        all_txs = recent_txs + [amount_inr]
        near_threshold = [tx for tx in all_txs if 850000.0 <= tx < 1000000.0]
        if len(near_threshold) >= 2:
            alert = {
                "alert_id": f"STR-{secrets.token_hex(6)}",
                "user_id": user_id,
                "report_type": "STR",
                "severity": "CRITICAL",
                "description": f"Structuring evasion: {len(near_threshold)} transactions near ₹10L reporting threshold",
            }
            result["str_triggered"] = True
            result["alerts"].append(alert)
            self.alerts.append(alert)

        # Mandatory CTR threshold (>= ₹10 Lakh)
        if amount_inr >= 1000000.0:
            alert = {
                "alert_id": f"CTR-{secrets.token_hex(6)}",
                "user_id": user_id,
                "report_type": "CTR",
                "severity": "HIGH",
                "description": f"Mandatory CTR threshold exceeded (₹{amount_inr:.2f} >= ₹10,00,000)",
            }
            result["ctr_triggered"] = True
            result["alerts"].append(alert)
            self.alerts.append(alert)

        # Travel Rule for crypto transfers > ₹50,000
        if is_crypto and amount_inr >= 50000.0:
            if not travel_rule:
                result["allowed"] = False
                result["travel_rule_compliant"] = False
                result["alerts"].append({"error": "Travel Rule payload missing for crypto transfer > ₹50,000"})
            else:
                required = ["originator_name", "originator_account_id", "beneficiary_name", "beneficiary_wallet"]
                missing = [f for f in required if not travel_rule.get(f)]
                if missing:
                    result["allowed"] = False
                    result["travel_rule_compliant"] = False
                    result["alerts"].append({"error": f"Travel Rule violation: missing fields {missing}"})

        return result

    # -------------------------------------------------------------
    # 7. Regulatory Sandbox Guardrails (Prompt 003, 047)
    # -------------------------------------------------------------
    def validate_sandbox_bounds(self, is_domestic: bool, current_users: int, portfolio_fiat: float) -> Tuple[bool, str]:
        if is_domestic:
            # SEBI Sandbox Phase 1
            if current_users >= 10000:
                return False, "SEBI Sandbox domestic user cohort limit (10,000) reached"
            if portfolio_fiat > 50000.0:
                return False, f"Portfolio ₹{portfolio_fiat:.2f} exceeds SEBI Sandbox domestic limit ₹50,000"
        else:
            # IFSCA GIFT City Sandbox
            if current_users >= 5000:
                return False, "IFSCA Sandbox foreign investor cohort limit (5,000) reached"
            if portfolio_fiat > 10000.0:
                return False, f"Portfolio ${portfolio_fiat:.2f} exceeds IFSCA Sandbox limit $10,000"

        return True, "Within sandbox boundaries"

    # -------------------------------------------------------------
    # 8. Zero-PII Cryptographic Identity Commitment (Prompt 000, 004)
    # -------------------------------------------------------------
    def generate_identity_commitment(self, pan_hash: str, investor_uuid: str) -> str:
        payload = f"{pan_hash}:{investor_uuid}:{self.salt}".encode("utf-8")
        h = hashlib.sha256(payload).hexdigest()
        return f"0x{h}"

    # -------------------------------------------------------------
    # 9. Full Domestic Onboarding Pipeline (Prompt 004)
    # -------------------------------------------------------------
    def onboard_domestic_investor(self, profile: Dict[str, Any]) -> Tuple[bool, Dict[str, Any], str]:
        # 1. Sanctions & PEP check
        screen = self.screen_sanctions_and_pep(profile["declared_name"], "IND")
        if screen["sanctioned"] or screen["frozen"]:
            return False, {}, f"Sanctions hit: {screen['reason']}"

        # 2. PAN verification
        pan_valid, entity_or_err = self.validate_pan(
            profile["pan"],
            is_individual=True,
            aadhaar_linked=profile.get("aadhaar_linked", True),
            status=profile.get("pan_status", "OPERATIVE")
        )
        if not pan_valid:
            return False, {}, f"PAN verification failed: {entity_or_err}"
        pan_hash = self.hash_pan(profile["pan"])

        # 3. Aadhaar masking check
        aadhaar_valid, last4_or_err = self.validate_aadhaar_masking(profile["aadhaar_masked"])
        if not aadhaar_valid:
            return False, {}, f"Aadhaar masking violation: {last4_or_err}"

        # 4. Bank Penny Drop
        pd_valid, score, pd_msg = self.validate_bank_penny_drop(
            profile["ifsc"],
            profile["declared_name"],
            profile["beneficiary_name"]
        )
        if not pd_valid:
            return False, {}, f"Bank penny drop failed: {pd_msg}"

        # 5. Face Liveness
        if profile.get("face_liveness_score", 0.0) < 0.90 or not profile.get("anti_spoof_passed", False):
            return False, {}, "Biometric face liveness failed (score < 0.90 or anti-spoof flagged)"

        # 6. Risk Tiering & KYCTier assignment
        risk_tier = RiskTier.HIGH if screen["pep"] else (RiskTier.MEDIUM if profile.get("annual_income", 0) > 2500000 else RiskTier.LOW)
        ckyc = profile.get("ckyc_number", "")
        if ckyc and len(ckyc) == 14:
            assigned_tier = KYCTier.TIER_2_FULL_CKYC
        else:
            assigned_tier = KYCTier.TIER_1_BASIC_OTP

        # 7. Commitment Generation (Zero-PII)
        commitment = self.generate_identity_commitment(pan_hash, profile["investor_uuid"])
        now = datetime.now(timezone.utc)
        re_kyc_due = now + self.RE_KYC_PERIODS[risk_tier]

        record = {
            "investor_uuid": profile["investor_uuid"],
            "assigned_tier": assigned_tier.name,
            "risk_tier": risk_tier.value,
            "name_match_score": score,
            "pep_status": screen["pep_category"].value,
            "sanctions_clear": True,
            "commitment_hash": commitment,
            "verified_at": now.isoformat(),
            "re_kyc_due_at": re_kyc_due.isoformat(),
        }

        self.investors[profile["investor_uuid"]] = record
        return True, record, "APPROVED"
