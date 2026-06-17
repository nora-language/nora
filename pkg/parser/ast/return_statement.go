package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
)

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) Pos() token.Position  { return rs.Token.Position }
func (rs *ReturnStatement) String() string       { return "return " + rs.ReturnValue.String() + ";" }
