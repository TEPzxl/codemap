package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	graphmodel "github.com/tepzxl/codemap/internal/graph"
	"github.com/tepzxl/codemap/internal/source"
)

type errorResponse struct {
	Error string `json:"error"`
}

func (p *Project) handleHealth(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (p *Project) handleSymbols(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"symbols":  p.Symbols,
		"warnings": p.Warnings,
	})
}

func (p *Project) handleGraph(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	query := r.URL.Query()
	entry := query.Get("entry")
	if entry == "" {
		writeError(w, http.StatusBadRequest, "entry is required")
		return
	}

	depth := 5
	if rawDepth := query.Get("depth"); rawDepth != "" {
		parsed, err := strconv.Atoi(rawDepth)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "depth must be a non-negative integer")
			return
		}
		depth = parsed
	}

	showExternal, ok := parseBoolQuery(w, query.Get("show_external"), "show_external")
	if !ok {
		return
	}
	showUnresolved, ok := parseBoolQuery(w, query.Get("show_unresolved"), "show_unresolved")
	if !ok {
		return
	}
	showInterface, ok := parseBoolQuery(w, query.Get("show_interface"), "show_interface")
	if !ok {
		return
	}
	expandInterface, ok := parseBoolQuery(w, query.Get("expand_interface"), "expand_interface")
	if !ok {
		return
	}

	nodeLimit := 0
	if rawNodeLimit := query.Get("node_limit"); rawNodeLimit != "" {
		parsed, err := strconv.Atoi(rawNodeLimit)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "node_limit must be a non-negative integer")
			return
		}
		nodeLimit = parsed
	}

	graph, err := p.BuildGraph(graphmodel.BuildOptions{
		Entry:           entry,
		Depth:           depth,
		ShowExternal:    showExternal,
		ShowUnresolved:  showUnresolved,
		ShowInterface:   showInterface,
		ExpandInterface: expandInterface,
		PackagePrefixes: query["package"],
		NodeLimit:       nodeLimit,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

func parseBoolQuery(w http.ResponseWriter, raw string, name string) (bool, bool) {
	if raw == "" {
		return false, true
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, name+" must be a boolean")
		return false, false
	}
	return parsed, true
}

func (p *Project) handleSource(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		writeError(w, http.StatusBadRequest, "node_id is required")
		return
	}

	symbol, ok := p.symbolByID[nodeID]
	if !ok {
		writeError(w, http.StatusNotFound, "node_id not found")
		return
	}

	snippet, err := source.ReadSnippet(p.Root, source.Location{
		NodeID:    symbol.ID,
		File:      symbol.File,
		StartLine: symbol.StartLine,
		EndLine:   symbol.EndLine,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snippet)
}

func (p *Project) handleWarnings(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"warnings": p.Warnings})
}

func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
