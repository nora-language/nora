package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

// arr[start:end]
type SliceExpression struct {
	Token token.Token // ':'
	Start Expression  // Optional (nil if omitted: [:end])
	End   Expression  // Optional (nil if omitted: [start:])
}

func (se *SliceExpression) expressionNode()      {}
func (se *SliceExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SliceExpression) Pos() token.Position  { return se.Token.Position }

func (se *SliceExpression) String() string {
	var out bytes.Buffer
	if se.Start != nil {
		out.WriteString(se.Start.String())
	}
	out.WriteString(":")
	if se.End != nil {
		out.WriteString(se.End.String())
	}
	return out.String()
}
