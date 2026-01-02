package scela

import (
	"encoding/json"
	"fmt"
)

// Serializer defines the interface for message serialization.
type Serializer interface {
	// Serialize converts a message payload to bytes.
	Serialize(payload interface{}) ([]byte, error)

	// Deserialize converts bytes back to a payload.
	Deserialize(data []byte, target interface{}) error
}

// JSONSerializer is a JSON-based serializer.
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// Serialize implements the Serializer interface.
func (s *JSONSerializer) Serialize(payload interface{}) ([]byte, error) {
	return json.Marshal(payload)
}

// Deserialize implements the Serializer interface.
func (s *JSONSerializer) Deserialize(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// GOBSerializer would use encoding/gob (not implemented - example)
// ProtobufSerializer would use protobuf (not implemented - example)

// SerializableMessage wraps a message with serialization capability.
type SerializableMessage struct {
	msg        Message
	serializer Serializer
}

// NewSerializableMessage creates a new serializable message.
func NewSerializableMessage(msg Message, serializer Serializer) *SerializableMessage {
	if serializer == nil {
		serializer = NewJSONSerializer()
	}
	return &SerializableMessage{
		msg:        msg,
		serializer: serializer,
	}
}

// Message returns the underlying message.
func (sm *SerializableMessage) Message() Message {
	return sm.msg
}

// Serialize serializes the message payload.
func (sm *SerializableMessage) Serialize() ([]byte, error) {
	return sm.serializer.Serialize(sm.msg.Payload())
}

// SerializeMessage serializes an entire message including metadata.
func (sm *SerializableMessage) SerializeMessage() ([]byte, error) {
	data := map[string]interface{}{
		"id":        sm.msg.ID(),
		"topic":     sm.msg.Topic(),
		"payload":   sm.msg.Payload(),
		"metadata":  sm.msg.Metadata(),
		"timestamp": sm.msg.Timestamp(),
	}
	return sm.serializer.Serialize(data)
}

// DeserializeMessage deserializes a complete message.
func DeserializeMessage(data []byte, serializer Serializer) (Message, error) {
	if serializer == nil {
		serializer = NewJSONSerializer()
	}

	var msgData map[string]interface{}
	if err := serializer.Deserialize(data, &msgData); err != nil {
		return nil, err
	}

	topic, ok := msgData["topic"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid message format: missing topic")
	}

	payload := msgData["payload"]

	msg := NewMessage(topic, payload)
	return msg, nil
}
