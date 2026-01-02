package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
	// Create a new message bus
	bus := scela.New()
	defer bus.Close()

	// Subscribe to user events with wildcard
	_, err := bus.Subscribe("user.*", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("[Handler 1] Received %s: %v\n", msg.Topic(), msg.Payload())
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe specifically to user.created
	_, err = bus.Subscribe("user.created", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("[Handler 2] New user created: %v\n", msg.Payload())
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Publish messages synchronously
	fmt.Println("\n=== Synchronous Publishing ===")
	if err := bus.PublishSync(ctx, "user.created", map[string]interface{}{
		"id":    "123",
		"email": "alice@example.com",
		"name":  "Alice",
	}); err != nil {
		log.Fatal(err)
	}

	if err := bus.PublishSync(ctx, "user.updated", map[string]interface{}{
		"id":    "123",
		"email": "alice.updated@example.com",
	}); err != nil {
		log.Fatal(err)
	}

	// Publish messages asynchronously
	fmt.Println("\n=== Asynchronous Publishing ===")
	for i := 0; i < 3; i++ {
		if err := bus.Publish(ctx, "user.deleted", map[string]interface{}{
			"id": fmt.Sprintf("user-%d", i),
		}); err != nil {
			log.Fatal(err)
		}
	}

	// Wait a bit for async messages to process
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== Done ===")
}
