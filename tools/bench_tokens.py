"""Token-cost benchmark: ANAL vs JSON vs TOON vs CSV.

Generates payloads of several realistic shapes, encodes each with every
format, and reports character and token counts (tiktoken cl100k_base).

Usage:
    python tools/bench_tokens.py            # markdown table to stdout
    python tools/bench_tokens.py --json     # raw numbers as JSON
"""

from __future__ import annotations

import argparse
import csv
import io
import json
import sys
from pathlib import Path

import tiktoken

sys.path.insert(0, str(Path(__file__).parent))
import anal  # noqa: E402

try:
    import toon_format
except ImportError:  # pragma: no cover - optional dep
    toon_format = None

ENC = tiktoken.get_encoding("cl100k_base")


def tokens(text: str) -> int:
    return len(ENC.encode(text))


def encoders(data, flat: bool) -> dict[str, str]:
    out = {
        "JSON pretty": json.dumps(data, indent=2),
        "JSON min": json.dumps(data, separators=(",", ":")),
    }
    if flat:
        buf = io.StringIO()
        keys = sorted({k for row in data for k in row})
        w = csv.DictWriter(buf, fieldnames=keys, lineterminator="\n")
        w.writeheader()
        w.writerows(data)
        out["CSV"] = buf.getvalue()
    if toon_format is not None:
        out["TOON"] = toon_format.encode(data)
    out["ANAL"] = anal.dumps(data)
    return out


def payload_orders():
    return [
        {"id": i, "sku": f"P{1000 + i}", "qty": i % 5 + 1, "price": round(9.99 + i * 1.37, 2), "active": i % 3 != 0}
        for i in range(1, 4)
    ]


def payload_events(n=100):
    kinds = ["page.view", "cart.add", "order.paid", "user.login", "item.click"]
    return [
        {
            "id": 100000 + i,
            "event": kinds[i % len(kinds)],
            "user": f"u{4200 + i % 90}",
            "status": "ok" if i % 11 else "error",
            "latency_ms": 12 + (i * 7) % 480,
            "retry": i % 17 == 0,
        }
        for i in range(n)
    ]


def payload_catalog(n=500):
    cats = ["tools", "garden", "kitchen", "office"]
    return [
        {
            "sku": f"SKU-{i:05d}",
            "name": f"Product number {i} with a reasonably long name",
            "category": cats[i % len(cats)],
            "price": round(4.5 + (i * 13) % 900 + i * 0.07, 2),
            "qty": (i * 31) % 200,
            "active": i % 9 != 0,
        }
        for i in range(n)
    ]


def payload_wide(n=50):
    return [
        {f"field_{j:02d}": f"value-{i}-{j}" if j % 3 else i * j for j in range(20)}
        for i in range(n)
    ]


def payload_nested(n=100):
    return [
        {
            "id": i,
            "user": {"name": f"user{i}", "plan": {"tier": "pro" if i % 4 else "free", "seats": 1 + i % 8}},
            "tags": [f"t{i}a", f"t{i}b"],
            "score": round(i * 0.73, 2),
        }
        for i in range(n)
    ]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", action="store_true")
    args = ap.parse_args()

    cases = [
        ("orders_3", payload_orders(), True),
        ("events_100", payload_events(), True),
        ("catalog_500", payload_catalog(), True),
        ("wide_50x20", payload_wide(), True),
        ("nested_100", payload_nested(), False),
    ]

    report = {}
    for name, data, flat in cases:
        report[name] = {}
        for fmt, text in encoders(data, flat).items():
            report[name][fmt] = {"chars": len(text), "tokens": tokens(text)}

    if args.json:
        print(json.dumps(report, indent=2))
        return

    for name, results in report.items():
        print(f"\n### {name}\n")
        print("| format | chars | tokens |")
        print("|---|---:|---:|")
        for fmt, m in results.items():
            print(f"| {fmt} | {m['chars']} | {m['tokens']} |")


if __name__ == "__main__":
    main()
