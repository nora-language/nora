package optimize

import (
	"fmt"

	"github.com/DwiYI/Project-Nora/pkg/hir"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
)

type Cloner struct {
	varMap  map[string]string
	tempID  int
}

func NewCloner() *Cloner {
	return &Cloner{
		varMap: make(map[string]string),
	}
}

func (c *Cloner) NextTemp() string {
	c.tempID++
	return fmt.Sprintf("_inline_var_%d", c.tempID)
}

func (c *Cloner) CloneBlock(block *hir.HIRBlock) *hir.HIRBlock {
	if block == nil {
		return nil
	}
	newBlock := hir.NewHIRBlock()
	for _, el := range block.Elements {
		newBlock.AddElement(c.CloneElement(el))
	}
	return newBlock
}

func (c *Cloner) CloneElement(el hir.BlockElement) hir.BlockElement {
	switch e := el.(type) {
	case *hir.InstElement:
		return &hir.InstElement{Inst: c.CloneInstruction(e.Inst)}
	case *hir.HIRIf:
		return &hir.HIRIf{
			Condition: c.CloneOperand(e.Condition),
			Then:      c.CloneBlock(e.Then),
			Else:      c.CloneBlock(e.Else),
		}
	case *hir.HIRLoop:
		return &hir.HIRLoop{
			Init:      c.CloneBlock(e.Init),
			Condition: c.CloneOperand(e.Condition),
			Step:      c.CloneBlock(e.Step),
			Body:      c.CloneBlock(e.Body),
		}
	case *hir.Select:
		var newCases []hir.SelectCase
		for _, sc := range e.Cases {
			newCases = append(newCases, hir.SelectCase{
				IsDefault: sc.IsDefault,
				IsSend:    sc.IsSend,
				Chan:      c.CloneOperand(sc.Chan),
				Val:       c.CloneOperand(sc.Val),
				VarName:   c.mapVarName(sc.VarName),
				VarType:   sc.VarType,
				Body:      c.CloneBlock(sc.Body),
			})
		}
		return &hir.Select{Cases: newCases}
	}
	return el
}

func (c *Cloner) mapVarName(name string) string {
	if name == "" {
		return ""
	}
	if newName, ok := c.varMap[name]; ok {
		return newName
	}
	// It's a global variable or undeclared, return as is
	return name
}

func (c *Cloner) CloneOperand(op hir.Operand) hir.Operand {
	if op == nil {
		return nil
	}
	switch o := op.(type) {
	case *hir.LiteralOperand:
		return o // Literals are immutable
	case *hir.VarOperand:
		newName := c.mapVarName(o.Name)
		var newSym *semantic.Symbol
		if newName == o.Name {
			newSym = o.Symbol
		}
		return &hir.VarOperand{
			Name:   newName,
			Type:   o.Type,
			Symbol: newSym,
		}
	case *hir.InstOperand:
		return &hir.InstOperand{Inst: c.CloneInstruction(o.Inst)}
	}
	return op
}

func (c *Cloner) CloneInstruction(inst hir.Instruction) hir.Instruction {
	if inst == nil {
		return nil
	}
	switch i := inst.(type) {
	case *hir.Alloca:
		newName := c.NextTemp()
		c.varMap[i.Name] = newName
		return &hir.Alloca{
			Name:   newName,
			Type:   i.Type,
			Symbol: nil,
		}
	case *hir.Load:
		return &hir.Load{
			Src:  c.CloneOperand(i.Src),
			Type: i.Type,
		}
	case *hir.Store:
		return &hir.Store{
			Dest: c.CloneOperand(i.Dest),
			Val:  c.CloneOperand(i.Val),
		}
	case *hir.AddressOf:
		return &hir.AddressOf{
			Val:      c.CloneOperand(i.Val),
			Type:     i.Type,
			Operator: i.Operator,
		}
	case *hir.Deref:
		return &hir.Deref{
			Val:  c.CloneOperand(i.Val),
			Type: i.Type,
		}
	case *hir.Call:
		args := make([]hir.Operand, len(i.Args))
		for idx, arg := range i.Args {
			args[idx] = c.CloneOperand(arg)
		}
		return &hir.Call{
			ASTNode:    i.ASTNode,
			FuncSymbol: i.FuncSymbol,
			FuncName:   i.FuncName,
			Args:       args,
			Type:       i.Type,
		}
	case *hir.Ret:
		return &hir.Ret{
			Val: c.CloneOperand(i.Val),
		}
	case *hir.FieldAccess:
		return &hir.FieldAccess{
			Base:      c.CloneOperand(i.Base),
			FieldName: i.FieldName,
			Type:      i.Type,
		}
	case *hir.IndexAccess:
		return &hir.IndexAccess{
			Base:          c.CloneOperand(i.Base),
			Index:         c.CloneOperand(i.Index),
			Type:          i.Type,
			NoBoundsCheck: i.NoBoundsCheck,
		}
	case *hir.Cast:
		return &hir.Cast{
			Val:  c.CloneOperand(i.Val),
			Type: i.Type,
		}
	case *hir.Assign:
		return &hir.Assign{
			Dest: c.CloneOperand(i.Dest),
			Val:  c.CloneOperand(i.Val),
		}
	case *hir.Expression:
		return &hir.Expression{
			Expr: i.Expr,
			Type: i.Type,
		}
	case *hir.BinOp:
		return &hir.BinOp{
			Left:  c.CloneOperand(i.Left),
			Op:    i.Op,
			Right: c.CloneOperand(i.Right),
			Type:  i.Type,
		}
	case *hir.UnOp:
		return &hir.UnOp{
			Op:   i.Op,
			Val:  c.CloneOperand(i.Val),
			Type: i.Type,
		}
	case *hir.VariantConstructor:
		args := make([]hir.Operand, len(i.Args))
		for idx, arg := range i.Args {
			args[idx] = c.CloneOperand(arg)
		}
		return &hir.VariantConstructor{
			SumType:     i.SumType,
			VariantName: i.VariantName,
			Args:        args,
			Type:        i.Type,
		}
	case *hir.Alloc:
		return &hir.Alloc{
			Type:    i.Type,
			Val:     c.CloneOperand(i.Val),
			IsArray: i.IsArray,
			PosFile: i.PosFile,
			PosLine: i.PosLine,
		}
	case *hir.Try:
		return &hir.Try{
			ASTNode: i.ASTNode,
			Val:     c.CloneOperand(i.Val),
			Type:    i.Type,
		}
	case *hir.ASTExpr:
		return &hir.ASTExpr{
			ASTNode: i.ASTNode,
			Type:    i.Type,
		}
	case *hir.Drop:
		return &hir.Drop{
			Symbol: i.Symbol,
			Field:  i.Field,
			Index:  i.Index,
		}
	case *hir.InterfaceCast:
		return &hir.InterfaceCast{
			Val:  c.CloneOperand(i.Val),
			Type: i.Type,
		}
	case *hir.InterfaceCall:
		args := make([]hir.Operand, len(i.Args))
		for idx, arg := range i.Args {
			args[idx] = c.CloneOperand(arg)
		}
		return &hir.InterfaceCall{
			Base:       c.CloneOperand(i.Base),
			MethodName: i.MethodName,
			Args:       args,
			Type:       i.Type,
		}
	case *hir.Spawn:
		newCall := c.CloneInstruction(i.Call).(*hir.Call)
		return &hir.Spawn{
			Call:           newCall,
			MonitorChannel: c.CloneOperand(i.MonitorChannel),
			Type:           i.Type,
		}
	case *hir.ChanSend:
		return &hir.ChanSend{
			Chan: c.CloneOperand(i.Chan),
			Val:  c.CloneOperand(i.Val),
		}
	case *hir.ChanRecv:
		return &hir.ChanRecv{
			Chan: c.CloneOperand(i.Chan),
			Type: i.Type,
		}
	case *hir.Lambda:
		return &hir.Lambda{
			FuncName: i.FuncName,
			ASTNode:  i.ASTNode,
			Type:     i.Type,
		}
	}
	return inst
}
