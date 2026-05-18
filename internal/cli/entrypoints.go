package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tepzxl/codemap/internal/analyzer"
)

func runEntrypoints(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("entrypoints", flag.ContinueOnError)
	fs.SetOutput(stderr)

	rootPath, flagArgs := splitGraphArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}
	if rootPath == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: codemap entrypoints <path>")
		return 1
	}

	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "entrypoints failed: %v\n", err)
		return 1
	}
	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "entrypoints failed: %v\n", err)
		return 1
	}
	entrypoints, err := analyzer.DiscoverEntrypoints(loadResult, symbols)
	if err != nil {
		fmt.Fprintf(stderr, "entrypoints failed: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(map[string]any{
		"entrypoints": entrypoints,
		"warnings":    loadResult.Warnings,
		"note":        analyzer.EntrypointDiscoveryNote,
	}); err != nil {
		fmt.Fprintf(stderr, "encode entrypoints result: %v\n", err)
		return 1
	}

	return 0
}
