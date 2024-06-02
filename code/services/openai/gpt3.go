package openai

import (
	"context"
	"errors"
	"io"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
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

func (gpt *ChatGPT) Completions(ctx context.Context, msg []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	resp, err := gpt.Client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:            openai.GPT4o,
		Messages:         msg,
		MaxTokens:        gpt.MaxTokens,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	},
	)

	if err != nil {
		gpt.logger.Error("ChatCompletion error", zap.Error(err))
		return openai.ChatCompletionMessage{}, err
	}

	return resp.Choices[0].Message, nil
}

func (gpt *ChatGPT) StreamChat(ctx context.Context, msgs []openai.ChatCompletionMessage, responseStream chan<- string) error {
	defer close(responseStream)
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4o,
		Messages:  msgs,
		MaxTokens: 2000,
		Stream:    true,
	}
	stream, err := gpt.Client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		gpt.logger.Error("ChatCompletionStream error", zap.Error(err))
		return err
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			gpt.logger.Error("Stream error", zap.Error(err))
			return err
		}
		if len(response.Choices) > 0 {
			responseStream <- response.Choices[0].Delta.Content
			gpt.logger.Debug("response", zap.String("content", response.Choices[0].Delta.Content))
		}

	}
}
