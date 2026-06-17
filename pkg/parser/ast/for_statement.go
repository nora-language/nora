package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

type ForStatement struct {
	Token token.Token // The 'for' token

	// For-In loops
	Key      *Identifier
	Value    *Identifier
	Iterable Expression

	Body *BlockStatement

	NextCall *CallExpression // Synthetic call for generic Next() methods
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) Pos() token.Position  { return fs.Token.Position }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	if fs.Value != nil {
		if fs.Key != nil {
			out.WriteString(fs.Key.String() + ", ")
		}
		out.WriteString(fs.Value.String() + " in " + fs.Iterable.String())
	}
	out.WriteString(" ")
	out.WriteString(fs.Body.String())

	return out.String()
}
