package openai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type AIMode float64

const (
	Fresh      AIMode = 0.1
	Warmth     AIMode = 0.7
	Balance    AIMode = 1.2
	Creativity AIMode = 1.7
)

var AIModeMap = map[string]AIMode{
	"严谨": Fresh,
	"简洁": Warmth,
	"标准": Balance,
	"发散": Creativity,
}

var AIModeStrs = []string{
	"严谨",
	"简洁",
	"标准",
	"发散",
}

type Messages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatGPTChoiceItem    `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChatGPTChoiceItem struct {
	Message      Messages `json:"message"`
	Index        int      `json:"index"`
	FinishReason string   `json:"finish_reason"`
}

func (gpt *ChatGPT) Completions(msg []openai.ChatCompletionMessage, aiMode AIMode) (openai.ChatCompletionMessage, error) {
	resp, err := gpt.Client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:            openai.GPT4o,
		Messages:         msg,
		MaxTokens:        gpt.MaxTokens,
		Temperature:      float32(aiMode),
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return openai.ChatCompletionMessage{}, err
	}

	return resp.Choices[0].Message, nil
}
