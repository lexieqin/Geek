package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexieqin/Geek/K8sGpt/cmd/utils"
)

type ListToolParam struct {
	Resource  string `json:"resource"`
	Namespace string `json:"namespace"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
}

type APIResponse struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

// ListTool represents a tool for listing k8s resource commands.
type ListTool struct {
	Name        string
	Description string
	ArgsSchema  string
}

// NewListTool creates a new ListTool instance.
func NewListTool() *ListTool {
	return &ListTool{
		Name:        "ListTool",
		Description: "Used to list and get details of Kubernetes resources. Can list all resources of a type in a namespace, get specific resource details, or filter resources by type.",
		ArgsSchema:  `{"type":"object","properties":{"resource":{"type":"string", "description": "Specified k8s resource type, e.g. pod, service, etc."}, "namespace":{"type":"string", "description": "Specified k8s namespace"}, "name":{"type":"string", "description": "Optional: Name of specific resource to get details for"}, "type":{"type":"string", "description": "Optional: Filter resources by type"}}}`,
	}
}

// Run executes the command and returns the output.
func (l *ListTool) Run(resource string, ns string, name string, resourceType string) (string, error) {
	resource = strings.ToLower(resource)
	var url string

	if name != "" {
		// Get specific resource details
		url = fmt.Sprintf("http://localhost:8080/%s?ns=%s&name=%s", resource, ns, name)
	} else if resourceType != "" {
		// Filter resources by type
		url = fmt.Sprintf("http://localhost:8080/get/resource?resource=%s&type=%s", resource, resourceType)
	} else {
		// List all resources
		if resource == "pod" || resource == "pods" {
			// Use the namespace-specific endpoint for pods
			url = fmt.Sprintf("http://localhost:8080/namespaces/%s/pods", ns)
		} else {
			// Use the generic resource endpoint for other resources
			url = fmt.Sprintf("http://localhost:8080/%s?ns=%s", resource, ns)
		}
	}

	response, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to get resource: %v", err)
	}

	// Parse the API response
	var apiResp APIResponse
	if err := json.Unmarshal([]byte(response), &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %v", err)
	}

	// Convert the data to a readable string
	data, err := json.MarshalIndent(apiResp.Data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response data: %v", err)
	}

	return string(data), nil
}
