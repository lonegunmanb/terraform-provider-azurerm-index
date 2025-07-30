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
