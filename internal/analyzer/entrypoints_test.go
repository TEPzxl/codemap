package analyzer

import (
	"path/filepath"
	"testing"
)

func TestDiscoverEntrypoints(t *testing.T) {
	t.Run("examples include main first", func(t *testing.T) {
		repoRoot := findRepoRoot(t)
		entrypoints := loadEntrypoints(t, filepath.Join(repoRoot, "examples", "layered-service"))

		if len(entrypoints) == 0 {
			t.Fatal("expected entrypoints")
		}
		if entrypoints[0].ID != "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main" {
			t.Fatalf("first entrypoint = %q, want layered-service main", entrypoints[0].ID)
		}
		requireEntrypointReason(t, entrypoints, "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main", "main-function")
	})

	t.Run("layered service includes exported handler and service methods", func(t *testing.T) {
		repoRoot := findRepoRoot(t)
		entrypoints := loadEntrypoints(t, filepath.Join(repoRoot, "examples", "layered-service"))

		requireEntrypointReason(t, entrypoints, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", "exported-method")
		requireEntrypointReason(t, entrypoints, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", "receiver:Handler")
		requireEntrypointReason(t, entrypoints, "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser", "exported-method")
	})

	t.Run("recognizes common starters handlers and goroutine starters", func(t *testing.T) {
		moduleDir := t.TempDir()
		writeFile(t, filepath.Join(moduleDir, "go.mod"), "module example.com/entryfixture\n\ngo 1.25.0\n")
		writeFile(t, filepath.Join(moduleDir, "main.go"), `package main

type Server struct{}

func (s *Server) Run() {}

func Start() {}

func Serve() {}

type APIHandler struct{}

func (h *APIHandler) HandleUser() {}

func Worker() {
	go Start()
}

func helper() {}
`)
		writeFile(t, filepath.Join(moduleDir, "main_test.go"), `package main

func TestMain() {}
func TestOnly() {}
`)

		entrypoints := loadEntrypoints(t, moduleDir)

		requireEntrypointReason(t, entrypoints, "example.com/entryfixture.(*Server).Run", "name:Run")
		requireEntrypointReason(t, entrypoints, "example.com/entryfixture.Start", "name:Start")
		requireEntrypointReason(t, entrypoints, "example.com/entryfixture.Serve", "name:Serve")
		requireEntrypointReason(t, entrypoints, "example.com/entryfixture.(*APIHandler).HandleUser", "name:Handle")
		requireEntrypointReason(t, entrypoints, "example.com/entryfixture.(*APIHandler).HandleUser", "receiver:Handler")
		requireEntrypointReason(t, entrypoints, "example.com/entryfixture.Worker", "contains-goroutine")
		forbidEntrypoint(t, entrypoints, "example.com/entryfixture.TestOnly")
	})
}

func loadEntrypoints(t *testing.T, rootPath string) []Entrypoint {
	t.Helper()

	result, err := LoadPackages(LoadRequest{RootPath: rootPath})
	if err != nil {
		t.Fatalf("LoadPackages returned error: %v", err)
	}
	symbols, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols returned error: %v", err)
	}
	entrypoints, err := DiscoverEntrypoints(result, symbols)
	if err != nil {
		t.Fatalf("DiscoverEntrypoints returned error: %v", err)
	}
	return entrypoints
}

func requireEntrypointReason(t *testing.T, entrypoints []Entrypoint, id string, reason string) {
	t.Helper()
	for _, entrypoint := range entrypoints {
		if entrypoint.ID != id {
			continue
		}
		if entrypoint.Label == "" || entrypoint.Package == "" || entrypoint.File == "" || entrypoint.Kind == "" {
			t.Fatalf("entrypoint has missing required fields: %#v", entrypoint)
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

func forbidEntrypoint(t *testing.T, entrypoints []Entrypoint, id string) {
	t.Helper()
	for _, entrypoint := range entrypoints {
		if entrypoint.ID == id {
			t.Fatalf("unexpected entrypoint %q in %#v", id, entrypoints)
		}
	}
}
