package ast

import (
	"github.com/nora-language/nora/pkg/token"
)

// FieldDefinition: <name>: <type> (e.g., age: u32)
type FieldDefinition struct {
	Doc        *CommentGroup
	Token      token.Token // The identifier token of the field name
	Name       *Identifier // "age"
	Type       TypeNode    // "u32" (for declarations)
	Value      Expression  // for literals (e.g. name: "Alice")
	Attributes []Attribute
}

func (fd *FieldDefinition) Pos() token.Position {
	if len(fd.Attributes) > 0 {
		return fd.Attributes[0].Pos()
	}
	return fd.Token.Position
}
func (fd *FieldDefinition) TokenLiteral() string { return fd.Token.Literal }

func (fd *FieldDefinition) String() string {
	res := ""
	for _, attr := range fd.Attributes {
		res += attr.String() + "\n"
	}
	res += fd.Name.String() + ": "
	if fd.Type != nil {
		res += fd.Type.String()
	} else if fd.Value != nil {
		res += fd.Value.String()
	}
	return res
}
