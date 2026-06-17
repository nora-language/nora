package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type Boolean struct {
	Token token.Token
	Value bool

	Type types.NRType
}

func (b *Boolean) expressionNode()               {}
func (b *Boolean) TokenLiteral() string          { return b.Token.Literal }
func (b *Boolean) Pos() token.Position           { return b.Token.Position }
func (b *Boolean) String() string                { return b.Token.Literal }
func (b *Boolean) GetResolvedType() types.NRType { return b.Type }
