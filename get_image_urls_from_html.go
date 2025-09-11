package main

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// getImagesFromHTML extracts all image srcs as absolute URLs
func getImagesFromHTML(htmlBody string, baseURL *url.URL) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}
	var urls []string
	doc.Find("img[src]").Each(func(_ int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists || src == "" {
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
