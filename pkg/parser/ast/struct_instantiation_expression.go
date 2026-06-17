package ast

import "github.com/nora-language/nora/pkg/token"

type StructInstantiation struct {
	Token  token.Token // '{'
	Type   Expression  // Identifier "User"
	Fields map[string]Expression
}

func (si *StructInstantiation) expressionNode()      {}
func (si *StructInstantiation) TokenLiteral() string { return si.Token.Literal }
func (si *StructInstantiation) Pos() token.Position  { return si.Token.Position }
func (si *StructInstantiation) String() string       { return "StructInstantiation" } // Simplify
