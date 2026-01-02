package scela

import (
"testing"
)

func TestJSONSerializer(t *testing.T) {
serializer := NewJSONSerializer()

// Test simple types
payload := map[string]interface{}{
"name":  "test",
"count": 42,
}

data, err := serializer.Serialize(payload)
if err != nil {
t.Fatalf("Serialize() error = %v", err)
}

var result map[string]interface{}
err = serializer.Deserialize(data, &result)
if err != nil {
t.Fatalf("Deserialize() error = %v", err)
}

if result["name"] != "test" {
t.Errorf("Expected name 'test', got %v", result["name"])
}
}

func TestSerializableMessage(t *testing.T) {
msg := NewMessage("test.topic", map[string]string{"key": "value"})
sm := NewSerializableMessage(msg, nil) // Uses default JSON serializer

data, err := sm.Serialize()
if err != nil {
t.Fatalf("Serialize() error = %v", err)
}

if len(data) == 0 {
t.Error("Expected non-empty serialized data")
}
}

func TestSerializableMessage_Full(t *testing.T) {
msg := NewMessage("test.topic", "test payload")
sm := NewSerializableMessage(msg, NewJSONSerializer())

data, err := sm.SerializeMessage()
if err != nil {
t.Fatalf("SerializeMessage() error = %v", err)
}

deserializedMsg, err := DeserializeMessage(data, NewJSONSerializer())
if err != nil {
t.Fatalf("DeserializeMessage() error = %v", err)
}

if deserializedMsg.Topic() != "test.topic" {
t.Errorf("Expected topic 'test.topic', got %s", deserializedMsg.Topic())
}
}
