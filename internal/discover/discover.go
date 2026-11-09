// Package discover is Stage 2: it finds the three catalogue pages and every
// book link on them. It lets the site tell it what the pages are by following
// the catalogue's own "next" link, rather than hardcoding sixty book URLs.
package discover

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/nezzyjoe/polite-scraper/internal/fetch"
)

// BookLink is one discovered book plus the catalogue page it was found on. The
// source page is kept so it can become the record's source_page provenance.
type BookLink struct {
	URL        string
	SourcePage string
}

// Result is the honest summary of a discovery run.
type Result struct {
	CataloguePages int
	Discovered     int // total links seen (before de-duplication)
	Links          []BookLink
}

// Run walks up to maxPages catalogue pages starting at startURL, collecting book
// links and de-duplicating them while preserving order.
func Run(client *fetch.Client, startURL string, maxPages int) (Result, error) {
	var res Result
	seen := make(map[string]bool)

	current := startURL
	for page := 1; page <= maxPages && current != ""; page++ {
		got, err := client.Get(current)
		if err != nil {
			return res, fmt.Errorf("catalogue page %q: %w", current, err)
		}
		res.CataloguePages++

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(got.Body)))
		if err != nil {
			return res, fmt.Errorf("parse %q: %w", current, err)
		}

		base, err := url.Parse(current)
		if err != nil {
			return res, err
		}

		// Collect book links from this page's product cards.
		doc.Find("article.product_pod h3 a").Each(func(_ int, s *goquery.Selection) {
			href, ok := s.Attr("href")
			if !ok {
				return
			}
			abs := resolve(base, href)
			res.Discovered++
			if !seen[abs] {
				seen[abs] = true
				res.Links = append(res.Links, BookLink{URL: abs, SourcePage: current})
			}
		})

		// Follow the catalogue's own "next" link; stop when there isn't one.
		current = ""
		if href, ok := doc.Find("li.next a").Attr("href"); ok {
			current = resolve(base, href)
		}
	}

	return res, nil
}

// resolve turns a relative href into an absolute URL against the page it was
// found on, using proper URL tools rather than gluing strings together, and
// normalizes the scheme to https (the live site is https-only).
func resolve(base *url.URL, href string) string {
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme == "http" {
		abs.Scheme = "https"
	}
	return abs.String()
}
