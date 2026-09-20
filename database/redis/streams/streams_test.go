package streams

import (
	"context"
	"testing"
	"time"
)

func TestMarketFeedConflatorCalculations(t *testing.T) {
	var emitted []*ConflatedTicker
	emitter := func(tk *ConflatedTicker) {
		emitted = append(emitted, tk)
	}

	// 10-second window so it doesn't auto-flush during test
	conflator := NewMarketFeedConflator(10*time.Second, emitter)
	defer conflator.Stop()

	// Ingest 4 ticks for INFY
	t0 := time.Now().UTC()
	conflator.IngestTick(&MarketTick{ISIN: "INE009A01021", Symbol: "INFY", Price: 1500.0, Quantity: 10, SequenceNo: 1, Timestamp: t0})
	conflator.IngestTick(&MarketTick{ISIN: "INE009A01021", Symbol: "INFY", Price: 1520.0, Quantity: 20, SequenceNo: 2, Timestamp: t0})
	conflator.IngestTick(&MarketTick{ISIN: "INE009A01021", Symbol: "INFY", Price: 1490.0, Quantity: 10, SequenceNo: 3, Timestamp: t0})
	conflator.IngestTick(&MarketTick{ISIN: "INE009A01021", Symbol: "INFY", Price: 1510.0, Quantity: 10, SequenceNo: 4, Timestamp: t0})

	tickers := conflator.FlushAll()
	if len(tickers) != 1 {
		t.Fatalf("expected 1 conflated ticker, got %d", len(tickers))
	}

	tk := tickers[0]
	if tk.Open != 1500.0 {
		t.Errorf("expected Open 1500.0, got %f", tk.Open)
	}
	if tk.High != 1520.0 {
		t.Errorf("expected High 1520.0, got %f", tk.High)
	}
	if tk.Low != 1490.0 {
		t.Errorf("expected Low 1490.0, got %f", tk.Low)
	}
	if tk.Close != 1510.0 {
		t.Errorf("expected Close 1510.0, got %f", tk.Close)
	}
	if tk.Volume != 50.0 {
		t.Errorf("expected Volume 50.0, got %f", tk.Volume)
	}

	// Expected Turnover: (1500*10) + (1520*20) + (1490*10) + (1510*10) = 15000 + 30400 + 14900 + 15100 = 75400
	// Expected VWAP: 75400 / 50 = 1508.0
	if tk.VWAP != 1508.0 {
		t.Errorf("expected VWAP 1508.0, got %f", tk.VWAP)
	}
	if tk.TickCount != 4 {
		t.Errorf("expected TickCount 4, got %d", tk.TickCount)
	}
	if tk.SequenceFrom != 1 || tk.SequenceTo != 4 {
		t.Errorf("expected sequence 1 -> 4, got %d -> %d", tk.SequenceFrom, tk.SequenceTo)
	}
}

func TestStreamBrokerMaxLenTrimming(t *testing.T) {
	broker := NewStreamBroker()
	stream := "growww:stream:test"

	// Add 10 messages with MAXLEN=5
	for i := 1; i <= 10; i++ {
		_, err := broker.XAdd(stream, 5, map[string]interface{}{"val": i})
		if err != nil {
			t.Fatalf("failed to xadd: %v", err)
		}
	}

	if broker.XLen(stream) != 5 {
		t.Fatalf("expected stream length capped at 5, got %d", broker.XLen(stream))
	}

	// Last 5 elements should be values 6..10
	msgs := broker.XRange(stream, "-", 10)
	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages from xrange, got %d", len(msgs))
	}
	if msgs[0].Payload["val"] != 6 || msgs[4].Payload["val"] != 10 {
		t.Errorf("expected trimmed messages 6 to 10, got %v to %v", msgs[0].Payload["val"], msgs[4].Payload["val"])
	}
}

func TestConsumerGroupReadAckAndAutoClaim(t *testing.T) {
	broker := NewStreamBroker()
	stream := "growww:stream:orders"

	// Publish 3 messages
	msg1, _ := broker.XAdd(stream, 100, map[string]interface{}{"id": 1})
	msg2, _ := broker.XAdd(stream, 100, map[string]interface{}{"id": 2})
	msg3, _ := broker.XAdd(stream, 100, map[string]interface{}{"id": 3})

	cg := NewConsumerGroup("order_processors", stream, broker, 2)

	// Consumer A reads 2 messages
	msgsA, err := cg.XReadGroup(context.Background(), "consumer-A", 2)
	if err != nil || len(msgsA) != 2 {
		t.Fatalf("expected 2 messages read by consumer A, got %d", len(msgsA))
	}

	// PEL should contain 2 pending messages
	if len(cg.XPending()) != 2 {
		t.Errorf("expected 2 pending messages in PEL, got %d", len(cg.XPending()))
	}

	// Consumer A acknowledges msg1
	acked := cg.XAck(context.Background(), msg1)
	if acked != 1 {
		t.Errorf("expected 1 acked message, got %d", acked)
	}

	// Now PEL should have only 1 pending (msg2)
	pending := cg.XPending()
	if len(pending) != 1 || pending[0].ID != msg2 {
		t.Errorf("expected only msg2 pending in PEL")
	}

	// Simulate Consumer A crashing and leaving msg2 idle for 10ms
	time.Sleep(15 * time.Millisecond)

	// Consumer B auto-claims idle messages (minIdleTime = 5ms)
	claimed, dlq := cg.XAutoClaim(context.Background(), "consumer-B", 5*time.Millisecond, 10)
	if len(claimed) != 1 || claimed[0].ID != msg2 || claimed[0].Consumer != "consumer-B" {
		t.Fatalf("expected msg2 to be claimed by consumer-B")
	}
	if len(dlq) != 0 {
		t.Errorf("expected 0 DLQ messages on first reclaim")
	}

	// Consumer B also crashes, idle again
	time.Sleep(15 * time.Millisecond)

	// Another reclaim: delivery attempts will reach 3 > maxRetries(2) -> routed to DLQ!
	claimed2, dlq2 := cg.XAutoClaim(context.Background(), "consumer-C", 5*time.Millisecond, 10)
	if len(claimed2) != 0 {
		t.Errorf("expected 0 claimed because msg2 exceeded retry limit")
	}
	if len(dlq2) != 1 || dlq2[0].ID != msg2 {
		t.Errorf("expected msg2 to be routed to DLQ")
	}

	// Consumer C reads remaining msg3
	msgsC, _ := cg.XReadGroup(context.Background(), "consumer-C", 10)
	if len(msgsC) != 1 || msgsC[0].ID != msg3 {
		t.Errorf("expected consumer-C to read msg3")
	}
}

func TestProtobufServiceContract(t *testing.T) {
	conflator := NewMarketFeedConflator(10*time.Second, nil)
	defer conflator.Stop()

	req := &RedisstreamsconflatedmarketfeedRequest{
		RequestID:   "req-conflated-1",
		EntityID:    "inst-001",
		AmountE8:    100000000,
		TimestampMs: uint64(time.Now().UnixMilli()),
		Metadata:    map[string]string{"feed": "ticker"},
	}

	resp, err := conflator.ExecuteService(context.Background(), req)
	if err != nil || !resp.Success {
		t.Fatalf("expected successful execution: %v", err)
	}
	if resp.RequestID != "req-conflated-1" {
		t.Errorf("request_id mismatch: %s", resp.RequestID)
	}
}
