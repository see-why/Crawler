package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
)

func main() {
	// Get command line arguments (excluding program name)
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("Usage: crawler <URL> [max_concurrency]")
		fmt.Println("  URL: The website URL to crawl")
		fmt.Println("  max_concurrency: Maximum number of concurrent goroutines (default: 10)")
		fmt.Println("  Environment variable CRAWLER_MAX_CONCURRENCY can also be used")
		os.Exit(1)
	}

	if len(args) > 2 {
		fmt.Println("too many arguments provided")
		fmt.Println("Usage: crawler <URL> [max_concurrency]")
		os.Exit(1)
	}

	// First argument - the BASE_URL
	baseURLString := args[0]

	// Second argument or environment variable - maxConcurrency
	maxConcurrency := 10 // Default value

	// Check if maxConcurrency was provided as command line argument
	if len(args) == 2 {
		if parsed, err := strconv.Atoi(args[1]); err != nil {
			fmt.Printf("Error parsing max_concurrency '%s': %v\n", args[1], err)
			fmt.Println("max_concurrency must be a positive integer")
			os.Exit(1)
		} else if parsed <= 0 {
			fmt.Println("max_concurrency must be a positive integer")
			os.Exit(1)
		} else {
			maxConcurrency = parsed
		}
	} else if envVar := os.Getenv("CRAWLER_MAX_CONCURRENCY"); envVar != "" {
		// Check environment variable if no command line argument provided
		if parsed, err := strconv.Atoi(envVar); err != nil {
			fmt.Printf("Error parsing CRAWLER_MAX_CONCURRENCY '%s': %v\n", envVar, err)
			fmt.Println("CRAWLER_MAX_CONCURRENCY must be a positive integer")
			os.Exit(1)
		} else if parsed <= 0 {
			fmt.Println("CRAWLER_MAX_CONCURRENCY must be a positive integer")
			os.Exit(1)
		} else {
			maxConcurrency = parsed
		}
	}

	fmt.Printf("starting crawl of: %s (max concurrency: %d)\n", baseURLString, maxConcurrency)

	// Parse the base URL
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		fmt.Printf("Error parsing base URL: %v\n", err)
		os.Exit(1)
	}

	// Initialize the config struct
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
