// Package fetch is the polite front door of the scraper. Every request to the
// real site goes through here so the politeness rules live in exactly one
// place: an honest user-agent, a timeout, a status-code check, a delay between
// real requests, a local cache so development never hammers the site, and a
// single retry on transient failures.
package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Client is a reusable polite fetcher.
type Client struct {
	httpClient *http.Client
	userAgent  string
	cacheDir   string
	minDelay   time.Duration // minimum gap between real (non-cached) requests
	lastReal   time.Time     // when the last real request went out
	refresh    bool          // when true, ignore the cache and always fetch

	// Stats are honest counters surfaced in the run report.
	Stats Stats
}

// Stats are the counters the run report reads.
type Stats struct {
	PagesFetched int // real network requests that returned 200
	CacheHits    int // reads served from the local cache
}

// Result carries the outcome of a fetch.
type Result struct {
	Body      []byte
	Status    int
	FromCache bool
}

// New builds a Client. userAgent should name you and link back to your repo;
// minDelay must be at least 500ms per the assignment; refresh bypasses cache.
func New(userAgent, cacheDir string, minDelay time.Duration, refresh bool) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second}, // never wait forever
		userAgent:  userAgent,
		cacheDir:   cacheDir,
		minDelay:   minDelay,
		refresh:    refresh,
	}
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// cachePath turns a URL into a readable, collision-free cache filename, e.g.
// https://books.toscrape.com/catalogue/page-1.html -> catalogue__page-1.html
func (c *Client) cachePath(url string) string {
	name := url
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "books.toscrape.com/")
	name = unsafeChars.ReplaceAllString(name, "__")
	if !strings.HasSuffix(name, ".html") {
		name += ".html"
	}
	return filepath.Join(c.cacheDir, name)
}

// Get fetches a URL politely. It serves from cache when possible; otherwise it
// makes one real, rate-limited request, retrying once on a timeout or 5xx but
// never on a 404 or 403. transient reports whether a non-200 is worth retrying.
func (c *Client) Get(url string) (Result, error) {
	path := c.cachePath(url)

	if !c.refresh {
		if body, err := os.ReadFile(path); err == nil {
			c.Stats.CacheHits++
			fmt.Printf("  CACHE HIT  %-55s %d bytes\n", short(url), len(body))
			return Result{Body: body, Status: 200, FromCache: true}, nil
		}
	}

	res, err := c.getWithRetry(url)
	if err != nil {
		return res, err
	}

	// Cache the good copy for next time.
	if err := os.MkdirAll(c.cacheDir, 0o755); err == nil {
		_ = os.WriteFile(path, res.Body, 0o644)
	}
	c.Stats.PagesFetched++
	fmt.Printf("  FETCH      %-55s %d bytes\n", short(url), len(res.Body))
	return res, nil
}

// getWithRetry performs the real request, waiting out the politeness delay and
// retrying once on a transient failure.
func (c *Client) getWithRetry(url string) (Result, error) {
	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		c.waitPolitely()
		res, transient, err := c.doOnce(url)
		if err == nil {
			return res, nil
		}
		lastErr = err
		if !transient || attempt == 2 {
			break
		}
		time.Sleep(time.Second) // a short pause before the single retry
	}
	return Result{}, lastErr
}

// doOnce makes a single request. transient is true for timeouts and 5xx (retry
// is worthwhile) and false for 404/403 and other client errors (retrying a 404
// will not create the page; retrying a 403 is how a polite robot becomes a pest).
func (c *Client) doOnce(url string) (res Result, transient bool, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Result{}, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	c.lastReal = time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Result{}, true, fmt.Errorf("request failed (treated as transient): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		isTransient := resp.StatusCode >= 500
		return Result{Status: resp.StatusCode}, isTransient,
			fmt.Errorf("non-200 status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{Status: resp.StatusCode}, true, err
	}
	return Result{Body: body, Status: resp.StatusCode}, false, nil
}

// waitPolitely sleeps so consecutive real requests stay at least minDelay apart.
func (c *Client) waitPolitely() {
	if c.lastReal.IsZero() {
		return
	}
	elapsed := time.Since(c.lastReal)
	if elapsed < c.minDelay {
		time.Sleep(c.minDelay - elapsed)
	}
}

func short(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "books.toscrape.com/")
}
