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
func NormalChat(message []openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	c := NewOpenAiClient()
	rsp, err := c.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:    "qwen-max",
		Messages: message,
	})
	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	return rsp.Choices[0].Message
}

// chat对话
func Chat(message []openai.ChatCompletionMessage, tools []openai.Tool) openai.ChatCompletionMessage {
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
type ChatMessages []*ChatMessage
type ChatMessage struct {
	Msg openai.ChatCompletionMessage
}

// 枚举出角色
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// 定义人设
func (cm *ChatMessages) Clear() {
	*cm = make([]*ChatMessage, 0) //重新初始化
	cm.AddForSystem("你是一个数学家")
}

// 添加角色和对应的prompt
func (cm *ChatMessages) AddFor(msg string, role string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:    role,
			Content: msg,
		},
	})
}

// 添加角色和对应的prompt
func (cm *ChatMessages) AddForToolCall(rsp openai.ChatCompletionMessage, role string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:         role,
			Content:      rsp.Content,
			FunctionCall: rsp.FunctionCall,
			ToolCalls:    rsp.ToolCalls,
		},
	})
}

// 添加Assistant角色的prompt
func (cm *ChatMessages) AddForAssistant(rsp openai.ChatCompletionMessage) {
	cm.AddForToolCall(rsp, RoleAssistant)

}

// 添加System角色的prompt
func (cm *ChatMessages) AddForSystem(msg string) {
	cm.AddFor(msg, RoleSystem)
}

// 添加User角色的prompt
func (cm *ChatMessages) AddForUser(msg string) {
	cm.AddFor(msg, RoleUser)
}

// 添加Tool角色的prompt
func (cm *ChatMessages) AddForTool(msg string, name string, toolCallID string) {
	*cm = append(*cm, &ChatMessage{
		Msg: openai.ChatCompletionMessage{
			Role:       RoleTool,
			Content:    msg,
			Name:       name,
			ToolCallID: toolCallID,
		},
	})
}

// 组装prompt
func (cm *ChatMessages) ToMessage() []openai.ChatCompletionMessage {
	ret := make([]openai.ChatCompletionMessage, len(*cm))
	for index, c := range *cm {
		ret[index] = c.Msg
	}
	return ret
}

// 得到返回的消息
func (cm *ChatMessages) GetLast() string {
	if len(*cm) == 0 {
		return "什么都没找到"
	}

	return (*cm)[len(*cm)-1].Msg.Content
}
