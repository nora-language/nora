package ast

import (
	"bytes"
	"strings"

	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

// ArrayLiteral: [1, 2, 3]
type ArrayLiteral struct {
	Token    token.Token // Token '['
	Elements []Expression

	Type types.NRType
}

func (al *ArrayLiteral) expressionNode()               {}
func (al *ArrayLiteral) TokenLiteral() string          { return al.Token.Literal }
func (al *ArrayLiteral) Pos() token.Position           { return al.Token.Position }
func (al *ArrayLiteral) GetResolvedType() types.NRType { return al.Type }

func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
