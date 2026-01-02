package scela

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SQLStore provides database persistence for messages.
// It works with any database/sql compatible driver.
type SQLStore struct {
	db         *sql.DB
	tableName  string
	serializer Serializer
	mu         sync.Mutex
}

// SQLStoreConfig configures a SQL store.
type SQLStoreConfig struct {
	DB         *sql.DB
	TableName  string
	Serializer Serializer
}

// NewSQLStore creates a new SQL-based message store.
func NewSQLStore(config SQLStoreConfig) (*SQLStore, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if config.TableName == "" {
		config.TableName = "scela_messages"
	}

	if config.Serializer == nil {
		config.Serializer = NewJSONSerializer()
	}

	store := &SQLStore{
		db:         config.DB,
		tableName:  config.TableName,
		serializer: config.Serializer,
	}

	// Create table if it doesn't exist
	if err := store.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return store, nil
}

// createTable creates the messages table if it doesn't exist.
func (s *SQLStore) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			topic TEXT NOT NULL,
			payload TEXT NOT NULL,
			metadata TEXT,
			timestamp TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, s.tableName)

	_, err := s.db.Exec(query)
	return err
}

// Store implements MessageStore.
func (s *SQLStore) Store(ctx context.Context, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize payload
	payloadData, err := s.serializer.Serialize(msg.Payload())
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	// Serialize metadata
	metadataData, err := json.Marshal(msg.Metadata())
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, topic, payload, metadata, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`, s.tableName)

	_, err = s.db.ExecContext(ctx, query,
		msg.ID(),
		msg.Topic(),
		string(payloadData),
		string(metadataData),
		msg.Timestamp(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	return nil
}

// scanMessages is a helper function to scan and deserialize message rows.
func (s *SQLStore) scanMessages(rows *sql.Rows) ([]Message, error) {
	messages := make([]Message, 0)

	for rows.Next() {
		var (
			id          string
			topic       string
			payloadData string
			metadataStr string
			timestamp   time.Time
		)

		if err := rows.Scan(&id, &topic, &payloadData, &metadataStr, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var payload interface{}
		if err := s.serializer.Deserialize([]byte(payloadData), &payload); err != nil {
			return nil, fmt.Errorf("failed to deserialize payload: %w", err)
		}

		var metadata map[string]interface{}
		if metadataStr != "" {
			if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
				return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
			}
		}

		msg := &message{
			id:        id,
			topic:     topic,
			payload:   payload,
			metadata:  metadata,
			timestamp: timestamp,
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return messages, nil
}

// Load implements MessageStore.
func (s *SQLStore) Load(ctx context.Context) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := fmt.Sprintf(`
		SELECT id, topic, payload, metadata, timestamp
		FROM %s
		ORDER BY timestamp ASC
	`, s.tableName)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanMessages(rows)
}

// LoadByTopic loads messages for a specific topic.
func (s *SQLStore) LoadByTopic(ctx context.Context, topic string) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := fmt.Sprintf(`
		SELECT id, topic, payload, metadata, timestamp
		FROM %s
		WHERE topic = ?
		ORDER BY timestamp ASC
	`, s.tableName)

	rows, err := s.db.QueryContext(ctx, query, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanMessages(rows)
}

// LoadAfter loads messages after a specific timestamp.
func (s *SQLStore) LoadAfter(ctx context.Context, after time.Time) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := fmt.Sprintf(`
		SELECT id, topic, payload, metadata, timestamp
		FROM %s
		WHERE timestamp > ?
		ORDER BY timestamp ASC
	`, s.tableName)

	rows, err := s.db.QueryContext(ctx, query, after)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanMessages(rows)
}

// Clear implements MessageStore.
func (s *SQLStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := fmt.Sprintf("DELETE FROM %s", s.tableName)
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clear messages: %w", err)
	}

	return nil
}

// ClearBefore removes messages older than the specified time.
func (s *SQLStore) ClearBefore(ctx context.Context, before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := fmt.Sprintf("DELETE FROM %s WHERE timestamp < ?", s.tableName)
	_, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to clear old messages: %w", err)
	}

	return nil
}

// Count returns the number of stored messages.
func (s *SQLStore) Count(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", s.tableName)
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// Close implements MessageStore.
func (s *SQLStore) Close() error {
	// Note: We don't close the DB here as it might be shared
	// The caller is responsible for closing the database connection
	return nil
}
