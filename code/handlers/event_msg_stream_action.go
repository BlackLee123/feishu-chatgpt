package handlers

import (
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

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
		Role: openai.ChatMessageRoleUser, Content: a.info.qParsed,
	})
	//if new topic
	var ifNewTopic bool
	if len(msg) <= 1 {
		ifNewTopic = true
	} else {
		ifNewTopic = false
	}

	cardId, err := sendOnProcessCard(*a.ctx, a.info.sessionId, a.info.msgId, ifNewTopic)
	if err != nil {
		return false
	}

	answer := ""
	chatResponseStream := make(chan string)
	done := make(chan struct{}) // 添加 done 信号，保证 goroutine 正确退出
	go func() {
		aiMode := a.handler.sessionCache.GetAIMode(*a.info.sessionId)
		if err := a.handler.gpt.StreamChat(*a.ctx, msg, aiMode, chatResponseStream, done); err != nil {
			fmt.Printf("StreamChat error: %v\n", err)
			err := updateFinalCard(*a.ctx, "聊天失败", cardId, ifNewTopic)
			if err != nil {
				fmt.Printf("updateFinalCard error: %v\n", err)
				return
			}
		}
	}()

	t := time.NewTicker(700 * time.Millisecond)
	defer t.Stop()
	
	for {
		select {
		case <-t.C:
			err := UpdateTextCard(*a.ctx, answer, cardId, ifNewTopic)
			if err != nil {
				fmt.Printf("UpdateTextCard error: %v\n", err)
				return false
			}
		case res, ok := <-chatResponseStream:
			if !ok {
				fmt.Println("chatResponseStream closed")
				return false
			}
			answer += res

		case <-done: // 添加 done 信号的处理
			err := updateFinalCard(*a.ctx, answer, cardId, ifNewTopic)
			if err != nil {
				fmt.Printf("updateFinalCard error: %v\n", err)
				return false
			}
			msg := append(msg, openai.ChatCompletionMessage{
				Role: "assistant", Content: answer,
			})
			a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)
			close(chatResponseStream)
			return false
		}
	}
}
