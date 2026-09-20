package src

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProblemDetails_RFC7807Formatting(t *testing.T) {
	prob := NewProblemDetails(
		422,
		"Insufficient Margin Available",
		"Account available margin is insufficient for requested order size.",
		"/api/v1/orders/ord_9901",
		"ERROR_CODE_INSUFFICIENT_FUNDS",
		"req_test_12345",
	)
	prob.AddViolation("amount_e8", "Exceeds margin capability")

	data, err := json.Marshal(prob)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if parsed["status"].(float64) != 422 {
		t.Errorf("expected status 422, got %v", parsed["status"])
	}
	if parsed["error_code"] != "ERROR_CODE_INSUFFICIENT_FUNDS" {
		t.Errorf("expected ERROR_CODE_INSUFFICIENT_FUNDS, got %v", parsed["error_code"])
	}
}

func TestCursorPagination_Roundtrip(t *testing.T) {
	nowNano := time.Now().UnixNano()
	seq := uint64(88421)
	rowID := "ord_7741_uuid"

	token := EncodeCursor(nowNano, seq, rowID)
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	decoded, err := DecodeCursor(token)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.TimestampNanos != nowNano {
		t.Errorf("expected ts %d, got %d", nowNano, decoded.TimestampNanos)
	}
	if decoded.SequenceNo != seq {
		t.Errorf("expected seq %d, got %d", seq, decoded.SequenceNo)
	}
	if decoded.ID != rowID {
		t.Errorf("expected id %s, got %s", rowID, decoded.ID)
	}
}

func TestHTTPAndGRPCMappings(t *testing.T) {
	cases := []struct {
		httpStatus int
		grpcCode   string
	}{
		{400, "INVALID_ARGUMENT"},
		{401, "UNAUTHENTICATED"},
		{403, "PERMISSION_DENIED"},
		{404, "NOT_FOUND"},
		{409, "ALREADY_EXISTS"},
		{422, "FAILED_PRECONDITION"},
		{429, "RESOURCE_EXHAUSTED"},
		{500, "INTERNAL"},
	}

	for _, c := range cases {
		mappedGRPC := HTTPToGRPCCode(c.httpStatus)
		if mappedGRPC != c.grpcCode {
			t.Errorf("HTTP %d: expected gRPC %s, got %s", c.httpStatus, c.grpcCode, mappedGRPC)
		}
		mappedHTTP := GRPCToHTTPStatus(c.grpcCode)
		if mappedHTTP != c.httpStatus {
			t.Errorf("gRPC %s: expected HTTP %d, got %d", c.grpcCode, c.httpStatus, mappedHTTP)
		}
	}
}
