package analyzer

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	SymbolKindFunction = "function"
	SymbolKindMethod   = "method"
)

type Symbol struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	Package   string `json:"package"`
	Receiver  string `json:"receiver,omitempty"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

func ExtractSymbols(loadResult LoadResult) ([]Symbol, error) {
	symbols := make([]Symbol, 0)

	for _, pkg := range loadResult.loadedPackages {
		if pkg == nil {
			continue
		}
		pkgSymbols, err := extractPackageSymbols(loadResult.rootPath, pkg)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, pkgSymbols...)
	}

	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].ID < symbols[j].ID
	})
	return symbols, nil
}

func extractPackageSymbols(rootPath string, pkg *packages.Package) ([]Symbol, error) {
	symbols := make([]Symbol, 0)

	for _, file := range pkg.Syntax {
		if file == nil {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			start := pkg.Fset.Position(fn.Pos())
			end := pkg.Fset.Position(fn.End())
			if strings.HasSuffix(start.Filename, "_test.go") {
				continue
			}

			relFile, err := relativeFile(rootPath, start.Filename)
			if err != nil {
				return nil, err
			}

			symbol := Symbol{
				Name:      fn.Name.Name,
				Package:   pkg.PkgPath,
				File:      relFile,
				StartLine: start.Line,
				EndLine:   end.Line,
			}

			if fn.Recv == nil || len(fn.Recv.List) == 0 {
				symbol.Kind = SymbolKindFunction
				symbol.ID = pkg.PkgPath + "." + fn.Name.Name
				symbol.Label = fn.Name.Name
			} else {
				receiver, receiverID, err := receiverName(fn.Recv.List[0].Type)
				if err != nil {
					return nil, fmt.Errorf("parse receiver for %s.%s: %w", pkg.PkgPath, fn.Name.Name, err)
				}
				symbol.Kind = SymbolKindMethod
				symbol.Receiver = receiver
				symbol.ID = pkg.PkgPath + "." + receiverID + "." + fn.Name.Name
				symbol.Label = receiverLabel(receiver) + "." + fn.Name.Name
			}

			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

func relativeFile(rootPath string, filename string) (string, error) {
	if rootPath == "" {
		return filepath.ToSlash(filename), nil
	}

	rel, err := filepath.Rel(rootPath, filename)
	if err != nil {
		return "", fmt.Errorf("resolve relative file %q: %w", filename, err)
	}
	return filepath.ToSlash(rel), nil
}

func receiverName(expr ast.Expr) (receiver string, receiverID string, err error) {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name, typed.Name, nil
	case *ast.StarExpr:
		inner, _, err := receiverName(typed.X)
		if err != nil {
			return "", "", err
		}
		return "*" + inner, "(*" + inner + ")", nil
	case *ast.IndexExpr:
		base, _, err := receiverName(typed.X)
		if err != nil {
			return "", "", err
		}
		return base, base, nil
	case *ast.IndexListExpr:
		base, _, err := receiverName(typed.X)
		if err != nil {
			return "", "", err
		}
		return base, base, nil
	default:
		return "", "", fmt.Errorf("unsupported receiver expression %T", expr)
	}
}

func receiverLabel(receiver string) string {
	return strings.TrimPrefix(receiver, "*")
}
