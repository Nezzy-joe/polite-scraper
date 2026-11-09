// Command polite-scraper downloads the first three catalogue pages of the Books
// to Scrape practice sandbox, visits all sixty book pages, turns the messy HTML
// into clean, schema-checked JSON, survives a broken page without crashing, and
// ends every run with an honest report.
//
// It is a chain of provable stages: classify (see README) -> fetch -> extract ->
// normalize -> validate -> store -> report.
//
// Usage:
//
//	go run .                 # normal run (reads cache if present)
//	go run . --refresh       # ignore the cache and fetch everything fresh
//	go run . --inject-broken # add one fake book URL to prove one bad page is
//	                         # logged and skipped without killing the run
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/nezzyjoe/polite-scraper/internal/discover"
	"github.com/nezzyjoe/polite-scraper/internal/extract"
	"github.com/nezzyjoe/polite-scraper/internal/fetch"
	"github.com/nezzyjoe/polite-scraper/internal/normalize"
	"github.com/nezzyjoe/polite-scraper/internal/report"
	"github.com/nezzyjoe/polite-scraper/internal/schema"
	"github.com/nezzyjoe/polite-scraper/internal/store"
)

const (
	// A polite robot introduces itself. Point the link at your own repo.
	userAgent = "FlyRankInternship-A9/1.0 (+https://github.com/nezzyjoe/polite-scraper)"

	startURL    = "https://books.toscrape.com/catalogue/page-1.html"
	maxPages    = 3 // the assignment's scope: the first three catalogue pages only
	politeDelay = 600 * time.Millisecond

	cacheDir = "cache"
	outputDir = "output"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	refresh := flag.Bool("refresh", false, "ignore the cache and fetch every page fresh")
	injectBroken := flag.Bool("inject-broken", false, "append one fake book URL to prove a bad page is skipped")
	flag.Parse()

	start := time.Now()
	client := fetch.New(userAgent, cacheDir, politeDelay, *refresh)

	// --- Stage 2: discover the catalogue pages and book links ---------------
	fmt.Println("Stage 2: discovering catalogue pages...")
	disc, err := discover.Run(client, startURL, maxPages)
	if err != nil {
		return err
	}
	links := disc.Links
	fmt.Printf("  catalogue_pages=%d discovered=%d unique_urls=%d\n",
		disc.CataloguePages, disc.Discovered, len(links))

	if *injectBroken {
		links = append(links, discover.BookLink{
			URL:        "https://books.toscrape.com/catalogue/this-book-does-not-exist_9999/index.html",
			SourcePage: "https://books.toscrape.com/catalogue/page-1.html",
		})
		fmt.Println("  (injected one deliberately broken URL for the failure test)")
	}

	// --- Stages 3-4: extract, normalize, validate each book -----------------
	fmt.Printf("Stages 3-4: extracting and validating %d book pages...\n", len(links))
	var (
		valid  []schema.Record
		errs   []schema.ErrorRecord
		failed int
	)

	for _, link := range links {
		fetchedAt := time.Now().UTC().Format(time.RFC3339)

		got, err := client.Get(link.URL)
		if err != nil {
			// A broken page is logged and skipped — it never takes the run down.
			failed++
			fmt.Printf("  SKIP (fetch failed): %s -- %v\n", link.URL, err)
			errs = append(errs, schema.ErrorRecord{
				ProductURL: link.URL, Reason: "fetch failed: " + err.Error(), FetchedAt: fetchedAt,
			})
			continue
		}

		raw, err := extract.FromHTML(got.Body, link.URL, link.SourcePage, fetchedAt)
		if err != nil {
			errs = append(errs, schema.ErrorRecord{
				ProductURL: link.URL, Reason: "extract failed: " + err.Error(), FetchedAt: fetchedAt,
			})
			continue
		}

		rec, err := normalize.Record(raw)
		if err != nil {
			errs = append(errs, schema.ErrorRecord{
				ProductURL: link.URL, Reason: "normalize failed: " + err.Error(), FetchedAt: fetchedAt,
			})
			continue
		}

		// Validate before storing: a record that fails the schema goes to
		// errors.json with the reason and never sneaks into books.json.
		if err := schema.Validate(rec); err != nil {
			errs = append(errs, schema.ErrorRecord{
				ProductURL: link.URL, Reason: "validation failed: " + err.Error(), FetchedAt: fetchedAt,
			})
			continue
		}
		valid = append(valid, rec)
	}

	// --- Stage 4: dedupe + store --------------------------------------------
	books := store.Dedupe(valid)
	if err := store.WriteJSON(outputDir+"/books.json", books); err != nil {
		return err
	}
	if err := store.WriteJSON(outputDir+"/errors.json", errs); err != nil {
		return err
	}

	// --- Stage 5: the honest run report -------------------------------------
	rep := report.Report{
		StartedAt:       start.UTC().Format(time.RFC3339),
		DurationSeconds: time.Since(start).Seconds(),
		CataloguePages:  disc.CataloguePages,
		Discovered:      disc.Discovered,
		UniqueURLs:      len(disc.Links),
		PagesFetched:    client.Stats.PagesFetched,
		CacheHits:       client.Stats.CacheHits,
		ValidRecords:    len(books),
		InvalidRecords:  len(errs),
		FailedPages:     failed,
	}
	if err := store.WriteJSON(outputDir+"/run-report.json", rep); err != nil {
		return err
	}

	fmt.Println("Done.")
	fmt.Printf("  valid_records=%d invalid_records=%d failed_pages=%d\n",
		rep.ValidRecords, rep.InvalidRecords, rep.FailedPages)
	fmt.Printf("  pages_fetched=%d cache_hits=%d duration=%.1fs\n",
		rep.PagesFetched, rep.CacheHits, rep.DurationSeconds)
	fmt.Printf("  wrote %s/books.json, %s/errors.json, %s/run-report.json\n",
		outputDir, outputDir, outputDir)
	return nil
}
