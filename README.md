# ANAL

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

A field or item may itself be an object or array — encoded inline,
recursively, as `{key:value,...}` or `[value,...]` — so any JSON document
round-trips through ANAL:

```text
@obj: {event,order_id,meta}
order.created|8123|{tags:[vip,rush],address:{city:NYC,zip:10001}}
```

ANAL still shines most on flat, tabular data; the token savings shrink as
nesting gets deeper, since nested values fall back to JSON-like punctuation.

## API

`Marshal` accepts the same kinds of values as `encoding/json` — structs, maps,
slices, and scalars — and honors `json` struct tags. `Unmarshal` does the
reverse.

Nested objects and arrays are supported (encoded inline); ANAL's compression
advantage is largest on flat, tabular data and shrinks as nesting gets deep.

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

Token counts for the same data, measured with `tiktoken` (`cl100k_base`):

| encoding | chars | tokens |
|---|---:|---:|
| JSON | 133 | 61 |
| TOON | 81 | 38 |
| ANAL | 63 | 35 |

CPU benchmarks:

```sh
go test -bench=. -benchmem
```

## License

MIT — see [LICENSE](LICENSE).
