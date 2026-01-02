package scela

import (
	"strings"
)

// patternMatcher handles wildcard pattern matching for topics.
type patternMatcher struct{}

// newPatternMatcher creates a new pattern matcher.
func newPatternMatcher() *patternMatcher {
	return &patternMatcher{}
}

// Match returns true if the topic matches the pattern.
// Patterns support:
//   - exact match: "user.created"
//   - single wildcard: "user.*" matches "user.created", "user.updated"
//   - suffix wildcard: "*.created" matches "user.created", "order.created"
//   - all wildcard: "*" or "#" matches everything
func (pm *patternMatcher) Match(pattern, topic string) bool {
	// All wildcard
	if pattern == "*" || pattern == "#" {
		return true
	}

	// Exact match
	if pattern == topic {
		return true
	}

	// No wildcards
	if !strings.Contains(pattern, "*") {
		return false
	}

	// Split pattern and topic by dots
	patternParts := strings.Split(pattern, ".")
	topicParts := strings.Split(topic, ".")

	// Different segment counts can still match with wildcards
	if len(patternParts) != len(topicParts) {
		return false
	}

	// Match each segment
	for i := range patternParts {
		if patternParts[i] == "*" {
			continue // Wildcard matches anything
		}
		if patternParts[i] != topicParts[i] {
			return false
		}
	}

	return true
}

// MatchMultiple returns all patterns that match the topic.
func (pm *patternMatcher) MatchMultiple(patterns []string, topic string) []string {
	var matches []string
	for _, pattern := range patterns {
		if pm.Match(pattern, topic) {
			matches = append(matches, pattern)
		}
	}
	return matches
}
