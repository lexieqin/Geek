package tools

import (
	"fmt"
	"os"
)

type HumanToolParam struct {
	Prompt string `json:"prompt"`
}

// HumanTool represents a tool for human interaction.
type HumanTool struct {
	Name        string
	Description string
	ArgsSchema  string
	ServerMode  bool
}

// NewHumanTool creates a new HumanTool instance.
func NewHumanTool() *HumanTool {
	// Check if running in server mode
	serverMode := os.Getenv("K8SGPT_SERVER_MODE") == "true"
	
	return &HumanTool{
		Name:        "HumanTool",
		Description: "When you need to perform irreversible dangerous operations, such as deletion actions, use this tool to request human confirmation first",
		ArgsSchema:  `{"type":"object","properties":{"prompt":{"type":"string", "description": "Content for which you need human assistance", "example": "Please confirm if you want to delete the foo-app pod in the default namespace"}}}`,
		ServerMode:  serverMode,
	}
}

// Run executes the command and returns the output.
func (d *HumanTool) Run(prompt string) string {
	if d.ServerMode {
		// In server mode, return a special response that the UI can handle
		return fmt.Sprintf("[HUMAN_CONFIRMATION_REQUIRED]: %s", prompt)
	}
	
	// In CLI mode, use the original behavior
	fmt.Print(prompt, " ")
	var input string
	fmt.Scanln(&input)
	return input
}
