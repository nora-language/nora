package ast

import (
	"bytes"

	"github.com/nora-language/nora/pkg/token"
)

type BlockStatement struct {
	Token      token.Token // '{'
	Rbrace     token.Token // '}'
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) expressionNode()      {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) Pos() token.Position  { return bs.Token.Position }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString("}")
	return out.String()
}
