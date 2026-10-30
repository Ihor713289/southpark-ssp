import json
import os
import random
import signal
import sys
import threading
import urllib.error
import urllib.request
from pathlib import Path

from dotenv import load_dotenv
load_dotenv()

API_URL = os.getenv("API_URL", "http://localhost:8080/messages")
INTERVAL_SECONDS = int(os.getenv("INTERVAL_SECONDS", "300"))
QUOTES_FILE = Path(os.getenv("QUOTES_FILE", Path(__file__).with_name("quotes.json")))
REQUEST_TIMEOUT = 10
RETRY_ATTEMPTS = 5
RETRY_DELAY_SECONDS = 3

stop_event = threading.Event()


def load_quotes(path):
    with open(path, encoding="utf-8") as f:
        quotes = json.load(f)
    quotes = [q for q in quotes if q.get("author", "").strip() and q.get("body", "").strip()]
    if not quotes:
        sys.exit(f"No valid quotes found in {path}")
    return quotes


def send_quote(quote):
    # Only send the fields the API knows about: it rejects unknown fields.
    payload = json.dumps({"author": quote["author"], "body": quote["body"]}).encode("utf-8")
    request = urllib.request.Request(
        API_URL,
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    for attempt in range(1, RETRY_ATTEMPTS + 1):
        try:
            with urllib.request.urlopen(request, timeout=REQUEST_TIMEOUT) as response:
                print(f"Sent [{response.status}] {quote['author']}: {quote['body']}", flush=True)
            return
        except urllib.error.HTTPError as err:
            print(f"API rejected quote [{err.code}]: {err.read().decode(errors='replace')}", flush=True)
            return
        except (urllib.error.URLError, TimeoutError) as err:
            print(f"Could not reach API at {API_URL} (attempt {attempt}/{RETRY_ATTEMPTS}): {err}", flush=True)
            if stop_event.wait(RETRY_DELAY_SECONDS):
                return


def main():
    quotes = load_quotes(QUOTES_FILE)
    signal.signal(signal.SIGTERM, lambda *_: stop_event.set())
    signal.signal(signal.SIGINT, lambda *_: stop_event.set())

    print(f"Loaded {len(quotes)} quotes, sending one to {API_URL} every {INTERVAL_SECONDS}s", flush=True)
    while not stop_event.is_set():
        send_quote(random.choice(quotes))
        stop_event.wait(INTERVAL_SECONDS)
    print("Shutting down", flush=True)


if __name__ == "__main__":
    main()
