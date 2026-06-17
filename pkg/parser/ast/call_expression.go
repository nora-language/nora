package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type CallExpression struct {
	Token         token.Token // '('
	Function      Expression  // Identifier or FunctionLiteral
	TypeArguments []TypeNode  // add[i32](...)
	Arguments     []*ArgumentsExpression
	Type          types.NRType
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) Pos() token.Position {
	if ce.Function != nil {
		return ce.Function.Pos()
	}
	return ce.Token.Position
}

func (ce *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function.String())
	if len(ce.TypeArguments) > 0 {
		out.WriteString("[")
		targs := []string{}
		for _, ta := range ce.TypeArguments {
			targs = append(targs, ta.String())
		}
		out.WriteString(strings.Join(targs, ", "))
		out.WriteString("]")
	}
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString("(" + strings.Join(args, ", ") + ")")
	return out.String()
}
