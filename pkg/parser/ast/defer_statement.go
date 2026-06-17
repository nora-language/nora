package ast

import "github.com/nora-language/nora/pkg/token"

type DeferStatement struct {
	Token token.Token
	Call  *CallExpression
}

func (ds *DeferStatement) statementNode()       {}
func (ds *DeferStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DeferStatement) Pos() token.Position  { return ds.Token.Position }
func (ds *DeferStatement) String() string       { return "defer " + ds.Call.String() }
