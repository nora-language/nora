package codegen

import (
	"fmt"
	"strings"

	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/topology"
	"github.com/nora-language/nora/pkg/types"
)

func (g *Generator) genStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	g.emitLine(stmt)

	switch s := stmt.(type) {
	case *ast.BlockStatement:
		g.genBlock(s)
	case *ast.VarStatement:
		g.genVarStatement(s)
	case *ast.AssignmentStatement:
		g.genAssignmentStatement(s)
	case *ast.ReturnStatement:
		g.genReturnStatement(s)
	case *ast.ForStatement:
		g.genForStatement(s)
	case *ast.WhileStatement:
		g.genWhileStatement(s)
	case *ast.ExpressionStatement:
		g.genExpressionStatement(s)
	case *ast.DeferStatement:
		g.genDeferStatement(s)
	case *ast.SelectStatement:
		g.genSelectStatement(s)
	case *ast.BreakStatement:
		g.emit("break;")
	case *ast.ContinueStatement:
		g.emit("continue;")
	}

	g.emit("nr_flush_temps();")
}

func (g *Generator) genBlock(b *ast.BlockStatement) {
	g.genBlockWithTarget(b, "")
}

func (g *Generator) genBlockWithTarget(b *ast.BlockStatement, targetVar string) {
	g.emit("{")
	g.emitDropsAt(b, 0)

	// Save current block context
	oldBlock := g.CurrentBlock
	oldIdx := g.CurrentStmtIndex
	g.CurrentBlock = b

	for i, stmt := range b.Statements {
		g.CurrentStmtIndex = i
		g.emitPreDropsAt(b, i)
		if i == len(b.Statements)-1 && targetVar != "" {
			if es, ok := stmt.(*ast.ExpressionStatement); ok {
				g.emitLine(es)
				g.buf.WriteString(targetVar + " = ")
				g.genExpression(es.Expression)
				g.emit(";")
			} else {
				g.genStatement(stmt)
			}
		} else {
			g.genStatement(stmt)
		}
		g.emitDropsAt(b, i+1)
	}

	// Restore block context
	g.CurrentBlock = oldBlock
	g.CurrentStmtIndex = oldIdx

	g.emit("}")
}

func (g *Generator) genVarStatement(s *ast.VarStatement) {
	var finalType types.NRType
	sym := g.SemanticInfo.Defs[s.Name]
	if sym == nil {
		sym = g.SemanticInfo.Uses[s.Name]
	}

	if sym != nil {
		finalType = sym.Type
	} else if s.Type != nil {
		// Fallback to resolving type node if sym not found (shouldn't happen)
		finalType = g.SemanticInfo.Types[s.Name]
	}

	if finalType == nil {
		finalType = g.SemanticInfo.Types[s.Value]
	}

	if finalType == nil {
		finalType = types.I32 // Fallback
	}

	if s.Value != nil {
		g.buf.WriteString(fmt.Sprintf("%s %s = ", g.cType(finalType), s.Name.Value))

		oldNoTemp := g.NoTempWrap
		g.NoTempWrap = true
		oldTargetIsValue := g.TargetIsValue
		g.TargetIsValue = !g.isPointerTypeInC(finalType)
		if proto, ok := finalType.(*types.ProtocolType); ok {
			g.genInterfaceCast(s.Value, proto)
		} else {
			g.genOwnedValue(s.Value, finalType)
		}
		g.TargetIsValue = oldTargetIsValue
		g.NoTempWrap = oldNoTemp
		g.emit(";")
		if sym != nil && g.hasDropFlag(sym) {
			g.emit("bool %s = true;", g.dropFlagName(sym))
		}
	} else {
		g.buf.WriteString(fmt.Sprintf("%s %s;", g.cType(finalType), s.Name.Value))
		g.emit("")
		if sym != nil && g.hasDropFlag(sym) {
			g.emit("bool %s = false;", g.dropFlagName(sym))
		}
	}
}

func (g *Generator) genAssignment(s *ast.AssignmentStatement) {
	oldTargetIsValue := g.TargetIsValue
	g.TargetIsValue = !g.isPointerInC(s.Left)
	defer func() { g.TargetIsValue = oldTargetIsValue }()

	if idx, ok := s.Left.(*ast.IndexExpression); ok {
		t := g.SemanticInfo.Types[idx.Left]
		ut := t
		if pt, ok := t.(*types.PointerType); ok && !pt.IsArray {
			ut = pt.Base
		}

		// Struct operator overload: index_mut
		if st, ok := ut.(*types.StructType); ok && len(idx.Indices) == 1 {
			if methodType, exists := st.Methods["index_mut"]; exists {
				if mt, ok := methodType.(*types.FunctionType); ok && len(mt.Params) == 1 {
					g.buf.WriteString("*(")
					g.buf.WriteString(g.mangledTypeName(st) + "_index_mut(NULL, ")
					g.emitArgument(idx.Left, st, mt.ReceiverLease)
					g.buf.WriteString(", ")
					g.emitArgument(idx.Indices[0], mt.Params[0], mt.ParamLeases[0])
					g.buf.WriteString(")) = ")
					g.genExpression(s.Value)
					return
				}
			}
		}

		if mt, ok := ut.(*types.MapType); ok {
			g.buf.WriteString("map_set(")
			g.genExpression(idx.Left)
			g.buf.WriteString(", ")
			g.buf.WriteString(fmt.Sprintf("&(%s){", g.cType(mt.Key)))
			g.genExpression(idx.Indices[0])
			g.buf.WriteString("}, ")
			g.buf.WriteString(fmt.Sprintf("&(%s){", g.cType(mt.Value)))
			g.genExpression(s.Value)
			g.buf.WriteString("})")
			return
		}
	}

	shouldDeref := g.shouldDereferenceInC(s.Left)
	if shouldDeref {
		// Don't dereference owned pointers (@struct) for assignment
		t := g.SemanticInfo.Types[s.Left]
		if ident, ok := s.Left.(*ast.Identifier); ok {
			if sym := g.SemanticInfo.Uses[ident]; sym != nil {
				t = sym.Type
			} else if sym := g.SemanticInfo.Defs[ident]; sym != nil {
				t = sym.Type
			}
		}
		if pt, ok := t.(*types.PointerType); ok && pt.Kind == types.LeaseMove {
			ut := types.UnwrapLease(t)
			if _, ok := ut.(*types.StructType); ok || ut.GetKind() == types.KindSum {
				shouldDeref = false
			}
		}

		// [NEW] If LHS and RHS have the exact same lease type, this is a pointer REBIND, not a MUTATION.
		if _, isIdent := s.Left.(*ast.Identifier); isIdent {
			rhsType := g.SemanticInfo.Types[s.Value]
			if pt, ok := t.(*types.PointerType); ok && pt.Leased {
				if types.Equals(t, rhsType) {
					shouldDeref = false
				}
			}
		}
	}

	if shouldDeref {
		g.buf.WriteString("*")
	}

	g.genExpression(s.Left)
	g.buf.WriteString(" = ")

	var targetType types.NRType
	leftIdent, isIdent := s.Left.(*ast.Identifier)

	if isIdent {
		sym := g.SemanticInfo.Uses[leftIdent]
		if sym == nil {
			sym = g.SemanticInfo.Defs[leftIdent]
		}
		if sym == nil {
			sym = g.findSymbolByName(leftIdent.Value)
		}
		if sym != nil {
			targetType = sym.Type
		}
	} else {
		targetType = g.SemanticInfo.Types[s.Left]
	}

	if g.Solver != nil && g.Solver.AssignDrops != nil && g.Solver.AssignDrops[s] {
		cType := g.cType(targetType)
		g.buf.WriteString(fmt.Sprintf("({ %s _tmp = ", cType))

		oldNoTemp := g.NoTempWrap
		g.NoTempWrap = true
		if proto, ok := targetType.(*types.ProtocolType); ok {
			g.genInterfaceCast(s.Value, proto)
		} else {
			g.genOwnedValue(s.Value, targetType)
		}
		g.NoTempWrap = oldNoTemp

		g.buf.WriteString("; ")

		leftStr := g.exprToString(s.Left)
		g.emitDrop(leftStr, targetType, g.isPointerInC(s.Left))

		g.buf.WriteString(" _tmp; })")
	} else {
		oldNoTemp := g.NoTempWrap
		g.NoTempWrap = true
		if proto, ok := targetType.(*types.ProtocolType); ok {
			g.genInterfaceCast(s.Value, proto)
		} else {
			g.genOwnedValue(s.Value, targetType)
		}
		g.NoTempWrap = oldNoTemp
	}
}

func (g *Generator) isSameSelector(e1, e2 ast.Node) bool {
	if id1, ok1 := e1.(*ast.Identifier); ok1 {
		if id2, ok2 := e2.(*ast.Identifier); ok2 {
			return id1.Value == id2.Value
		}
	}
	s1, ok1 := e1.(*ast.SelectorExpression)
	s2, ok2 := e2.(*ast.SelectorExpression)
	if !ok1 || !ok2 {
		return false
	}
	if s1.Field.Value != s2.Field.Value {
		return false
	}
	return g.isSameSelector(s1.Left, s2.Left)
}

func (g *Generator) genAssignmentStatement(s *ast.AssignmentStatement) {
	g.genAssignment(s)
	g.emit(";")
	if ident, ok := s.Left.(*ast.Identifier); ok {
		sym := g.SemanticInfo.Uses[ident]
		if sym == nil {
			sym = g.SemanticInfo.Defs[ident]
		}
		if sym == nil {
			sym = g.findSymbolByName(ident.Value)
		}
		if sym != nil && g.hasDropFlag(sym) {
			g.emit("%s = true;", g.dropFlagName(sym))
		}
	}
}

func (g *Generator) genReturnStatement(s *ast.ReturnStatement) {
	var hasDrops bool
	var drops []topology.DropInfo
	if g.Solver != nil && g.CurrentBlock != nil && g.Solver.PreDrops[g.CurrentBlock] != nil {
		drops = g.Solver.PreDrops[g.CurrentBlock][g.CurrentStmtIndex]
		if len(drops) > 0 {
			hasDrops = true
		}
	}

	oldTargetIsValue := g.TargetIsValue
	if g.CurrentFunc != nil {
		if ft, ok := g.CurrentFunc.Type.(*types.FunctionType); ok && ft.Return != nil {
			g.TargetIsValue = !g.isPointerTypeInC(ft.Return)
		}
	}
	defer func() { g.TargetIsValue = oldTargetIsValue }()

	if hasDrops {
		if s.ReturnValue != nil {
			// Save return expression to _ret temp
			retCType := "void*"
			if g.CurrentFunc != nil {
				if ft, ok := g.CurrentFunc.Type.(*types.FunctionType); ok && ft.Return != nil {
					retCType = g.cType(ft.Return)
				}
			}
			g.emit("{")
			g.buf.WriteString(fmt.Sprintf("    %s _ret = ", retCType))
			if g.CurrentFunc != nil {
				if ft, ok := g.CurrentFunc.Type.(*types.FunctionType); ok && ft.Return != nil {
					if !strings.HasSuffix(retCType, "*") {
						if g.isGeneratedExpressionPointer(s.ReturnValue, nil) {
							g.buf.WriteString("*")
						}
					} else {
						if !g.isGeneratedExpressionPointer(s.ReturnValue, nil) {
							g.buf.WriteString("&")
						}
					}
				}
			}
			oldNoTemp := g.NoTempWrap
			g.NoTempWrap = true
			g.genExpression(s.ReturnValue)
			g.NoTempWrap = oldNoTemp
			g.emit(";")

			// Emit drops
			g.emitReturnDrops(drops)

			g.emit("    return _ret;")
			g.emit("}")
		} else {
			// No return value, just emit drops and return
			g.emit("{")
			g.emitReturnDrops(drops)
			g.emit("    return;")
			g.emit("}")
		}
	} else {
		// Normal return
		g.buf.WriteString("return")
		if s.ReturnValue != nil {
			g.buf.WriteString(" ")
			if g.CurrentFunc != nil {
				if ft, ok := g.CurrentFunc.Type.(*types.FunctionType); ok && ft.Return != nil {
					retCType := g.cType(ft.Return)
					if !strings.HasSuffix(retCType, "*") {
						if g.isGeneratedExpressionPointer(s.ReturnValue, nil) {
							g.buf.WriteString("*")
						}
					} else {
						if !g.isGeneratedExpressionPointer(s.ReturnValue, nil) {
							g.buf.WriteString("&")
						}
					}
				}
			}
			oldNoTemp := g.NoTempWrap
			g.NoTempWrap = true
			g.genExpression(s.ReturnValue)
			g.NoTempWrap = oldNoTemp
		}
		g.emit(";")
	}
}

func (g *Generator) emitReturnDrops(drops []topology.DropInfo) {
	for _, drop := range drops {
		sym := drop.Symbol
		if sym == nil {
			if drop.Field != nil {
				exprStr := g.genSelectorString(drop.Field.Left, drop.Field.Field.Value)
				t := g.SemanticInfo.Types[drop.Field]
				g.emitDrop(exprStr, t, g.isPointerTypeInC(t))
			} else if drop.Index != nil {
				t := g.SemanticInfo.Types[drop.Index]
				left := g.exprToString(drop.Index.Left)
				idx := g.exprToString(drop.Index.Indices[0])
				elemExpr := fmt.Sprintf("((%s*)array_data(%s))[%s]", g.cType(t), left, idx)
				g.emitDrop(elemExpr, t, g.isPointerTypeInC(t))
			}
			continue
		}

		name := g.variableName(sym)
		t := sym.Type
		isPtr := g.isSymbolPointerInC(sym)
		if g.hasDropFlag(sym) {
			dfName := g.dropFlagName(sym)
			g.emit("if (%s) {", dfName)
			g.emit("    %s = false;", dfName)
			g.emitDrop(name, t, isPtr)
			g.emit("}")
		} else {
			g.emitDrop(name, t, isPtr)
		}
	}
}

func (g *Generator) genIfExpression(s *ast.IfExpression) {
	g.genIfWithTarget(s, "")
}

func (g *Generator) genIfWithTarget(s *ast.IfExpression, targetVar string) {
	g.buf.WriteString("if (")
	g.genExpression(s.Condition)
	g.buf.WriteString(") ")
	g.genBlockWithTarget(s.Consequence, targetVar)

	if s.Alternative != nil {
		g.buf.WriteString(" else ")
		if block, ok := s.Alternative.(*ast.BlockStatement); ok {
			g.genBlockWithTarget(block, targetVar)
		} else if ifExpr, ok := s.Alternative.(*ast.IfExpression); ok {
			g.genIfWithTarget(ifExpr, targetVar)
		}
	}
}

func (g *Generator) genForStatement(s *ast.ForStatement) {
	// For-In Loop
	if s.Iterable != nil {
		t := g.SemanticInfo.Types[s.Iterable]
		g.emit("{")

		// Special case for Range struct
		if st, ok := t.(*types.StructType); ok && st.TypeName == "Range" {
			g.emit("    Range _range = ")
			g.genExpression(s.Iterable)
			g.emit(";")
			g.emit("    for (int _i = _range.start; _i < _range.end; _i++) {")
			g.emitYieldCheckpoint()

			if s.Key != nil {
				g.emit("        int %s = _i;", s.Key.Value)
			}

			varName := "_item"
			if s.Value != nil {
				varName = s.Value.Value
			}
			g.emit("        int %s = _i;", varName)

			g.genBlock(s.Body)
			g.emit("    }")
			g.emit("}")
			return
		}

		g.emit("    void* _arr = ")
		g.genExpression(s.Iterable)
		g.emit(";")
		g.emit("    int _len = array_count(_arr);")
		g.emit("    for (int _i = 0; _i < _len; _i++) {")
		g.emitYieldCheckpoint()

		var elemType types.NRType = types.I32
		if lt, ok := t.(*types.ListType); ok {
			elemType = lt.ElementType
		}

		if s.Key != nil {
			g.emit("        int %s = _i;", s.Key.Value)
		}

		varName := "_item"
		if s.Value != nil {
			varName = s.Value.Value
		}

		ct := g.cType(elemType)
		g.emit("        %s %s = ((%s*)_arr)[_i];", ct, varName, ct)
		g.genBlock(s.Body)
		g.emit("    }")
		g.emit("}")
		return
	}

	// Infinite Loop
	g.buf.WriteString("for (;;) ")
	g.emit("{")
	g.emitYieldCheckpoint()
	oldBlock := g.CurrentBlock
	oldIdx := g.CurrentStmtIndex
	g.CurrentBlock = s.Body

	for i, stmt := range s.Body.Statements {
		g.CurrentStmtIndex = i
		g.emitPreDropsAt(s.Body, i)
		g.genStatement(stmt)
		g.emitDropsAt(s.Body, i+1)
	}

	g.CurrentBlock = oldBlock
	g.CurrentStmtIndex = oldIdx
	g.emit("}")
}

func (g *Generator) genWhileStatement(s *ast.WhileStatement) {
	g.buf.WriteString("while (")
	g.genExpression(s.Condition)
	g.buf.WriteString(") ")

	g.emit("{")
	g.emitYieldCheckpoint()
	oldBlock := g.CurrentBlock
	oldIdx := g.CurrentStmtIndex
	g.CurrentBlock = s.Body

	for i, stmt := range s.Body.Statements {
		g.CurrentStmtIndex = i
		g.emitPreDropsAt(s.Body, i)
		g.genStatement(stmt)
		g.emitDropsAt(s.Body, i+1)
	}

	g.CurrentBlock = oldBlock
	g.CurrentStmtIndex = oldIdx
	g.emit("}")
}

func (g *Generator) genMatchExpression(s *ast.MatchExpression) {
	g.genMatchWithTarget(s, "")
}

func (g *Generator) genMatchWithTarget(s *ast.MatchExpression, targetVar string) {
	t := g.SemanticInfo.Types[s.Target]
	if st, ok := types.UnwrapLease(t).(*types.SumType); ok {
		mangledType := g.mangledTypeName(st)
		isLeased := false
		if pt, ok := t.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
			isLeased = true
		}

		isMove := false
		var moveSource ast.Expression
		if pref, ok := s.Target.(*ast.PrefixExpression); ok && pref.Operator == "@" {
			isMove = true
			moveSource = pref.Right
		} else if !isLeased {
			isMove = true
			moveSource = s.Target
		}

		g.emit("{")
		if isMove {
			if g.isPointerInC(moveSource) {
				g.buf.WriteString(fmt.Sprintf("    %s _target = *(", mangledType))
				g.genExpression(moveSource)
				g.buf.WriteString(");")
				g.emit("")
			} else {
				g.buf.WriteString(fmt.Sprintf("    %s _target = ", mangledType))
				g.genExpression(moveSource)
				g.emit(";")
			}

			if id, ok := moveSource.(*ast.Identifier); ok {
				sym := g.SemanticInfo.Uses[id]
				isLocalVarOrParam := false
				if sym != nil && (sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam) {
					isLocalVarOrParam = true
				}
				if isLocalVarOrParam {
					tId := g.SemanticInfo.Types[id]
					if g.isPointerTypeInC(tId) {
						if g.isPointerInC(id) {
							g.emit("    *%s = NULL;", id.Value)
						} else {
							g.emit("    %s = NULL;", id.Value)
						}
					} else {
						if g.isPointerInC(id) {
							g.emit("    memset(%s, 0, sizeof(%s));", id.Value, g.cType(tId))
						} else {
							g.emit("    memset(&%s, 0, sizeof(%s));", id.Value, g.cType(tId))
						}
					}
				}
			} else if sel, ok := moveSource.(*ast.SelectorExpression); ok {
				tSel := g.SemanticInfo.Types[sel]
				exprStr := g.genSelectorString(sel.Left, sel.Field.Value)
				if g.isPointerTypeInC(tSel) {
					if g.isPointerInC(sel) {
						g.emit("    *%s = NULL;", exprStr)
					} else {
						g.emit("    %s = NULL;", exprStr)
					}
				} else {
					if g.isPointerInC(sel) {
						g.emit("    memset(%s, 0, sizeof(%s));", exprStr, g.cType(tSel))
					} else {
						g.emit("    memset(&%s, 0, sizeof(%s));", exprStr, g.cType(tSel))
					}
				}
			} else if idx, ok := moveSource.(*ast.IndexExpression); ok {
				tIdx := g.SemanticInfo.Types[idx]
				if g.isPointerTypeInC(tIdx) {
					g.emit("    (*((%s*)array_bounds_check(%s, %s, \"%s\", %d))) = NULL; ",
						g.cType(tIdx), g.exprToString(idx.Left), g.exprToString(idx.Indices[0]),
						strings.ReplaceAll(idx.Pos().Filename, "\\", "/"), idx.Pos().Line)
				} else {
					g.emit("    memset(array_bounds_check(%s, %s, \"%s\", %d), 0, sizeof(%s)); ",
						g.exprToString(idx.Left), g.exprToString(idx.Indices[0]),
						strings.ReplaceAll(idx.Pos().Filename, "\\", "/"), idx.Pos().Line, g.cType(tIdx))
				}
			}
		} else if g.isPointerInC(s.Target) {
			g.buf.WriteString(fmt.Sprintf("    %s _target = *(", mangledType))
			g.genExpression(s.Target)
			g.buf.WriteString(");")
			g.emit("")
		} else {
			g.buf.WriteString(fmt.Sprintf("    %s _target = ", mangledType))
			g.genExpression(s.Target)
			g.emit(";")
		}

		for i, case_ := range s.Cases {
			vName := ""
			isWildcard := false
			if p, ok := case_.Pattern.(*ast.Identifier); ok {
				if p.Value == "_" {
					isWildcard = true
				} else {
					vName = p.Value
				}
			} else if p, ok := case_.Pattern.(*ast.CallExpression); ok {
				if ident, ok := p.Function.(*ast.Identifier); ok {
					vName = ident.Value
				}
			}

			prefix := "if"
			if i > 0 {
				prefix = "else if"
			}

			if isWildcard {
				if i == 0 {
					g.emit("    {")
				} else {
					g.emit("    else {")
				}
			} else if i == len(s.Cases)-1 && i > 0 {
				g.emit("    else {")
			} else {
				g.emit("    %s (_target.tag == %s_TAG_%s) {", prefix, mangledType, vName)
			}

			// Extract fields using compilePattern logic
			if p, ok := case_.Pattern.(*ast.CallExpression); ok {
				variant := st.Variants[vName]
				for j, arg := range p.Arguments {
					if j >= len(variant.FieldNames) {
						break
					}
					fName := variant.FieldNames[j]
					fType := variant.Fields[fName]

					targetPath := fmt.Sprintf("_target.data.%s.%s", vName, fName)
					if len(variant.FieldNames) == 1 {
						// Nora optimization: data.VariantName instead of data.VariantName.FieldName
						targetPath = fmt.Sprintf("_target.data.%s", vName)
					}

					_, captures := g.compilePattern(targetPath, arg.Value, fType, isMove)
					for _, cap := range captures {
						g.emit("        %s;", cap)
					}
				}
			}

			g.genBlockWithTarget(case_.Body, targetVar)
			g.emit("    }")
		}

		if isMove && g.needsDrop(st) {
			// SumTypes are values in Nora (not heap pointers themselves, though they contain pointers)
			// So we just need to drop it.
			dropMethod := g.requestAutoDrop(st)
			g.emit(fmt.Sprintf("    %s(&_target);", dropMethod))
		}
		g.emit("}")
		return
	} else if pt, ok := types.UnwrapLease(t).(*types.ProtocolType); ok {
		isLeased := false
		if ptype, ok := t.(*types.PointerType); ok && ptype.Leased && ptype.Kind != types.LeaseMove {
			isLeased = true
		}

		isMove := false
		var moveSource ast.Expression
		if pref, ok := s.Target.(*ast.PrefixExpression); ok && pref.Operator == "@" {
			isMove = true
			moveSource = pref.Right
		} else if !isLeased {
			isMove = true
			moveSource = s.Target
		}

		g.emit("{")
		if isMove {
			if g.isPointerInC(moveSource) {
				g.buf.WriteString(fmt.Sprintf("    %s _target = *(", g.cType(t)))
				g.genExpression(moveSource)
				g.buf.WriteString(");")
				g.emit("")
			} else {
				g.buf.WriteString(fmt.Sprintf("    %s _target = ", g.cType(t)))
				g.genExpression(moveSource)
				g.emit(";")
			}
		} else if g.isPointerInC(s.Target) {
			g.buf.WriteString(fmt.Sprintf("    %s _target = *(", g.cType(t)))
			g.genExpression(s.Target)
			g.buf.WriteString(");")
			g.emit("")
		} else {
			g.buf.WriteString(fmt.Sprintf("    %s _target = ", g.cType(t)))
			g.genExpression(s.Target)
			g.emit(";")
		}

		for i, case_ := range s.Cases {
			vName := ""
			isWildcard := false
			var expectedType types.NRType

			if p, ok := case_.Pattern.(*ast.Identifier); ok {
				if p.Value == "_" {
					isWildcard = true
				}
			} else if p, ok := case_.Pattern.(*ast.CallExpression); ok {
				if ident, ok := p.Function.(*ast.Identifier); ok {
					castSym := g.SemanticInfo.Uses[ident]
					if castSym == nil {
						castSym = g.findSymbolByName(ident.Value)
					}
					if castSym != nil && castSym.Kind == semantic.SymType {
						expectedType = castSym.Type
					} else {
						for _, st := range g.Structs {
							if st != nil && (st.TypeName == ident.Value || (st.BaseType != nil && st.BaseType.TypeName == ident.Value)) {
								expectedType = st
								break
							}
						}
					}
					if len(p.Arguments) == 1 {
						if argId, ok := p.Arguments[0].Value.(*ast.Identifier); ok {
							vName = argId.Value
						}
					}
				}
			}

			prefix := "if"
			if i > 0 {
				prefix = "else if"
			}

			if isWildcard {
				if i == 0 {
					g.emit("    {")
				} else {
					g.emit("    else {")
				}
			} else {
				if expectedType != nil {
					vtableName := g.requestVTable(expectedType, pt)
					g.emit("    %s (_target.vtable == %s) {", prefix, vtableName)
				} else {
					g.emit("    %s (false) { // error: expected type not found", prefix)
				}
			}

			if vName != "" {
				scope := g.SemanticInfo.Scopes[case_.Body]
				if scope != nil {
					sym, _ := scope.Lookup(vName)
					if sym != nil {
						tSym := sym.Type
						dfStr := ""
						if g.hasDropFlag(sym) {
							dfStr = fmt.Sprintf(" bool %s = true;", g.dropFlagName(sym))
						}

						if g.isPointerTypeInC(tSym) {
							g.emit("        %s %s = (%s)_target.data;%s", g.cType(tSym), vName, g.cType(tSym), dfStr)
						} else {
							g.emit("        %s %s = *(%s*)_target.data;%s", g.cType(tSym), vName, g.cType(tSym), dfStr)
						}
					}
				}
			}

			g.genBlockWithTarget(case_.Body, targetVar)
			g.emit("    }")
		}

		if isMove && !isLeased {
			g.emit("    if (_target.vtable && ((void**)_target.vtable)[0] != NULL) {")
			g.emit("        ((void(*)(void*))((void**)_target.vtable)[0])(_target.data);")
			g.emit("    }")
			g.emit("    if (_target.data) nr_free(_target.data);")
		}

		g.emit("}")
		return
	}

	isMove := false
	if pref, ok := s.Target.(*ast.PrefixExpression); ok && pref.Operator == "@" {
		isMove = true
	}

	g.emit("{")
	targetType := t
	if isMove {
		// For moved targets, we keep them as pointers to avoid copying the whole struct
		g.buf.WriteString(fmt.Sprintf("    %s _target = ", g.cType(t)))
		g.genExpression(s.Target)
	} else if g.isPointerInC(s.Target) {
		// For borrowed pointers, we dereference into a local copy for matching
		ut := types.UnwrapLease(t)
		if pt, ok := ut.(*types.PointerType); ok && !pt.IsArray {
			ut = pt.Base
		}
		g.buf.WriteString(fmt.Sprintf("    %s _target = *(", g.cType(ut)))
		g.genExpression(s.Target)
		g.buf.WriteString(")")
		targetType = ut
	} else {
		g.buf.WriteString(fmt.Sprintf("    %s _target = ", g.cType(t)))
		g.genExpression(s.Target)
	}
	g.emit(";")
	for i, case_ := range s.Cases {
		prefix := "if"
		if i > 0 {
			prefix = "else if"
		}

		cond, captures := g.compilePattern("_target", case_.Pattern, targetType, isMove)

		if cond == "true" {
			g.emit("    %s (true) {", prefix)
		} else {
			g.emit("    %s (%s) {", prefix, cond)
		}

		for _, cap := range captures {
			g.emit("        %s", cap)
		}

		g.genBlockWithTarget(case_.Body, targetVar)
		g.emit("    }")
	}

	if isMove {
		ut := types.UnwrapLease(t)
		if pt, ok := ut.(*types.PointerType); ok && !pt.IsArray {
			ut = pt.Base
		}
		if g.needsDrop(ut) {
			dropMethod := g.requestAutoDrop(ut)
			g.emit(fmt.Sprintf("    %s(_target);", dropMethod))
		}
		if g.isHeapAllocated(t) {
			g.emit("    nr_free(_target);")
		}
	}

	g.emit("}")
}

func (g *Generator) compilePattern(target string, pattern ast.Expression, t types.NRType, isMove bool) (string, []string) {
	origT := t
	t = types.UnwrapLease(t)
	if pt, ok := t.(*types.PointerType); ok && !pt.IsArray {
		t = pt.Base
	}

	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Value == "_" {
			return "true", nil
		}
		// Find the symbol of the captured identifier
		var sym *semantic.Symbol
		if g.SemanticInfo != nil {
			sym = g.SemanticInfo.Uses[p]
			if sym == nil {
				sym = g.SemanticInfo.Defs[p]
			}
		}
		if !isMove && sym != nil && types.IsOwnedType(origT) {
			// Mutate the symbol's type and AST node type to be a leased type so they are represented as pointers and not dropped
			if _, ok := sym.Type.(*types.PointerType); !ok || !sym.Type.IsLeased() {
				leasedType := &types.PointerType{Base: sym.Type, Leased: true, Kind: types.LeaseRead}
				sym.Type = leasedType
				sym.LeaseKind = types.LeaseRead
				if g.SemanticInfo != nil {
					g.SemanticInfo.Types[p] = leasedType
				}
			}
		}

		// Variable capture
		dfStr := ""
		if sym != nil && g.hasDropFlag(sym) {
			dfStr = fmt.Sprintf(" bool %s = true;", g.dropFlagName(sym))
		}

		if isMove && types.IsOwnedType(origT) {
			if g.isPointerTypeInC(origT) {
				return "true", []string{fmt.Sprintf("%s %s = %s; %s = NULL;%s", g.cType(origT), p.Value, target, target, dfStr)}
			} else {
				return "true", []string{fmt.Sprintf("%s %s = %s; memset(&%s, 0, sizeof(%s));%s", g.cType(origT), p.Value, target, target, g.cType(origT), dfStr)}
			}
		}

		if !isMove && types.IsOwnedType(origT) && !g.isPointerTypeInC(origT) {
			return "true", []string{fmt.Sprintf("%s* %s = &(%s);", g.cType(origT), p.Value, target)}
		}
		return "true", []string{fmt.Sprintf("%s %s = %s;%s", g.cType(origT), p.Value, target, dfStr)}

	case *ast.StructLiteral:
		st, ok := t.(*types.StructType)
		if !ok {
			return "false", nil
		}

		var conds []string
		var captures []string

		for _, field := range p.Fields {
			fieldName := field.Name.Value
			fieldType := st.Fields[fieldName]

			sep := "."
			if strings.HasPrefix(target, "*") || (isMove && !strings.Contains(target, ".data.")) {
				sep = "->"
			}
			fCond, fCaps := g.compilePattern(target+sep+fieldName, field.Value, fieldType, isMove)
			if fCond != "true" {
				conds = append(conds, fCond)
			}
			captures = append(captures, fCaps...)
		}

		if len(conds) == 0 {
			return "true", captures
		}
		return strings.Join(conds, " && "), captures

	case *ast.IntegerLiteral, *ast.Boolean, *ast.StringLiteral:
		return fmt.Sprintf("%s == %s", target, g.exprToString(pattern)), nil

	default:
		return "false", nil
	}
}

func (g *Generator) genExpressionStatement(s *ast.ExpressionStatement) {
	if as, ok := s.Expression.(*ast.AssignmentStatement); ok {
		g.genAssignmentStatement(as)
	} else {
		g.genExpression(s.Expression)
		g.emit(";")
	}
}

func (g *Generator) genDeferStatement(s *ast.DeferStatement) {
	g.emit("/* defer not implemented in direct codegen */")
}

// --- RAII DROP EMISSION ---

func (g *Generator) emitDropsAt(block *ast.BlockStatement, index int) {
	if g.Solver == nil || g.Solver.Drops[block] == nil {
		return
	}

	drops := g.Solver.Drops[block][index]
	for _, drop := range drops {
		sym := drop.Symbol
		if sym == nil {
			continue
		}
		if g.DebugSemantic {
			fmt.Printf("[DEBUG-emitDropsAt] block=%p, index=%d, sym.Name=%s, sym.Type=%T (%v), sym.Kind=%v\n", block, index, sym.Name, sym.Type, sym.Type, sym.Kind)
		}

		t := sym.Type
		name := g.variableName(sym)

		if g.EnableDebug {
			g.emit("// End of life for stack var: %s", name)
		}

		isPtr := g.isSymbolPointerInC(sym)
		if g.hasDropFlag(sym) {
			dfName := g.dropFlagName(sym)
			g.emit("if (%s) {", dfName)
			g.emit("    %s = false;", dfName)
			g.emitDrop(name, t, isPtr)
			g.emit("}")
		} else {
			g.emitDrop(name, t, isPtr)
		}
	}
}

func (g *Generator) emitPreDropsAt(block *ast.BlockStatement, index int) {
	if g.Solver == nil || g.Solver.PreDrops[block] == nil {
		return
	}

	// Skip return statements. We evaluate the expression and do drops in genReturnStatement
	if index >= 0 && index < len(block.Statements) {
		if _, ok := block.Statements[index].(*ast.ReturnStatement); ok {
			return
		}
	}

	drops := g.Solver.PreDrops[block][index]
	for _, drop := range drops {
		sym := drop.Symbol
		if sym == nil {
			if drop.Field != nil {
				exprStr := g.genSelectorString(drop.Field.Left, drop.Field.Field.Value)
				t := g.SemanticInfo.Types[drop.Field]
				g.emitDrop(exprStr, t, g.isPointerTypeInC(t))
			} else if drop.Index != nil {
				// Handle array index re-assignment drop
				t := g.SemanticInfo.Types[drop.Index]
				left := g.exprToString(drop.Index.Left)
				idx := g.exprToString(drop.Index.Indices[0])

				// Calculate pointer to element
				elemExpr := fmt.Sprintf("((%s*)array_data(%s))[%s]", g.cType(t), left, idx)
				g.emitDrop(elemExpr, t, g.isPointerTypeInC(t))
			}
			continue
		}

		name := g.variableName(sym)
		t := sym.Type
		isPtr := g.isSymbolPointerInC(sym)
		if g.hasDropFlag(sym) {
			dfName := g.dropFlagName(sym)
			g.emit("if (%s) {", dfName)
			g.emit("    %s = false;", dfName)
			g.emitDrop(name, t, isPtr)
			g.emit("}")
		} else {
			g.emitDrop(name, t, isPtr)
		}
	}
}
func (g *Generator) genSelectStatement(s *ast.SelectStatement) {
	numCases := 0
	hasDefault := false
	for _, c := range s.Cases {
		if c.Condition != nil {
			numCases++
		} else {
			hasDefault = true
		}
	}

	g.emit("{")
	if numCases > 0 {
		g.emit("    select_op_t _ops[%d];", numCases)
		opIdx := 0
		for _, c := range s.Cases {
			if c.Condition == nil {
				continue
			}

			var cond interface{} = c.Condition
			if es, ok := cond.(*ast.ExpressionStatement); ok {
				cond = es.Expression
			}

			if se, ok := cond.(*ast.SendExpression); ok {
				g.buf.WriteString(fmt.Sprintf("    _ops[%d].chan = ", opIdx))
				if g.shouldDereferenceInC(se.Left) {
					g.buf.WriteString("*")
				}
				g.genExpression(se.Left)
				g.emit(";")
				g.emit("    _ops[%d].is_send = true;", opIdx)
				t := g.SemanticInfo.Types[se.Right]
				if t == nil {
					t = types.I32
				}
				g.buf.WriteString(fmt.Sprintf("    %s _send_val_%d = ", g.cType(t), opIdx))
				g.genExpression(se.Right)
				g.emit(";")
				g.emit("    _ops[%d].data = &_send_val_%d;", opIdx, opIdx)
				opIdx++
			} else if re, ok := cond.(*ast.ReceiveExpression); ok {
				g.buf.WriteString(fmt.Sprintf("    _ops[%d].chan = ", opIdx))
				if g.shouldDereferenceInC(re.Value) {
					g.buf.WriteString("*")
				}
				g.genExpression(re.Value)
				g.emit(";")
				g.emit("    _ops[%d].is_send = false;", opIdx)
				t := g.SemanticInfo.Types[re]
				if t == nil {
					t = types.I32
				}
				g.emit("    %s _tmp_%d;", g.cType(t), opIdx)
				g.emit("    _ops[%d].data = &_tmp_%d;", opIdx, opIdx)
				opIdx++
			} else if assign, ok := cond.(*ast.AssignmentStatement); ok {
				if re, ok := assign.Value.(*ast.ReceiveExpression); ok {
					g.buf.WriteString(fmt.Sprintf("    _ops[%d].chan = ", opIdx))
					if g.shouldDereferenceInC(re.Value) {
						g.buf.WriteString("*")
					}
					g.genExpression(re.Value)
					g.emit(";")
					g.emit("    _ops[%d].is_send = false;", opIdx)
					g.buf.WriteString(fmt.Sprintf("    _ops[%d].data = &", opIdx))
					g.genExpression(assign.Left)
					g.emit(";")
					opIdx++
				}
			} else if vs, ok := cond.(*ast.VarStatement); ok {
				if re, ok := vs.Value.(*ast.ReceiveExpression); ok {
					g.buf.WriteString(fmt.Sprintf("    _ops[%d].chan = ", opIdx))
					if g.shouldDereferenceInC(re.Value) {
						g.buf.WriteString("*")
					}
					g.genExpression(re.Value)
					g.emit(";")
					g.emit("    _ops[%d].is_send = false;", opIdx)
					t := g.SemanticInfo.Types[vs.Value]
					if t == nil {
						t = types.I32
					}
					g.emit("    %s %s;", g.cType(t), vs.Name.Value)
					g.emit("    _ops[%d].data = &%s;", opIdx, vs.Name.Value)
					opIdx++
				}
			}
		}
		g.emit("    int _res = channel_select(_ops, %d, %v);", numCases, hasDefault)

		// Clean up any unselected send expressions to prevent memory leaks
		opIdx = 0
		for _, c := range s.Cases {
			if c.Condition == nil {
				continue
			}

			var cond interface{} = c.Condition
			if es, ok := cond.(*ast.ExpressionStatement); ok {
				cond = es.Expression
			}

			if se, ok := cond.(*ast.SendExpression); ok {
				t := g.SemanticInfo.Types[se.Right]
				if t == nil {
					t = types.I32
				}
				if types.IsOwnedType(t) {
					g.emit("    if (_res != %d) {", opIdx)
					expr := fmt.Sprintf("*( (%s*)_ops[%d].data )", g.cType(t), opIdx)
					g.emitDrop(expr, t, g.isPointerTypeInC(t))
					g.emit("    }")
				}
			}
			opIdx++
		}

		emitted := 0
		for _, c := range s.Cases {
			if c.Condition == nil {
				continue
			}
			prefix := "if"
			if emitted > 0 {
				prefix = "else if"
			}
			g.emitLine(c)
			g.emit("    %s (_res == %d) ", prefix, emitted)
			g.genBlock(c.Body)
			emitted++
		}
		if hasDefault {
			for _, c := range s.Cases {
				if c.Condition == nil {
					g.emitLine(c)
					g.emit("    else ")
					g.genBlock(c.Body)
					break
				}
			}
		}
	} else if hasDefault {
		for _, c := range s.Cases {
			if c.Condition == nil {
				g.emitLine(c)
				g.genBlock(c.Body)
				break
			}
		}
	}
	g.emit("}")
}
