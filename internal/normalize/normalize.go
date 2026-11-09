// Package normalize is Stage 4: it turns raw strings into clean, typed values a
// program can sort and compare — "£51.77" becomes the number 51.77, the word
// "Three" becomes 3 — while keeping the original text alongside the clean value.
package normalize

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nezzyjoe/polite-scraper/internal/schema"
)

var ratings = map[string]int{
	"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
}

// keep only digits and the decimal point, dropping currency symbols/whitespace.
var nonPrice = regexp.MustCompile(`[^0-9.]`)

// Record converts a RawRecord into a normalized Record. It returns an error when
// a value cannot be cleaned (e.g. an unparseable price), so the caller can route
// the record to errors.json instead of guessing.
func Record(raw schema.RawRecord) (schema.Record, error) {
	price, err := priceToFloat(raw.PriceText)
	if err != nil {
		return schema.Record{}, err
	}

	rating, err := ratingToInt(raw.RatingText)
	if err != nil {
		return schema.Record{}, err
	}

	return schema.Record{
		Title:            strings.TrimSpace(raw.Title),
		ProductURL:       raw.ProductURL, // already an absolute, canonical https URL
		PriceText:        raw.PriceText,
		PriceGBP:         price,
		Rating:           rating,
		RatingText:       raw.RatingText,
		AvailabilityText: raw.AvailabilityText,
		Description:      raw.Description,
		SourcePage:       raw.SourcePage,
		FetchedAt:        raw.FetchedAt,
	}, nil
}

// priceToFloat parses "£51.77" (or "Â£51.77", "$51.77", etc.) into 51.77.
func priceToFloat(text string) (float64, error) {
	cleaned := nonPrice.ReplaceAllString(text, "")
	if cleaned == "" {
		return 0, fmt.Errorf("no numeric price in %q", text)
	}
	v, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("bad price %q: %w", text, err)
	}
	return v, nil
}

// ratingToInt maps the rating word to 1..5, case-insensitively.
func ratingToInt(word string) (int, error) {
	if n, ok := ratings[strings.ToLower(strings.TrimSpace(word))]; ok {
		return n, nil
	}
	return 0, fmt.Errorf("unknown rating %q", word)
}
