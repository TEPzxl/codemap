package graph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ExportFormat string

const (
	ExportFormatJSON    ExportFormat = "json"
	ExportFormatMermaid ExportFormat = "mermaid"
	ExportFormatDOT     ExportFormat = "dot"
)

func (f ExportFormat) IsValid() bool {
	switch f {
	case ExportFormatJSON, ExportFormatMermaid, ExportFormatDOT:
		return true
	default:
		return false
	}
}

func ExportGraph(graph Graph, format ExportFormat) (string, string, error) {
	switch format {
	case ExportFormatJSON:
		data, err := json.MarshalIndent(graph, "", "  ")
		if err != nil {
			return "", "", err
		}
		return string(data) + "\n", "application/json", nil
	case ExportFormatMermaid:
		return exportMermaid(graph), "text/plain; charset=utf-8", nil
	case ExportFormatDOT:
		return exportDOT(graph), "text/plain; charset=utf-8", nil
	default:
		return "", "", fmt.Errorf("format must be one of json, mermaid, dot")
	}
}

func exportMermaid(graph Graph) string {
	nodeIDs := stableExportNodeIDs(graph.Nodes)
	var builder strings.Builder
	builder.WriteString("flowchart LR\n")
	for _, node := range sortedExportNodes(graph.Nodes) {
		fmt.Fprintf(&builder, "  %s[\"%s\"]\n", nodeIDs[node.ID], escapeMermaidLabel(node.Label))
	}
	for _, edge := range sortedExportEdges(graph.Edges) {
		fromID, fromOK := nodeIDs[edge.From]
		toID, toOK := nodeIDs[edge.To]
		if !fromOK || !toOK {
			continue
		}
		fmt.Fprintf(&builder, "  %s --> %s\n", fromID, toID)
	}
	return builder.String()
}

func exportDOT(graph Graph) string {
	nodeIDs := stableExportNodeIDs(graph.Nodes)
	var builder strings.Builder
	builder.WriteString("digraph codemap {\n")
	builder.WriteString("  rankdir=LR;\n")
	for _, node := range sortedExportNodes(graph.Nodes) {
		fmt.Fprintf(&builder, "  %s [label=%s];\n", strconv.Quote(nodeIDs[node.ID]), strconv.Quote(node.Label))
	}
	for _, edge := range sortedExportEdges(graph.Edges) {
		fromID, fromOK := nodeIDs[edge.From]
		toID, toOK := nodeIDs[edge.To]
		if !fromOK || !toOK {
			continue
		}
		fmt.Fprintf(&builder, "  %s -> %s [label=%s];\n", strconv.Quote(fromID), strconv.Quote(toID), strconv.Quote(string(edge.Resolution)))
	}
	builder.WriteString("}\n")
	return builder.String()
}

func stableExportNodeIDs(nodes []Node) map[string]string {
	result := make(map[string]string, len(nodes))
	for index, node := range sortedExportNodes(nodes) {
		result[node.ID] = fmt.Sprintf("n%d", index)
	}
	return result
}

func sortedExportNodes(nodes []Node) []Node {
	result := append([]Node(nil), nodes...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func sortedExportEdges(edges []Edge) []Edge {
	result := append([]Edge(nil), edges...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].From != result[j].From {
			return result[i].From < result[j].From
		}
		if result[i].To != result[j].To {
			return result[i].To < result[j].To
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func escapeMermaidLabel(label string) string {
	replacer := strings.NewReplacer(
		"\\", " ",
		"\r", " ",
		"\n", " ",
		"\"", "#quot;",
		"[", "#91;",
		"]", "#93;",
	)
	return replacer.Replace(label)
}
