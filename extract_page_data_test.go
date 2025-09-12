package main

import (
	"reflect"
	"testing"
)

func TestExtractPageDataBasic(t *testing.T) {
	inputURL := "https://blog.boot.dev"
	inputBody := `<html><body>
		<h1>Test Title</h1>
		<p>This is the first paragraph.</p>
		<a href="/link1">Link 1</a>
		<img src="/image1.jpg" alt="Image 1">
	</body></html>`
	actual := extractPageData(inputBody, inputURL)
	expected := PageData{
		URL:            "https://blog.boot.dev",
		H1:             "Test Title",
		FirstParagraph: "This is the first paragraph.",
		OutgoingLinks:  []string{"https://blog.boot.dev/link1"},
		ImageURLs:      []string{"https://blog.boot.dev/image1.jpg"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestExtractPageDataNoH1OrParagraph(t *testing.T) {
	inputURL := "https://blog.boot.dev"
	inputBody := `<html><body>
		<a href="/link1">Link 1</a>
		<img src="/image1.jpg" alt="Image 1">
	</body></html>`
	actual := extractPageData(inputBody, inputURL)
	expected := PageData{
		URL:            "https://blog.boot.dev",
		H1:             "",
		FirstParagraph: "",
		OutgoingLinks:  []string{"https://blog.boot.dev/link1"},
		ImageURLs:      []string{"https://blog.boot.dev/image1.jpg"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestExtractPageDataMultipleLinksAndImages(t *testing.T) {
	inputURL := "https://blog.boot.dev"
	inputBody := `<html><body>
		<h1>Title</h1>
		<p>Paragraph</p>
		<a href="/one">One</a><a href="/two">Two</a>
		<img src="/img1.png"><img src="/img2.png">
	</body></html>`
	actual := extractPageData(inputBody, inputURL)
	expected := PageData{
		URL:            "https://blog.boot.dev",
		H1:             "Title",
		FirstParagraph: "Paragraph",
		OutgoingLinks:  []string{"https://blog.boot.dev/one", "https://blog.boot.dev/two"},
		ImageURLs:      []string{"https://blog.boot.dev/img1.png", "https://blog.boot.dev/img2.png"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestExtractPageDataEmptyHTML(t *testing.T) {
	inputURL := "https://blog.boot.dev"
	inputBody := ""
	actual := extractPageData(inputBody, inputURL)
	expected := PageData{
		URL:            "https://blog.boot.dev",
		H1:             "",
		FirstParagraph: "",
		OutgoingLinks:  nil,
		ImageURLs:      nil,
	}
	// Compare slices as empty or nil
	if len(actual.OutgoingLinks) != 0 {
		t.Errorf("expected OutgoingLinks to be empty, got %v", actual.OutgoingLinks)
	}
	if len(actual.ImageURLs) != 0 {
		t.Errorf("expected ImageURLs to be empty, got %v", actual.ImageURLs)
	}
	// Compare other fields
	if actual.URL != expected.URL || actual.H1 != expected.H1 || actual.FirstParagraph != expected.FirstParagraph {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
