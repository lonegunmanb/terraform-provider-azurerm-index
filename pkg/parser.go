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
	return extractMappingsFromMethod(node, "SupportedResources")
}

// ExtractSupportedDataSourcesMappings extracts mappings from SupportedDataSources method in the AST
func ExtractSupportedDataSourcesMappings(node *ast.File) map[string]string {
	return extractMappingsFromMethod(node, "SupportedDataSources")
}

// ExtractDataSourcesStructTypes extracts struct type names from DataSources method in the AST
func ExtractDataSourcesStructTypes(node *ast.File) []string {
	return extractStructTypesFromMethod(node, "DataSources")
}

// ExtractResourcesStructTypes extracts struct type names from Resources method in the AST
func ExtractResourcesStructTypes(node *ast.File) []string {
	return extractStructTypesFromMethod(node, "Resources")
}

// extractMappingsFromMethod extracts mappings from any method that returns map[string]*pluginsdk.Resource
func extractMappingsFromMethod(node *ast.File, methodName string) map[string]string {
	mappings := make(map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != methodName {
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
					mappings = mergeMap(mappings, extractFromMapLiteral(mapLit))
				}

				// Handle variable reference (like "resources" or "dataSources" variable)
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
							mappings = mergeMap(mappings, extractFromMapLiteral(mapLit))
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

// extractStructTypesFromMethod extracts struct type names from any method that returns []sdk.DataSource or []sdk.Resource
func extractStructTypesFromMethod(node *ast.File, methodName string) []string {
	var types []string

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != methodName {
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
				// Handle direct slice literal return
				if sliceLit, ok := result.(*ast.CompositeLit); ok {
					types = append(types, extractFromSliceLiteral(sliceLit)...)
				}

				// Handle variable reference (like "dataSources" variable)
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
						if sliceLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit); ok {
							types = append(types, extractFromSliceLiteral(sliceLit)...)
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return types
}

// extractFromMapLiteral extracts key-value pairs from a map literal
func extractFromMapLiteral(mapLit *ast.CompositeLit) map[string]string {
	mappings := make(map[string]string)
	for _, elt := range mapLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		// Extract the key (terraform resource type)
		var key string
		if keyLit, ok := kv.Key.(*ast.BasicLit); ok && keyLit.Kind == token.STRING {
			key = strings.Trim(keyLit.Value, `"`)
		}

		// Extract the value (function call name)
		var value string
		callExpr, ok := kv.Value.(*ast.CallExpr)
		if !ok {
			continue
		}
		if fnIdent, ok := callExpr.Fun.(*ast.Ident); ok {
			value = fnIdent.Name
		}

		if key != "" && value != "" {
			mappings[key] = value
		}
	}
	return mappings
}

// extractFromSliceLiteral extracts struct type names from a slice literal
func extractFromSliceLiteral(sliceLit *ast.CompositeLit) []string {
	var types []string
	for _, elt := range sliceLit.Elts {
		// Handle struct literals like StructName{}
		compLit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}

		// Extract the struct type name
		if ident, ok := compLit.Type.(*ast.Ident); ok {
			types = append(types, ident.Name)
		}
	}
	return types
}

func mergeMap[TK comparable, TV any](m1, m2 map[TK]TV) map[TK]TV {
	m := make(map[TK]TV)
	for tk, tv := range m1 {
		m[tk] = tv
	}
	for tk, tv := range m2 {
		m[tk] = tv
	}
	return m
}
