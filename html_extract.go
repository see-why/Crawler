package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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
