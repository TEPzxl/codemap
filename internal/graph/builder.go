package graph

import (
	"fmt"
	"sort"
	"strings"
)

type Symbol struct {
	ID        string
	Label     string
	Kind      string
	Package   string
	Receiver  string
	File      string
	StartLine int
	EndLine   int
}

type Call struct {
	From       string
	To         string
	Kind       string
	Resolution EdgeResolution
	Callsite   Callsite
	Candidate  bool
}

type BuildOptions struct {
	Entry           string
	Depth           int
	Direction       Direction
	ShowExternal    bool
	ShowUnresolved  bool
	ShowInterface   bool
	ExpandInterface bool
	PackagePrefixes []string
	NodeLimit       int
}

func BuildGraph(symbols []Symbol, calls []Call, options BuildOptions) (Graph, error) {
	symbolByID := make(map[string]Symbol, len(symbols))
	for _, symbol := range symbols {
		symbolByID[symbol.ID] = symbol
	}

	entry, err := resolveEntry(options.Entry, symbols)
	if err != nil {
		return Graph{}, err
	}
	if options.Depth < 0 {
		return Graph{}, fmt.Errorf("depth must be >= 0")
	}
	if options.NodeLimit < 0 {
		return Graph{}, fmt.Errorf("node limit must be >= 0")
	}
	direction := options.Direction.Normalized()
	if !direction.IsValid() {
		return Graph{}, fmt.Errorf("direction must be one of downstream, upstream, both")
	}

	adjacency := make(map[string][]Call)
	reverseAdjacency := make(map[string][]Call)
	for _, call := range calls {
		if !includeCall(call, options) {
			continue
		}
		adjacency[call.From] = append(adjacency[call.From], call)
		reverseAdjacency[call.To] = append(reverseAdjacency[call.To], call)
	}
	sortCallMap(adjacency, false)
	sortCallMap(reverseAdjacency, true)

	result := Graph{
		Entry:    entry,
		Nodes:    make([]Node, 0),
		Edges:    make([]Edge, 0),
		Warnings: make([]Warning, 0),
	}
	nodeSet := make(map[string]struct{})
	edgeSet := make(map[string]struct{})
	expanded := make(map[string]int)

	addNode(&result, nodeSet, nodeForSymbol(symbolByID[entry]))
	switch direction {
	case DirectionDownstream:
		traverseDownstream(entry, 0, options.Depth, adjacency, symbolByID, &result, nodeSet, edgeSet, expanded)
	case DirectionUpstream:
		traverseUpstream(entry, 0, options.Depth, reverseAdjacency, symbolByID, &result, nodeSet, edgeSet, expanded)
	case DirectionBoth:
		traverseDownstream(entry, 0, options.Depth, adjacency, symbolByID, &result, nodeSet, edgeSet, make(map[string]int))
		traverseUpstream(entry, 0, options.Depth, reverseAdjacency, symbolByID, &result, nodeSet, edgeSet, make(map[string]int))
	}

	sort.Slice(result.Nodes, func(i, j int) bool {
		return result.Nodes[i].ID < result.Nodes[j].ID
	})
	sort.Slice(result.Edges, func(i, j int) bool {
		return result.Edges[i].ID < result.Edges[j].ID
	})

	if err := applyGraphFilters(&result, options); err != nil {
		return Graph{}, err
	}

	return result, nil
}

func sortCallMap(calls map[string][]Call, upstream bool) {
	for key := range calls {
		sort.Slice(calls[key], func(i, j int) bool {
			if calls[key][i].Callsite.Line != calls[key][j].Callsite.Line {
				return calls[key][i].Callsite.Line < calls[key][j].Callsite.Line
			}
			if calls[key][i].Callsite.Column != calls[key][j].Callsite.Column {
				return calls[key][i].Callsite.Column < calls[key][j].Callsite.Column
			}
			if upstream && calls[key][i].From != calls[key][j].From {
				return calls[key][i].From < calls[key][j].From
			}
			return calls[key][i].To < calls[key][j].To
		})
	}
}

func resolveEntry(entry string, symbols []Symbol) (string, error) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return "", fmt.Errorf("entry symbol is required")
	}

	for _, symbol := range symbols {
		if symbol.ID == entry {
			return symbol.ID, nil
		}
	}

	matches := make([]string, 0)
	for _, symbol := range symbols {
		if symbol.ID == entry || symbol.Label == entry || strings.HasSuffix(symbol.ID, "."+entry) || shortQuery(symbol) == entry {
			matches = append(matches, symbol.ID)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		sort.Strings(matches)
		return "", fmt.Errorf("entry symbol is ambiguous %q: %s", entry, strings.Join(matches, ", "))
	}
	return "", fmt.Errorf("entry symbol not found: %s", entry)
}

func shortQuery(symbol Symbol) string {
	if symbol.Label == "main" {
		return "main.main"
	}
	if symbol.Receiver == "" {
		return packageName(symbol.Package) + "." + symbol.Label
	}
	return packageName(symbol.Package) + "." + symbol.Label
}

func packageName(pkgPath string) string {
	if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
		return pkgPath[idx+1:]
	}
	return pkgPath
}

func includeCall(call Call, options BuildOptions) bool {
	switch call.Resolution {
	case EdgeResolutionExternal:
		return options.ShowExternal
	case EdgeResolutionUnresolved:
		return options.ShowUnresolved
	case EdgeResolutionInterface:
		if call.Candidate {
			return options.ExpandInterface
		}
		return options.ShowInterface || options.ExpandInterface
	default:
		return true
	}
}

func applyGraphFilters(result *Graph, options BuildOptions) error {
	if len(options.PackagePrefixes) > 0 {
		filterByPackage(result, options.PackagePrefixes)
		if len(result.Nodes) == 0 {
			return fmt.Errorf("package filter removed all nodes")
		}
	}
	if options.NodeLimit > 0 && len(result.Nodes) > options.NodeLimit {
		originalCount := len(result.Nodes)
		result.Nodes = result.Nodes[:options.NodeLimit]
		filterEdgesToNodes(result)
		result.Warnings = append(result.Warnings, Warning{
			Code:    "node-limit-exceeded",
			Message: fmt.Sprintf("node limit %d truncated graph from %d nodes", options.NodeLimit, originalCount),
		})
	}
	return nil
}

func filterByPackage(result *Graph, prefixes []string) {
	nodes := make([]Node, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		if packageMatches(node.Package, prefixes) {
			nodes = append(nodes, node)
		}
	}
	result.Nodes = nodes
	filterEdgesToNodes(result)
}

func packageMatches(pkg string, prefixes []string) bool {
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" && strings.HasPrefix(pkg, prefix) {
			return true
		}
	}
	return false
}

func filterEdgesToNodes(result *Graph) {
	nodeSet := make(map[string]struct{}, len(result.Nodes))
	for _, node := range result.Nodes {
		nodeSet[node.ID] = struct{}{}
	}

	edges := make([]Edge, 0, len(result.Edges))
	for _, edge := range result.Edges {
		if _, ok := nodeSet[edge.From]; !ok {
			continue
		}
		if _, ok := nodeSet[edge.To]; !ok {
			continue
		}
		edges = append(edges, edge)
	}
	result.Edges = edges
}

func traverseDownstream(current string, depth int, maxDepth int, adjacency map[string][]Call, symbols map[string]Symbol, result *Graph, nodeSet map[string]struct{}, edgeSet map[string]struct{}, expanded map[string]int) {
	if depth >= maxDepth {
		return
	}
	if previousDepth, ok := expanded[current]; ok && previousDepth <= depth {
		return
	}
	expanded[current] = depth

	for _, call := range adjacency[current] {
		toNode := nodeForCallTarget(call, symbols)
		addNode(result, nodeSet, toNode)
		addEdge(result, edgeSet, edgeForCall(call, len(result.Edges)+1))

		if _, ok := symbols[call.To]; ok {
			traverseDownstream(call.To, depth+1, maxDepth, adjacency, symbols, result, nodeSet, edgeSet, expanded)
		}
	}
}

func traverseUpstream(current string, depth int, maxDepth int, reverseAdjacency map[string][]Call, symbols map[string]Symbol, result *Graph, nodeSet map[string]struct{}, edgeSet map[string]struct{}, expanded map[string]int) {
	if depth >= maxDepth {
		return
	}
	if previousDepth, ok := expanded[current]; ok && previousDepth <= depth {
		return
	}
	expanded[current] = depth

	for _, call := range reverseAdjacency[current] {
		fromSymbol, ok := symbols[call.From]
		if !ok {
			continue
		}
		addNode(result, nodeSet, nodeForSymbol(fromSymbol))
		addEdge(result, edgeSet, edgeForCall(call, len(result.Edges)+1))
		traverseUpstream(call.From, depth+1, maxDepth, reverseAdjacency, symbols, result, nodeSet, edgeSet, expanded)
		if _, ok := symbols[call.To]; ok {
			addNode(result, nodeSet, nodeForSymbol(symbols[call.To]))
		}
	}
}

func nodeForSymbol(symbol Symbol) Node {
	return Node{
		ID:         symbol.ID,
		Label:      symbol.Label,
		Kind:       NodeKind(symbol.Kind),
		Package:    symbol.Package,
		Receiver:   symbol.Receiver,
		File:       symbol.File,
		StartLine:  symbol.StartLine,
		EndLine:    symbol.EndLine,
		IsExternal: false,
	}
}

func nodeForCallTarget(call Call, symbols map[string]Symbol) Node {
	if symbol, ok := symbols[call.To]; ok {
		return nodeForSymbol(symbol)
	}

	kind := NodeKindUnresolved
	isExternal := false
	if call.Resolution == EdgeResolutionExternal {
		kind = NodeKindExternal
		isExternal = true
	}

	return Node{
		ID:         call.To,
		Label:      shortLabel(call.To),
		Kind:       kind,
		Package:    packageFromTarget(call.To),
		File:       "",
		StartLine:  0,
		EndLine:    0,
		IsExternal: isExternal,
	}
}

func edgeForCall(call Call, index int) Edge {
	return Edge{
		ID:         fmt.Sprintf("edge-%06d", index),
		From:       call.From,
		To:         call.To,
		Kind:       call.Kind,
		Resolution: call.Resolution,
		Callsite:   call.Callsite,
		Candidate:  call.Candidate,
	}
}

func addNode(result *Graph, nodeSet map[string]struct{}, node Node) {
	if _, ok := nodeSet[node.ID]; ok {
		return
	}
	nodeSet[node.ID] = struct{}{}
	result.Nodes = append(result.Nodes, node)
}

func addEdge(result *Graph, edgeSet map[string]struct{}, edge Edge) {
	key := edge.From + "\x00" + edge.To + "\x00" + edge.Kind + "\x00" + string(edge.Resolution) + "\x00" + fmt.Sprintf("%t", edge.Candidate) + "\x00" + edge.Callsite.File + fmt.Sprintf(":%d:%d", edge.Callsite.Line, edge.Callsite.Column)
	if _, ok := edgeSet[key]; ok {
		return
	}
	edgeSet[key] = struct{}{}
	result.Edges = append(result.Edges, edge)
}

func shortLabel(id string) string {
	if idx := strings.LastIndex(id, "."); idx >= 0 && idx < len(id)-1 {
		return id[idx+1:]
	}
	return id
}

func packageFromTarget(id string) string {
	if idx := strings.LastIndex(id, "."); idx > 0 {
		return id[:idx]
	}
	return ""
}
