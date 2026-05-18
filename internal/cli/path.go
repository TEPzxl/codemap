package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func runPath(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("path", flag.ContinueOnError)
	fs.SetOutput(stderr)

	from := fs.String("from", "", "source symbol id or query")
	to := fs.String("to", "", "target symbol id or query")
	maxDepth := fs.Int("max-depth", 8, "maximum path depth")
	limit := fs.Int("limit", 5, "maximum paths to return")
	showExternal := fs.Bool("show-external", false, "include external calls")
	showUnresolved := fs.Bool("show-unresolved", false, "include unresolved calls")
	showInterface := fs.Bool("show-interface", false, "include interface calls")
	expandInterface := fs.Bool("expand-interface", false, "include candidate concrete implementations for interface calls")
	var packagePrefixes stringListFlag
	fs.Var(&packagePrefixes, "package", "include only path nodes whose package has this prefix; repeatable")

	rootPath, flagArgs := splitGraphArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}
	if rootPath == "" || fs.NArg() != 0 || *from == "" || *to == "" {
		fmt.Fprintln(stderr, "usage: codemap path <path> --from <symbol> --to <symbol> --max-depth <n> --limit <n>")
		return 1
	}

	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "path failed: %v\n", err)
		return 1
	}

	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "path failed: %v\n", err)
		return 1
	}

	calls, err := analyzer.ExtractCallsWithOptions(loadResult, symbols, analyzer.CallOptions{
		ExpandInterface: *expandInterface,
	})
	if err != nil {
		fmt.Fprintf(stderr, "path failed: %v\n", err)
		return 1
	}

	result, err := graphmodel.FindPaths(toGraphSymbols(symbols), toGraphCalls(calls), graphmodel.PathOptions{
		From:            *from,
		To:              *to,
		MaxDepth:        *maxDepth,
		Limit:           *limit,
		ShowExternal:    *showExternal,
		ShowUnresolved:  *showUnresolved,
		ShowInterface:   *showInterface,
		ExpandInterface: *expandInterface,
		PackagePrefixes: packagePrefixes.values(),
	})
	if err != nil {
		fmt.Fprintf(stderr, "path failed: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(stderr, "encode path result: %v\n", err)
		return 1
	}

	return 0
}
