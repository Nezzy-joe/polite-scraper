# The polite scraper

FlyRank Internship · Backend Track · Week 5 · Assignment A9

A small, polite scraping pipeline. It downloads the first three catalogue pages
of the [Books to Scrape](https://books.toscrape.com) practice sandbox, visits
all 60 book pages, turns messy HTML into clean, schema-checked JSON, survives a
broken page without crashing, and ends every run with an honest report.

It is a chain of provable stages:

```
classify → fetch → extract → normalize → validate → store → report
```

> **Language note:** the assignment suggests JavaScript or Python. I built it in
> **Go** on purpose — the standard library gives a first-class, honest HTTP
> client (user-agent, timeout, status codes), and a typed `struct` + the
> `go-playground/validator` tags are a direct equivalent of Zod/Pydantic. Every
> deliverable requirement is language-agnostic, so nothing on the rubric is lost.

---

## Run it (≈5 minutes)

Requires [Go 1.21+](https://go.dev/dl/) and nothing else — no database, no paid
proxy, no cloud account, no credit card.

```bash
git clone https://github.com/nezzyjoe/polite-scraper.git
cd polite-scraper
go run .
```

The first run makes 63 real, politely-spaced requests (3 catalogue + 60 books)
and takes about 45 seconds. It writes:

- `output/books.json` — the 60 clean, validated records
- `output/errors.json` — any record that failed validation, with the reason
- `output/run-report.json` — honest numbers about the run

Run it again and it finishes in under a second, entirely from the local cache,
producing **the same 60 records** (see *Idempotency* below).

### Flags

| Command | What it does |
| --- | --- |
| `go run .` | Normal run. Reads the cache when present. |
| `go run . --refresh` | Ignore the cache and fetch every page fresh. |
| `go run . --inject-broken` | Add one fake book URL to prove a bad page is logged and skipped without killing the run. |

Run the tests (no network needed — they use fixtures in `fixtures/`):

```bash
go test ./...
```

---

## Stage 0 — Target classification

- **Site:** [books.toscrape.com](https://books.toscrape.com) — a public sandbox
  built by [Zyte/Scrapinghub](https://toscrape.com) *specifically so people can
  practise scraping on it*. That stated purpose is the permission, and a site
  like this is the only kind this project touches.
- **Scope:** the first **3 catalogue pages only** (`catalogue/page-1.html`
  through `page-3.html`), which is **60 book pages**.
- **Data collected:** for each book — title, URL, price, availability, rating,
  and description (the eight raw fields below).
- **robots result:** `GET https://books.toscrape.com/robots.txt` returns
  **HTTP 404 — no robots file found**. A missing robots file is not permission;
  it is simply an absent file. Nothing on the site disallows this practice use.

> I will not reuse this code on another site without checking its rules and
> terms first.

---

## The record schema

Every `output/books.json` record has this shape. The raw `price_text` is kept
right next to the clean numeric `price_gbp`, and `source_page` + `fetched_at`
are the **provenance** — the receipt showing where and when each fact came from.

| Field | Type | Notes |
| --- | --- | --- |
| `title` | string, required | |
| `product_url` | string, required | absolute `https://` URL; the record's canonical identity |
| `price_text` | string, required | the raw value, e.g. `"£51.77"` |
| `price_gbp` | number, > 0 | the clean value, e.g. `51.77` |
| `rating` | int, 1–5 | `"Three"` → `3` |
| `rating_text` | string, required | the raw rating word |
| `availability_text` | string, required | e.g. `"In stock (22 available)"` |
| `description` | string **or `null`** | `null` when the page has none — never invented |
| `source_page` | string, required | the catalogue page the book was discovered on |
| `fetched_at` | string, required | RFC 3339 UTC timestamp of the fetch |

Records are validated against this schema **before** storage. A record that
fails lands in `errors.json` with the reason and never sneaks into `books.json`.

---

## Politeness rules

Every request to the real site goes through one place (`internal/fetch`), where:

- **User-agent** names me and links back to this repo:
  `FlyRankInternship-A9/1.0 (+https://github.com/nezzyjoe/polite-scraper)`.
- **Timeout** of 10 s — a request gives up instead of hanging forever.
- **Status check** — only `200` is treated as a page; anything else is a failed
  fetch, not HTML to parse.
- **Delay** of at least 600 ms between real requests. Cached reads wait for
  nothing — they never leave the computer.
- **Cache** — the first fetch of a page is saved to `cache/`; development reads
  the saved copy instead of asking the site again. (`cache/` is `.gitignore`d.)
- **One retry** on a timeout or `5xx`, after a short pause. A `404` or `403` is
  **never** retried — asking again will not create a missing page, and hammering
  a `403` is how a polite robot becomes a pest.

---

## Idempotency — why a rerun gives the same 60

Each record's identity is its canonical `product_url`. Before writing, records
are de-duplicated by that URL and sorted, so running the scraper twice produces
exactly the same 60 records, never 120. That is what makes re-running a
half-finished job safe.

---

## Surviving a broken page

Each page is handled on its own, so one broken page is logged and skipped while
the other 59 survive. To prove it:

```bash
go run . --inject-broken
```

The run still finishes, `books.json` still holds its 60 good records, the fake
URL is recorded in `errors.json`, and `run-report.json` shows `failed_pages: 1`.

---

## A real run report

Pasted from `output/run-report.json` after a clean run:

```json
{
  "started_at": "2026-11-09T07:04:08Z",
  "duration_seconds": 44.5,
  "catalogue_pages": 3,
  "discovered": 60,
  "unique_urls": 60,
  "pages_fetched": 63,
  "cache_hits": 0,
  "valid_records": 60,
  "invalid_records": 0,
  "failed_pages": 0
}
```

## Why this assignment needs no browser

The data is already in the HTML the server sends — titles, prices, ratings and
descriptions are all in the markup of a plain `GET`. A headless browser would
only re-render that same HTML, adding seconds and hundreds of megabytes of
memory per page for nothing. (The deeper case — sites that hide their data in
JSON endpoints behind JavaScript — is next week's assignment, A16.)

---

## One honest limitation

The sandbox's own HTML **repeats each book's description** inside a single `<p>`
— a truncated preview followed by the full text. The scraper stores that `<p>`'s
text verbatim rather than guessing which half to drop, because inventing a rule
to trim it risks corrupting legitimate descriptions and would break provenance
("this is exactly what the page said"). A future version could add a documented,
tested de-duplication pass if downstream consumers need it.

---

## Ethics note

- Use an official API when one exists; scrape only when there isn't one.
- Never bypass logins, paywalls, or blocks, and never scrape behind
  authentication you were not given.
- Collect only what you need — here, the eight fields above from a sandbox built
  for practice, and nothing more.
- Identify yourself honestly, go slowly, and cache so you never hammer a site.

---

## Project layout

```
polite-scraper/
├── main.go                  orchestrates the stages end to end
├── internal/
│   ├── fetch/               polite HTTP: user-agent, timeout, status, delay, cache, retry
│   ├── discover/            Stage 2: find 3 catalogue pages and 60 book links
│   ├── extract/             Stage 3: read the 8 raw fields off a book page
│   ├── normalize/           Stage 4: "£51.77" → 51.77, "Three" → 3
│   ├── schema/              the record shape + the validation rule
│   ├── store/               de-dupe by canonical URL + write JSON
│   └── report/              the run-report shape
├── fixtures/                small HTML files for offline parser tests
├── output/                  books.json, errors.json, run-report.json (sample committed)
└── cache/                   generated, git-ignored
```
