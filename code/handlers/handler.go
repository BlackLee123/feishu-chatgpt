package handlers

import (
	"context"
	"fmt"
	"start-feishubot/initialization"
	"start-feishubot/services"
	"start-feishubot/services/openai"
	"strings"

	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
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

type MessageHandler struct {
	sessionCache services.SessionServiceCacheInterface
	msgCache     services.MsgCacheInterface
	gpt          *openai.ChatGPT
	config       initialization.Config
	logger       *zap.Logger
}

func (m MessageHandler) cardHandler(ctx context.Context,
	cardAction *larkcard.CardAction) (interface{}, error) {
	messageHandler := NewCardHandler(m)
	return messageHandler(ctx, cardAction)
}

func judgeMsgType(event *larkim.P2MessageReceiveV1) (string, error) {
	msgType := event.Event.Message.MessageType

	switch *msgType {
	case "text", "image", "audio":
		return *msgType, nil
	default:
		return "", fmt.Errorf("unknown message type: %v", *msgType)
	}

}

func (m MessageHandler) msgReceivedHandler(ctx context.Context, event *larkim.P2MessageReceiveV1) error {

	m.logger.Info("[receive]", zap.String("messageid", *event.Event.Message.MessageId), zap.String("message", *event.Event.Message.Content))
	// alert(ctx, fmt.Sprintf("收到消息: messageId %v", *event.Event.Message.MessageId))
	handlerType := judgeChatType(event)
	if handlerType == "otherChat" {
		replyMsg(ctx, "unknown chat type", event.Event.Message.MessageId)
		m.logger.Error("unknown chat type")
		return nil
	}
	//fmt.Println(larkcore.Prettify(event.Event.Message))

	msgType, err := judgeMsgType(event)
	if err != nil {
		replyMsg(ctx, "🥹不支持的消息类型, 当前仅支持文本消息、图片消息、语音消息", event.Event.Message.MessageId)
		m.logger.Error("error getting message type", zap.Error(err))
		return nil
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
	msgInfo := MsgInfo{
		handlerType: handlerType,
		msgType:     msgType,
		msgId:       msgId,
		chatId:      chatId,
		qParsed:     strings.Trim(parseContent(*content), " "),
		fileKey:     parseFileKey(*content),
		imageKey:    parseImageKey(*content),
		sessionId:   sessionId,
		mention:     mention,
	}
	data := &ActionInfo{
		ctx:     &ctx,
		handler: &m,
		info:    &msgInfo,
		logger:  m.logger,
	}
	actions := []Action{
		&ProcessedUniqueAction{}, //避免重复处理
		&ProcessMentionAction{},  //判断机器人是否应该被调用
		&AudioAction{},           //语音处理
		&EmptyAction{},           //空消息处理
		&ClearAction{},           //清除消息处理
		&PicAction{},             //图片处理
		&RoleListAction{},        //角色列表处理
		&HelpAction{},            //帮助处理
		&BalanceAction{},         //余额处理
		&RolePlayAction{},        //角色扮演处理
		&MessageAction{},         //消息处理

	}
	go chain(data, actions...)
	return nil
}

var _ MessageHandlerInterface = (*MessageHandler)(nil)

func NewMessageHandler(gpt *openai.ChatGPT,
	config initialization.Config, logger *zap.Logger) MessageHandlerInterface {
	return &MessageHandler{
		sessionCache: services.GetSessionCache(),
		msgCache:     services.GetMsgCache(),
		gpt:          gpt,
		config:       config,
		logger:       logger,
	}
}

func (m MessageHandler) judgeIfMentionMe(mention []*larkim.
	MentionEvent) bool {
	if len(mention) != 1 {
		return false
	}
	return *mention[0].Name == m.config.FeishuBotName
}
