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
func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	count, exists := cfg.pages[normalizedURL]
	if exists {
		cfg.pages[normalizedURL] = count + 1
		return false
	}

	cfg.pages[normalizedURL] = 1
	return true
}

// crawlPage recursively crawls pages starting from rawCurrentURL, staying within the same domain as baseURL
func (cfg *config) crawlPage(rawCurrentURL string) {
	// Acquire concurrency control
	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()

	// Check if we've reached the maximum number of pages
	cfg.mu.Lock()
	currentPageCount := len(cfg.pages)
	cfg.mu.Unlock()

	if currentPageCount >= cfg.maxPages {
		return
	}

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

	// Check if this is the first visit to this page
	isFirst := cfg.addPageVisit(normalizedURL)
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
