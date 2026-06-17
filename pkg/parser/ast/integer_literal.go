package ast

import (
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

type IntegerLiteral struct {
	Token  token.Token
	Value  int64
	Suffix string

	Type types.NRType
}

func (il *IntegerLiteral) expressionNode()               {}
func (il *IntegerLiteral) TokenLiteral() string          { return il.Token.Literal }
func (il *IntegerLiteral) Pos() token.Position           { return il.Token.Position }
func (il *IntegerLiteral) String() string                { return il.Token.Literal }
func (il *IntegerLiteral) GetResolvedType() types.NRType { return il.Type }
