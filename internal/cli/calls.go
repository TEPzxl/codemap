package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tepzxl/codemap/internal/analyzer"
)

type callsOutput struct {
	Root     string                    `json:"root"`
	Calls    []analyzer.Call           `json:"calls"`
	Warnings []analyzer.AnalyzeWarning `json:"warnings"`
}

func runCalls(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("calls", flag.ContinueOnError)
	fs.SetOutput(stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: codemap calls <path>")
		return 1
	}

	rootPath := fs.Arg(0)
	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "calls failed: %v\n", err)
		return 1
	}

	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "calls failed: %v\n", err)
		return 1
	}

	calls, err := analyzer.ExtractCalls(loadResult, symbols)
	if err != nil {
		fmt.Fprintf(stderr, "calls failed: %v\n", err)
		return 1
	}

	output := callsOutput{
		Root:     rootPath,
		Calls:    calls,
		Warnings: loadResult.Warnings,
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fmt.Fprintf(stderr, "encode calls result: %v\n", err)
		return 1
	}

	return 0
}
