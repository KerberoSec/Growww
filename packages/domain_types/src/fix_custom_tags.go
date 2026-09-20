package src

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const SOH = "\x01"

// Custom FIX 5.0 SP2 tags for institutional trading & regulatory reporting.
const (
	TagExchangeOrderID       int = 20001
	TagAlgoStrategyID        int = 20002
	TagExecutionLatencyNanos int = 20003
	TagZeroPIIPanToken       int = 20004
	TagSebiCategory          int = 20005
	TagSGFContributionE8     int = 20006
)

var (
	ErrFIXInvalidChecksum = errors.New("fix: checksum verification failed")
	ErrFIXTruncatedMsg    = errors.New("fix: truncated message or missing SOH delimiter")
	ErrFIXMissingTag      = errors.New("fix: mandatory tag missing")
)

// FIXMessage represents a parsed FIX tag-value message.
type FIXMessage struct {
	Tags map[int]string
}

// NewFIXMessage creates a new FIX message.
func NewFIXMessage() *FIXMessage {
	return &FIXMessage{Tags: make(map[int]string)}
}

// Set sets a tag and string value.
func (m *FIXMessage) Set(tag int, val string) {
	m.Tags[tag] = val
}

// SetInt sets an integer tag.
func (m *FIXMessage) SetInt(tag int, val int64) {
	m.Tags[tag] = strconv.FormatInt(val, 10)
}

// Get returns tag value.
func (m *FIXMessage) Get(tag int) (string, bool) {
	val, ok := m.Tags[tag]
	return val, ok
}

// Encode serializes message into raw FIX format with standard Tag 8, 9, and Tag 10 Checksum.
func (m *FIXMessage) Encode() string {
	var bodyBuilder strings.Builder

	// Write standard MsgType if present, then other tags (excluding 8, 9, 10)
	for tag, val := range m.Tags {
		if tag != 8 && tag != 9 && tag != 10 {
			bodyBuilder.WriteString(fmt.Sprintf("%d=%s%s", tag, val, SOH))
		}
	}

	body := bodyBuilder.String()
	bodyLen := len(body)

	// Prepend 8=FIX.5.0SP2 and 9=BodyLength
	header := fmt.Sprintf("8=FIX.5.0SP2%s9=%d%s", SOH, bodyLen, SOH)
	fullUntilChecksum := header + body

	// Calculate Tag 10 Checksum: sum of all bytes modulo 256
	var sum int
	for i := 0; i < len(fullUntilChecksum); i++ {
		sum += int(fullUntilChecksum[i])
	}
	checksum := sum % 256

	return fmt.Sprintf("%s10=%03d%s", fullUntilChecksum, checksum, SOH)
}

// ParseFIXMessage parses wire FIX string and validates Tag 10 checksum.
func ParseFIXMessage(raw string) (*FIXMessage, error) {
	if !strings.HasSuffix(raw, SOH) {
		return nil, ErrFIXTruncatedMsg
	}

	// Verify Checksum
	lastTagIdx := strings.LastIndex(raw[:len(raw)-1], "10=")
	if lastTagIdx == -1 {
		return nil, errors.New("fix: missing Tag 10 checksum")
	}

	msgWithoutChecksum := raw[:lastTagIdx]
	var sum int
	for i := 0; i < len(msgWithoutChecksum); i++ {
		sum += int(msgWithoutChecksum[i])
	}
	expectedChecksum := sum % 256

	fields := strings.Split(raw[:len(raw)-1], SOH)
	msg := NewFIXMessage()

	for _, field := range fields {
		if field == "" {
			continue
		}
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}
		tagNum, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		msg.Tags[tagNum] = parts[1]
	}

	chkStr, hasChecksum := msg.Get(10)
	if !hasChecksum {
		return nil, ErrFIXInvalidChecksum
	}
	chkNum, _ := strconv.Atoi(chkStr)
	if chkNum != expectedChecksum {
		return nil, fmt.Errorf("%w: expected %03d, got %03d", ErrFIXInvalidChecksum, expectedChecksum, chkNum)
	}

	return msg, nil
}
