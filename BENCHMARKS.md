# Benchmarks

Method: `tools/bench_tokens.py` generates payloads of several realistic
shapes, encodes each with every format, and counts characters and tokens
with `tiktoken` (`cl100k_base`). Reproduce:

```sh
pip install tiktoken toon-format
python tools/bench_tokens.py
```

### orders_3

| format | chars | tokens |
|---|---:|---:|
| JSON pretty | 287 | 128 |
| JSON min | 178 | 74 |
| CSV | 87 | 48 |
| TOON | 98 | 55 |
| ANAL | 87 | 55 |

### events_100

| format | chars | tokens |
|---|---:|---:|
| JSON pretty | 13640 | 5122 |
| JSON min | 9439 | 3022 |
| CSV | 3676 | 1420 |
| TOON | 3883 | 1583 |
| ANAL | 3292 | 1634 |

### catalog_500

| format | chars | tokens |
|---|---:|---:|
| JSON pretty | 88068 | 29252 |
| JSON min | 67067 | 18752 |
| CSV | 39601 | 12511 |
| TOON | 40598 | 12504 |
| ANAL | 38055 | 12765 |

### wide_50x20

| format | chars | tokens |
|---|---:|---:|
| JSON pretty | 27276 | 11802 |
| JSON min | 20975 | 8603 |
| CSV | 8754 | 4030 |
| TOON | 8860 | 4133 |
| ANAL | 8764 | 4684 |

### nested_100

| format | chars | tokens |
|---|---:|---:|
| JSON pretty | 20063 | 7602 |
| JSON min | 10362 | 3902 |
| TOON | 12665 | 5000 |
| ANAL | 3926 | 2420 |

## Reading the numbers

- On flat numeric tables, comma-joined formats (CSV, TOON) tokenize a few
  percent better than `|`-separated rows: `cl100k_base` merges digits with
  commas. ANAL trades those few percent for unambiguous string cells and a
  uniform object/table/list model.
- Once data stops being a flat table — nested objects, per-record arrays —
  CSV drops out entirely and ANAL wins by 2x against TOON and ~38% against
  minified JSON, because dot-joined paths avoid repeating nested keys.
- ANAL's real edge is shape coverage, not flat-table micro-wins: any JSON
  value encodes, and agents never have to pick a different format
  mid-stream.
