package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
)

type ArgumentsExpression struct {
	Token token.Token // The identifier token
	Name  *Identifier
	Value Expression
	Lease LeaseKind
}

func (p *ArgumentsExpression) expressionNode()      {}
func (p *ArgumentsExpression) TokenLiteral() string { return p.Token.Literal }
func (p *ArgumentsExpression) Pos() token.Position  { return p.Token.Position }

func (ae *ArgumentsExpression) String() string {
	if ae.Value != nil {
		return ae.Value.String()
	}
	return ""
}
