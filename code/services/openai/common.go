package openai

import (
	"github.com/blacklee123/feishu-openai/initialization"

	openai "github.com/sashabaranov/go-openai"
)

type PlatForm string

const (
	OpenAI PlatForm = "openai"
	Azure  PlatForm = "azure"
)

type ChatGPT struct {
	ApiKey    []string
	ApiUrl    string
	HttpProxy string
	Model     string
	MaxTokens int
	Platform  PlatForm
	Client    *openai.Client
}

func NewChatGPT(config initialization.Config) *ChatGPT {
	var client *openai.Client
	platform := OpenAI
	if config.AzureOn {
		platform = Azure
		azureConfig := openai.DefaultAzureConfig(config.AzureOpenaiToken, config.AzureEndpoint)
		azureConfig.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				config.OpenaiModel: config.AzureDeploymentName,
				"dall-e-3":         "pandada-dall-e-3",
			}
			return azureModelMapping[model]
		}
		azureConfig.APIVersion = "2024-02-01"
		client = openai.NewClientWithConfig(azureConfig)

	} else {
		client = openai.NewClient(config.OpenaiApiKeys[0])
	}

	return &ChatGPT{
		ApiKey:    config.OpenaiApiKeys,
		ApiUrl:    config.OpenaiApiUrl,
		HttpProxy: config.HttpProxy,
		Model:     config.OpenaiModel,
		MaxTokens: config.OpenaiMaxTokens,
		Platform:  platform,
		Client:    client,
	}
}
