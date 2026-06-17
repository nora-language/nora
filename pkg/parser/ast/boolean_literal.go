package ast

import (
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
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
