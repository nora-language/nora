package ast

import (
	"bytes"

	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type SpawnExpression struct {
	Token          token.Token // 'spawn'
	MonitorChannel Expression  // Optional channel for panic monitoring
	Call           *CallExpression
	CapturesId     []string
	Type           types.NRType
	Body           *BlockStatement
}

func (se *SpawnExpression) expressionNode()               {}
func (se *SpawnExpression) TokenLiteral() string          { return se.Token.Literal }
func (se *SpawnExpression) Pos() token.Position           { return se.Token.Position }
func (se *SpawnExpression) GetResolvedType() types.NRType { return se.Type }

func (se *SpawnExpression) String() string {
	var out bytes.Buffer

	out.WriteString("spawn ")
	if se.Call != nil {
		out.WriteString(se.Call.String())
	}

	return out.String()
}
