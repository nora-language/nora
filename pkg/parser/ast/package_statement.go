package ast

import "github.com/nora-language/nora/pkg/token"

type PackageStatement struct {
	Doc   *CommentGroup
	Token token.Token
	Name  *Identifier
}

func (ps *PackageStatement) statementNode()       {}
func (ps *PackageStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PackageStatement) Pos() token.Position  { return ps.Token.Position }
func (ps *PackageStatement) String() string       { return "package " + ps.Name.String() + ";" }
