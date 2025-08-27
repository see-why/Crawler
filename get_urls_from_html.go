package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// getURLsFromHTML extracts all URLs from anchor tags in the HTML and converts relative URLs to absolute using rawBaseURL.
func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	var urls []string
	base, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					// Skip empty hrefs
					if href == "" {
						resolved := base.ResolveReference(&url.URL{})
						urls = append(urls, resolved.String())
						continue
					}
					
					u, err := url.Parse(href)
					if err != nil {
						// Skip malformed URLs rather than creating invalid ones
						continue
					}
					resolved := base.ResolveReference(u)
					urls = append(urls, resolved.String())
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return urls, nil
}
