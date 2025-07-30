package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseSource(source string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, "", source, parser.ParseComments)
}

func TestExtractSupportedResourcesMappings(t *testing.T) {
	// Test case with the exact example provided
	source := `package resource

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_management_lock":                      resourceManagementLock(),
		"azurerm_management_group_template_deployment": managementGroupTemplateDeploymentResource(),
		"azurerm_resource_group":                       resourceResourceGroup(),
		"azurerm_resource_group_template_deployment":   resourceGroupTemplateDeploymentResource(),
		"azurerm_subscription_template_deployment":     subscriptionTemplateDeploymentResource(),
		"azurerm_tenant_template_deployment":           tenantTemplateDeploymentResource(),
	}

	return resources
}`

	expected := map[string]string{
		"azurerm_management_lock":                      "resourceManagementLock",
		"azurerm_management_group_template_deployment": "managementGroupTemplateDeploymentResource",
		"azurerm_resource_group":                       "resourceResourceGroup",
		"azurerm_resource_group_template_deployment":   "resourceGroupTemplateDeploymentResource",
		"azurerm_subscription_template_deployment":     "subscriptionTemplateDeploymentResource",
		"azurerm_tenant_template_deployment":           "tenantTemplateDeploymentResource",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractSupportedResourcesMappings(node)
	assert.Equal(t, expected, result)
}

func TestExtractSupportedResourcesDirectReturn(t *testing.T) {
	// Test case with direct map return (no intermediate variable)
	source := `package resource

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault":        resourceKeyVault(),
		"azurerm_key_vault_secret": resourceKeyVaultSecret(),
	}
}`

	expected := map[string]string{
		"azurerm_key_vault":        "resourceKeyVault",
		"azurerm_key_vault_secret": "resourceKeyVaultSecret",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractSupportedResourcesMappings(node)

	assert.Equal(t, expected, result)
}

func TestExtractSupportedResourcesEmptyMethod(t *testing.T) {
	// Test case with empty method
	source := `package resource

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{}
}`

	node, err := parseSource(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	result := ExtractSupportedResourcesMappings(node)

	assert.Empty(t, result)
}

func TestExtractSupportedResourcesNoMethod(t *testing.T) {
	// Test case with no SupportedResources method
	source := `package resource

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault": dataSourceKeyVault(),
	}
}`

	node, err := parseSource(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	result := ExtractSupportedResourcesMappings(node)

	assert.Empty(t, result)
}

func TestExtractSupportedDataSourcesMappings(t *testing.T) {
	// Test case based on the actual keyvault service example
	source := `package keyvault

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault_access_policy":      dataSourceKeyVaultAccessPolicy(),
		"azurerm_key_vault_certificate":        dataSourceKeyVaultCertificate(),
		"azurerm_key_vault_certificate_data":   dataSourceKeyVaultCertificateData(),
		"azurerm_key_vault_certificate_issuer": dataSourceKeyVaultCertificateIssuer(),
		"azurerm_key_vault_key":                dataSourceKeyVaultKey(),
		"azurerm_key_vault_secret":             dataSourceKeyVaultSecret(),
		"azurerm_key_vault_secrets":            dataSourceKeyVaultSecrets(),
		"azurerm_key_vault":                    dataSourceKeyVault(),
		"azurerm_key_vault_certificates":       dataSourceKeyVaultCertificates(),
	}
}`

	expected := map[string]string{
		"azurerm_key_vault_access_policy":      "dataSourceKeyVaultAccessPolicy",
		"azurerm_key_vault_certificate":        "dataSourceKeyVaultCertificate",
		"azurerm_key_vault_certificate_data":   "dataSourceKeyVaultCertificateData",
		"azurerm_key_vault_certificate_issuer": "dataSourceKeyVaultCertificateIssuer",
		"azurerm_key_vault_key":                "dataSourceKeyVaultKey",
		"azurerm_key_vault_secret":             "dataSourceKeyVaultSecret",
		"azurerm_key_vault_secrets":            "dataSourceKeyVaultSecrets",
		"azurerm_key_vault":                    "dataSourceKeyVault",
		"azurerm_key_vault_certificates":       "dataSourceKeyVaultCertificates",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractSupportedDataSourcesMappings(node)
	assert.Equal(t, expected, result)
}

func TestExtractSupportedDataSourcesWithVariable(t *testing.T) {
	// Test case with intermediate variable (like SupportedResources pattern)
	source := `package resource

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	dataSources := map[string]*pluginsdk.Resource{
		"azurerm_client_config":    dataSourceArmClientConfig(),
		"azurerm_resource_group":   dataSourceArmResourceGroup(),
		"azurerm_subscription":     dataSourceArmSubscription(),
	}
	return dataSources
}`

	expected := map[string]string{
		"azurerm_client_config":  "dataSourceArmClientConfig",
		"azurerm_resource_group": "dataSourceArmResourceGroup",
		"azurerm_subscription":   "dataSourceArmSubscription",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractSupportedDataSourcesMappings(node)
	assert.Equal(t, expected, result)
}

func TestExtractSupportedDataSourcesEmpty(t *testing.T) {
	// Test case with empty SupportedDataSources method
	source := `package resource

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractSupportedDataSourcesMappings(node)
	assert.Empty(t, result)
}

func TestExtractSupportedDataSourcesNoMethod(t *testing.T) {
	// Test case with no SupportedDataSources method
	source := `package resource

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_resource_group": resourceResourceGroup(),
	}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractSupportedDataSourcesMappings(node)
	assert.Empty(t, result)
}

func TestExtractDataSourcesStructTypes(t *testing.T) {
	// Test case based on the actual keyvault service example
	source := `package keyvault

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		EncryptedValueDataSource{},
	}
}`

	expected := []string{"EncryptedValueDataSource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractDataSourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractDataSourcesStructTypesMultiple(t *testing.T) {
	// Test case with multiple struct types
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		FirstDataSource{},
		SecondDataSource{},
		ThirdDataSource{},
	}
}`

	expected := []string{"FirstDataSource", "SecondDataSource", "ThirdDataSource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractDataSourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractDataSourcesStructTypesWithVariable(t *testing.T) {
	// Test case with intermediate variable
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	dataSources := []sdk.DataSource{
		ConfigDataSource{},
		InfoDataSource{},
	}
	return dataSources
}`

	expected := []string{"ConfigDataSource", "InfoDataSource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractDataSourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractDataSourcesStructTypesEmpty(t *testing.T) {
	// Test case with empty DataSources method
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractDataSourcesStructTypes(node)
	assert.Empty(t, result)
}

func TestExtractDataSourcesStructTypesNoMethod(t *testing.T) {
	// Test case with no DataSources method
	source := `package service

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		SomeResource{},
	}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := ExtractDataSourcesStructTypes(node)
	assert.Empty(t, result)
}
