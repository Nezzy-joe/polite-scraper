#!/usr/bin/env python3
"""Plain HTTP fetch of quotes.toscrape.com/js — no browser.

It downloads the page with a single GET and counts the rendered `.quote`
elements in the raw HTML. The point: a plain HTML parser sees ZERO quotes,
because the quotes are built by JavaScript at runtime and are not in the markup
the server sends. (The underlying data does sit in a <script> as a JS array —
which is the real lesson: reach for that data layer, not a browser.)

Run under `/usr/bin/time -v` to capture wall time and peak memory.
"""
import re
import time
import urllib.request

URL = "https://quotes.toscrape.com/js/"
UA = "FlyRankInternship-A9-stretch/1.0 (+https://github.com/nezzyjoe/polite-scraper)"


def main() -> None:
    start = time.perf_counter()

    req = urllib.request.Request(URL, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=10) as resp:
        html = resp.read().decode("utf-8", "replace")

    elapsed = time.perf_counter() - start

    # What a normal HTML scraper would select: rendered .quote DOM nodes.
    rendered_quotes = len(re.findall(r'class="quote"', html))
    # Where the data actually hides on this page: a JS array in a <script>.
    data_in_script = "var data" in html

    print(f"[plain]  html bytes            : {len(html)}")
    print(f"[plain]  rendered .quote nodes : {rendered_quotes}   <-- a parser sees nothing")
    print(f"[plain]  data hidden in <script> var data: {data_in_script}")
    print(f"[plain]  fetch time            : {elapsed*1000:.0f} ms")


if __name__ == "__main__":
    main()
