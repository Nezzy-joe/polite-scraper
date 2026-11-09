// Package report is Stage 5's honest ending: every run writes a handful of real
// numbers so a scraper can never fail silently for weeks before anyone notices.
package report

// Report is the shape of output/run-report.json.
type Report struct {
	StartedAt       string `json:"started_at"`
	DurationSeconds float64 `json:"duration_seconds"`
	CataloguePages  int    `json:"catalogue_pages"`
	Discovered      int    `json:"discovered"`
	UniqueURLs      int    `json:"unique_urls"`
	PagesFetched    int    `json:"pages_fetched"`
	CacheHits       int    `json:"cache_hits"`
	ValidRecords    int    `json:"valid_records"`
	InvalidRecords  int    `json:"invalid_records"`
	FailedPages     int    `json:"failed_pages"`
}
