package graph

import (
	"fmt"
)

type NodeKind string

const (
	NodeKindFunction   NodeKind = "function"
	NodeKindMethod     NodeKind = "method"
	NodeKindExternal   NodeKind = "external"
	NodeKindUnresolved NodeKind = "unresolved"
)

func (k NodeKind) IsValid() bool {
	switch k {
	case NodeKindFunction, NodeKindMethod, NodeKindExternal, NodeKindUnresolved:
		return true
	default:
		return false
	}
}

type EdgeResolution string

const (
	EdgeResolutionResolved   EdgeResolution = "resolved"
	EdgeResolutionInterface  EdgeResolution = "interface"
	EdgeResolutionExternal   EdgeResolution = "external"
	EdgeResolutionUnresolved EdgeResolution = "unresolved"
)

func (r EdgeResolution) IsValid() bool {
	switch r {
	case EdgeResolutionResolved, EdgeResolutionInterface, EdgeResolutionExternal, EdgeResolutionUnresolved:
		return true
	default:
		return false
	}
}

type Direction string

const (
	DirectionDownstream Direction = "downstream"
	DirectionUpstream   Direction = "upstream"
	DirectionBoth       Direction = "both"
)

func (d Direction) Normalized() Direction {
	if d == "" {
		return DirectionDownstream
	}
	return d
}

func (d Direction) IsValid() bool {
	switch d.Normalized() {
	case DirectionDownstream, DirectionUpstream, DirectionBoth:
		return true
	default:
		return false
	}
}

type Graph struct {
	Entry    string    `json:"entry"`
	Nodes    []Node    `json:"nodes"`
	Edges    []Edge    `json:"edges"`
	Warnings []Warning `json:"warnings"`
}

type Node struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Kind       NodeKind `json:"kind"`
	Package    string   `json:"package"`
	Receiver   string   `json:"receiver,omitempty"`
	File       string   `json:"file"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	IsExternal bool     `json:"is_external"`
}

type Edge struct {
	ID         string         `json:"id"`
	From       string         `json:"from"`
	To         string         `json:"to"`
	Kind       string         `json:"kind"`
	Resolution EdgeResolution `json:"resolution"`
	Callsite   Callsite       `json:"callsite"`
	Candidate  bool           `json:"candidate,omitempty"`
}

type Callsite struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
}

func (g Graph) Validate() error {
	nodeSet := make(map[string]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node id is empty")
		}
		if !node.Kind.IsValid() {
			return fmt.Errorf("invalid node kind: %q", node.Kind)
		}
		nodeSet[node.ID] = struct{}{}
	}

	for _, edge := range g.Edges {
		if !edge.Resolution.IsValid() {
			return fmt.Errorf("invalid edge resolution: %q", edge.Resolution)
		}
		if edge.From == "" || edge.To == "" {
			return fmt.Errorf("edge from/to must not be empty")
		}
		_, fromOK := nodeSet[edge.From]
		_, toOK := nodeSet[edge.To]
		if !fromOK || !toOK {
			// external 或 unresolved 策略下允许边指向未加载节点。
			if edge.Resolution != EdgeResolutionExternal && edge.Resolution != EdgeResolutionUnresolved {
				return fmt.Errorf("edge %q references unknown node: from_exists=%t,to_exists=%t", edge.ID, fromOK, toOK)
			}
		}
	}
	return nil
}
