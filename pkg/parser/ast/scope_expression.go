package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

type ScopeExpression struct {
	Token token.Token // The 'scope' token
	Body  *BlockStatement
}

func (pe *ScopeExpression) expressionNode()      {}
func (pe *ScopeExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *ScopeExpression) Pos() token.Position  { return pe.Token.Position }

func (pe *ScopeExpression) String() string {
	var out bytes.Buffer

	out.WriteString("scope ")
	out.WriteString(pe.Body.String())

	return out.String()
}
