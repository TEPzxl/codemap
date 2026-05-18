package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func TestPackagesCommand(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"packages",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "main.main",
		"--depth", "5",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("packages command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.PackageGraph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("packages command output is not package graph json: %v\n%s", err, stdout.String())
	}
	requirePackageGraphNode(t, output, "internal/service")
	requirePackageGraphEdge(t, output, "internal/handler", "internal/service", 2)
	requirePackageGraphEdge(t, output, "internal/service", "internal/repository", 2)
}

func TestPackagesCommandNoEntryUsesWholeProject(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"packages",
		filepath.Join(repoRoot, "examples", "layered-service"),
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("packages command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.PackageGraph
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("packages command output is not package graph json: %v\n%s", err, stdout.String())
	}
	requirePackageGraphEdge(t, output, "cmd/api", "internal/handler", 2)
	requirePackageGraphEdge(t, output, "internal/handler", "internal/service", 2)
	requirePackageGraphEdge(t, output, "internal/service", "internal/repository", 2)
}

func TestPackagesCommandInvalidEntry(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"packages",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--entry", "not.exists",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected packages command to fail for unknown entry")
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected stdout to be empty on error, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "entry symbol not found") {
		t.Fatalf("expected unknown entry error, got %q", stderr.String())
	}
}

func requirePackageGraphNode(t *testing.T, output graphmodel.PackageGraph, id string) {
	t.Helper()
	for _, node := range output.Nodes {
		if node.ID == id {
			return
		}
	}
	t.Fatalf("missing package node %q in %#v", id, output.Nodes)
}

func requirePackageGraphEdge(t *testing.T, output graphmodel.PackageGraph, from string, to string, calls int) {
	t.Helper()
	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to {
			if edge.Calls != calls {
				t.Fatalf("package edge %q -> %q calls = %d, want %d", from, to, edge.Calls, calls)
			}
			return
		}
	}
	t.Fatalf("missing package edge %q -> %q in %#v", from, to, output.Edges)
}
