package ast

import "github.com/DwiYI/Project-Nora/pkg/token"

// export fn my_api() { ... }
type ExportStatement struct {
	Token token.Token // 'export'
	Node  Node        // Can be a FunctionLiteral or TypeStatement
}

func (es *ExportStatement) statementNode()       {}
func (es *ExportStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExportStatement) Pos() token.Position  { return es.Token.Position }
func (es *ExportStatement) String() string       { return "export " + es.Node.String() }
