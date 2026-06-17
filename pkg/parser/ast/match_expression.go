package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type MatchExpression struct {
	Token  token.Token
	Target Expression
	Cases  []*MatchCase

	Type types.NRType
}

type MatchCase struct {
	Pattern Expression
	Body    *BlockStatement
}

func (me *MatchExpression) expressionNode()               {}
func (me *MatchExpression) TokenLiteral() string          { return me.Token.Literal }
func (me *MatchExpression) Pos() token.Position           { return me.Token.Position }
func (me *MatchExpression) GetResolvedType() types.NRType { return me.Type }

// MatchExpression String implementation
func (me *MatchExpression) String() string {
	var out bytes.Buffer

	out.WriteString("match ")
	out.WriteString(me.Target.String())
	out.WriteString(" { ")

	cases := []string{}
	for _, c := range me.Cases {
		cases = append(cases, c.String())
	}
	out.WriteString(strings.Join(cases, ", "))

	out.WriteString(" }")

	return out.String()
}

// MatchCase String implementation
func (mc *MatchCase) String() string {
	var out bytes.Buffer

	out.WriteString(mc.Pattern.String())
	out.WriteString(" => ")
	if mc.Body != nil {
		out.WriteString(mc.Body.String())
	}

	return out.String()
}
