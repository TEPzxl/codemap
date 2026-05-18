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
	p.mu.RLock()
	defer p.mu.RUnlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"symbols":  p.Symbols,
		"warnings": p.Warnings,
	})
}

func (p *Project) handleMeta(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, p.Meta())
}

func (p *Project) handleRescan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	meta, err := p.Rescan()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"meta": meta})
}

func (p *Project) handleGraph(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	options, ok := parseGraphOptions(w, r, true)
	if !ok {
		return
	}

	graph, err := p.BuildGraph(options)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, graph)
}

func (p *Project) handlePath(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	options, ok := parsePathOptions(w, r)
	if !ok {
		return
	}
	result, err := p.FindPaths(options)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func parseGraphOptions(w http.ResponseWriter, r *http.Request, requireEntry bool) (graphmodel.BuildOptions, bool) {
	query := r.URL.Query()
	entry := query.Get("entry")
	if entry == "" {
		if requireEntry {
			writeError(w, http.StatusBadRequest, "entry is required")
			return graphmodel.BuildOptions{}, false
		}
		entry = "main.main"
	}

	depth := 5
	if rawDepth := query.Get("depth"); rawDepth != "" {
		parsed, err := strconv.Atoi(rawDepth)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "depth must be a non-negative integer")
			return graphmodel.BuildOptions{}, false
		}
		depth = parsed
	}

	showExternal, ok := parseBoolQuery(w, query.Get("show_external"), "show_external")
	if !ok {
		return graphmodel.BuildOptions{}, false
	}
	showUnresolved, ok := parseBoolQuery(w, query.Get("show_unresolved"), "show_unresolved")
	if !ok {
		return graphmodel.BuildOptions{}, false
	}
	showInterface, ok := parseBoolQuery(w, query.Get("show_interface"), "show_interface")
	if !ok {
		return graphmodel.BuildOptions{}, false
	}
	expandInterface, ok := parseBoolQuery(w, query.Get("expand_interface"), "expand_interface")
	if !ok {
		return graphmodel.BuildOptions{}, false
	}

	nodeLimit := 0
	if rawNodeLimit := query.Get("node_limit"); rawNodeLimit != "" {
		parsed, err := strconv.Atoi(rawNodeLimit)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "node_limit must be a non-negative integer")
			return graphmodel.BuildOptions{}, false
		}
		nodeLimit = parsed
	}

	direction := graphmodel.Direction(query.Get("direction"))
	if !direction.IsValid() {
		writeError(w, http.StatusBadRequest, "direction must be one of downstream, upstream, both")
		return graphmodel.BuildOptions{}, false
	}

	return graphmodel.BuildOptions{
		Entry:           entry,
		Depth:           depth,
		Direction:       direction.Normalized(),
		ShowExternal:    showExternal,
		ShowUnresolved:  showUnresolved,
		ShowInterface:   showInterface,
		ExpandInterface: expandInterface,
		PackagePrefixes: query["package"],
		NodeLimit:       nodeLimit,
	}, true
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

func parsePathOptions(w http.ResponseWriter, r *http.Request) (graphmodel.PathOptions, bool) {
	query := r.URL.Query()
	from := query.Get("from")
	if from == "" {
		writeError(w, http.StatusBadRequest, "from is required")
		return graphmodel.PathOptions{}, false
	}
	to := query.Get("to")
	if to == "" {
		writeError(w, http.StatusBadRequest, "to is required")
		return graphmodel.PathOptions{}, false
	}

	maxDepth := 8
	if rawMaxDepth := query.Get("max_depth"); rawMaxDepth != "" {
		parsed, err := strconv.Atoi(rawMaxDepth)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "max_depth must be a non-negative integer")
			return graphmodel.PathOptions{}, false
		}
		maxDepth = parsed
	}

	limit := 5
	if rawLimit := query.Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "limit must be a non-negative integer")
			return graphmodel.PathOptions{}, false
		}
		limit = parsed
	}

	showExternal, ok := parseBoolQuery(w, query.Get("show_external"), "show_external")
	if !ok {
		return graphmodel.PathOptions{}, false
	}
	showUnresolved, ok := parseBoolQuery(w, query.Get("show_unresolved"), "show_unresolved")
	if !ok {
		return graphmodel.PathOptions{}, false
	}
	showInterface, ok := parseBoolQuery(w, query.Get("show_interface"), "show_interface")
	if !ok {
		return graphmodel.PathOptions{}, false
	}
	expandInterface, ok := parseBoolQuery(w, query.Get("expand_interface"), "expand_interface")
	if !ok {
		return graphmodel.PathOptions{}, false
	}

	return graphmodel.PathOptions{
		From:            from,
		To:              to,
		MaxDepth:        maxDepth,
		Limit:           limit,
		ShowExternal:    showExternal,
		ShowUnresolved:  showUnresolved,
		ShowInterface:   showInterface,
		ExpandInterface: expandInterface,
		PackagePrefixes: query["package"],
	}, true
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

	root, symbol, ok := p.sourceSymbol(nodeID)
	if !ok {
		writeError(w, http.StatusNotFound, "node_id not found")
		return
	}

	snippet, err := source.ReadSnippet(root, source.Location{
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

func (p *Project) handleCallsite(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}

	edgeID := r.URL.Query().Get("edge_id")
	if edgeID == "" {
		writeError(w, http.StatusBadRequest, "edge_id is required")
		return
	}

	options, ok := parseGraphOptions(w, r, false)
	if !ok {
		return
	}
	graph, err := p.BuildGraph(options)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	root := p.root()

	for _, edge := range graph.Edges {
		if edge.ID != edgeID {
			continue
		}
		snippet, err := source.ReadCallsite(root, source.CallsiteLocation{
			EdgeID:        edge.ID,
			File:          edge.Callsite.File,
			Line:          edge.Callsite.Line,
			Column:        edge.Callsite.Column,
			ContextBefore: 3,
			ContextAfter:  3,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, snippet)
		return
	}

	writeError(w, http.StatusNotFound, "edge_id not found")
}

func (p *Project) handleWarnings(w http.ResponseWriter, r *http.Request) {
	if !requireGET(w, r) {
		return
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
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
