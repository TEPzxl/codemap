package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func TestPathCommand(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"path",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--from", "main.main",
		"--to", "UserRepository.Save",
		"--max-depth", "8",
		"--limit", "5",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("path command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.PathResult
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("path command output is not path json: %v\n%s", err, stdout.String())
	}
	if len(output.Paths) != 1 {
		t.Fatalf("path count = %d, want 1: %#v", len(output.Paths), output.Paths)
	}
	wantLast := "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save"
	gotNodes := output.Paths[0].Nodes
	if gotNodes[len(gotNodes)-1] != wantLast {
		t.Fatalf("path last node = %q, want %q: %#v", gotNodes[len(gotNodes)-1], wantLast, gotNodes)
	}
	if len(output.Graph.Nodes) == 0 || len(output.Graph.Edges) == 0 {
		t.Fatalf("expected path graph, got %#v", output.Graph)
	}
}

func TestPathCommandNoPath(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"path",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--from", "main.main",
		"--to", "UserRepository.Save",
		"--max-depth", "1",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("path command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output graphmodel.PathResult
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("path command output is not path json: %v\n%s", err, stdout.String())
	}
	if len(output.Paths) != 0 {
		t.Fatalf("expected no paths, got %#v", output.Paths)
	}
	if len(output.Warnings) == 0 || output.Warnings[0].Code != "path-not-found" {
		t.Fatalf("expected path-not-found warning, got %#v", output.Warnings)
	}
}

func TestPathCommandInvalidArgs(t *testing.T) {
	repoRoot := findRepoRoot(t)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "unknown from",
			args: []string{"path", filepath.Join(repoRoot, "examples", "layered-service"), "--from", "not.exists", "--to", "UserRepository.Save"},
			want: "from symbol",
		},
		{
			name: "invalid max depth",
			args: []string{"path", filepath.Join(repoRoot, "examples", "layered-service"), "--from", "main.main", "--to", "UserRepository.Save", "--max-depth", "-1"},
			want: "max_depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code == 0 {
				t.Fatal("expected path command to fail")
			}
			if stdout.Len() != 0 {
				t.Fatalf("expected stdout to be empty on error, got %s", stdout.String())
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("stderr = %q, want substring %q", stderr.String(), tt.want)
			}
		})
	}
}
