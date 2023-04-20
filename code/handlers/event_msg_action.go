package handlers

import (
	"fmt"
	"start-feishubot/services/openai"

	"go.uber.org/zap"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	a.logger.Debug("[MessageAction] messageid:", zap.String("messageid", *a.info.msgId))
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	msg = append(msg, openai.Messages{
		Role: "user", Content: a.info.qParsed,
	})
	a.logger.Debug("[开始处理]", zap.String("messageid", *a.info.msgId))
	completions, err := a.handler.gpt.Completions(msg)
	if err != nil {
		a.logger.Error("============================== openai处理失败 ============================== \n", zap.String("messageid", *a.info.msgId), zap.Error(err))

		alert(*a.ctx, fmt.Sprintf("openai处理失败: messageId %v", *a.info.msgId))
		replyMsg(*a.ctx, fmt.Sprintf(
			"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}
	a.logger.Info("[reply]", zap.String("messageid", *a.info.msgId), zap.String("reply", completions.Content))
	msg = append(msg, completions)
	a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)
	//if new topic
	if len(msg) == 2 {
		//fmt.Println("new topic", msg[1].Content)
		sendNewTopicCard(*a.ctx, a.info.sessionId, a.info.msgId,
			completions.Content)
		return false
	}
	err = replyMsg(*a.ctx, completions.Content, a.info.msgId)
	if err != nil {
		a.logger.Error("============================== 消息回复失败 ==============================", zap.String("messageid", *a.info.msgId))
		replyMsg(*a.ctx, fmt.Sprintf(
			"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}
	return true
}
