package pkg

import (
	"encoding/json"
	"fmt"
	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

var outputFs = afero.NewOsFs()

// TerraformProviderIndex represents the complete index of a Terraform provider
type TerraformProviderIndex struct {
	Version    string                `json:"version"`    // Provider version
	Services   []ServiceRegistration `json:"services"`   // All service registrations
	Statistics ProviderStatistics    `json:"statistics"` // Summary statistics
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
		Version:    version,
		Services:   services,
		Statistics: stats,
	}, nil
}

// WriteIndexFiles writes all index files to the specified output directory
// This is the main method that orchestrates writing all index files
func (index *TerraformProviderIndex) WriteIndexFiles(outputDir string) error {
	// Create directory structure
	if err := index.CreateDirectoryStructure(outputDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Write main index file
	if err := index.WriteMainIndexFile(outputDir); err != nil {
		return fmt.Errorf("failed to write main index file: %w", err)
	}

	// Write individual resource files
	if err := index.WriteResourceFiles(outputDir); err != nil {
		return fmt.Errorf("failed to write resource files: %w", err)
	}

	// Write individual data source files
	if err := index.WriteDataSourceFiles(outputDir); err != nil {
		return fmt.Errorf("failed to write data source files: %w", err)
	}

	// Write individual ephemeral resource files
	if err := index.WriteEphemeralFiles(outputDir); err != nil {
		return fmt.Errorf("failed to write ephemeral files: %w", err)
	}

	return nil
}

// WriteMainIndexFile writes the main terraform-provider-azurerm-index.json file
func (index *TerraformProviderIndex) WriteMainIndexFile(outputDir string) error {
	mainIndexPath := filepath.Join(outputDir, "terraform-provider-azurerm-index.json")
	return index.WriteJSONFile(mainIndexPath, index)
}

// WriteResourceFiles writes individual JSON files for each resource
func (index *TerraformProviderIndex) WriteResourceFiles(outputDir string) error {
	resourcesDir := filepath.Join(outputDir, "resources")

	for _, service := range index.Services {
		// Process legacy resources
		for terraformType, registrationMethod := range service.SupportedResources {
			resourceInfo := NewTerraformResourceInfo(terraformType, "", registrationMethod, "legacy_pluginsdk", service)

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(resourcesDir, fileName)

			if err := index.WriteJSONFile(filePath, resourceInfo); err != nil {
				return fmt.Errorf("failed to write legacy resource file %s: %w", fileName, err)
			}
		}

		// Process modern resources
		for _, structType := range service.Resources {
			// Get the actual Terraform type from the mapping
			terraformType, exists := service.ResourceTerraformTypes[structType]
			if !exists {
				// Fallback to struct type if mapping doesn't exist
				terraformType = structType
			}

			resourceInfo := NewTerraformResourceInfo(terraformType, structType, "", "modern_sdk", service)

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(resourcesDir, fileName)

			if err := index.WriteJSONFile(filePath, resourceInfo); err != nil {
				return fmt.Errorf("failed to write modern resource file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// WriteDataSourceFiles writes individual JSON files for each data source
func (index *TerraformProviderIndex) WriteDataSourceFiles(outputDir string) error {
	dataSourcesDir := filepath.Join(outputDir, "datasources")

	for _, service := range index.Services {
		// Process legacy data sources
		for terraformType, registrationMethod := range service.SupportedDataSources {
			dataSourceInfo := NewTerraformDataSourceInfo(terraformType, "", registrationMethod, "legacy_pluginsdk", service)

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(dataSourcesDir, fileName)

			if err := index.WriteJSONFile(filePath, dataSourceInfo); err != nil {
				return fmt.Errorf("failed to write legacy data source file %s: %w", fileName, err)
			}
		}

		// Process modern data sources
		for _, structType := range service.DataSources {
			// Get the actual Terraform type from the mapping
			terraformType, exists := service.DataSourceTerraformTypes[structType]
			if !exists {
				// Fallback to struct type if mapping doesn't exist
				terraformType = structType
			}

			dataSourceInfo := NewTerraformDataSourceInfo(terraformType, structType, "", "modern_sdk", service)

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(dataSourcesDir, fileName)

			if err := index.WriteJSONFile(filePath, dataSourceInfo); err != nil {
				return fmt.Errorf("failed to write modern data source file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// WriteEphemeralFiles writes individual JSON files for each ephemeral resource
func (index *TerraformProviderIndex) WriteEphemeralFiles(outputDir string) error {
	ephemeralDir := filepath.Join(outputDir, "ephemeral")

	for _, service := range index.Services {
		for _, structType := range service.EphemeralResources {
			// Get the actual Terraform type from the mapping
			terraformType, exists := service.EphemeralTerraformTypes[structType]
			if !exists {
				// Fallback to struct type if mapping doesn't exist
				terraformType = structType
			}

			ephemeralInfo := NewTerraformEphemeralInfo(structType, service)

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(ephemeralDir, fileName)

			if err := index.WriteJSONFile(filePath, ephemeralInfo); err != nil {
				return fmt.Errorf("failed to write ephemeral resource file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// CreateDirectoryStructure creates the required directory structure for index files
func (index *TerraformProviderIndex) CreateDirectoryStructure(outputDir string) error {
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "resources"),
		filepath.Join(outputDir, "datasources"),
		filepath.Join(outputDir, "ephemeral"),
	}

	for _, dir := range dirs {
		if err := outputFs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// WriteJSONFile writes data as JSON to the specified file path
func (index *TerraformProviderIndex) WriteJSONFile(filePath string, data interface{}) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(filePath)
	if err := outputFs.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}

	// Marshal data to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Write to file
	if err := afero.WriteFile(outputFs, filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
