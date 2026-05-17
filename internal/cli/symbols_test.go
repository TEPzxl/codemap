package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSymbolsCommand(t *testing.T) {
	repoRoot := findRepoRoot(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"symbols", filepath.Join(repoRoot, "examples", "simple")}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("symbols command exit code = %d, stderr = %s", code, stderr.String())
	}

	var output struct {
		Root    string `json:"root"`
		Symbols []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Label     string `json:"label"`
			Kind      string `json:"kind"`
			Package   string `json:"package"`
			File      string `json:"file"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		} `json:"symbols"`
		Warnings []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"warnings"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("symbols command output is not json: %v\n%s", err, stdout.String())
	}
	if len(output.Symbols) == 0 {
		t.Fatal("expected symbols command to return symbols")
	}

	foundMain := false
	for _, symbol := range output.Symbols {
		if symbol.ID == "github.com/tepzxl/codemap/examples/simple.main" {
			foundMain = true
		}
		if symbol.ID == "" || symbol.Name == "" || symbol.Label == "" || symbol.Kind == "" || symbol.Package == "" || symbol.File == "" {
			t.Fatalf("symbol has missing required fields: %#v", symbol)
		}
		if symbol.StartLine <= 0 || symbol.EndLine < symbol.StartLine {
			t.Fatalf("symbol has invalid line range: %#v", symbol)
		}
	}
	if !foundMain {
		t.Fatal("expected symbols command to include main.main")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	dir := wd
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("cannot find repo root from %q", wd)
		}
		dir = parent
	}
}
