package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/toutaio/toutago-scela-bus/pkg/scela"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Create SQLite database
	db, err := sql.Open("sqlite3", "./messages.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create SQL store
	store, err := scela.NewSQLStore(scela.SQLStoreConfig{
		DB:        db,
		TableName: "messages",
	})
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Create bus with persistence
	bus := scela.New()
	persistentBus := scela.NewPersistentBus(bus, store)
	defer persistentBus.Close()

	// Subscribe to messages
	persistentBus.Subscribe("orders.*", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("Received: %s - %v\n", msg.Topic(), msg.Payload())
		return nil
	}))

	ctx := context.Background()

	// Publish some messages (they will be persisted to database)
	fmt.Println("Publishing messages...")
	persistentBus.Publish(ctx, "orders.created", map[string]interface{}{
		"order_id": "ORD-001",
		"amount":   99.99,
	})

	persistentBus.Publish(ctx, "orders.updated", map[string]interface{}{
		"order_id": "ORD-001",
		"status":   "shipped",
	})

	persistentBus.Publish(ctx, "orders.completed", map[string]interface{}{
		"order_id": "ORD-001",
	})

	// Wait for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Check how many messages are stored
	count, err := store.Count(ctx)
	if err == nil {
		fmt.Printf("\nMessages in database: %d\n", count)
	}

	// Load messages from database
	messages, err := store.LoadByTopic(ctx, "orders.created")
	if err == nil {
		fmt.Printf("\nLoaded %d 'orders.created' messages from database\n", len(messages))
	}

	fmt.Println("\n--- Simulating application restart ---")

	// Replay all persisted messages
	fmt.Println("\nReplaying persisted messages...")
	if err := persistentBus.Replay(ctx); err != nil {
		log.Fatalf("Failed to replay: %v", err)
	}

	// Wait for replay to complete
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nExample completed!")
}
