package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

type ConstStatement struct {
	Doc      *CommentGroup
	Token    token.Token
	Name     *Identifier
	Type     TypeNode // Optional
	Value    Expression
	IsPublic bool
}

func (ls *ConstStatement) statementNode()       {}
func (ls *ConstStatement) Pos() token.Position  { return ls.Token.Position }
func (ls *ConstStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *ConstStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " " + ls.Name.String())
	if ls.Type != nil {
		out.WriteString(": " + ls.Type.String())
	}
	if ls.Value != nil {
		out.WriteString(" = " + ls.Value.String())
	}
	return out.String()
}
