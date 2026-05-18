package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func runPackages(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("packages", flag.ContinueOnError)
	fs.SetOutput(stderr)

	entry := fs.String("entry", "", "optional entry symbol id or query")
	depth := fs.Int("depth", 5, "maximum traversal depth when entry is provided")
	showExternal := fs.Bool("show-external", false, "include external calls")
	showUnresolved := fs.Bool("show-unresolved", false, "include unresolved calls")
	showInterface := fs.Bool("show-interface", false, "include interface calls")
	expandInterface := fs.Bool("expand-interface", false, "include candidate concrete implementations for interface calls")
	showSelf := fs.Bool("show-self", false, "include package self edges")
	direction := fs.String("direction", string(graphmodel.DirectionDownstream), "graph traversal direction when entry is provided: downstream, upstream, or both")
	var packagePrefixes stringListFlag
	fs.Var(&packagePrefixes, "package", "include packages matching this prefix and directly connected packages; repeatable")

	rootPath, flagArgs := splitGraphArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}
	if rootPath == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: codemap packages <path> [--entry <symbol>] [--depth <n>]")
		return 1
	}
	graphDirection := graphmodel.Direction(strings.TrimSpace(*direction))
	if !graphDirection.IsValid() {
		fmt.Fprintln(stderr, "packages failed: direction must be one of downstream, upstream, both")
		return 1
	}

	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "packages failed: %v\n", err)
		return 1
	}

	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		fmt.Fprintf(stderr, "packages failed: %v\n", err)
		return 1
	}

	calls, err := analyzer.ExtractCallsWithOptions(loadResult, symbols, analyzer.CallOptions{
		ExpandInterface: *expandInterface,
	})
	if err != nil {
		fmt.Fprintf(stderr, "packages failed: %v\n", err)
		return 1
	}

	modulePath, _ := readModulePathForCLI(rootPath)
	result, err := graphmodel.BuildPackageGraph(toGraphSymbols(symbols), toGraphCalls(calls), graphmodel.PackageGraphOptions{
		BuildOptions: graphmodel.BuildOptions{
			Entry:           *entry,
			Depth:           *depth,
			Direction:       graphDirection,
			ShowExternal:    *showExternal,
			ShowUnresolved:  *showUnresolved,
			ShowInterface:   *showInterface,
			ExpandInterface: *expandInterface,
			PackagePrefixes: packagePrefixes.values(),
		},
		ModulePath:       modulePath,
		IncludeSelfEdges: *showSelf,
		Warnings:         toGraphWarnings(loadResult.Warnings),
	})
	if err != nil {
		fmt.Fprintf(stderr, "packages failed: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(stderr, "encode package graph result: %v\n", err)
		return 1
	}

	return 0
}

func readModulePathForCLI(root string) (string, error) {
	file, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module path not found")
}

func toGraphWarnings(warnings []analyzer.AnalyzeWarning) []graphmodel.Warning {
	result := make([]graphmodel.Warning, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, graphmodel.Warning{
			Code:    warning.Code,
			Message: warning.Message,
			File:    warning.PackageID,
		})
	}
	return result
}
