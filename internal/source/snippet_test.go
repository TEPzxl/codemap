package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSnippet(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "service", "user.go"), strings.Join([]string{
		"package service",
		"",
		"func CreateUser(name string) string {",
		"\treturn name",
		"}",
	}, "\n"))

	got, err := ReadSnippet(root, Location{
		NodeID:    "example.com/app/service.CreateUser",
		File:      "service/user.go",
		StartLine: 3,
		EndLine:   5,
	})
	if err != nil {
		t.Fatalf("ReadSnippet returned error: %v", err)
	}

	if got.NodeID != "example.com/app/service.CreateUser" {
		t.Fatalf("node id mismatch: got %q", got.NodeID)
	}
	if got.File != "service/user.go" {
		t.Fatalf("file mismatch: got %q", got.File)
	}
	if got.StartLine != 3 || got.EndLine != 5 {
		t.Fatalf("line range mismatch: got %d-%d", got.StartLine, got.EndLine)
	}
	if !strings.Contains(got.Source, "func CreateUser") {
		t.Fatalf("snippet missing function body: %q", got.Source)
	}
	if strings.Contains(got.Source, "package service") {
		t.Fatalf("snippet should not include lines before start_line: %q", got.Source)
	}
}

func TestReadSnippetRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "safe.go"), "package main\n")

	_, err := ReadSnippet(root, Location{
		NodeID:    "example.com/app.main",
		File:      "../outside.go",
		StartLine: 1,
		EndLine:   1,
	})
	if err == nil {
		t.Fatal("expected path traversal to be rejected")
	}
}

func TestReadSnippetRejectsSymlinkOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeTestFile(t, filepath.Join(outside, "secret.go"), "package secret\n")

	err := os.Symlink(filepath.Join(outside, "secret.go"), filepath.Join(root, "link.go"))
	if err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err = ReadSnippet(root, Location{
		NodeID:    "example.com/app.main",
		File:      "link.go",
		StartLine: 1,
		EndLine:   1,
	})
	if err == nil {
		t.Fatal("expected symlink outside root to be rejected")
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %q: %v", path, err)
	}
}
