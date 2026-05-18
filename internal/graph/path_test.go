package graph

import (
	"strings"
	"testing"
)

func TestFindPaths(t *testing.T) {
	t.Run("simple fixture finds path", func(t *testing.T) {
		got, err := FindPaths(simpleSymbols(), simpleCalls(), PathOptions{
			From:     "github.com/tepzxl/codemap/examples/simple.main",
			To:       "github.com/tepzxl/codemap/examples/simple/service.CreateUser",
			MaxDepth: 4,
			Limit:    5,
		})
		if err != nil {
			t.Fatalf("FindPaths returned error: %v", err)
		}

		requirePathNodes(t, got, []string{
			"github.com/tepzxl/codemap/examples/simple.main",
			"github.com/tepzxl/codemap/examples/simple/app.Run",
			"github.com/tepzxl/codemap/examples/simple/service.CreateUser",
		})
		requireNode(t, got.Graph, "github.com/tepzxl/codemap/examples/simple.main")
		requireEdge(t, got.Graph, "github.com/tepzxl/codemap/examples/simple/app.Run", "github.com/tepzxl/codemap/examples/simple/service.CreateUser", EdgeResolutionResolved)
	})

	t.Run("layered service supports shortcuts", func(t *testing.T) {
		got, err := FindPaths(layeredSymbols(), layeredCalls(), PathOptions{
			From:     "main.main",
			To:       "UserRepository.Save",
			MaxDepth: 8,
			Limit:    5,
		})
		if err != nil {
			t.Fatalf("FindPaths returned error: %v", err)
		}

		requirePathNodes(t, got, []string{
			"github.com/tepzxl/codemap/examples/layered-service/cmd/api.main",
			"github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
		})
	})

	t.Run("max depth limits search", func(t *testing.T) {
		got, err := FindPaths(layeredSymbols(), layeredCalls(), PathOptions{
			From:     "main.main",
			To:       "UserRepository.Save",
			MaxDepth: 2,
			Limit:    5,
		})
		if err != nil {
			t.Fatalf("FindPaths returned error: %v", err)
		}
		if len(got.Paths) != 0 {
			t.Fatalf("expected no paths with max depth 2, got %#v", got.Paths)
		}
		requireWarning(t, got.Graph, "path-not-found")
	})

	t.Run("unreachable returns empty paths and warning", func(t *testing.T) {
		got, err := FindPaths(layeredSymbols(), layeredCalls(), PathOptions{
			From:     "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
			To:       "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main",
			MaxDepth: 8,
			Limit:    5,
		})
		if err != nil {
			t.Fatalf("FindPaths returned error: %v", err)
		}
		if len(got.Paths) != 0 {
			t.Fatalf("expected no paths, got %#v", got.Paths)
		}
		requireWarning(t, got.Graph, "path-not-found")
	})

	t.Run("cycle does not recurse forever", func(t *testing.T) {
		got, err := FindPaths(cycleSymbols(), cycleCalls(), PathOptions{
			From:     "example.com/cycle.A",
			To:       "example.com/cycle.B",
			MaxDepth: 20,
			Limit:    5,
		})
		if err != nil {
			t.Fatalf("FindPaths returned error: %v", err)
		}
		requirePathNodes(t, got, []string{"example.com/cycle.A", "example.com/cycle.B"})
	})

	t.Run("unknown symbol returns clear error", func(t *testing.T) {
		_, err := FindPaths(simpleSymbols(), simpleCalls(), PathOptions{
			From:     "not.exists",
			To:       "github.com/tepzxl/codemap/examples/simple/service.CreateUser",
			MaxDepth: 4,
			Limit:    5,
		})
		if err == nil {
			t.Fatal("expected unknown symbol to fail")
		}
		if !strings.Contains(err.Error(), "from symbol") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid max depth and limit return clear errors", func(t *testing.T) {
		_, err := FindPaths(simpleSymbols(), simpleCalls(), PathOptions{
			From:     "github.com/tepzxl/codemap/examples/simple.main",
			To:       "github.com/tepzxl/codemap/examples/simple/service.CreateUser",
			MaxDepth: -1,
			Limit:    5,
		})
		if err == nil || !strings.Contains(err.Error(), "max_depth") {
			t.Fatalf("expected max_depth error, got %v", err)
		}

		_, err = FindPaths(simpleSymbols(), simpleCalls(), PathOptions{
			From:     "github.com/tepzxl/codemap/examples/simple.main",
			To:       "github.com/tepzxl/codemap/examples/simple/service.CreateUser",
			MaxDepth: 4,
			Limit:    -1,
		})
		if err == nil || !strings.Contains(err.Error(), "limit") {
			t.Fatalf("expected limit error, got %v", err)
		}
	})
}

func requirePathNodes(t *testing.T, result PathResult, want []string) {
	t.Helper()
	if len(result.Paths) == 0 {
		t.Fatalf("expected at least one path, got none")
	}
	got := result.Paths[0].Nodes
	if len(got) != len(want) {
		t.Fatalf("path node count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("path node[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
	if len(result.Paths[0].Edges) != len(want)-1 {
		t.Fatalf("path edge count = %d, want %d: %#v", len(result.Paths[0].Edges), len(want)-1, result.Paths[0].Edges)
	}
}
