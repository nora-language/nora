package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
	Type  types.NRType
}

func (i *Identifier) expressionNode()               {}
func (i *Identifier) TokenLiteral() string          { return i.Token.Literal }
func (i *Identifier) Pos() token.Position           { return i.Token.Position }
func (i *Identifier) GetResolvedType() types.NRType { return i.Type }
func (i *Identifier) String() string                { return i.Value }
