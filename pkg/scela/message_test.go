package scela

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	topic := "test.topic"
	payload := "test payload"

	msg := NewMessage(topic, payload)

	if msg.Topic() != topic {
		t.Errorf("Topic() = %v, want %v", msg.Topic(), topic)
	}

	if msg.Payload() != payload {
		t.Errorf("Payload() = %v, want %v", msg.Payload(), payload)
	}

	if msg.ID() == "" {
		t.Error("ID() returned empty string")
	}

	if msg.Timestamp().IsZero() {
		t.Error("Timestamp() is zero")
	}

	if len(msg.Metadata()) != 0 {
		t.Errorf("Metadata() length = %d, want 0", len(msg.Metadata()))
	}
}

func TestNewMessageWithPriority(t *testing.T) {
	msg := NewMessageWithPriority("test", "payload", PriorityHigh)

	impl, ok := msg.(*message)
	if !ok {
		t.Fatal("NewMessageWithPriority() did not return *message")
	}

	if impl.Priority() != PriorityHigh {
		t.Errorf("Priority() = %v, want %v", impl.Priority(), PriorityHigh)
	}
}

func TestMessage_Metadata(t *testing.T) {
	msg := NewMessage("test", nil)

	metadata := msg.Metadata()
	metadata["key1"] = "value1"
	metadata["key2"] = 42

	// Verify metadata persists
	if len(msg.Metadata()) != 2 {
		t.Errorf("Metadata() length = %d, want 2", len(msg.Metadata()))
	}

	if msg.Metadata()["key1"] != "value1" {
		t.Errorf("Metadata()[key1] = %v, want value1", msg.Metadata()["key1"])
	}
}

func TestMessage_Timestamp(t *testing.T) {
	before := time.Now()
	msg := NewMessage("test", nil)
	after := time.Now()

	ts := msg.Timestamp()

	if ts.Before(before) || ts.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", ts, before, after)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("generateID() returned empty string")
	}

	if id1 == id2 {
		t.Error("generateID() returned duplicate IDs")
	}

	// ID should be hex string (32 characters for 16 bytes)
	if len(id1) != 32 {
		t.Errorf("generateID() length = %d, want 32", len(id1))
	}
}

func BenchmarkNewMessage(b *testing.B) {
	payload := map[string]interface{}{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMessage("test.topic", payload)
	}
}

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateID()
	}
}
