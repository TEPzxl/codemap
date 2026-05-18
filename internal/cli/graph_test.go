package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func TestGraphCommand(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--depth", "5",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("graph command output is not graph json: %v\n%s", err, stdout.String())
	}
	if len(output.Nodes) == 0 {
		t.Fatal("expected graph command to return nodes")
	}
	if len(output.Edges) == 0 {
		t.Fatal("expected graph command to return edges")
	}
	if err := output.Validate(); err != nil {
		t.Fatalf("graph command returned invalid schema: %v", err)
	}
}

func TestGraphCommandExpandInterface(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "interface-call"),
		"--entry", "main.main",
		"--depth", "5",
		"--expand-interface",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("graph command output is not graph json: %v\n%s", err, stdout.String())
	}
	requireGraphEdge(t, output,
		"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
		"github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
		graphmodel.EdgeResolutionInterface,
		true,
	)
}

func TestGraphCommandPackageFilterAndNodeLimit(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--depth", "5",
		"--package", "github.com/tepzxl/codemap/examples/layered-service/internal/service",
		"--node-limit", "1",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("graph command output is not graph json: %v\n%s", err, stdout.String())
	}
	if len(output.Nodes) != 1 {
		t.Fatalf("node count = %d, want 1: %#v", len(output.Nodes), output.Nodes)
	}
	if output.Nodes[0].Package != "github.com/tepzxl/codemap/examples/layered-service/internal/service" {
		t.Fatalf("package filter did not apply: %#v", output.Nodes)
	}
	requireGraphWarning(t, output, "node-limit-exceeded")
}

func TestGraphCommandShowExternal(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--depth", "5",
		"--show-external",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("graph command output is not graph json: %v\n%s", err, stdout.String())
	}
	requireGraphEdge(t, output,
		"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
		"errors.New",
		graphmodel.EdgeResolutionExternal,
		false,
	)
}

func TestGraphCommandDirection(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
		"--depth", "2",
		"--direction", "upstream",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("graph command output is not graph json: %v\n%s", err, stdout.String())
	}
	requireGraphEdge(t, output,
		"github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
		"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
		graphmodel.EdgeResolutionResolved,
		false,
	)
	forbidGraphEdge(t, output,
		"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
		"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
	)
}

func TestGraphCommandInvalidDirection(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--direction", "sideways",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected graph command to fail for invalid direction")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected stdout to be empty on error, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "direction must be one of") {
		t.Fatalf("expected direction validation error, got %q", stderr.String())
	}
}

func TestGraphCommandUnknownEntry(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "not.exists",
		"--depth", "5",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected graph command to fail for unknown entry")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected stdout to be empty on error, got %s", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatal("expected stderr to contain error")
	}
}

func requireGraphEdge(t *testing.T, output graphmodel.Graph, from string, to string, resolution graphmodel.EdgeResolution, candidate bool) {
	t.Helper()

	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to && edge.Resolution == resolution && edge.Candidate == candidate {
			return
		}
	}
	t.Fatalf("missing graph edge from %q to %q resolution %q candidate %t in %#v", from, to, resolution, candidate, output.Edges)
}

func forbidGraphEdge(t *testing.T, output graphmodel.Graph, from string, to string) {
	t.Helper()

	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to {
			t.Fatalf("unexpected graph edge from %q to %q in %#v", from, to, output.Edges)
		}
	}
}

func requireGraphWarning(t *testing.T, output graphmodel.Graph, code string) {
	t.Helper()

	for _, warning := range output.Warnings {
		if warning.Code == code {
			return
		}
	}
	t.Fatalf("missing graph warning %q in %#v", code, output.Warnings)
}
