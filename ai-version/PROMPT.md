# The prompt — "AI vs me" bonus

This folder is **quarantined**. It is its own Go module so it never touches the
hand-built scraper or its `go test ./...`. The code here was generated from the
prompt below and nothing else — I did not paste in my hand-built code.

## Prompt v1 (written from memory, without looking at the assignment)

> Build me a small web scraper in Go.
>
> **Target:** `books.toscrape.com`, a public sandbox made for practising
> scraping. Only scrape the **first 3 catalogue pages** (that's 60 books). Start
> at the first catalogue page and follow the site's own "next" link to reach
> pages 2 and 3 — don't hardcode the book URLs.
>
> **For each book, collect:** title, the product page URL, the price,
> availability text, the star rating, and the description. Some books have no
> description — store null, don't make one up.
>
> Turn a price like `£51.77` into a number. Make the book URLs absolute (they're
> relative on the page).
>
> Before saving, check each record has the required fields; if one is bad, put it
> in `errors.json` with a reason instead of `books.json`. Don't save the same
> book twice — each URL should appear once. Save the good records to `books.json`.
>
> Be polite: send a user-agent that identifies me, set a timeout so it doesn't
> hang, check the response is 200 before parsing, wait at least half a second
> between requests, and cache each page so re-running doesn't hit the site again.
>
> If a page fails, retry it once, then log it and move on — one broken page
> shouldn't crash the run. At the end, write a `run-report.json` with: pages
> fetched, cache hits, valid records, failed pages, and how long it took.
>
> Use Go. Keep it simple.

See [v1 results and the diff in the top-level README](../README.md#bonus--ai-vs-me).

## Prompt v2 — the rematch (the five lines v1 forgot)

The diff below showed the AI did exactly what the prompt said and no more. The
gaps were in the *spec*, not the model. v2 adds the missing sentences:

> Additional requirements:
> 1. Keep the **raw `price_text`** (e.g. `"£51.77"`) on the record *next to* the
>    numeric `price_gbp`, so the original and the clean value sit side by side.
> 2. Also convert the star rating **word to a number 1–5** (`"Three"` → `3`) and
>    keep both.
> 3. Put **`source_page`** (the catalogue page the book was found on) and
>    **`fetched_at`** (a UTC timestamp) on every record — this is provenance.
> 4. **Do not retry a 404 or 403** — only retry timeouts and 5xx server errors.
> 5. Write `books.json` **sorted by URL** so two runs produce byte-identical
>    output, and validate records against a real schema (struct tags), not an
>    ad-hoc nil check.

Regenerated once into `main.go`; the changes are logged in the README.
