package tools

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/xingyunyang01/K8sGpt/cmd/ai"
	"github.com/xingyunyang01/K8sGpt/cmd/promptTpl"
	"github.com/xingyunyang01/K8sGpt/cmd/utils"
)

type CreateToolParam struct {
	Prompt   string `json:"prompt"`
	Resource string `json:"resource"`
}

// Define struct to parse JSON response
type response struct {
	Data string `json:"data"`
}

// CreateTool represents a tool for creating k8s resources.
type CreateTool struct {
	Name        string
	Description string
	ArgsSchema  string
}

// NewCreateTool creates a new CreateTool instance.
func NewCreateTool() *CreateTool {
	return &CreateTool{
		Name:        "CreateTool",
		Description: "Used to create specified Kubernetes resources in a given namespace, such as creating pods, services, etc.",
		ArgsSchema:  `{"type":"object","properties":{"prompt":{"type":"string", "description": "Place the user's resource creation prompt here exactly as provided, without any modifications"},"resource":{"type":"string", "description": "Specified k8s resource type, e.g. pod, service, etc."}}}`,
	}
}

// Run executes the command and returns the output.
func (c *CreateTool) Run(prompt string, resource string) string {
	// Let the large model generate yaml
	messages := make([]openai.ChatCompletionMessage, 2)

	messages[0] = openai.ChatCompletionMessage{Role: "system", Content: promptTpl.SystemPrompt}
	messages[1] = openai.ChatCompletionMessage{Role: "user", Content: prompt}

	rsp := ai.NormalChat(messages)
	fmt.Println("-----------------------")
	fmt.Println(rsp.Content)

	// Create JSON object {"yaml":"xxx"}
	body := map[string]string{"yaml": rsp.Content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err.Error()
	}

	url := "http://localhost:8080/" + resource
	s, err := utils.PostHTTP(url, jsonBody)
	if err != nil {
		return err.Error()
	}

	var response response
	// Parse JSON response
	err = json.Unmarshal([]byte(s), &response)
	if err != nil {
		return err.Error()
	}

	return response.Data
}
