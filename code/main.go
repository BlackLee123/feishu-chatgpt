package main

import (
	"context"

	"github.com/blacklee123/feishu-openai/handlers"
	"github.com/blacklee123/feishu-openai/initialization"
	"github.com/blacklee123/feishu-openai/services/openai"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/spf13/pflag"

	sdkginext "github.com/larksuite/oapi-sdk-gin"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
)

func main() {
	// configure logging
	logger, _ := initZap("debug")
	defer logger.Sync()
	stdLog := zap.RedirectStdLog(logger)
	defer stdLog()

	initialization.InitRoleList()
	pflag.Parse()
	config := initialization.GetConfig()
	initialization.LoadLarkClient(*config)
	gpt := openai.NewChatGPT(*config)
	handler := handlers.NewMessageHandler(gpt, *config, logger)

	eventHandler := dispatcher.NewEventDispatcher(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey).
		OnP2MessageReceiveV1(handler.MsgReceivedHandler)

	cardHandler := larkcard.NewCardActionHandler(
		config.FeishuAppVerificationToken,
		config.FeishuAppEncryptKey,
		handler.CardHandler)

	gin.ForceConsoleColor()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/webhook/card",
		sdkginext.NewCardActionHandlerFunc(
			cardHandler))
	go func() {
		larkWsClient := larkws.NewClient(config.FeishuAppId, config.FeishuAppSecret, larkws.WithEventHandler(eventHandler), larkws.WithLogLevel(larkcore.LogLevelDebug))
		err := larkWsClient.Start(context.Background())
		if err != nil {
			logger.Fatal("larkws  启动失败", zap.Error(err))
		}
	}()
	err := initialization.StartServer(*config, r)
	if err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}

}

func initZap(logLevel string) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	}

	zapEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zapEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return zapConfig.Build()
}
