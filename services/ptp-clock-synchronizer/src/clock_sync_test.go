package src

import (
	"testing"
	"time"
)

func TestClockSyncManager_NormalDriftAndTimestampMonotonicity(t *testing.T) {
	mgr := NewClockSyncManager()

	err := mgr.RegisterGrandmaster("GM_GPS_MUMBAI_01", "10.0.1.50", 1, true)
	if err != nil {
		t.Fatalf("failed to register GM: %v", err)
	}

	// Normal sync pulse: 2500 nanos offset = 2.5 microseconds (< 10us)
	now := time.Now()
	severity, err := mgr.RecordSyncPulse("GM_GPS_MUMBAI_01", 2500, 150, now)
	if err != nil {
		t.Fatalf("sync pulse error: %v", err)
	}
	if severity != DriftNormal {
		t.Errorf("expected NORMAL, got %v", severity)
	}

	// Test timestamp monotonicity across 1000 fast iterations
	var prevNano int64 = 0
	for i := 0; i < 1000; i++ {
		ts := mgr.InjectMonotonicTimestamp()
		if ts.MonotonicNanos <= prevNano {
			t.Fatalf("monotonicity violated: current %d <= prev %d", ts.MonotonicNanos, prevNano)
		}
		if ts.ComplianceStatus != "SEBI_MIFID_COMPLIANT" {
			t.Errorf("expected compliant status, got %s", ts.ComplianceStatus)
		}
		prevNano = ts.MonotonicNanos
	}
}

func TestClockSyncManager_WarningAndCriticalAlerts(t *testing.T) {
	mgr := NewClockSyncManager()

	alerts := make([]DriftSeverity, 0)
	mgr.SetAlertCallback(func(sev DriftSeverity, offset float64, gmid string) {
		alerts = append(alerts, sev)
	})

	_ = mgr.RegisterGrandmaster("GM_PRIMARY", "10.0.1.1", 1, true)
	_ = mgr.RegisterGrandmaster("GM_SECONDARY", "10.0.1.2", 1, false)

	now := time.Now()
	// 1. Warning: 25 microseconds offset (25,000 nanos)
	sev1, _ := mgr.RecordSyncPulse("GM_PRIMARY", 25000, 500, now)
	if sev1 != DriftWarning {
		t.Errorf("expected WARNING, got %v", sev1)
	}

	// 2. Critical: 60 microseconds offset (60,000 nanos) > 50us
	sev2, _ := mgr.RecordSyncPulse("GM_PRIMARY", 60000, 1200, now)
	if sev2 != DriftCritical {
		t.Errorf("expected CRITICAL, got %v", sev2)
	}

	// Verify failover to GM_SECONDARY
	active := mgr.GetActiveGrandmaster()
	if active == nil || active.GrandmasterID != "GM_SECONDARY" {
		t.Errorf("expected failover to GM_SECONDARY, got %+v", active)
	}

	// Non-compliant status when querying timestamps
	ts := mgr.InjectMonotonicTimestamp()
	if ts.ActiveGMID != "GM_SECONDARY" {
		t.Errorf("expected active GM_SECONDARY in timestamp, got %s", ts.ActiveGMID)
	}

	if len(alerts) != 2 {
		t.Errorf("expected 2 alerts dispatched, got %d", len(alerts))
	}
}
