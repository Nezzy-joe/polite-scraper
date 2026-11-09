#!/usr/bin/env python3
"""Headless-browser fetch of quotes.toscrape.com/js — with Playwright.

It launches headless Chromium, loads the page, lets the JavaScript run, and
then counts the `.quote` elements. Now there are 10 — the browser executed the
script that the plain fetch could only stare at. This is what a browser buys
you, and what it costs: a whole Chromium process in memory and seconds of
startup, versus a few milliseconds and a few megabytes for a plain GET.

Run under `/usr/bin/time -v` to capture wall time and peak memory.
"""
import time

from playwright.sync_api import sync_playwright

URL = "https://quotes.toscrape.com/js/"
UA = "FlyRankInternship-A9-stretch/1.0 (+https://github.com/nezzyjoe/polite-scraper)"


def main() -> None:
    start = time.perf_counter()

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(user_agent=UA)
        page.goto(URL, wait_until="networkidle", timeout=15000)
        rendered_quotes = page.locator(".quote").count()
        browser.close()

    elapsed = time.perf_counter() - start

    print(f"[browser] rendered .quote nodes : {rendered_quotes}   <-- JS ran, quotes appeared")
    print(f"[browser] render time           : {elapsed*1000:.0f} ms")


if __name__ == "__main__":
    main()
