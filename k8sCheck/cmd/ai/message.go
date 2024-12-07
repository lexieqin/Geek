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
	MessageStore.Clear() //清理和初始化

}

func NewOpenAiClient() *openai.Client {
	token := os.Getenv("DashScope")
	dashscope_url := "https://dashscope.aliyuncs.com/compatible-mode/v1"

	config := openai.DefaultConfig(token)
	config.BaseURL = dashscope_url

	return openai.NewClientWithConfig(config)
}

// chat对话
func NormalChat(message []openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	c := NewOpenAiClient()
	rsp, err := c.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:    "qwen-max-0403",
		Messages: message,
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
	cm.AddForSystem("You are a helpful k8s assistant!")
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

// 添加System角色的prompt
func (cm *ChatMessages) AddForSystem(msg string) {
	cm.AddFor(msg, RoleSystem)
}

// 添加User角色的prompt
func (cm *ChatMessages) AddForUser(msg string) {
	cm.AddFor(msg, RoleUser)
}

// 添加Assistant角色的prompt
func (cm *ChatMessages) AddForAssistant(msg string) {
	cm.AddFor(msg, RoleAssistant)
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
