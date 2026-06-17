package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

// struct { name: string, age: i32 }
type StructLiteral struct {
	Token  token.Token // 'struct'
	Name   Expression
	Fields []*FieldDefinition
}

func (sl *StructLiteral) expressionNode()      {}
func (sl *StructLiteral) MarkerTypeNode()      {}
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StructLiteral) Pos() token.Position {
	if sl.Name != nil {
		return sl.Name.Pos()
	}
	return sl.Token.Position
}

func (sl *StructLiteral) String() string {
	var out bytes.Buffer
	fields := []string{}
	for _, f := range sl.Fields {
		fields = append(fields, f.String())
	}
	out.WriteString("struct { " + strings.Join(fields, ", ") + " }")
	return out.String()
}
