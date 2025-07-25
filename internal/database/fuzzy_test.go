package database

import (
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		text     string
		pattern  string
		expected bool // whether it should match (score > 0)
		name     string
	}{
		// Basic matches
		{"project", "proj", true, "basic prefix match"},
		{"project", "prj", true, "consonant match"},
		{"project", "pj", true, "sparse match"},
		{"project", "xyz", false, "no match"},

		// Path matches (using basename)
		{"/home/user/my-project", "proj", true, "path basename match"},
		{"/home/user/documents", "doc", true, "path prefix match"},
		{"/very/long/path/to/project", "prj", true, "deep path match"},

		// Case sensitivity
		{"Project", "proj", true, "case insensitive"},
		{"PROJECT", "proj", true, "case insensitive upper"},
		{"project", "PROJ", true, "case insensitive pattern"},

		// Word boundaries
		{"my-awesome-project", "map", true, "word boundary match"},
		{"user_local_bin", "ulb", true, "underscore boundaries"},
		{"some.config.file", "scf", true, "dot boundaries"},

		// Edge cases
		{"", "test", false, "empty text"},
		{"test", "", false, "empty pattern"},
		{"a", "a", true, "single character match"},
		{"test", "test", true, "exact match"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := fuzzyMatch(tt.text, tt.pattern)
			hasMatch := score > 0

			if hasMatch != tt.expected {
				t.Errorf("fuzzyMatch(%q, %q) = %d (match: %v), expected match: %v",
					tt.text, tt.pattern, score, hasMatch, tt.expected)
			}
		})
	}
}

func TestFuzzyMatchScoring(t *testing.T) {
	// Test that better matches get higher scores
	tests := []struct {
		text1, text2 string
		pattern      string
		name         string
	}{
		// Prefix matches should score higher than non-prefix
		{"project", "myproject", "proj", "prefix vs non-prefix"},

		// Exact case matches should score higher
		{"Project", "project", "Proj", "case match bonus"},

		// Consecutive matches should score higher
		{"project", "p-r-o-j-e-c-t", "proj", "consecutive vs sparse"},

		// Word boundary matches should score higher
		{"my-project", "myproject", "mp", "word boundary bonus"},

		// Shorter matches should score higher (more specific)
		{"proj", "project", "proj", "shorter vs longer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score1 := fuzzyMatch(tt.text1, tt.pattern)
			score2 := fuzzyMatch(tt.text2, tt.pattern)

			if score1 <= score2 {
				t.Errorf("Expected %q to score higher than %q for pattern %q, got %d vs %d",
					tt.text1, tt.text2, tt.pattern, score1, score2)
			}
		})
	}
}

func TestCanMatch(t *testing.T) {
	tests := []struct {
		text, pattern string
		expected      bool
		name          string
	}{
		{"hello", "hlo", true, "can match with gaps"},
		{"hello", "hel", true, "can match prefix"},
		{"hello", "llo", true, "can match suffix"},
		{"hello", "xyz", false, "cannot match missing chars"},
		{"hello", "hlx", false, "cannot match with wrong char"},
		{"abc", "acb", false, "cannot match wrong order"},
		{"project", "pjt", true, "can match sparse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canMatch(tt.text, tt.pattern)
			if result != tt.expected {
				t.Errorf("canMatch(%q, %q) = %v, expected %v",
					tt.text, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsWordBoundary(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'/', true},
		{'-', true},
		{'_', true},
		{' ', true},
		{'.', true},
		{'a', false},
		{'A', false},
		{'1', false},
		{'@', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := isWordBoundary(tt.char)
			if result != tt.expected {
				t.Errorf("isWordBoundary(%c) = %v, expected %v",
					tt.char, result, tt.expected)
			}
		})
	}
}

// Benchmark the fuzzy matching performance
func BenchmarkFuzzyMatch(b *testing.B) {
	text := "/home/user/development/my-awesome-project"
	pattern := "map"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fuzzyMatch(text, pattern)
	}
}

func BenchmarkFuzzyMatchLong(b *testing.B) {
	text := "/very/long/path/to/some/deeply/nested/directory/with/many/components"
	pattern := "nested"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fuzzyMatch(text, pattern)
	}
}
