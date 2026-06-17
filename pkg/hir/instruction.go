package hir

import (
	"fmt"
	"strings"

	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/types"
)

type InstructionKind int

const (
	InstAlloca InstructionKind = iota
	InstLoad
	InstStore
	InstAddressOf
	InstDeref
	InstCall
	InstRet
	InstFieldAccess
	InstIndexAccess
	InstCast
	InstAssign
	InstExpression
	InstBinOp
	InstUnOp
	InstVariantConstructor
	InstAlloc
	InstTry
	InstASTExpr
	InstDrop
	InstInterfaceCast
	InstInterfaceCall
	InstSpawn
	InstChanSend
	InstChanRecv
	InstLambda
	InstLineInfo
)

type Instruction interface {
	GetInstructionKind() InstructionKind
	GetType() types.NRType
	String() string
}

type Operand interface {
	GetType() types.NRType
	String() string
}

// --- Operands ---

type LiteralOperand struct {
	Value string
	Type  types.NRType
}

func (l *LiteralOperand) GetType() types.NRType { return l.Type }
func (l *LiteralOperand) String() string        { return l.Value }

type VarOperand struct {
	Name   string
	Type   types.NRType
	Symbol *semantic.Symbol
}

func (v *VarOperand) GetType() types.NRType { return v.Type }
func (v *VarOperand) String() string        { return v.Name }

type InstOperand struct {
	Inst Instruction
}

func (i *InstOperand) GetType() types.NRType { return i.Inst.GetType() }
func (i *InstOperand) String() string        { return "(" + i.Inst.String() + ")" }

// --- Instructions ---

// Alloca: stack allocation of a local variable
type Alloca struct {
	Name   string
	Type   types.NRType
	Symbol *semantic.Symbol
}

func (a *Alloca) GetInstructionKind() InstructionKind { return InstAlloca }
func (a *Alloca) GetType() types.NRType               { return a.Type }
func (a *Alloca) String() string                      { return fmt.Sprintf("alloca %s : %s", a.Name, a.Type.Name()) }

// Load: load value from pointer Src
type Load struct {
	Src  Operand
	Type types.NRType
}

func (l *Load) GetInstructionKind() InstructionKind { return InstLoad }
func (l *Load) GetType() types.NRType               { return l.Type }
func (l *Load) String() string                      { return fmt.Sprintf("load %s", l.Src.String()) }

// Store: write value Val to Dest variable or memory location
type Store struct {
	Dest         Operand
	Val          Operand
	NeedsOldDrop bool
}

func (s *Store) GetInstructionKind() InstructionKind { return InstStore }
func (s *Store) GetType() types.NRType               { return nil }
func (s *Store) String() string {
	return fmt.Sprintf("store %s -> %s", s.Val.String(), s.Dest.String())
}

// AddressOf: take address of Val
type AddressOf struct {
	Val      Operand
	Type     types.NRType
	Operator string
}

func (a *AddressOf) GetInstructionKind() InstructionKind { return InstAddressOf }
func (a *AddressOf) GetType() types.NRType               { return a.Type }
func (a *AddressOf) String() string                      { return fmt.Sprintf("addressof %s", a.Val.String()) }

// Deref: prepends '*' in C
type Deref struct {
	Val  Operand
	Type types.NRType
}

func (d *Deref) GetInstructionKind() InstructionKind { return InstDeref }
func (d *Deref) GetType() types.NRType               { return d.Type }
func (d *Deref) String() string                      { return fmt.Sprintf("deref %s", d.Val.String()) }

// Call: call function with Args
type Call struct {
	ASTNode    *ast.CallExpression
	FuncSymbol *semantic.Symbol
	FuncName   string
	Args       []Operand
	Type       types.NRType
}

func (c *Call) GetInstructionKind() InstructionKind { return InstCall }
func (c *Call) GetType() types.NRType               { return c.Type }
func (c *Call) String() string {
	argsStr := make([]string, len(c.Args))
	for i, arg := range c.Args {
		argsStr[i] = arg.String()
	}
	name := c.FuncName
	if c.FuncSymbol != nil {
		name = c.FuncSymbol.Name
	}
	return fmt.Sprintf("call %s(%s)", name, strings.Join(argsStr, ", "))
}

// Ret: return Val
type Ret struct {
	Val Operand
}

func (r *Ret) GetInstructionKind() InstructionKind { return InstRet }
func (r *Ret) GetType() types.NRType               { return nil }
func (r *Ret) String() string                      { return fmt.Sprintf("ret %s", r.Val.String()) }

// FieldAccess: access struct field
type FieldAccess struct {
	Base      Operand
	FieldName string
	Type      types.NRType
}

func (f *FieldAccess) GetInstructionKind() InstructionKind { return InstFieldAccess }
func (f *FieldAccess) GetType() types.NRType               { return f.Type }
func (f *FieldAccess) String() string                      { return fmt.Sprintf("%s.%s", f.Base.String(), f.FieldName) }

// IndexAccess: access index of array
type IndexAccess struct {
	Base          Operand
	Index         Operand
	Type          types.NRType
	NoBoundsCheck bool
}

func (i *IndexAccess) GetInstructionKind() InstructionKind { return InstIndexAccess }
func (i *IndexAccess) GetType() types.NRType               { return i.Type }
func (i *IndexAccess) String() string {
	return fmt.Sprintf("%s[%s]", i.Base.String(), i.Index.String())
}

// Cast: explicitly convert type of Val
type Cast struct {
	Val  Operand
	Type types.NRType
}

func (c *Cast) GetInstructionKind() InstructionKind { return InstCast }
func (c *Cast) GetType() types.NRType               { return c.Type }
func (c *Cast) String() string                      { return fmt.Sprintf("cast %s to %s", c.Val.String(), c.Type.Name()) }

// Assign: write value Val to Dest variable
type Assign struct {
	Dest Operand
	Val  Operand
}

func (a *Assign) GetInstructionKind() InstructionKind { return InstAssign }
func (a *Assign) GetType() types.NRType               { return nil }
func (a *Assign) String() string {
	return fmt.Sprintf("assign %s = %s", a.Dest.String(), a.Val.String())
}

// Expression: a raw custom expression fallback
type Expression struct {
	Expr string
	Type types.NRType
}

func (e *Expression) GetInstructionKind() InstructionKind { return InstExpression }
func (e *Expression) GetType() types.NRType               { return e.Type }
func (e *Expression) String() string                      { return e.Expr }

type BinOp struct {
	Left  Operand
	Op    string
	Right Operand
	Type  types.NRType
}

func (b *BinOp) GetInstructionKind() InstructionKind { return InstBinOp }
func (b *BinOp) GetType() types.NRType               { return b.Type }
func (b *BinOp) String() string {
	return fmt.Sprintf("%s %s %s", b.Left.String(), b.Op, b.Right.String())
}

type UnOp struct {
	Op   string
	Val  Operand
	Type types.NRType
}

func (u *UnOp) GetInstructionKind() InstructionKind { return InstUnOp }
func (u *UnOp) GetType() types.NRType               { return u.Type }
func (u *UnOp) String() string                      { return fmt.Sprintf("%s%s", u.Op, u.Val.String()) }

type VariantConstructor struct {
	SumType     *types.SumType
	VariantName string
	Args        []Operand
	Type        types.NRType
}

func (vc *VariantConstructor) GetInstructionKind() InstructionKind { return InstVariantConstructor }
func (vc *VariantConstructor) GetType() types.NRType               { return vc.Type }
func (vc *VariantConstructor) String() string {
	argsStr := make([]string, len(vc.Args))
	for i, arg := range vc.Args {
		argsStr[i] = arg.String()
	}
	return fmt.Sprintf("variant %s::%s(%s)", vc.SumType.Name(), vc.VariantName, strings.Join(argsStr, ", "))
}

type Alloc struct {
	Type    types.NRType
	Val     Operand
	IsArray bool
	PosFile string
	PosLine int
}

func (a *Alloc) GetInstructionKind() InstructionKind { return InstAlloc }
func (a *Alloc) GetType() types.NRType               { return a.Type }
func (a *Alloc) String() string {
	if a.IsArray {
		return fmt.Sprintf("alloc %s[%s]", a.Type.Name(), a.Val.String())
	}
	return fmt.Sprintf("alloc %s(%s)", a.Type.Name(), a.Val.String())
}

type Try struct {
	ASTNode *ast.TryExpression
	Val     Operand
	Type    types.NRType
}

func (t *Try) GetInstructionKind() InstructionKind { return InstTry }
func (t *Try) GetType() types.NRType               { return t.Type }
func (t *Try) String() string                      { return fmt.Sprintf("try %s", t.Val.String()) }

type ASTExpr struct {
	ASTNode ast.Expression
	Type    types.NRType
}

func (a *ASTExpr) GetInstructionKind() InstructionKind { return InstASTExpr }
func (a *ASTExpr) GetType() types.NRType               { return a.Type }
func (a *ASTExpr) String() string                      { return a.ASTNode.String() }

type Drop struct {
	Symbol *semantic.Symbol
	Field  ast.Expression
	Index  ast.Expression
}

func (d *Drop) GetInstructionKind() InstructionKind { return InstDrop }
func (d *Drop) GetType() types.NRType               { return nil }
func (d *Drop) String() string {
	if d.Symbol != nil {
		return "drop " + d.Symbol.Name
	}
	if d.Field != nil {
		return "drop field " + d.Field.String()
	}
	if d.Index != nil {
		return "drop index " + d.Index.String()
	}
	return "drop unknown"
}

// InterfaceCast: wrap a concrete value with an interface/protocol vtable
type InterfaceCast struct {
	Val  Operand
	Type *types.ProtocolType
}

func (ic *InterfaceCast) GetInstructionKind() InstructionKind { return InstInterfaceCast }
func (ic *InterfaceCast) GetType() types.NRType               { return ic.Type }
func (ic *InterfaceCast) String() string {
	return fmt.Sprintf("interface_cast %s to %s", ic.Val.String(), ic.Type.Name())
}

// InterfaceCall: call a method on a protocol/interface value
type InterfaceCall struct {
	Base       Operand
	MethodName string
	Args       []Operand
	Type       types.NRType
}

func (ic *InterfaceCall) GetInstructionKind() InstructionKind { return InstInterfaceCall }
func (ic *InterfaceCall) GetType() types.NRType               { return ic.Type }
func (ic *InterfaceCall) String() string {
	argsStr := make([]string, len(ic.Args))
	for i, arg := range ic.Args {
		argsStr[i] = arg.String()
	}
	return fmt.Sprintf("interface_call %s.%s(%s)", ic.Base.String(), ic.MethodName, strings.Join(argsStr, ", "))
}

// Spawn: schedule a function call inside a green-thread fiber
type Spawn struct {
	Call           *Call
	MonitorChannel Operand // Optional channel for panic monitoring
	Type           types.NRType
}

func (s *Spawn) GetInstructionKind() InstructionKind { return InstSpawn }
func (s *Spawn) GetType() types.NRType               { return s.Type }
func (s *Spawn) String() string                      { return fmt.Sprintf("spawn %s", s.Call.String()) }

// ChanSend: send value to channel
type ChanSend struct {
	Chan Operand
	Val  Operand
}

func (cs *ChanSend) GetInstructionKind() InstructionKind { return InstChanSend }
func (cs *ChanSend) GetType() types.NRType               { return nil }
func (cs *ChanSend) String() string {
	return fmt.Sprintf("send %s <- %s", cs.Chan.String(), cs.Val.String())
}

// ChanRecv: receive value from channel
type ChanRecv struct {
	Chan Operand
	Type types.NRType
}

func (cr *ChanRecv) GetInstructionKind() InstructionKind { return InstChanRecv }
func (cr *ChanRecv) GetType() types.NRType               { return cr.Type }
func (cr *ChanRecv) String() string                      { return fmt.Sprintf("recv <- %s", cr.Chan.String()) }

// Lambda: representation of an inline lambda closure
type Lambda struct {
	FuncName string
	ASTNode  *ast.LambdaExpression
	Type     types.NRType
}

func (l *Lambda) GetInstructionKind() InstructionKind { return InstLambda }
func (l *Lambda) GetType() types.NRType               { return l.Type }
func (l *Lambda) String() string                      { return fmt.Sprintf("lambda %s", l.FuncName) }

type IteratorLoop struct {
	Iterator    Operand
	NextMangled string
	NextSymbol  *semantic.Symbol
	ElemName    string
	ElemType    types.NRType
	KeyName     string
	KeyType     types.NRType
	Body        *HIRBlock
}

func (i *IteratorLoop) isElement() {}

func (i *IteratorLoop) GetElementKind() BlockElementKind {
	return BlockIteratorLoop
}

func (i *IteratorLoop) String() string {
	return "for " + i.ElemName + " in " + i.Iterator.String() + " { ... }"
}

// LineInfo: carries original AST node to preserve line numbers in codegen
type LineInfo struct {
	ASTNode ast.Node
}

func (l *LineInfo) GetInstructionKind() InstructionKind { return InstLineInfo }
func (l *LineInfo) GetType() types.NRType               { return nil }
func (l *LineInfo) String() string {
	if l.ASTNode != nil {
		pos := l.ASTNode.Pos()
		return fmt.Sprintf("lineinfo %d", pos.Line)
	}
	return "lineinfo unknown"
}
