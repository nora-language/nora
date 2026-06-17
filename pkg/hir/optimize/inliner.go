package optimize

import (
	"strings"

	"github.com/nora-language/nora/pkg/hir"
	"github.com/nora-language/nora/pkg/semantic"
)

// Optimizer provides a suite of compiler passes that operate on HIR.
type Optimizer struct {
	SemanticInfo *semantic.SemanticInfo
}

// NewOptimizer creates a new HIR optimizer.
func NewOptimizer(sem *semantic.SemanticInfo) *Optimizer {
	return &Optimizer{
		SemanticInfo: sem,
	}
}

// OptimizeProgram runs all registered HIR optimization passes on the program.
func (opt *Optimizer) OptimizeProgram(prog *hir.Program) *hir.Program {
	return opt.runInlinePass(prog)
}

// Inliner performs the inline pass on the HIR program.
type Inliner struct {
	opt      *Optimizer
	hirFuncs map[string]*hir.Function
	cloner   *Cloner
}

func (opt *Optimizer) runInlinePass(prog *hir.Program) *hir.Program {
	inliner := &Inliner{
		opt:      opt,
		hirFuncs: make(map[string]*hir.Function),
		cloner:   NewCloner(),
	}

	for _, hf := range prog.Functions {
		name := hf.Name
		if hf.FuncSymbol != nil {
			name = hf.FuncSymbol.Name
		}
		inliner.hirFuncs[name] = hf
	}

	for _, hf := range prog.Functions {
		inliner.inlineFunction(hf)
	}

	return prog
}

func (inl *Inliner) inlineFunction(hf *hir.Function) {
	if hf.Body != nil {
		hf.Body = inl.processBlock(hf.Body)
	}
}

func (inl *Inliner) processBlock(block *hir.HIRBlock) *hir.HIRBlock {
	if block == nil {
		return nil
	}

	newBlock := hir.NewHIRBlock()

	for _, el := range block.Elements {
		// First recursively process structural blocks
		switch e := el.(type) {
		case *hir.HIRIf:
			e.Condition = inl.processOperand(e.Condition, newBlock)
			e.Then = inl.processBlock(e.Then)
			e.Else = inl.processBlock(e.Else)
			newBlock.AddElement(e)
		case *hir.HIRLoop:
			if e.Init != nil {
				e.Init = inl.processBlock(e.Init)
			}
			e.Condition = inl.processOperand(e.Condition, newBlock)
			if e.Step != nil {
				e.Step = inl.processBlock(e.Step)
			}
			e.Body = inl.processBlock(e.Body)
			newBlock.AddElement(e)
		case *hir.Select:
			for i, sc := range e.Cases {
				sc.Chan = inl.processOperand(sc.Chan, newBlock)
				sc.Val = inl.processOperand(sc.Val, newBlock)
				sc.Body = inl.processBlock(sc.Body)
				e.Cases[i] = sc
			}
			newBlock.AddElement(e)
		case *hir.InstElement:
			newInst := inl.processInstruction(e.Inst, newBlock)
			if newInst != nil {
				newBlock.AddInst(newInst)
			}
		default:
			newBlock.AddElement(el)
		}
	}

	return newBlock
}

func (inl *Inliner) processOperand(op hir.Operand, targetBlock *hir.HIRBlock) hir.Operand {
	if op == nil {
		return nil
	}
	if instOp, ok := op.(*hir.InstOperand); ok {
		newInst := inl.processInstruction(instOp.Inst, targetBlock)
		// If inlined, it might return a VarOperand for the return value
		if varOp, isVar := newInst.(*hir.Expression); isVar && strings.HasPrefix(varOp.Expr, "_inline_ret_") {
			return &hir.VarOperand{Name: varOp.Expr, Type: varOp.Type}
		}
		if newInst == nil {
			return nil
		}
		return &hir.InstOperand{Inst: newInst}
	}
	return op
}

func (inl *Inliner) processInstruction(inst hir.Instruction, targetBlock *hir.HIRBlock) hir.Instruction {
	if inst == nil {
		return nil
	}

	// Process operands recursively
	switch i := inst.(type) {
	case *hir.Assign:
		i.Val = inl.processOperand(i.Val, targetBlock)
		// If it was a call that got inlined
		if instOp, ok := i.Val.(*hir.InstOperand); ok {
			if expr, ok := instOp.Inst.(*hir.Expression); ok && strings.HasPrefix(expr.Expr, "_inline_ret_") {
				i.Val = &hir.VarOperand{Name: expr.Expr, Type: i.Dest.GetType()}
			}
		}
		return i
	case *hir.Store:
		i.Val = inl.processOperand(i.Val, targetBlock)
		if instOp, ok := i.Val.(*hir.InstOperand); ok {
			if expr, ok := instOp.Inst.(*hir.Expression); ok && strings.HasPrefix(expr.Expr, "_inline_ret_") {
				i.Val = &hir.VarOperand{Name: expr.Expr, Type: instOp.GetType()} // approximate
			}
		}
		return i
	case *hir.Call:
		// Check if it's an inline call
		targetName := i.FuncName
		if i.FuncSymbol != nil {
			targetName = i.FuncSymbol.Name
		}

		targetFunc, exists := inl.hirFuncs[targetName]
		if exists && targetFunc.FuncSymbol != nil && targetFunc.FuncSymbol.IsInline {
			// INLINE THIS FUNCTION!
			retVar := inl.cloner.NextTemp()

			// 1. Map parameters to arguments
			for idx, paramName := range targetFunc.Params {
				argOp := i.Args[idx]
				// Create a temporary variable for the argument
				tempArgName := inl.cloner.NextTemp()
				targetBlock.AddInst(&hir.Alloca{
					Name: tempArgName,
					Type: argOp.GetType(),
				})
				targetBlock.AddInst(&hir.Assign{
					Dest: &hir.VarOperand{Name: tempArgName, Type: argOp.GetType()},
					Val:  argOp,
				})
				inl.cloner.varMap[paramName] = tempArgName
			}

			// Allocate return variable
			if i.Type != nil {
				targetBlock.AddInst(&hir.Alloca{
					Name: retVar,
					Type: i.Type,
				})
			}

			// Clone body
			clonedBody := inl.cloner.CloneBlock(targetFunc.Body)

			// Transform Ret instructions into Assignments to retVar
			inl.transformReturns(clonedBody, retVar)

			// Append cloned body to current block
			for _, el := range clonedBody.Elements {
				targetBlock.AddElement(el)
			}

			// Clean up param mappings
			for _, paramName := range targetFunc.Params {
				delete(inl.cloner.varMap, paramName)
			}

			// The "Call" instruction itself is replaced by reading from retVar
			// Since we might be inside an expression, we return a mock expression
			// that will be converted to a VarOperand by the caller.
			// Using "_inline_ret_" prefix to identify it.
			return &hir.Expression{Expr: retVar, Type: i.Type}
		}

		for idx, arg := range i.Args {
			i.Args[idx] = inl.processOperand(arg, targetBlock)
		}
		return i
	}

	// Default: return unchanged for now (ideally we should walk all Operands deeply)
	return inst
}

func (inl *Inliner) transformReturns(block *hir.HIRBlock, retVar string) {
	if block == nil {
		return
	}
	var newElements []hir.BlockElement
	for _, el := range block.Elements {
		switch e := el.(type) {
		case *hir.InstElement:
			if ret, isRet := e.Inst.(*hir.Ret); isRet {
				if ret.Val != nil {
					newElements = append(newElements, &hir.InstElement{
						Inst: &hir.Assign{
							Dest: &hir.VarOperand{Name: retVar, Type: ret.Val.GetType()},
							Val:  ret.Val,
						},
					})
				}
				// Skip the actual Ret instruction!
			} else {
				newElements = append(newElements, e)
			}
		case *hir.HIRIf:
			inl.transformReturns(e.Then, retVar)
			inl.transformReturns(e.Else, retVar)
			newElements = append(newElements, e)
		case *hir.HIRLoop:
			inl.transformReturns(e.Body, retVar)
			newElements = append(newElements, e)
		default:
			newElements = append(newElements, e)
		}
	}
	block.Elements = newElements
}
