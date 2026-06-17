package ast

import (
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

type StringLiteral struct {
	Token token.Token
	Value string
	Type  types.NRType
}

func (sl *StringLiteral) expressionNode()               {}
func (sl *StringLiteral) TokenLiteral() string          { return sl.Token.Literal }
func (sl *StringLiteral) Pos() token.Position           { return sl.Token.Position }
func (sl *StringLiteral) String() string                { return `"` + sl.Value + `"` }
func (sl *StringLiteral) GetResolvedType() types.NRType { return sl.Type }
