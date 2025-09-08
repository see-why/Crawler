package main

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const (
	// Maximum number of URLs to extract from a single page
	maxURLsPerPage = 1000
	// Maximum depth to traverse in the HTML tree
	maxTraversalDepth = 50
)

// getURLsFromHTML extracts all URLs from anchor tags in the HTML and converts relative URLs to absolute using rawBaseURL.
func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	// Early validation
	if len(htmlBody) == 0 {
		return []string{}, nil
	}

	if len(htmlBody) > 10*1024*1024 { // 10MB limit
		return nil, fmt.Errorf("HTML body too large (%d bytes, max 10MB)", len(htmlBody))
	}

	var urls []string
	base, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	urlSet := make(map[string]bool) // Use map to deduplicate URLs

	var traverse func(*html.Node, int)
	traverse = func(n *html.Node, depth int) {
		// Prevent infinite recursion and excessive depth
		if depth > maxTraversalDepth {
			return
		}

		// Stop if we've found enough URLs
		if len(urlSet) >= maxURLsPerPage {
			return
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := strings.TrimSpace(attr.Val)

					// If href is empty, resolve to base URL
					if href == "" {
						normalizedURL := base.String()
						if !urlSet[normalizedURL] {
							urlSet[normalizedURL] = true
							urls = append(urls, normalizedURL)
						}
						break
					}
					// Skip fragments and common non-page links
					if href == "#" ||
						strings.HasPrefix(href, "mailto:") ||
						strings.HasPrefix(href, "tel:") ||
						strings.HasPrefix(href, "javascript:") ||
						strings.HasPrefix(href, "data:") {
						continue
					}

					// Parse and resolve the URL
					parsed, parseErr := url.Parse(href)
					if parseErr != nil {
						continue // Skip malformed URLs
					}

					resolved := base.ResolveReference(parsed)
					if resolved != nil {
						normalizedURL := resolved.String()
						// Only add if we haven't seen this URL before
						if !urlSet[normalizedURL] {
							urlSet[normalizedURL] = true
							urls = append(urls, normalizedURL)
						}
					}
					break // Only process first href attribute
				}
			}
		}

		// Recursively traverse child nodes
		for c := n.FirstChild; c != nil && len(urlSet) < maxURLsPerPage; c = c.NextSibling {
			traverse(c, depth+1)
		}
	}

	// Start traversal from the root
	traverse(doc, 0)

	return urls, nil
}
