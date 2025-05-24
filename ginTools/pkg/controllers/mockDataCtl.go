package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type MockDataController struct {
	staticFilePath string
}

func NewMockDataController() *MockDataController {
	return &MockDataController{
		staticFilePath: "pkg/staticfile",
	}
}

// GetJobByTenantAndUUID handles /tenant/{tenant}/jobs?requuid={jobid}
func (c *MockDataController) GetJobByTenantAndUUID(ctx *gin.Context) {
	tenant := ctx.Param("tenant")
	jobID := ctx.Query("requuid")
	
	if tenant == "" || jobID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing tenant or requuid parameter",
		})
		return
	}
	
	// Read the static job.json file
	jobData, err := ioutil.ReadFile(filepath.Join(c.staticFilePath, "job.json"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read job data: %v", err),
		})
		return
	}
	
	// Return the raw JSON to preserve field order
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, string(jobData))
}

// GetDatadogTrace handles datadog trace requests
func (c *MockDataController) GetDatadogTrace(ctx *gin.Context) {
	// Extract trace ID from URL path or query
	traceID := ctx.Query("trace_id")
	if traceID == "" {
		// Try to extract from path if it's provided differently
		traceID = ctx.Param("trace_id")
	}
	
	// Read the static datadogtrace.json file
	traceData, err := ioutil.ReadFile(filepath.Join(c.staticFilePath, "datadogtrace.json"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read trace data: %v", err),
		})
		return
	}
	
	// Return the raw JSON to preserve field order
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, string(traceData))
}

// GetSandboxLog handles sandbox log requests
func (c *MockDataController) GetSandboxLog(ctx *gin.Context) {
	// Extract parameters from the request
	_ = ctx.Query("path")  // path parameter (used in real implementation)
	_ = ctx.Query("hostip") // hostIP parameter (used in real implementation)
	logFile := ctx.Query("file")
	
	// Default to containers.log if no specific file requested
	if logFile == "" {
		logFile = "containers.log"
	}
	
	// Read the containers.log file
	logData, err := ioutil.ReadFile(filepath.Join(c.staticFilePath, logFile))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read log data: %v", err),
		})
		return
	}
	
	// If search parameter is provided, filter logs
	search := ctx.Query("search")
	if search != "" {
		filteredLogs := filterLogs(string(logData), search)
		ctx.String(http.StatusOK, filteredLogs)
		return
	}
	
	// Return raw log data
	ctx.String(http.StatusOK, string(logData))
}

// GetSandboxLogSmart handles smart log retrieval with critical log extraction
func (c *MockDataController) GetSandboxLogSmart(ctx *gin.Context) {
	// Read the containers.log file
	logData, err := ioutil.ReadFile(filepath.Join(c.staticFilePath, "containers.log"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read log data: %v", err),
		})
		return
	}
	
	logs := string(logData)
	
	// Extract critical logs based on patterns
	criticalLogs := extractCriticalLogs(logs)
	
	// Return structured response
	ctx.JSON(http.StatusOK, gin.H{
		"total_lines": strings.Count(logs, "\n"),
		"critical_logs": criticalLogs,
		"summary": summarizeLogs(criticalLogs),
	})
}

// Helper function to filter logs based on search criteria
func filterLogs(logs string, search string) string {
	var filteredLines []string
	lines := strings.Split(logs, "\n")
	
	searchLower := strings.ToLower(search)
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), searchLower) {
			filteredLines = append(filteredLines, line)
		}
	}
	
	return strings.Join(filteredLines, "\n")
}

// Helper function to extract critical logs
func extractCriticalLogs(logs string) []map[string]interface{} {
	var criticalLogs []map[string]interface{}
	lines := strings.Split(logs, "\n")
	
	// Patterns for critical logs
	errorPattern := regexp.MustCompile(`(?i)(error|exception|fail|fatal|panic)`)
	warningPattern := regexp.MustCompile(`(?i)(warning|warn)`)
	
	for i, line := range lines {
		logEntry := map[string]interface{}{
			"line_number": i + 1,
			"content":     line,
		}
		
		if errorPattern.MatchString(line) {
			logEntry["level"] = "ERROR"
			logEntry["category"] = categorizeError(line)
			criticalLogs = append(criticalLogs, logEntry)
		} else if warningPattern.MatchString(line) {
			logEntry["level"] = "WARNING"
			criticalLogs = append(criticalLogs, logEntry)
		}
		
		// Limit to most recent/relevant entries
		if len(criticalLogs) > 100 {
			break
		}
	}
	
	return criticalLogs
}

// Helper function to categorize errors
func categorizeError(logLine string) string {
	if strings.Contains(strings.ToLower(logLine), "connection") {
		return "CONNECTION_ERROR"
	}
	if strings.Contains(strings.ToLower(logLine), "timeout") {
		return "TIMEOUT_ERROR"
	}
	if strings.Contains(strings.ToLower(logLine), "permission") || strings.Contains(strings.ToLower(logLine), "denied") {
		return "PERMISSION_ERROR"
	}
	if strings.Contains(strings.ToLower(logLine), "memory") || strings.Contains(strings.ToLower(logLine), "oom") {
		return "MEMORY_ERROR"
	}
	if strings.Contains(strings.ToLower(logLine), "database") || strings.Contains(strings.ToLower(logLine), "sql") {
		return "DATABASE_ERROR"
	}
	return "GENERAL_ERROR"
}

// Helper function to summarize logs
func summarizeLogs(criticalLogs []map[string]interface{}) map[string]interface{} {
	summary := map[string]int{
		"total_critical": len(criticalLogs),
		"errors":         0,
		"warnings":       0,
	}
	
	errorCategories := make(map[string]int)
	
	for _, log := range criticalLogs {
		if level, ok := log["level"].(string); ok {
			if level == "ERROR" {
				summary["errors"]++
				if category, ok := log["category"].(string); ok {
					errorCategories[category]++
				}
			} else if level == "WARNING" {
				summary["warnings"]++
			}
		}
	}
	
	return map[string]interface{}{
		"counts":           summary,
		"error_categories": errorCategories,
	}
}