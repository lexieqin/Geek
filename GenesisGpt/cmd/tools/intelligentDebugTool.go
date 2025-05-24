package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexieqin/Geek/GenesisGpt/cmd/utils"
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
			"tenant": {
				"type": "string",
				"description": "The tenant name",
				"default": "default-tenant"
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
		Tenant     string `json:"tenant"`
		Namespace  string `json:"namespace"`
		DebugLevel string `json:"debugLevel"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid input: %v", err)
	}

	// Set defaults
	if args.Tenant == "" {
		args.Tenant = "default-tenant"
	}
	if args.Namespace == "" {
		args.Namespace = "default"
	}
	if args.DebugLevel == "" {
		args.DebugLevel = "quick"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("=== Debugging Job: %s (Tenant: %s) ===\n\n", args.JobID, args.Tenant))

	// Step 1: Get job details
	jobDetails, err := t.getJobDetails(args.Tenant, args.Namespace, args.JobID)
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
			result.WriteString(fmt.Sprintf("Sandbox Path: %s\n", sandboxPath))
			result.WriteString("Note: In real scenario, would navigate to csi-35b32b03db27ff2bad14579ebc29e3f67602aa1b7171eb061d66967c46c7cc16\n\n")

			// For demo, we know containers.log is available
			// In production, this would list files in the sandbox directory
			result.WriteString("Available log files:\n")
			result.WriteString("- containers.log\n\n")

			// Analyze containers.log
			errors := t.analyzeLogFile(sandboxPath, "containers.log")
			if errors != "" {
				result.WriteString("Critical errors found (showing first 3):\n")
				errorLines := strings.Split(errors, "\n")
				for i, line := range errorLines {
					if i >= 3 {
						result.WriteString("... (more errors in file)\n")
						break
					}
					result.WriteString(fmt.Sprintf("- %s\n", line))
				}
				result.WriteString("\n")
			} else {
				result.WriteString("No critical errors found in containers.log\n")
			}

			// Use smart analysis for deeper insights
			smartAnalysis := t.getSmartLogAnalysis(sandboxPath)
			if smartAnalysis != "" {
				result.WriteString("\nSmart Log Analysis Summary:\n")
				result.WriteString(smartAnalysis)
				result.WriteString("\n")
			}
		} else {
			result.WriteString("=== Sandbox Logs ===\n")
			result.WriteString("No sandbox path found in job details.\n\n")
		}
	}

	// Step 5: Provide summary
	result.WriteString("=== Debug Summary ===\n")
	result.WriteString(t.generateDebugSummary(jobDetails, args.DebugLevel))
	result.WriteString("\n")
	
	// Format the complete debug report
	debugReport := result.String()
	
	// Return with clear structure
	return fmt.Sprintf("Debug Report for Job %s:\n\n%s", args.JobID, debugReport), nil
}

func (t *IntelligentDebugTool) getJobDetails(tenant, namespace, jobID string) (map[string]interface{}, error) {
	// Use the new static data endpoint for demo/test purposes
	// The trace=true flag provides additional debugging information
	url := fmt.Sprintf("http://localhost:8080/tenant/%s/jobs?requuid=%s&trace=true", tenant, jobID)
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %v", err)
	}

	var jobDetails map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &jobDetails); err != nil {
		return nil, fmt.Errorf("failed to parse job details: %v", err)
	}

	return jobDetails, nil
}

func (t *IntelligentDebugTool) extractJobError(jobDetails map[string]interface{}) string {
	// Look for JobError section in the response
	if jobError, ok := jobDetails["jobError"].(map[string]interface{}); ok {
		if errMessages, ok := jobError["errMessage"].([]interface{}); ok {
			var errorMsg strings.Builder
			for _, errMsg := range errMessages {
				if errMap, ok := errMsg.(map[string]interface{}); ok {
					if error, ok := errMap["error"].(map[string]interface{}); ok {
						if category, ok := error["category"].(string); ok {
							errorMsg.WriteString(fmt.Sprintf("Category: %s\n", category))
						}
						if subCategory, ok := error["sub-category"].(string); ok {
							errorMsg.WriteString(fmt.Sprintf("Sub-category: %s\n", subCategory))
						}
						if component, ok := error["component"].(string); ok {
							errorMsg.WriteString(fmt.Sprintf("Component: %s\n", component))
						}
						if message, ok := error["message"].(string); ok {
							errorMsg.WriteString(fmt.Sprintf("Message: %s\n", message))
						}
						errorMsg.WriteString("\n")
					}
				}
			}
			return errorMsg.String()
		}
	}
	return ""
}

func (t *IntelligentDebugTool) extractDatadogTraceID(jobDetails map[string]interface{}) string {
	// Extract trace ID from contextData
	if contextData, ok := jobDetails["contextData"].(map[string]interface{}); ok {
		if traceURL, ok := contextData["Genesis-TraceID"].(string); ok {
			// Extract trace ID from URL
			// URL format: https://company-qa.datadoghq.com/apm/trace/81325fc3b05e4d9aada2d2399aebe135
			parts := strings.Split(traceURL, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}

func (t *IntelligentDebugTool) extractSandboxPath(jobDetails map[string]interface{}) string {
	// Look for sandbox log links in jobLogLinks
	if jobLogLinks, ok := jobDetails["jobLogLinks"].(map[string]interface{}); ok {
		// Get the first log link
		if logLink, ok := jobLogLinks["logLink"].(string); ok {
			// Extract path from URL like:
			// http://genesis.dev.companyinc.com:9101/sandboxlogs/#/katbox/browse?path=/csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704&hostip=000.000.000.000
			if strings.Contains(logLink, "path=") {
				parts := strings.Split(logLink, "path=")
				if len(parts) > 1 {
					pathAndParams := parts[1]
					// Extract just the path part before &
					pathParts := strings.Split(pathAndParams, "&")
					return pathParts[0]
				}
			}
		}
	}
	return "/csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704" // Default for demo
}

func (t *IntelligentDebugTool) fetchDatadogTraces(traceID string) string {
	// Call our static datadog trace endpoint
	url := fmt.Sprintf("http://localhost:8080/api/datadog/trace/%s", traceID)
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return fmt.Sprintf("Failed to fetch traces: %v", err)
	}

	var traceData map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &traceData); err != nil {
		return "Failed to parse trace data"
	}

	// Extract errors from trace data
	var errorSpans []map[string]interface{}
	
	// Navigate to the actual spans location: data.attributes.spans
	var spans []interface{}
	if data, ok := traceData["data"].(map[string]interface{}); ok {
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			if spansData, ok := attributes["spans"].([]interface{}); ok {
				spans = spansData
			}
		}
	}
	
	// If not found in nested structure, try top level (for backward compatibility)
	if len(spans) == 0 {
		if spansData, ok := traceData["spans"].([]interface{}); ok {
			spans = spansData
		}
	}
	
	if len(spans) > 0 {
		for _, span := range spans {
			if spanMap, ok := span.(map[string]interface{}); ok {
				hasError := false
				errorDetails := ""
				
				// Check meta fields for error information
				if meta, ok := spanMap["meta"].(map[string]interface{}); ok {
					// Check for OpenTelemetry error status
					if otelStatus, ok := meta["otel.status_code"].(string); ok && otelStatus == "ERROR" {
						hasError = true
					}
					
					// Extract error details from meta
					if hasError {
						if errMsg, ok := meta["error.message"].(string); ok {
							errorDetails = errMsg
						} else if errMsg, ok := meta["err.msg"].(string); ok {
							errorDetails = errMsg
						}
						
						// Add error type and category if available
						if errType, ok := meta["err.type"].(string); ok {
							errorDetails = fmt.Sprintf("[%s] %s", errType, errorDetails)
						}
						if errSubCat, ok := meta["err.sub_category"].(string); ok {
							errorDetails = fmt.Sprintf("[%s] %s", errSubCat, errorDetails)
						}
					}
				}
				
				// Also check numeric error field
				if errorFlag, ok := spanMap["error"].(float64); ok && errorFlag == 1 {
					hasError = true
				}
				
				if hasError {
					errorSpan := map[string]interface{}{
						"service": spanMap["service"],
						"operation": spanMap["name"],  // Using "name" field from actual trace
						"resource": spanMap["resource"],
						"error": errorDetails,
					}
					errorSpans = append(errorSpans, errorSpan)
				}
			}
		}
	}

	if len(errorSpans) > 0 {
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Found %d error spans in trace:\n\n", len(errorSpans)))
		
		// Find root cause - usually ppregistrator or the first error
		var rootCause map[string]interface{}
		for _, span := range errorSpans {
			if service, ok := span["service"].(string); ok && service == "ppregistrator" {
				rootCause = span
				break
			}
		}
		if rootCause == nil && len(errorSpans) > 0 {
			rootCause = errorSpans[0]
		}
		
		// Show root cause prominently
		if rootCause != nil {
			result.WriteString("Root Cause:\n")
			if service, ok := rootCause["service"].(string); ok {
				result.WriteString(fmt.Sprintf("  Service: %s\n", service))
			}
			if resource, ok := rootCause["resource"].(string); ok {
				result.WriteString(fmt.Sprintf("  Resource: %s\n", resource))
			}
			if error, ok := rootCause["error"].(string); ok && error != "" {
				result.WriteString(fmt.Sprintf("  Error: %s\n", error))
			}
			result.WriteString("\n")
		}
		
		// Show error propagation chain (unique services only)
		if len(errorSpans) > 1 {
			result.WriteString("Error Propagation Chain:\n  ")
			seenServices := make(map[string]bool)
			var chain []string
			
			// Start from the root cause
			for _, span := range errorSpans {
				if service, ok := span["service"].(string); ok {
					if !seenServices[service] {
						seenServices[service] = true
						chain = append(chain, service)
					}
				}
			}
			result.WriteString(strings.Join(chain, " → "))
			result.WriteString("\n")
		}
		
		return result.String()
	}
	return "No error spans found in Datadog traces (all spans have OK status)"
}

func (t *IntelligentDebugTool) analyzeLogFile(sandboxPath, logFile string) string {
	// For the demo, directly use the sandbox log endpoint
	url := fmt.Sprintf("http://localhost:8080/api/sandbox/logs?path=%s&file=%s", sandboxPath, logFile)
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return ""
	}
	
	// Simple error extraction
	var errors []string
	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "exception") ||
			strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "fatal") {
			errors = append(errors, strings.TrimSpace(line))
			if len(errors) >= 5 {
				errors = append(errors, "... (more errors in file)")
				break
			}
		}
	}
	return strings.Join(errors, "\n")
}

func (t *IntelligentDebugTool) getSmartLogAnalysis(sandboxPath string) string {
	// Use the smart log endpoint for comprehensive analysis
	url := fmt.Sprintf("http://localhost:8080/api/sandbox/logs/smart?path=%s", sandboxPath)
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return ""
	}

	var logAnalysis map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &logAnalysis); err != nil {
		return ""
	}

	var summary strings.Builder
	
	// Extract summary information
	if summaryData, ok := logAnalysis["summary"].(map[string]interface{}); ok {
		if counts, ok := summaryData["counts"].(map[string]interface{}); ok {
			if totalCritical, ok := counts["total_critical"].(float64); ok {
				summary.WriteString(fmt.Sprintf("Total critical issues: %d\n", int(totalCritical)))
			}
			if errors, ok := counts["errors"].(float64); ok {
				summary.WriteString(fmt.Sprintf("Errors: %d\n", int(errors)))
			}
			if warnings, ok := counts["warnings"].(float64); ok {
				summary.WriteString(fmt.Sprintf("Warnings: %d\n", int(warnings)))
			}
		}
		
		if errorCategories, ok := summaryData["error_categories"].(map[string]interface{}); ok {
			summary.WriteString("\nError Categories:\n")
			for category, count := range errorCategories {
				if countFloat, ok := count.(float64); ok {
					summary.WriteString(fmt.Sprintf("- %s: %d\n", category, int(countFloat)))
				}
			}
		}
	}
	
	return summary.String()
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

// Helper function to safely extract string from map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}