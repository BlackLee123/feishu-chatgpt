package handlers

type PreAction struct { /*图片*/
}

func (*PreAction) Execute(a *ActionInfo) bool {

	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)

	//if new topic
	var ifNewTopic bool
	if len(msg) <= 0 {
		ifNewTopic = true
	} else {
		ifNewTopic = false
	}

	cardId, err := a.sendOnProcessCard(*a.ctx, a.info.sessionId, a.info.msgId, ifNewTopic)
	if err != nil {
		return false
	}
	a.info.cardId = cardId
	a.info.newTopic = ifNewTopic
	return true
}
