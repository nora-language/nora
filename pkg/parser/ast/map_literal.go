package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

// MapLiteral: { "key": value }
type MapLiteral struct {
	Token token.Token // Token '{'
	Pairs map[Expression]Expression

	Type types.NRType
}

func (ml *MapLiteral) expressionNode()               {}
func (ml *MapLiteral) TokenLiteral() string          { return ml.Token.Literal }
func (ml *MapLiteral) Pos() token.Position           { return ml.Token.Position }
func (ml *MapLiteral) GetResolvedType() types.NRType { return ml.Type }

func (ml *MapLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range ml.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
