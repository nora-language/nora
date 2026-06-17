package ast

import (
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

// 5i, 10j

type ImaginaryLiteral struct {
	Token  token.Token // The 'IMAG' token (e.g., 5i)
	Value  float64     // The numeric part (5.0)
	Suffix string

	Type types.NRType
}

func (il *ImaginaryLiteral) Pos() token.Position           { return il.Token.Position }
func (il *ImaginaryLiteral) expressionNode()               {}
func (il *ImaginaryLiteral) TokenLiteral() string          { return il.Token.Literal }
func (il *ImaginaryLiteral) String() string                { return il.Token.Literal }
func (il *ImaginaryLiteral) GetResolvedType() types.NRType { return il.Type }
