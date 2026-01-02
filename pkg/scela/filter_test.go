package scela

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFilterMiddleware(t *testing.T) {
	bus := New()
	defer bus.Close()

	var received []string
	var mu sync.Mutex

	handler := HandlerFunc(func(ctx context.Context, msg Message) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.Topic())
		return nil
	})

	// Add filter that only accepts "user.created" topic
	filter := TopicFilter("user.created")
	bus.Use(FilterMiddleware(filter))

	_, err := bus.Subscribe("*", handler)
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	ctx := context.Background()

	// Publish multiple topics
	bus.PublishSync(ctx, "user.created", "data1")
	bus.PublishSync(ctx, "user.updated", "data2")
	bus.PublishSync(ctx, "user.created", "data3")
	bus.PublishSync(ctx, "user.deleted", "data4")

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Only user.created should be received
	if len(received) != 2 {
		t.Errorf("Expected 2 messages, got %d: %v", len(received), received)
	}

	for _, topic := range received {
		if topic != "user.created" {
			t.Errorf("Unexpected topic: %s", topic)
		}
	}
}

func TestMetadataFilter(t *testing.T) {
	filter := MetadataFilter("priority", "high")

	msg := NewMessage("test", "data")
	// Metadata is empty by default
	if filter(msg) {
		t.Error("Expected filter to fail for message without metadata")
	}
}

func TestAndFilter(t *testing.T) {
	filter1 := TopicFilter("user.created", "user.updated")
	filter2 := func(msg Message) bool {
		return msg.Payload() != nil
	}

	combined := AndFilter(filter1, filter2)

	msg1 := NewMessage("user.created", "data")
	if !combined(msg1) {
		t.Error("Expected filter to pass for user.created with data")
	}

	msg2 := NewMessage("user.deleted", "data")
	if combined(msg2) {
		t.Error("Expected filter to fail for user.deleted")
	}

	msg3 := NewMessage("user.created", nil)
	if combined(msg3) {
		t.Error("Expected filter to fail for nil payload")
	}
}

func TestOrFilter(t *testing.T) {
	filter1 := TopicFilter("user.created")
	filter2 := TopicFilter("user.updated")

	combined := OrFilter(filter1, filter2)

	msg1 := NewMessage("user.created", "data")
	if !combined(msg1) {
		t.Error("Expected filter to pass for user.created")
	}

	msg2 := NewMessage("user.updated", "data")
	if !combined(msg2) {
		t.Error("Expected filter to pass for user.updated")
	}

	msg3 := NewMessage("user.deleted", "data")
	if combined(msg3) {
		t.Error("Expected filter to fail for user.deleted")
	}
}

func TestNotFilter(t *testing.T) {
	filter := TopicFilter("user.deleted")
	inverted := NotFilter(filter)

	msg1 := NewMessage("user.deleted", "data")
	if inverted(msg1) {
		t.Error("Expected inverted filter to fail for user.deleted")
	}

	msg2 := NewMessage("user.created", "data")
	if !inverted(msg2) {
		t.Error("Expected inverted filter to pass for user.created")
	}
}

func TestPayloadFilter(t *testing.T) {
	// Filter for string payloads only
	filter := PayloadFilter(func(p interface{}) bool {
		_, ok := p.(string)
		return ok
	})

	msg1 := NewMessage("test", "string data")
	if !filter(msg1) {
		t.Error("Expected filter to pass for string payload")
	}

	msg2 := NewMessage("test", 123)
	if filter(msg2) {
		t.Error("Expected filter to fail for int payload")
	}
}
