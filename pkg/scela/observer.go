package scela

import (
	"context"
	"sync"
)

// Observer is called when bus events occur.
type Observer interface {
	OnPublish(ctx context.Context, topic string, msg Message)
	OnSubscribe(pattern string)
	OnUnsubscribe(pattern string)
	OnMessageProcessed(ctx context.Context, msg Message, err error)
	OnClose()
}

// ObserverFunc is a function adapter for Observer interface.
type observerRegistry struct {
	mu        sync.RWMutex
	observers []Observer
}

func newObserverRegistry() *observerRegistry {
	return &observerRegistry{
		observers: make([]Observer, 0),
	}
}

func (r *observerRegistry) Add(observer Observer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.observers = append(r.observers, observer)
}

func (r *observerRegistry) NotifyPublish(ctx context.Context, topic string, msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, obs := range r.observers {
		obs.OnPublish(ctx, topic, msg)
	}
}

func (r *observerRegistry) NotifySubscribe(pattern string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, obs := range r.observers {
		obs.OnSubscribe(pattern)
	}
}

func (r *observerRegistry) NotifyUnsubscribe(pattern string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, obs := range r.observers {
		obs.OnUnsubscribe(pattern)
	}
}

func (r *observerRegistry) NotifyMessageProcessed(ctx context.Context, msg Message, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, obs := range r.observers {
		obs.OnMessageProcessed(ctx, msg, err)
	}
}

func (r *observerRegistry) NotifyClose() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, obs := range r.observers {
		obs.OnClose()
	}
}

// WithObserver adds an observer to the bus.
func WithObserver(observer Observer) Option {
	return func(b *bus) {
		b.observers.Add(observer)
	}
}
