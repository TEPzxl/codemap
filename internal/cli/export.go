package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func runExport(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(stderr)

	entry := fs.String("entry", "", "entry symbol id or query")
	depth := fs.Int("depth", 5, "maximum traversal depth")
	format := fs.String("format", string(graphmodel.ExportFormatJSON), "export format: json, mermaid, or dot")
	out := fs.String("out", "", "write export output to a file")
	showExternal := fs.Bool("show-external", false, "include external calls")
	showUnresolved := fs.Bool("show-unresolved", false, "include unresolved calls")
	showInterface := fs.Bool("show-interface", false, "include interface calls")
	expandInterface := fs.Bool("expand-interface", false, "include candidate concrete implementations for interface calls")
	direction := fs.String("direction", string(graphmodel.DirectionDownstream), "graph traversal direction: downstream, upstream, or both")
	var packagePrefixes stringListFlag
	fs.Var(&packagePrefixes, "package", "include only graph nodes whose package has this prefix; repeatable")
	nodeLimit := fs.Int("node-limit", 0, "maximum graph nodes to return; 0 means unlimited")

	rootPath, flagArgs := splitGraphArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}
	if rootPath == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: codemap export <path> --entry <symbol> --depth <n> --format <json|mermaid|dot>")
		return 1
	}
	graphDirection := graphmodel.Direction(strings.TrimSpace(*direction))
	if !graphDirection.IsValid() {
		fmt.Fprintln(stderr, "export failed: direction must be one of downstream, upstream, both")
		return 1
	}
	exportFormat := graphmodel.ExportFormat(strings.TrimSpace(*format))
	if !exportFormat.IsValid() {
		fmt.Fprintln(stderr, "export failed: format must be one of json, mermaid, dot")
		return 1
	}

	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "export failed: %v\n", err)
		return 1
	}
	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "export failed: %v\n", err)
		return 1
	}
	calls, err := analyzer.ExtractCallsWithOptions(loadResult, symbols, analyzer.CallOptions{
		ExpandInterface: *expandInterface,
	})
	if err != nil {
		fmt.Fprintf(stderr, "export failed: %v\n", err)
		return 1
	}

	graph, err := graphmodel.BuildGraph(toGraphSymbols(symbols), toGraphCalls(calls), graphmodel.BuildOptions{
		Entry:           *entry,
		Depth:           *depth,
		Direction:       graphDirection,
		ShowExternal:    *showExternal,
		ShowUnresolved:  *showUnresolved,
		ShowInterface:   *showInterface,
		ExpandInterface: *expandInterface,
		PackagePrefixes: packagePrefixes.values(),
		NodeLimit:       *nodeLimit,
	})
	if err != nil {
		fmt.Fprintf(stderr, "export failed: %v\n", err)
		return 1
	}
	output, _, err := graphmodel.ExportGraph(graph, exportFormat)
	if err != nil {
		fmt.Fprintf(stderr, "export failed: %v\n", err)
		return 1
	}

	if strings.TrimSpace(*out) != "" {
		if err := os.WriteFile(*out, []byte(output), 0o644); err != nil {
			fmt.Fprintf(stderr, "write export output: %v\n", err)
			return 1
		}
		return 0
	}
	if _, err := io.WriteString(stdout, output); err != nil {
		fmt.Fprintf(stderr, "write export output: %v\n", err)
		return 1
	}
	return 0
}
