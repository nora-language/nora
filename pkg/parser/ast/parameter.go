package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

type Parameter struct {
	Token      token.Token // The identifier token
	Name       *Identifier
	Type       TypeNode // The type annotation (e.g., i32, User, [str])
	LeaseKind  LeaseKind
	IsVariadic bool
}

func (p *Parameter) TokenLiteral() string { return p.Token.Literal }
func (p *Parameter) Pos() token.Position  { return p.Token.Position }

func (p *Parameter) String() string {
	var out bytes.Buffer

	// 2. Print Name
	out.WriteString(p.Name.String())
	out.WriteString(": ")

	// 3. Print Type
	if p.IsVariadic {
		out.WriteString("...")
	} else if p.Type != nil {
		out.WriteString(p.Type.String())
	}

	return out.String()
}
