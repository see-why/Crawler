package main

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected string
	}{
		{
			name:     "remove scheme",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove http scheme",
			inputURL: "http://example.com",
			expected: "example.com",
		},
		{
			name:     "remove trailing slash",
			inputURL: "https://example.com/",
			expected: "example.com",
		},
		{
			name:     "keep path",
			inputURL: "https://example.com/foo/bar",
			expected: "example.com/foo/bar",
		},
		{
			name:     "remove www",
			inputURL: "https://www.example.com",
			expected: "example.com",
		},
		{
			name:     "remove query",
			inputURL: "https://example.com/path?query=1",
			expected: "example.com/path",
		},
		{
			name:     "remove fragment",
			inputURL: "https://example.com/path#section",
			expected: "example.com/path",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeURL(tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - %s FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if actual != tc.expected {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
