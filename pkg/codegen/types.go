package codegen

import (
	"bytes"
	"strings"

	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/types"
)

func (g *Generator) cType(t types.NRType) string {
	if t == nil {
		return "void"
	}
	res := ""
	switch t := t.(type) {
	case *types.PrimitiveType:
		res = g.mapPrimitive(t.Name())
	case *types.StructType:
		res = g.mangledTypeName(t)
	case *types.ChanType:
		res = "channel_t*"
	case *types.ProtocolType:
		res = g.mangledTypeName(t)
	case *types.SumType:
		res = g.mangledTypeName(t)
	case *types.PointerType:
		if t.Leased && !t.IsArray {
			if !g.shouldPassByPointer(t.Base, t.Kind) {
				return g.cType(t.Base)
			}
		}
		res = g.cType(t.Base) + "*"
	case *types.FunctionType:
		res = "nr_closure_t" // Use fat pointer struct instead of void*
	case *types.ListType:
		res = g.cType(t.ElementType) + "*"
	default:
		res = "void*"
	}
	return res
}

func (g *Generator) cParamType(t types.NRType, l types.LeaseKind) string {
	if l == types.LeaseRead || l == types.LeaseWrite {
		if pt, ok := t.(*types.PointerType); ok && pt.Leased && !pt.IsArray && (pt.Kind == types.LeaseRead || pt.Kind == types.LeaseWrite || pt.Kind == types.LeaseMove) {
			t = pt.Base
		}
	}
	ct := g.cType(t)
	if g.shouldPassByPointer(t, l) {
		return ct + "*"
	}
	return ct
}

func (g *Generator) mapPrimitive(name string) string {
	switch name {
	case "i32":
		return "int"
	case "i64":
		return "long long"
	case "u64":
		return "unsigned long long"
	case "u32":
		return "unsigned int"
	case "u16":
		return "unsigned short"
	case "u8":
		return "unsigned char"
	case "i16":
		return "short"
	case "i8":
		return "char"
	case "byte":
		return "unsigned char"
	case "f32":
		return "float"
	case "f64":
		return "double"
	case "bool":
		return "bool"
	case "str":
		return "char*" // C strings
	case "ptr":
		return "void*" // Raw pointer
	case "fiber":
		return "fiber_t"
	case "void":
		return "void"
	}
	if strings.HasPrefix(name, "chan[") {
		return "channel_t*"
	}
	return "int"
}

func (g *Generator) isHeapAllocated(t types.NRType) bool {
	if t == nil {
		return false
	}
	if pt, ok := t.(*types.PointerType); ok {
		// Owned pointers are heap-allocated:
		// 1. Raw owned pointers (!pt.Leased)
		// 2. Heap-allocated arrays (pt.IsArray)
		// 3. Move-leases (@T) which represent ownership transfer of heap objects
		return !pt.Leased || pt.IsArray || pt.Kind == types.LeaseMove
	}
	if t.GetKind() == types.KindList || t.GetKind() == types.KindChan || t.GetKind() == types.KindMap || t.Name() == "str" || t.Name() == "ptr" {
		return true
	}
	return false
}

func (g *Generator) isPointerInC(e ast.Expression) bool {
	if _, ok := e.(*ast.NoneLiteral); ok {
		return true
	}
	t := g.SemanticInfo.Types[e]
	if ident, ok := e.(*ast.Identifier); ok {
		sym := g.SemanticInfo.Uses[ident]
		if sym == nil {
			sym = g.findSymbolByName(ident.Value)
		}
		if sym != nil {
			t = sym.Type
		} else {
		}
	}
	if t == nil {
		return false
	}
	if _, ok := e.(*ast.IndexExpression); ok {
		return strings.HasSuffix(g.cType(t), "*")
	}

	ut := types.UnwrapLease(t)

	// Parameters are pointers in C if they are struct/sum/protocol types
	if ident, ok := e.(*ast.Identifier); ok {
		sym := g.SemanticInfo.Uses[ident]
		if sym == nil {
			sym = g.findSymbolByName(ident.Value)
		}
		if sym != nil && sym.Kind == semantic.SymParam {
			if _, ok := ut.(*types.StructType); ok || ut.GetKind() == types.KindSum || ut.GetKind() == types.KindProtocol {
				return true
			}
		}
	}

	if pt, ok := t.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
		if !g.isPointerTypeInC(pt.Base) && !g.shouldPassByPointer(pt.Base, pt.Kind) {
			return false
		}
		return true
	}

	if g.isPointerTypeInC(t) {
		return true
	}

	return false
}

func (g *Generator) cPointerLevel(t types.NRType, isValue bool) int {
	if t == nil {
		return 0
	}
	underlying := t
	ptrCount := 0
	for {
		if pt, ok := underlying.(*types.PointerType); ok {
			ptrCount++
			underlying = pt.Base
			continue
		}
		break
	}
	if _, isStruct := underlying.(*types.StructType); isStruct {
		if ptrCount > 0 {
			return 1
		}
		if isValue {
			return 0
		}
		return 1
	}
	if underlying.GetKind() == types.KindSum {
		if ptrCount > 0 {
			return 1
		}
		if isValue {
			return 0
		}
		return 1
	}
	if underlying.GetKind() == types.KindProtocol {
		if ptrCount > 0 {
			return 1
		}
		if isValue {
			return 0
		}
		return 1
	}
	if g.isPointerTypeInC(underlying) {
		return 1
	}
	cTypeStr := g.cType(t)
	return strings.Count(cTypeStr, "*")
}

func (g *Generator) shouldDereferenceInC(e ast.Expression) bool {
	if !g.isPointerInC(e) {
		return false
	}
	t := g.SemanticInfo.Types[e]
	if ident, ok := e.(*ast.Identifier); ok {
		if sym := g.SemanticInfo.Uses[ident]; sym != nil {
			t = sym.Type
		} else if sym := g.SemanticInfo.Defs[ident]; sym != nil {
			t = sym.Type
		}
	}
	if t == nil {
		return false
	}
	if t.Name() == "ptr" || t.Name() == "str" || types.UnwrapLease(t).Name() == "ptr" || types.UnwrapLease(t).Name() == "str" {
		return false
	}

	if g.Solver != nil && g.Solver.Moves[e] {
		return false
	}

	// IndexExpressions are already dereferenced by genIndexExpression
	if _, ok := e.(*ast.IndexExpression); ok {
		return false
	}

	if pt, ok := t.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
		ut := types.UnwrapLease(t)
		if _, ok := ut.(*types.StructType); ok || ut.GetKind() == types.KindSum {
			return false
		}
		// Only dereference if it's NOT actually passed by pointer in C (like LeaseRead primitives)
		if !g.shouldPassByPointer(pt.Base, pt.Kind) {
			// [NEW] If the base type is already a pointer in C (List, Map, Chan, str),
			// we MUST NOT dereference it even if it's a lease.
			if g.isPointerTypeInC(pt.Base) {
				return false
			}

			// [NEW] If it's an identifier, only dereference if it's a Write lease.
			// Read leases on variables should be treated as the pointer itself for assignment/usage.
			if _, ok := e.(*ast.Identifier); ok {
				if pt.Kind == types.LeaseRead {
					return false
				}
			}

			// [NEW] If it's a parameter, it might have been passed by value already
			if id, ok := e.(*ast.Identifier); ok {
				if sym := g.SemanticInfo.Uses[id]; sym != nil && sym.Kind == semantic.SymParam {
					return false
				}
			}
			return true
		} else {
			// If it IS passed by pointer (e.g. LeaseWrite), we must dereference it to get the value
			// UNLESS it's a struct/sum, where we use -> for field access instead.
			ut := types.UnwrapLease(t)
			if _, ok := ut.(*types.StructType); !ok && ut.GetKind() != types.KindSum {
				return true
			}
		}
	}

	return false
}

func (g *Generator) isPointerTypeInC(t types.NRType) bool {
	if t == nil {
		return false
	}
	if pt, ok := t.(*types.PointerType); ok {
		if pt.Leased && !pt.IsArray {
			return g.shouldPassByPointer(pt.Base, pt.Kind) || g.isPointerTypeInC(pt.Base)
		}
		return true
	}
	ut := types.UnwrapLease(t)
	if ut.GetKind() == types.KindPointer || ut.Name() == "str" || ut.Name() == "ptr" {
		return true
	}

	// Lists, Maps, Chans are always pointers in C
	if _, ok := t.(*types.ListType); ok {
		return true
	}
	if _, ok := t.(*types.MapType); ok {
		return true
	}
	if _, ok := t.(*types.ChanType); ok {
		return true
	}
	if _, ok := t.(*types.ProtocolType); ok {
		return false
	}
	return false
}

func (g *Generator) shouldPassByPointer(t types.NRType, l types.LeaseKind) bool {
	if t == nil {
		return false
	}
	if l == types.LeaseWrite {
		return true
	}
	if g.isPointerTypeInC(t) {
		return false
	}
	if l == types.LeaseMove || l == types.LeaseRead {
		if _, ok := t.(*types.StructType); ok || t.GetKind() == types.KindSum || t.GetKind() == types.KindProtocol {
			return true
		}
	}
	return false
}

func (g *Generator) getCFunctionPointerType(ft *types.FunctionType) string {
	if ft == nil {
		return "void*"
	}
	var out bytes.Buffer
	out.WriteString(g.cType(ft.Return))
	out.WriteString(" (*)")
	out.WriteString("(")
	// First parameter is always the environment for closures
	out.WriteString("void*")
	for i, p := range ft.Params {
		out.WriteString(", ")
		lease := types.LeaseRead
		if i < len(ft.ParamLeases) {
			lease = ft.ParamLeases[i]
		}
		out.WriteString(g.cParamType(p, lease))
	}
	if ft.IsVariadic {
		if len(ft.Params) > 0 {
			out.WriteString(", ")
		}
		out.WriteString("...")
	}
	out.WriteString(")")
	return out.String()
}

func (g *Generator) isVariantConstructor(expr ast.Expression) (*types.SumType, string) {
	t := g.SemanticInfo.Types[expr]
	if t == nil {
		return nil, ""
	}

	var st *types.SumType
	var ok bool

	if ft, ok2 := t.(*types.FunctionType); ok2 {
		st, ok = ft.Return.(*types.SumType)
	} else {
		st, ok = t.(*types.SumType)
	}

	if !ok {
		return nil, ""
	}

	var vName string
	switch fn := expr.(type) {
	case *ast.Identifier:
		vName = fn.Value
	case *ast.SelectorExpression:
		vName = fn.Field.Value
	case *ast.IndexExpression:
		if vIdent, ok := fn.Left.(*ast.Identifier); ok {
			vName = vIdent.Value
		} else if vSel, ok := fn.Left.(*ast.SelectorExpression); ok {
			vName = vSel.Field.Value
		}
	}

	if vName != "" {
		if _, exists := st.Variants[vName]; exists {
			return st, vName
		}
	}
	return nil, ""
}

func (g *Generator) isSymbolPointerInC(sym *semantic.Symbol) bool {
	if sym == nil {
		return false
	}
	if g.isPointerTypeInC(sym.Type) {
		return true
	}
	if sym.Kind == semantic.SymParam && g.shouldPassByPointer(sym.Type, sym.LeaseKind) {
		return true
	}
	return false
}

func (g *Generator) isGeneratedExpressionPointer(e ast.Expression, t types.NRType) bool {
	if t == nil {
		t = g.SemanticInfo.Types[e]
	}
	if g.Solver != nil && g.Solver.Moves[e] {
		if pe, ok := e.(*ast.PrefixExpression); ok && pe.Operator == "@" {
			tRight := g.SemanticInfo.Types[pe.Right]
			if tRight == nil {
				if id, ok := pe.Right.(*ast.Identifier); ok {
					if sym := g.SemanticInfo.Uses[id]; sym != nil {
						tRight = sym.Type
					} else if sym := g.SemanticInfo.Defs[id]; sym != nil {
						tRight = sym.Type
					}
				}
			}
			if tRight != nil {
				if pt, ok := tRight.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
					tRight = pt.Base
				}
				hasHeapAlloc := g.isPointerInC(e) && !g.isPointerTypeInC(tRight) && !g.TargetIsValue
				if hasHeapAlloc {
					return true
				}
				return strings.HasSuffix(g.cType(tRight), "*")
			}
		}
		return strings.HasSuffix(g.cType(t), "*")
	}

	if pe, ok := e.(*ast.PrefixExpression); ok && pe.Operator == "@" {
		tRight := g.SemanticInfo.Types[pe.Right]
		if tRight == nil {
			if id, ok := pe.Right.(*ast.Identifier); ok {
				if sym := g.SemanticInfo.Uses[id]; sym != nil {
					tRight = sym.Type
				} else if sym := g.SemanticInfo.Defs[id]; sym != nil {
					tRight = sym.Type
				}
			}
		}
		if tRight != nil {
			if pt, ok := tRight.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
				if _, isProto := pt.Base.(*types.ProtocolType); !isProto {
					tRight = pt.Base
				}
			}
			return strings.HasSuffix(g.cType(tRight), "*")
		}
	}

	return g.isPointerInC(e)
}

func (g *Generator) isChanType(t types.NRType) bool {
	if t == nil {
		return false
	}
	t = types.UnwrapLease(t)
	if pt, ok := t.(*types.PointerType); ok {
		return g.isChanType(pt.Base)
	}
	_, ok := t.(*types.ChanType)
	return ok
}
