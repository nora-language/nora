package ast

import (
	"github.com/nora-language/nora/pkg/token"
)

// chan[i32]
type ChanType struct {
	Token token.Token // 'chan'
	Value TypeNode    // The element type
}

func (ct *ChanType) MarkerTypeNode()      {}
func (ct *ChanType) expressionNode()      {}
func (ct *ChanType) TokenLiteral() string { return ct.Token.Literal }
func (ct *ChanType) Pos() token.Position  { return ct.Token.Position }
func (ct *ChanType) String() string       { return "chan[" + ct.Value.String() + "]" }
