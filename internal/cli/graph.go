package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func runGraph(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("graph", flag.ContinueOnError)
	fs.SetOutput(stderr)

	entry := fs.String("entry", "", "entry symbol id or query")
	depth := fs.Int("depth", 5, "maximum traversal depth")
	showExternal := fs.Bool("show-external", false, "include external calls")
	showUnresolved := fs.Bool("show-unresolved", false, "include unresolved calls")
	showInterface := fs.Bool("show-interface", false, "include interface calls")

	rootPath, flagArgs := splitGraphArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}
	if rootPath == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: codemap graph <path> --entry <symbol> --depth <n>")
		return 1
	}

	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "graph failed: %v\n", err)
		return 1
	}

	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "graph failed: %v\n", err)
		return 1
	}

	calls, err := analyzer.ExtractCalls(loadResult, symbols)
	if err != nil {
		fmt.Fprintf(stderr, "graph failed: %v\n", err)
		return 1
	}

	graph, err := graphmodel.BuildGraph(toGraphSymbols(symbols), toGraphCalls(calls), graphmodel.BuildOptions{
		Entry:          *entry,
		Depth:          *depth,
		ShowExternal:   *showExternal,
		ShowUnresolved: *showUnresolved,
		ShowInterface:  *showInterface,
	})
	if err != nil {
		fmt.Fprintf(stderr, "graph failed: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(graph); err != nil {
		fmt.Fprintf(stderr, "encode graph result: %v\n", err)
		return 1
	}

	return 0
}

func splitGraphArgs(args []string) (string, []string) {
	if len(args) == 0 {
		return "", args
	}
	if args[0] != "" && args[0][0] != '-' {
		return args[0], args[1:]
	}
	return "", args
}

func toGraphSymbols(symbols []analyzer.Symbol) []graphmodel.Symbol {
	result := make([]graphmodel.Symbol, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, graphmodel.Symbol{
			ID:        symbol.ID,
			Label:     symbol.Label,
			Kind:      symbol.Kind,
			Package:   symbol.Package,
			Receiver:  symbol.Receiver,
			File:      symbol.File,
			StartLine: symbol.StartLine,
			EndLine:   symbol.EndLine,
		})
	}
	return result
}

func toGraphCalls(calls []analyzer.Call) []graphmodel.Call {
	result := make([]graphmodel.Call, 0, len(calls))
	for _, call := range calls {
		result = append(result, graphmodel.Call{
			From:       call.From,
			To:         call.To,
			Kind:       call.Kind,
			Resolution: call.Resolution,
			Callsite:   call.Callsite,
		})
	}
	return result
}
