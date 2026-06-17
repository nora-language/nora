package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type TryExpression struct {
	Token token.Token // '?'
	Value Expression
	Type  types.NRType
}

func (te *TryExpression) expressionNode()               {}
func (te *TryExpression) TokenLiteral() string          { return te.Token.Literal }
func (te *TryExpression) Pos() token.Position           { return te.Token.Position }
func (te *TryExpression) GetResolvedType() types.NRType { return te.Type }

// To this (if you want the ? inside the grouped expression):
func (te *TryExpression) String() string {
	return te.Value.String() + "?"
}
