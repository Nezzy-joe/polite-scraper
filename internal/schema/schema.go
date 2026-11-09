// Package schema defines the record shapes and the validation rule that every
// scraped record must pass before it is allowed into books.json.
//
// This is the "trust nothing you scraped" habit made concrete: a web page is
// untrusted input, so a record only becomes real data after it satisfies this
// schema. go-playground/validator plays the role Zod/Pydantic play in the two
// documented lanes.
package schema

import "github.com/go-playground/validator/v10"

// RawRecord is the unprocessed receipt taken straight from a book page in
// Stage 3. It has exactly the eight keys the assignment asks for. Description is
// a pointer so a book with no description serializes as `"description": null`
// rather than inventing text that was never on the page.
type RawRecord struct {
	Title            string  `json:"title"`
	ProductURL       string  `json:"product_url"`
	PriceText        string  `json:"price_text"`
	AvailabilityText string  `json:"availability_text"`
	RatingText       string  `json:"rating_text"`
	Description      *string `json:"description"`
	SourcePage       string  `json:"source_page"`
	FetchedAt        string  `json:"fetched_at"`
}

// Record is a normalized, validated book ready to store. It keeps the original
// price_text next to the clean price_gbp so the raw value and the derived value
// live side by side, and it keeps source_page + fetched_at as provenance.
type Record struct {
	Title            string   `json:"title"              validate:"required"`
	ProductURL       string   `json:"product_url"        validate:"required,url,startswith=https://"`
	PriceText        string   `json:"price_text"         validate:"required"`
	PriceGBP         float64  `json:"price_gbp"          validate:"gt=0"`
	Rating           int      `json:"rating"             validate:"min=1,max=5"`
	RatingText       string   `json:"rating_text"        validate:"required"`
	AvailabilityText string   `json:"availability_text"  validate:"required"`
	Description      *string  `json:"description"`
	SourcePage       string   `json:"source_page"        validate:"required,url,startswith=https://"`
	FetchedAt        string   `json:"fetched_at"         validate:"required"`
}

// ErrorRecord is a record that failed validation (or extraction). It lands in
// errors.json together with the reason, and never sneaks into books.json.
type ErrorRecord struct {
	ProductURL string `json:"product_url"`
	Reason     string `json:"reason"`
	FetchedAt  string `json:"fetched_at"`
}

var validate = validator.New()

// Validate checks a record against the schema. A nil error means the record is
// safe to store; a non-nil error is a human-readable reason for errors.json.
func Validate(r Record) error {
	return validate.Struct(r)
}
