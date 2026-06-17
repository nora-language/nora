package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
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
