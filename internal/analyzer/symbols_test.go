package analyzer

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSymbols(t *testing.T) {
	repoRoot := findRepoRoot(t)

	t.Run("examples simple includes plain functions", func(t *testing.T) {
		result, err := LoadPackages(LoadRequest{
			RootPath: filepath.Join(repoRoot, "examples", "simple"),
		})
		if err != nil {
			t.Fatalf("LoadPackages returned error: %v", err)
		}

		symbols, err := ExtractSymbols(result)
		if err != nil {
			t.Fatalf("ExtractSymbols returned error: %v", err)
		}

		requireSymbol(t, symbols, "github.com/tepzxl/codemap/examples/simple.main")
		requireSymbol(t, symbols, "github.com/tepzxl/codemap/examples/simple/app.Run")
		requireSymbol(t, symbols, "github.com/tepzxl/codemap/examples/simple/service.CreateUser")
		assertUniqueSymbolIDs(t, symbols)
		assertSymbolLocations(t, symbols)
	})

	t.Run("examples layered service includes pointer receiver methods", func(t *testing.T) {
		result, err := LoadPackages(LoadRequest{
			RootPath: filepath.Join(repoRoot, "examples", "layered-service"),
		})
		if err != nil {
			t.Fatalf("LoadPackages returned error: %v", err)
		}

		symbols, err := ExtractSymbols(result)
		if err != nil {
			t.Fatalf("ExtractSymbols returned error: %v", err)
		}

		handler := requireSymbol(t, symbols, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser")
		if handler.Receiver != "*UserHandler" {
			t.Fatalf("handler receiver mismatch: got %q want %q", handler.Receiver, "*UserHandler")
		}
		if handler.Kind != "method" {
			t.Fatalf("handler kind mismatch: got %q want method", handler.Kind)
		}

		requireSymbol(t, symbols, "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser")
		requireSymbol(t, symbols, "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save")
		assertUniqueSymbolIDs(t, symbols)
		assertSymbolLocations(t, symbols)
	})

	t.Run("value receiver methods and test files", func(t *testing.T) {
		moduleDir := t.TempDir()
		writeFile(t, filepath.Join(moduleDir, "go.mod"), "module example.com/symbolfixture\n\ngo 1.25.0\n")
		writeFile(t, filepath.Join(moduleDir, "thing.go"), `package symbolfixture

type Thing struct{}

func BuildThing() Thing {
	return Thing{}
}

func (t Thing) Name() string {
	return "thing"
}
`)
		writeFile(t, filepath.Join(moduleDir, "thing_test.go"), `package symbolfixture

func TestOnlyHelper() {}
`)

		result, err := LoadPackages(LoadRequest{
			RootPath: moduleDir,
		})
		if err != nil {
			t.Fatalf("LoadPackages returned error: %v", err)
		}

		symbols, err := ExtractSymbols(result)
		if err != nil {
			t.Fatalf("ExtractSymbols returned error: %v", err)
		}

		valueMethod := requireSymbol(t, symbols, "example.com/symbolfixture.Thing.Name")
		if valueMethod.Receiver != "Thing" {
			t.Fatalf("value receiver mismatch: got %q want Thing", valueMethod.Receiver)
		}
		requireSymbol(t, symbols, "example.com/symbolfixture.BuildThing")

		for _, symbol := range symbols {
			if strings.Contains(symbol.ID, "TestOnlyHelper") {
				t.Fatalf("expected _test.go symbol to be excluded, got %q", symbol.ID)
			}
			if strings.HasSuffix(symbol.File, "_test.go") {
				t.Fatalf("expected _test.go file to be excluded, got %q", symbol.File)
			}
		}
	})
}

func requireSymbol(t *testing.T, symbols []Symbol, id string) Symbol {
	t.Helper()

	for _, symbol := range symbols {
		if symbol.ID == id {
			return symbol
		}
	}
	t.Fatalf("missing symbol %q in %#v", id, symbols)
	return Symbol{}
}

func assertUniqueSymbolIDs(t *testing.T, symbols []Symbol) {
	t.Helper()

	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		if _, ok := seen[symbol.ID]; ok {
			t.Fatalf("duplicate symbol id %q", symbol.ID)
		}
		seen[symbol.ID] = struct{}{}
	}
}

func assertSymbolLocations(t *testing.T, symbols []Symbol) {
	t.Helper()

	for _, symbol := range symbols {
		if symbol.ID == "" {
			t.Fatal("symbol id is empty")
		}
		if symbol.Label == "" {
			t.Fatalf("symbol %q label is empty", symbol.ID)
		}
		if symbol.Kind == "" {
			t.Fatalf("symbol %q kind is empty", symbol.ID)
		}
		if symbol.Package == "" {
			t.Fatalf("symbol %q package is empty", symbol.ID)
		}
		if symbol.File == "" {
			t.Fatalf("symbol %q file is empty", symbol.ID)
		}
		if symbol.StartLine <= 0 {
			t.Fatalf("symbol %q start line must be positive, got %d", symbol.ID, symbol.StartLine)
		}
		if symbol.EndLine < symbol.StartLine {
			t.Fatalf("symbol %q end line %d before start line %d", symbol.ID, symbol.EndLine, symbol.StartLine)
		}
	}
}
