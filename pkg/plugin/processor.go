package plugin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/plugin/api"
)

// ProcessMacroForFile invokes WASM plugins for any nodes containing attributes (e.g., [serialize]) in a single file.
func (m *PluginManager) ProcessMacroForFile(file *ast.File) error {
	if file.Name == "serialization_generated" || file.Name == "macro_generated" {
		return nil
	}

	var additionalStatements []ast.Statement

	for _, stmt := range file.Statements {
		if ts, ok := stmt.(*ast.TypeStatement); ok {
			if len(ts.Attributes) > 0 {
				for _, attr := range ts.Attributes {
					// e.g. [serialize] -> map to "serialize" plugin
					pluginName := attr.Name
					
					// Try to process using plugin
					generated, err := m.processAttribute(pluginName, &attr, ts)
					if err != nil {
						// If plugin not found, we just ignore. It might not be a macro attribute.
						if strings.Contains(err.Error(), "not loaded") || strings.Contains(err.Error(), "does not export macro") {
							continue
						}
						return fmt.Errorf("error processing attribute [%s] on type %s: %v", pluginName, ts.Name.Value, err)
					}

					if generated != "" {
						// Parse the generated code and append to the file
						l := lexer.New(generated, "macro_generated")
						p := parser.New(l)
						p.DisableMacros = true // prevent recursive macros!
						genFile := p.Parse("macro_generated")
						
						if len(p.Errors()) > 0 {
							return fmt.Errorf("macro '%s' generated invalid syntax: %v\n---\n%s\n---", pluginName, p.Errors(), generated)
						}

						additionalStatements = append(additionalStatements, genFile.Statements...)
					}
				}
			}
		}
	}

	if len(additionalStatements) > 0 {
		file.Statements = append(file.Statements, additionalStatements...)
	}
	return nil
}

func (m *PluginManager) processAttribute(pluginName string, attr *ast.Attribute, node ast.Node) (string, error) {
	nodeDTO, err := MapASTToDTO(node)
	if err != nil {
		return "", err
	}

	var args []string
	for _, a := range attr.Args {
		args = append(args, a)
	}

	req := api.PluginRequest{
		AttributeName: pluginName,
		Args:          args,
		Node:          nodeDTO,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	macroName := "macro_" + pluginName
	resBytes, err := m.ExecuteMacro(pluginName, macroName, reqBytes)
	if err != nil {
		return "", err
	}

	var res api.PluginResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return "", fmt.Errorf("failed to unmarshal macro response: %v", err)
	}

	if res.Error != "" {
		return "", fmt.Errorf("macro returned error: %s", res.Error)
	}

	// Map any modifications back to the original node
	if err := MapDTOToAST(node, res.Node); err != nil {
		return "", err
	}

	return res.GeneratedCode, nil
}
