package cli

import (
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/tepzxl/codemap/internal/server"
)

func runServe(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	port := fs.Int("port", 8080, "http server port")

	rootPath, flagArgs := splitPathFirstArgs(args)
	if err := fs.Parse(flagArgs); err != nil {
		return 1
	}
	if rootPath == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: codemap serve <path> --port 8080")
		return 1
	}
	if *port < 1 || *port > 65535 {
		fmt.Fprintln(stderr, "port must be between 1 and 65535")
		return 1
	}

	project, err := server.LoadProject(rootPath)
	if err != nil {
		fmt.Fprintf(stderr, "serve failed: %v\n", err)
		return 1
	}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Fprintf(stdout, "codemap api listening on http://localhost:%d\n", *port)
	if err := http.ListenAndServe(addr, server.NewHandler(project)); err != nil {
		fmt.Fprintf(stderr, "serve failed: %v\n", err)
		return 1
	}
	return 0
}

func splitPathFirstArgs(args []string) (string, []string) {
	if len(args) == 0 {
		return "", args
	}
	if args[0] != "" && args[0][0] != '-' {
		return args[0], args[1:]
	}
	return "", args
}
