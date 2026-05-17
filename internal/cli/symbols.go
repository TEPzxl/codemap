package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tepzxl/codemap/internal/analyzer"
)

type symbolsOutput struct {
	Root     string                    `json:"root"`
	Symbols  []analyzer.Symbol         `json:"symbols"`
	Warnings []analyzer.AnalyzeWarning `json:"warnings"`
}

func runSymbols(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("symbols", flag.ContinueOnError)
	fs.SetOutput(stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: codemap symbols <path>")
		return 1
	}

	rootPath := fs.Arg(0)
	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "symbols failed: %v\n", err)
		return 1
	}

	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "symbols failed: %v\n", err)
		return 1
	}

	output := symbolsOutput{
		Root:     rootPath,
		Symbols:  symbols,
		Warnings: loadResult.Warnings,
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fmt.Fprintf(stderr, "encode symbols result: %v\n", err)
		return 1
	}

	return 0
}
