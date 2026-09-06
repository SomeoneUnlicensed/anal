---
name: anal
description: Reduce token cost by encoding large structured data (JSON lists/objects) as ANAL before including or emitting it. Use when the user asks to cut token usage or when a big JSON payload would otherwise enter context verbatim; skip for tiny or irregular data, or when deep nesting would erase the savings.
metadata:
  short-description: Cut token cost by encoding data as ANAL
---

# ANAL

Encode large, flat structured data as ANAL (Agent Notation, Actually Lean)
instead of verbose JSON when tokens matter. ANAL sends field names once in a
header, then values only, and uses single-token markers for booleans and null.

## When to use

- The user asks to reduce token cost.
- A tool result or file contains large flat JSON (a list or object) that would
  otherwise sit in context verbatim.
- You are about to emit a large structured output.

ANAL saves the most on flat, tabular data (arrays of similar objects). It
still handles nested objects/arrays, but the savings shrink as nesting gets
deep. Do not use ANAL for tiny payloads, or when correctness beats cost —
fall back to JSON there.

## How

Prefer the installed binary:

```sh
anal enc   # JSON on stdin -> ANAL on stdout
anal dec   # ANAL on stdin -> JSON on stdout
```

If `anal` is not installed, run the bundled codec from this skill's directory:

```sh
python3 scripts/anal.py enc
python3 scripts/anal.py dec
```

When ANAL text will be read by a model, prepend this legend so it parses
reliably:

```text
ANAL: header lists fields; rows are values; T/F/~ mean true/false/null.
```

## Format

- `@table: {fields}` then `|`-separated rows -> a list of objects.
- `@obj: {fields}` then one row -> a single object.
- `@list:` then one value per line -> a list of scalars.
- Values: `T`/`F` for booleans, `~` for null, numbers as-is, strings as-is
  unless ambiguous or containing `` ` `` `\` `|` or newlines — then
  backtick-quoted.
- Escapes: `\\` `` \` `` `\|` `\n` `\t` `\r`.
- A value can itself be an object or array, encoded inline as
  `{key:value,...}` or `[value,...]`, recursively.
