package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
)

// ch <- val
type SendExpression struct {
	Token token.Token // '<-'
	Left  Expression  // ch
	Right Expression  // val
}

func (se *SendExpression) expressionNode()      {}
func (se *SendExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SendExpression) Pos() token.Position  { return se.Token.Position }
func (se *SendExpression) String() string {
	return se.Left.String() + " <- " + se.Right.String()
}
