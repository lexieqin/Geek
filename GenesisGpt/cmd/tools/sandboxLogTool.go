package tools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/lexieqin/Geek/K8sGpt/cmd/utils"
)

type SandboxLogTool struct{}

func NewSandboxLogTool() *SandboxLogTool {
	return &SandboxLogTool{}
}

func (t *SandboxLogTool) Name() string {
	return "SandboxLogTool"
}

func (t *SandboxLogTool) Description() string {
	return "Read and analyze sandbox log files (std.out, std.err, decout, decerr) from failed jobs. Can search for errors, read specific lines, or analyze the entire log file."
}

func (t *SandboxLogTool) ArgsSchema() string {
	return `{
		"type": "object",
		"properties": {
			"sandboxPath": {
				"type": "string",
				"description": "The sandbox directory path containing log files"
			},
			"action": {
				"type": "string",
				"enum": ["read", "analyze", "search"],
				"description": "Action to perform: read (show lines), analyze (find errors), search (grep for pattern)",
				"default": "analyze"
			},
			"logFile": {
				"type": "string",
				"enum": ["std.out", "std.err", "decout", "decerr"],
				"description": "The log file to read. Required for 'read' and 'search' actions",
				"default": "std.out"
			},
			"startLine": {
				"type": "integer",
				"description": "Starting line number for 'read' action (0-based)",
				"default": 0
			},
			"numLines": {
				"type": "integer",
				"description": "Number of lines to read for 'read' action",
				"default": 100
			},
			"searchPattern": {
				"type": "string",
				"description": "Pattern to search for in 'search' action"
			}
		},
		"required": ["sandboxPath"]
	}`
}

func (t *SandboxLogTool) Run(input string) (string, error) {
	var args struct {
		SandboxPath   string `json:"sandboxPath"`
		Action        string `json:"action"`
		LogFile       string `json:"logFile"`
		StartLine     int    `json:"startLine"`
		NumLines      int    `json:"numLines"`
		SearchPattern string `json:"searchPattern"`
	}

	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid input: %v", err)
	}

	// Set defaults
	if args.Action == "" {
		args.Action = "analyze"
	}
	if args.LogFile == "" {
		args.LogFile = "std.out"
	}
	if args.NumLines == 0 {
		args.NumLines = 100
	}

	switch args.Action {
	case "read":
		return t.readLogFile(args.SandboxPath, args.LogFile, args.StartLine, args.NumLines)
	case "analyze":
		return t.analyzeAllLogs(args.SandboxPath)
	case "search":
		if args.SearchPattern == "" {
			return "", fmt.Errorf("searchPattern is required for search action")
		}
		return t.searchInLog(args.SandboxPath, args.LogFile, args.SearchPattern)
	default:
		return "", fmt.Errorf("invalid action: %s", args.Action)
	}
}

func (t *SandboxLogTool) readLogFile(sandboxPath, logFile string, startLine, numLines int) (string, error) {
	encodedPath := url.QueryEscape(sandboxPath)
	url := fmt.Sprintf("http://localhost:8080/sandbox/read?path=%s&file=%s&start=%d&lines=%d",
		encodedPath, logFile, startLine, numLines)

	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %v", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return resp, nil // Return raw response if parsing fails
	}

	if content, ok := result["content"].(string); ok {
		return fmt.Sprintf("=== Content of %s (lines %d-%d) ===\n%s", 
			logFile, startLine, startLine+numLines, content), nil
	}

	return resp, nil
}

func (t *SandboxLogTool) analyzeAllLogs(sandboxPath string) (string, error) {
	var result strings.Builder
	result.WriteString("=== Analyzing Sandbox Logs ===\n\n")

	// Analyze each log file
	logFiles := []string{"std.out", "std.err", "decout", "decerr"}
	errorSummary := make(map[string][]string)
	
	for _, logFile := range logFiles {
		// Read first 500 lines to look for errors
		encodedPath := url.QueryEscape(sandboxPath)
		url := fmt.Sprintf("http://localhost:8080/sandbox/read?path=%s&file=%s&start=0&lines=500",
			encodedPath, logFile)
		
		resp, err := utils.GetHTTP(url)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		var logResp map[string]interface{}
		if err := json.Unmarshal([]byte(resp), &logResp); err != nil {
			continue
		}

		if content, ok := logResp["content"].(string); ok {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				lowerLine := strings.ToLower(line)
				if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "exception") ||
				   strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "fatal") {
					if errorSummary[logFile] == nil {
						errorSummary[logFile] = []string{}
					}
					errorSummary[logFile] = append(errorSummary[logFile], strings.TrimSpace(line))
				}
			}
		}
	}

	// Format error summary
	if len(errorSummary) > 0 {
		result.WriteString("Found errors in the following files:\n\n")
		for file, errors := range errorSummary {
			result.WriteString(fmt.Sprintf("File: %s\n", file))
			result.WriteString("Errors found:\n")
			// Show first 5 errors per file
			maxErrors := 5
			if len(errors) < maxErrors {
				maxErrors = len(errors)
			}
			for i := 0; i < maxErrors; i++ {
				result.WriteString(fmt.Sprintf("  - %s\n", errors[i]))
			}
			if len(errors) > 5 {
				result.WriteString(fmt.Sprintf("  ... and %d more errors\n", len(errors)-5))
			}
			result.WriteString("\n")
		}
	} else {
		result.WriteString("No obvious errors found in log files.\n")
		result.WriteString("You may want to read specific files for more details.\n")
	}

	return result.String(), nil
}

func (t *SandboxLogTool) searchInLog(sandboxPath, logFile, pattern string) (string, error) {
	// Read the entire file (up to 10000 lines)
	encodedPath := url.QueryEscape(sandboxPath)
	url := fmt.Sprintf("http://localhost:8080/sandbox/read?path=%s&file=%s&start=0&lines=10000",
		encodedPath, logFile)

	resp, err := utils.GetHTTP(url)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %v", err)
	}

	var logResp map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &logResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("=== Searching for '%s' in %s ===\n", pattern, logFile))

	if content, ok := logResp["content"].(string); ok {
		lines := strings.Split(content, "\n")
		matches := 0
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
				result.WriteString(fmt.Sprintf("%s\n", strings.TrimSpace(line)))
				matches++
				if matches >= 20 { // Limit to 20 matches
					result.WriteString("\n... (showing first 20 matches)\n")
					break
				}
			}
		}
		if matches == 0 {
			result.WriteString(fmt.Sprintf("No matches found for '%s'\n", pattern))
		} else {
			result.WriteString(fmt.Sprintf("\nTotal matches shown: %d\n", matches))
		}
	}

	return result.String(), nil
}