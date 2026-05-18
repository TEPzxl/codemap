package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

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
		{name: "missing graph entry", path: "/api/graph?depth=5", want: http.StatusBadRequest},
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
