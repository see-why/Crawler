package main

import (
	"fmt"
	"net/url"
	"strings"
)

// crawlPage recursively crawls pages starting from rawCurrentURL, staying within the same domain as rawBaseURL
func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
	// Parse the base URL and current URL to compare domains
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		fmt.Printf("Error parsing base URL %s: %v\n", rawBaseURL, err)
		return
	}

	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error parsing current URL %s: %v\n", rawCurrentURL, err)
		return
	}

	// Check if current URL is on the same domain as base URL
	if !strings.EqualFold(currentURL.Hostname(), baseURL.Hostname()) {
		return
	}

	// Get normalized version of the current URL
	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error normalizing URL %s: %v\n", rawCurrentURL, err)
		return
	}

	// Check if we've already crawled this page
	if count, exists := pages[normalizedURL]; exists {
		pages[normalizedURL] = count + 1
		return
	}

	// Add this page to our map
	pages[normalizedURL] = 1

	// Print what we're crawling
	fmt.Printf("Crawling: %s\n", rawCurrentURL)

	// Get the HTML from the current URL
	htmlBody, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error getting HTML from %s: %v\n", rawCurrentURL, err)
		return
	}

	// Get all URLs from the HTML
	urls, err := getURLsFromHTML(htmlBody, rawBaseURL)
	if err != nil {
		fmt.Printf("Error getting URLs from HTML of %s: %v\n", rawCurrentURL, err)
		return
	}

	// Recursively crawl each URL found on the page
	for _, foundURL := range urls {
		crawlPage(rawBaseURL, foundURL, pages)
	}
}
