package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

type InterpolatedString struct {
	Token token.Token
	Parts []Expression // Mixture of *StringLiteral and other Expressions
}

func (is *InterpolatedString) expressionNode()      {}
func (is *InterpolatedString) TokenLiteral() string { return is.Token.Literal }
func (is *InterpolatedString) Pos() token.Position  { return is.Token.Position }
func (is *InterpolatedString) String() string {
	var out bytes.Buffer
	out.WriteString("\"")
	for _, part := range is.Parts {
		if sl, ok := part.(*StringLiteral); ok {
			out.WriteString(sl.Value)
		} else {
			out.WriteString("${")
			out.WriteString(part.String())
			out.WriteString("}")
		}
	}
	out.WriteString("\"")
	return out.String()
}
