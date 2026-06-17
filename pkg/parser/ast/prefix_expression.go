package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

type PrefixExpression struct {
	Token    token.Token // !, -
	Operator string
	Right    Expression
	Type     types.NRType
}

func (pe *PrefixExpression) expressionNode()               {}
func (pe *PrefixExpression) TokenLiteral() string          { return pe.Token.Literal }
func (pe *PrefixExpression) Pos() token.Position           { return pe.Token.Position }
func (pe *PrefixExpression) MarkerTypeNode()               {}
func (pe *PrefixExpression) GetResolvedType() types.NRType { return pe.Type }

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	if pe.Right != nil && !IsNil(pe.Right) {
		out.WriteString(pe.Right.String())
	}
	out.WriteString(")")
	return out.String()
}
