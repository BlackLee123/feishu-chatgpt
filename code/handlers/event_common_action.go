package handlers

import (
	"context"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"go.uber.org/zap"
)

type MsgInfo struct {
	newTopic    bool
	cardId      *string
	handlerType HandlerType
	msgType     string
	msgId       *string
	userId      *string
	chatId      *string
	qParsed     string
	fileKey     string
	imageKeys   []string // post 消息卡片中的图片组
	sessionId   *string
}
type ActionInfo struct {
	handler    *MessageHandler
	ctx        *context.Context
	info       *MsgInfo
	logger     *zap.Logger
	config     Config
	larkClient *lark.Client
}

type Action interface {
	Execute(a *ActionInfo) bool
}
