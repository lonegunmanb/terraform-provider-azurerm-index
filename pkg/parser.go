package pkg

import (
	"go/ast"
	"go/token"
	"strings"
)

// TerraformResourceMapping represents a mapping between terraform resource type and its registration method
type TerraformResourceMapping struct {
	TerraformType      string `json:"terraform_type"`      // e.g., "azurerm_resource_group"
	RegistrationMethod string `json:"registration_method"` // e.g., "resourceResourceGroup"
}

// ExtractSupportedResourcesMappings extracts mappings from SupportedResources method in the AST
func ExtractSupportedResourcesMappings(node *ast.File) map[string]string {
	return extractSupportedResourcesMappings(node)
}

// extractSupportedResourcesMappings extracts mappings from SupportedResources method in the AST
func extractSupportedResourcesMappings(node *ast.File) map[string]string {
	mappings := make(map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "SupportedResources" {
			return true
		}

		// Look for return statements in the function body
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			// Process each return expression
			for _, result := range returnStmt.Results {
				// Handle direct map literal return
				if mapLit, ok := result.(*ast.CompositeLit); ok {
					extractFromMapLiteral(mapLit, mappings)
				}

				// Handle variable reference (like "resources" variable)
				ident, ok := result.(*ast.Ident)
				if !ok {
					continue
				}
				// Find the variable definition in the function
				ast.Inspect(fn.Body, func(varNode ast.Node) bool {
					assignStmt, ok := varNode.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for i, lhs := range assignStmt.Lhs {
						lhsIdent, ok := lhs.(*ast.Ident)
						if !ok || lhsIdent.Name != ident.Name {
							return true
						}
						if i >= len(assignStmt.Rhs) {
							return true
						}
						if mapLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit); ok {
							extractFromMapLiteral(mapLit, mappings)
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return mappings
}

// extractFromMapLiteral extracts key-value pairs from a map literal
func extractFromMapLiteral(mapLit *ast.CompositeLit, mappings map[string]string) {
	for _, elt := range mapLit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			// Extract the key (terraform resource type)
			var key string
			if keyLit, ok := kv.Key.(*ast.BasicLit); ok && keyLit.Kind == token.STRING {
				key = strings.Trim(keyLit.Value, `"`)
			}

			// Extract the value (function call name)
			var value string
			if callExpr, ok := kv.Value.(*ast.CallExpr); ok {
				if fnIdent, ok := callExpr.Fun.(*ast.Ident); ok {
					value = fnIdent.Name
				}
			}

			if key != "" && value != "" {
				mappings[key] = value
			}
		}
	}
}
