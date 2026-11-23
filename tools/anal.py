"""ANAL codec (reference implementation in Python).

Mirrors the Go package github.com/SomeoneUnlicensed/anal.
"""

from __future__ import annotations

import json
import re

_TRUE = "T"
_FALSE = "F"
_NULL = "~"

_NUMBER = re.compile(r"^-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?$")

_ESCAPE = {
    "\\": "\\\\",
    "`": "\\`",
    "|": "\\|",
    "\n": "\\n",
    "\t": "\\t",
    "\r": "\\r",
}
_UNESCAPE = {v: k for k, v in _ESCAPE.items()}


def dumps(value) -> str:
    """Serialize a dict or list to ANAL text."""
    if isinstance(value, dict):
        return _dump_object(value)
    if isinstance(value, list):
        return _dump_list(value)
    raise TypeError("top-level value must be an object or list")


def loads(text):
    """Parse ANAL text back into Python data."""
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
    flat = _flatten(obj)
    fields = sorted(flat)
    header = "@obj: {" + ",".join(fields) + "}\n"
    if not fields:
        return header
    return header + "|".join(_encode(flat[field]) for field in fields) + "\n"


def _dump_list(items) -> str:
    if not items:
        return "@list:\n"
    if all(isinstance(item, dict) for item in items):
        rows = [_flatten(item) for item in items]
        fields = _union_fields(rows)
        header = "@table: {" + ",".join(fields) + "}\n"
        if not fields:
            return header + "\n" * len(items)
        return header + "".join(
            "|".join(_encode(row.get(field)) for field in fields) + "\n"
            for row in rows
        )
    return "@list:\n" + "".join(_encode(item) + "\n" for item in items)


def _union_fields(rows):
    fields = []
    for row in rows:
        for key in row:
            if key not in fields:
                fields.append(key)
    return sorted(fields)


def _flatten(obj, prefix="", out=None):
    """Collapse nested dicts into dot-joined header paths."""
    if out is None:
        out = {}
    for key, value in obj.items():
        path = _escape_field(key)
        if prefix:
            path = prefix + "." + path
        if isinstance(value, dict) and value:
            _flatten(value, path, out)
        else:
            out[path] = value
    return out


def _escape_field(text) -> str:
    return "".join("\\" + ch if ch in "\\.,{}" else ch for ch in str(text))


def _encode(value) -> str:
    if value is True:
        return _TRUE
    if value is False:
        return _FALSE
    if value is None:
        return _NULL
    if isinstance(value, (dict, list, tuple)):
        return "&" + _escape(json.dumps(value, separators=(",", ":")))
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
        or text.startswith("&")
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
    if raw.startswith("&"):
        try:
            return json.loads(_unescape(raw[1:]))
        except ValueError:
            return raw
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
        return [
            field.strip()
            for field in _split_unescaped(text[1:-1], ",")
            if field.strip()
        ]
    raise ValueError(f"bad field header: {text!r}")


def _split_unescaped(text, sep):
    out, cur = [], []
    i = 0
    while i < len(text):
        ch = text[i]
        if ch == "\\" and i + 1 < len(text):
            cur.append(ch)
            cur.append(text[i + 1])
            i += 2
        elif ch == sep:
            out.append("".join(cur))
            cur = []
            i += 1
        else:
            cur.append(ch)
            i += 1
    out.append("".join(cur))
    return out


def _split_row(row):
    return _split_unescaped(row, "|")


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
    out = {}
    for i, field in enumerate(fields):
        value = _decode(values[i]) if i < len(values) else None
        _set_path(out, field, value)
    return out


def _set_path(root, raw_path, value):
    segments = _split_unescaped(raw_path, ".")
    current = root
    for segment in segments[:-1]:
        current = current.setdefault(_unescape(segment), {})
    current[_unescape(segments[-1])] = value
