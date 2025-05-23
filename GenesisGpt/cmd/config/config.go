package config

import "os"

// GetGinToolsURL returns the ginTools service URL
func GetGinToolsURL() string {
	url := os.Getenv("GINTOOLS_URL")
	if url == "" {
		// Default for local development
		return "http://localhost:8080"
	}
	return url
}