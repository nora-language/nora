package ast

import (
	"strings"
	"testing"

	"github.com/DwiYI/Project-Nora/pkg/token"
)

func TestCompleteAST(t *testing.T) {
	program := &Program{
		Files: []*File{
			{
				Statements: []Statement{
					// 1. Package & Import
					&PackageStatement{
						Token: token.Token{Type: token.PACKAGE, Literal: "package"},
						Name:  &Identifier{Value: "main"},
					},
					&ImportStatement{
						Token: token.Token{Type: token.IMPORT, Literal: "import"},
						Path:  &Identifier{Value: "std/io"},
					},

					// 2. Type Definition (Struct)
					&TypeStatement{
						Token: token.Token{Type: token.TYPE, Literal: "type"},
						Name:  &Identifier{Value: "User"},
						Value: &StructLiteral{
							Token: token.Token{Type: token.STRUCT, Literal: "struct"},
							Fields: []*FieldDefinition{
								{
									// We use the IDENT token of the field name
									Token: token.Token{Type: token.IDENT, Literal: "id"},
									Name:  &Identifier{Value: "id"},
									Type:  &Identifier{Value: "u64"},
								},
							},
						},
					},

					&VarStatement{
						Token: token.Token{Type: token.VAR, Literal: "var"},
						Name:  &Identifier{Value: "x"},
						Value: &IntegerLiteral{Token: token.Token{Literal: "5"}, Value: 5},
					},

					// 4. Function with Concurrency (Spawn)
					&FunctionStatement{
						Token: token.Token{Type: token.FN, Literal: "fn"},
						Name:  &Identifier{Value: "main"},
						Body: &BlockStatement{
							Statements: []Statement{
								&ExpressionStatement{
									Expression: &SpawnExpression{
										Token: token.Token{Type: token.SPAWN, Literal: "spawn"},
										Call: &CallExpression{
											Function: &Identifier{Value: "worker"},
											Arguments: []*ArgumentsExpression{
												{
													Token: token.Token{Type: token.IDENT, Literal: "x"},
													Value: &Identifier{Value: "x"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	output := program.String()

	expectedParts := []string{
		"package main",
		"import \"std/io\"",
		"type User = struct { id: u64 }",
		"var x = 5",
		"fn main() {spawn worker(x)}",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected AST output to contain %q. Got: %q", part, output)
		}
	}
}
