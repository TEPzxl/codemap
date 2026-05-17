package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/tepzxl/codemap/internal/analyzer"
)

type scanOutput struct {
	Root     string                    `json:"root"`
	Packages []analyzer.PackageInfo    `json:"packages"`
	Warnings []analyzer.AnalyzeWarning `json:"warnings"`
}

func runScan(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(stderr)

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: codemap scan <path>")
		return 1
	}

	rootPath := fs.Arg(0)
	result, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootPath,
		IncludeTests: false,
	})
	if err != nil {
		fmt.Fprintf(stderr, "scan failed: %v\n", err)
		return 1
	}

	output := scanOutput{
		Root:     rootPath,
		Packages: result.Packages,
		Warnings: result.Warnings,
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fmt.Fprintf(stderr, "encode scan result: %v\n", err)
		return 1
	}

	return 0
}
