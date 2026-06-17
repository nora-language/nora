package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

// type Status = enum { Active, Inactive, Pending(u32) }
type SumTypeLiteral struct {
	Token    token.Token // The 'enum' or 'type' token
	Variants []*VariantDefinition
	Type     types.NRType
}

func (sl *SumTypeLiteral) expressionNode()               {}
func (sl *SumTypeLiteral) MarkerTypeNode()               {}
func (sl *SumTypeLiteral) TokenLiteral() string          { return sl.Token.Literal }
func (sl *SumTypeLiteral) Pos() token.Position           { return sl.Token.Position }
func (sl *SumTypeLiteral) GetResolvedType() types.NRType { return sl.Type }

func (sl *SumTypeLiteral) String() string {
	var out bytes.Buffer
	variants := []string{}
	for _, v := range sl.Variants {
		variants = append(variants, v.String())
	}
	out.WriteString("enum { " + strings.Join(variants, ", ") + " }")
	return out.String()
}

type VariantDefinition struct {
	Doc    *CommentGroup
	Token  token.Token // The identifier of the variant
	Name   *Identifier
	Fields []*FieldDefinition // Optional data fields (for Enums with data)
}

func (vd *VariantDefinition) Pos() token.Position  { return vd.Token.Position }
func (vd *VariantDefinition) TokenLiteral() string { return vd.Token.Literal }

func (vd *VariantDefinition) String() string {
	var out bytes.Buffer
	out.WriteString(vd.Name.Value)
	if len(vd.Fields) > 0 {
		out.WriteString("(")
		// ... format fields ...
		out.WriteString(")")
	}
	return out.String()
}
