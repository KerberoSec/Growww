package src

import (
	"testing"
)

func TestTraceContext_GenerateAndParseRoundtrip(t *testing.T) {
	ctx := NewTraceContext(true)
	header := ctx.String()

	parsed, err := ParseTraceparent(header)
	if err != nil {
		t.Fatalf("failed to parse generated traceparent '%s': %v", header, err)
	}

	if parsed.Version != "00" {
		t.Errorf("expected version 00, got %s", parsed.Version)
	}
	if parsed.TraceID != ctx.TraceID {
		t.Errorf("expected traceID %s, got %s", ctx.TraceID, parsed.TraceID)
	}
	if parsed.SpanID != ctx.SpanID {
		t.Errorf("expected spanID %s, got %s", ctx.SpanID, parsed.SpanID)
	}
	if !parsed.Sampled {
		t.Errorf("expected sampled=true")
	}
}

func TestTraceContext_ChildSpanGeneration(t *testing.T) {
	parent := NewTraceContext(false)
	child := parent.CreateChildSpan()

	if child.TraceID != parent.TraceID {
		t.Errorf("child should inherit parent traceID: %s != %s", child.TraceID, parent.TraceID)
	}
	if child.SpanID == parent.SpanID {
		t.Errorf("child should have a new distinct spanID")
	}
	if child.Sampled != parent.Sampled {
		t.Errorf("child should inherit parent sampling decision")
	}
}

func TestTraceContext_InvalidHeaders(t *testing.T) {
	invalidCases := []string{
		"",
		"invalid-format",
		"01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", // unsupported version 01
		"00-shorttrace-00f067aa0ba902b7-01",                        // short trace id
		"00-4bf92f3577b34da6a3ce929d0e0e4736-shortspan-01",         // short span id
	}

	for _, bad := range invalidCases {
		_, err := ParseTraceparent(bad)
		if err == nil {
			t.Errorf("expected error for invalid header '%s'", bad)
		}
	}
}
