package handlers

import (
	"fmt"
	"start-feishubot/services/openai"
	"time"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	fmt.Printf("[MessageAction] messageid:%v \n", *a.info.msgId)
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	msg = append(msg, openai.Messages{
		Role: "user", Content: a.info.qParsed,
	})
	fmt.Printf("[开始处理] messageid:%v time: %v msg: %v \n", *a.info.msgId, time.Now(), msg)
	completions, err := a.handler.gpt.Completions(msg)
	if err != nil {
		fmt.Printf("============================== messageid:%v openai处理失败 ============================== \n", *a.info.msgId)
		replyMsg(*a.ctx, fmt.Sprintf(
			"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}
	fmt.Printf("[reply] messageid:%v time: %v reply: %v \n", *a.info.msgId, time.Now(), completions.Content)
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
		fmt.Printf("============================== messageid:%v 消息回复失败 ============================== \n", *a.info.msgId)
		replyMsg(*a.ctx, fmt.Sprintf(
			"🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}
	return true
}
