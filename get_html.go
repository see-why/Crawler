package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// getHTML fetches the HTML content from the given URL
func getHTML(rawURL string) (string, error) {
	// Make HTTP GET request
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for HTTP error status codes (400+)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Check content-type header
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		return "", fmt.Errorf("content-type is not text/html, got: %s", contentType)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
