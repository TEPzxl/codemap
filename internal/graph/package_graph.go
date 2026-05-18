package graph

import (
	"fmt"
	"sort"
	"strings"
)

type PackageGraphOptions struct {
	BuildOptions
	ModulePath       string
	IncludeSelfEdges bool
	Warnings         []Warning
}

type PackageGraph struct {
	Nodes    []PackageNode `json:"nodes"`
	Edges    []PackageEdge `json:"edges"`
	Warnings []Warning     `json:"warnings"`
}

type PackageNode struct {
	ID          string `json:"id"`
	Package     string `json:"package"`
	FullPackage string `json:"full_package,omitempty"`
	Symbols     int    `json:"symbols"`
	Calls       int    `json:"calls"`
}

type PackageEdge struct {
	ID    string `json:"id"`
	From  string `json:"from"`
	To    string `json:"to"`
	Calls int    `json:"calls"`
}

func BuildPackageGraph(symbols []Symbol, calls []Call, options PackageGraphOptions) (PackageGraph, error) {
	if options.Depth < 0 {
		return PackageGraph{}, fmt.Errorf("depth must be >= 0")
	}
	if options.NodeLimit < 0 {
		return PackageGraph{}, fmt.Errorf("node limit must be >= 0")
	}
	direction := options.Direction.Normalized()
	if !direction.IsValid() {
		return PackageGraph{}, fmt.Errorf("direction must be one of downstream, upstream, both")
	}

	symbolByID := make(map[string]Symbol, len(symbols))
	for _, symbol := range symbols {
		symbolByID[symbol.ID] = symbol
	}

	warnings := append([]Warning{}, options.Warnings...)
	if strings.TrimSpace(options.Entry) != "" {
		graphOptions := options.BuildOptions
		graphOptions.PackagePrefixes = nil
		graph, err := BuildGraph(symbols, calls, graphOptions)
		if err != nil {
			return PackageGraph{}, err
		}
		warnings = append(warnings, graph.Warnings...)
		return aggregatePackageGraph(graph.Nodes, graph.Edges, options, warnings), nil
	}

	nodes := make([]Node, 0, len(symbols))
	for _, symbol := range symbols {
		nodes = append(nodes, nodeForSymbol(symbol))
	}

	edges := make([]Edge, 0, len(calls))
	for _, call := range calls {
		if !includeCall(call, options.BuildOptions) {
			continue
		}
		fromSymbol, ok := symbolByID[call.From]
		if !ok {
			continue
		}
		nodes = append(nodes, nodeForCallTarget(call, symbolByID))
		edges = append(edges, Edge{
			ID:         edgeForCall(call, len(edges)+1).ID,
			From:       fromSymbol.ID,
			To:         call.To,
			Kind:       call.Kind,
			Resolution: call.Resolution,
			Callsite:   call.Callsite,
			Candidate:  call.Candidate,
		})
	}

	return aggregatePackageGraph(nodes, edges, options, warnings), nil
}

func aggregatePackageGraph(nodes []Node, edges []Edge, options PackageGraphOptions, warnings []Warning) PackageGraph {
	aggregator := newPackageAggregator(options)

	nodeByID := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		nodeByID[node.ID] = node
	}

	includedEdges := make([]Edge, 0, len(edges))
	includedPackages := make(map[string]struct{})
	for _, edge := range edges {
		fromNode, fromOK := nodeByID[edge.From]
		toNode, toOK := nodeByID[edge.To]
		if !fromOK || !toOK {
			continue
		}
		if len(options.PackagePrefixes) > 0 && !packageMatches(fromNode.Package, options.PackagePrefixes) && !packageMatches(toNode.Package, options.PackagePrefixes) {
			continue
		}
		includedEdges = append(includedEdges, edge)
		includedPackages[fromNode.Package] = struct{}{}
		includedPackages[toNode.Package] = struct{}{}
	}

	for _, node := range nodeByID {
		if len(options.PackagePrefixes) > 0 && !packageMatches(node.Package, options.PackagePrefixes) {
			if _, ok := includedPackages[node.Package]; !ok {
				continue
			}
		}
		aggregator.addSymbol(node.Package)
	}

	for _, edge := range includedEdges {
		fromNode := nodeByID[edge.From]
		toNode := nodeByID[edge.To]
		aggregator.addCall(fromNode.Package, toNode.Package)
	}

	result := PackageGraph{
		Nodes:    aggregator.sortedNodes(),
		Edges:    aggregator.sortedEdges(),
		Warnings: warnings,
	}
	return result
}

type packageAggregator struct {
	options   PackageGraphOptions
	nodeByID  map[string]*PackageNode
	edgeByKey map[string]*PackageEdge
}

func newPackageAggregator(options PackageGraphOptions) *packageAggregator {
	return &packageAggregator{
		options:   options,
		nodeByID:  make(map[string]*PackageNode),
		edgeByKey: make(map[string]*PackageEdge),
	}
}

func (a *packageAggregator) addSymbol(pkg string) {
	node := a.ensureNode(pkg)
	node.Symbols++
}

func (a *packageAggregator) addCall(fromPackage string, toPackage string) {
	from := a.ensureNode(fromPackage)
	to := a.ensureNode(toPackage)
	if from.ID == to.ID {
		from.Calls++
		if !a.options.IncludeSelfEdges {
			return
		}
	} else {
		from.Calls++
		to.Calls++
	}

	key := from.ID + "\x00" + to.ID
	edge, ok := a.edgeByKey[key]
	if !ok {
		edge = &PackageEdge{
			ID:   "pkg-edge-" + edgeIDPart(from.ID) + "-to-" + edgeIDPart(to.ID),
			From: from.ID,
			To:   to.ID,
		}
		a.edgeByKey[key] = edge
	}
	edge.Calls++
}

func (a *packageAggregator) ensureNode(pkg string) *PackageNode {
	id := relativePackagePath(pkg, a.options.ModulePath)
	node, ok := a.nodeByID[id]
	if ok {
		if node.FullPackage == "" && pkg != "" {
			node.FullPackage = pkg
		}
		return node
	}
	node = &PackageNode{
		ID:          id,
		Package:     id,
		FullPackage: pkg,
	}
	a.nodeByID[id] = node
	return node
}

func (a *packageAggregator) sortedNodes() []PackageNode {
	nodes := make([]PackageNode, 0, len(a.nodeByID))
	for _, node := range a.nodeByID {
		nodes = append(nodes, *node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	return nodes
}

func (a *packageAggregator) sortedEdges() []PackageEdge {
	edges := make([]PackageEdge, 0, len(a.edgeByKey))
	for _, edge := range a.edgeByKey {
		edges = append(edges, *edge)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	return edges
}

func relativePackagePath(pkg string, modulePath string) string {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return "(unknown)"
	}
	modulePath = strings.TrimSuffix(strings.TrimSpace(modulePath), "/")
	if modulePath == "" {
		return pkg
	}
	if pkg == modulePath {
		return "."
	}
	prefix := modulePath + "/"
	if strings.HasPrefix(pkg, prefix) {
		return strings.TrimPrefix(pkg, prefix)
	}
	return pkg
}

func edgeIDPart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	if value == "." {
		return "root"
	}
	replacer := strings.NewReplacer("/", "-", ".", "-", "*", "ptr", "(", "", ")", "", "[", "", "]", "", " ", "-")
	return replacer.Replace(value)
}
