package pkg

import "fmt"

// TerraformEphemeral represents information about a Terraform ephemeral resource
type TerraformEphemeral struct {
	TerraformType      string `json:"terraform_type"`         // "azurerm_key_vault_certificate"
	StructType         string `json:"struct_type"`            // "KeyVaultCertificateEphemeralResource"
	Namespace          string `json:"namespace"`              // "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault"
	RegistrationMethod string `json:"registration_method"`    // "EphemeralResources"
	SDKType            string `json:"sdk_type"`               // "ephemeral"
	SchemaIndex        string `json:"schema_index,omitempty"` // "method.KeyVaultSecretEphemeralResource.Schema.goindex" (optional)
	OpenIndex          string `json:"open_index,omitempty"`   // "method.KeyVaultSecretEphemeralResource.Open.goindex" (optional)
	RenewIndex         string `json:"renew_index,omitempty"`  // "method.KeyVaultSecretEphemeralResource.Renew.goindex" (optional)
	CloseIndex         string `json:"close_index,omitempty"`  // "method.KeyVaultSecretEphemeralResource.Close.goindex" (optional)
}

// NewTerraformEphemeralInfo creates a TerraformEphemeral struct
func NewTerraformEphemeralInfo(structType string, service ServiceRegistration) TerraformEphemeral {
	return TerraformEphemeral{
		TerraformType:      service.EphemeralTerraformTypes[structType],
		StructType:         structType,
		Namespace:          service.PackagePath,
		RegistrationMethod: "EphemeralResources",
		SDKType:            "ephemeral",
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaIndex: fmt.Sprintf("method.%s.Schema.goindex", structType),
		OpenIndex:   fmt.Sprintf("method.%s.Open.goindex", structType),
		RenewIndex:  fmt.Sprintf("method.%s.Renew.goindex", structType),
		CloseIndex:  fmt.Sprintf("method.%s.Close.goindex", structType),
	}
}
