package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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
	if len(msg) == 3 {
		//fmt.Println("new topic", msg[1].Content)
		sendNewTopicCard(*a.ctx, a.info.sessionId, a.info.msgId,
			completions.Content)
		return false
	}
	if len(msg) != 3 {
		sendOldTopicCard(*a.ctx, a.info.sessionId, a.info.msgId,
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
	a.logger.Info("[msgType]", zap.String("msgType", a.info.msgType))
	return true
}

// 判断msg中的是否包含system role
func hasSystemRole(msg []openai.ChatCompletionMessage) bool {
	for _, m := range msg {
		if m.Role == "system" {
			return true
		}
	}
	return false
}

type StreamMessageAction struct { /*消息*/
}

func (m *StreamMessageAction) Execute(a *ActionInfo) bool {
	if !a.handler.config.StreamMode {
		return true
	}
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	// 如果没有提示词，默认模拟ChatGPT
	// msg = setDefaultPrompt(msg)
	msg = append(msg, openai.ChatCompletionMessage{
		Role: "user", Content: a.info.qParsed,
	})
	//if new topic
	var ifNewTopic bool
	if len(msg) <= 2 {
		ifNewTopic = true
	} else {
		ifNewTopic = false
	}

	cardId, err2 := sendOnProcess(a, ifNewTopic)
	if err2 != nil {
		return false
	}

	answer := ""
	chatResponseStream := make(chan string)
	done := make(chan struct{}) // 添加 done 信号，保证 goroutine 正确退出
	noContentTimeout := time.AfterFunc(10*time.Second, func() {
		log.Println("no content timeout")
		close(done)
		err := updateFinalCard(*a.ctx, "请求超时", cardId, ifNewTopic)
		if err != nil {
			return
		}
		return
	})
	defer noContentTimeout.Stop()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				err := updateFinalCard(*a.ctx, "聊天失败", cardId, ifNewTopic)
				if err != nil {
					return
				}
			}
		}()

		//log.Printf("UserId: %s , Request: %s", a.info.userId, msg)
		aiMode := a.handler.sessionCache.GetAIMode(*a.info.sessionId)
		if err := a.handler.gpt.StreamChat(*a.ctx, msg, aiMode,
			chatResponseStream); err != nil {
			err := updateFinalCard(*a.ctx, "聊天失败", cardId, ifNewTopic)
			if err != nil {
				return
			}
			close(done) // 关闭 done 信号
		}

		close(done) // 关闭 done 信号
	}()
	ticker := time.NewTicker(700 * time.Millisecond)
	defer ticker.Stop() // 注意在函数结束时停止 ticker
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := updateTextCard(*a.ctx, answer, cardId, ifNewTopic)
				if err != nil {
					return
				}
			}
		}
	}()
	for {
		select {
		case res, ok := <-chatResponseStream:
			if !ok {
				return false
			}
			noContentTimeout.Stop()
			answer += res
			//pp.Println("answer", answer)
		case <-done: // 添加 done 信号的处理
			err := updateFinalCard(*a.ctx, answer, cardId, ifNewTopic)
			if err != nil {
				return false
			}
			ticker.Stop()
			msg := append(msg, openai.ChatCompletionMessage{
				Role: "assistant", Content: answer,
			})
			a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)
			close(chatResponseStream)
			log.Printf("\n\n\n")
			jsonByteArray, err := json.Marshal(msg)
			if err != nil {
				log.Println(err)
			}
			jsonStr := strings.ReplaceAll(string(jsonByteArray), "\\n", "")
			jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
			log.Printf("\n\n\n")
			return false
		}
	}
}

func sendOnProcess(a *ActionInfo, ifNewTopic bool) (*string, error) {
	// send 正在处理中
	cardId, err := sendOnProcessCard(*a.ctx, a.info.sessionId,
		a.info.msgId, ifNewTopic)
	if err != nil {
		return nil, err
	}
	return cardId, nil

}
