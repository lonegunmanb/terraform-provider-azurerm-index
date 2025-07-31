package pkg

import (
	pkg2 "github.com/lonegunmanb/gophon/pkg"
	"os"
)

// ServiceRegistration represents all registration methods found in a single service package
type ServiceRegistration struct {
	Package              *pkg2.PackageInfo
	ServiceName          string                                `json:"service_name"`           // "keyvault", "resource", etc.
	PackagePath          string                                `json:"package_path"`           // "internal/services/keyvault"
	SupportedResources   map[string]string                     `json:"supported_resources"`    // Legacy map-based resources
	SupportedDataSources map[string]string                     `json:"supported_data_sources"` // Legacy map-based data sources
	Resources            []string                              `json:"resources"`              // Modern slice-based resources
	DataSources          []string                              `json:"data_sources"`           // Modern slice-based data sources
	EphemeralResources   []string                              `json:"ephemeral_resources"`    // Function-based ephemeral resources
	ResourceCRUDMethods  map[string]*LegacyResourceCRUDMethods `json:"resource_crud_methods"`  // CRUD methods for legacy resources
	DataSourceMethods    map[string]*LegacyDataSourceMethods   `json:"data_source_methods"`    // Methods for legacy data sources
	// New mappings between Terraform types and struct types
	ResourceTerraformTypes    map[string]string `json:"resource_terraform_types"`     // StructType -> TerraformType for modern resources
	DataSourceTerraformTypes  map[string]string `json:"data_source_terraform_types"`  // StructType -> TerraformType for modern data sources
	EphemeralTerraformTypes   map[string]string `json:"ephemeral_terraform_types"`    // StructType -> TerraformType for ephemeral resources
}

func newServiceRegistration(packageInfo *pkg2.PackageInfo, entry os.DirEntry) ServiceRegistration {
	return ServiceRegistration{
		Package:                   packageInfo,
		ServiceName:               entry.Name(),
		PackagePath:               packageInfo.Files[0].Package,
		SupportedResources:        make(map[string]string),
		SupportedDataSources:      make(map[string]string),
		Resources:                 []string{},
		DataSources:               []string{},
		EphemeralResources:        []string{},
		ResourceCRUDMethods:       make(map[string]*LegacyResourceCRUDMethods),
		DataSourceMethods:         make(map[string]*LegacyDataSourceMethods),
		ResourceTerraformTypes:    make(map[string]string),
		DataSourceTerraformTypes:  make(map[string]string),
		EphemeralTerraformTypes:   make(map[string]string),
	}
}
