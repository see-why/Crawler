package main

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"
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

	// Check if we've reached the maximum number of pages before adding a NEW page
	if len(cfg.pages) >= cfg.maxPages {
		return false, true
	}

	cfg.pages[normalizedURL] = 1
	return true, false
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
		fmt.Printf("Error parsing current URL %s: %v\n", rawCurrentURL, err)
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
	requestCtx, cancel := context.WithTimeout(cfg.ctx, 15*time.Second)
	defer cancel()

	// Get the HTML from the current URL with context
	htmlBody, err := getHTMLWithContext(requestCtx, rawCurrentURL)
	if err != nil {
		fmt.Printf("Error getting HTML from %s: %v\n", rawCurrentURL, err)
		return
	}

	// Get all URLs from the HTML
	urls, err := getURLsFromHTML(htmlBody, cfg.baseURL.String())
	if err != nil {
		fmt.Printf("Error getting URLs from HTML of %s: %v\n", rawCurrentURL, err)
		return
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
