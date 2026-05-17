package analyzer

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/tepzxl/codemap/internal/graph"
)

func TestExtractCalls(t *testing.T) {
	repoRoot := findRepoRoot(t)

	t.Run("simple fixture includes resolved function calls", func(t *testing.T) {
		calls := loadSymbolsAndCalls(t, filepath.Join(repoRoot, "examples", "simple"))

		requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/simple.main",
			"github.com/tepzxl/codemap/examples/simple/app.Run",
			graph.EdgeResolutionResolved,
		)
		requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/simple/app.Run",
			"github.com/tepzxl/codemap/examples/simple/service.CreateUser",
			graph.EdgeResolutionResolved,
		)
		assertCallsites(t, calls)
	})

	t.Run("layered service includes resolved cross package method calls", func(t *testing.T) {
		calls := loadSymbolsAndCalls(t, filepath.Join(repoRoot, "examples", "layered-service"))

		requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/layered-service/cmd/api.main",
			"github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
			graph.EdgeResolutionResolved,
		)
		requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			graph.EdgeResolutionResolved,
		)
		requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
			graph.EdgeResolutionResolved,
		)
		assertCallsites(t, calls)
	})

	t.Run("interface fixture marks interface method calls conservatively", func(t *testing.T) {
		calls := loadSymbolsAndCalls(t, filepath.Join(repoRoot, "examples", "interface-call"))

		call := requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save",
			graph.EdgeResolutionInterface,
		)
		if strings.Contains(call.To, "MemoryUserRepository") {
			t.Fatalf("interface call must not be resolved to concrete implementation, got %q", call.To)
		}
	})

	t.Run("external unresolved and test file calls", func(t *testing.T) {
		moduleDir := t.TempDir()
		writeFile(t, filepath.Join(moduleDir, "go.mod"), "module example.com/callfixture\n\ngo 1.25.0\n")
		writeFile(t, filepath.Join(moduleDir, "main.go"), `package main

import "fmt"

func main() {
	helper("alice")
	fmt.Println("hello")
	var dynamic func()
	dynamic()
}

func helper(name string) string {
	return name
}
`)
		writeFile(t, filepath.Join(moduleDir, "main_test.go"), `package main

func TestOnly() {
	helper("test")
}
`)

		calls := loadSymbolsAndCalls(t, moduleDir)

		requireCall(t, calls,
			"example.com/callfixture.main",
			"example.com/callfixture.helper",
			graph.EdgeResolutionResolved,
		)
		requireResolution(t, calls, graph.EdgeResolutionExternal)
		requireResolution(t, calls, graph.EdgeResolutionUnresolved)

		for _, call := range calls {
			if strings.HasSuffix(call.Callsite.File, "_test.go") {
				t.Fatalf("expected _test.go call to be excluded, got %#v", call)
			}
		}
	})
}

func loadSymbolsAndCalls(t *testing.T, rootPath string) []Call {
	t.Helper()

	result, err := LoadPackages(LoadRequest{RootPath: rootPath})
	if err != nil {
		t.Fatalf("LoadPackages returned error: %v", err)
	}

	symbols, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols returned error: %v", err)
	}

	calls, err := ExtractCalls(result, symbols)
	if err != nil {
		t.Fatalf("ExtractCalls returned error: %v", err)
	}
	return calls
}

func requireCall(t *testing.T, calls []Call, from string, to string, resolution graph.EdgeResolution) Call {
	t.Helper()

	for _, call := range calls {
		if call.From == from && call.To == to && call.Resolution == resolution {
			return call
		}
	}
	t.Fatalf("missing call from %q to %q with resolution %q in %#v", from, to, resolution, calls)
	return Call{}
}

func requireResolution(t *testing.T, calls []Call, resolution graph.EdgeResolution) {
	t.Helper()

	for _, call := range calls {
		if call.Resolution == resolution {
			return
		}
	}
	t.Fatalf("missing call with resolution %q in %#v", resolution, calls)
}

func assertCallsites(t *testing.T, calls []Call) {
	t.Helper()

	for _, call := range calls {
		if call.Kind != "call" {
			t.Fatalf("call kind mismatch: got %q want call", call.Kind)
		}
		if call.From == "" {
			t.Fatalf("call from is empty: %#v", call)
		}
		if call.To == "" {
			t.Fatalf("call to is empty: %#v", call)
		}
		if call.Callsite.File == "" {
			t.Fatalf("callsite file is empty: %#v", call)
		}
		if call.Callsite.Line <= 0 {
			t.Fatalf("callsite line must be positive: %#v", call)
		}
		if call.Callsite.Column <= 0 {
			t.Fatalf("callsite column must be positive: %#v", call)
		}
	}
}
