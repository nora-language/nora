package ast

import "github.com/nora-language/nora/pkg/token"

// continue / break
type BranchStatement struct {
	Token token.Token // token.CONTINUE or token.BREAK
}

func (bs *BranchStatement) statementNode()       {}
func (bs *BranchStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BranchStatement) Pos() token.Position  { return bs.Token.Position }
func (bs *BranchStatement) String() string       { return bs.Token.Literal + ";" }
