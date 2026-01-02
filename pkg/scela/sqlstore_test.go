package scela

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	return db
}

func TestNewSQLStore(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{
		DB:        db,
		TableName: "test_messages",
	})

	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	if store == nil {
		t.Fatal("Store is nil")
	}

	// Verify table was created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_messages'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Table was not created: %v", err)
	}

	if tableName != "test_messages" {
		t.Errorf("Expected table name 'test_messages', got '%s'", tableName)
	}
}

func TestSQLStoreStoreAndLoad(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	// Store some messages
	msg1 := NewMessage("test.topic1", "payload1")
	msg2 := NewMessage("test.topic2", map[string]interface{}{"key": "value"})

	if err := store.Store(ctx, msg1); err != nil {
		t.Fatalf("Failed to store msg1: %v", err)
	}

	if err := store.Store(ctx, msg2); err != nil {
		t.Fatalf("Failed to store msg2: %v", err)
	}

	// Load messages
	messages, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load messages: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Verify first message
	if messages[0].Topic() != "test.topic1" {
		t.Errorf("Expected topic 'test.topic1', got '%s'", messages[0].Topic())
	}

	if messages[0].Payload() != "payload1" {
		t.Errorf("Expected payload 'payload1', got '%v'", messages[0].Payload())
	}

	// Verify second message
	if messages[1].Topic() != "test.topic2" {
		t.Errorf("Expected topic 'test.topic2', got '%s'", messages[1].Topic())
	}
}

func TestSQLStoreLoadByTopic(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	// Store messages with different topics
	msg1 := NewMessage("orders.created", "order1")
	msg2 := NewMessage("orders.updated", "order2")
	msg3 := NewMessage("orders.created", "order3")
	msg4 := NewMessage("users.created", "user1")

	store.Store(ctx, msg1)
	store.Store(ctx, msg2)
	store.Store(ctx, msg3)
	store.Store(ctx, msg4)

	// Load only "orders.created" messages
	messages, err := store.LoadByTopic(ctx, "orders.created")
	if err != nil {
		t.Fatalf("Failed to load by topic: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	for _, msg := range messages {
		if msg.Topic() != "orders.created" {
			t.Errorf("Expected topic 'orders.created', got '%s'", msg.Topic())
		}
	}
}

func TestSQLStoreLoadAfter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	now := time.Now()
	past := now.Add(-1 * time.Hour)

	// Store messages
	msg1 := NewMessage("test.topic", "old")
	store.Store(ctx, msg1)

	time.Sleep(100 * time.Millisecond)
	marker := time.Now()
	time.Sleep(100 * time.Millisecond)

	msg2 := NewMessage("test.topic", "recent")
	store.Store(ctx, msg2)

	// Load messages after marker
	messages, err := store.LoadAfter(ctx, marker)
	if err != nil {
		t.Fatalf("Failed to load after: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Payload() != "recent" {
		t.Errorf("Expected payload 'recent', got '%v'", messages[0].Payload())
	}

	// Load all messages after past
	messages, err = store.LoadAfter(ctx, past)
	if err != nil {
		t.Fatalf("Failed to load after past: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}
}

func TestSQLStoreClear(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	// Store messages
	msg1 := NewMessage("test.topic", "data1")
	msg2 := NewMessage("test.topic", "data2")

	store.Store(ctx, msg1)
	store.Store(ctx, msg2)

	// Verify stored
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Clear
	if err := store.Clear(ctx); err != nil {
		t.Fatalf("Failed to clear: %v", err)
	}

	// Verify cleared
	count, err = store.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count after clear: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 after clear, got %d", count)
	}
}

func TestSQLStoreClearBefore(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	// Store messages with delays
	msg1 := NewMessage("test.topic", "old")
	store.Store(ctx, msg1)

	time.Sleep(50 * time.Millisecond)
	marker := time.Now()
	time.Sleep(50 * time.Millisecond)

	msg2 := NewMessage("test.topic", "new")
	store.Store(ctx, msg2)

	// Verify we have 2 messages
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Clear messages before marker
	if err := store.ClearBefore(ctx, marker); err != nil {
		t.Fatalf("Failed to clear before: %v", err)
	}

	// Should have 1 message left
	count, err = store.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count after clear: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1 after clear, got %d", count)
	}

	// Verify it's the new message
	messages, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Payload() != "new" {
		t.Errorf("Expected payload 'new', got '%v'", messages[0].Payload())
	}
}

func TestSQLStoreCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	// Initially empty
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Add messages
	for i := 0; i < 5; i++ {
		msg := NewMessage("test.topic", i)
		store.Store(ctx, msg)
	}

	count, err = store.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestSQLStoreWithMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store, err := NewSQLStore(SQLStoreConfig{DB: db})
	if err != nil {
		t.Fatalf("Failed to create SQL store: %v", err)
	}

	ctx := context.Background()

	// Create message with metadata
	msg := NewMessage("test.topic", "data")
	msg.Metadata()["key1"] = "value1"
	msg.Metadata()["key2"] = 123

	// Store
	if err := store.Store(ctx, msg); err != nil {
		t.Fatalf("Failed to store: %v", err)
	}

	// Load
	messages, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Verify metadata
	loadedMsg := messages[0]
	if loadedMsg.Metadata()["key1"] != "value1" {
		t.Errorf("Expected metadata key1='value1', got '%v'", loadedMsg.Metadata()["key1"])
	}

	// Note: JSON unmarshaling will convert numbers to float64
	if loadedMsg.Metadata()["key2"] != float64(123) {
		t.Errorf("Expected metadata key2=123, got '%v'", loadedMsg.Metadata()["key2"])
	}
}
