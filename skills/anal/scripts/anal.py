#!/usr/bin/env python3
"""ANAL codec and a tiny enc/dec CLI. Self-contained for skill use."""

from __future__ import annotations

import argparse
import json
import re
import sys

_TRUE, _FALSE, _NULL = "T", "F", "~"
_NUMBER = re.compile(r"^-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?$")
_ESCAPE = {"\\": "\\\\", "`": "\\`", "|": "\\|", "\n": "\\n", "\t": "\\t", "\r": "\\r"}
_UNESCAPE = {v: k for k, v in _ESCAPE.items()}


def dumps(value) -> str:
    if isinstance(value, dict):
        return _dump_object(value)
    if isinstance(value, list):
        return _dump_list(value)
    raise TypeError("top-level value must be an object or list")


def loads(text):
    block = None
    for line in text.splitlines():
        if not line.strip():
            if block is None:
                continue
            block["rows"].append("")
            continue
        if line.startswith("@"):
            if block is not None:
                raise ValueError("multiple blocks are not supported")
            block = _parse_header(line)
        else:
            if block is None:
                raise ValueError("row before header")
            block["rows"].append(line)
    return _build(block) if block is not None else None


def _dump_object(obj) -> str:
    fields = sorted(obj)
    header = "@obj: {" + ",".join(fields) + "}\n"
    if not fields:
        return header
    return header + "|".join(_encode(obj[field]) for field in fields) + "\n"


def _dump_list(items) -> str:
    if not items:
        return "@list:\n"
    if all(isinstance(item, dict) for item in items):
        fields = _union_fields(items)
        header = "@table: {" + ",".join(fields) + "}\n"
        if not fields:
            return header + "\n" * len(items)
        return header + "".join(
            "|".join(_encode(item.get(field)) for field in fields) + "\n" for item in items
        )
    if any(isinstance(item, dict) for item in items):
        raise TypeError("mixed lists are not supported")
    return "@list:\n" + "".join(_encode(item) + "\n" for item in items)


def _union_fields(items):
    fields = []
    for item in items:
        for key in item:
            if key not in fields:
                fields.append(key)
    return sorted(fields)


def _encode(value) -> str:
    if value is True:
        return _TRUE
    if value is False:
        return _FALSE
    if value is None:
        return _NULL
    if isinstance(value, (dict, list, tuple)):
        raise TypeError("nested structures are not supported")
    if isinstance(value, bool):
        return _TRUE if value else _FALSE
    if isinstance(value, (int, float)):
        return str(value)
    text = str(value)
    return "`" + _escape(text) + "`" if _needs_quotes(text) else text


def _needs_quotes(text) -> bool:
    return (
        text == ""
        or _NUMBER.fullmatch(text) is not None
        or text in (_TRUE, _FALSE, _NULL)
        or any(ch in text for ch in "`\\|\n\t\r")
    )


def _escape(text) -> str:
    return "".join(_ESCAPE.get(ch, ch) for ch in text)


def _unescape(text) -> str:
    out = []
    i = 0
    while i < len(text):
        if text[i] == "\\" and i + 1 < len(text):
            out.append(_UNESCAPE.get(text[i : i + 2], text[i + 1]))
            i += 2
        else:
            out.append(text[i])
            i += 1
    return "".join(out)


def _decode(raw):
    if raw == _TRUE:
        return True
    if raw == _FALSE:
        return False
    if raw == _NULL:
        return None
    if _NUMBER.fullmatch(raw):
        return float(raw) if any(ch in raw for ch in ".eE") else int(raw)
    if raw.startswith("`") and raw.endswith("`") and len(raw) >= 2:
        return _unescape(raw[1:-1])
    return raw


def _parse_header(line):
    left, _, right = line[1:].partition(":")
    parts = left.split()
    return {"kind": parts[0] if parts else "", "fields": _parse_fields(right), "rows": []}


def _parse_fields(text):
    text = text.strip()
    if not text or text == "{}":
        return []
    if text.startswith("{") and text.endswith("}"):
        return [field.strip() for field in text[1:-1].split(",") if field.strip()]
    raise ValueError(f"bad field header: {text!r}")


def _split_row(row):
    out, cur = [], []
    i = 0
    while i < len(row):
        ch = row[i]
        if ch == "\\" and i + 1 < len(row):
            cur.append(ch)
            cur.append(row[i + 1])
            i += 2
        elif ch == "|":
            out.append("".join(cur))
            cur = []
            i += 1
        else:
            cur.append(ch)
            i += 1
    out.append("".join(cur))
    return out


def _build(block):
    kind, fields, rows = block["kind"], block["fields"], block["rows"]
    if kind == "list":
        return [_decode(row) for row in rows]
    if kind == "obj":
        return {} if not fields else _row_to_dict(fields, rows[0] if rows else "")
    if kind == "table":
        if not fields:
            return [{} for _ in rows]
        return [_row_to_dict(fields, row) for row in rows]
    raise ValueError(f"unknown block kind: {kind!r}")


def _row_to_dict(fields, row):
    values = _split_row(row)
    return {field: (_decode(values[i]) if i < len(values) else None) for i, field in enumerate(fields)}


def main():
    parser = argparse.ArgumentParser(prog="anal.py", description=__doc__)
    parser.add_argument("command", choices=["enc", "dec"])
    args = parser.parse_args()
    if args.command == "enc":
        sys.stdout.write(dumps(json.load(sys.stdin)))
    else:
        sys.stdout.write(json.dumps(loads(sys.stdin.read()), indent=2))


if __name__ == "__main__":
    main()
