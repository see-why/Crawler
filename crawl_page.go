package main

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// Maximum number of retry attempts for failed requests
	maxRetries = 3
	// Base delay for exponential backoff
	baseRetryDelay = 1 * time.Second
	// Maximum number of errors to track per host
	maxErrorsPerHost = 10
	// Maximum delay for exponential backoff (cap at 30 seconds)
	maxRetryBackoffDelay = 30 * time.Second
)

type config struct {
	pages              map[string]int
	externalLinks      map[string]int
	baseURL            *url.URL
	maxPages           int
	batchSize          int
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	ctx                context.Context
	// Error tracking for circuit breaker pattern
	hostErrors   map[string]*int64
	hostErrorsMu *sync.RWMutex
	// Statistics
	totalRequests  *int64
	failedRequests *int64
}

// addPageVisit safely adds a page visit to the map and returns whether this is the first visit
// and whether adding this page would exceed the maxPages limit
func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool, exceedsLimit bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	count, exists := cfg.pages[normalizedURL]
	if exists {
		cfg.pages[normalizedURL] = count + 1
		return false, false
	}

	// Check if adding this page would exceed the limit
	if len(cfg.pages) >= cfg.maxPages {
		return false, true
	}

	// This is a new page, add it
	cfg.pages[normalizedURL] = 1
	return true, false
}

// incrementHostError tracks errors per host for circuit breaker pattern
func (cfg *config) incrementHostError(host string) {
	cfg.hostErrorsMu.Lock()
	defer cfg.hostErrorsMu.Unlock()

	if cfg.hostErrors[host] == nil {
		var zero int64 = 0
		cfg.hostErrors[host] = &zero
	}
	atomic.AddInt64(cfg.hostErrors[host], 1)
}

// shouldSkipHost determines if we should skip a host due to too many errors
func (cfg *config) shouldSkipHost(host string) bool {
	cfg.hostErrorsMu.RLock()
	defer cfg.hostErrorsMu.RUnlock()

	if errorCount := cfg.hostErrors[host]; errorCount != nil {
		return atomic.LoadInt64(errorCount) >= maxErrorsPerHost
	}
	return false
}

// Use shared CalculateBackoffDelay from backoff.go

// incrementStats updates request statistics
func (cfg *config) incrementStats(failed bool) {
	atomic.AddInt64(cfg.totalRequests, 1)
	if failed {
		atomic.AddInt64(cfg.failedRequests, 1)
	}
}

// retryWithBackoff implements exponential backoff retry logic
func (cfg *config) retryWithBackoff(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Safe exponential backoff calculation with overflow protection
			delay := CalculateBackoffDelay(attempt, baseRetryDelay, maxRetryBackoffDelay)

			select {
			case <-cfg.ctx.Done():
				return cfg.ctx.Err()
			case <-time.After(delay):
			}
		}

		if err := operation(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("operation failed after %d retries, last error: %w", maxRetries, lastErr)
}

// crawlPage recursively crawls pages starting from rawCurrentURL, staying within the same domain as baseURL
func (cfg *config) crawlPage(rawCurrentURL string) {
	// Check if context is cancelled
	select {
	case <-cfg.ctx.Done():
		cfg.wg.Done() // Decrement WaitGroup since we're not doing any work
		return
	default:
	}

	// Acquire concurrency control
	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done() // Decrement WaitGroup after releasing concurrency control
	}()

	// Parse the current URL
	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		cfg.incrementStats(true)
		fmt.Printf("Error parsing current URL %s: %v\n", rawCurrentURL, err)
		return
	}

	// Check circuit breaker - skip hosts with too many errors
	if cfg.shouldSkipHost(currentURL.Hostname()) {
		cfg.incrementStats(true)
		fmt.Printf("Skipping %s due to too many previous errors\n", currentURL.Hostname())
		return
	}

	// Check if current URL is on the same domain as base URL
	if currentURL.Hostname() != cfg.baseURL.Hostname() {
		// Track external link
		cfg.mu.Lock()
		cfg.externalLinks[rawCurrentURL]++
		cfg.mu.Unlock()
		return
	}

	// Get normalized version of the current URL
	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		cfg.incrementStats(true)
		cfg.incrementHostError(currentURL.Hostname())
		fmt.Printf("Error normalizing URL %s: %v\n", rawCurrentURL, err)
		return
	}

	// Atomically check if this is the first visit and if we've reached the page limit
	isFirst, exceedsLimit := cfg.addPageVisit(normalizedURL)
	if exceedsLimit {
		return
	}
	if !isFirst {
		return
	}

	// Print what we're crawling
	fmt.Printf("Crawling: %s\n", rawCurrentURL)

	// Create a context with timeout for this specific request
	requestCtx, cancel := context.WithTimeout(cfg.ctx, 30*time.Second)
	defer cancel()

	// Use retry mechanism for getting HTML
	var htmlBody string
	err = cfg.retryWithBackoff(func() error {
		var htmlErr error
		htmlBody, htmlErr = getHTMLWithContext(requestCtx, rawCurrentURL)
		return htmlErr
	})

	if err != nil {
		cfg.incrementStats(true)
		cfg.incrementHostError(currentURL.Hostname())
		fmt.Printf("Error getting HTML from %s after retries: %v\n", rawCurrentURL, err)
		return
	}

	cfg.incrementStats(false) // Successful request

	// Get all URLs from the HTML with error handling
	urls, err := getURLsFromHTML(htmlBody, cfg.baseURL.String())
	if err != nil {
		fmt.Printf("Error getting URLs from HTML of %s: %v\n", rawCurrentURL, err)
		return
	}

	// Limit the number of URLs to process to avoid memory explosion
	if len(urls) > maxURLsPerPage {
		urls = urls[:maxURLsPerPage]
		fmt.Printf("Limiting URLs from %s to %d (originally %d)\n", rawCurrentURL, maxURLsPerPage, len(urls))
	}

	// Process URLs in batches to avoid creating too many goroutines at once
	batchSize := cfg.batchSize
	for i := 0; i < len(urls); i += batchSize {
		end := i + batchSize
		if end > len(urls) {
			end = len(urls)
		}

		// Process this batch of URLs
		for j := i; j < end; j++ {
			foundURL := urls[j]

			// Add to WaitGroup first to avoid race condition
			cfg.wg.Add(1)

			// Check context before starting new goroutine
			select {
			case <-cfg.ctx.Done():
				// Context cancelled, decrement WaitGroup and return
				cfg.wg.Done()
				return
			default:
				go cfg.crawlPage(foundURL)
			}
		}
	}
}
