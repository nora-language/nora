package hir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/topology"
	"github.com/nora-language/nora/pkg/types"
)

var globalLambdaCounter int64

type Lowerer struct {
	SemanticInfo        *semantic.SemanticInfo
	Solver              *topology.Solver
	CurrentBlock        *HIRBlock
	currentASTBlock     *ast.BlockStatement
	currentASTStmtIndex int
	tempCounter         int
	DebugLog            bool
	CurrentFunc         *semantic.Symbol
	lambdaFuncs         []*Function
	lambdaCounter       int
	activeDefers        []ast.Expression
	LambdaTemps         map[ast.Node]*semantic.Symbol
}

func NewLowerer(sem *semantic.SemanticInfo, solver *topology.Solver) *Lowerer {
	return &Lowerer{
		SemanticInfo: sem,
		Solver:       solver,
		tempCounter:  0,
		lambdaFuncs:  []*Function{},
		LambdaTemps:  make(map[ast.Node]*semantic.Symbol),
	}
}

func (l *Lowerer) makeTempName() string {
	l.tempCounter++
	return fmt.Sprintf("_hir_tmp_%d", l.tempCounter)
}

func (l *Lowerer) LowerProgram(prog *ast.Program) *Program {
	hirProg := &Program{Functions: []*Function{}}
	for _, file := range prog.Files {
		for _, stmt := range file.Statements {
			if fn, ok := stmt.(*ast.FunctionStatement); ok {
				// Find function symbol
				var fnSym *semantic.Symbol
				for _, sym := range l.SemanticInfo.Defs {
					if sym == nil {
						continue
					}
					if sym.DefNode == fn {
						fnSym = sym
						break
					}
				}
				if fnSym == nil {
					for _, methods := range l.SemanticInfo.MethodSymbols {
						for _, sym := range methods {
							if sym == nil {
								continue
							}
							if sym.DefNode == fn {
								fnSym = sym
								break
							}
						}
						if fnSym != nil {
							break
						}
					}
				}
				if fnSym == nil {
					for _, sym := range l.SemanticInfo.Uses {
						if sym == nil {
							continue
						}
						if sym.DefNode == fn {
							fnSym = sym
							break
						}
					}
				}
				if fnSym == nil {
					fnSym = &semantic.Symbol{Name: fn.Name.Value}
				}

				if l.shouldSkipHIR(fnSym, fn) {
					continue
				}

				hirFn := l.lowerFunction(fnSym, fn)
				hirProg.Functions = append(hirProg.Functions, hirFn)
			}
		}
	}

	for _, instMap := range l.SemanticInfo.Instances {
		for _, inst := range instMap {
			sym := l.SemanticInfo.Defs[inst.Name]
			if sym == nil {
				// Fallback just in case
				sym = &semantic.Symbol{Name: inst.Name.Value, DefNode: inst}
			}
			if l.shouldSkipHIR(sym, inst) {
				continue
			}
			hirFn := l.lowerFunction(sym, inst)
			hirProg.Functions = append(hirProg.Functions, hirFn)
		}
	}

	hirProg.Functions = append(hirProg.Functions, l.lambdaFuncs...)
	return hirProg
}

func (l *Lowerer) emitDrop(d topology.DropInfo, field ast.Expression, index ast.Expression) {
	if d.Lambda != nil && l.LambdaTemps[d.Lambda] != nil {
		l.CurrentBlock.AddInst(&Drop{Symbol: l.LambdaTemps[d.Lambda], Lambda: d.Lambda})
	} else {
		l.CurrentBlock.AddInst(&Drop{Symbol: d.Symbol, Field: field, Index: index, Lambda: d.Lambda})
	}
}

func (l *Lowerer) lowerFunction(sym *semantic.Symbol, fn *ast.FunctionStatement) *Function {
	prevFunc := l.CurrentFunc
	l.CurrentFunc = sym
	prevDefers := l.activeDefers
	l.activeDefers = nil
	defer func() {
		l.CurrentFunc = prevFunc
		l.activeDefers = prevDefers
	}()

	hirFn := &Function{
		Name:       fn.Name.Value,
		FuncSymbol: sym,
		Params:     []string{},
		Body:       NewHIRBlock(),
	}

	for _, p := range fn.Parameters {
		hirFn.Params = append(hirFn.Params, p.Name.Value)
	}

	if fn.Body != nil {
		hirFn.Body = l.lowerBlock(fn.Body)
	}

	if len(l.activeDefers) > 0 {
		hasRet := false
		if len(hirFn.Body.Elements) > 0 {
			last := hirFn.Body.Elements[len(hirFn.Body.Elements)-1]
			if instEl, ok := last.(*InstElement); ok {
				if _, isRet := instEl.Inst.(*Ret); isRet {
					hasRet = true
				}
			}
		}
		if !hasRet {
			prevBlock := l.CurrentBlock
			l.CurrentBlock = hirFn.Body
			for i := len(l.activeDefers) - 1; i >= 0; i-- {
				l.lowerExpression(l.activeDefers[i])
			}
			l.CurrentBlock = prevBlock
		}
	}

	return hirFn
}

func (l *Lowerer) lowerBlock(block *ast.BlockStatement) *HIRBlock {
	prevBlock := l.CurrentBlock
	prevASTBlock := l.currentASTBlock
	prevASTStmtIndex := l.currentASTStmtIndex

	hirBlock := NewHIRBlock()
	l.CurrentBlock = hirBlock
	l.currentASTBlock = block

	// 1. Emit initial drops at index 0 (e.g. unused parameters)
	if l.Solver != nil && l.Solver.Drops[block] != nil {
		drops := l.Solver.Drops[block][0]
		for _, d := range drops {
			var field ast.Expression
			if d.Field != nil {
				field = d.Field
			}
			var index ast.Expression
			if d.Index != nil {
				index = d.Index
			}
			l.emitDrop(d, field, index)
		}
	}

	for i, stmt := range block.Statements {
		l.currentASTStmtIndex = i
		// 2. Emit pre-drops for this statement if they exist, EXCEPT if it is a return statement
		if _, isRet := stmt.(*ast.ReturnStatement); !isRet {
			if l.Solver != nil && l.Solver.PreDrops[block] != nil {
				drops := l.Solver.PreDrops[block][i]
				for _, d := range drops {
					var field ast.Expression
					if d.Field != nil {
						field = d.Field
					}
					var index ast.Expression
					if d.Index != nil {
						index = d.Index
					}
					l.emitDrop(d, field, index)
				}
			}
		}

		l.lowerStatement(stmt)

		// 3. Emit post-drops after this statement
		if l.Solver != nil && l.Solver.Drops[block] != nil {
			drops := l.Solver.Drops[block][i+1]
			for _, d := range drops {
				var field ast.Expression
				if d.Field != nil {
					field = d.Field
				}
				var index ast.Expression
				if d.Index != nil {
					index = d.Index
				}
				l.emitDrop(d, field, index)
			}
		}
	}

	// 4. Emit end-of-block pre-drops
	if l.Solver != nil && l.Solver.PreDrops[block] != nil {
		drops := l.Solver.PreDrops[block][len(block.Statements)]
		for _, d := range drops {
			var field ast.Expression
			if d.Field != nil {
				field = d.Field
			}
			var index ast.Expression
			if d.Index != nil {
				index = d.Index
			}
			l.emitDrop(d, field, index)
		}
	}

	l.CurrentBlock = prevBlock
	l.currentASTBlock = prevASTBlock
	l.currentASTStmtIndex = prevASTStmtIndex
	return hirBlock
}

func (l *Lowerer) emitPreDrops() {
	if l.Solver != nil && l.currentASTBlock != nil && l.Solver.PreDrops[l.currentASTBlock] != nil {
		drops := l.Solver.PreDrops[l.currentASTBlock][l.currentASTStmtIndex]
		for _, d := range drops {
			var field ast.Expression
			if d.Field != nil {
				field = d.Field
			}
			var index ast.Expression
			if d.Index != nil {
				index = d.Index
			}
			l.emitDrop(d, field, index)
		}
	}
}

func (l *Lowerer) lowerStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	// Inject LineInfo to preserve original AST statement source line mapping for Codegen
	if l.CurrentBlock != nil {
		l.CurrentBlock.AddInst(&LineInfo{ASTNode: stmt})
	}

	switch s := stmt.(type) {
	case *ast.BlockStatement:
		subBlock := l.lowerBlock(s)
		for _, el := range subBlock.Elements {
			l.CurrentBlock.AddElement(el)
		}

	case *ast.ExpressionStatement:
		op := l.lowerExpression(s.Expression)
		if instOp, ok := op.(*InstOperand); ok {
			l.CurrentBlock.AddInst(instOp.Inst)
		}

	case *ast.VarStatement:
		t := l.getType(s.Name)
		sym := l.SemanticInfo.Defs[s.Name]
		if sym == nil {
			sym = l.SemanticInfo.Uses[s.Name]
		}

		// Emit explicit VarDecl (alloca)
		alloca := &Alloca{Name: s.Name.Value, Type: t, Symbol: sym}
		l.CurrentBlock.AddInst(alloca)

		if s.Value != nil {
			valOp := l.lowerExpression(s.Value)
			if proto, ok := t.(*types.ProtocolType); ok {
				valType := l.getType(s.Value)
				if valProto, valIsProto := valType.(*types.ProtocolType); !valIsProto || !types.Equals(proto, valProto) {
					valOp = &InstOperand{Inst: &InterfaceCast{Val: valOp, Type: proto}}
				}
			}
			// Explicit store to local variable
			store := &Store{
				Dest: &VarOperand{Name: s.Name.Value, Type: t, Symbol: sym},
				Val:  valOp,
			}
			l.CurrentBlock.AddInst(store)
		}

	case *ast.AssignmentStatement:
		destOp := l.lowerExpression(s.Left)
		valOp := l.lowerExpression(s.Value)
		if proto, ok := destOp.GetType().(*types.ProtocolType); ok {
			valType := l.getType(s.Value)
			if valProto, valIsProto := valType.(*types.ProtocolType); !valIsProto || !types.Equals(proto, valProto) {
				valOp = &InstOperand{Inst: &InterfaceCast{Val: valOp, Type: proto}}
			}
		}
		store := &Store{
			Dest:         destOp,
			Val:          valOp,
			NeedsOldDrop: l.Solver != nil && l.Solver.AssignDrops[s],
		}
		l.CurrentBlock.AddInst(store)

	case *ast.ReturnStatement:
		var valOp Operand
		var tempVar *VarOperand
		var retProto *types.ProtocolType
		if l.CurrentFunc != nil {
			if ft, ok := l.CurrentFunc.Type.(*types.FunctionType); ok && ft.Return != nil {
				if proto, ok := ft.Return.(*types.ProtocolType); ok {
					retProto = proto
				}
			}
		}
		if s.ReturnValue != nil {
			t := l.getType(s.ReturnValue)
			if retProto != nil {
				t = retProto
			}

			var drops []topology.DropInfo
			if l.Solver != nil && l.currentASTBlock != nil && l.Solver.PreDrops[l.currentASTBlock] != nil {
				drops = l.Solver.PreDrops[l.currentASTBlock][l.currentASTStmtIndex]
			}

			if len(drops) > 0 && t != nil && t.Name() != "void" && t != types.ErrorType {
				tempName := l.makeTempName()
				l.CurrentBlock.AddInst(&Alloca{Name: tempName, Type: t})
				valOp = l.lowerExpression(s.ReturnValue)
				if retProto != nil {
					valType := l.getType(s.ReturnValue)
					if valProto, valIsProto := valType.(*types.ProtocolType); !valIsProto || !types.Equals(retProto, valProto) {
						valOp = &InstOperand{Inst: &InterfaceCast{Val: valOp, Type: retProto}}
					}
				}
				l.CurrentBlock.AddInst(&Store{
					Dest: &VarOperand{Name: tempName, Type: t},
					Val:  valOp,
				})
				tempVar = &VarOperand{Name: tempName, Type: t}
			} else {
				valOp = l.lowerExpression(s.ReturnValue)
				if retProto != nil {
					valType := l.getType(s.ReturnValue)
					if valProto, valIsProto := valType.(*types.ProtocolType); !valIsProto || !types.Equals(retProto, valProto) {
						valOp = &InstOperand{Inst: &InterfaceCast{Val: valOp, Type: retProto}}
					}
				}
			}
		}
		for i := len(l.activeDefers) - 1; i >= 0; i-- {
			l.lowerExpression(l.activeDefers[i])
		}
		l.emitPreDrops()
		if tempVar != nil {
			l.CurrentBlock.AddInst(&Ret{Val: tempVar})
		} else if valOp != nil {
			l.CurrentBlock.AddInst(&Ret{Val: valOp})
		} else {
			l.CurrentBlock.AddInst(&Ret{Val: &LiteralOperand{Value: "none", Type: types.ErrorType}})
		}

	case *ast.ForStatement:
		if s.Iterable != nil {
			if rangeExpr, isRange := s.Iterable.(*ast.RangeExpression); isRange {
				startOp := l.lowerExpression(rangeExpr.Start)
				endOp := l.lowerExpression(rangeExpr.End)

				endTemp := l.makeTempName()
				l.CurrentBlock.AddInst(&Alloca{Name: endTemp, Type: endOp.GetType()})
				l.CurrentBlock.AddInst(&Store{Dest: &VarOperand{Name: endTemp, Type: endOp.GetType()}, Val: endOp})

				idxTemp := l.makeTempName()
				l.CurrentBlock.AddInst(&Alloca{Name: idxTemp, Type: startOp.GetType()})
				l.CurrentBlock.AddInst(&Store{
					Dest: &VarOperand{Name: idxTemp, Type: startOp.GetType()},
					Val:  startOp,
				})

				bodyBlock := NewHIRBlock()
				prevBlock := l.CurrentBlock
				l.CurrentBlock = bodyBlock

				if s.Value != nil {
					sym := l.SemanticInfo.Defs[s.Value]
					if sym == nil {
						sym = l.SemanticInfo.Uses[s.Value]
					}
					l.CurrentBlock.AddInst(&Alloca{Name: s.Value.Value, Type: startOp.GetType(), Symbol: sym})
					l.CurrentBlock.AddInst(&Store{
						Dest: &VarOperand{Name: s.Value.Value, Type: startOp.GetType(), Symbol: sym},
						Val:  &VarOperand{Name: idxTemp, Type: startOp.GetType()},
					})
				}

				subBody := l.lowerBlock(s.Body)
				for _, el := range subBody.Elements {
					l.CurrentBlock.AddElement(el)
				}

				l.CurrentBlock.AddInst(&Store{
					Dest: &VarOperand{Name: idxTemp, Type: startOp.GetType()},
					Val: &InstOperand{Inst: &BinOp{
						Left:  &VarOperand{Name: idxTemp, Type: startOp.GetType()},
						Op:    "+",
						Right: &LiteralOperand{Value: "1", Type: startOp.GetType()},
						Type:  startOp.GetType(),
					}},
				})

				l.CurrentBlock = prevBlock

				condOp := &InstOperand{Inst: &BinOp{
					Left:  &VarOperand{Name: idxTemp, Type: startOp.GetType()},
					Op:    "<",
					Right: &VarOperand{Name: endTemp, Type: endOp.GetType()},
					Type:  types.Bool,
				}}

				hirLoop := &HIRLoop{
					Condition: condOp,
					Body:      bodyBlock,
				}
				l.CurrentBlock.AddElement(hirLoop)
			} else {
				iterOp := l.lowerExpression(s.Iterable)
				iterType := types.UnwrapLease(iterOp.GetType())
				isIterator := false

				if st, ok := iterType.(*types.StructType); ok {
					if method, exists := st.Methods["Next"]; exists {
						if mt, ok := method.(*types.FunctionType); ok {
							if retSum, ok := mt.Return.(*types.SumType); ok && strings.HasPrefix(retSum.TypeName, "Option") {
								isIterator = true
							}
						}
					}
				}

				if isIterator {
					iterTemp := l.makeTempName()
					l.CurrentBlock.AddInst(&Alloca{Name: iterTemp, Type: iterOp.GetType()})
					l.CurrentBlock.AddInst(&Store{Dest: &VarOperand{Name: iterTemp, Type: iterOp.GetType()}, Val: iterOp})

					bodyBlock := NewHIRBlock()
					prevBlock := l.CurrentBlock
					l.CurrentBlock = bodyBlock

					subBody := l.lowerBlock(s.Body)
					for _, el := range subBody.Elements {
						l.CurrentBlock.AddElement(el)
					}

					l.CurrentBlock = prevBlock

					var elemName, keyName string
					var elemType, keyType types.NRType
					if s.Value != nil {
						elemName = s.Value.Value
						sym := l.SemanticInfo.Defs[s.Value]
						if sym == nil {
							sym = l.SemanticInfo.Uses[s.Value]
						}
						elemType = sym.Type
					}
					if s.Key != nil {
						keyName = s.Key.Value
						sym := l.SemanticInfo.Defs[s.Key]
						if sym == nil {
							sym = l.SemanticInfo.Uses[s.Key]
						}
						keyType = sym.Type
					}

					var nextMangled string
					var nextSymbol *semantic.Symbol
					if s.NextCall != nil {
						nextMangled = l.SemanticInfo.MonomorphizedNames[s.NextCall]
						if sel, ok := s.NextCall.Function.(*ast.SelectorExpression); ok {
							nextSymbol = l.SemanticInfo.Uses[sel.Field]
						}
						if l.DebugLog {
							fmt.Printf("[DEBUG-LOWER-FOR] NextCall: %p, Function: %T, nextMangled: %q, nextSymbol: %p\n", s.NextCall, s.NextCall.Function, nextMangled, nextSymbol)
						}
					}

					iterLoop := &IteratorLoop{
						Iterator:    &VarOperand{Name: iterTemp, Type: iterOp.GetType()},
						NextMangled: nextMangled,
						NextSymbol:  nextSymbol,
						ElemName:    elemName,
						ElemType:    elemType,
						KeyName:     keyName,
						KeyType:     keyType,
						Body:        bodyBlock,
					}
					l.CurrentBlock.AddElement(iterLoop)
				} else {
					arrTemp := l.makeTempName()
					l.CurrentBlock.AddInst(&Alloca{Name: arrTemp, Type: iterOp.GetType()})
					l.CurrentBlock.AddInst(&Store{Dest: &VarOperand{Name: arrTemp, Type: iterOp.GetType()}, Val: iterOp})

					lenTemp := l.makeTempName()
					l.CurrentBlock.AddInst(&Alloca{Name: lenTemp, Type: types.I32})
					l.CurrentBlock.AddInst(&Store{
						Dest: &VarOperand{Name: lenTemp, Type: types.I32},
						Val: &InstOperand{Inst: &Expression{
							Expr: fmt.Sprintf("array_count(%s)", arrTemp),
							Type: types.I32,
						}},
					})

					idxTemp := l.makeTempName()
					l.CurrentBlock.AddInst(&Alloca{Name: idxTemp, Type: types.I32})
					l.CurrentBlock.AddInst(&Store{
						Dest: &VarOperand{Name: idxTemp, Type: types.I32},
						Val:  &LiteralOperand{Value: "0", Type: types.I32},
					})

					bodyBlock := NewHIRBlock()
					prevBlock := l.CurrentBlock
					l.CurrentBlock = bodyBlock

					if s.Key != nil {
						keyType := types.I32
						sym := l.SemanticInfo.Defs[s.Key]
						if sym == nil {
							sym = l.SemanticInfo.Uses[s.Key]
						}
						l.CurrentBlock.AddInst(&Alloca{Name: s.Key.Value, Type: keyType, Symbol: sym})
						l.CurrentBlock.AddInst(&Store{
							Dest: &VarOperand{Name: s.Key.Value, Type: keyType, Symbol: sym},
							Val:  &VarOperand{Name: idxTemp, Type: types.I32},
						})
					}

					if s.Value != nil {
						var elemType types.NRType = types.I32
						if lt, ok := iterOp.GetType().(*types.ListType); ok {
							elemType = lt.ElementType
						} else if iterOp.GetType() != nil && iterOp.GetType().Name() == "str" {
							elemType = types.I8
						}
						sym := l.SemanticInfo.Defs[s.Value]
						if sym == nil {
							sym = l.SemanticInfo.Uses[s.Value]
						}
						l.CurrentBlock.AddInst(&Alloca{Name: s.Value.Value, Type: elemType, Symbol: sym})
						l.CurrentBlock.AddInst(&Store{
							Dest: &VarOperand{Name: s.Value.Value, Type: elemType, Symbol: sym},
							Val: &InstOperand{Inst: &IndexAccess{
								Base:          &VarOperand{Name: arrTemp, Type: iterOp.GetType()},
								Index:         &VarOperand{Name: idxTemp, Type: types.I32},
								Type:          elemType,
								NoBoundsCheck: true,
							}},
						})
					}

					subBody := l.lowerBlock(s.Body)
					for _, el := range subBody.Elements {
						l.CurrentBlock.AddElement(el)
					}

					l.CurrentBlock.AddInst(&Store{
						Dest: &VarOperand{Name: idxTemp, Type: types.I32},
						Val: &InstOperand{Inst: &BinOp{
							Left:  &VarOperand{Name: idxTemp, Type: types.I32},
							Op:    "+",
							Right: &LiteralOperand{Value: "1", Type: types.I32},
							Type:  types.I32,
						}},
					})

					l.CurrentBlock = prevBlock

					condOp := &InstOperand{Inst: &BinOp{
						Left:  &VarOperand{Name: idxTemp, Type: types.I32},
						Op:    "<",
						Right: &VarOperand{Name: lenTemp, Type: types.I32},
						Type:  types.Bool,
					}}

					hirLoop := &HIRLoop{
						Condition: condOp,
						Body:      bodyBlock,
					}
					l.CurrentBlock.AddElement(hirLoop)
				}
			}
		} else {
			// Infinite loop: for { ... }
			bodyBlock := l.lowerBlock(s.Body)
			hirLoop := &HIRLoop{
				Condition: nil,
				Body:      bodyBlock,
			}
			l.CurrentBlock.AddElement(hirLoop)
		}

	case *ast.WhileStatement:
		condOp := l.lowerExpression(s.Condition)
		bodyBlock := l.lowerBlock(s.Body)

		hirLoop := &HIRLoop{
			Condition: condOp,
			Body:      bodyBlock,
		}
		l.CurrentBlock.AddElement(hirLoop)

	case *ast.BreakStatement:
		l.emitPreDrops()
		l.CurrentBlock.AddInst(&Expression{Expr: "break", Type: types.Void})

	case *ast.ContinueStatement:
		l.emitPreDrops()
		l.CurrentBlock.AddInst(&Expression{Expr: "continue", Type: types.Void})

	case *ast.BranchStatement:
		l.emitPreDrops()
		if s.Token.Literal == "break" {
			l.CurrentBlock.AddInst(&Expression{Expr: "break", Type: types.Void})
		} else {
			l.CurrentBlock.AddInst(&Expression{Expr: "continue", Type: types.Void})
		}

	case *ast.SelectStatement:
		cases := []SelectCase{}
		for _, c := range s.Cases {
			sc := SelectCase{
				IsDefault: c.Condition == nil,
			}
			if c.Condition != nil {
				var cond ast.Node = c.Condition
				if es, ok := cond.(*ast.ExpressionStatement); ok {
					cond = es.Expression
				}

				if se, ok := cond.(*ast.SendExpression); ok {
					sc.IsSend = true
					sc.Chan = l.lowerExpression(se.Left)
					sc.Val = l.lowerExpression(se.Right)
				} else if re, ok := cond.(*ast.ReceiveExpression); ok {
					sc.IsSend = false
					sc.Chan = l.lowerExpression(re.Value)
				} else if assign, ok := cond.(*ast.AssignmentStatement); ok {
					if re, ok2 := assign.Value.(*ast.ReceiveExpression); ok2 {
						sc.IsSend = false
						sc.Chan = l.lowerExpression(re.Value)
						sc.Val = l.lowerExpression(assign.Left)
					}
				} else if vs, ok := cond.(*ast.VarStatement); ok {
					if re, ok2 := vs.Value.(*ast.ReceiveExpression); ok2 {
						sc.IsSend = false
						sc.Chan = l.lowerExpression(re.Value)
						sc.VarName = vs.Name.Value
						sc.VarType = l.getType(vs.Name)
					}
				}
			}
			sc.Body = l.lowerBlock(c.Body)
			cases = append(cases, sc)
		}
		l.CurrentBlock.AddElement(&Select{Cases: cases})

	case *ast.DeferStatement:
		l.activeDefers = append(l.activeDefers, s.Call)
	}

}

func (l *Lowerer) lowerExpression(expr ast.Expression) Operand {
	if expr == nil {
		return &LiteralOperand{Value: "none", Type: types.ErrorType}
	}

	t := l.getType(expr)

	switch e := expr.(type) {
	case *ast.ArgumentsExpression:
		return l.lowerExpression(e.Value)
	case *ast.Identifier:
		if st, vName := l.isVariantConstructor(e); st != nil {
			vc := &VariantConstructor{SumType: st, VariantName: vName, Args: nil, Type: t}
			return &InstOperand{Inst: vc}
		}
		sym := l.SemanticInfo.Uses[e]
		if sym == nil {
			sym = l.SemanticInfo.Defs[e]
		}
		return &VarOperand{Name: e.Value, Type: t, Symbol: sym}

	case *ast.IntegerLiteral:
		return &LiteralOperand{Value: strconv.FormatInt(e.Value, 10), Type: t}

	case *ast.FloatLiteral:
		s := fmt.Sprintf("%g", e.Value)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") {
			s += ".0"
		}
		return &LiteralOperand{Value: s, Type: t}

	case *ast.Boolean:
		return &LiteralOperand{Value: strconv.FormatBool(e.Value), Type: t}

	case *ast.StringLiteral:
		return &LiteralOperand{Value: fmt.Sprintf("%q", e.Value), Type: t}

	case *ast.NoneLiteral:
		return &LiteralOperand{Value: "none", Type: t}

	case *ast.PrefixExpression:
		if e.Operator == "!" || e.Operator == "-" || e.Operator == "~" {
			rt := l.getType(e.Right)
			if rt != nil {
				urt := rt
				if pt, ok := rt.(*types.PointerType); ok && !pt.IsArray {
					urt = pt.Base
				}
				if _, isStruct := urt.(*types.StructType); isStruct {
					astExpr := &ASTExpr{ASTNode: e, Type: t}
					return &InstOperand{Inst: astExpr}
				}
			}
		}
		valOp := l.lowerExpression(e.Right)
		if e.Operator == "@" || e.Operator == "#" || e.Operator == "&" {
			// Explicit borrow / move address-of
			addr := &AddressOf{Val: valOp, Type: t, Operator: e.Operator}
			return &InstOperand{Inst: addr}
		}
		if e.Operator == "*" {
			// Explicit deref
			deref := &Deref{Val: valOp, Type: t}
			return &InstOperand{Inst: deref}
		}
		unOp := &UnOp{Op: e.Operator, Val: valOp, Type: t}
		return &InstOperand{Inst: unOp}

	case *ast.SelectorExpression:
		if st, vName := l.isVariantConstructor(e); st != nil {
			vc := &VariantConstructor{SumType: st, VariantName: vName, Args: nil, Type: t}
			return &InstOperand{Inst: vc}
		}
		baseOp := l.lowerExpression(e.Left)
		fieldAcc := &FieldAccess{Base: baseOp, FieldName: e.Field.Value, Type: t}
		return &InstOperand{Inst: fieldAcc}

	case *ast.CallExpression:
		if ident, ok := e.Function.(*ast.Identifier); ok {
			val := ident.Value
			if prim, ok := types.LookupPrimitive(val); ok {
				if _, isPrim := prim.(*types.PrimitiveType); isPrim {
					valOp := l.lowerExpression(e.Arguments[0])
					cast := &Cast{Val: valOp, Type: t}
					return &InstOperand{Inst: cast}
				}
			}
			if len(e.Arguments) == 1 {
				argType := l.getType(e.Arguments[0])
				if argType != nil {
					unwrapped := types.UnwrapLease(argType)
					if _, isProto := unwrapped.(*types.ProtocolType); isProto {
						retType := l.getType(e)
						if retType != nil {
							if pt, ok := retType.(*types.PointerType); ok {
								if _, isStruct := pt.Base.(*types.StructType); isStruct {
									return &InstOperand{Inst: &ASTExpr{ASTNode: e, Type: t}}
								}
							}
						}
					}
				}
			}
		}
		if st, vName := l.isVariantConstructor(e.Function); st != nil {
			argsOps := []Operand{}
			for _, arg := range e.Arguments {
				argsOps = append(argsOps, l.lowerExpression(arg))
			}
			vc := &VariantConstructor{SumType: st, VariantName: vName, Args: argsOps, Type: t}
			return &InstOperand{Inst: vc}
		}

		if sel, ok := e.Function.(*ast.SelectorExpression); ok {
			if sel.Field.Value == "unchecked_get" || sel.Field.Value == "unchecked_set" {
				actualType := l.getType(sel.Left)
				if actualType != nil {
					unwrapped := types.UnwrapLease(actualType)
					isCollection := false
					if _, isList := unwrapped.(*types.ListType); isList {
						isCollection = true
					} else if pt, isPtr := unwrapped.(*types.PointerType); isPtr && pt.IsArray {
						isCollection = true
					}
					if isCollection {
						if sel.Field.Value == "unchecked_get" && len(e.Arguments) == 1 {
							idxOp := l.lowerExpression(e.Arguments[0])
							baseOp := l.lowerExpression(sel.Left)
							return &InstOperand{Inst: &IndexAccess{
								Base:          baseOp,
								Index:         idxOp,
								Type:          t,
								NoBoundsCheck: true,
							}}
						} else if sel.Field.Value == "unchecked_set" && len(e.Arguments) == 2 {
							idxOp := l.lowerExpression(e.Arguments[0])
							valOp := l.lowerExpression(e.Arguments[1])
							baseOp := l.lowerExpression(sel.Left)

							idxAcc := &IndexAccess{
								Base:          baseOp,
								Index:         idxOp,
								Type:          valOp.GetType(),
								NoBoundsCheck: true,
							}

							return &InstOperand{Inst: &Store{
								Dest:         &InstOperand{Inst: idxAcc},
								Val:          valOp,
								NeedsOldDrop: false, // Array values are uninitialized or we manually drop them
							}}
						}
					}
				}
			}
		}

		var funcSym *semantic.Symbol
		if ident, ok := e.Function.(*ast.Identifier); ok {
			funcSym = l.SemanticInfo.Uses[ident]
		} else if sel, ok := e.Function.(*ast.SelectorExpression); ok {
			funcSym = l.SemanticInfo.Uses[sel.Field]
			if sel.Field.Value == "Take" {
				if l.DebugLog {
					fmt.Printf("[DEBUG-LOWER-CALL] Selector for 'Take'. funcSym=%p. Left=%T\n", funcSym, sel.Left)
				}
			}
		}

		isBuiltin := false
		if ident, ok := e.Function.(*ast.Identifier); ok {
			val := ident.Value
			if val == "len" || val == "panic" || val == "make" || val == "append" {
				isBuiltin = true
			}
		} else if sel, ok := e.Function.(*ast.SelectorExpression); ok {
			if sel.Field.Value == "clone" {
				if lt := l.getType(sel.Left); lt != nil {
					unwrapped := types.UnwrapLease(lt)
					if unwrapped.GetKind() == types.KindChan {
						isBuiltin = true
					}
				}
			}
		}
		isMacro := false
		if funcSym != nil && funcSym.DefNode != nil {
			if fnStmt, ok := funcSym.DefNode.(*ast.FunctionStatement); ok {
				if ast.GetAttribute(fnStmt.Attributes, "macro") != nil {
					isMacro = true
				}
			}
		}

		if isBuiltin || isMacro {
			astExpr := &ASTExpr{ASTNode: e, Type: t}
			return &InstOperand{Inst: astExpr}
		}

		var protoType *types.ProtocolType
		var methodSelector *ast.SelectorExpression
		if sel, ok := e.Function.(*ast.SelectorExpression); ok {
			recType := l.getType(sel.Left)
			if recType != nil {
				unwrapped := types.UnwrapLease(recType)
				if pt, ok := unwrapped.(*types.PointerType); ok {
					unwrapped = pt.Base
				}
				if proto, ok := unwrapped.(*types.ProtocolType); ok {
					protoType = proto
					methodSelector = sel
				}
			}
		}

		if protoType != nil {
			baseOp := l.lowerExpression(methodSelector.Left)
			argsOps := []Operand{}
			mName := methodSelector.Field.Value
			mType := protoType.Methods[mName]

			for i, arg := range e.Arguments {
				argOp := l.lowerExpression(arg)
				if mType != nil && i < len(mType.Params) {
					paramType := mType.Params[i]
					argType := l.getType(arg)
					if paramType != nil && argType != nil && argType != types.ErrorType {
						if proto, ok := paramType.(*types.ProtocolType); ok {
							if argProto, argIsProto := argType.(*types.ProtocolType); !argIsProto || !types.Equals(proto, argProto) {
								argOp = &InstOperand{Inst: &InterfaceCast{Val: argOp, Type: proto}}
							}
						} else {
							lease := types.LeaseRead
							if i < len(mType.ParamLeases) {
								lease = mType.ParamLeases[i]
							}
							baseType := paramType
							if pt, ok := paramType.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
								baseType = pt.Base
								lease = pt.Kind
							}
							if l.shouldPassByPointer(baseType, lease) {
								if !types.IsPointerLike(argType) {
									var ptrType types.NRType
									if _, ok := paramType.(*types.PointerType); ok {
										ptrType = paramType
									} else {
										ptrType = &types.PointerType{Base: paramType, Leased: true, Kind: lease}
									}
									borrowAddr := &AddressOf{Val: argOp, Type: ptrType}
									argOp = &InstOperand{Inst: borrowAddr}
								}
							}
						}
					}
				}
				argsOps = append(argsOps, argOp)
			}

			icall := &InterfaceCall{
				Base:       baseOp,
				MethodName: mName,
				Args:       argsOps,
				Type:       t,
			}
			return &InstOperand{Inst: icall}
		}

		funcName := e.Function.String()

		var funcType *types.FunctionType
		if ft, ok := l.getType(e.Function).(*types.FunctionType); ok {
			funcType = ft
		}

		isMethod := false
		var receiver ast.Expression
		if funcType != nil && funcType.Receiver != nil {
			if sel, ok := e.Function.(*ast.SelectorExpression); ok {
				receiver = sel.Left
				isMethod = true
			}
		}

		argsOps := []Operand{}

		// Prepend receiver if this is a method call
		if isMethod && receiver != nil {
			recOp := l.lowerExpression(receiver)
			if funcType != nil && funcType.Receiver != nil {
				paramType := funcType.Receiver
				recType := l.getType(receiver)
				if recType != nil && recType != types.ErrorType {
					lease := funcType.ReceiverLease
					baseType := paramType
					if pt, ok := paramType.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
						baseType = pt.Base
						lease = pt.Kind
					}
					if l.shouldPassByPointer(baseType, lease) {
						if !types.IsPointerLike(recType) {
							var ptrType types.NRType
							if _, ok := paramType.(*types.PointerType); ok {
								ptrType = paramType
							} else {
								ptrType = &types.PointerType{Base: paramType, Leased: true, Kind: lease}
							}
							borrowAddr := &AddressOf{Val: recOp, Type: ptrType}
							recOp = &InstOperand{Inst: borrowAddr}
						}
					}
				}
			}
			argsOps = append(argsOps, recOp)
		}

		for i, arg := range e.Arguments {
			argOp := l.lowerExpression(arg)
			if funcType != nil {
				paramIdx := i
				if paramIdx < len(funcType.Params) {
					paramType := funcType.Params[paramIdx]
					argType := l.getType(arg)
					if paramType != nil && argType != nil && argType != types.ErrorType {
						if proto, ok := paramType.(*types.ProtocolType); ok {
							if argProto, argIsProto := argType.(*types.ProtocolType); !argIsProto || !types.Equals(proto, argProto) {
								argOp = &InstOperand{Inst: &InterfaceCast{Val: argOp, Type: proto}}
							}
							lease := types.LeaseRead
							if paramIdx < len(funcType.ParamLeases) {
								lease = funcType.ParamLeases[paramIdx]
							}
							if l.shouldPassByPointer(proto, lease) {
								if !types.IsPointerLike(argType) {
									ptrType := &types.PointerType{Base: proto, Leased: true, Kind: lease}
									borrowAddr := &AddressOf{Val: argOp, Type: ptrType}
									argOp = &InstOperand{Inst: borrowAddr}
								}
							}
						} else {
							lease := types.LeaseRead
							if paramIdx < len(funcType.ParamLeases) {
								lease = funcType.ParamLeases[paramIdx]
							}
							baseType := paramType
							if pt, ok := paramType.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
								baseType = pt.Base
								lease = pt.Kind
							}
							if l.shouldPassByPointer(baseType, lease) {
								if !types.IsPointerLike(argType) {
									var ptrType types.NRType
									if _, ok := paramType.(*types.PointerType); ok {
										ptrType = paramType
									} else {
										ptrType = &types.PointerType{Base: paramType, Leased: true, Kind: lease}
									}
									borrowAddr := &AddressOf{Val: argOp, Type: ptrType}
									argOp = &InstOperand{Inst: borrowAddr}
								}
							}
						}
					}
				}
			}
			argsOps = append(argsOps, argOp)
		}

		call := &Call{
			ASTNode:    e,
			FuncSymbol: funcSym,
			FuncName:   funcName,
			Args:       argsOps,
			Type:       t,
		}
		return &InstOperand{Inst: call}

	case *ast.InfixExpression:
		lt := l.getType(e.Left)
		rt := l.getType(e.Right)
		ult := types.UnwrapLease(lt)
		urt := types.UnwrapLease(rt)

		isStrConcat := e.Operator == "+" && (lt != nil && lt.Name() == "str" || rt != nil && rt.Name() == "str")
		isStrEq := (e.Operator == "==" || e.Operator == "!=") && (ult != nil && ult.Name() == "str" && urt != nil && urt.Name() == "str")
		isFuncEq := (e.Operator == "==" || e.Operator == "!=") && (ult != nil && ult.GetKind() == types.KindFunction && urt != nil && urt.GetKind() == types.KindFunction)

		_, isStructL := ult.(*types.StructType)
		_, isStructR := urt.(*types.StructType)
		isStructEq := (e.Operator == "==" || e.Operator == "!=") && isStructL && isStructR && lt.GetKind() != types.KindPointer && rt.GetKind() != types.KindPointer
		isStructOp := e.Operator != "==" && e.Operator != "!=" && isStructL

		if isStrConcat || isStrEq || isFuncEq || isStructEq || isStructOp {
			astExpr := &ASTExpr{ASTNode: e, Type: t}
			return &InstOperand{Inst: astExpr}
		}

		leftOp := l.lowerExpression(e.Left)
		rightOp := l.lowerExpression(e.Right)
		binOp := &BinOp{Left: leftOp, Op: e.Operator, Right: rightOp, Type: t}
		return &InstOperand{Inst: binOp}

	case *ast.IfExpression:
		condOp := l.lowerExpression(e.Condition)
		var tempOp Operand
		if t != nil && t.Name() != "void" {
			tempName := l.makeTempName()
			l.CurrentBlock.AddInst(&Alloca{Name: tempName, Type: t})
			tempOp = &VarOperand{Name: tempName, Type: t}
		}

		var thenBlock *HIRBlock
		if e.Consequence != nil {
			thenBlock = l.lowerStatementsWithDrops(e.Consequence, tempOp)
		} else {
			thenBlock = NewHIRBlock()
		}

		var elseBlock *HIRBlock
		if e.Alternative != nil {
			elseBlock = NewHIRBlock()
			prevBlock := l.CurrentBlock
			l.CurrentBlock = elseBlock
			if block, ok := e.Alternative.(*ast.BlockStatement); ok {
				sub := l.lowerStatementsWithDrops(block, tempOp)
				for _, el := range sub.Elements {
					elseBlock.AddElement(el)
				}
			} else {
				altOp := l.lowerExpression(e.Alternative)
				if tempOp != nil {
					elseBlock.AddInst(&Store{Dest: tempOp, Val: altOp})
				}
			}
			l.CurrentBlock = prevBlock
		}

		hirIf := &HIRIf{
			Condition: condOp,
			Then:      thenBlock,
			Else:      elseBlock,
		}
		l.CurrentBlock.AddElement(hirIf)
		if tempOp != nil {
			return tempOp
		}
		return &LiteralOperand{Value: "none", Type: types.ErrorType}
	case *ast.RangeExpression:
		astExpr := &ASTExpr{ASTNode: e, Type: t}
		return &InstOperand{Inst: astExpr}

	case *ast.AllocExpression:
		var valOp Operand
		isArray := false
		if pt, ok := t.(*types.PointerType); ok && pt.IsArray {
			isArray = true
			var countExpr ast.Expression
			if ie, ok := e.Value.(*ast.IndexExpression); ok {
				countExpr = ie.Indices[0]
			} else if pe, ok := e.Value.(*ast.PrefixExpression); ok {
				if ie, ok := pe.Right.(*ast.IndexExpression); ok {
					countExpr = ie.Indices[0]
				}
			}
			if countExpr != nil {
				valOp = l.lowerExpression(countExpr)
			} else {
				valOp = &LiteralOperand{Value: "0", Type: types.ErrorType}
			}
		} else {
			valOp = l.lowerExpression(e.Value)
		}

		alloc := &Alloc{
			Type:    t,
			Val:     valOp,
			IsArray: isArray,
			PosFile: e.Pos().Filename,
			PosLine: e.Pos().Line,
		}
		return &InstOperand{Inst: alloc}

	case *ast.TryExpression:
		valOp := l.lowerExpression(e.Value)
		tryInst := &Try{
			ASTNode: e,
			Val:     valOp,
			Type:    t,
		}
		return &InstOperand{Inst: tryInst}

	case *ast.AssignmentStatement:
		destOp := l.lowerExpression(e.Left)
		valOp := l.lowerExpression(e.Value)
		store := &Store{
			Dest:         destOp,
			Val:          valOp,
			NeedsOldDrop: l.Solver != nil && l.Solver.AssignDrops[e],
		}
		return &InstOperand{Inst: store}

	case *ast.SpawnExpression:
		var spawnCall *Call
		if e.Call != nil {
			callOp := l.lowerExpression(e.Call)
			if instOp, ok := callOp.(*InstOperand); ok {
				if call, ok2 := instOp.Inst.(*Call); ok2 {
					spawnCall = call
				}
			}
		}
		var monitorChan Operand
		if e.MonitorChannel != nil {
			monitorChan = l.lowerExpression(e.MonitorChannel)
		}
		spawn := &Spawn{
			Call:           spawnCall,
			MonitorChannel: monitorChan,
			Type:           t,
		}
		return &InstOperand{Inst: spawn}

	case *ast.SendExpression:
		chanOp := l.lowerExpression(e.Left)
		valOp := l.lowerExpression(e.Right)
		send := &ChanSend{
			Chan: chanOp,
			Val:  valOp,
		}
		return &InstOperand{Inst: send}

	case *ast.ReceiveExpression:
		chanOp := l.lowerExpression(e.Value)
		recv := &ChanRecv{
			Chan: chanOp,
			Type: t,
		}
		return &InstOperand{Inst: recv}

	case *ast.LambdaExpression:
		tempName := fmt.Sprintf("nr_lambda_tmp_%d", l.lambdaCounter)
		l.lambdaCounter++
		sym := &semantic.Symbol{Name: tempName, Type: t, Kind: semantic.SymVar}
		l.LambdaTemps[e] = sym
		
		l.CurrentBlock.AddInst(&Alloca{Symbol: sym, Type: t})
		l.CurrentBlock.AddInst(&Store{
			Dest: &VarOperand{Name: tempName, Type: t, Symbol: sym},
			Val:  &InstOperand{Inst: &ASTExpr{ASTNode: e, Type: t}},
		})

		lf := l.lowerLambdaFunction(e)
		l.lambdaFuncs = append(l.lambdaFuncs, lf)

		return &VarOperand{Name: tempName, Type: t, Symbol: sym}

	default:
		astExpr := &ASTExpr{ASTNode: expr, Type: t}
		return &InstOperand{Inst: astExpr}
	}
}

func (l *Lowerer) lowerStatementsWithDrops(block *ast.BlockStatement, tempOp Operand) *HIRBlock {
	prevBlock := l.CurrentBlock
	prevASTBlock := l.currentASTBlock
	prevASTStmtIndex := l.currentASTStmtIndex

	hirBlock := NewHIRBlock()
	l.CurrentBlock = hirBlock
	l.currentASTBlock = block

	// 1. Emit initial drops at index 0
	if l.Solver != nil && l.Solver.Drops[block] != nil {
		drops := l.Solver.Drops[block][0]
		for _, d := range drops {
			var field ast.Expression
			if d.Field != nil {
				field = d.Field
			}
			var index ast.Expression
			if d.Index != nil {
				index = d.Index
			}
			l.emitDrop(d, field, index)
		}
	}

	// 2. Emit statements and their drops
	for i, stmt := range block.Statements {
		l.currentASTStmtIndex = i

		if l.Solver != nil && l.Solver.PreDrops[block] != nil {
			drops := l.Solver.PreDrops[block][i]
			for _, d := range drops {
				var field ast.Expression
				if d.Field != nil {
					field = d.Field
				}
				var index ast.Expression
				if d.Index != nil {
					index = d.Index
				}
				l.emitDrop(d, field, index)
			}
		}

		// Handle the tempOp assignment if it's the last statement of the block
		if tempOp != nil && i == len(block.Statements)-1 {
			if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
				valOp := l.lowerExpression(exprStmt.Expression)
				l.CurrentBlock.AddInst(&Store{Dest: tempOp, Val: valOp})

				// Post-drops after this statement
				if l.Solver != nil && l.Solver.Drops[block] != nil {
					drops := l.Solver.Drops[block][i+1]
					for _, d := range drops {
						var field ast.Expression
						if d.Field != nil {
							field = d.Field
						}
						var index ast.Expression
						if d.Index != nil {
							index = d.Index
						}
						l.emitDrop(d, field, index)
					}
				}
				continue
			}
		}

		l.lowerStatement(stmt)

		if l.Solver != nil && l.Solver.Drops[block] != nil {
			drops := l.Solver.Drops[block][i+1]
			for _, d := range drops {
				var field ast.Expression
				if d.Field != nil {
					field = d.Field
				}
				var index ast.Expression
				if d.Index != nil {
					index = d.Index
				}
				l.emitDrop(d, field, index)
			}
		}
	}

	// 3. Emit end-of-block pre-drops
	if l.Solver != nil && l.Solver.PreDrops[block] != nil {
		drops := l.Solver.PreDrops[block][len(block.Statements)]
		for _, d := range drops {
			var field ast.Expression
			if d.Field != nil {
				field = d.Field
			}
			var index ast.Expression
			if d.Index != nil {
				index = d.Index
			}
			l.emitDrop(d, field, index)
		}
	}

	l.CurrentBlock = prevBlock
	l.currentASTBlock = prevASTBlock
	l.currentASTStmtIndex = prevASTStmtIndex
	return hirBlock
}

func (l *Lowerer) shouldPassByPointer(t types.NRType, lease types.LeaseKind) bool {
	if t == nil {
		return false
	}
	if lease == types.LeaseWrite {
		return true
	}
	// Check if it's already a pointer in C
	if pt, ok := t.(*types.PointerType); ok && !pt.Leased {
		return false
	}
	if t.GetKind() == types.KindList || t.GetKind() == types.KindChan || t.GetKind() == types.KindMap || t.Name() == "str" || t.Name() == "ptr" {
		return false
	}
	if lease == types.LeaseMove || lease == types.LeaseRead {
		if _, ok := t.(*types.StructType); ok || t.GetKind() == types.KindSum || t.GetKind() == types.KindProtocol {
			return true
		}
	}
	return false
}

func (l *Lowerer) getType(expr ast.Expression) types.NRType {
	if expr == nil {
		return types.ErrorType
	}
	if ae, ok := expr.(*ast.ArgumentsExpression); ok {
		return l.getType(ae.Value)
	}
	t := l.SemanticInfo.Types[expr]
	if t == nil {
		if ident, ok := expr.(*ast.Identifier); ok {
			if sym := l.SemanticInfo.Uses[ident]; sym != nil {
				t = sym.Type
			} else if sym := l.SemanticInfo.Defs[ident]; sym != nil {
				t = sym.Type
			}
		}
	}
	if t == nil {
		return types.ErrorType
	}
	return t
}

func (l *Lowerer) isVariantConstructor(expr ast.Expression) (*types.SumType, string) {
	if ae, ok := expr.(*ast.ArgumentsExpression); ok {
		return l.isVariantConstructor(ae.Value)
	}
	t := l.SemanticInfo.Types[expr]
	if t == nil {
		if ident, ok := expr.(*ast.Identifier); ok {
			if sym := l.SemanticInfo.Uses[ident]; sym != nil {
				t = sym.Type
			} else if sym := l.SemanticInfo.Defs[ident]; sym != nil {
				t = sym.Type
			}
		}
	}
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

func isUnsupportedType(t types.NRType) bool {
	return false
}

func (l *Lowerer) shouldSkipHIR(sym *semantic.Symbol, fn *ast.FunctionStatement) bool {
	if fn.IsGenericTemplate || len(fn.TypeParameters) > 0 {
		return true
	}
	return false
}

func (l *Lowerer) lowerLambdaFunction(e *ast.LambdaExpression) *Function {
	t := l.getType(e)
	ft, _ := t.(*types.FunctionType)

	h := sha256.New()
	h.Write([]byte(filepath.Base(e.Pos().Filename)))
	h.Write([]byte(fmt.Sprintf(":%d:%d", e.Pos().Line, e.Pos().Column)))
	if ft != nil {
		h.Write([]byte(fmt.Sprintf("%v", ft)))
	}
	if l.CurrentFunc != nil {
		h.Write([]byte(l.CurrentFunc.Name))
	}
	hashVal := hex.EncodeToString(h.Sum(nil))[:8]
	fnName := fmt.Sprintf("nr_lambda_%s", hashVal)

	sym := &semantic.Symbol{
		Name: fnName,
		Type: ft,
	}
	if l.CurrentFunc != nil {
		sym.DefScope = l.CurrentFunc.DefScope
	} else {
		sym.DefScope = l.SemanticInfo.Scopes[e]
	}

	prevDefers := l.activeDefers
	l.activeDefers = nil
	defer func() {
		l.activeDefers = prevDefers
	}()

	hirFn := &Function{
		Name:       fnName,
		FuncSymbol: sym,
		Params:     []string{},
		Body:       NewHIRBlock(),
		LambdaExpr: e,
	}

	for _, p := range e.Parameters {
		hirFn.Params = append(hirFn.Params, p.Name.Value)
	}

	if e.Body != nil {
		hirFn.Body = l.lowerBlock(e.Body)
	}

	if len(l.activeDefers) > 0 {
		hasRet := false
		if len(hirFn.Body.Elements) > 0 {
			last := hirFn.Body.Elements[len(hirFn.Body.Elements)-1]
			if instEl, ok := last.(*InstElement); ok {
				if _, isRet := instEl.Inst.(*Ret); isRet {
					hasRet = true
				}
			}
		}
		if !hasRet {
			prevBlock := l.CurrentBlock
			l.CurrentBlock = hirFn.Body
			for i := len(l.activeDefers) - 1; i >= 0; i-- {
				l.lowerExpression(l.activeDefers[i])
			}
			l.CurrentBlock = prevBlock
		}
	}

	return hirFn
}
