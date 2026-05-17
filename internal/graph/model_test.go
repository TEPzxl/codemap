package graph

import (
	"encoding/json"
	"testing"
)

func TestNodeKindIsValid(t *testing.T) {
	valid := []NodeKind{
		NodeKindFunction,
		NodeKindMethod,
		NodeKindExternal,
		NodeKindUnresolved,
	}
	for _, kind := range valid {
		if !kind.IsValid() {
			t.Fatalf("expected node kind %q to be valid", kind)
		}
	}

	if NodeKind("invalid").IsValid() {
		t.Fatal("expected invalid node kind to fail validation")
	}
}

func TestEdgeResolutionIsValid(t *testing.T) {
	valid := []EdgeResolution{
		EdgeResolutionResolved,
		EdgeResolutionInterface,
		EdgeResolutionExternal,
		EdgeResolutionUnresolved,
	}
	for _, resolution := range valid {
		if !resolution.IsValid() {
			t.Fatalf("expected edge resolution %q to be valid", resolution)
		}
	}

	if EdgeResolution("invalid").IsValid() {
		t.Fatal("expected invalid edge resolution to fail validation")
	}
}

func TestGraphJSONRoundTrip(t *testing.T) {
	graph := Graph{
		Entry: "github.com/tepzxl/codemap/examples/simple.main",
		Nodes: []Node{
			{
				ID:        "github.com/tepzxl/codemap/examples/simple.main",
				Label:     "main",
				Kind:      NodeKindFunction,
				Package:   "github.com/tepzxl/codemap/examples/simple",
				File:      "main.go",
				StartLine: 5,
				EndLine:   9,
			},
			{
				ID:        "github.com/tepzxl/codemap/examples/simple/app.Run",
				Label:     "app.Run",
				Kind:      NodeKindFunction,
				Package:   "github.com/tepzxl/codemap/examples/simple/app",
				File:      "app/app.go",
				StartLine: 5,
				EndLine:   7,
			},
		},
		Edges: []Edge{
			{
				ID:         "edge-1",
				From:       "github.com/tepzxl/codemap/examples/simple.main",
				To:         "github.com/tepzxl/codemap/examples/simple/app.Run",
				Kind:       "call",
				Resolution: EdgeResolutionResolved,
				Callsite: Callsite{
					File:   "main.go",
					Line:   7,
					Column: 2,
				},
			},
		},
		Warnings: []Warning{
			{
				Code:    "partial-package-load",
				Message: "some packages failed to load",
			},
		},
	}

	if err := graph.Validate(); err != nil {
		t.Fatalf("expected graph to be valid before marshal: %v", err)
	}

	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal graph: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal graph into map: %v", err)
	}
	if _, ok := decoded["nodes"]; !ok {
		t.Fatal("json output missing nodes field")
	}
	if _, ok := decoded["edges"]; !ok {
		t.Fatal("json output missing edges field")
	}

	var got Graph
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal graph: %v", err)
	}

	if err := got.Validate(); err != nil {
		t.Fatalf("expected graph to be valid after round trip: %v", err)
	}
	if got.Entry != graph.Entry {
		t.Fatalf("entry mismatch: got %q want %q", got.Entry, graph.Entry)
	}
	if len(got.Nodes) != len(graph.Nodes) {
		t.Fatalf("node count mismatch: got %d want %d", len(got.Nodes), len(graph.Nodes))
	}
	if len(got.Edges) != len(graph.Edges) {
		t.Fatalf("edge count mismatch: got %d want %d", len(got.Edges), len(graph.Edges))
	}
	if len(got.Warnings) != len(graph.Warnings) {
		t.Fatalf("warning count mismatch: got %d want %d", len(got.Warnings), len(graph.Warnings))
	}
}

func TestGraphValidateAllowsExternalOrUnresolvedUnknownNodes(t *testing.T) {
	graph := Graph{
		Nodes: []Node{
			{
				ID:        "github.com/tepzxl/codemap/examples/simple.main",
				Label:     "main",
				Kind:      NodeKindFunction,
				Package:   "github.com/tepzxl/codemap/examples/simple",
				File:      "main.go",
				StartLine: 1,
				EndLine:   3,
			},
		},
		Edges: []Edge{
			{
				ID:         "edge-external",
				From:       "github.com/tepzxl/codemap/examples/simple.main",
				To:         "fmt.Println",
				Kind:       "call",
				Resolution: EdgeResolutionExternal,
				Callsite: Callsite{
					File:   "main.go",
					Line:   2,
					Column: 2,
				},
			},
		},
	}

	if err := graph.Validate(); err != nil {
		t.Fatalf("expected external unknown node edge to be allowed: %v", err)
	}
}
