package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lexieqin/Geek/GenesisGpt/cmd/config"
)

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Get(url string, headers map[string]string) (string, error)
	Post(url string, body []byte, headers map[string]string) (string, error)
	Delete(url string, headers map[string]string) (string, error)
}

// DefaultHTTPClient is the default HTTP client implementation
type DefaultHTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with timeout
func NewHTTPClient() *DefaultHTTPClient {
	cfg := config.GetConfig()
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: cfg.Common.Timeout,
		},
	}
}

// Get performs HTTP GET request with optional headers
func (c *DefaultHTTPClient) Get(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Post performs HTTP POST request with optional headers
func (c *DefaultHTTPClient) Post(url string, body []byte, headers map[string]string) (string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// Delete performs HTTP DELETE request with optional headers
func (c *DefaultHTTPClient) Delete(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return "", err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// GetHTTP executes a GET HTTP request to the specified URL and returns the response body.
// This is kept for backward compatibility
func GetHTTP(url string) (string, error) {
	client := NewHTTPClient()
	return client.Get(url, nil)
}

// PostHTTP executes a POST HTTP request to the specified URL and returns the response body.
// This is kept for backward compatibility
func PostHTTP(url string, body []byte) (string, error) {
	client := NewHTTPClient()
	return client.Post(url, body, nil)
}

// DeleteHTTP executes a DELETE HTTP request to the specified URL and returns the response body.
// This is kept for backward compatibility
func DeleteHTTP(url string) (string, error) {
	client := NewHTTPClient()
	return client.Delete(url, nil)
}

// GetHTTPWithAuth performs HTTP GET with authentication based on config
func GetHTTPWithAuth(url string, authType string) (string, error) {
	client := NewHTTPClient()
	headers := make(map[string]string)

	// Add authentication headers if in production mode
	if !config.IsMockMode() {
		authConfig := config.GetAuthConfig()
		if authConfig != nil {
			switch authType {
			case "job":
				addAuthHeaders(headers, authConfig.JobAPI)
			case "datadog":
				addDatadogHeaders(headers, authConfig.Datadog)
			case "sandbox":
				addAuthHeaders(headers, authConfig.Sandbox)
			}
		}
	}

	return client.Get(url, headers)
}

func addAuthHeaders(headers map[string]string, auth config.AuthMethod) {
	switch auth.Type {
	case "bearer":
		if auth.Token != "" {
			headers["Authorization"] = "Bearer " + auth.Token
		}
	case "api-key":
		if auth.APIKey != "" {
			headers["X-API-Key"] = auth.APIKey
		}
	case "basic":
		// Add basic auth support if needed
	}
}

func addDatadogHeaders(headers map[string]string, auth config.AuthMethod) {
	if auth.APIKey != "" {
		headers["DD-API-KEY"] = auth.APIKey
	}
	if auth.AppKey != "" {
		headers["DD-APPLICATION-KEY"] = auth.AppKey
	}
}