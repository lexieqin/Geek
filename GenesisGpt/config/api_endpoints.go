package config

import (
	"github.com/lexieqin/Geek/GenesisGpt/cmd/config"
)

type APIEndpoints struct {
	JobAPI      string
	TraceAPI    string
	SandboxAPI  string
}

// GetAPIEndpoints returns API endpoints based on the YAML configuration
func GetAPIEndpoints() *APIEndpoints {
	apiConfig := config.GetAPIConfig()
	
	return &APIEndpoints{
		JobAPI:     apiConfig.JobAPIURL,
		TraceAPI:   apiConfig.DatadogAPIURL,
		SandboxAPI: apiConfig.SandboxLogsAPIURL,
	}
}