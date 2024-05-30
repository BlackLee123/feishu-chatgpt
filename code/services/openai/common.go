package openai

import (
	"start-feishubot/initialization"

	openai "github.com/sashabaranov/go-openai"
)

type PlatForm string

const (
	OpenAI PlatForm = "openai"
	Azure  PlatForm = "azure"
)

type AzureConfig struct {
	ResourceName   string
	DeploymentName string
	ApiVersion     string
	ApiToken       string
}

type ChatGPT struct {
	ApiKey      []string
	ApiUrl      string
	HttpProxy   string
	Model       string
	MaxTokens   int
	Platform    PlatForm
	AzureConfig AzureConfig
	Client      *openai.Client
}

func NewChatGPT(config initialization.Config) *ChatGPT {
	var client *openai.Client
	platform := OpenAI
	if config.AzureOn {
		platform = Azure
		config := openai.DefaultAzureConfig("1869c1c15b6a4966ac1b93d51283a374", "https://pandada-ai.openai.azure.com/")
		config.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"gpt-4o": "pandada-gpt4o",
			}
			return azureModelMapping[model]
		}
		client = openai.NewClientWithConfig(config)

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
		AzureConfig: AzureConfig{
			ResourceName:   config.AzureResourceName,
			DeploymentName: config.AzureDeploymentName,
			ApiVersion:     config.AzureApiVersion,
			ApiToken:       config.AzureOpenaiToken,
		},
		Client: client,
	}
}
