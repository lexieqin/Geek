package tools

import "fmt"

type HumanToolParam struct {
	Prompt string `json:"prompt"`
}

// HumanTool represents a tool for human interaction.
type HumanTool struct {
	Name        string
	Description string
	ArgsSchema  string
}

// NewHumanTool creates a new HumanTool instance.
func NewHumanTool() *HumanTool {
	return &HumanTool{
		Name:        "HumanTool",
		Description: "When you need to perform irreversible dangerous operations, such as deletion actions, use this tool to request human confirmation first",
		ArgsSchema:  `{"type":"object","properties":{"prompt":{"type":"string", "description": "Content for which you need human assistance", "example": "Please confirm if you want to delete the foo-app pod in the default namespace"}}}`,
	}
}

// Run executes the command and returns the output.
func (d *HumanTool) Run(prompt string) string {
	fmt.Print(prompt, " ")
	var input string
	fmt.Scanln(&input)
	return input
}
