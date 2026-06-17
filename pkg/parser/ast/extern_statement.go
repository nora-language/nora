package ast

import "github.com/nora-language/nora/pkg/token"

// extern fn printf(fmt: str, ...)
type ExternStatement struct {
	Token    token.Token // 'extern'
	Function *FunctionStatement
}

func (es *ExternStatement) statementNode()       {}
func (es *ExternStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExternStatement) Pos() token.Position  { return es.Token.Position }
func (es *ExternStatement) String() string       { return "extern " + es.Function.String() }
