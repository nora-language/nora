package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
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
