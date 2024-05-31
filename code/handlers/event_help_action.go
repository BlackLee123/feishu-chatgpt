package handlers

import (
	"fmt"

	"github.com/blacklee123/feishu-openai/utils"
)

type HelpAction struct { /*帮助*/
}

func (*HelpAction) Execute(a *ActionInfo) bool {
	if len(a.info.qParsed) == 0 {
		sendMsg(*a.ctx, "🤖️：你想知道什么呢~", a.info.chatId)
		fmt.Println("msgId", *a.info.msgId,
			"message.text is empty")

		return false
	}
	if _, foundHelp := utils.EitherTrimEqual(a.info.qParsed, "/help", "帮助"); foundHelp {
		sendHelpCard(*a.ctx, a.info.sessionId, a.info.msgId)
		return false
	}
	return true
}
