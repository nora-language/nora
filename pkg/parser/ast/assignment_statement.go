package ast

import "github.com/nora-language/nora/pkg/token"

// AssignmentStatement represents: x = 10; or arr[i] = 10;
type AssignmentStatement struct {
	Token token.Token // The identifier or [ or . token
	Left  Expression
	Value Expression
}

func (as *AssignmentStatement) statementNode()  {} // Implements Statement
func (as *AssignmentStatement) expressionNode() {} // Implements Expression
func (as *AssignmentStatement) Pos() token.Position {
	if as.Left != nil {
		return as.Left.Pos()
	}
	return as.Token.Position
}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }

func (as *AssignmentStatement) String() string {
	return as.Left.String() + " = " + as.Value.String()
}
