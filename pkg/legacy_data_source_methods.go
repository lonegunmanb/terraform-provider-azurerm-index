package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"go/ast"
	"go/token"
)

type LegacyDataSourceMethods struct {
	ReadMethod string `json:"read_method,omitempty"` // "dataSourceReadFunc"
}

func extractDataSourceMethodsFromPackage(registrationMethod string, packageInfo *gophon.PackageInfo) *LegacyDataSourceMethods {
	if packageInfo == nil || packageInfo.Functions == nil {
		return nil
	}

	// Find the registration function in the gophon function data
	for _, funcInfo := range packageInfo.Functions {
		if funcInfo.Name == registrationMethod && funcInfo.FuncDecl != nil {
			// Extract data source methods from the function declaration
			return extractDataSourceMethodsFromFunction(funcInfo.FuncDecl)
		}
	}

	return nil
}

// extractDataSourceMethodsFromFunction extracts data source methods from a data source function body
func extractDataSourceMethodsFromFunction(fn *ast.FuncDecl) *LegacyDataSourceMethods {
	methods := &LegacyDataSourceMethods{}

	if fn.Body == nil {
		return methods
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// Look for return statements
		returnStmt, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		// Process each return expression
		for _, result := range returnStmt.Results {
			unaryExpr, ok := result.(*ast.UnaryExpr)
			// Handle direct return of composite literal
			if !ok || unaryExpr.Op != token.AND {
				return true
			}
			if compLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
				extractFromDataSourceLiteral(compLit, methods)
			}
		}

		return true
	})

	// Also look for variable assignments in case of variable assignment pattern
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assignStmt, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for _, rhs := range assignStmt.Rhs {
			unaryExpr, ok := rhs.(*ast.UnaryExpr)
			if !ok || unaryExpr.Op != token.AND {
				return true
			}
			if compLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
				extractFromDataSourceLiteral(compLit, methods)
			}
		}

		return true
	})

	return methods
}
