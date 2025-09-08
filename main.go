package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Page represents a page with its URL and count for sorting
type Page struct {
	URL   string
	Count int
}

// printReport sorts and prints the crawl results in a formatted report
func printReport(pages map[string]int, externalLinks map[string]int, baseURL string) error {
	fmt.Println()
	fmt.Println("=============================")
	fmt.Printf("  REPORT for %s\n", baseURL)
	fmt.Println("=============================")

	// Parse the baseURL to get the original scheme
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("error parsing base URL: %v", err)
	}

	// Convert map to slice of structs for sorting
	var pageList []Page
	for normalizedURL, count := range pages {
		// Reconstruct full URL from normalized URL using the parsed base URL
		// Split normalized URL to get host and path
		parts := strings.SplitN(normalizedURL, "/", 2)
		host := parts[0]
		path := ""
		if len(parts) > 1 {
			path = "/" + parts[1]
		}

		// Create full URL using the original scheme and port from base URL
		fullURL := &url.URL{
			Scheme: parsedBaseURL.Scheme,
			Host:   host,
			Path:   path,
		}
		pageList = append(pageList, Page{URL: fullURL.String(), Count: count})
	}

	// Sort by count (descending), then by URL (ascending) for ties
	sort.Slice(pageList, func(i, j int) bool {
		if pageList[i].Count != pageList[j].Count {
			return pageList[i].Count > pageList[j].Count // Higher counts first
		}
		return pageList[i].URL < pageList[j].URL // Alphabetical for ties
	})

	// Print each internal page
	for _, page := range pageList {
		fmt.Printf("Found %d internal links to %s\n", page.Count, page.URL)
	}

	// Print external links summary
	fmt.Println()
	fmt.Println("-----------------------------")
	fmt.Println("  EXTERNAL LINKS REPORT")
	fmt.Println("-----------------------------")
	// Convert externalLinks map to slice for sorting
	var externalList []Page
	for url, count := range externalLinks {
		externalList = append(externalList, Page{URL: url, Count: count})
	}
	sort.Slice(externalList, func(i, j int) bool {
		if externalList[i].Count != externalList[j].Count {
			return externalList[i].Count > externalList[j].Count
		}
		return externalList[i].URL < externalList[j].URL
	})
	for _, ext := range externalList {
		fmt.Printf("Found %d external links to %s\n", ext.Count, ext.URL)
	}

	return nil
}

// printCrawlStatistics prints crawling statistics and performance metrics
func printCrawlStatistics(cfg *config) {
	totalReqs := atomic.LoadInt64(cfg.totalRequests)
	failedReqs := atomic.LoadInt64(cfg.failedRequests)

	fmt.Println()
	fmt.Println("=============================")
	fmt.Println("  CRAWLING STATISTICS")
	fmt.Println("=============================")
	fmt.Printf("Total HTTP requests: %d\n", totalReqs)
	fmt.Printf("Failed HTTP requests: %d\n", failedReqs)

	if totalReqs > 0 {
		successRate := float64(totalReqs-failedReqs) / float64(totalReqs) * 100
		fmt.Printf("Success rate: %.1f%%\n", successRate)
	}

	fmt.Printf("Unique pages discovered: %d\n", len(cfg.pages))
	fmt.Printf("External links found: %d\n", len(cfg.externalLinks))

	// Show error summary per host
	cfg.hostErrorsMu.RLock()
	if len(cfg.hostErrors) > 0 {
		fmt.Println("\nError summary by host:")
		for host, errorCount := range cfg.hostErrors {
			if errorCount != nil {
				count := atomic.LoadInt64(errorCount)
				if count > 0 {
					fmt.Printf("  %s: %d errors\n", host, count)
				}
			}
		}
	}
	cfg.hostErrorsMu.RUnlock()
}

func main() {
	// Get command line arguments (excluding program name)
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("Usage: crawler <URL> [max_concurrency] [max_pages] [batch_size] [--graph]")
		fmt.Println("  URL: The website URL to crawl")
		fmt.Println("  max_concurrency: Maximum number of concurrent goroutines (default: 10)")
		fmt.Println("  max_pages: Maximum number of pages to crawl (default: 10)")
		fmt.Println("  batch_size: Number of URLs to process in each batch (default: 5)")
		fmt.Println("  --graph: Generate a graph visualization (saves as graph.png)")
		fmt.Println("  Environment variable CRAWLER_MAX_CONCURRENCY can also be used")
		os.Exit(1)
	}

	// Check for graph flag first and remove it from args for cleaner processing
	generateGraph := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--graph" {
			generateGraph = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	if len(args) > 4 {
		fmt.Println("too many arguments provided")
		fmt.Println("Usage: crawler <URL> [max_concurrency] [max_pages] [batch_size] [--graph]")
		os.Exit(1)
	}

	// First argument - the BASE_URL
	baseURLString := args[0]

	// Second argument or environment variable - maxConcurrency
	maxConcurrency := 10 // Default value

	// Third argument - maxPages
	maxPages := 10 // Default value

	// Fourth argument - batchSize
	batchSize := 5 // Default value

	// Check if maxConcurrency was provided as command line argument
	if len(args) >= 2 {
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

	// Check if maxPages was provided as command line argument
	if len(args) >= 3 {
		if parsed, err := strconv.Atoi(args[2]); err != nil {
			fmt.Printf("Error parsing max_pages '%s': %v\n", args[2], err)
			fmt.Println("max_pages must be a positive integer")
			os.Exit(1)
		} else if parsed <= 0 {
			fmt.Println("max_pages must be a positive integer")
			os.Exit(1)
		} else {
			maxPages = parsed
		}
	}

	// Check if batchSize was provided as command line argument
	if len(args) >= 4 {
		if parsed, err := strconv.Atoi(args[3]); err != nil {
			fmt.Printf("Error parsing batch_size '%s': %v\n", args[3], err)
			fmt.Println("batch_size must be a positive integer")
			os.Exit(1)
		} else if parsed <= 0 {
			fmt.Println("batch_size must be a positive integer")
			os.Exit(1)
		} else {
			batchSize = parsed
		}
	}

	if generateGraph {
		fmt.Printf("starting crawl of: %s (max concurrency: %d, max pages: %d, batch size: %d) [Graph generation enabled]\n", baseURLString, maxConcurrency, maxPages, batchSize)
	} else {
		fmt.Printf("starting crawl of: %s (max concurrency: %d, max pages: %d, batch size: %d)\n", baseURLString, maxConcurrency, maxPages, batchSize)
	}

	// Parse the base URL
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		fmt.Printf("Error parsing base URL: %v\n", err)
		os.Exit(1)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up graceful shutdown on interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v, initiating graceful shutdown...\n", sig)
		cancel() // Cancel the context to stop all crawling
	}()

	// Initialize the config struct
	var totalRequests, failedRequests int64
	cfg := &config{
		pages:              make(map[string]int),
		externalLinks:      make(map[string]int),
		baseURL:            baseURL,
		maxPages:           maxPages,
		batchSize:          batchSize,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		ctx:                ctx, // Use the cancellable context
		hostErrors:         make(map[string]*int64),
		hostErrorsMu:       &sync.RWMutex{},
		totalRequests:      &totalRequests,
		failedRequests:     &failedRequests,
	}

	// Start crawling from the base URL
	cfg.wg.Add(1)
	go cfg.crawlPage(baseURLString)

	// Create a timeout context for very large crawls (maximum 10 minutes)
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 10*time.Minute)
	defer timeoutCancel()

	// Wait for all goroutines to complete or timeout
	done := make(chan struct{})
	go func() {
		cfg.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Normal completion
	case <-timeoutCtx.Done():
		fmt.Printf("\nCrawl timed out after 10 minutes, stopping...\n")
		cancel() // Cancel the main context
		// Give goroutines a moment to clean up
		time.Sleep(2 * time.Second)
	}

	// Print crawling statistics
	printCrawlStatistics(cfg)

	// Print the formatted report
	if err := printReport(cfg.pages, cfg.externalLinks, baseURLString); err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}

	// Generate graph visualization if requested
	if generateGraph {
		fmt.Println()
		fmt.Println("Generating graph visualization...")
		filename := "graph.png"
		if err := GenerateGraphVisualization(cfg.pages, cfg.externalLinks, baseURLString, filename); err != nil {
			fmt.Printf("Error generating graph: %v\n", err)
		}
	}
}
