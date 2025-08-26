package main

import (
	"net/url"
	"strings"
)

// normalizeURL takes a URL string and returns its normalized form.
func normalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := strings.TrimPrefix(u.Hostname(), "www.")

	path := strings.TrimSuffix(u.EscapedPath(), "/")

	// Rebuild normalized URL: host + path
	normalized := host
	if path != "" {
		normalized += path
	}

	return normalized, nil
}
