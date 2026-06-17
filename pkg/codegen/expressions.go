package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/plugin/api"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/types"
)

func (g *Generator) genExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	// [NEW] Wrap temporary function/method calls returning str in nr_temp_str to prevent leaks
	isTempStrCall := false
	if call, ok := expr.(*ast.CallExpression); ok {
		t := g.SemanticInfo.Types[call]
		if t != nil && t.Name() == "str" && !g.NoTempWrap {
			isUncheckedGet := false
			if sel, isSel := call.Function.(*ast.SelectorExpression); isSel && sel.Field.Value == "unchecked_get" {
				isUncheckedGet = true
			}
			if !isUncheckedGet {
				isTempStrCall = true
				g.buf.WriteString("nr_temp_str(")
			}
		}
	}

	// Check if this expression is a macro call
	if call, ok := expr.(*ast.CallExpression); ok {
		if g.tryExpandMacro(call) {
			if isTempStrCall {
				g.buf.WriteString(")")
			}
			return
		}
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		g.buf.WriteString(strconv.FormatInt(e.Value, 10))
	case *ast.FloatLiteral:
		s := e.Token.Literal
		s = strings.ReplaceAll(s, "_", "")
		s = strings.TrimSuffix(s, "f32")
		s = strings.TrimSuffix(s, "f64")
		s = strings.TrimSuffix(s, "f")
		g.buf.WriteString(s)
	case *ast.ImaginaryLiteral:
		s := e.Token.Literal
		s = strings.ReplaceAll(s, "_", "")
		s = strings.TrimSuffix(s, "i")
		s = strings.TrimSuffix(s, "j")
		g.buf.WriteString(s)
	case *ast.StringLiteral:
		varName := g.registerStringLiteral(e.Value)
		g.buf.WriteString(fmt.Sprintf("(char*)%s.data", varName))
	case *ast.Boolean:
		g.buf.WriteString(strconv.FormatBool(e.Value))
	case *ast.Identifier:
		g.genIdentifier(e)
	case *ast.GroupedExpression:
		g.buf.WriteString("(")
		g.genExpression(e.Expression)
		g.buf.WriteString(")")
	case *ast.InfixExpression:
		g.genInfixExpression(e)
	case *ast.PrefixExpression:
		g.genPrefixExpression(e)
	case *ast.CallExpression:
		g.genCallExpression(e)
	case *ast.SelectorExpression:
		g.genSelectorExpression(e)
	case *ast.IndexExpression:
		g.genIndexExpression(e)
	case *ast.ArrayLiteral:
		g.genArrayLiteral(e)
	case *ast.MapLiteral:
		g.genMapLiteral(e)
	case *ast.StructLiteral:
		g.genStructLiteral(e)
	case *ast.RangeExpression:
		t := g.SemanticInfo.Types[e]
		if t != nil {
			g.buf.WriteString("(")
			g.buf.WriteString(g.cType(t))
			g.buf.WriteString("){.start = ")
			g.genExpression(e.Start)
			g.buf.WriteString(", .end = ")
			g.genExpression(e.End)
			g.buf.WriteString(", .current = ")
			g.genExpression(e.Start)
			g.buf.WriteString("}")
		} else {
			g.buf.WriteString("/* error: RangeExpression has no type */")
		}
	case *ast.SpawnExpression:
		g.genSpawnExpression(e)
	case *ast.ReceiveExpression:
		g.genReceiveExpression(e)
	case *ast.SendExpression:
		g.genSendExpression(e)
	case *ast.ScopeExpression:
		g.genScopeExpression(e)
	case *ast.ParallelExpression:
		g.genParallelExpression(e)
	case *ast.IfExpression:
		t := g.SemanticInfo.Types[e]
		if t != nil && t != types.Void {
			g.buf.WriteString("({ ")
			g.buf.WriteString(fmt.Sprintf("%s _res; ", g.cType(t)))
			g.genIfWithTarget(e, "_res")
			g.buf.WriteString(" _res; })")
		} else {
			g.genIfExpression(e)
		}
	case *ast.MatchExpression:
		t := g.SemanticInfo.Types[e]
		if t != nil && t != types.Void {
			g.buf.WriteString("({ ")
			g.buf.WriteString(fmt.Sprintf("%s _res; ", g.cType(t)))
			g.genMatchWithTarget(e, "_res")
			g.buf.WriteString(" _res; })")
		} else {
			g.genMatchExpression(e)
		}
	case *ast.AllocExpression:
		g.genAllocExpression(e)
	case *ast.AssignmentStatement:
		g.genAssignment(e)
	case *ast.NoneLiteral:
		g.buf.WriteString("NULL")
	case *ast.InterpolatedString:
		g.genInterpolatedString(e)
	case *ast.RuneLiteral:
		g.genRuneLiteral(e)
	case *ast.TryExpression:
		g.genTryExpression(e)
	case *ast.LambdaExpression:
		g.genLambdaExpression(e)
	default:
		g.buf.WriteString(fmt.Sprintf("/* unknown expr: %T */", expr))
	}

	if isTempStrCall {
		g.buf.WriteString(")")
	}
}

func (g *Generator) genIdentifier(e *ast.Identifier) {
	sym := g.SemanticInfo.Uses[e]
	if sym == nil {
		sym = g.SemanticInfo.Defs[e]
	}

	isLocalVarOrParam := false
	if sym != nil && (sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam) {
		isLocalVarOrParam = true
	}
	if g.Solver != nil && g.Solver.Moves[e] && !g.InMoveOperator && isLocalVarOrParam {
		t := g.SemanticInfo.Types[e]
		g.buf.WriteString("({ ")
		g.buf.WriteString(fmt.Sprintf("%s _tmp = ", g.cType(t)))
		if g.isPointerInC(e) && !strings.HasSuffix(g.cType(t), "*") {
			g.buf.WriteString("*")
		}
		oldInMove := g.InMoveOperator
		g.InMoveOperator = true
		g.genIdentifier(e)
		g.InMoveOperator = oldInMove
		g.buf.WriteString("; ")

		var name string
		if sym != nil {
			name = g.variableName(sym)
		} else {
			name = e.Value
		}

		if g.isPointerTypeInC(t) && strings.HasSuffix(g.cType(t), "*") {
			g.buf.WriteString(fmt.Sprintf("%s = NULL; ", name))
		} else {
			if g.isPointerInC(e) {
				g.buf.WriteString(fmt.Sprintf("memset(%s, 0, sizeof(%s)); ", name, g.cType(t)))
			} else {
				g.buf.WriteString(fmt.Sprintf("memset(&%s, 0, sizeof(%s)); ", name, g.cType(t)))
			}
		}
		if sym != nil && g.hasDropFlag(sym) {
			g.buf.WriteString(fmt.Sprintf("%s = false; ", g.dropFlagName(sym)))
		}
		g.buf.WriteString(" _tmp; })")
		return
	}

	if sym != nil {
		if sym.Kind == semantic.SymVariant {
			st := sym.Type.(*types.SumType)
			g.buf.WriteString(fmt.Sprintf("%s_%s_make()", g.mangledTypeName(st), sym.Name))
		} else if sym.Kind == semantic.SymFunc && !g.InCallExpression {
			isExtern := false
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
				isExtern = true
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			}
			if cStdlibFunctions[sym.Name] && sym.Name == g.mangleName(sym) {
				isExtern = true
			}

			if isExtern {
				g.buf.WriteString(g.mangleName(sym))
			} else {
				g.buf.WriteString(fmt.Sprintf("((nr_closure_t){ .env = NULL, .fn_ptr = %s, .drop_fn = NULL })", g.mangleName(sym)))
			}
		} else {
			g.buf.WriteString(g.variableName(sym))
		}
	} else {
		g.buf.WriteString(e.Value)
	}
}

func (g *Generator) genInfixExpression(e *ast.InfixExpression) {
	lt := g.SemanticInfo.Types[e.Left]
	rt := g.SemanticInfo.Types[e.Right]

	// String concatenation
	if e.Operator == "+" && (lt != nil && lt.Name() == "str" || rt != nil && rt.Name() == "str") {
		oldNoTemp := g.NoTempWrap
		if g.NoTempWrap {
			g.buf.WriteString("nr_str_concat_free(")
		} else {
			g.buf.WriteString("nr_temp_str(nr_str_concat_free(")
		}
		g.NoTempWrap = true
		g.genInterpolatedPart(e.Left)
		g.buf.WriteString(", ")
		g.genInterpolatedPart(e.Right)
		g.NoTempWrap = oldNoTemp
		if oldNoTemp {
			g.buf.WriteString(fmt.Sprintf(", %v, %v)", g.isTempInterpolatedPart(e.Left), g.isTempInterpolatedPart(e.Right)))
		} else {
			g.buf.WriteString(fmt.Sprintf(", %v, %v))", g.isTempInterpolatedPart(e.Left), g.isTempInterpolatedPart(e.Right)))
		}
		return
	}

	// String equality
	ult := types.UnwrapLease(lt)
	urt := types.UnwrapLease(rt)
	if (e.Operator == "==" || e.Operator == "!=") && (ult != nil && ult.Name() == "str" && urt != nil && urt.Name() == "str") {
		if e.Operator == "!=" {
			g.buf.WriteString("!")
		}
		g.buf.WriteString("nr_str_eq(")
		g.genExpression(e.Left)
		g.buf.WriteString(", ")
		g.genExpression(e.Right)
		g.buf.WriteString(")")
		return
	}

	// Function/Closure equality
	if (e.Operator == "==" || e.Operator == "!=") && (ult != nil && ult.GetKind() == types.KindFunction && urt != nil && urt.GetKind() == types.KindFunction) {
		g.buf.WriteString("(")
		op := "=="
		logicalOp := "&&"
		if e.Operator == "!=" {
			op = "!="
			logicalOp = "||"
		}
		g.buf.WriteString("(")
		g.genExpression(e.Left)
		g.buf.WriteString(fmt.Sprintf(".env %s ", op))
		g.genExpression(e.Right)
		g.buf.WriteString(".env)")
		g.buf.WriteString(" " + logicalOp + " ")
		g.buf.WriteString("(")
		g.genExpression(e.Left)
		g.buf.WriteString(fmt.Sprintf(".fn_ptr %s ", op))
		g.genExpression(e.Right)
		g.buf.WriteString(".fn_ptr)")
		g.buf.WriteString(")")
		return
	}

	// Struct equality
	if e.Operator == "==" || e.Operator == "!=" {
		utL := types.UnwrapLease(lt)
		utR := types.UnwrapLease(rt)
		_, isStructL := utL.(*types.StructType)
		_, isStructR := utR.(*types.StructType)

		// We only trigger structural equality if:
		// 1. Both are structs
		// 2. Neither is explicitly a pointer/lease in the Nora code (to allow reference equality via #a == #b)
		if isStructL && isStructR && lt.GetKind() != types.KindPointer && rt.GetKind() != types.KindPointer {
			if e.Operator == "!=" {
				g.buf.WriteString("!")
			}
			eqMethod := g.getEqMethod(utL)
			g.buf.WriteString(eqMethod + "(")
			if !strings.HasPrefix(eqMethod, "nr_eq_") {
				g.buf.WriteString("NULL, ")
			}
			if g.isPointerInC(e.Left) {
				g.genExpression(e.Left)
			} else {
				g.buf.WriteString("&")
				g.genExpression(e.Left)
			}
			g.buf.WriteString(", ")
			if g.isPointerInC(e.Right) {
				g.genExpression(e.Right)
			} else {
				g.buf.WriteString("&")
				g.genExpression(e.Right)
			}
			g.buf.WriteString(")")
			return
		}
	}

	// Special handling for none (null) comparisons
	_, isLeftNone := e.Left.(*ast.NoneLiteral)
	_, isRightNone := e.Right.(*ast.NoneLiteral)
	if isLeftNone || isRightNone {
		g.buf.WriteString("(")
		g.genNoneComparisonOperand(e.Left)
		g.buf.WriteString(" " + e.Operator + " ")
		g.genNoneComparisonOperand(e.Right)
		g.buf.WriteString(")")
		return
	}

	// Operator overloads for structs (Arithmetic & Comparison)
	if e.Operator != "==" && e.Operator != "!=" {
		utL := types.UnwrapLease(lt)
		if isStruct, ok := utL.(*types.StructType); ok {
			methodName := ""
			switch e.Operator {
			case "+":
				methodName = "add"
			case "-":
				methodName = "sub"
			case "*":
				methodName = "mul"
			case "/":
				methodName = "div"
			case "%":
				methodName = "mod"
			case "&":
				methodName = "bitand"
			case "|":
				methodName = "bitor"
			case "^":
				methodName = "bitxor"
			case "<<":
				methodName = "shl"
			case ">>":
				methodName = "shr"
			case "<", ">", "<=", ">=":
				methodName = "cmp"
			}
			if methodName != "" {
				if methodType, exists := isStruct.Methods[methodName]; exists {
					if mt, ok := methodType.(*types.FunctionType); ok && len(mt.Params) == 1 {
						if methodName == "cmp" {
							g.buf.WriteString("(")
						}
						g.buf.WriteString(g.mangledTypeName(isStruct) + "_" + methodName + "(")
						g.buf.WriteString("NULL, ")
						g.emitArgument(e.Left, isStruct, mt.ReceiverLease)
						g.buf.WriteString(", ")
						g.emitArgument(e.Right, mt.Params[0], mt.ParamLeases[0])
						g.buf.WriteString(")")

						if methodName == "cmp" {
							g.buf.WriteString(" " + e.Operator + " 0)")
						}
						return
					}
				}
			}
		}
	}

	g.buf.WriteString("(")
	g.genValueExpression(e.Left)
	g.buf.WriteString(" " + e.Operator + " ")
	g.genValueExpression(e.Right)
	g.buf.WriteString(")")
}

func (g *Generator) genValueExpression(expr ast.Expression) {
	if g.shouldDereferenceInC(expr) {
		g.buf.WriteString("(*")
		g.genExpression(expr)
		g.buf.WriteString(")")
	} else {
		g.genExpression(expr)
	}
}

func (g *Generator) genNoneComparisonOperand(expr ast.Expression) {
	t := g.SemanticInfo.Types[expr]
	if ident, ok := expr.(*ast.Identifier); ok {
		if sym := g.SemanticInfo.Uses[ident]; sym != nil {
			t = sym.Type
		} else if sym := g.SemanticInfo.Defs[ident]; sym != nil {
			t = sym.Type
		}
	}

	if t != nil {
		cTypeName := g.cType(t)
		if strings.HasSuffix(cTypeName, "**") {
			g.buf.WriteString("(*")
			g.genExpression(expr)
			g.buf.WriteString(")")
			return
		}
	}

	g.genExpression(expr)
}

func (g *Generator) genPrefixExpression(e *ast.PrefixExpression) {

	// Struct operator overloads
	if e.Operator == "!" || e.Operator == "-" || e.Operator == "~" {
		rt := g.SemanticInfo.Types[e.Right]
		if rt != nil {
			urt := rt
			if pt, ok := rt.(*types.PointerType); ok && !pt.IsArray {
				urt = pt.Base
			}
			if st, ok := urt.(*types.StructType); ok {
				methodName := ""
				switch e.Operator {
				case "!":
					methodName = "not"
				case "-":
					methodName = "neg"
				case "~":
					methodName = "bitnot"
				}
				g.buf.WriteString(g.mangledTypeName(st) + "_" + methodName + "(")
				g.buf.WriteString("NULL, ")

				// We need to pass it by address if the method takes a lease
				if methodType, exists := st.Methods[methodName]; exists {
					if mt, ok := methodType.(*types.FunctionType); ok {
						g.emitArgument(e.Right, st, mt.ReceiverLease)
					} else {
						g.emitArgument(e.Right, st, types.LeaseRead)
					}
				} else {
					g.emitArgument(e.Right, st, types.LeaseRead)
				}
				g.buf.WriteString(")")
				return
			}
		}
	}

	if e.Operator == "@" {
		t := g.SemanticInfo.Types[e.Right]

		// Strip the lease to get the underlying value, EXCEPT for ProtocolType
		// (interfaces are always represented as pointers in C, even when moved).
		if pt, ok := t.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
			if _, isProto := pt.Base.(*types.ProtocolType); !isProto {
				t = pt.Base
			}
		}

		g.buf.WriteString("({ ")
		g.buf.WriteString(fmt.Sprintf("%s _tmp = ", g.cType(t)))
		if g.isPointerInC(e.Right) && !strings.HasSuffix(g.cType(t), "*") {
			g.buf.WriteString("*")
		}
		oldInMove := g.InMoveOperator
		g.InMoveOperator = true
		g.genExpression(e.Right)
		g.InMoveOperator = oldInMove
		g.buf.WriteString("; ")
		if id, ok := e.Right.(*ast.Identifier); ok {
			sym := g.SemanticInfo.Uses[id]
			isLocalVarOrParam := false
			if sym != nil && (sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam) {
				isLocalVarOrParam = true
			}
			if isLocalVarOrParam {
				if g.isPointerInC(id) {
					isHeapPointerLease := false
					if sym != nil {
						if pt, ok := sym.Type.(*types.PointerType); ok && pt.Kind == types.LeaseMove {
							isHeapPointerLease = true
						}
					}

					if isHeapPointerLease {
						if !g.isPointerTypeInC(t) {
							g.buf.WriteString(fmt.Sprintf("nr_free(%s); %s = NULL; ", id.Value, id.Value))
						} else {
							g.buf.WriteString(fmt.Sprintf("%s = NULL; ", id.Value))
						}
					} else {
						if !g.isPointerTypeInC(t) {
							g.buf.WriteString(fmt.Sprintf("memset(%s, 0, sizeof(%s)); ", id.Value, g.cType(t)))
						} else {
							g.buf.WriteString(fmt.Sprintf("%s = NULL; ", id.Value))
						}
					}
				} else {
					g.buf.WriteString(fmt.Sprintf("memset(&%s, 0, sizeof(%s)); ", id.Value, g.cType(t)))
				}
				if sym != nil && g.hasDropFlag(sym) {
					g.buf.WriteString(fmt.Sprintf("%s = false; ", g.dropFlagName(sym)))
				}
			}
		} else if sel, ok := e.Right.(*ast.SelectorExpression); ok {
			exprStr := g.genSelectorString(sel.Left, sel.Field.Value)
			if g.isPointerTypeInC(t) && strings.HasSuffix(g.cType(t), "*") {
				g.buf.WriteString(fmt.Sprintf("%s = NULL; ", exprStr))
			} else {
				if g.isPointerInC(sel) {
					g.buf.WriteString(fmt.Sprintf("memset(%s, 0, sizeof(%s)); ", exprStr, g.cType(t)))
				} else {
					g.buf.WriteString(fmt.Sprintf("memset(&%s, 0, sizeof(%s)); ", exprStr, g.cType(t)))
				}
			}
		} else if idx, ok := e.Right.(*ast.IndexExpression); ok {
			if g.isPointerTypeInC(t) && strings.HasSuffix(g.cType(t), "*") {
				g.buf.WriteString(fmt.Sprintf("(*((%s*)array_bounds_check(%s, %s, \"%s\", %d))) = NULL; ",
					g.cType(t), g.exprToString(idx.Left), g.exprToString(idx.Indices[0]),
					strings.ReplaceAll(idx.Pos().Filename, "\\", "/"), idx.Pos().Line))
			} else {
				g.buf.WriteString(fmt.Sprintf("memset(array_bounds_check(%s, %s, \"%s\", %d), 0, sizeof(%s)); ",
					g.exprToString(idx.Left), g.exprToString(idx.Indices[0]),
					strings.ReplaceAll(idx.Pos().Filename, "\\", "/"), idx.Pos().Line, g.cType(t)))
			}
		} else if call, ok := e.Right.(*ast.CallExpression); ok {
			if sel, isSel := call.Function.(*ast.SelectorExpression); isSel && sel.Field.Value == "unchecked_get" {
				if g.isPointerTypeInC(t) && strings.HasSuffix(g.cType(t), "*") {
					g.buf.WriteString(fmt.Sprintf("(%s) = NULL; ", g.exprToString(call)))
				} else {
					g.buf.WriteString(fmt.Sprintf("memset(&(%s), 0, sizeof(%s)); ", g.exprToString(call), g.cType(t)))
				}
			}
		}
		ct := g.cType(t)
		if g.isPointerInC(e) && !g.isPointerTypeInC(t) && !g.TargetIsValue {
			g.buf.WriteString(fmt.Sprintf("%s* _p = nr_malloc_debug(sizeof(%s), \"%s\", %d); ", ct, ct, strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
			g.buf.WriteString("*_p = _tmp; ")
			g.buf.WriteString("_p; })")
		} else {
			g.buf.WriteString(" _tmp; })")
		}
		return
	}
	if e.Operator == "#" || e.Operator == "&" {
		t := g.SemanticInfo.Types[e.Right]
		ut := types.UnwrapLease(t)
		isPrimitive := false
		if ut != nil {
			if _, ok := ut.(*types.PrimitiveType); ok {
				isPrimitive = true
			}
		}

		tExpr := g.SemanticInfo.Types[e]
		cTypeExpr := g.cType(tExpr)
		cTypeRight := g.cType(t)
		if id, ok := e.Right.(*ast.Identifier); ok {
			if sym := g.SemanticInfo.Uses[id]; sym != nil && sym.Kind == semantic.SymParam {
				cTypeRight = g.cParamType(sym.Type, sym.LeaseKind)
			}
		}

		writeAddrOf := true
		if e.Operator == "#" && isPrimitive {
			writeAddrOf = false
		} else if cTypeExpr == cTypeRight {
			writeAddrOf = false
		}

		if writeAddrOf {
			g.buf.WriteString("&")
		}
		g.genExpression(e.Right)
		return
	}
	g.buf.WriteString("(")
	g.buf.WriteString(e.Operator)
	g.genExpression(e.Right)
	g.buf.WriteString(")")
}

func (g *Generator) genCallExpression(e *ast.CallExpression) {
	oldNoTemp := g.NoTempWrap
	g.NoTempWrap = false
	defer func() { g.NoTempWrap = oldNoTemp }()
	// 0. Type cast check (e.g. i64(x), i32(y), collections.Vector[T](z))
	var matchedStruct *types.StructType
	var castSym *semantic.Symbol
	var isSelector = false

	funcExpr := e.Function
	if idx, ok := funcExpr.(*ast.IndexExpression); ok {
		funcExpr = idx.Left
	}
	// Removed debug selector checking

	exprType := g.SemanticInfo.Types[e]
	if exprType != nil {
		underlying := types.UnwrapLease(exprType)
		for {
			if pt, ok := underlying.(*types.PointerType); ok {
				underlying = pt.Base
			} else {
				break
			}
		}
		if st, ok := underlying.(*types.StructType); ok {
			matchedStruct = st
		}
	}

	if matchedStruct == nil {
		if ident, ok := funcExpr.(*ast.Identifier); ok {
			for _, st := range g.Structs {
				if st != nil && (st.TypeName == ident.Value || (st.BaseType != nil && st.BaseType.TypeName == ident.Value)) {
					matchedStruct = st
					break
				}
			}
		} else if sel, ok := funcExpr.(*ast.SelectorExpression); ok {
			for _, st := range g.Structs {
				if st != nil && (st.TypeName == sel.Field.Value || (st.BaseType != nil && st.BaseType.TypeName == sel.Field.Value)) {
					matchedStruct = st
					break
				}
			}
		}
	}

	if ident, ok := funcExpr.(*ast.Identifier); ok {
		castSym = g.SemanticInfo.Uses[ident]
		if castSym == nil {
			castSym = g.findSymbolByName(ident.Value)
		}
	} else if sel, ok := funcExpr.(*ast.SelectorExpression); ok {
		isSelector = true
		castSym = g.SemanticInfo.Uses[sel.Field]
	}

	if matchedStruct != nil && len(e.Arguments) == 1 {
		argType := g.SemanticInfo.Types[e.Arguments[0].Value]
		if argType != nil {
			unwrapped := types.UnwrapLease(argType)
			if _, isProto := unwrapped.(*types.ProtocolType); isProto {
				g.buf.WriteString("((")
				g.buf.WriteString(g.mangledTypeName(matchedStruct))
				g.buf.WriteString("*)(")
				g.buf.WriteString("(") // Added opening parenthesis
				g.genExpression(e.Arguments[0].Value)
				g.buf.WriteString(")") // Added closing parenthesis
				if g.isPointerInC(e.Arguments[0].Value) {
					g.buf.WriteString("->data")
				} else {
					g.buf.WriteString(".data")
				}
				g.buf.WriteString("))")
				return
			}
		}
	}

	if castSym != nil && castSym.Kind == semantic.SymType && len(e.Arguments) == 1 {
		if _, isPrim := castSym.Type.(*types.PrimitiveType); isPrim {
			argType := g.SemanticInfo.Types[e.Arguments[0].Value]
			if argType != nil {
				unwrapped := types.UnwrapLease(argType)
				if _, isProto := unwrapped.(*types.ProtocolType); isProto {
					exprType := g.SemanticInfo.Types[e]
					isLease := false
					if pt, ok := exprType.(*types.PointerType); ok && pt.Leased {
						isLease = true
					}
					if isLease || !g.TargetIsValue {
						g.buf.WriteString("((")
						g.buf.WriteString(g.cType(castSym.Type))
						g.buf.WriteString("*)(")
						g.buf.WriteString("(")
						g.genExpression(e.Arguments[0].Value)
						g.buf.WriteString(")")
						if g.isPointerInC(e.Arguments[0].Value) {
							g.buf.WriteString("->data")
						} else {
							g.buf.WriteString(".data")
						}
						g.buf.WriteString("))")
					} else {
						g.buf.WriteString("(*((")
						g.buf.WriteString(g.cType(castSym.Type))
						g.buf.WriteString("*)(")
						g.buf.WriteString("(")
						g.genExpression(e.Arguments[0].Value)
						g.buf.WriteString(")")
						if g.isPointerInC(e.Arguments[0].Value) {
							g.buf.WriteString("->data")
						} else {
							g.buf.WriteString(".data")
						}
						g.buf.WriteString(")))")
					}
					return
				}
			}
			g.buf.WriteString("((")
			g.buf.WriteString(g.cType(castSym.Type))
			g.buf.WriteString(")(")
			g.genExpression(e.Arguments[0].Value)
			g.buf.WriteString("))")
			return
		} else if _, isStruct := castSym.Type.(*types.StructType); isStruct {
			argType := g.SemanticInfo.Types[e.Arguments[0].Value]
			if argType != nil {
				unwrapped := types.UnwrapLease(argType)
				if _, isProto := unwrapped.(*types.ProtocolType); isProto {
					castType := g.SemanticInfo.Types[e]
					if castType == nil {
						castType = castSym.Type
					}
					g.buf.WriteString("((")
					g.buf.WriteString(g.cType(castType))
					g.buf.WriteString("*)(")
					g.buf.WriteString("(")
					g.genExpression(e.Arguments[0].Value)
					g.buf.WriteString(")")
					if g.isPointerInC(e.Arguments[0].Value) {
						g.buf.WriteString("->data")
					} else {
						g.buf.WriteString(".data")
					}
					g.buf.WriteString("))")
					return
				}
			}
		}
	} else if !isSelector {
		// Fallback in case semantic uses is not populated for type identifier (only for simple identifiers)
		if ident, ok := e.Function.(*ast.Identifier); ok {
			if prim, exists := types.LookupPrimitive(ident.Value); exists && len(e.Arguments) == 1 {
				g.buf.WriteString("((")
				g.buf.WriteString(g.cType(prim))
				g.buf.WriteString(")(")
				g.genExpression(e.Arguments[0].Value)
				g.buf.WriteString("))")
				return
			}
		}
	}

	// 0. Generic monomorphization check
	if mangledName, ok := g.SemanticInfo.MonomorphizedNames[e]; ok {
		g.buf.WriteString(mangledName)
		g.buf.WriteString("(")

		// Determine parameter types from the SPECIALIZED type
		var ft *types.FunctionType
		if t, ok := g.SemanticInfo.Types[e.Function].(*types.FunctionType); ok {
			ft = t
		}

		isMethod := false
		var receiver ast.Expression
		if ft != nil && ft.Receiver != nil {
			if sel, ok := e.Function.(*ast.SelectorExpression); ok {
				receiver = sel.Left
				isMethod = true
			}
		}

		if isMethod && receiver != nil {
			g.buf.WriteString("NULL, ")
			g.emitArgument(receiver, ft.Receiver, ft.ReceiverLease)
			for i, arg := range e.Arguments {
				g.buf.WriteString(", ")
				var targetType types.NRType
				lease := types.LeaseRead
				if i < len(ft.Params) {
					targetType = ft.Params[i]
				}
				if i < len(ft.ParamLeases) {
					lease = ft.ParamLeases[i]
				}
				g.emitArgument(arg.Value, targetType, lease)
			}
		} else {
			g.buf.WriteString("NULL")
			if len(e.Arguments) > 0 {
				g.buf.WriteString(", ")
			}
			for i, arg := range e.Arguments {
				if i > 0 {
					g.buf.WriteString(", ")
				}
				var targetType types.NRType
				lease := types.LeaseRead
				if ft != nil {
					if i < len(ft.Params) {
						targetType = ft.Params[i]
					}
					if i < len(ft.ParamLeases) {
						lease = ft.ParamLeases[i]
					}
				}
				g.emitArgument(arg.Value, targetType, lease)
			}
		}
		g.buf.WriteString(")")
		return
	}

	// 1. Variant Constructor Check
	if st, vName := g.isVariantConstructor(e.Function); st != nil {
		g.buf.WriteString(fmt.Sprintf("%s_%s_make(", g.mangledTypeName(st), vName))
		oldTargetVal := g.TargetIsValue
		g.TargetIsValue = true
		for i, arg := range e.Arguments {
			if i > 0 {
				g.buf.WriteString(", ")
			}
			g.genExpression(arg.Value)
		}
		g.TargetIsValue = oldTargetVal
		g.buf.WriteString(")")
		return
	}

	// 1.35 Built-in: unchecked_get and unchecked_set
	if sel, ok := e.Function.(*ast.SelectorExpression); ok && (sel.Field.Value == "unchecked_get" || sel.Field.Value == "unchecked_set") {
		if lt := g.SemanticInfo.Types[sel.Left]; lt != nil {
			unwrapped := types.UnwrapLease(lt)
			isCollection := false
			var elemType types.NRType
			if listType, ok := unwrapped.(*types.ListType); ok {
				isCollection = true
				elemType = listType.ElementType
			} else if pt, ok := unwrapped.(*types.PointerType); ok && pt.IsArray {
				isCollection = true
				elemType = pt.Base
			}
			if isCollection {
				if sel.Field.Value == "unchecked_get" && len(e.Arguments) == 1 {
					g.buf.WriteString("(((")
					g.buf.WriteString(g.cType(elemType))
					g.buf.WriteString("*)")
					if g.isPointerInC(sel.Left) {
						g.buf.WriteString("(")
						g.genExpression(sel.Left)
						g.buf.WriteString(")")
					} else {
						g.buf.WriteString("(")
						g.genExpression(sel.Left)
						g.buf.WriteString(".data)")
					}
					g.buf.WriteString(")[")
					g.genExpression(e.Arguments[0].Value)
					g.buf.WriteString("])")
					return
				} else if sel.Field.Value == "unchecked_set" && len(e.Arguments) == 2 {
					g.buf.WriteString("(((")
					g.buf.WriteString(g.cType(elemType))
					g.buf.WriteString("*)")
					if g.isPointerInC(sel.Left) {
						g.buf.WriteString("(")
						g.genExpression(sel.Left)
						g.buf.WriteString(")")
					} else {
						g.buf.WriteString("(")
						g.genExpression(sel.Left)
						g.buf.WriteString(".data)")
					}
					g.buf.WriteString(")[")
					g.genExpression(e.Arguments[0].Value)
					g.buf.WriteString("] = ")

					oldVal := g.TargetIsValue
					g.TargetIsValue = true
					g.genExpression(e.Arguments[1].Value)
					g.TargetIsValue = oldVal

					g.buf.WriteString(")")
					return
				}
			}
		}
	}

	// 1.4. Built-in: .clone() for channels -> channel_ref(c)
	if sel, ok := e.Function.(*ast.SelectorExpression); ok && sel.Field.Value == "clone" {
		if lt := g.SemanticInfo.Types[sel.Left]; lt != nil {
			unwrapped := types.UnwrapLease(lt)
			for {
				if pt, ok := unwrapped.(*types.PointerType); ok {
					unwrapped = pt.Base
				} else {
					break
				}
			}
			if unwrapped.GetKind() == types.KindChan {
				g.buf.WriteString("({ channel_t* _tmp_chan = ")
				g.genExpression(sel.Left)
				g.buf.WriteString("; channel_ref(_tmp_chan); _tmp_chan; })")
				return
			}
		}
	}

	// 1.5. Built-in: append(list, item) -> array_append(list, item)
	if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "append" {
		if len(e.Arguments) == 2 {
			listExpr := e.Arguments[0].Value
			itemExpr := e.Arguments[1].Value
			listType := g.SemanticInfo.Types[listExpr]

			g.buf.WriteString("array_append(")
			g.genExpression(listExpr)
			g.buf.WriteString(", ")

			if lt, ok := listType.(*types.ListType); ok {
				itemType := lt.ElementType
				if types.IsOwnedType(itemType) || itemType.GetKind() == types.KindProtocol {
					oldBuf := g.buf
					var tempBuf bytes.Buffer
					g.buf = &tempBuf
					g.emitArgument(itemExpr, itemType, types.LeaseMove)
					argStr := tempBuf.String()
					g.buf = oldBuf

					if itemType.GetKind() == types.KindProtocol {
						g.buf.WriteString(argStr)
					} else if strings.Contains(argStr, "({") {
						g.buf.WriteString(fmt.Sprintf("((%s[]){ %s })", g.cType(itemType), argStr))
					} else {
						g.buf.WriteString(fmt.Sprintf("&(%s)", argStr))
					}
				} else {
					g.buf.WriteString(fmt.Sprintf("&(%s){", g.cType(itemType)))
					g.genExpression(itemExpr)
					g.buf.WriteString("}")
				}
			} else {
				g.genExpression(itemExpr)
			}
			g.buf.WriteString(")")
			return
		}
	}

	// 2. Built-in: make(chan[T], cap) -> channel_make(cap, sizeof(T))
	if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "make" {
		if len(e.Arguments) >= 1 {
			typeArg := e.Arguments[0].Value
			capArg := "0"
			if len(e.Arguments) >= 2 {
				capArg = g.exprToString(e.Arguments[1].Value)
			}
			if chanTypeNode, ok := typeArg.(*ast.ChanType); ok {
				elemType := g.SemanticInfo.Types[chanTypeNode.Value]
				if elemType == nil {
					elemType = types.I32
				}
				g.buf.WriteString(fmt.Sprintf("channel_make(%s, sizeof(%s))", capArg, g.cType(elemType)))
				return
			} else if lt, ok := g.SemanticInfo.Types[typeArg].(*types.ListType); ok {
				g.buf.WriteString(fmt.Sprintf("array_make_empty(%s, sizeof(%s), \"%s\", %d)", capArg, g.cType(lt.ElementType), strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
				return
			} else if mt, ok := g.SemanticInfo.Types[typeArg].(*types.MapType); ok {
				isStrKey := mt.Key.Name() == "str"
				g.buf.WriteString(fmt.Sprintf("map_make(sizeof(%s), sizeof(%s), %v, \"%s\", %d)", g.cType(mt.Key), g.cType(mt.Value), isStrKey, strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
				return
			}
		}
	}

	// 3. Built-in: panic(msg) -> nr_panic(msg, __FILE__, __LINE__)
	if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "panic" {
		g.buf.WriteString("nr_panic(")
		if len(e.Arguments) > 0 {
			g.genExpression(e.Arguments[0].Value)
		} else {
			g.buf.WriteString("\"panic\"")
		}
		g.buf.WriteString(fmt.Sprintf(", \"%s\", %d)", strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
		return
	}

	// 4. Built-in: len(x) -> array_count(x)
	if ident, ok := e.Function.(*ast.Identifier); ok && ident.Value == "len" {
		if len(e.Arguments) >= 1 {
			g.buf.WriteString("array_count(")
			g.genExpression(e.Arguments[0].Value)
			g.buf.WriteString(")")
			return
		}
	}

	// 3.5 Interface Call
	if sel, ok := e.Function.(*ast.SelectorExpression); ok {
		leftType := types.UnwrapLease(g.SemanticInfo.Types[sel.Left])
		if proto, ok := leftType.(*types.ProtocolType); ok {
			g.genInterfaceCall(e, sel, proto)
			return
		}
	}

	// 4. Method Call
	var receiver ast.Expression
	var methodSym *semantic.Symbol
	if sel, ok := e.Function.(*ast.SelectorExpression); ok {
		if sym := g.SemanticInfo.Uses[sel.Field]; sym != nil && sym.Kind == semantic.SymFunc && sym.DefNode != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.Receiver != nil {
				receiver = sel.Left
				methodSym = sym
			}
		}
	}

	if methodSym != nil {
		g.buf.WriteString(g.mangleName(methodSym))
		g.buf.WriteString("(NULL, ")

		var receiverType types.NRType
		lease := types.LeaseRead
		if ft, ok := methodSym.Type.(*types.FunctionType); ok {
			receiverType = ft.Receiver
			lease = ft.ReceiverLease
		}
		g.emitArgument(receiver, receiverType, lease)

		for i, arg := range e.Arguments {
			g.buf.WriteString(", ")
			var targetType types.NRType
			argLease := types.LeaseRead
			if ft, ok := methodSym.Type.(*types.FunctionType); ok {
				if i < len(ft.Params) {
					targetType = ft.Params[i]
				}
				if i < len(ft.ParamLeases) {
					argLease = ft.ParamLeases[i]
				}
			}
			g.emitArgument(arg.Value, targetType, argLease)
		}
		g.buf.WriteString(")")
		return
	}

	var ft *types.FunctionType
	unwrapped := types.UnwrapLease(g.SemanticInfo.Types[e.Function])
	if t, ok := unwrapped.(*types.FunctionType); ok {
		ft = t
	} else if ident, ok := e.Function.(*ast.Identifier); ok {
		if sym := g.SemanticInfo.Uses[ident]; sym != nil {
			unwrappedSym := types.UnwrapLease(sym.Type)
			if t, ok := unwrappedSym.(*types.FunctionType); ok {
				ft = t
			}
		}
	}

	isVariableCall := true
	if ident, ok := e.Function.(*ast.Identifier); ok {
		if sym := g.SemanticInfo.Uses[ident]; sym != nil && sym.Kind == semantic.SymFunc {
			isVariableCall = false
		}
	} else if sel, ok := e.Function.(*ast.SelectorExpression); ok {
		// Method calls are already handled above, if we are here it's a field or something else
		if sym := g.SemanticInfo.Uses[sel.Field]; sym != nil && sym.Kind == semantic.SymFunc {
			isVariableCall = false
		}
	}

	isExtern := false
	if ident, ok := e.Function.(*ast.Identifier); ok {
		if sym := g.SemanticInfo.Uses[ident]; sym != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
				isExtern = true
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			} else if sym.DefNode == nil && sym.Kind == semantic.SymFunc {
				isExtern = true
			}
			if cStdlibFunctions[sym.Name] && sym.Name == g.mangleName(sym) {
				isExtern = true
			}
		}
	} else if sel, ok := e.Function.(*ast.SelectorExpression); ok {
		if sym := g.SemanticInfo.Uses[sel.Field]; sym != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
				isExtern = true
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			} else if sym.DefNode == nil && sym.Kind == semantic.SymFunc {
				isExtern = true
			}
		}
	}

	if ft != nil && isVariableCall {
		g.buf.WriteString("({ nr_closure_t _c = ")
		g.genExpression(e.Function)
		g.buf.WriteString(fmt.Sprintf("; ((%s)_c.fn_ptr)(_c.env", g.getCFunctionPointerType(ft)))
		if len(e.Arguments) > 0 {
			g.buf.WriteString(", ")
		}
	} else {
		oldInCall := g.InCallExpression
		g.InCallExpression = true
		g.genExpression(e.Function)
		g.InCallExpression = oldInCall
		g.buf.WriteString("(")
		if !isExtern {
			g.buf.WriteString("NULL")
			if len(e.Arguments) > 0 {
				g.buf.WriteString(", ")
			}
		}
	}

	for i, arg := range e.Arguments {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		var targetType types.NRType
		lease := types.LeaseRead
		if ft != nil {
			if i < len(ft.Params) {
				targetType = ft.Params[i]
			}
			if i < len(ft.ParamLeases) {
				lease = ft.ParamLeases[i]
			}
		}
		g.emitArgument(arg.Value, targetType, lease)
	}
	if ft != nil && isVariableCall {
		g.buf.WriteString("); })")
	} else {
		g.buf.WriteString(")")
	}
}

func (g *Generator) genLambdaExpression(e *ast.LambdaExpression) {
	fnName := g.generateLambdaFunction(e)

	envStructName := fnName + "_env_t"

	scope := g.SemanticInfo.Scopes[e]
	hasCaptures := false
	if scope != nil && scope.Captures != nil && len(scope.Captures) > 0 {
		hasCaptures = true
	}

	g.buf.WriteString("({ nr_closure_t _c; ")
	if hasCaptures {
		g.buf.WriteString(fmt.Sprintf("%s* _env_local = nr_malloc(sizeof(%s)); ", envStructName, envStructName))
		g.populateLambdaEnv(e, scope)
		g.buf.WriteString("_c.env = _env_local; ")
		g.buf.WriteString(fmt.Sprintf("_c.drop_fn = %s_env_drop; ", fnName))
	} else {
		g.buf.WriteString("_c.env = NULL; ")
		g.buf.WriteString("_c.drop_fn = NULL; ")
	}
	g.buf.WriteString(fmt.Sprintf("_c.fn_ptr = %s; ", fnName))
	g.buf.WriteString("_c; })")
}

func (g *Generator) genAddressOf(expr ast.Expression, targetType types.NRType, lease types.LeaseKind) {
	isLVal := false
	if g.Solver == nil || !g.Solver.Moves[expr] {
		switch expr.(type) {
		case *ast.Identifier, *ast.SelectorExpression, *ast.IndexExpression:
			isLVal = true
		}
	}

	if isLVal {
		g.buf.WriteString("&")
		if lease == types.LeaseMove {
			g.genOwnedValue(expr, targetType)
		} else {
			g.genExpression(expr)
		}
	} else {
		// RValue: use C99 array compound literal to safely get its address with block-scope lifetime
		t := g.SemanticInfo.Types[expr]
		if pt, ok := t.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
			t = pt.Base
		}
		ct := g.cType(t)
		g.buf.WriteString(fmt.Sprintf("(%s[]){ ", ct))

		oldTargetIsValue := g.TargetIsValue
		g.TargetIsValue = true

		if lease == types.LeaseMove {
			g.genOwnedValue(expr, targetType)
		} else {
			g.genExpression(expr)
		}

		g.TargetIsValue = oldTargetIsValue

		g.buf.WriteString(" }")
	}
}
func (g *Generator) emitArgument(expr ast.Expression, targetType types.NRType, lease types.LeaseKind) {
	if proto, ok := targetType.(*types.ProtocolType); ok {
		if g.shouldPassByPointer(targetType, lease) {
			isLVal := false
			if g.Solver == nil || !g.Solver.Moves[expr] {
				switch expr.(type) {
				case *ast.Identifier, *ast.SelectorExpression, *ast.IndexExpression:
					isLVal = true
				}
			}
			if isLVal {
				g.buf.WriteString("&(")
				g.genInterfaceCast(expr, proto)
				g.buf.WriteString(")")
			} else {
				g.buf.WriteString(fmt.Sprintf("(%s[]){ ", g.cType(proto)))
				oldTarget := g.TargetIsValue
				g.TargetIsValue = true
				g.genInterfaceCast(expr, proto)
				g.TargetIsValue = oldTarget
				g.buf.WriteString(" }")
			}
		} else {
			g.genInterfaceCast(expr, proto)
		}
		return
	}
	argType := g.SemanticInfo.Types[expr]
	// Handle PrefixExpression (@u1, #u1) where Semantic Analyzer might have set it to Void
	if prefix, ok := expr.(*ast.PrefixExpression); ok {
		if prefix.Operator == "@" || prefix.Operator == "#" {
			argType = g.SemanticInfo.Types[prefix.Right]
			if argType == nil {
				if ident, ok := prefix.Right.(*ast.Identifier); ok {
					if sym := g.SemanticInfo.Uses[ident]; sym != nil {
						argType = sym.Type
					}
				}
			}
		}
	} else if argType == nil {
		if ident, ok := expr.(*ast.Identifier); ok {
			if sym := g.SemanticInfo.Uses[ident]; sym != nil {
				argType = sym.Type
			}
		}
	}

	if (lease == types.LeaseRead || lease == types.LeaseWrite || lease == types.LeaseMove) && targetType != nil && argType != nil {
		targetCType := g.cParamType(targetType, lease)
		argCType := g.cType(argType)
		targetLevel := strings.Count(targetCType, "*")
		argLevel := strings.Count(argCType, "*")
		if argLevel > targetLevel {
			shouldDeref := true
			if targetType != nil {
				unwrapped := types.UnwrapLease(targetType)
				if targetType.Name() == "ptr" || unwrapped.Name() == "ptr" || targetType.Name() == "str" || unwrapped.Name() == "str" {
					shouldDeref = false
				}
			}
			if shouldDeref {
				for i := 0; i < argLevel-targetLevel; i++ {
					g.buf.WriteString("*")
				}
			}
		}
		if targetCType == argCType+"*" {
			isAlreadyPointerPointer := false
			if prefix, ok := expr.(*ast.PrefixExpression); ok && (prefix.Operator == "@" || prefix.Operator == "#" || prefix.Operator == "&") {
				isAlreadyPointerPointer = true
			} else if _, ok := expr.(*ast.NoneLiteral); ok {
				isAlreadyPointerPointer = true
			} else if ident, ok := expr.(*ast.Identifier); ok {
				if sym := g.SemanticInfo.Uses[ident]; sym != nil {
					if sym.Kind == semantic.SymParam && g.shouldPassByPointer(sym.Type, sym.LeaseKind) {
						isAlreadyPointerPointer = true
					}
				}
			}
			if !isAlreadyPointerPointer {
				g.genAddressOf(expr, targetType, lease)
				return
			}
		}
	}

	if g.shouldPassByPointer(argType, lease) {
		isPointer := g.isPointerInC(expr)
		if isPointer && lease == types.LeaseMove && (argType.GetKind() == types.KindStruct || argType.GetKind() == types.KindSum) {
			isPointer = false
		}
		if isPointer {
			if lease == types.LeaseMove {
				g.genOwnedValue(expr, targetType)
			} else {
				g.genExpression(expr)
			}
		} else {
			g.genAddressOf(expr, targetType, lease)
		}
	} else {
		// Only dereference if the target type is NOT a pointer in C
		if g.shouldDereferenceInC(expr) && !g.isPointerTypeInC(targetType) {
			g.buf.WriteString("*")
		}
		if lease == types.LeaseMove {
			g.genOwnedValue(expr, targetType)
		} else {
			g.genExpression(expr)
		}
	}
}

func (g *Generator) genSelectorExpression(e *ast.SelectorExpression) {
	// [NEW] Handle implicit moves for SelectorExpression
	if g.Solver != nil && g.Solver.Moves[e] && !g.InMoveOperator {
		t := g.SemanticInfo.Types[e]
		g.buf.WriteString("({ ")
		g.buf.WriteString(fmt.Sprintf("%s _tmp = ", g.cType(t)))
		if g.isPointerInC(e) && !strings.HasSuffix(g.cType(t), "*") {
			g.buf.WriteString("*")
		}
		oldInMove := g.InMoveOperator
		g.InMoveOperator = true
		g.genSelectorExpression(e)
		g.InMoveOperator = oldInMove
		g.buf.WriteString("; ")

		exprStr := g.genSelectorString(e.Left, e.Field.Value)
		if g.isPointerTypeInC(t) && strings.HasSuffix(g.cType(t), "*") {
			g.buf.WriteString(fmt.Sprintf("%s = NULL; ", exprStr))
		} else {
			if g.isPointerInC(e) {
				g.buf.WriteString(fmt.Sprintf("memset(%s, 0, sizeof(%s)); ", exprStr, g.cType(t)))
			} else {
				g.buf.WriteString(fmt.Sprintf("memset(&%s, 0, sizeof(%s)); ", exprStr, g.cType(t)))
			}
		}
		g.buf.WriteString(" _tmp; })")
		return
	}

	// Check if this is a package access (e.g., io.println)
	if leftIdent, ok := e.Left.(*ast.Identifier); ok {
		if leftSym := g.SemanticInfo.Uses[leftIdent]; leftSym != nil && (leftSym.Kind == semantic.SymPackage || leftSym.Kind == semantic.SymModule) {
			if sym := g.SemanticInfo.Uses[e.Field]; sym != nil {
				g.buf.WriteString(g.mangleName(sym))
				return
			}
		}
	}

	g.buf.WriteString(g.genSelectorString(e.Left, e.Field.Value))
}

func (g *Generator) genSelectorString(leftExpr ast.Expression, field string) string {
	op := "."
	if g.isPointerInC(leftExpr) {
		op = "->"
	}
	leftStr := g.exprToString(leftExpr)
	cTypeName := g.cType(g.SemanticInfo.Types[leftExpr])

	if g.SemanticInfo.Types[leftExpr] == nil {
		if ident, ok := leftExpr.(*ast.Identifier); ok {
			var t types.NRType
			if sym := g.SemanticInfo.Uses[ident]; sym != nil {
				t = sym.Type
			} else if sym := g.SemanticInfo.Defs[ident]; sym != nil {
				t = sym.Type
			}
			if t != nil {
				cTypeName = g.cType(t)
			}
		}
	}

	if strings.HasSuffix(cTypeName, "**") {
		return fmt.Sprintf("(*%s)%s%s", leftStr, op, field)
	}
	return fmt.Sprintf("%s%s%s", leftStr, op, field)
}

func (g *Generator) genIndexExpression(e *ast.IndexExpression) {
	// [NEW] Handle implicit moves for IndexExpression
	if g.Solver != nil && g.Solver.Moves[e] && !g.InMoveOperator {
		t := g.SemanticInfo.Types[e]
		g.buf.WriteString("({ ")
		g.buf.WriteString(fmt.Sprintf("%s _tmp = ", g.cType(t)))
		if g.isPointerInC(e) && !strings.HasSuffix(g.cType(t), "*") {
			g.buf.WriteString("*")
		}
		oldInMove := g.InMoveOperator
		g.InMoveOperator = true
		g.genIndexExpression(e)
		g.InMoveOperator = oldInMove
		g.buf.WriteString("; ")

		if g.isPointerTypeInC(t) && strings.HasSuffix(g.cType(t), "*") {
			if e.NoBoundsCheck {
				g.buf.WriteString(fmt.Sprintf("((((%s*)%s)[%s]) = NULL; ",
					g.cType(t), g.exprToString(e.Left), g.exprToString(e.Indices[0])))
			} else {
				g.buf.WriteString(fmt.Sprintf("(*((%s*)array_bounds_check(%s, %s, \"%s\", %d))) = NULL; ",
					g.cType(t), g.exprToString(e.Left), g.exprToString(e.Indices[0]),
					strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
			}
		} else {
			if e.NoBoundsCheck {
				g.buf.WriteString(fmt.Sprintf("memset(&(((%s*)%s)[%s]), 0, sizeof(%s)); ",
					g.cType(t), g.exprToString(e.Left), g.exprToString(e.Indices[0]), g.cType(t)))
			} else {
				g.buf.WriteString(fmt.Sprintf("memset(array_bounds_check(%s, %s, \"%s\", %d), 0, sizeof(%s)); ",
					g.exprToString(e.Left), g.exprToString(e.Indices[0]),
					strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line, g.cType(t)))
			}
		}
		g.buf.WriteString(" _tmp; })")
		return
	}

	if len(e.Indices) == 0 {
		g.genExpression(e.Left)
		return
	}

	// [NEW] Check for generic variant constructor: None[i32]
	if st, vName := g.isVariantConstructor(e); st != nil {
		g.buf.WriteString(fmt.Sprintf("%s_%s_make()", g.mangledTypeName(st), vName))
		return
	}

	// Check if it's a slice: arr[start:end]
	if se, ok := e.Indices[0].(*ast.SliceExpression); ok {
		g.genSliceExpression(e.Left, se)
		return
	}

	t := g.SemanticInfo.Types[e.Left]
	if t != nil {
		ut := t
		if pt, ok := t.(*types.PointerType); ok && !pt.IsArray {
			ut = pt.Base
		}

		// Struct operator overload: index
		if st, ok := ut.(*types.StructType); ok && len(e.Indices) == 1 {
			if methodType, exists := st.Methods["index"]; exists {
				if mt, ok := methodType.(*types.FunctionType); ok && len(mt.Params) == 1 {
					g.buf.WriteString(g.mangledTypeName(st) + "_index(NULL, ")
					g.emitArgument(e.Left, st, mt.ReceiverLease)
					g.buf.WriteString(", ")
					g.emitArgument(e.Indices[0], mt.Params[0], mt.ParamLeases[0])
					g.buf.WriteString(")")
					return
				}
			}
		}

		if ut.Name() == "str" {
			g.buf.WriteString("((char*)")
			g.genExpression(e.Left)
			g.buf.WriteString(")[")
			g.genExpression(e.Indices[0])
			g.buf.WriteString("]")
			return
		}

		if mt, ok := ut.(*types.MapType); ok {
			g.buf.WriteString("(*(")
			g.buf.WriteString(fmt.Sprintf("(%s*)", g.cType(mt.Value)))
			g.buf.WriteString("map_get(")
			g.genValueExpression(e.Left)
			g.buf.WriteString(", ")
			g.buf.WriteString(fmt.Sprintf("&(%s){", g.cType(mt.Key)))
			g.genExpression(e.Indices[0])
			g.buf.WriteString("})))")
			return
		}
	}

	// Bounds checking
	if e.NoBoundsCheck {
		g.buf.WriteString("(((")
		g.buf.WriteString(fmt.Sprintf("%s*)", g.cType(g.SemanticInfo.Types[e])))
		g.genValueExpression(e.Left)
		g.buf.WriteString(")[")
		g.genValueExpression(e.Indices[0])
		g.buf.WriteString("])")
		return
	}

	g.buf.WriteString("(*(")
	g.buf.WriteString(fmt.Sprintf("(%s*)", g.cType(g.SemanticInfo.Types[e])))
	g.buf.WriteString("array_bounds_check(")
	g.genValueExpression(e.Left)
	g.buf.WriteString(", ")
	g.genValueExpression(e.Indices[0])
	g.buf.WriteString(fmt.Sprintf(", \"%s\", %d)))", strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
}

func (g *Generator) genArrayLiteral(e *ast.ArrayLiteral) {
	oldNoTemp := g.NoTempWrap
	g.NoTempWrap = false
	defer func() { g.NoTempWrap = oldNoTemp }()
	lt, ok := g.SemanticInfo.Types[e].(*types.ListType)
	var elemType types.NRType
	if ok {
		elemType = lt.ElementType
	} else {
		elemType = types.I32 // Fallback
	}
	g.buf.WriteString(fmt.Sprintf("array_make(%d, sizeof(%s), \"%s\", %d", len(e.Elements), g.cType(elemType), strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
	oldTargetIsValue := g.TargetIsValue
	g.TargetIsValue = !g.isPointerTypeInC(elemType)
	for _, el := range e.Elements {
		g.buf.WriteString(", ")
		g.genOwnedValue(el, elemType)
	}
	g.TargetIsValue = oldTargetIsValue
	g.buf.WriteString(")")
}

func (g *Generator) isTempStringExpr(expr ast.Expression) bool {
	if _, ok := expr.(*ast.InterpolatedString); ok {
		return true
	}
	if infix, ok := expr.(*ast.InfixExpression); ok && infix.Operator == "+" {
		lt := g.SemanticInfo.Types[infix.Left]
		rt := g.SemanticInfo.Types[infix.Right]
		if (lt != nil && lt.Name() == "str") || (rt != nil && rt.Name() == "str") {
			return true
		}
	}
	return false
}

func (g *Generator) shouldDereferenceOwned(expr ast.Expression, targetType types.NRType) bool {
	if expr == nil || targetType == nil {
		return false
	}
	if _, ok := expr.(*ast.NoneLiteral); ok {
		return false
	}
	if prefix, ok := expr.(*ast.PrefixExpression); ok && prefix.Operator == "@" {
		return false
	}
	exprType := g.SemanticInfo.Types[expr]
	if exprType == nil {
		return false
	}

	exprIsValue := !g.isPointerInC(expr)
	if g.Solver != nil && g.Solver.Moves[expr] {
		exprIsValue = true
	}

	exprLevel := g.cPointerLevel(exprType, exprIsValue)
	targetLevel := g.cPointerLevel(targetType, g.TargetIsValue)

	if g.Solver != nil && g.Solver.Moves[expr] {
		if exprLevel > targetLevel {
			return true
		}
		return false
	}

	return exprLevel > targetLevel
}

func (g *Generator) isTemporaryHeapPointer(expr ast.Expression, exprType types.NRType) bool {
	if expr == nil {
		return false
	}
	if exprType == nil {
		exprType = g.SemanticInfo.Types[expr]
	}
	if exprType == nil {
		return false
	}

	// Must be a pointer type and heap-allocated (an owned pointer)
	_, isPointer := exprType.(*types.PointerType)
	if !isPointer {
		return false
	}
	if !g.isHeapAllocated(exprType) {
		return false
	}

	// Temporary expressions are anonymous and evaluated on the fly (e.g. not stored in a variable)
	switch expr.(type) {
	case *ast.CallExpression, *ast.ReceiveExpression, *ast.AllocExpression, *ast.IfExpression, *ast.MatchExpression, *ast.TryExpression:
		return true
	}
	return false
}

func (g *Generator) genOwnedValue(expr ast.Expression, targetType types.NRType) {
	oldNoTemp := g.NoTempWrap
	g.NoTempWrap = true

	if g.shouldDereferenceOwned(expr, targetType) {
		exprType := g.SemanticInfo.Types[expr]
		if g.isTemporaryHeapPointer(expr, exprType) {
			cTypeStr := g.cType(exprType) // e.g. "T*"
			baseType := exprType.(*types.PointerType).Base
			cBaseTypeStr := g.cType(baseType) // e.g. "T"

			g.buf.WriteString("({ ")
			g.buf.WriteString(fmt.Sprintf("%s _temp_ptr = ", cTypeStr))
			g.genExpression(expr)
			g.buf.WriteString("; ")
			g.buf.WriteString(fmt.Sprintf("%s _val; memset(&_val, 0, sizeof(_val)); ", cBaseTypeStr))
			g.buf.WriteString("if (_temp_ptr) { _val = *_temp_ptr; nr_free(_temp_ptr); } ")
			g.buf.WriteString("_val; })")
		} else {
			exprType := g.SemanticInfo.Types[expr]
			if exprType != nil && (exprType.Name() == "ptr" || types.UnwrapLease(exprType).Name() == "ptr") {
				g.buf.WriteString(fmt.Sprintf("(*(%s*)", g.cType(targetType)))
				g.genExpression(expr)
				g.buf.WriteString(")")
			} else {
				g.buf.WriteString("(*")
				g.genExpression(expr)
				g.buf.WriteString(")")
			}
		}
	} else {
		g.genExpression(expr)
	}

	g.NoTempWrap = oldNoTemp
}

func (g *Generator) genMapLiteral(e *ast.MapLiteral) {
	oldNoTemp := g.NoTempWrap
	g.NoTempWrap = false
	defer func() { g.NoTempWrap = oldNoTemp }()
	mt := g.SemanticInfo.Types[e].(*types.MapType)
	isStrKey := mt.Key.Name() == "str"
	g.buf.WriteString("({ ")
	g.buf.WriteString(fmt.Sprintf("void* _m = map_make(sizeof(%s), sizeof(%s), %v, \"%s\", %d); ", g.cType(mt.Key), g.cType(mt.Value), isStrKey, strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))

	// Sort keys for deterministic output
	type pair struct {
		k, v ast.Expression
	}
	var pairs []pair
	for k, v := range e.Pairs {
		pairs = append(pairs, pair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].k.String() < pairs[j].k.String()
	})

	for _, p := range pairs {
		g.buf.WriteString("map_set(_m, ")
		g.buf.WriteString(fmt.Sprintf("&(%s){", g.cType(mt.Key)))
		g.genOwnedValue(p.k, mt.Key)
		g.buf.WriteString("}, ")
		g.buf.WriteString(fmt.Sprintf("&(%s){", g.cType(mt.Value)))
		g.genOwnedValue(p.v, mt.Value)
		g.buf.WriteString("}); ")
	}
	g.buf.WriteString("_m; })")
}

func (g *Generator) genStructLiteral(e *ast.StructLiteral) {
	oldNoTemp := g.NoTempWrap
	g.NoTempWrap = false
	defer func() { g.NoTempWrap = oldNoTemp }()
	t := g.SemanticInfo.Types[e]
	ctype := g.cType(t)
	if ctype == "void" && e.Name != nil {
		if ident, ok := e.Name.(*ast.Identifier); ok {
			ctype = ident.Value
		}
	}
	g.buf.WriteString(fmt.Sprintf("(%s){", ctype))
	for i, field := range e.Fields {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(fmt.Sprintf(".%s = ", field.Name.Value))
		if field.Value != nil {
			var fieldType types.NRType
			if st, ok := t.(*types.StructType); ok {
				fieldType = st.Fields[field.Name.Value]
			}
			oldTargetIsValue := g.TargetIsValue
			g.TargetIsValue = !g.isPointerTypeInC(fieldType)
			g.genOwnedValue(field.Value, fieldType)
			g.TargetIsValue = oldTargetIsValue
		}
	}
	g.buf.WriteString("}")
}

func (g *Generator) genSliceExpression(left ast.Expression, e *ast.SliceExpression) {
	t := g.SemanticInfo.Types[left]
	if t != nil && t.Name() == "str" {
		g.buf.WriteString("string_slice(")
		g.genExpression(left)
		g.buf.WriteString(", ")
		if e.Start != nil {
			g.genExpression(e.Start)
		} else {
			g.buf.WriteString("0")
		}
		g.buf.WriteString(", ")
		if e.End != nil {
			g.genExpression(e.End)
		} else {
			g.buf.WriteString("-1")
		}
		g.buf.WriteString(")")
		return
	}

	var elemType types.NRType
	if lt, ok := t.(*types.ListType); ok {
		elemType = lt.ElementType
	} else {
		elemType = types.I32
	}
	g.buf.WriteString("array_slice(")
	g.genExpression(left)
	g.buf.WriteString(", ")
	if e.Start != nil {
		g.genExpression(e.Start)
	} else {
		g.buf.WriteString("0")
	}
	g.buf.WriteString(", ")
	if e.End != nil {
		g.genExpression(e.End)
	} else {
		g.buf.WriteString("-1")
	}
	g.buf.WriteString(fmt.Sprintf(", sizeof(%s))", g.cType(elemType)))
}

func (g *Generator) unwrapSpawnArgType(t types.NRType) types.NRType {
	if t == nil {
		return nil
	}
	if pt, ok := t.(*types.PointerType); ok && !pt.IsArray {
		baseKind := pt.Base.GetKind()
		if baseKind == types.KindStruct || baseKind == types.KindSum || baseKind == types.KindProtocol {
			return g.unwrapSpawnArgType(pt.Base)
		}
	}
	return t
}

func (g *Generator) genSpawnExpression(e *ast.SpawnExpression) {
	if e.Call == nil {
		g.buf.WriteString("scheduler_spawn(NULL, NULL)")
		return
	}

	g.spawnCounter++
	structName := fmt.Sprintf("__spawn_args_%d", g.spawnCounter)
	wrapperName := fmt.Sprintf("__spawn_wrapper_%d", g.spawnCounter)

	// 1. Generate Struct
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("struct %s {\n", structName))

	hasScope := len(g.ScopeWaitgroups) > 0
	if hasScope {
		sb.WriteString("    void* _scope_wg;\n")
	}

	for i, arg := range e.Call.Arguments {
		t := g.SemanticInfo.Types[arg.Value]
		targetType := g.unwrapSpawnArgType(t)
		sb.WriteString(fmt.Sprintf("    %s arg%d;\n", g.cType(targetType), i))
	}
	sb.WriteString("};\n")
	g.SpawnStructs = append(g.SpawnStructs, sb.String())

	// 2. Generate Wrapper
	sb.Reset()
	if g.EnableDebug {
		pos := e.Pos()
		filename := normalizeDebugPath(pos.Filename)
		sb.WriteString(fmt.Sprintf("#line %d \"%s\"\n", pos.Line, filename))
	}
	sb.WriteString(fmt.Sprintf("static void %s(void* p) {\n", wrapperName))
	if g.EnableDebug {
		sb.WriteString("    if (nr_fiber_current() == __nora_step_target_fiber) {\n")
		sb.WriteString("        __nora_step_target_fiber = NULL;\n")
		sb.WriteString("        NR_DEBUGBREAK();\n")
		sb.WriteString("    }\n")
		sb.WriteString("    __nora_fiber_started(nr_fiber_parent());\n")
		pos := e.Pos()
		filename := normalizeDebugPath(pos.Filename)
		sb.WriteString(fmt.Sprintf("#line %d \"%s\"\n", pos.Line, filename))
	}
	sb.WriteString(fmt.Sprintf("    struct %s* args = (struct %s*)p;\n", structName, structName))

	// Call the function
	oldBuf := g.buf
	g.buf = new(bytes.Buffer)
	oldInCall := g.InCallExpression
	g.InCallExpression = true
	g.genExpression(e.Call.Function)
	g.InCallExpression = oldInCall
	fnName := g.buf.String()
	g.buf = oldBuf

	var ft *types.FunctionType
	if t, ok := g.SemanticInfo.Types[e.Call.Function].(*types.FunctionType); ok {
		ft = t
	}

	isExtern := false
	if ident, ok := e.Call.Function.(*ast.Identifier); ok {
		if sym := g.SemanticInfo.Uses[ident]; sym != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
				isExtern = true
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			}
			if cStdlibFunctions[sym.Name] && sym.Name == g.mangleName(sym) {
				isExtern = true
			}
		}
	} else if sel, ok := e.Call.Function.(*ast.SelectorExpression); ok {
		if sym := g.SemanticInfo.Uses[sel.Field]; sym != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
				isExtern = true
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			}
		}
	}

	sb.WriteString("    void* self = nr_fiber_current();\n")
	sb.WriteString("    if (self && setjmp(*nr_fiber_panic_buf_ptr(self)) != 0) {\n")
	if hasScope {
		sb.WriteString("        nr_sync_waitgroup_panic(args->_scope_wg, nr_fiber_panic_msg(self), nr_fiber_panic_file(self), nr_fiber_panic_line(self));\n")
	}
	sb.WriteString("    } else {\n")
	sb.WriteString(fmt.Sprintf("        %s(", fnName))
	if !isExtern {
		sb.WriteString("NULL")
		if len(e.Call.Arguments) > 0 {
			sb.WriteString(", ")
		}
	}
	for i := range e.Call.Arguments {
		if i > 0 {
			sb.WriteString(", ")
		}
		t := g.SemanticInfo.Types[e.Call.Arguments[i].Value]
		unwrapped := g.unwrapSpawnArgType(t)
		paramCType := ""
		if ft != nil && i < len(ft.Params) {
			lease := types.LeaseRead
			if i < len(ft.ParamLeases) {
				lease = ft.ParamLeases[i]
			}
			paramCType = g.cParamType(ft.Params[i], lease)
		}
		memberCType := g.cType(unwrapped)
		if strings.HasSuffix(paramCType, "*") && !strings.HasSuffix(memberCType, "*") {
			sb.WriteString(fmt.Sprintf("&args->arg%d", i))
		} else {
			sb.WriteString(fmt.Sprintf("args->arg%d", i))
		}
	}
	sb.WriteString(");\n")
	sb.WriteString("    }\n")
	// Clean up arguments if they are owned types (like channels)
	for i, arg := range e.Call.Arguments {
		t := g.SemanticInfo.Types[arg.Value]
		if g.isChanType(t) {
			sb.WriteString(fmt.Sprintf("    channel_free(args->arg%d);\n", i))
		}
	}
	sb.WriteString("    nr_flush_temps();\n")
	if hasScope {
		sb.WriteString("    nr_sync_waitgroup_done(args->_scope_wg);\n")
	}
	sb.WriteString("    nr_free_untracked(args);\n")
	sb.WriteString("}\n")
	g.SpawnWrappers = append(g.SpawnWrappers, sb.String())

	// 3. Emit the call to scheduler_spawn
	g.buf.WriteString(fmt.Sprintf("({ struct %s* _args = (struct %s*)nr_malloc_untracked(sizeof(struct %s)); ", structName, structName, structName))
	for i, arg := range e.Call.Arguments {
		targetType := g.unwrapSpawnArgType(g.SemanticInfo.Types[arg.Value])
		targetCType := g.cType(targetType)

		// Get the expression representation
		oldBuf := g.buf
		g.buf = new(bytes.Buffer)
		oldTargetIsValue := g.TargetIsValue
		g.TargetIsValue = !g.isPointerTypeInC(targetType)
		g.genExpression(arg.Value)
		g.TargetIsValue = oldTargetIsValue
		valStr := g.buf.String()
		g.buf = oldBuf

		operandCType := g.cType(g.SemanticInfo.Types[arg.Value])
		// If TargetIsValue was true, and the expression is a move (@), it generated a value, not a pointer.
		if prefix, ok := arg.Value.(*ast.PrefixExpression); ok && prefix.Operator == "@" && !g.isPointerTypeInC(targetType) {
			operandCType = g.cType(targetType)
		}

		// Count asterisks to align pointers
		targetStars := 0
		for idx := len(targetCType) - 1; idx >= 0; idx-- {
			if targetCType[idx] == '*' {
				targetStars++
			} else if targetCType[idx] != ' ' {
				break
			}
		}
		operandStars := 0
		for idx := len(operandCType) - 1; idx >= 0; idx-- {
			if operandCType[idx] == '*' {
				operandStars++
			} else if operandCType[idx] != ' ' {
				break
			}
		}

		if operandStars > targetStars {
			if targetType.Name() != "ptr" && targetType.Name() != "str" {
				valStr = strings.Repeat("*", operandStars-targetStars) + valStr
			}
		} else if targetStars > operandStars {
			valStr = "&(" + valStr + ")"
		}

		g.emit(fmt.Sprintf("    _args->arg%d = %s;\n", i, valStr))

		// If it's a channel, increment ref count
		if g.isChanType(targetType) {
			g.emit(fmt.Sprintf("    channel_ref(_args->arg%d);\n", i))
		}
	}
	funcName := "anonymous"
	if id, ok := e.Call.Function.(*ast.Identifier); ok {
		funcName = id.Value
	} else if sel, ok := e.Call.Function.(*ast.SelectorExpression); ok {
		funcName = sel.Field.Value
	}

	if len(g.ScopeWaitgroups) > 0 {
		activeWg := g.ScopeWaitgroups[len(g.ScopeWaitgroups)-1]
		g.buf.WriteString(fmt.Sprintf("_args->_scope_wg = %s; ", activeWg))
		g.buf.WriteString(fmt.Sprintf("nr_sync_waitgroup_add(%s, 1); ", activeWg))
	}

	g.buf.WriteString(fmt.Sprintf("scheduler_spawn(%s, _args, \"%s\", \"%s\", %d); })", wrapperName, funcName, strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
}

func (g *Generator) genReceiveExpression(e *ast.ReceiveExpression) {
	g.buf.WriteString("({ ")
	t := g.SemanticInfo.Types[e]
	g.buf.WriteString(fmt.Sprintf("%s _res; ", g.cType(t)))
	g.buf.WriteString("channel_recv(")
	if g.shouldDereferenceInC(e.Value) {
		g.buf.WriteString("*")
	}
	g.genExpression(e.Value)
	g.buf.WriteString(", &_res); _res; })")
}

func (g *Generator) genSendExpression(e *ast.SendExpression) {
	t := g.SemanticInfo.Types[e.Right]
	if t == nil {
		t = types.I32
	}
	g.buf.WriteString("({ ")
	g.buf.WriteString("channel_t* _c = ")
	if g.shouldDereferenceInC(e.Left) {
		g.buf.WriteString("*")
	}
	g.genExpression(e.Left)
	g.buf.WriteString("; ")
	g.buf.WriteString(fmt.Sprintf("%s _send_val = ", g.cType(t)))
	g.genExpression(e.Right)
	g.buf.WriteString("; channel_send(_c, &_send_val); })")
}

func (g *Generator) genParallelExpression(e *ast.ParallelExpression) {
	numStmts := len(e.Body.Statements)
	g.buf.WriteString("({ ")

	isResult := false
	var cResultTypeName string
	if st, ok := g.SemanticInfo.Types[e].(*types.SumType); ok && st.CoreIntrinsic == "Result" {
		isResult = true
		cResultTypeName = g.cType(st)
	}

	if isResult {
		g.buf.WriteString(fmt.Sprintf("channel_t* _wg = channel_make(%d, sizeof(char*)); ", numStmts))
	} else {
		g.buf.WriteString(fmt.Sprintf("channel_t* _wg = channel_make(%d, sizeof(int)); ", numStmts))
	}

	for _, stmt := range e.Body.Statements {
		g.spawnCounter++
		wrapperName := fmt.Sprintf("__parallel_wrapper_%d", g.spawnCounter)

		// Generate wrapper
		var sb strings.Builder
		if g.EnableDebug {
			pos := stmt.Pos()
			filename := normalizeDebugPath(pos.Filename)
			sb.WriteString(fmt.Sprintf("#line %d \"%s\"\n", pos.Line, filename))
		}
		sb.WriteString(fmt.Sprintf("static void %s(void* p) {\n", wrapperName))
		if g.EnableDebug {
			sb.WriteString("    if (nr_fiber_current() == __nora_step_target_fiber) {\n")
			sb.WriteString("        __nora_step_target_fiber = NULL;\n")
			sb.WriteString("        NR_DEBUGBREAK();\n")
			sb.WriteString("    }\n")
			sb.WriteString("    __nora_fiber_started(nr_fiber_parent());\n")
			pos := stmt.Pos()
			filename := normalizeDebugPath(pos.Filename)
			sb.WriteString(fmt.Sprintf("#line %d \"%s\"\n", pos.Line, filename))
		}
		sb.WriteString("    channel_t* _wg = (channel_t*)p;\n")

		stmtType := g.SemanticInfo.Types[stmt]
		var exprStmt *ast.ExpressionStatement
		var ok bool
		if exprStmt, ok = stmt.(*ast.ExpressionStatement); ok {
			stmtType = g.SemanticInfo.Types[exprStmt.Expression]
		}

		isStmtResult := false
		var stmtResultTypeName string
		if st, ok := stmtType.(*types.SumType); ok && st.CoreIntrinsic == "Result" {
			isStmtResult = true
			stmtResultTypeName = g.cType(st)
		}

		sb.WriteString("    void* self = nr_fiber_current();\n")
		sb.WriteString("    if (self && setjmp(*nr_fiber_panic_buf_ptr(self)) != 0) {\n")
		if isResult {
			sb.WriteString("        char* _err = (char*)nr_fiber_panic_msg(self);\n")
			sb.WriteString("        channel_send(_wg, &_err);\n")
		} else {
			sb.WriteString("        int _zero = -1;\n")
			sb.WriteString("        channel_send(_wg, &_zero);\n")
		}
		sb.WriteString("    } else {\n")

		if isStmtResult && exprStmt != nil {
			oldBuf := g.buf
			g.buf = new(bytes.Buffer)
			g.genExpression(exprStmt.Expression)
			exprStr := g.buf.String()
			g.buf = oldBuf

			sb.WriteString(fmt.Sprintf("        %s _sub_res = %s;\n", stmtResultTypeName, exprStr))
			sb.WriteString("        char* _err = NULL;\n")
			sb.WriteString("        if (_sub_res.tag == 0) {\n") // TAG_Err is 0
			sb.WriteString("            _err = _sub_res.data.Err;\n")
			sb.WriteString("        }\n")
			sb.WriteString("        channel_send(_wg, &_err);\n")
		} else {
			oldBuf := g.buf
			g.buf = new(bytes.Buffer)
			g.genStatement(stmt)
			// Indent the statement block
			stmtStr := strings.ReplaceAll(g.buf.String(), "\n", "\n        ")
			sb.WriteString("        " + stmtStr)
			g.buf = oldBuf

			if isResult {
				sb.WriteString("\n        char* _err = NULL;\n")
				sb.WriteString("        channel_send(_wg, &_err);\n")
			} else {
				sb.WriteString("\n        int _zero = 0;\n")
				sb.WriteString("        channel_send(_wg, &_zero);\n")
			}
		}
		sb.WriteString("    }\n")

		sb.WriteString("}\n")
		g.SpawnWrappers = append(g.SpawnWrappers, sb.String())

		g.buf.WriteString(fmt.Sprintf("scheduler_spawn(%s, _wg, \"parallel_block\", \"%s\", %d); ", wrapperName, strings.ReplaceAll(stmt.Pos().Filename, "\\", "/"), stmt.Pos().Line))
	}

	if isResult {
		g.buf.WriteString("char* _first_err = NULL; ")
		g.buf.WriteString(fmt.Sprintf("for (int _i = 0; _i < %d; _i++) { ", numStmts))
		g.buf.WriteString("    char* _task_err; ")
		g.buf.WriteString("    channel_recv(_wg, &_task_err); ")
		g.buf.WriteString("    if (_task_err != NULL && _first_err == NULL) { _first_err = _task_err; } ")
		g.buf.WriteString("} ")
		g.buf.WriteString("channel_free(_wg); ")
		g.buf.WriteString(fmt.Sprintf("_first_err != NULL ? %s_Err_make(_first_err) : %s_Ok_make(true); })", cResultTypeName, cResultTypeName))
	} else {
		g.buf.WriteString("int _dummy; ")
		g.buf.WriteString(fmt.Sprintf("for (int _i = 0; _i < %d; _i++) channel_recv(_wg, &_dummy); ", numStmts))
		g.buf.WriteString("channel_free(_wg); })")
	}
}

// --- MACRO EXPANSION ---

func (g *Generator) tryExpandMacro(call *ast.CallExpression) bool {
	var sym *semantic.Symbol
	switch fn := call.Function.(type) {
	case *ast.Identifier:
		sym = g.SemanticInfo.Uses[fn]
	case *ast.SelectorExpression:
		sym = g.SemanticInfo.Uses[fn.Field]
	}

	if sym == nil || sym.Kind != semantic.SymFunc || sym.DefNode == nil {
		return false
	}

	fnStmt, ok := sym.DefNode.(*ast.FunctionStatement)
	if !ok || ast.GetAttribute(fnStmt.Attributes, "macro") == nil {
		return false
	}

	expanded := g.expandMacro(sym, call)
	if expanded == "" {
		return false
	}

	g.buf.WriteString(expanded)
	return true
}

func (g *Generator) expandMacro(sym *semantic.Symbol, call *ast.CallExpression) string {
	if g.PluginMgr == nil {
		return ""
	}

	req := api.CallRequest{
		FunctionName: sym.Name,
		Arguments:    make([]api.CallArgumentDTO, len(call.Arguments)),
	}

	for i, arg := range call.Arguments {
		oldBuf := g.buf
		g.buf = new(bytes.Buffer)
		g.genExpression(arg.Value)
		req.Arguments[i] = api.CallArgumentDTO{
			Value: g.buf.String(),
			Type:  g.SemanticInfo.Types[arg.Value].Name(),
		}
		g.buf = oldBuf
	}

	pluginName := g.getSymbolPackage(sym)
	macroName := sym.Name

	if sym.DefNode != nil {
		if fn, ok := sym.DefNode.(*ast.FunctionStatement); ok {
			if attr := ast.GetAttribute(fn.Attributes, "macro"); attr != nil && len(attr.Args) >= 2 {
				pluginName = attr.Args[0]
				macroName = attr.Args[1]
			}
		}
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.Encode(req)
	reqJSON := bytes.TrimSpace(buf.Bytes())

	resJSON, err := g.PluginMgr.ExecuteMacro(pluginName, macroName, reqJSON)
	if err != nil {
		return fmt.Sprintf("/* macro error: %v */", err)
	}

	var resp api.CallResponse
	if err := json.Unmarshal(resJSON, &resp); err != nil {
		return fmt.Sprintf("/* macro json error: %v */", err)
	}

	return resp.ReplacementCode
}

func (g *Generator) genAllocExpression(e *ast.AllocExpression) {
	if e.Type == nil {
		g.buf.WriteString("NULL")
		return
	}

	// If it's an array allocation: alloc T[n] or alloc @T[n]
	if pt, ok := e.Type.(*types.PointerType); ok && pt.IsArray {
		// Calculate size
		var countExpr ast.Expression
		if ie, ok := e.Value.(*ast.IndexExpression); ok {
			countExpr = ie.Indices[0]
		} else if pe, ok := e.Value.(*ast.PrefixExpression); ok {
			if ie, ok := pe.Right.(*ast.IndexExpression); ok {
				countExpr = ie.Indices[0]
			}
		}

		if countExpr != nil {
			g.buf.WriteString("({ ")
			g.buf.WriteString("int _n = ")
			g.genExpression(countExpr)
			g.buf.WriteString("; ")

			elemType := pt.Base
			cElemType := g.cType(elemType)

			g.buf.WriteString(fmt.Sprintf("void* _p = nr_malloc_debug(_n * sizeof(%s), \"%s\", %d); ", cElemType, strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))

			// Set up Nora collection header (size, capacity, etc. at negative offsets)
			g.buf.WriteString("nr_header_t* _h = (nr_header_t*)((char*)_p - NR_HEADER_SIZE); ")
			g.buf.WriteString("_h->count = _n; ")
			g.buf.WriteString(fmt.Sprintf("_h->elem_size = sizeof(%s); ", cElemType))
			g.buf.WriteString("_h->magic = NR_HEADER_MAGIC; ")

			g.buf.WriteString(fmt.Sprintf("%s* _data = (%s*)_p; ", cElemType, cElemType))
			g.buf.WriteString(fmt.Sprintf("memset(_data, 0, _n * sizeof(%s)); ", cElemType))
			g.buf.WriteString("_data; ")
			g.buf.WriteString("})")
			return
		}
	}

	// Single value allocation: alloc Point{x: 1, y: 2}
	g.buf.WriteString("({ ")
	pt, ok := e.Type.(*types.PointerType)
	var ct string
	var baseType types.NRType
	if ok {
		ct = g.cType(pt.Base)
		baseType = pt.Base
	} else {
		ct = g.cType(e.Type)
		baseType = e.Type
	}
	g.buf.WriteString(fmt.Sprintf("%s* _p = nr_malloc_debug(sizeof(%s), \"%s\", %d); ", ct, ct, strings.ReplaceAll(e.Pos().Filename, "\\", "/"), e.Pos().Line))
	g.buf.WriteString("*_p = ")
	oldTargetIsValue := g.TargetIsValue
	g.TargetIsValue = !g.isPointerTypeInC(baseType)
	g.genOwnedValue(e.Value, baseType)
	g.TargetIsValue = oldTargetIsValue
	g.buf.WriteString("; _p; })")
}

func (g *Generator) genInterpolatedString(e *ast.InterpolatedString) {
	if len(e.Parts) == 0 {
		g.buf.WriteString("nr_strdup(\"\")")
		return
	}

	if len(e.Parts) == 1 {
		g.buf.WriteString("nr_temp_str(")
		g.genInterpolatedPart(e.Parts[0])
		g.buf.WriteString(")")
		return
	}

	oldNoTemp := g.NoTempWrap
	g.NoTempWrap = true

	oldBuf := g.buf
	g.buf = new(bytes.Buffer)

	// First pair
	g.buf.WriteString("nr_str_concat_free(")
	g.genInterpolatedPart(e.Parts[0])
	g.buf.WriteString(", ")
	g.genInterpolatedPart(e.Parts[1])
	g.buf.WriteString(fmt.Sprintf(", %v, %v)", g.isTempInterpolatedPart(e.Parts[0]), g.isTempInterpolatedPart(e.Parts[1])))

	for i := 2; i < len(e.Parts); i++ {
		current := g.buf.String()
		g.buf.Reset()
		g.buf.WriteString("nr_str_concat_free(")
		g.buf.WriteString(current)
		g.buf.WriteString(", ")
		g.genInterpolatedPart(e.Parts[i])
		g.buf.WriteString(fmt.Sprintf(", true, %v)", g.isTempInterpolatedPart(e.Parts[i])))
	}

	final := g.buf.String()
	g.buf = oldBuf
	g.NoTempWrap = oldNoTemp

	if g.NoTempWrap {
		g.buf.WriteString(final)
	} else {
		g.buf.WriteString("nr_temp_str(")
		g.buf.WriteString(final)
		g.buf.WriteString(")")
	}
}

func (g *Generator) isTempInterpolatedPart(e ast.Expression) bool {
	t := g.SemanticInfo.Types[e]
	if t == nil {
		return false
	}
	ut := types.UnwrapLease(t)
	if ut.Name() != "str" {
		// Non-string types are converted via nr_xxx_to_str which dynamically allocates a temp string
		return true
	}
	if t.IsLeased() {
		return false
	}

	// If the solver says this expression is moved (ownership transfer), we must free it!
	if g.Solver != nil && g.Solver.Moves[e] {
		return true
	}

	// At this point, it is either a natively typed string, or a fallback.
	// If it's a string identifier/literal, we don't free it (it's managed by RAII or static)
	switch e.(type) {
	case *ast.StringLiteral, *ast.Identifier, *ast.IndexExpression, *ast.SelectorExpression:
		return false
	}

	// Complex string expressions (like concatenation results) are temps
	return true
}

func (g *Generator) genInterpolatedPart(expr ast.Expression) {
	t := g.SemanticInfo.Types[expr]
	if t != nil {
		t = types.UnwrapLease(t)
	}
	if t != nil && t.Name() == "str" {
		g.genExpression(expr)
		return
	}

	// For non-string types, we need a conversion helper
	if t != nil {
		switch t.Name() {
		case "i32", "int":
			g.buf.WriteString("nr_i32_to_str(")
			g.genExpression(expr)
			g.buf.WriteString(")")
			return
		case "i64":
			g.buf.WriteString("nr_i64_to_str(")
			g.genExpression(expr)
			g.buf.WriteString(")")
			return
		case "f64":
			g.buf.WriteString("nr_f64_to_str(")
			g.genExpression(expr)
			g.buf.WriteString(")")
			return
		case "bool":
			g.buf.WriteString("nr_bool_to_str(")
			g.genExpression(expr)
			g.buf.WriteString(")")
			return
		case "byte", "i8":
			g.buf.WriteString("nr_i32_to_str((int32_t)(")
			g.genExpression(expr)
			g.buf.WriteString("))")
			return
		}
	}

	// Fallback/Generic
	g.buf.WriteString("nr_to_str(")
	g.genExpression(expr)
	g.buf.WriteString(")")
}

func (g *Generator) genRuneLiteral(e *ast.RuneLiteral) {
	g.buf.WriteString(fmt.Sprintf("%d", e.Value))
}

// exprToString generates an expression into a temporary string without modifying g.buf.
func (g *Generator) exprToString(expr ast.Expression) string {
	saved := g.buf
	g.buf = new(bytes.Buffer)
	g.genExpression(expr)
	result := g.buf.String()
	g.buf = saved
	return result
}
func (g *Generator) genTryExpression(e *ast.TryExpression) {
	g.buf.WriteString("({ ")
	valType := g.SemanticInfo.Types[e.Value]
	st, ok := valType.(*types.SumType)
	if !ok {
		g.buf.WriteString("/* error: Try on non-SumType type */ 0")
		g.buf.WriteString("; })")
		return
	}

	isResult := st.CoreIntrinsic == "Result"
	isOption := st.CoreIntrinsic == "Option"

	if !isResult && !isOption {
		g.buf.WriteString("/* error: Try on non-Result/Option type */ 0")
		g.buf.WriteString("; })")
		return
	}

	// Result/Option struct name
	resName := g.cType(valType)

	g.buf.WriteString(fmt.Sprintf("%s _res = ", resName))
	g.genExpression(e.Value)
	g.emit(";")

	// Early return condition
	if isResult {
		g.emit("if (_res.tag == %s_TAG_Err) {", g.mangledTypeName(st))
	} else if isOption {
		g.emit("if (_res.tag == %s_TAG_None) {", g.mangledTypeName(st))
	}

	// Emit RAII drops for early return
	if g.Solver != nil && g.Solver.TryDrops[e] != nil {
		for _, drop := range g.Solver.TryDrops[e] {
			sym := drop.Symbol
			isPtr := g.isPointerTypeInC(sym.Type)
			g.emitDrop(g.variableName(sym), sym.Type, isPtr)
		}
	}

	// Construct return value for current function
	if g.CurrentFunc != nil {
		ft := g.CurrentFunc.Type.(*types.FunctionType)
		retType := ft.Return
		retResName := g.cType(retType)

		if isResult {
			// Return Result_Err_make(_res.data.Err)
			g.emit("    return %s_Err_make(_res.data.Err);", retResName)
		} else {
			// Return Option_None_make()
			g.emit("    return %s_None_make();", retResName)
		}
	} else {
		g.emit("    return; /* error: no current function */")
	}
	g.emit("}")

	// Unwrapped value
	if isResult {
		// Ok variant has one field (Value), accessed via data.Ok in C
		g.emit("_res.data.Ok;")
	} else {
		// Some variant has one field (val), accessed via data.Some in C
		g.emit("_res.data.Some;")
	}
	g.buf.WriteString(" })")
}
func (g *Generator) genInterfaceCast(expr ast.Expression, targetProto *types.ProtocolType) {
	valType := g.SemanticInfo.Types[expr]
	if valType == nil {
		if ident, ok := expr.(*ast.Identifier); ok {
			sym := g.SemanticInfo.Uses[ident]
			if sym == nil {
				sym = g.SemanticInfo.Defs[ident]
			}
			if sym == nil {
				sym = g.findSymbolByName(ident.Value)
			}
			if sym != nil {
				valType = sym.Type
			}
		}
	}
	if valType == nil {
		valType = types.ErrorType
	}

	unwrappedValType := types.UnwrapLease(valType)
	exprIsValue := !g.isPointerInC(expr)
	if g.Solver != nil && g.Solver.Moves[expr] {
		exprIsValue = true
	}

	if types.Equals(targetProto, unwrappedValType) {
		if !exprIsValue && g.TargetIsValue {
			g.buf.WriteString("*")
		}
		g.genExpression(expr)
		return
	}
	if _, ok := unwrappedValType.(*types.ProtocolType); ok {
		g.buf.WriteString(fmt.Sprintf("(%s){ .data = ", g.cType(targetProto)))
		g.genExpression(expr)
		if !exprIsValue && g.TargetIsValue {
			g.buf.WriteString("->data, .vtable = ")
		} else {
			g.buf.WriteString(".data, .vtable = ")
		}
		g.genExpression(expr)
		if !exprIsValue && g.TargetIsValue {
			g.buf.WriteString("->vtable }")
		} else {
			g.buf.WriteString(".vtable }")
		}
		return
	}

	vtableName := g.requestVTable(valType, targetProto)

	g.buf.WriteString(fmt.Sprintf("(%s){ .data = ", g.cType(targetProto)))

	// For data pointer, we need the address of the concrete value.
	isPointerType := g.isPointerTypeInC(valType)
	isAddrOf := false
	if ue, ok := expr.(*ast.PrefixExpression); ok && (ue.Operator == "&" || ue.Operator == "#") {
		isAddrOf = true
	}
	isCptr := g.isPointerInC(expr) || isPointerType || isAddrOf
	if ue, ok := expr.(*ast.PrefixExpression); ok && ue.Operator == "@" {
		operandType := g.SemanticInfo.Types[ue.Right]
		if operandType == nil {
			if id, ok := ue.Right.(*ast.Identifier); ok {
				if sym := g.SemanticInfo.Uses[id]; sym != nil {
					operandType = sym.Type
				} else if sym := g.SemanticInfo.Defs[id]; sym != nil {
					operandType = sym.Type
				}
			}
		}
		if operandType != nil && g.isPointerTypeInC(operandType) {
			isCptr = true
		} else {
			underlyingType := types.UnwrapLease(valType)
			if pt, ok := underlyingType.(*types.PointerType); ok {
				underlyingType = pt.Base
			}
			if !g.isPointerTypeInC(underlyingType) {
				isCptr = false
			}
		}
	}
	if isCptr && (!g.TargetIsValue || isPointerType || isAddrOf) {
		g.genExpression(expr)
	} else {
		castType := valType
		if pt, ok := valType.(*types.PointerType); ok && !pt.IsArray {
			castType = pt.Base
		}
		ct := g.cType(castType)
		g.buf.WriteString(fmt.Sprintf("({ %s* _p = nr_malloc_debug(sizeof(%s), \"%s\", %d); *_p = ", ct, ct, strings.ReplaceAll(expr.Pos().Filename, "\\", "/"), expr.Pos().Line))
		g.genExpression(expr)
		g.buf.WriteString("; _p; })")
	}

	g.buf.WriteString(fmt.Sprintf(", .vtable = %s }", vtableName))
}
func (g *Generator) getMethodIndex(proto *types.ProtocolType, mName string) int {
	names := []string{}
	for name := range proto.Methods {
		names = append(names, name)
	}
	sort.Strings(names)
	for i, name := range names {
		if name == mName {
			return i + 1
		}
	}
	return -1
}

func (g *Generator) getCFunctionPointerTypeWithSelf(ft *types.FunctionType) string {
	res := g.cType(ft.Return) + " (*)"
	res += "(void*, void*"
	for _, p := range ft.Params {
		res += ", " + g.cType(p)
	}
	res += ")"
	return res
}

func (g *Generator) genInterfaceCall(e *ast.CallExpression, sel *ast.SelectorExpression, proto *types.ProtocolType) {
	mName := sel.Field.Value
	mType := proto.Methods[mName]
	idx := g.getMethodIndex(proto, mName)

	isPtr := g.isPointerInC(sel.Left)

	g.buf.WriteString("((")
	g.buf.WriteString(g.getCFunctionPointerTypeWithSelf(mType))
	g.buf.WriteString(")")
	g.genExpression(sel.Left)
	if isPtr {
		g.buf.WriteString(fmt.Sprintf("->vtable[%d])(NULL, ", idx))
	} else {
		g.buf.WriteString(fmt.Sprintf(".vtable[%d])(NULL, ", idx))
	}

	// Self
	g.genExpression(sel.Left)
	if isPtr {
		g.buf.WriteString("->data")
	} else {
		g.buf.WriteString(".data")
	}

	for i, arg := range e.Arguments {
		g.buf.WriteString(", ")
		var targetType types.NRType
		lease := types.LeaseRead
		if mType != nil {
			if i < len(mType.Params) {
				targetType = mType.Params[i]
			}
			if i < len(mType.ParamLeases) {
				lease = mType.ParamLeases[i]
			}
		}
		g.emitArgument(arg.Value, targetType, lease)
	}
	g.buf.WriteString(")")
}

func (g *Generator) genScopeExpression(e *ast.ScopeExpression) {
	g.scopeCounter++
	wgName := fmt.Sprintf("__scope_wg_%d", g.scopeCounter)

	g.buf.WriteString("({ ")
	g.buf.WriteString(fmt.Sprintf("void* %s = nr_sync_waitgroup_create(); ", wgName))

	g.ScopeWaitgroups = append(g.ScopeWaitgroups, wgName)

	oldBlock := g.CurrentBlock
	oldIdx := g.CurrentStmtIndex
	g.CurrentBlock = e.Body

	g.emitDropsAt(e.Body, 0)
	for i, stmt := range e.Body.Statements {
		g.CurrentStmtIndex = i
		g.emitPreDropsAt(e.Body, i)
		g.genStatement(stmt)
		g.emitDropsAt(e.Body, i+1)
	}

	g.CurrentBlock = oldBlock
	g.CurrentStmtIndex = oldIdx

	g.ScopeWaitgroups = g.ScopeWaitgroups[:len(g.ScopeWaitgroups)-1]

	g.buf.WriteString(fmt.Sprintf("nr_sync_waitgroup_wait(%s); ", wgName))
	g.buf.WriteString(fmt.Sprintf("nr_sync_waitgroup_destroy(%s); ", wgName))
	g.buf.WriteString(" })")
}
