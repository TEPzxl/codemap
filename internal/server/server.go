package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

type Project struct {
	Root     string
	Symbols  []analyzer.Symbol
	Calls    []analyzer.Call
	Warnings []analyzer.AnalyzeWarning

	symbolByID map[string]analyzer.Symbol
}

func LoadProject(rootPath string) (*Project, error) {
	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}

	loadResult, err := analyzer.LoadPackages(analyzer.LoadRequest{
		RootPath:     rootAbs,
		IncludeTests: false,
	})
	if err != nil {
		return nil, err
	}

	symbols, err := analyzer.ExtractSymbols(loadResult)
	if err != nil {
		return nil, err
	}
	calls, err := analyzer.ExtractCalls(loadResult, symbols)
	if err != nil {
		return nil, err
	}

	symbolByID := make(map[string]analyzer.Symbol, len(symbols))
	for _, symbol := range symbols {
		symbolByID[symbol.ID] = symbol
	}

	return &Project{
		Root:       rootAbs,
		Symbols:    symbols,
		Calls:      calls,
		Warnings:   loadResult.Warnings,
		symbolByID: symbolByID,
	}, nil
}

func NewHandler(project *Project) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", project.handleHealth)
	mux.HandleFunc("/api/symbols", project.handleSymbols)
	mux.HandleFunc("/api/graph", project.handleGraph)
	mux.HandleFunc("/api/source", project.handleSource)
	mux.HandleFunc("/api/warnings", project.handleWarnings)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		webHandler().ServeHTTP(w, r)
	})
	return mux
}

func (p *Project) BuildGraph(options graphmodel.BuildOptions) (graphmodel.Graph, error) {
	return graphmodel.BuildGraph(toGraphSymbols(p.Symbols), toGraphCalls(p.Calls), options)
}

func toGraphSymbols(symbols []analyzer.Symbol) []graphmodel.Symbol {
	result := make([]graphmodel.Symbol, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, graphmodel.Symbol{
			ID:        symbol.ID,
			Label:     symbol.Label,
			Kind:      symbol.Kind,
			Package:   symbol.Package,
			Receiver:  symbol.Receiver,
			File:      symbol.File,
			StartLine: symbol.StartLine,
			EndLine:   symbol.EndLine,
		})
	}
	return result
}

func toGraphCalls(calls []analyzer.Call) []graphmodel.Call {
	result := make([]graphmodel.Call, 0, len(calls))
	for _, call := range calls {
		result = append(result, graphmodel.Call{
			From:       call.From,
			To:         call.To,
			Kind:       call.Kind,
			Resolution: call.Resolution,
			Callsite:   call.Callsite,
		})
	}
	return result
}
