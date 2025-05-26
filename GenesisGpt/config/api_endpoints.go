package config

import (
	"os"
)

type APIEndpoints struct {
	JobAPI      string
	TraceAPI    string
	SandboxAPI  string
}

func GetAPIEndpoints() *APIEndpoints {
	env := os.Getenv("ENVIRONMENT")
	
	if env == "production" {
		return &APIEndpoints{
			JobAPI:     os.Getenv("PROD_JOB_API"),      // e.g., https://api.yourplatform.com
			TraceAPI:   os.Getenv("PROD_TRACE_API"),    // e.g., https://api.datadoghq.com
			SandboxAPI: os.Getenv("PROD_SANDBOX_API"),  // e.g., https://sandbox.yourplatform.com
		}
	}
	
	// Default to mock/development endpoints
	return &APIEndpoints{
		JobAPI:     "http://localhost:8080",
		TraceAPI:   "http://localhost:8080/api/datadog",
		SandboxAPI: "http://localhost:8080/api/sandbox",
	}
}