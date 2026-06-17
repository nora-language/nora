package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

// none
type NoneLiteral struct {
	Token token.Token
	Type  types.NRType
}

func (nl *NoneLiteral) expressionNode()               {}
func (nl *NoneLiteral) TokenLiteral() string          { return nl.Token.Literal }
func (nl *NoneLiteral) Pos() token.Position           { return nl.Token.Position }
func (nl *NoneLiteral) String() string                { return "none" }
func (nl *NoneLiteral) GetResolvedType() types.NRType { return nl.Type }
