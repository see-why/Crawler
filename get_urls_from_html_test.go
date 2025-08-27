package main

import (
	"reflect"
	"testing"
)

func TestGetURLsFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="/path/one">
			<span>Boot.dev</span>
		</a>
		<a href="https://other.com/path/one">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:     "multiple relative URLs",
			inputURL: "https://example.com",
			inputBody: `
<html>
	<body>
		<a href="/foo">foo</a>
		<a href="/bar">bar</a>
	</body>
</html>
`,
			expected: []string{"https://example.com/foo", "https://example.com/bar"},
		},
		{
			name:     "no anchor tags",
			inputURL: "https://none.com",
			inputBody: `
<html>
	<body>
		<p>No links here!</p>
	</body>
</html>
`,
			expected: []string{},
		},
		{
			name:     "anchor with no href",
			inputURL: "https://nohref.com",
			inputBody: `
<html>
	<body>
		<a>no href</a>
	</body>
</html>
`,
			expected: []string{},
		},
		{
			name:     "anchor with empty href",
			inputURL: "https://emptyhref.com",
			inputBody: `
<html>
	<body>
		<a href="">empty href</a>
	</body>
</html>
`,
			expected: []string{"https://emptyhref.com"},
		},
		{
			name:     "anchor with invalid href",
			inputURL: "https://malformed.com",
			inputBody: `
<html>
	<body>
		<a href="://bad:url">bad url</a>
	</body>
</html>
`,
			expected: []string{"https://malformed.com://bad:url"},
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if len(actual) == 0 && len(tc.expected) == 0 {
				// Both are empty, this is a pass
				return
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected URLs: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
