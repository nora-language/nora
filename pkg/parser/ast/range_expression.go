package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

type RangeExpression struct {
	Token token.Token // The '..' token
	Start Expression
	End   Expression
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) Pos() token.Position  { return re.Token.Position }
func (re *RangeExpression) String() string {
	var out bytes.Buffer

	out.WriteString(re.Start.String())
	out.WriteString(" .. ")
	out.WriteString(re.End.String())

	return out.String()
}
