package main

import "testing"

func TestGetH1FromHTMLBasic(t *testing.T) {
	inputBody := "<html><body><h1>Test Title</h1></body></html>"
	actual := getH1FromHTML(inputBody)
	expected := "Test Title"
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetH1FromHTMLNone(t *testing.T) {
	inputBody := "<html><body><h2>No H1 here</h2></body></html>"
	actual := getH1FromHTML(inputBody)
	expected := ""
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetH1FromHTMLWhitespace(t *testing.T) {
	inputBody := "<html><body><h1>   Lots of space   </h1></body></html>"
	actual := getH1FromHTML(inputBody)
	expected := "Lots of space"
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetFirstParagraphFromHTMLMainPriority(t *testing.T) {
	inputBody := `<html><body>
		<p>Outside paragraph.</p>
		<main>
			<p>Main paragraph.</p>
		</main>
	</body></html>`
	actual := getFirstParagraphFromHTML(inputBody)
	expected := "Main paragraph."
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetFirstParagraphFromHTMLNoMain(t *testing.T) {
	inputBody := "<html><body><p>First paragraph.</p><p>Second paragraph.</p></body></html>"
	actual := getFirstParagraphFromHTML(inputBody)
	expected := "First paragraph."
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func TestGetFirstParagraphFromHTMLNone(t *testing.T) {
	inputBody := "<html><body><div>No paragraphs here</div></body></html>"
	actual := getFirstParagraphFromHTML(inputBody)
	expected := ""
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
