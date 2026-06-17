package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

type FunctionType struct {
	Token      token.Token // The 'fn' token
	Parameters []TypeNode
	ReturnType TypeNode
}

func (ft *FunctionType) TokenLiteral() string { return ft.Token.Literal }
func (ft *FunctionType) Pos() token.Position  { return ft.Token.Position }
func (ft *FunctionType) expressionNode()      {}
func (ft *FunctionType) MarkerTypeNode()      {}

func (ft *FunctionType) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range ft.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if ft.ReturnType != nil {
		out.WriteString(" ")
		out.WriteString(ft.ReturnType.String())
	}
	return out.String()
}
