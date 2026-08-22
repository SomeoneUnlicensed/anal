#!/usr/bin/env python3
"""Drop-in helpers for feeding compact data to an LLM.

Models read compact tabular data fine, but a one-line legend removes any
guessing. wrap_tool_result() returns a token-lean fragment ready to inject
into context in place of a verbose JSON dump.
"""

from __future__ import annotations

import anal

_LEGEND = "ANAL: header lists fields; rows are values; T/F/~ mean true/false/null."


def wrap_tool_result(data, legend: bool = True) -> str:
    """Encode tool output as a compact context fragment."""
    text = anal.dumps(data)
    return _LEGEND + "\n" + text if legend else text


def decode(text: str):
    """Parse an ANAL fragment back into data, ignoring any leading legend."""
    lines = text.splitlines()
    start = 0
    while start < len(lines) and not lines[start].lstrip().startswith("@"):
        start += 1
    return anal.loads("\n".join(lines[start:]))


def main():
    import json
    import sys

    data = json.load(sys.stdin)
    output = wrap_tool_result(data)
    print(output, end="")
    roundtrip = decode(output)
    assert roundtrip == data, (roundtrip, data)


if __name__ == "__main__":
    main()
