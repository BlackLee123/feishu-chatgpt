package handlers

import (
	"fmt"
	"os"

	"github.com/blacklee123/feishu-openai/utils/audio"
)

type AudioAction struct { /*语音*/
}

func (*AudioAction) Execute(a *ActionInfo) bool {
	// 只有私聊才解析语音,其他不解析
	if a.info.handlerType != UserHandler || a.info.msgType != "audio" {
		return true
	}

	fileKey := a.info.fileKey
	msgId := a.info.msgId

	f, err := a.downloadOpus(fileKey, msgId)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer os.Remove(f)

	output := fmt.Sprintf("%s.mp3", fileKey)
	audio.OggToWavByPath(f, output)
	defer os.Remove(output)

	text, err := a.handler.gpt.AudioToText(*a.ctx, output)
	if err != nil {
		fmt.Println(err)

		a.sendMsg(*a.ctx, fmt.Sprintf("🤖️：语音转换失败，请稍后再试～\n错误信息: %v", err), a.info.msgId)
		return false
	}

	// replyMsg(*a.ctx, fmt.Sprintf("🤖️：%s", text), a.info.msgId)
	//fmt.Println("text: ", text)
	a.info.qParsed = text
	if len(a.info.qParsed) == 0 {
		a.sendMsg(*a.ctx, "🤖️：你想知道什么呢~", a.info.chatId)
		fmt.Println("msgId", *a.info.msgId,
			"message.text is empty")

		return false
	}
	return true
}
