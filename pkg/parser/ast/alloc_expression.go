package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type AllocExpression struct {
	Token token.Token // The 'alloc' token
	Value Expression  // The value to allocate (e.g., StructLiteral)

	Type types.NRType
}

func (ae *AllocExpression) expressionNode()      {}
func (ae *AllocExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AllocExpression) Pos() token.Position  { return ae.Token.Position }

func (ae *AllocExpression) GetResolvedType() types.NRType { return ae.Type }

func (ae *AllocExpression) String() string {
	var out bytes.Buffer
	out.WriteString("alloc ")
	out.WriteString(ae.Value.String())
	return out.String()
}
