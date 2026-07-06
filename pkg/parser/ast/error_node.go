package ast

import "github.com/nora-language/nora/pkg/token"

type ErrorNode struct {
	Token token.Token
}

func (en *ErrorNode) expressionNode()      {}
func (en *ErrorNode) TokenLiteral() string { return en.Token.Literal }
func (en *ErrorNode) String() string       { return "<error>" }
func (en *ErrorNode) Pos() token.Position  { return en.Token.Position }
func (en *ErrorNode) MarkerTypeNode()      {}
