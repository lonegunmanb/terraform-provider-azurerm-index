package pkg

import (
	"fmt"
	"github.com/lonegunmanb/gophon/pkg"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// TerraformResourceInfo represents information about a Terraform resource
type TerraformResourceInfo struct {
	TerraformType      string `json:"terraform_type"`             // "azurerm_resource_group"
	StructType         string `json:"struct_type"`                // "ResourceGroupResource"
	Namespace          string `json:"namespace"`                  // "github.com/hashicorp/terraform-provider-azurerm/internal/services/resource"
	RegistrationMethod string `json:"registration_method"`        // "SupportedResources", "Resources", etc.
	SDKType            string `json:"sdk_type"`                   // "legacy_pluginsdk", "modern_sdk"
	SchemaMethod       string `json:"schema_method,omitempty"`    // "resourceGroupSchema" (optional)
	CreateMethod       string `json:"create_method,omitempty"`    // "resourceGroupCreateFunc" (optional)
	ReadMethod         string `json:"read_method,omitempty"`      // "resourceGroupReadFunc" (optional)
	UpdateMethod       string `json:"update_method,omitempty"`    // "resourceGroupUpdateFunc" (optional)
	DeleteMethod       string `json:"delete_method,omitempty"`    // "resourceGroupDeleteFunc" (optional)
	AttributeMethod    string `json:"attribute_method,omitempty"` // "resourceGroupAttributes" (optional)
}

// TerraformDataSourceInfo represents information about a Terraform data source
type TerraformDataSourceInfo struct {
	TerraformType      string `json:"terraform_type"`             // "azurerm_client_config"
	StructType         string `json:"struct_type"`                // "ClientConfigDataSource"
	Namespace          string `json:"namespace"`                  // "github.com/hashicorp/terraform-provider-azurerm/internal/services/client"
	RegistrationMethod string `json:"registration_method"`        // "SupportedDataSources", "DataSources", etc.
	SDKType            string `json:"sdk_type"`                   // "legacy_pluginsdk", "modern_sdk"
	SchemaMethod       string `json:"schema_method,omitempty"`    // "dataSourceArmClientConfigSchema" (optional)
	ReadMethod         string `json:"read_method,omitempty"`      // "dataSourceArmClientConfigRead" (optional)
	AttributeMethod    string `json:"attribute_method,omitempty"` // "dataSourceArmClientConfigAttributes" (optional)
}

// TerraformEphemeralInfo represents information about a Terraform ephemeral resource
type TerraformEphemeralInfo struct {
	TerraformType      string `json:"terraform_type"`             // "azurerm_key_vault_certificate"
	StructType         string `json:"struct_type"`                // "KeyVaultCertificateEphemeralResource"
	Namespace          string `json:"namespace"`                  // "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault"
	RegistrationMethod string `json:"registration_method"`        // "EphemeralResources"
	SDKType            string `json:"sdk_type"`                   // "ephemeral"
	SchemaMethod       string `json:"schema_method,omitempty"`    // "keyVaultCertificateEphemeralSchema" (optional)
	OpenMethod         string `json:"open_method,omitempty"`      // "keyVaultCertificateEphemeralOpen" (optional)
	RenewMethod        string `json:"renew_method,omitempty"`     // "keyVaultCertificateEphemeralRenew" (optional)
	CloseMethod        string `json:"close_method,omitempty"`     // "keyVaultCertificateEphemeralClose" (optional)
	AttributeMethod    string `json:"attribute_method,omitempty"` // "keyVaultCertificateEphemeralAttributes" (optional)
}

// TerraformProviderIndex represents the complete index of a Terraform provider
type TerraformProviderIndex struct {
	Version    string                `json:"version"`     // Provider version
	Services   []ServiceRegistration `json:"services"`    // All service registrations
	GlobalMaps GlobalMappings        `json:"global_maps"` // Complete mapping across all services
	Statistics ProviderStatistics    `json:"statistics"`  // Summary statistics
}

// ServiceRegistration represents all registration methods found in a single service package
type ServiceRegistration struct {
	Package              *pkg.PackageInfo
	ServiceName          string                                `json:"service_name"`           // "keyvault", "resource", etc.
	PackagePath          string                                `json:"package_path"`           // "internal/services/keyvault"
	SupportedResources   map[string]string                     `json:"supported_resources"`    // Legacy map-based resources
	SupportedDataSources map[string]string                     `json:"supported_data_sources"` // Legacy map-based data sources
	Resources            []string                              `json:"resources"`              // Modern slice-based resources
	DataSources          []string                              `json:"data_sources"`           // Modern slice-based data sources
	EphemeralResources   []string                              `json:"ephemeral_resources"`    // Function-based ephemeral resources
	ResourceCRUDMethods  map[string]*LegacyResourceCRUDMethods `json:"resource_crud_methods"`  // CRUD methods for legacy resources
	DataSourceMethods    map[string]*LegacyDataSourceMethods   `json:"data_source_methods"`    // Methods for legacy data sources
}

// GlobalMappings represents complete mappings across all services
type GlobalMappings struct {
	AllDataSources map[string]string `json:"all_data_sources"` // Complete mapping across all services
	AllResources   map[string]string `json:"all_resources"`    // Complete mapping across all services
}

// ProviderStatistics represents summary statistics for the provider
type ProviderStatistics struct {
	ServiceCount       int `json:"service_count"`
	TotalDataSources   int `json:"total_data_sources"`
	TotalResources     int `json:"total_resources"`
	LegacyResources    int `json:"legacy_resources"`
	ModernResources    int `json:"modern_resources"`
	EphemeralResources int `json:"ephemeral_resources"`
}

// TerraformResourceMapping represents a mapping between terraform resource type and its registration method
type TerraformResourceMapping struct {
	TerraformType      string `json:"terraform_type"`      // e.g., "azurerm_resource_group"
	RegistrationMethod string `json:"registration_method"` // e.g., "resourceResourceGroup"
}

// LegacyResourceCRUDMethods represents CRUD methods extracted from legacy plugin SDK resources
type LegacyResourceCRUDMethods struct {
	CreateMethod string `json:"create_method,omitempty"` // "keyVaultCreateFunc"
	ReadMethod   string `json:"read_method,omitempty"`   // "keyVaultReadFunc"
	UpdateMethod string `json:"update_method,omitempty"` // "keyVaultUpdateFunc"
	DeleteMethod string `json:"delete_method,omitempty"` // "keyVaultDeleteFunc"
}

// LegacyDataSourceMethods represents methods extracted from legacy plugin SDK data sources
type LegacyDataSourceMethods struct {
	ReadMethod string `json:"read_method,omitempty"` // "dataSourceReadFunc"
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

// ExtractEphemeralResourcesFunctions extracts function names from EphemeralResources method in the AST
func ExtractEphemeralResourcesFunctions(node *ast.File) []string {
	return extractFunctionNamesFromMethod(node, "EphemeralResources")
}

// ExtractLegacyResourceCRUDMethods analyzes a legacy plugin SDK resource function
// and extracts CRUD method names from the returned pluginsdk.Resource struct
// The input ast.File should contain the registration function's source code
// It will find any function that returns *pluginsdk.Resource and parse its CRUD methods
func ExtractLegacyResourceCRUDMethods(node *ast.File) (*LegacyResourceCRUDMethods, error) {
	// Find the resource function that returns *pluginsdk.Resource
	resourceFunc := findResourceFunction(node)
	if resourceFunc == nil {
		return &LegacyResourceCRUDMethods{}, nil // No resource function found, return empty
	}

	// Extract CRUD methods from the function body
	crudMethods := extractCRUDFromFunction(resourceFunc)
	return crudMethods, nil
}

// ScanTerraformProviderServices scans the specified directory for Terraform provider services
// and extracts all registration information into a structured index
func ScanTerraformProviderServices(dir, basePkgUrl string, version string) (*TerraformProviderIndex, error) {

	// Read the services directory to get all service subdirectories
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read services directory: %w", err)
	}

	var services []ServiceRegistration
	globalResources := make(map[string]string)
	globalDataSources := make(map[string]string)

	stats := ProviderStatistics{}

	// Iterate through each service directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		servicePath := filepath.Join(dir, entry.Name())

		// Scan the individual service package
		packageInfo, err := pkg.ScanSinglePackage(servicePath, basePkgUrl)
		if err != nil || packageInfo == nil || len(packageInfo.Files) == 0 {
			// Skip services that can't be scanned (might not be valid Go packages)
			continue
		}

		serviceReg := newServiceRegistration(packageInfo, entry)

		// Process each file in the package
		for _, fileInfo := range packageInfo.Files {
			if fileInfo.File == nil {
				continue
			}

			// Extract all registration methods from this file
			supportedResources := ExtractSupportedResourcesMappings(fileInfo.File)
			supportedDataSources := ExtractSupportedDataSourcesMappings(fileInfo.File)
			resources := ExtractResourcesStructTypes(fileInfo.File)
			dataSources := ExtractDataSourcesStructTypes(fileInfo.File)
			ephemeralResources := ExtractEphemeralResourcesFunctions(fileInfo.File)

			// Merge results into service registration
			serviceReg.SupportedResources = mergeMap(serviceReg.SupportedResources, supportedResources)
			serviceReg.SupportedDataSources = mergeMap(serviceReg.SupportedDataSources, supportedDataSources)
			serviceReg.Resources = append(serviceReg.Resources, resources...)
			serviceReg.DataSources = append(serviceReg.DataSources, dataSources...)
			serviceReg.EphemeralResources = append(serviceReg.EphemeralResources, ephemeralResources...)
		}

		// After processing all files, extract CRUD methods for legacy resources using gophon function data
		for terraformType, registrationMethod := range serviceReg.SupportedResources {
			if crudMethods := extractCRUDFromPackage(registrationMethod, packageInfo); crudMethods != nil {
				serviceReg.ResourceCRUDMethods[terraformType] = crudMethods
			}
		}

		// Extract methods for legacy data sources
		for terraformType, registrationMethod := range serviceReg.SupportedDataSources {
			if methods := extractDataSourceMethodsFromGophonPackage(registrationMethod, packageInfo); methods != nil {
				serviceReg.DataSourceMethods[terraformType] = methods
			}
		}

		// Only include services that have at least one registration method
		if len(serviceReg.SupportedResources) > 0 || len(serviceReg.SupportedDataSources) > 0 ||
			len(serviceReg.Resources) > 0 || len(serviceReg.DataSources) > 0 || len(serviceReg.EphemeralResources) > 0 {
			services = append(services, serviceReg)
			stats.ServiceCount++

			// Add to global maps
			globalResources = mergeMap(globalResources, serviceReg.SupportedResources)
			globalDataSources = mergeMap(globalDataSources, serviceReg.SupportedDataSources)

			// Update statistics
			stats.LegacyResources += len(serviceReg.SupportedResources)
			stats.TotalDataSources += len(serviceReg.SupportedDataSources)
			stats.ModernResources += len(serviceReg.Resources)
			stats.TotalDataSources += len(serviceReg.DataSources)
			stats.EphemeralResources += len(serviceReg.EphemeralResources)
		}
	}

	stats.TotalResources = stats.LegacyResources + stats.ModernResources + stats.EphemeralResources

	return &TerraformProviderIndex{
		Version:  version,
		Services: services,
		GlobalMaps: GlobalMappings{
			AllResources:   globalResources,
			AllDataSources: globalDataSources,
		},
		Statistics: stats,
	}, nil
}

func newServiceRegistration(packageInfo *pkg.PackageInfo, entry os.DirEntry) ServiceRegistration {
	return ServiceRegistration{
		Package:              packageInfo,
		ServiceName:          entry.Name(),
		PackagePath:          packageInfo.Files[0].Package,
		SupportedResources:   make(map[string]string),
		SupportedDataSources: make(map[string]string),
		Resources:            []string{},
		DataSources:          []string{},
		EphemeralResources:   []string{},
		ResourceCRUDMethods:  make(map[string]*LegacyResourceCRUDMethods),
		DataSourceMethods:    make(map[string]*LegacyDataSourceMethods),
	}
}

// GenerateIndividualResourceFiles generates individual JSON files for each resource and data source
func GenerateIndividualResourceFiles(index *TerraformProviderIndex) (map[string]TerraformResourceInfo, map[string]TerraformDataSourceInfo, map[string]TerraformEphemeralInfo) {
	resources := make(map[string]TerraformResourceInfo)
	dataSources := make(map[string]TerraformDataSourceInfo)
	ephemeralResources := make(map[string]TerraformEphemeralInfo)

	for _, service := range index.Services {
		namespace := fmt.Sprintf("github.com/hashicorp/terraform-provider-azurerm/%s", service.PackagePath)

		// Process legacy resources (SupportedResources)
		for terraformType, registrationMethod := range service.SupportedResources {
			resources[terraformType] = TerraformResourceInfo{
				TerraformType:      terraformType,
				StructType:         "", // Would need additional parsing to determine
				Namespace:          namespace,
				RegistrationMethod: registrationMethod,
				SDKType:            "legacy_pluginsdk",
			}
		}

		// Process legacy data sources (SupportedDataSources)
		for terraformType, registrationMethod := range service.SupportedDataSources {
			dataSources[terraformType] = TerraformDataSourceInfo{
				TerraformType:      terraformType,
				StructType:         "", // Would need additional parsing to determine
				Namespace:          namespace,
				RegistrationMethod: registrationMethod,
				SDKType:            "legacy_pluginsdk",
			}
		}

		// Process modern resources (Resources)
		for _, structType := range service.Resources {
			// For modern resources, we'd need to map struct types back to terraform types
			// This would require additional AST parsing of ResourceType() methods
			terraformType := fmt.Sprintf("azurerm_%s", strings.ToLower(structType)) // Placeholder
			resources[terraformType] = TerraformResourceInfo{
				TerraformType:      terraformType,
				StructType:         structType,
				Namespace:          namespace,
				RegistrationMethod: "Resources",
				SDKType:            "modern_sdk",
			}
		}

		// Process modern data sources (DataSources)
		for _, structType := range service.DataSources {
			// For modern data sources, we'd need to map struct types back to terraform types
			terraformType := fmt.Sprintf("azurerm_%s", strings.ToLower(structType)) // Placeholder
			dataSources[terraformType] = TerraformDataSourceInfo{
				TerraformType:      terraformType,
				StructType:         structType,
				Namespace:          namespace,
				RegistrationMethod: "DataSources",
				SDKType:            "modern_sdk",
			}
		}

		// Process ephemeral resources (EphemeralResources)
		for _, functionName := range service.EphemeralResources {
			// For ephemeral resources, we'd need to map function names to terraform types
			terraformType := fmt.Sprintf("azurerm_%s", strings.ToLower(functionName)) // Placeholder
			ephemeralResources[terraformType] = TerraformEphemeralInfo{
				TerraformType:      terraformType,
				StructType:         "", // Would need additional parsing to determine
				Namespace:          namespace,
				RegistrationMethod: functionName,
				SDKType:            "ephemeral",
			}
		}
	}

	return resources, dataSources, ephemeralResources
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

// extractFunctionNamesFromMethod extracts function names from any method that returns []func() ephemeral.EphemeralResource
func extractFunctionNamesFromMethod(node *ast.File, methodName string) []string {
	var functions []string

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
					functions = append(functions, extractFromFunctionSliceLiteral(sliceLit)...)
				}

				// Handle variable reference (like "ephemeralResources" variable)
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
							functions = append(functions, extractFromFunctionSliceLiteral(sliceLit)...)
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return functions
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

// extractFromFunctionSliceLiteral extracts function names from a slice literal
func extractFromFunctionSliceLiteral(sliceLit *ast.CompositeLit) []string {
	var functions []string
	for _, elt := range sliceLit.Elts {
		// Handle function identifiers like FuncName (without parentheses)
		if ident, ok := elt.(*ast.Ident); ok {
			functions = append(functions, ident.Name)
		}
	}
	return functions
}

// findResourceFunction locates any function declaration that returns *pluginsdk.Resource
func findResourceFunction(node *ast.File) *ast.FuncDecl {
	var resourceFunc *ast.FuncDecl

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Type.Results == nil {
			return true
		}

		// Check if function returns *pluginsdk.Resource
		for _, result := range fn.Type.Results.List {
			if starExpr, ok := result.Type.(*ast.StarExpr); ok {
				if selectorExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
					if ident, ok := selectorExpr.X.(*ast.Ident); ok {
						if ident.Name == "pluginsdk" && selectorExpr.Sel.Name == "Resource" {
							resourceFunc = fn
							return false // Found it, stop searching
						}
					}
				}
			}
		}
		return true
	})

	return resourceFunc
}

// extractCRUDFromFunction extracts CRUD method names from a resource function body
func extractCRUDFromFunction(fn *ast.FuncDecl) *LegacyResourceCRUDMethods {
	methods := &LegacyResourceCRUDMethods{}

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
				extractFromResourceLiteral(compLit, methods)
			}
		}

		return true
	})

	// Also look for variable assignments in case of Pattern 2 (variable assignment)
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
				extractFromResourceLiteral(compLit, methods)
			}
		}

		return true
	})

	return methods
}

// extractFromResourceLiteral parses a pluginsdk.Resource composite literal
// and extracts only CRUD method names
func extractFromResourceLiteral(compLit *ast.CompositeLit, methods *LegacyResourceCRUDMethods) {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Get the field name
		var fieldName string
		if ident, ok := kv.Key.(*ast.Ident); ok {
			fieldName = ident.Name
		}

		// Extract function reference from the value
		funcName := extractFunctionReference(kv.Value)
		if funcName == "" {
			continue
		}

		// Map field names to CRUD methods
		switch fieldName {
		case "CreateContext", "CreateFunc", "CreateWithoutTimeout":
			methods.CreateMethod = funcName
		case "ReadContext", "ReadFunc", "ReadWithoutTimeout":
			methods.ReadMethod = funcName
		case "UpdateContext", "UpdateFunc", "UpdateWithoutTimeout":
			methods.UpdateMethod = funcName
		case "DeleteContext", "DeleteFunc", "DeleteWithoutTimeout":
			methods.DeleteMethod = funcName
		}
	}
}

// extractFunctionReference extracts function name from various AST patterns:
// - Direct identifier: funcName
// - Selector expression: package.FuncName
func extractFunctionReference(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// Direct identifier: funcName
		return e.Name
	case *ast.SelectorExpr:
		// Selector expression: package.FuncName
		return e.Sel.Name
	default:
		return ""
	}
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

// extractCRUDFromPackage extracts CRUD methods from a gophon PackageInfo by finding the registration function
func extractCRUDFromPackage(registrationMethod string, packageInfo *pkg.PackageInfo) *LegacyResourceCRUDMethods {
	if packageInfo == nil || packageInfo.Functions == nil {
		return nil
	}

	// Find the registration function in the gophon function data
	for _, funcInfo := range packageInfo.Functions {
		if funcInfo.Name == registrationMethod && funcInfo.FuncDecl != nil {
			return extractCRUDFromFunction(funcInfo.FuncDecl)
		}
	}

	return nil
}

// extractDataSourceMethodsFromGophonPackage extracts data source methods from a gophon PackageInfo
func extractDataSourceMethodsFromGophonPackage(registrationMethod string, packageInfo *pkg.PackageInfo) *LegacyDataSourceMethods {
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

// extractFromDataSourceLiteral parses a pluginsdk.Resource composite literal for data sources
// and extracts only read method names
func extractFromDataSourceLiteral(compLit *ast.CompositeLit, methods *LegacyDataSourceMethods) {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Get the field name
		var fieldName string
		if ident, ok := kv.Key.(*ast.Ident); ok {
			fieldName = ident.Name
		}

		// Extract function reference from the value
		funcName := extractFunctionReference(kv.Value)
		if funcName == "" {
			continue
		}

		// Map field names to data source methods (only ReadContext/ReadFunc for data sources)
		switch fieldName {
		case "ReadContext", "ReadFunc", "ReadWithoutTimeout":
			methods.ReadMethod = funcName
		}
	}
}
