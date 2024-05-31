package handlers

import (
	"context"

	"github.com/blacklee123/feishu-openai/initialization"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

type MsgInfo struct {
	cardId      *string
	handlerType HandlerType
	msgType     string
	msgId       *string
	chatId      *string
	qParsed     string
	fileKey     string
	imageKey    string
	imageKeys   []string // post 消息卡片中的图片组
	sessionId   *string
	mention     []*larkim.MentionEvent
}
type ActionInfo struct {
	handler *MessageHandler
	ctx     *context.Context
	info    *MsgInfo
	logger  *zap.Logger
	config  initialization.Config
}

type Action interface {
	Execute(a *ActionInfo) bool
}
