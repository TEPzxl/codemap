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
	Candidate  bool                 `json:"candidate,omitempty"`
}

type CallOptions struct {
	ExpandInterface bool
}

func ExtractCalls(loadResult LoadResult, symbols []Symbol) ([]Call, error) {
	return ExtractCallsWithOptions(loadResult, symbols, CallOptions{})
}

func ExtractCallsWithOptions(loadResult LoadResult, symbols []Symbol, options CallOptions) ([]Call, error) {
	symbolIDs := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolIDs[symbol.ID] = struct{}{}
	}

	var candidates *interfaceCandidateIndex
	if options.ExpandInterface {
		candidates = newInterfaceCandidateIndex(loadResult, symbolIDs)
	}

	calls := make([]Call, 0)
	for _, pkg := range loadResult.loadedPackages {
		if pkg == nil {
			continue
		}
		pkgCalls, err := extractPackageCalls(loadResult.rootPath, pkg, symbolIDs, candidates)
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

func extractPackageCalls(rootPath string, pkg *packages.Package, symbolIDs map[string]struct{}, candidates *interfaceCandidateIndex) ([]Call, error) {
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

				target := resolveCallTarget(pkg, callExpr.Fun, symbolIDs)
				call := Call{
					From:       from,
					To:         target.To,
					Kind:       CallKind,
					Resolution: target.Resolution,
					Callsite: graph.Callsite{
						File:   relFile,
						Line:   callsite.Line,
						Column: callsite.Column,
					},
				}
				calls = append(calls, call)
				if candidates != nil && target.Resolution == graph.EdgeResolutionInterface {
					calls = append(calls, candidates.callsFor(call, target)...)
				}
				return true
			})
		}
	}

	return calls, nil
}

type callTarget struct {
	To              string
	Resolution      graph.EdgeResolution
	InterfaceMethod *types.Func
	InterfaceType   types.Type
}

func resolveCallTarget(pkg *packages.Package, fun ast.Expr, symbolIDs map[string]struct{}) callTarget {
	switch expr := fun.(type) {
	case *ast.Ident:
		obj := pkg.TypesInfo.Uses[expr]
		fn, ok := obj.(*types.Func)
		if !ok {
			return callTarget{To: expr.Name, Resolution: graph.EdgeResolutionUnresolved}
		}
		to, resolution := resolveFuncObject(fn, nil, symbolIDs)
		return callTarget{To: to, Resolution: resolution}
	case *ast.SelectorExpr:
		if selection := pkg.TypesInfo.Selections[expr]; selection != nil {
			fn, ok := selection.Obj().(*types.Func)
			if !ok {
				return callTarget{To: expr.Sel.Name, Resolution: graph.EdgeResolutionUnresolved}
			}
			if isInterfaceReceiver(selection.Recv()) {
				return resolveInterfaceMethodCallTarget(fn, selection.Recv(), symbolIDs)
			}
			to, resolution := resolveFuncObject(fn, selection.Recv(), symbolIDs)
			return callTarget{To: to, Resolution: resolution}
		}

		obj := pkg.TypesInfo.Uses[expr.Sel]
		fn, ok := obj.(*types.Func)
		if !ok {
			return callTarget{To: expr.Sel.Name, Resolution: graph.EdgeResolutionUnresolved}
		}
		to, resolution := resolveFuncObject(fn, nil, symbolIDs)
		return callTarget{To: to, Resolution: resolution}
	default:
		return callTarget{To: "unresolved", Resolution: graph.EdgeResolutionUnresolved}
	}
}

type interfaceCandidateIndex struct {
	symbolIDs map[string]struct{}
	types     []types.Type
}

func newInterfaceCandidateIndex(loadResult LoadResult, symbolIDs map[string]struct{}) *interfaceCandidateIndex {
	index := &interfaceCandidateIndex{
		symbolIDs: symbolIDs,
		types:     make([]types.Type, 0),
	}
	for _, pkg := range loadResult.loadedPackages {
		if pkg == nil || pkg.Types == nil || pkg.Types.Scope() == nil {
			continue
		}
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			typeName, ok := scope.Lookup(name).(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := types.Unalias(typeName.Type()).(*types.Named)
			if !ok || isInterfaceReceiver(named) {
				continue
			}
			index.types = append(index.types, named)
		}
	}
	return index
}

func (i *interfaceCandidateIndex) callsFor(original Call, target callTarget) []Call {
	if target.InterfaceMethod == nil || target.InterfaceType == nil {
		return nil
	}
	iface := interfaceType(target.InterfaceType)
	if iface == nil {
		return nil
	}

	result := make([]Call, 0)
	seen := make(map[string]struct{})
	for _, typ := range i.types {
		i.addImplementationCall(&result, seen, original, iface, target.InterfaceMethod, typ)
		i.addImplementationCall(&result, seen, original, iface, target.InterfaceMethod, types.NewPointer(typ))
	}
	sort.Slice(result, func(a, b int) bool {
		return result[a].To < result[b].To
	})
	return result
}

func (i *interfaceCandidateIndex) addImplementationCall(result *[]Call, seen map[string]struct{}, original Call, iface *types.Interface, interfaceMethod *types.Func, typ types.Type) {
	if !types.Implements(typ, iface) {
		return
	}
	method := matchingMethod(typ, interfaceMethod)
	if method == nil {
		return
	}

	id, resolution := resolveFuncObject(method, nil, i.symbolIDs)
	if resolution != graph.EdgeResolutionResolved {
		return
	}
	if _, ok := seen[id]; ok {
		return
	}
	seen[id] = struct{}{}

	*result = append(*result, Call{
		From:       original.From,
		To:         id,
		Kind:       original.Kind,
		Resolution: graph.EdgeResolutionInterface,
		Callsite:   original.Callsite,
		Candidate:  true,
	})
}

func interfaceType(typ types.Type) *types.Interface {
	typ = types.Unalias(typ)
	if named, ok := typ.(*types.Named); ok {
		typ = named.Underlying()
	}
	iface, ok := typ.Underlying().(*types.Interface)
	if !ok {
		return nil
	}
	return iface.Complete()
}

func matchingMethod(typ types.Type, interfaceMethod *types.Func) *types.Func {
	methodSet := types.NewMethodSet(typ)
	for idx := 0; idx < methodSet.Len(); idx++ {
		method, ok := methodSet.At(idx).Obj().(*types.Func)
		if !ok || method.Name() != interfaceMethod.Name() {
			continue
		}
		if signaturesMatch(method, interfaceMethod) {
			return method
		}
	}
	return nil
}

func signaturesMatch(candidate *types.Func, interfaceMethod *types.Func) bool {
	candidateSig, candidateOK := candidate.Type().(*types.Signature)
	interfaceSig, interfaceOK := interfaceMethod.Type().(*types.Signature)
	if !candidateOK || !interfaceOK {
		return false
	}
	return types.IdenticalIgnoreTags(signatureWithoutReceiver(candidateSig), signatureWithoutReceiver(interfaceSig))
}

func signatureWithoutReceiver(sig *types.Signature) *types.Signature {
	return types.NewSignatureType(nil, nil, nil, sig.Params(), sig.Results(), sig.Variadic())
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

func resolveInterfaceMethodCallTarget(fn *types.Func, recv types.Type, symbolIDs map[string]struct{}) callTarget {
	to, resolution := resolveFuncObject(fn, recv, symbolIDs)
	if resolution == graph.EdgeResolutionUnresolved {
		return callTarget{To: to, Resolution: graph.EdgeResolutionUnresolved}
	}
	return callTarget{
		To:              to,
		Resolution:      graph.EdgeResolutionInterface,
		InterfaceMethod: fn,
		InterfaceType:   recv,
	}
}

func funcObjectID(fn *types.Func, recv types.Type) string {
	if fn == nil || fn.Pkg() == nil {
		if fn == nil {
			return "unresolved"
		}
		return fn.Name()
	}
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
	if fn == nil || fn.Pkg() == nil {
		name := "unresolved"
		if fn != nil {
			name = fn.Name()
		}
		return name
	}
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
