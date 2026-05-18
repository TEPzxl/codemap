package cli

import (
	"fmt"
	"io"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printRootHelp(stdout)
		return 0
	}

	switch args[0] {
	case "calls":
		return runCalls(args[1:], stdout, stderr)
	case "entrypoints":
		return runEntrypoints(args[1:], stdout, stderr)
	case "graph":
		return runGraph(args[1:], stdout, stderr)
	case "packages":
		return runPackages(args[1:], stdout, stderr)
	case "path":
		return runPath(args[1:], stdout, stderr)
	case "scan":
		return runScan(args[1:], stdout, stderr)
	case "serve":
		return runServe(args[1:], stdout, stderr)
	case "symbols":
		return runSymbols(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printRootHelp(stderr)
		return 1
	}
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "codemap - local-first go call graph tool")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  codemap <command> [arguments]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available Commands:")
	fmt.Fprintln(w, "  calls   list go function and method calls under a local path")
	fmt.Fprintln(w, "  entrypoints   list heuristic graph entrypoint candidates")
	fmt.Fprintln(w, "  graph   build a depth-limited call graph from an entry symbol")
	fmt.Fprintln(w, "  packages   build a package-level call overview")
	fmt.Fprintln(w, "  path   find call paths between two symbols")
	fmt.Fprintln(w, "  scan   scan go packages under a local path")
	fmt.Fprintln(w, "  serve   start local http api server for a go project")
	fmt.Fprintln(w, "  symbols   list go function and method symbols under a local path")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -h, --help   help for codemap")
}
