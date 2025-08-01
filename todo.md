# TODO: Terraform Provider AzureRM Index Generation Plan

## ✅ COMPLETED: Progress Bar Enhancement
**Status**: COMPLETED ✅  
**Date**: August 1, 2025

### Summary
Successfully implemented rich progress bars for the Terraform provider indexing operations, inspired by the gophon CLI tool. The enhancement provides visual feedback during the two main phases:
1. **Scanning Terraform Provider Services** - Shows progress while analyzing Go source files
2. **Writing Index Files** - Shows progress while generating JSON output files

### Implementation Details
- **Progress Infrastructure**: Created `pkg/progress.go` with ProgressInfo struct and callback system
- **Rich Display**: Unicode progress bars with percentage, ETA calculations, processing rates, and emoji indicators
- **Clean API**: Implemented ProgressTracker helper class to avoid complex parameter passing
- **Thread-Safe**: Uses atomic operations for concurrent progress updates
- **Flexible**: Supports both rich and simple progress display modes

### Key Features
- 🔄 Real-time progress bars with Unicode characters
- 📊 Percentage completion and item counts
- ⏱️ Elapsed time tracking
- 🔮 ETA (Estimated Time to Arrival) calculations
- ⚡ Processing rate display (items/second)
- 📦 Current item being processed
- ✅ Completion notifications

### Code Structure
- `pkg/progress.go`: Core progress tracking types and display functions
- `pkg/progress_tracker.go`: ProgressTracker helper class for clean API
- `pkg/terraform_provider_index.go`: Integration with scanning and writing operations
- `main.go`: Rich progress callback setup
- Comprehensive test coverage for all progress functionality

---

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
  "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/resource",
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
  "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/client",
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
  "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
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
    Namespace          string `json:"namespace"`                    // "github.com/hashicorp/terraform-provider-azurerm/internal/services/resource"
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
    Namespace          string `json:"namespace"`                    // "github.com/hashicorp/terraform-provider-azurerm/internal/services/client"
    RegistrationMethod string `json:"registration_method"`          // "SupportedDataSources", "DataSources", etc.
    SDKType            string `json:"sdk_type"`                     // "legacy_pluginsdk", "modern_sdk"
    SchemaMethod       string `json:"schema_method,omitempty"`      // "dataSourceArmClientConfigSchema" (optional)
    ReadMethod         string `json:"read_method,omitempty"`        // "dataSourceArmClientConfigRead" (optional)
    AttributeMethod    string `json:"attribute_method,omitempty"`   // "dataSourceArmClientConfigAttributes" (optional)
}

type TerraformEphemeralInfo struct {
    TerraformType      string `json:"terraform_type"`               // "azurerm_key_vault_certificate"
    StructType         string `json:"struct_type"`                  // "KeyVaultCertificateEphemeralResource"
    Namespace          string `json:"namespace"`                    // "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault"
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
- [x] **Modern SDK Resources parser**: Added `ExtractResourcesStructTypes()` function for parsing slice-based registration methods
- [x] **EphemeralResources parser**: Added `ExtractEphemeralResourcesFunctions()` function for parsing function slice registration methods
- [x] **Integration & Cross-Service Analysis**: Implemented `ScanTerraformProviderServices()` function with gophon integration, structured output types, and test harness
- [x] **Unit tests**: Comprehensive test coverage for all five parsers (legacy map-based, modern slice-based, and ephemeral function-based)
- [x] **Integration tests**: Added `TestScanTerraformProviderServices()` with complete test harness including all registration method types
- [x] **Project structure**: Set up `pkg/` folder with proper Go module structure and test harness under `pkg/testharness/`
- [x] **Code refactoring**: Created reusable functions for map-based, slice-based, and function-based parsing patterns
- [x] **Structured data types**: Implemented all Phase 3 data structures for TerraformProviderIndex, ServiceRegistration, and individual resource info types
- [x] **ExtractCRUDFromPackage function**: Enhanced to support basic CRUD field names (`Create`, `Read`, `Update`, `Delete`) in addition to context variants
- [x] **ExtractCRUDFromPackage unit tests**: Comprehensive test coverage including real Azure provider function examples

### 🚧 Next Tasks (In Priority Order)

#### 1. **Unified CRUD Method Extraction for Legacy Plugin SDK** (Next Priority)
- [ ] **Problem**: Currently legacy plugin SDK resources only have `RegistrationMethod` set, while modern framework has all `xxxMethod` fields populated
- [ ] **Goal**: Unify handling by analyzing legacy plugin registration method source code to extract CRUD operations
- [ ] **Implementation Plan**:
  - [ ] Create `ExtractLegacyResourceCRUDMethods()` function to parse `pluginsdk.Resource` return values
  - [ ] Parse `Schema`, `CreateFunc`, `ReadFunc`, `UpdateFunc`, `DeleteFunc` fields from resource function bodies
  - [ ] Handle both direct return and variable assignment patterns
  - [ ] Add comprehensive tests for legacy resource CRUD extraction
  - [ ] Update `GenerateIndividualResourceFiles()` to use extracted CRUD methods for legacy resources
- [ ] **Expected Result**: Both legacy and modern resources will have complete method information in their JSON output

#### 2. **Unified CRUD Method Extraction for Legacy Plugin SDK Data Sources** (Next)
- [ ] Create `ExtractLegacyDataSourceMethods()` function to parse legacy data source functions
- [ ] Parse `Schema`, `ReadFunc` fields from data source function bodies 
- [ ] Update data source generation to include extracted methods

#### 3. **JSON Output Generation & File Writing** (After CRUD extraction)
- [ ] Add functions to generate individual JSON files for resources, data sources, and ephemeral resources
- [ ] Create directory structure (resources/, datasources/) for organized output  
- [ ] Implement file writing functionality for structured JSON indexes
- [ ] Add tests for JSON generation and file output

## Technical Implementation Plan: Legacy Plugin SDK CRUD Extraction

### Current State vs Target State
**Current**: Legacy plugin SDK resources only have `RegistrationMethod` populated:
```json
{
  "terraform_type": "azurerm_key_vault",
  "registration_method": "resourceKeyVault", 
  "sdk_type": "legacy_pluginsdk",
  "schema_method": "",      // Empty
  "create_method": "",      // Empty  
  "read_method": "",        // Empty
  "update_method": "",      // Empty
  "delete_method": ""       // Empty
}
```

**Target**: Legacy plugin SDK resources have all CRUD methods populated like modern framework:
```json
{
  "terraform_type": "azurerm_key_vault",
  "registration_method": "resourceKeyVault",
  "sdk_type": "legacy_pluginsdk", 
  "schema_method": "keyVaultSchema",
  "create_method": "keyVaultCreateFunc",
  "read_method": "keyVaultReadFunc", 
  "update_method": "keyVaultUpdateFunc",
  "delete_method": "keyVaultDeleteFunc"
}
```

### Legacy Plugin SDK Resource Structure Analysis
Legacy plugin SDK resources return `*pluginsdk.Resource` structs like this:
```go
func resourceKeyVault() *pluginsdk.Resource {
    return &pluginsdk.Resource{
        CreateContext: keyVaultCreateFunc,
        ReadContext:   keyVaultReadFunc,
        UpdateContext: keyVaultUpdateFunc, 
        DeleteContext: keyVaultDeleteFunc,
        
        Schema: map[string]*pluginsdk.Schema{
            "name": {Type: pluginsdk.TypeString, Required: true},
            // ... more schema fields
        },
        
        Importer: &pluginsdk.ResourceImporter{
            StateContext: keyVaultImporter,
        },
        
        Timeouts: &pluginsdk.ResourceTimeout{
            Create: pluginsdk.DefaultTimeout(30 * time.Minute),
            Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
            Update: pluginsdk.DefaultTimeout(30 * time.Minute), 
            Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
        },
    }
}
```

### AST Parsing Strategy for Legacy Plugin SDK

#### Phase 1: Function Body Analysis
1. **Find Resource Function**: Scan the AST file for any function that returns `*pluginsdk.Resource`
2. **Extract Return Statement**: Look for `return &pluginsdk.Resource{...}` or variable assignment patterns
3. **Parse Composite Literal**: Extract field assignments from the struct literal

#### Phase 2: CRUD Method Field Extraction  
Extract function names/references from these fields:
- `CreateContext` / `CreateFunc` → `create_method`
- `ReadContext` / `ReadFunc` → `read_method` 
- `UpdateContext` / `UpdateFunc` → `update_method`
- `DeleteContext` / `DeleteFunc` → `delete_method`

### Implementation Functions

#### Core Function: `ExtractLegacyResourceCRUDMethods()`
```go
// LegacyResourceCRUDMethods represents CRUD methods extracted from legacy plugin SDK resources
type LegacyResourceCRUDMethods struct {
    CreateMethod string `json:"create_method,omitempty"`    // "keyVaultCreateFunc"
    ReadMethod   string `json:"read_method,omitempty"`      // "keyVaultReadFunc"
    UpdateMethod string `json:"update_method,omitempty"`    // "keyVaultUpdateFunc"
    DeleteMethod string `json:"delete_method,omitempty"`    // "keyVaultDeleteFunc"
}

// ExtractLegacyResourceCRUDMethods analyzes a legacy plugin SDK resource function 
// and extracts CRUD method names from the returned pluginsdk.Resource struct
// The input ast.File should contain the registration function's source code
// It will find any function that returns *pluginsdk.Resource and parse its CRUD methods
func ExtractLegacyResourceCRUDMethods(node *ast.File) (*LegacyResourceCRUDMethods, error)
```

#### Supporting Functions:
```go
// extractFromResourceLiteral parses a pluginsdk.Resource composite literal
// and extracts only CRUD method names
func extractFromResourceLiteral(compLit *ast.CompositeLit) *LegacyResourceCRUDMethods

// extractFunctionReference extracts function name from various AST patterns:
// - Direct identifier: funcName
// - Selector expression: package.FuncName  
func extractFunctionReference(expr ast.Expr) string

// findResourceFunction locates any function declaration that returns *pluginsdk.Resource
func findResourceFunction(node *ast.File) *ast.FuncDecl
```

### AST Parsing Patterns to Handle

#### Pattern 1: Direct Return
```go
func resourceKeyVault() *pluginsdk.Resource {
    return &pluginsdk.Resource{
        CreateContext: keyVaultCreateFunc,
        ReadContext:   keyVaultReadFunc,
        UpdateContext: keyVaultUpdateFunc,
        DeleteContext: keyVaultDeleteFunc,
    }
}
```

#### Pattern 2: Variable Assignment  
```go
func resourceKeyVault() *pluginsdk.Resource {
    resource := &pluginsdk.Resource{
        CreateContext: keyVaultCreateFunc,
        ReadContext:   keyVaultReadFunc,
        UpdateContext: keyVaultUpdateFunc,
        DeleteContext: keyVaultDeleteFunc,
    }
    return resource
}
```

#### Focus: CRUD Fields Only
Extract function names/references from these specific fields:
- `CreateContext` / `CreateFunc` → `create_method`
- `ReadContext` / `ReadFunc` → `read_method` 
- `UpdateContext` / `UpdateFunc` → `update_method`
- `DeleteContext` / `DeleteFunc` → `delete_method`

### Integration with Current Codebase

#### Workflow: 
1. **Registration Method Parsing**: Use existing functions to get `registrationMethod` name with `Namespace`
2. **Source Code Reading**: You read the registration function's source code and parse it into `*ast.File`
3. **CRUD Extraction**: Pass the `*ast.File` to `ExtractLegacyResourceCRUDMethods()`
4. **Population**: Populate the CRUD method fields in `TerraformResourceInfo`

#### Update `GenerateIndividualResourceFiles()` (Conceptual):
```go
// For legacy resources, extract CRUD methods after getting registration mapping
for terraformType, registrationMethod := range service.SupportedResources {
    // YOU will read the function source code and parse to *ast.File
    functionAST := readAndParseFunction(namespace, registrationMethod)
    
    // NEW: Extract CRUD methods from the registration function
    crudMethods, err := ExtractLegacyResourceCRUDMethods(functionAST)
    if err != nil {
        // Log warning but continue
        crudMethods = &LegacyResourceCRUDMethods{}
    }
    
    resources[terraformType] = TerraformResourceInfo{
        TerraformType:      terraformType,
        StructType:         "", // Still empty for legacy
        Namespace:          namespace,
        RegistrationMethod: registrationMethod,
        SDKType:            "legacy_pluginsdk",
        // NEW: Populate CRUD methods from extraction
        CreateMethod:       crudMethods.CreateMethod,
        ReadMethod:         crudMethods.ReadMethod,
        UpdateMethod:       crudMethods.UpdateMethod,
        DeleteMethod:       crudMethods.DeleteMethod,
    }
}
```

### Testing Strategy

#### Unit Tests for CRUD Extraction
- Test direct return pattern parsing
- Test variable assignment pattern parsing  
- Test error handling for malformed functions
- Test with different CRUD field variations (CreateContext vs CreateFunc)

#### Integration Tests
- Test with real Terraform provider function examples
- Test error handling for missing functions
- Test performance with various function sizes

### Benefits of This Approach
1. **Unified Data Model**: Both legacy and modern resources have complete CRUD method information
2. **Better AI Agent Support**: AI agents can find CRUD methods for any resource type
3. **Focused Implementation**: Only parses essential CRUD operations, keeping it simple
4. **Clear Separation**: You handle source code reading, function focuses on AST parsing
5. **Backward Compatible**: Doesn't break existing functionality for modern framework resources

## Session Progress (Current Development Session)

### ✅ **COMPLETED: Modern Framework Terraform Type Extraction** 
*Session Date: July 31, 2025*

Successfully implemented comprehensive Terraform type extraction for modern Plugin Framework resources, data sources, and ephemeral resources. **Critical Issue Resolved**: The system now extracts actual Terraform types (e.g., `"azurerm_key_vault"`) instead of just struct names (e.g., `"KeyVaultResource"`).

#### **1. Enhanced Data Structures** ✅
- **Modified `ServiceRegistration`** with new mapping fields:
  - `ResourceTerraformTypes map[string]string` - Maps struct names to Terraform types for resources
  - `DataSourceTerraformTypes map[string]string` - Maps struct names to Terraform types for data sources  
  - `EphemeralTerraformTypes map[string]string` - Maps struct names to Terraform types for ephemeral resources

#### **2. AST Parsing Functions** ✅
- **`extractTerraformTypeFromResourceTypeMethod()`** - Extracts Terraform types from `ResourceType()` methods for resources and data sources
- **`extractTerraformTypeFromMetadataMethod()`** - Extracts Terraform types from `Metadata()` methods for ephemeral resources
- **Receiver Type Support**: Both functions handle value receivers (`func (r StructName)`) and pointer receivers (`func (r *StructName)`)
- **Robust Parsing**: Handles different method signatures and patterns

#### **3. Updated JSON Generation** ✅
- **Modified Generation Functions**:
  - `writeResourceFiles()` - Now uses actual Terraform types from mappings
  - `writeDataSourceFiles()` - Now uses actual Terraform types from mappings
  - `writeEphemeralFiles()` - Now uses actual Terraform types from mappings
- **Accuracy Improvement**: JSON output now contains correct Terraform resource type names instead of Go struct names

#### **4. Comprehensive Test Coverage** ✅
- **Table-Driven Tests** implemented for all extraction scenarios:
  - `TestExtractTerraformTypeFromResourceTypeMethod` - Tests resource `ResourceType()` method extraction
  - `TestExtractTerraformTypeFromDataSourceResourceTypeMethod` - Tests data source `ResourceType()` method extraction  
  - `TestExtractTerraformTypeFromMetadataMethod` - Tests ephemeral resource `Metadata()` method extraction
- **Test Coverage**: Both value and pointer receivers, various method patterns, edge cases
- **Test Framework**: Using `gophon.PackageInfo` mocks with `github.com/stretchr/testify` assertions

#### **5. Key Technical Achievements** ✅
- **Framework Compatibility**: Supports both legacy Provider SDK and modern Terraform Plugin Framework patterns
- **AST-Based Extraction**: Uses Go AST parsing to accurately extract types from method implementations
- **Type Safety**: Proper handling of different receiver types and method signatures
- **Future-Proof**: The AST-based approach can easily be extended for new patterns

#### **6. Impact & Benefits** ✅
- **Accurate Indexing**: Generated indexes now contain actual Terraform types that users reference in configurations
- **AI Agent Compatibility**: AI agents can now correctly identify and work with Terraform resource types
- **Developer Experience**: Improved accuracy makes the index more valuable for development tools
- **Quality Assurance**: Comprehensive tests ensure reliability and prevent regressions

### **Current State**: Ready for Integration Testing
The implementation is **complete and tested** for modern framework types. The system can now correctly extract Terraform types like:
- Resources: `"azurerm_container_app_environment_dapr_component"` from `ContainerAppEnvironmentDaprComponentResource`
- Data Sources: `"azurerm_key_vault"` from `KeyVaultDataSource`  
- Ephemeral Resources: `"azurerm_key_vault_certificate"` from `KeyVaultCertificateEphemeralResource`

### **Next Session Goals**:
- Integration testing with real Terraform provider codebase
- Performance validation on large codebases
- End-to-end workflow testing

---

## 🎯 **PROGRESS BAR ENHANCEMENT PLAN**
*Session Date: August 1, 2025*

### **Current State Analysis**
The existing code has basic progress output with simple `fmt.Printf` statements:
- **Scanning Phase**: `fmt.Printf("%s scanned.\n", servicePath)` in `ScanTerraformProviderServices`
- **Indexing Phase**: Multiple `fmt.Printf("resource %s indexed. \n", ...)` statements in write operations
- **No Progress Context**: No percentage, ETA, or visual progress indicators
- **No Rate Information**: No indication of processing speed or remaining time

### **Target State: Rich Visual Progress Bars**
Inspired by the `gophon` CLI tool's excellent progress implementation, implement comprehensive progress tracking with:
- **Visual Progress Bars**: Rich Unicode progress bars with percentage completion
- **ETA Calculations**: Estimated time remaining based on current processing rate
- **Processing Rate**: Items processed per second indicator
- **Current Item Display**: Show what's currently being processed
- **Phase Indicators**: Clear indication of which phase is running (scanning vs indexing)
- **Elapsed Time**: Real-time elapsed time display

### **Implementation Plan**

#### **Phase 1: Progress Infrastructure** 
1. **Create Progress Types**:
   ```go
   // ProgressInfo represents progress information for a long-running operation
   type ProgressInfo struct {
       Phase       string    // "scanning" or "indexing"
       Current     string    // Current item being processed
       Completed   int       // Number of items completed
       Total       int       // Total number of items
       Percentage  float64   // Completion percentage (0-100)
       StartTime   time.Time // When the operation started
   }
   
   // ProgressCallback is called to report progress updates
   type ProgressCallback func(ProgressInfo)
   ```

2. **Progress Display Utilities**:
   ```go
   // createProgressBar generates a Unicode progress bar string
   func createProgressBar(percentage float64, width int) string
   
   // formatProgress creates a formatted progress line with all indicators
   func formatProgress(info ProgressInfo) string
   
   // calculateETA estimates time remaining based on current progress
   func calculateETA(elapsed time.Duration, percentage float64) time.Duration
   ```

#### **Phase 2: ScanTerraformProviderServices Progress**
1. **Modify Function Signature**:
   ```go
   func ScanTerraformProviderServices(dir, basePkgUrl string, version string, 
       progressCallback ProgressCallback) (*TerraformProviderIndex, error)
   ```

2. **Implement Scanning Progress**:
   - Track total number of service directories
   - Report progress for each service being scanned
   - Update progress as workers complete service processing
   - Use atomic counters for thread-safe progress tracking

3. **Integration Points**:
   - Before worker goroutines: Initialize progress with total count
   - In worker goroutines: Report progress after each service scan
   - After all workers: Final completion progress update

#### **Phase 3: WriteIndexFiles Progress**
1. **Modify Function Signatures**:
   ```go
   func (index *TerraformProviderIndex) WriteIndexFiles(outputDir string, 
       progressCallback ProgressCallback) error
   func (index *TerraformProviderIndex) WriteResourceFiles(outputDir string, 
       progressCallback ProgressCallback) error
   func (index *TerraformProviderIndex) WriteDataSourceFiles(outputDir string, 
       progressCallback ProgressCallback) error
   func (index *TerraformProviderIndex) WriteEphemeralFiles(outputDir string, 
       progressCallback ProgressCallback) error
   ```

2. **Implement Indexing Progress**:
   - Calculate total number of files to be written (resources + data sources + ephemeral)
   - Track progress across all writing operations
   - Report current file being written and completion percentage
   - Use atomic counters for thread-safe progress in parallel writing

#### **Phase 4: Main Integration**
1. **Create Rich Progress Display Function**:
   ```go
   // createRichProgressCallback creates a callback that displays rich progress information
   func createRichProgressCallback() ProgressCallback {
       return func(progress ProgressInfo) {
           elapsed := time.Since(progress.StartTime)
           eta := calculateETA(elapsed, progress.Percentage)
           rate := calculateProcessingRate(progress.Completed, elapsed)
           
           bar := createProgressBar(progress.Percentage, 50)
           
           // Display rich progress with Unicode indicators
           fmt.Printf("\r🔄 %s | [%s] %.1f%% (%d/%d) | ⏱️ %.1fs | 🔮 ETA: %.1fs | ⚡ %.1f/s | 📦 %s",
               progress.Phase, bar, progress.Percentage, progress.Completed, progress.Total,
               elapsed.Seconds(), eta.Seconds(), rate, truncateString(progress.Current, 30))
           
           if progress.Percentage >= 100 {
               fmt.Printf("\n✅ %s completed!\n", progress.Phase)
           }
       }
   }
   ```

2. **Update main.go Usage**:
   ```go
   progressCallback := createRichProgressCallback()
   
   fmt.Printf("🚀 Starting Terraform Provider Indexing...\n")
   
   // Scanning phase
   index, err := ScanTerraformProviderServices(dir, basePkgUrl, version, progressCallback)
   if err != nil {
       return err
   }
   
   // Writing phase
   err = index.WriteIndexFiles(outputDir, progressCallback)
   if err != nil {
       return err
   }
   
   fmt.Printf("🎉 All operations completed successfully!\n")
   ```

### **Technical Considerations**

#### **Thread Safety**
- Use `sync/atomic` for counters shared between goroutines
- Ensure progress callbacks are called from a single goroutine to avoid display conflicts
- Consider using channels to serialize progress updates

#### **Performance Impact**
- Progress updates should be lightweight and not significantly impact processing speed
- Use buffered channels to avoid blocking worker goroutines
- Consider rate-limiting progress updates (e.g., max 10 updates per second)

#### **Testing Strategy**
- Mock progress callbacks for unit tests
- Test progress calculation accuracy with known datasets
- Verify thread safety under concurrent operations
- Test edge cases (0 items, single item, completion scenarios)

#### **Backward Compatibility**
- Make progress callbacks optional (nil callback = no progress display)
- Maintain existing function signatures in internal APIs
- Provide wrapper functions for existing callers

### **Expected Benefits**
1. **Better User Experience**: Users can see real-time progress and estimated completion times
2. **Debugging Aid**: Progress information helps identify which services/resources are slow to process
3. **Professional Appearance**: Rich progress bars make the tool feel more polished and modern
4. **Performance Monitoring**: Processing rates help identify performance regressions
5. **Long-Running Operation Support**: ETA helps users plan for long indexing operations

### **Implementation Priority**
1. **High Priority**: Basic progress bars with percentage and current item
2. **Medium Priority**: ETA calculations and processing rates
3. **Low Priority**: Advanced features like memory tracking and cancellation support

This enhancement will transform the user experience from basic text output to rich, informative progress tracking similar to modern CLI tools.

---

## Next Steps
1. ~~Review this plan~~ ✅
2. ~~**Decide**: Custom AST parsing vs gophon vs hybrid approach~~ ✅ (Chose hybrid)
3. ~~Implement Phase 1 changes~~ ✅ **COMPLETED** (Modern framework Terraform type extraction)
4. Complete remaining registration method parsers (DataSources, Resources, EphemeralResources) - **Note**: These are complete for modern framework
5. Integration testing with actual terraform-provider-azurerm codebase
6. Test with a known terraform-provider-azurerm version
7. Implement GitHub Actions workflow adaptation
8. Iterate and improve based on results
