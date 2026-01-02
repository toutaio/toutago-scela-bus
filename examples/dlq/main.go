package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
	// Track DLQ messages
	dlqMessages := make([]scela.Message, 0)

	// Create bus with retry and DLQ configuration
	bus := scela.New(
		scela.WithMaxRetries(3),
		scela.WithDeadLetterHandler(scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
			fmt.Printf("[DLQ] Message failed after max retries: %s - %v\n", msg.Topic(), msg.Payload())
			dlqMessages = append(dlqMessages, msg)
			return nil
		})),
	)
	defer bus.Close()

	attemptCount := 0

	// Subscribe with a handler that fails first 2 times, then succeeds
	_, err := bus.Subscribe("payment.process", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		attemptCount++
		fmt.Printf("[HANDLER] Attempt %d to process payment: %v\n", attemptCount, msg.Payload())

		if attemptCount < 3 {
			return errors.New("payment gateway timeout")
		}

		fmt.Printf("[HANDLER] Payment processed successfully!\n")
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe with a handler that always fails (will go to DLQ)
	_, err = bus.Subscribe("email.send", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("[HANDLER] Attempting to send email: %v\n", msg.Payload())
		return errors.New("email service unavailable")
	}))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== Testing Retry Logic ===\n")

	// This will succeed after 3 attempts
	if err := bus.Publish(ctx, "payment.process", map[string]interface{}{
		"amount":      100.50,
		"customer_id": "CUST-123",
	}); err != nil {
		log.Fatal(err)
	}

	// Wait for retries
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n=== Testing Dead Letter Queue ===\n")

	// Reset attempt counter
	attemptCount = 0

	// This will fail and go to DLQ
	if err := bus.Publish(ctx, "email.send", map[string]interface{}{
		"to":      "user@example.com",
		"subject": "Welcome!",
	}); err != nil {
		log.Fatal(err)
	}

	// Wait for retries and DLQ
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("\n=== DLQ Summary ===\n")
	fmt.Printf("Total messages in DLQ: %d\n", len(dlqMessages))
	for i, msg := range dlqMessages {
		fmt.Printf("  %d. Topic: %s, Payload: %v\n", i+1, msg.Topic(), msg.Payload())
	}
}
