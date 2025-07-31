package pkg

// TerraformResourceMapping represents a mapping between terraform resource type and its registration method
type TerraformResourceMapping struct {
	TerraformType      string `json:"terraform_type"`      // e.g., "azurerm_resource_group"
	RegistrationMethod string `json:"registration_method"` // e.g., "resourceResourceGroup"
}
