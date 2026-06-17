package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

// FunctionStatement represents: fn add(a: i32, b: i32) i32 { ... }
type FunctionStatement struct {
	Doc               *CommentGroup
	Token             token.Token      // The 'fn' token
	Name              *Identifier      // Optional name (e.g. "add")
	Receiver          *Parameter       // Optional receiver for methods: (self: #Point)
	TypeParameters    []*TypeParameter // [T: Printable, U]
	Parameters        []*Parameter     // Changed from []*Identifier
	ReturnType        TypeNode         // Optional return type (e.g. "i32")
	Body              *BlockStatement
	IsExtern          bool        // Whether this function is an FFI extern
	IsExport          bool        // Whether this function is exported to C
	Attributes        []Attribute // Custom annotations like [memoize]
	IsGenericTemplate bool        // True if this is an instance but still generic (e.g. from T passed to T)
	IsPublic          bool
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) Pos() token.Position {
	if len(fs.Attributes) > 0 {
		// Calculate the starting position from the first attribute
		// However, ast.Attribute's Pos() is just its name, we want the '[' if possible.
		// Since we don't store '[' position, we can just return the attribute name's pos.
		// It's still before the 'fn' keyword.
		return fs.Attributes[0].Pos()
	}
	return fs.Token.Position
}

func (fl *FunctionStatement) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral()) // "fn"

	if fl.Receiver != nil {
		out.WriteString(" (" + fl.Receiver.String() + ")")
	}

	if fl.Name != nil {
		out.WriteString(" " + fl.Name.String())
	}

	if len(fl.TypeParameters) > 0 {
		out.WriteString("[")
		tparams := []string{}
		for _, tp := range fl.TypeParameters {
			tparams = append(tparams, tp.String())
		}
		out.WriteString(strings.Join(tparams, ", "))
		out.WriteString("]")
	}

	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")

	if fl.ReturnType != nil {
		out.WriteString(" " + fl.ReturnType.String())
	}

	out.WriteString(" ")
	if fl.Body != nil {
		out.WriteString(fl.Body.String())
	}

	return out.String()
}
