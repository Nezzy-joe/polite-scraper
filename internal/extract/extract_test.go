package extract

import (
	"os"
	"testing"
)

const (
	testURL    = "https://books.toscrape.com/catalogue/example_1/index.html"
	testSource = "https://books.toscrape.com/catalogue/page-1.html"
	testTime   = "2026-08-06T10:00:00Z"
)

// Parser handles messy whitespace: title, availability and description are all
// collapsed to single spaces and trimmed.
func TestExtractCollapsesWhitespace(t *testing.T) {
	html := mustRead(t, "../../fixtures/book_whitespace.html")
	rec, err := FromHTML(html, testURL, testSource, testTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Title != "The Messy Whitespace Book" {
		t.Errorf("title = %q, want collapsed single spaces", rec.Title)
	}
	if rec.AvailabilityText != "In stock (5 available)" {
		t.Errorf("availability = %q, want collapsed text", rec.AvailabilityText)
	}
	if rec.RatingText != "Four" {
		t.Errorf("rating_text = %q, want \"Four\"", rec.RatingText)
	}
	if rec.Description == nil || *rec.Description != "A description with irregular spacing that should be collapsed." {
		t.Errorf("description not collapsed as expected: %v", rec.Description)
	}
}

// Parser survives a page with no description: it stores null, never invented text.
func TestExtractMissingDescriptionIsNil(t *testing.T) {
	html := mustRead(t, "../../fixtures/book_no_description.html")
	rec, err := FromHTML(html, testURL, testSource, testTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Description != nil {
		t.Errorf("expected nil description, got %q", *rec.Description)
	}
	if rec.Title != "The Book Without A Description" {
		t.Errorf("title = %q", rec.Title)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return b
}
