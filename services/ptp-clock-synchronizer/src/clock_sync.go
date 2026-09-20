package src

import (
	"errors"
	"sync"
	"time"
)

// PTPPortState defines the IEEE 1588 port state
type PTPPortState string

const (
	PTPStateMaster  PTPPortState = "MASTER"
	PTPStateSlave   PTPPortState = "SLAVE"
	PTPStatePassive PTPPortState = "PASSIVE"
	PTPStateFaulty  PTPPortState = "FAULTY"
)

// DriftSeverity denotes compliance alert level
type DriftSeverity string

const (
	DriftNormal   DriftSeverity = "NORMAL"   // <= 10 us
	DriftWarning  DriftSeverity = "WARNING"  // > 10 us
	DriftCritical DriftSeverity = "CRITICAL" // > 50 us (SEBI/MiFID II violation threshold)
)

const (
	WarningDriftThresholdMicros  = 10.0
	CriticalDriftThresholdMicros = 50.0
)

// GrandmasterClock represents a Stratum 1 GPS/atomic clock source
type GrandmasterClock struct {
	GrandmasterID   string       `json:"grandmaster_id"`
	IPAddress       string       `json:"ip_address"`
	Stratum         int          `json:"stratum"`
	State           PTPPortState `json:"state"`
	OffsetNanos     int64        `json:"offset_nanos"`
	JitterNanos     int64        `json:"jitter_nanos"`
	LastSyncTime    time.Time    `json:"last_sync_time"`
	IsActivePrimary bool         `json:"is_active_primary"`
}

// OrderEventTimestamp provides microsecond and nanosecond audited precision
type OrderEventTimestamp struct {
	MonotonicNanos   int64     `json:"monotonic_nanos"`
	UtcMicroseconds  int64     `json:"utc_microseconds"`
	UtcIsoFormatted  string    `json:"utc_iso_formatted"`
	ActiveGMID       string    `json:"active_gmid"`
	ClockDriftMicros float64   `json:"clock_drift_micros"`
	ComplianceStatus string    `json:"compliance_status"`
}

// ClockSyncManager coordinates IEEE 1588v2 hardware PTP tracking
type ClockSyncManager struct {
	mu            sync.RWMutex
	grandmasters  map[string]*GrandmasterClock
	activeGMID    string
	lastNano      int64
	alertCallback func(severity DriftSeverity, offsetMicros float64, gmid string)
}

func NewClockSyncManager() *ClockSyncManager {
	return &ClockSyncManager{
		grandmasters: make(map[string]*GrandmasterClock),
	}
}

func (m *ClockSyncManager) SetAlertCallback(cb func(severity DriftSeverity, offsetMicros float64, gmid string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertCallback = cb
}

// RegisterGrandmaster adds a GPS/PTP time source
func (m *ClockSyncManager) RegisterGrandmaster(id, ip string, stratum int, isPrimary bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" || ip == "" {
		return errors.New("invalid grandmaster configuration")
	}

	gm := &GrandmasterClock{
		GrandmasterID:   id,
		IPAddress:       ip,
		Stratum:         stratum,
		State:           PTPStateSlave,
		OffsetNanos:     0,
		JitterNanos:     0,
		LastSyncTime:    time.Now(),
		IsActivePrimary: isPrimary,
	}

	m.grandmasters[id] = gm
	if isPrimary || m.activeGMID == "" {
		m.activeGMID = id
	}
	return nil
}

// RecordSyncPulse records a hardware PTP sync pulse and evaluates drift
func (m *ClockSyncManager) RecordSyncPulse(gmid string, offsetNanos, jitterNanos int64, now time.Time) (DriftSeverity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	gm, exists := m.grandmasters[gmid]
	if !exists {
		return DriftNormal, errors.New("unknown grandmaster")
	}

	gm.OffsetNanos = offsetNanos
	gm.JitterNanos = jitterNanos
	gm.LastSyncTime = now

	offsetMicros := float64(abs(offsetNanos)) / 1000.0
	severity := DriftNormal

	if offsetMicros > CriticalDriftThresholdMicros {
		severity = DriftCritical
		gm.State = PTPStateFaulty
		// Attempt failover to secondary if primary is faulty
		m.failoverToHealthyGM(gmid)
	} else if offsetMicros > WarningDriftThresholdMicros {
		severity = DriftWarning
		gm.State = PTPStateSlave
	} else {
		gm.State = PTPStateSlave
	}

	if (severity == DriftWarning || severity == DriftCritical) && m.alertCallback != nil {
		m.alertCallback(severity, offsetMicros, gmid)
	}

	return severity, nil
}

// InjectMonotonicTimestamp injects microsecond-accurate audited timestamp
func (m *ClockSyncManager) InjectMonotonicTimestamp() OrderEventTimestamp {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	nowNano := now.UnixNano()

	// Guarantee strict monotonicity
	if nowNano <= m.lastNano {
		nowNano = m.lastNano + 1
	}
	m.lastNano = nowNano

	activeGM, exists := m.grandmasters[m.activeGMID]
	driftMicros := 0.0
	compliance := "SEBI_MIFID_COMPLIANT"

	if exists {
		driftMicros = float64(abs(activeGM.OffsetNanos)) / 1000.0
		if driftMicros > CriticalDriftThresholdMicros {
			compliance = "NON_COMPLIANT_OFFSET_EXCEEDED"
		}
	}

	return OrderEventTimestamp{
		MonotonicNanos:   nowNano,
		UtcMicroseconds:  nowNano / 1000,
		UtcIsoFormatted:  now.Format(time.RFC3339Nano),
		ActiveGMID:       m.activeGMID,
		ClockDriftMicros: driftMicros,
		ComplianceStatus: compliance,
	}
}

func (m *ClockSyncManager) failoverToHealthyGM(faultyID string) {
	for id, gm := range m.grandmasters {
		if id != faultyID && gm.State != PTPStateFaulty {
			m.activeGMID = id
			gm.IsActivePrimary = true
			if old, ok := m.grandmasters[faultyID]; ok {
				old.IsActivePrimary = false
			}
			return
		}
	}
}

func (m *ClockSyncManager) GetActiveGrandmaster() *GrandmasterClock {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.grandmasters[m.activeGMID]
}

func abs(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
