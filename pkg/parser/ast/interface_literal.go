package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

// InterfaceLiteral represents: interface { fn say_hello(self: Greeter) }
type InterfaceLiteral struct {
	Token    token.Token          // The 'interface' or 'protocol' token
	Methods  []*FunctionStatement // Method signatures (Body will be nil)
	Embedded []*Identifier        // Embedded interfaces
}

func (il *InterfaceLiteral) expressionNode()      {}
func (il *InterfaceLiteral) MarkerTypeNode()      {}
func (il *InterfaceLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *InterfaceLiteral) Pos() token.Position  { return il.Token.Position }

func (il *InterfaceLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(il.Token.Literal)
	out.WriteString(" {\n")

	for _, e := range il.Embedded {
		out.WriteString("    ")
		out.WriteString(e.String())
		out.WriteString("\n")
	}

	for _, m := range il.Methods {
		out.WriteString("    ")
		out.WriteString(m.String())
		out.WriteString("\n")
	}

	out.WriteString("}")

	return out.String()
}
