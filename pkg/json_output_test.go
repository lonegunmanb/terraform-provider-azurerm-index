package pkg

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJSONOutput(t *testing.T) {
	// Setup mock filesystem
	mockFS := afero.NewMemMapFs()

	// Stub the filesystem variable
	stubs := gostub.Stub(&fileSystem, mockFS)
	defer stubs.Reset()

	// Create test data
	testIndex := &TerraformProviderIndex{
		Version: "v4.0.0",
		Services: []ServiceRegistration{
			{
				ServiceName: "keyvault",
				PackagePath: "internal/services/keyvault",
				SupportedResources: map[string]string{
					"azurerm_key_vault":             "resourceKeyVault",
					"azurerm_key_vault_certificate": "resourceKeyVaultCertificate",
				},
				SupportedDataSources: map[string]string{
					"azurerm_key_vault": "dataSourceKeyVault",
				},
				Resources: []string{
					"KeyVaultAccessPolicyResource",
					"KeyVaultCertificateContactsResource",
				},
				DataSources: []string{
					"KeyVaultDataSource",
				},
				EphemeralResources: []string{
					"NewKeyVaultCertificateEphemeralResource",
					"NewKeyVaultSecretEphemeralResource",
				},
			},
			{
				ServiceName: "resource",
				PackagePath: "internal/services/resource",
				SupportedResources: map[string]string{
					"azurerm_resource_group": "resourceResourceGroup",
				},
				SupportedDataSources: map[string]string{
					"azurerm_resource_group": "dataSourceResourceGroup",
				},
				Resources:          []string{"ResourceGroupResource"},
				DataSources:        []string{"ResourceGroupDataSource"},
				EphemeralResources: []string{},
			},
		},
		GlobalMaps: GlobalMappings{
			AllResources: map[string]string{
				"azurerm_key_vault":             "resourceKeyVault",
				"azurerm_key_vault_certificate": "resourceKeyVaultCertificate",
				"azurerm_resource_group":        "resourceResourceGroup",
			},
			AllDataSources: map[string]string{
				"azurerm_key_vault":      "dataSourceKeyVault",
				"azurerm_resource_group": "dataSourceResourceGroup",
			},
		},
		Statistics: ProviderStatistics{
			ServiceCount:       2,
			TotalResources:     5, // 3 legacy + 2 modern
			TotalDataSources:   3, // 2 legacy + 1 modern
			LegacyResources:    3,
			ModernResources:    2,
			EphemeralResources: 2,
		},
	}

	outputDir := "/test-output"

	// Execute
	err := testIndex.GenerateJSONOutput(outputDir)
	require.NoError(t, err)

	// Verify directory structure was created
	t.Run("DirectoryStructure", func(t *testing.T) {
		exists, err := afero.DirExists(mockFS, outputDir)
		require.NoError(t, err)
		assert.True(t, exists, "Output directory should exist")

		exists, err = afero.DirExists(mockFS, filepath.Join(outputDir, "resources"))
		require.NoError(t, err)
		assert.True(t, exists, "Resources directory should exist")

		exists, err = afero.DirExists(mockFS, filepath.Join(outputDir, "datasources"))
		require.NoError(t, err)
		assert.True(t, exists, "Datasources directory should exist")

		exists, err = afero.DirExists(mockFS, filepath.Join(outputDir, "ephemeral"))
		require.NoError(t, err)
		assert.True(t, exists, "Ephemeral directory should exist")
	})

	// Verify main index file
	t.Run("MainIndexFile", func(t *testing.T) {
		indexPath := filepath.Join(outputDir, "terraform-provider-azurerm-index.json")
		exists, err := afero.Exists(mockFS, indexPath)
		require.NoError(t, err)
		assert.True(t, exists, "Main index file should exist")

		content, err := afero.ReadFile(mockFS, indexPath)
		require.NoError(t, err)

		var loadedIndex TerraformProviderIndex
		err = json.Unmarshal(content, &loadedIndex)
		require.NoError(t, err)

		assert.Equal(t, testIndex.Version, loadedIndex.Version)
		assert.Equal(t, testIndex.Statistics, loadedIndex.Statistics)
		assert.Equal(t, len(testIndex.Services), len(loadedIndex.Services))
	})

	// Verify individual resource files
	t.Run("ResourceFiles", func(t *testing.T) {
		expectedResources := []struct {
			fileName      string
			terraformType string
			structType    string
			sdkType       string
		}{
			{"azurerm_key_vault.json", "azurerm_key_vault", "", "legacy_pluginsdk"},
			{"azurerm_key_vault_certificate.json", "azurerm_key_vault_certificate", "", "legacy_pluginsdk"},
			{"azurerm_resource_group.json", "azurerm_resource_group", "", "legacy_pluginsdk"},
		}

		for _, expected := range expectedResources {
			resourcePath := filepath.Join(outputDir, "resources", expected.fileName)
			exists, err := afero.Exists(mockFS, resourcePath)
			require.NoError(t, err, "Resource file %s should exist", expected.fileName)
			assert.True(t, exists)

			content, err := afero.ReadFile(mockFS, resourcePath)
			require.NoError(t, err)

			var resourceInfo TerraformResourceInfo
			err = json.Unmarshal(content, &resourceInfo)
			require.NoError(t, err)

			assert.Equal(t, expected.terraformType, resourceInfo.TerraformType)
			assert.Equal(t, expected.sdkType, resourceInfo.SDKType)
		}
	})

	// Verify individual data source files
	t.Run("DataSourceFiles", func(t *testing.T) {
		expectedDataSources := []struct {
			fileName      string
			terraformType string
			sdkType       string
		}{
			{"azurerm_key_vault.json", "azurerm_key_vault", "legacy_pluginsdk"},
			{"azurerm_resource_group.json", "azurerm_resource_group", "legacy_pluginsdk"},
		}

		for _, expected := range expectedDataSources {
			dataSourcePath := filepath.Join(outputDir, "datasources", expected.fileName)
			exists, err := afero.Exists(mockFS, dataSourcePath)
			require.NoError(t, err, "Data source file %s should exist", expected.fileName)
			assert.True(t, exists)

			content, err := afero.ReadFile(mockFS, dataSourcePath)
			require.NoError(t, err)

			var dataSourceInfo TerraformDataSourceInfo
			err = json.Unmarshal(content, &dataSourceInfo)
			require.NoError(t, err)

			assert.Equal(t, expected.terraformType, dataSourceInfo.TerraformType)
			assert.Equal(t, expected.sdkType, dataSourceInfo.SDKType)
		}
	})

	// Verify ephemeral resource files
	t.Run("EphemeralFiles", func(t *testing.T) {
		expectedEphemeral := []struct {
			fileName      string
			terraformType string
			structType    string
		}{
			{"NewKeyVaultCertificateEphemeralResource.json", "NewKeyVaultCertificateEphemeralResource", ""},
			{"NewKeyVaultSecretEphemeralResource.json", "NewKeyVaultSecretEphemeralResource", ""},
		}

		for _, expected := range expectedEphemeral {
			ephemeralPath := filepath.Join(outputDir, "ephemeral", expected.fileName)
			exists, err := afero.Exists(mockFS, ephemeralPath)
			require.NoError(t, err, "Ephemeral file %s should exist", expected.fileName)
			assert.True(t, exists)

			content, err := afero.ReadFile(mockFS, ephemeralPath)
			require.NoError(t, err)

			var ephemeralInfo TerraformEphemeralInfo
			err = json.Unmarshal(content, &ephemeralInfo)
			require.NoError(t, err)

			assert.Equal(t, expected.terraformType, ephemeralInfo.TerraformType)
			assert.Equal(t, "ephemeral", ephemeralInfo.SDKType)
		}
	})
}

func TestGenerateJSONOutputErrorHandling(t *testing.T) {
	// Setup mock filesystem that denies write access
	mockFS := afero.NewReadOnlyFs(afero.NewMemMapFs())

	// Stub the filesystem variable
	stubs := gostub.Stub(&fileSystem, mockFS)
	defer stubs.Reset()

	testIndex := &TerraformProviderIndex{
		Version:  "v4.0.0",
		Services: []ServiceRegistration{},
		GlobalMaps: GlobalMappings{
			AllResources:   map[string]string{},
			AllDataSources: map[string]string{},
		},
		Statistics: ProviderStatistics{},
	}

	// Execute - should fail due to read-only filesystem
	err := testIndex.GenerateJSONOutput("/test-output")
	assert.Error(t, err, "Should fail on read-only filesystem")
}

func TestWriteJSONFile(t *testing.T) {
	// Setup mock filesystem
	mockFS := afero.NewMemMapFs()

	// Stub the filesystem variable
	stubs := gostub.Stub(&fileSystem, mockFS)
	defer stubs.Reset()

	testData := map[string]interface{}{
		"test_field": "test_value",
		"number":     42,
	}

	// Test successful write
	err := writeJSONFile("/test/file.json", testData)
	require.NoError(t, err)

	// Verify file exists and content is correct
	exists, err := afero.Exists(mockFS, "/test/file.json")
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := afero.ReadFile(mockFS, "/test/file.json")
	require.NoError(t, err)

	var loadedData map[string]interface{}
	err = json.Unmarshal(content, &loadedData)
	require.NoError(t, err)

	assert.Equal(t, "test_value", loadedData["test_field"])
	assert.Equal(t, float64(42), loadedData["number"]) // JSON unmarshals numbers as float64
}

func TestCreateResourceInfo(t *testing.T) {
	testCases := []struct {
		name               string
		terraformType      string
		serviceName        string
		packagePath        string
		registrationMethod string
		sdkType            string
		expected           TerraformResourceInfo
	}{
		{
			name:               "Legacy Resource",
			terraformType:      "azurerm_key_vault",
			serviceName:        "keyvault",
			packagePath:        "internal/services/keyvault",
			registrationMethod: "resourceKeyVault",
			sdkType:            "legacy_pluginsdk",
			expected: TerraformResourceInfo{
				TerraformType:      "azurerm_key_vault",
				StructType:         "",
				Namespace:          "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
				RegistrationMethod: "SupportedResources",
				SDKType:            "legacy_pluginsdk",
			},
		},
		{
			name:               "Modern Resource",
			terraformType:      "KeyVaultResource",
			serviceName:        "keyvault",
			packagePath:        "internal/services/keyvault",
			registrationMethod: "",
			sdkType:            "modern_sdk",
			expected: TerraformResourceInfo{
				TerraformType:      "KeyVaultResource",
				StructType:         "KeyVaultResource",
				Namespace:          "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
				RegistrationMethod: "Resources",
				SDKType:            "modern_sdk",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := createResourceInfo(tc.terraformType, tc.serviceName, tc.packagePath, tc.registrationMethod, tc.sdkType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateDataSourceInfo(t *testing.T) {
	testCases := []struct {
		name               string
		terraformType      string
		serviceName        string
		packagePath        string
		registrationMethod string
		sdkType            string
		expected           TerraformDataSourceInfo
	}{
		{
			name:               "Legacy Data Source",
			terraformType:      "azurerm_key_vault",
			serviceName:        "keyvault",
			packagePath:        "internal/services/keyvault",
			registrationMethod: "dataSourceKeyVault",
			sdkType:            "legacy_pluginsdk",
			expected: TerraformDataSourceInfo{
				TerraformType:      "azurerm_key_vault",
				StructType:         "",
				Namespace:          "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
				RegistrationMethod: "SupportedDataSources",
				SDKType:            "legacy_pluginsdk",
			},
		},
		{
			name:               "Modern Data Source",
			terraformType:      "KeyVaultDataSource",
			serviceName:        "keyvault",
			packagePath:        "internal/services/keyvault",
			registrationMethod: "",
			sdkType:            "modern_sdk",
			expected: TerraformDataSourceInfo{
				TerraformType:      "KeyVaultDataSource",
				StructType:         "KeyVaultDataSource",
				Namespace:          "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
				RegistrationMethod: "DataSources",
				SDKType:            "modern_sdk",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := createDataSourceInfo(tc.terraformType, tc.serviceName, tc.packagePath, tc.registrationMethod, tc.sdkType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateEphemeralInfo(t *testing.T) {
	result := createEphemeralInfo("NewKeyVaultCertificateEphemeralResource", "keyvault", "internal/services/keyvault")

	expected := TerraformEphemeralInfo{
		TerraformType:      "NewKeyVaultCertificateEphemeralResource",
		StructType:         "",
		Namespace:          "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
		RegistrationMethod: "EphemeralResources",
		SDKType:            "ephemeral",
	}

	assert.Equal(t, expected, result)
}
