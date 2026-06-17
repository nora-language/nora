package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Expression // Optional 'else'

	Type types.NRType
}

func (ie *IfExpression) expressionNode()               {}
func (ie *IfExpression) TokenLiteral() string          { return ie.Token.Literal }
func (ie *IfExpression) Pos() token.Position           { return ie.Token.Position }
func (ie *IfExpression) GetResolvedType() types.NRType { return ie.Type }

func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString(" else ")
		// This will now call either BlockStatement.String()
		// OR IfExpression.String() recursively
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}
