package pkg

// GlobalMappings represents complete mappings across all services
type GlobalMappings struct {
	AllDataSources map[string]string `json:"all_data_sources"` // Complete mapping across all services
	AllResources   map[string]string `json:"all_resources"`    // Complete mapping across all services
}
