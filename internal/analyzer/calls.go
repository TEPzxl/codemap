package analyzer

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"github.com/tepzxl/codemap/internal/graph"
	"golang.org/x/tools/go/packages"
)

const CallKind = "call"

type Call struct {
	From       string               `json:"from"`
	To         string               `json:"to"`
	Kind       string               `json:"kind"`
	Resolution graph.EdgeResolution `json:"resolution"`
	Callsite   graph.Callsite       `json:"callsite"`
}

func ExtractCalls(loadResult LoadResult, symbols []Symbol) ([]Call, error) {
	symbolIDs := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolIDs[symbol.ID] = struct{}{}
	}

	calls := make([]Call, 0)
	for _, pkg := range loadResult.loadedPackages {
		if pkg == nil {
			continue
		}
		pkgCalls, err := extractPackageCalls(loadResult.rootPath, pkg, symbolIDs)
		if err != nil {
			return nil, err
		}
		calls = append(calls, pkgCalls...)
	}

	sort.Slice(calls, func(i, j int) bool {
		if calls[i].From != calls[j].From {
			return calls[i].From < calls[j].From
		}
		if calls[i].Callsite.File != calls[j].Callsite.File {
			return calls[i].Callsite.File < calls[j].Callsite.File
		}
		if calls[i].Callsite.Line != calls[j].Callsite.Line {
			return calls[i].Callsite.Line < calls[j].Callsite.Line
		}
		if calls[i].Callsite.Column != calls[j].Callsite.Column {
			return calls[i].Callsite.Column < calls[j].Callsite.Column
		}
		return calls[i].To < calls[j].To
	})
	return calls, nil
}

func extractPackageCalls(rootPath string, pkg *packages.Package, symbolIDs map[string]struct{}) ([]Call, error) {
	calls := make([]Call, 0)

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

			from, err := funcDeclSymbolID(pkg.PkgPath, fn)
			if err != nil {
				return nil, err
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				callsite := pkg.Fset.Position(callExpr.Lparen)
				relFile, relErr := relativeFile(rootPath, callsite.Filename)
				if relErr != nil {
					calls = append(calls, unresolvedCall(from, "unresolved", graph.Callsite{
						File:   callsite.Filename,
						Line:   callsite.Line,
						Column: callsite.Column,
					}))
					return true
				}

				to, resolution := resolveCallTarget(pkg, callExpr.Fun, symbolIDs)
				calls = append(calls, Call{
					From:       from,
					To:         to,
					Kind:       CallKind,
					Resolution: resolution,
					Callsite: graph.Callsite{
						File:   relFile,
						Line:   callsite.Line,
						Column: callsite.Column,
					},
				})
				return true
			})
		}
	}

	return calls, nil
}

func resolveCallTarget(pkg *packages.Package, fun ast.Expr, symbolIDs map[string]struct{}) (string, graph.EdgeResolution) {
	switch expr := fun.(type) {
	case *ast.Ident:
		obj := pkg.TypesInfo.Uses[expr]
		fn, ok := obj.(*types.Func)
		if !ok {
			return expr.Name, graph.EdgeResolutionUnresolved
		}
		return resolveFuncObject(fn, nil, symbolIDs)
	case *ast.SelectorExpr:
		if selection := pkg.TypesInfo.Selections[expr]; selection != nil {
			fn, ok := selection.Obj().(*types.Func)
			if !ok {
				return expr.Sel.Name, graph.EdgeResolutionUnresolved
			}
			if isInterfaceReceiver(selection.Recv()) {
				return methodID(fn, selection.Recv()), graph.EdgeResolutionInterface
			}
			return resolveFuncObject(fn, selection.Recv(), symbolIDs)
		}

		obj := pkg.TypesInfo.Uses[expr.Sel]
		fn, ok := obj.(*types.Func)
		if !ok {
			return expr.Sel.Name, graph.EdgeResolutionUnresolved
		}
		return resolveFuncObject(fn, nil, symbolIDs)
	default:
		return "unresolved", graph.EdgeResolutionUnresolved
	}
}

func resolveFuncObject(fn *types.Func, recv types.Type, symbolIDs map[string]struct{}) (string, graph.EdgeResolution) {
	if fn == nil || fn.Pkg() == nil {
		name := "unresolved"
		if fn != nil {
			name = fn.Name()
		}
		return name, graph.EdgeResolutionUnresolved
	}

	id := funcObjectID(fn, recv)
	if _, ok := symbolIDs[id]; ok {
		return id, graph.EdgeResolutionResolved
	}
	return id, graph.EdgeResolutionExternal
}

func funcObjectID(fn *types.Func, recv types.Type) string {
	sig, _ := fn.Type().(*types.Signature)
	if recv == nil && sig != nil && sig.Recv() != nil {
		recv = sig.Recv().Type()
	}
	if recv != nil {
		return methodID(fn, recv)
	}
	return fn.Pkg().Path() + "." + fn.Name()
}

func methodID(fn *types.Func, recv types.Type) string {
	return fn.Pkg().Path() + "." + receiverTypeID(recv) + "." + fn.Name()
}

func funcDeclSymbolID(pkgPath string, fn *ast.FuncDecl) (string, error) {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return pkgPath + "." + fn.Name.Name, nil
	}

	_, receiverID, err := receiverName(fn.Recv.List[0].Type)
	if err != nil {
		return "", fmt.Errorf("parse receiver for %s.%s: %w", pkgPath, fn.Name.Name, err)
	}
	return pkgPath + "." + receiverID + "." + fn.Name.Name, nil
}

func receiverTypeID(typ types.Type) string {
	typ = types.Unalias(typ)
	switch t := typ.(type) {
	case *types.Pointer:
		return "(*" + receiverTypeID(t.Elem()) + ")"
	case *types.Named:
		return t.Obj().Name()
	default:
		return sanitizeTypeName(types.TypeString(typ, func(pkg *types.Package) string {
			return pkg.Name()
		}))
	}
}

func sanitizeTypeName(name string) string {
	name = strings.TrimPrefix(name, "*")
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

func isInterfaceReceiver(typ types.Type) bool {
	typ = types.Unalias(typ)
	if named, ok := typ.(*types.Named); ok {
		typ = named.Underlying()
	}
	_, ok := typ.Underlying().(*types.Interface)
	return ok
}

func unresolvedCall(from string, to string, callsite graph.Callsite) Call {
	return Call{
		From:       from,
		To:         to,
		Kind:       CallKind,
		Resolution: graph.EdgeResolutionUnresolved,
		Callsite:   callsite,
	}
}
