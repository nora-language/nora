package ast

import (
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

// 3.14
type FloatLiteral struct {
	Token  token.Token
	Value  float64
	Suffix string
	Type   types.NRType
}

func (fl *FloatLiteral) expressionNode()               {}
func (fl *FloatLiteral) TokenLiteral() string          { return fl.Token.Literal }
func (fl *FloatLiteral) Pos() token.Position           { return fl.Token.Position }
func (fl *FloatLiteral) String() string                { return fl.Token.Literal }
func (fl *FloatLiteral) GetResolvedType() types.NRType { return fl.Type }
