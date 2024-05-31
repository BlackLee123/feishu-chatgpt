package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/blacklee123/feishu-openai/initialization"
	myopenai "github.com/blacklee123/feishu-openai/services/openai"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	//if new topic
	var ifNewTopic bool
	if len(msg) <= 1 {
		ifNewTopic = true
	} else {
		ifNewTopic = false
	}
	if a.info.msgType == "post" && a.info.imageKeys != nil && len(a.info.imageKeys) > 0 {
		var base64s []string
		for _, imageKey := range a.info.imageKeys {
			if imageKey == "" {
				continue
			}
			base64, err := downloadAndEncodeImage(imageKey, a.info.msgId)
			if err != nil {
				replyWithErrorMsg(*a.ctx, err, a.info.msgId)
				return false
			}
			base64s = append(base64s, base64)
		}
		msg = append(msg, createMultipleVisionMessages(a.info.qParsed, base64s))

	} else {
		msg = append(msg, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser, Content: a.info.qParsed,
		})
	}
	answer := ""
	chatResponseStream := make(chan string)
	go func() {
		if err := a.handler.gpt.StreamChat(*a.ctx, msg, chatResponseStream); err != nil {
			fmt.Printf("StreamChat error: %v\n", err)
			err := updateFinalCard(*a.ctx, "聊天失败", a.info.cardId, ifNewTopic)
			if err != nil {
				fmt.Printf("updateFinalCard error: %v\n", err)
				return
			}
		}
	}()
	timer := time.NewTicker(700 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			log.Println("answer:", answer)
			if answer != "" {
				err := UpdateTextCard(*a.ctx, answer, a.info.cardId, ifNewTopic)
				if err != nil {
					fmt.Printf("UpdateTextCard error: %v\n", err)
				}
			}

		case res, ok := <-chatResponseStream:
			if ok {
				answer += res
			} else {
				fmt.Println("chatResponseStream closed")
				timer.Stop()
				err := updateFinalCard(*a.ctx, answer, a.info.cardId, ifNewTopic)
				if err != nil {
					fmt.Printf("updateFinalCard error: %v\n", err)
					return false
				}
				msg := append(msg, openai.ChatCompletionMessage{
					Role: "assistant", Content: answer,
				})
				a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)
				if a.info.msgType == "audio" {
					fileName := *a.info.msgId + ".opus"
					err := a.handler.gpt.TextToSpeech(answer, fileName)
					if err != nil {
						return false
					}
					f, err := os.Open(fileName)
					if err != nil {
						fmt.Println("Error opening file:", err)
						return false
					}
					defer f.Close()
					fileKey, err := uploadOpus(f, fileName)
					if err != nil {
						a.logger.Error("文件上传失败", zap.Error(err))
					}
					replyAudio(*a.ctx, fileKey, a.info.msgId)
				}
				return false
			}

		}
	}

}

func downloadAndEncodeImage(imageKey string, msgId *string) (string, error) {
	f := fmt.Sprintf("%s.png", imageKey)
	defer os.Remove(f)

	req := larkim.NewGetMessageResourceReqBuilder().MessageId(*msgId).FileKey(imageKey).Type("image").Build()
	resp, err := initialization.GetLarkClient().Im.MessageResource.Get(context.Background(), req)
	if err != nil {
		return "", err
	}

	resp.WriteFile(f)
	return myopenai.GetBase64FromImage(f)
}

func replyWithErrorMsg(ctx context.Context, err error, msgId *string) {
	replyMsg(ctx, fmt.Sprintf("🤖️：图片下载失败，请稍后再试～\n 错误信息: %v", err), msgId)
}

func createMultipleVisionMessages(query string, base64Images []string) openai.ChatCompletionMessage {
	content := []openai.ChatMessagePart{{Type: "text", Text: query}}
	for _, base64Image := range base64Images {
		content = append(content, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: "data:image/jpeg;base64," + base64Image,
			},
		})
	}
	return openai.ChatCompletionMessage{Role: "user", MultiContent: content}
}
