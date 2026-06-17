package ast

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/nora-language/nora/pkg/token"
)

// Node is the base interface for everything in the AST
type Node interface {
	TokenLiteral() string
	String() string
	Pos() token.Position
}

// IsNil checks if an interface is truly nil or contains a nil pointer.
func IsNil(n Node) bool {
	if n == nil {
		return true
	}
	// Fast path for common pointer types to avoid reflection
	switch v := n.(type) {
	case *Identifier:
		return v == nil
	case *IntegerLiteral:
		return v == nil
	case *StringLiteral:
		return v == nil
	case *FunctionStatement:
		return v == nil
	case *BlockStatement:
		return v == nil
	case *VarStatement:
		return v == nil
	case *ExpressionStatement:
		return v == nil
	case *CallExpression:
		return v == nil
	case *InfixExpression:
		return v == nil
	case *PrefixExpression:
		return v == nil
	case *IfExpression:
		return v == nil
	case *MatchExpression:
		return v == nil
	case *ReturnStatement:
		return v == nil
	case *TypeStatement:
		return v == nil
	case *File:
		return v == nil
	case *Program:
		return v == nil
	}
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// Statement represents syntax like 'let x = 5' or 'return'
type Statement interface {
	Node
	statementNode()
}

// Expression represents syntax like '5 + 5' or 'fn()'
type Expression interface {
	Node
	expressionNode()
}

// TypeNode is a marker interface for nodes that represent types.
// Valid examples: *Identifier ("i32"), *ArrayType ("[i32]"), *MapType
type TypeNode interface {
	Node
	Expression       // Allow types to appear in expression contexts (e.g. make(chan[T]))
	MarkerTypeNode() // Marker method
}

type GroupedExpression struct {
	Token      token.Token // The '(' token
	Expression Expression
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupedExpression) String() string       { return "(" + ge.Expression.String() + ")" }
func (ge *GroupedExpression) Pos() token.Position  { return ge.Token.Position }
func (ge *GroupedExpression) MarkerTypeNode()      {}

// Ensure Identifier can be used as a TypeNode (e.g. "User", "i32")
func (i *Identifier) MarkerTypeNode() {}

type LeaseKind int

const (
	LeaseRead  LeaseKind = iota
	LeaseWrite           // #
	LeaseMove            // @
)

// ----------------------------------------------------------------------------
// 1. The Program (The Project Root)
// ----------------------------------------------------------------------------

// Program represents the entire build (all files combined).
// This is what you pass to the Semantic Analyzer.
type Program struct {
	Files []*File
}

func (p *Program) TokenLiteral() string {
	if len(p.Files) > 0 {
		return p.Files[0].TokenLiteral()
	}
	return ""
}

func (p *Program) Pos() token.Position {
	if len(p.Files) > 0 {
		return p.Files[0].Pos()
	}
	return token.Position{}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, f := range p.Files {
		out.WriteString("File: " + f.Name + "\n")
		out.WriteString("----------------\n")
		out.WriteString(f.String())
		out.WriteString("\n")
	}
	return out.String()
}

// ----------------------------------------------------------------------------
// 2. The File (Formerly "Program")
// ----------------------------------------------------------------------------

// File represents a single parsed source file (e.g., "main.nr").
// Your Parser will return *this* struct for each file it parses.
type File struct {
	Name        string      // e.g., "main.nr"
	Statements  []Statement // The top-level code in this file
	Comments    []*CommentGroup
	BlankLines  []int             // Lines that are blank in the source file
	StmtEndLine map[Statement]int // Maps each statement to its original end line
}

func (f *File) TokenLiteral() string {
	if len(f.Statements) > 0 {
		return f.Statements[0].TokenLiteral()
	}
	return ""
}

// Pos returns the position of the first statement in the file
func (f *File) Pos() token.Position {
	if len(f.Statements) > 0 {
		return f.Statements[0].Pos()
	}
	return token.Position{Filename: f.Name, Line: 1, Column: 1}
}

func (f *File) String() string {
	var out bytes.Buffer
	for _, s := range f.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

type RuneLiteral struct {
	Token token.Token
	Value int32
}

func (rl *RuneLiteral) expressionNode()      {}
func (rl *RuneLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RuneLiteral) String() string       { return rl.Token.Literal }
func (rl *RuneLiteral) Pos() token.Position  { return rl.Token.Position }

type TypeParameter struct {
	Token      token.Token // The identifier token
	Name       *Identifier
	Constraint TypeNode // Optional: e.g. Printable
}

func (tp *TypeParameter) TokenLiteral() string { return tp.Token.Literal }
func (tp *TypeParameter) String() string {
	if tp.Constraint != nil {
		return tp.Name.String() + ": " + tp.Constraint.String()
	}
	return tp.Name.String()
}
func (tp *TypeParameter) Pos() token.Position { return tp.Token.Position }
func (tp *TypeParameter) expressionNode()     {}

// Attribute represents a custom annotation like [builtin("println")]
type Attribute struct {
	Token token.Token // The name token
	Name  string      // "builtin"
	Args  []string    // ["println"]
}

func (a *Attribute) TokenLiteral() string { return a.Token.Literal }
func (a *Attribute) String() string {
	if len(a.Args) > 0 {
		return "[" + a.Name + "(\"" + strings.Join(a.Args, "\", \"") + "\")]"
	}
	return "[" + a.Name + "]"
}
func (a *Attribute) Pos() token.Position { return a.Token.Position }

func GetAttribute(attributes []Attribute, name string) *Attribute {
	for _, attr := range attributes {
		if attr.Name == name {
			return &attr
		}
	}
	return nil
}
