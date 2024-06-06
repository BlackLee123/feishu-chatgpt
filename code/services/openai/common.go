package openai

import (
	"go.uber.org/zap"

	openai "github.com/sashabaranov/go-openai"
)

type ChatGPT struct {
	ApiKey        string
	ApiUrl        string
	HttpProxy     string
	Model         string
	MaxTokens     int
	AzureOn       bool
	Client        *openai.Client
	WhisperClient *openai.Client
	TtsClient     *openai.Client
	Logger        *zap.Logger
}
