package codegen

import (
	"sort"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

func (g *Generator) cTypePrefix(t types.NRType) string {
	for {
		if pt, ok := t.(*types.PointerType); ok {
			t = pt.Base
			continue
		}
		break
	}
	if t == nil {
		return "void"
	}
	name := g.mangledTypeName(t)
	// Fallback: strip lease prefixes if they still exist for some reason
	name = strings.TrimPrefix(name, "@")
	name = strings.TrimPrefix(name, "#")
	name = strings.TrimPrefix(name, "&")
	return name
}

func (g *Generator) getSymbolPackage(sym *semantic.Symbol) string {
	if sym == nil || sym.DefScope == nil {
		return ""
	}
	s := sym.DefScope
	for s != nil {
		if s.Kind == semantic.ScopePackage {
			return s.PackageName
		}
		s = s.Parent
	}
	return ""
}

func (g *Generator) getPackageName(file *ast.File) string {
	if file == nil {
		return "main"
	}
	for _, stmt := range file.Statements {
		if pkg, ok := stmt.(*ast.PackageStatement); ok && pkg != nil && pkg.Name != nil {
			return pkg.Name.Value
		}
	}
	return "main"
}

func (g *Generator) mangleName(sym *semantic.Symbol) string {
	if sym == nil {
		return ""
	}

	if sym.Kind == semantic.SymType && sym.Type != nil {
		if erased := g.getErasedTypeName(sym.Type); erased != "" {
			return erased
		}
	}

	if sym.Kind == semantic.SymFunc && sym.Type != nil {
		if ft, ok := sym.Type.(*types.FunctionType); ok {
			if ft.Receiver != nil {
				erasedRec := g.getErasedTypeName(ft.Receiver)
				if erasedRec != "" {
					baseName := sym.Name
					if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok {
						baseName = fnStmt.Name.Value
					}

					methodBase := baseName
					if len(methodBase) > 9 && methodBase[len(methodBase)-9] == '_' {
						isHex := true
						for _, c := range methodBase[len(methodBase)-8:] {
							if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
								isHex = false
								break
							}
						}
						if isHex {
							methodBase = methodBase[:len(methodBase)-9]
						}
					}
					return erasedRec + "_" + methodBase
				}
			}
		}
	}

	pkg := g.getSymbolPackage(sym)
	if strings.HasPrefix(sym.Name, "nr_lambda_") {
		return sym.Name
	}
	if sym.Name == "main" && (pkg == "" || pkg == "main") && sym.Kind == semantic.SymFunc {
		return "nr_main"
	}

	// 0. Extern check: do not mangle FFI functions
	if sym.Kind == semantic.SymFunc && sym.DefNode != nil {
		if fn, ok := sym.DefNode.(*ast.FunctionStatement); ok && fn.IsExtern {
			return sym.Name
		}
	}

	// 1. Check for Monomorphized name if it's a generic instance
	// (This would need access to SemanticInfo.MonomorphizedNames, which I'll check later)

	// 2. If it's a local variable or parameter, don't mangle
	if sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam {
		if sym.DefScope != nil && (sym.DefScope.Kind == semantic.ScopeBlock || sym.DefScope.Kind == semantic.ScopeFunction || sym.DefScope.Kind == semantic.ScopeLoop || sym.DefScope.Kind == semantic.ScopeClosure || sym.DefScope.Kind == semantic.ScopeSpawn) {
			return sym.Name
		}
	}

	if pkg == "" || pkg == "main" {
		return sym.Name
	}

	// Replace / and . with _ for C compatibility
	safePkg := strings.ReplaceAll(pkg, "/", "_")
	safePkg = strings.ReplaceAll(safePkg, ".", "_")

	if strings.HasPrefix(sym.Name, safePkg+"_") {
		return sym.Name
	}
	return safePkg + "_" + sym.Name
}

func (g *Generator) sortedVariantNames(st *types.SumType) []string {
	names := make([]string, 0, len(st.Variants))
	for name := range st.Variants {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
func (g *Generator) sortedFunctionNames() []string {
	names := make([]string, 0, len(g.Functions))
	for name := range g.Functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}



func (g *Generator) getErasedTypeName(t types.NRType) string {
	if t == nil {
		return ""
	}

	// Unwrap pointer/lease wrapper types
	underlying := t
	for {
		if pt, ok := underlying.(*types.PointerType); ok {
			underlying = pt.Base
			continue
		}
		break
	}

	if st, ok := underlying.(*types.StructType); ok && st.BaseType != nil && len(st.TypeArgs) > 0 {
		allPointerLike := true
		for _, arg := range st.TypeArgs {
			if !types.IsPointerLike(arg) || types.IsOwnedType(arg) {
				allPointerLike = false
				break
			}
		}
		if allPointerLike {
			hashSuffix := types.GetHashSuffix(st.BaseType.TypeName, st.TypeArgs)
			suffix := "_" + hashSuffix
			if strings.HasSuffix(st.TypeName, suffix) {
				baseMangled := strings.TrimSuffix(st.TypeName, suffix)
				return baseMangled + "_ptr"
			}
			return st.TypeName
		}
	}
	if sumT, ok := underlying.(*types.SumType); ok && sumT.BaseType != nil && len(sumT.TypeArgs) > 0 {
		allPointerLike := true
		for _, arg := range sumT.TypeArgs {
			if !types.IsPointerLike(arg) || types.IsOwnedType(arg) {
				allPointerLike = false
				break
			}
		}
		if allPointerLike {
			hashSuffix := types.GetHashSuffix(sumT.BaseType.TypeName, sumT.TypeArgs)
			suffix := "_" + hashSuffix
			if strings.HasSuffix(sumT.TypeName, suffix) {
				baseMangled := strings.TrimSuffix(sumT.TypeName, suffix)
				return baseMangled + "_ptr"
			}
			return sumT.TypeName
		}
	}
	return ""
}

func sanitizeCIdentifier(name string) string {
	name = strings.ReplaceAll(name, "@", "")
	name = strings.ReplaceAll(name, "#", "")
	name = strings.ReplaceAll(name, "&", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")
	name = strings.ReplaceAll(name, "*", "")
	return name
}

func (g *Generator) mangledTypeName(t types.NRType) string {
	if t == nil {
		return "void"
	}

	if erased := g.getErasedTypeName(t); erased != "" {
		return erased
	}

	// 1. Check Structs
	for mangled, st := range g.Structs {
		if st == t {
			return mangled
		}
	}

	if st, ok := t.(*types.StructType); ok && st.BaseType != nil {
		for mangled, base := range g.Structs {
			if base == st.BaseType {
				return mangled
			}
		}
	}

	// 2. Check SumTypes
	for mangled, sum := range g.SumTypes {
		if sum == t {
			return mangled
		}
	}

	if sumT, ok := t.(*types.SumType); ok && sumT.BaseType != nil {
		for mangled, base := range g.SumTypes {
			if base == sumT.BaseType {
				return mangled
			}
		}
	}

	// 3. Check Protocols
	for mangled, proto := range g.Protocols {
		if proto == t {
			return mangled
		}
	}

	if pt, ok := t.(*types.ProtocolType); ok && pt.BaseType != nil {
		for mangled, base := range g.Protocols {
			if base == pt.BaseType {
				return mangled
			}
		}
	}

	return t.Name()
}
