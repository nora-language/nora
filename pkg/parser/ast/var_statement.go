package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

// LetMutStatement handles both cases:
// 1. "var x = 5"
type VarStatement struct {
	Doc      *CommentGroup
	Token    token.Token
	Name     *Identifier
	Type     TypeNode // Optional: let x: i32 = ...
	Value    Expression
	IsPublic bool
}

func (ls *VarStatement) statementNode()       {}
func (ls *VarStatement) Pos() token.Position  { return ls.Token.Position }
func (ls *VarStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *VarStatement) String() string {
	var out bytes.Buffer
	// Will print "let" or "mut" depending on the token
	out.WriteString(ls.TokenLiteral() + " " + ls.Name.String())
	if ls.Type != nil {
		out.WriteString(": " + ls.Type.String())
	}
	if ls.Value != nil {
		out.WriteString(" = " + ls.Value.String())
	}
	return out.String()
}
