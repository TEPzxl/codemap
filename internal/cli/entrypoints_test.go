package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/tepzxl/codemap/internal/analyzer"
)

func TestEntrypointsCommand(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"entrypoints", filepath.Join(repoRoot, "examples", "layered-service")}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("entrypoints command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output struct {
		Entrypoints []analyzer.Entrypoint `json:"entrypoints"`
		Note        string                `json:"note"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("entrypoints command output is not json: %v\n%s", err, stdout.String())
	}
	if output.Note == "" {
		t.Fatal("expected heuristic note")
	}
	if len(output.Entrypoints) == 0 {
		t.Fatal("expected entrypoints")
	}
	if output.Entrypoints[0].ID != "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main" {
		t.Fatalf("first entrypoint = %q, want main", output.Entrypoints[0].ID)
	}
	requireCLIEntrypointReason(t, output.Entrypoints, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", "receiver:Handler")
}

func requireCLIEntrypointReason(t *testing.T, entrypoints []analyzer.Entrypoint, id string, reason string) {
	t.Helper()
	for _, entrypoint := range entrypoints {
		if entrypoint.ID != id {
			continue
		}
		for _, gotReason := range entrypoint.Reasons {
			if gotReason == reason {
				return
			}
		}
		t.Fatalf("entrypoint %q missing reason %q in %#v", id, reason, entrypoint.Reasons)
	}
	t.Fatalf("missing entrypoint %q in %#v", id, entrypoints)
}
