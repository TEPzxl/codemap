package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

func TestAssetExistsAllowsHashedStaticFilenames(t *testing.T) {
	files := fstest.MapFS{
		"_next/static/chunks/0qymxd.i09ub..css": &fstest.MapFile{Data: []byte("body{}")},
		"index.html":                            &fstest.MapFile{Data: []byte("<!doctype html>")},
	}

	if !assetExists(files, "_next/static/chunks/0qymxd.i09ub..css") {
		t.Fatal("expected hashed CSS filename containing consecutive dots to be accepted")
	}
	if assetExists(files, "../index.html") {
		t.Fatal("expected parent directory traversal to be rejected")
	}
	if assetExists(files, "_next/static/../index.html") {
		t.Fatal("expected normalized traversal path to be rejected")
	}
}

func TestStaticAssetRequestDetection(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/_next/static/chunks/app.js", want: true},
		{path: "/favicon.ico", want: true},
		{path: "/docs/readme.html", want: true},
		{path: "/graph/deep-link", want: false},
		{path: "/", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isStaticAssetRequest(tt.path); got != tt.want {
				t.Fatalf("isStaticAssetRequest(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestAPIHandlers(t *testing.T) {
	project := loadTestProject(t)
	handler := NewHandler(project)

	t.Run("health", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got struct {
			Status string `json:"status"`
		}
		decodeJSON(t, rr, &got)
		if got.Status != "ok" {
			t.Fatalf("status body mismatch: %#v", got)
		}
	})

	t.Run("symbols", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/symbols", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got struct {
			Symbols []struct {
				ID string `json:"id"`
			} `json:"symbols"`
			Warnings []struct {
				Code string `json:"code"`
			} `json:"warnings"`
		}
		decodeJSON(t, rr, &got)
		if len(got.Symbols) == 0 {
			t.Fatal("expected symbols response to include symbols")
		}
		requireSymbolID(t, got.Symbols, "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main")
	})

	t.Run("entrypoints", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/entrypoints", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got struct {
			Entrypoints []analyzer.Entrypoint `json:"entrypoints"`
			Note        string                `json:"note"`
		}
		decodeJSON(t, rr, &got)
		if got.Note == "" || len(got.Entrypoints) == 0 {
			t.Fatalf("expected entrypoints and heuristic note, got %#v", got)
		}
		if got.Entrypoints[0].ID != "github.com/tepzxl/codemap/examples/layered-service/cmd/api.main" {
			t.Fatalf("first entrypoint = %q, want main", got.Entrypoints[0].ID)
		}
		requireServerEntrypointReason(t, got.Entrypoints, "github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser", "receiver:Handler")
	})

	t.Run("meta", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got ProjectMeta
		decodeJSON(t, rr, &got)
		if got.Root == "" {
			t.Fatal("expected root in meta response")
		}
		if got.Packages == 0 || got.Symbols == 0 || got.Calls == 0 {
			t.Fatalf("expected non-zero counts, got %#v", got)
		}
		if got.AnalyzedAt == "" || got.AnalysisDurationMS < 0 || got.Version == "" {
			t.Fatalf("expected analysis metadata, got %#v", got)
		}
	})

	t.Run("graph", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/graph?entry=main.main&depth=5", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got graphmodel.Graph
		decodeJSON(t, rr, &got)
		if len(got.Nodes) == 0 || len(got.Edges) == 0 {
			t.Fatalf("expected graph nodes and edges, got %#v", got)
		}
		if err := got.Validate(); err != nil {
			t.Fatalf("invalid graph schema: %v", err)
		}
	})

	t.Run("export json", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/export?entry=main.main&depth=5&format=json&direction=downstream", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
			t.Fatalf("content type = %q, want json", contentType)
		}
		var got graphmodel.Graph
		decodeJSON(t, rr, &got)
		requireServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
			graphmodel.EdgeResolutionResolved,
			false,
		)
	})

	t.Run("export mermaid and dot", func(t *testing.T) {
		tests := []struct {
			format string
			want   string
		}{
			{format: "mermaid", want: "flowchart LR\n"},
			{format: "dot", want: "digraph codemap {\n"},
		}
		for _, tt := range tests {
			t.Run(tt.format, func(t *testing.T) {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/export?entry=main.main&depth=5&format="+tt.format, nil)
				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
				}
				if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
					t.Fatalf("content type = %q, want text/plain", contentType)
				}
				if !strings.Contains(rr.Body.String(), tt.want) || !strings.Contains(rr.Body.String(), "UserService.CreateUser") {
					t.Fatalf("export %s output mismatch:\n%s", tt.format, rr.Body.String())
				}
			})
		}
	})

	t.Run("package graph", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/package-graph?entry=main.main&depth=5", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got graphmodel.PackageGraph
		decodeJSON(t, rr, &got)
		requireServerPackageNode(t, got, "internal/service")
		requireServerPackageEdge(t, got, "internal/handler", "internal/service", 2)
		requireServerPackageEdge(t, got, "internal/service", "internal/repository", 2)
	})

	t.Run("graph upstream direction", func(t *testing.T) {
		entry := "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser"
		got := requestGraph(t, handler, "/api/graph?entry="+entry+"&depth=2&direction=upstream")
		requireServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			graphmodel.EdgeResolutionResolved,
			false,
		)
		forbidServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
		)
	})

	t.Run("graph both direction", func(t *testing.T) {
		entry := "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser"
		got := requestGraph(t, handler, "/api/graph?entry="+entry+"&depth=1&direction=both")
		requireServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/handler.(*UserHandler).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			graphmodel.EdgeResolutionResolved,
			false,
		)
		requireServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
			graphmodel.EdgeResolutionResolved,
			false,
		)
	})

	t.Run("source", func(t *testing.T) {
		nodeID := "github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser"
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/source?node_id="+nodeID, nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got struct {
			NodeID    string `json:"node_id"`
			File      string `json:"file"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
			Source    string `json:"source"`
			Language  string `json:"language"`
		}
		decodeJSON(t, rr, &got)
		if got.NodeID != nodeID {
			t.Fatalf("node id mismatch: got %q", got.NodeID)
		}
		if got.File != "internal/service/user.go" {
			t.Fatalf("file mismatch: got %q", got.File)
		}
		if got.StartLine != 15 || got.EndLine != 17 {
			t.Fatalf("line range mismatch: got %d-%d", got.StartLine, got.EndLine)
		}
		if !strings.Contains(got.Source, "func (s *UserService) CreateUser") {
			t.Fatalf("source snippet mismatch: %q", got.Source)
		}
		if got.Language != "go" {
			t.Fatalf("language mismatch: got %q", got.Language)
		}
	})

	t.Run("callsite", func(t *testing.T) {
		graph := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5")
		edge := findServerGraphEdge(t, graph,
			"github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
		)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/callsite?entry=main.main&depth=5&edge_id="+edge.ID, nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got struct {
			EdgeID        string `json:"edge_id"`
			File          string `json:"file"`
			Line          int    `json:"line"`
			Column        int    `json:"column"`
			StartLine     int    `json:"start_line"`
			EndLine       int    `json:"end_line"`
			Source        string `json:"source"`
			HighlightLine int    `json:"highlight_line"`
			Language      string `json:"language"`
		}
		decodeJSON(t, rr, &got)
		if got.EdgeID != edge.ID {
			t.Fatalf("edge id mismatch: got %q want %q", got.EdgeID, edge.ID)
		}
		if got.File != "internal/service/user.go" {
			t.Fatalf("file mismatch: got %q", got.File)
		}
		if got.Line != edge.Callsite.Line || got.Column != edge.Callsite.Column || got.HighlightLine != edge.Callsite.Line {
			t.Fatalf("callsite mismatch: got line=%d column=%d highlight=%d, edge=%#v", got.Line, got.Column, got.HighlightLine, edge.Callsite)
		}
		if got.StartLine > got.Line || got.EndLine < got.Line {
			t.Fatalf("line range %d-%d does not contain callsite line %d", got.StartLine, got.EndLine, got.Line)
		}
		if !strings.Contains(got.Source, "s.repo.Save") {
			t.Fatalf("callsite source mismatch: %q", got.Source)
		}
		if got.Language != "go" {
			t.Fatalf("language mismatch: got %q", got.Language)
		}
	})

	t.Run("warnings", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/warnings", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
		}
		var got struct {
			Warnings []struct {
				Code string `json:"code"`
			} `json:"warnings"`
		}
		decodeJSON(t, rr, &got)
		if got.Warnings == nil {
			t.Fatal("expected warnings field to be present")
		}
	})
}

func TestAPIRescan(t *testing.T) {
	project := loadTestProject(t)
	handler := NewHandler(project)

	before := requestMeta(t, handler)
	time.Sleep(time.Millisecond)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rescan", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var got struct {
		Meta ProjectMeta `json:"meta"`
	}
	decodeJSON(t, rr, &got)
	if got.Meta.Symbols != before.Symbols || got.Meta.Calls != before.Calls || got.Meta.Packages != before.Packages {
		t.Fatalf("rescan changed stable counts unexpectedly: before=%#v after=%#v", before, got.Meta)
	}
	if got.Meta.AnalyzedAt == before.AnalyzedAt {
		t.Fatalf("expected analyzed_at to update, before=%q after=%q", before.AnalyzedAt, got.Meta.AnalyzedAt)
	}

	after := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5")
	if len(after.Nodes) == 0 || len(after.Edges) == 0 {
		t.Fatalf("graph unavailable after rescan: %#v", after)
	}
}

func TestAPIRescanFailureKeepsOldIndex(t *testing.T) {
	project := loadTestProject(t)
	project.loadProject = func(string) (*Project, error) {
		return nil, os.ErrNotExist
	}
	handler := NewHandler(project)
	before := requestMeta(t, handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rescan", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
	var errBody struct {
		Error string `json:"error"`
	}
	decodeJSON(t, rr, &errBody)
	if errBody.Error == "" {
		t.Fatalf("expected JSON error, got %s", rr.Body.String())
	}

	after := requestMeta(t, handler)
	if after.Symbols != before.Symbols || after.Calls != before.Calls || after.Packages != before.Packages || after.AnalyzedAt != before.AnalyzedAt {
		t.Fatalf("old index was modified after failed rescan: before=%#v after=%#v", before, after)
	}
	graph := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5")
	if len(graph.Nodes) == 0 || len(graph.Edges) == 0 {
		t.Fatalf("old graph unavailable after failed rescan: %#v", graph)
	}
}

func TestAPIConcurrentRequestsDuringRescan(t *testing.T) {
	project := loadTestProject(t)
	handler := NewHandler(project)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = requestMeta(t, handler)
				_ = requestGraph(t, handler, "/api/graph?entry=main.main&depth=5")
			}
		}()
	}

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/rescan", nil)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("rescan status = %d, body = %s", rr.Code, rr.Body.String())
		}
	}
	wg.Wait()
}

func TestAPIGraphExpandInterface(t *testing.T) {
	project, err := LoadProject(filepath.Join(findRepoRoot(t), "examples", "interface-call"))
	if err != nil {
		t.Fatalf("LoadProject returned error: %v", err)
	}
	handler := NewHandler(project)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/graph?entry=main.main&depth=5&expand_interface=true", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var got graphmodel.Graph
	decodeJSON(t, rr, &got)
	requireServerGraphEdge(t, got,
		"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
		"github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
		graphmodel.EdgeResolutionInterface,
		true,
	)
}

func TestAPIGraphFilters(t *testing.T) {
	project, err := LoadProject(filepath.Join(findRepoRoot(t), "examples", "interface-call"))
	if err != nil {
		t.Fatalf("LoadProject returned error: %v", err)
	}
	handler := NewHandler(project)

	t.Run("show interface", func(t *testing.T) {
		got := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5&show_interface=true")
		requireServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/service.UserRepository.Save",
			graphmodel.EdgeResolutionInterface,
			false,
		)
		forbidServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/interface-call/service.(*UserService).CreateUser",
			"github.com/tepzxl/codemap/examples/interface-call/repository.(*MemoryUserRepository).Save",
		)
	})

	t.Run("package and node limit", func(t *testing.T) {
		got := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5&package=github.com/tepzxl/codemap/examples/interface-call/service&node_limit=1")
		if len(got.Nodes) != 1 {
			t.Fatalf("node count = %d, want 1: %#v", len(got.Nodes), got.Nodes)
		}
		if got.Nodes[0].Package != "github.com/tepzxl/codemap/examples/interface-call/service" {
			t.Fatalf("package filter did not apply: %#v", got.Nodes)
		}
		requireServerGraphWarning(t, got, "node-limit-exceeded")
	})
}

func TestAPIGraphExternalAndUnresolvedFilters(t *testing.T) {
	t.Run("show external", func(t *testing.T) {
		project := loadTestProject(t)
		handler := NewHandler(project)

		got := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5&show_external=true")
		requireServerGraphEdge(t, got,
			"github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save",
			"errors.New",
			graphmodel.EdgeResolutionExternal,
			false,
		)
	})

	t.Run("show unresolved", func(t *testing.T) {
		project := &Project{
			Symbols: []analyzer.Symbol{
				{
					ID:        "example.com/app.main",
					Label:     "main",
					Kind:      "function",
					Package:   "example.com/app",
					File:      "main.go",
					StartLine: 1,
					EndLine:   3,
				},
			},
			Calls: []analyzer.Call{
				{
					From:       "example.com/app.main",
					To:         "dynamic",
					Kind:       "call",
					Resolution: graphmodel.EdgeResolutionUnresolved,
					Callsite:   graphmodel.Callsite{File: "main.go", Line: 2, Column: 9},
				},
			},
		}
		handler := NewHandler(project)

		hidden := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5")
		forbidServerGraphEdge(t, hidden, "example.com/app.main", "dynamic")

		visible := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5&show_unresolved=true")
		requireServerGraphEdge(t, visible,
			"example.com/app.main",
			"dynamic",
			graphmodel.EdgeResolutionUnresolved,
			false,
		)
	})
}

func TestAPIPath(t *testing.T) {
	project := loadTestProject(t)
	handler := NewHandler(project)

	t.Run("finds layered service path", func(t *testing.T) {
		got := requestPath(t, handler, "/api/path?from=main.main&to=UserRepository.Save&max_depth=8&limit=5")
		if len(got.Paths) != 1 {
			t.Fatalf("path count = %d, want 1: %#v", len(got.Paths), got.Paths)
		}
		wantTo := "github.com/tepzxl/codemap/examples/layered-service/internal/repository.(*UserRepository).Save"
		if got.To != wantTo {
			t.Fatalf("to = %q, want %q", got.To, wantTo)
		}
		if len(got.Graph.Nodes) == 0 || len(got.Graph.Edges) == 0 {
			t.Fatalf("expected path graph, got %#v", got.Graph)
		}
	})

	t.Run("unreachable returns empty paths with warning", func(t *testing.T) {
		got := requestPath(t, handler, "/api/path?from=main.main&to=UserRepository.Save&max_depth=1&limit=5")
		if len(got.Paths) != 0 {
			t.Fatalf("expected no paths, got %#v", got.Paths)
		}
		if len(got.Warnings) == 0 || got.Warnings[0].Code != "path-not-found" {
			t.Fatalf("expected path-not-found warning, got %#v", got.Warnings)
		}
	})
}

func TestAPIPathMatchesCLI(t *testing.T) {
	repoRoot := findRepoRoot(t)
	project := loadTestProject(t)
	handler := NewHandler(project)

	apiPath := requestPath(t, handler, "/api/path?from=main.main&to=UserRepository.Save&max_depth=8&limit=5")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := cliRunGraphForServerTest([]string{
		"path",
		filepath.Join(repoRoot, "examples", "layered-service"),
		"--from", "main.main",
		"--to", "UserRepository.Save",
		"--max-depth", "8",
		"--limit", "5",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("path command exit code = %d, stderr = %s", code, stderr.String())
	}
	var cliPath graphmodel.PathResult
	if err := json.Unmarshal(stdout.Bytes(), &cliPath); err != nil {
		t.Fatalf("CLI path output is not json: %v\n%s", err, stdout.String())
	}

	left, _ := json.Marshal(apiPath)
	right, _ := json.Marshal(cliPath)
	if string(left) != string(right) {
		t.Fatalf("API and CLI path differ:\napi=%s\ncli=%s", left, right)
	}
}

func TestAPIGraphMatchesCLIForSameFilters(t *testing.T) {
	repoRoot := findRepoRoot(t)
	project, err := LoadProject(filepath.Join(repoRoot, "examples", "interface-call"))
	if err != nil {
		t.Fatalf("LoadProject returned error: %v", err)
	}
	handler := NewHandler(project)

	apiGraph := requestGraph(t, handler, "/api/graph?entry=main.main&depth=5&expand_interface=true&package=github.com/tepzxl/codemap/examples/interface-call/service")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := cliRunGraphForServerTest([]string{
		"graph",
		filepath.Join(repoRoot, "examples", "interface-call"),
		"--entry", "main.main",
		"--depth", "5",
		"--expand-interface",
		"--package", "github.com/tepzxl/codemap/examples/interface-call/service",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("graph command exit code = %d, stderr = %s", code, stderr.String())
	}
	var cliGraph graphmodel.Graph
	if err := json.Unmarshal(stdout.Bytes(), &cliGraph); err != nil {
		t.Fatalf("CLI graph output is not json: %v\n%s", err, stdout.String())
	}

	if !graphsEqual(apiGraph, cliGraph) {
		t.Fatalf("API and CLI graph differ:\napi=%#v\ncli=%#v", apiGraph, cliGraph)
	}
}

func TestAPIErrors(t *testing.T) {
	project := loadTestProject(t)
	handler := NewHandler(project)

	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "unknown source node", path: "/api/source?node_id=not.exists", want: http.StatusNotFound},
		{name: "invalid graph entry", path: "/api/graph?entry=not.exists&depth=5", want: http.StatusBadRequest},
		{name: "invalid graph depth", path: "/api/graph?entry=main.main&depth=-1", want: http.StatusBadRequest},
		{name: "invalid graph direction", path: "/api/graph?entry=main.main&direction=sideways", want: http.StatusBadRequest},
		{name: "invalid graph bool", path: "/api/graph?entry=main.main&show_external=maybe", want: http.StatusBadRequest},
		{name: "invalid graph node limit", path: "/api/graph?entry=main.main&node_limit=-1", want: http.StatusBadRequest},
		{name: "invalid export format", path: "/api/export?entry=main.main&format=svg", want: http.StatusBadRequest},
		{name: "invalid package graph entry", path: "/api/package-graph?entry=not.exists&depth=5", want: http.StatusBadRequest},
		{name: "invalid package graph depth", path: "/api/package-graph?depth=-1", want: http.StatusBadRequest},
		{name: "invalid package graph bool", path: "/api/package-graph?show_external=maybe", want: http.StatusBadRequest},
		{name: "missing path from", path: "/api/path?to=UserRepository.Save", want: http.StatusBadRequest},
		{name: "unknown path from", path: "/api/path?from=not.exists&to=UserRepository.Save", want: http.StatusBadRequest},
		{name: "invalid path max depth", path: "/api/path?from=main.main&to=UserRepository.Save&max_depth=-1", want: http.StatusBadRequest},
		{name: "invalid path limit", path: "/api/path?from=main.main&to=UserRepository.Save&limit=-1", want: http.StatusBadRequest},
		{name: "package filter removes all nodes", path: "/api/graph?entry=main.main&package=example.com/no-match", want: http.StatusBadRequest},
		{name: "missing graph entry", path: "/api/graph?depth=5", want: http.StatusBadRequest},
		{name: "missing callsite edge", path: "/api/callsite?entry=main.main&depth=5", want: http.StatusBadRequest},
		{name: "unknown callsite edge", path: "/api/callsite?entry=main.main&depth=5&edge_id=edge-missing", want: http.StatusNotFound},
		{name: "invalid callsite bool", path: "/api/callsite?entry=main.main&edge_id=edge-000001&show_external=maybe", want: http.StatusBadRequest},
		{name: "rescan requires post", path: "/api/rescan", want: http.StatusMethodNotAllowed},
		{name: "unknown endpoint", path: "/api/not-found", want: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.want {
				t.Fatalf("status = %d, want %d, body = %s", rr.Code, tt.want, rr.Body.String())
			}
			var got struct {
				Error string `json:"error"`
			}
			decodeJSON(t, rr, &got)
			if got.Error == "" {
				t.Fatalf("expected json error, got %s", rr.Body.String())
			}
		})
	}
}

func requestPath(t *testing.T, handler http.Handler, path string) graphmodel.PathResult {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var got graphmodel.PathResult
	decodeJSON(t, rr, &got)
	return got
}

func TestSourcePathTraversalReturnsJSONError(t *testing.T) {
	project := &Project{
		Root: t.TempDir(),
		Symbols: []analyzer.Symbol{
			{
				ID:        "example.com/app.Escape",
				Label:     "Escape",
				Kind:      "function",
				Package:   "example.com/app",
				File:      "../outside.go",
				StartLine: 1,
				EndLine:   1,
			},
		},
		symbolByID: map[string]analyzer.Symbol{},
	}
	project.symbolByID[project.Symbols[0].ID] = project.Symbols[0]

	handler := NewHandler(project)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/source?node_id=example.com/app.Escape", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	var got struct {
		Error string `json:"error"`
	}
	decodeJSON(t, rr, &got)
	if !strings.Contains(got.Error, "escapes project root") {
		t.Fatalf("expected path traversal error, got %q", got.Error)
	}
}

func TestCallsitePathTraversalReturnsJSONError(t *testing.T) {
	project := &Project{
		Root: t.TempDir(),
		Symbols: []analyzer.Symbol{
			{
				ID:        "example.com/app.main",
				Label:     "main",
				Kind:      analyzer.SymbolKindFunction,
				Package:   "example.com/app",
				File:      "main.go",
				StartLine: 1,
				EndLine:   3,
			},
			{
				ID:        "example.com/app.helper",
				Label:     "helper",
				Kind:      analyzer.SymbolKindFunction,
				Package:   "example.com/app",
				File:      "helper.go",
				StartLine: 1,
				EndLine:   3,
			},
		},
		Calls: []analyzer.Call{
			{
				From:       "example.com/app.main",
				To:         "example.com/app.helper",
				Kind:       analyzer.CallKind,
				Resolution: graphmodel.EdgeResolutionResolved,
				Callsite:   graphmodel.Callsite{File: "../outside.go", Line: 2, Column: 2},
			},
		},
	}

	handler := NewHandler(project)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/callsite?entry=main.main&depth=5&edge_id=edge-000001", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	var got struct {
		Error string `json:"error"`
	}
	decodeJSON(t, rr, &got)
	if !strings.Contains(got.Error, "escapes project root") {
		t.Fatalf("expected path traversal error, got %q", got.Error)
	}
}

func TestWebUIStaticRootIsServed(t *testing.T) {
	project := loadTestProject(t)
	handler := NewHandler(project)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("content type = %q, want html", contentType)
	}
	if !strings.Contains(rr.Body.String(), "codemap") {
		t.Fatalf("expected embedded web UI html, got %q", rr.Body.String())
	}
}

func loadTestProject(t *testing.T) *Project {
	t.Helper()
	project, err := LoadProject(filepath.Join(findRepoRoot(t), "examples", "layered-service"))
	if err != nil {
		t.Fatalf("LoadProject returned error: %v", err)
	}
	return project
}

func decodeJSON(t *testing.T, rr *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), out); err != nil {
		t.Fatalf("response is not json: %v\n%s", err, rr.Body.String())
	}
}

func requestGraph(t *testing.T, handler http.Handler, path string) graphmodel.Graph {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var got graphmodel.Graph
	decodeJSON(t, rr, &got)
	return got
}

func requestMeta(t *testing.T, handler http.Handler) ProjectMeta {
	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	var got ProjectMeta
	decodeJSON(t, rr, &got)
	return got
}

func cliRunGraphForServerTest(args []string, stdout *bytes.Buffer, stderr *bytes.Buffer) int {
	cmd := exec.Command("go", append([]string{"run", "./cmd/codemap"}, args...)...)
	cmd.Dir = findRepoRootForExec()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		stderr.WriteString(err.Error())
		return 1
	}
	return 0
}

func findRepoRootForExec() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

func requireSymbolID(t *testing.T, symbols []struct {
	ID string `json:"id"`
}, id string) {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.ID == id {
			return
		}
	}
	t.Fatalf("missing symbol %q in %#v", id, symbols)
}

func requireServerGraphEdge(t *testing.T, output graphmodel.Graph, from string, to string, resolution graphmodel.EdgeResolution, candidate bool) {
	t.Helper()

	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to && edge.Resolution == resolution && edge.Candidate == candidate {
			return
		}
	}
	t.Fatalf("missing graph edge from %q to %q resolution %q candidate %t in %#v", from, to, resolution, candidate, output.Edges)
}

func findServerGraphEdge(t *testing.T, output graphmodel.Graph, from string, to string) graphmodel.Edge {
	t.Helper()

	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to {
			return edge
		}
	}
	t.Fatalf("missing graph edge from %q to %q in %#v", from, to, output.Edges)
	return graphmodel.Edge{}
}

func forbidServerGraphEdge(t *testing.T, output graphmodel.Graph, from string, to string) {
	t.Helper()

	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to {
			t.Fatalf("unexpected graph edge from %q to %q in %#v", from, to, output.Edges)
		}
	}
}

func requireServerGraphWarning(t *testing.T, output graphmodel.Graph, code string) {
	t.Helper()

	for _, warning := range output.Warnings {
		if warning.Code == code {
			return
		}
	}
	t.Fatalf("missing graph warning %q in %#v", code, output.Warnings)
}

func requireServerEntrypointReason(t *testing.T, entrypoints []analyzer.Entrypoint, id string, reason string) {
	t.Helper()

	for _, entrypoint := range entrypoints {
		if entrypoint.ID != id {
			continue
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

func requireServerPackageNode(t *testing.T, output graphmodel.PackageGraph, id string) {
	t.Helper()

	for _, node := range output.Nodes {
		if node.ID == id {
			return
		}
	}
	t.Fatalf("missing package node %q in %#v", id, output.Nodes)
}

func requireServerPackageEdge(t *testing.T, output graphmodel.PackageGraph, from string, to string, calls int) {
	t.Helper()

	for _, edge := range output.Edges {
		if edge.From == from && edge.To == to {
			if edge.Calls != calls {
				t.Fatalf("package edge %q -> %q calls = %d, want %d", from, to, edge.Calls, calls)
			}
			return
		}
	}
	t.Fatalf("missing package edge %q -> %q in %#v", from, to, output.Edges)
}

func graphsEqual(a graphmodel.Graph, b graphmodel.Graph) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(left) == string(right)
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
