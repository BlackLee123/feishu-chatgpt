package handlers

import (
	"context"
	"fmt"
	"os"
	"time"

	myopenai "github.com/blacklee123/feishu-openai/services/openai"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	if a.info.newTopic {
		userName := ""
		user, err := a.retrieveUserInfo(*a.ctx, *a.info.userId)
		if err != nil {
			userName = *a.info.userId
		} else {
			userName = *user.Name
		}
		msg = append(msg, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(`你是ChatGPT-4, 一个被OpenAI训练出来的大语言模型。
			我的名字是%s, 请使用这个名字和我交流。`, userName),
			Name: "ChatGPT-4",
		})
	}
	if (a.info.msgType == "post" && a.info.imageKeys != nil && len(a.info.imageKeys) > 0) || a.info.msgType == "image" {
		var base64s []string
		for _, imageKey := range a.info.imageKeys {
			if imageKey == "" {
				continue
			}
			base64, err := a.downloadAndEncodeImage(imageKey, a.info.msgId)
			if err != nil {
				a.replyWithErrorMsg(*a.ctx, err, a.info.msgId)
				return false
			}
			base64s = append(base64s, base64)
		}
		msg = append(msg, createMultipleVisionMessages(a.info.qParsed, base64s, *a.info.userId))

	} else {
		msg = append(msg, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: a.info.qParsed,
			Name:    *a.info.userId,
		})
	}
	answer := ""
	if a.info.msgType == "audio" {
		answer = "🧑‍💻：" + a.info.qParsed + "\n" + "🤖: "
	}
	chatResponseStream := make(chan string)
	go func() {
		if err := a.handler.gpt.StreamChat(*a.ctx, msg, chatResponseStream); err != nil {
			a.logger.Error("StreamChat error", zap.Error(err))
			err := a.updateFinalCard(*a.ctx, "聊天失败", a.info.cardId, a.info.newTopic)
			if err != nil {
				a.logger.Error("updateFinalCard error", zap.Error(err))
				return
			}
		}
	}()
	timer := time.NewTicker(700 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			a.logger.Debug("answer", zap.String("answer", answer))
			if answer != "" {
				err := a.UpdateTextCard(*a.ctx, answer, a.info.cardId, a.info.newTopic)
				if err != nil {
					a.logger.Error("UpdateTextCard error", zap.Error(err))
				}
			}

		case res, ok := <-chatResponseStream:
			if ok {
				answer += res
			} else {
				timer.Stop()
				err := a.updateFinalCard(*a.ctx, answer, a.info.cardId, a.info.newTopic)
				if err != nil {
					a.logger.Error("updateFinalCard error", zap.Error(err))
					return false
				}
				msg := append(msg, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: answer,
					Name:    msg[0].Name,
				})
				a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)
				if a.info.msgType == "audio" {
					fileName := *a.info.msgId + ".opus"
					err := a.handler.gpt.TextToSpeech(*a.ctx, answer, fileName)
					if err != nil {
						a.logger.Error("TTS error", zap.Error(err))
						return false
					}
					f, err := os.Open(fileName)
					if err != nil {
						a.logger.Error("Error opening file", zap.Error(err))
						return false
					}
					defer f.Close()
					fileKey, err := a.uploadOpus(f, fileName)
					if err != nil {
						a.logger.Error("文件上传失败", zap.Error(err))
						return false
					}
					a.replyAudio(*a.ctx, fileKey, a.info.msgId)
				}
				return false
			}

		}
	}

}

func (a *ActionInfo) downloadAndEncodeImage(imageKey string, msgId *string) (string, error) {
	f := fmt.Sprintf("%s.png", imageKey)
	defer os.Remove(f)

	req := larkim.NewGetMessageResourceReqBuilder().MessageId(*msgId).FileKey(imageKey).Type("image").Build()
	resp, err := a.larkClient.Im.MessageResource.Get(context.Background(), req)
	if err != nil {
		return "", err
	}

	resp.WriteFile(f)
	return myopenai.GetBase64FromImage(f)
}

func (a *ActionInfo) replyWithErrorMsg(ctx context.Context, err error, msgId *string) {
	a.replyMsg(ctx, fmt.Sprintf("🤖️：图片下载失败，请稍后再试～\n 错误信息: %v", err), msgId)
}

func createMultipleVisionMessages(query string, base64Images []string, userId string) openai.ChatCompletionMessage {
	content := []openai.ChatMessagePart{{Type: "text", Text: query}}
	for _, base64Image := range base64Images {
		content = append(content, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: "data:image/jpeg;base64," + base64Image,
			},
		})
	}
	return openai.ChatCompletionMessage{
		Role:         openai.ChatMessageRoleUser,
		MultiContent: content,
		Name:         userId,
	}
}
