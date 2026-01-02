package scela

import (
"context"
"sync"
"time"
)

// Batch represents a collection of messages.
type Batch struct {
Messages []Message
mu       sync.RWMutex
}

// NewBatch creates a new batch.
func NewBatch() *Batch {
return &Batch{
Messages: make([]Message, 0),
}
}

// Add adds a message to the batch.
func (b *Batch) Add(msg Message) {
b.mu.Lock()
defer b.mu.Unlock()
b.Messages = append(b.Messages, msg)
}

// Size returns the number of messages in the batch.
func (b *Batch) Size() int {
b.mu.RLock()
defer b.mu.RUnlock()
return len(b.Messages)
}

// Clear removes all messages from the batch.
func (b *Batch) Clear() []Message {
b.mu.Lock()
defer b.mu.Unlock()
messages := b.Messages
b.Messages = make([]Message, 0)
return messages
}

// BatchPublisher collects messages and publishes them in batches.
type BatchPublisher struct {
bus       Bus
batch     *Batch
maxSize   int
maxWait   time.Duration
mu        sync.Mutex
timer     *time.Timer
done      chan struct{}
wg        sync.WaitGroup
onPublish func(messages []Message)
}

// BatchPublisherOption is a functional option for configuring a batch publisher.
type BatchPublisherOption func(*BatchPublisher)

// WithBatchSize sets the maximum batch size.
func WithBatchSize(size int) BatchPublisherOption {
return func(bp *BatchPublisher) {
if size > 0 {
bp.maxSize = size
}
}
}

// WithBatchWait sets the maximum wait time before publishing a batch.
func WithBatchWait(wait time.Duration) BatchPublisherOption {
return func(bp *BatchPublisher) {
if wait > 0 {
bp.maxWait = wait
}
}
}

// WithBatchCallback sets a callback to be called when a batch is published.
func WithBatchCallback(fn func(messages []Message)) BatchPublisherOption {
return func(bp *BatchPublisher) {
bp.onPublish = fn
}
}

// NewBatchPublisher creates a new batch publisher.
func NewBatchPublisher(bus Bus, opts ...BatchPublisherOption) *BatchPublisher {
bp := &BatchPublisher{
bus:     bus,
batch:   NewBatch(),
maxSize: 100,
maxWait: 1 * time.Second,
done:    make(chan struct{}),
}

for _, opt := range opts {
opt(bp)
}

bp.timer = time.NewTimer(bp.maxWait)
bp.wg.Add(1)
go bp.processTimer()

return bp
}

// Publish adds a message to the batch.
func (bp *BatchPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
bp.mu.Lock()
defer bp.mu.Unlock()

msg := NewMessage(topic, payload)
bp.batch.Add(msg)

if bp.batch.Size() >= bp.maxSize {
bp.flush(ctx)
}

return nil
}

// Flush publishes all pending messages immediately.
func (bp *BatchPublisher) Flush(ctx context.Context) error {
bp.mu.Lock()
defer bp.mu.Unlock()
return bp.flush(ctx)
}

// flush publishes the current batch (must be called with lock held).
func (bp *BatchPublisher) flush(ctx context.Context) error {
if bp.batch.Size() == 0 {
return nil
}

messages := bp.batch.Clear()

// Reset timer
if !bp.timer.Stop() {
select {
case <-bp.timer.C:
default:
}
}
bp.timer.Reset(bp.maxWait)

// Publish all messages
for _, msg := range messages {
if err := bp.bus.Publish(ctx, msg.Topic(), msg.Payload()); err != nil {
return err
}
}

// Call callback if set
if bp.onPublish != nil {
bp.onPublish(messages)
}

return nil
}

// processTimer handles periodic flushing.
func (bp *BatchPublisher) processTimer() {
defer bp.wg.Done()

for {
select {
case <-bp.timer.C:
ctx := context.Background()
bp.Flush(ctx)
bp.timer.Reset(bp.maxWait)
case <-bp.done:
return
}
}
}

// Close stops the batch publisher and flushes any remaining messages.
func (bp *BatchPublisher) Close() error {
close(bp.done)
bp.wg.Wait()

ctx := context.Background()
return bp.Flush(ctx)
}
