package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
	// Create bus with middleware
	bus := scela.New()
	defer bus.Close()

	// Add logging middleware
	loggingMiddleware := func(next scela.Handler) scela.Handler {
		return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
			start := time.Now()
			fmt.Printf("[LOG] Processing message %s on topic %s\n", msg.ID(), msg.Topic())

			err := next.Handle(ctx, msg)

			duration := time.Since(start)
			if err != nil {
				fmt.Printf("[LOG] Failed after %v: %v\n", duration, err)
			} else {
				fmt.Printf("[LOG] Completed in %v\n", duration)
			}

			return err
		})
	}

	// Add metrics middleware
	metricsMiddleware := func(next scela.Handler) scela.Handler {
		return scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
			// Add to message metadata
			msg.Metadata()["processed_at"] = time.Now()

			err := next.Handle(ctx, msg)

			msg.Metadata()["completed_at"] = time.Now()
			msg.Metadata()["success"] = err == nil

			return err
		})
	}

	// Register middleware (executes in order: logging -> metrics -> handler)
	bus.Use(loggingMiddleware, metricsMiddleware)

	// Subscribe to events
	_, err := bus.Subscribe("order.*", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("[HANDLER] Processing order event: %v\n", msg.Payload())

		// Simulate some work
		time.Sleep(50 * time.Millisecond)

		// Print metadata added by middleware
		fmt.Printf("[HANDLER] Metadata: %v\n", msg.Metadata())

		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Publish messages
	fmt.Println("=== Publishing messages with middleware ===\n")

	if err := bus.PublishSync(ctx, "order.created", map[string]interface{}{
		"order_id": "ORD-001",
		"amount":   99.99,
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Println()

	if err := bus.PublishSync(ctx, "order.completed", map[string]interface{}{
		"order_id": "ORD-001",
		"status":   "delivered",
	}); err != nil {
		log.Fatal(err)
	}
}
