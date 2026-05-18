package graph

import (
	"strings"
	"testing"
)

func TestGraphTraversal(t *testing.T) {
	t.Run("simple fixture expands from main", func(t *testing.T) {
		got, err := BuildGraph(simpleSymbols(), simpleCalls(), BuildOptions{
			Entry: "github.com/tepzxl/codemap/examples/simple.main",
			Depth: 5,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}

		requireNode(t, got, "github.com/tepzxl/codemap/examples/simple.main")
		requireNode(t, got, "github.com/tepzxl/codemap/examples/simple/app.Run")
		requireNode(t, got, "github.com/tepzxl/codemap/examples/simple/service.CreateUser")
		requireEdge(t, got, "github.com/tepzxl/codemap/examples/simple.main", "github.com/tepzxl/codemap/examples/simple/app.Run", EdgeResolutionResolved)
		if err := got.Validate(); err != nil {
			t.Fatalf("graph schema validation failed: %v", err)
		}
	})

	t.Run("layered service expands to repository", func(t *testing.T) {
		got, err := BuildGraph(layeredSymbols(), layeredCalls(), BuildOptions{
			Entry: "main.main",
			Depth: 5,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}

		requireNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main")
		requireNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser")
		requireNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser")
		requireNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save")
		requireEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
			EdgeResolutionResolved,
		)
	})

	t.Run("depth limits expansion", func(t *testing.T) {
		got, err := BuildGraph(layeredSymbols(), layeredCalls(), BuildOptions{
			Entry: "main.main",
			Depth: 1,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}

		requireNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main")
		requireNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser")
		forbidNode(t, got, "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser")
	})

	t.Run("depth zero returns only entry", func(t *testing.T) {
		got, err := BuildGraph(layeredSymbols(), layeredCalls(), BuildOptions{
			Entry: "main.main",
			Depth: 0,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		if len(got.Nodes) != 1 {
			t.Fatalf("depth 0 node count = %d, want 1", len(got.Nodes))
		}
		if len(got.Edges) != 0 {
			t.Fatalf("depth 0 edge count = %d, want 0", len(got.Edges))
		}
	})

	t.Run("external is hidden by default and visible with flag", func(t *testing.T) {
		hidden, err := BuildGraph(externalSymbols(), externalCalls(), BuildOptions{
			Entry: "example.com/app.main",
			Depth: 2,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		for _, edge := range hidden.Edges {
			if edge.Resolution == EdgeResolutionExternal {
				t.Fatalf("external edge should be hidden by default: %#v", edge)
			}
		}

		visible, err := BuildGraph(externalSymbols(), externalCalls(), BuildOptions{
			Entry:        "example.com/app.main",
			Depth:        2,
			ShowExternal: true,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		requireNode(t, visible, "fmt.Println")
		requireEdge(t, visible, "example.com/app.main", "fmt.Println", EdgeResolutionExternal)
		if err := visible.Validate(); err != nil {
			t.Fatalf("graph schema validation failed: %v", err)
		}
	})

	t.Run("interface edge is controlled by flag", func(t *testing.T) {
		hidden, err := BuildGraph(interfaceSymbols(), interfaceCalls(), BuildOptions{
			Entry: "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			Depth: 2,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		forbidNode(t, hidden, "github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save")

		visible, err := BuildGraph(interfaceSymbols(), interfaceCalls(), BuildOptions{
			Entry:         "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			Depth:         2,
			ShowInterface: true,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		requireNode(t, visible, "github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save")
		requireEdge(t, visible,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save",
			EdgeResolutionInterface,
		)
	})

	t.Run("interface candidates are visible only when expansion is enabled", func(t *testing.T) {
		calls := append(interfaceCalls(), Call{
			From:       "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			To:         "github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
			Kind:       "call",
			Resolution: EdgeResolutionInterface,
			Candidate:  true,
			Callsite:   Callsite{File: "service/user.go", Line: 21, Column: 20},
		})
		symbols := append(interfaceSymbols(), Symbol{
			ID:        "github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
			Label:     "MemoryUserRepository.Save",
			Kind:      "method",
			Package:   "github.com/tepzxl/codemap/examples/interface-call/repository",
			Receiver:  "*MemoryUserRepository",
			File:      "repository/repo.go",
			StartLine: 9,
			EndLine:   12,
		})

		visibleWithoutCandidates, err := BuildGraph(symbols, calls, BuildOptions{
			Entry:         "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			Depth:         2,
			ShowInterface: true,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		requireEdge(t, visibleWithoutCandidates,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save",
			EdgeResolutionInterface,
		)
		forbidEdge(t, visibleWithoutCandidates,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
		)

		expanded, err := BuildGraph(symbols, calls, BuildOptions{
			Entry:           "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			Depth:           2,
			ExpandInterface: true,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		edge := requireEdge(t, expanded,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
			EdgeResolutionInterface,
		)
		if !edge.Candidate {
			t.Fatalf("expected expanded implementation edge to be marked candidate: %#v", edge)
		}
	})

	t.Run("cycle does not recurse forever", func(t *testing.T) {
		got, err := BuildGraph(cycleSymbols(), cycleCalls(), BuildOptions{
			Entry: "example.com/cycle.A",
			Depth: 20,
		})
		if err != nil {
			t.Fatalf("BuildGraph returned error: %v", err)
		}
		if len(got.Nodes) != 2 {
			t.Fatalf("cycle graph node count = %d, want 2", len(got.Nodes))
		}
		if len(got.Edges) != 2 {
			t.Fatalf("cycle graph edge count = %d, want 2", len(got.Edges))
		}
	})

	t.Run("unknown entry returns clear error", func(t *testing.T) {
		_, err := BuildGraph(simpleSymbols(), simpleCalls(), BuildOptions{
			Entry: "not.exists",
			Depth: 1,
		})
		if err == nil {
			t.Fatal("expected unknown entry to return error")
		}
		if !strings.Contains(err.Error(), "entry symbol not found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func simpleSymbols() []Symbol {
	return []Symbol{
		{ID: "github.com/tepzxl/codemap/examples/simple.main", Label: "main", Kind: "function", Package: "github.com/tepzxl/codemap/examples/simple", File: "main.go", StartLine: 5, EndLine: 7},
		{ID: "github.com/tepzxl/codemap/examples/simple/app.Run", Label: "Run", Kind: "function", Package: "github.com/tepzxl/codemap/examples/simple/app", File: "app/app.go", StartLine: 5, EndLine: 7},
		{ID: "github.com/tepzxl/codemap/examples/simple/service.CreateUser", Label: "CreateUser", Kind: "function", Package: "github.com/tepzxl/codemap/examples/simple/service", File: "service/user.go", StartLine: 3, EndLine: 8},
	}
}

func simpleCalls() []Call {
	return []Call{
		{From: "github.com/tepzxl/codemap/examples/simple.main", To: "github.com/tepzxl/codemap/examples/simple/app.Run", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "main.go", Line: 6, Column: 9}},
		{From: "github.com/tepzxl/codemap/examples/simple/app.Run", To: "github.com/tepzxl/codemap/examples/simple/service.CreateUser", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "app/app.go", Line: 6, Column: 20}},
	}
}

func layeredSymbols() []Symbol {
	return []Symbol{
		{ID: "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main", Label: "main", Kind: "function", Package: "github.com/tepzxl/codemap/examples/layered-service/cmd/api", File: "cmd/api/main.go", StartLine: 5, EndLine: 8},
		{ID: "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", Label: "UserHandler.CreateUser", Kind: "method", Package: "github.com/tepzxl/codemap/examples/layered-service/internal/handler", Receiver: "*UserHandler", File: "internal/handler/user.go", StartLine: 15, EndLine: 17},
		{ID: "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser", Label: "UserService.CreateUser", Kind: "method", Package: "github.com/tepzxl/codemap/examples/layered-service/internal/service", Receiver: "*UserService", File: "internal/service/user.go", StartLine: 15, EndLine: 17},
		{ID: "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save", Label: "UserRepository.Save", Kind: "method", Package: "github.com/tepzxl/codemap/examples/layered-service/internal/repository", Receiver: "*UserRepository", File: "internal/repository/user.go", StartLine: 11, EndLine: 16},
	}
}

func layeredCalls() []Call {
	return []Call{
		{From: "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main", To: "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "cmd/api/main.go", Line: 7, Column: 18}},
		{From: "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", To: "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "internal/handler/user.go", Line: 16, Column: 29}},
		{From: "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser", To: "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "internal/service/user.go", Line: 16, Column: 20}},
	}
}

func externalSymbols() []Symbol {
	return []Symbol{{ID: "example.com/app.main", Label: "main", Kind: "function", Package: "example.com/app", File: "main.go", StartLine: 1, EndLine: 5}}
}

func externalCalls() []Call {
	return []Call{{From: "example.com/app.main", To: "fmt.Println", Kind: "call", Resolution: EdgeResolutionExternal, Callsite: Callsite{File: "main.go", Line: 3, Column: 13}}}
}

func interfaceSymbols() []Symbol {
	return []Symbol{{ID: "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser", Label: "UserService.CreateUser", Kind: "method", Package: "github.com/tepzxl/codemap/examples/interface-call/service", Receiver: "*UserService", File: "service/user.go", StartLine: 17, EndLine: 22}}
}

func interfaceCalls() []Call {
	return []Call{{From: "github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser", To: "github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save", Kind: "call", Resolution: EdgeResolutionInterface, Callsite: Callsite{File: "service/user.go", Line: 21, Column: 20}}}
}

func cycleSymbols() []Symbol {
	return []Symbol{
		{ID: "example.com/cycle.A", Label: "A", Kind: "function", Package: "example.com/cycle", File: "cycle.go", StartLine: 1, EndLine: 3},
		{ID: "example.com/cycle.B", Label: "B", Kind: "function", Package: "example.com/cycle", File: "cycle.go", StartLine: 5, EndLine: 7},
	}
}

func cycleCalls() []Call {
	return []Call{
		{From: "example.com/cycle.A", To: "example.com/cycle.B", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "cycle.go", Line: 2, Column: 3}},
		{From: "example.com/cycle.B", To: "example.com/cycle.A", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "cycle.go", Line: 6, Column: 3}},
	}
}

func requireNode(t *testing.T, graph Graph, id string) {
	t.Helper()
	for _, node := range graph.Nodes {
		if node.ID == id {
			return
		}
	}
	t.Fatalf("missing node %q in %#v", id, graph.Nodes)
}

func forbidNode(t *testing.T, graph Graph, id string) {
	t.Helper()
	for _, node := range graph.Nodes {
		if node.ID == id {
			t.Fatalf("unexpected node %q in %#v", id, graph.Nodes)
		}
	}
}

func requireEdge(t *testing.T, graph Graph, from string, to string, resolution EdgeResolution) Edge {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to && edge.Resolution == resolution {
			return edge
		}
	}
	t.Fatalf("missing edge %q -> %q (%s) in %#v", from, to, resolution, graph.Edges)
	return Edge{}
}

func forbidEdge(t *testing.T, graph Graph, from string, to string) {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to {
			t.Fatalf("unexpected edge %q -> %q in %#v", from, to, graph.Edges)
		}
	}
}
