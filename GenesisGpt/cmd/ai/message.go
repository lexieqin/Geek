package ai

import (
	"context"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

var MessageStore ChatMessages

func init() {
	MessageStore = make(ChatMessages, 0)
	MessageStore.Clear() // Clean and initialize
}

func NewOpenAiClient() *openai.Client {
	// Get API token from environment variable
	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}
	
	dashscope_url := os.Getenv("OPENAI_BASE_URL")
	if dashscope_url == "" {
		dashscope_url = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}

	config := openai.DefaultConfig(token)
	config.BaseURL = dashscope_url

	return openai.NewClientWithConfig(config)
}

// GetModelName returns the AI model to use from environment or default
func GetModelName() string {
	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "qwen-max" // Default model
	}
	return model
}

// NormalChat handles the chat conversation
func NormalChat(message []openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	c := NewOpenAiClient()
	model := GetModelName()
	
	// Add model-specific system message for Claude to prevent hallucination
	if model == "claude-3-5-sonnet-20241022" || model == "claude-3-7-sonnet-20250219" {
		// Prepend a strong system message for Claude
		systemMsg := openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleSystem,
			Content: "CRITICAL: When using tools, you MUST wait for actual responses. Never generate fake tool outputs or observations. Tool outputs will contain specific timestamps and data that you cannot predict.",
		}
		message = append([]openai.ChatCompletionMessage{systemMsg}, message...)
	}
	
	rsp, err := c.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:    model,
		Messages: message,
	})
	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	return rsp.Choices[0].Message
}

// Define chat model
type ChatMessages []*ChatMessage
type ChatMessage struct {
	Msg openai.ChatCompletionMessage
}

// Define roles
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// Define personality
func (cm *ChatMessages) Clear() {
	*cm = make([]*ChatMessage, 0) // Reinitialize
	cm.AddForSystem("You are a helpful k8s assistant!")
}

// Add role and corresponding prompt
func (cm *ChatMessages) AddFor(msg string, role string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:    role,
			Content: msg,
		},
	})
}

// Add System role prompt
func (cm *ChatMessages) AddForSystem(msg string) {
	cm.AddFor(msg, RoleSystem)
}

// Add User role prompt
func (cm *ChatMessages) AddForUser(msg string) {
	cm.AddFor(msg, RoleUser)
}

// Add Assistant role prompt
func (cm *ChatMessages) AddForAssistant(msg string) {
	cm.AddFor(msg, RoleAssistant)
}

// Assemble prompt
func (cm *ChatMessages) ToMessage() []openai.ChatCompletionMessage {
	ret := make([]openai.ChatCompletionMessage, len(*cm))
	for index, c := range *cm {
		ret[index] = c.Msg
	}
	return ret
}

// Get the last message
func (cm *ChatMessages) GetLast() string {
	if len(*cm) == 0 {
		return "Nothing found"
	}

	return (*cm)[len(*cm)-1].Msg.Content
}
