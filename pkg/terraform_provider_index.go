package pkg

import (
	"fmt"
	gophon "github.com/lonegunmanb/gophon/pkg"
	"os"
	"path/filepath"
)

// TerraformProviderIndex represents the complete index of a Terraform provider
type TerraformProviderIndex struct {
	Version    string                `json:"version"`     // Provider version
	Services   []ServiceRegistration `json:"services"`    // All service registrations
	GlobalMaps GlobalMappings        `json:"global_maps"` // Complete mapping across all services
	Statistics ProviderStatistics    `json:"statistics"`  // Summary statistics
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
		packageInfo, err := gophon.ScanSinglePackage(servicePath, basePkgUrl)
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
			supportedResources := extractSupportedResourcesMappings(fileInfo.File)
			supportedDataSources := extractSupportedDataSourcesMappings(fileInfo.File)
			resources := extractResourcesStructTypes(fileInfo.File)
			dataSources := extractDataSourcesStructTypes(fileInfo.File)
			ephemeralResources := extractEphemeralResourcesFunctions(fileInfo.File)

			// Merge results into service registration
			serviceReg.SupportedResources = mergeMap(serviceReg.SupportedResources, supportedResources)
			serviceReg.SupportedDataSources = mergeMap(serviceReg.SupportedDataSources, supportedDataSources)
			serviceReg.Resources = append(serviceReg.Resources, resources...)
			serviceReg.DataSources = append(serviceReg.DataSources, dataSources...)
			serviceReg.EphemeralResources = append(serviceReg.EphemeralResources, ephemeralResources...)
		}

		// After processing all files, extract Terraform types for modern resources and data sources
		serviceReg.ResourceTerraformTypes = extractResourceTerraformTypes(packageInfo, serviceReg.Resources)
		serviceReg.DataSourceTerraformTypes = extractDataSourceTerraformTypes(packageInfo, serviceReg.DataSources)
		serviceReg.EphemeralTerraformTypes = extractEphemeralTerraformTypes(packageInfo, serviceReg.EphemeralResources)

		// Extract CRUD methods for legacy resources using gophon function data
		for terraformType, registrationMethod := range serviceReg.SupportedResources {
			if crudMethods := extractCRUDFromPackage(registrationMethod, packageInfo); crudMethods != nil {
				serviceReg.ResourceCRUDMethods[terraformType] = crudMethods
			}
		}

		// Extract methods for legacy data sources
		for terraformType, registrationMethod := range serviceReg.SupportedDataSources {
			if methods := extractDataSourceMethodsFromPackage(registrationMethod, packageInfo); methods != nil {
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
