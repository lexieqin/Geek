package tools

import (
	"fmt"

	"github.com/xingyunyang01/K8sGpt/cmd/utils"
)

type PodToolParam struct {
	Namespace string `json:"namespace"`
	PodName   string `json:"podName"`
	Container string `json:"container,omitempty"`
	Tail      int    `json:"tail,omitempty"`
	EventType string `json:"eventType,omitempty"`
	Operation string `json:"operation"` // "logs" or "events"
}

// PodTool represents a tool for pod-specific operations.
type PodTool struct {
	Name        string
	Description string
	ArgsSchema  string
}

// NewPodTool creates a new PodTool instance.
func NewPodTool() *PodTool {
	return &PodTool{
		Name:        "PodTool",
		Description: "Used for pod-specific operations like getting logs and events. Can retrieve pod logs with optional container and line count, and get pod events with optional event type filtering.",
		ArgsSchema:  `{"type":"object","properties":{"namespace":{"type":"string", "description": "Namespace where the pod is located"}, "podName":{"type":"string", "description": "Name of the pod"}, "container":{"type":"string", "description": "Optional: Specific container name for logs"}, "tail":{"type":"integer", "description": "Optional: Number of log lines to retrieve"}, "eventType":{"type":"string", "description": "Optional: Filter events by type (e.g., Warning)"}, "operation":{"type":"string", "description": "Operation to perform: 'logs' or 'events'"}}}`,
	}
}

// Run executes the command and returns the output.
func (p *PodTool) Run(param PodToolParam) (string, error) {
	var url string

	if param.Operation == "logs" {
		url = fmt.Sprintf("http://localhost:8080/namespaces/%s/pods/%s/logs", param.Namespace, param.PodName)
		if param.Container != "" {
			url += "?container=" + param.Container
		}
		if param.Tail > 0 {
			if param.Container != "" {
				url += "&"
			} else {
				url += "?"
			}
			url += fmt.Sprintf("tail=%d", param.Tail)
		}
	} else if param.Operation == "events" {
		url = fmt.Sprintf("http://localhost:8080/namespaces/%s/pods/%s/events", param.Namespace, param.PodName)
		if param.EventType != "" {
			url += "?type=" + param.EventType
		}
	} else {
		return "", fmt.Errorf("invalid operation: %s", param.Operation)
	}

	s, err := utils.GetHTTP(url)
	return s, err
}
