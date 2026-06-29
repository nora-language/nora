package codegen

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nora-language/nora/pkg/hir"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/types"
)

func collectAllocas(b *hir.HIRBlock) []*hir.Alloca {
	var allocas []*hir.Alloca
	if b == nil {
		return allocas
	}
	for _, el := range b.Elements {
		switch e := el.(type) {
		case *hir.InstElement:
			if alloca, ok := e.Inst.(*hir.Alloca); ok {
				allocas = append(allocas, alloca)
			}
		case *hir.HIRIf:
			allocas = append(allocas, collectAllocas(e.Then)...)
			if e.Else != nil {
				allocas = append(allocas, collectAllocas(e.Else)...)
			}
		case *hir.IteratorLoop:
			allocas = append(allocas, collectAllocas(e.Body)...)
		case *hir.HIRLoop:
			allocas = append(allocas, collectAllocas(e.Init)...)
			allocas = append(allocas, collectAllocas(e.Step)...)
			allocas = append(allocas, collectAllocas(e.Body)...)
		case *hir.Select:
			for _, c := range e.Cases {
				allocas = append(allocas, collectAllocas(c.Body)...)
			}
		}
	}
	return allocas
}

func (g *Generator) genHIRFunction(hf *hir.Function) {
	g.CurrentFunc = hf.FuncSymbol
	if hf.LambdaExpr != nil {
		g.CurrentLambda = hf.LambdaExpr
	}
	defer func() {
		g.CurrentLambda = nil
	}()

	// Print signature
	name := g.mangleName(hf.FuncSymbol)
	ft := hf.FuncSymbol.Type.(*types.FunctionType)
	
	// Use type-erased signature if this is a shared generic monomorphization
	if declSym, ok := g.Functions[name]; ok && declSym.Type != nil {
		if declFT, ok := declSym.Type.(*types.FunctionType); ok {
			ft = declFT
		}
	}
	
	retType := g.cType(ft.Return)
	mangledName := name

	var params string
	fn, hasFn := hf.FuncSymbol.DefNode.(*ast.FunctionStatement)
	if !(hasFn && fn.IsExport) {
		params = "void* _env_ptr"
	}

	if hasFn && fn.Receiver != nil {
		t := g.cParamType(ft.Receiver, ft.ReceiverLease)
		if params != "" {
			params += ", "
		}
		params += t + " " + fn.Receiver.Name.Value
	}

	var lambdaParams []*ast.Parameter
	if hf.LambdaExpr != nil {
		lambdaParams = hf.LambdaExpr.Parameters
	}

	for i, p := range ft.Params {
		if params != "" {
			params += ", "
		}
		lease := types.LeaseRead
		if i < len(ft.ParamLeases) {
			lease = ft.ParamLeases[i]
		}
		t := g.cParamType(p, lease)
		params += t
		if hasFn && i < len(fn.Parameters) && fn.Parameters[i].Name != nil {
			params += " " + fn.Parameters[i].Name.Value
		} else if hf.LambdaExpr != nil && i < len(lambdaParams) && lambdaParams[i].Name != nil {
			params += " " + lambdaParams[i].Name.Value
		} else {
			if i < len(hf.Params) {
				params += " " + hf.Params[i]
			} else {
				params += fmt.Sprintf(" _param_%d", i)
			}
		}
	}

	if g.EnableDebug {
		if hasFn && fn != nil {
			g.emitLine(fn)
		} else if hf.LambdaExpr != nil {
			g.emitLine(hf.LambdaExpr)
		}
	}

	g.emit(fmt.Sprintf("%s %s(%s) {", retType, mangledName, params))
	g.emitYieldCheckpoint()

	if hasFn && fn.IsExport {
		g.emit("    void* _env_ptr = NULL;")
	}

	if hf.LambdaExpr != nil {
		scope := g.SemanticInfo.Scopes[hf.LambdaExpr]
		if scope != nil && scope.Captures != nil && len(scope.Captures) > 0 {
			envStructName := mangledName + "_env_t"
			g.emit(fmt.Sprintf("    %s* _env = (%s*)_env_ptr;", envStructName, envStructName))
		}
	}

	// Add NULL check for drop methods to prevent segfaults
	if hasFn && fn.Receiver != nil && strings.HasPrefix(fn.Name.Value, "drop") {
		g.emit("if (%s == NULL) return;", fn.Receiver.Name.Value)
	}

	// Drop flags initialization for parameters and receiver
	if hasFn && fn.Receiver != nil {
		sym := g.SemanticInfo.Defs[fn.Receiver.Name]
		if sym == nil {
			sym = g.SemanticInfo.Uses[fn.Receiver.Name]
		}
		if sym != nil && g.hasDropFlag(sym) {
			g.emit("    bool %s = true;", g.dropFlagName(sym))
		}
	}
	if hasFn {
		for _, param := range fn.Parameters {
			if param.Name != nil {
				sym := g.SemanticInfo.Defs[param.Name]
				if sym == nil {
					sym = g.SemanticInfo.Uses[param.Name]
				}
				if sym != nil && g.hasDropFlag(sym) {
					g.emit("    bool %s = true;", g.dropFlagName(sym))
				}
			}
		}
	}
	if hf.LambdaExpr != nil {
		for _, param := range hf.LambdaExpr.Parameters {
			if param.Name != nil {
				sym := g.SemanticInfo.Defs[param.Name]
				if sym == nil {
					sym = g.SemanticInfo.Uses[param.Name]
				}
				if sym != nil && g.hasDropFlag(sym) {
					g.emit("    bool %s = true;", g.dropFlagName(sym))
				}
			}
		}
	}

	// Drop flags initialization for local variables
	allAllocas := collectAllocas(hf.Body)
	declaredDropFlags := make(map[string]bool)
	for _, alloca := range allAllocas {
		sym := alloca.Symbol
		if sym != nil && g.hasDropFlag(sym) {
			dfName := g.dropFlagName(sym)
			if !declaredDropFlags[dfName] {
				g.emit(fmt.Sprintf("    bool %s = false;", dfName))
				declaredDropFlags[dfName] = true
			}
		}
	}

	// Generate block elements
	g.genHIRBlock(hf.Body)

	g.emit("    nr_flush_temps();")
	g.emit("}")

	g.emit("")
}

func (g *Generator) genHIRBlock(b *hir.HIRBlock) {
	for _, el := range b.Elements {
		switch e := el.(type) {
		case *hir.InstElement:
			g.genHIRInstruction(e.Inst)
		case *hir.HIRIf:
			g.emit(fmt.Sprintf("    if (%s) {", g.hirOperandStr(e.Condition)))
			g.genHIRBlock(e.Then)
			if e.Else != nil && len(e.Else.Elements) > 0 {
				g.emit("    } else {")
				g.genHIRBlock(e.Else)
			}
			g.emit("    }")
		case *hir.IteratorLoop:
			g.emit("    {") // loop scope
			iterType := e.Iterator.GetType()
			iterName := g.mangledTypeName(iterType)
			iterVar := g.hirOperandStr(e.Iterator)

			nextName := e.NextMangled
			if nextName == "" && e.NextSymbol != nil {
				nextName = g.mangleName(e.NextSymbol)
			}
			if nextName == "" {
				nextName = iterName + "_Next" // fallback
			}

			g.emit(fmt.Sprintf("    while (1) {"))
			g.emit(fmt.Sprintf("        __auto_type _opt = %s(_env_ptr, &%s);", nextName, iterVar))
			g.emit(fmt.Sprintf("        if (_opt.tag == 0) { break; }")) // None
			if e.ElemName != "" {
				cTypeStr := g.cType(e.ElemType)
				g.emit(fmt.Sprintf("        %s %s = (%s)_opt.data.Some;", cTypeStr, e.ElemName, cTypeStr))
			}
			g.genHIRBlock(e.Body)
			g.emit("        nr_flush_temps();")
			g.emit("    }")
			g.emit("    }")
		case *hir.HIRLoop:
			g.emit("    {") // loop scope
			if e.Init != nil {
				g.genHIRBlock(e.Init)
			}
			cond := "1"
			if e.Condition != nil {
				cond = g.hirOperandStr(e.Condition)
			}
			g.emit(fmt.Sprintf("    while (%s) {", cond))
			g.genHIRBlock(e.Body)
			if e.Step != nil {
				g.genHIRBlock(e.Step)
			}
			g.emit("        nr_flush_temps();")
			g.emit("    }")
			g.emit("    }")
		case *hir.Select:
			numCases := 0
			hasDefault := false
			for _, c := range e.Cases {
				if !c.IsDefault {
					numCases++
				} else {
					hasDefault = true
				}
			}

			g.emit("    {")
			if numCases > 0 {
				g.emit(fmt.Sprintf("    select_op_t _ops[%d];", numCases))
				opIdx := 0
				for _, c := range e.Cases {
					if c.IsDefault {
						continue
					}

					g.emit(fmt.Sprintf("    _ops[%d].chan = %s;", opIdx, g.hirOperandStr(c.Chan)))
					if c.IsSend {
						g.emit(fmt.Sprintf("    _ops[%d].is_send = true;", opIdx))
						g.emit(fmt.Sprintf("    %s _send_val_%d = %s;", g.cType(c.Val.GetType()), opIdx, g.hirOperandStr(c.Val)))
						g.emit(fmt.Sprintf("    _ops[%d].data = &_send_val_%d;", opIdx, opIdx))
					} else {
						g.emit(fmt.Sprintf("    _ops[%d].is_send = false;", opIdx))
						if c.VarName != "" {
							g.emit(fmt.Sprintf("    %s %s;", g.cType(c.VarType), c.VarName))
							g.emit(fmt.Sprintf("    _ops[%d].data = &%s;", opIdx, c.VarName))
						} else if c.Val != nil {
							g.emit(fmt.Sprintf("    _ops[%d].data = &(%s);", opIdx, g.hirOperandStr(c.Val)))
						} else {
							chanType := types.UnwrapLease(c.Chan.GetType())
							if pt, ok := chanType.(*types.PointerType); ok {
								chanType = pt.Base
							}
							var elemType types.NRType = types.I32
							if ct, ok := chanType.(*types.ChanType); ok {
								elemType = ct.Elem
							}
							g.emit(fmt.Sprintf("    %s _tmp_%d;", g.cType(elemType), opIdx))
							g.emit(fmt.Sprintf("    _ops[%d].data = &_tmp_%d;", opIdx, opIdx))
						}
					}
					opIdx++
				}

				g.emit(fmt.Sprintf("    int _res = channel_select(_ops, %d, %v);", numCases, hasDefault))

				opIdx = 0
				for _, c := range e.Cases {
					if c.IsDefault {
						continue
					}
					if c.IsSend {
						t := c.Val.GetType()
						if _, ok := types.UnwrapLease(t).(*types.ChanType); ok {
							g.emit(fmt.Sprintf("    if (_res != %d) { channel_free(_send_val_%d); }", opIdx, opIdx))
						}
					}
					opIdx++
				}

				emitted := 0
				for _, c := range e.Cases {
					if c.IsDefault {
						continue
					}
					prefix := "if"
					if emitted > 0 {
						prefix = "else if"
					}
					g.emit(fmt.Sprintf("    %s (_res == %d) {", prefix, emitted))
					g.genHIRBlock(c.Body)
					g.emit("    }")
					emitted++
				}
				if hasDefault {
					g.emit("    else {")
					for _, c := range e.Cases {
						if c.IsDefault {
							g.genHIRBlock(c.Body)
						}
					}
					g.emit("    }")
				}
			} else {
				for _, c := range e.Cases {
					if c.IsDefault {
						g.genHIRBlock(c.Body)
					}
				}
			}
			g.emit("    }")
		}
	}
}

func (g *Generator) genHIRInstruction(inst hir.Instruction) {
	switch i := inst.(type) {
	case *hir.LineInfo:
		if i.ASTNode != nil {
			g.emitLine(i.ASTNode)
		}
	case *hir.Alloca:
		name := i.Name
		if i.Symbol != nil {
			name = g.variableName(i.Symbol)
		}
		g.emit(fmt.Sprintf("    %s %s;", g.cType(i.Type), name))
	case *hir.Store:
		isMapOrIndexSet := false
		if instOp, ok := i.Dest.(*hir.InstOperand); ok {
			if astExpr, ok := instOp.Inst.(*hir.ASTExpr); ok {
				if idx, ok := astExpr.ASTNode.(*ast.IndexExpression); ok {
					t := g.SemanticInfo.Types[idx.Left]
					ut := t
					if pt, ok := t.(*types.PointerType); ok && !pt.IsArray {
						ut = pt.Base
					}

					// Struct operator overload: index_mut
					if st, ok := ut.(*types.StructType); ok && len(idx.Indices) == 1 {
						if methodType, exists := st.Methods["index_mut"]; exists {
							if mt, ok := methodType.(*types.FunctionType); ok && len(mt.Params) == 1 {
								oldNoTemp := g.NoTempWrap
								g.NoTempWrap = true
								valStr := g.hirOperandStr(i.Val)
								g.NoTempWrap = oldNoTemp

								g.emit(fmt.Sprintf("    *(%s_index_mut(NULL, ", g.mangledTypeName(st)))
								oldBuf := g.buf
								var argBuf1 bytes.Buffer
								g.buf = &argBuf1
								g.emitArgument(idx.Left, st, mt.ReceiverLease)
								g.buf = oldBuf

								var argBuf2 bytes.Buffer
								g.buf = &argBuf2
								g.emitArgument(idx.Indices[0], mt.Params[0], mt.ParamLeases[0])
								g.buf = oldBuf

								g.emit(fmt.Sprintf("%s, %s)) = %s;", argBuf1.String(), argBuf2.String(), valStr))
								isMapOrIndexSet = true
							}
						}
					}

					if !isMapOrIndexSet {
						if mt, ok := types.UnwrapLease(t).(*types.MapType); ok {
							mapStr := g.exprToString(idx.Left)
							keyStr := g.exprToString(idx.Indices[0])
							oldNoTemp := g.NoTempWrap
							g.NoTempWrap = true
							valStr := g.hirOperandStr(i.Val)
							g.NoTempWrap = oldNoTemp
							g.emit(fmt.Sprintf("    map_set(%s, &(%s){%s}, &(%s){%s});",
								mapStr, g.cType(mt.Key), keyStr, g.cType(mt.Value), valStr))
							isMapOrIndexSet = true
						}
					}
				}
			}
		}
		if !isMapOrIndexSet {
			oldNoTemp := g.NoTempWrap
			g.NoTempWrap = true
			valStr := g.hirOperandStr(i.Val)
			g.NoTempWrap = oldNoTemp
			valIsPtr := g.isOperandPointerInC(i.Val)
			destIsPtr := g.isOperandPointerInC(i.Dest)
			// fmt.Printf("[DEBUG-STORE] Dest=%s (Type=%s, CType=%s, isPtr=%v), Val=%s (Type=%s, CType=%s, isPtr=%v)\n",
			// 	i.Dest.String(), i.Dest.GetType().Name(), g.cTypeOfOperand(i.Dest), destIsPtr,
			// 	i.Val.String(), i.Val.GetType().Name(), g.cTypeOfOperand(i.Val), valIsPtr)
			if valIsPtr && !destIsPtr {
				if g.isHIRTemporaryHeapPointer(i.Val) {
					valStr = g.wrapHIRTemporaryHeapPointer(i.Val, valStr)
				} else {
					valStr = "*" + valStr
				}
			}
			destStr := g.hirOperandStr(i.Dest)
			if !valIsPtr && destIsPtr {
				baseType := i.Dest.GetType()
				if pt, ok := baseType.(*types.PointerType); ok {
					baseType = pt.Base
				}
				valStr = fmt.Sprintf("((%s[]){ %s })", g.cType(baseType), valStr)
			}
			varOp, isVar := i.Dest.(*hir.VarOperand)
			if i.NeedsOldDrop {
				oldBuf := g.buf
				var tempBuf bytes.Buffer
				g.buf = &tempBuf
				oldEnableDebug := g.EnableDebug
				g.EnableDebug = false
				targetType := i.Dest.GetType()
				g.emitDrop("_old", targetType, g.isOperandPointerInC(i.Dest))
				g.EnableDebug = oldEnableDebug
				dropStr := strings.TrimSpace(tempBuf.String())
				g.buf = oldBuf

				if isVar && varOp.Symbol != nil && g.hasDropFlag(varOp.Symbol) {
					dfName := g.dropFlagName(varOp.Symbol)
					g.emit(fmt.Sprintf("    { %s _old = %s; %s = %s; if (%s) { %s } }", g.cType(targetType), destStr, destStr, valStr, dfName, dropStr))
				} else {
					g.emit(fmt.Sprintf("    { %s _old = %s; %s = %s; %s }", g.cType(targetType), destStr, destStr, valStr, dropStr))
				}
			} else {
				g.emit(fmt.Sprintf("    %s = %s;", destStr, valStr))
			}
			if isVar && varOp.Symbol != nil {
				if g.hasDropFlag(varOp.Symbol) {
					g.emit(fmt.Sprintf("    %s = true;", g.dropFlagName(varOp.Symbol)))
				}
			}
			if sym := g.getMovedHeapPointerSymbol(i.Val); sym != nil {
				destType := i.Dest.GetType()
				if destType != nil {
					if !types.IsPointerLike(destType) {
						g.emit(fmt.Sprintf("    nr_free(%s); %s = NULL;", g.variableName(sym), g.variableName(sym)))
					}
				}
			}
		}
	case *hir.Assign:
		isMapOrIndexSet := false
		if instOp, ok := i.Dest.(*hir.InstOperand); ok {
			if astExpr, ok := instOp.Inst.(*hir.ASTExpr); ok {
				if idx, ok := astExpr.ASTNode.(*ast.IndexExpression); ok {
					t := g.SemanticInfo.Types[idx.Left]
					ut := t
					if pt, ok := t.(*types.PointerType); ok && !pt.IsArray {
						ut = pt.Base
					}

					// Struct operator overload: index_mut
					if st, ok := ut.(*types.StructType); ok && len(idx.Indices) == 1 {
						if methodType, exists := st.Methods["index_mut"]; exists {
							if mt, ok := methodType.(*types.FunctionType); ok && len(mt.Params) == 1 {
								oldNoTemp := g.NoTempWrap
								g.NoTempWrap = true
								valStr := g.hirOperandStr(i.Val)
								g.NoTempWrap = oldNoTemp

								g.emit(fmt.Sprintf("    *(%s_index_mut(NULL, ", g.mangledTypeName(st)))
								oldBuf := g.buf
								var argBuf1 bytes.Buffer
								g.buf = &argBuf1
								g.emitArgument(idx.Left, st, mt.ReceiverLease)
								g.buf = oldBuf

								var argBuf2 bytes.Buffer
								g.buf = &argBuf2
								g.emitArgument(idx.Indices[0], mt.Params[0], mt.ParamLeases[0])
								g.buf = oldBuf

								g.emit(fmt.Sprintf("%s, %s)) = %s;", argBuf1.String(), argBuf2.String(), valStr))
								isMapOrIndexSet = true
							}
						}
					}

					if !isMapOrIndexSet {
						if mt, ok := types.UnwrapLease(t).(*types.MapType); ok {
							mapStr := g.exprToString(idx.Left)
							keyStr := g.exprToString(idx.Indices[0])
							oldNoTemp := g.NoTempWrap
							g.NoTempWrap = true
							valStr := g.hirOperandStr(i.Val)
							g.NoTempWrap = oldNoTemp
							g.emit(fmt.Sprintf("    map_set(%s, &(%s){%s}, &(%s){%s});",
								mapStr, g.cType(mt.Key), keyStr, g.cType(mt.Value), valStr))
							isMapOrIndexSet = true
						}
					}
				}
			}
		}
		if !isMapOrIndexSet {
			oldNoTemp := g.NoTempWrap
			g.NoTempWrap = true
			valStr := g.hirOperandStr(i.Val)
			g.NoTempWrap = oldNoTemp
			valIsPtr := g.isOperandPointerInC(i.Val)
			destIsPtr := g.isOperandPointerInC(i.Dest)
			if valIsPtr && !destIsPtr {
				if g.isHIRTemporaryHeapPointer(i.Val) {
					valStr = g.wrapHIRTemporaryHeapPointer(i.Val, valStr)
				} else {
					valStr = "*" + valStr
				}
			}
			destStr := g.hirOperandStr(i.Dest)
			if !g.isOperandPointerInC(i.Val) && g.isOperandPointerInC(i.Dest) {
				baseType := i.Dest.GetType()
				if pt, ok := baseType.(*types.PointerType); ok {
					baseType = pt.Base
				}
				valStr = fmt.Sprintf("((%s[]){ %s })", g.cType(baseType), valStr)
			}
			varOp, isVar := i.Dest.(*hir.VarOperand)
			// Assign doesn't have NeedsOldDrop since it's mostly unused.
			g.emit(fmt.Sprintf("    %s = %s;", destStr, valStr))
			if isVar && varOp.Symbol != nil {
				if g.hasDropFlag(varOp.Symbol) {
					g.emit(fmt.Sprintf("    %s = true;", g.dropFlagName(varOp.Symbol)))
				}
			}
		}

	case *hir.Ret:
		oldNoTemp := g.NoTempWrap
		g.NoTempWrap = true
		valStr := g.hirOperandStr(i.Val)
		g.NoTempWrap = oldNoTemp

		if g.CurrentFunc != nil {
			if ft, ok := g.CurrentFunc.Type.(*types.FunctionType); ok {
				if ft.Return == nil || ft.Return.Name() == "void" {
					g.emit("    nr_flush_temps();")
					g.emit("    return;")
					break
				}
				retCType := g.cType(ft.Return)
				valCType := g.cTypeOfOperand(i.Val)
				if strings.HasSuffix(retCType, "*") && !strings.HasSuffix(valCType, "*") && valStr != "NULL" {
					if pt, isPtr := ft.Return.(*types.PointerType); isPtr && pt.Leased && (pt.Kind == types.LeaseRead || pt.Kind == types.LeaseWrite) {
						g.emit("    nr_flush_temps();")
						g.emit(fmt.Sprintf("    return &(%s);", valStr))
						break
					} else {
						// We must heap-allocate the return value because the caller owns it (move lease return)
						ct := strings.TrimSuffix(retCType, "*")
						g.emit(fmt.Sprintf("    %s* _ret_heap = nr_malloc(sizeof(%s));", ct, ct))
						g.emit(fmt.Sprintf("    *_ret_heap = %s;", valStr))
						g.emit("    nr_flush_temps();")
						g.emit("    return _ret_heap;")
						break
					}
				} else if !strings.HasSuffix(retCType, "*") && strings.HasSuffix(valCType, "*") {
					g.emit("    nr_flush_temps();")
					g.emit(fmt.Sprintf("    return *%s;", valStr))
					break
				}
				g.emit("    nr_flush_temps();")
				g.emit(fmt.Sprintf("    return %s;", valStr))
				break
			}
		}
		g.emit("    nr_flush_temps();")
		g.emit(fmt.Sprintf("    return %s;", valStr))

	case *hir.Call:
		g.emit(fmt.Sprintf("    %s;", g.hirInstructionStr(i)))
	case *hir.Drop:
		sym := i.Symbol
		if sym == nil {
			if i.Field != nil {
				exprStr := g.exprToString(i.Field)
				t := g.SemanticInfo.Types[i.Field]
				g.emitDrop(exprStr, t, g.isPointerTypeInC(t))
			} else if i.Index != nil {
				t := g.SemanticInfo.Types[i.Index]
				if idx, ok := i.Index.(*ast.IndexExpression); ok {
					left := g.exprToString(idx.Left)
					index := g.exprToString(idx.Indices[0])
					elemExpr := fmt.Sprintf("((%s*)array_data(%s))[%s]", g.cType(t), left, index)
					g.emitDrop(elemExpr, t, g.isPointerTypeInC(t))
				}
			}
			break
		}

		name := g.variableName(sym)
		t := sym.Type
		isPtr := g.isSymbolPointerInC(sym)
		if g.hasDropFlag(sym) {
			dfName := g.dropFlagName(sym)
			g.emit(fmt.Sprintf("    if (%s) {", dfName))
			g.emit(fmt.Sprintf("        %s = false;", dfName))
			g.emitDrop(name, t, isPtr)
			g.emit("    }")
		} else {
			g.emitDrop(name, t, isPtr)
		}
	default:
		str := g.hirInstructionStr(i)
		if str != "" {
			g.emit(fmt.Sprintf("    %s;", str))
		}
	}
	g.cleanMovedHeapPointers(inst)
}

func (g *Generator) hirInstructionStr(inst hir.Instruction) string {
	switch i := inst.(type) {
	case *hir.Alloca:
		return ""
	case *hir.Drop:
		return ""
	case *hir.Load:
		opStr := g.hirOperandStr(i.Src)
		if strings.Contains(opStr, "({") {
			return opStr
		}
		if g.isPointerTypeInC(i.Type) && g.isOperandPointerInC(i.Src) {
			return opStr
		}
		if g.cType(i.Type) == g.cTypeOfOperand(i.Src) {
			return opStr
		}
		if !g.isOperandPointerInC(i.Src) {
			return opStr
		}
		return fmt.Sprintf("*(%s)", opStr)
	case *hir.AddressOf:
		isLVal := false
		switch o := i.Val.(type) {
		case *hir.VarOperand:
			isLVal = true
		case *hir.InstOperand:
			switch inst := o.Inst.(type) {
			case *hir.FieldAccess, *hir.IndexAccess, *hir.Load, *hir.Deref:
				isLVal = true
			case *hir.ASTExpr:
				switch inst.ASTNode.(type) {
				case *ast.Identifier, *ast.SelectorExpression, *ast.IndexExpression:
					isLVal = true
				}
			}
		}

		opStr := g.hirOperandStr(i.Val)
		if strings.Contains(opStr, "({") {
			isLVal = false
		}

		cType1 := g.cType(i.Type)
		cType2 := g.cTypeOfOperand(i.Val)

		if i.Operator == "@" {
			if strings.Contains(opStr, "({") {
				return opStr
			}

			if varOp, ok := i.Val.(*hir.VarOperand); ok && varOp.Symbol != nil && g.hasDropFlag(varOp.Symbol) {
				dfName := g.dropFlagName(varOp.Symbol)
				return fmt.Sprintf("({ %s _tmp = %s; %s = false; _tmp; })", cType2, opStr, dfName)
			} else if instOp, ok := i.Val.(*hir.InstOperand); ok {
				needsNullify := false
				if _, ok := instOp.Inst.(*hir.IndexAccess); ok {
					needsNullify = true
				} else if _, ok := instOp.Inst.(*hir.FieldAccess); ok {
					needsNullify = true
				} else if astExpr, ok := instOp.Inst.(*hir.ASTExpr); ok {
					if _, ok := astExpr.ASTNode.(*ast.IndexExpression); ok {
						needsNullify = true
					} else if _, ok := astExpr.ASTNode.(*ast.SelectorExpression); ok {
						needsNullify = true
					}
				}

				if needsNullify {
					nullifyStr := ""
					if strings.HasSuffix(cType2, "*") {
						nullifyStr = fmt.Sprintf("%s = NULL; ", opStr)
					} else {
						nullifyStr = fmt.Sprintf("memset(&(%s), 0, sizeof(%s)); ", opStr, cType2)
					}
					return fmt.Sprintf("({ %s _tmp = %s; %s_tmp; })", cType2, opStr, nullifyStr)
				}
			}
		}

		if cType1 == cType2 {
			return opStr
		}
		if isLVal {
			return fmt.Sprintf("&(%s)", opStr)
		} else {
			baseType := i.Type
			if pt, ok := i.Type.(*types.PointerType); ok {
				baseType = pt.Base
			}
			return fmt.Sprintf("((%s[]){ %s })", g.cType(baseType), opStr)
		}
	case *hir.Deref:
		opStr := g.hirOperandStr(i.Val)
		if strings.Contains(opStr, "({") {
			return opStr
		}
		if g.isPointerTypeInC(i.Type) && g.isOperandPointerInC(i.Val) {
			return opStr
		}
		if g.cType(i.Type) == g.cTypeOfOperand(i.Val) {
			return opStr
		}
		if !g.isOperandPointerInC(i.Val) {
			return opStr
		}
		return fmt.Sprintf("*(%s)", opStr)
	case *hir.Call:
		if i.ASTNode != nil && len(i.Args) == 1 {
			funcExpr := i.ASTNode.Function
			if idx, ok := funcExpr.(*ast.IndexExpression); ok {
				funcExpr = idx.Left
			}
			var castSym *semantic.Symbol
			if ident, ok := funcExpr.(*ast.Identifier); ok {
				castSym = g.SemanticInfo.Uses[ident]
				if castSym == nil {
					castSym = g.findSymbolByName(ident.Value)
				}
			} else if sel, ok := funcExpr.(*ast.SelectorExpression); ok {
				castSym = g.SemanticInfo.Uses[sel.Field]
			}

			if castSym != nil && castSym.Kind == semantic.SymType {
				if _, isPrim := castSym.Type.(*types.PrimitiveType); isPrim {
					argType := i.Args[0].GetType()
					if argType != nil {
						unwrapped := types.UnwrapLease(argType)
						if _, isProto := unwrapped.(*types.ProtocolType); isProto {
							castType := g.SemanticInfo.Types[i.ASTNode]
							if castType == nil {
								castType = castSym.Type
							}
							isLease := false
							if pt, ok := castType.(*types.PointerType); ok && pt.Leased {
								isLease = true
							}
							argStr := g.hirOperandStr(i.Args[0])
							memberOp := ".data"
							if g.isOperandPointerInC(i.Args[0]) {
								memberOp = "->data"
							}
							if isLease || !g.TargetIsValue {
								return fmt.Sprintf("((%s*)((%s)%s))", g.cType(castSym.Type), argStr, memberOp)
							} else {
								return fmt.Sprintf("(*((%s*)((%s)%s)))", g.cType(castSym.Type), argStr, memberOp)
							}
						}
					}
					argStr := g.hirOperandStr(i.Args[0])
					return fmt.Sprintf("((%s)(%s))", g.cType(castSym.Type), argStr)
				} else if _, isStruct := castSym.Type.(*types.StructType); isStruct {
					argType := i.Args[0].GetType()
					if argType != nil {
						unwrapped := types.UnwrapLease(argType)
						if _, isProto := unwrapped.(*types.ProtocolType); isProto {
							castType := g.SemanticInfo.Types[i.ASTNode]
							if castType == nil {
								castType = castSym.Type
							}
							argStr := g.hirOperandStr(i.Args[0])
							memberOp := ".data"
							if g.isOperandPointerInC(i.Args[0]) {
								memberOp = "->data"
							}
							return fmt.Sprintf("((%s*)((%s)%s))", g.cType(castType), argStr, memberOp)
						}
					}
				}
			}

			// Fallback check based on resolved expression types
			exprType := g.SemanticInfo.Types[i.ASTNode]
			argType := i.Args[0].GetType()
			if exprType != nil && argType != nil {
				unwrappedArg := types.UnwrapLease(argType)
				if _, isProto := unwrappedArg.(*types.ProtocolType); isProto {
					unwrappedExpr := types.UnwrapLease(exprType)
					if pt, ok := unwrappedExpr.(*types.PointerType); ok {
						unwrappedExpr = pt.Base
					}
					if st, ok := unwrappedExpr.(*types.StructType); ok {
						argStr := g.hirOperandStr(i.Args[0])
						memberOp := ".data"
						if g.isOperandPointerInC(i.Args[0]) {
							memberOp = "->data"
						}
						return fmt.Sprintf("((%s*)((%s)%s))", g.mangledTypeName(st), argStr, memberOp)
					}
				}
			}
		}

		var ft *types.FunctionType
		if i.ASTNode != nil {
			t := types.UnwrapLease(g.SemanticInfo.Types[i.ASTNode.Function])
			if ftType, ok := t.(*types.FunctionType); ok {
				ft = ftType
			}
		}
		if ft == nil && i.FuncSymbol != nil {
			if fType, ok := i.FuncSymbol.Type.(*types.FunctionType); ok {
				ft = fType
			}
		}

		var callStr string
		if i.ASTNode != nil {
			isVariableCall := true
			if i.FuncSymbol != nil && i.FuncSymbol.Kind == semantic.SymFunc {
				isVariableCall = false
			}
			if ft != nil && isVariableCall {
				var argsStr []string
				argsStr = append(argsStr, "_c.env")
				argsStr = append(argsStr, g.packCallArguments(i.Args, ft)...)
				callStr = fmt.Sprintf("({ nr_closure_t _c = %s; ((%s)_c.fn_ptr)(%s); })",
					g.exprToString(i.ASTNode.Function),
					g.getCFunctionPointerType(ft),
					strings.Join(argsStr, ", "))
			}
		}

		if callStr == "" {
			var argsStr []string
			isExtern := false
			if i.FuncSymbol != nil {
				sym := i.FuncSymbol
				if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
					if ast.GetAttribute(fnStmt.Attributes, "plugin_override") == nil {
						isExtern = true
					}
				} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
					isExtern = true
				} else if sym.DefNode == nil && sym.Kind == semantic.SymFunc {
					isExtern = true
				}
				
				if cStdlibFunctions[sym.Name] && sym.Name == g.mangleName(sym) {
					isExtern = true
				}


			}
			if !isExtern {
				argsStr = append(argsStr, "_env_ptr")
			}

			name := i.FuncName
			if i.ASTNode != nil {
				if mangled, ok := g.SemanticInfo.MonomorphizedNames[i.ASTNode]; ok {
					name = mangled
				} else if i.FuncSymbol != nil {
					name = g.mangleName(i.FuncSymbol)
				}
			} else if i.FuncSymbol != nil {
				name = g.mangleName(i.FuncSymbol)
			}

			if strings.Contains(name, "Take") {
				_, ok := g.SemanticInfo.MonomorphizedNames[i.ASTNode]
				if g.DebugSemantic {
					fmt.Printf("[DEBUG-CODEGEN-CALL] name=%s, pkg=%s, FuncSymbol.Name=%s, in_MonomorphizedNames=%v\n", name, g.getSymbolPackage(i.FuncSymbol), i.FuncSymbol.Name, ok)
				}
			}

			if strings.Contains(name, "LoadPtr") {
				if g.DebugSemantic {
					fmt.Printf("[DEBUG-LOADPTR] name=%s argsStr=%v\n", name, argsStr)
				}
			}

			if ft == nil {
				if specSym, ok := g.Functions[name]; ok && specSym.Type != nil {
					if t, ok := specSym.Type.(*types.FunctionType); ok {
						ft = t
					}
				}
			}

			argsStr = append(argsStr, g.packCallArguments(i.Args, ft)...)
			callStr = fmt.Sprintf("%s(%s)", name, strings.Join(argsStr, ", "))
		}

		if i.Type != nil && i.Type.Name() == "str" && !g.NoTempWrap {
			return fmt.Sprintf("nr_temp_str(%s)", callStr)
		}
		return callStr
	case *hir.VariantConstructor:
		var argsStr []string
		oldNoTemp := g.NoTempWrap
		g.NoTempWrap = true
		oldTargetVal := g.TargetIsValue
		g.TargetIsValue = true
		for _, arg := range i.Args {
			argsStr = append(argsStr, g.hirOperandStr(arg))
		}
		g.TargetIsValue = oldTargetVal
		g.NoTempWrap = oldNoTemp
		return fmt.Sprintf("%s_%s_make(%s)", g.mangledTypeName(i.SumType), i.VariantName, strings.Join(argsStr, ", "))
	case *hir.Alloc:
		if i.IsArray {
			pt := i.Type.(*types.PointerType)
			cElemType := g.cType(pt.Base)
			return fmt.Sprintf("({ int _n = %s; void* _p = nr_malloc_debug(_n * sizeof(%s), \"%s\", %d); nr_header_t* _h = (nr_header_t*)((char*)_p - NR_HEADER_SIZE); _h->count = _n; _h->elem_size = sizeof(%s); %s* _data = (%s*)_p; _data; })",
				g.hirOperandStr(i.Val), cElemType, strings.ReplaceAll(i.PosFile, "\\", "/"), i.PosLine, cElemType, cElemType, cElemType)
		} else {
			pt, ok := i.Type.(*types.PointerType)
			var ct string
			var baseType types.NRType
			if ok {
				ct = g.cType(pt.Base)
				baseType = pt.Base
			} else {
				ct = g.cType(i.Type)
				baseType = i.Type
			}
			oldNoTemp := g.NoTempWrap
			g.NoTempWrap = true
			valStr := g.hirOperandStr(i.Val)
			g.NoTempWrap = oldNoTemp
			if g.isOperandPointerInC(i.Val) && !g.isPointerTypeInC(baseType) && !strings.Contains(valStr, "({") {
				valStr = "*" + valStr
			}
			return fmt.Sprintf("({ %s* _p = nr_malloc_debug(sizeof(%s), \"%s\", %d); *_p = %s; _p; })",
				ct, ct, strings.ReplaceAll(i.PosFile, "\\", "/"), i.PosLine, valStr)
		}
	case *hir.FieldAccess:
		opStr := g.hirOperandStr(i.Base)
		res := ""
		ptrLevel := strings.Count(g.cType(i.Base.GetType()), "*")
		if _, ok := i.Base.(*hir.VarOperand); ok && g.CurrentFunc != nil {
			if g.isOperandPointerInC(i.Base) && ptrLevel == 0 {
				ptrLevel = 1
			}
		}
		if ptrLevel == 0 {
			res = fmt.Sprintf("%s.%s", opStr, i.FieldName)
		} else if ptrLevel == 1 {
			res = fmt.Sprintf("%s->%s", opStr, i.FieldName)
		} else {
			prefix := strings.Repeat("*(", ptrLevel-1)
			suffix := strings.Repeat(")", ptrLevel-1)
			res = fmt.Sprintf("(%s%s%s)->%s", prefix, opStr, suffix, i.FieldName)
		}
		// fmt.Printf("[DEBUG] hirInstructionStr FieldAccess: %s -> %s\n", i.String(), res)
		return res
	case *hir.IndexAccess:
		cElemType := g.cType(i.Type)
		if i.NoBoundsCheck {
			return fmt.Sprintf("((((%s*)%s)[%s]))", cElemType, g.hirOperandStr(i.Base), g.hirOperandStr(i.Index))
		}
		return fmt.Sprintf("(*((%s*)array_bounds_check(%s, %s, \"\", 0)))", cElemType, g.hirOperandStr(i.Base), g.hirOperandStr(i.Index))
	case *hir.Cast:
		if _, isFn := i.Val.GetType().(*types.FunctionType); isFn {
			if pt, isPrim := i.Type.(*types.PrimitiveType); isPrim && pt.Name() == "ptr" {
				if varOp, ok := i.Val.(*hir.VarOperand); ok {
					sym := varOp.Symbol
					if sym == nil {
						sym = g.findDefSymbol(varOp.Name)
					}
					if sym != nil && sym.Kind == semantic.SymFunc {
						return fmt.Sprintf("((void*)%s)", g.mangleName(sym))
					}
				}
			}
		}
		return fmt.Sprintf("((%s)%s)", g.cType(i.Type), g.hirOperandStr(i.Val))
	case *hir.BinOp:
		leftStr := g.hirOperandStr(i.Left)
		rightStr := g.hirOperandStr(i.Right)
		if i.Op == "==" || i.Op == "!=" {
			if rightStr == "NULL" {
				ptrLevel := strings.Count(g.cType(i.Left.GetType()), "*")
				if ptrLevel >= 2 {
					leftStr = fmt.Sprintf("(*%s)", leftStr)
				}
			} else if leftStr == "NULL" {
				ptrLevel := strings.Count(g.cType(i.Right.GetType()), "*")
				if ptrLevel >= 2 {
					rightStr = fmt.Sprintf("(*%s)", rightStr)
				}
			}
		}

		if i.Op == "/" || i.Op == "%" {
			isFloat := false
			if pt, ok := types.UnwrapLease(i.Type).(*types.PrimitiveType); ok {
				if pt.Name() == "f32" || pt.Name() == "f64" {
					isFloat = true
				}
			}
			if !isFloat {
				return fmt.Sprintf("({ __auto_type _left = %s; __auto_type _right = %s; if (_right == 0) nr_panic(\"division by zero\", \"\", 0); _left %s _right; })", leftStr, rightStr, i.Op)
			}
		}

		return fmt.Sprintf("(%s %s %s)", leftStr, i.Op, rightStr)
	case *hir.UnOp:
		return fmt.Sprintf("(%s%s)", i.Op, g.hirOperandStr(i.Val))
	case *hir.Try:
		return g.exprToString(i.ASTNode)
	case *hir.ASTExpr:
		return g.exprToString(i.ASTNode)
	case *hir.Expression:
		return i.Expr
	case *hir.Store:
		return fmt.Sprintf("(%s = %s)", g.hirOperandStr(i.Dest), g.hirOperandStr(i.Val))
	case *hir.Assign:
		return fmt.Sprintf("(%s = %s)", g.hirOperandStr(i.Dest), g.hirOperandStr(i.Val))
	case *hir.InterfaceCast:
		valType := i.Val.GetType()
		if valType == nil {
			valType = types.ErrorType
		}
		valStr := g.hirOperandStr(i.Val)
		unwrapped := types.UnwrapLease(valType)
		if pt, ok := unwrapped.(*types.PointerType); ok {
			unwrapped = pt.Base
		}
		if _, ok := unwrapped.(*types.ProtocolType); ok {
			memberOp := ".data"
			memberVTable := ".vtable"
			if g.isOperandPointerInC(i.Val) {
				memberOp = "->data"
				memberVTable = "->vtable"
			}
			if strings.Contains(valStr, "({") {
				valCType := g.cTypeOfOperand(i.Val)
				return fmt.Sprintf("({ %s _iface_tmp = %s; ((%s){ .data = _iface_tmp%s, .vtable = _iface_tmp%s }); })", valCType, valStr, g.cType(i.Type), memberOp, memberVTable)
			}
			return fmt.Sprintf("((%s){ .data = (%s)%s, .vtable = (%s)%s })", g.cType(i.Type), valStr, memberOp, valStr, memberVTable)
		}
		vtableName := g.requestVTable(valType, i.Type)
		isPtrInC := strings.HasSuffix(g.cTypeOfOperand(i.Val), "*")
		isDestLeased := false

		originalValType := valType
		if instOp, ok := i.Val.(*hir.InstOperand); ok {
			if addrOf, isAddr := instOp.Inst.(*hir.AddressOf); isAddr && addrOf.Operator == "@" {
				originalValType = addrOf.Val.GetType()
			}
		}
		isPtrInNora := types.IsPointerLike(originalValType)

		var dataExpr string
		if !isDestLeased && !isPtrInNora {
			// We MUST heap allocate value types for owned interfaces
			if isPtrInC {
				// The value was passed by pointer as an optimization in C, so we dereference it to copy the value
				dataExpr = fmt.Sprintf("({ %s* _box = nr_malloc_debug(sizeof(%s), \"\", 0); *_box = *%s; _box; })", g.cType(originalValType), g.cType(originalValType), valStr)
			} else {
				// The value is passed by value in C
				dataExpr = fmt.Sprintf("({ %s* _box = nr_malloc_debug(sizeof(%s), \"\", 0); *_box = %s; _box; })", g.cType(originalValType), g.cType(originalValType), valStr)
			}
		} else {
			// Either it's already a heap pointer in Nora, or we only need a read-only borrow (stack pointer is fine)
			if isPtrInC {
				dataExpr = valStr
			} else {
				if varOp, ok := i.Val.(*hir.VarOperand); ok {
					varName := varOp.Name
					if varOp.Symbol != nil {
						varName = g.variableName(varOp.Symbol)
					}
					dataExpr = fmt.Sprintf("&%s", varName)
				} else {
					dataExpr = fmt.Sprintf("((%s[]){ %s })", g.cTypeOfOperand(i.Val), valStr)
				}
			}
		}
		return fmt.Sprintf("((%s){ .data = %s, .vtable = %s })", g.cType(i.Type), dataExpr, vtableName)
	case *hir.InterfaceCall:
		proto, ok := types.UnwrapLease(i.Base.GetType()).(*types.ProtocolType)
		if !ok {
			if pt, ok2 := types.UnwrapLease(i.Base.GetType()).(*types.PointerType); ok2 {
				proto, _ = pt.Base.(*types.ProtocolType)
			}
		}
		mName := i.MethodName
		mType := proto.Methods[mName]
		idx := g.getMethodIndex(proto, mName)
		baseStr := g.hirOperandStr(i.Base)
		isPtr := strings.HasSuffix(g.cTypeOfOperand(i.Base), "*")
		baseAccess := baseStr
		if isPtr {
			baseAccess = fmt.Sprintf("(*%s)", baseStr)
		}
		var argsStr []string
		argsStr = append(argsStr, "NULL")
		argsStr = append(argsStr, fmt.Sprintf("%s.data", baseAccess))
		for aIdx, arg := range i.Args {
			var paramType types.NRType
			lease := types.LeaseRead
			if mType != nil && aIdx < len(mType.Params) {
				paramType = mType.Params[aIdx]
				if aIdx < len(mType.ParamLeases) {
					lease = mType.ParamLeases[aIdx]
				}
			}
			argStr := g.hirOperandStr(arg)
			argStr = g.alignCallArgument(arg, paramType, lease, argStr)
			argsStr = append(argsStr, argStr)
		}
		callStr := fmt.Sprintf("(((%s)%s.vtable[%d])(%s))",
			g.getCFunctionPointerTypeWithSelf(mType),
			baseAccess,
			idx,
			strings.Join(argsStr, ", "))
		if i.Type != nil && i.Type.Name() == "str" && !g.NoTempWrap {
			return fmt.Sprintf("nr_temp_str(%s)", callStr)
		}
		return callStr
	case *hir.Spawn:
		g.spawnCounter++
		structName := fmt.Sprintf("__spawn_args_%d", g.spawnCounter)
		wrapperName := fmt.Sprintf("__spawn_wrapper_%d", g.spawnCounter)
		var structSB strings.Builder
		structSB.WriteString(fmt.Sprintf("struct %s {\n", structName))
		if i.MonitorChannel != nil {
			structSB.WriteString("    channel_t* monitor_chan;\n")
		}
		if i.Call != nil {
			for idx, arg := range i.Call.Args {
				structSB.WriteString(fmt.Sprintf("    %s arg%d;\n", g.cType(g.unwrapSpawnArgType(arg.GetType())), idx))
			}
		}
		structSB.WriteString("};\n")
		g.SpawnStructs = append(g.SpawnStructs, structSB.String())

		filename := ""
		line := 0
		if i.Call != nil && i.Call.ASTNode != nil {
			filename = normalizeDebugPath(i.Call.ASTNode.Pos().Filename)
			line = i.Call.ASTNode.Pos().Line
		}

		var wrapSB strings.Builder
		if g.EnableDebug && filename != "" {
			wrapSB.WriteString(fmt.Sprintf("#line %d \"%s\"\n", line, filename))
		}
		wrapSB.WriteString(fmt.Sprintf("static void %s(void* p) {\n", wrapperName))
		if g.EnableDebug && filename != "" {
			wrapSB.WriteString("    if (nr_fiber_current() == __nora_step_target_fiber) {\n")
			wrapSB.WriteString("        __nora_step_target_fiber = NULL;\n")
			wrapSB.WriteString("        NR_DEBUGBREAK();\n")
			wrapSB.WriteString("    }\n")
			wrapSB.WriteString("    __nora_fiber_started(nr_fiber_parent());\n")
			wrapSB.WriteString(fmt.Sprintf("#line %d \"%s\"\n", line, filename))
		}
		wrapSB.WriteString(fmt.Sprintf("    struct %s* args = (struct %s*)p;\n", structName, structName))
		isExtern := false
		if i.Call != nil && i.Call.FuncSymbol != nil {
			sym := i.Call.FuncSymbol
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && (fnStmt.IsExtern || fnStmt.IsExport) {
				if ast.GetAttribute(fnStmt.Attributes, "plugin_override") == nil {
					isExtern = true
				}
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			}


		}
		funcName := "anonymous"
		if i.Call != nil {
			funcName = i.Call.FuncName
			if i.Call.FuncSymbol != nil {
				funcName = g.mangleName(i.Call.FuncSymbol)
			}
		}
		wrapSB.WriteString("    void* self = nr_fiber_current();\n")
		wrapSB.WriteString("    if (self && setjmp(*nr_fiber_panic_buf_ptr(self)) != 0) {\n")
		// Catch the panic
		if i.MonitorChannel != nil {
			wrapSB.WriteString("        if (args->monitor_chan != NULL) {\n")
			wrapSB.WriteString("            const char* msg = nr_fiber_panic_msg(self);\n")
			wrapSB.WriteString("            if (msg == NULL) msg = \"unknown panic\";\n")
			wrapSB.WriteString("            char* panic_str = nr_str_from_cstring((void*)msg);\n")
			wrapSB.WriteString("            channel_send(args->monitor_chan, &panic_str);\n")
			wrapSB.WriteString("        }\n")
		}
		// leave the else branch.
		wrapSB.WriteString("    } else {\n")
		wrapSB.WriteString(fmt.Sprintf("        %s(", funcName))
		if !isExtern {
			wrapSB.WriteString("NULL")
			if i.Call != nil && len(i.Call.Args) > 0 {
				wrapSB.WriteString(", ")
			}
		}
		var ft *types.FunctionType
		if i.Call != nil && i.Call.FuncSymbol != nil {
			if fType, ok := i.Call.FuncSymbol.Type.(*types.FunctionType); ok {
				ft = fType
			}
		}
		if i.Call != nil {
			for idx := range i.Call.Args {
				if idx > 0 {
					wrapSB.WriteString(", ")
				}
				t := i.Call.Args[idx].GetType()
				unwrapped := g.unwrapSpawnArgType(t)
				paramCType := ""
				if ft != nil && idx < len(ft.Params) {
					lease := types.LeaseRead
					if idx < len(ft.ParamLeases) {
						lease = ft.ParamLeases[idx]
					}
					paramCType = g.cParamType(ft.Params[idx], lease)
				}
				memberCType := g.cType(unwrapped)
				if strings.HasSuffix(paramCType, "*") && !strings.HasSuffix(memberCType, "*") {
					wrapSB.WriteString(fmt.Sprintf("&args->arg%d", idx))
				} else {
					wrapSB.WriteString(fmt.Sprintf("args->arg%d", idx))
				}
			}
		}
		wrapSB.WriteString(");\n")
		wrapSB.WriteString("    }\n")
		if i.MonitorChannel != nil {
			wrapSB.WriteString("    if (args->monitor_chan) channel_free(args->monitor_chan);\n")
		}
		if i.Call != nil {
			for idx, arg := range i.Call.Args {
				t := arg.GetType()
				if g.isChanType(t) {
					wrapSB.WriteString(fmt.Sprintf("    channel_free(args->arg%d);\n", idx))
				}
			}
		}
		wrapSB.WriteString("    nr_flush_temps();\n")
		wrapSB.WriteString("    nr_free_untracked(args);\n")
		wrapSB.WriteString("}\n")
		g.SpawnWrappers = append(g.SpawnWrappers, wrapSB.String())

		var initBlock strings.Builder
		initBlock.WriteString(fmt.Sprintf("({ struct %s* _args = (struct %s*)nr_malloc_untracked(sizeof(struct %s)); ", structName, structName, structName))
		if i.MonitorChannel != nil {
			valStr := g.hirOperandStr(i.MonitorChannel)
			initBlock.WriteString(fmt.Sprintf("_args->monitor_chan = %s; ", valStr))
			initBlock.WriteString(fmt.Sprintf("if (_args->monitor_chan) channel_ref(_args->monitor_chan); "))
		}
		if i.Call != nil {
			for idx, arg := range i.Call.Args {
				valStr := g.hirOperandStr(arg)
				targetType := g.unwrapSpawnArgType(arg.GetType())
				targetCType := g.cType(targetType)
				operandCType := g.cTypeOfOperand(arg)

				// Count asterisks to align pointers
				targetStars := 0
				for i := len(targetCType) - 1; i >= 0; i-- {
					if targetCType[i] == '*' {
						targetStars++
					} else if targetCType[i] != ' ' {
						break
					}
				}
				operandStars := 0
				for i := len(operandCType) - 1; i >= 0; i-- {
					if operandCType[i] == '*' {
						operandStars++
					} else if operandCType[i] != ' ' {
						break
					}
				}

				if operandStars > targetStars {
					// Only dereference if we are not working with raw 'str' or 'ptr' (since they are built-in pointers and shouldn't be blindly dereferenced)
					if targetType.Name() != "ptr" && targetType.Name() != "str" {
						valStr = strings.Repeat("*", operandStars-targetStars) + valStr
					}
				} else if targetStars > operandStars {
					valStr = "&(" + valStr + ")"
				}

				initBlock.WriteString(fmt.Sprintf("_args->arg%d = %s; ", idx, valStr))
				if g.isChanType(targetType) {
					initBlock.WriteString(fmt.Sprintf("channel_ref(_args->arg%d); ", idx))
				}
			}
		}
		callName := "anonymous"
		filename = ""
		line = 0
		if i.Call != nil {
			callName = i.Call.FuncName
			if i.Call.ASTNode != nil {
				filename = normalizeDebugPath(i.Call.ASTNode.Pos().Filename)
				line = i.Call.ASTNode.Pos().Line
			}
		}
		initBlock.WriteString(fmt.Sprintf("scheduler_spawn(%s, _args, \"%s\", \"%s\", %d); })", wrapperName, callName, filename, line))
		return initBlock.String()
	case *hir.ChanSend:
		t := i.Val.GetType()
		if t == nil {
			t = types.I32
		}
		return fmt.Sprintf("({ channel_t* _c = %s; %s _send_val = %s; channel_send(_c, &_send_val); })",
			g.hirOperandStr(i.Chan), g.cType(t), g.hirOperandStr(i.Val))
	case *hir.ChanRecv:
		return fmt.Sprintf("({ %s _res; channel_recv(%s, &_res); _res; })",
			g.cType(i.Type), g.hirOperandStr(i.Chan))
	case *hir.Lambda:
		fnName := i.FuncName
		envStructName := fnName + "_env_t"

		scope := g.SemanticInfo.Scopes[i.ASTNode]
		hasCaptures := false
		if scope != nil && scope.Captures != nil && len(scope.Captures) > 0 {
			hasCaptures = true
		}

		var sb strings.Builder
		sb.WriteString("({ nr_closure_t _c; ")
		if hasCaptures {
			sb.WriteString(fmt.Sprintf("%s* _env_local = nr_malloc(sizeof(%s)); ", envStructName, envStructName))

			var captures []*semantic.Symbol
			for sym := range scope.Captures {
				captures = append(captures, sym)
			}
			sort.Slice(captures, func(x, y int) bool {
				return captures[x].Name < captures[y].Name
			})
			for _, cap := range captures {
				name := g.mangleName(cap)
				rhs := name
				if g.CurrentLambda != nil {
					parentScope := g.SemanticInfo.Scopes[g.CurrentLambda]
					if parentScope != nil && parentScope.Captures != nil && parentScope.Captures[cap] {
						rhs = fmt.Sprintf("_env->%s", name)
					}
				}

				if rhs == name {
					if cap.Kind == semantic.SymParam {
						if g.shouldPassByPointer(cap.Type, cap.LeaseKind) && !g.isPointerTypeInC(cap.Type) {
							rhs = "*" + rhs
						}
					}
				}

				sb.WriteString(fmt.Sprintf("_env_local->%s = %s; ", name, rhs))
			}

			sb.WriteString("_c.env = _env_local; ")
			sb.WriteString(fmt.Sprintf("_c.drop_fn = %s_env_drop; ", fnName))
		} else {
			sb.WriteString("_c.env = NULL; ")
			sb.WriteString("_c.drop_fn = NULL; ")
		}
		sb.WriteString(fmt.Sprintf("_c.fn_ptr = %s; ", fnName))
		sb.WriteString("_c; })")
		return sb.String()
	}
	return ""
}

func (g *Generator) hirOperandStr(op hir.Operand) string {
	switch o := op.(type) {
	case *hir.LiteralOperand:
		if o.Value == "none" {
			return "NULL"
		}
		if o.Type != nil && o.Type.Name() == "str" {
			rawStr, err := strconv.Unquote(o.Value)
			if err == nil {
				varName := g.registerStringLiteral(rawStr)
				return fmt.Sprintf("((char*)%s.data)", varName)
			}
		}
		return o.Value
	case *hir.VarOperand:
		var sym *semantic.Symbol
		if o.Symbol != nil {
			sym = o.Symbol
		} else {
			sym = g.findDefSymbol(o.Name)
		}
		if sym != nil && sym.Kind == semantic.SymFunc {
			isExtern := false
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && (fnStmt.IsExtern || fnStmt.IsExport) {
				if ast.GetAttribute(fnStmt.Attributes, "plugin_override") == nil {
					isExtern = true
				}
			} else if _, ok := sym.DefNode.(*ast.ExternStatement); ok {
				isExtern = true
			}
			
			if cStdlibFunctions[sym.Name] && sym.Name == g.mangleName(sym) {
				isExtern = true
			}


			if isExtern {
				return g.mangleName(sym)
			}
			return fmt.Sprintf("((nr_closure_t){ .env = NULL, .fn_ptr = %s, .drop_fn = NULL })", g.mangleName(sym))
		}
		if o.Symbol != nil {
			return g.variableName(o.Symbol)
		}
		return o.Name
	case *hir.InstOperand:
		res := g.hirInstructionStr(o.Inst)
		if res == "" {
			if g.DebugSemantic {
				fmt.Printf("DEBUG: hirInstructionStr returned EMPTY for %T: %+v\n", o.Inst, o.Inst)
			}
		}
		return res
	}
	if g.DebugSemantic {
		fmt.Printf("DEBUG: hirOperandStr returned EMPTY for %T\n", op)
	}
	return ""
}

func (g *Generator) findCurrentParamSymbol(name string) *semantic.Symbol {
	if g.CurrentFunc != nil && g.CurrentFunc.DefNode != nil {
		if fnStmt, ok := g.CurrentFunc.DefNode.(*ast.FunctionStatement); ok {
			for _, p := range fnStmt.Parameters {
				if p.Name != nil && p.Name.Value == name {
					if sym := g.SemanticInfo.Defs[p.Name]; sym != nil {
						return sym
					}
				}
			}
		}
	}
	if g.CurrentLambda != nil {
		for _, p := range g.CurrentLambda.Parameters {
			if p.Name != nil && p.Name.Value == name {
				if sym := g.SemanticInfo.Defs[p.Name]; sym != nil {
					return sym
				}
			}
		}
	}
	return nil
}

func (g *Generator) cTypeOfOperand(op hir.Operand) string {
	if op == nil {
		return "void"
	}
	t := op.GetType()
	if t == nil {
		return "void"
	}
	if instOp, ok := op.(*hir.InstOperand); ok {
		switch inst := instOp.Inst.(type) {
		case *hir.AddressOf:
			if inst.Operator == "@" {
				return g.cTypeOfOperand(inst.Val)
			}
			if g.cType(inst.Type) == g.cTypeOfOperand(inst.Val) {
				return g.cTypeOfOperand(inst.Val)
			}
		case *hir.Load:
			if g.cType(inst.Type) == g.cTypeOfOperand(inst.Src) {
				return g.cTypeOfOperand(inst.Src)
			}
		case *hir.Deref:
			if g.cType(inst.Type) == g.cTypeOfOperand(inst.Val) {
				return g.cTypeOfOperand(inst.Val)
			}
		}
	}
	if varOp, ok := op.(*hir.VarOperand); ok {
		if sym := g.findCurrentParamSymbol(varOp.Name); sym != nil {
			return g.cParamType(sym.Type, sym.LeaseKind)
		}
		if varOp.Symbol != nil {
			if varOp.Symbol.Kind == semantic.SymParam {
				return g.cParamType(varOp.Symbol.Type, varOp.Symbol.LeaseKind)
			}
			return g.cType(t)
		}
	}
	return g.cType(t)
}

func (g *Generator) alignCallArgument(arg hir.Operand, paramType types.NRType, lease types.LeaseKind, argStr string) string {
	passByPointer := false
	if paramType != nil && g.shouldPassByPointer(types.UnwrapLease(paramType), lease) {
		passByPointer = true
	}

	targetCType := "void*"
	if paramType != nil {
		targetCType = g.cParamType(paramType, lease)
	}
	operandCType := g.cTypeOfOperand(arg)
	targetStars := 0
	for i := len(targetCType) - 1; i >= 0; i-- {
		if targetCType[i] == '*' {
			targetStars++
		} else if targetCType[i] != ' ' {
			break
		}
	}
	operandStars := 0
	for i := len(operandCType) - 1; i >= 0; i-- {
		if operandCType[i] == '*' {
			operandStars++
		} else if operandCType[i] != ' ' {
			break
		}
	}

	if operandStars > targetStars {
		if paramType != nil && paramType.Name() != "ptr" && paramType.Name() != "str" && types.UnwrapLease(paramType).Name() != "ptr" && types.UnwrapLease(paramType).Name() != "str" {
			argStr = strings.Repeat("*", operandStars-targetStars) + argStr
		}
	} else if passByPointer && !g.isOperandPointerInC(arg) {
		argStr = fmt.Sprintf("((%s[]){ %s })", g.cType(types.UnwrapLease(paramType)), argStr)
	} else if !passByPointer && g.isOperandPointerInC(arg) {
		if paramType == nil || (!g.isPointerTypeInC(types.UnwrapLease(paramType)) && paramType.Name() != "ptr" && types.UnwrapLease(paramType).Name() != "ptr" && paramType.Name() != "str" && types.UnwrapLease(paramType).Name() != "str") {
			argStr = "*" + argStr
		}
	}
	return argStr
}

func (g *Generator) packCallArguments(args []hir.Operand, ft *types.FunctionType) []string {
	var argsStr []string
	for idx, arg := range args {
		var paramType types.NRType
		lease := types.LeaseRead
		if ft != nil {
			hasReceiver := false
			if ft.Receiver != nil && len(args) == len(ft.Params)+1 {
				hasReceiver = true
			}

			if hasReceiver {
				if idx == 0 {
					paramType = ft.Receiver
					lease = ft.ReceiverLease
				} else if idx-1 < len(ft.Params) {
					paramType = ft.Params[idx-1]
					if idx-1 < len(ft.ParamLeases) {
						lease = ft.ParamLeases[idx-1]
					}
				}
			} else {
				if idx < len(ft.Params) {
					paramType = ft.Params[idx]
					if idx < len(ft.ParamLeases) {
						lease = ft.ParamLeases[idx]
					}
				}
			}
		}

		isLease := false
		if paramType != nil {
			if pt, ok := paramType.(*types.PointerType); ok && pt.Leased {
				isLease = true
			}
		}
		isOwnedStr := (arg.GetType() != nil && arg.GetType().Name() == "str") && !isLease

		oldNoTemp := g.NoTempWrap
		if isOwnedStr {
			g.NoTempWrap = true
		}
		argStr := g.hirOperandStr(arg)
		g.NoTempWrap = oldNoTemp

		argStr = g.alignCallArgument(arg, paramType, lease, argStr)
		argsStr = append(argsStr, argStr)
	}
	return argsStr
}

func (g *Generator) isOperandPointerInC(op hir.Operand) bool {
	if op == nil {
		return false
	}
	if instOp, ok := op.(*hir.InstOperand); ok {
		switch inst := instOp.Inst.(type) {
		case *hir.ASTExpr:
			switch node := inst.ASTNode.(type) {
			case *ast.IndexExpression, *ast.SelectorExpression:
				t := g.SemanticInfo.Types[node]
				if t != nil && g.isPointerTypeInC(t) {
					return true
				}
				return false
			}
		}
	}
	cTypeStr := g.cTypeOfOperand(op)
	hasSuffix := strings.HasSuffix(cTypeStr, "*")
	// fmt.Printf("[DEBUG] isOperandPointerInC: op %s, Nora Type: %s, C Type: %s, hasSuffix: %v\n", op.String(), op.GetType().Name(), cTypeStr, hasSuffix)
	return hasSuffix
}

func (g *Generator) isHIRTemporaryHeapPointer(op hir.Operand) bool {
	if op == nil {
		return false
	}
	instOp, ok := op.(*hir.InstOperand)
	if !ok {
		return false
	}
	switch instOp.Inst.(type) {
	case *hir.AddressOf, *hir.Deref:
		return false
	}
	t := op.GetType()
	if t == nil {
		return false
	}
	pt, isPointer := t.(*types.PointerType)
	if !isPointer {
		return false
	}
	if pt.Leased && pt.Kind != types.LeaseMove {
		return false
	}
	return g.isHeapAllocated(t)
}

func (g *Generator) wrapHIRTemporaryHeapPointer(op hir.Operand, valStr string) string {
	t := op.GetType()
	cTypeStr := g.cType(t)
	baseType := t.(*types.PointerType).Base
	cBaseTypeStr := g.cType(baseType)

	return fmt.Sprintf("({ %s _temp_ptr = %s; %s _val; memset(&_val, 0, sizeof(_val)); if (_temp_ptr) { _val = *_temp_ptr; nr_free(_temp_ptr); } _val; })",
		cTypeStr, valStr, cBaseTypeStr)
}

func (g *Generator) getMovedHeapPointerSymbol(op hir.Operand) *semantic.Symbol {
	var valOp hir.Operand
	isExplicitMove := false
	if addr, ok := op.(*hir.InstOperand); ok {
		if addrOf, ok := addr.Inst.(*hir.AddressOf); ok {
			valOp = addrOf.Val
			if addrOf.Operator == "@" {
				isExplicitMove = true
			}
		}
	} else {
		valOp = op
	}
	if !isExplicitMove {
		return nil
	}
	if valOp == nil {
		return nil
	}
	if instOp, ok := valOp.(*hir.InstOperand); ok {
		if loadInst, ok := instOp.Inst.(*hir.Load); ok {
			valOp = loadInst.Src
		}
	}
	if varOp, ok := valOp.(*hir.VarOperand); ok {
		sym := varOp.Symbol
		if sym == nil {
			sym = g.findDefSymbol(varOp.Name)
		}
		if sym != nil {
			if types.IsPointerLike(sym.Type) {
				if types.IsOwnedType(sym.Type) {
					return sym
				}
				if pt, ok := sym.Type.(*types.PointerType); ok && pt.Kind == types.LeaseMove {
					return sym
				}
			}
		}
	}
	return nil
}

func (g *Generator) cleanMovedHeapPointers(inst hir.Instruction) {
	processCall := func(c *hir.Call) {
		var ft *types.FunctionType
		if c.ASTNode != nil {
			if t, ok := g.SemanticInfo.Types[c.ASTNode.Function].(*types.FunctionType); ok {
				ft = t
			}
		}
		if ft == nil && c.FuncSymbol != nil {
			if fType, ok := c.FuncSymbol.Type.(*types.FunctionType); ok {
				ft = fType
			}
		}
		for argIdx, arg := range c.Args {
			if sym := g.getMovedHeapPointerSymbol(arg); sym != nil {
				var paramType types.NRType
				if ft != nil {
					paramIdx := argIdx
					if ft.Receiver != nil {
						if argIdx == 0 {
							paramType = ft.Receiver
						} else {
							paramIdx = argIdx - 1
						}
					}
					if paramType == nil && paramIdx < len(ft.Params) {
						paramType = ft.Params[paramIdx]
					}
				}
				if paramType != nil {
					if !types.IsPointerLike(paramType) {
						g.emit(fmt.Sprintf("    nr_free(%s); %s = NULL;", g.variableName(sym), g.variableName(sym)))
					}
				}
			}
		}
	}

	processStore := func(dest hir.Operand, val hir.Operand) {
		if sym := g.getMovedHeapPointerSymbol(val); sym != nil {
			destType := dest.GetType()
			if destType != nil {
				if !types.IsPointerLike(destType) {
					g.emit(fmt.Sprintf("    nr_free(%s); %s = NULL;", g.variableName(sym), g.variableName(sym)))
				}
			}
		}
	}

	switch i := inst.(type) {
	case *hir.Call:
		processCall(i)
	case *hir.Store:
		processStore(i.Dest, i.Val)
		if instOp, ok := i.Val.(*hir.InstOperand); ok {
			if call, ok := instOp.Inst.(*hir.Call); ok {
				processCall(call)
			}
		}
	case *hir.Assign:
		processStore(i.Dest, i.Val)
		if instOp, ok := i.Val.(*hir.InstOperand); ok {
			if call, ok := instOp.Inst.(*hir.Call); ok {
				processCall(call)
			}
		}
	}
}
