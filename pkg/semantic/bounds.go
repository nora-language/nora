package semantic

import (
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
)

// VarBounds represents the mathematical bounds of an integer variable for Range Analysis (BCE).
type VarBounds struct {
	MinBound  int64          // Lower bound (e.g., 0)
	MaxSymbol ast.Expression // Upper bound expression (e.g., len(arr))
}

// IsSemanticallyEquivalent safely compares two AST expressions for structural equality.
// This is used to determine if the `MaxSymbol` matches the array being indexed (e.g., `len(arr)` vs `arr`).
func IsSemanticallyEquivalent(e1, e2 ast.Expression) bool {
	if e1 == nil || e2 == nil {
		return e1 == e2
	}

	switch n1 := e1.(type) {
	case *ast.Identifier:
		if n2, ok := e2.(*ast.Identifier); ok {
			return n1.Value == n2.Value
		}
	case *ast.SelectorExpression:
		if n2, ok := e2.(*ast.SelectorExpression); ok {
			return IsSemanticallyEquivalent(n1.Left, n2.Left) && IsSemanticallyEquivalent(n1.Field, n2.Field)
		}
	case *ast.CallExpression:
		if n2, ok := e2.(*ast.CallExpression); ok {
			if !IsSemanticallyEquivalent(n1.Function, n2.Function) {
				return false
			}
			if len(n1.Arguments) != len(n2.Arguments) {
				return false
			}
			for i := range n1.Arguments {
				if !IsSemanticallyEquivalent(n1.Arguments[i].Value, n2.Arguments[i].Value) {
					return false
				}
			}
			return true
		}
	case *ast.IndexExpression:
		if n2, ok := e2.(*ast.IndexExpression); ok {
			if !IsSemanticallyEquivalent(n1.Left, n2.Left) {
				return false
			}
			if len(n1.Indices) != len(n2.Indices) {
				return false
			}
			for i := range n1.Indices {
				if !IsSemanticallyEquivalent(n1.Indices[i], n2.Indices[i]) {
					return false
				}
			}
			return true
		}
	case *ast.IntegerLiteral:
		if n2, ok := e2.(*ast.IntegerLiteral); ok {
			return n1.Value == n2.Value
		}
	}

	// We only support simple expressions for BCE guarantees to remain conservative.
	return false
}
