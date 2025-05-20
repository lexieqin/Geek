package ai

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	"log"
)

var MessageStore ChatMessages

func init() {
	MessageStore = make(ChatMessages, 0)
	MessageStore.Clear() //清理和初始化

}

func NewOpenAiClient() *openai.Client {
	token := "sk-07d4040e83824cea8df0da757f10844f"
	dashscope_url := "https://dashscope.aliyuncs.com/compatible-mode/v1"
	config := openai.DefaultConfig(token)
	config.BaseURL = dashscope_url

	return openai.NewClientWithConfig(config)
}

// chat对话
func Chat(message []openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	c := NewOpenAiClient()
	rsp, err := c.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:    "qwen-plus",
		Messages: message,
	})
	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	return rsp.Choices[0].Message
}

// 带tools的chat对话
func ToolsChat(message []openai.ChatCompletionMessage, tools []openai.Tool) openai.ChatCompletionMessage {
	c := NewOpenAiClient()
	rsp, err := c.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:      "qwen-plus",
		Messages:   message,
		Tools:      tools,
		ToolChoice: "auto",
	})
	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	return rsp.Choices[0].Message
}

// 定义chat模型
type ChatMessages []openai.ChatCompletionMessage

// 枚举出角色
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

func (cm *ChatMessages) Clear() {
	*cm = make([]openai.ChatCompletionMessage, 0) //重新初始化
}

// 添加角色和对应的prompt
func (cm *ChatMessages) AddFor(role string, msg string, toolCalls []openai.ToolCall) {
	*cm = append(*cm, openai.ChatCompletionMessage{
		Role:      role,
		Content:   msg,
		ToolCalls: toolCalls,
	})
}

// 添加Tool角色的prompt
func (cm *ChatMessages) AddForTool(msg string, name string, toolCallID string) {
	*cm = append(*cm, openai.ChatCompletionMessage{
		Role:       RoleTool,
		Content:    msg,
		Name:       name,
		ToolCallID: toolCallID,
	})
}

func (cm *ChatMessages) ToMessage() []openai.ChatCompletionMessage {
	ret := make([]openai.ChatCompletionMessage, len(*cm))
	for index, c := range *cm {
		ret[index] = c
	}
	return ret
}
