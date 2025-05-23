package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexieqin/Geek/K8sGpt/cmd/utils"
)

type JobDebugTool struct{}

func NewJobDebugTool() *JobDebugTool {
	return &JobDebugTool{}
}

func (t *JobDebugTool) Name() string {
	return "JobDebugTool"
}

func (t *JobDebugTool) Description() string {
	return "Debug failed Kubernetes jobs by retrieving comprehensive information including Datadog traces, error details, sandbox logs, and associated pod information. Can find jobs by name or UUID."
}

func (t *JobDebugTool) ArgsSchema() string {
	return `{
		"type": "object",
		"properties": {
			"uuid": {
				"type": "string",
				"description": "The UUID of the job to debug (use this OR name+namespace)"
			},
			"name": {
				"type": "string",
				"description": "The name of the job to debug"
			},
			"namespace": {
				"type": "string",
				"description": "The namespace of the job (required if using name, optional if using UUID)"
			},
			"debug_type": {
				"type": "string",
				"enum": ["full", "traces", "errors", "logs", "pods"],
				"description": "Type of debug information to retrieve. Default is 'full' for all information",
				"default": "full"
			}
		},
		"oneOf": [
			{"required": ["uuid"]},
			{"required": ["name", "namespace"]}
		]
	}`
}

func (t *JobDebugTool) Run(input string) (string, error) {
	var args struct {
		UUID      string `json:"uuid"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		DebugType string `json:"debug_type"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid input: %v", err)
	}

	// Default to full debug info
	if args.DebugType == "" {
		args.DebugType = "full"
	}

	// If UUID is provided, first find the job
	if args.UUID != "" {
		job, err := t.findJobByUUID(args.UUID, args.Namespace)
		if err != nil {
			return "", err
		}
		// Extract name and namespace from found job
		if jobMap, ok := job.(map[string]interface{}); ok {
			if name, ok := jobMap["name"].(string); ok {
				args.Name = name
			}
			if namespace, ok := jobMap["namespace"].(string); ok {
				args.Namespace = namespace
			}
			// If metadata exists, try to get from there
			if metadata, ok := jobMap["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["name"].(string); ok {
					args.Name = name
				}
				if namespace, ok := metadata["namespace"].(string); ok {
					args.Namespace = namespace
				}
			}
		}
	}

	// Validate we have name and namespace
	if args.Name == "" || args.Namespace == "" {
		return "", fmt.Errorf("job name and namespace are required")
	}

	// Get debug information based on type
	switch args.DebugType {
	case "full":
		return t.getFullDebugInfo(args.Namespace, args.Name)
	case "traces":
		return t.getTraces(args.Namespace, args.Name)
	case "errors":
		return t.getErrors(args.Namespace, args.Name)
	case "logs":
		return t.getSandboxLogs(args.Namespace, args.Name)
	case "pods":
		return t.getJobPods(args.Namespace, args.Name)
	default:
		return "", fmt.Errorf("invalid debug_type: %s", args.DebugType)
	}
}

func (t *JobDebugTool) findJobByUUID(uuid, namespace string) (interface{}, error) {
	url := fmt.Sprintf("http://localhost:8080/jobs/uuid/%s", uuid)
	if namespace != "" {
		url += fmt.Sprintf("?namespace=%s", namespace)
	}

	resp, err := utils.GetHTTP(url)
	if err != nil {
		return nil, fmt.Errorf("failed to find job by UUID: %v", err)
	}

	var job interface{}
	if err := json.Unmarshal([]byte(resp), &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %v", err)
	}

	return job, nil
}

func (t *JobDebugTool) getFullDebugInfo(namespace, name string) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/jobs/%s/%s/debug", namespace, name)
	
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to get job debug info: %v", err)
	}

	// Parse and format the debug info
	var debugInfo map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &debugInfo); err != nil {
		return resp, nil // Return raw response if parsing fails
	}

	return t.formatDebugInfo(debugInfo), nil
}

func (t *JobDebugTool) getTraces(namespace, name string) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/jobs/%s/%s/traces", namespace, name)
	
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to get job traces: %v", err)
	}

	return resp, nil
}

func (t *JobDebugTool) getErrors(namespace, name string) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/jobs/%s/%s/errors", namespace, name)
	
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to get job errors: %v", err)
	}

	return resp, nil
}

func (t *JobDebugTool) getSandboxLogs(namespace, name string) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/jobs/%s/%s/sandbox", namespace, name)
	
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to get sandbox logs: %v", err)
	}

	return resp, nil
}

func (t *JobDebugTool) getJobPods(namespace, name string) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/jobs/%s/%s/pods", namespace, name)
	
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to get job pods: %v", err)
	}

	return resp, nil
}

func (t *JobDebugTool) formatDebugInfo(debugInfo map[string]interface{}) string {
	var result strings.Builder

	// Format job summary
	if job, ok := debugInfo["job"].(map[string]interface{}); ok {
		result.WriteString("=== Job Summary ===\n")
		result.WriteString(fmt.Sprintf("Name: %s/%s\n", job["namespace"], job["name"]))
		result.WriteString(fmt.Sprintf("UUID: %s\n", job["uuid"]))
		result.WriteString(fmt.Sprintf("Status: %s\n", job["status"]))
		result.WriteString("\n")
	}

	// Format trace information
	if traces, ok := debugInfo["traces"].(map[string]interface{}); ok {
		result.WriteString("=== Trace Information ===\n")
		if datadogUrl, ok := traces["datadogUrl"].(string); ok && datadogUrl != "" {
			result.WriteString(fmt.Sprintf("Datadog URL: %s\n", datadogUrl))
		}
		if traceId, ok := traces["traceId"].(string); ok && traceId != "" {
			result.WriteString(fmt.Sprintf("Trace ID: %s\n", traceId))
		}
		if traceLink, ok := traces["traceLink"].(string); ok && traceLink != "" {
			result.WriteString(fmt.Sprintf("Trace Link: %s\n", traceLink))
		}
		result.WriteString("\n")
	}

	// Format error information
	if errors, ok := debugInfo["errors"].(map[string]interface{}); ok {
		result.WriteString("=== Error Details ===\n")
		if reason, ok := errors["reason"].(string); ok && reason != "" {
			result.WriteString(fmt.Sprintf("Reason: %s\n", reason))
		}
		if message, ok := errors["message"].(string); ok && message != "" {
			result.WriteString(fmt.Sprintf("Message: %s\n", message))
		}
		if podErrors, ok := errors["podErrors"].([]interface{}); ok && len(podErrors) > 0 {
			result.WriteString("\nPod Errors:\n")
			for _, podError := range podErrors {
				if pe, ok := podError.(map[string]interface{}); ok {
					result.WriteString(fmt.Sprintf("  - Pod: %s, Container: %s\n", pe["podName"], pe["container"]))
					result.WriteString(fmt.Sprintf("    Reason: %s\n", pe["reason"]))
					result.WriteString(fmt.Sprintf("    Message: %s\n", pe["message"]))
				}
			}
		}
		result.WriteString("\n")
	}

	// Format logs information
	if logs, ok := debugInfo["logs"].(map[string]interface{}); ok {
		result.WriteString("=== Log Information ===\n")
		if sandboxPath, ok := logs["sandboxPath"].(string); ok && sandboxPath != "" {
			result.WriteString(fmt.Sprintf("Sandbox Path: %s\n", sandboxPath))
			result.WriteString("  Available log files: std.out, std.err, decout, decerr\n")
			result.WriteString("  To read logs, I can analyze them for errors\n")
		}
		if sandboxUrl, ok := logs["sandboxUrl"].(string); ok && sandboxUrl != "" {
			result.WriteString(fmt.Sprintf("Sandbox URL: %s\n", sandboxUrl))
		}
		if logFiles, ok := logs["logFiles"].(map[string]interface{}); ok && len(logFiles) > 0 {
			result.WriteString("\nLog Files:\n")
			for name, file := range logFiles {
				result.WriteString(fmt.Sprintf("  - %s: %s\n", name, file))
			}
		}
		if containers, ok := logs["containers"].(map[string]interface{}); ok && len(containers) > 0 {
			result.WriteString("\nContainer Logs Available:\n")
			for container, logInfo := range containers {
				result.WriteString(fmt.Sprintf("  - %s: %s\n", container, logInfo))
			}
		}
		result.WriteString("\n")
	}

	// Format events
	if events, ok := debugInfo["events"].([]interface{}); ok && len(events) > 0 {
		result.WriteString("=== Events ===\n")
		for _, event := range events {
			result.WriteString(fmt.Sprintf("  - %s\n", event))
		}
		result.WriteString("\n")
	}

	// Format pods
	if pods, ok := debugInfo["pods"].([]interface{}); ok && len(pods) > 0 {
		result.WriteString("=== Associated Pods ===\n")
		for _, pod := range pods {
			if p, ok := pod.(map[string]interface{}); ok {
				result.WriteString(fmt.Sprintf("  - %s (Status: %s, Node: %s)\n", 
					p["name"], p["status"], p["node"]))
			}
		}
	}

	return result.String()
}

// readSandboxLog reads a specific sandbox log file
func (t *JobDebugTool) readSandboxLog(sandboxPath, logFile string, startLine, numLines int) (string, error) {
	url := fmt.Sprintf("http://localhost:8080/sandbox/read?path=%s&file=%s&start=%d&lines=%d", 
		sandboxPath, logFile, startLine, numLines)
	
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to read sandbox log: %v", err)
	}

	return resp, nil
}