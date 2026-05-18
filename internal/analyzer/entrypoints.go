package analyzer

import (
	"go/ast"
	"sort"
	"strings"
)

const EntrypointDiscoveryNote = "Entrypoint discovery is heuristic and may miss valid entrypoints or include non-entry utility symbols."

type Entrypoint struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Package string   `json:"package"`
	File    string   `json:"file"`
	Kind    string   `json:"kind"`
	Reasons []string `json:"reasons"`
}

func DiscoverEntrypoints(loadResult LoadResult, symbols []Symbol) ([]Entrypoint, error) {
	goroutineStarters, err := discoverGoroutineStarters(loadResult)
	if err != nil {
		return nil, err
	}

	entrypoints := make([]Entrypoint, 0)
	for _, symbol := range symbols {
		reasons := entrypointReasons(symbol, goroutineStarters[symbol.ID])
		if len(reasons) == 0 {
			continue
		}
		entrypoints = append(entrypoints, Entrypoint{
			ID:      symbol.ID,
			Label:   symbol.Label,
			Package: symbol.Package,
			File:    symbol.File,
			Kind:    symbol.Kind,
			Reasons: reasons,
		})
	}

	sort.Slice(entrypoints, func(i, j int) bool {
		leftScore := entrypointScore(entrypoints[i])
		rightScore := entrypointScore(entrypoints[j])
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		if entrypoints[i].Package != entrypoints[j].Package {
			return entrypoints[i].Package < entrypoints[j].Package
		}
		if entrypoints[i].Label != entrypoints[j].Label {
			return entrypoints[i].Label < entrypoints[j].Label
		}
		return entrypoints[i].ID < entrypoints[j].ID
	})
	return entrypoints, nil
}

func discoverGoroutineStarters(loadResult LoadResult) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, pkg := range loadResult.loadedPackages {
		if pkg == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			if file == nil {
				continue
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				start := pkg.Fset.Position(fn.Pos())
				if strings.HasSuffix(start.Filename, "_test.go") {
					continue
				}
				if !containsGoroutine(fn.Body) {
					continue
				}
				id, err := funcDeclSymbolID(pkg.PkgPath, fn)
				if err != nil {
					return nil, err
				}
				result[id] = true
			}
		}
	}
	return result, nil
}

func containsGoroutine(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if _, ok := node.(*ast.GoStmt); ok {
			found = true
			return false
		}
		return !found
	})
	return found
}

func entrypointReasons(symbol Symbol, containsGo bool) []string {
	reasons := make([]string, 0, 4)
	if symbol.Kind == SymbolKindFunction && symbol.Name == "main" {
		reasons = append(reasons, "main-function")
	}
	if ast.IsExported(symbol.Name) {
		if symbol.Kind == SymbolKindMethod {
			reasons = append(reasons, "exported-method")
		} else {
			reasons = append(reasons, "exported-function")
		}
	}
	if isStarterName(symbol.Name) {
		reasons = append(reasons, "name:"+symbol.Name)
	}
	if strings.Contains(symbol.Name, "Handler") {
		reasons = append(reasons, "name:Handler")
	} else if strings.Contains(symbol.Name, "Handle") {
		reasons = append(reasons, "name:Handle")
	}
	if strings.Contains(strings.TrimPrefix(symbol.Receiver, "*"), "Handler") {
		reasons = append(reasons, "receiver:Handler")
	}
	if containsGo {
		reasons = append(reasons, "contains-goroutine")
	}
	return dedupeStrings(reasons)
}

func isStarterName(name string) bool {
	switch name {
	case "Run", "Start", "Serve", "Listen", "Execute":
		return true
	default:
		return false
	}
}

func entrypointScore(entrypoint Entrypoint) int {
	score := 0
	for _, reason := range entrypoint.Reasons {
		switch {
		case reason == "main-function":
			score += 1000
		case strings.HasPrefix(reason, "name:"):
			score += 300
		case reason == "receiver:Handler":
			score += 250
		case reason == "contains-goroutine":
			score += 200
		case reason == "exported-function" || reason == "exported-method":
			score += 100
		}
	}
	return score
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
