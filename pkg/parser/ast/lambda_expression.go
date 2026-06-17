package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

// LambdaExpression represents: fn(a: i32) i32 { ... }
type LambdaExpression struct {
	Token      token.Token // The 'fn' token
	Parameters []*Parameter
	ReturnType TypeNode // Optional return type
	Body       *BlockStatement
}

func (le *LambdaExpression) expressionNode()      {}
func (le *LambdaExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LambdaExpression) Pos() token.Position  { return le.Token.Position }

func (le *LambdaExpression) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range le.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(le.TokenLiteral()) // "fn"
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")

	if le.ReturnType != nil {
		out.WriteString(" " + le.ReturnType.String())
	}

	out.WriteString(" ")
	if le.Body != nil {
		out.WriteString(le.Body.String())
	}

	return out.String()
}
