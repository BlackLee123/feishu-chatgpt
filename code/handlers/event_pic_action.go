package handlers

import (
	"fmt"

	"github.com/blacklee123/feishu-openai/utils"
)

type PicAction struct { /*图片*/
}

func (*PicAction) Execute(a *ActionInfo) bool {
	// 正则表达式匹配 `/picture 图片描述`, 并提取图片描述
	if matched, prompt := utils.MatchPicture(a.info.qParsed); matched {
		bs64, err := a.handler.gpt.GenerateOneImage(prompt, *a.info.userId)
		if err != nil {
			replyMsg(*a.ctx, fmt.Sprintf(
				"🤖️：图片生成失败，请稍后再试～\n错误信息: %v", err), a.info.msgId)
			return false
		}
		UpdateImageCard(*a.ctx, bs64, a.info.cardId, a.info.sessionId, a.info.qParsed)
		// replayImageCardByBase64(*a.ctx, bs64, a.info.msgId, a.info.sessionId, a.info.qParsed)
		return false
	}

	return true
}
