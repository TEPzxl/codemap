package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/tepzxl/codemap/internal/graph"
)

func TestCallsCommand(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"calls", filepath.Join(repoRoot, "examples", "simple")}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("calls command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output struct {
		Root  string `json:"root"`
		Calls []struct {
			From       string               `json:"from"`
			To         string               `json:"to"`
			Kind       string               `json:"kind"`
			Resolution graph.EdgeResolution `json:"resolution"`
			Callsite   graph.Callsite       `json:"callsite"`
		} `json:"calls"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("calls command output is not json: %v\n%s", err, stdout.String())
	}
	if len(output.Calls) == 0 {
		t.Fatal("expected calls command to return calls")
	}

	foundResolved := false
	for _, call := range output.Calls {
		if call.From == "" || call.To == "" || call.Kind != "call" {
			t.Fatalf("call has invalid required fields: %#v", call)
		}
		if call.Callsite.File == "" || call.Callsite.Line <= 0 {
			t.Fatalf("call has invalid callsite: %#v", call)
		}
		if call.Resolution == graph.EdgeResolutionResolved {
			foundResolved = true
		}
	}
	if !foundResolved {
		t.Fatal("expected at least one resolved call")
	}
}
