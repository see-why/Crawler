package main

import (
	"fmt"
	"os"
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
	baseURL := args[0]
	fmt.Printf("starting crawl of: %s\n", baseURL)

	// Initialize the pages map to track crawled pages
	pages := make(map[string]int)

	// Start crawling from the base URL
	crawlPage(baseURL, baseURL, pages)

	// Print the results
	fmt.Println("\n=== Crawl Results ===")
	for url, count := range pages {
		fmt.Printf("%s: %d\n", url, count)
	}
}
