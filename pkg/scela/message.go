package scela

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// message is the default implementation of Message interface.
type message struct {
	id        string
	topic     string
	payload   interface{}
	metadata  map[string]interface{}
	timestamp time.Time
	priority  Priority
}

// generateID generates a random message ID.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// NewMessage creates a new message.
func NewMessage(topic string, payload interface{}) Message {
	return &message{
		id:        generateID(),
		topic:     topic,
		payload:   payload,
		metadata:  make(map[string]interface{}),
		timestamp: time.Now(),
		priority:  PriorityNormal,
	}
}

// NewMessageWithPriority creates a new message with specified priority.
func NewMessageWithPriority(topic string, payload interface{}, priority Priority) Message {
	msg := NewMessage(topic, payload).(*message)
	msg.priority = priority
	return msg
}

// ID returns the message ID.
func (m *message) ID() string {
	return m.id
}

// Topic returns the message topic.
func (m *message) Topic() string {
	return m.topic
}

// Payload returns the message payload.
func (m *message) Payload() interface{} {
	return m.payload
}

// Metadata returns the message metadata.
func (m *message) Metadata() map[string]interface{} {
	return m.metadata
}

// Timestamp returns when the message was created.
func (m *message) Timestamp() time.Time {
	return m.timestamp
}

// Priority returns the message priority (not part of Message interface, internal use).
func (m *message) Priority() Priority {
	return m.priority
}
