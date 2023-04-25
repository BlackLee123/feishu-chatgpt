package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"start-feishubot/initialization"
	"start-feishubot/services/openai"
	"start-feishubot/utils/audio"

	"github.com/creack/pty"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

type MessageAction struct { /*消息*/
}

func (*MessageAction) Execute(a *ActionInfo) bool {
	a.logger.Debug("[MessageAction] messageid:", zap.String("messageid", *a.info.msgId))
	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	msg = append(msg, openai.Messages{
		Role: "user", Content: a.info.qParsed,
	})
	a.logger.Debug("[开始处理]", zap.String("messageid", *a.info.msgId))
	// get ai mode as temperature
	aiMode := a.handler.sessionCache.GetAIMode(*a.info.sessionId)
	completions, err := a.handler.gpt.Completions(msg, aiMode, a.config.OpenaiModel)
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
	if a.info.msgType == "audio" && a.config.XFAppId != "" && a.config.XFApiSecret != "" && a.config.XFApiKey != "" {
		fmt.Printf("msgId: %v content: %v\n", *a.info.msgId, completions.Content)
		err = audio.TextToAudio(*a.info.msgId, completions.Content, a.config.XFAppId, a.config.XFApiSecret, a.config.XFApiKey)
		if err == nil {
			cmd := exec.Command("ffmpeg", "-i", fmt.Sprintf("%s.mp3", *a.info.msgId), "-acodec", "libopus", "-ac", "1", "-ar", "16000", fmt.Sprintf("%s.opus", *a.info.msgId))
			fmt.Println(cmd.String())
			pf, err := pty.Start(cmd)
			if err != nil {
				fmt.Println("error", err)
			}
			io.Copy(os.Stdout, pf)
			f, _ := os.Open(fmt.Sprintf("%s.opus", *a.info.msgId))
			defer f.Close()
			audioReq := larkim.NewCreateFileReqBuilder().
				Body(larkim.NewCreateFileReqBodyBuilder().
					FileType("opus").
					FileName(fmt.Sprintf("%s.opus", *a.info.msgId)).
					File(f).
					Build()).
				Build()
			resp, err := initialization.GetLarkClient().Im.File.Create(context.Background(), audioReq)
			if err == nil && resp.Success() {
				replyAudio(*a.ctx, *resp.Data.FileKey, a.info.msgId)
			} else {
				a.logger.Error("文件上传失败", zap.Error(err))
			}
		} else {
			a.logger.Error("讯飞转换失败", zap.String("msgType", a.info.msgType), zap.Error(err))
		}
	}
	return true
}
