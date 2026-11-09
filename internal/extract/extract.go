// Package extract is Stage 3: it opens one book page and reads the raw record
// off it. Selectors are aimed at the product area of the page, not the whole
// document, so a second price elsewhere on the page can never be mistaken for
// the book's price.
package extract

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nezzyjoe/polite-scraper/internal/schema"
)

// FromHTML parses a book detail page into a RawRecord. productURL and sourcePage
// are passed in (they come from discovery) and fetchedAt is the receipt time.
// A book with no description yields a nil Description rather than invented text.
func FromHTML(html []byte, productURL, sourcePage, fetchedAt string) (schema.RawRecord, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return schema.RawRecord{}, fmt.Errorf("parse html: %w", err)
	}

	main := doc.Find("div.product_main").First()

	title := collapse(main.Find("h1").First().Text())
	priceText := strings.TrimSpace(main.Find("p.price_color").First().Text())
	availability := collapse(main.Find("p.availability").First().Text())
	ratingText := ratingWord(main.Find("p.star-rating").First())

	rec := schema.RawRecord{
		Title:            title,
		ProductURL:       productURL,
		PriceText:        priceText,
		AvailabilityText: availability,
		RatingText:       ratingText,
		Description:      description(doc),
		SourcePage:       sourcePage,
		FetchedAt:        fetchedAt,
	}
	if title == "" {
		return rec, fmt.Errorf("no title found in product area")
	}
	return rec, nil
}

// description reads the paragraph that follows the "Product Description" heading.
// Many books have one; some have none, in which case it returns nil.
func description(doc *goquery.Document) *string {
	p := doc.Find("#product_description").NextFiltered("p")
	if p.Length() == 0 {
		return nil
	}
	text := collapse(p.First().Text())
	if text == "" {
		return nil
	}
	return &text
}

// ratingWord pulls the word ("One".."Five") out of the "star-rating Three" class.
func ratingWord(sel *goquery.Selection) string {
	classes, _ := sel.Attr("class")
	for _, c := range strings.Fields(classes) {
		if c != "star-rating" && c != "" {
			return c
		}
	}
	return ""
}

// collapse trims and squeezes runs of whitespace into single spaces.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
