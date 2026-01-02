package scela

import (
	"testing"
)

func TestPatternMatcher_Match(t *testing.T) {
	pm := newPatternMatcher()

	tests := []struct {
		name    string
		pattern string
		topic   string
		want    bool
	}{
		// Exact matches
		{"exact match", "user.created", "user.created", true},
		{"exact mismatch", "user.created", "user.updated", false},

		// Single wildcard
		{"single wildcard prefix", "user.*", "user.created", true},
		{"single wildcard prefix 2", "user.*", "user.updated", true},
		{"single wildcard prefix mismatch", "user.*", "order.created", false},

		// Suffix wildcard
		{"suffix wildcard", "*.created", "user.created", true},
		{"suffix wildcard 2", "*.created", "order.created", true},
		{"suffix wildcard mismatch", "*.created", "user.updated", false},

		// All wildcard
		{"all wildcard *", "*", "user.created", true},
		{"all wildcard #", "#", "order.updated", true},
		{"all wildcard matches anything", "*", "any.topic.here", true},

		// Multiple segments
		{"multi segment exact", "user.profile.updated", "user.profile.updated", true},
		{"multi segment wildcard", "user.*.updated", "user.profile.updated", true},
		{"multi segment wildcard mismatch length", "user.*", "user.profile.updated", false},

		// Edge cases
		{"empty topic", "user.created", "", false},
		{"empty pattern", "", "user.created", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pm.Match(tt.pattern, tt.topic)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.topic, got, tt.want)
			}
		})
	}
}

func TestPatternMatcher_MatchMultiple(t *testing.T) {
	pm := newPatternMatcher()

	patterns := []string{"user.*", "*.created", "order.updated", "*"}
	topic := "user.created"

	matches := pm.MatchMultiple(patterns, topic)

	// Should match: user.*, *.created, *
	expectedCount := 3
	if len(matches) != expectedCount {
		t.Errorf("MatchMultiple() returned %d matches, want %d", len(matches), expectedCount)
	}
}

func BenchmarkPatternMatcher_Match(b *testing.B) {
	pm := newPatternMatcher()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pm.Match("user.*", "user.created")
	}
}
