// Package store is Stage 4/5's writer: it deduplicates records by their
// canonical product_url (so a rerun produces the same set, never doubles it) and
// writes pretty-printed JSON that another program can read.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/nezzyjoe/polite-scraper/internal/schema"
)

// Dedupe collapses records to one per canonical product_url and returns them in
// a stable, sorted order. This is what makes re-running the scraper idempotent:
// running twice yields exactly the same records, not twice as many.
func Dedupe(records []schema.Record) []schema.Record {
	byURL := make(map[string]schema.Record, len(records))
	for _, r := range records {
		byURL[r.ProductURL] = r // last write wins; identity is the URL
	}
	out := make([]schema.Record, 0, len(byURL))
	for _, r := range byURL {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ProductURL < out[j].ProductURL })
	return out
}

// WriteJSON writes v as indented JSON, creating parent directories as needed.
func WriteJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
