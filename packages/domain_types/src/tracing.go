package src

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// TraceContext encapsulates W3C TraceContext Level 1
type TraceContext struct {
	Version  string `json:"version"`   // "00"
	TraceID  string `json:"trace_id"`  // 32-hex chars
	SpanID   string `json:"span_id"`   // 16-hex chars
	Sampled  bool   `json:"sampled"`
}

// NewTraceContext generates a fresh root trace context
func NewTraceContext(sampled bool) *TraceContext {
	traceBytes := make([]byte, 16)
	spanBytes := make([]byte, 8)
	_, _ = rand.Read(traceBytes)
	_, _ = rand.Read(spanBytes)

	return &TraceContext{
		Version: "00",
		TraceID: hex.EncodeToString(traceBytes),
		SpanID:  hex.EncodeToString(spanBytes),
		Sampled: sampled,
	}
}

// String formats context into W3C traceparent header: 00-{trace_id}-{span_id}-{flags}
func (t *TraceContext) String() string {
	flags := "00"
	if t.Sampled {
		flags = "01"
	}
	return fmt.Sprintf("%s-%s-%s-%s", t.Version, t.TraceID, t.SpanID, flags)
}

// ParseTraceparent parses standard W3C traceparent string
func ParseTraceparent(header string) (*TraceContext, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil, errors.New("empty traceparent header")
	}

	parts := strings.Split(header, "-")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid traceparent format (expected 4 segments, got %d)", len(parts))
	}

	version, traceID, spanID, flags := parts[0], parts[1], parts[2], parts[3]

	if version != "00" {
		return nil, fmt.Errorf("unsupported traceparent version '%s'", version)
	}
	if len(traceID) != 32 {
		return nil, fmt.Errorf("invalid trace_id length: %d (expected 32 hex chars)", len(traceID))
	}
	if len(spanID) != 16 {
		return nil, fmt.Errorf("invalid span_id length: %d (expected 16 hex chars)", len(spanID))
	}

	sampled := (flags == "01")

	return &TraceContext{
		Version: version,
		TraceID: traceID,
		SpanID:  spanID,
		Sampled: sampled,
	}, nil
}

// CreateChildSpan creates a child span preserving the parent trace ID
func (t *TraceContext) CreateChildSpan() *TraceContext {
	childSpanBytes := make([]byte, 8)
	_, _ = rand.Read(childSpanBytes)

	return &TraceContext{
		Version: t.Version,
		TraceID: t.TraceID,
		SpanID:  hex.EncodeToString(childSpanBytes),
		Sampled: t.Sampled,
	}
}
