package scela

import (
"context"
"os"
"testing"
"time"
)

func TestInMemoryStore(t *testing.T) {
store := NewInMemoryStore(100)
ctx := context.Background()

msg1 := NewMessage("test1", "data1")
msg2 := NewMessage("test2", "data2")

// Store messages
if err := store.Store(ctx, msg1); err != nil {
t.Fatalf("Store() error = %v", err)
}
if err := store.Store(ctx, msg2); err != nil {
t.Fatalf("Store() error = %v", err)
}

// Load messages
messages, err := store.Load(ctx)
if err != nil {
t.Fatalf("Load() error = %v", err)
}

if len(messages) != 2 {
t.Errorf("Expected 2 messages, got %d", len(messages))
}

// Clear messages
if err := store.Clear(ctx); err != nil {
t.Fatalf("Clear() error = %v", err)
}

messages, _ = store.Load(ctx)
if len(messages) != 0 {
t.Errorf("Expected 0 messages after clear, got %d", len(messages))
}
}

func TestInMemoryStore_MaxSize(t *testing.T) {
store := NewInMemoryStore(5)
ctx := context.Background()

// Store 10 messages
for i := 0; i < 10; i++ {
msg := NewMessage("test", i)
store.Store(ctx, msg)
}

messages, _ := store.Load(ctx)
if len(messages) != 5 {
t.Errorf("Expected max 5 messages, got %d", len(messages))
}
}

func TestFileStore(t *testing.T) {
filepath := "test_messages.json"
defer os.Remove(filepath)

store := NewFileStore(filepath)
defer store.Close()

ctx := context.Background()

msg1 := NewMessage("test1", "data1")
msg2 := NewMessage("test2", map[string]string{"key": "value"})

// Store messages
if err := store.Store(ctx, msg1); err != nil {
t.Fatalf("Store() error = %v", err)
}
if err := store.Store(ctx, msg2); err != nil {
t.Fatalf("Store() error = %v", err)
}

// Load messages
messages, err := store.Load(ctx)
if err != nil {
t.Fatalf("Load() error = %v", err)
}

if len(messages) != 2 {
t.Errorf("Expected 2 messages, got %d", len(messages))
}

// Verify persistence (create new store instance)
store2 := NewFileStore(filepath)
messages2, err := store2.Load(ctx)
if err != nil {
t.Fatalf("Load() from new store error = %v", err)
}

if len(messages2) != 2 {
t.Errorf("Expected 2 persisted messages, got %d", len(messages2))
}
}

func TestPersistentBus(t *testing.T) {
bus := New()
defer bus.Close()

store := NewInMemoryStore(100)
pbus := NewPersistentBus(bus, store)

ctx := context.Background()

// Publish messages
pbus.Publish(ctx, "test1", "data1")
pbus.Publish(ctx, "test2", "data2")

// Verify messages were stored
messages, err := store.Load(ctx)
if err != nil {
t.Fatalf("Load() error = %v", err)
}

if len(messages) != 2 {
t.Errorf("Expected 2 stored messages, got %d", len(messages))
}
}

func TestReplayableStore(t *testing.T) {
store := NewInMemoryStore(100)
ctx := context.Background()

// Store some old messages
oldMsg := NewMessage("old", "data")
store.Store(ctx, oldMsg)

// Wait a bit
time.Sleep(10 * time.Millisecond)
cutoff := time.Now()
time.Sleep(10 * time.Millisecond)

// Store newer messages
newMsg := NewMessage("new", "data")
store.Store(ctx, newMsg)

// Create replayable store with cutoff time
replayStore := NewReplayableStore(store, cutoff)

// Should only load messages after cutoff
messages, err := replayStore.Load(ctx)
if err != nil {
t.Fatalf("Load() error = %v", err)
}

// Should get at least the new message
if len(messages) < 1 {
t.Errorf("Expected at least 1 message after cutoff, got %d", len(messages))
}
}
