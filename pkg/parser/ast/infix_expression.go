package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression

	Type types.NRType
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) Pos() token.Position {
	if ie.Left != nil {
		return ie.Left.Pos()
	}
	return ie.Token.Position
}
func (ie *InfixExpression) GetResolvedType() types.NRType { return ie.Type }

func (ie *InfixExpression) String() string {
	leftStr := ""
	if ie.Left != nil && !IsNil(ie.Left) {
		leftStr = ie.Left.String()
	}
	rightStr := ""
	if ie.Right != nil && !IsNil(ie.Right) {
		rightStr = ie.Right.String()
	}
	return "(" + leftStr + " " + ie.Operator + " " + rightStr + ")"
}
