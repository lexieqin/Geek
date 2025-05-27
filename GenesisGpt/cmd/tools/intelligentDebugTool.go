package tools

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/lexieqin/Geek/GenesisGpt/cmd/config"
	"github.com/lexieqin/Geek/GenesisGpt/cmd/utils"
	apiconfig "github.com/lexieqin/Geek/GenesisGpt/config"
)

type IntelligentDebugTool struct{
	apiEndpoints *apiconfig.APIEndpoints
}

// Removed SandboxInfo struct - we'll use the URL directly

func NewIntelligentDebugTool() *IntelligentDebugTool {
	return &IntelligentDebugTool{
		apiEndpoints: apiconfig.GetAPIEndpoints(),
	}
}

func (t *IntelligentDebugTool) Name() string {
	return "IntelligentDebugTool"
}

func (t *IntelligentDebugTool) Description() string {
	return "Debug failed deployment platform jobs (NOT Kubernetes Jobs) by analyzing job details, errors, Datadog traces, and sandbox logs. IMPORTANT: This tool returns a structured report that MUST be preserved exactly as formatted. Do not summarize or reorganize the output before presenting it. The tool will return actual data with specific timestamps, error messages, and trace IDs - never generate these yourself."
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
		sandboxURL := t.extractSandboxURL(jobDetails)
		if sandboxURL != "" {
			result.WriteString("=== Sandbox Log Analysis ===\n")
			result.WriteString(fmt.Sprintf("Sandbox URL: %s\n\n", sandboxURL))

			// Check if sandbox logs are available
			logAvailable, logFiles := t.checkSandboxLogsAvailable(sandboxURL)
			if !logAvailable {
				result.WriteString("❌ Sandbox logs not available\n")
				result.WriteString("This usually indicates the deployment failed at an early stage before application logs were generated.\n")
				result.WriteString("Check the Job Error section above for deployment-level issues.\n\n")
			} else {
				result.WriteString("📁 Discovered log files:\n")
				if len(logFiles) > 0 {
					for _, file := range logFiles {
						result.WriteString(fmt.Sprintf("  • %s\n", file))
					}
					result.WriteString("\n")

					// Analyze each discovered log file layer by layer
					result.WriteString("📋 Layer-by-Layer Log Analysis:\n\n")
					
					// Priority order for log analysis
					priorityFiles := []string{"containers.log", "std.err", "std.out", "deploy.log"}
					analyzedFiles := make(map[string]bool)
					
					// First analyze priority files in order
					for _, priorityFile := range priorityFiles {
						if t.containsFile(logFiles, priorityFile) {
							result.WriteString(fmt.Sprintf("▶ Analyzing %s:\n", priorityFile))
							analysis := t.analyzeSpecificLogFile(sandboxURL, priorityFile)
							result.WriteString(analysis)
							result.WriteString("\n")
							analyzedFiles[priorityFile] = true
						}
					}
					
					// Then analyze any other discovered files
					for _, file := range logFiles {
						if !analyzedFiles[file] {
							result.WriteString(fmt.Sprintf("▶ Analyzing %s:\n", file))
							analysis := t.analyzeSpecificLogFile(sandboxURL, file)
							result.WriteString(analysis)
							result.WriteString("\n")
						}
					}

					// Provide comprehensive smart analysis across all logs
					smartAnalysis := t.getSmartLogAnalysis(sandboxURL)
					if smartAnalysis != "" {
						result.WriteString("📊 Aggregated Analysis Across All Logs:\n")
						result.WriteString(smartAnalysis)
						result.WriteString("\n")
					}
					
					// Root cause analysis based on all findings
					rootCause := t.determineRootCauseFromLogs(logFiles, sandboxURL)
					if rootCause != "" {
						result.WriteString("🎯 Root Cause Analysis:\n")
						result.WriteString(rootCause)
						result.WriteString("\n")
					}
				} else {
					result.WriteString("  ❌ No log files discovered in sandbox\n")
					result.WriteString("  This indicates the application never started or sandbox was not properly initialized.\n\n")
				}
			}
		} else {
			result.WriteString("=== Sandbox Logs ===\n")
			result.WriteString("❌ No sandbox path found in job details\n")
			result.WriteString("This indicates the job failed before sandbox initialization.\n")
			result.WriteString("Check the Job Error section for early-stage deployment issues.\n\n")
		}
	}

	// Step 5: Provide summary
	result.WriteString("=== Debug Summary ===\n")
	result.WriteString(t.generateDebugSummary(jobDetails, args.DebugLevel))
	result.WriteString("\n")
	
	// Format the complete debug report
	debugReport := result.String()
	
	// Return with clear structure and unique markers to prevent hallucination
	timestamp := time.Now().Unix()
	return fmt.Sprintf("=== ACTUAL TOOL OUTPUT START [TS:%d] ===\nDebug Report for Job %s:\n\n%s\n=== ACTUAL TOOL OUTPUT END [TS:%d] ===", 
		timestamp, args.JobID, debugReport, timestamp), nil
}

func (t *IntelligentDebugTool) getJobDetails(tenant, namespace, jobID string) (map[string]interface{}, error) {
	// Use the correct job API endpoint: /tenant/{tenant}/job/{jobid}?trace=true
	// This calls your deployment platform's job API, not Kubernetes Job resources
	url := fmt.Sprintf("%s/tenant/%s/job/%s?trace=true", t.apiEndpoints.JobAPI, tenant, jobID)
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details from deployment platform: %v", err)
	}

	var jobDetails map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &jobDetails); err != nil {
		return nil, fmt.Errorf("failed to parse job details response: %v", err)
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

func (t *IntelligentDebugTool) extractSandboxURL(jobDetails map[string]interface{}) string {
	// Look for DEPLOY_SB_SAMPLE_LINK in dzStatus -> <dz> -> dzProgress -> msg
	if dzStatus, ok := jobDetails["dzStatus"].(map[string]interface{}); ok {
		// Iterate through each DZ (data zone)
		for dzName, dzData := range dzStatus {
			// Skip GFSM as it's the global FSM, not a deployment zone
			if dzName == "GFSM" {
				continue
			}
			
			if dzMap, ok := dzData.(map[string]interface{}); ok {
				if dzProgress, ok := dzMap["dzProgress"].([]interface{}); ok {
					// Look through each step in dzProgress
					for _, step := range dzProgress {
						if stepMap, ok := step.(map[string]interface{}); ok {
							if msgs, ok := stepMap["msg"].([]interface{}); ok {
								// Look through messages in this step
								for _, msg := range msgs {
									if msgMap, ok := msg.(map[string]interface{}); ok {
										if msgType, ok := msgMap["msgType"].(string); ok && msgType == "DEPLOY_SB_SAMPLE_LINK" {
											// Found the sandbox link message
											if fields, ok := msgMap["fields"].(map[string]interface{}); ok {
												// Check different status fields (STARTING, SUCCESS, FAILED, etc.)
												for _, value := range fields {
													if links, ok := value.([]interface{}); ok && len(links) > 0 {
														if link, ok := links[0].(string); ok {
															return link // Return the full URL directly
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Fallback: check jobLogLinks (this is for FSM logs, not application logs)
	// But we can try to use them if no sandbox link is found
	if jobLogLinks, ok := jobDetails["jobLogLinks"].(map[string]interface{}); ok {
		// Skip GFSM and look for actual deployment zone logs
		for dzName, links := range jobLogLinks {
			if dzName != "GFSM" {
				if linkArray, ok := links.([]interface{}); ok && len(linkArray) > 0 {
					if firstLink, ok := linkArray[0].(map[string]interface{}); ok {
						if _, ok := firstLink["logLink"].(string); ok {
							// Note: These are FSM logs, not application logs
							// Skip FSM logs for now
							continue
						}
					}
				}
			}
		}
	}
	
	return "" // Return empty string when no sandbox URL found
}

// extractPathFromURL extracts the path parameter from sandbox URL
// Used only in mock mode to work with our mock endpoints
func extractPathFromURL(url string) string {
	if strings.Contains(url, "path=") {
		parts := strings.Split(url, "path=")
		if len(parts) > 1 {
			pathAndParams := parts[1]
			// Extract just the path part before &
			pathParts := strings.Split(pathAndParams, "&")
			return pathParts[0]
		}
	}
	return ""
}

func (t *IntelligentDebugTool) fetchDatadogTraces(traceID string) string {
	// Call datadog trace endpoint
	url := fmt.Sprintf("%s/trace/%s", t.apiEndpoints.TraceAPI, traceID)
	resp, err := utils.GetHTTPWithAuth(url, "datadog")
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
					// Check for OpenTelemetry error status (case insensitive)
					if otelStatus, ok := meta["otel.status_code"].(string); ok && strings.EqualFold(otelStatus, "ERROR") {
						hasError = true
					}
					
					// Extract error details from meta (check regardless of hasError)
					// Try multiple fields for error message
					if errMsg, ok := meta["error.message"].(string); ok && errMsg != "" {
						errorDetails = errMsg
					} else if errMsg, ok := meta["err.msg"].(string); ok && errMsg != "" {
						errorDetails = errMsg
					} else if errMsg, ok := meta["message"].(string); ok && errMsg != "" {
						errorDetails = errMsg
					} else if errStack, ok := meta["err.stack"].(string); ok && errStack != "" {
						// Extract first line of stack trace as error message
						if lines := strings.Split(errStack, "\n"); len(lines) > 0 {
							errorDetails = lines[0]
						}
					}
					
					// Add error type and category if available
					if errType, ok := meta["err.type"].(string); ok && errType != "" {
						if errorDetails == "" {
							errorDetails = fmt.Sprintf("[%s]", errType)
						} else {
							errorDetails = fmt.Sprintf("[%s] %s", errType, errorDetails)
						}
					}
					if errSubCat, ok := meta["err.sub_category"].(string); ok && errSubCat != "" {
						if errorDetails == "" {
							errorDetails = fmt.Sprintf("[%s]", errSubCat)
						} else {
							errorDetails = fmt.Sprintf("[%s] %s", errSubCat, errorDetails)
						}
					}
					
					// If still no error details, check for error in span description
					if errorDetails == "" {
						if spanName, ok := spanMap["name"].(string); ok && strings.Contains(strings.ToLower(spanName), "error") {
							errorDetails = spanName
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
		
		// Add high-level summary first
		result.WriteString("📍 Summary: ")
		
		// Look for the most specific error (usually ppregistrator or environment errors)
		summaryFound := false
		for _, span := range errorSpans {
			if errorMsg, ok := span["error"].(string); ok && errorMsg != "" {
				// Look for specific error patterns that indicate root cause
				if strings.Contains(errorMsg, "LCM Error:") || 
				   strings.Contains(errorMsg, "Unable to properly resolve") ||
				   strings.Contains(errorMsg, "Environment-Error") {
					// Extract just the core error message
					if idx := strings.Index(errorMsg, "LCM Error:"); idx != -1 {
						result.WriteString(errorMsg[idx:])
					} else if idx := strings.Index(errorMsg, "Unable to"); idx != -1 {
						result.WriteString(errorMsg[idx:])
					} else {
						// Remove error categories and show just the message
						msg := errorMsg
						msg = strings.TrimPrefix(msg, "[Environment-Error] ")
						msg = strings.TrimPrefix(msg, "[System-Error] ")
						result.WriteString(msg)
					}
					summaryFound = true
					break
				}
			}
		}
		
		// If no specific root cause found, use the first ppregistrator error
		if !summaryFound {
			for _, span := range errorSpans {
				if service, ok := span["service"].(string); ok && service == "ppregistrator" {
					if errorMsg, ok := span["error"].(string); ok && errorMsg != "" {
						// Extract core message
						msg := errorMsg
						// Remove common prefixes
						prefixes := []string{"[Environment-Error] ", "[System-Error] ", "[Deploy-Driver-Error] ", "[Fsm-Error] "}
						for _, prefix := range prefixes {
							msg = strings.TrimPrefix(msg, prefix)
						}
						result.WriteString(msg)
						summaryFound = true
						break
					}
				}
			}
		}
		
		// Fallback to replica set failure if nothing else found
		if !summaryFound {
			for _, span := range errorSpans {
				if errorMsg, ok := span["error"].(string); ok && strings.Contains(errorMsg, "replica set") && strings.Contains(errorMsg, "failed") {
					// Just show the replica set failure
					if idx := strings.Index(errorMsg, "replica set"); idx != -1 {
						result.WriteString(errorMsg[idx:])
					} else {
						result.WriteString("Deployment failed")
					}
					break
				}
			}
		}
		
		result.WriteString("\n\n")
		
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
		
		// Show error propagation chain with error details
		if len(errorSpans) > 1 {
			result.WriteString("Error Propagation Chain:\n")
			seenServices := make(map[string]bool)
			
			for _, span := range errorSpans {
				if service, ok := span["service"].(string); ok {
					if !seenServices[service] {
						seenServices[service] = true
						result.WriteString(fmt.Sprintf("  • %s", service))
						
						// Add error details if available
						if errorMsg, ok := span["error"].(string); ok && errorMsg != "" {
							result.WriteString(fmt.Sprintf(": %s", errorMsg))
						}
						result.WriteString("\n")
					}
				}
			}
		}
		
		return result.String()
	}
	return "No error spans found in Datadog traces (all spans have OK status)"
}

func (t *IntelligentDebugTool) checkSandboxLogsAvailable(sandboxURL string) (bool, []string) {
	// In mock mode, use our mock endpoint instead
	var url string
	if config.IsMockMode() {
		// Extract path from sandbox URL and use mock endpoint
		path := extractPathFromURL(sandboxURL)
		url = fmt.Sprintf("%s/logs/list?path=%s", t.apiEndpoints.SandboxAPI, path)
	} else {
		// In production, use the sandbox URL directly
		url = sandboxURL + "&action=list"
	}
	resp, err := utils.GetHTTP(url)
	if err != nil {
		// If listing fails, sandbox might not exist or be accessible
		return false, nil
	}

	// Try to parse the response as JSON array of filenames
	var files []string
	if err := json.Unmarshal([]byte(resp), &files); err != nil {
		// If JSON parsing fails, try to parse as plain text list
		lines := strings.Split(strings.TrimSpace(resp), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "error") && !strings.Contains(line, "not found") {
				files = append(files, line)
			}
		}
	}

	return len(files) > 0, files
}

func (t *IntelligentDebugTool) containsFile(files []string, targetFile string) bool {
	for _, file := range files {
		if file == targetFile {
			return true
		}
	}
	return false
}

func (t *IntelligentDebugTool) analyzeLogFile(sandboxURL, logFile string) string {
	// In mock mode, use our mock endpoint instead
	var url string
	if config.IsMockMode() {
		path := extractPathFromURL(sandboxURL)
		url = fmt.Sprintf("%s/logs?path=%s&file=%s", t.apiEndpoints.SandboxAPI, path, logFile)
	} else {
		// In production, use the sandbox URL directly
		url = fmt.Sprintf("%s&action=read&file=%s", sandboxURL, logFile)
	}
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return fmt.Sprintf("Failed to read %s: %v", logFile, err)
	}
	
	return t.performSemanticLogAnalysis(resp, logFile)
}

type LogEntry struct {
	LineNumber  int
	Timestamp   string
	Level       string
	Message     string
	Source      string // stdout/stderr
	Component   string // container name
	Context     string // additional context
	RawLine     string
}

// performSemanticLogAnalysis provides intelligent analysis of log content
func (t *IntelligentDebugTool) performSemanticLogAnalysis(logContent, logFile string) string {
	lines := strings.Split(logContent, "\n")
	
	var criticalErrors []LogEntry
	var warnings []LogEntry
	var importantInfo []LogEntry
	var rootCauseIndicators []LogEntry
	
	// Phase 1: Parse and categorize all log entries
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		entry := t.parseLogEntry(line, i+1)
		if entry == nil {
			continue
		}
		
		// Categorize based on semantic analysis
		switch {
		case t.isCriticalError(entry):
			criticalErrors = append(criticalErrors, *entry)
		case t.isRootCauseIndicator(entry):
			rootCauseIndicators = append(rootCauseIndicators, *entry)
		case t.isWarning(entry):
			warnings = append(warnings, *entry)
		case t.isImportantInfo(entry):
			importantInfo = append(importantInfo, *entry)
		}
	}
	
	// Phase 2: Determine root cause through semantic analysis
	rootCause := t.analyzeRootCause(criticalErrors, rootCauseIndicators, importantInfo)
	
	// Phase 3: Generate intelligent summary
	var result strings.Builder
	
	if rootCause != "" {
		result.WriteString("🎯 Root Cause Analysis:\n")
		result.WriteString(fmt.Sprintf("  %s\n\n", rootCause))
	}
	
	if len(criticalErrors) > 0 {
		result.WriteString("❌ Critical Errors:\n")
		for _, err := range criticalErrors {
			result.WriteString(fmt.Sprintf("  [%s] %s\n", err.Level, err.Message))
			if err.Context != "" {
				result.WriteString(fmt.Sprintf("    Context: %s\n", err.Context))
			}
		}
		result.WriteString("\n")
	}
	
	if len(rootCauseIndicators) > 0 && rootCause == "" {
		result.WriteString("🔍 Potential Root Cause Indicators:\n")
		for _, indicator := range rootCauseIndicators[:t.min(3, len(rootCauseIndicators))] {
			result.WriteString(fmt.Sprintf("  %s\n", indicator.Message))
		}
		result.WriteString("\n")
	}
	
	if len(warnings) > 0 && len(criticalErrors) == 0 {
		result.WriteString("⚠️  Notable Warnings (may be related):\n")
		for _, warn := range warnings[:t.min(2, len(warnings))] {
			result.WriteString(fmt.Sprintf("  %s\n", warn.Message))
		}
		result.WriteString("\n")
	}
	
	if result.Len() == 0 {
		result.WriteString("✅ No critical issues found in this log file\n")
	}
	
	return result.String()
}

// parseLogEntry extracts structured information from a log line
func (t *IntelligentDebugTool) parseLogEntry(line string, lineNum int) *LogEntry {
	// Parse container log format: [container] timestamp source F [timestamp] LEVEL message
	if !strings.Contains(line, "] ") {
		return nil
	}
	
	entry := &LogEntry{
		LineNumber: lineNum,
		RawLine:    line,
	}
	
	// Extract component (container name)
	if strings.HasPrefix(line, "[") {
		endBracket := strings.Index(line, "]")
		if endBracket > 0 {
			entry.Component = line[1:endBracket]
		}
	}
	
	// Extract source (stdout/stderr)
	if strings.Contains(line, " stdout F ") {
		entry.Source = "stdout"
	} else if strings.Contains(line, " stderr F ") {
		entry.Source = "stderr"
	}
	
	// Extract the actual log message part
	parts := strings.Split(line, "] ")
	if len(parts) < 3 {
		return entry
	}
	
	logMessage := parts[len(parts)-1]
	
	// Extract log level and message from the log content
	if strings.Contains(logMessage, "] INFO ") {
		entry.Level = "INFO"
		entry.Message = strings.SplitN(logMessage, "] INFO ", 2)[1]
	} else if strings.Contains(logMessage, "] ERROR ") {
		entry.Level = "ERROR"
		entry.Message = strings.SplitN(logMessage, "] ERROR ", 2)[1]
	} else if strings.Contains(logMessage, "] WARNING ") || strings.Contains(logMessage, "] WARN ") {
		entry.Level = "WARNING"
		if strings.Contains(logMessage, "] WARNING ") {
			entry.Message = strings.SplitN(logMessage, "] WARNING ", 2)[1]
		} else {
			entry.Message = strings.SplitN(logMessage, "] WARN ", 2)[1]
		}
	} else if strings.Contains(logMessage, "] FATAL ") {
		entry.Level = "FATAL"
		entry.Message = strings.SplitN(logMessage, "] FATAL ", 2)[1]
	} else {
		// No explicit log level, use the whole message
		entry.Message = logMessage
		// Infer level from content and source
		if entry.Source == "stderr" && (strings.Contains(strings.ToLower(logMessage), "error") || 
			strings.Contains(strings.ToLower(logMessage), "failed") ||
			strings.Contains(strings.ToLower(logMessage), "exception")) {
			entry.Level = "ERROR"
		}
	}
	
	return entry
}

// isCriticalError determines if a log entry represents a critical error
func (t *IntelligentDebugTool) isCriticalError(entry *LogEntry) bool {
	if entry.Level == "ERROR" || entry.Level == "FATAL" {
		return true
	}
	
	// Check for critical patterns in stderr
	if entry.Source == "stderr" {
		lowerMsg := strings.ToLower(entry.Message)
		criticalPatterns := []string{
			"unable to", "cannot", "failed to", "connection refused",
			"permission denied", "no such file", "command not found",
			"timeout", "connection reset", "segmentation fault",
		}
		
		for _, pattern := range criticalPatterns {
			if strings.Contains(lowerMsg, pattern) {
				return true
			}
		}
	}
	
	return false
}

// isRootCauseIndicator identifies entries that indicate root causes
func (t *IntelligentDebugTool) isRootCauseIndicator(entry *LogEntry) bool {
	lowerMsg := strings.ToLower(entry.Message)
	
	// Exit codes and return codes indicate failure points
	if strings.Contains(lowerMsg, "exit ") && !strings.Contains(lowerMsg, "exit 0") {
		return true
	}
	
	if strings.Contains(lowerMsg, "rc=") && !strings.Contains(lowerMsg, "rc=0") {
		return true
	}
	
	// LCM error codes are root causes
	if strings.Contains(lowerMsg, "lcm") && strings.Contains(lowerMsg, ":") {
		return true
	}
	
	// "Unable to successfully complete" indicates a failure summary
	if strings.Contains(lowerMsg, "unable to successfully complete") {
		return true
	}
	
	return false
}

// isWarning determines if entry is a warning
func (t *IntelligentDebugTool) isWarning(entry *LogEntry) bool {
	return entry.Level == "WARNING" || entry.Level == "WARN"
}

// isImportantInfo identifies important informational entries
func (t *IntelligentDebugTool) isImportantInfo(entry *LogEntry) bool {
	if entry.Level != "INFO" {
		return false
	}
	
	lowerMsg := strings.ToLower(entry.Message)
	importantPatterns := []string{
		"lcmstepend", "lcmphase", "application name:",
		"partner_stage:", "bootstrap_token", "environment",
	}
	
	for _, pattern := range importantPatterns {
		if strings.Contains(lowerMsg, pattern) {
			return true
		}
	}
	
	return false
}

// analyzeRootCause performs semantic analysis to determine the root cause
func (t *IntelligentDebugTool) analyzeRootCause(criticalErrors, indicators, info []LogEntry) string {
	// Look for LCM error codes first - these are usually the root cause
	for _, err := range criticalErrors {
		if strings.Contains(err.Message, "LCM") && strings.Contains(err.Message, ":") {
			return fmt.Sprintf("LCM Error: %s", err.Message)
		}
	}
	
	// Look for DNS/network resolution issues
	for _, err := range criticalErrors {
		if strings.Contains(strings.ToLower(err.Message), "unable to properly resolve") ||
		   strings.Contains(strings.ToLower(err.Message), "host") {
			return fmt.Sprintf("DNS/Network Resolution Issue: %s", err.Message)
		}
	}
	
	// Look for exit codes in indicators
	for _, indicator := range indicators {
		if strings.Contains(strings.ToLower(indicator.Message), "exit ") && 
		   !strings.Contains(strings.ToLower(indicator.Message), "exit 0") {
			return fmt.Sprintf("Process Exit Failure: %s", indicator.Message)
		}
	}
	
	// Look for configuration issues
	for _, err := range criticalErrors {
		lowerMsg := strings.ToLower(err.Message)
		if strings.Contains(lowerMsg, "no such file") || 
		   strings.Contains(lowerMsg, "command not found") ||
		   strings.Contains(lowerMsg, "permission denied") {
			return fmt.Sprintf("Configuration/Environment Issue: %s", err.Message)
		}
	}
	
	// If we have critical errors but no specific root cause, return the first critical error
	if len(criticalErrors) > 0 {
		return fmt.Sprintf("Primary Error: %s", criticalErrors[0].Message)
	}
	
	return ""
}

// min helper function
func (t *IntelligentDebugTool) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (t *IntelligentDebugTool) getSmartLogAnalysis(sandboxURL string) string {
	// In mock mode, use our mock endpoint instead
	var url string
	if config.IsMockMode() {
		path := extractPathFromURL(sandboxURL)
		url = fmt.Sprintf("%s/logs/smart?path=%s", t.apiEndpoints.SandboxAPI, path)
	} else {
		// In production, use the sandbox URL directly
		url = sandboxURL + "&action=analyze"
	}
	resp, err := utils.GetHTTP(url)
	if err != nil {
		return ""
	}

	var logAnalysis map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &logAnalysis); err != nil {
		return ""
	}

	var summary strings.Builder
	
	// Extract summary information (focus on errors only, ignore warnings)
	if summaryData, ok := logAnalysis["summary"].(map[string]interface{}); ok {
		if counts, ok := summaryData["counts"].(map[string]interface{}); ok {
			if errors, ok := counts["errors"].(float64); ok && errors > 0 {
				summary.WriteString(fmt.Sprintf("Total errors found: %d\n", int(errors)))
			} else {
				summary.WriteString("No errors found in aggregated analysis\n")
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

// analyzeSpecificLogFile provides detailed analysis for a specific log file
func (t *IntelligentDebugTool) analyzeSpecificLogFile(sandboxURL, logFile string) string {
	var result strings.Builder
	
	// In mock mode, use our mock endpoint instead
	var url string
	if config.IsMockMode() {
		path := extractPathFromURL(sandboxURL)
		url = fmt.Sprintf("%s/logs?path=%s&file=%s", t.apiEndpoints.SandboxAPI, path, logFile)
	} else {
		// In production, use the sandbox URL directly
		url = fmt.Sprintf("%s&action=read&file=%s", sandboxURL, logFile)
	}
	resp, err := utils.GetHTTP(url)
	if err != nil {
		result.WriteString(fmt.Sprintf("  ⚠️  Failed to read: %v\n", err))
		return result.String()
	}
	
	// Analyze based on file type
	baseName := filepath.Base(logFile)
	switch baseName {
	case "containers.log":
		result.WriteString("  Container runtime logs (all containers in pod):\n")
	case "std.err":
		result.WriteString("  Application error output:\n")
	case "std.out":
		result.WriteString("  Application standard output:\n")
	case "deploy.log":
		if logFile != baseName {
			result.WriteString(fmt.Sprintf("  Deployment process logs (%s):\n", logFile))
		} else {
			result.WriteString("  Deployment process logs:\n")
		}
	default:
		result.WriteString(fmt.Sprintf("  %s content:\n", logFile))
	}
	
	// Extract and categorize errors (ignore warnings)
	lines := strings.Split(resp, "\n")
	errorCount := 0
	var criticalErrors []string
	var lastNonErrorLines []string
	
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		
		// Skip warnings - they're just noise
		if strings.Contains(lowerLine, "warning") || strings.Contains(lowerLine, "warn") {
			continue
		}
		
		// Keep track of last few non-error lines for context
		if !strings.Contains(lowerLine, "error") && !strings.Contains(lowerLine, "exception") &&
			!strings.Contains(lowerLine, "failed") && !strings.Contains(lowerLine, "fatal") {
			lastNonErrorLines = append(lastNonErrorLines, line)
			if len(lastNonErrorLines) > 3 {
				lastNonErrorLines = lastNonErrorLines[1:]
			}
		}
		
		// Identify actual errors and exceptions
		if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "exception") ||
			strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "fatal") {
			errorCount++
			
			// For critical errors, include context
			if errorCount <= 3 {
				// Include previous lines for context
				if len(lastNonErrorLines) > 0 && i > 0 {
					criticalErrors = append(criticalErrors, "    [Context] "+lastNonErrorLines[len(lastNonErrorLines)-1])
				}
				criticalErrors = append(criticalErrors, fmt.Sprintf("    [Line %d] %s", i+1, strings.TrimSpace(line)))
			}
		}
	}
	
	// Report findings
	if errorCount == 0 {
		result.WriteString("  ✅ No errors found\n")
	} else {
		result.WriteString(fmt.Sprintf("  ❌ Found %d error(s)\n", errorCount))
		if len(criticalErrors) > 0 {
			result.WriteString("  Critical errors with context:\n")
			for _, err := range criticalErrors {
				result.WriteString(err + "\n")
			}
			if errorCount > 3 {
				result.WriteString(fmt.Sprintf("    ... and %d more errors\n", errorCount-3))
			}
		}
	}
	
	// File-specific insights
	if logFile == "containers.log" && errorCount > 0 {
		// Check for specific container issues
		if containsContainerError(resp, "OOMKilled") {
			result.WriteString("  💥 Container was killed due to Out of Memory\n")
		}
		if containsContainerError(resp, "CrashLoopBackOff") {
			result.WriteString("  🔄 Container is in crash loop\n")
		}
	}
	
	return result.String()
}

// determineRootCauseFromLogs analyzes all logs to determine root cause
func (t *IntelligentDebugTool) determineRootCauseFromLogs(logFiles []string, sandboxURL string) string {
	var rootCauses []string
	
	// Check each log file for specific patterns
	for _, file := range logFiles {
		var url string
		if config.IsMockMode() {
			path := extractPathFromURL(sandboxURL)
			url = fmt.Sprintf("%s/logs?path=%s&file=%s", t.apiEndpoints.SandboxAPI, path, file)
		} else {
			// In production, use the sandbox URL directly
			url = fmt.Sprintf("%s&action=read&file=%s", sandboxURL, file)
		}
		resp, err := utils.GetHTTP(url)
		if err != nil {
			continue
		}
		
		// Look for specific root cause patterns
		if strings.Contains(resp, "Unable to properly resolve the host") {
			rootCauses = append(rootCauses, "DNS resolution failure - host cannot be resolved")
		}
		if strings.Contains(resp, "Connection refused") {
			rootCauses = append(rootCauses, "Service connection refused - target service may be down")
		}
		if strings.Contains(resp, "OOMKilled") {
			rootCauses = append(rootCauses, "Out of Memory - container exceeded memory limits")
		}
		if strings.Contains(resp, "Permission denied") {
			rootCauses = append(rootCauses, "Permission denied - check RBAC or file permissions")
		}
		if strings.Contains(resp, "No such file or directory") {
			rootCauses = append(rootCauses, "Missing required files or directories")
		}
		if strings.Contains(resp, "timeout") || strings.Contains(resp, "Timeout") {
			rootCauses = append(rootCauses, "Operation timeout - check network connectivity or increase timeout")
		}
	}
	
	if len(rootCauses) == 0 {
		return ""
	}
	
	// Build root cause summary
	var result strings.Builder
	result.WriteString("Based on log analysis, the following root causes were identified:\n")
	for i, cause := range rootCauses {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, cause))
	}
	
	return result.String()
}

// containsContainerError checks if specific container error exists
func containsContainerError(logContent, errorType string) bool {
	return strings.Contains(logContent, errorType)
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