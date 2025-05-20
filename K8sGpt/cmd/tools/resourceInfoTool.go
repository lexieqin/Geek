package tools

import (
	"fmt"

	"github.com/xingyunyang01/K8sGpt/cmd/utils"
)

type ResourceInfoToolParam struct {
	Resource string `json:"resource"`
	InfoType string `json:"infoType"` // "gvr" or "list"
}

// ResourceInfoTool represents a tool for getting resource type information.
type ResourceInfoTool struct {
	Name        string
	Description string
	ArgsSchema  string
}

// NewResourceInfoTool creates a new ResourceInfoTool instance.
func NewResourceInfoTool() *ResourceInfoTool {
	return &ResourceInfoTool{
		Name:        "ResourceInfoTool",
		Description: "Used to get information about Kubernetes resource types. Can retrieve GVR (GroupVersionResource) information or list available resources of a specific type.",
		ArgsSchema:  `{"type":"object","properties":{"resource":{"type":"string", "description": "Resource type to get information for"}, "infoType":{"type":"string", "description": "Type of information to retrieve: 'gvr' for GroupVersionResource info or 'list' for resource list"}}}`,
	}
}

// Run executes the command and returns the output.
func (r *ResourceInfoTool) Run(param ResourceInfoToolParam) (string, error) {
	var url string

	if param.InfoType == "gvr" {
		url = "http://localhost:8080/get/gvr?resource=" + param.Resource
	} else if param.InfoType == "list" {
		url = "http://localhost:8080/get/resource?resource=" + param.Resource
	} else {
		return "", fmt.Errorf("invalid info type: %s", param.InfoType)
	}

	s, err := utils.GetHTTP(url)
	return s, err
}
