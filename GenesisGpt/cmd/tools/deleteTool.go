package tools

import (
	"strings"

	"github.com/lexieqin/Geek/K8sGpt/cmd/utils"
)

type DeleteToolParam struct {
	Resource  string `json:"resource"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DeleteTool represents a tool for deleting k8s resources.
type DeleteTool struct {
	Name        string
	Description string
	ArgsSchema  string
}

// NewDeleteTool creates a new DeleteTool instance.
func NewDeleteTool() *DeleteTool {
	return &DeleteTool{
		Name:        "DeleteTool",
		Description: "Used to delete specified Kubernetes resources in a given namespace, such as deleting pods, services, etc.",
		ArgsSchema:  `{"type":"object","properties":{"resource":{"type":"string", "description": "Specified k8s resource type, e.g. pod, service, etc."}, "name":{"type":"string", "description": "Name of the specified k8s resource instance"}, "namespace":{"type":"string", "description": "Namespace where the specified k8s resource is located"}}`,
	}
}

// Run executes the command and returns the output.
func (d *DeleteTool) Run(resource, name, ns string) error {
	resource = strings.ToLower(resource)

	url := "http://localhost:8080/" + resource + "?ns=" + ns + "&name=" + name

	_, err := utils.DeleteHTTP(url)

	return err
}
