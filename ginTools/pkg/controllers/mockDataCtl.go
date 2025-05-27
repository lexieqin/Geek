package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

// GetJobByTenantAndJobID handles /tenant/{tenant}/job/{jobid}?trace=true
func (c *MockDataController) GetJobByTenantAndJobID(ctx *gin.Context) {
	tenant := ctx.Param("tenant")
	jobID := ctx.Param("jobid")
	_ = ctx.Query("trace") // Optional trace parameter (used in production)
	
	if tenant == "" || jobID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing tenant or jobid parameter",
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
	_ = ctx.Query("path")  // e.g., /csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704
	_ = ctx.Query("hostip") // hostIP parameter (used in real implementation)
	logFile := ctx.Query("file")
	
	// Default to containers.log if no specific file requested
	if logFile == "" {
		logFile = "containers.log"
	}
	
	// For mock purposes, we map any sandbox path to our sandboxlink directory
	// This simulates having different sandboxes with their own log files
	sandboxLinkPath := filepath.Join(c.staticFilePath, "sandboxlink")
	
	// Try to read the requested log file from sandboxlink directory
	// The logFile parameter can include subdirectories (e.g., "applog/deploy.log")
	logFilePath := filepath.Join(sandboxLinkPath, logFile)
	
	// Clean the path to prevent directory traversal attacks
	logFilePath = filepath.Clean(logFilePath)
	logData, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		// If file doesn't exist, return appropriate error
		if os.IsNotExist(err) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Log file %s not found in sandbox", logFile),
			})
			return
		}
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

// GetSandboxLogList handles listing files in sandbox directory
func (c *MockDataController) GetSandboxLogList(ctx *gin.Context) {
	// Check if path parameter is provided (for realistic behavior)
	sandboxPath := ctx.Query("path")
	if sandboxPath == "" {
		// If no path provided, return empty list (simulating no sandbox)
		ctx.JSON(http.StatusOK, []string{})
		return
	}
	
	// For mock purposes, list files from our sandboxlink directory
	// This simulates listing actual files in a sandbox
	sandboxLinkPath := filepath.Join(c.staticFilePath, "sandboxlink")
	
	// Recursively find all files in the directory tree
	availableFiles := []string{}
	err := filepath.Walk(sandboxLinkPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		
		// Skip directories and README
		if !info.IsDir() && info.Name() != "README.md" {
			// Get relative path from sandboxlink directory
			relPath, err := filepath.Rel(sandboxLinkPath, path)
			if err == nil {
				// Convert to forward slashes for consistency
				relPath = filepath.ToSlash(relPath)
				availableFiles = append(availableFiles, relPath)
			}
		}
		return nil
	})
	
	if err != nil {
		// If walk fails, return empty list
		ctx.JSON(http.StatusOK, []string{})
		return
	}
	
	// Return the list of all discovered files with their relative paths
	ctx.JSON(http.StatusOK, availableFiles)
}

// GetSandboxLogSmart handles smart log retrieval with critical log extraction
func (c *MockDataController) GetSandboxLogSmart(ctx *gin.Context) {
	// For mock purposes, analyze all logs in sandboxlink directory
	sandboxLinkPath := filepath.Join(c.staticFilePath, "sandboxlink")
	
	// Try to read containers.log as the primary log file
	logData, err := ioutil.ReadFile(filepath.Join(sandboxLinkPath, "containers.log"))
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

// HandleSandboxURL handles the new sandbox URL pattern with action parameters
// This mimics the real sandbox log service behavior
func (c *MockDataController) HandleSandboxURL(ctx *gin.Context) {
	// Extract parameters from query string
	path := ctx.Query("path")       // e.g., /csi-data-dir/7d1f4a89-b6ec-44e4-b047-d34d6d3f9704
	action := ctx.Query("action")   // e.g., list, read, analyze
	file := ctx.Query("file")       // e.g., containers.log (for read action)
	
	// Route based on action
	switch action {
	case "list":
		c.handleSandboxList(ctx, path)
	case "read":
		c.handleSandboxRead(ctx, path, file)
	case "analyze":
		c.handleSandboxAnalyze(ctx, path)
	default:
		// Default to read if no action specified
		c.handleSandboxRead(ctx, path, file)
	}
}

// handleSandboxList lists files in the sandbox
func (c *MockDataController) handleSandboxList(ctx *gin.Context, path string) {
	if path == "" {
		ctx.JSON(http.StatusOK, []string{})
		return
	}
	
	// Use the same logic as GetSandboxLogList
	sandboxLinkPath := filepath.Join(c.staticFilePath, "sandboxlink")
	
	availableFiles := []string{}
	err := filepath.Walk(sandboxLinkPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && info.Name() != "README.md" {
			relPath, err := filepath.Rel(sandboxLinkPath, filePath)
			if err == nil {
				relPath = filepath.ToSlash(relPath)
				availableFiles = append(availableFiles, relPath)
			}
		}
		return nil
	})
	
	if err != nil {
		ctx.JSON(http.StatusOK, []string{})
		return
	}
	
	ctx.JSON(http.StatusOK, availableFiles)
}

// handleSandboxRead reads a specific log file
func (c *MockDataController) handleSandboxRead(ctx *gin.Context, path string, file string) {
	if file == "" {
		file = "containers.log"
	}
	
	sandboxLinkPath := filepath.Join(c.staticFilePath, "sandboxlink")
	logFilePath := filepath.Join(sandboxLinkPath, file)
	logFilePath = filepath.Clean(logFilePath)
	
	logData, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Log file %s not found in sandbox", file),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read log data: %v", err),
		})
		return
	}
	
	ctx.String(http.StatusOK, string(logData))
}

// handleSandboxAnalyze provides smart log analysis
func (c *MockDataController) handleSandboxAnalyze(ctx *gin.Context, path string) {
	// Use the same logic as GetSandboxLogSmart
	sandboxLinkPath := filepath.Join(c.staticFilePath, "sandboxlink")
	
	var allCriticalLogs []map[string]interface{}
	errorCounts := make(map[string]int)
	warningCount := 0
	totalErrors := 0
	
	// Analyze all log files
	err := filepath.Walk(sandboxLinkPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() == "README.md" {
			return nil
		}
		
		logData, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil
		}
		
		relPath, _ := filepath.Rel(sandboxLinkPath, filePath)
		relPath = filepath.ToSlash(relPath)
		
		// Extract critical logs from this file
		criticalLogs := extractCriticalLogsWithFile(string(logData), relPath)
		for _, log := range criticalLogs {
			if level, ok := log["level"].(string); ok {
				if level == "ERROR" {
					totalErrors++
					if category, ok := log["category"].(string); ok {
						errorCounts[category]++
					}
				} else if level == "WARNING" {
					warningCount++
				}
			}
		}
		allCriticalLogs = append(allCriticalLogs, criticalLogs...)
		
		return nil
	})
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze logs",
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"counts": gin.H{
				"errors":   totalErrors,
				"warnings": warningCount,
			},
			"error_categories": errorCounts,
		},
		"critical_logs": allCriticalLogs,
	})
}

// extractCriticalLogsWithFile is similar to extractCriticalLogs but includes file info
func extractCriticalLogsWithFile(logs string, fileName string) []map[string]interface{} {
	var criticalLogs []map[string]interface{}
	lines := strings.Split(logs, "\n")
	
	errorPattern := regexp.MustCompile(`(?i)(error|exception|fail|fatal|panic)`)
	warningPattern := regexp.MustCompile(`(?i)(warning|warn)`)
	
	for i, line := range lines {
		logEntry := map[string]interface{}{
			"file":        fileName,
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
		
		if len(criticalLogs) > 50 { // Limit per file
			break
		}
	}
	
	return criticalLogs
}