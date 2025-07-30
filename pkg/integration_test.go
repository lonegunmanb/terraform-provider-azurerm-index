package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestScanTerraformProviderServices(t *testing.T) {
	// Use test harness instead of actual terraform provider source
	testHarnessPath := filepath.Join("testharness", "internal", "services")

	// Scan the test harness services
	index, err := ScanTerraformProviderServices(testHarnessPath, "github.com/lonegunmanb/terraform-provider-azurerm-index", "test-version")
	require.NoError(t, err)
	require.NotNil(t, index)
	for _, svc := range index.Services {
		assert.Equal(t, fmt.Sprintf("github.com/lonegunmanb/terraform-provider-azurerm-index/testharness/internal/services/%s", svc.ServiceName), svc.PackagePath)
	}

	// Basic validation
	assert.Equal(t, "test-version", index.Version)
	assert.NotEmpty(t, index.Services, "Should find at least some services")
	assert.Greater(t, index.Statistics.ServiceCount, 0, "Should have at least one service")

	// Should find our test services
	serviceNames := make(map[string]bool)
	for _, service := range index.Services {
		serviceNames[service.ServiceName] = true
	}
	assert.Contains(t, serviceNames, "keyvault", "Should find keyvault service")
	assert.Contains(t, serviceNames, "resource", "Should find resource service")
	assert.Contains(t, serviceNames, "compute", "Should find compute service")
	assert.Contains(t, serviceNames, "storage", "Should find storage service")

	// Validate keyvault service specifically (has all registration types)
	var keyvaultService *ServiceRegistration
	for _, service := range index.Services {
		if service.ServiceName == "keyvault" {
			keyvaultService = &service
			break
		}
	}
	require.NotNil(t, keyvaultService, "Should find keyvault service")

	// Keyvault should have all registration types
	assert.NotEmpty(t, keyvaultService.SupportedResources, "Keyvault should have legacy resources")
	assert.NotEmpty(t, keyvaultService.SupportedDataSources, "Keyvault should have legacy data sources")
	assert.NotEmpty(t, keyvaultService.Resources, "Keyvault should have modern resources")
	assert.NotEmpty(t, keyvaultService.DataSources, "Keyvault should have modern data sources")
	assert.NotEmpty(t, keyvaultService.EphemeralResources, "Keyvault should have ephemeral resources")

	expectedSupportedResources := map[string]string{
		"azurerm_key_vault":               "resourceKeyVault",
		"azurerm_key_vault_certificate":   "resourceKeyVaultCertificate",
		"azurerm_key_vault_access_policy": "resourceKeyVaultAccessPolicy",
	}
	expectedSupportedDataSources := map[string]string{
		"azurerm_key_vault":             "dataSourceKeyVault",
		"azurerm_key_vault_certificate": "dataSourceKeyVaultCertificate",
		"azurerm_key_vault_secret":      "dataSourceKeyVaultSecret",
	}
	// Check specific expected values
	assert.Equal(t, expectedSupportedResources, keyvaultService.SupportedResources)
	assert.Equal(t, expectedSupportedDataSources, keyvaultService.SupportedDataSources)
	assert.Contains(t, keyvaultService.Resources, "KeyVaultCertificateContactsResource")
	assert.Contains(t, keyvaultService.DataSources, "EncryptedValueDataSource")
	assert.Contains(t, keyvaultService.EphemeralResources, "NewKeyVaultCertificateEphemeralResource")
	assert.Contains(t, keyvaultService.EphemeralResources, "NewKeyVaultSecretEphemeralResource")

	// Validate global maps are populated
	assert.NotEmpty(t, index.GlobalMaps.AllResources, "Should have global resource mappings")
	assert.NotEmpty(t, index.GlobalMaps.AllDataSources, "Should have global data source mappings")

	// Check for expected resources in global maps
	assert.Contains(t, index.GlobalMaps.AllResources, "azurerm_key_vault")
	assert.Contains(t, index.GlobalMaps.AllDataSources, "azurerm_key_vault")

	// Validate statistics make sense
	assert.Greater(t, index.Statistics.TotalResources, 0)
	assert.Greater(t, index.Statistics.TotalDataSources, 0)
	assert.Greater(t, index.Statistics.LegacyResources, 0)
	assert.Greater(t, index.Statistics.ModernResources, 0)
	assert.Greater(t, index.Statistics.EphemeralResources, 0)
}
