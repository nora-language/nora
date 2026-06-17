package ast

import "github.com/DwiYI/Project-Nora/pkg/token"

type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return bs.Token.Literal + ";" }
func (bs *BreakStatement) Pos() token.Position  { return bs.Token.Position }

type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return cs.Token.Literal + ";" }
func (cs *ContinueStatement) Pos() token.Position  { return cs.Token.Position }
