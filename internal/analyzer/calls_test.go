package analyzer

import (
	"go/token"
	"go/types"
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

func TestResolveInterfaceMethodWithNilPackageDoesNotPanic(t *testing.T) {
	method := types.NewFunc(
		token.NoPos,
		nil,
		"Save",
		types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(), false),
	)
	iface := types.NewInterfaceType([]*types.Func{method}, nil).Complete()
	recv := types.NewNamed(
		types.NewTypeName(token.NoPos, types.NewPackage("example.com/app", "app"), "Store", nil),
		iface,
		nil,
	)

	target := resolveInterfaceMethodCallTarget(method, recv, map[string]struct{}{})
	if target.Resolution != graph.EdgeResolutionUnresolved {
		t.Fatalf("resolution = %q, want %q", target.Resolution, graph.EdgeResolutionUnresolved)
	}
	if target.To != "Save" {
		t.Fatalf("to = %q, want Save", target.To)
	}
	if target.InterfaceMethod != nil || target.InterfaceType != nil {
		t.Fatalf("unresolved target should not carry interface expansion metadata: %#v", target)
	}
}

func TestExtractCallsExpandInterfaceCandidates(t *testing.T) {
	t.Run("interface fixture keeps original edge and adds pointer receiver candidate", func(t *testing.T) {
		repoRoot := findRepoRoot(t)
		calls := loadSymbolsAndCallsWithOptions(t, filepath.Join(repoRoot, "examples", "interface-call"), CallOptions{
			ExpandInterface: true,
		})

		requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save",
			graph.EdgeResolutionInterface,
		)
		candidate := requireCall(t, calls,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
			graph.EdgeResolutionInterface,
		)
		if !candidate.Candidate {
			t.Fatalf("expected concrete implementation edge to be marked candidate: %#v", candidate)
		}
	})

	t.Run("value pointer multiple and signature mismatch cases use go/types implementation checks", func(t *testing.T) {
		moduleDir := t.TempDir()
		writeFile(t, filepath.Join(moduleDir, "go.mod"), "module example.com/interfacecandidates\n\ngo 1.25.0\n")
		writeFile(t, filepath.Join(moduleDir, "main.go"), `package main

type Store interface {
	Save(name string) error
}

type ValueStore struct{}

func (ValueStore) Save(name string) error {
	_ = name
	return nil
}

type PointerStore struct{}

func (*PointerStore) Save(name string) error {
	_ = name
	return nil
}

type SecondPointerStore struct{}

func (*SecondPointerStore) Save(name string) error {
	_ = name
	return nil
}

type WrongSignatureStore struct{}

func (*WrongSignatureStore) Save(id int) error {
	_ = id
	return nil
}

type Service struct {
	store Store
}

func (s *Service) Create(name string) error {
	return s.store.Save(name)
}
`)

		calls := loadSymbolsAndCallsWithOptions(t, moduleDir, CallOptions{ExpandInterface: true})
		from := "example.com/interfacecandidates.(*Service).Create"

		requireCandidateCall(t, calls, from, "example.com/interfacecandidates.ValueStore.Save")
		requireCandidateCall(t, calls, from, "example.com/interfacecandidates.(*PointerStore).Save")
		requireCandidateCall(t, calls, from, "example.com/interfacecandidates.(*SecondPointerStore).Save")
		forbidCall(t, calls, from, "example.com/interfacecandidates.(*WrongSignatureStore).Save")
	})
}

func loadSymbolsAndCalls(t *testing.T, rootPath string) []Call {
	t.Helper()
	return loadSymbolsAndCallsWithOptions(t, rootPath, CallOptions{})
}

func loadSymbolsAndCallsWithOptions(t *testing.T, rootPath string, options CallOptions) []Call {
	t.Helper()

	result, err := LoadPackages(LoadRequest{RootPath: rootPath})
	if err != nil {
		t.Fatalf("LoadPackages returned error: %v", err)
	}

	symbols, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols returned error: %v", err)
	}

	calls, err := ExtractCallsWithOptions(result, symbols, options)
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

func requireCandidateCall(t *testing.T, calls []Call, from string, to string) Call {
	t.Helper()

	call := requireCall(t, calls, from, to, graph.EdgeResolutionInterface)
	if !call.Candidate {
		t.Fatalf("expected call to be candidate: %#v", call)
	}
	return call
}

func forbidCall(t *testing.T, calls []Call, from string, to string) {
	t.Helper()

	for _, call := range calls {
		if call.From == from && call.To == to {
			t.Fatalf("unexpected call from %q to %q in %#v", from, to, calls)
		}
	}
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
