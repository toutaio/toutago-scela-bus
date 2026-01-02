package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
	// Create message history for audit trail
	history := scela.NewMessageHistory(1000)

	// Create bus with audit trail
	bus := scela.New()
	auditBus := scela.NewAuditableBus(bus, history)
	defer auditBus.Close()

	// Add history middleware to track message delivery
	auditBus.Subscribe("orders.*", scela.HistoryMiddleware(history)(scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("Processing order: %s\n", msg.Topic())
		return nil
	})))

	auditBus.Subscribe("payments.*", scela.HistoryMiddleware(history)(scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		fmt.Printf("Processing payment: %s\n", msg.Topic())
		// Simulate a failure
		if msg.Topic() == "payments.declined" {
			return fmt.Errorf("payment declined")
		}
		return nil
	})))

	ctx := context.Background()

	// Publish various messages
	fmt.Println("Publishing messages...")
	auditBus.Publish(ctx, "orders.created", "ORD-001")
	auditBus.Publish(ctx, "payments.processed", "PAY-001")
	auditBus.Publish(ctx, "payments.declined", "PAY-002")
	auditBus.Publish(ctx, "orders.completed", "ORD-001")

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Query audit trail
	fmt.Println("\n=== Audit Trail ===")
	
	fmt.Printf("\nTotal events: %d\n", history.Count())

	// Get published events
	published := history.GetByEvent("published")
	fmt.Printf("\nPublished messages: %d\n", len(published))
	for _, entry := range published {
		fmt.Printf("  - %s [%s] at %s\n", 
			entry.Message.Topic(), 
			entry.Message.ID()[:8], 
			entry.Timestamp.Format("15:04:05"))
	}

	// Get delivered events
	delivered := history.GetByEvent("delivered")
	fmt.Printf("\nDelivered messages: %d\n", len(delivered))

	// Get failed events
	failed := history.GetByEvent("failed")
	fmt.Printf("\nFailed messages: %d\n", len(failed))
	for _, entry := range failed {
		fmt.Printf("  - %s: %s\n", entry.Message.Topic(), entry.Error)
	}

	// Query by topic
	fmt.Println("\n=== Orders Topic History ===")
	orderEvents := history.GetByTopic("orders.created")
	for _, entry := range orderEvents {
		fmt.Printf("  - Event: %s at %s\n", entry.Event, entry.Timestamp.Format("15:04:05"))
	}

	// Query by time range
	fmt.Println("\n=== Recent Activity (last 5 seconds) ===")
	fiveSecondsAgo := time.Now().Add(-5 * time.Second)
	recentEvents := history.GetInTimeRange(fiveSecondsAgo, time.Now())
	fmt.Printf("Events in last 5 seconds: %d\n", len(recentEvents))

	fmt.Println("\nAudit trail example completed!")
}
