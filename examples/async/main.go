package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/toutaio/toutago-scela-bus/pkg/scela"
)

func main() {
	// Create a bus
	bus := scela.New(scela.WithWorkers(5))
	defer bus.Close()

	var wg sync.WaitGroup
	processed := make(map[string]int)
	var mu sync.Mutex

	// Subscribe to various topics
	topics := []string{"logs", "metrics", "events", "analytics"}

	for _, topic := range topics {
		t := topic // Capture for closure
		_, err := bus.Subscribe(t, scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
			// Simulate processing time
			time.Sleep(10 * time.Millisecond)

			mu.Lock()
			processed[t]++
			count := processed[t]
			mu.Unlock()

			if count%10 == 0 {
				fmt.Printf("[%s] Processed %d messages\n", t, count)
			}

			return nil
		}))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Subscribe to all messages with wildcard
	_, err := bus.Subscribe("*", scela.HandlerFunc(func(ctx context.Context, msg scela.Message) error {
		mu.Lock()
		processed["all"]++
		mu.Unlock()
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== Publishing messages asynchronously ===\n")

	// Publish many messages asynchronously
	start := time.Now()
	numMessages := 100

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			topic := topics[idx%len(topics)]
			payload := map[string]interface{}{
				"index": idx,
				"data":  fmt.Sprintf("Message %d", idx),
			}

			if err := bus.Publish(ctx, topic, payload); err != nil {
				log.Printf("Error publishing: %v", err)
			}
		}(i)
	}

	// Wait for all publishes to complete
	wg.Wait()

	// Wait for async processing
	time.Sleep(2 * time.Second)

	elapsed := time.Since(start)

	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Published %d messages in %v\n", numMessages, elapsed)
	fmt.Printf("Throughput: %.0f msg/sec\n\n", float64(numMessages)/elapsed.Seconds())

	mu.Lock()
	fmt.Println("Messages processed by topic:")
	for topic, count := range processed {
		fmt.Printf("  %s: %d\n", topic, count)
	}
	mu.Unlock()
}
