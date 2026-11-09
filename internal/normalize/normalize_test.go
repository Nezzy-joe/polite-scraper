package normalize

import (
	"testing"

	"github.com/nezzyjoe/polite-scraper/internal/schema"
)

func TestPriceToFloat(t *testing.T) {
	cases := []struct {
		in      string
		want    float64
		wantErr bool
	}{
		{"£51.77", 51.77, false},
		{"Â£51.77", 51.77, false}, // mojibake pound sign still parses
		{"$12.00", 12.00, false},
		{"  £7.50  ", 7.50, false},
		{"free", 0, true}, // no digits -> error, not a silent zero
		{"", 0, true},
	}
	for _, c := range cases {
		got, err := priceToFloat(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("priceToFloat(%q): expected error, got %v", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("priceToFloat(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("priceToFloat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRatingToInt(t *testing.T) {
	cases := map[string]int{"One": 1, "two": 2, "THREE": 3, "Four": 4, "Five": 5}
	for word, want := range cases {
		got, err := ratingToInt(word)
		if err != nil {
			t.Errorf("ratingToInt(%q): unexpected error %v", word, err)
		}
		if got != want {
			t.Errorf("ratingToInt(%q) = %d, want %d", word, got, want)
		}
	}
	if _, err := ratingToInt("Eleven"); err == nil {
		t.Error("ratingToInt(\"Eleven\"): expected error for unknown rating")
	}
}

func TestRecordKeepsRawTextAndProvenance(t *testing.T) {
	raw := schema.RawRecord{
		Title:            "A Light in the Attic",
		ProductURL:       "https://books.toscrape.com/catalogue/a-light-in-the-attic_1000/index.html",
		PriceText:        "£51.77",
		AvailabilityText: "In stock (22 available)",
		RatingText:       "Three",
		Description:      nil,
		SourcePage:       "https://books.toscrape.com/catalogue/page-1.html",
		FetchedAt:        "2026-08-06T10:00:00Z",
	}
	rec, err := normalizeHelper(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.PriceGBP != 51.77 {
		t.Errorf("PriceGBP = %v, want 51.77", rec.PriceGBP)
	}
	if rec.PriceText != "£51.77" {
		t.Errorf("raw price_text should be preserved, got %q", rec.PriceText)
	}
	if rec.Rating != 3 {
		t.Errorf("Rating = %d, want 3", rec.Rating)
	}
	if rec.SourcePage != raw.SourcePage || rec.FetchedAt != raw.FetchedAt {
		t.Error("provenance (source_page / fetched_at) must be carried through unchanged")
	}
}

// normalizeHelper exists so the test reads clearly; Record is the real entry point.
func normalizeHelper(raw schema.RawRecord) (schema.Record, error) { return Record(raw) }
