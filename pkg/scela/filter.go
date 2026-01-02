package scela

import "context"

// Filter is a function that determines whether a message should be processed.
type Filter func(msg Message) bool

// FilterMiddleware creates middleware that filters messages based on a predicate.
func FilterMiddleware(filter Filter) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Message) error {
			if !filter(msg) {
				// Skip this message
				return nil
			}
			return next.Handle(ctx, msg)
		})
	}
}

// TopicFilter returns a filter that matches specific topics.
func TopicFilter(topics ...string) Filter {
	topicMap := make(map[string]bool)
	for _, topic := range topics {
		topicMap[topic] = true
	}

	return func(msg Message) bool {
		return topicMap[msg.Topic()]
	}
}

// PayloadFilter returns a filter based on payload type.
func PayloadFilter(typeCheck func(interface{}) bool) Filter {
	return func(msg Message) bool {
		return typeCheck(msg.Payload())
	}
}

// MetadataFilter returns a filter based on message metadata.
func MetadataFilter(key string, value interface{}) Filter {
	return func(msg Message) bool {
		v, exists := msg.Metadata()[key]
		if !exists {
			return false
		}
		return v == value
	}
}

// AndFilter combines multiple filters with AND logic.
func AndFilter(filters ...Filter) Filter {
	return func(msg Message) bool {
		for _, f := range filters {
			if !f(msg) {
				return false
			}
		}
		return true
	}
}

// OrFilter combines multiple filters with OR logic.
func OrFilter(filters ...Filter) Filter {
	return func(msg Message) bool {
		for _, f := range filters {
			if f(msg) {
				return true
			}
		}
		return false
	}
}

// NotFilter inverts a filter.
func NotFilter(filter Filter) Filter {
	return func(msg Message) bool {
		return !filter(msg)
	}
}
