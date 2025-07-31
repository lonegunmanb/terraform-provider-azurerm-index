package pkg

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// fileSystem is a global variable for the filesystem interface, allowing us to stub it in tests
var fileSystem afero.Fs = afero.NewOsFs()

// JSONOutputConfig holds configuration for JSON output generation
type JSONOutputConfig struct {
	OutputDir         string
	MainIndexFileName string
	ResourcesSubDir   string
	DataSourcesSubDir string
	EphemeralSubDir   string
	BaseNamespace     string
}

// DefaultJSONOutputConfig returns the default configuration
func DefaultJSONOutputConfig(outputDir string) *JSONOutputConfig {
	return &JSONOutputConfig{
		OutputDir:         outputDir,
		MainIndexFileName: "terraform-provider-azurerm-index.json",
		ResourcesSubDir:   "resources",
		DataSourcesSubDir: "datasources",
		EphemeralSubDir:   "ephemeral",
		BaseNamespace:     "github.com/hashicorp/terraform-provider-azurerm",
	}
}

// FileWriter interface for writing files
type FileWriter interface {
	WriteJSONFile(filePath string, data interface{}) error
	CreateDirectories(dirs []string) error
}

// DefaultFileWriter implements FileWriter using afero
type DefaultFileWriter struct {
	fs afero.Fs
}

// NewDefaultFileWriter creates a new DefaultFileWriter
func NewDefaultFileWriter(fs afero.Fs) *DefaultFileWriter {
	return &DefaultFileWriter{fs: fs}
}

// WriteJSONFile writes data as JSON to the specified file path
func (w *DefaultFileWriter) WriteJSONFile(filePath string, data interface{}) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(filePath)
	if err := w.fs.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}

	// Marshal data to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Write to file
	if err := afero.WriteFile(w.fs, filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// CreateDirectories creates all required directories
func (w *DefaultFileWriter) CreateDirectories(dirs []string) error {
	for _, dir := range dirs {
		if err := w.fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func NewTerraformResourceInfo(terraformType, structType, serviceName, packagePath, registrationMethod, sdkType, baseNamespace string) TerraformResourceInfo {
	namespace := fmt.Sprintf("%s/%s", baseNamespace, packagePath)

	if sdkType == "legacy_pluginsdk" {

		return TerraformResourceInfo{
			TerraformType:      terraformType,
			StructType:         "",
			Namespace:          namespace,
			RegistrationMethod: registrationMethod,
			SDKType:            sdkType,
			// Optional fields can be added later when we have more sophisticated AST parsing
			SchemaMethod:    "",
			CreateMethod:    "",
			ReadMethod:      "",
			UpdateMethod:    "",
			DeleteMethod:    "",
			AttributeMethod: "",
		}
	}
	return TerraformResourceInfo{
		TerraformType:      terraformType,
		StructType:         structType,
		Namespace:          namespace,
		RegistrationMethod: "",
		SDKType:            sdkType,
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaMethod:    fmt.Sprintf("method.%s.Arguments.goindex", structType),
		CreateMethod:    fmt.Sprintf("method.%s.Create.goindex", structType),
		ReadMethod:      fmt.Sprintf("method.%s.Read.goindex", structType),
		UpdateMethod:    fmt.Sprintf("method.%s.Update.goindex", structType),
		DeleteMethod:    fmt.Sprintf("method.%s.Delete.goindex", structType),
		AttributeMethod: fmt.Sprintf("method.%s.Attributes.goindex", structType),
	}
}

// NewTerraformDataSourceInfo creates a TerraformDataSourceInfo struct
func NewTerraformDataSourceInfo(terraformType, serviceName, packagePath, registrationMethod, sdkType, baseNamespace string) TerraformDataSourceInfo {
	namespace := fmt.Sprintf("%s/%s", baseNamespace, packagePath)

	var regMethod string
	var structType string

	if sdkType == "legacy_pluginsdk" {
		regMethod = "SupportedDataSources"
		structType = ""
	} else {
		regMethod = "DataSources"
		structType = terraformType
	}

	return TerraformDataSourceInfo{
		TerraformType:      terraformType,
		StructType:         structType,
		Namespace:          namespace,
		RegistrationMethod: regMethod,
		SDKType:            sdkType,
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaMethod:    "",
		ReadMethod:      "",
		AttributeMethod: "",
	}
}

// NewTerraformEphemeralInfo creates a TerraformEphemeralInfo struct
func NewTerraformEphemeralInfo(functionName, serviceName, packagePath, baseNamespace string) TerraformEphemeralInfo {
	namespace := fmt.Sprintf("%s/%s", baseNamespace, packagePath)

	return TerraformEphemeralInfo{
		TerraformType:      functionName,
		StructType:         "",
		Namespace:          namespace,
		RegistrationMethod: "EphemeralResources",
		SDKType:            "ephemeral",
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaMethod:    "",
		OpenMethod:      "",
		RenewMethod:     "",
		CloseMethod:     "",
		AttributeMethod: "",
	}
}

// JSONOutputGenerator handles the generation of JSON output files
type JSONOutputGenerator struct {
	config     *JSONOutputConfig
	fileWriter FileWriter
}

// NewJSONOutputGenerator creates a new JSONOutputGenerator
func NewJSONOutputGenerator(config *JSONOutputConfig, fileWriter FileWriter) *JSONOutputGenerator {
	return &JSONOutputGenerator{
		config:     config,
		fileWriter: fileWriter,
	}
}

// Generate generates all JSON output files for the given index
func (g *JSONOutputGenerator) Generate(index *TerraformProviderIndex) error {
	// Create directory structure
	if err := g.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Write main index file
	if err := g.writeMainIndexFile(index); err != nil {
		return fmt.Errorf("failed to write main index file: %w", err)
	}

	// Write individual resource files
	if err := g.writeResourceFiles(index); err != nil {
		return fmt.Errorf("failed to write resource files: %w", err)
	}

	// Write individual data source files
	if err := g.writeDataSourceFiles(index); err != nil {
		return fmt.Errorf("failed to write data source files: %w", err)
	}

	// Write individual ephemeral resource files
	if err := g.writeEphemeralFiles(index); err != nil {
		return fmt.Errorf("failed to write ephemeral files: %w", err)
	}

	return nil
}

// createDirectoryStructure creates the required directory structure
func (g *JSONOutputGenerator) createDirectoryStructure() error {
	dirs := []string{
		g.config.OutputDir,
		filepath.Join(g.config.OutputDir, g.config.ResourcesSubDir),
		filepath.Join(g.config.OutputDir, g.config.DataSourcesSubDir),
		filepath.Join(g.config.OutputDir, g.config.EphemeralSubDir),
	}

	return g.fileWriter.CreateDirectories(dirs)
}

// writeMainIndexFile writes the main index file
func (g *JSONOutputGenerator) writeMainIndexFile(index *TerraformProviderIndex) error {
	mainIndexPath := filepath.Join(g.config.OutputDir, g.config.MainIndexFileName)
	return g.fileWriter.WriteJSONFile(mainIndexPath, index)
}

// writeResourceFiles writes individual JSON files for each resource
func (g *JSONOutputGenerator) writeResourceFiles(index *TerraformProviderIndex) error {
	resourcesDir := filepath.Join(g.config.OutputDir, g.config.ResourcesSubDir)

	for _, service := range index.Services {
		// Process legacy resources
		for terraformType, registrationMethod := range service.SupportedResources {
			resourceInfo := NewTerraformResourceInfo(terraformType, "", service.ServiceName, service.PackagePath, registrationMethod, "legacy_pluginsdk", g.config.BaseNamespace)

			// Add CRUD methods if available
			if crudMethods, exists := service.ResourceCRUDMethods[terraformType]; exists && crudMethods != nil {
				resourceInfo.CreateMethod = crudMethods.CreateMethod
				resourceInfo.ReadMethod = crudMethods.ReadMethod
				resourceInfo.UpdateMethod = crudMethods.UpdateMethod
				resourceInfo.DeleteMethod = crudMethods.DeleteMethod
			}

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(resourcesDir, fileName)

			if err := g.fileWriter.WriteJSONFile(filePath, resourceInfo); err != nil {
				return fmt.Errorf("failed to write legacy resource file %s: %w", fileName, err)
			}
		}

		// Process modern resources
		for _, structType := range service.Resources {
			resourceInfo := NewTerraformResourceInfo(structType, structType, service.ServiceName, service.PackagePath, "", "modern_sdk", g.config.BaseNamespace)

			fileName := fmt.Sprintf("%s.json", structType)
			filePath := filepath.Join(resourcesDir, fileName)

			if err := g.fileWriter.WriteJSONFile(filePath, resourceInfo); err != nil {
				return fmt.Errorf("failed to write modern resource file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// writeDataSourceFiles writes individual JSON files for each data source
func (g *JSONOutputGenerator) writeDataSourceFiles(index *TerraformProviderIndex) error {
	dataSourcesDir := filepath.Join(g.config.OutputDir, g.config.DataSourcesSubDir)

	for _, service := range index.Services {
		// Process legacy data sources
		for terraformType, registrationMethod := range service.SupportedDataSources {
			dataSourceInfo := NewTerraformDataSourceInfo(terraformType, service.ServiceName, service.PackagePath, registrationMethod, "legacy_pluginsdk", g.config.BaseNamespace)

			// Add data source methods if available
			if methods, exists := service.DataSourceMethods[terraformType]; exists && methods != nil {
				dataSourceInfo.ReadMethod = methods.ReadMethod
			}

			fileName := fmt.Sprintf("%s.json", terraformType)
			filePath := filepath.Join(dataSourcesDir, fileName)

			if err := g.fileWriter.WriteJSONFile(filePath, dataSourceInfo); err != nil {
				return fmt.Errorf("failed to write legacy data source file %s: %w", fileName, err)
			}
		}

		// Process modern data sources
		for _, structType := range service.DataSources {
			dataSourceInfo := NewTerraformDataSourceInfo(structType, service.ServiceName, service.PackagePath, "", "modern_sdk", g.config.BaseNamespace)

			fileName := fmt.Sprintf("%s.json", structType)
			filePath := filepath.Join(dataSourcesDir, fileName)

			if err := g.fileWriter.WriteJSONFile(filePath, dataSourceInfo); err != nil {
				return fmt.Errorf("failed to write modern data source file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

// writeEphemeralFiles writes individual JSON files for each ephemeral resource
func (g *JSONOutputGenerator) writeEphemeralFiles(index *TerraformProviderIndex) error {
	ephemeralDir := filepath.Join(g.config.OutputDir, g.config.EphemeralSubDir)

	for _, service := range index.Services {
		for _, functionName := range service.EphemeralResources {
			ephemeralInfo := NewTerraformEphemeralInfo(functionName, service.ServiceName, service.PackagePath, g.config.BaseNamespace)

			fileName := fmt.Sprintf("%s.json", functionName)
			filePath := filepath.Join(ephemeralDir, fileName)

			if err := g.fileWriter.WriteJSONFile(filePath, ephemeralInfo); err != nil {
				return fmt.Errorf("failed to write ephemeral resource file %s: %w", fileName, err)
			}
		}
	}

	return nil
}

func (index *TerraformProviderIndex) GenerateJSONOutput(outputDir string) error {
	config := DefaultJSONOutputConfig(outputDir)
	fileWriter := NewDefaultFileWriter(fileSystem)
	generator := NewJSONOutputGenerator(config, fileWriter)

	return generator.Generate(index)
}

// Legacy functions for backward compatibility
// writeJSONFile is kept for backward compatibility with existing tests
func writeJSONFile(filePath string, data interface{}) error {
	fileWriter := NewDefaultFileWriter(fileSystem)
	return fileWriter.WriteJSONFile(filePath, data)
}

// createResourceInfo is kept for backward compatibility with existing tests
func createResourceInfo(terraformType, serviceName, packagePath, registrationMethod, sdkType string) TerraformResourceInfo {
	namespace := fmt.Sprintf("github.com/hashicorp/terraform-provider-azurerm/%s", packagePath)

	if sdkType == "legacy_pluginsdk" {
		return TerraformResourceInfo{
			TerraformType:      terraformType,
			StructType:         "",
			Namespace:          namespace,
			RegistrationMethod: "SupportedResources", // Old behavior: always "SupportedResources" for legacy
			SDKType:            sdkType,
			SchemaMethod:       "",
			CreateMethod:       "",
			ReadMethod:         "",
			UpdateMethod:       "",
			DeleteMethod:       "",
			AttributeMethod:    "",
		}
	} else {
		return TerraformResourceInfo{
			TerraformType:      terraformType,
			StructType:         terraformType, // Old behavior: StructType = terraformType for modern
			Namespace:          namespace,
			RegistrationMethod: "Resources", // Old behavior: always "Resources" for modern
			SDKType:            sdkType,
			SchemaMethod:       "", // Old behavior: empty method fields
			CreateMethod:       "",
			ReadMethod:         "",
			UpdateMethod:       "",
			DeleteMethod:       "",
			AttributeMethod:    "",
		}
	}
}

// createDataSourceInfo is kept for backward compatibility with existing tests
func createDataSourceInfo(terraformType, serviceName, packagePath, registrationMethod, sdkType string) TerraformDataSourceInfo {
	return NewTerraformDataSourceInfo(terraformType, serviceName, packagePath, registrationMethod, sdkType, "github.com/hashicorp/terraform-provider-azurerm")
}

// createEphemeralInfo is kept for backward compatibility with existing tests
func createEphemeralInfo(functionName, serviceName, packagePath string) TerraformEphemeralInfo {
	return NewTerraformEphemeralInfo(functionName, serviceName, packagePath, "github.com/hashicorp/terraform-provider-azurerm")
}
