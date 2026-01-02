package scela

import (
	"context"
	"sync"
	"time"
)

// MessageHistory provides message history and audit trail capabilities.
type MessageHistory struct {
	entries []HistoryEntry
	mu      sync.RWMutex
	maxSize int
}

// HistoryEntry represents a single entry in the message history.
type HistoryEntry struct {
	Message      Message
	Event        string // "published", "delivered", "failed", "retried"
	Timestamp    time.Time
	Metadata     map[string]interface{}
	SubscriberID string
	Error        string
}

// NewMessageHistory creates a new message history tracker.
func NewMessageHistory(maxSize int) *MessageHistory {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &MessageHistory{
		entries: make([]HistoryEntry, 0),
		maxSize: maxSize,
	}
}

// Record adds a new entry to the history.
func (h *MessageHistory) Record(entry HistoryEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	h.entries = append(h.entries, entry)

	// Trim if exceeded max size
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
}

// GetAll returns all history entries.
func (h *MessageHistory) GetAll() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]HistoryEntry, len(h.entries))
	copy(result, h.entries)
	return result
}

// GetByMessageID returns all history entries for a specific message.
func (h *MessageHistory) GetByMessageID(messageID string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]HistoryEntry, 0)
	for _, entry := range h.entries {
		if entry.Message.ID() == messageID {
			result = append(result, entry)
		}
	}
	return result
}

// GetByTopic returns all history entries for a specific topic.
func (h *MessageHistory) GetByTopic(topic string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]HistoryEntry, 0)
	for _, entry := range h.entries {
		if entry.Message.Topic() == topic {
			result = append(result, entry)
		}
	}
	return result
}

// GetByEvent returns all history entries for a specific event type.
func (h *MessageHistory) GetByEvent(event string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]HistoryEntry, 0)
	for _, entry := range h.entries {
		if entry.Event == event {
			result = append(result, entry)
		}
	}
	return result
}

// GetInTimeRange returns history entries within a time range.
func (h *MessageHistory) GetInTimeRange(start, end time.Time) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]HistoryEntry, 0)
	for _, entry := range h.entries {
		if (entry.Timestamp.After(start) || entry.Timestamp.Equal(start)) &&
			(entry.Timestamp.Before(end) || entry.Timestamp.Equal(end)) {
			result = append(result, entry)
		}
	}
	return result
}

// Clear removes all history entries.
func (h *MessageHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]HistoryEntry, 0)
}

// Count returns the number of history entries.
func (h *MessageHistory) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.entries)
}

// HistoryMiddleware creates a middleware that records message history.
func HistoryMiddleware(history *MessageHistory) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Message) error {
			// Record publication
			history.Record(HistoryEntry{
				Message:   msg,
				Event:     "delivered",
				Timestamp: time.Now(),
			})

			// Execute handler
			err := next.Handle(ctx, msg)

			// Record result
			if err != nil {
				history.Record(HistoryEntry{
					Message:   msg,
					Event:     "failed",
					Timestamp: time.Now(),
					Error:     err.Error(),
				})
			}

			return err
		})
	}
}

// AuditableBus wraps a bus with audit trail capabilities.
type AuditableBus struct {
	Bus
	history *MessageHistory
}

// NewAuditableBus creates a new auditable bus.
func NewAuditableBus(bus Bus, history *MessageHistory) *AuditableBus {
	return &AuditableBus{
		Bus:     bus,
		history: history,
	}
}

// Publish publishes a message and records it in the audit trail.
func (ab *AuditableBus) Publish(ctx context.Context, topic string, payload interface{}) error {
	msg := NewMessage(topic, payload)

	// Record publication
	ab.history.Record(HistoryEntry{
		Message:   msg,
		Event:     "published",
		Timestamp: time.Now(),
	})

	// Publish
	err := ab.Bus.Publish(ctx, topic, payload)
	if err != nil {
		ab.history.Record(HistoryEntry{
			Message:   msg,
			Event:     "publish_failed",
			Timestamp: time.Now(),
			Error:     err.Error(),
		})
	}

	return err
}

// GetHistory returns the audit history.
func (ab *AuditableBus) GetHistory() *MessageHistory {
	return ab.history
}
