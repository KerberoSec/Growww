package main

import (
	"strings"
	"testing"
)

func newTestCCIPService() *CCIPIngressService {
	svc := NewCCIPIngressService(1)
	svc.RegisterLane(CCIPLane{
		SourceChainSelector: 1,
		DestChainSelector:   137,
		OnRampAddress:       "0xOnRamp",
		OffRampAddress:      "0xOffRamp",
	})
	svc.RegisterLane(CCIPLane{
		SourceChainSelector: 137,
		DestChainSelector:   42161,
		OnRampAddress:       "0xOnRamp2",
		OffRampAddress:      "0xOffRamp2",
	})
	return svc
}

func TestCCIPIngestHappyPath(t *testing.T) {
	svc := newTestCCIPService()
	msg := CCIPMessage{
		MessageID:      "MSG-001",
		SourceChain:    1,
		DestChain:      137,
		SequenceNumber: 1,
		Sender:         "0xAlice",
		Receiver:       "0xBob",
		TokenAmounts:   []TokenAmount{{Token: "USDC", AmountE8: 100_00000000}},
		MessageType:    CCIPTokenTransfer,
		Nonce:          1,
	}

	event, err := svc.IngestMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.VerificationStatus != VerificationPending {
		t.Errorf("expected PENDING, got %s", event.VerificationStatus)
	}
	if !strings.HasPrefix(event.CommitRoot, "0x") {
		t.Errorf("commit root should start with 0x: %s", event.CommitRoot)
	}
	if !strings.HasPrefix(event.EventID, "CCIP-EVT-") {
		t.Errorf("event ID format wrong: %s", event.EventID)
	}
}

func TestCCIPDuplicateRejected(t *testing.T) {
	svc := newTestCCIPService()
	msg := CCIPMessage{
		MessageID: "MSG-DUP", SourceChain: 1, DestChain: 137,
		SequenceNumber: 1, Sender: "0xA", Receiver: "0xB", MessageType: CCIPTokenTransfer,
	}

	_, _ = svc.IngestMessage(msg)
	_, err := svc.IngestMessage(msg)
	if err == nil {
		t.Fatal("expected duplicate rejection")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCCIPUnsupportedLane(t *testing.T) {
	svc := newTestCCIPService()
	msg := CCIPMessage{
		MessageID: "MSG-BAD", SourceChain: 1, DestChain: 56, // BSC not registered
		SequenceNumber: 1, Sender: "0xA", Receiver: "0xB",
	}

	_, err := svc.IngestMessage(msg)
	if err == nil {
		t.Fatal("expected unsupported lane error")
	}
}

func TestCCIPSequenceOrdering(t *testing.T) {
	svc := newTestCCIPService()

	// Ingest sequence 5
	msg1 := CCIPMessage{
		MessageID: "MSG-S5", SourceChain: 1, DestChain: 137,
		SequenceNumber: 5, Sender: "0xA", Receiver: "0xB",
	}
	_, _ = svc.IngestMessage(msg1)

	// Try to ingest sequence 3 (out of order)
	msg2 := CCIPMessage{
		MessageID: "MSG-S3", SourceChain: 1, DestChain: 137,
		SequenceNumber: 3, Sender: "0xA", Receiver: "0xB",
	}
	_, err := svc.IngestMessage(msg2)
	if err == nil {
		t.Fatal("expected out-of-order error")
	}
}

func TestCCIPConfirmAndReject(t *testing.T) {
	svc := newTestCCIPService()
	msg := CCIPMessage{
		MessageID: "MSG-CFM", SourceChain: 1, DestChain: 137,
		SequenceNumber: 1, Sender: "0xA", Receiver: "0xB",
	}
	_, _ = svc.IngestMessage(msg)

	// Confirm
	err := svc.ConfirmEvent("MSG-CFM")
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	event, _ := svc.GetEvent("MSG-CFM")
	if event.VerificationStatus != VerificationFinalized {
		t.Errorf("expected FINALIZED, got %s", event.VerificationStatus)
	}

	// Try to reject a finalized event (should reject)
	msg2 := CCIPMessage{
		MessageID: "MSG-REJ", SourceChain: 1, DestChain: 137,
		SequenceNumber: 2, Sender: "0xA", Receiver: "0xB",
	}
	_, _ = svc.IngestMessage(msg2)
	_ = svc.RejectEvent("MSG-REJ", "fraud proof")
	evt2, _ := svc.GetEvent("MSG-REJ")
	if evt2.VerificationStatus != VerificationRejected {
		t.Errorf("expected REJECTED, got %s", evt2.VerificationStatus)
	}

	// Cannot confirm a rejected event
	err = svc.ConfirmEvent("MSG-REJ")
	if err == nil {
		t.Fatal("expected error confirming rejected event")
	}
}
