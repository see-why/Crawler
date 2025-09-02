package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Global HTTP client with optimized settings for concurrent requests
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

// getHTML fetches the HTML content from the given URL with timeout
func getHTML(rawURL string) (string, error) {
	return getHTMLWithContext(context.Background(), rawURL)
}

// getHTMLWithContext fetches HTML with context support for cancellation
func getHTMLWithContext(ctx context.Context, rawURL string) (string, error) {
	// Create a new HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", err
	}

	// Add User-Agent header to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Crawler/1.0)")

	// Make HTTP request using the global client
	resp, err := httpClient.Do(req)
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
