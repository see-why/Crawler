package main

import (
	"fmt"
	"net/url"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	maxPages           int
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

// addPageVisit safely adds a page visit to the map and returns whether this is the first visit
// and whether adding this page would exceed the maxPages limit
func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool, exceedsLimit bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// Check if we've reached the maximum number of pages before adding
	if len(cfg.pages) >= cfg.maxPages {
		return false, true
	}

	count, exists := cfg.pages[normalizedURL]
	if exists {
		cfg.pages[normalizedURL] = count + 1
		return false, false
	}

	cfg.pages[normalizedURL] = 1
	return true, false
}

// crawlPage recursively crawls pages starting from rawCurrentURL, staying within the same domain as baseURL
func (cfg *config) crawlPage(rawCurrentURL string) {
	// Acquire concurrency control
	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()

	// Parse the current URL
	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error parsing current URL %s: %v\n", rawCurrentURL, err)
		return
	}

	// Check if current URL is on the same domain as base URL
	if currentURL.Hostname() != cfg.baseURL.Hostname() {
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

	// Get the HTML from the current URL
	htmlBody, err := getHTML(rawCurrentURL)
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

	// Recursively crawl each URL found on the page using goroutines
	for _, foundURL := range urls {
		cfg.wg.Add(1)
		go cfg.crawlPage(foundURL)
	}
}
