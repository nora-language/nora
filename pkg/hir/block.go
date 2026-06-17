package hir

import (
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type BlockElementKind int

const (
	BlockInstruction BlockElementKind = iota
	BlockIf
	BlockLoop
	BlockIteratorLoop
	BlockSelect
)

type BlockElement interface {
	GetElementKind() BlockElementKind
	String() string
}

// Wrap Instruction as a BlockElement
type InstElement struct {
	Inst Instruction
}

func (i *InstElement) GetElementKind() BlockElementKind { return BlockInstruction }
func (i *InstElement) String() string                   { return i.Inst.String() }

// HIRBlock is a sequence of Instructions or control flow elements
type HIRBlock struct {
	Elements []BlockElement
}

func NewHIRBlock() *HIRBlock {
	return &HIRBlock{Elements: []BlockElement{}}
}

func (b *HIRBlock) AddInst(inst Instruction) {
	b.Elements = append(b.Elements, &InstElement{Inst: inst})
}

func (b *HIRBlock) AddElement(el BlockElement) {
	b.Elements = append(b.Elements, el)
}

func (b *HIRBlock) String() string {
	var sb strings.Builder
	for _, el := range b.Elements {
		sb.WriteString(el.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

// HIRIf represents structured conditional branches
type HIRIf struct {
	Condition Operand
	Then      *HIRBlock
	Else      *HIRBlock // optional
}

func (i *HIRIf) GetElementKind() BlockElementKind { return BlockIf }
func (i *HIRIf) String() string {
	var sb strings.Builder
	sb.WriteString("if " + i.Condition.String() + " {\n")
	sb.WriteString(i.Then.String())
	if i.Else != nil && len(i.Else.Elements) > 0 {
		sb.WriteString("} else {\n")
		sb.WriteString(i.Else.String())
	}
	sb.WriteString("}")
	return sb.String()
}

// HIRLoop represents structured loops (while / for loops)
type HIRLoop struct {
	Init      *HIRBlock
	Condition Operand
	Step      *HIRBlock
	Body      *HIRBlock
}

func (l *HIRLoop) GetElementKind() BlockElementKind { return BlockLoop }
func (l *HIRLoop) String() string {
	var sb strings.Builder
	sb.WriteString("loop {\n")
	if l.Init != nil {
		sb.WriteString("  init:\n" + l.Init.String())
	}
	if l.Condition != nil {
		sb.WriteString("  cond: " + l.Condition.String() + "\n")
	}
	sb.WriteString("  body:\n" + l.Body.String())
	if l.Step != nil {
		sb.WriteString("  step:\n" + l.Step.String())
	}
	sb.WriteString("}")
	return sb.String()
}

// Function represents a linearized function body
type Function struct {
	Name       string
	FuncSymbol *semantic.Symbol
	Params     []string
	Body       *HIRBlock
	LambdaExpr *ast.LambdaExpression // nil for top-level functions
}

// Program represents the entire lowering output
type Program struct {
	Functions []*Function
}

// SelectCase represents a single select branch inside Select
type SelectCase struct {
	IsDefault bool
	IsSend    bool
	Chan      Operand
	Val       Operand      // For send, value; for receive, destination expression (if assignment)
	VarName   string       // For receive variable declaration
	VarType   types.NRType // For receive variable declaration
	Body      *HIRBlock
}

// Select represents a structured channel select statement
type Select struct {
	Cases []SelectCase
}

func (s *Select) GetElementKind() BlockElementKind { return BlockSelect }
func (s *Select) String() string {
	return "select {\n}"
}
