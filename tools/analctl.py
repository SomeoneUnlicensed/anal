#!/usr/bin/env python3
"""analctl — convert between JSON and ANAL, and measure token savings."""

from __future__ import annotations

import argparse
import json
import sys

import anal


def _encode():
    sys.stdout.write(anal.dumps(json.load(sys.stdin)))


def _decode():
    sys.stdout.write(json.dumps(anal.loads(sys.stdin.read()), indent=2))


def _diff():
    data = json.load(sys.stdin)
    json_text = json.dumps(data)
    anal_text = anal.dumps(data)
    try:
        import tiktoken

        encoder = tiktoken.get_encoding("cl100k_base")
        json_tokens = len(encoder.encode(json_text))
        anal_tokens = len(encoder.encode(anal_text))
        metric = "tokens"
    except ImportError:
        json_tokens = len(json_text)
        anal_tokens = len(anal_text)
        metric = "chars (install tiktoken for real token counts)"
    saved = (1 - anal_tokens / json_tokens) * 100 if json_tokens else 0.0
    print(f"JSON: {json_tokens} {metric}")
    print(f"ANAL: {anal_tokens} {metric}")
    print(f"Saved: {saved:.1f}%")


def main():
    parser = argparse.ArgumentParser(prog="analctl", description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("enc", help="JSON -> ANAL")
    sub.add_parser("dec", help="ANAL -> JSON")
    sub.add_parser("diff", help="report token savings for JSON on stdin")
    args = parser.parse_args()
    {"enc": _encode, "dec": _decode, "diff": _diff}[args.command]()


if __name__ == "__main__":
    main()
