package src

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// InvalidParam details a specific field error in an RFC 7807 problem
type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// ProblemDetails implements RFC 7807 error envelopes
type ProblemDetails struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	Instance      string         `json:"instance"`
	ErrorCode     string         `json:"error_code"`
	RequestID     string         `json:"request_id"`
	Timestamp     int64          `json:"timestamp"`
	InvalidParams []InvalidParam `json:"invalid_params,omitempty"`
}

func NewProblemDetails(status int, title, detail, instance, errorCode, reqID string) *ProblemDetails {
	return &ProblemDetails{
		Type:      fmt.Sprintf("https://api.growww.trade/errors/%s", strings.ToLower(errorCode)),
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  instance,
		ErrorCode: errorCode,
		RequestID: reqID,
		Timestamp: time.Now().Unix(),
	}
}

func (p *ProblemDetails) AddViolation(field, reason string) {
	p.InvalidParams = append(p.InvalidParams, InvalidParam{Name: field, Reason: reason})
}

// CursorPayload represents opaque base64 cursor state
type CursorPayload struct {
	TimestampNanos int64  `json:"t"`
	SequenceNo     uint64 `json:"s"`
	ID             string `json:"id"`
}

// EncodeCursor produces base64 URL-safe opaque cursor token
func EncodeCursor(tsNanos int64, seq uint64, id string) string {
	raw := fmt.Sprintf("%d:%d:%s", tsNanos, seq, id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes opaque cursor token
func DecodeCursor(token string) (*CursorPayload, error) {
	if token == "" {
		return nil, errors.New("empty cursor token")
	}

	bytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor base64: %w", err)
	}

	parts := strings.Split(string(bytes), ":")
	if len(parts) < 3 {
		return nil, errors.New("malformed cursor payload")
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, errors.New("invalid cursor timestamp")
	}

	seq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, errors.New("invalid cursor sequence")
	}

	return &CursorPayload{
		TimestampNanos: ts,
		SequenceNo:     seq,
		ID:             parts[2],
	}, nil
}

// HTTPToGRPCCode maps standard HTTP statuses to canonical gRPC error codes
func HTTPToGRPCCode(status int) string {
	switch status {
	case 400:
		return "INVALID_ARGUMENT"
	case 401:
		return "UNAUTHENTICATED"
	case 403:
		return "PERMISSION_DENIED"
	case 404:
		return "NOT_FOUND"
	case 409:
		return "ALREADY_EXISTS"
	case 422:
		return "FAILED_PRECONDITION"
	case 429:
		return "RESOURCE_EXHAUSTED"
	default:
		return "INTERNAL"
	}
}

// GRPCToHTTPStatus maps canonical gRPC error codes to HTTP statuses
func GRPCToHTTPStatus(grpcCode string) int {
	switch grpcCode {
	case "INVALID_ARGUMENT":
		return 400
	case "UNAUTHENTICATED":
		return 401
	case "PERMISSION_DENIED":
		return 403
	case "NOT_FOUND":
		return 404
	case "ALREADY_EXISTS":
		return 409
	case "FAILED_PRECONDITION":
		return 422
	case "RESOURCE_EXHAUSTED":
		return 429
	default:
		return 500
	}
}
