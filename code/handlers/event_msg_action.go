package handlers

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"

	"go.uber.org/zap"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	if a.handler.config.StreamMode {
		return true
	}
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	// 如果没有提示词，默认模拟ChatGPT
	// msg = setDefaultPrompt(msg)
	msg = append(msg, openai.ChatCompletionMessage{
		Role: "user", Content: a.info.qParsed,
	})

	//fmt.Println("msg", msg)
	//logger.Debug("msg", msg)
	// get ai mode as temperature
	aiMode := a.handler.sessionCache.GetAIMode(*a.info.sessionId)
	completions, err := a.handler.gpt.Completions(msg, aiMode)
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
	} else {
		sendOldTopicCard(*a.ctx, a.info.sessionId, a.info.msgId,
			completions.Content)
	}
	return true
}
