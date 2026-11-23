# ANAL

[![CI](https://github.com/SomeoneUnlicensed/anal/actions/workflows/ci.yml/badge.svg)](https://github.com/SomeoneUnlicensed/anal/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/SomeoneUnlicensed/anal.svg)](https://pkg.go.dev/github.com/SomeoneUnlicensed/anal)
[![Go Report Card](https://goreportcard.com/badge/github.com/SomeoneUnlicensed/anal)](https://goreportcard.com/report/github.com/SomeoneUnlicensed/anal)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

ANAL (Agent Notation, Actually Lean) is a compact, token-lean serialization
format for data flowing between agents and models. It needs no model
retraining: it wins by matching the tokenizer rather than fighting it.

## Why

LLMs bill by tokens, not characters. JSON repeats field names, quotes, and
brackets for every value. ANAL declares field names once in a header and then
sends only values.

## Install

```sh
go get github.com/SomeoneUnlicensed/anal
```

## Quick start

```go
package main

import (
	"fmt"

	"github.com/SomeoneUnlicensed/anal"
)

func main() {
	rows := []map[string]any{
		{"id": 1, "sku": "P123", "qty": 2, "active": true},
		{"id": 2, "sku": "P124", "qty": 1, "active": false},
	}
	data, _ := anal.Marshal(rows)
	fmt.Print(string(data))
}
```

Output:

```text
@table: {active,id,qty,sku}
T|1|2|P123
F|2|1|P124
```

## Format

```text
@table: {id,sku,qty,price,active}
1|P123|2|100|T
2|P124|1|150|F

@obj: {event,order_id,status,total}
order.created|8123|paid|199.9

@list:
code
coffee
```

- `@table` — a list of objects, one `|`-separated row each.
- `@obj` — a single object.
- `@list` — a list of values, one per line.

Values: `T` / `F` for booleans, `~` for null, numbers as-is, strings as-is
unless ambiguous or containing special characters — then backtick-quoted.

Escapes: `\\` `` \` `` `\|` `\n` `\t` `\r`.

### Nested data

Nested objects flatten into dot-joined header paths; arrays and empty
objects become `&`-prefixed inline JSON cells:

```text
@table: {id,user.name,user.plan.tier,tags}
1|alice|pro|&["a","b"]
2|bob|free|&[]
```

decodes back to

```json
[
  {"id": 1, "user": {"name": "alice", "plan": {"tier": "pro"}}, "tags": ["a", "b"]},
  {"id": 2, "user": {"name": "bob", "plan": {"tier": "free"}}, "tags": []}
]
```

Dots, commas and braces in literal key names are escaped (`a\.b`). Any JSON
value encodes — agents never have to bail out to another format mid-stream.

## API

`Marshal` accepts the same kinds of values as `encoding/json` — structs, maps,
slices, and scalars — and honors `json` struct tags. `Unmarshal` does the
reverse, rebuilding nested objects from dot-joined paths.

## Integrate

Install the CLI:

```sh
go install github.com/SomeoneUnlicensed/anal/cmd/anal@latest
```

Install the ANAL skill so coding agents know to compress large structured data:

```sh
anal install          # Codex + Claude Code
anal install codex    # Codex only
anal install claude   # Claude Code only
```

```sh
echo '[{"id":1,"sku":"P123","qty":2}]' | anal enc
```

Python helpers live in `tools/`:

- `analctl.py` — convert JSON <-> ANAL and `diff` to measure token savings.
- `agent_shim.py` — `wrap_tool_result()` to drop compact tool results into
  context, and `decode()` to read them back.

```sh
python tools/analctl.py diff < payload.json
```

## Benchmark

Token counts with `tiktoken` (`cl100k_base`) — full tables and methodology
in [BENCHMARKS.md](BENCHMARKS.md):

| payload | JSON min | TOON | ANAL |
|---|---:|---:|---:|
| events_100 (flat) | 3022 | 1583 | 1634 |
| catalog_500 (flat) | 18752 | 12504 | 12765 |
| nested_100 | 3902 | 5000 | **2420** |

On flat numeric tables ANAL lands within a few percent of CSV/TOON; on
nested or mixed data it wins by ~2x against TOON and ~38% against minified
JSON, and keeps working where CSV can't encode at all.

```sh
pip install tiktoken toon-format
python tools/bench_tokens.py   # reproduce the tables
go test -bench=. -benchmem     # CPU benchmarks
```

## License

MIT — see [LICENSE](LICENSE).
