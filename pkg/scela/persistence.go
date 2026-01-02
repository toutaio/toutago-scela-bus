package scela

import (
"context"
"encoding/json"
"fmt"
"io"
"os"
"sync"
"time"
)

// MessageStore defines the interface for message persistence.
type MessageStore interface {
// Store persists a message.
Store(ctx context.Context, msg Message) error

// Load retrieves messages from storage.
Load(ctx context.Context) ([]Message, error)

// Clear removes all stored messages.
Clear(ctx context.Context) error

// Close closes the store.
Close() error
}

// InMemoryStore is a simple in-memory message store.
type InMemoryStore struct {
messages []Message
mu       sync.RWMutex
maxSize  int
}

// NewInMemoryStore creates a new in-memory store.
func NewInMemoryStore(maxSize int) *InMemoryStore {
if maxSize <= 0 {
maxSize = 10000
}
return &InMemoryStore{
messages: make([]Message, 0),
maxSize:  maxSize,
}
}

// Store implements MessageStore.
func (s *InMemoryStore) Store(ctx context.Context, msg Message) error {
s.mu.Lock()
defer s.mu.Unlock()

s.messages = append(s.messages, msg)

// Trim if exceeded max size
if len(s.messages) > s.maxSize {
s.messages = s.messages[len(s.messages)-s.maxSize:]
}

return nil
}

// Load implements MessageStore.
func (s *InMemoryStore) Load(ctx context.Context) ([]Message, error) {
s.mu.RLock()
defer s.mu.RUnlock()

// Return a copy
result := make([]Message, len(s.messages))
copy(result, s.messages)
return result, nil
}

// Clear implements MessageStore.
func (s *InMemoryStore) Clear(ctx context.Context) error {
s.mu.Lock()
defer s.mu.Unlock()

s.messages = make([]Message, 0)
return nil
}

// Close implements MessageStore.
func (s *InMemoryStore) Close() error {
return nil
}

// FileStore persists messages to a file.
type FileStore struct {
filepath   string
serializer Serializer
mu         sync.Mutex
}

// NewFileStore creates a new file-based store.
func NewFileStore(filepath string) *FileStore {
return &FileStore{
filepath:   filepath,
serializer: NewJSONSerializer(),
}
}

// Store implements MessageStore.
func (s *FileStore) Store(ctx context.Context, msg Message) error {
s.mu.Lock()
defer s.mu.Unlock()

// Load existing messages
messages, err := s.loadFromFile()
if err != nil && !os.IsNotExist(err) {
return err
}

// Append new message
messages = append(messages, msg)

// Save back to file
return s.saveToFile(messages)
}

// Load implements MessageStore.
func (s *FileStore) Load(ctx context.Context) ([]Message, error) {
s.mu.Lock()
defer s.mu.Unlock()

return s.loadFromFile()
}

// Clear implements MessageStore.
func (s *FileStore) Clear(ctx context.Context) error {
s.mu.Lock()
defer s.mu.Unlock()

return os.Remove(s.filepath)
}

// Close implements MessageStore.
func (s *FileStore) Close() error {
return nil
}

// loadFromFile loads messages from the file.
func (s *FileStore) loadFromFile() ([]Message, error) {
file, err := os.Open(s.filepath)
if err != nil {
if os.IsNotExist(err) {
return []Message{}, nil
}
return nil, err
}
defer file.Close()

data, err := io.ReadAll(file)
if err != nil {
return nil, err
}

if len(data) == 0 {
return []Message{}, nil
}

var messagesData []map[string]interface{}
if err := json.Unmarshal(data, &messagesData); err != nil {
return nil, err
}

messages := make([]Message, 0, len(messagesData))
for _, msgData := range messagesData {
topic, ok := msgData["topic"].(string)
if !ok {
continue
}
payload := msgData["payload"]
msg := NewMessage(topic, payload)
messages = append(messages, msg)
}

return messages, nil
}

// saveToFile saves messages to the file.
func (s *FileStore) saveToFile(messages []Message) error {
messagesData := make([]map[string]interface{}, 0, len(messages))

for _, msg := range messages {
msgData := map[string]interface{}{
"id":        msg.ID(),
"topic":     msg.Topic(),
"payload":   msg.Payload(),
"timestamp": msg.Timestamp(),
}
messagesData = append(messagesData, msgData)
}

data, err := json.MarshalIndent(messagesData, "", "  ")
if err != nil {
return err
}

return os.WriteFile(s.filepath, data, 0644)
}

// PersistentBus wraps a bus with message persistence.
type PersistentBus struct {
Bus
store MessageStore
mu    sync.Mutex
}

// NewPersistentBus creates a new persistent bus.
func NewPersistentBus(bus Bus, store MessageStore) *PersistentBus {
return &PersistentBus{
Bus:   bus,
store: store,
}
}

// Publish publishes and persists a message.
func (pb *PersistentBus) Publish(ctx context.Context, topic string, payload interface{}) error {
msg := NewMessage(topic, payload)

// Persist first
if err := pb.store.Store(ctx, msg); err != nil {
return fmt.Errorf("failed to persist message: %w", err)
}

// Then publish
return pb.Bus.Publish(ctx, topic, payload)
}

// Replay replays all stored messages.
func (pb *PersistentBus) Replay(ctx context.Context) error {
messages, err := pb.store.Load(ctx)
if err != nil {
return err
}

for _, msg := range messages {
if err := pb.Bus.Publish(ctx, msg.Topic(), msg.Payload()); err != nil {
return err
}
}

return nil
}

// GetStore returns the underlying message store.
func (pb *PersistentBus) GetStore() MessageStore {
return pb.store
}

// Close closes the persistent bus and its store.
func (pb *PersistentBus) Close() error {
if err := pb.store.Close(); err != nil {
return err
}
return pb.Bus.Close()
}

// ReplayableStore wraps a store with replay capability.
type ReplayableStore struct {
store     MessageStore
startTime time.Time
}

// NewReplayableStore creates a store that supports replay from a specific time.
func NewReplayableStore(store MessageStore, startTime time.Time) *ReplayableStore {
return &ReplayableStore{
store:     store,
startTime: startTime,
}
}

// Store implements MessageStore.
func (rs *ReplayableStore) Store(ctx context.Context, msg Message) error {
return rs.store.Store(ctx, msg)
}

// Load implements MessageStore, filtering by start time.
func (rs *ReplayableStore) Load(ctx context.Context) ([]Message, error) {
all, err := rs.store.Load(ctx)
if err != nil {
return nil, err
}

filtered := make([]Message, 0)
for _, msg := range all {
if msg.Timestamp().After(rs.startTime) || msg.Timestamp().Equal(rs.startTime) {
filtered = append(filtered, msg)
}
}

return filtered, nil
}

// Clear implements MessageStore.
func (rs *ReplayableStore) Clear(ctx context.Context) error {
return rs.store.Clear(ctx)
}

// Close implements MessageStore.
func (rs *ReplayableStore) Close() error {
return rs.store.Close()
}
