import unittest
from scripts.security.bounty_sla_tracker import BountySLATracker

class TestBountySLATracker(unittest.TestCase):
    def setUp(self):
        self.tracker = BountySLATracker()

    def test_classify_severity_by_cvss(self):
        self.assertEqual(self.tracker.classify_severity(9.8), "CRITICAL")
        self.assertEqual(self.tracker.classify_severity(7.5), "HIGH")
        self.assertEqual(self.tracker.classify_severity(5.3), "MEDIUM")
        self.assertEqual(self.tracker.classify_severity(2.1), "LOW")
        self.assertEqual(self.tracker.classify_severity(0.0), "INFORMATIONAL")

        with self.assertRaises(ValueError):
            self.tracker.classify_severity(11.0)

    def test_sla_status_on_track_and_expiry_calculation(self):
        submitted = "2026-09-20T10:00:00Z"
        expiry = self.tracker.calculate_sla_expiry(submitted, "CRITICAL")
        self.assertIn("2026-09-21T10:00:00", expiry)

        # Evaluate at 4 hours later (20 hours remaining > 25% of 24h)
        current = "2026-09-20T14:00:00Z"
        res = self.tracker.evaluate_sla_status(submitted, "CRITICAL", current)
        self.assertEqual(res["status"], "ON_TRACK")
        self.assertEqual(res["remaining_hours"], 20.0)
        self.assertFalse(res["escalate_pagerduty"])

    def test_sla_status_urgent_escalation_and_breach(self):
        submitted = "2026-09-20T00:00:00Z"

        # 20 hours elapsed out of 24 hours -> 4h remaining (16.6% <= 25%) -> URGENT
        current_urgent = "2026-09-20T20:00:00Z"
        res_urgent = self.tracker.evaluate_sla_status(submitted, "CRITICAL", current_urgent)
        self.assertEqual(res_urgent["status"], "URGENT_ESCALATION")
        self.assertTrue(res_urgent["escalate_pagerduty"])

        # 26 hours elapsed -> BREACHED
        current_breach = "2026-09-21T02:00:00Z"
        res_breach = self.tracker.evaluate_sla_status(submitted, "CRITICAL", current_breach)
        self.assertEqual(res_breach["status"], "SLA_BREACHED")
        self.assertTrue(res_breach["escalate_pagerduty"])

    def test_bounty_payout_calculation(self):
        crit_full = self.tracker.compute_bounty_recommendation("CRITICAL", 1.0)
        self.assertEqual(crit_full["recommended_inr"], 1000000)

        crit_min = self.tracker.compute_bounty_recommendation("CRITICAL", 0.0)
        self.assertEqual(crit_min["recommended_inr"], 500000)

        high_mid = self.tracker.compute_bounty_recommendation("HIGH", 0.5)
        self.assertEqual(high_mid["recommended_inr"], 225000)

if __name__ == "__main__":
    unittest.main()
