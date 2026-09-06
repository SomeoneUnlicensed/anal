"""ANAL codec (reference implementation in Python).

Mirrors the Go package github.com/SomeoneUnlicensed/anal.
"""

from __future__ import annotations

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
    fields = sorted(obj)
    header = "@obj: {" + ",".join(fields) + "}\n"
    if not fields:
        return header
    return header + "|".join(_encode(obj[field], nested=False) for field in fields) + "\n"


def _dump_list(items) -> str:
    if not items:
        return "@list:\n"
    if all(isinstance(item, dict) for item in items):
        fields = _union_fields(items)
        header = "@table: {" + ",".join(fields) + "}\n"
        if not fields:
            return header + "\n" * len(items)
        return header + "".join(
            "|".join(_encode(item.get(field), nested=False) for field in fields) + "\n"
            for item in items
        )
    return "@list:\n" + "".join(_encode(item, nested=False) + "\n" for item in items)


def _union_fields(items):
    fields = []
    for item in items:
        for key in item:
            if key not in fields:
                fields.append(key)
    return sorted(fields)


def _encode(value, *, nested: bool) -> str:
    """Encode value. nested is True when it will be embedded inside an
    object ({}) or array ([]) literal, where "," ":" "{" "}" "[" "]" also
    need to stay unambiguous."""
    if value is True:
        return _TRUE
    if value is False:
        return _FALSE
    if value is None:
        return _NULL
    if isinstance(value, dict):
        fields = sorted(value)
        body = ",".join(f"{_encode_key(k)}:{_encode(value[k], nested=True)}" for k in fields)
        return "{" + body + "}"
    if isinstance(value, (list, tuple)):
        return "[" + ",".join(_encode(item, nested=True) for item in value) + "]"
    if isinstance(value, (int, float)):
        return str(value)
    text = str(value)
    return "`" + _escape(text) + "`" if _needs_quotes(text, nested=nested) else text


def _encode_key(key: str) -> str:
    if key == "" or key.startswith("`") or any(ch in key for ch in ":`\\\n\t\r"):
        return "`" + _escape(key) + "`"
    return key


def _needs_quotes(text, *, nested: bool) -> bool:
    if (
        text == ""
        or _NUMBER.fullmatch(text) is not None
        or text in (_TRUE, _FALSE, _NULL)
        or text.startswith("{")
        or text.startswith("[")
        or any(ch in text for ch in "`\\|\n\t\r")
    ):
        return True
    return nested and any(ch in text for ch in ",:{}[]")


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


def _decode_scalar(raw):
    if raw == _TRUE:
        return True
    if raw == _FALSE:
        return False
    if raw == _NULL:
        return None
    if _NUMBER.fullmatch(raw):
        return float(raw) if any(ch in raw for ch in ".eE") else int(raw)
    return raw


def _parse_value(s, pos, stops):
    """Parse one value from s starting at pos. Returns (value, next_pos).

    stops names the characters (outside of a `` `-quoted `` string or a
    nested {}/[] literal) that terminate a bare scalar: "|" between row
    fields, "" when the value spans the rest of s, or ",}" / ",]" between
    object/array elements.
    """
    if pos >= len(s):
        return "", pos
    ch = s[pos]
    if ch == "`":
        return _parse_quoted(s, pos)
    if ch == "{":
        return _parse_object(s, pos)
    if ch == "[":
        return _parse_array(s, pos)
    end = pos
    while end < len(s) and s[end] not in stops:
        end += 1
    return _decode_scalar(s[pos:end]), end


def _parse_quoted(s, pos):
    pos += 1  # skip opening backtick
    start = pos
    while pos < len(s):
        if s[pos] == "\\" and pos + 1 < len(s):
            pos += 2
            continue
        if s[pos] == "`":
            return _unescape(s[start:pos]), pos + 1
        pos += 1
    raise ValueError(f"unterminated quoted string {s!r}")


def _parse_object(s, pos):
    pos += 1  # skip '{'
    obj = {}
    if pos < len(s) and s[pos] == "}":
        return obj, pos + 1
    while True:
        if pos < len(s) and s[pos] == "`":
            key, pos = _parse_quoted(s, pos)
        else:
            start = pos
            while pos < len(s) and s[pos] != ":":
                pos += 1
            if pos >= len(s):
                raise ValueError(f"expected ':' in object {s!r}")
            key = s[start:pos]
        if pos >= len(s) or s[pos] != ":":
            raise ValueError(f"expected ':' in object {s!r}")
        pos += 1  # skip ':'

        value, pos = _parse_value(s, pos, ",}")
        obj[key] = value

        if pos >= len(s):
            raise ValueError(f"unterminated object {s!r}")
        if s[pos] == ",":
            pos += 1
            continue
        if s[pos] == "}":
            return obj, pos + 1
        raise ValueError(f"expected ',' or '}}' in object {s!r}")


def _parse_array(s, pos):
    pos += 1  # skip '['
    arr = []
    if pos < len(s) and s[pos] == "]":
        return arr, pos + 1
    while True:
        value, pos = _parse_value(s, pos, ",]")
        arr.append(value)

        if pos >= len(s):
            raise ValueError(f"unterminated array {s!r}")
        if s[pos] == ",":
            pos += 1
            continue
        if s[pos] == "]":
            return arr, pos + 1
        raise ValueError(f"expected ',' or ']' in array {s!r}")


def _decode_value(s):
    value, pos = _parse_value(s, 0, "")
    if pos != len(s):
        raise ValueError(f"trailing data at offset {pos} in value {s!r}")
    return value


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


def _parse_row(row):
    out = []
    pos = 0
    while True:
        value, pos = _parse_value(row, pos, "|")
        out.append(value)
        if pos >= len(row):
            return out
        if row[pos] != "|":
            raise ValueError(f"expected '|' at offset {pos} in row {row!r}")
        pos += 1


def _build(block):
    kind, fields, rows = block["kind"], block["fields"], block["rows"]
    if kind == "list":
        return [_decode_value(row) for row in rows]
    if kind == "obj":
        return {} if not fields else _row_to_dict(fields, rows[0] if rows else "")
    if kind == "table":
        if not fields:
            return [{} for _ in rows]
        return [_row_to_dict(fields, row) for row in rows]
    raise ValueError(f"unknown block kind: {kind!r}")


def _row_to_dict(fields, row):
    values = _parse_row(row)
    return {field: (values[i] if i < len(values) else None) for i, field in enumerate(fields)}
