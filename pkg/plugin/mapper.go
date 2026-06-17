package plugin

import (
	"fmt"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/plugin/api"
	"github.com/nora-language/nora/pkg/semantic"
)

// MapASTToDTO converts an internal compiler AST node to a stable DTO for plugins.
func MapASTToDTO(node ast.Node) (api.NodeDTO, error) {
	switch n := node.(type) {
	case *ast.FunctionStatement:
		dto := api.FunctionDTO{
			Name:     n.Name.String(),
			IsExport: n.IsExport,
			IsExtern: n.IsExtern,
		}

		for _, p := range n.Parameters {
			dto.Parameters = append(dto.Parameters, p.String())
		}

		if n.ReturnType != nil {
			dto.ReturnType = n.ReturnType.String()
		}

		if n.Body != nil {
			dto.Body = n.Body.String()
		}

		return api.NodeDTO{
			Kind:     api.KindFunction,
			Function: &dto,
		}, nil

	case *ast.TypeStatement:
		dto := api.TypeDTO{
			Name: n.Name.String(),
		}
		if n.Value != nil {
			dto.Value = n.Value.String()
		}
		return api.NodeDTO{
			Kind: api.KindType,
			Type: &dto,
		}, nil

	default:
		return api.NodeDTO{}, fmt.Errorf("unsupported node type for plugin serialization: %T", node)
	}
}

// MapDTOToAST applies the DTO modifications back onto the original AST node.
func MapDTOToAST(original ast.Node, dto api.NodeDTO) error {
	switch n := original.(type) {
	case *ast.FunctionStatement:
		if dto.Kind != api.KindFunction || dto.Function == nil {
			return fmt.Errorf("expected Function DTO, got %s", dto.Kind)
		}

		fDTO := dto.Function

		// If the macro modified the body string, we must parse it back into an AST block.
		if fDTO.Body != "" && (n.Body == nil || fDTO.Body != n.Body.String()) {
			// Wrap the body in a dummy function so the parser can handle it easily
			// Or we can just parse it directly by instantiating a new parser.
			source := fDTO.Body
			// Strip leading/trailing braces if they exist in the DTO body string to parse its contents,
			// or just parse the whole block:
			if source[0] != '{' {
				source = "{\n" + source + "\n}"
			}

			// A trick to parse a block statement
			// The parser doesn't have an exported ParseBlock method directly accessible from outside without a hack,
			// but we can parse a dummy function and extract its body.
			dummySource := "fn dummy() " + source
			dummyL := lexer.New(dummySource, "macro_generated")
			dummyP := parser.New(dummyL)
			file := dummyP.Parse("macro_generated")

			if len(dummyP.Errors()) > 0 {
				return fmt.Errorf("failed to parse macro-generated body: %v", dummyP.Errors())
			}

			if len(file.Statements) > 0 {
				if fn, ok := file.Statements[0].(*ast.FunctionStatement); ok {
					n.Body = fn.Body
				}
			}
		}

	case *ast.TypeStatement:
		if dto.Kind != api.KindType || dto.Type == nil {
			return fmt.Errorf("expected Type DTO, got %s", dto.Kind)
		}
		// Logic to map Type modifications back can be added here
	}

	return nil
}

func MapCallToDTO(n *ast.CallExpression, info *semantic.SemanticInfo) (api.CallRequest, error) {
	dto := api.CallRequest{}

	if ident, ok := n.Function.(*ast.Identifier); ok {
		dto.FunctionName = ident.Value
	} else if sel, ok := n.Function.(*ast.SelectorExpression); ok {
		dto.FunctionName = sel.Field.Value
	}

	for _, arg := range n.Arguments {
		argDTO := api.CallArgumentDTO{
			Value: arg.Value.String(),
		}
		if t, ok := info.Types[arg.Value]; ok {
			argDTO.Type = t.Name()
		}
		dto.Arguments = append(dto.Arguments, argDTO)
	}

	return dto, nil
}
