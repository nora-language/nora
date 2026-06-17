package ast

import (
	"bytes"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

// arr[index] or Map[K, V]
type IndexExpression struct {
	Token   token.Token  // '['
	Left    Expression   // The array/map/generic
	Indices []Expression // Supports multiple indices for generics: Map[K, V]
	NoBoundsCheck bool   // True if static analysis proves this access is 100% safe

	Type types.NRType
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) MarkerTypeNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) Pos() token.Position {
	if ie.Left != nil {
		return ie.Left.Pos()
	}
	return ie.Token.Position
}

func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	if ie.Left != nil && !IsNil(ie.Left) {
		out.WriteString(ie.Left.String())
	} else {
		out.WriteString("")
	}
	out.WriteString("[")
	indices := []string{}
	for _, idx := range ie.Indices {
		if idx != nil && !IsNil(idx) {
			indices = append(indices, idx.String())
		} else {
			indices = append(indices, "")
		}
	}
	out.WriteString(strings.Join(indices, ", "))
	out.WriteString("])")
	return out.String()
}
