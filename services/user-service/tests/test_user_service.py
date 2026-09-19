import unittest
import time
import sys
import os

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../src')))
from user_service import UserService

class TestUserService(unittest.TestCase):
    def setUp(self):
        self.svc = UserService()

    def test_registration_and_login_success(self):
        user = self.svc.register_user("trader@growww.in", "+919876543210", "SuperSecretP@ss123")
        self.assertTrue(user.user_id.startswith("usr_"))
        self.assertEqual(user.email, "trader@growww.in")
        self.assertEqual(user.kyc_tier, 0)

        # Login
        auth_user, session = self.svc.authenticate("trader@growww.in", "SuperSecretP@ss123", "192.168.1.10", "Mozilla/5.0")
        self.assertEqual(auth_user.user_id, user.user_id)
        self.assertTrue(session.is_active)

        # Session validation
        validated = self.svc.validate_session(session.session_id)
        self.assertIsNotNone(validated)
        self.assertEqual(validated.user_id, user.user_id)

    def test_duplicate_registration_fails(self):
        self.svc.register_user("unique@growww.in", "+919876543210", "Password123!")
        with self.assertRaises(ValueError):
            self.svc.register_user("unique@growww.in", "+919876543211", "DifferentPassword123!")

    def test_invalid_password_lockout(self):
        self.svc.register_user("lockout@growww.in", "+919876543210", "Password123!")
        for _ in range(4):
            with self.assertRaises(ValueError):
                self.svc.authenticate("lockout@growww.in", "WrongPassword", "1.1.1.1", "curl")

        # 5th attempt triggers lockout
        with self.assertRaises(ValueError):
            self.svc.authenticate("lockout@growww.in", "WrongPassword", "1.1.1.1", "curl")

        # Subsequent attempts are locked
        with self.assertRaises(PermissionError):
            self.svc.authenticate("lockout@growww.in", "Password123!", "1.1.1.1", "curl")

    def test_kyc_tier_upgrade(self):
        user = self.svc.register_user("kyc@growww.in", "+919876543210", "Password123!")
        self.assertEqual(user.kyc_tier, 0)

        # Tier 1 upgrade
        self.svc.upgrade_kyc_tier(user.user_id, 1, pan="ABCDE1234F")
        self.assertEqual(user.kyc_tier, 1)
        self.assertEqual(user.pan, "ABCDE1234F")

        # Cannot skip to Tier 3
        with self.assertRaises(ValueError):
            self.svc.upgrade_kyc_tier(user.user_id, 3)

        # Step to Tier 2
        self.svc.upgrade_kyc_tier(user.user_id, 2, aadhaar_token="vault_tok_aadhaar_8765")
        self.assertEqual(user.kyc_tier, 2)
        self.assertEqual(user.aadhaar_vault_token, "vault_tok_aadhaar_8765")

    def test_anti_phishing_phrase(self):
        user = self.svc.register_user("phrase@growww.in", "+919876543210", "Password123!")
        self.svc.set_anti_phishing_phrase(user.user_id, "BlueLotus2026")
        self.assertEqual(user.anti_phishing_phrase, "BlueLotus2026")

        with self.assertRaises(ValueError):
            self.svc.set_anti_phishing_phrase(user.user_id, "abc")

if __name__ == '__main__':
    unittest.main()
