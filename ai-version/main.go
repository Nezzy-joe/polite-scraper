// AI-GENERATED (v2, the rematch) from PROMPT.md — quarantined for "AI vs me".
//
// v2 regenerates from prompt v2, which adds the five sentences v1 forgot:
// raw price_text kept beside price_gbp, a numeric rating, source_page +
// fetched_at provenance, no retry on 404/403, and sorted + schema-validated
// output. With those lines in the spec, the AI version converges on the
// hand-built one — the point of the bonus.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-playground/validator/v10"
)

const (
	userAgent = "my-book-scraper/2.0 (+https://github.com/nezzyjoe/polite-scraper)"
	startURL  = "https://books.toscrape.com/catalogue/page-1.html"
	maxPages  = 3
	delay     = 500 * time.Millisecond
	cacheDir  = "cache"
)

// Book is one scraped record. v2 keeps the raw price_text beside the numeric
// price_gbp, a numeric rating beside the word, and source_page + fetched_at.
type Book struct {
	Title        string  `json:"title"         validate:"required"`
	URL          string  `json:"url"           validate:"required,url,startswith=https://"`
	PriceText    string  `json:"price_text"    validate:"required"`
	PriceGBP     float64 `json:"price_gbp"     validate:"gt=0"`
	Rating       int     `json:"rating"        validate:"min=1,max=5"`
	RatingText   string  `json:"rating_text"   validate:"required"`
	Availability string  `json:"availability"  validate:"required"`
	Description  *string `json:"description"`
	SourcePage   string  `json:"source_page"   validate:"required"`
	FetchedAt    string  `json:"fetched_at"    validate:"required"`
}

type ErrRecord struct {
	URL    string `json:"url"`
	Reason string `json:"reason"`
}

type Report struct {
	PagesFetched int     `json:"pages_fetched"`
	CacheHits    int     `json:"cache_hits"`
	ValidRecords int     `json:"valid_records"`
	FailedPages  int     `json:"failed_pages"`
	DurationSec  float64 `json:"duration_sec"`
}

type link struct{ url, source string }

var (
	pagesFetched int
	cacheHits    int
	lastFetch    time.Time
	priceRe      = regexp.MustCompile(`[0-9.]+`)
	nameRe       = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	ratingWords  = map[string]int{"One": 1, "Two": 2, "Three": 3, "Four": 4, "Five": 5}
	validate     = validator.New()
)

func main() {
	start := time.Now()
	injectBroken := len(os.Args) > 1 && os.Args[1] == "--inject-broken"

	links, err := discover()
	if err != nil {
		fmt.Println("discovery failed:", err)
		os.Exit(1)
	}
	if injectBroken {
		links = append(links, link{
			url:    "https://books.toscrape.com/catalogue/nope-does-not-exist_9999/index.html",
			source: startURL,
		})
	}
	fmt.Printf("found %d book urls\n", len(links))

	books := map[string]Book{}
	var errRecords []ErrRecord
	failed := 0

	for _, l := range links {
		fetchedAt := time.Now().UTC().Format(time.RFC3339)
		html, err := fetch(l.url)
		if err != nil {
			failed++
			errRecords = append(errRecords, ErrRecord{URL: l.url, Reason: err.Error()})
			fmt.Println("skipping:", l.url)
			continue
		}
		book, err := parseBook(html, l.url, l.source, fetchedAt)
		if err != nil {
			errRecords = append(errRecords, ErrRecord{URL: l.url, Reason: err.Error()})
			continue
		}
		if err := validate.Struct(book); err != nil {
			errRecords = append(errRecords, ErrRecord{URL: l.url, Reason: err.Error()})
			continue
		}
		books[book.URL] = book
	}

	// Collect and sort by URL so output is byte-identical across runs.
	var list []Book
	for _, b := range books {
		list = append(list, b)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].URL < list[j].URL })

	writeJSON("books.json", list)
	writeJSON("errors.json", errRecords)
	writeJSON("run-report.json", Report{
		PagesFetched: pagesFetched,
		CacheHits:    cacheHits,
		ValidRecords: len(list),
		FailedPages:  failed,
		DurationSec:  time.Since(start).Seconds(),
	})
	fmt.Printf("done: %d books, %d failed\n", len(list), failed)
}

func discover() ([]link, error) {
	var links []link
	current := startURL
	for page := 0; page < maxPages && current != ""; page++ {
		html, err := fetch(current)
		if err != nil {
			return nil, err
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
		if err != nil {
			return nil, err
		}
		base, _ := url.Parse(current)
		doc.Find("article.product_pod h3 a").Each(func(_ int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				ref, _ := url.Parse(href)
				links = append(links, link{url: base.ResolveReference(ref).String(), source: current})
			}
		})
		current = ""
		if href, ok := doc.Find("li.next a").Attr("href"); ok {
			ref, _ := url.Parse(href)
			current = base.ResolveReference(ref).String()
		}
	}
	return links, nil
}

func parseBook(html []byte, u, source, fetchedAt string) (Book, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return Book{}, err
	}
	main := doc.Find(".product_main")
	title := strings.Join(strings.Fields(main.Find("h1").Text()), " ")
	priceText := strings.TrimSpace(main.Find(".price_color").Text())
	price, _ := strconv.ParseFloat(priceRe.FindString(priceText), 64)
	avail := strings.Join(strings.Fields(main.Find(".availability").Text()), " ")

	ratingText := ""
	if cls, ok := main.Find(".star-rating").Attr("class"); ok {
		ratingText = strings.TrimSpace(strings.Replace(cls, "star-rating", "", 1))
	}

	var desc *string
	if p := doc.Find("#product_description").NextFiltered("p"); p.Length() > 0 {
		t := strings.Join(strings.Fields(p.Text()), " ")
		desc = &t
	}

	return Book{
		Title:        title,
		URL:          u,
		PriceText:    priceText,
		PriceGBP:     price,
		Rating:       ratingWords[ratingText],
		RatingText:   ratingText,
		Availability: avail,
		Description:  desc,
		SourcePage:   source,
		FetchedAt:    fetchedAt,
	}, nil
}

// fetch gets a URL via cache, retrying ONCE on a timeout or 5xx but never on a
// 404/403 (v2 fix).
func fetch(u string) ([]byte, error) {
	path := filepath.Join(cacheDir, nameRe.ReplaceAllString(u, "_")+".html")
	if data, err := os.ReadFile(path); err == nil {
		cacheHits++
		return data, nil
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if !lastFetch.IsZero() {
			if d := time.Since(lastFetch); d < delay {
				time.Sleep(delay - d)
			}
		}
		lastFetch = time.Now()

		client := &http.Client{Timeout: 10 * time.Second}
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", userAgent)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue // network/timeout is transient — retry
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			if resp.StatusCode < 500 {
				break // 4xx (404/403) is not retried
			}
			continue // 5xx is transient — retry
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		os.MkdirAll(cacheDir, 0o755)
		os.WriteFile(path, body, 0o644)
		pagesFetched++
		return body, nil
	}
	return nil, lastErr
}

func writeJSON(name string, v any) {
	os.MkdirAll("output", 0o755)
	data, _ := json.MarshalIndent(v, "", "  ")
	os.WriteFile(filepath.Join("output", name), append(data, '\n'), 0o644)
}
