package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/blacklee123/feishu-openai/services"
	myopenai "github.com/blacklee123/feishu-openai/services/openai"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	openai "github.com/sashabaranov/go-openai"

	"go.uber.org/zap"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type Config struct {
	FeishuAppId             string `mapstructure:"FEISHU_APP_ID"`
	FeishuAppSecret         string `mapstructure:"FEISHU_APP_SECRET"`
	FeishuEncryptKey        string `mapstructure:"FEISHU_ENCRYPT_KEY"`
	FeishuVerificationToken string `mapstructure:"FEISHU_VERIFICATION_TOKEN"`

	OpenaiApiKey    string `mapstructure:"OPENAI_KEY"`
	OpenaiModel     string `mapstructure:"OPENAI_MODEL"`
	OpenaiMaxTokens int    `mapstructure:"OPENAI_MAX_TOKENS"`
	OpenaiApiUrl    string `mapstructure:"OPENAI_API_URL"`
	HttpProxy       string `mapstructure:"HTTP_PROXY"`

	AzureOn bool `mapstructure:"AZURE_ON"`

	AzureEndpoint            string `mapstructure:"AZURE_ENDPOINT"`
	AzureApiVersion          string `mapstructure:"AZURE_APIVERSION"`
	AzureOpenaiToken         string `mapstructure:"AZURE_OPENAI_TOKEN"`
	AzureDeploymentName      string `mapstructure:"AZURE_DEPLOYMENT_NAME"`
	AzureDalleDeploymentName string `mapstructure:"AZURE_DALLE_DEPLOYMENT_NAME"`

	AzureWhisperEndpoint       string `mapstructure:"AZURE_WHISPER_ENDPOINT"`
	AzureWhisperApiVersion     string `mapstructure:"AZURE_WHISPER_APIVERSION"`
	AzureWhisperOpenaiToken    string `mapstructure:"AZURE_WHISPER_OPENAI_TOKEN"`
	AzureWhisperDeploymentName string `mapstructure:"AZURE_WHISPER_DEPLOYMENT_NAME"`

	AzureTtsEndpoint       string `mapstructure:"AZURE_TTS_ENDPOINT"`
	AzureTtsApiVersion     string `mapstructure:"AZURE_TTS_APIVERSION"`
	AzureTtsOpenaiToken    string `mapstructure:"AZURE_TTS_OPENAI_TOKEN"`
	AzureTtsDeploymentName string `mapstructure:"AZURE_TTS_DEPLOYMENT_NAME"`
}

const (
	OpenAI string = "openai"
	Azure  string = "azure"
)

// 责任链
func chain(data *ActionInfo, actions ...Action) bool {
	for _, v := range actions {
		if !v.Execute(data) {
			return false
		}
	}
	return true
}

type MessageHandlerInterface interface {
	MsgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error
}

type HandlerType string

const (
	GroupHandler = "group"
	UserHandler  = "personal"
)

func judgeChatType(event *larkim.P2MessageReceiveV1) HandlerType {
	chatType := event.Event.Message.ChatType
	if *chatType == "group" {
		return GroupHandler
	}
	if *chatType == "p2p" {
		return UserHandler
	}
	return "otherChat"
}

type MessageHandler struct {
	sessionCache services.SessionServiceCacheInterface
	gpt          *myopenai.ChatGPT
	config       Config
	logger       *zap.Logger
	larkClient   *lark.Client
}

func judgeMsgType(event *larkim.P2MessageReceiveV1) (string, error) {
	msgType := event.Event.Message.MessageType

	switch *msgType {
	case "text", "image", "audio", "post":
		return *msgType, nil
	default:
		return "", fmt.Errorf("unknown message type: %v", *msgType)
	}

}

func (a MessageHandler) replyMsg(ctx context.Context, msg string, msgId *string) error {
	msg, i := processMessage(msg)
	if i != nil {
		return i
	}
	client := a.larkClient
	content := larkim.NewTextMsgBuilder().
		Text(msg).
		Build()

	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			Uuid(uuid.New().String()).
			Content(content).
			Build()).
		Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return errors.New(resp.Msg)
	}
	return nil
}

func (m MessageHandler) MsgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	go func() {
		m.logger.Info("[receive]", zap.String("messageid", *event.Event.Message.MessageId), zap.String("MessageType", *event.Event.Message.MessageType), zap.String("message", *event.Event.Message.Content))
		// alert(ctx, fmt.Sprintf("收到消息: messageId %v", *event.Event.Message.MessageId))
		handlerType := judgeChatType(event)
		if handlerType == "otherChat" {
			m.replyMsg(ctx, "unknown chat type", event.Event.Message.MessageId)
			m.logger.Error("unknown chat type")
			return
		}
		//fmt.Println(larkcore.Prettify(event.Event.Message))

		msgType, err := judgeMsgType(event)
		if err != nil {
			m.replyMsg(ctx, "🥹不支持的消息类型, 当前仅支持文本消息、图片消息、语音消息", event.Event.Message.MessageId)
			m.logger.Error("error getting message type", zap.Error(err))
			return
		}

		content := event.Event.Message.Content
		msgId := event.Event.Message.MessageId
		rootId := event.Event.Message.RootId
		chatId := event.Event.Message.ChatId

		sessionId := rootId
		if sessionId == nil || *sessionId == "" {
			sessionId = msgId
		}
		qParsed := strings.Trim(parseContent(*content, msgType), " ")
		m.logger.Info("[receive]", zap.String("messageid", *event.Event.Message.MessageId), zap.String("MessageType", *event.Event.Message.MessageType), zap.String("qParsed", qParsed))
		imageKeys := []string{}
		if msgType == "image" {
			imageKeys = []string{parseImageKey(*content)}
			qParsed = "描述一下图片中的内容"
		} else if msgType == "post" {
			imageKeys = parsePostImageKeys(*content)
		}
		msgInfo := MsgInfo{
			handlerType: handlerType,
			msgType:     msgType,
			msgId:       msgId,
			chatId:      chatId,
			userId:      event.Event.Sender.SenderId.OpenId,
			qParsed:     qParsed,
			fileKey:     parseFileKey(*content),
			imageKeys:   imageKeys,
			sessionId:   sessionId,
		}
		data := &ActionInfo{
			ctx:        &ctx,
			handler:    &m,
			info:       &msgInfo,
			logger:     m.logger,
			config:     m.config,
			larkClient: m.larkClient,
		}
		actions := []Action{
			&HelpAction{},    //帮助处理
			&PreAction{},     //预处理
			&PicAction{},     //图片处理
			&AudioAction{},   //语音处理
			&MessageAction{}, //消息处理
		}
		chain(data, actions...)
	}()
	return nil
}

var _ MessageHandlerInterface = (*MessageHandler)(nil)

func NewMessageHandler(gpt *myopenai.ChatGPT, config Config, logger *zap.Logger, larkClient *lark.Client) MessageHandlerInterface {
	return &MessageHandler{
		sessionCache: services.GetSessionCache(),
		gpt:          gpt,
		config:       config,
		logger:       logger,
		larkClient:   larkClient,
	}
}

func NewChatGPT(config Config, logger *zap.Logger) *myopenai.ChatGPT {
	var client *openai.Client
	var whisperClient *openai.Client
	var ttsClient *openai.Client
	if config.AzureOn {
		azureConfig := openai.DefaultAzureConfig(config.AzureOpenaiToken, config.AzureEndpoint)
		azureConfig.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				config.OpenaiModel: config.AzureDeploymentName,
				"dall-e-3":         config.AzureDalleDeploymentName,
			}
			return azureModelMapping[model]
		}
		azureConfig.APIVersion = config.AzureApiVersion
		client = openai.NewClientWithConfig(azureConfig)

		azureWhisperConfig := openai.DefaultAzureConfig(config.AzureWhisperOpenaiToken, config.AzureWhisperEndpoint)
		azureWhisperConfig.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"whisper-1": config.AzureWhisperDeploymentName,
			}
			return azureModelMapping[model]
		}
		azureWhisperConfig.APIVersion = config.AzureWhisperApiVersion
		whisperClient = openai.NewClientWithConfig(azureWhisperConfig)

		azureTtsConfig := openai.DefaultAzureConfig(config.AzureTtsOpenaiToken, config.AzureTtsEndpoint)
		azureTtsConfig.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"tts-1": config.AzureTtsDeploymentName,
			}
			return azureModelMapping[model]
		}
		azureTtsConfig.APIVersion = config.AzureTtsApiVersion
		ttsClient = openai.NewClientWithConfig(azureTtsConfig)
	} else {
		client = openai.NewClient(config.OpenaiApiKey)
	}

	return &myopenai.ChatGPT{
		ApiKey:        config.OpenaiApiKey,
		ApiUrl:        config.OpenaiApiUrl,
		HttpProxy:     config.HttpProxy,
		Model:         config.OpenaiModel,
		MaxTokens:     config.OpenaiMaxTokens,
		AzureOn:       config.AzureOn,
		Client:        client,
		WhisperClient: whisperClient,
		TtsClient:     ttsClient,
		Logger:        logger,
	}
}
