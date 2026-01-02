package main

import (
"context"
"fmt"
"log"
"sync/atomic"
"time"

"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

// MetricsObserver tracks bus metrics
type MetricsObserver struct {
publishCount   int64
processedCount int64
errorCount     int64
subscribeCount int64
}

func (m *MetricsObserver) OnPublish(ctx context.Context, topic string, msg scela.Message) {
atomic.AddInt64(&m.publishCount, 1)
}

func (m *MetricsObserver) OnSubscribe(pattern string) {
atomic.AddInt64(&m.subscribeCount, 1)
fmt.Printf("[METRICS] New subscription: %s\n", pattern)
}

func (m *MetricsObserver) OnUnsubscribe(pattern string) {
fmt.Printf("[METRICS] Unsubscribed: %s\n", pattern)
}

func (m *MetricsObserver) OnMessageProcessed(ctx context.Context, msg scela.Message, err error) {
atomic.AddInt64(&m.processedCount, 1)
if err != nil {
atomic.AddInt64(&m.errorCount, 1)
}
}

func (m *MetricsObserver) OnClose() {
fmt.Println("[METRICS] Bus closed")
}

func (m *MetricsObserver) PrintStats() {
fmt.Println("\n=== Metrics Summary ===")
fmt.Printf("Published: %d\n", atomic.LoadInt64(&m.publishCount))
fmt.Printf("Processed: %d\n", atomic.LoadInt64(&m.processedCount))
fmt.Printf("Errors: %d\n", atomic.LoadInt64(&m.errorCount))
fmt.Printf("Subscriptions: %d\n", atomic.LoadInt64(&m.subscribeCount))
}

func main() {
metrics := &MetricsObserver{}

// Create bus with metrics observer
bus := scela.New(scela.WithObserver(metrics))
defer bus.Close()

// Subscribe to events
_, err := bus.Subscribe("user.*", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
fmt.Printf("[HANDLER] Processing: %s\n", msg.Topic())
time.Sleep(10 * time.Millisecond)
return nil
}))
if err != nil {
log.Fatal(err)
}

_, err = bus.Subscribe("order.*", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
fmt.Printf("[HANDLER] Processing: %s\n", msg.Topic())
return nil
}))
if err != nil {
log.Fatal(err)
}

ctx := context.Background()

fmt.Println("=== Publishing messages ===\n")

// Publish messages
topics := []string{
"user.created",
"user.updated",
"order.created",
"order.shipped",
"user.deleted",
}

for _, topic := range topics {
if err := bus.PublishSync(ctx, topic, map[string]interface{}{
"id": fmt.Sprintf("item-%s", topic),
}); err != nil {
log.Fatal(err)
}
}

// Print metrics
metrics.PrintStats()
}
