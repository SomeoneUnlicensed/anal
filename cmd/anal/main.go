// Command anal converts between JSON and ANAL on the command line.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/SomeoneUnlicensed/anal"
)

const usageText = `anal — JSON <-> ANAL converter.

Usage:
  anal enc   read JSON from stdin, write ANAL to stdout
  anal dec   read ANAL from stdin, write JSON to stdout
  anal install [codex|claude|all]  install the ANAL skill for Codex / Claude Code
`

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usageText) }
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	switch flag.Arg(0) {
	case "enc", "encode":
		runEncode()
	case "dec", "decode":
		runDecode()
	case "install":
		if err := runInstall(flag.Args()[1:]); err != nil {
			fatal(err)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func runEncode() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatal(err)
	}
	var value any
	if err := json.Unmarshal(input, &value); err != nil {
		fatal(err)
	}
	output, err := anal.Marshal(value)
	if err != nil {
		fatal(err)
	}
	if _, err := os.Stdout.Write(output); err != nil {
		fatal(err)
	}
}

func runDecode() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatal(err)
	}
	var value any
	if err := anal.Unmarshal(input, &value); err != nil {
		fatal(err)
	}
	output, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(output))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "anal:", err)
	os.Exit(1)
}
