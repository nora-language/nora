package semantic

import (
	"fmt"

	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/types"
)

// EvalConstInteger recursively evaluates an AST expression into a compile-time constant int64.
// It supports simple literals, prefix, and infix arithmetic/bitwise operations.
func EvalConstInteger(expr ast.Expression, scope *Scope) (int64, error) {
	if expr == nil {
		return 0, fmt.Errorf("cannot evaluate nil expression")
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value, nil
	case *ast.PrefixExpression:
		right, err := EvalConstInteger(e.Right, scope)
		if err != nil {
			return 0, err
		}
		switch e.Operator {
		case "-":
			return -right, nil
		case "+":
			return right, nil
		case "~":
			return ^right, nil
		default:
			return 0, fmt.Errorf("unsupported prefix operator for constant evaluation: %s", e.Operator)
		}
	case *ast.InfixExpression:
		left, err := EvalConstInteger(e.Left, scope)
		if err != nil {
			return 0, err
		}
		right, err := EvalConstInteger(e.Right, scope)
		if err != nil {
			return 0, err
		}

		switch e.Operator {
		case "+":
			return left + right, nil
		case "-":
			return left - right, nil
		case "*":
			return left * right, nil
		case "/":
			if right == 0 {
				return 0, fmt.Errorf("division by zero in constant expression")
			}
			return left / right, nil
		case "%":
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero in constant expression")
			}
			return left % right, nil
		case "<<":
			return left << uint64(right), nil
		case ">>":
			return left >> uint64(right), nil
		case "&":
			return left & right, nil
		case "|":
			return left | right, nil
		case "^":
			return left ^ right, nil
		default:
			return 0, fmt.Errorf("unsupported infix operator for constant evaluation: %s", e.Operator)
		}
	case *ast.Identifier:
		// If it's a constant or variant, evaluate it from the scope
		sym, exists := scope.Resolve(e.Value)
		if !exists || sym == nil {
			return 0, fmt.Errorf("undefined identifier in constant expression: %s", e.Value)
		}
		if sym.Kind == SymConst || sym.Kind == SymVariant {
			// For variants, we use their evaluated values.
			return sym.VariantValue, nil
		}
		return 0, fmt.Errorf("identifier '%s' is not a constant", e.Value)
	case *ast.SelectorExpression:
		// e.g., Enum.Variant
		// For MVP, we can resolve the left, check if it's an enum, then get the variant.
		leftSym, exists := scope.Resolve(e.Left.(*ast.Identifier).Value)
		if exists && leftSym != nil && leftSym.Kind == SymType {
			if st, ok := types.UnwrapLease(leftSym.Type).(*types.SumType); ok {
				if v, exists := st.Variants[e.Field.Value]; exists {
					if !st.IsPrimitiveEnum {
						return 0, fmt.Errorf("cannot use variant of non-primitive enum '%s' as constant integer", st.Name())
					}
					return v.Value, nil
				}
			}
		}
		return 0, fmt.Errorf("unsupported selector in constant expression: %s", e.String())
	}
	return 0, fmt.Errorf("unsupported expression type in constant evaluation: %T", expr)
}
