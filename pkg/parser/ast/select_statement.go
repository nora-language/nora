package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

type SelectStatement struct {
	Token token.Token // The 'select' token
	Cases []*SelectCase
}

func (ss *SelectStatement) statementNode()       {}
func (ss *SelectStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SelectStatement) Pos() token.Position  { return ss.Token.Position }
func (ss *SelectStatement) String() string {
	var out bytes.Buffer
	out.WriteString("select {\n")
	for _, c := range ss.Cases {
		out.WriteString(c.String())
	}
	out.WriteString("}")
	return out.String()
}

type SelectCase struct {
	Token     token.Token // 'case' or 'default'
	Condition Statement   // A SendStatement or AssignmentStatement (for Recv) or nil (for default)
	Body      *BlockStatement
}

func (sc *SelectCase) Pos() token.Position  { return sc.Token.Position }
func (sc *SelectCase) TokenLiteral() string { return sc.Token.Literal }

func (sc *SelectCase) String() string {
	var out bytes.Buffer
	if sc.Condition == nil {
		out.WriteString("default:\n")
	} else {
		out.WriteString("case ")
		out.WriteString(sc.Condition.String())
		out.WriteString(":\n")
	}
	out.WriteString(sc.Body.String())
	return out.String()
}
