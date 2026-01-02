package scela

import (
	"context"
	"testing"
	"time"
)

const testTopic = "test.topic"

func TestMessageHistory(t *testing.T) {
	history := NewMessageHistory(100)

	msg1 := NewMessage(testTopic, "payload1")
	msg2 := NewMessage(testTopic, "payload2")
	msg3 := NewMessage("other.topic", "payload3")

	// Record some entries
	history.Record(HistoryEntry{
		Message:   msg1,
		Event:     "published",
		Timestamp: time.Now(),
	})

	time.Sleep(10 * time.Millisecond)

	history.Record(HistoryEntry{
		Message:   msg1,
		Event:     "delivered",
		Timestamp: time.Now(),
	})

	history.Record(HistoryEntry{
		Message:   msg2,
		Event:     "published",
		Timestamp: time.Now(),
	})

	history.Record(HistoryEntry{
		Message:   msg3,
		Event:     "failed",
		Timestamp: time.Now(),
		Error:     "handler error",
	})

	// Test GetAll
	all := history.GetAll()
	if len(all) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(all))
	}

	// Test GetByMessageID
	msg1Entries := history.GetByMessageID(msg1.ID())
	if len(msg1Entries) != 2 {
		t.Errorf("Expected 2 entries for msg1, got %d", len(msg1Entries))
	}

	// Test GetByTopic
	testTopicEntries := history.GetByTopic(testTopic)
	if len(testTopicEntries) != 3 {
		t.Errorf("Expected 3 entries for test.topic, got %d", len(testTopicEntries))
	}

	// Test GetByEvent
	publishedEntries := history.GetByEvent("published")
	if len(publishedEntries) != 2 {
		t.Errorf("Expected 2 published entries, got %d", len(publishedEntries))
	}

	failedEntries := history.GetByEvent("failed")
	if len(failedEntries) != 1 {
		t.Errorf("Expected 1 failed entry, got %d", len(failedEntries))
	}
	if failedEntries[0].Error != "handler error" {
		t.Errorf("Expected error 'handler error', got '%s'", failedEntries[0].Error)
	}

	// Test Count
	if history.Count() != 4 {
		t.Errorf("Expected count 4, got %d", history.Count())
	}

	// Test Clear
	history.Clear()
	if history.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", history.Count())
	}
}

func TestMessageHistoryMaxSize(t *testing.T) {
	history := NewMessageHistory(3)

	for i := 0; i < 5; i++ {
		msg := NewMessage(testTopic, i)
		history.Record(HistoryEntry{
			Message: msg,
			Event:   "published",
		})
	}

	// Should only keep last 3 entries
	if history.Count() != 3 {
		t.Errorf("Expected count 3, got %d", history.Count())
	}

	entries := history.GetAll()
	// Should have entries for payloads 2, 3, 4
	if entries[0].Message.Payload() != 2 {
		t.Errorf("Expected first entry payload 2, got %v", entries[0].Message.Payload())
	}
}

func TestMessageHistoryTimeRange(t *testing.T) {
	history := NewMessageHistory(100)

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	msg1 := NewMessage(testTopic, "old")
	msg2 := NewMessage(testTopic, "recent")
	msg3 := NewMessage(testTopic, "later")

	history.Record(HistoryEntry{
		Message:   msg1,
		Event:     "published",
		Timestamp: past,
	})

	history.Record(HistoryEntry{
		Message:   msg2,
		Event:     "published",
		Timestamp: now,
	})

	history.Record(HistoryEntry{
		Message:   msg3,
		Event:     "published",
		Timestamp: future,
	})

	// Get entries from past to now
	entries := history.GetInTimeRange(past, now)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in range, got %d", len(entries))
	}

	// Get entries from now to future
	entries = history.GetInTimeRange(now, future)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in range, got %d", len(entries))
	}
}

func TestHistoryMiddleware(t *testing.T) {
	history := NewMessageHistory(100)
	middleware := HistoryMiddleware(history)

	called := false
	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		called = true
		return nil
	})

	wrappedHandler := middleware(handler)

	msg := NewMessage(testTopic, "data")
	err := wrappedHandler.Handle(context.Background(), msg)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}

	// Should have recorded delivery
	entries := history.GetByEvent("delivered")
	if len(entries) != 1 {
		t.Errorf("Expected 1 delivered entry, got %d", len(entries))
	}
}

func TestHistoryMiddlewareWithError(t *testing.T) {
	history := NewMessageHistory(100)
	middleware := HistoryMiddleware(history)

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		return context.Canceled
	})

	wrappedHandler := middleware(handler)

	msg := NewMessage(testTopic, "data")
	err := wrappedHandler.Handle(context.Background(), msg)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	// Should have recorded both delivered and failed
	delivered := history.GetByEvent("delivered")
	if len(delivered) != 1 {
		t.Errorf("Expected 1 delivered entry, got %d", len(delivered))
	}

	failed := history.GetByEvent("failed")
	if len(failed) != 1 {
		t.Errorf("Expected 1 failed entry, got %d", len(failed))
	}

	if failed[0].Error != context.Canceled.Error() {
		t.Errorf("Expected error '%v', got '%s'", context.Canceled, failed[0].Error)
	}
}

func TestAuditableBus(t *testing.T) {
	bus := New()
	history := NewMessageHistory(100)
	auditBus := NewAuditableBus(bus, history)

	received := make(chan Message, 1)
	auditBus.Subscribe(testTopic, HandlerFunc(func(ctx context.Context, msg Message) error {
		received <- msg
		return nil
	}))

	// Publish a message
	err := auditBus.Publish(context.Background(), testTopic, "test data")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for delivery
	select {
	case <-received:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Message not received")
	}

	// Check audit trail
	published := history.GetByEvent("published")
	if len(published) != 1 {
		t.Errorf("Expected 1 published entry, got %d", len(published))
	}

	if published[0].Message.Topic() != testTopic {
		t.Errorf("Expected topic '%s', got '%s'", testTopic, published[0].Message.Topic())
	}
}

func TestAuditableBusPublishError(t *testing.T) {
	// Create a bus that will be closed immediately
	bus := New()
	bus.Close()

	history := NewMessageHistory(100)
	auditBus := NewAuditableBus(bus, history)

	// Try to publish to closed bus
	err := auditBus.Publish(context.Background(), testTopic, "test data")
	if err == nil {
		t.Error("Expected error publishing to closed bus")
	}

	// Should have recorded both published and publish_failed
	published := history.GetByEvent("published")
	if len(published) != 1 {
		t.Errorf("Expected 1 published entry, got %d", len(published))
	}

	failed := history.GetByEvent("publish_failed")
	if len(failed) != 1 {
		t.Errorf("Expected 1 publish_failed entry, got %d", len(failed))
	}
}
