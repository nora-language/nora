package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

// type User
type TypeStatement struct {
	Doc            *CommentGroup
	Token          token.Token // 'type'
	Name           *Identifier // "User"
	TypeParameters []*TypeParameter
	Value          Expression // usually *StructLiteral
	Attributes     []Attribute
	IsPublic       bool
}

func (ts *TypeStatement) statementNode()       {}
func (ts *TypeStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TypeStatement) Pos() token.Position {
	if len(ts.Attributes) > 0 {
		return ts.Attributes[0].Pos()
	}
	return ts.Token.Position
}

func (ts *TypeStatement) String() string {
	var out bytes.Buffer
	out.WriteString("type ")
	if ts.Name != nil {
		out.WriteString(ts.Name.String())
	}
	if len(ts.TypeParameters) > 0 {
		out.WriteString("[")
		tparams := []string{}
		for _, tp := range ts.TypeParameters {
			tparams = append(tparams, tp.String())
		}
		out.WriteString(strings.Join(tparams, ", "))
		out.WriteString("]")
	}
	out.WriteString(" = ")
	if ts.Value != nil {
		out.WriteString(ts.Value.String())
	}
	return out.String()
}
