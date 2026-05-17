package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPackages(t *testing.T) {
	repoRoot := findRepoRoot(t)
	simplePath := filepath.Join(repoRoot, "examples", "simple")

	got, err := LoadPackages(LoadRequest{
		RootPath: simplePath,
	})
	if err != nil {
		t.Fatalf("LoadPackages returned error: %v", err)
	}
	if len(got.Packages) == 0 {
		t.Fatal("expected at least one loaded package")
	}

	for _, pkg := range got.Packages {
		for _, file := range pkg.Files {
			if strings.HasSuffix(file, "_test.go") {
				t.Fatalf("expected _test.go files to be excluded, got %q", file)
			}
		}
	}
}

func TestLoadPackagesFixturesCanScan(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tests := []struct {
		name string
		path string
	}{
		{name: "simple", path: filepath.Join(repoRoot, "examples", "simple")},
		{name: "layered service", path: filepath.Join(repoRoot, "examples", "layered-service")},
		{name: "interface call", path: filepath.Join(repoRoot, "examples", "interface-call")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadPackages(LoadRequest{
				RootPath: tt.path,
			})
			if err != nil {
				t.Fatalf("LoadPackages returned error: %v", err)
			}
			if len(got.Packages) == 0 {
				t.Fatal("expected at least one loaded package")
			}
			for _, pkg := range got.Packages {
				for _, file := range pkg.Files {
					if strings.HasSuffix(file, "_test.go") {
						t.Fatalf("expected _test.go files to be excluded, got %q", file)
					}
				}
			}
		})
	}
}

func TestLoadPackagesDefaultExcludesTests(t *testing.T) {
	moduleDir := t.TempDir()
	writeFile(t, filepath.Join(moduleDir, "go.mod"), "module example.com/tmpmod\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(moduleDir, "main.go"), "package main\n\nfunc main() {}\n")
	writeFile(t, filepath.Join(moduleDir, "main_test.go"), "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n")

	got, err := LoadPackages(LoadRequest{
		RootPath: moduleDir,
	})
	if err != nil {
		t.Fatalf("LoadPackages returned error: %v", err)
	}

	for _, pkg := range got.Packages {
		for _, file := range pkg.Files {
			if strings.HasSuffix(file, "_test.go") {
				t.Fatalf("expected _test.go files to be excluded, got %q", file)
			}
		}
	}
}

func TestLoadPackagesInvalidPath(t *testing.T) {
	_, err := LoadPackages(LoadRequest{
		RootPath: filepath.Join(t.TempDir(), "not-exists"),
	})
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
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

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %q: %v", path, err)
	}
}
