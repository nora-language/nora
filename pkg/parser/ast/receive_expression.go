package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
)

// <-ch
type ReceiveExpression struct {
	Token token.Token // '<-'
	Value Expression  // ch
}

func (re *ReceiveExpression) expressionNode()      {}
func (re *ReceiveExpression) TokenLiteral() string { return re.Token.Literal }
func (re *ReceiveExpression) Pos() token.Position  { return re.Token.Position }
func (re *ReceiveExpression) String() string       { return "<-" + re.Value.String() }
