# Crawler

A high-performance concurrent web crawler built in Go. This crawler efficiently crawls websites with configurable concurrency, page limits, and batch processing while tracking both internal and external links.

## Features

- 🚀 **Concurrent crawling** with configurable goroutine limits
- 📊 **Page limit control** to prevent runaway crawling  
- 🔗 **Domain-restricted crawling** (stays within the same domain)
- 🌍 **External link tracking** (reports links to other domains)
- 📈 **Comprehensive reporting** with sorted link count analysis
- 🎨 **Graph visualization** (generates visual network graphs of page relationships)
- ⚡ **High performance** with HTTP connection pooling and batch processing
- 🛡️ **Thread-safe** with proper mutex synchronization and race condition prevention
- 🌐 **Protocol preservation** (works with HTTP, HTTPS, etc.)
- ⚙️ **Configurable batch processing** for optimal goroutine management
- 🕐 **Context-based timeouts** for robust error handling

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Git

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/see-why/Crawler.git
   cd Crawler
   ```

2. **Initialize Go modules and install dependencies:**
   ```bash
   go mod init github.com/see-why/Crawler
   go get github.com/fogleman/gg  # For graph visualization
   ```

3. **Build the crawler:**
   ```bash
   go build -o crawler
   ```

4. **Run the crawler:**
   ```bash
   ./crawler "https://example.com" 10 5
   ```

### Usage

```bash
# Basic syntax
./crawler <URL> [max_concurrency] [max_pages] [batch_size] [--graph]

# Or use go run directly
go run . <URL> [max_concurrency] [max_pages] [batch_size] [--graph]
```

#### Parameters

- **URL** (required): The website URL to crawl
- **max_concurrency** (optional): Maximum number of concurrent goroutines (default: 10)
- **max_pages** (optional): Maximum number of pages to crawl (default: 10)
- **batch_size** (optional): Number of URLs to process in each batch (default: 5)
- **--graph** (optional): Generate a visual graph of page relationships (saves as graph.png)

#### Examples

```bash
# Crawl with defaults (10 concurrent, 10 pages max, batch size 5)
./crawler "https://example.com"

# Crawl with custom concurrency and page limit
./crawler "https://wagslane.dev" 5 20

# Single-threaded crawling with small batches
./crawler "https://blog.example.com" 1 5 2

# High-performance crawling with large batches
./crawler "https://docs.example.com" 20 50 10

# Conservative crawling for memory-constrained environments
./crawler "https://site.com" 3 15 1

# Generate graph visualization
./crawler "https://example.com" 5 10 5 --graph

# High-performance crawling with graph generation
./crawler "https://docs.example.com" 20 50 10 --graph
```

#### Environment Variables

You can also set the concurrency via environment variable:

```bash
export CRAWLER_MAX_CONCURRENCY=15
./crawler "https://example.com"
```

### Sample Output

```
starting crawl of: https://example.com (max concurrency: 5, max pages: 10, batch size: 5) [Graph generation enabled]
Crawling: https://example.com
Crawling: https://example.com/about
Crawling: https://example.com/contact

=============================
  REPORT for https://example.com
=============================
Found 8 internal links to https://example.com
Found 3 internal links to https://example.com/about
Found 1 internal links to https://example.com/contact

-----------------------------
  EXTERNAL LINKS REPORT
-----------------------------
Found 5 external links to https://github.com/example/repo
Found 3 external links to https://twitter.com/example
Found 1 external links to https://linkedin.com/company/example

Generating graph visualization...
Graph visualization saved to: graph.png
```

## Graph Visualization

The crawler can generate visual network graphs showing the relationships between pages. The graph visualization includes:

- **Blue nodes**: Internal pages (within the crawled domain)
- **Orange nodes**: External links (to other domains)
- **Node size**: Proportional to the number of links to that page
- **Edge thickness**: Proportional to the link count between pages
- **Automatic layout**: Circular layout for internal pages, linear layout for external links

### Graph Features

- **PNG output**: High-quality raster graphics suitable for viewing and sharing
- **Smart labeling**: Shortened URLs for readability
- **Color coding**: Clear visual distinction between internal and external links
- **Scalable sizing**: Node and edge sizes reflect link importance
- **Legend**: Built-in legend explaining the visualization elements

## Performance

The crawler shows significant performance improvements with optimized concurrent processing:

- **Sequential (concurrency=1)**: ~1.07 seconds for test crawls
- **20 concurrent goroutines**: ~0.37 seconds (**2.9x faster**)
- **HTTP connection pooling**: Reuses connections for better performance
- **Batch processing**: Prevents goroutine explosion while maintaining speed

### Performance Optimizations

- **Global HTTP client** with connection pooling (MaxIdleConns: 100)
- **Context-based timeouts** (15 seconds per request)
- **Batch URL processing** to control goroutine creation
- **Atomic operations** for race condition prevention

## Architecture

### Core Components

- **`main.go`**: CLI interface and configuration management
- **`crawl_page.go`**: Concurrent crawling logic with thread-safe page tracking
- **`normalize_url.go`**: URL standardization for deduplication
- **`get_urls_from_html.go`**: HTML parsing to extract links
- **`get_html.go`**: Optimized HTTP client with connection pooling
- **`graph_visualizer.go`**: Graph generation and visualization engine

### Key Features

- **Thread-safe page counting** with mutex protection
- **External link tracking** for comprehensive link analysis
- **Graph visualization** with automatic layout and smart labeling
- **Atomic check-and-add pattern** to prevent race conditions
- **Domain boundary enforcement** to stay within target site
- **Configurable batch processing** for optimal resource usage
- **HTTP connection pooling** for improved performance
- **Context-based cancellation** with proper timeout handling
- **Graceful error handling** for network issues and invalid content

## Development

### Running Tests

```bash
go test ./...
```

### Code Structure

The crawler uses a config struct to manage shared state across goroutines:

```go
type config struct {
    pages              map[string]int      // Thread-safe internal page visit counter
    externalLinks      map[string]int      // Thread-safe external link counter
    baseURL            *url.URL           // Domain boundary enforcement
    maxPages           int                // Crawl limit
    batchSize          int                // Batch processing size
    mu                 *sync.Mutex        // Thread safety
    concurrencyControl chan struct{}      // Goroutine limiting
    wg                 *sync.WaitGroup    // Synchronization
    ctx                context.Context    // Context for cancellation
}
```

### Batch Processing

The crawler processes discovered URLs in configurable batches to prevent creating too many goroutines simultaneously:

- **Small batches (1-2)**: Conservative memory usage, good for limited resources
- **Medium batches (5-10)**: Balanced performance and resource usage (default: 5)
- **Large batches (15+)**: Maximum performance for high-resource environments

### Link Tracking

The crawler tracks two types of links:

1. **Internal Links**: Links within the same domain as the base URL
   - Counted and crawled recursively (up to `max_pages` limit)
   - Reported in the main report section

2. **External Links**: Links to different domains
   - Counted but not crawled (respects domain boundaries)
   - Reported in the "EXTERNAL LINKS REPORT" section

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source and available under the MIT License.
