package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/blacklee123/feishu-openai/initialization"
	"github.com/blacklee123/feishu-openai/services"
	"github.com/blacklee123/feishu-openai/services/openai"

	"go.uber.org/zap"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
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
	msgCache     services.MsgCacheInterface
	gpt          *openai.ChatGPT
	config       initialization.Config
	logger       *zap.Logger
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

func (m MessageHandler) MsgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	go func() {
		m.logger.Info("[receive]", zap.String("messageid", *event.Event.Message.MessageId), zap.String("MessageType", *event.Event.Message.MessageType), zap.String("message", *event.Event.Message.Content))
		// alert(ctx, fmt.Sprintf("收到消息: messageId %v", *event.Event.Message.MessageId))
		handlerType := judgeChatType(event)
		if handlerType == "otherChat" {
			replyMsg(ctx, "unknown chat type", event.Event.Message.MessageId)
			m.logger.Error("unknown chat type")
			return
		}
		//fmt.Println(larkcore.Prettify(event.Event.Message))

		msgType, err := judgeMsgType(event)
		if err != nil {
			replyMsg(ctx, "🥹不支持的消息类型, 当前仅支持文本消息、图片消息、语音消息", event.Event.Message.MessageId)
			m.logger.Error("error getting message type", zap.Error(err))
			return
		}

		content := event.Event.Message.Content
		msgId := event.Event.Message.MessageId
		rootId := event.Event.Message.RootId
		chatId := event.Event.Message.ChatId
		mention := event.Event.Message.Mentions

		sessionId := rootId
		if sessionId == nil || *sessionId == "" {
			sessionId = msgId
		}
		qParsed := strings.Trim(parseContent(*content, msgType), " ")
		m.logger.Info("[receive]", zap.String("messageid", *event.Event.Message.MessageId), zap.String("MessageType", *event.Event.Message.MessageType), zap.String("qParsed", qParsed))
		msgInfo := MsgInfo{
			handlerType: handlerType,
			msgType:     msgType,
			msgId:       msgId,
			chatId:      chatId,
			qParsed:     qParsed,
			fileKey:     parseFileKey(*content),
			imageKey:    parseImageKey(*content),
			imageKeys:   parsePostImageKeys(*content),
			sessionId:   sessionId,
			mention:     mention,
		}
		data := &ActionInfo{
			ctx:     &ctx,
			handler: &m,
			info:    &msgInfo,
			logger:  m.logger,
			config:  m.config,
		}
		actions := []Action{
			&HelpAction{},    //帮助处理
			&PreAction{},     //预处理
			&AudioAction{},   //语音处理
			&PicAction{},     //图片处理
			&MessageAction{}, //消息处理
		}
		chain(data, actions...)
	}()
	return nil
}

var _ MessageHandlerInterface = (*MessageHandler)(nil)

func NewMessageHandler(gpt *openai.ChatGPT, config initialization.Config, logger *zap.Logger) MessageHandlerInterface {
	return &MessageHandler{
		sessionCache: services.GetSessionCache(),
		msgCache:     services.GetMsgCache(),
		gpt:          gpt,
		config:       config,
		logger:       logger,
	}
}
