package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
)

func main() {
	// Get command line arguments (excluding program name)
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 1 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	// Exactly one argument - the BASE_URL
	baseURLString := args[0]
	fmt.Printf("starting crawl of: %s\n", baseURLString)

	// Parse the base URL
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		fmt.Printf("Error parsing base URL: %v\n", err)
		os.Exit(1)
	}

	// Initialize the config struct
	maxConcurrency := 10 // Test with higher concurrency
	cfg := &config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
	}

	// Start crawling from the base URL
	cfg.wg.Add(1)
	go cfg.crawlPage(baseURLString)

	// Wait for all goroutines to complete
	cfg.wg.Wait()

	// Print the results
	fmt.Println("\n=== Crawl Results ===")
	for url, count := range cfg.pages {
		fmt.Printf("%s: %d\n", url, count)
	}
}
