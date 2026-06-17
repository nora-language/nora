package ast

import "github.com/DwiYI/Project-Nora/pkg/token"

type DeferStatement struct {
	Token token.Token
	Call  *CallExpression
}

func (ds *DeferStatement) statementNode()       {}
func (ds *DeferStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DeferStatement) Pos() token.Position  { return ds.Token.Position }
func (ds *DeferStatement) String() string       { return "defer " + ds.Call.String() }
