package main

import (
	"golang.org/x/net/html"
	"net/url"
	"strings"
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
					u, err := url.Parse(href)
					if err == nil {
						resolved := base.ResolveReference(u)
						urls = append(urls, resolved.String())
					}
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
