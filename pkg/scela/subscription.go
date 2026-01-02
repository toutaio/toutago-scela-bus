package scela

import (
	"fmt"
	"sync"
)

// subscription implements the Subscription interface.
type subscription struct {
	id      string
	pattern string
	handler Handler
	bus     *bus
}

// Topic returns the subscription pattern.
func (s *subscription) Topic() string {
	return s.pattern
}

// Unsubscribe removes the subscription from the bus.
func (s *subscription) Unsubscribe() error {
	return s.bus.unsubscribe(s.id)
}

// subscriptionRegistry manages all subscriptions.
type subscriptionRegistry struct {
	mu            sync.RWMutex
	subscriptions map[string]*subscription // id -> subscription
	patterns      map[string][]string      // pattern -> []subscription IDs
	matcher       *patternMatcher
}

// newSubscriptionRegistry creates a new subscription registry.
func newSubscriptionRegistry() *subscriptionRegistry {
	return &subscriptionRegistry{
		subscriptions: make(map[string]*subscription),
		patterns:      make(map[string][]string),
		matcher:       newPatternMatcher(),
	}
}

// Add adds a new subscription.
func (sr *subscriptionRegistry) Add(pattern string, handler Handler, bus *bus) (*subscription, error) {
	if pattern == "" {
		return nil, fmt.Errorf("subscription pattern cannot be empty")
	}
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}

	sub := &subscription{
		id:      generateID(),
		pattern: pattern,
		handler: handler,
		bus:     bus,
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.subscriptions[sub.id] = sub
	sr.patterns[pattern] = append(sr.patterns[pattern], sub.id)

	return sub, nil
}

// Remove removes a subscription by ID.
func (sr *subscriptionRegistry) Remove(id string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sub, exists := sr.subscriptions[id]
	if !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	// Remove from subscriptions
	delete(sr.subscriptions, id)

	// Remove from patterns
	pattern := sub.pattern
	ids := sr.patterns[pattern]
	for i, sid := range ids {
		if sid == id {
			sr.patterns[pattern] = append(ids[:i], ids[i+1:]...)
			break
		}
	}

	// Clean up empty pattern list
	if len(sr.patterns[pattern]) == 0 {
		delete(sr.patterns, pattern)
	}

	return nil
}

// GetHandlers returns all handlers that match the topic.
func (sr *subscriptionRegistry) GetHandlers(topic string) []Handler {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var handlers []Handler
	seen := make(map[string]bool)

	// Check each pattern for matches
	for pattern, ids := range sr.patterns {
		if sr.matcher.Match(pattern, topic) {
			for _, id := range ids {
				if !seen[id] {
					if sub, ok := sr.subscriptions[id]; ok {
						handlers = append(handlers, sub.handler)
						seen[id] = true
					}
				}
			}
		}
	}

	return handlers
}

// Count returns the total number of subscriptions.
func (sr *subscriptionRegistry) Count() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return len(sr.subscriptions)
}

// Clear removes all subscriptions.
func (sr *subscriptionRegistry) Clear() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.subscriptions = make(map[string]*subscription)
	sr.patterns = make(map[string][]string)
}
