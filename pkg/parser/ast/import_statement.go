package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

type ImportStatement struct {
	Token token.Token // The 'import' keyword
	Path  *Identifier // The path identifier (virtual identifier created from string literal)
	Alias *Identifier // Optional: import m "math" (can be nil)
}

func (is *ImportStatement) statementNode() {}

// TokenLiteral returns the literal value of the 'import' token
func (is *ImportStatement) TokenLiteral() string {
	return is.Token.Literal
}

// Pos returns the position of the 'import' keyword
func (is *ImportStatement) Pos() token.Position {
	return is.Token.Position
}

// PathValue returns the actual string path without quotes (helper)
// Assumes token.Literal is like "\"math\""
func (is *ImportStatement) PathValue() string {
	if is.Path == nil {
		return ""
	}
	return is.Path.Value
}

func (is *ImportStatement) String() string {
	var out bytes.Buffer

	out.WriteString(is.TokenLiteral() + " ")

	if is.Alias != nil {
		out.WriteString(is.Alias.String() + " ")
	}

	if is.Path != nil {
		out.WriteString(`"` + is.Path.Value + `"`)
	}

	return out.String()
}
