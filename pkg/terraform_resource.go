package pkg

import "fmt"

// TerraformResource represents information about a Terraform resource
type TerraformResource struct {
	TerraformType      string `json:"terraform_type"`            // "azurerm_resource_group"
	StructType         string `json:"struct_type"`               // "ResourceGroupResource"
	Namespace          string `json:"namespace"`                 // "github.com/hashicorp/terraform-provider-azurerm/internal/services/resource"
	RegistrationMethod string `json:"registration_method"`       // "SupportedResources", "Resources", etc.
	SDKType            string `json:"sdk_type"`                  // "legacy_pluginsdk", "modern_sdk"
	SchemaIndex        string `json:"schema_index,omitempty"`    // "func.resourceGroup.goindex" or "method.ContainerAppResource.Arguments.goindex" (optional)
	CreateIndex        string `json:"create_index,omitempty"`    // "func.resourceGroupCreateFunc.goindex" or "method.ContainerAppResource.Create.goindex (optional)
	ReadIndex          string `json:"read_index,omitempty"`      // "func.resourceGroupReadFunc.goindex" or "method.ContainerAppResource.Read.goindex" (optional)
	UpdateIndex        string `json:"update_index,omitempty"`    // "func.resourceGroupUpdateFunc.goindex" or "method.ContainerAppResource.Update.goindex" (optional)
	DeleteIndex        string `json:"delete_index,omitempty"`    // "func.resourceGroupDeleteFunc.goindex" or "method.ContainerAppResource.Delete.goindex" (optional)
	AttributeIndex     string `json:"attribute_index,omitempty"` // "func.resourceGroup.goindex" "method.ContainerAppResource.Attributes.goindex"(optional)
}

func NewTerraformResourceInfo(terraformType, structType, registrationMethod, sdkType string, serviceReg ServiceRegistration) TerraformResource {
	if sdkType == "legacy_pluginsdk" {
		result := TerraformResource{
			TerraformType:      terraformType,
			StructType:         "",
			Namespace:          serviceReg.PackagePath,
			RegistrationMethod: registrationMethod,
			SDKType:            sdkType,
			// Optional fields can be added later when we have more sophisticated AST parsing
			SchemaIndex:    fmt.Sprintf("func.%s.goindex", registrationMethod),
			CreateIndex:    "",
			ReadIndex:      "",
			UpdateIndex:    "",
			DeleteIndex:    "",
			AttributeIndex: fmt.Sprintf("func.%s.goindex", registrationMethod),
		}
		// Add CRUD methods if available
		if crudMethods, exists := serviceReg.ResourceCRUDMethods[terraformType]; exists && crudMethods != nil {
			result.CreateIndex = fmt.Sprintf("func.%s.goindex", crudMethods.CreateMethod)
			result.ReadIndex = fmt.Sprintf("func.%s.goindex", crudMethods.ReadMethod)
			result.UpdateIndex = fmt.Sprintf("func.%s.goindex", crudMethods.UpdateMethod)
			result.DeleteIndex = fmt.Sprintf("func.%s.goindex", crudMethods.DeleteMethod)
		}
		return result
	}
	return TerraformResource{
		TerraformType:      serviceReg.ResourceTerraformTypes[structType],
		StructType:         structType,
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "",
		SDKType:            sdkType,
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaIndex:    fmt.Sprintf("method.%s.Arguments.goindex", structType),
		CreateIndex:    fmt.Sprintf("method.%s.Create.goindex", structType),
		ReadIndex:      fmt.Sprintf("method.%s.Read.goindex", structType),
		UpdateIndex:    fmt.Sprintf("method.%s.Update.goindex", structType),
		DeleteIndex:    fmt.Sprintf("method.%s.Delete.goindex", structType),
		AttributeIndex: fmt.Sprintf("method.%s.Attributes.goindex", structType),
	}
}
