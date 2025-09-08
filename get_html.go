package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// Network-level errors that are typically retryable
var retryableErrors = []string{
	"timeout",
	"connection refused",
	"connection reset",
	"no such host",
	"network unreachable",
	"temporary failure",
	"i/o timeout",
}

// HTTP status codes that are retryable
var retryableHTTPCodes = []string{
	"HTTP error 429", // Too Many Requests
	"HTTP error 502", // Bad Gateway
	"HTTP error 503", // Service Unavailable
	"HTTP error 504", // Gateway Timeout
	"HTTP error 520", // Cloudflare unknown error
	"HTTP error 521", // Cloudflare web server down
	"HTTP error 522", // Cloudflare connection timeout
	"HTTP error 523", // Cloudflare origin unreachable
	"HTTP error 524", // Cloudflare timeout
}

const (
	// Maximum response body size (10MB)
	maxResponseSize = 10 * 1024 * 1024
	// Request timeout for individual requests
	defaultRequestTimeout = 15 * time.Second
	// Rate limiting delay between requests
	requestDelay = 100 * time.Millisecond
	// Maximum number of retries for failed requests
	maxHTTPRetries = 3
	// Base delay for HTTP retry backoff
	httpRetryDelay = 500 * time.Millisecond
	// Maximum delay for exponential backoff (cap at 30 seconds)
	maxBackoffDelay = 30 * time.Second
)

// Global HTTP client with optimized settings for concurrent requests
var httpClient = &http.Client{
	Timeout: defaultRequestTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		MaxConnsPerHost:     20, // Limit connections per host
	},
}

// calculateBackoffDelay calculates exponential backoff delay with overflow protection
func calculateBackoffDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Prevent negative or zero attempts
	if attempt <= 0 {
		return 0
	}

	// Use math.Pow for safe calculation, but with reasonable limits
	// Limit attempt to prevent extreme values
	if attempt > 10 {
		attempt = 10 // Cap at 10 to prevent overflow
	}

	// Calculate: baseDelay * 2^(attempt-1)
	// Using float64 to avoid integer overflow, then converting back
	multiplier := math.Pow(2, float64(attempt-1))

	// Convert to duration, checking for overflow
	if multiplier > float64(maxDelay/baseDelay) {
		return maxDelay
	}

	delay := time.Duration(float64(baseDelay) * multiplier)

	// Ensure we don't exceed the maximum
	if delay > maxDelay {
		return maxDelay
	}

	return delay
}

// getHTMLWithContext fetches HTML with context support for cancellation and robust error handling
func getHTMLWithContext(ctx context.Context, rawURL string) (string, error) {
	var lastErr error

	// Retry logic with exponential backoff
	for attempt := 0; attempt <= maxHTTPRetries; attempt++ {
		if attempt > 0 {
			// Safe exponential backoff calculation with overflow protection
			delay := calculateBackoffDelay(attempt, httpRetryDelay, maxBackoffDelay)

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
		}

		// Add rate limiting delay (only on first attempt to avoid double delay)
		if attempt == 0 {
			time.Sleep(requestDelay)
		}

		body, err := performHTTPRequest(ctx, rawURL)
		if err != nil {
			lastErr = err
			// Check if this is a retryable error
			if !isRetryableError(err) {
				return "", fmt.Errorf("non-retryable error: %w", err)
			}
			continue
		}

		return body, nil
	}

	return "", fmt.Errorf("HTTP request failed after %d retries for URL %s: %w", maxHTTPRetries, rawURL, lastErr)
}

// performHTTPRequest performs a single HTTP request
func performHTTPRequest(ctx context.Context, rawURL string) (string, error) {
	// Create a new HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add comprehensive headers to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Crawler/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Make HTTP request using the global client
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body for %s: %v\n", rawURL, closeErr)
		}
	}()

	// Check for HTTP error status codes
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error %d (%s) for URL %s", resp.StatusCode, resp.Status, rawURL)
	}

	// Check content-type header
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "text/html") {
		return "", fmt.Errorf("content-type is not HTML (got: %s) for URL %s", contentType, rawURL)
	}

	// Check content-length if provided to avoid reading massive files
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if resp.ContentLength > maxResponseSize {
			return "", fmt.Errorf("content too large (%d bytes, max %d) for URL %s", resp.ContentLength, maxResponseSize, rawURL)
		}
	}

	// Create a limited reader to prevent reading massive responses
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)

	// Read the response body with size limit
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if we hit the size limit
	if len(body) >= maxResponseSize {
		return "", fmt.Errorf("response body too large (>= %d bytes) for URL %s", maxResponseSize, rawURL)
	}

	return string(body), nil
}

// isRetryableError determines if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryable) {
			return true
		}
	}

	for _, retryableCode := range retryableHTTPCodes {
		if strings.Contains(errStr, retryableCode) {
			return true
		}
	}

	return false
}
