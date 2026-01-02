package main

import (
"context"
"fmt"
"log"
"time"

"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
// Create bus
bus := scela.New(scela.WithWorkers(2))
defer bus.Close()

// Subscribe to all messages
_, err := bus.Subscribe("*", scela.HandlerFunc(
func(ctx context.Context, msg scela.Message) error {
fmt.Printf("[%s] Processed: %s\n", time.Now().Format("15:04:05.000"), msg.Payload())
// Simulate processing time
time.Sleep(100 * time.Millisecond)
return nil
},
))
if err != nil {
log.Fatal(err)
}

ctx := context.Background()

fmt.Println("Publishing messages with different priorities...")
fmt.Println()

// Publish messages in mixed priority order
bus.PublishWithPriority(ctx, "task", "Low priority task 1", scela.PriorityLow)
bus.PublishWithPriority(ctx, "task", "Normal priority task 1", scela.PriorityNormal)
bus.PublishWithPriority(ctx, "task", "URGENT: Critical task!", scela.PriorityUrgent)
bus.PublishWithPriority(ctx, "task", "High priority task 1", scela.PriorityHigh)
bus.PublishWithPriority(ctx, "task", "Low priority task 2", scela.PriorityLow)
bus.PublishWithPriority(ctx, "task", "Normal priority task 2", scela.PriorityNormal)
bus.PublishWithPriority(ctx, "task", "High priority task 2", scela.PriorityHigh)

fmt.Println("Messages published. Processing...")
fmt.Println()

// Wait for all messages to be processed
time.Sleep(2 * time.Second)

fmt.Println()
fmt.Println("Note: Messages are processed based on priority:")
fmt.Println("  Urgent > High > Normal > Low")
}
