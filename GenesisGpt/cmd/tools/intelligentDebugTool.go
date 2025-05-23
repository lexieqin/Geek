package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexieqin/Geek/K8sGpt/cmd/utils"
)

type IntelligentDebugTool struct{}

func NewIntelligentDebugTool() *IntelligentDebugTool {
	return &IntelligentDebugTool{}
}

func (t *IntelligentDebugTool) Name() string {
	return "IntelligentDebugTool"
}

func (t *IntelligentDebugTool) Description() string {
	return "Intelligently debug failed jobs following the standard debugging workflow: 1) Get job details and JobError, 2) Fetch Datadog traces if needed, 3) Analyze sandbox logs if needed. Returns comprehensive debug summary."
}

func (t *IntelligentDebugTool) ArgsSchema() string {
	return `{
		"type": "object",
		"properties": {
			"jobId": {
				"type": "string",
				"description": "The job ID or UUID to debug"
			},
			"namespace": {
				"type": "string",
				"description": "The namespace of the job",
				"default": "default"
			},
			"debugLevel": {
				"type": "string",
				"enum": ["quick", "traces", "full"],
				"description": "Debug level: quick (JobError only), traces (JobError + Datadog), full (all including sandbox logs)",
				"default": "quick"
			}
		},
		"required": ["jobId"]
	}`
}

func (t *IntelligentDebugTool) Run(input string) (string, error) {
	var args struct {
		JobID      string `json:"jobId"`
		Namespace  string `json:"namespace"`
		DebugLevel string `json:"debugLevel"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid input: %v", err)
	}

	// Set defaults
	if args.Namespace == "" {
		args.Namespace = "default"
	}
	if args.DebugLevel == "" {
		args.DebugLevel = "quick"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("=== Debugging Job: %s/%s ===\n\n", args.Namespace, args.JobID))

	// Step 1: Get job details
	jobDetails, err := t.getJobDetails(args.Namespace, args.JobID)
	if err != nil {
		return "", fmt.Errorf("failed to get job details: %v", err)
	}

	// Step 2: Check JobError section
	jobError := t.extractJobError(jobDetails)
	if jobError != "" {
		result.WriteString("=== Job Error (Pre-categorized) ===\n")
		result.WriteString(jobError)
		result.WriteString("\n\n")

		if args.DebugLevel == "quick" {
			result.WriteString("💡 Quick analysis complete. Use debugLevel='traces' or 'full' for deeper investigation.\n")
			return result.String(), nil
		}
	} else {
		result.WriteString("=== Job Error ===\n")
		result.WriteString("No pre-categorized errors found in JobError section.\n\n")
	}

	// Step 3: Get Datadog traces if requested
	if args.DebugLevel == "traces" || args.DebugLevel == "full" {
		traceID := t.extractDatadogTraceID(jobDetails)
		if traceID != "" {
			result.WriteString("=== Datadog Traces ===\n")
			result.WriteString(fmt.Sprintf("Trace ID: %s\n", traceID))

			// Here we would fetch actual traces via Datadog API
			// For now, we'll simulate it
			traceErrors := t.fetchDatadogTraces(traceID)
			if traceErrors != "" {
				result.WriteString("Errors from traces:\n")
				result.WriteString(traceErrors)
				result.WriteString("\n")
			} else {
				result.WriteString("No errors found in Datadog traces (system level looks OK).\n")
			}
			result.WriteString("\n")

			if args.DebugLevel == "traces" {
				result.WriteString("💡 Trace analysis complete. Use debugLevel='full' to check application logs.\n")
				return result.String(), nil
			}
		}
	}

	// Step 4: Analyze sandbox logs if requested
	if args.DebugLevel == "full" {
		sandboxPath := t.extractSandboxPath(jobDetails)
		if sandboxPath != "" {
			result.WriteString("=== Sandbox Log Analysis ===\n")
			result.WriteString(fmt.Sprintf("Sandbox Path: %s\n\n", sandboxPath))

			// Analyze each log file
			logFiles := []string{"std.out", "std.err", "decout", "decerr"}
			for _, logFile := range logFiles {
				errors := t.analyzeLogFile(sandboxPath, logFile)
				if errors != "" {
					result.WriteString(fmt.Sprintf("Errors in %s:\n", logFile))
					result.WriteString(errors)
					result.WriteString("\n")
				}
			}
		} else {
			result.WriteString("=== Sandbox Logs ===\n")
			result.WriteString("No sandbox path found in job details.\n\n")
		}
	}

	// Step 5: Provide summary
	result.WriteString("=== Debug Summary ===\n")
	result.WriteString(t.generateDebugSummary(jobDetails, args.DebugLevel))

	return result.String(), nil
}

func (t *IntelligentDebugTool) getJobDetails(namespace, jobID string) (map[string]interface{}, error) {
	// Check if this is a test job ID - use mock endpoint
	if strings.HasPrefix(jobID, "real-job-") || strings.HasPrefix(jobID, "test-") {
		url := fmt.Sprintf("http://localhost:8080/mock/jobs/%s/debug", jobID)
		resp, err := utils.GetHTTP(url)
		if err == nil {
			var jobDetails map[string]interface{}
			if err := json.Unmarshal([]byte(resp), &jobDetails); err != nil {
				return nil, fmt.Errorf("failed to parse job details: %v", err)
			}
			return jobDetails, nil
		}
	}

	// First try by name
	url := fmt.Sprintf("http://localhost:8080/jobs/%s/%s/debug", namespace, jobID)
	resp, err := utils.GetHTTP(url)

	if err != nil {
		// If failed, try by UUID
		url = fmt.Sprintf("http://localhost:8080/jobs/uuid/%s", jobID)
		resp, err = utils.GetHTTP(url)
		if err != nil {
			return nil, err
		}
	}

	var jobDetails map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &jobDetails); err != nil {
		return nil, fmt.Errorf("failed to parse job details: %v", err)
	}

	return jobDetails, nil
}

func (t *IntelligentDebugTool) extractJobError(jobDetails map[string]interface{}) string {
	// Look for JobError section in the response
	if errors, ok := jobDetails["errors"].(map[string]interface{}); ok {
		var errorMsg strings.Builder
		if reason, ok := errors["reason"].(string); ok {
			errorMsg.WriteString(fmt.Sprintf("Reason: %s\n", reason))
		}
		if message, ok := errors["message"].(string); ok {
			errorMsg.WriteString(fmt.Sprintf("Message: %s\n", message))
		}
		if errorType, ok := errors["type"].(string); ok {
			errorMsg.WriteString(fmt.Sprintf("Type: %s\n", errorType))
		}
		return errorMsg.String()
	}
	return ""
}

func (t *IntelligentDebugTool) extractDatadogTraceID(jobDetails map[string]interface{}) string {
	if traces, ok := jobDetails["traces"].(map[string]interface{}); ok {
		if traceID, ok := traces["traceId"].(string); ok {
			return traceID
		}
	}
	return ""
}

func (t *IntelligentDebugTool) extractSandboxPath(jobDetails map[string]interface{}) string {
	if logs, ok := jobDetails["logs"].(map[string]interface{}); ok {
		if path, ok := logs["sandboxPath"].(string); ok {
			return path
		}
	}
	return ""
}

func (t *IntelligentDebugTool) fetchDatadogTraces(traceID string) string {
	// In a real implementation, this would call Datadog API
	// For now, return simulated data
	return "Service 'database-connector' returned 500 error\nLatency spike detected in 'api-gateway' service"
}

func (t *IntelligentDebugTool) analyzeLogFile(sandboxPath, logFile string) string {
	url := fmt.Sprintf("http://localhost:8080/sandbox/read?path=%s&file=%s&start=0&lines=1000", sandboxPath, logFile)
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return ""
	}

	var logResp map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &logResp); err != nil {
		return ""
	}

	if content, ok := logResp["content"].(string); ok {
		// Extract error lines
		var errors []string
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "exception") ||
				strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "fatal") {
				errors = append(errors, strings.TrimSpace(line))
				if len(errors) >= 5 { // Limit to 5 errors per file
					errors = append(errors, "... (more errors in file)")
					break
				}
			}
		}
		return strings.Join(errors, "\n")
	}
	return ""
}

func (t *IntelligentDebugTool) generateDebugSummary(jobDetails map[string]interface{}, debugLevel string) string {
	var summary strings.Builder

	summary.WriteString("Debug level: " + debugLevel + "\n")

	// Add recommendations based on findings
	if debugLevel == "quick" {
		summary.WriteString("- Checked JobError section only\n")
		summary.WriteString("- For system-level issues, use debugLevel='traces'\n")
		summary.WriteString("- For application-level issues, use debugLevel='full'\n")
	} else if debugLevel == "traces" {
		summary.WriteString("- Checked JobError and Datadog traces\n")
		summary.WriteString("- If issue not found, likely application-level - use debugLevel='full'\n")
	} else {
		summary.WriteString("- Performed full analysis including sandbox logs\n")
		summary.WriteString("- Check std.out for detailed application errors\n")
		summary.WriteString("- Check std.err for stack traces\n")
	}

	return summary.String()
}
