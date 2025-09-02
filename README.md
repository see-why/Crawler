# Crawler

A high-performance concurrent web crawler built in Go. This crawler can efficiently crawl websites with configurable concurrency and page limits while respecting domain boundaries.

## Features

- 🚀 **Concurrent crawling** with configurable goroutine limits
- 📊 **Page limit control** to prevent runaway crawling  
- 🔗 **Domain-restricted crawling** (stays within the same domain)
- 📈 **Sorted reporting** with link count analysis
- ⚡ **High performance** with significant speedup from concurrency
- 🛡️ **Thread-safe** with proper mutex synchronization
- 🌐 **Protocol preservation** (works with HTTP, HTTPS, etc.)

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

2. **Build the crawler:**
   ```bash
   go build -o crawler
   ```

3. **Run the crawler:**
   ```bash
   ./crawler "https://example.com" 10 5
   ```

### Usage

```bash
# Basic syntax
./crawler <URL> [max_concurrency] [max_pages]

# Or use go run directly
go run . <URL> [max_concurrency] [max_pages]
```

#### Parameters

- **URL** (required): The website URL to crawl
- **max_concurrency** (optional): Maximum number of concurrent goroutines (default: 10)
- **max_pages** (optional): Maximum number of pages to crawl (default: 10)

#### Examples

```bash
# Crawl with defaults (10 concurrent, 10 pages max)
./crawler "https://example.com"

# Crawl with custom concurrency and page limit
./crawler "https://wagslane.dev" 5 20

# Single-threaded crawling
./crawler "https://blog.example.com" 1 5

# High-performance crawling
./crawler "https://docs.example.com" 20 50
```

#### Environment Variables

You can also set the concurrency via environment variable:

```bash
export CRAWLER_MAX_CONCURRENCY=15
./crawler "https://example.com"
```

### Sample Output

```
starting crawl of: https://example.com (max concurrency: 5, max pages: 10)
Crawling: https://example.com
Crawling: https://example.com/about
Crawling: https://example.com/contact

=============================
  REPORT for https://example.com
=============================
Found 8 internal links to https://example.com
Found 3 internal links to https://example.com/about
Found 1 internal links to https://example.com/contact
```

## Performance

The crawler shows significant performance improvements with higher concurrency:

- **Sequential (concurrency=1)**: ~3.2 seconds for wagslane.dev
- **5 concurrent goroutines**: ~1.0 seconds (**3.3x faster**)
- **10 concurrent goroutines**: ~0.7 seconds (**4.6x faster**)

## Architecture

### Core Components

- **`main.go`**: CLI interface and configuration
- **`crawl_page.go`**: Concurrent crawling logic with thread-safe page tracking
- **`normalize_url.go`**: URL standardization for deduplication
- **`get_urls_from_html.go`**: HTML parsing to extract links
- **`get_html.go`**: HTTP client for fetching web pages

### Key Features

- **Thread-safe page counting** with mutex protection
- **Atomic check-and-add** to prevent race conditions
- **Domain boundary enforcement** to stay within target site
- **Graceful error handling** for network issues and invalid content
- **Configurable timeouts** and proper resource cleanup

## Development

### Running Tests

```bash
go test ./...
```

### Code Structure

The crawler uses a config struct to manage shared state across goroutines:

```go
type config struct {
    pages              map[string]int      // Thread-safe page visit counter
    baseURL            *url.URL           // Domain boundary enforcement
    maxPages           int                // Crawl limit
    mu                 *sync.Mutex        // Thread safety
    concurrencyControl chan struct{}      // Goroutine limiting
    wg                 *sync.WaitGroup    // Synchronization
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source and available under the MIT License.
