package graph

import (
	"fmt"
	"sort"
)

type PathOptions struct {
	From            string
	To              string
	MaxDepth        int
	Limit           int
	ShowExternal    bool
	ShowUnresolved  bool
	ShowInterface   bool
	ExpandInterface bool
	PackagePrefixes []string
}

type PathResult struct {
	From     string       `json:"from"`
	To       string       `json:"to"`
	Paths    []SymbolPath `json:"paths"`
	Graph    Graph        `json:"graph"`
	Warnings []Warning    `json:"warnings"`
}

type SymbolPath struct {
	Nodes []string `json:"nodes"`
	Edges []string `json:"edges"`
}

type pathState struct {
	nodes []string
	calls []Call
}

func FindPaths(symbols []Symbol, calls []Call, options PathOptions) (PathResult, error) {
	if options.MaxDepth < 0 {
		return PathResult{}, fmt.Errorf("max_depth must be >= 0")
	}
	if options.Limit < 0 {
		return PathResult{}, fmt.Errorf("limit must be >= 0")
	}
	limit := options.Limit
	if limit == 0 {
		limit = 1
	}

	from, err := resolveEntry(options.From, symbols)
	if err != nil {
		return PathResult{}, fmt.Errorf("from symbol: %w", err)
	}
	to, err := resolveEntry(options.To, symbols)
	if err != nil {
		return PathResult{}, fmt.Errorf("to symbol: %w", err)
	}

	symbolByID := make(map[string]Symbol, len(symbols))
	for _, symbol := range symbols {
		symbolByID[symbol.ID] = symbol
	}

	result := PathResult{
		From:  from,
		To:    to,
		Paths: make([]SymbolPath, 0),
		Graph: Graph{
			Entry:    from,
			Nodes:    make([]Node, 0),
			Edges:    make([]Edge, 0),
			Warnings: make([]Warning, 0),
		},
		Warnings: make([]Warning, 0),
	}

	if !pathSymbolAllowed(symbolByID[from], options.PackagePrefixes) || !pathSymbolAllowed(symbolByID[to], options.PackagePrefixes) {
		addPathNotFoundWarning(&result, options.MaxDepth)
		return result, nil
	}

	adjacency := make(map[string][]Call)
	buildOptions := BuildOptions{
		ShowExternal:    options.ShowExternal,
		ShowUnresolved:  options.ShowUnresolved,
		ShowInterface:   options.ShowInterface,
		ExpandInterface: options.ExpandInterface,
	}
	for _, call := range calls {
		if !includeCall(call, buildOptions) {
			continue
		}
		toSymbol, ok := symbolByID[call.To]
		if !ok {
			continue
		}
		fromSymbol, ok := symbolByID[call.From]
		if !ok {
			continue
		}
		if !pathSymbolAllowed(fromSymbol, options.PackagePrefixes) || !pathSymbolAllowed(toSymbol, options.PackagePrefixes) {
			continue
		}
		adjacency[call.From] = append(adjacency[call.From], call)
	}
	sortCallMap(adjacency, false)

	if from == to {
		result.Paths = append(result.Paths, SymbolPath{Nodes: []string{from}, Edges: []string{}})
		result.Graph.Nodes = append(result.Graph.Nodes, nodeForSymbol(symbolByID[from]))
		return result, nil
	}

	queue := []pathState{{nodes: []string{from}, calls: []Call{}}}
	foundCalls := make([][]Call, 0)
	for len(queue) > 0 && len(foundCalls) < limit {
		current := queue[0]
		queue = queue[1:]
		if len(current.calls) >= options.MaxDepth {
			continue
		}
		last := current.nodes[len(current.nodes)-1]
		for _, call := range adjacency[last] {
			if containsNode(current.nodes, call.To) {
				continue
			}
			nextNodes := appendCopy(current.nodes, call.To)
			nextCalls := appendCopyCall(current.calls, call)
			if call.To == to {
				foundCalls = append(foundCalls, nextCalls)
				if len(foundCalls) >= limit {
					break
				}
				continue
			}
			queue = append(queue, pathState{nodes: nextNodes, calls: nextCalls})
		}
	}

	if len(foundCalls) == 0 {
		addPathNotFoundWarning(&result, options.MaxDepth)
		return result, nil
	}

	buildPathGraph(&result, foundCalls, symbolByID)
	return result, nil
}

func buildPathGraph(result *PathResult, paths [][]Call, symbols map[string]Symbol) {
	nodeSet := make(map[string]struct{})
	edgeByKey := make(map[string]string)
	for _, calls := range paths {
		path := SymbolPath{
			Nodes: make([]string, 0, len(calls)+1),
			Edges: make([]string, 0, len(calls)),
		}
		if len(calls) > 0 {
			path.Nodes = append(path.Nodes, calls[0].From)
			addNode(&result.Graph, nodeSet, nodeForSymbol(symbols[calls[0].From]))
		}
		for _, call := range calls {
			path.Nodes = append(path.Nodes, call.To)
			addNode(&result.Graph, nodeSet, nodeForSymbol(symbols[call.To]))
			key := callKey(call)
			edgeID, ok := edgeByKey[key]
			if !ok {
				edge := edgeForCall(call, len(result.Graph.Edges)+1)
				edgeByKey[key] = edge.ID
				edgeID = edge.ID
				result.Graph.Edges = append(result.Graph.Edges, edge)
			}
			path.Edges = append(path.Edges, edgeID)
		}
		result.Paths = append(result.Paths, path)
	}

	sort.Slice(result.Graph.Nodes, func(i, j int) bool {
		return result.Graph.Nodes[i].ID < result.Graph.Nodes[j].ID
	})
	sort.Slice(result.Graph.Edges, func(i, j int) bool {
		return result.Graph.Edges[i].ID < result.Graph.Edges[j].ID
	})
}

func addPathNotFoundWarning(result *PathResult, maxDepth int) {
	warning := Warning{
		Code:    "path-not-found",
		Message: fmt.Sprintf("no path found within max_depth %d", maxDepth),
	}
	result.Warnings = append(result.Warnings, warning)
	result.Graph.Warnings = append(result.Graph.Warnings, warning)
}

func pathSymbolAllowed(symbol Symbol, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	return packageMatches(symbol.Package, prefixes)
}

func containsNode(nodes []string, node string) bool {
	for _, current := range nodes {
		if current == node {
			return true
		}
	}
	return false
}

func appendCopy(nodes []string, node string) []string {
	next := make([]string, 0, len(nodes)+1)
	next = append(next, nodes...)
	next = append(next, node)
	return next
}

func appendCopyCall(calls []Call, call Call) []Call {
	next := make([]Call, 0, len(calls)+1)
	next = append(next, calls...)
	next = append(next, call)
	return next
}

func callKey(call Call) string {
	return call.From + "\x00" + call.To + "\x00" + call.Kind + "\x00" + string(call.Resolution) + "\x00" + fmt.Sprintf("%t", call.Candidate) + "\x00" + call.Callsite.File + fmt.Sprintf(":%d:%d", call.Callsite.Line, call.Callsite.Column)
}
