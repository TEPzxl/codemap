package server

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tepzxl/codemap/internal/analyzer"
	graphmodel "github.com/tepzxl/codemap/internal/graph"
)

const Version = "v0.2.0"

type Project struct {
	mu sync.RWMutex

	Root          string
	Module        string
	Symbols       []analyzer.Symbol
	Calls         []analyzer.Call
	ExpandedCalls []analyzer.Call
	Warnings      []analyzer.AnalyzeWarning
	Packages      int
	AnalyzedAt    time.Time
	Duration      time.Duration

	symbolByID  map[string]analyzer.Symbol
	loadProject func(string) (*Project, error)
}

type ProjectMeta struct {
	Root               string `json:"root"`
	Module             string `json:"module"`
	Packages           int    `json:"packages"`
	Symbols            int    `json:"symbols"`
	Calls              int    `json:"calls"`
	Warnings           int    `json:"warnings"`
	AnalyzedAt         string `json:"analyzed_at"`
	AnalysisDurationMS int64  `json:"analysis_duration_ms"`
	Version            string `json:"version"`
}

func LoadProject(rootPath string) (*Project, error) {
	started := time.Now()
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
	expandedCalls, err := analyzer.ExtractCallsWithOptions(loadResult, symbols, analyzer.CallOptions{
		ExpandInterface: true,
	})
	if err != nil {
		return nil, err
	}

	symbolByID := make(map[string]analyzer.Symbol, len(symbols))
	for _, symbol := range symbols {
		symbolByID[symbol.ID] = symbol
	}

	modulePath, err := readModulePath(rootAbs)
	if err != nil {
		modulePath = ""
	}

	return &Project{
		Root:          rootAbs,
		Module:        modulePath,
		Symbols:       symbols,
		Calls:         calls,
		ExpandedCalls: expandedCalls,
		Warnings:      loadResult.Warnings,
		Packages:      len(loadResult.Packages),
		AnalyzedAt:    time.Now(),
		Duration:      time.Since(started),
		symbolByID:    symbolByID,
	}, nil
}

func NewHandler(project *Project) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", project.handleHealth)
	mux.HandleFunc("/api/meta", project.handleMeta)
	mux.HandleFunc("/api/rescan", project.handleRescan)
	mux.HandleFunc("/api/symbols", project.handleSymbols)
	mux.HandleFunc("/api/graph", project.handleGraph)
	mux.HandleFunc("/api/source", project.handleSource)
	mux.HandleFunc("/api/callsite", project.handleCallsite)
	mux.HandleFunc("/api/warnings", project.handleWarnings)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		webHandler().ServeHTTP(w, r)
	})
	return mux
}

func (p *Project) Meta() ProjectMeta {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.metaLocked()
}

func (p *Project) Rescan() (ProjectMeta, error) {
	root := p.root()
	load := p.loadProject
	if load == nil {
		load = LoadProject
	}

	next, err := load(root)
	if err != nil {
		return ProjectMeta{}, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.Root = next.Root
	p.Module = next.Module
	p.Symbols = next.Symbols
	p.Calls = next.Calls
	p.ExpandedCalls = next.ExpandedCalls
	p.Warnings = next.Warnings
	p.Packages = next.Packages
	p.AnalyzedAt = next.AnalyzedAt
	p.Duration = next.Duration
	p.symbolByID = next.symbolByID
	return p.metaLocked(), nil
}

func (p *Project) BuildGraph(options graphmodel.BuildOptions) (graphmodel.Graph, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	calls := p.Calls
	if options.ExpandInterface {
		calls = p.ExpandedCalls
	}
	return graphmodel.BuildGraph(toGraphSymbols(p.Symbols), toGraphCalls(calls), options)
}

func (p *Project) sourceSymbol(nodeID string) (string, analyzer.Symbol, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	symbol, ok := p.symbolByID[nodeID]
	if !ok {
		for _, candidate := range p.Symbols {
			if candidate.ID == nodeID {
				symbol = candidate
				ok = true
				break
			}
		}
	}
	return p.Root, symbol, ok
}

func (p *Project) root() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Root
}

func (p *Project) metaLocked() ProjectMeta {
	analyzedAt := ""
	if !p.AnalyzedAt.IsZero() {
		analyzedAt = p.AnalyzedAt.UTC().Format(time.RFC3339Nano)
	}
	return ProjectMeta{
		Root:               p.Root,
		Module:             p.Module,
		Packages:           p.Packages,
		Symbols:            len(p.Symbols),
		Calls:              len(p.Calls),
		Warnings:           len(p.Warnings),
		AnalyzedAt:         analyzedAt,
		AnalysisDurationMS: p.Duration.Milliseconds(),
		Version:            Version,
	}
}

func readModulePath(root string) (string, error) {
	file, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module path not found")
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
			Candidate:  call.Candidate,
		})
	}
	return result
}
