package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func TestExportCommandJSON(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"export",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--depth", "5",
		"--format", "json",
		"--direction", "downstream",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("export command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("export json output is not graph json: %v\n%s", err, stdout.String())
	}
	if output.Entry != "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main" {
		t.Fatalf("entry = %q", output.Entry)
	}
	requireGraphEdge(t, output,
		"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
		"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
		graphmodel.EdgeResolutionResolved,
		false,
	)
}

func TestExportCommandMermaidAndDOT(t *testing.T) {
	repoRoot := findRepoRoot(t)

	tests := []struct {
		format string
		want   string
	}{
		{format: "mermaid", want: "flowchart LR\n"},
		{format: "dot", want: "digraph codemap {\n"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run([]string{
				"export",
				filepath.Join(repoRoot, "examples", "layered-service"),
				"--entry", "main.main",
				"--depth", "5",
				"--format", tt.format,
			}, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("export command exit code = %d, stderr = %s", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("export %s output missing %q:\n%s", tt.format, tt.want, stdout.String())
			}
			if !strings.Contains(stdout.String(), "UserService.CreateUser") {
				t.Fatalf("export %s output missing expected node label:\n%s", tt.format, stdout.String())
			}
		})
	}
}

func TestExportCommandOutFileAndInvalidFormat(t *testing.T) {
	repoRoot := findRepoRoot(t)
	outFile := filepath.Join(t.TempDir(), "graph.mmd")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"export",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--depth", "5",
		"--format", "mermaid",
		"--out", outFile,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("export command exit code = %d, stderr = %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected stdout to be empty when --out is used, got %s", stdout.String())
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected out file: %v", err)
	}
	if !strings.HasPrefix(string(data), "flowchart LR\n") {
		t.Fatalf("out file content mismatch: %s", string(data))
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"export",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--format", "svg",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected invalid format to fail")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected stdout to be empty on error, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "format must be one of") {
		t.Fatalf("expected invalid format error, got %q", stderr.String())
	}
}
