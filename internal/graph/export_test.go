package graph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExportGraphJSON(t *testing.T) {
	graph := exportFixtureGraph()

	got, contentType, err := ExportGraph(graph, ExportFormatJSON)
	if err != nil {
		t.Fatalf("ExportGraph returned error: %v", err)
	}
	if contentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
	var decoded Graph
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("json export is invalid: %v\n%s", err, got)
	}
	if decoded.Entry != graph.Entry || len(decoded.Nodes) != len(graph.Nodes) || len(decoded.Edges) != len(graph.Edges) {
		t.Fatalf("json export mismatch: %#v", decoded)
	}
}

func TestExportGraphMermaidEscapesSpecialCharacters(t *testing.T) {
	got, contentType, err := ExportGraph(exportFixtureGraph(), ExportFormatMermaid)
	if err != nil {
		t.Fatalf("ExportGraph returned error: %v", err)
	}
	if contentType != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q, want text/plain", contentType)
	}
	if !strings.HasPrefix(got, "flowchart LR\n") {
		t.Fatalf("mermaid output should start with flowchart LR, got %q", got)
	}
	if !strings.Contains(got, `n0["main#91;root#93; #quot;start#quot;"]`) {
		t.Fatalf("mermaid label was not escaped: %s", got)
	}
	if !strings.Contains(got, "n0 --> n1") {
		t.Fatalf("missing mermaid edge: %s", got)
	}
}

func TestExportGraphDOTEscapesSpecialCharacters(t *testing.T) {
	got, contentType, err := ExportGraph(exportFixtureGraph(), ExportFormatDOT)
	if err != nil {
		t.Fatalf("ExportGraph returned error: %v", err)
	}
	if contentType != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q, want text/plain", contentType)
	}
	if !strings.HasPrefix(got, "digraph codemap {\n") {
		t.Fatalf("dot output should start with digraph, got %q", got)
	}
	if !strings.Contains(got, `"n0" [label="main[root] \"start\""];`) {
		t.Fatalf("dot label was not escaped with quoted string: %s", got)
	}
	if !strings.Contains(got, `"n0" -> "n1"`) {
		t.Fatalf("missing dot edge: %s", got)
	}
}

func TestExportGraphInvalidFormat(t *testing.T) {
	_, _, err := ExportGraph(exportFixtureGraph(), ExportFormat("svg"))
	if err == nil {
		t.Fatal("expected invalid format to return error")
	}
	if !strings.Contains(err.Error(), "format must be one of") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func exportFixtureGraph() Graph {
	return Graph{
		Entry: "example.com/app.main",
		Nodes: []Node{
			{
				ID:      "example.com/app.main",
				Label:   `main[root] "start"`,
				Kind:    NodeKindFunction,
				Package: "example.com/app",
				File:    "cmd/app/main.go",
			},
			{
				ID:      "example.com/app/internal/service.(*UserService).CreateUser",
				Label:   "UserService.CreateUser",
				Kind:    NodeKindMethod,
				Package: "example.com/app/internal/service",
				File:    "internal/service/user.go",
			},
		},
		Edges: []Edge{
			{
				ID:         "edge-000001",
				From:       "example.com/app.main",
				To:         "example.com/app/internal/service.(*UserService).CreateUser",
				Kind:       "call",
				Resolution: EdgeResolutionResolved,
				Callsite:   Callsite{File: "cmd/app/main.go", Line: 4, Column: 2},
			},
		},
	}
}
