package graph

import "testing"

func TestBuildPackageGraphLayeredService(t *testing.T) {
	got, err := BuildPackageGraph(layeredSymbols(), layeredCalls(), PackageGraphOptions{
		BuildOptions: BuildOptions{
			Entry: "main.main",
			Depth: 5,
		},
		ModulePath: "github.com/tepzxl/codemap/examples/layered-service",
	})
	if err != nil {
		t.Fatalf("BuildPackageGraph returned error: %v", err)
	}

	requirePackageNode(t, got, "cmd/api", 1, 1)
	requirePackageNode(t, got, "internal/handler", 1, 2)
	requirePackageNode(t, got, "internal/service", 1, 2)
	requirePackageNode(t, got, "internal/repository", 1, 1)
	requirePackageEdge(t, got, "cmd/api", "internal/handler", 1)
	requirePackageEdge(t, got, "internal/handler", "internal/service", 1)
	requirePackageEdge(t, got, "internal/service", "internal/repository", 1)
}

func TestBuildPackageGraphHidesSelfEdgesByDefault(t *testing.T) {
	symbols := []Symbol{
		{ID: "example.com/app.A", Label: "A", Kind: "function", Package: "example.com/app", File: "app.go", StartLine: 1, EndLine: 2},
		{ID: "example.com/app.B", Label: "B", Kind: "function", Package: "example.com/app", File: "app.go", StartLine: 4, EndLine: 5},
	}
	calls := []Call{
		{From: "example.com/app.A", To: "example.com/app.B", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "app.go", Line: 2, Column: 2}},
	}

	got, err := BuildPackageGraph(symbols, calls, PackageGraphOptions{
		ModulePath: "example.com/app",
	})
	if err != nil {
		t.Fatalf("BuildPackageGraph returned error: %v", err)
	}

	requirePackageNode(t, got, ".", 2, 1)
	forbidPackageEdge(t, got, ".", ".")
}

func TestBuildPackageGraphCanShowSelfEdges(t *testing.T) {
	symbols := []Symbol{
		{ID: "example.com/app.A", Label: "A", Kind: "function", Package: "example.com/app", File: "app.go", StartLine: 1, EndLine: 2},
		{ID: "example.com/app.B", Label: "B", Kind: "function", Package: "example.com/app", File: "app.go", StartLine: 4, EndLine: 5},
	}
	calls := []Call{
		{From: "example.com/app.A", To: "example.com/app.B", Kind: "call", Resolution: EdgeResolutionResolved, Callsite: Callsite{File: "app.go", Line: 2, Column: 2}},
	}

	got, err := BuildPackageGraph(symbols, calls, PackageGraphOptions{
		ModulePath:       "example.com/app",
		IncludeSelfEdges: true,
	})
	if err != nil {
		t.Fatalf("BuildPackageGraph returned error: %v", err)
	}

	requirePackageEdge(t, got, ".", ".", 1)
}

func TestBuildPackageGraphAccumulatesCallCount(t *testing.T) {
	symbols := layeredSymbols()
	calls := append(layeredCalls(), Call{
		From:       "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
		To:         "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
		Kind:       "call",
		Resolution: EdgeResolutionResolved,
		Callsite:   Callsite{File: "internal/handler/user.go", Line: 17, Column: 29},
	})

	got, err := BuildPackageGraph(symbols, calls, PackageGraphOptions{
		BuildOptions: BuildOptions{
			Entry: "main.main",
			Depth: 5,
		},
		ModulePath: "github.com/tepzxl/codemap/examples/layered-service",
	})
	if err != nil {
		t.Fatalf("BuildPackageGraph returned error: %v", err)
	}

	requirePackageEdge(t, got, "internal/handler", "internal/service", 2)
	requirePackageNode(t, got, "internal/handler", 1, 3)
	requirePackageNode(t, got, "internal/service", 1, 3)
}

func TestBuildPackageGraphNoEntryUsesWholeProject(t *testing.T) {
	got, err := BuildPackageGraph(layeredSymbols(), layeredCalls(), PackageGraphOptions{
		ModulePath: "github.com/tepzxl/codemap/examples/layered-service",
	})
	if err != nil {
		t.Fatalf("BuildPackageGraph returned error: %v", err)
	}

	requirePackageEdge(t, got, "cmd/api", "internal/handler", 1)
	requirePackageEdge(t, got, "internal/handler", "internal/service", 1)
	requirePackageEdge(t, got, "internal/service", "internal/repository", 1)
}

func TestBuildPackageGraphPackageFilterKeepsCrossPackageContext(t *testing.T) {
	got, err := BuildPackageGraph(layeredSymbols(), layeredCalls(), PackageGraphOptions{
		BuildOptions: BuildOptions{
			Entry:           "main.main",
			Depth:           5,
			PackagePrefixes: []string{"github.com/tepzxl/codemap/examples/layered-service/internal/service"},
		},
		ModulePath: "github.com/tepzxl/codemap/examples/layered-service",
	})
	if err != nil {
		t.Fatalf("BuildPackageGraph returned error: %v", err)
	}

	requirePackageNode(t, got, "internal/handler", 1, 1)
	requirePackageNode(t, got, "internal/service", 1, 2)
	requirePackageNode(t, got, "internal/repository", 1, 1)
	requirePackageEdge(t, got, "internal/handler", "internal/service", 1)
	requirePackageEdge(t, got, "internal/service", "internal/repository", 1)
	forbidPackageEdge(t, got, "cmd/api", "internal/handler")
}

func requirePackageNode(t *testing.T, graph PackageGraph, id string, symbols int, calls int) {
	t.Helper()
	for _, node := range graph.Nodes {
		if node.ID == id {
			if node.Package != id {
				t.Fatalf("package node %q package = %q, want %q", id, node.Package, id)
			}
			if node.Symbols != symbols || node.Calls != calls {
				t.Fatalf("package node %q counts = symbols:%d calls:%d, want symbols:%d calls:%d", id, node.Symbols, node.Calls, symbols, calls)
			}
			return
		}
	}
	t.Fatalf("missing package node %q in %#v", id, graph.Nodes)
}

func requirePackageEdge(t *testing.T, graph PackageGraph, from string, to string, calls int) {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to {
			if edge.Calls != calls {
				t.Fatalf("package edge %q -> %q calls = %d, want %d", from, to, edge.Calls, calls)
			}
			return
		}
	}
	t.Fatalf("missing package edge %q -> %q in %#v", from, to, graph.Edges)
}

func forbidPackageEdge(t *testing.T, graph PackageGraph, from string, to string) {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to {
			t.Fatalf("unexpected package edge %q -> %q in %#v", from, to, graph.Edges)
		}
	}
}
