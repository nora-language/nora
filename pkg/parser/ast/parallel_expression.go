package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

type ParallelExpression struct {
	Token token.Token // 'parallel'
	Body  *BlockStatement
	Type  types.NRType
}

func (pe *ParallelExpression) expressionNode()               {}
func (pe *ParallelExpression) TokenLiteral() string          { return pe.Token.Literal }
func (pe *ParallelExpression) Pos() token.Position           { return pe.Token.Position }
func (pe *ParallelExpression) String() string                { return "parallel " + pe.Body.String() }
func (pe *ParallelExpression) GetResolvedType() types.NRType { return pe.Type }
