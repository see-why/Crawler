package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type PageData struct {
	URL            string
	H1             string
	FirstParagraph string
	OutgoingLinks  []string
	ImageURLs      []string
}

func extractPageData(html, pageURL string) PageData {
	var pd PageData
	pd.URL = pageURL
	pd.H1 = getH1FromHTML(html)
	pd.FirstParagraph = getFirstParagraphFromHTML(html)
	links, err := getURLsFromHTML(html, pageURL)
	if err == nil {
		pd.OutgoingLinks = links
	} else {
		pd.OutgoingLinks = []string{}
	}
	base, err := url.Parse(pageURL)
	if err == nil {
		imgs, imgErr := getImagesFromHTML(html, base)
		if imgErr == nil {
			pd.ImageURLs = imgs
		} else {
			pd.ImageURLs = []string{}
		}
	} else {
		pd.ImageURLs = []string{}
	}
	return pd
}

// getH1FromHTML returns the text content of the first <h1> tag, or "" if not found
func getH1FromHTML(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(doc.Find("h1").First().Text())
}

// getFirstParagraphFromHTML returns the text content of the first <p> tag in <main>, or first <p> in document if no <main> exists
func getFirstParagraphFromHTML(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ""
	}
	main := doc.Find("main")
	if main.Length() > 0 {
		p := main.Find("p").First()
		if p.Length() > 0 {
			return strings.TrimSpace(p.Text())
		}
	}
	p := doc.Find("p").First()
	if p.Length() > 0 {
		return strings.TrimSpace(p.Text())
	}
	return ""
}

// getURLsFromHTML extracts all URLs from anchor tags in the HTML and converts relative URLs to absolute using rawBaseURL.
const maxTraversalDepth = 50
const maxURLsPerPage = 1000

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	if len(htmlBody) == 0 {
		return []string{}, nil
	}
	if len(htmlBody) > 10*1024*1024 {
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
	urlSet := make(map[string]bool)
	var traverse func(*html.Node, int)
	traverse = func(n *html.Node, depth int) {
		if depth > maxTraversalDepth {
			return
		}
		if len(urlSet) >= maxURLsPerPage {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := strings.TrimSpace(attr.Val)
					if href == "" {
						resolved := base.ResolveReference(&url.URL{})
						if resolved != nil {
							normalizedURL := resolved.String()
							if !urlSet[normalizedURL] {
								urlSet[normalizedURL] = true
								urls = append(urls, normalizedURL)
							}
						}
					} else if href == "#" ||
						strings.HasPrefix(href, "mailto:") ||
						strings.HasPrefix(href, "tel:") ||
						strings.HasPrefix(href, "javascript:") ||
						strings.HasPrefix(href, "data:") {
						// skip
					} else {
						parsed, parseErr := url.Parse(href)
						if parseErr == nil {
							resolved := base.ResolveReference(parsed)
							if resolved != nil {
								normalizedURL := resolved.String()
								if !urlSet[normalizedURL] {
									urlSet[normalizedURL] = true
									urls = append(urls, normalizedURL)
								}
							}
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil && len(urlSet) < maxURLsPerPage; c = c.NextSibling {
			traverse(c, depth+1)
		}
	}
	traverse(doc, 0)
	return urls, nil
}

// getImagesFromHTML extracts all image srcs as absolute URLs
func getImagesFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}
	var urls []string
	doc.Find("img[src]").Each(func(_ int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src == "" {
			return
		}
		parsed, err := url.Parse(src)
		if err != nil {
			return
		}
		abs := baseURL.ResolveReference(parsed)
		urls = append(urls, abs.String())
	})
	return urls, nil
}
