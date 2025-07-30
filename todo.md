# TODO: Terraform Provider AzureRM Index Generation Plan

## Overview
Adapt the existing GitHub Actions workflow to monitor the `hashicorp/terraform-provider-azurerm` repository for new releases, and automatically generate golang indexes using `gophon` when new versions are available.

## Current State Analysis
- **Existing workflow**: Currently configured for `google-gemini/gemini-cli` with Docker build/push
- **Target**: Need to adapt for `hashicorp/terraform-provider-azurerm` with golang indexing
- **Storage**: Index files will be stored in `index/` folder instead of Docker images
- **Tool**: Use `gophon` to generate `.goindex` files for all Go symbols

## Required Changes

### 1. Update Version Detection Step
- **Current**: Checks `repos/google-gemini/gemini-cli/tags`
- **Change to**: Check `repos/hashicorp/terraform-provider-azurerm/tags`
- **Logic**: If manual version is provided via workflow input, use that version; otherwise, fetch latest tag
- **Filter**: Keep the semver regex pattern `^[vV]?[0-9]+\.[0-9]+\.[0-9]+$`
- **Note**: Terraform provider uses `vx.x.x` format

### 2. Remove Docker-Related Steps
- **Remove**: QEMU setup
- **Remove**: Docker Buildx setup  
- **Remove**: DockerHub login
- **Remove**: GHCR login
- **Remove**: Docker build and push step
- **Remove**: All Docker secrets dependencies

### 3. Add Go Environment Setup
- **Add**: Go setup action (use git hash of latest tag for `actions/setup-go@v4`)
- **Version**: Use Go 1.21+ (required by gophon)
- **Add**: Install gophon tool

### 4. Add Source Code Checkout for Target Repo
- **Add**: Checkout terraform-provider-azurerm at the specific tag
- **Location**: Use a `./tmp` folder (don't commit provider's source code)
- **Method**: Either git clone or use actions/checkout with repository parameter

### 5. Clean Previous Index Files
- **Add**: Remove current `index` folder if it exists
- **Command**: `rm -rf ./index` or equivalent
- **Purpose**: Prepare clean slate for new index files from the new version

### 6. Add Gophon Index Generation
- **Working Dir**: `terraform-provider-azurerm`'s folder
- **Command**: `gophon -base=internal -dest=<path-to-this-repo>/index`
- **Input**: The checked out terraform-provider-azurerm source
- **Output**: `.goindex` files in `./index` directory
- **Note**: May need to handle large codebase - consider if need to limit to specific packages

### 7. Commit and Push Index Files
- **Add**: Git configuration (already exists)
- **Add**: Add all files in `index/` directory
- **Add**: Commit with meaningful message including version
- **Add**: Push the changes to main branch
- **Keep**: Tag creation and push (already exists)

### 8. Update Workflow Metadata
- **Name**: Change from "cronjob" to something like "terraform-azurerm-index"
- **Schedule**: Keep the `0 */6 * * *` (every 6 hours)
- **Permissions**: May need to adjust if different permissions needed

## Implementation Steps

### Phase 1: Core Workflow Adaptation
1. Update repository reference in version check
2. Remove all Docker-related steps
3. Add Go setup and gophon installation
4. Add terraform-provider-azurerm checkout

### Phase 2: Index Generation
1. Implement gophon execution
2. Test with a specific version first
3. Handle potential errors (large codebase, memory issues)

### Phase 3: Git Operations
1. Add index files to git
2. Commit and push changes
3. Ensure tag creation works correctly

### Phase 4: Testing & Optimization
1. Test with manual workflow dispatch
2. Verify index files are generated correctly
3. Check that AI agents can access the indexes properly
4. Optimize for performance if needed

## Potential Challenges

### 1. Large Codebase
- **Issue**: terraform-provider-azurerm is very large
- **Solution**: May need to limit indexing to specific packages or use pagination
- **Consideration**: Monitor GitHub Actions execution time limits

### 2. Memory Usage
- **Issue**: Gophon may use significant memory on large codebases
- **Solution**: Consider running on larger GitHub runner if needed

### 3. Storage Space
- **Issue**: Large number of index files
- **Solution**: Monitor repository size, may need cleanup of old versions

### 4. Rate Limits
- **Issue**: GitHub API rate limits for checking tags
- **Solution**: Current implementation should be fine with 6-hour intervals

## File Structure After Implementation
```
terraform-provider-azurerm-index/
├── .github/
│   └── workflows/
│       └── index.yaml (updated)
├── index/                    (new)
│   ├── func.NewSomething.goindex
│   ├── type.SomeType.goindex
│   ├── method.Service.Method.goindex
│   └── ... (many more .goindex files)
├── todo.md                   (this file)
└── README.md                 (to be created)
```

## Go AST Parsing Analysis for Terraform Provider Indexing

Based on sample code analysis from [AST parsing gist](https://gist.githubusercontent.com/lonegunmanb/d518cdc08ec636b922fc0f24c757e825/raw/0538b5e38fc0e511536f96e50fcf833d78e57470/ast.go), we can implement custom Go AST scanning instead of or in addition to `gophon` for more targeted indexing.

### Key Insights from Sample Code

The sample demonstrates three distinct AST parsing approaches for Terraform providers:

#### 1. Resource Type to Schema Function Mapping
Extracts mappings from provider registration methods like `SupportedResources()` and `SupportedDataSources()`:

```go
// Example mapping extracted:
map[string]string{
    "azurerm_availability_set": "availabilitySetResource",
    "azurerm_capacity_reservation": "capacityReservationResource",
}
```

**Key Implementation Pattern:**
```go
func extractMappingsFromMethod(node *ast.File, methodName string, variableName string) map[string]string {
    mappings := make(map[string]string)
    ast.Inspect(node, func(n ast.Node) bool {
        fn, ok := n.(*ast.FuncDecl)
        if !ok || fn.Name.Name != methodName {
            return true
        }
        // Extract map literal key-value pairs
        return extractFromReturn(inner, mappings)
    })
    return mappings
}
```

#### 2. Type Extraction from Registration Methods
Extracts struct types from `DataSources()` and `Resources()` method return slices:

```go
// Extracts type names like: []string{"DataSourceType1", "ResourceType1", ...}
func extractTypesFromMethod(node *ast.File, methodName string) []string {
    var types []string
    ast.Inspect(node, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == methodName {
            // Look for return statements with slice literals
            extractTypesFromReturn(inner, &types)
        }
        return true
    })
    return types
}
```

#### 3. ResourceType Method Implementation Mapping
Maps struct types to their resource type strings by analyzing `ResourceType()` method implementations:

```go
// Example: AvailabilitySetResource -> "azurerm_availability_set"
func extractResourceTypeMethods(node *ast.File) map[string]string {
    mappings := make(map[string]string)
    ast.Inspect(node, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "ResourceType" {
            receiverType := extractReceiverType(fn)      // "AvailabilitySetResource"
            resourceType := extractReturnedString(fn)    // "azurerm_availability_set"
            mappings[receiverType] = resourceType
        }
        return true
    })
    return mappings
}
```

### Proposed Custom Indexing Strategy

Instead of only using `gophon`, we could implement a hybrid approach:

#### Phase 1: AST-Based Terraform-Specific Indexing
1. **Comprehensive Package Scanning**: Use `gophon.ScanPackagesRecursively()` to discover all packages under `internal/services`
2. **Multi-File Analysis**: Parse all Go files in each package for any registration methods they may contain
3. **Selective Method Extraction**: Extract whichever registration method types are present in each file (most files will have none or only a few)
4. **Cross-Service Aggregation**: Combine mappings from all services into global indexes

#### Phase 2: Enhanced Index Generation
Create individual JSON files for each terraform resource/data source, similar to gophon's approach:

```
index/
├── resources/
│   ├── azurerm_resource_group.json               # Individual resource mapping
│   ├── azurerm_key_vault.json                    # Contains entry method, struct type, etc.
│   ├── azurerm_virtual_machine.json              # Quick lookup for specific resource
│   └── ... (one file per resource)
├── datasources/
│   ├── azurerm_resource_group.json               # Individual data source mapping
│   ├── azurerm_key_vault.json                    # Contains entry method, struct type, etc.
│   ├── azurerm_client_config.json                # Quick lookup for specific data source
│   └── ... (one file per data source)
├── ephemeral/
│   ├── azurerm_key_vault_certificate.json        # Individual ephemeral resource mapping
│   └── ... (one file per ephemeral resource)
└── func.*.goindex                                # gophon-generated function indexes (if hybrid)
```

**Example `resources/azurerm_resource_group.json`:**
```json
{
  "terraform_type": "azurerm_resource_group",
  "struct_type": "ResourceGroupResource",
  "registration_method": "resourceResourceGroup",
  "sdk_type": "legacy_pluginsdk",
  "schema_method": "resourceGroupSchema",
  "create_method": "resourceGroupCreateFunc",
  "read_method": "resourceGroupReadFunc", 
  "update_method": "resourceGroupUpdateFunc",
  "delete_method": "resourceGroupDeleteFunc",
  "attribute_method": "resourceGroupAttributes"
}
```

**Example `datasources/azurerm_client_config.json`:**
```json
{
  "terraform_type": "azurerm_client_config",
  "struct_type": "ClientConfigDataSource",
  "registration_method": "SupportedDataSources",
  "sdk_type": "legacy_pluginsdk",
  "schema_method": "dataSourceArmClientConfigSchema",
  "read_method": "dataSourceArmClientConfigRead"
}
```

**Example `ephemeral/azurerm_key_vault_certificate.json`:**
```json
{
  "terraform_type": "azurerm_key_vault_certificate",
  "struct_type": "KeyVaultCertificateEphemeralResource",
  "registration_method": "EphemeralResources",
  "sdk_type": "ephemeral",
  "schema_method": "keyVaultCertificateEphemeralSchema",
  "open_method": "keyVaultCertificateEphemeralOpen",
  "renew_method": "keyVaultCertificateEphemeralRenew",
  "close_method": "keyVaultCertificateEphemeralClose",
  "attribute_method": "keyVaultCertificateEphemeralAttributes"
}
```

#### Phase 3: Implementation Code Structure
```go
type TerraformResourceInfo struct {
    TerraformType      string `json:"terraform_type"`               // "azurerm_resource_group"
    StructType         string `json:"struct_type"`                  // "ResourceGroupResource"
    RegistrationMethod string `json:"registration_method"`          // "SupportedResources", "Resources", etc.
    SDKType            string `json:"sdk_type"`                     // "legacy_pluginsdk", "modern_sdk"
    SchemaMethod       string `json:"schema_method,omitempty"`      // "resourceGroupSchema" (optional)
    CreateMethod       string `json:"create_method,omitempty"`      // "resourceGroupCreateFunc" (optional)
    ReadMethod         string `json:"read_method,omitempty"`        // "resourceGroupReadFunc" (optional)
    UpdateMethod       string `json:"update_method,omitempty"`      // "resourceGroupUpdateFunc" (optional)
    DeleteMethod       string `json:"delete_method,omitempty"`      // "resourceGroupDeleteFunc" (optional)
    AttributeMethod    string `json:"attribute_method,omitempty"`   // "resourceGroupAttributes" (optional)
}

type TerraformDataSourceInfo struct {
    TerraformType      string `json:"terraform_type"`               // "azurerm_client_config"
    StructType         string `json:"struct_type"`                  // "ClientConfigDataSource"
    RegistrationMethod string `json:"registration_method"`          // "SupportedDataSources", "DataSources", etc.
    SDKType            string `json:"sdk_type"`                     // "legacy_pluginsdk", "modern_sdk"
    SchemaMethod       string `json:"schema_method,omitempty"`      // "dataSourceArmClientConfigSchema" (optional)
    ReadMethod         string `json:"read_method,omitempty"`        // "dataSourceArmClientConfigRead" (optional)
    AttributeMethod    string `json:"attribute_method,omitempty"`   // "dataSourceArmClientConfigAttributes" (optional)
}

type TerraformEphemeralInfo struct {
    TerraformType      string `json:"terraform_type"`               // "azurerm_key_vault_certificate"
    StructType         string `json:"struct_type"`                  // "KeyVaultCertificateEphemeralResource"
    RegistrationMethod string `json:"registration_method"`          // "EphemeralResources"
    SDKType            string `json:"sdk_type"`                     // "ephemeral"
    SchemaMethod       string `json:"schema_method,omitempty"`      // "keyVaultCertificateEphemeralSchema" (optional)
    OpenMethod         string `json:"open_method,omitempty"`        // "keyVaultCertificateEphemeralOpen" (optional)
    RenewMethod        string `json:"renew_method,omitempty"`       // "keyVaultCertificateEphemeralRenew" (optional)
    CloseMethod        string `json:"close_method,omitempty"`       // "keyVaultCertificateEphemeralClose" (optional)
    AttributeMethod    string `json:"attribute_method,omitempty"`   // "keyVaultCertificateEphemeralAttributes" (optional)
}
```

### Current Terraform Provider AzureRM Structure Analysis

Based on the cloned source code in `./tmp/terraform-provider-azurerm`, the provider has:

- **130+ service packages** in `internal/services/` directory
- **Modern registration pattern** with multiple registration methods per service
- **Mixed architecture**: Legacy pluginsdk + modern framework + ephemeral resources

#### Registration Method Types Found
From analyzing `keyvault/registration.go`, each service implements:

```go
// Legacy SDK methods (return map[string]*pluginsdk.Resource)
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource

// Modern SDK methods (return []sdk.DataSource/Resource)
func (r Registration) DataSources() []sdk.DataSource  
func (r Registration) Resources() []sdk.Resource

// New ephemeral resources (introduced recently)
func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource
```

#### Example Registration Patterns

**Legacy SDK Pattern (Map-based):**
```go
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
    return map[string]*pluginsdk.Resource{
        "azurerm_key_vault":              resourceKeyVault(),
        "azurerm_key_vault_certificate":  resourceKeyVaultCertificate(),
        // ... more mappings
    }
}
```

**Modern SDK Pattern (Slice-based):**
```go
func (r Registration) Resources() []sdk.Resource {
    return []sdk.Resource{
        KeyVaultCertificateContactsResource{},
    }
}
```

**Framework Pattern (Function slice):**
```go
func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource {
    return []func() ephemeral.EphemeralResource{
        NewKeyVaultCertificateEphemeralResource,
        NewKeyVaultSecretEphemeralResource,
    }
}
```

### Enhanced AST Parsing Strategy

Need to extend the sample code to handle **5 registration method types**:

#### Required AST Parsing Adaptations

1. **Map-based extractors** (for SupportedDataSources/SupportedResources):
```go
// From sample: extractMappingsFromMethod() - already handles this pattern
func extractMappingsFromMethod(node *ast.File, methodName string) map[string]string
```

2. **Slice literal extractors** (for DataSources/Resources):
```go
// From sample: extractTypesFromMethod() - needs adaptation for struct literals
func extractTypesFromMethod(node *ast.File, methodName string) []string
```

3. **Function slice extractors** (for EphemeralResources):
```go
// NEW: Need to implement function name extraction from slice literals
func extractFunctionNamesFromMethod(node *ast.File, methodName string) []string
```

**Phase 2: Enhanced AST Parsing**
Based on sample code patterns, need these extractors that work with `gophon.FileInfo`:

```go
// Parse AST from gophon FileInfo
func parseFileForRegistrations(fileInfo *gophon.FileInfo) (*FileRegistrations, error) {
    // Parse AST from fileInfo.Path or fileInfo.Content
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, fileInfo.Path, nil, parser.ParseComments)
    if err != nil {
        return nil, err
    }
    
    // Only extract methods that are actually present in this file
    // Most files will return empty maps/slices for most of these
    return &FileRegistrations{
        SupportedDataSources: extractMappingsFromMethod(node, "SupportedDataSources"), // may be empty
        SupportedResources:   extractMappingsFromMethod(node, "SupportedResources"),   // may be empty  
        DataSources:         extractStructTypesFromMethod(node, "DataSources"),        // may be empty
        Resources:           extractStructTypesFromMethod(node, "Resources"),          // may be empty
        EphemeralResources:  extractFunctionNamesFromMethod(node, "EphemeralResources"), // may be empty
    }, nil
}

// Existing pattern from sample (works for SupportedDataSources/SupportedResources)
func extractMappingsFromMethod(node *ast.File, methodName string) map[string]string

// Enhanced for struct literals (DataSources/Resources methods)
func extractStructTypesFromMethod(node *ast.File, methodName string) []string {
    // Handle: []sdk.Resource{StructName{}, AnotherStruct{}}
}

// NEW: For function names (EphemeralResources methods)  
func extractFunctionNamesFromMethod(node *ast.File, methodName string) []string {
    // Handle: []func() ephemeral.EphemeralResource{FuncName, AnotherFunc}
}
```

**Phase 3: Structured Output Generation**
```go
type TerraformProviderIndex struct {
    Version      string                        // Provider version
    Services     []ServiceRegistration         // All service registrations
    GlobalMaps   struct {
        AllDataSources map[string]string       // Complete mapping across all services
        AllResources   map[string]string       // Complete mapping across all services
    }
    Statistics   struct {
        ServiceCount        int
        TotalDataSources    int
        TotalResources      int
        LegacyResources     int
        ModernResources     int
        EphemeralResources  int
    }
}
```

### Integration with Current Workflow Plan

This enhanced AST approach will **complement gophon** with more targeted indexing:

- **Pros**: Complete Terraform provider understanding, smaller targeted indexes, cross-service analysis
- **Cons**: More complex than gophon alone, provider-specific implementation  
- **Output**: Structured JSON + individual service indexes alongside gophon's function/type indexes

### Implementation Considerations for 130+ Services

1. **Performance**: Leverage `gophon.ScanPackagesRecursively()` for efficient package discovery
2. **File-level parsing**: Process all `pkg.Files` in each `PackageInfo`, not just assumed filenames
3. **Error handling**: Graceful handling of parsing failures per file/package
4. **Memory management**: Stream processing using gophon's package discovery
5. **Validation**: Check all files in each package for registration methods
6. **Progress tracking**: Report scanning progress across packages and files
7. **Flexible discovery**: Don't assume registration code location - scan all Go files
8. **Sparse results handling**: Most files will have no registration methods - handle empty results efficiently

## Success Criteria
1. Workflow triggers on new terraform-provider-azurerm releases
2. Successfully generates gophon indexes for the entire codebase
3. Commits and pushes index files to repository
4. Creates matching tags for version tracking
5. AI agents can access individual symbol files via predictable URLs
6. Process completes within GitHub Actions time limits

## Progress Status

### ✅ Completed
- [x] **SupportedResources parser**: Created `pkg/parser.go` with `ExtractSupportedResourcesMappings()` function
- [x] **SupportedDataSources parser**: Added `ExtractSupportedDataSourcesMappings()` function with shared `extractMappingsFromMethod()` logic
- [x] **Modern SDK DataSources parser**: Added `ExtractDataSourcesStructTypes()` function for parsing slice-based registration methods
- [x] **Unit tests**: Comprehensive test coverage for all three parsers (legacy map-based and modern slice-based)
- [x] **Project structure**: Set up `pkg/` folder with proper Go module structure
- [x] **Code refactoring**: Created reusable functions for both map-based and slice-based parsing patterns

### 🚧 Next Tasks (In Priority Order)

#### 1. **Modern SDK Resources Parser** (Next)
- [ ] Add `ExtractResourcesStructTypes(node *ast.File) []string` function
- [ ] Handle slice literal parsing: `[]sdk.Resource{StructName{}, AnotherStruct{}}`
- [ ] Create unit tests for Resources method parsing
- [ ] Extract struct type names from composite literals

#### 2. **EphemeralResources Parser**
- [ ] Add `ExtractEphemeralResourcesFunctions(node *ast.File) []string` function
- [ ] Handle function slice parsing: `[]func() ephemeral.EphemeralResource{FuncName, AnotherFunc}`
- [ ] Create unit tests for EphemeralResources method parsing
- [ ] Extract function names from slice literals

#### 3. **Integration & Cross-Service Analysis**
- [ ] Create unified parser that handles all registration method types
- [ ] Implement package-level scanning across all services
- [ ] Create structured output JSON generation
- [ ] Add comprehensive integration tests

## Next Steps
1. ~~Review this plan~~ ✅
2. ~~**Decide**: Custom AST parsing vs gophon vs hybrid approach~~ ✅ (Chose hybrid)
3. ~~Implement Phase 1 changes~~ 🚧 (In progress - SupportedResources and SupportedDataSources done)
4. Complete remaining registration method parsers (DataSources, Resources, EphemeralResources)
5. Test with a known terraform-provider-azurerm version
6. Implement GitHub Actions workflow adaptation
7. Iterate and improve based on results
