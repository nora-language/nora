package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

// SelectorExpression represents dot notation: Left.Field (e.g. "math.Pi" or "user.id")
type SelectorExpression struct {
	Token token.Token // The '.' token
	Left  Expression  // The expression before the dot
	Field *Identifier // The identifier after the dot
}

func (se *SelectorExpression) expressionNode() {}
func (se *SelectorExpression) MarkerTypeNode() {}
func (se *SelectorExpression) Pos() token.Position {
	if se.Left != nil {
		return se.Left.Pos()
	}
	return se.Token.Position
}
func (se *SelectorExpression) TokenLiteral() string { return se.Token.Literal }

func (se *SelectorExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	if se.Left != nil && !IsNil(se.Left) {
		out.WriteString(se.Left.String())
	}
	out.WriteString(".")
	if se.Field != nil && !IsNil(se.Field) {
		out.WriteString(se.Field.String())
	}
	out.WriteString(")")
	return out.String()
}
