# Stretch — browser cost comparison

> *Fetch one page of `quotes.toscrape.com/js` with a plain HTTP request (the
> rendered quotes are not in the source), then with Playwright. Report the time
> and memory of both, and why the core assignment needed no browser.*

The core A9 target, `books.toscrape.com`, ships its data **in the HTML the server
sends** — a plain `GET` already contains every title, price and rating. This page,
`quotes.toscrape.com/js`, is the opposite: it sends an almost-empty shell and
**builds the quotes in the browser with JavaScript**. It is the small version of
next week's problem (A16), and the cleanest way to feel what a browser costs.

## What the two approaches see

- **[`plain.py`](plain.py)** — one `urllib` GET, then count `.quote` elements in
  the raw HTML. Result: **0**. The rendered quotes are not in the markup. (They
  *do* exist, as a JavaScript `var data = [...]` array inside a `<script>` tag —
  so the honest move is to read that data layer directly, not boot a browser.)
- **[`browser.py`](browser.py)** — launch headless Chromium with Playwright, load
  the page, let the JS run, then count `.quote` elements. Result: **10**. The
  browser executed the script the plain fetch could only stare at.

## Measured cost

Both were run under `/usr/bin/time -v` (same machine, same Python runtime, so the
only variable is *plain HTTP vs headless browser*). Two representative runs:

| Approach | `.quote` found | Wall time | Peak memory (RSS) |
| --- | :---: | :---: | :---: |
| Plain HTTP (`urllib`) | 0 | ~1.1–1.4 s | **~21 MB** |
| Headless Chromium (Playwright) | 10 | ~6–7 s | **~135 MB** |

The browser is roughly **5× slower and 6× heavier** — and most of the plain
fetch's ~1 s is just network latency to the live site, so the *compute* gap is
larger still. Multiply by 60 pages and the browser approach costs minutes and
gigabytes to collect data that a plain fetch gets in seconds and megabytes.

## Why the core assignment needed no browser

Because the data was already in the HTML. A headless browser only re-renders the
same markup the server already sent for `books.toscrape.com`, adding a whole
Chromium process — seconds of startup and ~130 MB of memory per page — for
nothing. You reach for a browser only when the data genuinely is not in the
source, and even then the cheaper move is often to find the JSON/JS data layer
behind the page (exactly what `var data` is here). That is next week's
assignment, A16.

## Reproduce

```bash
cd stretch/browser-cost
python3 -m venv .venv && source .venv/bin/activate
pip install playwright && playwright install chromium
./compare.sh
```
